package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/ericanthonywu/maremereso-olga/backend/internal/dto"
)

// BiteshipClient wraps Biteship's REST API for instant courier rates, order
// creation, lookup, and cancellation.
type BiteshipClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewBiteshipClient(apiKey, baseURL string) *BiteshipClient {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = "https://api.biteship.com"
	}
	return &BiteshipClient{
		apiKey:  strings.TrimSpace(apiKey),
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *BiteshipClient) Enabled() bool {
	return c != nil && c.apiKey != ""
}

// CourierDisplayName returns the customer-facing Indonesian label for a
// supported Biteship instant courier.
func CourierDisplayName(company string) string {
	switch strings.ToLower(strings.TrimSpace(company)) {
	case "gojek":
		return "GoSend Instant"
	case "grab":
		return "GrabExpress Instant"
	default:
		return strings.TrimSpace(company)
	}
}

// IsSupportedInstantCourier reports whether company is one of the configured
// instant delivery providers.
func IsSupportedInstantCourier(company string) bool {
	switch strings.ToLower(strings.TrimSpace(company)) {
	case "gojek", "grab":
		return true
	default:
		return false
	}
}

// sanitizePhoneForBiteship normalizes Indonesian contact numbers so Biteship
// accepts landline or mobile numbers cleanly.
func sanitizePhoneForBiteship(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, raw)
	if strings.HasPrefix(digits, "62") {
		return "0" + digits[2:]
	}
	return digits
}

type BiteshipRateItem struct {
	Name     string `json:"name"`
	Value    int    `json:"value"`
	Quantity int    `json:"quantity"`
	Weight   int    `json:"weight"`
}

type biteshipRatesRequest struct {
	OriginLatitude       float64            `json:"origin_latitude"`
	OriginLongitude      float64            `json:"origin_longitude"`
	DestinationLatitude  float64            `json:"destination_latitude"`
	DestinationLongitude float64            `json:"destination_longitude"`
	Couriers             string             `json:"couriers"`
	Items                []BiteshipRateItem `json:"items"`
}

type biteshipRatesResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Pricing []struct {
		Company            string `json:"company"`
		CourierCode        string `json:"courier_code"`
		CourierName        string `json:"courier_name"`
		CourierServiceCode string `json:"courier_service_code"`
		CourierServiceName string `json:"courier_service_name"`
		Duration           string `json:"duration"`
		Type               string `json:"type"`
		Price              int    `json:"price"`
	} `json:"pricing"`
}

// GetInstantRates queries Biteship for GoSend Instant and GrabExpress Instant
// rates between two coordinates.
func (c *BiteshipClient) GetInstantRates(
	ctx context.Context,
	originLat, originLon, destLat, destLon float64,
	items []BiteshipRateItem,
) ([]dto.CourierOption, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("biteship is not configured")
	}
	if len(items) == 0 {
		items = []BiteshipRateItem{{
			Name:     "Pesanan Makanan & Minuman",
			Value:    25000,
			Quantity: 1,
			Weight:   500,
		}}
	}

	reqBody := biteshipRatesRequest{
		OriginLatitude:       originLat,
		OriginLongitude:      originLon,
		DestinationLatitude:  destLat,
		DestinationLongitude: destLon,
		Couriers:             "gojek,grab",
		Items:                items,
	}

	var resp biteshipRatesResponse
	if err := c.doJSON(ctx, http.MethodPost, "/v1/rates/couriers", reqBody, &resp); err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("biteship rates failed: %s", resp.Error)
	}

	seen := make(map[string]bool)
	options := make([]dto.CourierOption, 0, 2)
	for _, p := range resp.Pricing {
		company := strings.ToLower(strings.TrimSpace(p.Company))
		if company == "" {
			company = strings.ToLower(strings.TrimSpace(p.CourierCode))
		}
		serviceCode := strings.ToLower(strings.TrimSpace(p.CourierServiceCode))
		if serviceCode == "" {
			serviceCode = strings.ToLower(strings.TrimSpace(p.Type))
		}
		if !IsSupportedInstantCourier(company) || serviceCode != "instant" {
			continue
		}
		if seen[company] {
			continue
		}
		seen[company] = true
		options = append(options, dto.CourierOption{
			Company:     company,
			Type:        "instant",
			CourierName: CourierDisplayName(company),
			Duration:    strings.TrimSpace(p.Duration),
			Price:       p.Price,
		})
	}

	// Keep deterministic order: GoSend first, then Grab.
	sort.SliceStable(options, func(i, j int) bool {
		if options[i].Company == options[j].Company {
			return false
		}
		return options[i].Company == "gojek"
	})
	return options, nil
}

