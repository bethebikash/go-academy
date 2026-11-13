package bookmarks_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"go-academy/backend/internal/bookmarks"
	"go-academy/backend/internal/config"
	"go-academy/backend/internal/content"
	"go-academy/backend/internal/database"
	"go-academy/backend/internal/database/sqlc"
)

func setupTestDB(t *testing.T) (*database.DB, func()) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, cfg)
	if err != nil {
		t.Skipf("skipping live database test: %v", err)
		return nil, nil
	}

	db := database.NewDB(pool)
	cleanup := func() {
		pool.Close()
	}

	return db, cleanup
}

func TestBookmarksLifecycle(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	bookmarkService := bookmarks.NewService(db)
	contentService := content.NewService(db)

	suffix := time.Now().UnixNano()
	userRow, err := db.Queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        fmt.Sprintf("bm_user_%d@example.com", suffix),
		Username:     fmt.Sprintf("bm_usr_%d", suffix),
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
		Title:  "BM Path",
		Status: content.StatusPublished,
	}, userID)
	if err != nil {
		t.Fatalf("failed to create path: %v", err)
	}
	defer contentService.DeleteLearningPath(ctx, path.ID)

	course, err := contentService.CreateCourse(ctx, content.CreateCourseRequest{
		LearningPathID: path.ID,
		Title:          "BM Course",
		Status:         content.StatusPublished,
	})
	if err != nil {
		t.Fatalf("failed to create course: %v", err)
	}

	module, err := contentService.CreateModule(ctx, content.CreateModuleRequest{
		CourseID: course.ID,
		Title:    "BM Module",
	})
	if err != nil {
		t.Fatalf("failed to create module: %v", err)
	}

	lesson, err := contentService.CreateLesson(ctx, content.CreateLessonRequest{
		ModuleID: module.ID,
		Title:    "BM Lesson",
		Content:  "Lesson content",
		Status:   content.StatusPublished,
	})
	if err != nil {
		t.Fatalf("failed to create lesson: %v", err)
	}

	// 1. Initial check: not bookmarked
	isBookmarked, err := bookmarkService.IsLessonBookmarked(ctx, userID, lesson.ID)
	if err != nil {
		t.Fatalf("failed to check bookmark: %v", err)
	}
	if isBookmarked {
		t.Fatal("expected lesson not to be bookmarked initially")
	}

	// 2. Add bookmark
	bm, err := bookmarkService.AddLessonBookmark(ctx, userID, lesson.ID)
	if err != nil {
		t.Fatalf("failed to add bookmark: %v", err)
	}
	if bm.LessonID != lesson.ID || bm.LessonTitle != "BM Lesson" || bm.CourseTitle != "BM Course" {
		t.Fatalf("unexpected bookmark metadata: %+v", bm)
	}

	// 3. Repeated addition is idempotent
	bmRepeat, err := bookmarkService.AddLessonBookmark(ctx, userID, lesson.ID)
	if err != nil {
		t.Fatalf("failed on repeated bookmark add: %v", err)
	}
	if bmRepeat.ID != bm.ID {
		t.Fatalf("expected same bookmark ID on idempotent add")
	}

	// 4. Verify bookmarked status
	isBookmarkedNow, err := bookmarkService.IsLessonBookmarked(ctx, userID, lesson.ID)
	if err != nil || !isBookmarkedNow {
		t.Fatalf("expected lesson to be bookmarked")
	}

	// 5. List bookmarks
	list, err := bookmarkService.ListUserBookmarks(ctx, userID)
	if err != nil {
		t.Fatalf("failed to list bookmarks: %v", err)
	}
	if len(list) != 1 || list[0].ID != bm.ID {
		t.Fatalf("expected 1 bookmark, got %d", len(list))
	}

	// 6. Remove bookmark
	err = bookmarkService.RemoveLessonBookmark(ctx, userID, lesson.ID)
	if err != nil {
		t.Fatalf("failed to remove bookmark: %v", err)
	}

	// 7. Verify removed
	isBookmarkedAfter, err := bookmarkService.IsLessonBookmarked(ctx, userID, lesson.ID)
	if err != nil || isBookmarkedAfter {
		t.Fatal("expected lesson to no longer be bookmarked")
	}
}
