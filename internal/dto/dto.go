package dto

import (
	"regexp"
	"strings"
	"time"

	"github.com/ericanthonywu/maremereso-olga/backend/internal/apperror"
	"github.com/google/uuid"
)

// NormalizeIndonesianPhone transforms 0812..., 62812..., +62812... into canonical "+62812..."
func NormalizeIndonesianPhone(phone string) (string, error) {
	cleaned := strings.TrimSpace(phone)
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")

	if strings.HasPrefix(cleaned, "+62") {
		// e.g. +628123456789
		cleaned = cleaned[3:]
	} else if strings.HasPrefix(cleaned, "62") {
		cleaned = cleaned[2:]
	} else if strings.HasPrefix(cleaned, "0") {
		cleaned = cleaned[1:]
	}

	// Now it must start with 8 and have length between 8 and 13 digits (total 9-14 with 8)
	match, _ := regexp.MatchString(`^8[1-9][0-9]{7,11}$`, cleaned)
	if !match {
		return "", apperror.ErrInvalidPhone
	}

	return "+62" + cleaned, nil
}

// Request DTOs
type CustomerPhoneLoginRequest struct {
	Phone string `json:"phone" validate:"required"`
	Name  string `json:"name"`
}

type AdminLoginRequest struct {
	Identifier string `json:"identifier" validate:"required"` // email or phone
	Password   string `json:"password" validate:"required"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  UserSummary `json:"user"`
}

type UserSummary struct {
	ID       uuid.UUID  `json:"id"`
	Name     string     `json:"name"`
	Phone    string     `json:"phone"`
	Role     string     `json:"role"`
	BranchID *uuid.UUID `json:"branch_id,omitempty"`
}

type CreateOrderItemRequest struct {
	MenuItemID uuid.UUID `json:"menu_item_id" validate:"required"`
	Quantity   int       `json:"quantity" validate:"required,min=1"`
	Notes      string    `json:"notes"`
}

type CreateOrderRequest struct {
	BranchID        uuid.UUID                `json:"branch_id" validate:"required"`
	OrderType       string                   `json:"order_type" validate:"required,oneof=delivery pickup scheduled"`
	CustomerName    string                   `json:"customer_name" validate:"required"`
	CustomerPhone   string                   `json:"customer_phone" validate:"required"`
	DeliveryAddress string                   `json:"delivery_address"`
	DeliveryNotes   string                   `json:"delivery_notes"`
	DeliveryLat     *float64                 `json:"delivery_lat"`
	DeliveryLon     *float64                 `json:"delivery_lon"`
	PromoCode       string                   `json:"promo_code"`
	ScheduledAt     *time.Time               `json:"scheduled_at"`
	Items           []CreateOrderItemRequest `json:"items" validate:"required,min=1"`
}

type UpdateOrderStatusRequest struct {
	Status          string `json:"status" validate:"required"`
	RejectionReason string `json:"rejection_reason"`
	ExpectedVersion int    `json:"expected_version"`
}

type CreatePaymentRequest struct {
	OrderID        uuid.UUID `json:"order_id" validate:"required"`
	PaymentMethod  string    `json:"payment_method" validate:"required,oneof=qris gopay shopeepay"`
	IdempotencyKey string    `json:"idempotency_key" validate:"required"`
}

type ValidatePromoRequest struct {
	Code        string `json:"code" validate:"required"`
	BranchID    string `json:"branch_id"`
	Subtotal    int    `json:"subtotal" validate:"required"`
	DeliveryFee int    `json:"delivery_fee"`
}

type ValidatePromoResponse struct {
	Code           string `json:"code"`
	DiscountAmount int    `json:"discount_amount"`
	FinalTotal     int    `json:"final_total"`
	Message        string `json:"message"`
}

type CreateMenuItemRequest struct {
	BranchID    uuid.UUID `json:"branch_id" validate:"required"`
	CategoryID  uuid.UUID `json:"category_id" validate:"required"`
	Name        string    `json:"name" validate:"required"`
	Description string    `json:"description"`
	Price       int       `json:"price" validate:"required,min=1000"`
	Icon        string    `json:"icon"`
	IconBgClass string    `json:"icon_bg_class"`
	Tag         string    `json:"tag"`
	IsAvailable bool      `json:"is_available"`
}

type UpdateMenuItemRequest struct {
	CategoryID  *uuid.UUID `json:"category_id"`
	Name        *string    `json:"name"`
	Description *string    `json:"description"`
	Price       *int       `json:"price"`
	Icon        *string    `json:"icon"`
	IconBgClass *string    `json:"icon_bg_class"`
	Tag         *string    `json:"tag"`
	IsAvailable *bool      `json:"is_available"`
}

type ToggleAvailabilityRequest struct {
	IsAvailable bool `json:"is_available"`
}

type UpdateBranchSettingsRequest struct {
	OperatingHours         map[string]any `json:"operating_hours"`
	MaxDeliveryRadiusKm    int            `json:"max_delivery_radius_km"`
	BaseDeliveryFeeNear    int            `json:"base_delivery_fee_near"`
	BaseDeliveryFeeMid     int            `json:"base_delivery_fee_mid"`
	BaseDeliveryFeeFar     int            `json:"base_delivery_fee_far"`
	NearThresholdKm        int            `json:"near_threshold_km"`
	MidThresholdKm         int            `json:"mid_threshold_km"`
	MinOrderAmount         int            `json:"min_order_amount"`
	FreeDeliveryThreshold  int            `json:"free_delivery_threshold"`
	WhatsappNumber         string         `json:"whatsapp_number"`
	Description            string         `json:"description"`
}

type CommonResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}
