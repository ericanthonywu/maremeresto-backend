package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ericanthonywu/maremereso-olga/backend/internal/apperror"
	"github.com/ericanthonywu/maremereso-olga/backend/internal/dto"
)

// The geocoder is proxied through the backend rather than called from the
// browser: Nominatim forbids anonymous browser traffic, and routing it here
// lets us cache, rate-limit and swap providers without shipping a new PWA.
const (
	geocodeTimeout    = 8 * time.Second
	geocodeCacheTTL   = 24 * time.Hour
	geocodeMinSpacing = 1100 * time.Millisecond // Nominatim fair-use: max 1 req/s
	geocodeMaxResults = 6
)

// Viewport biasing the search towards Surakarta and the surrounding
// Soloraya area, where every outlet trades.
const (
	soloViewbox = "110.6500,-7.7200,110.9500,-7.4200"
)

type geocodeCacheEntry struct {
	body      []byte
	expiresAt time.Time
}

type geocoder struct {
	baseURL string
	email   string
	client  *http.Client

	mu    sync.Mutex
	cache map[string]geocodeCacheEntry
	last  time.Time
}

func newGeocoder(baseURL, email string) *geocoder {
	return &geocoder{
		baseURL: strings.TrimRight(baseURL, "/"),
		email:   email,
		client:  &http.Client{Timeout: geocodeTimeout},
		cache:   make(map[string]geocodeCacheEntry),
	}
}

func (g *geocoder) fetch(ctx context.Context, path string, params url.Values) ([]byte, error) {
	params.Set("format", "jsonv2")
	if g.email != "" {
		params.Set("email", g.email)
	}
	cacheKey := path + "?" + params.Encode()

	g.mu.Lock()
	if entry, ok := g.cache[cacheKey]; ok && time.Now().Before(entry.expiresAt) {
		g.mu.Unlock()
		return entry.body, nil
	}
	// Serialise outbound calls and space them out to stay inside the
	// provider's fair-use policy.
	if wait := geocodeMinSpacing - time.Since(g.last); wait > 0 {
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			g.mu.Unlock()
			return nil, ctx.Err()
		}
	}
	g.last = time.Now()
	g.mu.Unlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.baseURL+path+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "MaremeresoOlga/1.0 (+cafe ordering PWA)")
	req.Header.Set("Accept", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperror.ErrGeocoderUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: upstream status %d", apperror.ErrGeocoderUnavailable, resp.StatusCode)
	}

	body := make([]byte, 0, 8192)
	buf := make([]byte, 4096)
	for {
		n, readErr := resp.Body.Read(buf)
		body = append(body, buf[:n]...)
		if readErr != nil {
			break
		}
		if len(body) > 1<<20 {
			break
		}
	}

	g.mu.Lock()
	g.cache[cacheKey] = geocodeCacheEntry{body: body, expiresAt: time.Now().Add(geocodeCacheTTL)}
	// Keep the cache from growing without bound on a long-lived process.
	if len(g.cache) > 2000 {
		now := time.Now()
		for k, v := range g.cache {
			if now.After(v.expiresAt) {
				delete(g.cache, k)
			}
		}
	}
	g.mu.Unlock()

	return body, nil
}

type nominatimPlace struct {
	DisplayName string `json:"display_name"`
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	Type        string `json:"type"`
	Address     struct {
		Road         string `json:"road"`
		Suburb       string `json:"suburb"`
		Village      string `json:"village"`
		CityDistrict string `json:"city_district"`
		City         string `json:"city"`
		Town         string `json:"town"`
		County       string `json:"county"`
		State        string `json:"state"`
		Postcode     string `json:"postcode"`
	} `json:"address"`
}

func (p nominatimPlace) toResult() (dto.GeocodeResult, bool) {
	lat, err1 := strconv.ParseFloat(p.Lat, 64)
	lon, err2 := strconv.ParseFloat(p.Lon, 64)
	if err1 != nil || err2 != nil {
		return dto.GeocodeResult{}, false
	}

	// A short label is far more usable in a one-line address chip than
	// Nominatim's full comma-separated display name.
	parts := make([]string, 0, 4)
	for _, v := range []string{
		p.Address.Road,
		firstNonEmpty(p.Address.Suburb, p.Address.Village),
		firstNonEmpty(p.Address.CityDistrict, p.Address.City, p.Address.Town, p.Address.County),
	} {
		if v != "" {
			parts = append(parts, v)
		}
	}
	short := strings.Join(parts, ", ")
	if short == "" {
		short = p.DisplayName
	}

	return dto.GeocodeResult{
		Label:       short,
		FullAddress: p.DisplayName,
		Latitude:    lat,
		Longitude:   lon,
		Postcode:    p.Address.Postcode,
	}, true
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// SearchAddress resolves free-text typed by the customer into real
// coordinates. Previously the app pinned every typed address to one fixed
// coordinate, which silently mispriced delivery.
func (s *Service) SearchAddress(ctx context.Context, query string) ([]dto.GeocodeResult, error) {
	query = strings.TrimSpace(query)
	if len(query) < 3 {
		return nil, apperror.Invalid("masukkan minimal 3 karakter untuk mencari alamat")
	}
	if len(query) > 200 {
		return nil, apperror.Invalid("kata kunci alamat terlalu panjang")
	}

	params := url.Values{}
	params.Set("q", query)
	params.Set("limit", strconv.Itoa(geocodeMaxResults))
	params.Set("addressdetails", "1")
	params.Set("countrycodes", "id")
	params.Set("viewbox", soloViewbox)
	params.Set("bounded", "0")

	body, err := s.geocoder.fetch(ctx, "/search", params)
	if err != nil {
		return nil, err
	}

	var places []nominatimPlace
	if err := json.Unmarshal(body, &places); err != nil {
		return nil, apperror.ErrGeocoderUnavailable
	}

	results := make([]dto.GeocodeResult, 0, len(places))
	for _, p := range places {
		if r, ok := p.toResult(); ok {
			results = append(results, r)
		}
	}
	return results, nil
}

// ReverseGeocode turns the GPS fix from the browser into a human-readable
// address so the customer can confirm where the courier is being sent.
func (s *Service) ReverseGeocode(ctx context.Context, lat, lon float64) (*dto.GeocodeResult, error) {
	if lat < -90 || lat > 90 || lon < -180 || lon > 180 {
		return nil, apperror.Invalid("koordinat tidak valid")
	}

	params := url.Values{}
	params.Set("lat", strconv.FormatFloat(lat, 'f', 7, 64))
	params.Set("lon", strconv.FormatFloat(lon, 'f', 7, 64))
	params.Set("zoom", "18")
	params.Set("addressdetails", "1")

	body, err := s.geocoder.fetch(ctx, "/reverse", params)
	if err != nil {
		return nil, err
	}

	var place nominatimPlace
	if err := json.Unmarshal(body, &place); err != nil {
		return nil, apperror.ErrGeocoderUnavailable
	}

	result, ok := place.toResult()
	if !ok {
		return nil, apperror.ErrNotFound
	}
	// Always answer with the caller's own fix, not the centroid of whatever
	// feature matched, so distance maths stays anchored to the real position.
	result.Latitude = lat
	result.Longitude = lon
	return &result, nil
}
