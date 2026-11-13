package bookmarks

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

// CreateBookmark handles POST /api/v1/bookmarks.
func (h *Handler) CreateBookmark(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	var req CreateBookmarkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Malformed request body", nil)
		return
	}

	if req.LessonID == uuid.Nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "lesson_id is required", nil)
		return
	}

	b, err := h.service.AddLessonBookmark(r.Context(), *userID, req.LessonID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create bookmark", nil)
		return
	}

	response.JSON(w, http.StatusCreated, b, nil)
}

// ListBookmarks handles GET /api/v1/bookmarks.
func (h *Handler) ListBookmarks(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	bookmarks, err := h.service.ListUserBookmarks(r.Context(), *userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list bookmarks", nil)
		return
	}

	response.JSON(w, http.StatusOK, bookmarks, nil)
}

// CheckLesson handles GET /api/v1/bookmarks/check?lesson_id=...
func (h *Handler) CheckLesson(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	lessonIDStr := r.URL.Query().Get("lesson_id")
	lessonID, err := uuid.Parse(lessonIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid lesson_id", nil)
		return
	}

	isBookmarked, err := h.service.IsLessonBookmarked(r.Context(), *userID, lessonID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to check bookmark", nil)
		return
	}

	response.JSON(w, http.StatusOK, map[string]bool{"bookmarked": isBookmarked}, nil)
}

// DeleteBookmark handles DELETE /api/v1/bookmarks/{id}.
func (h *Handler) DeleteBookmark(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid bookmark ID", nil)
		return
	}

	if err := h.service.RemoveBookmarkByID(r.Context(), *userID, id); err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete bookmark", nil)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Bookmark removed successfully"}, nil)
}

// DeleteLessonBookmark handles DELETE /api/v1/bookmarks/lessons/{lessonId}.
func (h *Handler) DeleteLessonBookmark(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	lessonIDStr := chi.URLParam(r, "lessonId")
	lessonID, err := uuid.Parse(lessonIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid lesson ID", nil)
		return
	}

	if err := h.service.RemoveLessonBookmark(r.Context(), *userID, lessonID); err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to remove lesson bookmark", nil)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Lesson bookmark removed successfully"}, nil)
}
