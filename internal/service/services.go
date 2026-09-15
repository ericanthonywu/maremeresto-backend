package service

import (
	"bytes"
	"context"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"log/slog"
	"math"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/ericanthonywu/maremereso-olga/backend/internal/apperror"
	"github.com/ericanthonywu/maremereso-olga/backend/internal/config"
	"github.com/ericanthonywu/maremereso-olga/backend/internal/dto"
	"github.com/ericanthonywu/maremereso-olga/backend/internal/middleware"
	"github.com/ericanthonywu/maremereso-olga/backend/internal/model"
	"github.com/ericanthonywu/maremereso-olga/backend/internal/repository"
	ws "github.com/ericanthonywu/maremereso-olga/backend/internal/websocket"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo       *repository.Repository
	cfg        *config.Config
	hub        *ws.Hub
	snapClient snap.Client
	geocoder   *geocoder
}

func NewService(repo *repository.Repository, cfg *config.Config, hub *ws.Hub) *Service {
	s := &Service{
		repo:     repo,
		cfg:      cfg,
		hub:      hub,
		geocoder: newGeocoder(cfg.GeocoderURL, cfg.GeocoderEmail),
	}

	env := midtrans.Sandbox
	if cfg.MidtransIsProd {
		env = midtrans.Production
	}
	s.snapClient.New(cfg.MidtransServerKey, env)

	hub.SetAuthorizer(s.authorizeRoom)

	return s
}

// authorizeRoom is the websocket access policy. Order payloads carry the
// customer's name, phone and address, so every subscription is checked here.
func (s *Service) authorizeRoom(id *ws.Identity, room string) bool {
	switch {
	case room == ws.RoomOwner:
		return id != nil && id.Role == "owner"

	case strings.HasPrefix(room, ws.RoomBranchPfx):
		if id == nil {
			return false
		}
		if id.Role == "owner" {
			return true
		}
		branchID, err := uuid.Parse(strings.TrimPrefix(room, ws.RoomBranchPfx))
		if err != nil {
			return false
		}
		return id.Role == "branch_admin" && id.BranchID != nil && *id.BranchID == branchID

	case strings.HasPrefix(room, ws.RoomOrderPfx):
		orderID, err := uuid.Parse(strings.TrimPrefix(room, ws.RoomOrderPfx))
		if err != nil || id == nil {
			return false
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		order, err := s.repo.FindOrderByID(ctx, orderID)
		if err != nil || order == nil {
			return false
		}
		return s.canAccessOrder(id.Role, id.UserID, id.BranchID, order)

	default:
		return false
	}
}

// canAccessOrder centralises the ownership rule: the customer who placed the
// order, staff of the branch fulfilling it, or the owner.
func (s *Service) canAccessOrder(role string, userID uuid.UUID, branchID *uuid.UUID, order *model.Order) bool {
	switch role {
	case "owner":
		return true
	case "branch_admin":
		return branchID != nil && *branchID == order.BranchID
	default:
		return order.UserID != nil && *order.UserID == userID
	}
}

// CanAccessOrder is the exported form used by HTTP handlers.
func (s *Service) CanAccessOrder(claims *middleware.JWTClaims, order *model.Order) bool {
	if claims == nil || order == nil {
		return false
	}
	return s.canAccessOrder(claims.Role, claims.UserID, claims.BranchID, order)
}

// ---------------------------------------------------------------------
// Auth Service
// ---------------------------------------------------------------------
func (s *Service) CustomerLogin(ctx context.Context, rawPhone, name string) (*dto.AuthResponse, error) {
	name = strings.TrimSpace(name)
	if len(name) < 2 || len(name) > 100 {
		return nil, apperror.Invalid("nama pemesan harus 2-100 karakter")
	}
	// Strict phone validation and canonicalization to +628...
	normalizedPhone, err := dto.NormalizeIndonesianPhone(rawPhone)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.FindUserByPhone(ctx, normalizedPhone)
	if err != nil {
		return nil, err
	}

	if user == nil {
		user, err = s.repo.CreateCustomer(ctx, normalizedPhone, name)
		if err != nil {
			return nil, err
		}
	}

	// Staff accounts must authenticate with a password through the admin
	// portal; the passwordless phone flow is for customers only.
	if user.Role != "customer" {
		return nil, apperror.ErrUnauthorized
	}
	// Checkout submits the current required name. Keep an existing account in
	// sync without making the customer visit a separate profile screen first.
	if user.Name != name {
		updated, updateErr := s.repo.UpdateCustomerProfile(ctx, user.ID, name, normalizedPhone)
		if updateErr != nil {
			return nil, updateErr
		}
		if updated != nil {
			user = updated
		}
	}

	tokenStr, err := s.issueToken(user, 30*24*time.Hour)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: tokenStr,
		User: dto.UserSummary{
			ID:       user.ID,
			Name:     user.Name,
			Phone:    user.Phone,
			Role:     user.Role,
			BranchID: user.BranchID,
		},
	}, nil
}

func (s *Service) UpdateCustomerProfile(ctx context.Context, userID uuid.UUID, req *dto.UpdateCustomerProfileRequest) (*dto.AuthResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	user, err := s.repo.UpdateCustomerProfile(ctx, userID, req.Name, req.Phone)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperror.ErrNotFound
	}
	token, err := s.issueToken(user, 30*24*time.Hour)
	if err != nil {
		return nil, err
	}
	return &dto.AuthResponse{Token: token, User: dto.UserSummary{
		ID: user.ID, Name: user.Name, Phone: user.Phone, Role: user.Role, BranchID: user.BranchID,
	}}, nil
}

func (s *Service) AdminLogin(ctx context.Context, identifier, password string) (*dto.AuthResponse, error) {
	identifier = strings.TrimSpace(identifier)
	if identifier == "" || password == "" {
		return nil, apperror.ErrUnauthorized
	}

	var user *model.User
	var err error

	if strings.Contains(identifier, "@") {
		user, err = s.repo.FindUserByEmail(ctx, strings.ToLower(identifier))
	} else {
		normalized, errNorm := dto.NormalizeIndonesianPhone(identifier)
		if errNorm != nil {
			return nil, apperror.ErrUnauthorized
		}
		user, err = s.repo.FindUserByPhone(ctx, normalized)
	}
	if err != nil {
		return nil, err
	}

	// Compare against a dummy hash when the account is missing or has no
	// password so the response time does not reveal which accounts exist.
	storedHash := dummyBcryptHash
	if user != nil && user.PasswordHash != nil {
		storedHash = *user.PasswordHash
	}
	bcryptErr := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))

	if user == nil || user.PasswordHash == nil || bcryptErr != nil {
		slog.Warn("failed admin login attempt", "identifier", identifier)
		return nil, apperror.ErrUnauthorized
	}
	if user.Role != "branch_admin" && user.Role != "owner" {
		return nil, apperror.ErrUnauthorized
	}
	if user.Role == "branch_admin" && user.BranchID == nil {
		slog.Error("branch admin has no branch assigned", "user_id", user.ID)
		return nil, apperror.ErrUnauthorized
	}

	tokenStr, err := s.issueToken(user, 12*time.Hour)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: tokenStr,
		User: dto.UserSummary{
			ID:       user.ID,
			Name:     user.Name,
			Phone:    user.Phone,
			Role:     user.Role,
			BranchID: user.BranchID,
		},
	}, nil
}

