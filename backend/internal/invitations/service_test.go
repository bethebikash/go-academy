package invitations_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"go-academy/backend/internal/config"
	"go-academy/backend/internal/database"
	"go-academy/backend/internal/invitations"
	"go-academy/backend/internal/users"
)

func setupInvitationTest(t *testing.T) (*database.DB, *config.Config, *users.Service, *invitations.Service, *users.User) {
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
	invitationService := invitations.NewService(db, cfg, userService)

	adminUser, err := userService.GetUserByEmail(ctx, "test_admin@example.com")
	if err != nil {
		t.Fatalf("expected admin user to exist: %v", err)
	}

	return db, cfg, userService, invitationService, adminUser
}

func TestInvitation_FullLifecycleAndDoubleAcceptance(t *testing.T) {
	db, _, _, invitationService, adminUser := setupInvitationTest(t)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now().UnixNano()
	email := fmt.Sprintf("friend_invitee_%d@example.com", now)

	// 1. Create Invitation
	inv, err := invitationService.CreateInvitation(ctx, adminUser.ID, email, 7)
	if err != nil {
		t.Fatalf("failed to create invitation: %v", err)
	}
	if inv.Email != email {
		t.Fatalf("expected email %s, got %s", email, inv.Email)
	}
	if inv.RawToken == "" {
		t.Fatal("expected non-empty raw invitation token")
	}
	if inv.Status != invitations.StatusPending {
		t.Fatalf("expected status pending, got %s", inv.Status)
	}

	// 2. Validate token preview
	preview, err := invitationService.ValidateInvitationToken(ctx, inv.RawToken)
	if err != nil {
		t.Fatalf("token validation failed: %v", err)
	}
	if preview.Email != email {
		t.Fatalf("expected preview email %s, got %s", email, preview.Email)
	}

	// 3. Prevent duplicate invite for already registered user
	_, err = invitationService.CreateInvitation(ctx, adminUser.ID, adminUser.Email, 7)
	if !errors.Is(err, invitations.ErrEmailAlreadyRegistered) {
		t.Fatalf("expected ErrEmailAlreadyRegistered, got: %v", err)
	}

	// 4. Accept Invitation and Register
	username := fmt.Sprintf("friend_%d", now%1000000)
	password := "SecureP@ssword123!"

	user, accessToken, rawRefreshToken, err := invitationService.AcceptInvitationAndRegister(
		ctx,
		inv.RawToken,
		username,
		password,
	)
	if err != nil {
		t.Fatalf("expected registration to succeed: %v", err)
	}
	if user.Email != email {
		t.Fatalf("expected user email %s, got %s", email, user.Email)
	}
	if user.Role != users.RoleLearner {
		t.Fatalf("expected learner role, got %s", user.Role)
	}
	if accessToken == "" || rawRefreshToken == "" {
		t.Fatal("expected non-empty access and refresh tokens")
	}

	// 5. Test Double Acceptance Rejection (Single-use invariant)
	_, _, _, err = invitationService.AcceptInvitationAndRegister(
		ctx,
		inv.RawToken,
		"another_username",
		"AnotherPass123!",
	)
	if !errors.Is(err, invitations.ErrInvitationAlreadyUsed) {
		t.Fatalf("expected ErrInvitationAlreadyUsed on second acceptance, got: %v", err)
	}

	// 6. Token validation on already consumed invitation
	_, err = invitationService.ValidateInvitationToken(ctx, inv.RawToken)
	if !errors.Is(err, invitations.ErrInvitationAlreadyUsed) {
		t.Fatalf("expected ErrInvitationAlreadyUsed during token validation, got: %v", err)
	}
}

func TestInvitation_Revocation(t *testing.T) {
	db, _, _, invitationService, adminUser := setupInvitationTest(t)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	email := fmt.Sprintf("friend_revoke_%d@example.com", time.Now().UnixNano())
	inv, err := invitationService.CreateInvitation(ctx, adminUser.ID, email, 7)
	if err != nil {
		t.Fatalf("failed to create invitation: %v", err)
	}

	// Revoke invitation
	revoked, err := invitationService.RevokeInvitation(ctx, inv.ID)
	if err != nil {
		t.Fatalf("failed to revoke invitation: %v", err)
	}
	if revoked.Status != invitations.StatusRevoked {
		t.Fatalf("expected revoked status, got %s", revoked.Status)
	}

	// Validation must fail with ErrInvitationRevoked
	_, err = invitationService.ValidateInvitationToken(ctx, inv.RawToken)
	if !errors.Is(err, invitations.ErrInvitationRevoked) {
		t.Fatalf("expected ErrInvitationRevoked, got: %v", err)
	}

	// Registration attempt must fail with ErrInvitationRevoked
	_, _, _, err = invitationService.AcceptInvitationAndRegister(ctx, inv.RawToken, "revoke_usr", "SecureP@ssword123!")
	if !errors.Is(err, invitations.ErrInvitationRevoked) {
		t.Fatalf("expected ErrInvitationRevoked on registration attempt, got: %v", err)
	}
}

func TestInvitation_ConcurrentAcceptanceSafety(t *testing.T) {
	db, _, _, invitationService, adminUser := setupInvitationTest(t)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now().UnixNano()
	email := fmt.Sprintf("concurrent_%d@example.com", now)
	inv, err := invitationService.CreateInvitation(ctx, adminUser.ID, email, 7)
	if err != nil {
		t.Fatalf("failed to create invitation: %v", err)
	}

	// Concurrently attempt to accept the same invitation
	var wg sync.WaitGroup
	results := make([]error, 2)

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, _, _, regErr := invitationService.AcceptInvitationAndRegister(
				context.Background(),
				inv.RawToken,
				fmt.Sprintf("cuser_%d", now%1000000),
				"Password123!Safe",
			)
			results[idx] = regErr
		}(i)
	}

	wg.Wait()

	// Exactly ONE must succeed (nil error), and the other must fail with ErrInvitationAlreadyUsed or ErrUsernameTaken
	successCount := 0
	failureCount := 0
	for _, res := range results {
		if res == nil {
			successCount++
		} else {
			failureCount++
		}
	}

	if successCount != 1 || failureCount != 1 {
		t.Fatalf("expected exactly 1 success and 1 failure in concurrent acceptance, got %d successes and %d failures (%v)",
			successCount, failureCount, results)
	}
}