type BiteshipCoordinate struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type BiteshipCreateOrderRequest struct {
	OriginContactName      string             `json:"origin_contact_name"`
	OriginContactPhone     string             `json:"origin_contact_phone"`
	OriginAddress          string             `json:"origin_address"`
	OriginNote             string             `json:"origin_note,omitempty"`
	OriginCoordinate       BiteshipCoordinate `json:"origin_coordinate"`
	DestinationContactName  string            `json:"destination_contact_name"`
	DestinationContactPhone string            `json:"destination_contact_phone"`
	DestinationAddress     string             `json:"destination_address"`
	DestinationNote        string             `json:"destination_note,omitempty"`
	DestinationCoordinate  BiteshipCoordinate `json:"destination_coordinate"`
	CourierCompany         string             `json:"courier_company"`
	CourierType            string             `json:"courier_type"`
	DeliveryType           string             `json:"delivery_type"`
	ReferenceID            string             `json:"reference_id,omitempty"`
	Items                  []BiteshipRateItem `json:"items"`
}

type BiteshipOrderCourier struct {
	TrackingID        string  `json:"tracking_id"`
	WaybillID         string  `json:"waybill_id"`
	Company           string  `json:"company"`
	Type              string  `json:"type"`
	Link              string  `json:"link"`
	Name              *string `json:"name"`
	Phone             *string `json:"phone"`
	DriverName        *string `json:"driver_name"`
	DriverPhone       *string `json:"driver_phone"`
	DriverPlateNumber *string `json:"driver_plate_number"`
	DriverPhotoURL    *string `json:"driver_photo_url"`
}

type BiteshipOrderResponse struct {
	Success     bool                  `json:"success"`
	Message     string                `json:"message"`
	Error       string                `json:"error"`
	ID          string                `json:"id"`
	Status      string                `json:"status"`
	Price       int                   `json:"price"`
	ReferenceID string                `json:"reference_id"`
	Courier     BiteshipOrderCourier  `json:"courier"`
}

func (c *BiteshipClient) CreateOrder(ctx context.Context, req *BiteshipCreateOrderRequest) (*BiteshipOrderResponse, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("biteship is not configured")
	}
	var resp BiteshipOrderResponse
	if err := c.doJSON(ctx, http.MethodPost, "/v1/orders", req, &resp); err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("biteship create order failed: %s", resp.Error)
	}
	return &resp, nil
}

func (c *BiteshipClient) GetOrder(ctx context.Context, biteshipOrderID string) (*BiteshipOrderResponse, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("biteship is not configured")
	}
	var resp BiteshipOrderResponse
	if err := c.doJSON(ctx, http.MethodGet, "/v1/orders/"+biteshipOrderID, nil, &resp); err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("biteship get order failed: %s", resp.Error)
	}
	return &resp, nil
}

func (c *BiteshipClient) doJSON(ctx context.Context, method, path string, body any, dst any) error {
	var bodyReader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode biteship request: %w", err)
		}
		bodyReader = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("build biteship request: %w", err)
	}
	req.Header.Set("Authorization", c.apiKey)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("biteship request failed: %w", err)
	}
	defer resp.Body.Close()

	rawResp, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read biteship response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errPayload struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		_ = json.Unmarshal(rawResp, &errPayload)
		msg := strings.TrimSpace(errPayload.Error)
		if msg == "" {
			msg = strings.TrimSpace(errPayload.Message)
		}
		if msg == "" {
			msg = string(rawResp)
		}
		return fmt.Errorf("biteship returned %d: %s", resp.StatusCode, msg)
	}

	if dst != nil {
		if err := json.Unmarshal(rawResp, dst); err != nil {
			return fmt.Errorf("decode biteship response: %w", err)
		}
	}
	return nil
}
