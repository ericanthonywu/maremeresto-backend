package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ericanthonywu/maremereso-olga/backend/internal/apperror"
	"github.com/ericanthonywu/maremereso-olga/backend/internal/dto"
	"github.com/ericanthonywu/maremereso-olga/backend/internal/middleware"
	"github.com/ericanthonywu/maremereso-olga/backend/internal/service"
	ws "github.com/ericanthonywu/maremereso-olga/backend/internal/websocket"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Controller struct {
	svc *service.Service
	hub *ws.Hub
}

func NewController(svc *service.Service, hub *ws.Hub) *Controller {
	return &Controller{
		svc: svc,
		hub: hub,
	}
}

// Helpers
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, err error) {
	appErr := apperror.FromError(err)
	writeJSON(w, appErr.Code, map[string]any{
		"success": false,
		"error":   appErr.Message,
	})
}

// ---------------------------------------------------------------------
// Auth Handlers
// ---------------------------------------------------------------------
func (c *Controller) CustomerPhoneLogin(w http.ResponseWriter, r *http.Request) {
	var req dto.CustomerPhoneLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperror.ErrBadRequest)
		return
	}

	resp, err := c.svc.CustomerLogin(r.Context(), req.Phone, req.Name)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    resp,
	})
}

func (c *Controller) AdminLogin(w http.ResponseWriter, r *http.Request) {
	var req dto.AdminLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperror.ErrBadRequest)
		return
	}

	resp, err := c.svc.AdminLogin(r.Context(), req.Identifier, req.Password)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    resp,
	})
}

func (c *Controller) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, apperror.ErrUnauthorized)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    claims,
	})
}

// ---------------------------------------------------------------------
// Branch Handlers
// ---------------------------------------------------------------------
func (c *Controller) ListBranches(w http.ResponseWriter, r *http.Request) {
	branches, err := c.svc.ListBranches(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    branches,
	})
}

func (c *Controller) GetBranch(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	branch, err := c.svc.GetBranchBySlug(r.Context(), slug)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    branch,
	})
}

