package progress

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"go-academy/backend/internal/database"
	"go-academy/backend/internal/database/sqlc"
)

var (
	ErrNotFound           = errors.New("progress not found")
	ErrInvalidInput       = errors.New("invalid input")
	ErrPrerequisitesUnmet = errors.New("cannot complete lesson: unmet prerequisites")
)

type Progress struct {
	ID             uuid.UUID  `json:"id"`
	UserID         uuid.UUID  `json:"user_id"`
	LessonID       uuid.UUID  `json:"lesson_id"`
	Status         string     `json:"status"` // started, completed
	StartedAt      time.Time  `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	LastAccessedAt time.Time  `json:"last_accessed_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type ActivityRecorder interface {
	RecordActivity(ctx context.Context, userID uuid.UUID, activityType string, metadata map[string]interface{}) error
}

type Service struct {
	db       *database.DB
	recorder ActivityRecorder
}

func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

func (s *Service) SetActivityRecorder(r ActivityRecorder) {
	s.recorder = r
}

func (s *Service) GetLessonProgress(ctx context.Context, userID, lessonID uuid.UUID) (*Progress, error) {
	row, err := s.db.Queries.GetLessonProgress(ctx, sqlc.GetLessonProgressParams{
		UserID:   toPgUUID(userID),
		LessonID: toPgUUID(lessonID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get lesson progress: %w", err)
	}

	p := fromSQLCLessonProgress(row)
	return &p, nil
}

func (s *Service) UpdateLessonProgress(ctx context.Context, userID, lessonID uuid.UUID, status string) (*Progress, error) {
	if status != "started" && status != "completed" {
		return nil, fmt.Errorf("%w: status must be started or completed", ErrInvalidInput)
	}

	// If attempting to mark as completed, enforce that all prerequisites are completed
	if status == "completed" {
		uncompleted, err := s.db.Queries.GetUncompletedPrerequisitesForUser(ctx, sqlc.GetUncompletedPrerequisitesForUserParams{
			LessonID: toPgUUID(lessonID),
			UserID:   toPgUUID(userID),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to verify prerequisites: %w", err)
		}
		if len(uncompleted) > 0 {
			return nil, ErrPrerequisitesUnmet
		}
	}

	var completedAt pgtype.Timestamptz
	if status == "completed" {
		completedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
	}

	row, err := s.db.Queries.UpsertLessonProgress(ctx, sqlc.UpsertLessonProgressParams{
		UserID:      toPgUUID(userID),
		LessonID:    toPgUUID(lessonID),
		Status:      status,
		CompletedAt: completedAt,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upsert lesson progress: %w", err)
	}

	p := fromSQLCLessonProgress(row)

	if status == "completed" && s.recorder != nil {
		_ = s.recorder.RecordActivity(ctx, userID, "lesson_completed", map[string]interface{}{
			"lesson_id": lessonID.String(),
		})
	}

	return &p, nil
}

func fromSQLCLessonProgress(lp sqlc.LessonProgress) Progress {
	p := Progress{
		ID:             uuid.UUID(lp.ID.Bytes),
		UserID:         uuid.UUID(lp.UserID.Bytes),
		LessonID:       uuid.UUID(lp.LessonID.Bytes),
		Status:         lp.Status,
		StartedAt:      lp.StartedAt.Time,
		LastAccessedAt: lp.LastAccessedAt.Time,
		UpdatedAt:      lp.UpdatedAt.Time,
	}
	if lp.CompletedAt.Valid {
		t := lp.CompletedAt.Time
		p.CompletedAt = &t
	}
	return p
}

func toPgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}
