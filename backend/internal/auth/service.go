package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"go-academy/backend/internal/config"
	"go-academy/backend/internal/database"
	"go-academy/backend/internal/database/sqlc"
	"go-academy/backend/internal/platform/hasher"
	"go-academy/backend/internal/users"
)

var (
	// ErrInvalidCredentials indicates incorrect email or password.
	ErrInvalidCredentials = errors.New("invalid email or password")
	// ErrAccountDisabled indicates the account is deactivated.
	ErrAccountDisabled = errors.New("account is disabled")
	// ErrSessionNotFound indicates no matching active refresh session exists.
	ErrSessionNotFound = errors.New("refresh session not found or expired")
	// ErrTokenReuseDetected indicates an already-revoked refresh token was presented.
	ErrTokenReuseDetected = errors.New("token reuse detected; session family revoked")
)

// Service coordinates authentication, session issuance, and refresh token rotation.
type Service struct {
	db          *database.DB
	cfg         *config.Config
	userService *users.Service
}

// NewService constructs a new auth Service.
func NewService(db *database.DB, cfg *config.Config, userService *users.Service) *Service {
	return &Service{
		db:          db,
		cfg:         cfg,
		userService: userService,
	}
}

// Login authenticates a user and establishes a new rotating refresh session.
func (s *Service) Login(ctx context.Context, email, password string) (*users.User, string, string, error) {
	user, err := s.userService.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			return nil, "", "", ErrInvalidCredentials
		}
		return nil, "", "", fmt.Errorf("failed to lookup user: %w", err)
	}

	if user.Status != users.StatusActive {
		return nil, "", "", ErrAccountDisabled
	}

	match, err := hasher.VerifyPassword(password, user.PasswordHash)
	if err != nil || !match {
		return nil, "", "", ErrInvalidCredentials
	}

	// Generate access token
	accessToken, err := GenerateAccessToken(user, s.cfg.JWTSecret, s.cfg.JWTAccessExpiry)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to create access token: %w", err)
	}

	// Generate rotating refresh token
	rawRefreshToken, tokenHash, err := GenerateOpaqueToken()
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to create refresh token: %w", err)
	}

	tokenFamilyID := uuid.New()
	expiresAt := time.Now().UTC().Add(s.cfg.RefreshTokenExpiry)

	_, err = s.db.Queries.CreateRefreshSession(ctx, sqlc.CreateRefreshSessionParams{
		UserID:        pgtype.UUID{Bytes: user.ID, Valid: true},
		TokenHash:     tokenHash,
		TokenFamilyID: pgtype.UUID{Bytes: tokenFamilyID, Valid: true},
		ExpiresAt:     pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to store refresh session: %w", err)
	}

	// Update last login timestamp asynchronously
	_ = s.db.Queries.UpdateUserLastLogin(ctx, pgtype.UUID{Bytes: user.ID, Valid: true})

	return user, accessToken, rawRefreshToken, nil
}

// Refresh rotates the refresh token and returns a new access token.
// Enforces automatic family revocation upon reuse detection.
func (s *Service) Refresh(ctx context.Context, rawRefreshToken string) (*users.User, string, string, error) {
	if rawRefreshToken == "" {
		return nil, "", "", ErrSessionNotFound
	}

	tokenHash := HashOpaqueToken(rawRefreshToken)

	session, err := s.db.Queries.GetRefreshSessionByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", "", ErrSessionNotFound
		}
		return nil, "", "", fmt.Errorf("failed to lookup refresh session: %w", err)
	}

	// Reuse detection: If token was already revoked, revoke its entire family!
	if session.RevokedAt.Valid {
		_ = s.db.Queries.RevokeRefreshFamily(ctx, session.TokenFamilyID)
		return nil, "", "", ErrTokenReuseDetected
	}

	// Check expiration
	if session.ExpiresAt.Time.Before(time.Now().UTC()) {
		return nil, "", "", ErrSessionNotFound
	}

	// Verify user status
	userID := uuid.UUID(session.UserID.Bytes)
	user, err := s.userService.GetUserByID(ctx, userID)
	if err != nil {
		return nil, "", "", ErrSessionNotFound
	}
	if user.Status != users.StatusActive {
		return nil, "", "", ErrAccountDisabled
	}

	// Revoke current session
	if err := s.db.Queries.RevokeRefreshSession(ctx, session.ID); err != nil {
		return nil, "", "", fmt.Errorf("failed to revoke previous session: %w", err)
	}

	// Rotate: Issue new refresh token within the same token family
	newRawRefreshToken, newTokenHash, err := GenerateOpaqueToken()
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate rotated refresh token: %w", err)
	}

	expiresAt := time.Now().UTC().Add(s.cfg.RefreshTokenExpiry)
	_, err = s.db.Queries.CreateRefreshSession(ctx, sqlc.CreateRefreshSessionParams{
		UserID:        session.UserID,
		TokenHash:     newTokenHash,
		TokenFamilyID: session.TokenFamilyID, // Preserves token family lineage
		ExpiresAt:     pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to store rotated refresh session: %w", err)
	}

	// Generate new access token
	newAccessToken, err := GenerateAccessToken(user, s.cfg.JWTSecret, s.cfg.JWTAccessExpiry)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to create access token: %w", err)
	}

	return user, newAccessToken, newRawRefreshToken, nil
}

// Logout revokes the active refresh session.
func (s *Service) Logout(ctx context.Context, rawRefreshToken string) error {
	if rawRefreshToken == "" {
		return nil
	}

	tokenHash := HashOpaqueToken(rawRefreshToken)
	session, err := s.db.Queries.GetRefreshSessionByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("failed to lookup refresh session: %w", err)
	}

	return s.db.Queries.RevokeRefreshSession(ctx, session.ID)
}
