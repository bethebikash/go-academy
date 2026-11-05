package auth_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"go-academy/backend/internal/auth"
	"go-academy/backend/internal/users"
)

func TestAccessTokenGenerationAndValidation(t *testing.T) {
	secret := "test_secret_must_be_sufficiently_long_for_hmac_sha256"
	user := &users.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Username: "tester",
		Role:     users.RoleLearner,
		Status:   users.StatusActive,
	}

	// 1. Valid Token
	tokenStr, err := auth.GenerateAccessToken(user, secret, 15*time.Minute)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	claims, err := auth.ValidateAccessToken(tokenStr, secret)
	if err != nil {
		t.Fatalf("expected valid token, got error: %v", err)
	}

	if claims.UserID != user.ID.String() {
		t.Fatalf("expected user ID %s, got %s", user.ID.String(), claims.UserID)
	}
	if claims.Email != user.Email {
		t.Fatalf("expected email %s, got %s", user.Email, claims.Email)
	}
	if claims.Role != user.Role {
		t.Fatalf("expected role %s, got %s", user.Role, claims.Role)
	}

	// 2. Tampered Secret
	_, err = auth.ValidateAccessToken(tokenStr, "wrong_secret_key_used_to_tamper_signature")
	if !errors.Is(err, auth.ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken for wrong secret, got: %v", err)
	}

	// 3. Expired Token
	expiredToken, err := auth.GenerateAccessToken(user, secret, -1*time.Minute)
	if err != nil {
		t.Fatalf("failed to generate expired token: %v", err)
	}

	_, err = auth.ValidateAccessToken(expiredToken, secret)
	if !errors.Is(err, auth.ErrExpiredToken) {
		t.Fatalf("expected ErrExpiredToken, got: %v", err)
	}
}

func TestOpaqueTokenGeneration(t *testing.T) {
	raw, hash, err := auth.GenerateOpaqueToken()
	if err != nil {
		t.Fatalf("failed to generate opaque token: %v", err)
	}

	if len(raw) == 0 || len(hash) != 64 {
		t.Fatalf("unexpected token lengths: raw=%d, hash=%d", len(raw), len(hash))
	}

	recomputedHash := auth.HashOpaqueToken(raw)
	if recomputedHash != hash {
		t.Fatalf("expected recomputed hash to match: %s vs %s", recomputedHash, hash)
	}
}
