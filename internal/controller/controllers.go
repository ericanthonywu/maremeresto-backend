package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ericanthonywu/maremereso-olga/backend/internal/apperror"
	"github.com/ericanthonywu/maremereso-olga/backend/internal/config"
	"github.com/ericanthonywu/maremereso-olga/backend/internal/dto"
	"github.com/ericanthonywu/maremereso-olga/backend/internal/middleware"
	"github.com/ericanthonywu/maremereso-olga/backend/internal/service"
	ws "github.com/ericanthonywu/maremereso-olga/backend/internal/websocket"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// maxJSONBody caps request bodies so a single client cannot exhaust memory.
const maxJSONBody = 1 << 20 // 1 MiB

type Controller struct {
	svc *service.Service
	hub *ws.Hub
	cfg *config.Config
}

func NewController(svc *service.Service, hub *ws.Hub, cfg *config.Config) *Controller {
	return &Controller{svc: svc, hub: hub, cfg: cfg}
}

// ---------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func ok(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": data})
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	appErr := apperror.FromError(err)
	// 5xx hides the real cause from the client, so the log is the only place
	// it can be found. 4xx are expected outcomes and stay at debug.
	if appErr.Code >= 500 {
		slog.ErrorContext(r.Context(), "request failed", "status", appErr.Code, "method", r.Method, "path", r.URL.Path, "err", err)
	} else {
		slog.DebugContext(r.Context(), "request rejected", "status", appErr.Code, "method", r.Method, "path", r.URL.Path, "err", err)
	}
	writeJSON(w, appErr.Code, map[string]any{"success": false, "error": appErr.Error()})
}

// decode reads a JSON body, rejecting oversized payloads and unknown fields so
// a typo in the client surfaces as an error rather than a silently ignored value.
func decode(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBody)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return apperror.Invalid("data yang dikirim terlalu besar")
		}
		return apperror.Invalid(fmt.Sprintf("format data tidak valid: %v", err))
	}
	// Reject trailing content after the JSON document.
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return apperror.Invalid("format data tidak valid")
	}
	return nil
}

func urlUUID(r *http.Request, key string) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, key))
	if err != nil {
		return uuid.Nil, apperror.Invalid("identitas tidak valid")
	}
	return id, nil
}

func queryUUID(r *http.Request, key string) (*uuid.UUID, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, apperror.Invalid("parameter " + key + " tidak valid")
	}
	return &id, nil
}

func actor(r *http.Request) *middleware.JWTClaims {
	claims, _ := middleware.GetUserFromContext(r.Context())
	return claims
}

// ---------------------------------------------------------------------
// Auth
// ---------------------------------------------------------------------

