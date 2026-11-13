package invitations

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrInvalidInvitation indicates the invitation token does not exist or is malformed.
	ErrInvalidInvitation = errors.New("invitation not found or invalid")
	// ErrInvitationExpired indicates the invitation token has passed its expiry date.
	ErrInvitationExpired = errors.New("invitation has expired")
	// ErrInvitationRevoked indicates the invitation has been explicitly revoked by an administrator.
	ErrInvitationRevoked = errors.New("invitation has been revoked")
	// ErrInvitationAlreadyUsed indicates the invitation token has already been accepted.
	ErrInvitationAlreadyUsed = errors.New("invitation has already been accepted")
	// ErrEmailAlreadyRegistered indicates an account with the invited email already exists.
	ErrEmailAlreadyRegistered = errors.New("a user with this email address is already registered")
	// ErrInvalidEmail indicates the email address format is invalid.
	ErrInvalidEmail = errors.New("invalid email address")
)

const (
	StatusPending  = "pending"
	StatusAccepted = "accepted"
	StatusExpired  = "expired"
	StatusRevoked  = "revoked"
)

// Invitation represents an invite record for controlled platform onboarding.
type Invitation struct {
	ID                uuid.UUID  `json:"id"`
	Email             string     `json:"email"`
	Status            string     `json:"status"`
	InvitedBy         uuid.UUID  `json:"invited_by"`
	InvitedByUsername string     `json:"invited_by_username,omitempty"`
	ExpiresAt         time.Time  `json:"expires_at"`
	AcceptedAt        *time.Time `json:"accepted_at,omitempty"`
	RevokedAt         *time.Time `json:"revoked_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	InvitationURL     string     `json:"invitation_url,omitempty"`
	RawToken          string     `json:"invitation_token,omitempty"`
}

// ComputeStatus determines the lifecycle state of an invitation record.
func ComputeStatus(expiresAt time.Time, acceptedAt, revokedAt *time.Time) string {
	if revokedAt != nil {
		return StatusRevoked
	}
	if acceptedAt != nil {
		return StatusAccepted
	}
	if time.Now().UTC().After(expiresAt) {
		return StatusExpired
	}
	return StatusPending
}
