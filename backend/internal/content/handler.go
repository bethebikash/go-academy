package content

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"go-academy/backend/internal/auth"
	"go-academy/backend/internal/platform/response"
)

// Handler serves HTTP endpoints for curriculum and content management.
type Handler struct {
	service *Service
}

// NewHandler constructs a new content Handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func getUserIDFromContext(r *http.Request) *uuid.UUID {
	claims, ok := auth.GetClaims(r.Context())
	if !ok {
		return nil
	}
	id, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil
	}
	return &id
}

// --- Learner Endpoints ---

// ListPaths handles GET /api/v1/learning-paths.
func (h *Handler) ListPaths(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	paths, err := h.service.ListPublishedLearningPaths(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list learning paths", nil)
		return
	}

	response.JSON(w, http.StatusOK, paths, nil)
}

// GetPath handles GET /api/v1/learning-paths/{slug}.
func (h *Handler) GetPath(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Path slug is required", nil)
		return
	}

	userID := getUserIDFromContext(r)
	path, courses, err := h.service.GetLearningPathBySlug(r.Context(), slug, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Learning path not found", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get learning path", nil)
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"path":    path,
		"courses": courses,
	}, nil)
}

// GetCourse handles GET /api/v1/courses/{slug}.
func (h *Handler) GetCourse(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Course slug is required", nil)
		return
	}

	userID := getUserIDFromContext(r)
	courseDetail, err := h.service.GetCourseBySlug(r.Context(), slug, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Course not found", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get course syllabus", nil)
		return
	}

	response.JSON(w, http.StatusOK, courseDetail, nil)
}

// GetLesson handles GET /api/v1/lessons/{slug}.
func (h *Handler) GetLesson(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Lesson slug is required", nil)
		return
	}

	userID := getUserIDFromContext(r)
	lessonDetail, err := h.service.GetLessonBySlug(r.Context(), slug, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Lesson not found", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get lesson details", nil)
		return
	}

	response.JSON(w, http.StatusOK, lessonDetail, nil)
}

// --- Admin Endpoints: Learning Paths ---

// AdminListPaths handles GET /api/v1/admin/learning-paths.
func (h *Handler) AdminListPaths(w http.ResponseWriter, r *http.Request) {
	paths, err := h.service.ListAllLearningPaths(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list learning paths", nil)
		return
	}

	response.JSON(w, http.StatusOK, paths, nil)
}

// AdminCreatePath handles POST /api/v1/admin/learning-paths.
func (h *Handler) AdminCreatePath(w http.ResponseWriter, r *http.Request) {
	adminID := getUserIDFromContext(r)
	if adminID == nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	var req CreateLearningPathRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Malformed request body", nil)
		return
	}

	path, err := h.service.CreateLearningPath(r.Context(), req, *adminID)
	if err != nil {
		if errors.Is(err, ErrInvalidInput) {
			response.Error(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error(), nil)
			return
		}
		if errors.Is(err, ErrSlugExists) {
			response.Error(w, http.StatusConflict, "CONFLICT", "A learning path with this slug already exists", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create learning path", nil)
		return
	}

	response.JSON(w, http.StatusCreated, path, nil)
}

// AdminGetPath handles GET /api/v1/admin/learning-paths/{id}.
func (h *Handler) AdminGetPath(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid path ID", nil)
		return
	}

	path, err := h.service.GetLearningPathByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Learning path not found", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get learning path", nil)
		return
	}

	response.JSON(w, http.StatusOK, path, nil)
}

// AdminUpdatePath handles PUT /api/v1/admin/learning-paths/{id}.
func (h *Handler) AdminUpdatePath(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid path ID", nil)
		return
	}

	var req UpdateLearningPathRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Malformed request body", nil)
		return
	}

	path, err := h.service.UpdateLearningPath(r.Context(), id, req)
	if err != nil {
		if errors.Is(err, ErrInvalidInput) {
			response.Error(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error(), nil)
			return
		}
		if errors.Is(err, ErrSlugExists) {
			response.Error(w, http.StatusConflict, "CONFLICT", "A learning path with this slug already exists", nil)
			return
		}
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Learning path not found", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update learning path", nil)
		return
	}

	response.JSON(w, http.StatusOK, path, nil)
}

// AdminUpdatePathStatus handles PATCH /api/v1/admin/learning-paths/{id}/status.
func (h *Handler) AdminUpdatePathStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid path ID", nil)
		return
	}

	var req UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Malformed request body", nil)
		return
	}

	path, err := h.service.UpdateLearningPathStatus(r.Context(), id, req.Status)
	if err != nil {
		if errors.Is(err, ErrInvalidInput) {
			response.Error(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error(), nil)
			return
		}
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Learning path not found", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update status", nil)
		return
	}

	response.JSON(w, http.StatusOK, path, nil)
}

