package invitations

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"go-academy/backend/internal/auth"
	"go-academy/backend/internal/config"
	"go-academy/backend/internal/database"
	"go-academy/backend/internal/database/sqlc"
	"go-academy/backend/internal/platform/hasher"
	"go-academy/backend/internal/users"
)

// Service provides invitation lifecycle management and transactional onboarding logic.
type Service struct {
	db          *database.DB
	cfg         *config.Config
	userService *users.Service
}

// NewService constructs a new invitations Service.
func NewService(db *database.DB, cfg *config.Config, userService *users.Service) *Service {
	return &Service{
		db:          db,
		cfg:         cfg,
		userService: userService,
	}
}

// CreateInvitation generates a new cryptographic single-use invitation record.
func (s *Service) CreateInvitation(ctx context.Context, adminID uuid.UUID, email string, expiresInDays int) (*Invitation, error) {
	trimmedEmail := strings.ToLower(strings.TrimSpace(email))
	if _, err := mail.ParseAddress(trimmedEmail); err != nil {
		return nil, ErrInvalidEmail
	}

	// Verify the invited email is not already a registered account
	if _, err := s.userService.GetUserByEmail(ctx, trimmedEmail); err == nil {
		return nil, ErrEmailAlreadyRegistered
	} else if !errors.Is(err, users.ErrUserNotFound) {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	if expiresInDays <= 0 || expiresInDays > 30 {
		expiresInDays = 7
	}

	expiresAt := time.Now().UTC().Add(time.Duration(expiresInDays) * 24 * time.Hour)

	// Generate high-entropy 32-byte cryptographic token
	rawToken, tokenHash, err := auth.GenerateOpaqueToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate invitation token: %w", err)
	}

	row, err := s.db.Queries.CreateInvitation(ctx, sqlc.CreateInvitationParams{
		Email:     trimmedEmail,
		TokenHash: tokenHash,
		InvitedBy: pgtype.UUID{Bytes: adminID, Valid: true},
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to persist invitation: %w", err)
	}

	invID := uuid.UUID(row.ID.Bytes)
	invURL := fmt.Sprintf("%s/register?token=%s", s.cfg.FrontendURL, rawToken)

	return &Invitation{
		ID:            invID,
		Email:         row.Email,
		Status:        StatusPending,
		InvitedBy:     adminID,
		ExpiresAt:     row.ExpiresAt.Time,
		CreatedAt:     row.CreatedAt.Time,
		InvitationURL: invURL,
		RawToken:      rawToken,
	}, nil
}

// ListInvitations lists invitations with pagination and computed status.
func (s *Service) ListInvitations(ctx context.Context, page, limit int) ([]*Invitation, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	total, err := s.db.Queries.CountInvitations(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count invitations: %w", err)
	}

	rows, err := s.db.Queries.ListInvitations(ctx, sqlc.ListInvitationsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list invitations: %w", err)
	}

	items := make([]*Invitation, 0, len(rows))
	for _, r := range rows {
		var acceptedAt, revokedAt *time.Time
		if r.AcceptedAt.Valid {
			t := r.AcceptedAt.Time
			acceptedAt = &t
		}
		if r.RevokedAt.Valid {
			t := r.RevokedAt.Time
			revokedAt = &t
		}

		status := ComputeStatus(r.ExpiresAt.Time, acceptedAt, revokedAt)

		items = append(items, &Invitation{
			ID:                uuid.UUID(r.ID.Bytes),
			Email:             r.Email,
			Status:            status,
			InvitedBy:         uuid.UUID(r.InvitedBy.Bytes),
			InvitedByUsername: r.InvitedByUsername,
			ExpiresAt:         r.ExpiresAt.Time,
			AcceptedAt:        acceptedAt,
			RevokedAt:         revokedAt,
			CreatedAt:         r.CreatedAt.Time,
		})
	}

	return items, total, nil
}

// RevokeInvitation revokes an active invitation.
func (s *Service) RevokeInvitation(ctx context.Context, id uuid.UUID) (*Invitation, error) {
	pgID := pgtype.UUID{Bytes: id, Valid: true}
	row, err := s.db.Queries.RevokeInvitation(ctx, pgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidInvitation
		}
		return nil, fmt.Errorf("failed to revoke invitation: %w", err)
	}

	return &Invitation{
		ID:        id,
		Email:     row.Email,
		Status:    StatusRevoked,
		RevokedAt: &row.RevokedAt.Time,
	}, nil
}

