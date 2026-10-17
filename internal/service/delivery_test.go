package service

import (
	"math"
	"testing"
	"time"

	"github.com/ericanthonywu/maremereso-olga/backend/internal/model"
)

// Real outlet coordinates from the branches table.
const (
	kertenLat, kertenLon       = -7.5597, 110.7942
	makamhajiLat, makamhajiLon = -7.5662, 110.7788
	makDjanLat, makDjanLon     = -7.5647, 110.8227
)

func TestHaversineKmMatchesKnownDistances(t *testing.T) {
	// Kerten -> Mak Djan is about 3.2 km straight line across central Solo.
	got := HaversineKm(kertenLat, kertenLon, makDjanLat, makDjanLon)
	if math.Abs(got-3.2) > 0.3 {
		t.Errorf("Kerten->Mak Djan = %.2f km, want ~3.2 km", got)
	}

	// Kerten -> Makamhaji is about 1.8 km.
	got = HaversineKm(kertenLat, kertenLon, makamhajiLat, makamhajiLon)
	if math.Abs(got-1.8) > 0.3 {
		t.Errorf("Kerten->Makamhaji = %.2f km, want ~1.8 km", got)
	}

	if d := HaversineKm(kertenLat, kertenLon, kertenLat, kertenLon); d != 0 {
		t.Errorf("distance to self = %v, want 0", d)
	}
}

func TestHaversineIsSymmetric(t *testing.T) {
	a := HaversineKm(kertenLat, kertenLon, makDjanLat, makDjanLon)
	b := HaversineKm(makDjanLat, makDjanLon, kertenLat, kertenLon)
	if math.Abs(a-b) > 1e-9 {
		t.Errorf("not symmetric: %v vs %v", a, b)
	}
}

func TestRoadDistanceAppliesFloorAndRounding(t *testing.T) {
	// Two points a few metres apart still bill the 0.5 km minimum.
	if got := RoadDistanceKm(kertenLat, kertenLon, kertenLat+0.0001, kertenLon); got != 0.5 {
		t.Errorf("minimum billable distance = %v, want 0.5", got)
	}

	// The result must be rounded to one decimal, because the quote endpoint
	// and order creation both display and charge from this number.
	got := RoadDistanceKm(kertenLat, kertenLon, makDjanLat, makDjanLon)
	if got != math.Round(got*10)/10 {
		t.Errorf("%v is not rounded to one decimal", got)
	}
}

func settings() *model.BranchSettings {
	return &model.BranchSettings{
		BaseDeliveryFeeNear:   0,
		BaseDeliveryFeeMid:    8000,
		BaseDeliveryFeeFar:    12000,
		NearThresholdKm:       1,
		MidThresholdKm:        5,
		FreeDeliveryThreshold: 0,
		ServiceFee:            2000,
	}
}

func TestDeliveryFeeTiers(t *testing.T) {
	s := settings()

	cases := []struct {
		name      string
		distance  float64
		subtotal  int
		orderType string
		want      int
	}{
		{"free at one kilometre", 1.0, 50000, "delivery", 0},
		{"just past free tier", 1.1, 50000, "delivery", 8000},
		{"exactly on five kilometre boundary", 5.0, 50000, "delivery", 8000},
		{"past five kilometre boundary", 5.1, 50000, "delivery", 12000},
		{"up to ten kilometres", 10.0, 50000, "delivery", 12000},
		{"pickup is never charged", 9.0, 50000, "pickup", 0},
		{"scheduled is charged like delivery", 5.1, 50000, "scheduled", 12000},
		{"subtotal does not change fixed fee", 9.0, 200000, "delivery", 12000},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := DeliveryFee(s, tc.distance, tc.subtotal, tc.orderType); got != tc.want {
				t.Errorf("DeliveryFee(%v km, Rp%d, %s) = %d, want %d",
					tc.distance, tc.subtotal, tc.orderType, got, tc.want)
			}
		})
	}
}

func TestDeliveryFeeIsNotChangedBySubtotal(t *testing.T) {
	s := settings()
	s.FreeDeliveryThreshold = 0

	if got := DeliveryFee(s, 2.0, 10_000_000, "delivery"); got != 8000 {
		t.Errorf("a large order should keep the fixed distance fee, got %d", got)
	}
}

func TestEtaGrowsWithDistance(t *testing.T) {
	near := EtaMinutes(1.0)
	far := EtaMinutes(10.0)
	if near >= far {
		t.Errorf("ETA should grow with distance: %d vs %d", near, far)
	}
	if near <= preparationMinute {
		t.Errorf("ETA must include preparation time, got %d", near)
	}
}

func TestIsWithinOperatingHours(t *testing.T) {
	weekdaySchedule := map[string]any{
		"weekday": map[string]any{"open": "08:00", "close": "22:00"},
		"weekend": map[string]any{"open": "09:00", "close": "21:00"},
	}

	at := func(day int, hour, min int) time.Time {
		// 2026-09-07 is a Monday.
		return time.Date(2026, 9, 6+day, hour, min, 0, 0, OutletLocation)
	}

	cases := []struct {
		name string
		when time.Time
		want bool
	}{
		{"monday mid-morning", at(1, 10, 0), true},
		{"monday exactly at opening", at(1, 8, 0), true},
		{"monday one minute before opening", at(1, 7, 59), false},
		{"monday exactly at closing", at(1, 22, 0), false},
		{"monday just before closing", at(1, 21, 59), true},
		{"sunday before weekend opening", at(0, 8, 30), false},
		{"sunday after weekend opening", at(0, 9, 30), true},
		{"sunday after weekend closing", at(0, 21, 30), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsWithinOperatingHours(weekdaySchedule, tc.when); got != tc.want {
				t.Errorf("IsWithinOperatingHours(%s) = %v, want %v", tc.when.Format(time.RFC3339), got, tc.want)
			}
		})
	}
}

