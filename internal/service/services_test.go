package service_test

import (
	"testing"
	"time"

	"github.com/ericanthonywu/maremereso-olga/backend/internal/model"
)

func TestPromoRedeemable(t *testing.T) {
	now := time.Now()
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)
	maxRedemptions := 5

	promo := model.Promo{
		IsActive:        true,
		ValidFrom:       &past,
		ValidUntil:      &future,
		MaxRedemptions:  &maxRedemptions,
		RedemptionCount: 2,
	}

	if !promo.Redeemable(now) {
		t.Errorf("expected promo to be redeemable")
	}

	// Test max redemptions reached
	promo.RedemptionCount = 5
	if promo.Redeemable(now) {
		t.Errorf("expected promo not to be redeemable when max redemptions reached")
	}

	// Test inactive
	promo.RedemptionCount = 2
	promo.IsActive = false
	if promo.Redeemable(now) {
		t.Errorf("expected inactive promo not to be redeemable")
	}
}
