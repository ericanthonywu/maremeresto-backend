package websocket

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait       = 10 * time.Second
	pongWait        = 60 * time.Second
	pingPeriod      = (pongWait * 9) / 10
	maxMessageSize  = 512
	maxRoomsPerConn = 24
)

// Room name prefixes. Every room is access-controlled by Authorizer.
const (
	RoomOwner     = "owner"   // network-wide feed, owner only
	RoomBranchPfx = "branch:" // one outlet's order feed, that branch's admin or the owner
	RoomOrderPfx  = "order:"  // a single order's status, the customer who placed it or staff
)

// Identity is the authenticated principal behind a socket. A nil Identity is
// an anonymous connection, which may only join rooms the Authorizer allows
// without credentials.
type Identity struct {
	UserID   uuid.UUID
	Role     string // customer, branch_admin, owner
	BranchID *uuid.UUID
}

// Authorizer decides whether a connection may subscribe to a room. It is
// supplied by the service layer because ownership of an order is a database
// question, not a transport one.
type Authorizer func(id *Identity, room string) bool

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	identity *Identity

	mu    sync.Mutex
	rooms map[string]bool
}

type Message struct {
	Room    string `json:"room"`
	Event   string `json:"event"` // new_order, status_updated, order_paid, connected, joined, error
	Payload any    `json:"payload"`
}

type Hub struct {
	clients    map[*Client]bool
	rooms      map[string]map[*Client]bool
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	authorize  Authorizer
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		rooms:      make(map[string]map[*Client]bool),
		broadcast:  make(chan Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		// Deny by default until the service layer installs a real policy, so a
		// wiring mistake cannot fall open.
		authorize: func(*Identity, string) bool { return false },
	}
}

// SetAuthorizer installs the room access policy. Called once during startup.
func (h *Hub) SetAuthorizer(a Authorizer) {
	h.mu.Lock()
	h.authorize = a
	h.mu.Unlock()
}

func (h *Hub) authorizerFn() Authorizer {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.authorize
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.mu.Lock()
				for room := range client.rooms {
					if clients, ok := h.rooms[room]; ok {
						delete(clients, client)
						if len(clients) == 0 {
							delete(h.rooms, room)
						}
					}
				}
				client.mu.Unlock()
				close(client.send)
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.deliver(msg)
		}
	}
}

// deliver fans a message out to exactly one room. There is deliberately no
// "broadcast to every socket" path: order payloads carry customer names,
// phone numbers and addresses, and must only reach subscribers of a room they
// were authorised for.
func (h *Hub) deliver(msg Message) {
	if msg.Room == "" {
		slog.Warn("dropping websocket broadcast with no room", "event", msg.Event)
		return
	}

	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("failed to marshal websocket message", "err", err, "event", msg.Event)
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	clients, ok := h.rooms[msg.Room]
	if !ok {
		return
	}
	for client := range clients {
		select {
		case client.send <- data:
		default:
			// Slow consumer: drop it rather than block the hub.
			close(client.send)
			delete(clients, client)
			delete(h.clients, client)
		}
	}
	if len(clients) == 0 {
		delete(h.rooms, msg.Room)
	}
}

// JoinRoom subscribes a client after checking the authorizer. It reports
// whether the join was permitted.
func (h *Hub) JoinRoom(client *Client, room string) bool {
	room = strings.TrimSpace(room)
	if room == "" || len(room) > 128 {
		return false
	}
	if !h.authorizerFn()(client.identity, room) {
		return false
	}

	client.mu.Lock()
	if len(client.rooms) >= maxRoomsPerConn && !client.rooms[room] {
		client.mu.Unlock()
		return false
	}
	client.rooms[room] = true
	client.mu.Unlock()

	h.mu.Lock()
	if _, ok := h.rooms[room]; !ok {
		h.rooms[room] = make(map[*Client]bool)
	}
	h.rooms[room][client] = true
	h.mu.Unlock()
	return true
}

// Broadcast queues a message for one room. It never blocks the caller: if the
// hub is saturated the message is dropped and logged, because an order must
// still be committed even when the realtime feed is congested.
func (h *Hub) Broadcast(room string, event string, payload any) {
	select {
	case h.broadcast <- Message{Room: room, Event: event, Payload: payload}:
	default:
		slog.Warn("websocket broadcast buffer full, dropping event", "room", room, "event", event)
	}
}

// BranchRoom and OrderRoom centralise room naming so producers and the
// authorizer cannot drift apart.
func BranchRoom(branchID uuid.UUID) string { return RoomBranchPfx + branchID.String() }
func OrderRoom(orderID uuid.UUID) string   { return RoomOrderPfx + orderID.String() }

func (c *Client) writeJSON(v any) {
	data, err := json.Marshal(v)
	if err != nil {
		return
	}
	select {
	case c.send <- data:
	default:
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Warn("WebSocket unexpected close", "err", err)
			}
			return
		}

		var req struct {
			Action string `json:"action"`
			Room   string `json:"room"`
		}
		if err := json.Unmarshal(message, &req); err != nil {
			continue
		}

		switch req.Action {
		case "join":
			if c.hub.JoinRoom(c, req.Room) {
				c.writeJSON(Message{Room: req.Room, Event: "joined"})
			} else {
				c.writeJSON(Message{Room: req.Room, Event: "error", Payload: map[string]string{
					"message": "tidak memiliki akses ke kanal ini",
				}})
			}
		case "ping":
			c.writeJSON(Message{Event: "pong"})
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			// One frame per message: coalescing several JSON documents into a
			// single frame separated by newlines breaks JSON.parse on the client.
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWs upgrades the request. allowedOrigins restricts which sites may open
// a socket; an empty list allows same-origin requests only.
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request, identity *Identity, allowedOrigins []string) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     originChecker(allowedOrigins),
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("WebSocket upgrade error", "err", err)
		return
	}

	client := &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		identity: identity,
		rooms:    make(map[string]bool),
	}
	hub.register <- client

	go client.writePump()

	// Auto-join the room named in the query string, if the caller is allowed it.
	if room := r.URL.Query().Get("room"); room != "" {
		if hub.JoinRoom(client, room) {
			client.writeJSON(Message{Room: room, Event: "joined"})
		} else {
			client.writeJSON(Message{Room: room, Event: "error", Payload: map[string]string{
				"message": "tidak memiliki akses ke kanal ini",
			}})
		}
	}

	role := "anonymous"
	if identity != nil {
		role = identity.Role
	}
	client.writeJSON(Message{Event: "connected", Payload: map[string]any{
		"role":      role,
		"timestamp": time.Now().Format(time.RFC3339),
	}})

	go client.readPump()
}

func originChecker(allowed []string) func(*http.Request) bool {
	normalized := make(map[string]bool, len(allowed))
	for _, o := range allowed {
		if o = strings.TrimRight(strings.TrimSpace(o), "/"); o != "" {
			normalized[strings.ToLower(o)] = true
		}
	}

	return func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			// Non-browser client (mobile app, health check): no Origin to forge.
			return true
		}
		origin = strings.ToLower(strings.TrimRight(origin, "/"))
		if normalized[origin] {
			return true
		}
		// Same-origin requests are always fine.
		if host := r.Host; host != "" {
			if origin == "http://"+strings.ToLower(host) || origin == "https://"+strings.ToLower(host) {
				return true
			}
		}
		slog.Warn("rejected websocket origin", "origin", origin)
		return false
	}
}