// dummyBcryptHash is a valid bcrypt digest of a random value. Comparing
// against it keeps the failure path the same cost as the success path.
const dummyBcryptHash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

func (s *Service) issueToken(user *model.User, ttl time.Duration) (string, error) {
	claims := middleware.JWTClaims{
		UserID:   user.ID,
		Phone:    user.Phone,
		Role:     user.Role,
		BranchID: user.BranchID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.cfg.JWTSecret))
}

// ---------------------------------------------------------------------
// Branch & Menu Service
// ---------------------------------------------------------------------
func (s *Service) ListBranches(ctx context.Context) ([]model.Branch, error) {
	branches, err := s.repo.ListBranches(ctx)
	if err != nil {
		return nil, err
	}

	// One query for every branch's settings rather than one per branch.
	settings, err := s.repo.ListSettingsByBranch(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	for i := range branches {
		set, ok := settings[branches[i].ID]
		if !ok {
			branches[i].IsOpenNow = branches[i].IsOpen
			continue
		}
		branches[i].IsOpenNow = branches[i].IsOpen && IsWithinOperatingHours(set.OperatingHours, now)
		branches[i].TodayHours = FormatOperatingHours(set.OperatingHours, now)
		branches[i].WhatsappNumber = set.WhatsappNumber
		branches[i].HalalCertificateID = set.HalalCertificateID
	}
	return branches, nil
}

func (s *Service) GetBranchBySlug(ctx context.Context, slug string) (*model.Branch, error) {
	b, err := s.repo.FindBranchBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, apperror.ErrNotFound
	}

	set := s.settingsFor(ctx, b.ID)
	now := time.Now()
	b.IsOpenNow = b.IsOpen && IsWithinOperatingHours(set.OperatingHours, now)
	b.TodayHours = FormatOperatingHours(set.OperatingHours, now)
	b.WhatsappNumber = set.WhatsappNumber
	b.HalalCertificateID = set.HalalCertificateID
	return b, nil
}

func (s *Service) ListCategories(ctx context.Context) ([]model.Category, error) {
	return s.repo.ListCategories(ctx)
}

func slugify(val string) string {
	val = strings.ToLower(strings.TrimSpace(val))
	var b strings.Builder
	lastDash := false
	for _, r := range val {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDash = false
		} else if !lastDash && b.Len() > 0 {
			b.WriteRune('-')
			lastDash = true
		}
	}
	res := strings.Trim(b.String(), "-")
	if res == "" {
		res = "kategori"
	}
	if len(res) > 40 {
		res = res[:40]
	}
	return res
}

func (s *Service) CreateCategory(ctx context.Context, req *dto.CreateCategoryRequest) (*model.Category, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	baseSlug := slugify(req.Name)
	slug := baseSlug
	suffix := 1
	for {
		exists, err := s.repo.CategorySlugExists(ctx, slug, nil)
		if err != nil {
			return nil, err
		}
		if !exists {
			break
		}
		slug = fmt.Sprintf("%s-%d", baseSlug, suffix)
		suffix++
	}

	cat := &model.Category{
		ID:        uuid.New(),
		Name:      req.Name,
		Slug:      slug,
		Emoji:     req.Emoji,
		SortOrder: req.SortOrder,
		CreatedAt: time.Now(),
	}
	if err := s.repo.CreateCategory(ctx, cat); err != nil {
		return nil, err
	}
	return cat, nil
}

func (s *Service) UpdateCategory(ctx context.Context, id uuid.UUID, req *dto.UpdateCategoryRequest) (*model.Category, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	existing, err := s.repo.FindCategoryByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, apperror.ErrNotFound
	}

	existing.Name = req.Name
	existing.Emoji = req.Emoji
	existing.SortOrder = req.SortOrder
	if err := s.repo.UpdateCategory(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *Service) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	existing, err := s.repo.FindCategoryByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.ErrNotFound
	}
	count, err := s.repo.CountMenuItemsByCategoryID(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return apperror.Invalid(fmt.Sprintf("kategori tidak dapat dihapus karena masih digunakan oleh %d item menu", count))
	}
	return s.repo.DeleteCategory(ctx, id)
}

func (s *Service) ListMenuByBranch(ctx context.Context, branchID uuid.UUID) ([]model.MenuItem, error) {
	return s.repo.ListMenuItemsByBranch(ctx, branchID)
}

// CreateMenuItem adds an item. A branch admin may only write to their own
// outlet; the owner must name the branch explicitly.
func (s *Service) CreateMenuItem(ctx context.Context, actor *middleware.JWTClaims, req *dto.CreateMenuItemRequest) (*model.MenuItem, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	branchID, err := resolveWritableBranch(actor, req.BranchID)
	if err != nil {
		return nil, err
	}
	if err := s.assertBranchAndCategoryExist(ctx, branchID, req.CategoryID); err != nil {
		return nil, err
	}

	item := &model.MenuItem{
		BranchID:    branchID,
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Icon:        defaultIfEmpty(req.Icon, "fa-mug-hot"),
		IconBgClass: defaultIfEmpty(req.IconBgClass, "bg-amber-50"),
		IsAvailable: req.IsAvailable,
		SortOrder:   req.SortOrder,
	}
	if req.Tag != "" {
		tag := req.Tag
		item.Tag = &tag
	}
	if req.ImageURL != "" {
		url := req.ImageURL
		item.ImageURL = &url
	}

	if err := s.repo.CreateMenuItem(ctx, item); err != nil {
		return nil, err
	}
	return s.repo.FindMenuItemByID(ctx, item.ID)
}

