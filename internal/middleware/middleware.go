package middleware

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/ericanthonywu/maremereso-olga/backend/internal/alert"
	"github.com/ericanthonywu/maremereso-olga/backend/internal/config"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/cors"
)

type contextKey string

const (
	UserContextKey        contextKey = "user"
	IdempotencyContextKey contextKey = "idempotency_key"
)

type JWTClaims struct {
	UserID   uuid.UUID  `json:"user_id"`
	Name     string     `json:"name,omitempty"`
	Phone    string     `json:"phone"`
	Role     string     `json:"role"`
	BranchID *uuid.UUID `json:"branch_id,omitempty"`
	jwt.RegisteredClaims
}

func writeAuthError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"success": false, "error": message})
}

func bearerToken(r *http.Request) string {
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
	}
	if cookie, err := r.Cookie("token"); err == nil {
		return cookie.Value
	}
	return ""
}

// ParseToken validates a signed token and returns its claims. HMAC is pinned
// explicitly so a token claiming "alg":"none" (or an asymmetric algorithm) can
// never be accepted.
func ParseToken(cfg *config.Config, tokenStr string) (*JWTClaims, error) {
	claims := &JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		return []byte(cfg.JWTSecret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}

// AuthRequired rejects the request unless a valid token is present and, when
// roles are given, the caller holds one of them.
func AuthRequired(cfg *config.Config, requiredRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := bearerToken(r)
			if tokenStr == "" {
				writeAuthError(w, http.StatusUnauthorized, "sesi tidak ditemukan, silakan masuk kembali")
				return
			}

			claims, err := ParseToken(cfg, tokenStr)
			if err != nil {
				writeAuthError(w, http.StatusUnauthorized, "sesi tidak valid atau telah berakhir, silakan masuk kembali")
				return
			}

			if len(requiredRoles) > 0 && !hasRole(claims.Role, requiredRoles) {
				writeAuthError(w, http.StatusForbidden, "akun Anda tidak memiliki izin untuk tindakan ini")
				return
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), UserContextKey, claims)))
		})
	}
}

