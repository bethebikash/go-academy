package challenges

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"go-academy/backend/internal/platform/response"
)

// Handler serves HTTP endpoints for coding challenges.
type Handler struct {
	service *Service
}

// NewHandler constructs a new challenges Handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// --- Learner Endpoints ---

// ListChallenges handles GET /api/v1/challenges.
func (h *Handler) ListChallenges(w http.ResponseWriter, r *http.Request) {
	difficulty := r.URL.Query().Get("difficulty")
	challenges, err := h.service.ListPublishedChallenges(r.Context(), difficulty)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list challenges", nil)
		return
	}

	response.JSON(w, http.StatusOK, challenges, nil)
}

// GetChallenge handles GET /api/v1/challenges/{slug}.
func (h *Handler) GetChallenge(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Challenge slug is required", nil)
		return
	}

	challenge, err := h.service.GetPublishedChallengeBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Challenge not found", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get challenge", nil)
		return
	}

	response.JSON(w, http.StatusOK, challenge, nil)
}

// --- Admin Endpoints ---

// AdminListChallenges handles GET /api/v1/admin/challenges.
func (h *Handler) AdminListChallenges(w http.ResponseWriter, r *http.Request) {
	challenges, err := h.service.ListAllChallenges(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list all challenges", nil)
		return
	}

	response.JSON(w, http.StatusOK, challenges, nil)
}

// AdminCreateChallenge handles POST /api/v1/admin/challenges.
func (h *Handler) AdminCreateChallenge(w http.ResponseWriter, r *http.Request) {
	var req CreateChallengeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Malformed request body", nil)
		return
	}

	challenge, err := h.service.CreateChallenge(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidInput) {
			response.Error(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error(), nil)
			return
		}
		if errors.Is(err, ErrSlugExists) {
			response.Error(w, http.StatusConflict, "SLUG_EXISTS", "A challenge with this slug already exists", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create challenge", nil)
		return
	}

	response.JSON(w, http.StatusCreated, challenge, nil)
}

// AdminGetChallenge handles GET /api/v1/admin/challenges/{id}.
func (h *Handler) AdminGetChallenge(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid challenge ID", nil)
		return
	}

	challenge, err := h.service.GetChallengeByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Challenge not found", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get challenge", nil)
		return
	}

	response.JSON(w, http.StatusOK, challenge, nil)
}

// AdminUpdateChallenge handles PUT /api/v1/admin/challenges/{id}.
func (h *Handler) AdminUpdateChallenge(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid challenge ID", nil)
		return
	}

	var req UpdateChallengeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Malformed request body", nil)
		return
	}

	challenge, err := h.service.UpdateChallenge(r.Context(), id, req)
	if err != nil {
		if errors.Is(err, ErrInvalidInput) {
			response.Error(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error(), nil)
			return
		}
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Challenge not found", nil)
			return
		}
		if errors.Is(err, ErrSlugExists) {
			response.Error(w, http.StatusConflict, "SLUG_EXISTS", "A challenge with this slug already exists", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update challenge", nil)
		return
	}

	response.JSON(w, http.StatusOK, challenge, nil)
}

// AdminUpdateChallengeStatus handles PATCH /api/v1/admin/challenges/{id}/status.
func (h *Handler) AdminUpdateChallengeStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid challenge ID", nil)
		return
	}

	var req UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Malformed request body", nil)
		return
	}

	challenge, err := h.service.UpdateChallengeStatus(r.Context(), id, req.Status)
	if err != nil {
		if errors.Is(err, ErrInvalidInput) {
			response.Error(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error(), nil)
			return
		}
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Challenge not found", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update challenge status", nil)
		return
	}

	response.JSON(w, http.StatusOK, challenge, nil)
}

// AdminDeleteChallenge handles DELETE /api/v1/admin/challenges/{id}.
func (h *Handler) AdminDeleteChallenge(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid challenge ID", nil)
		return
	}

	if err := h.service.DeleteChallenge(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete challenge", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