// CreateMenuItemsBulk validates every request row before creating the menu
// items. The editor uses this for its spreadsheet-style bulk entry action.
func (s *Service) CreateMenuItemsBulk(ctx context.Context, actor *middleware.JWTClaims, req *dto.CreateMenuItemsBulkRequest) ([]model.MenuItem, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	items := make([]model.MenuItem, 0, len(req.Items))
	for i := range req.Items {
		item, err := s.CreateMenuItem(ctx, actor, &req.Items[i])
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, nil
}

func (s *Service) UpdateMenuItem(ctx context.Context, actor *middleware.JWTClaims, id uuid.UUID, req *dto.UpdateMenuItemRequest) (*model.MenuItem, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	item, err := s.loadWritableMenuItem(ctx, actor, id)
	if err != nil {
		return nil, err
	}

	if req.CategoryID != nil {
		exists, err := s.repo.CategoryExists(ctx, *req.CategoryID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, apperror.Invalid("kategori tidak ditemukan")
		}
		item.CategoryID = *req.CategoryID
	}
	if req.Name != nil {
		item.Name = *req.Name
	}
	if req.Description != nil {
		item.Description = *req.Description
	}
	if req.Price != nil {
		item.Price = *req.Price
	}
	if req.Icon != nil {
		item.Icon = defaultIfEmpty(*req.Icon, "fa-mug-hot")
	}
	if req.IconBgClass != nil {
		item.IconBgClass = defaultIfEmpty(*req.IconBgClass, "bg-amber-50")
	}
	if req.Tag != nil {
		if *req.Tag == "" {
			item.Tag = nil
		} else {
			item.Tag = req.Tag
		}
	}
	if req.ImageURL != nil {
		if *req.ImageURL == "" {
			item.ImageURL = nil
		} else {
			item.ImageURL = req.ImageURL
		}
	}
	if req.IsAvailable != nil {
		item.IsAvailable = *req.IsAvailable
	}
	if req.SortOrder != nil {
		item.SortOrder = *req.SortOrder
	}

	if err := s.repo.UpdateMenuItem(ctx, item); err != nil {
		return nil, err
	}
	return s.repo.FindMenuItemByID(ctx, item.ID)
}

func (s *Service) ToggleMenuAvailability(ctx context.Context, actor *middleware.JWTClaims, itemID uuid.UUID, isAvailable bool) error {
	if _, err := s.loadWritableMenuItem(ctx, actor, itemID); err != nil {
		return err
	}
	return s.repo.ToggleMenuItemAvailability(ctx, itemID, isAvailable)
}

func (s *Service) DeleteMenuItem(ctx context.Context, actor *middleware.JWTClaims, id uuid.UUID) error {
	if _, err := s.loadWritableMenuItem(ctx, actor, id); err != nil {
		return err
	}
	return s.repo.DeleteMenuItem(ctx, id)
}

// loadWritableMenuItem fetches an item and verifies the caller may modify it.
// Without this a branch admin could edit or delete another outlet's menu just
// by knowing an item id.
func (s *Service) loadWritableMenuItem(ctx context.Context, actor *middleware.JWTClaims, id uuid.UUID) (*model.MenuItem, error) {
	item, err := s.repo.FindMenuItemByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, apperror.ErrNotFound
	}
	if _, err := resolveWritableBranch(actor, item.BranchID); err != nil {
		return nil, err
	}
	return item, nil
}

// resolveWritableBranch returns the branch the caller is allowed to write to.
func resolveWritableBranch(actor *middleware.JWTClaims, requested uuid.UUID) (uuid.UUID, error) {
	if actor == nil {
		return uuid.Nil, apperror.ErrUnauthorized
	}

	switch actor.Role {
	case "owner":
		if requested == uuid.Nil {
			return uuid.Nil, apperror.Invalid("outlet wajib dipilih")
		}
		return requested, nil

	case "branch_admin":
		if actor.BranchID == nil {
			return uuid.Nil, apperror.ErrForbidden
		}
		// A branch admin's own outlet always wins over whatever the client sent.
		if requested != uuid.Nil && requested != *actor.BranchID {
			return uuid.Nil, apperror.ErrForbidden
		}
		return *actor.BranchID, nil

	default:
		return uuid.Nil, apperror.ErrForbidden
	}
}

func (s *Service) assertBranchAndCategoryExist(ctx context.Context, branchID, categoryID uuid.UUID) error {
	branch, err := s.repo.FindBranchByID(ctx, branchID)
	if err != nil {
		return err
	}
	if branch == nil {
		return apperror.Invalid("outlet tidak ditemukan")
	}

	exists, err := s.repo.CategoryExists(ctx, categoryID)
	if err != nil {
		return err
	}
	if !exists {
		return apperror.Invalid("kategori tidak ditemukan")
	}
	return nil
}

func defaultIfEmpty(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}

// ---------------------------------------------------------------------
// Order Service (Concurrency & State Machine)
// ---------------------------------------------------------------------
// validTransitions is intentionally compact after payment: an outlet either
// marks the paid order delivered, rejects it, or refunds it through the
// dedicated refund flow. Historical statuses are normalised by migration 8.
var validTransitions = map[string][]string{
	"pending":  {"accepted", "rejected", "cancelled"},
	"accepted": {"completed", "rejected", "cancelled"},
}

// statusesCustomerMayCancel: once a barista has started making the drinks the
// ingredients are already spent, so self-service cancellation stops there.
var statusesCustomerMayCancel = map[string]bool{"pending": true}

// customerCancelWindow is the "5 minute cancellation guarantee" advertised in
// the customer app.
const customerCancelWindow = 5 * time.Minute

