package users_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"go-academy/backend/internal/config"
	"go-academy/backend/internal/database"
	"go-academy/backend/internal/platform/hasher"
	"go-academy/backend/internal/users"
)

func TestValidateRegistrationCredentials(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		username    string
		password    string
		expectError bool
	}{
		{
			name:        "valid credentials",
			email:       "admin@example.com",
			username:    "admin_user",
			password:    "SecureP@ss1234",
			expectError: false,
		},
		{
			name:        "invalid email format",
			email:       "invalid-email",
			username:    "admin_user",
			password:    "SecureP@ss1234",
			expectError: true,
		},
		{
			name:        "username too short",
			email:       "valid@example.com",
			username:    "ad",
			password:    "SecureP@ss1234",
			expectError: true,
		},
		{
			name:        "username with illegal characters",
			email:       "valid@example.com",
			username:    "admin user!",
			password:    "SecureP@ss1234",
			expectError: true,
		},
		{
			name:        "password too short",
			email:       "valid@example.com",
			username:    "admin_user",
			password:    "Short1!",
			expectError: true,
		},
		{
			name:        "password missing numbers/symbols",
			email:       "valid@example.com",
			username:    "admin_user",
			password:    "OnlyLettersHere",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := users.ValidateRegistrationCredentials(tc.email, tc.username, tc.password)
			if tc.expectError && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("expected success, got: %v", err)
			}
		})
	}
}

func TestBootstrapAdminIntegration(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load configuration: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, cfg)
	if err != nil {
		t.Skipf("skipping live database test: %v", err)
		return
	}
	defer pool.Close()

	db := database.NewDB(pool)
	service := users.NewService(db)

	adminCount, err := service.CountAdmins(ctx)
	if err != nil {
		t.Fatalf("failed to query admin count: %v", err)
	}

	testEmail := "test_admin@example.com"
	testUsername := "test_admin"
	testPassword := "SuperAdminPass123!"

	if adminCount == 0 {
		// First bootstrap should succeed
		admin, err := service.BootstrapAdmin(ctx, testEmail, testUsername, testPassword)
		if err != nil {
			t.Fatalf("expected bootstrap to succeed, got: %v", err)
		}

		if admin.Email != testEmail {
			t.Fatalf("expected email %q, got %q", testEmail, admin.Email)
		}
		if admin.Role != users.RoleAdmin {
			t.Fatalf("expected role %q, got %q", users.RoleAdmin, admin.Role)
		}
		if admin.Status != users.StatusActive {
			t.Fatalf("expected status %q, got %q", users.StatusActive, admin.Status)
		}

		// Verify password hash in DB
		fetched, err := service.GetUserByID(ctx, admin.ID)
		if err != nil {
			t.Fatalf("failed to fetch user by ID: %v", err)
		}
		valid, err := hasher.VerifyPassword(testPassword, fetched.PasswordHash)
		if err != nil || !valid {
			t.Fatalf("password verification against stored hash failed: %v", err)
		}
	}

	// Repeated bootstrap must be rejected with ErrAdminAlreadyExists
	_, err = service.BootstrapAdmin(ctx, "another@example.com", "another_admin", "AnotherSecurePass123!")
	if !errors.Is(err, users.ErrAdminAlreadyExists) {
		t.Fatalf("expected ErrAdminAlreadyExists on second bootstrap, got: %v", err)
	}
}
