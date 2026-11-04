package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"go-academy/backend/internal/platform/response"
)

// Recoverer returns a middleware that recovers from panics and logs the stack trace.
func Recoverer(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					stack := string(debug.Stack())
					logger.Error("panic recovered",
						slog.Any("error", rec),
						slog.String("path", r.URL.Path),
						slog.String("stack", stack),
					)

					response.Error(
						w,
						http.StatusInternalServerError,
						"INTERNAL_SERVER_ERROR",
						"An unexpected error occurred. Please try again later.",
						nil,
					)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