func (s *Service) CreateOrder(ctx context.Context, userID *uuid.UUID, req *dto.CreateOrderRequest) (*model.Order, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 1. Branch must exist, be switched on, and be inside its trading hours.
	branch, err := s.repo.FindBranchByID(ctx, req.BranchID)
	if err != nil {
		return nil, err
	}
	if branch == nil {
		return nil, apperror.ErrNotFound
	}
	settings := s.settingsFor(ctx, branch.ID)
	if !branch.IsOpen || !IsWithinOperatingHours(settings.OperatingHours, time.Now()) {
		return nil, apperror.ErrStoreClosed
	}

	normPhone, err := dto.NormalizeIndonesianPhone(req.CustomerPhone)
	if err != nil {
		return nil, err
	}

	// 2. Price the basket from the database, never from the client. Each item
	//    must belong to this branch, so a menu id copied from another outlet
	//    cannot be ordered at its neighbour's price.
	ids := make([]uuid.UUID, 0, len(req.Items))
	for _, it := range req.Items {
		ids = append(ids, it.MenuItemID)
	}
	menuItems, err := s.repo.FindMenuItemsByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	subtotal := 0
	orderItems := make([]model.OrderItem, 0, len(req.Items))
	for _, itReq := range req.Items {
		mItem, ok := menuItems[itReq.MenuItemID]
		if !ok || !mItem.IsAvailable {
			return nil, apperror.Invalid("menu yang dipilih sedang tidak tersedia, silakan muat ulang keranjang")
		}
		if mItem.BranchID != branch.ID {
			return nil, apperror.Invalid(fmt.Sprintf("%q tidak tersedia di outlet %s", mItem.Name, branch.Name))
		}

		lineTotal := mItem.Price * itReq.Quantity
		subtotal += lineTotal

		notes := strings.TrimSpace(itReq.Notes)
		menuItemID := mItem.ID
		orderItems = append(orderItems, model.OrderItem{
			MenuItemID: &menuItemID,
			ItemName:   mItem.Name,
			ItemPrice:  mItem.Price,
			ItemIcon:   mItem.Icon,
			Quantity:   itReq.Quantity,
			Notes:      &notes,
			LineTotal:  lineTotal,
		})
	}

	if settings.MinOrderAmount > 0 && subtotal < settings.MinOrderAmount {
		return nil, fmt.Errorf("%w (minimum Rp %s)", apperror.ErrBelowMinimumOrder, formatRupiah(settings.MinOrderAmount))
	}

	// 3. Distance and delivery fee come from the same helpers the quote
	//    endpoint uses, so the customer is charged exactly what they were shown.
	isDelivery := req.OrderType == "delivery" || req.OrderType == "scheduled"
	distanceKm := 0.0

	if isDelivery {
		if req.DeliveryLat == nil || req.DeliveryLon == nil || (*req.DeliveryLat == 0 && *req.DeliveryLon == 0) {
			return nil, apperror.ErrLocationRequired
		}
		if strings.TrimSpace(req.DeliveryAddress) == "" {
			return nil, apperror.Invalid("alamat pengantaran wajib diisi")
		}
		distanceKm = RoadDistanceKm(*req.DeliveryLat, *req.DeliveryLon, branch.Latitude, branch.Longitude)
		if distanceKm > maxServiceableKm {
			return nil, fmt.Errorf("%w (jarak %.1f km, maksimal %d km)",
				apperror.ErrOutOfDeliveryRange, distanceKm, int(maxServiceableKm))
		}
	}

	deliveryFee := DeliveryFee(settings, distanceKm, subtotal, req.OrderType)

	// 4. Promo validity, expiry and redemption cap are all enforced here; the
	//    client-side discount is never trusted.
	discount := 0
	promoCode := strings.ToUpper(strings.TrimSpace(req.PromoCode))
	if promoCode != "" {
		promo, err := s.repo.FindPromoByCode(ctx, promoCode)
		if err != nil {
			return nil, err
		}
		if promo == nil || !promo.Redeemable(time.Now()) || subtotal < promo.MinSpend {
			return nil, apperror.ErrInvalidPromo
		}
		switch promo.Type {
		case "fixed":
			discount = promo.DiscountAmount
		case "free_delivery":
			discount = deliveryFee
		}
		if discount > subtotal+deliveryFee {
			discount = subtotal + deliveryFee
		}
	}

	grandTotal := subtotal + deliveryFee + settings.ServiceFee - discount
	if grandTotal < 0 {
		grandTotal = 0
	}

	// 5. Persist atomically.
	tx, err := s.repo.DB().Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	orderNumber, err := s.repo.NextOrderNumber(ctx, tx, branch.Slug)
	if err != nil {
		return nil, err
	}

	deliveryAddress := strings.TrimSpace(req.DeliveryAddress)
	deliveryNotes := strings.TrimSpace(req.DeliveryNotes)

	order := &model.Order{
		OrderNumber:        orderNumber,
		UserID:             userID,
		BranchID:           branch.ID,
		OrderType:          req.OrderType,
		Status:             "pending",
		CustomerName:       strings.TrimSpace(req.CustomerName),
		CustomerPhone:      normPhone,
		DeliveryAddress:    &deliveryAddress,
		DeliveryNotes:      &deliveryNotes,
		DeliveryLat:        req.DeliveryLat,
		DeliveryLon:        req.DeliveryLon,
		DeliveryDistanceKm: distanceKm,
		Subtotal:           subtotal,
		DeliveryFee:        deliveryFee,
		ServiceFee:         settings.ServiceFee,
		Discount:           discount,
		GrandTotal:         grandTotal,
		ScheduledAt:        req.ScheduledAt,
		Items:              orderItems,
	}
	if promoCode != "" {
		order.PromoCode = &promoCode
	}

	if err := s.repo.CreateOrder(ctx, tx, order); err != nil {
		return nil, err
	}
	if promoCode != "" {
		if err := s.repo.IncrementPromoRedemption(ctx, tx, promoCode); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	order.Branch = branch

	// Do not notify admin rooms yet — the order enters the admin system with
	// notifications only after the customer completes payment.
	return order, nil
}

// formatRupiah renders an amount with Indonesian thousands separators for use
// inside error messages shown to the customer.
func formatRupiah(v int) string {
	str := strconv.Itoa(v)
	neg := ""
	if strings.HasPrefix(str, "-") {
		neg, str = "-", str[1:]
	}

	var out []byte
	for i, c := range []byte(str) {
		if i > 0 && (len(str)-i)%3 == 0 {
			out = append(out, '.')
		}
		out = append(out, c)
	}
	return neg + string(out)
}

func (s *Service) GetOrder(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	order, err := s.repo.FindOrderByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, apperror.ErrNotFound
	}
	return order, nil
}

func (s *Service) ListOrders(ctx context.Context, branchID *uuid.UUID, status, search string, limit, offset int) ([]model.Order, int, error) {
	return s.repo.ListOrders(ctx, branchID, status, search, limit, offset)
}

func (s *Service) CountOrdersByStatus(ctx context.Context, branchID *uuid.UUID) (map[string]int, error) {
	return s.repo.CountOrdersByStatus(ctx, branchID)
}

// ListOrdersForCustomer returns the signed-in customer's own order history.
func (s *Service) ListOrdersForCustomer(ctx context.Context, userID uuid.UUID, limit, offset int) ([]model.Order, int, error) {
	return s.repo.ListOrdersForCustomer(ctx, userID, limit, offset)
}

