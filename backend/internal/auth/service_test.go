package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"go-academy/backend/internal/auth"
	"go-academy/backend/internal/config"
	"go-academy/backend/internal/database"
	"go-academy/backend/internal/users"
)

func setupTestContext(t *testing.T) (*database.DB, *config.Config, *users.Service, *auth.Service) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, cfg)
	if err != nil {
		t.Skipf("skipping live database test: %v", err)
	}

	db := database.NewDB(pool)
	userService := users.NewService(db)
	authService := auth.NewService(db, cfg, userService)

	return db, cfg, userService, authService
}

func TestAuthService_FullLifecycle(t *testing.T) {
	db, _, userService, authService := setupTestContext(t)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Ensure an admin user exists (from bootstrap)
	adminUser, err := userService.GetUserByEmail(ctx, "test_admin@example.com")
	if err != nil {
		t.Fatalf("expected admin user to exist from bootstrap: %v", err)
	}

	password := "SuperAdminPass123!"

	// 1. Successful Login
	user, accessToken, rawRefreshToken, err := authService.Login(ctx, adminUser.Email, password)
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}
	if user.Email != adminUser.Email {
		t.Fatalf("expected user email %s, got %s", adminUser.Email, user.Email)
	}
	if accessToken == "" || rawRefreshToken == "" {
		t.Fatal("expected non-empty tokens")
	}

	// 2. Failed Login - Wrong Password
	_, _, _, err = authService.Login(ctx, adminUser.Email, "WrongPassword123!")
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got: %v", err)
	}

	// 3. Failed Login - Nonexistent Email
	_, _, _, err = authService.Login(ctx, "nonexistent@example.com", password)
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials for unknown user, got: %v", err)
	}

	// 4. Token Rotation (First Refresh)
	userRefreshed, newAccessToken, newRefreshToken, err := authService.Refresh(ctx, rawRefreshToken)
	if err != nil {
		t.Fatalf("expected successful refresh: %v", err)
	}
	if userRefreshed.ID != user.ID {
		t.Fatalf("expected same user ID, got %s vs %s", user.ID, userRefreshed.ID)
	}
	if newRefreshToken == rawRefreshToken {
		t.Fatal("expected rotated refresh token to be different from original")
	}
	if newAccessToken == accessToken {
		t.Fatal("expected new access token to be generated")
	}

	// 5. Token Reuse Detection
	// Presenting the OLD rawRefreshToken that was already rotated must trigger reuse detection!
	_, _, _, err = authService.Refresh(ctx, rawRefreshToken)
	if !errors.Is(err, auth.ErrTokenReuseDetected) {
		t.Fatalf("expected ErrTokenReuseDetected when presenting burned token, got: %v", err)
	}

	// 6. Family Revocation Check
	// Because reuse was detected, the newest token in the family must ALSO now be revoked!
	_, _, _, err = authService.Refresh(ctx, newRefreshToken)
	if !errors.Is(err, auth.ErrTokenReuseDetected) && !errors.Is(err, auth.ErrSessionNotFound) {
		t.Fatalf("expected session to be invalid after family revocation, got: %v", err)
	}

	// 7. Logout Flow
	// Perform fresh login and test explicit logout
	_, _, logoutRefreshToken, err := authService.Login(ctx, adminUser.Email, password)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	if err := authService.Logout(ctx, logoutRefreshToken); err != nil {
		t.Fatalf("logout failed: %v", err)
	}

	// Subsequent refresh on logged out token should fail
	_, _, _, err = authService.Refresh(ctx, logoutRefreshToken)
	if err == nil {
		t.Fatal("expected refresh to fail after logout")
	}
}

func TestAuthService_DisabledAccountRejection(t *testing.T) {
	db, _, userService, authService := setupTestContext(t)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	adminUser, err := userService.GetUserByEmail(ctx, "test_admin@example.com")
	if err != nil {
		t.Fatalf("failed to fetch user: %v", err)
	}

	// Temporarily disable user
	_, err = userService.UpdateUserStatus(ctx, adminUser.ID, users.StatusDisabled)
	if err != nil {
		t.Fatalf("failed to disable user: %v", err)
	}
	// Restore status on cleanup
	defer func() {
		_, _ = userService.UpdateUserStatus(ctx, adminUser.ID, users.StatusActive)
	}()

	// Login must fail with ErrAccountDisabled
	_, _, _, err = authService.Login(ctx, adminUser.Email, "SuperAdminPass123!")
	if !errors.Is(err, auth.ErrAccountDisabled) {
		t.Fatalf("expected ErrAccountDisabled for disabled account, got: %v", err)
	}
}

func TestAuthService_NonexistentSession(t *testing.T) {
	db, _, _, authService := setupTestContext(t)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fakeToken := uuid.New().String()
	_, _, _, err := authService.Refresh(ctx, fakeToken)
	if !errors.Is(err, auth.ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound for fake token, got: %v", err)
	}
}
