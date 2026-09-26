package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID  `json:"id"`
	Phone        string     `json:"phone"` // strictly +628...
	Name         string     `json:"name"`
	Role         string     `json:"role"` // customer, branch_admin, owner
	BranchID     *uuid.UUID `json:"branch_id,omitempty"`
	PasswordHash *string    `json:"-"`
	Email        *string    `json:"email,omitempty"`
	Username     *string    `json:"username,omitempty"`
	Address      *string    `json:"address,omitempty"`
	Latitude     *float64   `json:"latitude,omitempty"`
	Longitude    *float64   `json:"longitude,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type Branch struct {
	ID            uuid.UUID `json:"id"`
	Slug          string    `json:"slug"`
	Name          string    `json:"name"`
	Address       string    `json:"address"`
	Phone         string    `json:"phone"`
	Latitude      float64   `json:"latitude"`
	Longitude     float64   `json:"longitude"`
	GradientTheme string    `json:"gradient_theme"` // brand, emerald, indigo
	Icon          string    `json:"icon"`
	FacilityTags  []string  `json:"facility_tags"`
	Rating        float64   `json:"rating"`
	IsOpen        bool      `json:"is_open"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Derived per-request from branch_settings.operating_hours. IsOpen is the
	// manual master switch; IsOpenNow also accounts for the clock, and is what
	// the storefront should trust.
	IsOpenNow          bool   `json:"is_open_now"`
	TodayHours         string `json:"today_hours,omitempty"`
	WhatsappNumber     string `json:"whatsapp_number,omitempty"`
	HalalCertificateID string `json:"halal_certificate_id,omitempty"`
}

type BranchSettings struct {
	ID                    uuid.UUID      `json:"id"`
	BranchID              uuid.UUID      `json:"branch_id"`
	OperatingHours        map[string]any `json:"operating_hours"`
	MaxDeliveryRadiusKm   int            `json:"max_delivery_radius_km"`
	BaseDeliveryFeeNear   int            `json:"base_delivery_fee_near"`
	BaseDeliveryFeeMid    int            `json:"base_delivery_fee_mid"`
	BaseDeliveryFeeFar    int            `json:"base_delivery_fee_far"`
	NearThresholdKm       int            `json:"near_threshold_km"`
	MidThresholdKm        int            `json:"mid_threshold_km"`
	ServiceFee            int            `json:"service_fee"`
	MinOrderAmount        int            `json:"min_order_amount"`
	FreeDeliveryThreshold int            `json:"free_delivery_threshold"`
	WhatsappNumber        string         `json:"whatsapp_number"`
	Description           *string        `json:"description,omitempty"`
	HalalCertificateID    string         `json:"halal_certificate_id"`
	UpdatedAt             time.Time      `json:"updated_at"`
}

type Category struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Emoji     string    `json:"emoji"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

type MenuItem struct {
	ID          uuid.UUID `json:"id"`
	BranchID    uuid.UUID `json:"branch_id"`
	CategoryID  uuid.UUID `json:"category_id"`
	Category    *Category `json:"category,omitempty"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       int       `json:"price"` // Rupiah
	Icon        string    `json:"icon"`
	IconBgClass string    `json:"icon_bg_class"`
	ImageURL    *string   `json:"image_url,omitempty"`
	Tag         *string   `json:"tag,omitempty"`
	IsAvailable bool      `json:"is_available"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Order struct {
	ID                 uuid.UUID      `json:"id"`
	OrderNumber        string         `json:"order_number"` // BRANCH-YYYYMMDD-XXXX
	UserID             *uuid.UUID     `json:"user_id,omitempty"`
	BranchID           uuid.UUID      `json:"branch_id"`
	Branch             *Branch        `json:"branch,omitempty"`
	OrderType          string         `json:"order_type"` // delivery, pickup, scheduled
	Status             string         `json:"status"`
	CustomerName       string         `json:"customer_name"`
	CustomerPhone      string         `json:"customer_phone"` // +628...
	DeliveryAddress    *string        `json:"delivery_address,omitempty"`
	DeliveryNotes      *string        `json:"delivery_notes,omitempty"`
	DeliveryLat        *float64       `json:"delivery_lat,omitempty"`
	DeliveryLon        *float64       `json:"delivery_lon,omitempty"`
	DeliveryDistanceKm float64        `json:"delivery_distance_km"`
	Subtotal           int            `json:"subtotal"`
	DeliveryFee        int            `json:"delivery_fee"`
	ServiceFee         int            `json:"service_fee"`
	Discount           int            `json:"discount"`
	GrandTotal         int            `json:"grand_total"`
	PromoCode          *string        `json:"promo_code,omitempty"`
	ScheduledAt        *time.Time     `json:"scheduled_at,omitempty"`
	AcknowledgedAt     *time.Time     `json:"acknowledged_at,omitempty"`
	RejectionReason    *string        `json:"rejection_reason,omitempty"`
	Version            int            `json:"version"`
	Items              []OrderItem    `json:"items,omitempty"`
	Payment            *Payment       `json:"payment,omitempty"`
	Feedback           *OrderFeedback `json:"feedback,omitempty"`
	IsFirstOrder       bool           `json:"is_first_order"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}

type OrderItemFeedback struct {
	OrderItemID uuid.UUID  `json:"order_item_id"`
	MenuItemID  *uuid.UUID `json:"menu_item_id,omitempty"`
	ItemName    string     `json:"item_name"`
	Rating      int        `json:"rating"`
	Reason      *string    `json:"reason,omitempty"`
}