func (s *Service) UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, newStatus string, reason string, expectedVersion int) (*model.Order, error) {
	order, err := s.repo.FindOrderByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, apperror.ErrNotFound
	}

	if !isValidTransition(order.Status, newStatus) {
		return nil, fmt.Errorf("%w: %s -> %s", apperror.ErrInvalidStatusTransition, order.Status, newStatus)
	}
	if newStatus == "rejected" && strings.TrimSpace(reason) == "" {
		return nil, apperror.Invalid("alasan penolakan wajib diisi")
	}
	tx, err := s.repo.DB().Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := s.repo.UpdateOrderStatus(ctx, tx, orderID, newStatus, reason, expectedVersion); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	updated, err := s.repo.FindOrderByID(ctx, orderID)
	if err != nil || updated == nil {
		// The write succeeded; failing to re-read it must not look like a failure.
		updated = order
		updated.Status = newStatus
	}

	s.broadcastStatus(updated, "")
	return updated, nil
}

func (s *Service) SubmitOrderFeedback(ctx context.Context, orderID uuid.UUID, req *dto.OrderFeedbackRequest) (*model.OrderFeedback, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	order, err := s.repo.FindOrderByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, apperror.ErrNotFound
	}
	if order.Status != "completed" {
		return nil, apperror.Invalid("feedback dapat diberikan setelah pesanan selesai")
	}
	return s.repo.UpsertOrderFeedback(ctx, orderID, req.Rating, req.Comment)
}

func isValidTransition(from, to string) bool {
	for _, allowed := range validTransitions[from] {
		if allowed == to {
			return true
		}
	}
	return false
}

// broadcastStatus notifies the one customer tracking this order plus the staff
// rooms that are allowed to see it.
func (s *Service) broadcastStatus(order *model.Order, message string) {
	payload := map[string]any{
		"order_id":     order.ID,
		"order_number": order.OrderNumber,
		"status":       order.Status,
		"version":      order.Version,
	}
	if message != "" {
		payload["message"] = message
	}
	if order.RejectionReason != nil && *order.RejectionReason != "" {
		payload["rejection_reason"] = *order.RejectionReason
	}

	s.hub.Broadcast(ws.OrderRoom(order.ID), "status_updated", payload)
	s.hub.Broadcast(ws.BranchRoom(order.BranchID), "status_updated", payload)
	s.hub.Broadcast(ws.RoomOwner, "status_updated", payload)
}

// CancelOrder is the customer-facing cancellation. Staff cancel through
// UpdateOrderStatus instead, which is not time-boxed.
func (s *Service) CancelOrder(ctx context.Context, orderID uuid.UUID) (*model.Order, error) {
	order, err := s.repo.FindOrderByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, apperror.ErrNotFound
	}

	if order.Status == "cancelled" {
		return order, nil
	}

	// Both conditions must hold. The original used && between them, so an order
	// that had sat pending for an hour was still cancellable.
	if !statusesCustomerMayCancel[order.Status] {
		return nil, apperror.Invalid("pesanan sudah diproses dan tidak dapat dibatalkan sendiri, silakan hubungi outlet")
	}
	if time.Since(order.CreatedAt) > customerCancelWindow {
		return nil, apperror.Invalid("batas waktu pembatalan mandiri (5 menit) telah lewat, silakan hubungi outlet")
	}

	// A paid order must be refunded through the payment gateway rather than
	// silently cancelled here.
	if order.Payment != nil && order.Payment.Status == "settlement" {
		return nil, apperror.Invalid("pesanan sudah dibayar, silakan hubungi outlet untuk pengembalian dana")
	}

	return s.UpdateOrderStatus(ctx, orderID, "cancelled", "Dibatalkan oleh pelanggan", order.Version)
}

// AcknowledgeOrders clears the admin's unread-order badge.
func (s *Service) AcknowledgeOrders(ctx context.Context, branchID *uuid.UUID, orderIDs []uuid.UUID, userID uuid.UUID) (int, error) {
	return s.repo.AcknowledgeOrders(ctx, branchID, orderIDs, userID)
}

func (s *Service) CountUnacknowledgedOrders(ctx context.Context, branchID *uuid.UUID) (int, error) {
	return s.repo.CountUnacknowledgedOrders(ctx, branchID)
}

// ---------------------------------------------------------------------
// Payment Service (Midtrans Integration + Double Payment Guard)
// ---------------------------------------------------------------------
func (s *Service) CreatePayment(ctx context.Context, req *dto.CreatePaymentRequest) (*model.Payment, error) {
	switch req.PaymentMethod {
	case "qris", "gopay", "shopeepay", "snap", "midtrans", "":
	default:
		return nil, apperror.Invalid("metode pembayaran tidak didukung")
	}
	if req.PaymentMethod == "" {
		req.PaymentMethod = "snap"
	}

	// Replaying the same Idempotency-Key returns the original payment rather
	// than charging twice.
	if existing, err := s.repo.FindPaymentByIdempotencyKey(ctx, req.IdempotencyKey); err == nil && existing != nil {
		if existing.OrderID != req.OrderID {
			return nil, apperror.Conflict("Idempotency-Key sudah dipakai untuk pesanan lain")
		}
		return existing, nil
	}

	tx, err := s.repo.DB().BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Serialise concurrent pay attempts for this order.
	lockKey := int64(crc32.ChecksumIEEE([]byte(req.OrderID.String())))
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", lockKey); err != nil {
		return nil, err
	}

	order, err := s.repo.FindOrderByID(ctx, req.OrderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, apperror.ErrNotFound
	}
	if order.Status != "pending" {
		return nil, apperror.ErrOrderNotPayable
	}

	if existingP, err := s.repo.FindPaymentByOrderID(ctx, req.OrderID); err == nil && existingP != nil {
		if existingP.Status == "settlement" {
			return nil, apperror.ErrOrderAlreadyPaid
		}
		// An unexpired pending attempt is reusable: hand back the same Snap
		// session instead of opening a second one.
		if existingP.Status == "pending" && time.Now().Before(existingP.ExpiresAt) {
			return existingP, nil
		}
	}

	expiresAt := time.Now().Add(paymentWindow)

	snapReq := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  order.OrderNumber,
			GrossAmt: int64(order.GrandTotal),
		},
		CustomerDetail: &midtrans.CustomerDetails{
			FName: order.CustomerName,
			Phone: order.CustomerPhone,
		},
		Items:           snapItems(order),
		EnabledPayments: nil, // Allow full Snap channels (QRIS, GoPay, ShopeePay, Bank Transfer, etc.)
		Expiry: &snap.ExpiryDetails{
			Unit:     "minute",
			Duration: int64(paymentWindow / time.Minute),
		},
		Callbacks: &snap.Callbacks{
			Finish: s.cfg.CustomerURL + "/order-success/" + order.ID.String(),
		},
	}

	snapResp, snapErr := s.snapClient.CreateTransaction(snapReq)
	if snapErr != nil {
		// Previously this branch invented a token and a QRIS payload and
		// reported success, so the customer saw "payment created" for a
		// transaction that did not exist. Surface the failure instead.
		slog.Error("midtrans snap transaction failed",
			"order_number", order.OrderNumber,
			"status_code", snapErr.StatusCode,
			"message", snapErr.Message,
		)
		return nil, fmt.Errorf("%w", apperror.ErrMidtransFailed)
	}
	if snapResp == nil || snapResp.Token == "" || snapResp.RedirectURL == "" {
		slog.Error("midtrans returned an empty snap session", "order_number", order.OrderNumber)
		return nil, apperror.ErrMidtransFailed
	}

	snapToken := snapResp.Token
	redirectURL := snapResp.RedirectURL

	payment := &model.Payment{
		OrderID:         order.ID,
		MidtransOrderID: order.OrderNumber,
		PaymentMethod:   req.PaymentMethod,
		PaymentType:     req.PaymentMethod,
		Status:          "pending",
		Amount:          order.GrandTotal,
		IdempotencyKey:  req.IdempotencyKey,
		SnapToken:       &snapToken,
		SnapRedirectURL: &redirectURL,
		ExpiresAt:       expiresAt,
	}

	if err := s.repo.CreatePayment(ctx, tx, payment); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return payment, nil
}

