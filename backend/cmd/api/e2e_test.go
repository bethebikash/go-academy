package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"go-academy/backend/internal/analytics"
	"go-academy/backend/internal/auth"
	"go-academy/backend/internal/bookmarks"
	"go-academy/backend/internal/challenges"
	"go-academy/backend/internal/config"
	"go-academy/backend/internal/content"
	"go-academy/backend/internal/database"
	"go-academy/backend/internal/invitations"
	"go-academy/backend/internal/middleware"
	"go-academy/backend/internal/notes"
	"go-academy/backend/internal/progress"
	"go-academy/backend/internal/runner"
	"go-academy/backend/internal/users"
)

func TestEndToEndPlatformFlow(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, cfg)
	if err != nil {
		t.Skipf("skipping live database test: %v", err)
		return
	}
	defer pool.Close()

	db := database.NewDB(pool)
	userService := users.NewService(db)
	authService := auth.NewService(db, cfg, userService)
	invitationService := invitations.NewService(db, cfg, userService)
	contentService := content.NewService(db)
	progressService := progress.NewService(db)
	noteService := notes.NewService(db)
	bookmarkService := bookmarks.NewService(db)
	codeRunner := runner.NewProcessRunner(5 * time.Second)
	challengeService := challenges.NewService(db, codeRunner)
	analyticsService := analytics.NewService(db)

	progressService.SetActivityRecorder(analyticsService)
	challengeService.SetActivityRecorder(analyticsService)
	noteService.SetActivityRecorder(analyticsService)

	authHandler := auth.NewHandler(authService, userService, cfg)
	invitationHandler := invitations.NewHandler(invitationService, cfg)
	userHandler := users.NewHandler(userService)
	contentHandler := content.NewHandler(contentService)
	progressHandler := progress.NewHandler(progressService)
	noteHandler := notes.NewHandler(noteService)
	bookmarkHandler := bookmarks.NewHandler(bookmarkService)
	challengeHandler := challenges.NewHandler(challengeService)
	analyticsHandler := analytics.NewHandler(analyticsService)

	// Build full API router
	r := chi.NewRouter()
	r.Use(middleware.Authenticate(cfg.JWTSecret))

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/certificates/verify/{code}", analyticsHandler.VerifyCertificate)

		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", authHandler.Login)
			r.Post("/register", invitationHandler.Register)
			r.Get("/invitations/validate", invitationHandler.Validate)
			r.Post("/refresh", authHandler.Refresh)
			r.Post("/logout", authHandler.Logout)
			r.With(middleware.RequireAuth).Get("/me", authHandler.Me)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth)
			r.Get("/learning-paths", contentHandler.ListPaths)
			r.Get("/learning-paths/{slug}", contentHandler.GetPath)
			r.Get("/courses/{slug}", contentHandler.GetCourse)
			r.Get("/lessons/{slug}", contentHandler.GetLesson)
			r.Get("/lessons/{id}/progress", progressHandler.GetProgress)
			r.Put("/lessons/{id}/progress", progressHandler.UpdateProgress)

			r.Route("/notes", func(r chi.Router) {
				r.Get("/", noteHandler.ListNotes)
				r.Post("/", noteHandler.CreateNote)
				r.Patch("/{id}", noteHandler.UpdateNote)
				r.Delete("/{id}", noteHandler.DeleteNote)
			})

			r.Route("/bookmarks", func(r chi.Router) {
				r.Get("/", bookmarkHandler.ListBookmarks)
				r.Post("/", bookmarkHandler.CreateBookmark)
				r.Get("/check", bookmarkHandler.CheckLesson)
				r.Delete("/lessons/{lessonId}", bookmarkHandler.DeleteLessonBookmark)
			})

			r.Get("/challenges", challengeHandler.ListChallenges)
			r.Get("/challenges/{slug}", challengeHandler.GetChallenge)
			r.Post("/challenges/{slug}/run", challengeHandler.RunSampleTests)
			r.Post("/challenges/{slug}/submit", challengeHandler.SubmitChallenge)
			r.Get("/challenges/{slug}/submissions", challengeHandler.ListSubmissions)
			r.Get("/challenges/{slug}/submissions/{id}", challengeHandler.GetSubmission)

			// Analytics & Certificates
			r.Get("/analytics/summary", analyticsHandler.GetSummary)
			r.Get("/analytics/activity", analyticsHandler.GetActivity)
			r.Get("/analytics/badges", analyticsHandler.GetBadges)
			r.Post("/certificates/claim", analyticsHandler.ClaimCertificate)
			r.Get("/certificates/me", analyticsHandler.ListCertificates)
		})

		r.Route("/admin", func(r chi.Router) {
			r.Use(middleware.RequireAuth)
			r.Use(middleware.RequireRole(users.RoleAdmin))

			r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			r.Route("/invitations", func(r chi.Router) {
				r.Post("/", invitationHandler.Create)
				r.Get("/", invitationHandler.List)
				r.Delete("/{id}", invitationHandler.Revoke)
			})

			r.Route("/users", func(r chi.Router) {
				r.Get("/", userHandler.List)
				r.Patch("/{id}/status", userHandler.UpdateStatus)
			})

			r.Route("/challenges", func(r chi.Router) {
				r.Get("/", challengeHandler.AdminListChallenges)
				r.Post("/", challengeHandler.AdminCreateChallenge)
				r.Get("/{id}", challengeHandler.AdminGetChallenge)
				r.Put("/{id}", challengeHandler.AdminUpdateChallenge)
				r.Patch("/{id}/status", challengeHandler.AdminUpdateChallengeStatus)
				r.Delete("/{id}", challengeHandler.AdminDeleteChallenge)
			})
		})
	})

	server := httptest.NewServer(r)
	defer server.Close()

	client := server.Client()

	now := time.Now().UnixNano()
	adminEmail := "test_admin@example.com"
	adminPass := "SuperAdminPass123!"

	// Step 1: Admin logs in
	loginPayload, _ := json.Marshal(map[string]string{
		"email":    adminEmail,
		"password": adminPass,
	})
	resp, err := client.Post(server.URL+"/api/v1/auth/login", "application/json", bytes.NewReader(loginPayload))
	if err != nil {
		t.Fatalf("admin login failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for admin login, got %d", resp.StatusCode)
	}

	var adminLoginResp struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&adminLoginResp)
	adminToken := adminLoginResp.Data.AccessToken
	if adminToken == "" {
		t.Fatal("expected non-empty admin token")
	}

	// Step 2: Admin creates an invitation for a friend
	friendEmail := fmt.Sprintf("e2e_friend_%d@example.com", now)
	invitePayload, _ := json.Marshal(map[string]any{
		"email":           friendEmail,
		"expires_in_days": 7,
	})
	req, _ := http.NewRequest(http.MethodPost, server.URL+"/api/v1/admin/invitations", bytes.NewReader(invitePayload))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("create invitation failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 for create invite, got %d", resp.StatusCode)
	}

	var inviteResp struct {
		Data struct {
			ID              string `json:"id"`
			InvitationToken string `json:"invitation_token"`
		} `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&inviteResp)
	rawInviteToken := inviteResp.Data.InvitationToken

	// Step 3: Invitee validates token preview
	resp, err = client.Get(server.URL + "/api/v1/auth/invitations/validate?token=" + rawInviteToken)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for token validation, got status %d", resp.StatusCode)
	}

	// Step 4: Invitee registers using the token
	friendUsername := fmt.Sprintf("friend_%d", now%1000000)
	friendPass := "FriendPass123!Secure"
	regPayload, _ := json.Marshal(map[string]string{
		"invitation_token": rawInviteToken,
		"username":         friendUsername,
		"password":         friendPass,
	})
	resp, err = client.Post(server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(regPayload))
	if err != nil || resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 for register, got %d", resp.StatusCode)
	}

	var regResp struct {
		Data struct {
			AccessToken string `json:"access_token"`
			User        struct {
				ID   string `json:"id"`
				Role string `json:"role"`
			} `json:"user"`
		} `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&regResp)
	learnerToken := regResp.Data.AccessToken
	learnerID := regResp.Data.User.ID

	if regResp.Data.User.Role != users.RoleLearner {
		t.Fatalf("expected learner role, got %s", regResp.Data.User.Role)
	}

	// Step 5: Learner accesses /auth/me (Permitted)
	req, _ = http.NewRequest(http.MethodGet, server.URL+"/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+learnerToken)
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for learner me, got %d", resp.StatusCode)
	}

	// Step 6: Learner attempts to access admin endpoint (Must be 403 Forbidden)
	req, _ = http.NewRequest(http.MethodGet, server.URL+"/api/v1/admin/invitations", nil)
	req.Header.Set("Authorization", "Bearer "+learnerToken)
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for learner accessing admin invitations, got %d", resp.StatusCode)
	}

	// Step 6.1: Learner browses published learning paths
	req, _ = http.NewRequest(http.MethodGet, server.URL+"/api/v1/learning-paths", nil)
	req.Header.Set("Authorization", "Bearer "+learnerToken)
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for learner browsing learning paths, got %d", resp.StatusCode)
	}

	// Step 6.2: Learner accesses lesson detail
	req, _ = http.NewRequest(http.MethodGet, server.URL+"/api/v1/lessons/variables-types-zero-values", nil)
	req.Header.Set("Authorization", "Bearer "+learnerToken)
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for learner reading lesson, got %d", resp.StatusCode)
	}
	var lessonResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&lessonResp)
	if lessonResp.Data.ID == "" {
		t.Fatal("expected non-empty lesson id")
	}

	// Step 6.3: Learner marks lesson as completed
	compPayload, _ := json.Marshal(map[string]string{"status": "completed"})
	req, _ = http.NewRequest(http.MethodPut, server.URL+"/api/v1/lessons/"+lessonResp.Data.ID+"/progress", bytes.NewReader(compPayload))
	req.Header.Set("Authorization", "Bearer "+learnerToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for marking lesson complete, got %d", resp.StatusCode)
	}

	// Step 6.4: Learner creates a personal note
	notePayload, _ := json.Marshal(map[string]any{
		"lesson_id": lessonResp.Data.ID,
		"title":     "Variables Note",
		"content":   "Zero values are applied automatically in Go.",
		"tags":      []string{"basics", "variables"},
	})
	req, _ = http.NewRequest(http.MethodPost, server.URL+"/api/v1/notes", bytes.NewReader(notePayload))
	req.Header.Set("Authorization", "Bearer "+learnerToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 for creating note, got %d", resp.StatusCode)
	}
	var createdNoteResp struct {
		Data struct {
			ID      string `json:"id"`
			Version int    `json:"version"`
		} `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&createdNoteResp)

	// Step 6.5: Learner updates note with valid version (Optimistic locking)
	updateNotePayload, _ := json.Marshal(map[string]any{
		"title":   "Variables Note - Updated",
		"content": "Expanded content.",
		"version": createdNoteResp.Data.Version,
	})
	req, _ = http.NewRequest(http.MethodPatch, server.URL+"/api/v1/notes/"+createdNoteResp.Data.ID, bytes.NewReader(updateNotePayload))
	req.Header.Set("Authorization", "Bearer "+learnerToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for updating note with valid version, got %d", resp.StatusCode)
	}

	// Step 6.6: Learner bookmarks the lesson
	bmPayload, _ := json.Marshal(map[string]string{
		"lesson_id": lessonResp.Data.ID,
	})
	req, _ = http.NewRequest(http.MethodPost, server.URL+"/api/v1/bookmarks", bytes.NewReader(bmPayload))
	req.Header.Set("Authorization", "Bearer "+learnerToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 for bookmarking lesson, got %d", resp.StatusCode)
	}

	// Step 6.7: Check bookmark status
	req, _ = http.NewRequest(http.MethodGet, server.URL+"/api/v1/bookmarks/check?lesson_id="+lessonResp.Data.ID, nil)
	req.Header.Set("Authorization", "Bearer "+learnerToken)
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for checking bookmark, got %d", resp.StatusCode)
	}
	var checkBmResp struct {
		Data struct {
			Bookmarked bool `json:"bookmarked"`
		} `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&checkBmResp)
	if !checkBmResp.Data.Bookmarked {
		t.Fatal("expected lesson to be bookmarked")
	}

	// Step 6.8: Remove bookmark
	req, _ = http.NewRequest(http.MethodDelete, server.URL+"/api/v1/bookmarks/lessons/"+lessonResp.Data.ID, nil)
	req.Header.Set("Authorization", "Bearer "+learnerToken)
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for removing bookmark, got %d", resp.StatusCode)
	}

	// Step 6.9: Admin creates and publishes a challenge; Learner browses and inspects
	challengeSlug := fmt.Sprintf("e2e-challenge-%d", time.Now().UnixNano())
	createChallengePayload, _ := json.Marshal(map[string]interface{}{
		"title":             "E2E Challenge",
		"slug":              challengeSlug,
		"description":       "Solve problem via e2e test",
		"difficulty":        "beginner",
		"estimated_minutes": 15,
		"starter_code":      "package main\n\nfunc Run() string { return \"\" }",
		"solution_code":     "package main\n\nfunc Run() string { return \"pass\" }",
		"status":            "published",
		"test_cases": []map[string]interface{}{
			{
				"title":           "Visible Test",
				"input":           "in1",
				"expected_output": "out1",
				"hidden":          false,
				"sort_order":      1,
			},
			{
				"title":           "Secret Test",
				"input":           "in2",
				"expected_output": "out2",
				"hidden":          true,
				"sort_order":      2,
			},
		},
	})
	req, _ = http.NewRequest(http.MethodPost, server.URL+"/api/v1/admin/challenges", bytes.NewReader(createChallengePayload))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 for creating challenge, got %d", resp.StatusCode)
	}
	var createChallengeResp struct {
		Data struct {
			ID           string `json:"id"`
			SolutionCode string `json:"solution_code"`
			TestCases    []struct {
				ID     string `json:"id"`
				Hidden bool   `json:"hidden"`
			} `json:"test_cases"`
		} `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&createChallengeResp)
	if len(createChallengeResp.Data.TestCases) != 2 {
		t.Fatalf("expected admin to see 2 test cases, got %d", len(createChallengeResp.Data.TestCases))
	}
	if createChallengeResp.Data.SolutionCode == "" {
		t.Fatal("expected admin to see solution code")
	}

	// Learner lists published challenges
	req, _ = http.NewRequest(http.MethodGet, server.URL+"/api/v1/challenges", nil)
	req.Header.Set("Authorization", "Bearer "+learnerToken)
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for listing published challenges, got %d", resp.StatusCode)
	}

	// Learner fetches challenge by slug
	req, _ = http.NewRequest(http.MethodGet, server.URL+"/api/v1/challenges/"+challengeSlug, nil)
	req.Header.Set("Authorization", "Bearer "+learnerToken)
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for getting challenge by slug, got %d", resp.StatusCode)
	}
	var learnerChallengeResp struct {
		Data struct {
			Title           string `json:"title"`
			SampleTestCases []struct {
				Hidden bool `json:"hidden"`
			} `json:"sample_test_cases"`
			SolutionCode *string `json:"solution_code,omitempty"`
		} `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&learnerChallengeResp)
	if len(learnerChallengeResp.Data.SampleTestCases) != 1 {
		t.Fatalf("expected learner to see exactly 1 sample test case, got %d", len(learnerChallengeResp.Data.SampleTestCases))
	}
	if learnerChallengeResp.Data.SampleTestCases[0].Hidden {
		t.Fatal("hidden test case exposed to learner!")
	}
	if learnerChallengeResp.Data.SolutionCode != nil {
		t.Fatal("solution code exposed to learner!")
	}

	// Learner runs sample tests
	runCodePayload, _ := json.Marshal(map[string]string{
		"code": "package main\n\nfunc Run() string { return \"\" }",
	})
	req, _ = http.NewRequest(http.MethodPost, server.URL+"/api/v1/challenges/"+challengeSlug+"/run", bytes.NewReader(runCodePayload))
	req.Header.Set("Authorization", "Bearer "+learnerToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for running sample tests, got %d", resp.StatusCode)
	}

	// Learner submits solution
	submitPayload, _ := json.Marshal(map[string]string{
		"code": "package main\n\nfunc Run() string { return \"pass\" }",
	})
	req, _ = http.NewRequest(http.MethodPost, server.URL+"/api/v1/challenges/"+challengeSlug+"/submit", bytes.NewReader(submitPayload))
	req.Header.Set("Authorization", "Bearer "+learnerToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 for submitting challenge, got %d", resp.StatusCode)
	}

	// Learner lists submissions
	req, _ = http.NewRequest(http.MethodGet, server.URL+"/api/v1/challenges/"+challengeSlug+"/submissions", nil)
	req.Header.Set("Authorization", "Bearer "+learnerToken)
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for listing challenge submissions, got %d", resp.StatusCode)
	}

	// Step 6.10: Learner checks analytics summary and badges
	req, _ = http.NewRequest(http.MethodGet, server.URL+"/api/v1/analytics/summary", nil)
	req.Header.Set("Authorization", "Bearer "+learnerToken)
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for analytics summary, got %d", resp.StatusCode)
	}
	var summaryResp struct {
		Data struct {
			Streak struct {
				CurrentStreak int `json:"current_streak"`
			} `json:"streak"`
			CompletedLessons int `json:"completed_lessons"`
		} `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&summaryResp)
	if summaryResp.Data.Streak.CurrentStreak < 1 {
		t.Fatalf("expected streak >= 1, got %d", summaryResp.Data.Streak.CurrentStreak)
	}

	req, _ = http.NewRequest(http.MethodGet, server.URL+"/api/v1/analytics/badges", nil)
	req.Header.Set("Authorization", "Bearer "+learnerToken)
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for user badges, got %d", resp.StatusCode)
	}

	// Step 7: Admin disables learner account
	disablePayload, _ := json.Marshal(map[string]string{"status": "disabled"})
	req, _ = http.NewRequest(http.MethodPatch, server.URL+"/api/v1/admin/users/"+learnerID+"/status", bytes.NewReader(disablePayload))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for user deactivation, got %d", resp.StatusCode)
	}

	// Step 8: Disabled learner attempts to login (Must fail)
	disabledLoginPayload, _ := json.Marshal(map[string]string{
		"email":    friendEmail,
		"password": friendPass,
	})
	resp, err = client.Post(server.URL+"/api/v1/auth/login", "application/json", bytes.NewReader(disabledLoginPayload))
	if err != nil || resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for disabled user login, got %d", resp.StatusCode)
	}
}
