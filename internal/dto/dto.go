package dto

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ericanthonywu/maremereso-olga/backend/internal/apperror"
	"github.com/ericanthonywu/maremereso-olga/backend/internal/model"
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
	Name  string `json:"name" validate:"required"`
}

func (r *CustomerPhoneLoginRequest) Validate() error {
	r.Name = strings.TrimSpace(r.Name)
	if len(r.Name) < 2 || len(r.Name) > 100 {
		return apperror.Invalid("nama pemesan harus 2-100 karakter")
	}
	return nil
}

type UpdateCustomerProfileRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

func (r *UpdateCustomerProfileRequest) Validate() error {
	r.Name = strings.TrimSpace(r.Name)
	if len(r.Name) < 2 || len(r.Name) > 100 {
		return apperror.Invalid("nama pemesan harus 2-100 karakter")
	}
	normalized, err := NormalizeIndonesianPhone(r.Phone)
	if err != nil {
		return err
	}
	r.Phone = normalized
	return nil
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
	Email    *string    `json:"email,omitempty"`
	Username *string    `json:"username,omitempty"`
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

// Validate enforces the shape of a checkout submission. Quantities are capped
// so a single request cannot be used to mint an enormous order.
func (r *CreateOrderRequest) Validate() error {
	r.CustomerName = strings.TrimSpace(r.CustomerName)
	r.DeliveryAddress = strings.TrimSpace(r.DeliveryAddress)
	r.DeliveryNotes = strings.TrimSpace(r.DeliveryNotes)
	r.PromoCode = strings.ToUpper(strings.TrimSpace(r.PromoCode))

	if r.BranchID == uuid.Nil {
		return apperror.Invalid("outlet wajib dipilih")
	}
	switch r.OrderType {
	case "delivery", "pickup", "scheduled":
	default:
		return apperror.Invalid("metode penerimaan tidak valid")
	}
	if len(r.CustomerName) < 2 || len(r.CustomerName) > 100 {
		return apperror.Invalid("nama pemesan harus 2-100 karakter")
	}
	if len(r.DeliveryAddress) > 500 {
		return apperror.Invalid("alamat maksimal 500 karakter")
	}
	if len(r.DeliveryNotes) > 300 {
		return apperror.Invalid("catatan maksimal 300 karakter")
	}
	if (r.OrderType == "delivery" || r.OrderType == "scheduled") && r.DeliveryNotes == "" {
		return apperror.Invalid("catatan untuk kurir wajib diisi")
	}
	if len(r.PromoCode) > 50 {
		return apperror.Invalid("kode promo tidak valid")
	}
	if len(r.Items) == 0 {
		return apperror.Invalid("keranjang pesanan kosong")
	}
	if len(r.Items) > 100 {
		return apperror.Invalid("terlalu banyak jenis item dalam satu pesanan")
	}

	totalQty := 0
	for i := range r.Items {
		it := &r.Items[i]
		it.Notes = strings.TrimSpace(it.Notes)
		if it.MenuItemID == uuid.Nil {
			return apperror.Invalid("item pesanan tidak valid")
		}
		if it.Quantity < 1 || it.Quantity > 99 {
			return apperror.Invalid("jumlah setiap item harus antara 1 dan 99")
		}
		if len(it.Notes) > 200 {
			return apperror.Invalid("catatan item maksimal 200 karakter")
		}
		totalQty += it.Quantity
	}
	if totalQty > 300 {
		return apperror.Invalid("total item dalam satu pesanan terlalu banyak")
	}

	if r.OrderType == "scheduled" {
		if r.ScheduledAt == nil {
			return apperror.Invalid("waktu penjadwalan wajib diisi")
		}
		if r.ScheduledAt.Before(time.Now().Add(-5 * time.Minute)) {
			return apperror.Invalid("waktu penjadwalan tidak boleh di masa lalu")
		}
		if r.ScheduledAt.After(time.Now().Add(7 * 24 * time.Hour)) {
			return apperror.Invalid("penjadwalan maksimal 7 hari ke depan")
		}
	} else {
		// A time picked before the customer switched away from "scheduled"
		// must not linger on the order.
		r.ScheduledAt = nil
	}

	if r.DeliveryLat != nil && (*r.DeliveryLat < -90 || *r.DeliveryLat > 90) {
		return apperror.Invalid("koordinat lokasi tidak valid")
	}
	if r.DeliveryLon != nil && (*r.DeliveryLon < -180 || *r.DeliveryLon > 180) {
		return apperror.Invalid("koordinat lokasi tidak valid")
	}

	return nil
}

type UpdateOrderStatusRequest struct {
	Status          string `json:"status" validate:"required"`
	RejectionReason string `json:"rejection_reason"`
	ExpectedVersion int    `json:"expected_version"`
}

type OrderItemFeedbackDTO struct {
	OrderItemID uuid.UUID  `json:"order_item_id"`
	MenuItemID  *uuid.UUID `json:"menu_item_id,omitempty"`
	ItemName    string     `json:"item_name"`
	Rating      int        `json:"rating"`
	Reason      string     `json:"reason"`
}

type OrderFeedbackRequest struct {
	Rating        int                    `json:"rating"`
	RestoRating   *int                   `json:"resto_rating,omitempty"`
	AppRating     *int                   `json:"app_rating,omitempty"`
	RestoReason   string                 `json:"resto_reason,omitempty"`
	AppReason     string                 `json:"app_reason,omitempty"`
	Comment       string                 `json:"comment,omitempty"`
	ItemsFeedback []OrderItemFeedbackDTO `json:"items_feedback,omitempty"`
}

func (r *OrderFeedbackRequest) Validate() error {
	r.RestoReason = strings.TrimSpace(r.RestoReason)
	r.AppReason = strings.TrimSpace(r.AppReason)
	r.Comment = strings.TrimSpace(r.Comment)

	if len(r.RestoReason) > 500 {
		return apperror.Invalid("alasan resto maksimal 500 karakter")
	}
	if len(r.AppReason) > 500 {
		return apperror.Invalid("alasan aplikasi maksimal 500 karakter")
	}
	if len(r.Comment) > 500 {
		return apperror.Invalid("ulasan maksimal 500 karakter")
	}

	if r.RestoRating != nil && (*r.RestoRating < 1 || *r.RestoRating > 5) {
		return apperror.Invalid("rating resto harus antara 1 dan 5")
	}
	if r.AppRating != nil && (*r.AppRating < 1 || *r.AppRating > 5) {
		return apperror.Invalid("rating aplikasi harus antara 1 dan 5")
	}

	for _, it := range r.ItemsFeedback {
		if it.Rating < 1 || it.Rating > 5 {
			name := it.ItemName
			if name == "" {
				name = "menu"
			}
			return apperror.Invalid(fmt.Sprintf("rating untuk %s harus antara 1 dan 5", name))
		}
		if len(strings.TrimSpace(it.Reason)) > 500 {
			name := it.ItemName
			if name == "" {
				name = "menu"
			}
			return apperror.Invalid(fmt.Sprintf("alasan untuk %s maksimal 500 karakter", name))
		}
	}

	// If overall Rating is not explicitly set between 1 and 5, derive it intelligently
	if r.Rating < 1 || r.Rating > 5 {
		if r.RestoRating != nil && *r.RestoRating >= 1 && *r.RestoRating <= 5 {
			r.Rating = *r.RestoRating
		} else if r.AppRating != nil && *r.AppRating >= 1 && *r.AppRating <= 5 {
			r.Rating = *r.AppRating
		} else if len(r.ItemsFeedback) > 0 {
			total := 0
			count := 0
			for _, it := range r.ItemsFeedback {
				if it.Rating >= 1 && it.Rating <= 5 {
					total += it.Rating
					count++
				}
			}
			if count > 0 {
				r.Rating = (total + count/2) / count
			}
		}
	}

	if r.Rating < 1 || r.Rating > 5 {
		return apperror.Invalid("pilih rating minimal salah satu (resto, aplikasi, atau menu)")
	}

	return nil
}

type RatingCount struct {
	Rating     int     `json:"rating"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

type BranchRatingSummary struct {
	BranchID         uuid.UUID `json:"branch_id"`
	BranchName       string    `json:"branch_name"`
	BranchSlug       string    `json:"branch_slug"`
	TotalReviews     int       `json:"total_reviews"`
	AvgOverallRating float64   `json:"avg_overall_rating"`
	AvgRestoRating   float64   `json:"avg_resto_rating"`
	AvgAppRating     float64   `json:"avg_app_rating"`
}

type MenuItemRatingSummary struct {
	ItemName      string   `json:"item_name"`
	TotalReviews  int      `json:"total_reviews"`
	AvgRating     float64  `json:"avg_rating"`
	PositiveCount int      `json:"positive_count"`
	NegativeCount int      `json:"negative_count"`
	SampleReasons []string `json:"sample_reasons"`
}

type CommonTagCount struct {
	Tag        string `json:"tag"`
	Count      int    `json:"count"`
	IsPositive bool   `json:"is_positive"`
}

type FeedbackAnalytics struct {
	TotalReviews        int                     `json:"total_reviews"`
	AvgOverallRating    float64                 `json:"avg_overall_rating"`
	AvgRestoRating      float64                 `json:"avg_resto_rating"`
	AvgAppRating        float64                 `json:"avg_app_rating"`
	SatisfactionRate    float64                 `json:"satisfaction_rate"` // % rating >= 4
	PositiveCount       int                     `json:"positive_count"`
	ConstructiveCount   int                     `json:"constructive_count"`
	RatingBreakdown     []RatingCount           `json:"rating_breakdown"`
	BranchSummaries     []BranchRatingSummary   `json:"branch_summaries"`
	TopMenuItems        []MenuItemRatingSummary `json:"top_menu_items"`
	NeedsAttentionItems []MenuItemRatingSummary `json:"needs_attention_items"`
	CommonTags          []CommonTagCount        `json:"common_tags"`
}

type FeedbackListResponse struct {
	Items      []model.OrderFeedbackAdminItem `json:"items"`
	Total      int                            `json:"total"`
	Page       int                            `json:"page"`
	Limit      int                            `json:"limit"`
	TotalPages int                            `json:"total_pages"`
}

type FeedbackAISummaryResponse struct {
	Configured            bool     `json:"configured"`
	Model                 string   `json:"model"`
	ExecutiveSummary      string   `json:"executive_summary"`
	DetailSummary         string   `json:"detail_summary"`
	ActionableSuggestions []string `json:"actionable_suggestions"`
	RawAnalysis           string   `json:"raw_analysis,omitempty"`
	GeneratedAt           string   `json:"generated_at"`
	Error                 string   `json:"error,omitempty"`
}

type CreatePaymentRequest struct {
	OrderID        uuid.UUID `json:"order_id" validate:"required"`
	PaymentMethod  string    `json:"payment_method"`
	IdempotencyKey string    `json:"idempotency_key" validate:"required"`
}

// RefundOrderRequest is issued by staff against a paid order. Amount is
// optional: zero means "refund whatever is still outstanding".
type RefundOrderRequest struct {
	Amount          int    `json:"amount"`
	Reason          string `json:"reason" validate:"required"`
	ExpectedVersion int    `json:"expected_version"`
}

func (r *RefundOrderRequest) Validate() error {
	r.Reason = strings.TrimSpace(r.Reason)
	if r.Reason == "" {
		return apperror.Invalid("alasan refund wajib diisi")
	}
	if len(r.Reason) > 500 {
		return apperror.Invalid("alasan refund maksimal 500 karakter")
	}
	if r.Amount < 0 {
		return apperror.Invalid("jumlah refund tidak boleh negatif")
	}
	return nil
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
	BranchID    uuid.UUID `json:"branch_id"`
	CategoryID  uuid.UUID `json:"category_id" validate:"required"`
	Name        string    `json:"name" validate:"required"`
	Description string    `json:"description"`
	Price       int       `json:"price" validate:"required,min=1000"`
	Icon        string    `json:"icon"`
	IconBgClass string    `json:"icon_bg_class"`
	ImageURL    string    `json:"image_url"`
	Tag         string    `json:"tag"`
	IsAvailable bool      `json:"is_available"`
	SortOrder   int       `json:"sort_order"`
}

type CreateMenuItemsBulkRequest struct {
	Items []CreateMenuItemRequest `json:"items"`
}

func (r *CreateMenuItemsBulkRequest) Validate() error {
	if len(r.Items) == 0 {
		return apperror.Invalid("minimal satu menu wajib diisi")
	}
	if len(r.Items) > 100 {
		return apperror.Invalid("maksimal 100 menu dalam sekali tambah")
	}
	branchID := r.Items[0].BranchID
	for i := range r.Items {
		if r.Items[i].BranchID != branchID {
			return apperror.Invalid("semua menu bulk harus untuk outlet yang sama")
		}
		if err := r.Items[i].Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Validate enforces the rules the admin form relies on. Every field is
// re-checked here because the client is untrusted.
func (r *CreateMenuItemRequest) Validate() error {
	r.Name = strings.TrimSpace(r.Name)
	r.Description = strings.TrimSpace(r.Description)
	r.Tag = strings.TrimSpace(r.Tag)

	if r.CategoryID == uuid.Nil {
		return apperror.Invalid("kategori wajib dipilih")
	}
	if len(r.Name) < 2 || len(r.Name) > 100 {
		return apperror.Invalid("nama menu harus 2-100 karakter")
	}
	if len(r.Description) > 500 {
		return apperror.Invalid("deskripsi maksimal 500 karakter")
	}
	if r.Price < 1000 || r.Price > 10000000 {
		return apperror.Invalid("harga harus antara Rp 1.000 dan Rp 10.000.000")
	}
	if len(r.Tag) > 50 {
		return apperror.Invalid("label maksimal 50 karakter")
	}
	if r.SortOrder < 0 || r.SortOrder > 9999 {
		return apperror.Invalid("urutan harus antara 0 dan 9999")
	}
	return nil
}

type UpdateMenuItemRequest struct {
	CategoryID  *uuid.UUID `json:"category_id"`
	Name        *string    `json:"name"`
	Description *string    `json:"description"`
	Price       *int       `json:"price"`
	Icon        *string    `json:"icon"`
	IconBgClass *string    `json:"icon_bg_class"`
	ImageURL    *string    `json:"image_url"`
	Tag         *string    `json:"tag"`
	IsAvailable *bool      `json:"is_available"`
	SortOrder   *int       `json:"sort_order"`
}

// Validate checks only the fields actually supplied, so a partial update
// stays partial.
func (r *UpdateMenuItemRequest) Validate() error {
	if r.CategoryID != nil && *r.CategoryID == uuid.Nil {
		return apperror.Invalid("kategori tidak valid")
	}
	if r.Name != nil {
		*r.Name = strings.TrimSpace(*r.Name)
		if len(*r.Name) < 2 || len(*r.Name) > 100 {
			return apperror.Invalid("nama menu harus 2-100 karakter")
		}
	}
	if r.Description != nil {
		*r.Description = strings.TrimSpace(*r.Description)
		if len(*r.Description) > 500 {
			return apperror.Invalid("deskripsi maksimal 500 karakter")
		}
	}
	if r.Price != nil && (*r.Price < 1000 || *r.Price > 10000000) {
		return apperror.Invalid("harga harus antara Rp 1.000 dan Rp 10.000.000")
	}
	if r.Tag != nil {
		*r.Tag = strings.TrimSpace(*r.Tag)
		if len(*r.Tag) > 50 {
			return apperror.Invalid("label maksimal 50 karakter")
		}
	}
	if r.SortOrder != nil && (*r.SortOrder < 0 || *r.SortOrder > 9999) {
		return apperror.Invalid("urutan harus antara 0 dan 9999")
	}
	return nil
}

type ToggleAvailabilityRequest struct {
	IsAvailable bool `json:"is_available"`
}

type CreateCategoryRequest struct {
	Name      string `json:"name"`
	Emoji     string `json:"emoji"`
	SortOrder int    `json:"sort_order"`
}

func (r *CreateCategoryRequest) Validate() error {
	r.Name = strings.TrimSpace(r.Name)
	r.Emoji = strings.TrimSpace(r.Emoji)
	if r.Name == "" || len(r.Name) > 100 {
		return apperror.Invalid("nama kategori wajib diisi, maksimal 100 karakter")
	}
	if r.Emoji == "" {
		r.Emoji = "🍽️"
	}
	if len(r.Emoji) > 20 {
		return apperror.Invalid("emoji maksimal 20 karakter")
	}
	if r.SortOrder < 0 {
		r.SortOrder = 0
	}
	return nil
}

type UpdateCategoryRequest struct {
	Name      string `json:"name"`
	Emoji     string `json:"emoji"`
	SortOrder int    `json:"sort_order"`
}

func (r *UpdateCategoryRequest) Validate() error {
	r.Name = strings.TrimSpace(r.Name)
	r.Emoji = strings.TrimSpace(r.Emoji)
	if r.Name == "" || len(r.Name) > 100 {
		return apperror.Invalid("nama kategori wajib diisi, maksimal 100 karakter")
	}
	if r.Emoji == "" {
		r.Emoji = "🍽️"
	}
	if len(r.Emoji) > 20 {
		return apperror.Invalid("emoji maksimal 20 karakter")
	}
	if r.SortOrder < 0 {
		r.SortOrder = 0
	}
	return nil
}

type UpdateBranchSettingsRequest struct {
	ID                    *uuid.UUID     `json:"id,omitempty"`
	BranchID              *uuid.UUID     `json:"branch_id"`
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
	Description           string         `json:"description"`
	HalalCertificateID    string         `json:"halal_certificate_id"`
	UpdatedAt             any            `json:"updated_at,omitempty"`
}

// Validate rejects settings that would make the storefront unusable, e.g. a
// mid tier cheaper than the near tier or a zero delivery radius.
func (r *UpdateBranchSettingsRequest) Validate() error {
	if r.NearThresholdKm < 1 || r.MidThresholdKm <= r.NearThresholdKm {
		return apperror.Invalid("batas jarak menengah harus lebih besar dari batas dekat")
	}
	if r.MaxDeliveryRadiusKm < r.MidThresholdKm {
		return apperror.Invalid("radius maksimal tidak boleh lebih kecil dari batas jarak menengah")
	}
	if r.MaxDeliveryRadiusKm > 50 {
		return apperror.Invalid("radius maksimal tidak boleh lebih dari 50 km")
	}
	for _, fee := range []int{r.BaseDeliveryFeeNear, r.BaseDeliveryFeeMid, r.BaseDeliveryFeeFar, r.ServiceFee} {
		if fee < 0 || fee > 1000000 {
			return apperror.Invalid("tarif harus antara Rp 0 dan Rp 1.000.000")
		}
	}
	if r.BaseDeliveryFeeMid < r.BaseDeliveryFeeNear || r.BaseDeliveryFeeFar < r.BaseDeliveryFeeMid {
		return apperror.Invalid("tarif ongkir harus naik sesuai jarak")
	}
	if r.MinOrderAmount < 0 || r.MinOrderAmount > 10000000 {
		return apperror.Invalid("minimal order tidak valid")
	}
	if r.FreeDeliveryThreshold < 0 || r.FreeDeliveryThreshold > 100000000 {
		return apperror.Invalid("ambang gratis ongkir tidak valid")
	}
	if strings.TrimSpace(r.WhatsappNumber) != "" {
		if len(r.WhatsappNumber) > 30 {
			return apperror.Invalid("nomor WhatsApp outlet maksimal 30 karakter")
		}
		if normalized, err := NormalizeIndonesianPhone(r.WhatsappNumber); err == nil {
			r.WhatsappNumber = normalized
		} else {
			// Allow landlines / international / alternate numbers with 6-20 digits
			cleaned := strings.Map(func(rn rune) rune {
				if rn >= '0' && rn <= '9' {
					return rn
				}
				return -1
			}, r.WhatsappNumber)
			if len(cleaned) < 6 || len(cleaned) > 20 {
				return apperror.Invalid("nomor WhatsApp outlet tidak valid")
			}
			r.WhatsappNumber = strings.TrimSpace(r.WhatsappNumber)
		}
	}
	if len(r.Description) > 500 {
		return apperror.Invalid("deskripsi maksimal 500 karakter")
	}
	r.HalalCertificateID = strings.TrimSpace(r.HalalCertificateID)
	if len(r.HalalCertificateID) > 160 {
		return apperror.Invalid("ID sertifikat halal maksimal 160 karakter")
	}
	for _, key := range []string{"weekday", "weekend"} {
		entry, ok := r.OperatingHours[key].(map[string]any)
		if !ok {
			continue
		}
		for _, field := range []string{"open", "close"} {
			v, _ := entry[field].(string)
			if !clockPattern.MatchString(v) {
				return apperror.Invalid("jam operasional harus berformat HH:MM")
			}
		}
	}
	return nil
}

var clockPattern = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)

// UpdateBranchProfileRequest edits the branch's public-facing identity
// (name/address/contact/coordinates) as distinct from its operational
// settings (hours, fees). Latitude/longitude are typically supplied by the
// same geocoder the customer app uses, so the delivery-fee calculation
// stays consistent with what the address says.
type UpdateBranchProfileRequest struct {
	Name      string  `json:"name"`
	Address   string  `json:"address"`
	Phone     string  `json:"phone"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func (r *UpdateBranchProfileRequest) Validate() error {
	r.Name = strings.TrimSpace(r.Name)
	r.Address = strings.TrimSpace(r.Address)
	r.Phone = strings.TrimSpace(r.Phone)

	if r.Name == "" || len(r.Name) > 100 {
		return apperror.Invalid("nama outlet wajib diisi, maksimal 100 karakter")
	}
	if r.Address == "" || len(r.Address) > 2000 {
		return apperror.Invalid("alamat outlet wajib diisi")
	}
	if r.Phone != "" {
		if len(r.Phone) > 30 {
			return apperror.Invalid("WhatsApp outlet maksimal 30 karakter")
		}
		// Normalise mobile numbers (+628...); allow landline numbers like (021), (0271) as-is
		if normalized, err := NormalizeIndonesianPhone(r.Phone); err == nil {
			r.Phone = normalized
		}
	}
	if r.Latitude < -90 || r.Latitude > 90 || r.Longitude < -180 || r.Longitude > 180 {
		return apperror.Invalid("koordinat outlet tidak valid")
	}
	if r.Latitude == 0 && r.Longitude == 0 {
		return apperror.Invalid("koordinat outlet wajib ditentukan melalui pencarian alamat")
	}
	return nil
}

type CommonResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// ---------------------------------------------------------------------
// Delivery quoting
// ---------------------------------------------------------------------

type DeliveryQuoteRequest struct {
	Lat       float64 `json:"lat"`
	Lon       float64 `json:"lon"`
	Subtotal  int     `json:"subtotal"`
	OrderType string  `json:"order_type"`
}

type BranchDeliveryQuote struct {
	BranchID         uuid.UUID `json:"branch_id"`
	BranchSlug       string    `json:"branch_slug"`
	BranchName       string    `json:"branch_name"`
	DistanceKm       float64   `json:"distance_km"`
	DistanceMeters   int       `json:"distance_meters"`
	DeliveryFee      int       `json:"delivery_fee"`
	ServiceFee       int       `json:"service_fee"`
	MinOrderAmount   int       `json:"min_order_amount"`
	EtaMinutes       int       `json:"eta_minutes"`
	MaxRadiusKm      int       `json:"max_radius_km"`
	WithinRadius     bool      `json:"within_radius"`
	FreeDeliveryFrom int       `json:"free_delivery_from"`
	IsOpenNow        bool      `json:"is_open_now"`
	IsNearest        bool      `json:"is_nearest"`
}

type DeliveryQuoteResponse struct {
	Quotes          []BranchDeliveryQuote `json:"quotes"`
	NearestBranchID *uuid.UUID            `json:"nearest_branch_id,omitempty"`
}

// ---------------------------------------------------------------------
// Geocoding
// ---------------------------------------------------------------------

type GeocodeResult struct {
	Label       string  `json:"label"`
	FullAddress string  `json:"full_address"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Postcode    string  `json:"postcode,omitempty"`
}

// ---------------------------------------------------------------------
// Analytics
// ---------------------------------------------------------------------

type HourlySalesPoint struct {
	Hour    string `json:"hour"`
	Orders  int    `json:"orders"`
	Revenue int    `json:"revenue"`
}

type BranchSalesPoint struct {
	BranchID   uuid.UUID `json:"branch_id"`
	BranchName string    `json:"branch_name"`
	Orders     int       `json:"orders"`
	Revenue    int       `json:"revenue"`
}

type DashboardStats struct {
	TotalOrders      int                `json:"total_orders"`
	TotalRevenue     int                `json:"total_revenue"`
	AvgOrder         int                `json:"avg_order"`
	PendingOrders    int                `json:"pending_orders"`
	CompletedOrders  int                `json:"completed_orders"`
	CancelledOrders  int                `json:"cancelled_orders"`
	OrdersDeltaPct   *float64           `json:"orders_delta_pct"`
	RevenueDeltaPct  *float64           `json:"revenue_delta_pct"`
	Hourly           []HourlySalesPoint `json:"hourly"`
	Branches         []BranchSalesPoint `json:"branches,omitempty"`
	UnacknowledgedID []uuid.UUID        `json:"-"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (r *ChangePasswordRequest) Validate() error {
	r.CurrentPassword = strings.TrimSpace(r.CurrentPassword)
	r.NewPassword = strings.TrimSpace(r.NewPassword)
	if r.CurrentPassword == "" {
		return apperror.Invalid("password saat ini wajib diisi")
	}
	if len(r.NewPassword) < 8 {
		return apperror.Invalid("password baru minimal 8 karakter")
	}
	if r.NewPassword == r.CurrentPassword {
		return apperror.Invalid("password baru tidak boleh sama dengan password saat ini")
	}
	return nil
}

type BranchCredentialsResponse struct {
	BranchID    uuid.UUID  `json:"branch_id"`
	BranchName  string     `json:"branch_name"`
	BranchSlug  string     `json:"branch_slug"`
	UserID      *uuid.UUID `json:"user_id,omitempty"`
	Name        string     `json:"name"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	Phone       string     `json:"phone"`
	HasPassword bool       `json:"has_password"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

var validUsernameRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

type UpdateBranchCredentialsRequest struct {
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	Phone    string `json:"phone,omitempty"`
}

func (r *UpdateBranchCredentialsRequest) Validate() error {
	r.Username = strings.TrimSpace(r.Username)
	r.Password = strings.TrimSpace(r.Password)
	r.Name = strings.TrimSpace(r.Name)
	r.Email = strings.TrimSpace(r.Email)
	r.Phone = strings.TrimSpace(r.Phone)

	if r.Username == "" {
		return apperror.Invalid("username cabang wajib diisi")
	}
	if len(r.Username) < 3 || len(r.Username) > 50 {
		return apperror.Invalid("username harus antara 3 sampai 50 karakter")
	}
	if !validUsernameRegex.MatchString(r.Username) {
		return apperror.Invalid("username hanya boleh berisi huruf, angka, titik (.), minus (-), atau underscore (_)")
	}
	if r.Password != "" && len(r.Password) < 8 {
		return apperror.Invalid("password minimal 8 karakter")
	}
	if r.Email != "" && !strings.Contains(r.Email, "@") {
		return apperror.Invalid("format email tidak valid")
	}
	if r.Phone != "" {
		normalized, err := NormalizeIndonesianPhone(r.Phone)
		if err != nil {
			return err
		}
		r.Phone = normalized
	}
	return nil
}

