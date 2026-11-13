package content_test

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

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Go Fundamentals", "go-fundamentals"},
		{"Concurrency in Go (Goroutines & Channels)!", "concurrency-in-go-goroutines-channels"},
		{"   Trim  Spaces   ", "trim-spaces"},
		{"Special -- Characters @@ 123", "special-characters-123"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := content.GenerateSlug(tt.input)
			if got != tt.expected {
				t.Fatalf("expected slug %q, got %q", tt.expected, got)
			}
		})
	}
}

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

func TestContentAndCurriculumFlow(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	contentService := content.NewService(db)
	progressService := progress.NewService(db)

	// Create test admin
	suffix := time.Now().UnixNano()
	adminRow, err := db.Queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        fmt.Sprintf("admin_content_%d@example.com", suffix),
		Username:     fmt.Sprintf("admin_%d", suffix),
		PasswordHash: "fakehash",
		Status:       "active",
		Timezone:     "UTC",
	})
	if err != nil {
		t.Fatalf("failed to create admin: %v", err)
	}
	adminUserID := uuid.UUID(adminRow.ID.Bytes)

	// Create test learner
	learnerRow, err := db.Queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        fmt.Sprintf("learner_content_%d@example.com", suffix),
		Username:     fmt.Sprintf("learner_%d", suffix),
		PasswordHash: "fakehash",
		Status:       "active",
		Timezone:     "UTC",
	})
	if err != nil {
		t.Fatalf("failed to create learner: %v", err)
	}
	learnerID := uuid.UUID(learnerRow.ID.Bytes)

	// 1. Create Learning Path
	pathSlug := fmt.Sprintf("test-path-%d", suffix)
	path, err := contentService.CreateLearningPath(ctx, content.CreateLearningPathRequest{
		Title:       "Test Go Path",
		Slug:        pathSlug,
		Description: "Master Go from scratch",
		Status:      content.StatusPublished,
		SortOrder:   1,
	}, adminUserID)
	if err != nil {
		t.Fatalf("failed to create learning path: %v", err)
	}
	if path.Slug != pathSlug {
		t.Fatalf("expected slug %s, got %s", pathSlug, path.Slug)
	}

	// 2. Duplicate slug should fail
	_, err = contentService.CreateLearningPath(ctx, content.CreateLearningPathRequest{
		Title:  "Duplicate Path",
		Slug:   pathSlug,
		Status: content.StatusPublished,
	}, adminUserID)
	if err != content.ErrSlugExists {
		t.Fatalf("expected ErrSlugExists, got %v", err)
	}

	// 3. Create Course under path
	courseSlug := fmt.Sprintf("test-course-%d", suffix)
	course, err := contentService.CreateCourse(ctx, content.CreateCourseRequest{
		LearningPathID: path.ID,
		Title:          "Go Basics",
		Slug:           courseSlug,
		Description:    "Introductory Go Course",
		Status:         content.StatusPublished,
		SortOrder:      1,
	})
	if err != nil {
		t.Fatalf("failed to create course: %v", err)
	}

	// 4. Create Module under course
	moduleSlug := fmt.Sprintf("test-module-%d", suffix)
	module, err := contentService.CreateModule(ctx, content.CreateModuleRequest{
		CourseID:    course.ID,
		Title:       "Getting Started",
		Slug:        moduleSlug,
		Description: "Syntax & Variables",
		SortOrder:   1,
	})
	if err != nil {
		t.Fatalf("failed to create module: %v", err)
	}

	// 5. Create Lesson 1 (Variables)
	lesson1Slug := fmt.Sprintf("lesson-variables-%d", suffix)
	lesson1, err := contentService.CreateLesson(ctx, content.CreateLessonRequest{
		ModuleID:         module.ID,
		Title:            "Variables and Types",
		Slug:             lesson1Slug,
		Summary:          "Learn how variables work in Go.",
		Content:          "# Variables\nIn Go, variables are declared with `var` or `:=`.",
		Difficulty:       content.DifficultyBeginner,
		EstimatedMinutes: 10,
		Status:           content.StatusPublished,
		SortOrder:        1,
	})
	if err != nil {
		t.Fatalf("failed to create lesson 1: %v", err)
	}

	// 6. Create Lesson 2 (Control Flow) with Lesson 1 as prerequisite
	lesson2Slug := fmt.Sprintf("lesson-control-flow-%d", suffix)
	lesson2, err := contentService.CreateLesson(ctx, content.CreateLessonRequest{
		ModuleID:         module.ID,
		Title:            "Control Flow",
		Slug:             lesson2Slug,
		Summary:          "If/else and loops in Go.",
		Content:          "# Control Flow\nGo only has one loop keyword: `for`.",
		Difficulty:       content.DifficultyBeginner,
		EstimatedMinutes: 15,
		Status:           content.StatusPublished,
		SortOrder:        2,
		PrerequisiteIDs:  []uuid.UUID{lesson1.ID},
	})
	if err != nil {
		t.Fatalf("failed to create lesson 2: %v", err)
	}

	// 7. Verify Lesson 2 is locked for learner because Lesson 1 is not completed
	lesson2Detail, err := contentService.GetLessonBySlug(ctx, lesson2Slug, &learnerID)
	if err != nil {
		t.Fatalf("failed to get lesson 2 detail: %v", err)
	}
	if !lesson2Detail.IsLocked {
		t.Fatalf("expected lesson 2 to be locked, but IsLocked was false")
	}
	if len(lesson2Detail.MissingPrerequisites) != 1 || lesson2Detail.MissingPrerequisites[0].ID != lesson1.ID {
		t.Fatalf("expected missing prerequisite %s, got %+v", lesson1.ID, lesson2Detail.MissingPrerequisites)
	}

	// 8. Attempting to mark Lesson 2 as completed should fail with unmet prerequisites
	_, err = progressService.UpdateLessonProgress(ctx, learnerID, lesson2.ID, "completed")
	if err != progress.ErrPrerequisitesUnmet {
		t.Fatalf("expected ErrPrerequisitesUnmet, got %v", err)
	}

	// 9. Mark Lesson 1 as completed
	prog1, err := progressService.UpdateLessonProgress(ctx, learnerID, lesson1.ID, "completed")
	if err != nil {
		t.Fatalf("failed to complete lesson 1: %v", err)
	}
	if prog1.Status != "completed" || prog1.CompletedAt == nil {
		t.Fatalf("expected completed status with timestamp, got %+v", prog1)
	}

	// 10. Verify Lesson 2 is now unlocked for learner
	lesson2DetailAfter, err := contentService.GetLessonBySlug(ctx, lesson2Slug, &learnerID)
	if err != nil {
		t.Fatalf("failed to get lesson 2 detail after prereq: %v", err)
	}
	if lesson2DetailAfter.IsLocked {
		t.Fatalf("expected lesson 2 to be unlocked, but IsLocked was true")
	}

	// 11. Now learner can complete Lesson 2
	prog2, err := progressService.UpdateLessonProgress(ctx, learnerID, lesson2.ID, "completed")
	if err != nil {
		t.Fatalf("failed to complete lesson 2: %v", err)
	}
	if prog2.Status != "completed" {
		t.Fatalf("expected lesson 2 to be completed")
	}

	// 12. Check course syllabus progress
	courseDetail, err := contentService.GetCourseBySlug(ctx, courseSlug, &learnerID)
	if err != nil {
		t.Fatalf("failed to get course syllabus: %v", err)
	}
	if courseDetail.TotalLessons != 2 || courseDetail.CompletedLessons != 2 || courseDetail.ProgressPercent != 100 {
		t.Fatalf("expected 2/2 lessons (100%%), got %d/%d (%d%%)",
			courseDetail.CompletedLessons, courseDetail.TotalLessons, courseDetail.ProgressPercent)
	}

	// Clean up
	_ = contentService.DeleteLearningPath(ctx, path.ID)
}
