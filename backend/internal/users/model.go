package users

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrAdminAlreadyExists indicates an administrator account already exists.
	ErrAdminAlreadyExists = errors.New("administrator already exists; repeated bootstrap is disallowed")
	// ErrUserNotFound indicates the target user does not exist.
	ErrUserNotFound = errors.New("user not found")
	// ErrEmailTaken indicates the email address is already in use.
	ErrEmailTaken = errors.New("email is already registered")
	// ErrUsernameTaken indicates the username is already in use.
	ErrUsernameTaken = errors.New("username is already taken")
	// ErrInvalidInput indicates validation failure on input arguments.
	ErrInvalidInput = errors.New("invalid user input")
	// ErrUserDisabled indicates the account is deactivated.
	ErrUserDisabled = errors.New("account is disabled")
)

const (
	RoleAdmin   = "admin"
	RoleLearner = "learner"

	StatusActive   = "active"
	StatusDisabled = "disabled"
)

// User represents an application user account.
type User struct {
	ID           uuid.UUID  `json:"id"`
	Email        string     `json:"email"`
	Username     string     `json:"username"`
	PasswordHash string     `json:"-"`
	Role         string     `json:"role"`
	Status       string     `json:"status"`
	Timezone     string     `json:"timezone"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
