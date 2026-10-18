package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIsWithinSurakarta(t *testing.T) {
	cases := []struct {
		name string
		lat  float64
		lon  float64
		want bool
	}{
		{"Central Surakarta / Slamet Riyadi", -7.5676, 110.8134, true},
		{"Kerten Laweyan", -7.5569, 110.7941, true},
		{"Banjarsari", -7.5422, 110.8103, true},
		{"Jebres", -7.5538, 110.8533, true},
		{"Pasar Kliwon", -7.5784, 110.8283, true},
		{"Serengan", -7.5804, 110.8123, true},
		{"Jakarta Monas", -6.1754, 106.8272, false},
		{"Yogyakarta Malioboro", -7.7926, 110.3658, false},
		{"Semarang Simpang Lima", -6.9904, 110.4229, false},
		{"Solo Baru The Park Mall (South outside city)", -7.5989, 110.8167, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := isWithinSurakarta(tc.lat, tc.lon)
			if got != tc.want {
				t.Errorf("isWithinSurakarta(%f, %f) = %v; want %v", tc.lat, tc.lon, got, tc.want)
			}
		})
	}
}

func TestIsSurakartaPlace(t *testing.T) {
	// A valid place inside Surakarta
	validPlace := nominatimPlace{
		DisplayName: "Jalan Slamet Riyadi, Kemlayan, Surakarta, Jawa Tengah, 57132, Indonesia",
		Lat:         "-7.5676",
		Lon:         "110.8134",
	}
	validPlace.Address.City = "Surakarta"
	validPlace.Address.Road = "Jalan Slamet Riyadi"

	if !isSurakartaPlace(validPlace) {
		t.Errorf("expected valid Surakarta place to pass, got rejected")
	}

	// Place in Sukoharjo (county set)
	sukoharjoPlace := validPlace
	sukoharjoPlace.Address.County = "Sukoharjo"
	if isSurakartaPlace(sukoharjoPlace) {
		t.Errorf("expected place with county=Sukoharjo to be rejected")
	}

	// Place outside coordinates
	outsidePlace := validPlace
	outsidePlace.Lat = "-6.2000"
	outsidePlace.Lon = "106.8000"
	if isSurakartaPlace(outsidePlace) {
		t.Errorf("expected place with Jakarta coordinates to be rejected")
	}
}

func TestNormalizeIndonesianSpelling(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"Dr radjiman 415", "Dr rajiman 415"},
		{"Jl Dr Radjiman 415", "Jl Dr Rajiman 415"},
		{"Soedirman", "Sudirman"},
		{"Tjokroaminoto", "Cokroaminoto"},
		{"Moewardi", "Muwardi"},
	}

	for _, tc := range cases {
		got := normalizeIndonesianSpelling(tc.in)
		if got != tc.want {
			t.Errorf("normalizeIndonesianSpelling(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestGenerateQueryCandidates(t *testing.T) {
	candidates := generateQueryCandidates("Dr radjiman 415")
	foundRajiman := false
	for _, c := range candidates {
		if c == "Dr rajiman 415" {
			foundRajiman = true
			break
		}
	}
	if !foundRajiman {
		t.Errorf("expected generateQueryCandidates to produce 'Dr rajiman 415', got %v", candidates)
	}

	candidatesJl := generateQueryCandidates("Jl. Dr. Radjiman No. 415")
	foundDrRajiman := false
	for _, c := range candidatesJl {
		if strings.Contains(c, "Rajiman") {
			foundDrRajiman = true
			break
		}
	}
	if !foundDrRajiman {
		t.Errorf("expected generateQueryCandidates to produce Rajiman variation, got %v", candidatesJl)
	}
}

func TestSearchAddressWithMockServer(t *testing.T) {
	// Mock Nominatim server
	nominatimHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		viewbox := r.URL.Query().Get("viewbox")
		bounded := r.URL.Query().Get("bounded")

		// Verify parameters sent
		if bounded != "1" {
			t.Errorf("expected bounded=1 for surakarta scope, got %s", bounded)
		}
		if viewbox != surakartaViewbox {
			t.Errorf("expected viewbox=%s, got %s", surakartaViewbox, viewbox)
		}

		if strings.HasPrefix(q, "Jakarta") {
			// Return empty for outside cities
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`[]`))
			return
		}

		// OSM doesn't have "radjiman" with dj, only "rajiman" with j
		if strings.Contains(strings.ToLower(q), "radjiman") {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`[]`))
			return
		}

		if strings.Contains(strings.ToLower(q), "rajiman") {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`[
				{
					"place_id": 99999,
					"lat": "-7.5717985",
					"lon": "110.8125492",
					"display_name": "Jalan Dokter Rajiman Widiodiningrat, Sriwedari, Surakarta, Jawa Tengah, 57149, Indonesia",
					"address": {
						"road": "Jalan Dokter Rajiman Widiodiningrat",
						"village": "Sriwedari",
						"city": "Surakarta",
						"state": "Jawa Tengah",
						"postcode": "57149",
						"country": "Indonesia"
					}
				}
			]`))
			return
		}

		// Return a valid Surakarta place for Slamet Riyadi
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[
			{
				"place_id": 12345,
				"lat": "-7.5676627",
				"lon": "110.8134472",
				"display_name": "Jalan Slamet Riyadi, Kemlayan, Surakarta, Jawa Tengah, 57132, Indonesia",
				"address": {
					"road": "Jalan Slamet Riyadi",
					"village": "Kemlayan",
					"city": "Surakarta",
					"state": "Jawa Tengah",
					"postcode": "57132",
					"country": "Indonesia"
				}
			}
		]`))
	})

	server := httptest.NewServer(nominatimHandler)
	defer server.Close()

	svc := &Service{
		geocoder: newGeocoder(server.URL, "test@example.com"),
	}

	// Test Surakarta query
	results, err := svc.SearchAddress(context.Background(), "Slamet Riyadi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Label == "" || results[0].Latitude == 0 {
		t.Errorf("invalid result content: %+v", results[0])
	}

	// Test Dr radjiman 415 query (should fall back to Dr rajiman 415 and succeed)
	radjimanResults, err := svc.SearchAddress(context.Background(), "Dr radjiman 415")
	if err != nil {
		t.Fatalf("unexpected error on Dr radjiman 415: %v", err)
	}
	if len(radjimanResults) != 1 {
		t.Fatalf("expected 1 result for 'Dr radjiman 415', got %d", len(radjimanResults))
	}
	if !strings.Contains(radjimanResults[0].FullAddress, "Rajiman") {
		t.Errorf("expected address to contain Rajiman, got %s", radjimanResults[0].FullAddress)
	}

	// Test outside query
	outsideResults, err := svc.SearchAddress(context.Background(), "Jakarta")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(outsideResults) != 0 {
		t.Errorf("expected 0 results for outside query, got %d", len(outsideResults))
	}
}