func (c *Controller) UpdateBranchStatus(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetUserFromContext(r.Context())
	branchIDStr := chi.URLParam(r, "id")
	branchID, err := uuid.Parse(branchIDStr)
	if err != nil {
		writeError(w, apperror.ErrBadRequest)
		return
	}

	// Branch isolation check
	if claims.Role == "branch_admin" && (claims.BranchID == nil || *claims.BranchID != branchID) {
		writeError(w, apperror.ErrForbidden)
		return
	}

	var req struct {
		IsOpen bool `json:"is_open"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperror.ErrBadRequest)
		return
	}

	if err := c.svc.UpdateBranchStatus(r.Context(), branchID, req.IsOpen); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Branch status updated successfully",
	})
}

// ---------------------------------------------------------------------
// Menu & Category Handlers
// ---------------------------------------------------------------------
func (c *Controller) ListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := c.svc.ListCategories(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    categories,
	})
}

func (c *Controller) ListMenuByBranch(w http.ResponseWriter, r *http.Request) {
	slugOrID := chi.URLParam(r, "branch")
	var branchID uuid.UUID

	if parsedID, err := uuid.Parse(slugOrID); err == nil {
		branchID = parsedID
	} else {
		b, err := c.svc.GetBranchBySlug(r.Context(), slugOrID)
		if err != nil {
			writeError(w, err)
			return
		}
		branchID = b.ID
	}

	items, err := c.svc.ListMenuByBranch(r.Context(), branchID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    items,
	})
}

func (c *Controller) CreateMenuItem(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateMenuItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperror.ErrBadRequest)
		return
	}

	claims, _ := middleware.GetUserFromContext(r.Context())
	if claims.Role == "branch_admin" && claims.BranchID != nil {
		req.BranchID = *claims.BranchID
	}

	item, err := c.svc.CreateMenuItem(r.Context(), &req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"success": true,
		"data":    item,
	})
}

func (c *Controller) UpdateMenuItem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, apperror.ErrBadRequest)
		return
	}

	var req dto.UpdateMenuItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperror.ErrBadRequest)
		return
	}

	item, err := c.svc.UpdateMenuItem(r.Context(), id, &req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    item,
	})
}

func (c *Controller) ToggleMenuItemAvailability(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, apperror.ErrBadRequest)
		return
	}

	var req dto.ToggleAvailabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperror.ErrBadRequest)
		return
	}

	if err := c.svc.ToggleMenuAvailability(r.Context(), id, req.IsAvailable); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Availability updated",
	})
}

func (c *Controller) DeleteMenuItem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, apperror.ErrBadRequest)
		return
	}

	if err := c.svc.DeleteMenuItem(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Item deleted",
	})
}

// ---------------------------------------------------------------------
// Order Handlers
// ---------------------------------------------------------------------
func (c *Controller) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperror.ErrBadRequest)
		return
	}

	var userID *uuid.UUID
	if claims, ok := middleware.GetUserFromContext(r.Context()); ok {
		userID = &claims.UserID
	}

	order, err := c.svc.CreateOrder(r.Context(), userID, &req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"success": true,
		"data":    order,
	})
}

func (c *Controller) GetOrder(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, apperror.ErrBadRequest)
		return
	}

	order, err := c.svc.GetOrder(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    order,
	})
}

func (c *Controller) ListOrders(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	search := r.URL.Query().Get("search")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	var branchID *uuid.UUID
	if bStr := r.URL.Query().Get("branch_id"); bStr != "" {
		if bid, err := uuid.Parse(bStr); err == nil {
			branchID = &bid
		}
	}

	// Branch admin scope enforcement
	claims, ok := middleware.GetUserFromContext(r.Context())
	if ok && claims.Role == "branch_admin" && claims.BranchID != nil {
		branchID = claims.BranchID
	}

	orders, total, err := c.svc.ListOrders(r.Context(), branchID, status, search, limit, offset)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    orders,
		"total":   total,
	})
}

func (c *Controller) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, apperror.ErrBadRequest)
		return
	}

	var req dto.UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperror.ErrBadRequest)
		return
	}

	if err := c.svc.UpdateOrderStatus(r.Context(), id, req.Status, req.RejectionReason, req.ExpectedVersion); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Order status updated",
	})
}

func (c *Controller) CancelOrder(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, apperror.ErrBadRequest)
		return
	}

	if err := c.svc.CancelOrder(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Order cancelled successfully",
	})
}

// ---------------------------------------------------------------------
// Payment Handlers (Midtrans Integration + Webhook)
// ---------------------------------------------------------------------
func (c *Controller) CreatePayment(w http.ResponseWriter, r *http.Request) {
	var req dto.CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperror.ErrBadRequest)
		return
	}

	// If idempotency key not provided in body, check header
	if req.IdempotencyKey == "" {
		req.IdempotencyKey = r.Header.Get("Idempotency-Key")
	}
	if req.IdempotencyKey == "" {
		req.IdempotencyKey = uuid.NewString()
	}

	payment, err := c.svc.CreatePayment(r.Context(), &req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    payment,
	})
}

func (c *Controller) HandleMidtransNotification(w http.ResponseWriter, r *http.Request) {
	var notification map[string]any
	if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if err := c.svc.HandleMidtransWebhook(r.Context(), notification); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// ---------------------------------------------------------------------
// Promo Handlers
// ---------------------------------------------------------------------
func (c *Controller) ValidatePromo(w http.ResponseWriter, r *http.Request) {
	var req dto.ValidatePromoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperror.ErrBadRequest)
		return
	}

	resp, err := c.svc.ValidatePromo(r.Context(), &req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    resp,
	})
}

// ---------------------------------------------------------------------
// Settings Handlers
// ---------------------------------------------------------------------
func (c *Controller) GetBranchSettings(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetUserFromContext(r.Context())
	var branchID uuid.UUID

	if bStr := r.URL.Query().Get("branch_id"); bStr != "" {
		branchID, _ = uuid.Parse(bStr)
	} else if claims != nil && claims.BranchID != nil {
		branchID = *claims.BranchID
	} else {
		// Default to Sudirman
		branchID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	}

	settings, err := c.svc.GetBranchSettings(r.Context(), branchID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    settings,
	})
}

func (c *Controller) UpdateBranchSettings(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetUserFromContext(r.Context())
	var branchID uuid.UUID

	if claims != nil && claims.BranchID != nil {
		branchID = *claims.BranchID
	} else {
		branchID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	}

	var req dto.UpdateBranchSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperror.ErrBadRequest)
		return
	}

	if err := c.svc.UpdateBranchSettings(r.Context(), branchID, &req); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Settings updated",
	})
}

// ---------------------------------------------------------------------
// Upload Handlers (Auto-Compression)
// ---------------------------------------------------------------------
func (c *Controller) UploadImage(w http.ResponseWriter, r *http.Request) {
	// Parse 10MB max form
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, apperror.ErrBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		writeError(w, apperror.ErrBadRequest)
		return
	}
	defer file.Close()

	url, err := c.svc.UploadAndCompressImage(r.Context(), fileHeader)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"url":     url,
	})
}

// ---------------------------------------------------------------------
// Analytics Handlers
// ---------------------------------------------------------------------
func (c *Controller) BranchDashboard(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetUserFromContext(r.Context())
	var branchID uuid.UUID
	if claims != nil && claims.BranchID != nil {
		branchID = *claims.BranchID
	} else {
		branchID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	}

	stats, err := c.svc.GetBranchAnalytics(r.Context(), branchID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    stats,
	})
}

func (c *Controller) OwnerDashboard(w http.ResponseWriter, r *http.Request) {
	stats, err := c.svc.GetOwnerAnalytics(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    stats,
	})
}

// ---------------------------------------------------------------------
// WebSocket Handler
// ---------------------------------------------------------------------
func (c *Controller) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	ws.ServeWs(c.hub, w, r)
}
