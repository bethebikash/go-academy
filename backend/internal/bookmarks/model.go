package bookmarks

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Bookmark struct {
	ID                     uuid.UUID `json:"id"`
	UserID                 uuid.UUID `json:"user_id"`
	LessonID               uuid.UUID `json:"lesson_id"`
	LessonTitle            string    `json:"lesson_title"`
	LessonSlug             string    `json:"lesson_slug"`
	LessonSummary          string    `json:"lesson_summary"`
	LessonDifficulty       string    `json:"lesson_difficulty"`
	LessonEstimatedMinutes int       `json:"lesson_estimated_minutes"`
	CourseTitle            string    `json:"course_title"`
	CourseSlug             string    `json:"course_slug"`
	ProgressStatus         string    `json:"progress_status"`
	CreatedAt              time.Time `json:"created_at"`
}

type CreateBookmarkRequest struct {
	LessonID uuid.UUID `json:"lesson_id"`
}

func toPgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}
