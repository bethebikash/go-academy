package analytics_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"go-academy/backend/internal/analytics"
	"go-academy/backend/internal/config"
	"go-academy/backend/internal/content"
	"go-academy/backend/internal/database"
	"go-academy/backend/internal/database/sqlc"
	"go-academy/backend/internal/progress"
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

func TestActivityAndStreakCalculation(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	svc := analytics.NewService(db)

	suffix := time.Now().UnixNano()
	userRow, err := db.Queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        fmt.Sprintf("streak_user_%d@example.com", suffix),
		Username:     fmt.Sprintf("streak_usr_%d", suffix),
		PasswordHash: "fakehash",
		Status:       "active",
		Timezone:     "UTC",
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	userID := uuid.UUID(userRow.ID.Bytes)

	// Day 1 Activity
	err = svc.RecordActivity(ctx, userID, analytics.ActivityLessonCompleted, map[string]interface{}{"num": 1})
	if err != nil {
		t.Fatalf("failed to record activity: %v", err)
	}

	stats, err := svc.GetSummaryStats(ctx, userID)
	if err != nil {
		t.Fatalf("failed to get summary stats: %v", err)
	}
	if stats.Streak.CurrentStreak != 1 {
		t.Errorf("expected current streak 1, got %d", stats.Streak.CurrentStreak)
	}
	if stats.Streak.LongestStreak != 1 {
		t.Errorf("expected longest streak 1, got %d", stats.Streak.LongestStreak)
	}

	// Same Day Second Activity - Streak should remain 1
	err = svc.RecordActivity(ctx, userID, analytics.ActivityChallengePassed, map[string]interface{}{"num": 2})
	if err != nil {
		t.Fatalf("failed to record second activity: %v", err)
	}

	stats, err = svc.GetSummaryStats(ctx, userID)
	if err != nil {
		t.Fatalf("failed to get summary stats: %v", err)
	}
	if stats.Streak.CurrentStreak != 1 {
		t.Errorf("expected current streak to stay 1 for same-day activity, got %d", stats.Streak.CurrentStreak)
	}

	// Check badges: should have both First Step and Bug Hunter
	badges, err := svc.GetUserBadges(ctx, userID)
	if err != nil {
		t.Fatalf("failed to get badges: %v", err)
	}

	earned := make(map[string]bool)
	for _, b := range badges {
		if b.Awarded {
			earned[b.Key] = true
		}
	}

	if !earned[analytics.BadgeFirstStep] {
		t.Errorf("expected BadgeFirstStep to be awarded")
	}
	if !earned[analytics.BadgeBugHunter] {
		t.Errorf("expected BadgeBugHunter to be awarded")
	}
}

func TestCertificateIssuanceAndVerification(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	analyticsSvc := analytics.NewService(db)
	contentSvc := content.NewService(db)
	progressSvc := progress.NewService(db)

	suffix := time.Now().UnixNano()
	adminRow, err := db.Queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        fmt.Sprintf("cert_admin_%d@example.com", suffix),
		Username:     fmt.Sprintf("cert_adm_%d", suffix),
		PasswordHash: "fakehash",
		Status:       "active",
		Timezone:     "UTC",
	})
	if err != nil {
		t.Fatalf("failed to create admin: %v", err)
	}
	adminID := uuid.UUID(adminRow.ID.Bytes)

	learnerRow, err := db.Queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        fmt.Sprintf("cert_learner_%d@example.com", suffix),
		Username:     fmt.Sprintf("cert_lrn_%d", suffix),
		PasswordHash: "fakehash",
		Status:       "active",
		Timezone:     "UTC",
	})
	if err != nil {
		t.Fatalf("failed to create learner: %v", err)
	}
	learnerID := uuid.UUID(learnerRow.ID.Bytes)

	// Create Path, Course, Module, 1 Lesson
	path, err := contentSvc.CreateLearningPath(ctx, content.CreateLearningPathRequest{
		Title:       fmt.Sprintf("Cert Path %d", suffix),
		Description: "Certificate test path",
		Status:      content.StatusPublished,
		SortOrder:   1,
	}, adminID)
	if err != nil {
		t.Fatalf("failed to create path: %v", err)
	}

	course, err := contentSvc.CreateCourse(ctx, content.CreateCourseRequest{
		LearningPathID: path.ID,
		Title:          fmt.Sprintf("Cert Course %d", suffix),
		Description:    "Cert Course Desc",
		Status:         content.StatusPublished,
		SortOrder:      1,
	})
	if err != nil {
		t.Fatalf("failed to create course: %v", err)
	}

	mod, err := contentSvc.CreateModule(ctx, content.CreateModuleRequest{
		CourseID:    course.ID,
		Title:       fmt.Sprintf("Cert Module %d", suffix),
		Description: "Cert Module Desc",
		SortOrder:   1,
	})
	if err != nil {
		t.Fatalf("failed to create module: %v", err)
	}

	lesson, err := contentSvc.CreateLesson(ctx, content.CreateLessonRequest{
		ModuleID:         mod.ID,
		Title:            fmt.Sprintf("Cert Lesson %d", suffix),
		Content:          "# Content",
		Difficulty:       content.DifficultyBeginner,
		EstimatedMinutes: 10,
		Status:           content.StatusPublished,
		SortOrder:        1,
	})
	if err != nil {
		t.Fatalf("failed to create lesson: %v", err)
	}

	// 1. Attempt to claim certificate before completing lesson (Must fail)
	_, err = analyticsSvc.ClaimCertificate(ctx, learnerID, path.ID)
	if !errors.Is(err, analytics.ErrPathNotCompleted) {
		t.Fatalf("expected ErrPathNotCompleted, got: %v", err)
	}

	// 2. Complete the lesson
	_, err = progressSvc.UpdateLessonProgress(ctx, learnerID, lesson.ID, "completed")
	if err != nil {
		t.Fatalf("failed to complete lesson: %v", err)
	}

	// 3. Claim Certificate (Must succeed)
	cert, err := analyticsSvc.ClaimCertificate(ctx, learnerID, path.ID)
	if err != nil {
		t.Fatalf("failed to claim certificate after 100%% completion: %v", err)
	}
	if cert.CertificateCode == "" {
		t.Fatal("expected non-empty certificate code")
	}
	if cert.RecipientName != learnerRow.Username {
		t.Errorf("expected recipient %s, got %s", learnerRow.Username, cert.RecipientName)
	}
	if cert.LearningPathTitle != path.Title {
		t.Errorf("expected path title %s, got %s", path.Title, cert.LearningPathTitle)
	}

	// 4. Verify certificate by code (Public verification)
	verified, err := analyticsSvc.GetCertificateByCode(ctx, cert.CertificateCode)
	if err != nil {
		t.Fatalf("failed to publicly verify certificate: %v", err)
	}
	if verified.ID != cert.ID {
		t.Errorf("expected matching certificate ID")
	}

	// 5. Check user badges - Path Master badge should be unlocked
	badges, err := analyticsSvc.GetUserBadges(ctx, learnerID)
	if err != nil {
		t.Fatalf("failed to get badges: %v", err)
	}
	var pathMasterAwarded bool
	for _, b := range badges {
		if b.Key == analytics.BadgePathMaster && b.Awarded {
			pathMasterAwarded = true
			break
		}
	}
	if !pathMasterAwarded {
		t.Error("expected BadgePathMaster to be awarded upon path completion")
	}
}