func (c *Controller) CustomerPhoneLogin(w http.ResponseWriter, r *http.Request) {
	var req dto.CustomerPhoneLoginRequest
	if err := decode(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	if err := req.Validate(); err != nil {
		writeError(w, r, err)
		return
	}

	resp, err := c.svc.CustomerLogin(r.Context(), req.Phone, req.Name)
	if err != nil {
		writeError(w, r, err)
		return
	}
	ok(w, resp)
}

func (c *Controller) UpdateCustomerProfile(w http.ResponseWriter, r *http.Request) {
	claims := actor(r)
	if claims == nil || claims.Role != "customer" {
		writeError(w, r, apperror.ErrUnauthorized)
		return
	}
	var req dto.UpdateCustomerProfileRequest
	if err := decode(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	resp, err := c.svc.UpdateCustomerProfile(r.Context(), claims.UserID, &req)
	if err != nil {
		writeError(w, r, err)
		return
	}
	ok(w, resp)
}

func (c *Controller) AdminLogin(w http.ResponseWriter, r *http.Request) {
	var req dto.AdminLoginRequest
	if err := decode(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}

	resp, err := c.svc.AdminLogin(r.Context(), req.Identifier, req.Password)
	if err != nil {
		writeError(w, r, err)
		return
	}
	ok(w, resp)
}

func (c *Controller) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	claims := actor(r)
	if claims == nil {
		writeError(w, r, apperror.ErrUnauthorized)
		return
	}

	user, err := c.svc.GetUserByID(r.Context(), claims.UserID)
	if err == nil && user != nil {
		ok(w, dto.UserSummary{
			ID:       user.ID,
			Name:     user.Name,
			Phone:    user.Phone,
			Email:    user.Email,
			Username: user.Username,
			Role:     user.Role,
			BranchID: user.BranchID,
		})
		return
	}

	// Fallback to claims if user query fails
	ok(w, dto.UserSummary{
		ID:       claims.UserID,
		Phone:    claims.Phone,
		Role:     claims.Role,
		BranchID: claims.BranchID,
	})
}

// ---------------------------------------------------------------------
// Branches
// ---------------------------------------------------------------------

func (c *Controller) ListBranches(w http.ResponseWriter, r *http.Request) {
	branches, err := c.svc.ListBranches(r.Context())
	if err != nil {
		writeError(w, r, err)
		return
	}
	ok(w, branches)
}

func (c *Controller) GetBranch(w http.ResponseWriter, r *http.Request) {
	branch, err := c.svc.GetBranchBySlug(r.Context(), chi.URLParam(r, "slug"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	ok(w, branch)
}

func (c *Controller) UpdateBranchStatus(w http.ResponseWriter, r *http.Request) {
	branchID, err := urlUUID(r, "id")
	if err != nil {
		writeError(w, r, err)
		return
	}
	scoped, err := service.ResolveBranchScope(actor(r), &branchID)
	if err != nil || scoped == nil {
		writeError(w, r, apperror.ErrForbidden)
		return
	}

	var req struct {
		IsOpen bool `json:"is_open"`
	}
	if err := decode(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}

	if err := c.svc.UpdateBranchStatus(r.Context(), *scoped, req.IsOpen); err != nil {
		writeError(w, r, err)
		return
	}
	ok(w, map[string]any{"branch_id": *scoped, "is_open": req.IsOpen})
}

func (c *Controller) UpdateBranchProfile(w http.ResponseWriter, r *http.Request) {
	branchID, err := urlUUID(r, "id")
	if err != nil {
		writeError(w, r, err)
		return
	}
	scoped, err := service.ResolveBranchScope(actor(r), &branchID)
	if err != nil || scoped == nil {
		writeError(w, r, apperror.ErrForbidden)
		return
	}

	var req dto.UpdateBranchProfileRequest
	if err := decode(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}

	branch, err := c.svc.UpdateBranchProfile(r.Context(), *scoped, &req)
	if err != nil {
		writeError(w, r, err)
		return
	}
	ok(w, branch)
}

// ---------------------------------------------------------------------
// Menu & categories
// ---------------------------------------------------------------------

func (c *Controller) ListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := c.svc.ListCategories(r.Context())
	if err != nil {
		writeError(w, r, err)
		return
	}
	ok(w, categories)
}

func (c *Controller) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateCategoryRequest
	if err := decode(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}

	cat, err := c.svc.CreateCategory(r.Context(), &req)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "data": cat})
}

func (c *Controller) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	id, err := urlUUID(r, "id")
	if err != nil {
		writeError(w, r, err)
		return
	}

	var req dto.UpdateCategoryRequest
	if err := decode(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}

	cat, err := c.svc.UpdateCategory(r.Context(), id, &req)
	if err != nil {
		writeError(w, r, err)
		return
	}
	ok(w, cat)
}

func (c *Controller) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id, err := urlUUID(r, "id")
	if err != nil {
		writeError(w, r, err)
		return
	}

	if err := c.svc.DeleteCategory(r.Context(), id); err != nil {
		writeError(w, r, err)
		return
	}
	ok(w, map[string]any{"id": id})
}

