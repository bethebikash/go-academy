package invitations

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"go-academy/backend/internal/auth"
	"go-academy/backend/internal/config"
	"go-academy/backend/internal/platform/response"
	"go-academy/backend/internal/users"
)

const (
	refreshTokenCookieName = "refresh_token"
)

// Handler serves HTTP endpoints for invitation management and registration.
type Handler struct {
	service *Service
	cfg     *config.Config
}

// NewHandler constructs a new invitations Handler.
func NewHandler(service *Service, cfg *config.Config) *Handler {
	return &Handler{
		service: service,
		cfg:     cfg,
	}
}

type createInvitationRequest struct {
	Email         string `json:"email"`
	ExpiresInDays int    `json:"expires_in_days"`
}

type registerRequest struct {
	InvitationToken string `json:"invitation_token"`
	Username        string `json:"username"`
	Password        string `json:"password"`
}

// Create handles admin invitation creation.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetClaims(r.Context())
	if !ok || claims == nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required", nil)
		return
	}

	adminID, err := uuid.Parse(claims.UserID)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Invalid administrator identity", nil)
		return
	}

	var req createInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Malformed JSON payload", nil)
		return
	}

	inv, err := h.service.CreateInvitation(r.Context(), adminID, req.Email, req.ExpiresInDays)
	if err != nil {
		if errors.Is(err, ErrInvalidEmail) {
			response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "A valid email address is required", nil)
			return
		}
		if errors.Is(err, ErrEmailAlreadyRegistered) {
			response.Error(w, http.StatusConflict, "CONFLICT", "A user with this email address is already registered", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create invitation", nil)
		return
	}

	response.JSON(w, http.StatusCreated, inv, nil)
}

// List handles admin invitation listing with pagination.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	invitations, total, err := h.service.ListInvitations(r.Context(), page, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list invitations", nil)
		return
	}

	meta := map[string]any{
		"pagination": map[string]any{
			"page":        page,
			"limit":       limit,
			"total_items": total,
		},
	}

	response.JSON(w, http.StatusOK, invitations, meta)
}

// Revoke handles admin invitation revocation.
func (h *Handler) Revoke(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid invitation ID", nil)
		return
	}

	inv, err := h.service.RevokeInvitation(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrInvalidInvitation) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Invitation not found or already processed", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to revoke invitation", nil)
		return
	}

	response.JSON(w, http.StatusOK, inv, nil)
}

// Validate checks whether an invitation token is valid and unconsumed.
func (h *Handler) Validate(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invitation token query parameter is required", nil)
		return
	}

	inv, err := h.service.ValidateInvitationToken(r.Context(), token)
	if err != nil {
		if errors.Is(err, ErrInvitationExpired) {
			response.Error(w, http.StatusBadRequest, "INVITATION_EXPIRED", "This invitation link has expired", nil)
			return
		}
		if errors.Is(err, ErrInvitationRevoked) {
			response.Error(w, http.StatusBadRequest, "INVITATION_REVOKED", "This invitation has been revoked", nil)
			return
		}
		if errors.Is(err, ErrInvitationAlreadyUsed) {
			response.Error(w, http.StatusConflict, "INVITATION_ALREADY_USED", "This invitation has already been accepted", nil)
			return
		}
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "Invalid or unknown invitation token", nil)
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"email":      inv.Email,
		"expires_at": inv.ExpiresAt,
		"valid":      true,
	}, nil)
}

// Register processes an invitation and creates the user account transactionally.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Malformed JSON payload", nil)
		return
	}

	user, accessToken, rawRefreshToken, err := h.service.AcceptInvitationAndRegister(
		r.Context(),
		req.InvitationToken,
		req.Username,
		req.Password,
	)
	if err != nil {
		if errors.Is(err, ErrInvalidInvitation) {
			response.Error(w, http.StatusBadRequest, "INVALID_INVITATION", "Invitation token is invalid", nil)
			return
		}
		if errors.Is(err, ErrInvitationExpired) {
			response.Error(w, http.StatusBadRequest, "INVITATION_EXPIRED", "This invitation link has expired", nil)
			return
		}
		if errors.Is(err, ErrInvitationRevoked) {
			response.Error(w, http.StatusBadRequest, "INVITATION_REVOKED", "This invitation has been revoked", nil)
			return
		}
		if errors.Is(err, ErrInvitationAlreadyUsed) {
			response.Error(w, http.StatusConflict, "INVITATION_ALREADY_USED", "This invitation has already been accepted", nil)
			return
		}
		if errors.Is(err, users.ErrUsernameTaken) {
			response.Error(w, http.StatusConflict, "USERNAME_TAKEN", "This username is already taken", nil)
			return
		}
		if errors.Is(err, users.ErrInvalidInput) {
			response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			return
		}

		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Registration failed", nil)
		return
	}

	// Set refresh token cookie
	h.setRefreshTokenCookie(w, rawRefreshToken)

	response.JSON(w, http.StatusCreated, map[string]any{
		"access_token": accessToken,
		"expires_in":   int(h.cfg.JWTAccessExpiry.Seconds()),
		"user":         user,
	}, nil)
}

func (h *Handler) setRefreshTokenCookie(w http.ResponseWriter, rawToken string) {
	cookie := &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    rawToken,
		Path:     "/api/v1/auth",
		MaxAge:   int(h.cfg.RefreshTokenExpiry.Seconds()),
		Expires:  time.Now().Add(h.cfg.RefreshTokenExpiry),
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	}

	if h.cfg.IsProduction() && h.cfg.CookieDomain != "" && h.cfg.CookieDomain != "localhost" {
		cookie.Domain = h.cfg.CookieDomain
	}

	http.SetCookie(w, cookie)
}
