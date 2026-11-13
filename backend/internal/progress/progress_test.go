package progress_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"go-academy/backend/internal/config"
	"go-academy/backend/internal/content"
	"go-academy/backend/internal/database"
	"go-academy/backend/internal/database/sqlc"
	"go-academy/backend/internal/progress"
)

func setupTestDB(t *testing.T) (*database.DB, *config.Config, func()) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, cfg)
	if err != nil {
		t.Skipf("skipping live database test: %v", err)
		return nil, nil, nil
	}

	db := database.NewDB(pool)
	cleanup := func() {
		pool.Close()
	}

	return db, cfg, cleanup
}

func TestLessonProgressIdempotencyAndValidation(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	contentService := content.NewService(db)
	progressService := progress.NewService(db)

	suffix := time.Now().UnixNano()
	userRow, err := db.Queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        fmt.Sprintf("progress_user_%d@example.com", suffix),
		Username:     fmt.Sprintf("prog_usr_%d", suffix),
		PasswordHash: "fakehash",
		Status:       "active",
		Timezone:     "UTC",
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	userID := uuid.UUID(userRow.ID.Bytes)

	// Setup hierarchy: Path -> Course -> Module -> Lesson
	path, err := contentService.CreateLearningPath(ctx, content.CreateLearningPathRequest{
		Title:  "Progress Test Path",
		Status: content.StatusPublished,
	}, userID)
	if err != nil {
		t.Fatalf("failed to create path: %v", err)
	}
	defer contentService.DeleteLearningPath(ctx, path.ID)

	course, err := contentService.CreateCourse(ctx, content.CreateCourseRequest{
		LearningPathID: path.ID,
		Title:          "Progress Course",
		Status:         content.StatusPublished,
	})
	if err != nil {
		t.Fatalf("failed to create course: %v", err)
	}

	module, err := contentService.CreateModule(ctx, content.CreateModuleRequest{
		CourseID: course.ID,
		Title:    "Progress Module",
	})
	if err != nil {
		t.Fatalf("failed to create module: %v", err)
	}

	lesson, err := contentService.CreateLesson(ctx, content.CreateLessonRequest{
		ModuleID: module.ID,
		Title:    "Lesson 1",
		Content:  "Content 1",
		Status:   content.StatusPublished,
	})
	if err != nil {
		t.Fatalf("failed to create lesson: %v", err)
	}

	// 1. Invalid status
	_, err = progressService.UpdateLessonProgress(ctx, userID, lesson.ID, "unknown_status")
	if err == nil {
		t.Fatal("expected error for invalid status")
	}

	// 2. Mark started
	p1, err := progressService.UpdateLessonProgress(ctx, userID, lesson.ID, "started")
	if err != nil {
		t.Fatalf("failed to mark started: %v", err)
	}
	if p1.Status != "started" || p1.CompletedAt != nil {
		t.Fatalf("unexpected progress state: %+v", p1)
	}

	// 3. Mark completed
	p2, err := progressService.UpdateLessonProgress(ctx, userID, lesson.ID, "completed")
	if err != nil {
		t.Fatalf("failed to mark completed: %v", err)
	}
	if p2.Status != "completed" || p2.CompletedAt == nil {
		t.Fatalf("expected completed with timestamp, got %+v", p2)
	}
	firstCompletedAt := *p2.CompletedAt

	// 4. Repeated complete call must be idempotent and preserve initial completed_at
	time.Sleep(10 * time.Millisecond)
	p3, err := progressService.UpdateLessonProgress(ctx, userID, lesson.ID, "completed")
	if err != nil {
		t.Fatalf("failed on repeated complete: %v", err)
	}
	if p3.Status != "completed" {
		t.Fatalf("status should remain completed")
	}
	if !p3.CompletedAt.Equal(firstCompletedAt) {
		t.Fatalf("completed_at was overwritten! original: %v, new: %v", firstCompletedAt, p3.CompletedAt)
	}

	// 5. Query progress directly
	pQuery, err := progressService.GetLessonProgress(ctx, userID, lesson.ID)
	if err != nil {
		t.Fatalf("failed to get lesson progress: %v", err)
	}
	if pQuery.ID != p2.ID {
		t.Fatalf("expected ID %s, got %s", p2.ID, pQuery.ID)
	}
}
