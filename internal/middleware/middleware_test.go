package middleware

import (
	"testing"
	"time"

	"github.com/ericanthonywu/maremereso-olga/backend/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestParseToken(t *testing.T) {
	secret := "super-secret-key-that-is-at-least-32-chars-long!"
	cfg := &config.Config{
		JWTSecret: secret,
	}

	userID := uuid.New()
	branchID := uuid.New()

	t.Run("valid HS256 token with all claims", func(t *testing.T) {
		claims := &JWTClaims{
			UserID:   userID,
			Phone:    "+6281234567890",
			Role:     "admin",
			BranchID: &branchID,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, err := token.SignedString([]byte(secret))
		if err != nil {
			t.Fatalf("failed to sign token: %v", err)
		}

		parsedClaims, err := ParseToken(cfg, tokenStr)
		if err != nil {
			t.Fatalf("expected valid token, got error: %v", err)
		}
		if parsedClaims.UserID != userID {
			t.Errorf("got UserID %v, want %v", parsedClaims.UserID, userID)
		}
		if parsedClaims.Phone != "+6281234567890" {
			t.Errorf("got Phone %q, want %q", parsedClaims.Phone, "+6281234567890")
		}
		if parsedClaims.Role != "admin" {
			t.Errorf("got Role %q, want %q", parsedClaims.Role, "admin")
		}
		if parsedClaims.BranchID == nil || *parsedClaims.BranchID != branchID {
			t.Errorf("got BranchID %v, want %v", parsedClaims.BranchID, branchID)
		}
	})

	t.Run("valid HS256 token without optional BranchID", func(t *testing.T) {
		claims := &JWTClaims{
			UserID: userID,
			Phone:  "+6281234567890",
			Role:   "customer",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, err := token.SignedString([]byte(secret))
		if err != nil {
			t.Fatalf("failed to sign token: %v", err)
		}

		parsedClaims, err := ParseToken(cfg, tokenStr)
		if err != nil {
			t.Fatalf("expected valid token, got error: %v", err)
		}
		if parsedClaims.BranchID != nil {
			t.Errorf("expected nil BranchID, got %v", parsedClaims.BranchID)
		}
	})

	t.Run("token signed with wrong secret", func(t *testing.T) {
		claims := &JWTClaims{
			UserID: userID,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, err := token.SignedString([]byte("wrong-secret-key-which-is-different!"))
		if err != nil {
			t.Fatalf("failed to sign token: %v", err)
		}

		_, err = ParseToken(cfg, tokenStr)
		if err == nil {
			t.Error("expected error for token signed with wrong secret, got nil")
		}
	})

	t.Run("token signed with unsupported signing method HS384", func(t *testing.T) {
		claims := &JWTClaims{
			UserID: userID,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
		tokenStr, err := token.SignedString([]byte(secret))
		if err != nil {
			t.Fatalf("failed to sign token: %v", err)
		}

		_, err = ParseToken(cfg, tokenStr)
		if err == nil {
			t.Error("expected error for token signed with HS384, got nil")
		}
	})

	t.Run("token signed with unsupported signing method HS512", func(t *testing.T) {
		claims := &JWTClaims{
			UserID: userID,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
		tokenStr, err := token.SignedString([]byte(secret))
		if err != nil {
			t.Fatalf("failed to sign token: %v", err)
		}

		_, err = ParseToken(cfg, tokenStr)
		if err == nil {
			t.Error("expected error for token signed with HS512, got nil")
		}
	})

	t.Run("token with 'none' signing method", func(t *testing.T) {
		claims := &JWTClaims{
			UserID: userID,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
		tokenStr, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
		if err != nil {
			t.Fatalf("failed to sign token: %v", err)
		}

		_, err = ParseToken(cfg, tokenStr)
		if err == nil {
			t.Error("expected error for token with 'none' alg, got nil")
		}
	})

	t.Run("expired token", func(t *testing.T) {
		claims := &JWTClaims{
			UserID: userID,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, err := token.SignedString([]byte(secret))
		if err != nil {
			t.Fatalf("failed to sign token: %v", err)
		}

		_, err = ParseToken(cfg, tokenStr)
		if err == nil {
			t.Error("expected error for expired token, got nil")
		}
	})

	t.Run("malformed tokens", func(t *testing.T) {
		malformedInputs := []string{
			"",
			"not.a.token",
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.signature",
			"abc123xyz",
		}
		for _, input := range malformedInputs {
			_, err := ParseToken(cfg, input)
			if err == nil {
				t.Errorf("expected error for malformed token %q, got nil", input)
			}
		}
	})
}