func (c *Controller) ListMenuByBranch(w http.ResponseWriter, r *http.Request) {
	slugOrID := chi.URLParam(r, "branch")

	branchID, err := uuid.Parse(slugOrID)
	if err != nil {
		b, lookupErr := c.svc.GetBranchBySlug(r.Context(), slugOrID)
		if lookupErr != nil {
			writeError(w, r, lookupErr)
			return
		}
		branchID = b.ID
	}

	items, err := c.svc.ListMenuByBranch(r.Context(), branchID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	ok(w, items)
}

func (c *Controller) ListBestSellers(w http.ResponseWriter, r *http.Request) {
	branchQuery := r.URL.Query().Get("branch")
	var branchID *uuid.UUID
	if branchQuery != "" {
		id, err := uuid.Parse(branchQuery)
		if err != nil {
			b, lookupErr := c.svc.GetBranchBySlug(r.Context(), branchQuery)
			if lookupErr != nil {
				writeError(w, r, lookupErr)
				return
			}
			branchID = &b.ID
		} else {
			branchID = &id
		}
	}

	items, err := c.svc.ListBestSellers(r.Context(), branchID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	ok(w, items)
}

func (c *Controller) CreateMenuItem(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateMenuItemRequest
	if err := decode(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}

	item, err := c.svc.CreateMenuItem(r.Context(), actor(r), &req)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "data": item})
}

func (c *Controller) CreateMenuItemsBulk(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateMenuItemsBulkRequest
	if err := decode(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	items, err := c.svc.CreateMenuItemsBulk(r.Context(), actor(r), &req)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "data": items})
}

func (c *Controller) UpdateMenuItem(w http.ResponseWriter, r *http.Request) {
	id, err := urlUUID(r, "id")
	if err != nil {
		writeError(w, r, err)
		return
	}

	var req dto.UpdateMenuItemRequest
	if err := decode(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}

	item, err := c.svc.UpdateMenuItem(r.Context(), actor(r), id, &req)
	if err != nil {
		writeError(w, r, err)
		return
	}
	ok(w, item)
}

func (c *Controller) ToggleMenuItemAvailability(w http.ResponseWriter, r *http.Request) {
	id, err := urlUUID(r, "id")
	if err != nil {
		writeError(w, r, err)
		return
	}

	var req dto.ToggleAvailabilityRequest
	if err := decode(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}

	if err := c.svc.ToggleMenuAvailability(r.Context(), actor(r), id, req.IsAvailable); err != nil {
		writeError(w, r, err)
		return
	}
	ok(w, map[string]any{"id": id, "is_available": req.IsAvailable})
}

func (c *Controller) ToggleMenuItemBestSeller(w http.ResponseWriter, r *http.Request) {
	id, err := urlUUID(r, "id")
	if err != nil {
		writeError(w, r, err)
		return
	}

	var req dto.ToggleBestSellerRequest
	if err := decode(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}

	if err := c.svc.ToggleMenuBestSeller(r.Context(), actor(r), id, req.IsBestSeller); err != nil {
		writeError(w, r, err)
		return
	}
	ok(w, map[string]any{"id": id, "is_best_seller": req.IsBestSeller})
}

func (c *Controller) DeleteMenuItem(w http.ResponseWriter, r *http.Request) {
	id, err := urlUUID(r, "id")
	if err != nil {
		writeError(w, r, err)
		return
	}

	if err := c.svc.DeleteMenuItem(r.Context(), actor(r), id); err != nil {
		writeError(w, r, err)
		return
	}
	ok(w, map[string]any{"id": id})
}

// ---------------------------------------------------------------------
// Delivery quoting & geocoding
// ---------------------------------------------------------------------

func (c *Controller) QuoteDelivery(w http.ResponseWriter, r *http.Request) {
	var req dto.DeliveryQuoteRequest
	if err := decode(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}

	quote, err := c.svc.QuoteDelivery(r.Context(), &req)
	if err != nil {
		writeError(w, r, err)
		return
	}
	ok(w, quote)
}