func TestOperatingHoursCrossingMidnight(t *testing.T) {
	schedule := map[string]any{
		"weekday": map[string]any{"open": "17:00", "close": "01:00"},
	}

	tue := func(h, m int) time.Time { return time.Date(2026, 9, 8, h, m, 0, 0, OutletLocation) }

	for _, tc := range []struct {
		when time.Time
		want bool
	}{
		{tue(18, 0), true},
		{tue(23, 59), true},
		{tue(0, 30), true},
		{tue(1, 0), false},
		{tue(16, 59), false},
	} {
		if got := IsWithinOperatingHours(schedule, tc.when); got != tc.want {
			t.Errorf("at %s got %v, want %v", tc.when.Format("15:04"), got, tc.want)
		}
	}
}

// A malformed or missing schedule must not lock an outlet out of trading.
func TestOperatingHoursFailOpen(t *testing.T) {
	now := time.Now()

	for name, schedule := range map[string]map[string]any{
		"nil":            nil,
		"empty":          {},
		"wrong shape":    {"weekday": "08:00-22:00"},
		"unparseable":    {"weekday": map[string]any{"open": "8am", "close": "10pm"}},
		"missing fields": {"weekday": map[string]any{"open": "08:00"}},
	} {
		if !IsWithinOperatingHours(schedule, now) {
			t.Errorf("%s schedule should fail open", name)
		}
	}
}

func TestPerDayOverrideBeatsWeekdayBucket(t *testing.T) {
	schedule := map[string]any{
		"weekday": map[string]any{"open": "08:00", "close": "22:00"},
		"monday":  map[string]any{"open": "11:00", "close": "15:00"},
	}

	monday9 := time.Date(2026, 9, 7, 9, 0, 0, 0, OutletLocation)
	if IsWithinOperatingHours(schedule, monday9) {
		t.Error("the monday override should keep the outlet shut at 09:00")
	}

	tuesday9 := time.Date(2026, 9, 8, 9, 0, 0, 0, OutletLocation)
	if !IsWithinOperatingHours(schedule, tuesday9) {
		t.Error("tuesday should still use the weekday bucket and be open at 09:00")
	}
}

func TestFormatRupiah(t *testing.T) {
	for in, want := range map[int]string{
		0: "0", 999: "999", 1000: "1.000", 20000: "20.000",
		150000: "150.000", 1234567: "1.234.567",
	} {
		if got := formatRupiah(in); got != want {
			t.Errorf("formatRupiah(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestIsWithinSoloArea(t *testing.T) {
	// Kerten (Solo) -> inside Solo Raya
	if !IsWithinSoloArea(-7.5561, 110.7931) {
		t.Fatalf("expected Kerten (-7.5561, 110.7931) to be within Solo area")
	}
	// Makamhaji (Kartasura / Sukoharjo border) -> inside Solo Raya
	if !IsWithinSoloArea(-7.5681, 110.7712) {
		t.Fatalf("expected Makamhaji (-7.5681, 110.7712) to be within Solo area")
	}
	// Jakarta -> outside Solo Raya
	if IsWithinSoloArea(-6.2088, 106.8456) {
		t.Fatalf("expected Jakarta (-6.2088, 106.8456) to be outside Solo area")
	}
	// Yogyakarta -> outside Solo Raya
	if IsWithinSoloArea(-7.7956, 110.3695) {
		t.Fatalf("expected Yogyakarta (-7.7956, 110.3695) to be outside Solo area")
	}
}

func TestEffectiveMaxRadiusKm(t *testing.T) {
	if got := effectiveMaxRadiusKm(nil); got != 10.0 {
		t.Fatalf("effectiveMaxRadiusKm(nil) = %v, want 10.0", got)
	}
	if got := effectiveMaxRadiusKm(&model.BranchSettings{MaxDeliveryRadiusKm: 7}); got != 7.0 {
		t.Fatalf("effectiveMaxRadiusKm(7) = %v, want 7.0", got)
	}
	if got := effectiveMaxRadiusKm(&model.BranchSettings{MaxDeliveryRadiusKm: 15}); got != 10.0 {
		t.Fatalf("effectiveMaxRadiusKm(15) = %v, want 10.0 (capped at maxServiceableKm)", got)
	}
}

func TestSnapItemsSumMatchesGrandTotal(t *testing.T) {
	order := &model.Order{
		Items: []model.OrderItem{
			{ItemName: "Nasi Goreng", ItemPrice: 22000, Quantity: 2, LineTotal: 44000},
			{ItemName: "Es Teh", ItemPrice: 5000, Quantity: 1, LineTotal: 5000},
		},
		Subtotal:    49000,
		DeliveryFee: 8000,
		ServiceFee:  2000,
		Discount:    5000,
		GrandTotal:  54000,
	}
	items := snapItems(order)
	var sum int64
	for _, it := range *items {
		sum += it.Price * int64(it.Qty)
	}
	if sum != int64(order.GrandTotal) {
		t.Fatalf("snapItems sum = %d, want GrandTotal %d", sum, order.GrandTotal)
	}
}

