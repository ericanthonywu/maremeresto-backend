package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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

// UpdateCustomerProfile keeps the account identity used for subsequent orders
// in sync with the customer's profile menu. The unique phone constraint is
// deliberately left to PostgreSQL, which safely rejects a number in use.
func (r *Repository) UpdateCustomerProfile(ctx context.Context, userID uuid.UUID, name, phone string) (*model.User, error) {
	query := `
		UPDATE users
		SET name = $1, phone = $2, updated_at = NOW()
		WHERE id = $3 AND role = 'customer'
		RETURNING id, phone, name, role, branch_id, address, latitude, longitude, created_at, updated_at
	`
	var u model.User
	err := r.db.QueryRow(ctx, query, name, phone, userID).Scan(
		&u.ID, &u.Phone, &u.Name, &u.Role, &u.BranchID, &u.Address, &u.Latitude, &u.Longitude, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
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

func (r *Repository) UpdateBranchProfile(ctx context.Context, id uuid.UUID, name, address, phone string, lat, lon float64) error {
	query := `UPDATE branches SET name = $1, address = $2, phone = $3, latitude = $4, longitude = $5, updated_at = NOW() WHERE id = $6`
	_, err := r.db.Exec(ctx, query, name, address, phone, lat, lon, id)
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
func (r *Repository) NextOrderNumber(ctx context.Context, tx pgx.Tx, branchSlug string) (string, error) {
	var seq int
	err := tx.QueryRow(ctx, `SELECT nextval('order_number_seq')`).Scan(&seq)
	if err != nil {
		return "", err
	}
	dateStr := time.Now().Format("20060102")
	prefix := strings.ToUpper(strings.TrimSpace(branchSlug))
	prefix = strings.Map(func(r rune) rune {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, prefix)
	if prefix == "" {
		prefix = "OUTLET"
	}
	return fmt.Sprintf("%s-%s-%04d", prefix, dateStr, seq), nil
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
		       o.promo_code, o.scheduled_at, o.acknowledged_at,
		       o.rejection_reason, o.version, o.created_at, o.updated_at,
		       b.name, b.slug, b.address, b.phone, COALESCE(bs.whatsapp_number, '')
		FROM orders o
		JOIN branches b ON o.branch_id = b.id
		LEFT JOIN branch_settings bs ON bs.branch_id = b.id
		WHERE o.id = $1
	`
	var o model.Order
	var bName, bSlug, bAddr, bPhone, bWhatsApp string
	err := r.db.QueryRow(ctx, query, id).Scan(
		&o.ID, &o.OrderNumber, &o.UserID, &o.BranchID, &o.OrderType, &o.Status,
		&o.CustomerName, &o.CustomerPhone, &o.DeliveryAddress, &o.DeliveryNotes,
		&o.DeliveryLat, &o.DeliveryLon, &o.DeliveryDistanceKm,
		&o.Subtotal, &o.DeliveryFee, &o.ServiceFee, &o.Discount, &o.GrandTotal,
		&o.PromoCode, &o.ScheduledAt, &o.AcknowledgedAt,
		&o.RejectionReason, &o.Version, &o.CreatedAt, &o.UpdatedAt,
		&bName, &bSlug, &bAddr, &bPhone, &bWhatsApp,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	o.Branch = &model.Branch{
		ID:             o.BranchID,
		Name:           bName,
		Slug:           bSlug,
		Address:        bAddr,
		Phone:          bPhone,
		WhatsappNumber: bWhatsApp,
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
	feedback, err := r.FindOrderFeedback(ctx, o.ID)
	if err != nil {
		return nil, err
	}
	o.Feedback = feedback

	return &o, nil
}

func (r *Repository) FindOrderFeedback(ctx context.Context, orderID uuid.UUID) (*model.OrderFeedback, error) {
	query := `SELECT rating, comment, created_at, updated_at FROM order_feedback WHERE order_id = $1`
	var feedback model.OrderFeedback
	err := r.db.QueryRow(ctx, query, orderID).Scan(&feedback.Rating, &feedback.Comment, &feedback.CreatedAt, &feedback.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &feedback, nil
}

func (r *Repository) UpsertOrderFeedback(ctx context.Context, orderID uuid.UUID, rating int, comment string) (*model.OrderFeedback, error) {
	query := `
		INSERT INTO order_feedback (order_id, rating, comment)
		VALUES ($1, $2, NULLIF($3, ''))
		ON CONFLICT (order_id) DO UPDATE
		SET rating = EXCLUDED.rating, comment = EXCLUDED.comment, updated_at = NOW()
		RETURNING rating, comment, created_at, updated_at
	`
	var feedback model.OrderFeedback
	err := r.db.QueryRow(ctx, query, orderID, rating, comment).Scan(
		&feedback.Rating, &feedback.Comment, &feedback.CreatedAt, &feedback.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &feedback, nil
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

// CountOrdersByStatus supplies the badges on the admin status filters. It is
// deliberately independent of the currently selected filter/search term so
// staff can see where work is waiting before switching tabs.
func (r *Repository) CountOrdersByStatus(ctx context.Context, branchID *uuid.UUID) (map[string]int, error) {
	query := `SELECT status, COUNT(*) FROM orders`
	args := []any{}
	if branchID != nil {
		query += ` WHERE branch_id = $1`
		args = append(args, *branchID)
	}
	query += ` GROUP BY status`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		counts[status] = count
	}
	return counts, rows.Err()
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
	} else if f.UserID == nil {
		// Admin orders list: do not include unpaid pending orders unless explicitly asked
		baseQuery += " AND o.status != 'pending'"
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
		       o.promo_code, o.scheduled_at, o.acknowledged_at,
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
			&o.PromoCode, &o.ScheduledAt, &o.AcknowledgedAt,
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
		       paid_at, expires_at, refund_amount, refund_reason, refunded_at, refunded_by,
		       created_at, updated_at
		FROM payments WHERE order_id = $1
	`
	var p model.Payment
	err := r.db.QueryRow(ctx, query, orderID).Scan(
		&p.ID, &p.OrderID, &p.MidtransOrderID, &p.PaymentMethod, &p.PaymentType, &p.Status, &p.Amount,
		&p.IdempotencyKey, &p.SnapToken, &p.SnapRedirectURL, &p.QRString, &p.MidtransTransactionID,
		&p.PaidAt, &p.ExpiresAt, &p.RefundAmount, &p.RefundReason, &p.RefundedAt, &p.RefundedBy,
		&p.CreatedAt, &p.UpdatedAt,
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

// RecordRefund persists a confirmed Midtrans refund on the payment row.
// refundedBy is nil-able so a system-initiated refund can still be recorded.
func (r *Repository) RecordRefund(ctx context.Context, tx pgx.Tx, orderID uuid.UUID, amount int, reason string, refundedBy *uuid.UUID, responseJSON map[string]any) error {
	jsonBytes, _ := json.Marshal(responseJSON)
	query := `
		UPDATE payments
		SET status = 'refund', refund_amount = refund_amount + $1, refund_reason = $2,
		    refunded_at = NOW(), refunded_by = $3, midtrans_refund_response = $4, updated_at = NOW()
		WHERE order_id = $5
	`
	res, err := tx.Exec(ctx, query, amount, reason, refundedBy, jsonBytes, orderID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}
	return nil
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
		       service_fee, min_order_amount, free_delivery_threshold, whatsapp_number, description,
		       halal_certificate_id, updated_at
		FROM branch_settings WHERE branch_id = $1
	`
	var s model.BranchSettings
	var hoursJSON []byte
	err := r.db.QueryRow(ctx, query, branchID).Scan(
		&s.ID, &s.BranchID, &hoursJSON, &s.MaxDeliveryRadiusKm, &s.BaseDeliveryFeeNear,
		&s.BaseDeliveryFeeMid, &s.BaseDeliveryFeeFar, &s.NearThresholdKm, &s.MidThresholdKm,
		&s.ServiceFee, &s.MinOrderAmount, &s.FreeDeliveryThreshold, &s.WhatsappNumber, &s.Description,
		&s.HalalCertificateID, &s.UpdatedAt,
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
		    whatsapp_number = $11, description = $12, halal_certificate_id = $13, updated_at = NOW()
		WHERE branch_id = $14
	`
	res, err := r.db.Exec(ctx, query,
		hoursBytes, s.MaxDeliveryRadiusKm, s.BaseDeliveryFeeNear, s.BaseDeliveryFeeMid, s.BaseDeliveryFeeFar,
		s.NearThresholdKm, s.MidThresholdKm, s.ServiceFee, s.MinOrderAmount, s.FreeDeliveryThreshold,
		s.WhatsappNumber, s.Description, s.HalalCertificateID, s.BranchID,
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
// Order lifecycle extras: staff acknowledgement
// ---------------------------------------------------------------------

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
		  AND status NOT IN ('pending', 'cancelled', 'rejected', 'completed')
		  AND ($1::uuid IS NULL OR branch_id = $1)
	`
	var count int
	err := r.db.QueryRow(ctx, query, branchID).Scan(&count)
	return count, err
}

// IncrementPromoRedemption is called once an order that used a promo is
// committed, so max_redemptions is actually enforced.
func (r *Repository) IncrementPromoRedemption(ctx context.Context, tx pgx.Tx, code string) error {
	cmd, err := tx.Exec(ctx, `
		UPDATE promos
		SET redemption_count = redemption_count + 1
		WHERE UPPER(code) = UPPER($1)
		  AND is_active = true
		  AND (max_redemptions IS NULL OR redemption_count < max_redemptions)
		  AND (valid_from IS NULL OR NOW() >= valid_from)
		  AND (valid_until IS NULL OR NOW() <= valid_until)
	`, code)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return apperror.ErrInvalidPromo
	}
	return nil
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

func (r *Repository) FindCategoryByID(ctx context.Context, id uuid.UUID) (*model.Category, error) {
	query := `SELECT id, name, slug, emoji, sort_order, created_at FROM categories WHERE id = $1`
	var c model.Category
	err := r.db.QueryRow(ctx, query, id).Scan(&c.ID, &c.Name, &c.Slug, &c.Emoji, &c.SortOrder, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *Repository) CategorySlugExists(ctx context.Context, slug string, excludeID *uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM categories WHERE slug = $1 AND ($2::uuid IS NULL OR id != $2))`
	var exists bool
	err := r.db.QueryRow(ctx, query, slug, excludeID).Scan(&exists)
	return exists, err
}

func (r *Repository) CountMenuItemsByCategoryID(ctx context.Context, categoryID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM menu_items WHERE category_id = $1`, categoryID).Scan(&count)
	return count, err
}

func (r *Repository) CreateCategory(ctx context.Context, c *model.Category) error {
	query := `INSERT INTO categories (id, name, slug, emoji, sort_order, created_at)
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	return r.db.QueryRow(ctx, query, c.ID, c.Name, c.Slug, c.Emoji, c.SortOrder, c.CreatedAt).Scan(&c.ID)
}

func (r *Repository) UpdateCategory(ctx context.Context, c *model.Category) error {
	query := `UPDATE categories SET name = $1, emoji = $2, sort_order = $3 WHERE id = $4`
	cmd, err := r.db.Exec(ctx, query, c.Name, c.Emoji, c.SortOrder, c.ID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}
	return nil
}

func (r *Repository) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	cmd, err := r.db.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}
	return nil
}

// ListSettingsByBranch returns every branch's settings keyed by branch id, so
// listing branches costs one query instead of one per branch.
func (r *Repository) ListSettingsByBranch(ctx context.Context) (map[uuid.UUID]model.BranchSettings, error) {
	query := `
		SELECT id, branch_id, operating_hours, max_delivery_radius_km, base_delivery_fee_near,
		       base_delivery_fee_mid, base_delivery_fee_far, near_threshold_km, mid_threshold_km,
		       service_fee, min_order_amount, free_delivery_threshold, whatsapp_number, description,
		       halal_certificate_id, updated_at
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
			&s.ServiceFee, &s.MinOrderAmount, &s.FreeDeliveryThreshold, &s.WhatsappNumber, &s.Description,
			&s.HalalCertificateID, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(hoursJSON, &s.OperatingHours)
		out[s.BranchID] = s
	}
	return out, rows.Err()
}
