package content

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"go-academy/backend/internal/database/sqlc"
)

// Content status constants
const (
	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusArchived  = "archived"
)

// Difficulty levels
const (
	DifficultyBeginner     = "beginner"
	DifficultyIntermediate = "intermediate"
	DifficultyAdvanced     = "advanced"
)

// LearningPath domain model
type LearningPath struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Description string     `json:"description,omitempty"`
	Status      string     `json:"status"`
	SortOrder   int        `json:"sort_order"`
	CreatedBy   uuid.UUID  `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
}

// LearningPathWithStats includes aggregate progress for a learner
type LearningPathWithStats struct {
	LearningPath
	TotalLessons     int `json:"total_lessons"`
	CompletedLessons int `json:"completed_lessons"`
	ProgressPercent  int `json:"progress_percent"`
}

// Course domain model
type Course struct {
	ID             uuid.UUID  `json:"id"`
	LearningPathID uuid.UUID  `json:"learning_path_id"`
	Title          string     `json:"title"`
	Slug           string     `json:"slug"`
	Description    string     `json:"description,omitempty"`
	Status         string     `json:"status"`
	SortOrder      int        `json:"sort_order"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	PublishedAt    *time.Time `json:"published_at,omitempty"`
}

// CourseWithStats includes aggregate progress for a learner
type CourseWithStats struct {
	Course
	TotalLessons     int `json:"total_lessons"`
	CompletedLessons int `json:"completed_lessons"`
	ProgressPercent  int `json:"progress_percent"`
}

