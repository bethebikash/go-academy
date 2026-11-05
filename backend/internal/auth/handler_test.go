package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"go-academy/backend/internal/auth"
	"go-academy/backend/internal/middleware"
	"go-academy/backend/internal/platform/response"
	"go-academy/backend/internal/users"
)

func setupTestServer(t *testing.T) (*chi.Mux, *auth.Handler, *auth.Service, *users.Service, string) {
	db, cfg, userService, authService := setupTestContext(t)

	authHandler := auth.NewHandler(authService, userService, cfg)

	r := chi.NewRouter()
	r.Use(middleware.Authenticate(cfg.JWTSecret))

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.Refresh)
			r.Post("/logout", authHandler.Logout)
			r.With(middleware.RequireAuth).Get("/me", authHandler.Me)
		})

		r.Route("/admin", func(r chi.Router) {
			r.Use(middleware.RequireAuth)
			r.Use(middleware.RequireRole(users.RoleAdmin))
			r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
				response.JSON(w, http.StatusOK, map[string]string{"status": "admin_verified"}, nil)
			})
		})
	})

	t.Cleanup(func() {
		db.Close()
	})

	return r, authHandler, authService, userService, cfg.JWTSecret
}

func TestAuthHTTPHandlers(t *testing.T) {
	router, _, _, userService, jwtSecret := setupTestServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	adminUser, err := userService.GetUserByEmail(ctx, "test_admin@example.com")
	if err != nil {
		t.Fatalf("failed to fetch admin user: %v", err)
	}

	password := "SuperAdminPass123!"

	// 1. Test POST /api/v1/auth/login
	loginBody, _ := json.Marshal(map[string]string{
		"email":    adminUser.Email,
		"password": password,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify Set-Cookie header contains refresh_token
	cookies := w.Result().Cookies()
	var refreshCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "refresh_token" {
			refreshCookie = c
			break
		}
	}
	if refreshCookie == nil || refreshCookie.Value == "" {
		t.Fatal("expected refresh_token cookie to be set")
	}
	if !refreshCookie.HttpOnly {
		t.Fatal("expected refresh_token cookie to be HttpOnly")
	}

	var loginResp struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("failed to parse login response: %v", err)
	}
	accessToken := loginResp.Data.AccessToken
	if accessToken == "" {
		t.Fatal("expected non-empty access token")
	}

	// 2. Test GET /api/v1/auth/me (Authenticated)
	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+accessToken)
	meW := httptest.NewRecorder()
	router.ServeHTTP(meW, meReq)

	if meW.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", meW.Code, meW.Body.String())
	}

	// 3. Test GET /api/v1/auth/me (Unauthenticated)
	unauthReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	unauthW := httptest.NewRecorder()
	router.ServeHTTP(unauthW, unauthReq)

	if unauthW.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 for unauthenticated request, got %d", unauthW.Code)
	}

	// 4. Test POST /api/v1/auth/refresh
	refreshReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	refreshReq.AddCookie(refreshCookie)
	refreshW := httptest.NewRecorder()
	router.ServeHTTP(refreshW, refreshReq)

	if refreshW.Code != http.StatusOK {
		t.Fatalf("expected status 200 for refresh, got %d: %s", refreshW.Code, refreshW.Body.String())
	}

	// 5. Test Admin Role Authorization
	adminReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/ping", nil)
	adminReq.Header.Set("Authorization", "Bearer "+accessToken)
	adminW := httptest.NewRecorder()
	router.ServeHTTP(adminW, adminReq)

	if adminW.Code != http.StatusOK {
		t.Fatalf("expected status 200 for admin user, got %d: %s", adminW.Code, adminW.Body.String())
	}

	// 6. Test Non-Admin Forbidden on Admin Endpoint
	learnerUser := &users.User{
		ID:       adminUser.ID,
		Email:    "learner@example.com",
		Username: "learner_one",
		Role:     users.RoleLearner,
		Status:   users.StatusActive,
	}
	learnerToken, _ := auth.GenerateAccessToken(learnerUser, jwtSecret, 15*time.Minute)

	learnerAdminReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/ping", nil)
	learnerAdminReq.Header.Set("Authorization", "Bearer "+learnerToken)
	learnerAdminW := httptest.NewRecorder()
	router.ServeHTTP(learnerAdminW, learnerAdminReq)

	if learnerAdminW.Code != http.StatusForbidden {
		t.Fatalf("expected status 403 Forbidden for learner on admin route, got %d", learnerAdminW.Code)
	}
}