// paymentWindow is how long a Snap session stays valid. It is mirrored into
// the Midtrans expiry so both sides agree.
const paymentWindow = 15 * time.Minute

// enabledPaymentsFor restricts the Snap page to the single channel the
// customer chose, so the hidden VA and card options cannot reappear.
func enabledPaymentsFor(method string) []snap.SnapPaymentType {
	switch method {
	case "gopay":
		return []snap.SnapPaymentType{snap.PaymentTypeGopay}
	case "shopeepay":
		return []snap.SnapPaymentType{snap.PaymentTypeShopeepay}
	default:
		return []snap.SnapPaymentType{snap.SnapPaymentType("other_qris")}
	}
}

// snapItems itemises the basket for Midtrans. The line items must sum to
// gross_amount or Snap rejects the transaction, so fees and discounts are sent
// as their own lines.
func snapItems(order *model.Order) *[]midtrans.ItemDetails {
	items := make([]midtrans.ItemDetails, 0, len(order.Items)+3)

	for _, it := range order.Items {
		id := it.ItemName
		if it.MenuItemID != nil {
			id = it.MenuItemID.String()
		}
		items = append(items, midtrans.ItemDetails{
			ID:    id,
			Name:  truncate(it.ItemName, 50),
			Price: int64(it.ItemPrice),
			Qty:   int32(it.Quantity),
		})
	}
	if order.DeliveryFee > 0 {
		items = append(items, midtrans.ItemDetails{ID: "delivery_fee", Name: "Ongkos Kirim", Price: int64(order.DeliveryFee), Qty: 1})
	}
	if order.ServiceFee > 0 {
		items = append(items, midtrans.ItemDetails{ID: "service_fee", Name: "Biaya Layanan", Price: int64(order.ServiceFee), Qty: 1})
	}
	if order.Discount > 0 {
		items = append(items, midtrans.ItemDetails{ID: "discount", Name: "Diskon Promo", Price: -int64(order.Discount), Qty: 1})
	}

	return &items
}

// Midtrans rejects item names longer than 50 characters.
func truncate(v string, max int) string {
	r := []rune(v)
	if len(r) <= max {
		return v
	}
	return string(r[:max-1]) + "…"
}

// GetPaymentForOrder lets the success screen poll for settlement without
// exposing the whole payment table.
func (s *Service) GetPaymentForOrder(ctx context.Context, orderID uuid.UUID) (*model.Payment, error) {
	p, err := s.repo.FindPaymentByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, apperror.ErrNotFound
	}
	return p, nil
}