// Module domain model
type Module struct {
	ID          uuid.UUID `json:"id"`
	CourseID    uuid.UUID `json:"course_id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Description string    `json:"description,omitempty"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Lesson domain model
type Lesson struct {
	ID               uuid.UUID  `json:"id"`
	ModuleID         uuid.UUID  `json:"module_id"`
	Title            string     `json:"title"`
	Slug             string     `json:"slug"`
	Summary          string     `json:"summary,omitempty"`
	Content          string     `json:"content"`
	Difficulty       string     `json:"difficulty"`
	EstimatedMinutes int        `json:"estimated_minutes"`
	Status           string     `json:"status"`
	SortOrder        int        `json:"sort_order"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	PublishedAt      *time.Time `json:"published_at,omitempty"`
}

// LessonSummary for module curriculum lists
type LessonSummary struct {
	ID               uuid.UUID `json:"id"`
	Title            string    `json:"title"`
	Slug             string    `json:"slug"`
	Summary          string    `json:"summary,omitempty"`
	Difficulty       string    `json:"difficulty"`
	EstimatedMinutes int       `json:"estimated_minutes"`
	Status           string    `json:"status"` // admin content status or learner progress status
	SortOrder        int       `json:"sort_order"`
	IsCompleted      bool      `json:"is_completed"`
	IsLocked         bool      `json:"is_locked"`
}

// ModuleWithLessons represents a module and its ordered lessons
type ModuleWithLessons struct {
	Module
	Lessons []LessonSummary `json:"lessons"`
}

// CourseDetail represents a full syllabus for a course
type CourseDetail struct {
	Course
	Modules          []ModuleWithLessons `json:"modules"`
	TotalLessons     int                 `json:"total_lessons"`
	CompletedLessons int                 `json:"completed_lessons"`
	ProgressPercent  int                 `json:"progress_percent"`
}

// PrerequisiteInfo minimal info for locked lessons
type PrerequisiteInfo struct {
	ID    uuid.UUID `json:"id"`
	Title string    `json:"title"`
	Slug  string    `json:"slug"`
}

// LessonDetail represents a full lesson view for the reader
type LessonDetail struct {
	Lesson
	CourseTitle          string             `json:"course_title"`
	CourseSlug           string             `json:"course_slug"`
	ModuleTitle          string             `json:"module_title"`
	ModuleSlug           string             `json:"module_slug"`
	ProgressStatus       string             `json:"progress_status"` // not_started, started, completed
	IsCompleted          bool               `json:"is_completed"`
	IsLocked             bool               `json:"is_locked"`
	MissingPrerequisites []PrerequisiteInfo `json:"missing_prerequisites"`
	Prerequisites        []PrerequisiteInfo `json:"prerequisites"`
	PrevLessonSlug       *string            `json:"prev_lesson_slug,omitempty"`
	NextLessonSlug       *string            `json:"next_lesson_slug,omitempty"`
}

// Request and DTO types

type CreateLearningPathRequest struct {
	Title       string `json:"title"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Status      string `json:"status"`
	SortOrder   int    `json:"sort_order"`
}

type UpdateLearningPathRequest struct {
	Title       string `json:"title"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

type UpdateStatusRequest struct {
	Status string `json:"status"`
}

type CreateCourseRequest struct {
	LearningPathID uuid.UUID `json:"learning_path_id"`
	Title          string    `json:"title"`
	Slug           string    `json:"slug"`
	Description    string    `json:"description"`
	Status         string    `json:"status"`
	SortOrder      int       `json:"sort_order"`
}

type UpdateCourseRequest struct {
	Title       string `json:"title"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

type CreateModuleRequest struct {
	CourseID    uuid.UUID `json:"course_id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	SortOrder   int       `json:"sort_order"`
}

type UpdateModuleRequest struct {
	Title       string `json:"title"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

type CreateLessonRequest struct {
	ModuleID         uuid.UUID   `json:"module_id"`
	Title            string      `json:"title"`
	Slug             string      `json:"slug"`
	Summary          string      `json:"summary"`
	Content          string      `json:"content"`
	Difficulty       string      `json:"difficulty"`
	EstimatedMinutes int         `json:"estimated_minutes"`
	Status           string      `json:"status"`
	SortOrder        int         `json:"sort_order"`
	PrerequisiteIDs  []uuid.UUID `json:"prerequisite_ids"`
}

type UpdateLessonRequest struct {
	Title            string      `json:"title"`
	Slug             string      `json:"slug"`
	Summary          string      `json:"summary"`
	Content          string      `json:"content"`
	Difficulty       string      `json:"difficulty"`
	EstimatedMinutes int         `json:"estimated_minutes"`
	SortOrder        int         `json:"sort_order"`
	PrerequisiteIDs  []uuid.UUID `json:"prerequisite_ids"`
}

// Converters from sqlc types

func fromSQLCLearningPath(lp sqlc.LearningPath) LearningPath {
	res := LearningPath{
		ID:        uuid.UUID(lp.ID.Bytes),
		Title:     lp.Title,
		Slug:      lp.Slug,
		Status:    lp.Status,
		SortOrder: int(lp.SortOrder),
		CreatedBy: uuid.UUID(lp.CreatedBy.Bytes),
		CreatedAt: lp.CreatedAt.Time,
		UpdatedAt: lp.UpdatedAt.Time,
	}
	if lp.Description.Valid {
		res.Description = lp.Description.String
	}
	if lp.PublishedAt.Valid {
		t := lp.PublishedAt.Time
		res.PublishedAt = &t
	}
	return res
}

func fromSQLCCourse(c sqlc.Course) Course {
	res := Course{
		ID:             uuid.UUID(c.ID.Bytes),
		LearningPathID: uuid.UUID(c.LearningPathID.Bytes),
		Title:          c.Title,
		Slug:           c.Slug,
		Status:         c.Status,
		SortOrder:      int(c.SortOrder),
		CreatedAt:      c.CreatedAt.Time,
		UpdatedAt:      c.UpdatedAt.Time,
	}
	if c.Description.Valid {
		res.Description = c.Description.String
	}
	if c.PublishedAt.Valid {
		t := c.PublishedAt.Time
		res.PublishedAt = &t
	}
	return res
}

func fromSQLCModule(m sqlc.Module) Module {
	res := Module{
		ID:        uuid.UUID(m.ID.Bytes),
		CourseID:  uuid.UUID(m.CourseID.Bytes),
		Title:     m.Title,
		Slug:      m.Slug,
		SortOrder: int(m.SortOrder),
		CreatedAt: m.CreatedAt.Time,
		UpdatedAt: m.UpdatedAt.Time,
	}
	if m.Description.Valid {
		res.Description = m.Description.String
	}
	return res
}

func fromSQLCLesson(l sqlc.Lesson) Lesson {
	res := Lesson{
		ID:               uuid.UUID(l.ID.Bytes),
		ModuleID:         uuid.UUID(l.ModuleID.Bytes),
		Title:            l.Title,
		Slug:             l.Slug,
		Content:          l.Content,
		Difficulty:       l.Difficulty,
		EstimatedMinutes: int(l.EstimatedMinutes),
		Status:           l.Status,
		SortOrder:        int(l.SortOrder),
		CreatedAt:        l.CreatedAt.Time,
		UpdatedAt:        l.UpdatedAt.Time,
	}
	if l.Summary.Valid {
		res.Summary = l.Summary.String
	}
	if l.PublishedAt.Valid {
		t := l.PublishedAt.Time
		res.PublishedAt = &t
	}
	return res
}

func toPgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func toPgText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}

func toPgTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}
