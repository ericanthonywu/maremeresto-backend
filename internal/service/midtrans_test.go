package service

import (
	"crypto/sha512"
	"encoding/hex"
	"testing"
	"time"
)

func TestVerifyMidtransSignature(t *testing.T) {
	orderID := "ORDER-1001"
	statusCode := "200"
	grossAmount := "50000.00"
	serverKey := "SB-Mid-server-KEY-12345"

	// Construct valid signature
	hashInput := orderID + statusCode + grossAmount + serverKey
	expectedHash := sha512.Sum512([]byte(hashInput))
	validSignature := hex.EncodeToString(expectedHash[:])

	tests := []struct {
		name         string
		orderID      string
		statusCode   string
		grossAmount  string
		sigKey       string
		serverKey    string
		want         bool
	}{
		{
			name:        "valid signature",
			orderID:     orderID,
			statusCode:  statusCode,
			grossAmount: grossAmount,
			sigKey:      validSignature,
			serverKey:   serverKey,
			want:        true,
		},
		{
			name:        "valid signature uppercase",
			orderID:     orderID,
			statusCode:  statusCode,
			grossAmount: grossAmount,
			sigKey:      hex.EncodeToString(expectedHash[:]), // ConstantTimeCompare lowers input strings
			serverKey:   serverKey,
			want:        true,
		},
		{
			name:        "invalid signature",
			orderID:     orderID,
			statusCode:  statusCode,
			grossAmount: grossAmount,
			sigKey:      "invalid_signature_hash_value",
			serverKey:   serverKey,
			want:        false,
		},
		{
			name:        "wrong server key",
			orderID:     orderID,
			statusCode:  statusCode,
			grossAmount: grossAmount,
			sigKey:      validSignature,
			serverKey:   "WRONG-SERVER-KEY",
			want:        false,
		},
		{
			name:        "wrong order ID",
			orderID:     "ORDER-9999",
			statusCode:  statusCode,
			grossAmount: grossAmount,
			sigKey:      validSignature,
			serverKey:   serverKey,
			want:        false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := verifyMidtransSignature(tc.orderID, tc.statusCode, tc.grossAmount, tc.sigKey, tc.serverKey)
			if got != tc.want {
				t.Errorf("verifyMidtransSignature() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDeterminePaymentStatus(t *testing.T) {
	now := time.Date(2026, 3, 30, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name              string
		txStatus          string
		fraudStatus       string
		wantPaymentStatus string
		wantPaidAt        bool
	}{
		{
			name:              "capture with accept fraud status",
			txStatus:          "capture",
			fraudStatus:       "accept",
			wantPaymentStatus: "settlement",
			wantPaidAt:        true,
		},
		{
			name:              "capture with empty fraud status",
			txStatus:          "capture",
			fraudStatus:       "",
			wantPaymentStatus: "settlement",
			wantPaidAt:        true,
		},
		{
			name:              "capture with deny fraud status",
			txStatus:          "capture",
			fraudStatus:       "deny",
			wantPaymentStatus: "pending",
			wantPaidAt:        false,
		},
		{
			name:              "capture with challenge fraud status",
			txStatus:          "capture",
			fraudStatus:       "challenge",
			wantPaymentStatus: "pending",
			wantPaidAt:        false,
		},
		{
			name:              "settlement status",
			txStatus:          "settlement",
			fraudStatus:       "",
			wantPaymentStatus: "settlement",
			wantPaidAt:        true,
		},
		{
			name:              "expire status",
			txStatus:          "expire",
			fraudStatus:       "",
			wantPaymentStatus: "expire",
			wantPaidAt:        false,
		},
		{
			name:              "cancel status",
			txStatus:          "cancel",
			fraudStatus:       "",
			wantPaymentStatus: "cancel",
			wantPaidAt:        false,
		},
		{
			name:              "deny status",
			txStatus:          "deny",
			fraudStatus:       "",
			wantPaymentStatus: "cancel",
			wantPaidAt:        false,
		},
		{
			name:              "failure status",
			txStatus:          "failure",
			fraudStatus:       "",
			wantPaymentStatus: "cancel",
			wantPaidAt:        false,
		},
		{
			name:              "pending status",
			txStatus:          "pending",
			fraudStatus:       "",
			wantPaymentStatus: "pending",
			wantPaidAt:        false,
		},
		{
			name:              "unhandled status defaults to pending",
			txStatus:          "unknown_status_xyz",
			fraudStatus:       "",
			wantPaymentStatus: "pending",
			wantPaidAt:        false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			status, paidAt := determinePaymentStatus(tc.txStatus, tc.fraudStatus, now)
			if status != tc.wantPaymentStatus {
				t.Errorf("determinePaymentStatus() status = %v, want %v", status, tc.wantPaymentStatus)
			}
			if tc.wantPaidAt {
				if paidAt == nil || *paidAt != now {
					t.Errorf("determinePaymentStatus() paidAt = %v, want %v", paidAt, now)
				}
			} else {
				if paidAt != nil {
					t.Errorf("determinePaymentStatus() paidAt = %v, want nil", paidAt)
				}
			}
		})
	}
}