func (s *Service) HandleMidtransWebhook(ctx context.Context, payload map[string]any) error {
	orderID, _ := payload["order_id"].(string)
	statusCode, _ := payload["status_code"].(string)
	grossAmount, _ := payload["gross_amount"].(string)
	signatureKey, _ := payload["signature_key"].(string)
	txStatus, _ := payload["transaction_status"].(string)
	fraudStatus, _ := payload["fraud_status"].(string)
	txID, _ := payload["transaction_id"].(string)

	if orderID == "" {
		return apperror.Invalid("order_id kosong pada notifikasi")
	}

	// SHA512(order_id + status_code + gross_amount + server_key). This is the
	// only thing standing between the endpoint and anyone who can POST to it,
	// so it is verified in every environment, not just production.
	expected := sha512.Sum512([]byte(orderID + statusCode + grossAmount + s.cfg.MidtransServerKey))
	if subtle.ConstantTimeCompare([]byte(strings.ToLower(signatureKey)), []byte(hex.EncodeToString(expected[:]))) != 1 {
		slog.Warn("rejected midtrans notification with bad signature", "order_id", orderID)
		return apperror.ErrUnauthorized
	}

	order, err := s.repo.FindOrderByOrderNumber(ctx, orderID)
	if err != nil {
		return err
	}
	if order == nil {
		slog.Warn("midtrans notification for unknown order", "order_id", orderID)
		return apperror.ErrNotFound
	}

	// Only act on a transaction this server actually opened. Without a
	// payment row there is nothing to reconcile, and advancing the order
	// would accept an unpaid basket.
	payment, err := s.repo.FindPaymentByOrderID(ctx, order.ID)
	if err != nil {
		return err
	}
	if payment == nil {
		slog.Warn("midtrans notification for an order with no payment record", "order_id", orderID)
		return apperror.ErrNotFound
	}

	// Refunds are applied synchronously by RefundPayment when staff act in the
	// admin portal. This notification is Midtrans's own async confirmation of
	// that same refund, so there is nothing left to do here — running it
	// through the generic status switch below would wipe paid_at (it only
	// ever sets it, never preserves it) and stamp over the refund bookkeeping.
	if txStatus == "refund" || txStatus == "partial_refund" {
		slog.Info("midtrans refund notification acknowledged", "order_id", orderID, "status", txStatus)
		return nil
	}

	// Confirm the amount the gateway settled matches what we billed.
	if grossAmount != "" {
		if amount, convErr := strconv.ParseFloat(grossAmount, 64); convErr == nil {
			if int(math.Round(amount)) != order.GrandTotal {
				slog.Error("midtrans amount mismatch",
					"order_id", orderID, "expected", order.GrandTotal, "received", amount)
				return apperror.Conflict("jumlah pembayaran tidak sesuai dengan tagihan")
			}
		}
	}

	now := time.Now()
	var paidAt *time.Time
	var newPaymentStatus string

	switch txStatus {
	case "capture":
		// A captured card transaction is only money in the bank once fraud
		// review has accepted it.
		if fraudStatus == "deny" || fraudStatus == "challenge" {
			newPaymentStatus = "pending"
		} else {
			newPaymentStatus = "settlement"
			paidAt = &now
		}
	case "settlement":
		newPaymentStatus = "settlement"
		paidAt = &now
	case "expire":
		newPaymentStatus = "expire"
	case "cancel", "deny", "failure":
		newPaymentStatus = "cancel"
	case "pending":
		newPaymentStatus = "pending"
	default:
		slog.Warn("unhandled midtrans transaction_status", "status", txStatus, "order_id", orderID)
		newPaymentStatus = "pending"
	}

	tx, err := s.repo.DB().Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := s.repo.UpdatePaymentStatus(ctx, tx, orderID, newPaymentStatus, txID, payload, paidAt); err != nil {
		return err
	}

	// Move the order forward only on a real settlement, and only from pending,
	// so a duplicate notification cannot rewind an order already being made.
	advanced := false
	if newPaymentStatus == "settlement" && order.Status == "pending" {
		if err := s.repo.UpdateOrderStatus(ctx, tx, order.ID, "accepted", "Pembayaran diterima", order.Version); err != nil {
			// A concurrent staff action already advanced it; the payment
			// record is what matters here.
			if !errors.Is(err, apperror.ErrConcurrentModification) {
				return err
			}
		} else {
			advanced = true
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	if advanced {
		order.Status = "accepted"
		s.broadcastStatus(order, "Pembayaran lunas, pesanan Anda mulai diproses.")
		s.hub.Broadcast(ws.BranchRoom(order.BranchID), "new_order", order)
		s.hub.Broadcast(ws.BranchRoom(order.BranchID), "order_paid", order)
		s.hub.Broadcast(ws.RoomOwner, "new_order", order)
		s.hub.Broadcast(ws.RoomOwner, "order_paid", order)
	}

	return nil
}

// SyncPaymentWithMidtrans actively polls Midtrans status API to reconcile pending payments
func (s *Service) SyncPaymentWithMidtrans(ctx context.Context, orderNumber string) (*model.Payment, error) {
	baseURL := "https://api.sandbox.midtrans.com"
	if s.cfg.MidtransIsProd {
		baseURL = "https://api.midtrans.com"
	}

	reqURL := fmt.Sprintf("%s/v2/%s/status", baseURL, url.PathEscape(orderNumber))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(s.cfg.MidtransServerKey, "")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("midtrans status check returned %d", resp.StatusCode)
	}

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	if txStatus, _ := payload["transaction_status"].(string); txStatus == "settlement" || txStatus == "capture" {
		if err := s.HandleMidtransWebhook(ctx, payload); err != nil {
			slog.Warn("sync midtrans webhook handler returned error", "err", err)
		}
	}

	order, err := s.repo.FindOrderByOrderNumber(ctx, orderNumber)
	if err != nil || order == nil {
		return nil, err
	}
	return s.repo.FindPaymentByOrderID(ctx, order.ID)
}

// RefundPayment is the admin-portal counterpart to CreatePayment: staff issue
// a refund (full or partial) against a settled payment through Midtrans, and
// the order moves to the terminal "refunded" status. Unlike the ordinary
// status transitions, a refund can be issued from any paid status (accepted
// through completed, and even a rejected/cancelled order that was already
// charged) — the only real requirement is that the payment settled and has
// not already been refunded.
func (s *Service) RefundPayment(ctx context.Context, orderID uuid.UUID, amount int, reason string, actorID uuid.UUID, expectedVersion int) (*model.Order, error) {
	order, err := s.repo.FindOrderByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, apperror.ErrNotFound
	}
	if order.Version != expectedVersion {
		return nil, apperror.ErrConcurrentModification
	}
	if order.Status == "refunded" {
		return nil, apperror.Invalid("pesanan ini sudah direfund")
	}
	if !order.Payment.Refundable() {
		return nil, apperror.Invalid("pesanan ini belum lunas, tidak dapat direfund")
	}

	remaining := order.Payment.RemainingRefundable()
	if remaining <= 0 {
		return nil, apperror.Invalid("tidak ada sisa dana yang dapat direfund untuk pesanan ini")
	}
	if amount <= 0 {
		amount = remaining
	}
	if amount > remaining {
		return nil, apperror.Invalid(fmt.Sprintf("jumlah refund melebihi sisa yang dapat dikembalikan (maks Rp %s)", formatRupiah(remaining)))
	}

	// Stable per order, so a retried request after a network hiccup reaches
	// Midtrans as the same refund rather than a second one.
	refundKey := order.OrderNumber + "-refund"

	response, refundErr := s.midtransRefund(ctx, order.OrderNumber, refundKey, amount, reason)
	if refundErr != nil {
		slog.Error("midtrans refund failed", "order_number", order.OrderNumber, "err", refundErr)
		return nil, fmt.Errorf("%w", apperror.ErrMidtransFailed)
	}

	tx, err := s.repo.DB().Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := s.repo.RecordRefund(ctx, tx, order.ID, amount, reason, &actorID, response); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateOrderStatus(ctx, tx, order.ID, "refunded", reason, expectedVersion); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	updated, err := s.repo.FindOrderByID(ctx, order.ID)
	if err != nil || updated == nil {
		updated = order
		updated.Status = "refunded"
	}

	s.broadcastStatus(updated, "Pesanan direfund, dana akan kembali sesuai kebijakan Midtrans/penyedia pembayaran.")
	return updated, nil
}

// midtransRefund calls Midtrans's Core API refund endpoint directly: the
// snap-go SDK used for CreateTransaction has no refund support of its own.
func (s *Service) midtransRefund(ctx context.Context, orderNumber, refundKey string, amount int, reason string) (map[string]any, error) {
	baseURL := "https://api.sandbox.midtrans.com"
	if s.cfg.MidtransIsProd {
		baseURL = "https://api.midtrans.com"
	}

	body, err := json.Marshal(map[string]any{
		"refund_key": refundKey,
		"amount":     amount,
		"reason":     reason,
	})
	if err != nil {
		return nil, err
	}

	reqURL := fmt.Sprintf("%s/v2/%s/refund", baseURL, url.PathEscape(orderNumber))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(s.cfg.MidtransServerKey, "")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("midtrans refund returned %d and an unreadable body", resp.StatusCode)
	}

	statusCode, _ := payload["status_code"].(string)
	if resp.StatusCode >= 300 || (statusCode != "" && statusCode != "200" && statusCode != "201") {
		msg, _ := payload["status_message"].(string)
		return payload, fmt.Errorf("midtrans refund rejected (http %d, status_code %s): %s", resp.StatusCode, statusCode, msg)
	}

	return payload, nil
}

