package health_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-academy/backend/internal/health"
)

type mockDB struct {
	err error
}

func (m *mockDB) Ping(ctx context.Context) error {
	return m.err
}

func TestHealthLive(t *testing.T) {
	handler := health.NewHandler(nil)
	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rr := httptest.NewRecorder()

	handler.Live(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var res map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	data, ok := res["data"].(map[string]any)
	if !ok || data["status"] != "ok" {
		t.Fatalf("expected data.status to be 'ok', got %v", res)
	}
}

func TestHealthReady(t *testing.T) {
	t.Run("ready when database is healthy", func(t *testing.T) {
		handler := health.NewHandler(&mockDB{err: nil})
		req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
		rr := httptest.NewRecorder()

		handler.Ready(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rr.Code)
		}
	})

	t.Run("unready when database fails", func(t *testing.T) {
		handler := health.NewHandler(&mockDB{err: errors.New("connection refused")})
		req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
		rr := httptest.NewRecorder()

		handler.Ready(rr, req)

		if rr.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected status 503, got %d", rr.Code)
		}
	})
}