// OptionalAuth attaches claims when a valid token is present but lets the
// request through either way. Used for guest checkout, where an order may be
// placed before the customer has a session.
func OptionalAuth(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if tokenStr := bearerToken(r); tokenStr != "" {
				if claims, err := ParseToken(cfg, tokenStr); err == nil {
					r = r.WithContext(context.WithValue(r.Context(), UserContextKey, claims))
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func hasRole(role string, required []string) bool {
	for _, want := range required {
		if role == want {
			return true
		}
	}
	// The owner is a superset of every branch-admin permission.
	if role == "owner" {
		for _, want := range required {
			if want == "branch_admin" {
				return true
			}
		}
	}
	return false
}

func GetUserFromContext(ctx context.Context) (*JWTClaims, bool) {
	claims, ok := ctx.Value(UserContextKey).(*JWTClaims)
	return claims, ok && claims != nil
}

func CORS(cfg *config.Config) func(http.Handler) http.Handler {
	origins := []string{cfg.CustomerURL, cfg.AdminURL}
	if !cfg.IsProduction() {
		// Vite dev servers, only outside production.
		origins = append(origins,
			"http://localhost:5173", "http://localhost:5174",
			"http://127.0.0.1:5173", "http://127.0.0.1:5174",
		)
	}

	cleaned := make([]string, 0, len(origins))
	seen := make(map[string]bool)
	for _, o := range origins {
		o = strings.TrimRight(strings.TrimSpace(o), "/")
		if o != "" && !seen[o] {
			seen[o] = true
			cleaned = append(cleaned, o)
		}
	}
	slog.Info("CORS allowed origins", "origins", cleaned)

	return cors.New(cors.Options{
		AllowedOrigins:   cleaned,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "Idempotency-Key"},
		ExposedHeaders:   []string{"Idempotency-Key"},
		AllowCredentials: true,
		MaxAge:           300,
	}).Handler
}

// ---------------------------------------------------------------------
// Rate limiting
// ---------------------------------------------------------------------

type rateBucket struct {
	tokens   float64
	lastSeen time.Time
}

// RateLimiter is a per-IP token bucket. It protects the endpoints that cost
// real money or reach a third party: login attempts, order creation, payment
// initiation and geocoder lookups.
type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*rateBucket
	capacity float64
	refill   float64 // tokens per second
}

func NewRateLimiter(burst int, perMinute int) *RateLimiter {
	rl := &RateLimiter{
		buckets:  make(map[string]*rateBucket),
		capacity: float64(burst),
		refill:   float64(perMinute) / 60.0,
	}
	go rl.reap()
	return rl
}

func (rl *RateLimiter) reap() {
	for range time.Tick(5 * time.Minute) {
		cutoff := time.Now().Add(-15 * time.Minute)
		rl.mu.Lock()
		for k, b := range rl.buckets {
			if b.lastSeen.Before(cutoff) {
				delete(rl.buckets, k)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) allow(key string) bool {
	now := time.Now()

	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, ok := rl.buckets[key]
	if !ok {
		rl.buckets[key] = &rateBucket{tokens: rl.capacity - 1, lastSeen: now}
		return true
	}

	b.tokens += now.Sub(b.lastSeen).Seconds() * rl.refill
	if b.tokens > rl.capacity {
		b.tokens = rl.capacity
	}
	b.lastSeen = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		if !rl.allow(clientIP(r)) {
			w.Header().Set("Retry-After", "30")
			writeAuthError(w, http.StatusTooManyRequests, "terlalu banyak permintaan, silakan tunggu sebentar")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func clientIP(r *http.Request) string {
	// chi's RealIP has already normalised X-Forwarded-For into RemoteAddr.
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

func Idempotency(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("Idempotency-Key")
		if key == "" {
			writeAuthError(w, http.StatusBadRequest, "Idempotency-Key header wajib untuk operasi pembayaran")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), IdempotencyContextKey, key)))
	})
}

// statusRecorder captures the response status so the access log is useful.
//
// It must forward Hijack and Flush to the wrapped writer: a ResponseWriter
// that does not implement http.Hijacker makes the WebSocket upgrade fail with
// "response does not implement http.Hijacker", which silently kills the entire
// live-order feed.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	if s.status == 0 {
		s.status = http.StatusOK
	}
	return s.ResponseWriter.Write(b)
}

// Hijack lets the WebSocket handler take over the connection.
func (s *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := s.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return hijacker.Hijack()
}

func (s *statusRecorder) Flush() {
	if flusher, ok := s.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Unwrap exposes the underlying writer to helpers such as
// http.ResponseController.
func (s *statusRecorder) Unwrap() http.ResponseWriter { return s.ResponseWriter }

func traceID(r *http.Request) string {
	if reqID := chimw.GetReqID(r.Context()); reqID != "" {
		return reqID
	}
	if reqID := r.Header.Get("X-Request-Id"); reqID != "" {
		return reqID
	}
	if reqID := r.Header.Get("X-Trace-Id"); reqID != "" {
		return reqID
	}
	return uuid.NewString()
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		tid := traceID(r)
		w.Header().Set("X-Trace-Id", tid)
		rec := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(rec, r)
		if rec.status == 0 {
			rec.status = http.StatusOK
		}
		slog.Info("HTTP Request",
			"trace_id", tid,
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration", time.Since(start),
			"ip", clientIP(r),
		)
	})
}

func Recovery(alertSvc *alert.AlertService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					tid := traceID(r)
					stack := string(debug.Stack())
					errMsg := fmt.Sprintf("%v", rec)
					slog.Error("PANIC recovered", "trace_id", tid, "error", rec, "path", r.URL.Path)

					if alertSvc != nil {
						alertSvc.SendErrorAlert(tid, errMsg, r.URL.Path, stack)
					}

					writeAuthError(w, http.StatusInternalServerError, "terjadi kesalahan pada server")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
