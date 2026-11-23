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
	if opts[0].Company != "gojek" || opts[0].Price != 17500 || opts[0].CourierName != "GoSend Instant" {
		t.Fatalf("unexpected first option: %+v", opts[0])
	}
	if opts[1].Company != "grab" || opts[1].Price != 14000 || opts[1].CourierName != "GrabExpress Instant" {
		t.Fatalf("unexpected second option: %+v", opts[1])
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
