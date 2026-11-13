package analytics

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"go-academy/backend/internal/auth"
	"go-academy/backend/internal/platform/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func getUserID(r *http.Request) (*uuid.UUID, error) {
	claims, ok := auth.GetClaims(r.Context())
	if !ok {
		return nil, errors.New("unauthorized")
	}
	id, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, errors.New("invalid user id in claims")
	}
	return &id, nil
}

// GetSummary handles GET /api/v1/analytics/summary.
func (h *Handler) GetSummary(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	stats, err := h.service.GetSummaryStats(r.Context(), *userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get user summary", nil)
		return
	}

	response.JSON(w, http.StatusOK, stats, nil)
}

// GetActivity handles GET /api/v1/analytics/activity?days=90.
func (h *Handler) GetActivity(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	days := 90
	if daysStr := r.URL.Query().Get("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	activity, err := h.service.GetActivityHistory(r.Context(), *userID, days)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get activity history", nil)
		return
	}

	response.JSON(w, http.StatusOK, activity, nil)
}

// GetBadges handles GET /api/v1/analytics/badges.
func (h *Handler) GetBadges(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	badges, err := h.service.GetUserBadges(r.Context(), *userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get user badges", nil)
		return
	}

	response.JSON(w, http.StatusOK, badges, nil)
}

// ClaimCertificate handles POST /api/v1/certificates/claim.
func (h *Handler) ClaimCertificate(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	var req ClaimCertificateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Malformed request body", nil)
		return
	}

	cert, err := h.service.ClaimCertificate(r.Context(), *userID, req.LearningPathID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Learning path not found", nil)
			return
		}
		if errors.Is(err, ErrPathNotCompleted) {
			response.Error(w, http.StatusBadRequest, "PATH_NOT_COMPLETED", err.Error(), nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to claim certificate", nil)
		return
	}

	response.JSON(w, http.StatusCreated, cert, nil)
}

// ListCertificates handles GET /api/v1/certificates/me.
func (h *Handler) ListCertificates(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	certs, err := h.service.ListUserCertificates(r.Context(), *userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list certificates", nil)
		return
	}

	response.JSON(w, http.StatusOK, certs, nil)
}

// VerifyCertificate handles GET /api/v1/certificates/verify/{code} (Public).
func (h *Handler) VerifyCertificate(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Certificate code is required", nil)
		return
	}

	cert, err := h.service.GetCertificateByCode(r.Context(), code)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Certificate not found or invalid", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to verify certificate", nil)
		return
	}

	response.JSON(w, http.StatusOK, cert, nil)
}
