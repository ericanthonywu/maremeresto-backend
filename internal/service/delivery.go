package service

import (
	"context"
	"log/slog"
	"math"
	"sync"
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
	maxServiceableKm  = 10.0

	// Greater Surakarta (Solo Raya) bounding box matching soloViewbox in geocode.go.
	soloMinLat = -7.72
	soloMaxLat = -7.42
	soloMinLon = 110.65
	soloMaxLon = 110.95
)

// IsWithinSoloArea reports whether a WGS-84 coordinate lies within Greater
// Surakarta (Solo Raya). Delivery and scheduled orders outside this bounding
// box are always out of range.
func IsWithinSoloArea(lat, lon float64) bool {
	return lat >= soloMinLat && lat <= soloMaxLat && lon >= soloMinLon && lon <= soloMaxLon
}

// effectiveMaxRadiusKm returns the branch's configured max delivery radius
// capped by the hard platform maximum (maxServiceableKm = 10 km).
func effectiveMaxRadiusKm(set *model.BranchSettings) float64 {
	if set != nil && set.MaxDeliveryRadiusKm > 0 && float64(set.MaxDeliveryRadiusKm) < maxServiceableKm {
		return float64(set.MaxDeliveryRadiusKm)
	}
	return maxServiceableKm
}

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

// DeliveryFee applies the fixed delivery policy used by every outlet. This is
// the ONE place the fee is derived: quotes and order creation therefore always
// agree. The subtotal and settings stay in the signature for compatibility
// with callers, but delivery pricing is intentionally no longer editable per
// branch.
func DeliveryFee(_ *model.BranchSettings, distanceKm float64, _ int, orderType string) int {
	if orderType == "pickup" {
		return 0
	}

	switch {
	case distanceKm <= 1:
		return 0
	case distanceKm <= 5:
		return 8000
	default:
		return 12000
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
		BaseDeliveryFeeNear:   0,
		BaseDeliveryFeeMid:    8000,
		BaseDeliveryFeeFar:    12000,
		NearThresholdKm:       1,
		MidThresholdKm:        5,
		MaxDeliveryRadiusKm:   10,
		MinOrderAmount:        0,
		FreeDeliveryThreshold: 0,
	}
}

// settingsFor returns the branch's settings, or the defaults when it has none.
// A lookup failure is returned rather than masked: silently pricing an order
// with default fees would charge the customer the wrong amount.
func (s *Service) settingsFor(ctx context.Context, branchID uuid.UUID) (*model.BranchSettings, error) {
	set, err := s.repo.FindSettingsByBranchID(ctx, branchID)
	if err != nil {
		return nil, err
	}
	if set == nil {
		return defaultSettings(branchID), nil
	}
	return set, nil
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
	withinSolo := IsWithinSoloArea(req.Lat, req.Lon)
	quotes := make([]dto.BranchDeliveryQuote, len(branches))
	serviceableFlags := make([]bool, len(branches))

	for i, b := range branches {
		set, err := s.settingsFor(ctx, b.ID)
		if err != nil {
			return nil, err
		}
		distance := RoadDistanceKm(req.Lat, req.Lon, b.Latitude, b.Longitude)
		fee := DeliveryFee(set, distance, req.Subtotal, orderType)

		maxRadius := effectiveMaxRadiusKm(set)
		serviceable := withinSolo && distance <= maxRadius
		serviceableFlags[i] = serviceable
		openNow := b.IsOpen && IsWithinOperatingHours(set.OperatingHours, now)

		quotes[i] = dto.BranchDeliveryQuote{
			BranchID:         b.ID,
			BranchSlug:       b.Slug,
			BranchName:       b.Name,
			BranchAddress:    b.Address,
			DistanceKm:       distance,
			DistanceMeters:   int(math.Round(distance * 1000)),
			DeliveryFee:      fee,
			ServiceFee:       set.ServiceFee,
			MinOrderAmount:   0, // Minimum order removed
			EtaMinutes:       EtaMinutes(distance),
			MaxRadiusKm:      int(maxRadius),
			WithinRadius:     serviceable,
			WithinSolo:       withinSolo,
			FreeDeliveryFrom: 0,
			IsOpenNow:        openNow,
		}
	}

	// Fetch real-time GoSend & Grab instant rates concurrently for serviceable branches.
	if orderType != "pickup" && s.biteship.Enabled() && withinSolo {
		rateCtx, cancel := context.WithTimeout(ctx, 6*time.Second)
		defer cancel()
		itemValue := req.Subtotal
		if itemValue < 10000 {
			itemValue = 25000
		}
		rateItems := []BiteshipRateItem{{
			Name:     "Pesanan Makanan & Minuman",
			Value:    itemValue,
			Quantity: 1,
			Weight:   500,
		}}

		var wg sync.WaitGroup
		for i, b := range branches {
			if !serviceableFlags[i] {
				continue
			}
			wg.Add(1)
			go func(idx int, br model.Branch) {
				defer wg.Done()
				opts, err := s.biteship.GetInstantRates(rateCtx, br.Latitude, br.Longitude, req.Lat, req.Lon, rateItems)
				if err != nil {
					slog.WarnContext(rateCtx, "biteship quote failed", "branch", br.Slug, "err", err)
					quotes[idx].CourierOptions = []dto.CourierOption{}
					quotes[idx].DeliveryFee = 0
					return
				}
				quotes[idx].CourierOptions = opts
				if len(opts) > 0 {
					quotes[idx].DeliveryFee = opts[0].Price
				} else {
					quotes[idx].DeliveryFee = 0
				}
			}(i, b)
		}
		wg.Wait()
	}

	nearestIdx := -1
	for i := range quotes {
		q := quotes[i]

		// Nearest = closest branch within radius that can actually take the order right now;
		// fall back to closest within radius even if closed.
		if !q.WithinRadius {
			continue
		}
		if nearestIdx == -1 {
			nearestIdx = i
			continue
		}
		best := quotes[nearestIdx]
		bestUsable := best.WithinRadius && best.IsOpenNow
		thisUsable := q.WithinRadius && q.IsOpenNow
		if (thisUsable && !bestUsable) || (thisUsable == bestUsable && q.DistanceKm < best.DistanceKm) {
			nearestIdx = i
		}
	}

	resp := &dto.DeliveryQuoteResponse{
		Quotes:     quotes,
		WithinSolo: withinSolo,
	}
	if nearestIdx >= 0 && quotes[nearestIdx].WithinRadius {
		quotes[nearestIdx].IsNearest = true
		resp.NearestBranchID = &quotes[nearestIdx].BranchID
	}
	return resp, nil
}