// AdminDeletePath handles DELETE /api/v1/admin/learning-paths/{id}.
func (h *Handler) AdminDeletePath(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid path ID", nil)
		return
	}

	if err := h.service.DeleteLearningPath(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete learning path", nil)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Learning path deleted successfully"}, nil)
}

// --- Admin Endpoints: Courses ---

// AdminListCourses handles GET /api/v1/admin/learning-paths/{pathId}/courses.
func (h *Handler) AdminListCourses(w http.ResponseWriter, r *http.Request) {
	pathIDStr := chi.URLParam(r, "pathId")
	pathID, err := uuid.Parse(pathIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid path ID", nil)
		return
	}

	courses, err := h.service.ListCoursesByPath(r.Context(), pathID, false)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list courses", nil)
		return
	}

	response.JSON(w, http.StatusOK, courses, nil)
}

// AdminCreateCourse handles POST /api/v1/admin/courses.
func (h *Handler) AdminCreateCourse(w http.ResponseWriter, r *http.Request) {
	var req CreateCourseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Malformed request body", nil)
		return
	}

	course, err := h.service.CreateCourse(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidInput) {
			response.Error(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error(), nil)
			return
		}
		if errors.Is(err, ErrSlugExists) {
			response.Error(w, http.StatusConflict, "CONFLICT", "A course with this slug already exists", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create course", nil)
		return
	}

	response.JSON(w, http.StatusCreated, course, nil)
}

// AdminGetCourse handles GET /api/v1/admin/courses/{id}.
func (h *Handler) AdminGetCourse(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid course ID", nil)
		return
	}

	course, err := h.service.GetCourseByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Course not found", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get course", nil)
		return
	}

	response.JSON(w, http.StatusOK, course, nil)
}

// AdminUpdateCourse handles PUT /api/v1/admin/courses/{id}.
func (h *Handler) AdminUpdateCourse(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid course ID", nil)
		return
	}

	var req UpdateCourseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Malformed request body", nil)
		return
	}

	course, err := h.service.UpdateCourse(r.Context(), id, req)
	if err != nil {
		if errors.Is(err, ErrInvalidInput) {
			response.Error(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error(), nil)
			return
		}
		if errors.Is(err, ErrSlugExists) {
			response.Error(w, http.StatusConflict, "CONFLICT", "A course with this slug already exists", nil)
			return
		}
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Course not found", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update course", nil)
		return
	}

	response.JSON(w, http.StatusOK, course, nil)
}

// AdminUpdateCourseStatus handles PATCH /api/v1/admin/courses/{id}/status.
func (h *Handler) AdminUpdateCourseStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid course ID", nil)
		return
	}

	var req UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Malformed request body", nil)
		return
	}

	course, err := h.service.UpdateCourseStatus(r.Context(), id, req.Status)
	if err != nil {
		if errors.Is(err, ErrInvalidInput) {
			response.Error(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error(), nil)
			return
		}
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Course not found", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update course status", nil)
		return
	}

	response.JSON(w, http.StatusOK, course, nil)
}

// AdminDeleteCourse handles DELETE /api/v1/admin/courses/{id}.
func (h *Handler) AdminDeleteCourse(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid course ID", nil)
		return
	}

	if err := h.service.DeleteCourse(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete course", nil)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Course deleted successfully"}, nil)
}

// --- Admin Endpoints: Modules ---

// AdminListModules handles GET /api/v1/admin/courses/{courseId}/modules.
func (h *Handler) AdminListModules(w http.ResponseWriter, r *http.Request) {
	courseIDStr := chi.URLParam(r, "courseId")
	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid course ID", nil)
		return
	}

	modules, err := h.service.ListModulesByCourse(r.Context(), courseID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list modules", nil)
		return
	}

	response.JSON(w, http.StatusOK, modules, nil)
}

// AdminCreateModule handles POST /api/v1/admin/modules.
func (h *Handler) AdminCreateModule(w http.ResponseWriter, r *http.Request) {
	var req CreateModuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Malformed request body", nil)
		return
	}

	module, err := h.service.CreateModule(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidInput) {
			response.Error(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error(), nil)
			return
		}
		if errors.Is(err, ErrSlugExists) {
			response.Error(w, http.StatusConflict, "CONFLICT", "A module with this slug already exists", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create module", nil)
		return
	}

	response.JSON(w, http.StatusCreated, module, nil)
}

