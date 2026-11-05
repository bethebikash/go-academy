package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	"go-academy/backend/internal/config"
	"go-academy/backend/internal/platform/response"
	"go-academy/backend/internal/users"
)

const (
	refreshTokenCookieName = "refresh_token"
)

// Handler serves HTTP endpoints for authentication.
type Handler struct {
	service     *Service
	userService *users.Service
	cfg         *config.Config
}

// NewHandler constructs a new auth Handler.
func NewHandler(service *Service, userService *users.Service, cfg *config.Config) *Handler {
	return &Handler{
		service:     service,
		userService: userService,
		cfg:         cfg,
	}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	AccessToken string      `json:"access_token"`
	ExpiresIn   int         `json:"expires_in"`
	User        *users.User `json:"user,omitempty"`
}

type refreshResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// Login handles user authentication via email & password.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Malformed JSON payload", nil)
		return
	}

	if req.Email == "" || req.Password == "" {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "Email and password are required", nil)
		return
	}

	user, accessToken, rawRefreshToken, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			response.Error(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password", nil)
			return
		}
		if errors.Is(err, ErrAccountDisabled) {
			response.Error(w, http.StatusForbidden, "ACCOUNT_DISABLED", "Your account has been deactivated", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Authentication failed", nil)
		return
	}

	h.setRefreshTokenCookie(w, rawRefreshToken)

	response.JSON(w, http.StatusOK, authResponse{
		AccessToken: accessToken,
		ExpiresIn:   int(h.cfg.JWTAccessExpiry.Seconds()),
		User:        user,
	}, nil)
}

// Refresh handles token rotation and issues a new access token.
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(refreshTokenCookieName)
	if err != nil || cookie.Value == "" {
		response.Error(w, http.StatusUnauthorized, "UNAUTHENTICATED", "No refresh session token provided", nil)
		return
	}

	_, newAccessToken, newRawRefreshToken, err := h.service.Refresh(r.Context(), cookie.Value)
	if err != nil {
		h.clearRefreshTokenCookie(w)

		if errors.Is(err, ErrTokenReuseDetected) {
			response.Error(
				w,
				http.StatusUnauthorized,
				"TOKEN_REUSE_DETECTED",
				"Potential session compromise detected. Please sign in again.",
				nil,
			)
			return
		}
		if errors.Is(err, ErrAccountDisabled) {
			response.Error(w, http.StatusForbidden, "ACCOUNT_DISABLED", "Your account has been deactivated", nil)
			return
		}

		response.Error(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Refresh session has expired or is invalid", nil)
		return
	}

	h.setRefreshTokenCookie(w, newRawRefreshToken)

	response.JSON(w, http.StatusOK, refreshResponse{
		AccessToken: newAccessToken,
		ExpiresIn:   int(h.cfg.JWTAccessExpiry.Seconds()),
	}, nil)
}

// Logout terminates the refresh session and deletes the session cookie.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(refreshTokenCookieName)
	if err == nil && cookie.Value != "" {
		_ = h.service.Logout(r.Context(), cookie.Value)
	}

	h.clearRefreshTokenCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// Me retrieves the profile of the currently authenticated user.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := GetClaims(r.Context())
	if !ok || claims == nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required", nil)
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Invalid user token identity", nil)
		return
	}

	user, err := h.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "User profile not found", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve user profile", nil)
		return
	}

	response.JSON(w, http.StatusOK, user, nil)
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

func (h *Handler) clearRefreshTokenCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    "",
		Path:     "/api/v1/auth",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	}

	if h.cfg.IsProduction() && h.cfg.CookieDomain != "" && h.cfg.CookieDomain != "localhost" {
		cookie.Domain = h.cfg.CookieDomain
	}

	http.SetCookie(w, cookie)
}
