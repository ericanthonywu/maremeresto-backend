package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBiteshipGetInstantRatesFiltersOnlyGojekAndGrabInstant(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/rates/couriers" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "biteship_test_key" {
			t.Fatalf("unexpected auth header: %q", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"pricing": []map[string]any{
				{"company": "grab", "courier_service_code": "same_day", "duration": "4 - 8 Hours", "price": 9000},
				{"company": "grab", "courier_service_code": "instant", "duration": "1 - 3 Hours", "price": 14000},
				{"company": "grab", "courier_service_code": "instant_car", "duration": "1 - 3 Hours", "price": 37000},
				{"company": "gojek", "courier_service_code": "instant", "duration": "1 - 2 Hours", "price": 17500},
			},
		})
	}))
	defer srv.Close()

	client := NewBiteshipClient("biteship_test_key", srv.URL)
	opts, err := client.GetInstantRates(context.Background(), -7.5597, 110.7942, -7.5647, 110.8227, nil)
	if err != nil {
		t.Fatalf("GetInstantRates failed: %v", err)
	}
	if len(opts) != 2 {
		t.Fatalf("got %d options, want 2", len(opts))
	}
	if opts[0].Company != "grab" || opts[0].Price != 14000 || opts[0].CourierName != "GrabExpress Instant" || !opts[0].IsCheapest {
		t.Fatalf("expected cheapest option first (Grab 14000): %+v", opts[0])
	}
	if opts[1].Company != "gojek" || opts[1].Price != 17500 || opts[1].CourierName != "GoSend Instant" || opts[1].IsCheapest {
		t.Fatalf("expected second option (GoSend 17500): %+v", opts[1])
	}
}

func TestSanitizePhoneForBiteship(t *testing.T) {
	cases := map[string]string{
		"+6282147551009": "082147551009",
		"(0271) 712452":  "0271712452",
		"08967-397-0088": "089673970088",
		"6281234567890":  "081234567890",
	}
	for in, want := range cases {
		if got := sanitizePhoneForBiteship(in); got != want {
			t.Errorf("sanitizePhoneForBiteship(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestBiteshipCreateOrderSendsExactOriginAndDestination(t *testing.T) {
	var captured BiteshipCreateOrderRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/orders" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"id":      "bs_order_1",
			"status":  "confirmed",
			"courier": map[string]any{
				"tracking_id": "trk_1",
				"waybill_id":  "wyb_1",
				"link":        "https://track.biteship.com/trk_1",
			},
		})
	}))
	defer srv.Close()

	client := NewBiteshipClient("biteship_test_key", srv.URL)
	resp, err := client.CreateOrder(context.Background(), &BiteshipCreateOrderRequest{
		OriginContactName:       "Mareme Kerten",
		OriginContactPhone:      "085867222301",
		OriginAddress:           "Jl. Samratulangi No.65, Kerten, Kec. Laweyan, Kota Surakarta, Jawa Tengah 57143",
		OriginNote:              "Outlet Mareme Kerten | KERTEN-20261123-0001",
		OriginCoordinate:        BiteshipCoordinate{Latitude: -7.5597, Longitude: 110.7942},
		DestinationContactName:  "Budi",
		DestinationContactPhone: "081234567890",
		DestinationAddress:      "Jl. Slamet Riyadi No. 100, Surakarta",
		DestinationNote:         "Pagar hitam",
		DestinationCoordinate:   BiteshipCoordinate{Latitude: -7.567, Longitude: 110.812},
		CourierCompany:          "gojek",
		CourierType:             "instant",
		DeliveryType:            "now",
		ReferenceID:             "KERTEN-20261123-0001",
		Items:                   []BiteshipRateItem{{Name: "Nasi Goreng", Value: 25000, Quantity: 1, Weight: 300}},
	})
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}
	if resp.ID != "bs_order_1" {
		t.Fatalf("unexpected order id: %q", resp.ID)
	}
	if captured.OriginAddress != "Jl. Samratulangi No.65, Kerten, Kec. Laweyan, Kota Surakarta, Jawa Tengah 57143" {
		t.Fatalf("unexpected origin address: %q", captured.OriginAddress)
	}
	if captured.OriginCoordinate.Latitude != -7.5597 || captured.OriginCoordinate.Longitude != 110.7942 {
		t.Fatalf("unexpected origin coordinate: %+v", captured.OriginCoordinate)
	}
	if captured.DestinationCoordinate.Latitude != -7.567 || captured.DestinationCoordinate.Longitude != 110.812 {
		t.Fatalf("unexpected destination coordinate: %+v", captured.DestinationCoordinate)
	}
}

func TestBiteshipStatusHelpers(t *testing.T) {
	fetchCases := map[string]bool{
		"confirmed":         false,
		"scheduled":         false,
		"allocated":         true,
		"picking_up":        true,
		"picked":            true,
		"in_transit":        true,
		"dropping_off":      true,
		"delivered":         true,
		"cancelled":         false,
		"courier_not_found": false,
	}
	for st, want := range fetchCases {
		if got := ShouldFetchBiteshipDriverInfo(st); got != want {
			t.Errorf("ShouldFetchBiteshipDriverInfo(%q) = %v, want %v", st, got, want)
		}
	}

	redispatchCases := map[string]bool{
		"confirmed":         false,
		"in_transit":        false,
		"delivered":         false,
		"cancelled":         true,
		"rejected":          true,
		"courier_not_found": true,
		"returned":          true,
	}
	for st, want := range redispatchCases {
		if got := IsBiteshipRedispatchable(st); got != want {
			t.Errorf("IsBiteshipRedispatchable(%q) = %v, want %v", st, got, want)
		}
	}
}
