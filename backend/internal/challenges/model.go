package challenges

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"go-academy/backend/internal/database/sqlc"
)

// Difficulty constants
const (
	DifficultyBeginner     = "beginner"
	DifficultyIntermediate = "intermediate"
	DifficultyAdvanced     = "advanced"
)

// Content status constants
const (
	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusArchived  = "archived"
)

// TestCase model
type TestCase struct {
	ID             uuid.UUID `json:"id"`
	ChallengeID    uuid.UUID `json:"challenge_id"`
	Title          string    `json:"title"`
	Input          string    `json:"input"`
	ExpectedOutput string    `json:"expected_output"`
	Hidden         bool      `json:"hidden"`
	SortOrder      int       `json:"sort_order"`
	CreatedAt      time.Time `json:"created_at"`
}

// ChallengeSummary for catalog listing
type ChallengeSummary struct {
	ID               uuid.UUID  `json:"id"`
	CourseID         *uuid.UUID `json:"course_id,omitempty"`
	LessonID         *uuid.UUID `json:"lesson_id,omitempty"`
	Title            string     `json:"title"`
	Slug             string     `json:"slug"`
	Description      string     `json:"description"`
	Difficulty       string     `json:"difficulty"`
	EstimatedMinutes int        `json:"estimated_minutes"`
	StarterCode      string     `json:"starter_code"`
	Status           string     `json:"status"`
	SortOrder        int        `json:"sort_order"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	PublishedAt      *time.Time `json:"published_at,omitempty"`
}

// ChallengeDetail for learner solving workspace (strictly no solution code or hidden test cases)
type ChallengeDetail struct {
	ChallengeSummary
	SampleTestCases []TestCase `json:"sample_test_cases"`
	CourseTitle     string     `json:"course_title,omitempty"`
	CourseSlug      string     `json:"course_slug,omitempty"`
}

// AdminChallengeDetail contains full metadata, solution code, and all test cases
type AdminChallengeDetail struct {
	ChallengeSummary
	SolutionCode string     `json:"solution_code"`
	TestCases    []TestCase `json:"test_cases"`
}

// Request DTOs

type CreateTestCaseRequest struct {
	Title          string `json:"title"`
	Input          string `json:"input"`
	ExpectedOutput string `json:"expected_output"`
	Hidden         bool   `json:"hidden"`
	SortOrder      int    `json:"sort_order"`
}

type CreateChallengeRequest struct {
	CourseID         *uuid.UUID              `json:"course_id"`
	LessonID         *uuid.UUID              `json:"lesson_id"`
	Title            string                  `json:"title"`
	Slug             string                  `json:"slug"`
	Description      string                  `json:"description"`
	Difficulty       string                  `json:"difficulty"`
	EstimatedMinutes int                     `json:"estimated_minutes"`
	StarterCode      string                  `json:"starter_code"`
	SolutionCode     string                  `json:"solution_code"`
	Status           string                  `json:"status"`
	SortOrder        int                     `json:"sort_order"`
	TestCases        []CreateTestCaseRequest `json:"test_cases"`
}

type UpdateChallengeRequest struct {
	CourseID         *uuid.UUID              `json:"course_id"`
	LessonID         *uuid.UUID              `json:"lesson_id"`
	Title            string                  `json:"title"`
	Slug             string                  `json:"slug"`
	Description      string                  `json:"description"`
	Difficulty       string                  `json:"difficulty"`
	EstimatedMinutes int                     `json:"estimated_minutes"`
	StarterCode      string                  `json:"starter_code"`
	SolutionCode     string                  `json:"solution_code"`
	SortOrder        int                     `json:"sort_order"`
	TestCases        []CreateTestCaseRequest `json:"test_cases"`
}

type UpdateStatusRequest struct {
	Status string `json:"status"`
}

// Helper conversions

func toPgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func toOptionalPgUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func toPgTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func fromSQLCChallenge(c sqlc.Challenge) ChallengeSummary {
	res := ChallengeSummary{
		ID:               uuid.UUID(c.ID.Bytes),
		Title:            c.Title,
		Slug:             c.Slug,
		Description:      c.Description,
		Difficulty:       c.Difficulty,
		EstimatedMinutes: int(c.EstimatedMinutes),
		StarterCode:      c.StarterCode,
		Status:           c.Status,
		SortOrder:        int(c.SortOrder),
		CreatedAt:        c.CreatedAt.Time,
		UpdatedAt:        c.UpdatedAt.Time,
	}
	if c.CourseID.Valid {
		id := uuid.UUID(c.CourseID.Bytes)
		res.CourseID = &id
	}
	if c.LessonID.Valid {
		id := uuid.UUID(c.LessonID.Bytes)
		res.LessonID = &id
	}
	if c.PublishedAt.Valid {
		t := c.PublishedAt.Time
		res.PublishedAt = &t
	}
	return res
}

func fromSQLCTestCase(tc sqlc.ChallengeTestCase) TestCase {
	return TestCase{
		ID:             uuid.UUID(tc.ID.Bytes),
		ChallengeID:    uuid.UUID(tc.ChallengeID.Bytes),
		Title:          tc.Title,
		Input:          tc.Input,
		ExpectedOutput: tc.ExpectedOutput,
		Hidden:         tc.Hidden,
		SortOrder:      int(tc.SortOrder),
		CreatedAt:      tc.CreatedAt.Time,
	}
}
