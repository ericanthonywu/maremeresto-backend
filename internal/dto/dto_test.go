package dto

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNormalizeIndonesianPhone(t *testing.T) {
	valid := map[string]string{
		"081234567890":      "+6281234567890",
		"6281234567890":     "+6281234567890",
		"+6281234567890":    "+6281234567890",
		"+62 812-3456-7890": "+6281234567890",
		"0812 3456 7890":    "+6281234567890",
		"(0812) 3456-7890":  "+6281234567890",
		"08123456789":       "+628123456789",
	}
	for in, want := range valid {
		got, err := NormalizeIndonesianPhone(in)
		if err != nil {
			t.Errorf("NormalizeIndonesianPhone(%q) returned %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("NormalizeIndonesianPhone(%q) = %q, want %q", in, got, want)
		}
	}

	invalid := []string{
		"", "   ", "123", "0812", "08023456789", // second digit may not be 0
		"+1234567890", "not a phone", "0812345678901234",
	}
	for _, in := range invalid {
		if got, err := NormalizeIndonesianPhone(in); err == nil {
			t.Errorf("NormalizeIndonesianPhone(%q) unexpectedly succeeded with %q", in, got)
		}
	}
}

// Every spelling of the same number must normalise identically, otherwise a
// returning customer is created as a second account.
func TestPhoneNormalizationIsStable(t *testing.T) {
	forms := []string{"081236391375", "6281236391375", "+6281236391375", "+62 812-3639-1375"}

	var first string
	for i, f := range forms {
		got, err := NormalizeIndonesianPhone(f)
		if err != nil {
			t.Fatalf("%q failed: %v", f, err)
		}
		if i == 0 {
			first = got
			continue
		}
		if got != first {
			t.Errorf("%q normalised to %q, expected %q", f, got, first)
		}
	}
}

func validOrder() CreateOrderRequest {
	return CreateOrderRequest{
		BranchID:      uuid.New(),
		OrderType:     "delivery",
		CustomerName:  "Budi Santoso",
		CustomerPhone: "081234567890",
		DeliveryNotes: "Ketuk pagar tiga kali",
		Items:         []CreateOrderItemRequest{{MenuItemID: uuid.New(), Quantity: 2}},
	}
}

func TestCreateOrderRequestValidate(t *testing.T) {
	base := validOrder()
	if err := base.Validate(); err != nil {
		t.Fatalf("a valid order was rejected: %v", err)
	}

	cases := map[string]func(*CreateOrderRequest){
		"no branch":              func(r *CreateOrderRequest) { r.BranchID = uuid.Nil },
		"unknown type":           func(r *CreateOrderRequest) { r.OrderType = "teleport" },
		"empty name":             func(r *CreateOrderRequest) { r.CustomerName = " " },
		"no items":               func(r *CreateOrderRequest) { r.Items = nil },
		"zero quantity":          func(r *CreateOrderRequest) { r.Items[0].Quantity = 0 },
		"negative quantity":      func(r *CreateOrderRequest) { r.Items[0].Quantity = -5 },
		"absurd quantity":        func(r *CreateOrderRequest) { r.Items[0].Quantity = 100000 },
		"nil menu item":          func(r *CreateOrderRequest) { r.Items[0].MenuItemID = uuid.Nil },
		"bad latitude":           func(r *CreateOrderRequest) { v := 95.0; r.DeliveryLat = &v },
		"bad longitude":          func(r *CreateOrderRequest) { v := -200.0; r.DeliveryLon = &v },
		"overlong address":       func(r *CreateOrderRequest) { r.DeliveryAddress = strings.Repeat("x", 501) },
		"delivery note required": func(r *CreateOrderRequest) { r.DeliveryNotes = " " },
	}

	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			r := validOrder()
			mutate(&r)
			if err := r.Validate(); err == nil {
				t.Errorf("%s should have been rejected", name)
			}
		})
	}
}

func TestScheduledOrderRequiresAFutureTime(t *testing.T) {
	r := validOrder()
	r.OrderType = "scheduled"

	if err := r.Validate(); err == nil {
		t.Error("a scheduled order with no time should be rejected")
	}

	past := time.Now().Add(-2 * time.Hour)
	r.ScheduledAt = &past
	if err := r.Validate(); err == nil {
		t.Error("a scheduled order in the past should be rejected")
	}

	tooFar := time.Now().Add(30 * 24 * time.Hour)
	r.ScheduledAt = &tooFar
	if err := r.Validate(); err == nil {
		t.Error("a schedule a month out should be rejected")
	}

	soon := time.Now().Add(2 * time.Hour)
	r.ScheduledAt = &soon
	if err := r.Validate(); err != nil {
		t.Errorf("a sensible schedule was rejected: %v", err)
	}
}

// A time left over from the "scheduled" tab must not ride along on an
// immediate order.
func TestNonScheduledOrderClearsScheduledAt(t *testing.T) {
	r := validOrder()
	later := time.Now().Add(3 * time.Hour)
	r.ScheduledAt = &later

	if err := r.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.ScheduledAt != nil {
		t.Error("ScheduledAt should be cleared for a non-scheduled order")
	}
}

