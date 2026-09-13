package service

import (
	"errors"
	"testing"

	"github.com/ericanthonywu/maremereso-olga/backend/internal/apperror"
	"github.com/ericanthonywu/maremereso-olga/backend/internal/model"
)

func TestCalculateOrderDelivery(t *testing.T) {
	bSettings := &model.BranchSettings{
		BaseDeliveryFeeNear: 0,
		BaseDeliveryFeeMid:  8000,
		BaseDeliveryFeeFar:  12000,
		NearThresholdKm:     1,
		MidThresholdKm:      5,
		ServiceFee:          2000,
	}

	branch := &model.Branch{
		Latitude:  kertenLat,
		Longitude: kertenLon,
	}

	t.Run("pickup order does not require location and has zero delivery fee", func(t *testing.T) {
		dist, fee, err := calculateOrderDelivery(bSettings, branch, "pickup", nil, nil, "", 50000)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dist != 0 || fee != 0 {
			t.Errorf("got dist=%.1f fee=%d, want dist=0 fee=0", dist, fee)
		}
	})

	t.Run("delivery missing lat/lon returns ErrLocationRequired", func(t *testing.T) {
		_, _, err := calculateOrderDelivery(bSettings, branch, "delivery", nil, nil, "Some Address", 50000)
		if !errors.Is(err, apperror.ErrLocationRequired) {
			t.Errorf("got error %v, want %v", err, apperror.ErrLocationRequired)
		}

		lat, lon := 0.0, 0.0
		_, _, err = calculateOrderDelivery(bSettings, branch, "delivery", &lat, &lon, "Some Address", 50000)
		if !errors.Is(err, apperror.ErrLocationRequired) {
			t.Errorf("got error %v, want %v", err, apperror.ErrLocationRequired)
		}
	})

	t.Run("delivery missing address returns invalid error", func(t *testing.T) {
		lat, lon := makDjanLat, makDjanLon
		_, _, err := calculateOrderDelivery(bSettings, branch, "delivery", &lat, &lon, "   ", 50000)
		if err == nil {
			t.Fatalf("expected error for empty address, got nil")
		}
	})

	t.Run("delivery out of range returns ErrOutOfDeliveryRange", func(t *testing.T) {
		// ~50km away from kertenLat/kertenLon
		farLat, farLon := -7.1, 110.4
		_, _, err := calculateOrderDelivery(bSettings, branch, "delivery", &farLat, &farLon, "Far away", 50000)
		if !errors.Is(err, apperror.ErrOutOfDeliveryRange) {
			t.Errorf("got error %v, want %v", err, apperror.ErrOutOfDeliveryRange)
		}
	})

	t.Run("valid delivery calculates distance and fee correctly", func(t *testing.T) {
		lat, lon := makDjanLat, makDjanLon
		dist, fee, err := calculateOrderDelivery(bSettings, branch, "delivery", &lat, &lon, "Mak Djan Address", 50000)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dist <= 0 {
			t.Errorf("expected positive distance, got %.2f", dist)
		}
		if fee <= 0 {
			t.Errorf("expected positive delivery fee, got %d", fee)
		}
	})
}
