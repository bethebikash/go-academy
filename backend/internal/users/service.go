package users

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"go-academy/backend/internal/database"
	"go-academy/backend/internal/database/sqlc"
	"go-academy/backend/internal/platform/hasher"
)

var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,30}$`)
)

// Service provides user management business logic.
type Service struct {
	db *database.DB
}

// NewService constructs a new users Service.
func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

// ValidateRegistrationCredentials validates email, username, and password requirements.
func ValidateRegistrationCredentials(email, username, password string) error {
	trimmedEmail := strings.TrimSpace(email)
	if _, err := mail.ParseAddress(trimmedEmail); err != nil {
		return fmt.Errorf("%w: invalid email address", ErrInvalidInput)
	}

	trimmedUsername := strings.TrimSpace(username)
	if !usernameRegex.MatchString(trimmedUsername) {
		return fmt.Errorf("%w: username must be 3-30 characters and contain only letters, numbers, hyphens, or underscores", ErrInvalidInput)
	}

	if len(password) < 10 {
		return fmt.Errorf("%w: password must be at least 10 characters long", ErrInvalidInput)
	}

	var hasLetter, hasDigitOrPunct bool
	for _, c := range password {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			hasLetter = true
		} else {
			hasDigitOrPunct = true
		}
	}
	if !hasLetter || !hasDigitOrPunct {
		return fmt.Errorf("%w: password must include both letters and numbers/symbols", ErrInvalidInput)
	}

	return nil
}

// BootstrapAdmin provisions the initial administrator account. Disallows repeated execution.
func (s *Service) BootstrapAdmin(ctx context.Context, email, username, password string) (*User, error) {
	if err := ValidateRegistrationCredentials(email, username, password); err != nil {
		return nil, err
	}

	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	cleanUsername := strings.TrimSpace(username)

	// Begin atomic transaction
	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) // Safe to call: no-op if committed

	qtx := s.db.Queries.WithTx(tx)

	// Check if any admin exists
	adminCount, err := qtx.CountAdmins(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing administrators: %w", err)
	}
	if adminCount > 0 {
		return nil, ErrAdminAlreadyExists
	}

	// Verify email not taken
	if _, err := qtx.GetUserByEmail(ctx, cleanEmail); err == nil {
		return nil, ErrEmailTaken
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to check existing email: %w", err)
	}

	// Verify username not taken
	if _, err := qtx.GetUserByUsername(ctx, cleanUsername); err == nil {
		return nil, ErrUsernameTaken
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to check existing username: %w", err)
	}

	// Hash password with Argon2id
	hashedPassword, err := hasher.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user record
	createdUser, err := qtx.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        cleanEmail,
		Username:     cleanUsername,
		PasswordHash: hashedPassword,
		Status:       StatusActive,
		Timezone:     "UTC",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Retrieve admin role
	adminRole, err := qtx.GetRoleByName(ctx, RoleAdmin)
	if err != nil {
		return nil, fmt.Errorf("admin role not found in database: %w", err)
	}

	// Assign admin role
	if err := qtx.AssignRoleToUser(ctx, sqlc.AssignRoleToUserParams{
		UserID: createdUser.ID,
		RoleID: adminRole.ID,
	}); err != nil {
		return nil, fmt.Errorf("failed to assign admin role: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit bootstrap transaction: %w", err)
	}

	userID := uuid.UUID(createdUser.ID.Bytes)

	return &User{
		ID:        userID,
		Email:     createdUser.Email,
		Username:  createdUser.Username,
		Role:      RoleAdmin,
		Status:    createdUser.Status,
		Timezone:  createdUser.Timezone,
		CreatedAt: createdUser.CreatedAt.Time,
		UpdatedAt: createdUser.UpdatedAt.Time,
	}, nil
}

// GetUserByID fetches a user by primary key.
func (s *Service) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	pgID := pgtype.UUID{Bytes: id, Valid: true}
	row, err := s.db.Queries.GetUserByID(ctx, pgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	roles, err := s.db.Queries.GetUserRoles(ctx, pgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	primaryRole := RoleLearner
	for _, r := range roles {
		if r.Name == RoleAdmin {
			primaryRole = RoleAdmin
			break
		}
	}

	var lastLogin *time.Time
	if row.LastLoginAt.Valid {
		t := row.LastLoginAt.Time
		lastLogin = &t
	}

	return &User{
		ID:           id,
		Email:        row.Email,
		Username:     row.Username,
		PasswordHash: row.PasswordHash,
		Role:         primaryRole,
		Status:       row.Status,
		Timezone:     row.Timezone,
		LastLoginAt:  lastLogin,
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}, nil
}

// GetUserByEmail fetches a user by case-insensitive email address.
func (s *Service) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	row, err := s.db.Queries.GetUserByEmail(ctx, cleanEmail)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	roles, err := s.db.Queries.GetUserRoles(ctx, row.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	primaryRole := RoleLearner
	for _, r := range roles {
		if r.Name == RoleAdmin {
			primaryRole = RoleAdmin
			break
		}
	}

	var lastLogin *time.Time
	if row.LastLoginAt.Valid {
		t := row.LastLoginAt.Time
		lastLogin = &t
	}

	return &User{
		ID:           uuid.UUID(row.ID.Bytes),
		Email:        row.Email,
		Username:     row.Username,
		PasswordHash: row.PasswordHash,
		Role:         primaryRole,
		Status:       row.Status,
		Timezone:     row.Timezone,
		LastLoginAt:  lastLogin,
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}, nil
}

// UpdateUserStatus changes a user account status (e.g. active <-> disabled).
func (s *Service) UpdateUserStatus(ctx context.Context, id uuid.UUID, newStatus string) (*User, error) {
	if newStatus != StatusActive && newStatus != StatusDisabled {
		return nil, fmt.Errorf("%w: invalid status %q", ErrInvalidInput, newStatus)
	}

	pgID := pgtype.UUID{Bytes: id, Valid: true}
	row, err := s.db.Queries.UpdateUserStatus(ctx, sqlc.UpdateUserStatusParams{
		ID:     pgID,
		Status: newStatus,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to update status: %w", err)
	}

	return &User{
		ID:        id,
		Email:     row.Email,
		Username:  row.Username,
		Status:    row.Status,
		UpdatedAt: row.UpdatedAt.Time,
	}, nil
}

// CountAdmins returns the current count of admin users.
func (s *Service) CountAdmins(ctx context.Context) (int64, error) {
	return s.db.Queries.CountAdmins(ctx)
}

// ListUsers retrieves users with pagination and total count.
func (s *Service) ListUsers(ctx context.Context, page, limit int) ([]*User, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	total, err := s.db.Queries.CountUsers(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	rows, err := s.db.Queries.ListUsers(ctx, sqlc.ListUsersParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	items := make([]*User, 0, len(rows))
	for _, r := range rows {
		var lastLogin *time.Time
		if r.LastLoginAt.Valid {
			t := r.LastLoginAt.Time
			lastLogin = &t
		}

		roleName := r.Role
		if roleName == "" {
			roleName = RoleLearner
		}

		items = append(items, &User{
			ID:          uuid.UUID(r.ID.Bytes),
			Email:       r.Email,
			Username:    r.Username,
			Role:        roleName,
			Status:      r.Status,
			Timezone:    r.Timezone,
			LastLoginAt: lastLogin,
			CreatedAt:   r.CreatedAt.Time,
			UpdatedAt:   r.UpdatedAt.Time,
		})
	}

	return items, total, nil
}