func TestCreateMenuItemValidate(t *testing.T) {
	valid := CreateMenuItemRequest{
		CategoryID: uuid.New(), Name: "Es Kopi Susu", Price: 22000,
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid menu item rejected: %v", err)
	}

	cases := map[string]CreateMenuItemRequest{
		"no category":   {Name: "Kopi", Price: 22000},
		"short name":    {CategoryID: uuid.New(), Name: "X", Price: 22000},
		"price too low": {CategoryID: uuid.New(), Name: "Kopi", Price: 500},
		"price absurd":  {CategoryID: uuid.New(), Name: "Kopi", Price: 99999999},
		"negative sort": {CategoryID: uuid.New(), Name: "Kopi", Price: 22000, SortOrder: -1},
	}
	for name, req := range cases {
		t.Run(name, func(t *testing.T) {
			r := req
			if err := r.Validate(); err == nil {
				t.Errorf("%s should have been rejected", name)
			}
		})
	}
}

func TestUpdateBranchSettingsValidate(t *testing.T) {
	base := func() UpdateBranchSettingsRequest {
		return UpdateBranchSettingsRequest{
			MaxDeliveryRadiusKm: 12, NearThresholdKm: 3, MidThresholdKm: 7,
			BaseDeliveryFeeNear: 8000, BaseDeliveryFeeMid: 12000, BaseDeliveryFeeFar: 18000,
			ServiceFee: 2000, MinOrderAmount: 20000, FreeDeliveryThreshold: 150000,
			WhatsappNumber: "081234567890",
			OperatingHours: map[string]any{
				"weekday": map[string]any{"open": "08:00", "close": "22:00"},
			},
		}
	}

	v := base()
	if err := v.Validate(); err != nil {
		t.Fatalf("valid settings rejected: %v", err)
	}
	if v.WhatsappNumber != "+6281234567890" {
		t.Errorf("WhatsApp number should be normalised, got %q", v.WhatsappNumber)
	}

	cases := map[string]func(*UpdateBranchSettingsRequest){
		"mid below near threshold": func(r *UpdateBranchSettingsRequest) { r.MidThresholdKm = 2 },
		"radius below mid":         func(r *UpdateBranchSettingsRequest) { r.MaxDeliveryRadiusKm = 5 },
		"radius absurd":            func(r *UpdateBranchSettingsRequest) { r.MaxDeliveryRadiusKm = 500 },
		"fees not increasing":      func(r *UpdateBranchSettingsRequest) { r.BaseDeliveryFeeFar = 1000 },
		"negative fee":             func(r *UpdateBranchSettingsRequest) { r.BaseDeliveryFeeNear = -1 },
		"bad whatsapp":             func(r *UpdateBranchSettingsRequest) { r.WhatsappNumber = "12345" },
		"bad clock": func(r *UpdateBranchSettingsRequest) {
			r.OperatingHours = map[string]any{"weekday": map[string]any{"open": "8", "close": "22:00"}}
		},
	}

	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			r := base()
			mutate(&r)
			if err := r.Validate(); err == nil {
				t.Errorf("%s should have been rejected", name)
			}
		})
	}
}

func TestAssignDriverValidate(t *testing.T) {
	r := AssignDriverRequest{DriverName: "Andi", DriverPhone: "0812 3456 7890", DriverPlate: "ad 1234 xy"}
	if err := r.Validate(); err != nil {
		t.Fatalf("valid driver rejected: %v", err)
	}
	if r.DriverPhone != "+6281234567890" {
		t.Errorf("driver phone not normalised: %q", r.DriverPhone)
	}
	if r.DriverPlate != "AD 1234 XY" {
		t.Errorf("plate should be upper-cased, got %q", r.DriverPlate)
	}

	bad := AssignDriverRequest{DriverName: "A", DriverPhone: "0812"}
	if err := bad.Validate(); err == nil {
		t.Error("an invalid driver should be rejected")
	}
}

func TestCreateCategoryRequestValidate(t *testing.T) {
	valid := CreateCategoryRequest{Name: "Snack & Bites", Emoji: "🍟", SortOrder: 6}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid category rejected: %v", err)
	}

	emptyName := CreateCategoryRequest{Name: "  "}
	if err := emptyName.Validate(); err == nil {
		t.Error("empty category name should be rejected")
	}
}

func TestUpdateBranchProfileAllowsLandlinePhone(t *testing.T) {
	cases := []string{
		"(021) 1234567",
		"021-7654321",
		"(0271) 712345",
		"081234567890",
		"+6281234567890",
	}

	for _, phone := range cases {
		r := UpdateBranchProfileRequest{
			Name:      "Mak Djan",
			Address:   "Jl. Gatot Subroto No. 10",
			Phone:     phone,
			Latitude:  -7.5532,
			Longitude: 110.8061,
		}
		if err := r.Validate(); err != nil {
			t.Errorf("phone %q should be accepted for branch profile, but got error: %v", phone, err)
		}
	}
}
