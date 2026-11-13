package invitations_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"go-academy/backend/internal/auth"
	"go-academy/backend/internal/invitations"
	"go-academy/backend/internal/middleware"
	"go-academy/backend/internal/users"
)

func setupInvitationServer(t *testing.T) (*chi.Mux, *invitations.Handler, *users.User, string, string) {
	db, cfg, _, invitationService, adminUser := setupInvitationTest(t)

	invitationHandler := invitations.NewHandler(invitationService, cfg)

	r := chi.NewRouter()
	r.Use(middleware.Authenticate(cfg.JWTSecret))

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", invitationHandler.Register)
			r.Get("/invitations/validate", invitationHandler.Validate)
		})

		r.Route("/admin", func(r chi.Router) {
			r.Use(middleware.RequireAuth)
			r.Use(middleware.RequireRole(users.RoleAdmin))

			r.Route("/invitations", func(r chi.Router) {
				r.Post("/", invitationHandler.Create)
				r.Get("/", invitationHandler.List)
				r.Delete("/{id}", invitationHandler.Revoke)
			})
		})
	})

	t.Cleanup(func() {
		db.Close()
	})

	adminToken, err := auth.GenerateAccessToken(adminUser, cfg.JWTSecret, 15*time.Minute)
	if err != nil {
		t.Fatalf("failed to generate admin token: %v", err)
	}

	return r, invitationHandler, adminUser, adminToken, cfg.JWTSecret
}

func TestInvitationHTTPHandlers(t *testing.T) {
	router, _, _, adminToken, jwtSecret := setupInvitationServer(t)

	now := time.Now().UnixNano()
	testEmail := fmt.Sprintf("http_invitee_%d@example.com", now)
	testUsername := fmt.Sprintf("http_user_%d", now%1000000)

	// 1. Admin creates invitation
	createPayload, _ := json.Marshal(map[string]any{
		"email":           testEmail,
		"expires_in_days": 7,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/invitations", bytes.NewReader(createPayload))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var createResp struct {
		Data struct {
			ID              string `json:"id"`
			Email           string `json:"email"`
			InvitationToken string `json:"invitation_token"`
			InvitationURL   string `json:"invitation_url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	token := createResp.Data.InvitationToken
	if token == "" {
		t.Fatal("expected non-empty invitation_token")
	}

	// 2. Public validate endpoint
	validateReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/invitations/validate?token="+token, nil)
	validateW := httptest.NewRecorder()
	router.ServeHTTP(validateW, validateReq)

	if validateW.Code != http.StatusOK {
		t.Fatalf("expected status 200 on validation, got %d: %s", validateW.Code, validateW.Body.String())
	}

	// 3. User registers using the invitation
	regPayload, _ := json.Marshal(map[string]string{
		"invitation_token": token,
		"username":         testUsername,
		"password":         "Password123!Secure",
	})
	regReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(regPayload))
	regReq.Header.Set("Content-Type", "application/json")
	regW := httptest.NewRecorder()
	router.ServeHTTP(regW, regReq)

	if regW.Code != http.StatusCreated {
		t.Fatalf("expected status 201 for register, got %d: %s", regW.Code, regW.Body.String())
	}

	// Verify refresh_token cookie was set
	cookies := regW.Result().Cookies()
	var hasRefreshCookie bool
	for _, c := range cookies {
		if c.Name == "refresh_token" && c.Value != "" {
			hasRefreshCookie = true
			break
		}
	}
	if !hasRefreshCookie {
		t.Fatal("expected refresh_token cookie to be set upon registration")
	}

	// 4. Repeated registration with same token should return 409 Conflict
	repeatReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(regPayload))
	repeatReq.Header.Set("Content-Type", "application/json")
	repeatW := httptest.NewRecorder()
	router.ServeHTTP(repeatW, repeatReq)

	if repeatW.Code != http.StatusConflict {
		t.Fatalf("expected status 409 for reused token, got %d: %s", repeatW.Code, repeatW.Body.String())
	}

	// 5. Admin lists invitations
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/invitations?page=1&limit=10", nil)
	listReq.Header.Set("Authorization", "Bearer "+adminToken)
	listW := httptest.NewRecorder()
	router.ServeHTTP(listW, listReq)

	if listW.Code != http.StatusOK {
		t.Fatalf("expected status 200 for list, got %d: %s", listW.Code, listW.Body.String())
	}

	// 6. Non-admin blocked from creating invitations
	learnerToken, _ := auth.GenerateAccessToken(&users.User{
		Role:   users.RoleLearner,
		Status: users.StatusActive,
	}, jwtSecret, 15*time.Minute)

	unauthReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/invitations", bytes.NewReader(createPayload))
	unauthReq.Header.Set("Authorization", "Bearer "+learnerToken)
	unauthReq.Header.Set("Content-Type", "application/json")
	unauthW := httptest.NewRecorder()
	router.ServeHTTP(unauthW, unauthReq)

	if unauthW.Code != http.StatusForbidden {
		t.Fatalf("expected status 403 Forbidden for learner, got %d", unauthW.Code)
	}
}
