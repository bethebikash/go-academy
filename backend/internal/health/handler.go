package health

import (
	"context"
	"net/http"

	"go-academy/backend/internal/platform/response"
)

// DBPinger defines the interface required to verify database connectivity.
type DBPinger interface {
	Ping(ctx context.Context) error
}

// Handler provides liveness and readiness HTTP handlers.
type Handler struct {
	db DBPinger
}

// NewHandler creates a new health Handler.
func NewHandler(db DBPinger) *Handler {
	return &Handler{db: db}
}

// Live handles liveness checks (indicates process is running).
func (h *Handler) Live(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	}, nil)
}

// Ready handles readiness checks (indicates process can accept traffic).
func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	if h.db != nil {
		if err := h.db.Ping(r.Context()); err != nil {
			response.Error(w, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "Database ping failed", nil)
			return
		}
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"status":   "ready",
		"database": "up",
	}, nil)
}
