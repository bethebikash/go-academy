package notes

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"go-academy/backend/internal/database"
	"go-academy/backend/internal/database/sqlc"
)

var (
	ErrNotFound        = errors.New("note not found")
	ErrVersionConflict = errors.New("note has been modified by another process")
	ErrInvalidInput    = errors.New("invalid input")
)

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

// CreateNote creates a new personal note and associates any provided tags.
func (s *Service) CreateNote(ctx context.Context, userID uuid.UUID, req CreateNoteRequest) (*Note, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, fmt.Errorf("%w: title cannot be empty", ErrInvalidInput)
	}

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.db.Queries.WithTx(tx)

	created, err := qtx.CreateNote(ctx, sqlc.CreateNoteParams{
		UserID:   toPgUUID(userID),
		LessonID: toOptionalPgUUID(req.LessonID),
		Title:    title,
		Content:  req.Content,
		Pinned:   req.Pinned,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	// Process and link tags
	tagNames := make([]string, 0, len(req.Tags))
	for _, rawTag := range req.Tags {
		cleanTag := strings.ToLower(strings.TrimSpace(rawTag))
		if cleanTag == "" {
			continue
		}
		tagRow, err := qtx.UpsertTag(ctx, sqlc.UpsertTagParams{
			UserID: toPgUUID(userID),
			Name:   cleanTag,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to upsert tag %q: %w", cleanTag, err)
		}

		err = qtx.AddNoteTag(ctx, sqlc.AddNoteTagParams{
			NoteID: created.ID,
			TagID:  tagRow.ID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to link tag to note: %w", err)
		}
		tagNames = append(tagNames, cleanTag)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	res := fromSQLCNote(created, tagNames)

	if s.recorder != nil {
		_ = s.recorder.RecordActivity(ctx, userID, "note_created", map[string]interface{}{
			"note_id": res.ID.String(),
		})
	}

	return &res, nil
}

// GetNoteByID fetches a single note by ID with its tags and optional lesson context.
func (s *Service) GetNoteByID(ctx context.Context, userID, noteID uuid.UUID) (*Note, error) {
	noteRow, err := s.db.Queries.GetNoteByID(ctx, sqlc.GetNoteByIDParams{
		ID:     toPgUUID(noteID),
		UserID: toPgUUID(userID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	tags, err := s.getTagsForNote(ctx, noteRow.ID)
	if err != nil {
		return nil, err
	}

	res := fromSQLCNote(noteRow, tags)
	if noteRow.LessonID.Valid {
		lesson, err := s.db.Queries.GetLessonByID(ctx, noteRow.LessonID)
		if err == nil {
			res.LessonTitle = lesson.Title
			res.LessonSlug = lesson.Slug
		}
	}

	return &res, nil
}

// ListNotes lists notes for user with optional tag, lesson_id, or full-text query filtering.
func (s *Service) ListNotes(ctx context.Context, userID uuid.UUID, tag string, lessonID *uuid.UUID, query string) ([]Note, error) {
	var noteRows []sqlc.Note
	var err error

	cleanTag := strings.TrimSpace(tag)
	cleanQuery := strings.TrimSpace(query)

	if cleanTag != "" {
		noteRows, err = s.db.Queries.ListNotesByTag(ctx, sqlc.ListNotesByTagParams{
			UserID: toPgUUID(userID),
			Name:   cleanTag,
		})
	} else if cleanQuery != "" {
		noteRows, err = s.db.Queries.SearchNotes(ctx, sqlc.SearchNotesParams{
			UserID:  toPgUUID(userID),
			PlaintoTsquery: cleanQuery,
		})
	} else if lessonID != nil {
		noteRows, err = s.db.Queries.ListNotesByLessonID(ctx, sqlc.ListNotesByLessonIDParams{
			UserID:   toPgUUID(userID),
			LessonID: toPgUUID(*lessonID),
		})
	} else {
		noteRows, err = s.db.Queries.ListNotesByUserID(ctx, toPgUUID(userID))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list notes: %w", err)
	}

	notes := make([]Note, 0, len(noteRows))
	for _, n := range noteRows {
		tags, _ := s.getTagsForNote(ctx, n.ID)
		note := fromSQLCNote(n, tags)
		if n.LessonID.Valid {
			lesson, err := s.db.Queries.GetLessonByID(ctx, n.LessonID)
			if err == nil {
				note.LessonTitle = lesson.Title
				note.LessonSlug = lesson.Slug
			}
		}
		notes = append(notes, note)
	}

	return notes, nil
}

// UpdateNote updates a note with optimistic concurrency check.
func (s *Service) UpdateNote(ctx context.Context, userID, noteID uuid.UUID, req UpdateNoteRequest) (*Note, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, fmt.Errorf("%w: title cannot be empty", ErrInvalidInput)
	}

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.db.Queries.WithTx(tx)

	updated, err := qtx.UpdateNote(ctx, sqlc.UpdateNoteParams{
		ID:      toPgUUID(noteID),
		UserID:  toPgUUID(userID),
		Version: int32(req.Version),
		Title:   title,
		Content: req.Content,
		Pinned:  req.Pinned,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Check if note exists to differentiate between 404 and 409
			_, checkErr := qtx.GetNoteByID(ctx, sqlc.GetNoteByIDParams{
				ID:     toPgUUID(noteID),
				UserID: toPgUUID(userID),
			})
			if errors.Is(checkErr, pgx.ErrNoRows) {
				return nil, ErrNotFound
			}
			return nil, ErrVersionConflict
		}
		return nil, fmt.Errorf("failed to update note: %w", err)
	}

	// Update tags if provided
	if req.Tags != nil {
		if err := qtx.ClearNoteTags(ctx, toPgUUID(noteID)); err != nil {
			return nil, fmt.Errorf("failed to clear note tags: %w", err)
		}

		for _, rawTag := range req.Tags {
			cleanTag := strings.ToLower(strings.TrimSpace(rawTag))
			if cleanTag == "" {
				continue
			}
			tagRow, err := qtx.UpsertTag(ctx, sqlc.UpsertTagParams{
				UserID: toPgUUID(userID),
				Name:   cleanTag,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to upsert tag: %w", err)
			}
			if err := qtx.AddNoteTag(ctx, sqlc.AddNoteTagParams{
				NoteID: toPgUUID(noteID),
				TagID:  tagRow.ID,
			}); err != nil {
				return nil, fmt.Errorf("failed to link tag: %w", err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	tags, _ := s.getTagsForNote(ctx, toPgUUID(noteID))
	res := fromSQLCNote(updated, tags)
	return &res, nil
}

// DeleteNote deletes a note by ID.
func (s *Service) DeleteNote(ctx context.Context, userID, noteID uuid.UUID) error {
	err := s.db.Queries.DeleteNote(ctx, sqlc.DeleteNoteParams{
		ID:     toPgUUID(noteID),
		UserID: toPgUUID(userID),
	})
	if err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}
	return nil
}

// ListTags lists all tags created by the user with active note counts.
func (s *Service) ListTags(ctx context.Context, userID uuid.UUID) ([]Tag, error) {
	tagRows, err := s.db.Queries.ListUserTags(ctx, toPgUUID(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}

	tags := make([]Tag, len(tagRows))
	for i, t := range tagRows {
		tags[i] = Tag{
			ID:        uuid.UUID(t.ID.Bytes),
			Name:      t.Name,
			NoteCount: int(t.NoteCount),
			CreatedAt: t.CreatedAt.Time,
		}
	}
	return tags, nil
}

func (s *Service) getTagsForNote(ctx context.Context, noteID pgtype.UUID) ([]string, error) {
	tagRows, err := s.db.Queries.GetNoteTags(ctx, noteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags for note: %w", err)
	}

	tagNames := make([]string, len(tagRows))
	for i, t := range tagRows {
		tagNames[i] = t.Name
	}
	return tagNames, nil
}
