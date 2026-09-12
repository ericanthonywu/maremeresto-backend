package service

import (
	"context"
	"math"
	"time"

	"github.com/ericanthonywu/maremereso-olga/backend/internal/apperror"
	"github.com/ericanthonywu/maremereso-olga/backend/internal/dto"
	"github.com/ericanthonywu/maremereso-olga/backend/internal/model"
	"github.com/google/uuid"
)

// roadFactor converts straight-line ("as the crow flies") distance into an
// estimate of the distance actually ridden on Solo's road grid.
const roadFactor = 1.3

// Courier speed assumption used for the ETA, in km/h, plus a fixed preparation
// allowance in minutes.
const (
	courierKmPerHour  = 18.0
	preparationMinute = 8
	minBillableKm     = 0.5
	maxServiceableKm  = 25.0
)

// HaversineKm returns the great-circle distance between two WGS-84 coordinates
// in kilometres.
func HaversineKm(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKm = 6371.0

	rad := func(deg float64) float64 { return deg * math.Pi / 180.0 }

	dLat := rad(lat2 - lat1)
	dLon := rad(lon2 - lon1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(rad(lat1))*math.Cos(rad(lat2))*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	return earthRadiusKm * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

// RoadDistanceKm is the distance a courier is expected to ride, rounded to one
// decimal so the figure quoted to the customer and the figure stored on the
// order are byte-identical.
func RoadDistanceKm(lat1, lon1, lat2, lon2 float64) float64 {
	d := HaversineKm(lat1, lon1, lat2, lon2) * roadFactor
	if d < minBillableKm {
		d = minBillableKm
	}
	return math.Round(d*10) / 10
}

// DeliveryFee applies the branch's own tiered pricing. This is the ONE place
// the fee is derived: the quote endpoint and order creation both call it, so a
// customer can never be quoted one price and charged another.
func DeliveryFee(s *model.BranchSettings, distanceKm float64, subtotal int, orderType string) int {
	if orderType == "pickup" {
		return 0
	}
	if s.FreeDeliveryThreshold > 0 && subtotal >= s.FreeDeliveryThreshold {
		return 0
	}

	switch {
	case distanceKm <= float64(s.NearThresholdKm):
		return s.BaseDeliveryFeeNear
	case distanceKm <= float64(s.MidThresholdKm):
		return s.BaseDeliveryFeeMid
	default:
		return s.BaseDeliveryFeeFar
	}
}

// EtaMinutes estimates door-to-door time including preparation.
func EtaMinutes(distanceKm float64) int {
	riding := distanceKm / courierKmPerHour * 60.0
	return preparationMinute + int(math.Ceil(riding))
}

// defaultSettings is used only when a branch has no settings row yet, so a
// missing row degrades to a sane price rather than a zero-rupiah delivery.
func defaultSettings(branchID uuid.UUID) *model.BranchSettings {
	return &model.BranchSettings{
		BranchID:              branchID,
		ServiceFee:            2000,
		BaseDeliveryFeeNear:   8000,
		BaseDeliveryFeeMid:    12000,
		BaseDeliveryFeeFar:    18000,
		NearThresholdKm:       3,
		MidThresholdKm:        7,
		MaxDeliveryRadiusKm:   12,
		MinOrderAmount:        0,
		FreeDeliveryThreshold: 150000,
	}
}

func (s *Service) settingsFor(ctx context.Context, branchID uuid.UUID) *model.BranchSettings {
	set, err := s.repo.FindSettingsByBranchID(ctx, branchID)
	if err != nil || set == nil {
		return defaultSettings(branchID)
	}
	return set
}

// QuoteDelivery prices delivery from every branch to the given coordinate. The
// customer app renders these numbers directly, so no arithmetic is duplicated
// on the client.
func (s *Service) QuoteDelivery(ctx context.Context, req *dto.DeliveryQuoteRequest) (*dto.DeliveryQuoteResponse, error) {
	if req.Lat == 0 && req.Lon == 0 {
		return nil, apperror.ErrLocationRequired
	}
	if req.Lat < -90 || req.Lat > 90 || req.Lon < -180 || req.Lon > 180 {
		return nil, apperror.Invalid("koordinat lokasi tidak valid")
	}

	orderType := req.OrderType
	if orderType == "" {
		orderType = "delivery"
	}

	branches, err := s.repo.ListBranches(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	quotes := make([]dto.BranchDeliveryQuote, 0, len(branches))
	nearestIdx := -1

	for _, b := range branches {
		set := s.settingsFor(ctx, b.ID)
		distance := RoadDistanceKm(req.Lat, req.Lon, b.Latitude, b.Longitude)
		fee := DeliveryFee(set, distance, req.Subtotal, orderType)

		serviceable := distance <= float64(set.MaxDeliveryRadiusKm) && distance <= maxServiceableKm
		openNow := b.IsOpen && IsWithinOperatingHours(set.OperatingHours, now)

		q := dto.BranchDeliveryQuote{
			BranchID:         b.ID,
			BranchSlug:       b.Slug,
			BranchName:       b.Name,
			DistanceKm:       distance,
			DistanceMeters:   int(math.Round(distance * 1000)),
			DeliveryFee:      fee,
			ServiceFee:       set.ServiceFee,
			MinOrderAmount:   0, // Minimum order removed
			EtaMinutes:       EtaMinutes(distance),
			MaxRadiusKm:      set.MaxDeliveryRadiusKm,
			WithinRadius:     serviceable,
			FreeDeliveryFrom: set.FreeDeliveryThreshold,
			IsOpenNow:        openNow,
		}
		quotes = append(quotes, q)

		// Nearest = closest branch that can actually take the order right now;
		// fall back to closest overall so the UI always has something selected.
		if nearestIdx == -1 {
			nearestIdx = len(quotes) - 1
			continue
		}
		best := quotes[nearestIdx]
		bestUsable := best.WithinRadius && best.IsOpenNow
		thisUsable := q.WithinRadius && q.IsOpenNow
		if (thisUsable && !bestUsable) || (thisUsable == bestUsable && q.DistanceKm < best.DistanceKm) {
			nearestIdx = len(quotes) - 1
		}
	}

	resp := &dto.DeliveryQuoteResponse{Quotes: quotes}
	if nearestIdx >= 0 {
		quotes[nearestIdx].IsNearest = true
		resp.NearestBranchID = &quotes[nearestIdx].BranchID
	}
	return resp, nil
}
