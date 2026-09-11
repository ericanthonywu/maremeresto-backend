package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ericanthonywu/maremereso-olga/backend/internal/apperror"
	"github.com/ericanthonywu/maremereso-olga/backend/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// DB getter for transactions
func (r *Repository) DB() *pgxpool.Pool {
	return r.db
}

// ---------------------------------------------------------------------
// User Repository
// ---------------------------------------------------------------------
func (r *Repository) FindUserByPhone(ctx context.Context, phone string) (*model.User, error) {
	query := `SELECT id, phone, name, role, branch_id, password_hash, address, latitude, longitude, created_at, updated_at FROM users WHERE phone = $1`
	var u model.User
	err := r.db.QueryRow(ctx, query, phone).Scan(
		&u.ID, &u.Phone, &u.Name, &u.Role, &u.BranchID, &u.PasswordHash, &u.Address, &u.Latitude, &u.Longitude, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// FindUserByEmail resolves a staff sign-in. Email replaces the hardcoded
// email->phone switch that previously lived in the service layer.
func (r *Repository) FindUserByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `SELECT id, phone, name, role, branch_id, password_hash, address, latitude, longitude, created_at, updated_at FROM users WHERE LOWER(email) = LOWER($1)`
	var u model.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.Phone, &u.Name, &u.Role, &u.BranchID, &u.PasswordHash, &u.Address, &u.Latitude, &u.Longitude, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) FindUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	query := `SELECT id, phone, name, role, branch_id, password_hash, address, latitude, longitude, created_at, updated_at FROM users WHERE id = $1`
	var u model.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Phone, &u.Name, &u.Role, &u.BranchID, &u.PasswordHash, &u.Address, &u.Latitude, &u.Longitude, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) CreateCustomer(ctx context.Context, phone, name string) (*model.User, error) {
	query := `INSERT INTO users (phone, name, role) VALUES ($1, $2, 'customer') RETURNING id, phone, name, role, branch_id, created_at, updated_at`
	var u model.User
	err := r.db.QueryRow(ctx, query, phone, name).Scan(
		&u.ID, &u.Phone, &u.Name, &u.Role, &u.BranchID, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// ---------------------------------------------------------------------
// Branch Repository
// ---------------------------------------------------------------------
func (r *Repository) ListBranches(ctx context.Context) ([]model.Branch, error) {
	query := `SELECT id, slug, name, address, phone, latitude, longitude, gradient_theme, icon, facility_tags, rating, is_open, created_at, updated_at FROM branches ORDER BY name ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	branches := make([]model.Branch, 0)
	for rows.Next() {
		var b model.Branch
		var tagsJSON []byte
		if err := rows.Scan(
			&b.ID, &b.Slug, &b.Name, &b.Address, &b.Phone, &b.Latitude, &b.Longitude,
			&b.GradientTheme, &b.Icon, &tagsJSON, &b.Rating, &b.IsOpen, &b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(tagsJSON, &b.FacilityTags)
		branches = append(branches, b)
	}
	return branches, nil
}

func (r *Repository) FindBranchBySlug(ctx context.Context, slug string) (*model.Branch, error) {
	query := `SELECT id, slug, name, address, phone, latitude, longitude, gradient_theme, icon, facility_tags, rating, is_open, created_at, updated_at FROM branches WHERE slug = $1`
	var b model.Branch
	var tagsJSON []byte
	err := r.db.QueryRow(ctx, query, slug).Scan(
		&b.ID, &b.Slug, &b.Name, &b.Address, &b.Phone, &b.Latitude, &b.Longitude,
		&b.GradientTheme, &b.Icon, &tagsJSON, &b.Rating, &b.IsOpen, &b.CreatedAt, &b.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(tagsJSON, &b.FacilityTags)
	return &b, nil
}

func (r *Repository) FindBranchByID(ctx context.Context, id uuid.UUID) (*model.Branch, error) {
	query := `SELECT id, slug, name, address, phone, latitude, longitude, gradient_theme, icon, facility_tags, rating, is_open, created_at, updated_at FROM branches WHERE id = $1`
	var b model.Branch
	var tagsJSON []byte
	err := r.db.QueryRow(ctx, query, id).Scan(
		&b.ID, &b.Slug, &b.Name, &b.Address, &b.Phone, &b.Latitude, &b.Longitude,
		&b.GradientTheme, &b.Icon, &tagsJSON, &b.Rating, &b.IsOpen, &b.CreatedAt, &b.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(tagsJSON, &b.FacilityTags)
	return &b, nil
}

func (r *Repository) UpdateBranchStatus(ctx context.Context, id uuid.UUID, isOpen bool) error {
	query := `UPDATE branches SET is_open = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, isOpen, id)
	return err
}

// ---------------------------------------------------------------------
// Menu & Category Repository
// ---------------------------------------------------------------------
func (r *Repository) ListCategories(ctx context.Context) ([]model.Category, error) {
	query := `SELECT id, name, slug, emoji, sort_order, created_at FROM categories ORDER BY sort_order ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]model.Category, 0)
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Emoji, &c.SortOrder, &c.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, nil
}

func (r *Repository) ListMenuItemsByBranch(ctx context.Context, branchID uuid.UUID) ([]model.MenuItem, error) {
	query := `
		SELECT m.id, m.branch_id, m.category_id, m.name, m.description, m.price, m.icon, m.icon_bg_class,
		       m.image_url, m.tag, m.is_available, m.sort_order, m.created_at, m.updated_at,
		       c.name, c.slug, c.emoji
		FROM menu_items m
		JOIN categories c ON m.category_id = c.id
		WHERE m.branch_id = $1
		ORDER BY c.sort_order ASC, m.sort_order ASC
	`
	rows, err := r.db.Query(ctx, query, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.MenuItem, 0)
	for rows.Next() {
		var m model.MenuItem
		var catName, catSlug, catEmoji string
		if err := rows.Scan(
			&m.ID, &m.BranchID, &m.CategoryID, &m.Name, &m.Description, &m.Price, &m.Icon, &m.IconBgClass,
			&m.ImageURL, &m.Tag, &m.IsAvailable, &m.SortOrder, &m.CreatedAt, &m.UpdatedAt,
			&catName, &catSlug, &catEmoji,
		); err != nil {
			return nil, err
		}
		m.Category = &model.Category{
			ID:    m.CategoryID,
			Name:  catName,
			Slug:  catSlug,
			Emoji: catEmoji,
		}
		items = append(items, m)
	}
	return items, nil
}

func (r *Repository) FindMenuItemByID(ctx context.Context, id uuid.UUID) (*model.MenuItem, error) {
	query := `
		SELECT m.id, m.branch_id, m.category_id, m.name, m.description, m.price, m.icon, m.icon_bg_class,
		       m.image_url, m.tag, m.is_available, m.sort_order, m.created_at, m.updated_at,
		       c.name, c.slug, c.emoji
		FROM menu_items m
		JOIN categories c ON m.category_id = c.id
		WHERE m.id = $1
	`
	var m model.MenuItem
	var catName, catSlug, catEmoji string
	err := r.db.QueryRow(ctx, query, id).Scan(
		&m.ID, &m.BranchID, &m.CategoryID, &m.Name, &m.Description, &m.Price, &m.Icon, &m.IconBgClass,
		&m.ImageURL, &m.Tag, &m.IsAvailable, &m.SortOrder, &m.CreatedAt, &m.UpdatedAt,
		&catName, &catSlug, &catEmoji,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	m.Category = &model.Category{
		ID:    m.CategoryID,
		Name:  catName,
		Slug:  catSlug,
		Emoji: catEmoji,
	}
	return &m, nil
}

func (r *Repository) CreateMenuItem(ctx context.Context, m *model.MenuItem) error {
	query := `
		INSERT INTO menu_items (branch_id, category_id, name, description, price, icon, icon_bg_class, image_url, tag, is_available, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query,
		m.BranchID, m.CategoryID, m.Name, m.Description, m.Price, m.Icon, m.IconBgClass, m.ImageURL, m.Tag, m.IsAvailable, m.SortOrder,
	).Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)
}