func (c *Controller) SearchAddress(w http.ResponseWriter, r *http.Request) {
	scope := r.URL.Query().Get("scope")
	results, err := c.svc.SearchAddress(r.Context(), r.URL.Query().Get("q"), scope)
	if err != nil {
		writeError(w, r, err)
		return
	}
	ok(w, results)
}

func (c *Controller) ReverseGeocode(w http.ResponseWriter, r *http.Request) {
	lat, errLat := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lon, errLon := strconv.ParseFloat(r.URL.Query().Get("lon"), 64)
	if errLat != nil || errLon != nil {
		writeError(w, r, apperror.Invalid("koordinat tidak valid"))
		return
	}

	result, err := c.svc.ReverseGeocode(r.Context(), lat, lon)
	if err != nil {
		writeError(w, r, err)
		return
	}
	ok(w, result)
}

// ---------------------------------------------------------------------
// Orders
// ---------------------------------------------------------------------

func (c *Controller) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateOrderRequest
	if err := decode(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}

	var userID *uuid.UUID
	if claims := actor(r); claims != nil {
		userID = &claims.UserID
	}

	order, err := c.svc.CreateOrder(r.Context(), userID, &req)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "data": order})
}

// GetOrder is restricted to the customer who placed the order and to staff of
// the branch fulfilling it. It used to be fully public, so anyone holding an
// order id could read the customer's name, phone number and home address.
func (c *Controller) GetOrder(w http.ResponseWriter, r *http.Request) {
	id, err := urlUUID(r, "id")
	if err != nil {
		writeError(w, r, err)
		return
	}

	order, err := c.svc.GetOrder(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if !c.svc.CanAccessOrder(actor(r), order) {
		// 404 rather than 403: confirming an id exists is itself a small leak.
		writeError(w, r, apperror.ErrNotFound)
		return
	}

	// Reconcile pending payments if the customer opens tracking/success before the Midtrans webhook arrives.
	if order.Status == "pending" && order.Payment != nil && order.Payment.Status == "pending" {
		if synced, syncErr := c.svc.SyncPaymentWithMidtrans(r.Context(), order.OrderNumber); syncErr == nil && synced != nil && synced.Status == "settlement" {
			if refreshed, rErr := c.svc.GetOrder(r.Context(), id); rErr == nil && refreshed != nil {
				order = refreshed
			}
		}
	}

	ok(w, order)
}

func (c *Controller) ListMyOrders(w http.ResponseWriter, r *http.Request) {
	claims := actor(r)
	if claims == nil {
		writeError(w, r, apperror.ErrUnauthorized)
		return
	}

	limit, offset := paging(r)
	orders, total, err := c.svc.ListOrdersForCustomer(r.Context(), claims.UserID, limit, offset)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": orders, "total": total})
}

func parseDateParam(raw string, endOfDay bool) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	if t, err := time.ParseInLocation("2006-01-02", raw, service.OutletLocation); err == nil {
		if endOfDay {
			nextDay := t.AddDate(0, 0, 1)
			return &nextDay, nil
		}
		return &t, nil
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return &t, nil
	}
	return nil, apperror.Invalid("format tanggal tidak valid (gunakan YYYY-MM-DD)")
}

func (c *Controller) ListOrders(w http.ResponseWriter, r *http.Request) {
	requested, err := queryUUID(r, "branch_id")
	if err != nil {
		writeError(w, r, err)
		return
	}
	branchID, err := service.ResolveBranchScope(actor(r), requested)
	if err != nil {
		writeError(w, r, err)
		return
	}

	dateFrom, err := parseDateParam(r.URL.Query().Get("date_from"), false)
	if err != nil {
		writeError(w, r, err)
		return
	}
	dateTo, err := parseDateParam(r.URL.Query().Get("date_to"), true)
	if err != nil {
		writeError(w, r, err)
		return
	}
	includeActive := r.URL.Query().Get("include_active") == "1" || r.URL.Query().Get("include_active") == "true"

	limit, offset := paging(r)
	orders, total, err := c.svc.ListOrders(
		r.Context(), branchID,
		r.URL.Query().Get("status"),
		r.URL.Query().Get("search"),
		dateFrom, dateTo, includeActive,
		limit, offset,
	)
	if err != nil {
		writeError(w, r, err)
		return
	}

	unread, err := c.svc.CountUnacknowledgedOrders(r.Context(), branchID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	statusCounts, err := c.svc.CountOrdersByStatus(r.Context(), branchID, dateFrom, dateTo, includeActive)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success":       true,
		"data":          orders,
		"total":         total,
		"unread":        unread,
		"status_counts": statusCounts,
	})
}

