package bookmarks

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"go-academy/backend/internal/database"
	"go-academy/backend/internal/database/sqlc"
)

var (
	ErrNotFound = errors.New("bookmark not found")
)

type Service struct {
	db *database.DB
}

func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

// AddLessonBookmark adds a lesson to user's bookmarks idempotently.
func (s *Service) AddLessonBookmark(ctx context.Context, userID, lessonID uuid.UUID) (*Bookmark, error) {
	_, err := s.db.Queries.CreateLessonBookmark(ctx, sqlc.CreateLessonBookmarkParams{
		UserID:   toPgUUID(userID),
		LessonID: toPgUUID(lessonID),
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to create lesson bookmark: %w", err)
	}

	// Retrieve full bookmark details
	bookmarks, err := s.ListUserBookmarks(ctx, userID)
	if err != nil {
		return nil, err
	}

	for _, b := range bookmarks {
		if b.LessonID == lessonID {
			return &b, nil
		}
	}

	return nil, ErrNotFound
}

// ListUserBookmarks returns all lesson bookmarks for a user with rich context.
func (s *Service) ListUserBookmarks(ctx context.Context, userID uuid.UUID) ([]Bookmark, error) {
	rows, err := s.db.Queries.ListUserLessonBookmarks(ctx, toPgUUID(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to list user bookmarks: %w", err)
	}

	bookmarks := make([]Bookmark, len(rows))
	for i, r := range rows {
		bookmarks[i] = Bookmark{
			ID:                     uuid.UUID(r.ID.Bytes),
			UserID:                 uuid.UUID(r.UserID.Bytes),
			LessonID:               uuid.UUID(r.LessonID.Bytes),
			LessonTitle:            r.LessonTitle,
			LessonSlug:             r.LessonSlug,
			LessonSummary:          r.LessonSummary,
			LessonDifficulty:       r.LessonDifficulty,
			LessonEstimatedMinutes: int(r.LessonEstimatedMinutes),
			CourseTitle:            r.CourseTitle,
			CourseSlug:             r.CourseSlug,
			ProgressStatus:         r.ProgressStatus,
			CreatedAt:              r.CreatedAt.Time,
		}
	}

	return bookmarks, nil
}

// IsLessonBookmarked checks if a lesson is currently bookmarked by the user.
func (s *Service) IsLessonBookmarked(ctx context.Context, userID, lessonID uuid.UUID) (bool, error) {
	_, err := s.db.Queries.GetLessonBookmark(ctx, sqlc.GetLessonBookmarkParams{
		UserID:   toPgUUID(userID),
		LessonID: toPgUUID(lessonID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check bookmark: %w", err)
	}
	return true, nil
}

// RemoveLessonBookmark removes a bookmark by user and lesson ID.
func (s *Service) RemoveLessonBookmark(ctx context.Context, userID, lessonID uuid.UUID) error {
	err := s.db.Queries.DeleteLessonBookmark(ctx, sqlc.DeleteLessonBookmarkParams{
		UserID:   toPgUUID(userID),
		LessonID: toPgUUID(lessonID),
	})
	if err != nil {
		return fmt.Errorf("failed to remove lesson bookmark: %w", err)
	}
	return nil
}

// RemoveBookmarkByID removes a bookmark by its primary key ID and user ID.
func (s *Service) RemoveBookmarkByID(ctx context.Context, userID, bookmarkID uuid.UUID) error {
	err := s.db.Queries.DeleteBookmarkByID(ctx, sqlc.DeleteBookmarkByIDParams{
		ID:     toPgUUID(bookmarkID),
		UserID: toPgUUID(userID),
	})
	if err != nil {
		return fmt.Errorf("failed to delete bookmark: %w", err)
	}
	return nil
}