func (r *Repository) UpdateMenuItem(ctx context.Context, m *model.MenuItem) error {
	query := `
		UPDATE menu_items
		SET category_id = $1, name = $2, description = $3, price = $4, icon = $5, icon_bg_class = $6, image_url = $7, tag = $8, is_available = $9, updated_at = NOW()
		WHERE id = $10
	`
	_, err := r.db.Exec(ctx, query, m.CategoryID, m.Name, m.Description, m.Price, m.Icon, m.IconBgClass, m.ImageURL, m.Tag, m.IsAvailable, m.ID)
	return err
}

func (r *Repository) ToggleMenuItemAvailability(ctx context.Context, id uuid.UUID, isAvailable bool) error {
	query := `UPDATE menu_items SET is_available = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, isAvailable, id)
	return err
}

func (r *Repository) DeleteMenuItem(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM menu_items WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// ---------------------------------------------------------------------
// Order Repository
// ---------------------------------------------------------------------
func (r *Repository) NextOrderNumber(ctx context.Context, tx pgx.Tx) (string, error) {
	var seq int
	err := tx.QueryRow(ctx, `SELECT nextval('order_number_seq')`).Scan(&seq)
	if err != nil {
		return "", err
	}
	dateStr := time.Now().Format("20060102")
	return fmt.Sprintf("OLG-%s-%04d", dateStr, seq), nil
}

func (r *Repository) CreateOrder(ctx context.Context, tx pgx.Tx, order *model.Order) error {
	query := `
		INSERT INTO orders (
			order_number, user_id, branch_id, order_type, status,
			customer_name, customer_phone, delivery_address, delivery_notes,
			delivery_lat, delivery_lon, delivery_distance_km,
			subtotal, delivery_fee, service_fee, discount, grand_total,
			promo_code, scheduled_at, version
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, 1
		) RETURNING id, version, created_at, updated_at
	`
	err := tx.QueryRow(ctx, query,
		order.OrderNumber, order.UserID, order.BranchID, order.OrderType, order.Status,
		order.CustomerName, order.CustomerPhone, order.DeliveryAddress, order.DeliveryNotes,
		order.DeliveryLat, order.DeliveryLon, order.DeliveryDistanceKm,
		order.Subtotal, order.DeliveryFee, order.ServiceFee, order.Discount, order.GrandTotal,
		order.PromoCode, order.ScheduledAt,
	).Scan(&order.ID, &order.Version, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return err
	}

	// Insert items
	itemQuery := `
		INSERT INTO order_items (order_id, menu_item_id, item_name, item_price, item_icon, quantity, notes, line_total)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at
	`
	for i := range order.Items {
		item := &order.Items[i]
		item.OrderID = order.ID
		err = tx.QueryRow(ctx, itemQuery,
			item.OrderID, item.MenuItemID, item.ItemName, item.ItemPrice, item.ItemIcon,
			item.Quantity, item.Notes, item.LineTotal,
		).Scan(&item.ID, &item.CreatedAt)
		if err != nil {
			return err
		}
	}

	// Add initial history
	_, err = tx.Exec(ctx, `INSERT INTO order_status_history (order_id, to_status, notes) VALUES ($1, $2, $3)`,
		order.ID, order.Status, "Pesanan dibuat")
	return err
}

func (r *Repository) FindOrderByID(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	query := `
		SELECT o.id, o.order_number, o.user_id, o.branch_id, o.order_type, o.status,
		       o.customer_name, o.customer_phone, o.delivery_address, o.delivery_notes,
		       o.delivery_lat, o.delivery_lon, o.delivery_distance_km,
		       o.subtotal, o.delivery_fee, o.service_fee, o.discount, o.grand_total,
		       o.promo_code, o.scheduled_at, o.driver_name, o.driver_phone, o.driver_vehicle,
		       o.driver_plate, o.driver_rating, o.driver_assigned_at, o.acknowledged_at,
		       o.rejection_reason, o.version, o.created_at, o.updated_at,
		       b.name, b.slug, b.address, b.phone
		FROM orders o
		JOIN branches b ON o.branch_id = b.id
		WHERE o.id = $1
	`
	var o model.Order
	var bName, bSlug, bAddr, bPhone string
	err := r.db.QueryRow(ctx, query, id).Scan(
		&o.ID, &o.OrderNumber, &o.UserID, &o.BranchID, &o.OrderType, &o.Status,
		&o.CustomerName, &o.CustomerPhone, &o.DeliveryAddress, &o.DeliveryNotes,
		&o.DeliveryLat, &o.DeliveryLon, &o.DeliveryDistanceKm,
		&o.Subtotal, &o.DeliveryFee, &o.ServiceFee, &o.Discount, &o.GrandTotal,
		&o.PromoCode, &o.ScheduledAt, &o.DriverName, &o.DriverPhone, &o.DriverVehicle,
		&o.DriverPlate, &o.DriverRating, &o.DriverAssignedAt, &o.AcknowledgedAt,
		&o.RejectionReason, &o.Version, &o.CreatedAt, &o.UpdatedAt,
		&bName, &bSlug, &bAddr, &bPhone,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	o.Branch = &model.Branch{
		ID:      o.BranchID,
		Name:    bName,
		Slug:    bSlug,
		Address: bAddr,
		Phone:   bPhone,
	}

	// Fetch items
	items, err := r.ListOrderItems(ctx, o.ID)
	if err != nil {
		return nil, err
	}
	o.Items = items

	// Fetch payment
	p, err := r.FindPaymentByOrderID(ctx, o.ID)
	if err == nil && p != nil {
		o.Payment = p
	}

	return &o, nil
}

func (r *Repository) FindOrderByOrderNumber(ctx context.Context, num string) (*model.Order, error) {
	query := `SELECT id FROM orders WHERE order_number = $1`
	var id uuid.UUID
	err := r.db.QueryRow(ctx, query, num).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.FindOrderByID(ctx, id)
}

func (r *Repository) ListOrderItems(ctx context.Context, orderID uuid.UUID) ([]model.OrderItem, error) {
	query := `SELECT id, order_id, menu_item_id, item_name, item_price, item_icon, quantity, notes, line_total, created_at FROM order_items WHERE order_id = $1`
	rows, err := r.db.Query(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]model.OrderItem, 0)
	for rows.Next() {
		var it model.OrderItem
		if err := rows.Scan(
			&it.ID, &it.OrderID, &it.MenuItemID, &it.ItemName, &it.ItemPrice, &it.ItemIcon,
			&it.Quantity, &it.Notes, &it.LineTotal, &it.CreatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, it)
	}
	return list, nil
}

// ListOrdersForCustomer backs the customer's own order history, scoped to the
// signed-in account.
func (r *Repository) ListOrdersForCustomer(ctx context.Context, userID uuid.UUID, limit, offset int) ([]model.Order, int, error) {
	return r.listOrders(ctx, orderFilter{UserID: &userID}, limit, offset)
}

func (r *Repository) ListOrders(ctx context.Context, branchID *uuid.UUID, status string, search string, limit, offset int) ([]model.Order, int, error) {
	return r.listOrders(ctx, orderFilter{BranchID: branchID, Status: status, Search: search}, limit, offset)
}

type orderFilter struct {
	BranchID *uuid.UUID
	UserID   *uuid.UUID
	Status   string
	Search   string
}

func (r *Repository) listOrders(ctx context.Context, f orderFilter, limit, offset int) ([]model.Order, int, error) {
	branchID, status, search := f.BranchID, f.Status, f.Search
	baseQuery := `FROM orders o JOIN branches b ON o.branch_id = b.id WHERE 1=1`
	args := []any{}
	idx := 1

	if f.UserID != nil {
		baseQuery += fmt.Sprintf(" AND o.user_id = $%d", idx)
		args = append(args, *f.UserID)
		idx++
	}
	if branchID != nil {
		baseQuery += fmt.Sprintf(" AND o.branch_id = $%d", idx)
		args = append(args, *branchID)
		idx++
	}
	if status != "" && status != "all" {
		baseQuery += fmt.Sprintf(" AND o.status = $%d", idx)
		args = append(args, status)
		idx++
	}
	if search != "" {
		baseQuery += fmt.Sprintf(" AND (o.order_number ILIKE $%d OR o.customer_name ILIKE $%d OR o.customer_phone ILIKE $%d)", idx, idx, idx)
		args = append(args, "%"+search+"%")
		idx++
	}

	var total int
	err := r.db.QueryRow(ctx, "SELECT count(*) "+baseQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	selectQuery := `
		SELECT o.id, o.order_number, o.user_id, o.branch_id, o.order_type, o.status,
		       o.customer_name, o.customer_phone, o.delivery_address, o.delivery_notes,
		       o.delivery_lat, o.delivery_lon, o.delivery_distance_km,
		       o.subtotal, o.delivery_fee, o.service_fee, o.discount, o.grand_total,
		       o.promo_code, o.scheduled_at, o.driver_name, o.driver_phone, o.driver_vehicle,
		       o.driver_plate, o.driver_rating, o.driver_assigned_at, o.acknowledged_at,
		       o.rejection_reason, o.version, o.created_at, o.updated_at,
		       b.name, b.slug
		` + baseQuery + fmt.Sprintf(" ORDER BY o.created_at DESC LIMIT $%d OFFSET $%d", idx, idx+1)

	args = append(args, limit, offset)
	rows, err := r.db.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	orders := make([]model.Order, 0)
	for rows.Next() {
		var o model.Order
		var bName, bSlug string
		if err := rows.Scan(
			&o.ID, &o.OrderNumber, &o.UserID, &o.BranchID, &o.OrderType, &o.Status,
			&o.CustomerName, &o.CustomerPhone, &o.DeliveryAddress, &o.DeliveryNotes,
			&o.DeliveryLat, &o.DeliveryLon, &o.DeliveryDistanceKm,
			&o.Subtotal, &o.DeliveryFee, &o.ServiceFee, &o.Discount, &o.GrandTotal,
			&o.PromoCode, &o.ScheduledAt, &o.DriverName, &o.DriverPhone, &o.DriverVehicle,
			&o.DriverPlate, &o.DriverRating, &o.DriverAssignedAt, &o.AcknowledgedAt,
			&o.RejectionReason, &o.Version, &o.CreatedAt, &o.UpdatedAt,
			&bName, &bSlug,
		); err != nil {
			return nil, 0, err
		}
		o.Branch = &model.Branch{
			ID:   o.BranchID,
			Name: bName,
			Slug: bSlug,
		}
		orders = append(orders, o)
	}

	// Attach items in one round trip rather than a query per order.
	if err := r.attachOrderItems(ctx, orders); err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// attachOrderItems loads the line items for a page of orders with a single
// query and fills them in place.
func (r *Repository) attachOrderItems(ctx context.Context, orders []model.Order) error {
	if len(orders) == 0 {
		return nil
	}

	ids := make([]uuid.UUID, len(orders))
	index := make(map[uuid.UUID]int, len(orders))
	for i := range orders {
		ids[i] = orders[i].ID
		index[orders[i].ID] = i
	}

	query := `
		SELECT id, order_id, menu_item_id, item_name, item_price, item_icon, quantity, notes, line_total, created_at
		FROM order_items
		WHERE order_id = ANY($1)
		ORDER BY created_at ASC, id ASC
	`
	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var it model.OrderItem
		if err := rows.Scan(
			&it.ID, &it.OrderID, &it.MenuItemID, &it.ItemName, &it.ItemPrice, &it.ItemIcon,
			&it.Quantity, &it.Notes, &it.LineTotal, &it.CreatedAt,
		); err != nil {
			return err
		}
		if i, ok := index[it.OrderID]; ok {
			orders[i].Items = append(orders[i].Items, it)
		}
	}
	return rows.Err()
}

func (r *Repository) UpdateOrderStatus(ctx context.Context, tx pgx.Tx, orderID uuid.UUID, newStatus string, rejectionReason string, expectedVersion int) error {
	query := `
		UPDATE orders
		SET status = $1, rejection_reason = $2, version = version + 1, updated_at = NOW()
		WHERE id = $3 AND version = $4
	`
	res, err := tx.Exec(ctx, query, newStatus, rejectionReason, orderID, expectedVersion)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return apperror.ErrConcurrentModification
	}

	// Insert history
	_, err = tx.Exec(ctx, `INSERT INTO order_status_history (order_id, to_status, notes) VALUES ($1, $2, $3)`,
		orderID, newStatus, rejectionReason)
	return err
}

// ---------------------------------------------------------------------
// Payment Repository (With Idempotency & Locking)
// ---------------------------------------------------------------------
func (r *Repository) FindPaymentByOrderID(ctx context.Context, orderID uuid.UUID) (*model.Payment, error) {
	query := `
		SELECT id, order_id, midtrans_order_id, payment_method, payment_type, status, amount,
		       idempotency_key, snap_token, snap_redirect_url, qr_string, midtrans_transaction_id,
		       paid_at, expires_at, created_at, updated_at
		FROM payments WHERE order_id = $1
	`
	var p model.Payment
	err := r.db.QueryRow(ctx, query, orderID).Scan(
		&p.ID, &p.OrderID, &p.MidtransOrderID, &p.PaymentMethod, &p.PaymentType, &p.Status, &p.Amount,
		&p.IdempotencyKey, &p.SnapToken, &p.SnapRedirectURL, &p.QRString, &p.MidtransTransactionID,
		&p.PaidAt, &p.ExpiresAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) FindPaymentByIdempotencyKey(ctx context.Context, key string) (*model.Payment, error) {
	query := `
		SELECT id, order_id, midtrans_order_id, payment_method, payment_type, status, amount,
		       idempotency_key, snap_token, snap_redirect_url, qr_string, midtrans_transaction_id,
		       paid_at, expires_at, created_at, updated_at
		FROM payments WHERE idempotency_key = $1
	`
	var p model.Payment
	err := r.db.QueryRow(ctx, query, key).Scan(
		&p.ID, &p.OrderID, &p.MidtransOrderID, &p.PaymentMethod, &p.PaymentType, &p.Status, &p.Amount,
		&p.IdempotencyKey, &p.SnapToken, &p.SnapRedirectURL, &p.QRString, &p.MidtransTransactionID,
		&p.PaidAt, &p.ExpiresAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) CreatePayment(ctx context.Context, tx pgx.Tx, p *model.Payment) error {
	query := `
		INSERT INTO payments (
			order_id, midtrans_order_id, payment_method, payment_type, status, amount,
			idempotency_key, snap_token, snap_redirect_url, qr_string, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`
	return tx.QueryRow(ctx, query,
		p.OrderID, p.MidtransOrderID, p.PaymentMethod, p.PaymentType, p.Status, p.Amount,
		p.IdempotencyKey, p.SnapToken, p.SnapRedirectURL, p.QRString, p.ExpiresAt,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *Repository) UpdatePaymentStatus(ctx context.Context, tx pgx.Tx, midtransOrderID string, status string, txID string, responseJSON map[string]any, paidAt *time.Time) error {
	jsonBytes, _ := json.Marshal(responseJSON)
	query := `
		UPDATE payments
		SET status = $1, midtrans_transaction_id = $2, midtrans_response = $3, paid_at = $4, updated_at = NOW()
		WHERE midtrans_order_id = $5
	`
	_, err := tx.Exec(ctx, query, status, txID, jsonBytes, paidAt, midtransOrderID)
	return err
}

// ---------------------------------------------------------------------
// Promo Repository
// ---------------------------------------------------------------------
func (r *Repository) FindPromoByCode(ctx context.Context, code string) (*model.Promo, error) {
	query := `
		SELECT id, code, type, discount_amount, is_active, min_spend,
		       valid_from, valid_until, max_redemptions, redemption_count, created_at
		FROM promos WHERE UPPER(code) = UPPER($1)
	`
	var p model.Promo
	err := r.db.QueryRow(ctx, query, code).Scan(
		&p.ID, &p.Code, &p.Type, &p.DiscountAmount, &p.IsActive, &p.MinSpend,
		&p.ValidFrom, &p.ValidUntil, &p.MaxRedemptions, &p.RedemptionCount, &p.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// ---------------------------------------------------------------------
// Branch Settings Repository
// ---------------------------------------------------------------------
func (r *Repository) FindSettingsByBranchID(ctx context.Context, branchID uuid.UUID) (*model.BranchSettings, error) {
	query := `
		SELECT id, branch_id, operating_hours, max_delivery_radius_km, base_delivery_fee_near,
		       base_delivery_fee_mid, base_delivery_fee_far, near_threshold_km, mid_threshold_km,
		       service_fee, min_order_amount, free_delivery_threshold, whatsapp_number, description, updated_at
		FROM branch_settings WHERE branch_id = $1
	`
	var s model.BranchSettings
	var hoursJSON []byte
	err := r.db.QueryRow(ctx, query, branchID).Scan(
		&s.ID, &s.BranchID, &hoursJSON, &s.MaxDeliveryRadiusKm, &s.BaseDeliveryFeeNear,
		&s.BaseDeliveryFeeMid, &s.BaseDeliveryFeeFar, &s.NearThresholdKm, &s.MidThresholdKm,
		&s.ServiceFee, &s.MinOrderAmount, &s.FreeDeliveryThreshold, &s.WhatsappNumber, &s.Description, &s.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(hoursJSON, &s.OperatingHours)
	return &s, nil
}

func (r *Repository) UpdateSettings(ctx context.Context, s *model.BranchSettings) error {
	hoursBytes, _ := json.Marshal(s.OperatingHours)
	query := `
		UPDATE branch_settings
		SET operating_hours = $1, max_delivery_radius_km = $2, base_delivery_fee_near = $3,
		    base_delivery_fee_mid = $4, base_delivery_fee_far = $5, near_threshold_km = $6,
		    mid_threshold_km = $7, service_fee = $8, min_order_amount = $9, free_delivery_threshold = $10,
		    whatsapp_number = $11, description = $12, updated_at = NOW()
		WHERE branch_id = $13
	`
	res, err := r.db.Exec(ctx, query,
		hoursBytes, s.MaxDeliveryRadiusKm, s.BaseDeliveryFeeNear, s.BaseDeliveryFeeMid, s.BaseDeliveryFeeFar,
		s.NearThresholdKm, s.MidThresholdKm, s.ServiceFee, s.MinOrderAmount, s.FreeDeliveryThreshold,
		s.WhatsappNumber, s.Description, s.BranchID,
	)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}
	return nil
}

// ---------------------------------------------------------------------
// Order lifecycle extras: driver assignment & staff acknowledgement
// ---------------------------------------------------------------------

// AssignDriver records the courier handling an order. Driver details used to
// come from column defaults ("Andi Pratama", "B 1234 ABC"), which meant every
// customer saw the same invented courier.
func (r *Repository) AssignDriver(ctx context.Context, orderID uuid.UUID, name, phone, vehicle, plate string) error {
	query := `
		UPDATE orders
		SET driver_name = $1, driver_phone = $2, driver_vehicle = $3, driver_plate = $4,
		    driver_assigned_at = NOW(), updated_at = NOW()
		WHERE id = $5
	`
	res, err := r.db.Exec(ctx, query, name, phone, vehicle, plate, orderID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}
	return nil
}

// AcknowledgeOrders marks orders as seen by staff so the admin unread badge is
// backed by the database and survives a page reload.
func (r *Repository) AcknowledgeOrders(ctx context.Context, branchID *uuid.UUID, orderIDs []uuid.UUID, userID uuid.UUID) (int, error) {
	query := `
		UPDATE orders
		SET acknowledged_at = NOW(), acknowledged_by = $1
		WHERE acknowledged_at IS NULL
		  AND ($2::uuid IS NULL OR branch_id = $2)
		  AND ($3::uuid[] IS NULL OR id = ANY($3))
	`
	var ids any
	if len(orderIDs) > 0 {
		ids = orderIDs
	}
	res, err := r.db.Exec(ctx, query, userID, branchID, ids)
	if err != nil {
		return 0, err
	}
	return int(res.RowsAffected()), nil
}

// CountUnacknowledgedOrders powers the admin notification badge.
func (r *Repository) CountUnacknowledgedOrders(ctx context.Context, branchID *uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*) FROM orders
		WHERE acknowledged_at IS NULL
		  AND status NOT IN ('cancelled', 'rejected', 'completed')
		  AND ($1::uuid IS NULL OR branch_id = $1)
	`
	var count int
	err := r.db.QueryRow(ctx, query, branchID).Scan(&count)
	return count, err
}

// IncrementPromoRedemption is called once an order that used a promo is
// committed, so max_redemptions is actually enforced.
func (r *Repository) IncrementPromoRedemption(ctx context.Context, tx pgx.Tx, code string) error {
	_, err := tx.Exec(ctx,
		`UPDATE promos SET redemption_count = redemption_count + 1 WHERE UPPER(code) = UPPER($1)`, code)
	return err
}

// FindMenuItemsByIDs loads several menu items at once for order validation,
// replacing a query per line item.
func (r *Repository) FindMenuItemsByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]model.MenuItem, error) {
	result := make(map[uuid.UUID]model.MenuItem, len(ids))
	if len(ids) == 0 {
		return result, nil
	}

	query := `
		SELECT id, branch_id, category_id, name, description, price, icon, icon_bg_class,
		       image_url, tag, is_available, sort_order
		FROM menu_items
		WHERE id = ANY($1)
	`
	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m model.MenuItem
		if err := rows.Scan(
			&m.ID, &m.BranchID, &m.CategoryID, &m.Name, &m.Description, &m.Price, &m.Icon, &m.IconBgClass,
			&m.ImageURL, &m.Tag, &m.IsAvailable, &m.SortOrder,
		); err != nil {
			return nil, err
		}
		result[m.ID] = m
	}
	return result, rows.Err()
}

