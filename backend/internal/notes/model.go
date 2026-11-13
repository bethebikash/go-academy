package notes

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"go-academy/backend/internal/database/sqlc"
)

// Note domain model
type Note struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	LessonID    *uuid.UUID `json:"lesson_id,omitempty"`
	LessonTitle string     `json:"lesson_title,omitempty"`
	LessonSlug  string     `json:"lesson_slug,omitempty"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	Pinned      bool       `json:"pinned"`
	Version     int        `json:"version"`
	Tags        []string   `json:"tags"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Tag domain model
type Tag struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	NoteCount int       `json:"note_count"`
	CreatedAt time.Time `json:"created_at"`
}

// Request DTOs

type CreateNoteRequest struct {
	LessonID *uuid.UUID `json:"lesson_id"`
	Title    string     `json:"title"`
	Content  string     `json:"content"`
	Pinned   bool       `json:"pinned"`
	Tags     []string   `json:"tags"`
}

type UpdateNoteRequest struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Pinned  bool     `json:"pinned"`
	Version int      `json:"version"` // Optimistic locking
	Tags    []string `json:"tags"`
}

// Helper conversions

func fromSQLCNote(n sqlc.Note, tags []string) Note {
	res := Note{
		ID:        uuid.UUID(n.ID.Bytes),
		UserID:    uuid.UUID(n.UserID.Bytes),
		Title:     n.Title,
		Content:   n.Content,
		Pinned:    n.Pinned,
		Version:   int(n.Version),
		Tags:      tags,
		CreatedAt: n.CreatedAt.Time,
		UpdatedAt: n.UpdatedAt.Time,
	}
	if n.LessonID.Valid {
		id := uuid.UUID(n.LessonID.Bytes)
		res.LessonID = &id
	}
	if res.Tags == nil {
		res.Tags = []string{}
	}
	return res
}

func toPgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func toOptionalPgUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}