// AdminGetModule handles GET /api/v1/admin/modules/{id}.
func (h *Handler) AdminGetModule(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid module ID", nil)
		return
	}

	module, err := h.service.GetModuleByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Module not found", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get module", nil)
		return
	}

	response.JSON(w, http.StatusOK, module, nil)
}

// AdminUpdateModule handles PUT /api/v1/admin/modules/{id}.
func (h *Handler) AdminUpdateModule(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid module ID", nil)
		return
	}

	var req UpdateModuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Malformed request body", nil)
		return
	}

	module, err := h.service.UpdateModule(r.Context(), id, req)
	if err != nil {
		if errors.Is(err, ErrInvalidInput) {
			response.Error(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error(), nil)
			return
		}
		if errors.Is(err, ErrSlugExists) {
			response.Error(w, http.StatusConflict, "CONFLICT", "A module with this slug already exists", nil)
			return
		}
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Module not found", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update module", nil)
		return
	}

	response.JSON(w, http.StatusOK, module, nil)
}

// AdminDeleteModule handles DELETE /api/v1/admin/modules/{id}.
func (h *Handler) AdminDeleteModule(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid module ID", nil)
		return
	}

	if err := h.service.DeleteModule(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete module", nil)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Module deleted successfully"}, nil)
}

// --- Admin Endpoints: Lessons ---

// AdminListLessons handles GET /api/v1/admin/modules/{moduleId}/lessons.
func (h *Handler) AdminListLessons(w http.ResponseWriter, r *http.Request) {
	moduleIDStr := chi.URLParam(r, "moduleId")
	moduleID, err := uuid.Parse(moduleIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid module ID", nil)
		return
	}

	lessons, err := h.service.ListLessonsByModule(r.Context(), moduleID, false)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list lessons", nil)
		return
	}

	response.JSON(w, http.StatusOK, lessons, nil)
}

// AdminCreateLesson handles POST /api/v1/admin/lessons.
func (h *Handler) AdminCreateLesson(w http.ResponseWriter, r *http.Request) {
	var req CreateLessonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Malformed request body", nil)
		return
	}

	lesson, err := h.service.CreateLesson(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidInput) {
			response.Error(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error(), nil)
			return
		}
		if errors.Is(err, ErrSlugExists) {
			response.Error(w, http.StatusConflict, "CONFLICT", "A lesson with this slug already exists", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create lesson", nil)
		return
	}

	response.JSON(w, http.StatusCreated, lesson, nil)
}

// AdminGetLesson handles GET /api/v1/admin/lessons/{id}.
func (h *Handler) AdminGetLesson(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid lesson ID", nil)
		return
	}

	lesson, prereqIDs, err := h.service.GetLessonByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Lesson not found", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get lesson", nil)
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"lesson":           lesson,
		"prerequisite_ids": prereqIDs,
	}, nil)
}

// AdminUpdateLesson handles PUT /api/v1/admin/lessons/{id}.
func (h *Handler) AdminUpdateLesson(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid lesson ID", nil)
		return
	}

	var req UpdateLessonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Malformed request body", nil)
		return
	}

	lesson, err := h.service.UpdateLesson(r.Context(), id, req)
	if err != nil {
		if errors.Is(err, ErrInvalidInput) {
			response.Error(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error(), nil)
			return
		}
		if errors.Is(err, ErrSlugExists) {
			response.Error(w, http.StatusConflict, "CONFLICT", "A lesson with this slug already exists", nil)
			return
		}
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Lesson not found", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update lesson", nil)
		return
	}

	response.JSON(w, http.StatusOK, lesson, nil)
}

// AdminUpdateLessonStatus handles PATCH /api/v1/admin/lessons/{id}/status.
func (h *Handler) AdminUpdateLessonStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid lesson ID", nil)
		return
	}

	var req UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Malformed request body", nil)
		return
	}

	lesson, err := h.service.UpdateLessonStatus(r.Context(), id, req.Status)
	if err != nil {
		if errors.Is(err, ErrInvalidInput) {
			response.Error(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error(), nil)
			return
		}
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Lesson not found", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update lesson status", nil)
		return
	}

	response.JSON(w, http.StatusOK, lesson, nil)
}

// AdminDeleteLesson handles DELETE /api/v1/admin/lessons/{id}.
func (h *Handler) AdminDeleteLesson(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid lesson ID", nil)
		return
	}

	if err := h.service.DeleteLesson(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete lesson", nil)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Lesson deleted successfully"}, nil)
}