// CategoryExists guards menu writes against a bogus category_id, which would
// otherwise surface as an opaque foreign-key error.
func (r *Repository) CategoryExists(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM categories WHERE id = $1)`, id).Scan(&exists)
	return exists, err
}

// ListSettingsByBranch returns every branch's settings keyed by branch id, so
// listing branches costs one query instead of one per branch.
func (r *Repository) ListSettingsByBranch(ctx context.Context) (map[uuid.UUID]model.BranchSettings, error) {
	query := `
		SELECT id, branch_id, operating_hours, max_delivery_radius_km, base_delivery_fee_near,
		       base_delivery_fee_mid, base_delivery_fee_far, near_threshold_km, mid_threshold_km,
		       service_fee, min_order_amount, free_delivery_threshold, whatsapp_number, description, updated_at
		FROM branch_settings
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[uuid.UUID]model.BranchSettings)
	for rows.Next() {
		var s model.BranchSettings
		var hoursJSON []byte
		if err := rows.Scan(
			&s.ID, &s.BranchID, &hoursJSON, &s.MaxDeliveryRadiusKm, &s.BaseDeliveryFeeNear,
			&s.BaseDeliveryFeeMid, &s.BaseDeliveryFeeFar, &s.NearThresholdKm, &s.MidThresholdKm,
			&s.ServiceFee, &s.MinOrderAmount, &s.FreeDeliveryThreshold, &s.WhatsappNumber, &s.Description, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(hoursJSON, &s.OperatingHours)
		out[s.BranchID] = s
	}
	return out, rows.Err()
}