// ---------------------------------------------------------------------
// Promos & Settings
// ---------------------------------------------------------------------
func (s *Service) ValidatePromo(ctx context.Context, req *dto.ValidatePromoRequest) (*dto.ValidatePromoResponse, error) {
	code := strings.ToUpper(strings.TrimSpace(req.Code))
	if code == "" || len(code) > 50 {
		return nil, apperror.ErrInvalidPromo
	}
	if req.Subtotal < 0 || req.DeliveryFee < 0 {
		return nil, apperror.Invalid("rincian keranjang tidak valid")
	}

	promo, err := s.repo.FindPromoByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if promo == nil || !promo.Redeemable(time.Now()) {
		return nil, apperror.ErrInvalidPromo
	}
	if req.Subtotal < promo.MinSpend {
		return nil, apperror.Invalid(fmt.Sprintf("minimal belanja untuk promo ini adalah Rp %s", formatRupiah(promo.MinSpend)))
	}

	discount := 0
	msg := ""
	switch promo.Type {
	case "fixed":
		discount = promo.DiscountAmount
		msg = fmt.Sprintf("Potongan Rp %s berhasil digunakan!", formatRupiah(discount))
	case "free_delivery":
		discount = req.DeliveryFee
		if discount == 0 {
			msg = "Gratis ongkir diterapkan (ongkir Anda sudah Rp 0)."
		} else {
			msg = "Gratis ongkir berhasil diterapkan!"
		}
	default:
		return nil, apperror.ErrInvalidPromo
	}

	if discount > req.Subtotal+req.DeliveryFee {
		discount = req.Subtotal + req.DeliveryFee
	}

	finalTotal := req.Subtotal + req.DeliveryFee - discount
	if finalTotal < 0 {
		finalTotal = 0
	}

	return &dto.ValidatePromoResponse{
		Code:           promo.Code,
		DiscountAmount: discount,
		FinalTotal:     finalTotal,
		Message:        msg,
	}, nil
}

// ResolveBranchScope works out which branch an admin request targets. A branch
// admin is always pinned to their own outlet; the owner must say which one.
// Previously both fell back to a hardcoded branch UUID.
func ResolveBranchScope(actor *middleware.JWTClaims, requested *uuid.UUID) (*uuid.UUID, error) {
	if actor == nil {
		return nil, apperror.ErrUnauthorized
	}

	if actor.Role == "branch_admin" {
		if actor.BranchID == nil {
			return nil, apperror.ErrForbidden
		}
		if requested != nil && *requested != *actor.BranchID {
			return nil, apperror.ErrForbidden
		}
		return actor.BranchID, nil
	}

	// Owner: nil means "the whole network".
	return requested, nil
}

func (s *Service) GetBranchSettings(ctx context.Context, branchID uuid.UUID) (*model.BranchSettings, error) {
	settings, err := s.repo.FindSettingsByBranchID(ctx, branchID)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		return nil, apperror.ErrNotFound
	}
	return settings, nil
}

func (s *Service) UpdateBranchSettings(ctx context.Context, branchID uuid.UUID, req *dto.UpdateBranchSettingsRequest) (*model.BranchSettings, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	settings := &model.BranchSettings{
		BranchID:              branchID,
		OperatingHours:        req.OperatingHours,
		MaxDeliveryRadiusKm:   10,
		BaseDeliveryFeeNear:   0,
		BaseDeliveryFeeMid:    8000,
		BaseDeliveryFeeFar:    12000,
		NearThresholdKm:       1,
		MidThresholdKm:        5,
		ServiceFee:            req.ServiceFee,
		MinOrderAmount:        req.MinOrderAmount,
		FreeDeliveryThreshold: 0,
		WhatsappNumber:        req.WhatsappNumber,
		HalalCertificateID:    req.HalalCertificateID,
	}
	if req.Description != "" {
		settings.Description = &req.Description
	}

	if err := s.repo.UpdateSettings(ctx, settings); err != nil {
		return nil, err
	}
	return s.repo.FindSettingsByBranchID(ctx, branchID)
}

func (s *Service) UpdateBranchStatus(ctx context.Context, branchID uuid.UUID, isOpen bool) error {
	return s.repo.UpdateBranchStatus(ctx, branchID, isOpen)
}

// UpdateBranchProfile lets a branch_admin (own outlet) or owner (any outlet)
// correct the outlet's public identity. This replaced the previous state
// where name/address/phone/coordinates could only be set once, at seed time,
// with no way to fix leftover placeholder data short of a manual DB edit.
func (s *Service) UpdateBranchProfile(ctx context.Context, branchID uuid.UUID, req *dto.UpdateBranchProfileRequest) (*model.Branch, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateBranchProfile(ctx, branchID, req.Name, req.Address, req.Phone, req.Latitude, req.Longitude); err != nil {
		return nil, err
	}
	return s.repo.FindBranchByID(ctx, branchID)
}

// ---------------------------------------------------------------------
// Upload Service (With Auto-Compression)
// ---------------------------------------------------------------------
func (s *Service) UploadAndCompressImage(ctx context.Context, fileHeader *multipart.FileHeader) (string, error) {
	src, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Ensure upload dir exists
	uploadDir := s.cfg.UploadDir
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	// Decode image
	img, _, err := image.Decode(src)
	if err != nil {
		return "", apperror.Invalid("format gambar tidak valid atau file rusak")
	}

	// Auto-resize max dimension 800px keeping aspect ratio
	resized := imaging.Fit(img, 800, 800, imaging.Lanczos)

	// Save compressed JPEG with 80% quality
	filename := fmt.Sprintf("%s.jpg", uuid.NewString())
	targetPath := filepath.Join(uploadDir, filename)
	out, err := os.Create(targetPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	err = jpeg.Encode(out, resized, &jpeg.Options{Quality: 80})
	if err != nil {
		return "", err
	}

	return "/uploads/" + filename, nil
}

// ---------------------------------------------------------------------
// Analytics Service
// ---------------------------------------------------------------------

// GetDashboardStats returns today's figures. A nil branchID means the whole
// network, which only the owner may request.
func (s *Service) GetDashboardStats(ctx context.Context, branchID *uuid.UUID) (*dto.DashboardStats, error) {
	return s.repo.GetDashboardStats(ctx, branchID, OutletLocation)
}