func (c *Controller) SubmitOrderFeedback(w http.ResponseWriter, r *http.Request) {
	id, err := urlUUID(r, "id")
	if err != nil {
		writeError(w, r, err)
		return
	}
	order, err := c.svc.GetOrder(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if !c.svc.CanAccessOrder(actor(r), order) {
		writeError(w, r, apperror.ErrNotFound)
		return
	}
	var req dto.OrderFeedbackRequest
	if err := decode(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	feedback, err := c.svc.SubmitOrderFeedback(r.Context(), id, &req)
	if err != nil {
		writeError(w, r, err)
		return
	}
	ok(w, feedback)
}

func (c *Controller) ListFeedback(w http.ResponseWriter, r *http.Request) {
	requested, _ := queryUUID(r, "branch_id")
	branchID, err := service.ResolveBranchScope(actor(r), requested)
	if err != nil {
		writeError(w, r, err)
		return
	}

	var ratingPtr *int
	if ratingStr := r.URL.Query().Get("rating"); ratingStr != "" {
		if rVal, err := strconv.Atoi(ratingStr); err == nil && rVal >= 1 && rVal <= 5 {
			ratingPtr = &rVal
		}
	}

	search := r.URL.Query().Get("search")
	limit, offset := paging(r)

	result, err := c.svc.ListOrderFeedbackAdmin(r.Context(), branchID, ratingPtr, search, limit, offset)
	if err != nil {
		writeError(w, r, err)
		return
	}
	ok(w, result)
}

func (c *Controller) FeedbackAnalytics(w http.ResponseWriter, r *http.Request) {
	requested, _ := queryUUID(r, "branch_id")
	branchID, err := service.ResolveBranchScope(actor(r), requested)
	if err != nil {
		writeError(w, r, err)
		return
	}

	analytics, err := c.svc.GetFeedbackAnalytics(r.Context(), branchID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	ok(w, analytics)
}

func (c *Controller) GenerateFeedbackAISummary(w http.ResponseWriter, r *http.Request) {
	requested, _ := queryUUID(r, "branch_id")
	branchID, err := service.ResolveBranchScope(actor(r), requested)
	if err != nil {
		writeError(w, r, err)
		return
	}

	aiSummary, err := c.svc.GenerateFeedbackAISummary(r.Context(), branchID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	ok(w, aiSummary)
}

func paging(r *http.Request) (limit, offset int) {
	limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ = strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func (c *Controller) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	id, err := urlUUID(r, "id")
	if err != nil {
		writeError(w, r, err)
		return
	}

	var req dto.UpdateOrderStatusRequest
	if err := decode(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}

	// Staff may only act on orders belonging to their own branch.
	existing, err := c.svc.GetOrder(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if !c.svc.CanAccessOrder(actor(r), existing) {
		writeError(w, r, apperror.ErrForbidden)
		return
	}

	order, err := c.svc.UpdateOrderStatus(r.Context(), id, req.Status, req.RejectionReason, req.ExpectedVersion)
	if err != nil {
		writeError(w, r, err)
		return
	}
	ok(w, order)
}

// RefundOrder issues a Midtrans refund (full or partial) against a paid
// order and moves it to the terminal "refunded" status.
func (c *Controller) RefundOrder(w http.ResponseWriter, r *http.Request) {
	id, err := urlUUID(r, "id")
	if err != nil {
		writeError(w, r, err)
		return
	}

	var req dto.RefundOrderRequest
	if err := decode(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	if err := req.Validate(); err != nil {
		writeError(w, r, err)
		return
	}

	existing, err := c.svc.GetOrder(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}
	claims := actor(r)
	if !c.svc.CanAccessOrder(claims, existing) {
		writeError(w, r, apperror.ErrForbidden)
		return
	}

	order, err := c.svc.RefundPayment(r.Context(), id, req.Amount, req.Reason, claims.UserID, req.ExpectedVersion)
	if err != nil {
		writeError(w, r, err)
		return
	}
	ok(w, order)
}

// AcknowledgeOrders clears the admin's unread badge. With no ids in the body
// every unread order in scope is marked as seen.
func (c *Controller) AcknowledgeOrders(w http.ResponseWriter, r *http.Request) {
	claims := actor(r)

	var req struct {
		OrderIDs []uuid.UUID `json:"order_ids"`
		BranchID *uuid.UUID  `json:"branch_id"`
	}
	if r.ContentLength > 0 {
		if err := decode(w, r, &req); err != nil {
			writeError(w, r, err)
			return
		}
	}

	// An owner must pass branch_id to scope this to the outlet they are
	// currently viewing — otherwise "mark as read" silently acknowledges every
	// outlet's backlog at once, which the admin portal used to do.
	branchID, err := service.ResolveBranchScope(claims, req.BranchID)
	if err != nil {
		writeError(w, r, err)
		return
	}

	count, err := c.svc.AcknowledgeOrders(r.Context(), branchID, req.OrderIDs, claims.UserID)
	if err != nil {
		writeError(w, r, err)
		return
	}

	unread, err := c.svc.CountUnacknowledgedOrders(r.Context(), branchID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	ok(w, map[string]any{"acknowledged": count, "unread": unread})
}

func (c *Controller) CancelOrder(w http.ResponseWriter, r *http.Request) {
	id, err := urlUUID(r, "id")
	if err != nil {
		writeError(w, r, err)
		return
	}

	// Only the customer who placed the order (or staff) may cancel it.
	existing, err := c.svc.GetOrder(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if !c.svc.CanAccessOrder(actor(r), existing) {
		writeError(w, r, apperror.ErrNotFound)
		return
	}

	order, err := c.svc.CancelOrder(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}
	ok(w, order)
}

// ---------------------------------------------------------------------
// Payments
// ---------------------------------------------------------------------

func (c *Controller) CreatePayment(w http.ResponseWriter, r *http.Request) {
	var req dto.CreatePaymentRequest
	if err := decode(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}

	// The header is authoritative; a body value is accepted for older clients.
	if key := r.Header.Get("Idempotency-Key"); key != "" {
		req.IdempotencyKey = key
	}
	if strings.TrimSpace(req.IdempotencyKey) == "" {
		writeError(w, r, apperror.Invalid("Idempotency-Key wajib dikirim untuk pembayaran"))
		return
	}
	if len(req.IdempotencyKey) > 100 {
		writeError(w, r, apperror.Invalid("Idempotency-Key terlalu panjang"))
		return
	}
	if strings.TrimSpace(req.PaymentMethod) == "" {
		req.PaymentMethod = "snap"
	}

	existing, err := c.svc.GetOrder(r.Context(), req.OrderID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if !c.svc.CanAccessOrder(actor(r), existing) {
		writeError(w, r, apperror.ErrNotFound)
		return
	}

	payment, err := c.svc.CreatePayment(r.Context(), &req)
	if err != nil {
		writeError(w, r, err)
		return
	}
	ok(w, payment)
}

func (c *Controller) GetPaymentStatus(w http.ResponseWriter, r *http.Request) {
	orderID, err := urlUUID(r, "id")
	if err != nil {
		writeError(w, r, err)
		return
	}

	order, err := c.svc.GetOrder(r.Context(), orderID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if !c.svc.CanAccessOrder(actor(r), order) {
		writeError(w, r, apperror.ErrNotFound)
		return
	}

	payment, err := c.svc.GetPaymentForOrder(r.Context(), orderID)
	if err != nil {
		writeError(w, r, err)
		return
	}

	if payment != nil && payment.Status == "pending" {
		if synced, syncErr := c.svc.SyncPaymentWithMidtrans(r.Context(), order.OrderNumber); syncErr == nil && synced != nil {
			payment = synced
			if updatedOrder, oErr := c.svc.GetOrder(r.Context(), orderID); oErr == nil && updatedOrder != nil {
				order = updatedOrder
			}
		}
	}

	ok(w, map[string]any{
		"order_status":   order.Status,
		"payment_status": payment.Status,
		"expires_at":     payment.ExpiresAt,
		"paid_at":        payment.PaidAt,
		"redirect_url":   payment.SnapRedirectURL,
		"snap_token":     payment.SnapToken,
	})
}

func (c *Controller) HandleMidtransNotification(w http.ResponseWriter, r *http.Request) {
	var notification map[string]any
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBody)
	if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if err := c.svc.HandleMidtransWebhook(r.Context(), notification); err != nil {
		// Midtrans retries on a non-2xx. A bad signature is never going to
		// become valid, so acknowledge it and stop the retry storm; genuine
		// server faults still return 500 so the notification is redelivered.
		if errors.Is(err, apperror.ErrUnauthorized) || errors.Is(err, apperror.ErrNotFound) {
			writeJSON(w, http.StatusOK, map[string]string{"status": "ignored"})
			return
		}
		http.Error(w, "processing error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ---------------------------------------------------------------------
// Promos
// ---------------------------------------------------------------------

func (c *Controller) ValidatePromo(w http.ResponseWriter, r *http.Request) {
	var req dto.ValidatePromoRequest
	if err := decode(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}

	resp, err := c.svc.ValidatePromo(r.Context(), &req)
	if err != nil {
		writeError(w, r, err)
		return
	}
	ok(w, resp)
}

// ---------------------------------------------------------------------
// Settings
// ---------------------------------------------------------------------

func (c *Controller) GetBranchSettings(w http.ResponseWriter, r *http.Request) {
	branchID, err := c.resolveSettingsBranch(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	settings, err := c.svc.GetBranchSettings(r.Context(), branchID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	ok(w, settings)
}

func (c *Controller) UpdateBranchSettings(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateBranchSettingsRequest
	if err := decode(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}

	scoped, err := service.ResolveBranchScope(actor(r), req.BranchID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if scoped == nil {
		writeError(w, r, apperror.Invalid("outlet wajib dipilih"))
		return
	}

	settings, err := c.svc.UpdateBranchSettings(r.Context(), *scoped, &req)
	if err != nil {
		writeError(w, r, err)
		return
	}
	ok(w, settings)
}

// resolveSettingsBranch picks the outlet a settings request targets. The owner
// must name one explicitly; there is no default outlet.
func (c *Controller) resolveSettingsBranch(r *http.Request) (uuid.UUID, error) {
	requested, err := queryUUID(r, "branch_id")
	if err != nil {
		return uuid.Nil, err
	}
	scoped, err := service.ResolveBranchScope(actor(r), requested)
	if err != nil {
		return uuid.Nil, err
	}
	if scoped == nil {
		return uuid.Nil, apperror.Invalid("parameter branch_id wajib diisi untuk akun owner")
	}
	return *scoped, nil
}

// ---------------------------------------------------------------------
// Uploads
// ---------------------------------------------------------------------

func (c *Controller) UploadImage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		writeError(w, r, apperror.Invalid("gambar terlalu besar (maksimal 8 MB)"))
		return
	}

	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		writeError(w, r, apperror.Invalid("file gambar tidak ditemukan pada field 'image'"))
		return
	}
	defer file.Close()

	url, err := c.svc.UploadAndCompressImage(fileHeader)
	if err != nil {
		writeError(w, r, err)
		return
	}
	ok(w, map[string]string{"url": url})
}

const maxUploadBytes = 8 << 20

// ---------------------------------------------------------------------
// Analytics
// ---------------------------------------------------------------------

func (c *Controller) BranchDashboard(w http.ResponseWriter, r *http.Request) {
	requested, err := queryUUID(r, "branch_id")
	if err != nil {
		writeError(w, r, err)
		return
	}
	branchID, err := service.ResolveBranchScope(actor(r), requested)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if branchID == nil {
		writeError(w, r, apperror.Invalid("parameter branch_id wajib diisi untuk akun owner"))
		return
	}

	stats, err := c.svc.GetDashboardStats(r.Context(), branchID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	ok(w, stats)
}

// OwnerDashboard aggregates every outlet, or one outlet when branch_id is given.
func (c *Controller) OwnerDashboard(w http.ResponseWriter, r *http.Request) {
	branchID, err := queryUUID(r, "branch_id")
	if err != nil {
		writeError(w, r, err)
		return
	}

	stats, err := c.svc.GetDashboardStats(r.Context(), branchID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	ok(w, stats)
}

// ---------------------------------------------------------------------
// WebSocket
// ---------------------------------------------------------------------

// HandleWebSocket authenticates the socket before upgrading. The token is
// accepted from the query string because the browser WebSocket API cannot set
// an Authorization header.
func (c *Controller) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	var identity *ws.Identity

	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
			token = strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
		} else if cookie, err := r.Cookie("token"); err == nil {
			token = cookie.Value
		}
	}

	if token != "" {
		claims, err := middleware.ParseToken(c.cfg, token)
		if err != nil {
			writeError(w, r, apperror.ErrUnauthorized)
			return
		}
		identity = &ws.Identity{UserID: claims.UserID, Role: claims.Role, BranchID: claims.BranchID}
	}

	origins := []string{c.cfg.CustomerURL, c.cfg.AdminURL}
	if !c.cfg.IsProduction() {
		origins = append(origins,
			"http://localhost:5173", "http://localhost:5174",
			"http://127.0.0.1:5173", "http://127.0.0.1:5174",
		)
	}
	ws.ServeWs(c.hub, w, r, identity, origins)
}

// Health is the liveness probe for nginx and the process manager.
func (c *Controller) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ---------------------------------------------------------------------
// Credentials & Password Management
// ---------------------------------------------------------------------

func (c *Controller) ChangePassword(w http.ResponseWriter, r *http.Request) {
	claims := actor(r)
	if claims == nil {
		writeError(w, r, apperror.ErrUnauthorized)
		return
	}

	var req dto.ChangePasswordRequest
	if err := decode(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}

	if err := c.svc.ChangePassword(r.Context(), claims.UserID, req.CurrentPassword, req.NewPassword); err != nil {
		writeError(w, r, err)
		return
	}

	ok(w, map[string]any{
		"success": true,
		"message": "Password berhasil diubah",
	})
}

func (c *Controller) ListBranchCredentials(w http.ResponseWriter, r *http.Request) {
	claims := actor(r)
	list, err := c.svc.ListBranchCredentials(r.Context(), claims)
	if err != nil {
		writeError(w, r, err)
		return
	}
	ok(w, list)
}

func (c *Controller) GetBranchCredentials(w http.ResponseWriter, r *http.Request) {
	claims := actor(r)
	branchID, err := urlUUID(r, "id")
	if err != nil {
		writeError(w, r, err)
		return
	}

	creds, err := c.svc.GetBranchCredentials(r.Context(), claims, branchID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	ok(w, creds)
}

func (c *Controller) UpdateBranchCredentials(w http.ResponseWriter, r *http.Request) {
	claims := actor(r)
	branchID, err := urlUUID(r, "id")
	if err != nil {
		writeError(w, r, err)
		return
	}

	var req dto.UpdateBranchCredentialsRequest
	if err := decode(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}

	creds, err := c.svc.UpdateBranchCredentials(r.Context(), claims, branchID, &req)
	if err != nil {
		writeError(w, r, err)
		return
	}

	ok(w, creds)
}

