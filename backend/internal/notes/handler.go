package notes

import (
	"encoding/json"
	"errors"
	"net/http"

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

// CreateNote handles POST /api/v1/notes.
func (h *Handler) CreateNote(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	var req CreateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Malformed request body", nil)
		return
	}

	note, err := h.service.CreateNote(r.Context(), *userID, req)
	if err != nil {
		if errors.Is(err, ErrInvalidInput) {
			response.Error(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error(), nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create note", nil)
		return
	}

	response.JSON(w, http.StatusCreated, note, nil)
}

// GetNote handles GET /api/v1/notes/{id}.
func (h *Handler) GetNote(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	noteIDStr := chi.URLParam(r, "id")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid note ID", nil)
		return
	}

	note, err := h.service.GetNoteByID(r.Context(), *userID, noteID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Note not found", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get note", nil)
		return
	}

	response.JSON(w, http.StatusOK, note, nil)
}

// ListNotes handles GET /api/v1/notes.
func (h *Handler) ListNotes(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	tag := r.URL.Query().Get("tag")
	query := r.URL.Query().Get("q")
	lessonIDStr := r.URL.Query().Get("lesson_id")

	var lessonID *uuid.UUID
	if lessonIDStr != "" {
		parsed, err := uuid.Parse(lessonIDStr)
		if err == nil {
			lessonID = &parsed
		}
	}

	notes, err := h.service.ListNotes(r.Context(), *userID, tag, lessonID, query)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list notes", nil)
		return
	}

	response.JSON(w, http.StatusOK, notes, nil)
}

// UpdateNote handles PATCH /api/v1/notes/{id} and PUT /api/v1/notes/{id}.
func (h *Handler) UpdateNote(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	noteIDStr := chi.URLParam(r, "id")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid note ID", nil)
		return
	}

	var req UpdateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Malformed request body", nil)
		return
	}

	note, err := h.service.UpdateNote(r.Context(), *userID, noteID, req)
	if err != nil {
		if errors.Is(err, ErrVersionConflict) {
			response.Error(w, http.StatusConflict, "VERSION_CONFLICT", "Note was modified elsewhere. Please refresh and try again.", nil)
			return
		}
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Note not found", nil)
			return
		}
		if errors.Is(err, ErrInvalidInput) {
			response.Error(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error(), nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update note", nil)
		return
	}

	response.JSON(w, http.StatusOK, note, nil)
}

// DeleteNote handles DELETE /api/v1/notes/{id}.
func (h *Handler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	noteIDStr := chi.URLParam(r, "id")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid note ID", nil)
		return
	}

	if err := h.service.DeleteNote(r.Context(), *userID, noteID); err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete note", nil)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Note deleted successfully"}, nil)
}

// ListTags handles GET /api/v1/tags.
func (h *Handler) ListTags(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	tags, err := h.service.ListTags(r.Context(), *userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list tags", nil)
		return
	}

	response.JSON(w, http.StatusOK, tags, nil)
}
