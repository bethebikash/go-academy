package challenges_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"go-academy/backend/internal/challenges"
	"go-academy/backend/internal/config"
	"go-academy/backend/internal/database"
	"go-academy/backend/internal/database/sqlc"
	"go-academy/backend/internal/runner"

	"github.com/google/uuid"
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

func TestChallengeLifecycleAndSecurityBoundary(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	codeRunner := runner.NewProcessRunner(5 * time.Second)
	svc := challenges.NewService(db, codeRunner)

	suffix := time.Now().UnixNano()
	challengeTitle := fmt.Sprintf("Challenge Security Test %d", suffix)
	challengeSlug := fmt.Sprintf("challenge-security-test-%d", suffix)

	// 1. Admin creates draft challenge with 1 sample test case and 1 hidden test case
	created, err := svc.CreateChallenge(ctx, challenges.CreateChallengeRequest{
		Title:            challengeTitle,
		Slug:             challengeSlug,
		Description:      "Solve this algorithm problem securely.",
		Difficulty:       challenges.DifficultyIntermediate,
		EstimatedMinutes: 20,
		StarterCode:      "package main\n\nfunc Solve(n int) int {\n\treturn 0\n}\n",
		SolutionCode:     "package main\n\nfunc Solve(n int) int {\n\treturn n * 2\n}\n",
		Status:           challenges.StatusDraft,
		SortOrder:        1,
		TestCases: []challenges.CreateTestCaseRequest{
			{
				Title:          "Sample: n = 2",
				Input:          "2",
				ExpectedOutput: "4",
				Hidden:         false,
				SortOrder:      1,
			},
			{
				Title:          "Hidden: n = 100",
				Input:          "100",
				ExpectedOutput: "200",
				Hidden:         true,
				SortOrder:      2,
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create challenge: %v", err)
	}
	if created.ID.String() == "" {
		t.Fatalf("expected non-empty challenge ID")
	}

	// 2. Admin fetches challenge by ID - must see solution code and all test cases
	adminDetail, err := svc.GetChallengeByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("admin failed to get challenge by ID: %v", err)
	}
	if adminDetail.SolutionCode == "" {
		t.Fatalf("expected admin to see solution code")
	}
	if len(adminDetail.TestCases) != 2 {
		t.Fatalf("expected admin to see 2 test cases, got %d", len(adminDetail.TestCases))
	}

	// 3. While in draft status, learner cannot see it
	_, err = svc.GetPublishedChallengeBySlug(ctx, challengeSlug)
	if !errors.Is(err, challenges.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for draft challenge via learner slug query, got: %v", err)
	}

	// 4. Publish the challenge
	_, err = svc.UpdateChallengeStatus(ctx, created.ID, challenges.StatusPublished)
	if err != nil {
		t.Fatalf("failed to publish challenge: %v", err)
	}

	// 5. Learner fetches published challenge by slug
	learnerDetail, err := svc.GetPublishedChallengeBySlug(ctx, challengeSlug)
	if err != nil {
		t.Fatalf("learner failed to get published challenge: %v", err)
	}
	if learnerDetail.Title != challengeTitle {
		t.Errorf("expected title %q, got %q", challengeTitle, learnerDetail.Title)
	}
	// Security check: Only visible test cases (Hidden == false) must be returned
	if len(learnerDetail.SampleTestCases) != 1 {
		t.Fatalf("expected exactly 1 sample test case for learner, got %d", len(learnerDetail.SampleTestCases))
	}
	if learnerDetail.SampleTestCases[0].Hidden {
		t.Errorf("hidden test case leaked to learner!")
	}
	if learnerDetail.SampleTestCases[0].Title != "Sample: n = 2" {
		t.Errorf("unexpected test case title: %s", learnerDetail.SampleTestCases[0].Title)
	}

	// 6. Test slug uniqueness check
	_, err = svc.CreateChallenge(ctx, challenges.CreateChallengeRequest{
		Title:       "Duplicate Slug Challenge",
		Slug:        challengeSlug,
		Description: "Duplicate slug should fail",
	})
	if !errors.Is(err, challenges.ErrSlugExists) {
		t.Fatalf("expected ErrSlugExists, got: %v", err)
	}

	// 7. Update challenge metadata and test cases
	updated, err := svc.UpdateChallenge(ctx, created.ID, challenges.UpdateChallengeRequest{
		Title:            challengeTitle + " Updated",
		Slug:             challengeSlug,
		Description:      "Updated problem description",
		Difficulty:       challenges.DifficultyAdvanced,
		EstimatedMinutes: 30,
		StarterCode:      "package main\n\nfunc Solve(n int) int {\n\treturn 1\n}\n",
		SolutionCode:     "package main\n\nfunc Solve(n int) int {\n\treturn n * 3\n}\n",
		SortOrder:        2,
		TestCases: []challenges.CreateTestCaseRequest{
			{
				Title:          "Sample 1",
				Input:          "5",
				ExpectedOutput: "15",
				Hidden:         false,
				SortOrder:      1,
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to update challenge: %v", err)
	}
	if updated.Title != challengeTitle+" Updated" {
		t.Errorf("expected updated title, got %q", updated.Title)
	}
	if len(updated.TestCases) != 1 {
		t.Errorf("expected 1 test case after update, got %d", len(updated.TestCases))
	}

	// 8. Delete challenge
	err = svc.DeleteChallenge(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to delete challenge: %v", err)
	}

	// Verify deleted
	_, err = svc.GetChallengeByID(ctx, created.ID)
	if !errors.Is(err, challenges.ErrNotFound) {
		t.Fatalf("expected ErrNotFound after deletion, got: %v", err)
	}
}

func TestChallengeSubmissionsAndExecutionFlow(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	codeRunner := runner.NewProcessRunner(5 * time.Second)
	svc := challenges.NewService(db, codeRunner)

	// Create test user
	suffix := time.Now().UnixNano()
	userRow, err := db.Queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        fmt.Sprintf("sub_user_%d@example.com", suffix),
		Username:     fmt.Sprintf("sub_usr_%d", suffix),
		PasswordHash: "fakehash",
		Status:       "active",
		Timezone:     "UTC",
	})
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	userID := uuid.UUID(userRow.ID.Bytes)

	// Admin creates TwoSum challenge
	slug := fmt.Sprintf("two-sum-test-%d", suffix)
	created, err := svc.CreateChallenge(ctx, challenges.CreateChallengeRequest{
		Title:            "Two Sum Test",
		Slug:             slug,
		Description:      "Find two numbers",
		Difficulty:       challenges.DifficultyBeginner,
		EstimatedMinutes: 15,
		StarterCode:      "package main",
		Status:           challenges.StatusPublished,
		TestCases: []challenges.CreateTestCaseRequest{
			{
				Title:          "Sample 1",
				Input:          "sample in",
				ExpectedOutput: "[0, 1]",
				Hidden:         false,
				SortOrder:      1,
			},
			{
				Title:          "Secret 1",
				Input:          "secret in",
				ExpectedOutput: "[2, 4]",
				Hidden:         true,
				SortOrder:      2,
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create challenge: %v", err)
	}
	_ = created

	// 1. Run Sample Tests (Passing code)
	validCode := `package main

func TwoSum(nums []int, target int) []int {
	seen := make(map[int]int, len(nums))
	for i, num := range nums {
		diff := target - num
		if j, ok := seen[diff]; ok {
			return []int{j, i}
		}
		seen[num] = i
	}
	return nil
}
`
	runRes, err := svc.RunSampleTests(ctx, slug, validCode)
	if err != nil {
		t.Fatalf("failed to run sample tests: %v", err)
	}
	if runRes.Status != runner.StatusPassed {
		t.Fatalf("expected passed status for sample tests, got: %s", runRes.Status)
	}

	// 2. Submit solution
	subDetail, err := svc.SubmitChallenge(ctx, userID, slug, validCode)
	if err != nil {
		t.Fatalf("failed to submit challenge: %v", err)
	}
	if subDetail.Status != runner.StatusPassed {
		t.Fatalf("expected submission to pass all tests, got: %s", subDetail.Status)
	}
	if subDetail.TotalTestCases < 1 {
		t.Errorf("expected at least 1 test case evaluated, got %d", subDetail.TotalTestCases)
	}

	// Verify hidden test sanitization
	for _, o := range subDetail.TestOutcomes {
		if o.Hidden {
			if o.Input != "[Hidden test case]" {
				t.Errorf("expected hidden test input to be sanitized, got %q", o.Input)
			}
			if o.ExpectedOutput != "[Hidden test case]" {
				t.Errorf("expected hidden test expected output to be sanitized, got %q", o.ExpectedOutput)
			}
		}
	}

	// 3. List submissions
	subs, err := svc.ListSubmissions(ctx, userID, slug)
	if err != nil {
		t.Fatalf("failed to list submissions: %v", err)
	}
	if len(subs) != 1 {
		t.Fatalf("expected 1 submission in history, got %d", len(subs))
	}
	if subs[0].ID != subDetail.ID {
		t.Errorf("expected submission ID %s, got %s", subDetail.ID, subs[0].ID)
	}

	// 4. Get submission detail
	fetched, err := svc.GetSubmission(ctx, userID, subDetail.ID)
	if err != nil {
		t.Fatalf("failed to get submission detail: %v", err)
	}
	if fetched.ID != subDetail.ID {
		t.Errorf("expected matching submission ID")
	}

	// 5. Unauthorized access check (another user cannot access this submission)
	otherUser := uuid.New()
	_, err = svc.GetSubmission(ctx, otherUser, subDetail.ID)
	if !errors.Is(err, challenges.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized for other user, got: %v", err)
	}
}
