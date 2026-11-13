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

	"go-academy/backend/internal/auth"
	"go-academy/backend/internal/config"
	"go-academy/backend/internal/content"
	"go-academy/backend/internal/database"
	"go-academy/backend/internal/invitations"
	"go-academy/backend/internal/middleware"
	"go-academy/backend/internal/progress"
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

	authHandler := auth.NewHandler(authService, userService, cfg)
	invitationHandler := invitations.NewHandler(invitationService, cfg)
	userHandler := users.NewHandler(userService)
	contentHandler := content.NewHandler(contentService)
	progressHandler := progress.NewHandler(progressService)

	// Build full API router
	r := chi.NewRouter()
	r.Use(middleware.Authenticate(cfg.JWTSecret))

	r.Route("/api/v1", func(r chi.Router) {
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