// OrderFeedback is submitted by the customer after a completed order. It is
// intentionally separate from a public branch rating: staff can use it to
// improve service without exposing an unmoderated review feed.
type OrderFeedback struct {
	Rating        int                 `json:"rating"`
	RestoRating   *int                `json:"resto_rating,omitempty"`
	AppRating     *int                `json:"app_rating,omitempty"`
	RestoReason   *string             `json:"resto_reason,omitempty"`
	AppReason     *string             `json:"app_reason,omitempty"`
	Comment       *string             `json:"comment,omitempty"`
	ItemsFeedback []OrderItemFeedback `json:"items_feedback,omitempty"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
}

type OrderFeedbackAdminItem struct {
	OrderID       uuid.UUID           `json:"order_id"`
	OrderNumber   string              `json:"order_number"`
	CustomerName  string              `json:"customer_name"`
	CustomerPhone string              `json:"customer_phone"`
	BranchID      uuid.UUID           `json:"branch_id"`
	BranchName    string              `json:"branch_name"`
	BranchSlug    string              `json:"branch_slug"`
	Rating        int                 `json:"rating"`
	RestoRating   *int                `json:"resto_rating,omitempty"`
	AppRating     *int                `json:"app_rating,omitempty"`
	RestoReason   *string             `json:"resto_reason,omitempty"`
	AppReason     *string             `json:"app_reason,omitempty"`
	Comment       *string             `json:"comment,omitempty"`
	ItemsFeedback []OrderItemFeedback `json:"items_feedback"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
}

type OrderItem struct {
	ID           uuid.UUID  `json:"id"`
	OrderID      uuid.UUID  `json:"order_id"`
	MenuItemID   *uuid.UUID `json:"menu_item_id,omitempty"`
	ItemName     string     `json:"item_name"`
	ItemPrice    int        `json:"item_price"`
	ItemIcon     string     `json:"item_icon"`
	Quantity     int        `json:"quantity"`
	Notes        *string    `json:"notes,omitempty"`
	LineTotal    int        `json:"line_total"`
	Rating       *int       `json:"rating,omitempty"`
	ReviewReason *string    `json:"review_reason,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

type Payment struct {
	ID                     uuid.UUID      `json:"id"`
	OrderID                uuid.UUID      `json:"order_id"`
	MidtransOrderID        string         `json:"midtrans_order_id"`
	PaymentMethod          string         `json:"payment_method"` // qris, gopay, shopeepay
	PaymentType            string         `json:"payment_type"`
	Status                 string         `json:"status"` // pending, settlement, expire, cancel
	Amount                 int            `json:"amount"`
	IdempotencyKey         string         `json:"idempotency_key"`
	SnapToken              *string        `json:"snap_token,omitempty"`
	SnapRedirectURL        *string        `json:"snap_redirect_url,omitempty"`
	QRString               *string        `json:"qr_string,omitempty"`
	MidtransTransactionID  *string        `json:"midtrans_transaction_id,omitempty"`
	MidtransResponse       map[string]any `json:"midtrans_response,omitempty"`
	PaidAt                 *time.Time     `json:"paid_at,omitempty"`
	ExpiresAt              time.Time      `json:"expires_at"`
	RefundAmount           int            `json:"refund_amount"`
	RefundReason           *string        `json:"refund_reason,omitempty"`
	RefundedAt             *time.Time     `json:"refunded_at,omitempty"`
	RefundedBy             *uuid.UUID     `json:"refunded_by,omitempty"`
	MidtransRefundResponse map[string]any `json:"midtrans_refund_response,omitempty"`
	CreatedAt              time.Time      `json:"created_at"`
	UpdatedAt              time.Time      `json:"updated_at"`
}

// Refundable reports whether this payment can still be sent to Midtrans for a
// refund: it must have actually settled, and not have been refunded already.
func (p *Payment) Refundable() bool {
	return p != nil && p.Status == "settlement"
}

// RemainingRefundable is the amount still eligible for refund (the original
// charge minus whatever has already been refunded).
func (p *Payment) RemainingRefundable() int {
	if p == nil {
		return 0
	}
	remaining := p.Amount - p.RefundAmount
	if remaining < 0 {
		return 0
	}
	return remaining
}

type Promo struct {
	ID              uuid.UUID  `json:"id"`
	Code            string     `json:"code"`
	Type            string     `json:"type"` // fixed, free_delivery
	DiscountAmount  int        `json:"discount_amount"`
	IsActive        bool       `json:"is_active"`
	MinSpend        int        `json:"min_spend"`
	ValidFrom       *time.Time `json:"valid_from,omitempty"`
	ValidUntil      *time.Time `json:"valid_until,omitempty"`
	MaxRedemptions  *int       `json:"max_redemptions,omitempty"`
	RedemptionCount int        `json:"redemption_count"`
	CreatedAt       time.Time  `json:"created_at"`
}

// Redeemable reports whether the promo may still be applied at the given time.
func (p *Promo) Redeemable(at time.Time) bool {
	if !p.IsActive {
		return false
	}
	if p.ValidFrom != nil && at.Before(*p.ValidFrom) {
		return false
	}
	if p.ValidUntil != nil && at.After(*p.ValidUntil) {
		return false
	}
	if p.MaxRedemptions != nil && p.RedemptionCount >= *p.MaxRedemptions {
		return false
	}
	return true
}