// ValidateInvitationToken checks token validity without consuming it (used by registration UI).
func (s *Service) ValidateInvitationToken(ctx context.Context, rawToken string) (*Invitation, error) {
	if rawToken == "" {
		return nil, ErrInvalidInvitation
	}

	tokenHash := auth.HashOpaqueToken(rawToken)
	inv, err := s.db.Queries.GetInvitationByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidInvitation
		}
		return nil, fmt.Errorf("failed to query invitation: %w", err)
	}

	var acceptedAt, revokedAt *time.Time
	if inv.AcceptedAt.Valid {
		t := inv.AcceptedAt.Time
		acceptedAt = &t
	}
	if inv.RevokedAt.Valid {
		t := inv.RevokedAt.Time
		revokedAt = &t
	}

	status := ComputeStatus(inv.ExpiresAt.Time, acceptedAt, revokedAt)
	if status == StatusRevoked {
		return nil, ErrInvitationRevoked
	}
	if status == StatusAccepted {
		return nil, ErrInvitationAlreadyUsed
	}
	if status == StatusExpired {
		return nil, ErrInvitationExpired
	}

	return &Invitation{
		ID:         uuid.UUID(inv.ID.Bytes),
		Email:      inv.Email,
		Status:     status,
		ExpiresAt:  inv.ExpiresAt.Time,
		AcceptedAt: acceptedAt,
		RevokedAt:  revokedAt,
		CreatedAt:  inv.CreatedAt.Time,
	}, nil
}

// AcceptInvitationAndRegister consumes an invitation and registers a new learner in an atomic transaction.
func (s *Service) AcceptInvitationAndRegister(
	ctx context.Context,
	rawToken, username, password string,
) (*users.User, string, string, error) {
	if rawToken == "" {
		return nil, "", "", ErrInvalidInvitation
	}

	tokenHash := auth.HashOpaqueToken(rawToken)

	// Begin atomic transaction
	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.db.Queries.WithTx(tx)

	// Row-level lock on invitation to prevent race conditions & double acceptance
	inv, err := qtx.GetInvitationByTokenHashForUpdate(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", "", ErrInvalidInvitation
		}
		return nil, "", "", fmt.Errorf("failed to acquire invitation lock: %w", err)
	}

	// Validate status inside lock
	if inv.RevokedAt.Valid {
		return nil, "", "", ErrInvitationRevoked
	}
	if inv.AcceptedAt.Valid {
		return nil, "", "", ErrInvitationAlreadyUsed
	}
	if inv.ExpiresAt.Time.Before(time.Now().UTC()) {
		return nil, "", "", ErrInvitationExpired
	}

	// Validate credentials
	cleanUsername := strings.TrimSpace(username)
	if err := users.ValidateRegistrationCredentials(inv.Email, cleanUsername, password); err != nil {
		return nil, "", "", err
	}

	// Verify username not taken
	if _, err := qtx.GetUserByUsername(ctx, cleanUsername); err == nil {
		return nil, "", "", users.ErrUsernameTaken
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, "", "", fmt.Errorf("failed to check username: %w", err)
	}

	// Hash password with Argon2id
	hashedPassword, err := hasher.HashPassword(password)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user record
	createdUser, err := qtx.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        inv.Email,
		Username:     cleanUsername,
		PasswordHash: hashedPassword,
		Status:       users.StatusActive,
		Timezone:     "UTC",
	})
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to create user: %w", err)
	}

	// Retrieve learner role
	learnerRole, err := qtx.GetRoleByName(ctx, users.RoleLearner)
	if err != nil {
		return nil, "", "", fmt.Errorf("learner role not found: %w", err)
	}

	// Assign learner role
	if err := qtx.AssignRoleToUser(ctx, sqlc.AssignRoleToUserParams{
		UserID: createdUser.ID,
		RoleID: learnerRole.ID,
	}); err != nil {
		return nil, "", "", fmt.Errorf("failed to assign learner role: %w", err)
	}

	// Mark invitation accepted
	if err := qtx.MarkInvitationAccepted(ctx, inv.ID); err != nil {
		return nil, "", "", fmt.Errorf("failed to mark invitation accepted: %w", err)
	}

	// Commit atomic registration
	if err := tx.Commit(ctx); err != nil {
		return nil, "", "", fmt.Errorf("failed to commit registration: %w", err)
	}

	userID := uuid.UUID(createdUser.ID.Bytes)
	registeredUser := &users.User{
		ID:        userID,
		Email:     createdUser.Email,
		Username:  createdUser.Username,
		Role:      users.RoleLearner,
		Status:    createdUser.Status,
		Timezone:  createdUser.Timezone,
		CreatedAt: createdUser.CreatedAt.Time,
		UpdatedAt: createdUser.UpdatedAt.Time,
	}

	// Generate access token
	accessToken, err := auth.GenerateAccessToken(registeredUser, s.cfg.JWTSecret, s.cfg.JWTAccessExpiry)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to create access token: %w", err)
	}

	// Establish initial rotating refresh session
	rawRefreshToken, refreshTokenHash, err := auth.GenerateOpaqueToken()
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to create refresh token: %w", err)
	}

	tokenFamilyID := uuid.New()
	sessionExpiresAt := time.Now().UTC().Add(s.cfg.RefreshTokenExpiry)

	_, err = s.db.Queries.CreateRefreshSession(ctx, sqlc.CreateRefreshSessionParams{
		UserID:        createdUser.ID,
		TokenHash:     refreshTokenHash,
		TokenFamilyID: pgtype.UUID{Bytes: tokenFamilyID, Valid: true},
		ExpiresAt:     pgtype.Timestamptz{Time: sessionExpiresAt, Valid: true},
	})
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to store initial refresh session: %w", err)
	}

	return registeredUser, accessToken, rawRefreshToken, nil
}
