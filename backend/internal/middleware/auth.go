package middleware

import (
	"net/http"
	"strings"

	"go-academy/backend/internal/auth"
	"go-academy/backend/internal/platform/response"
)

// Authenticate verifies the Bearer JWT and injects claims into context if present.
func Authenticate(jwtSecret string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				next.ServeHTTP(w, r)
				return
			}

			tokenStr := parts[1]
			claims, err := auth.ValidateAccessToken(tokenStr, jwtSecret)
			if err != nil {
				// Don't fail immediately: RequireAuth middleware will reject if endpoint requires auth
				next.ServeHTTP(w, r)
				return
			}

			ctx := auth.WithClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuth ensures that a valid authenticated user is present in context.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := auth.GetClaims(r.Context())
		if !ok || claims == nil {
			response.Error(
				w,
				http.StatusUnauthorized,
				"UNAUTHENTICATED",
				"Authentication is required to access this resource.",
				nil,
			)
			return
		}

		if claims.Status != "active" {
			response.Error(
				w,
				http.StatusForbidden,
				"ACCOUNT_DISABLED",
				"Your account has been deactivated.",
				nil,
			)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireRole enforces role-based access control (e.g. "admin").
func RequireRole(role string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := auth.GetClaims(r.Context())
			if !ok || claims == nil {
				response.Error(
					w,
					http.StatusUnauthorized,
					"UNAUTHENTICATED",
					"Authentication is required to access this resource.",
					nil,
				)
				return
			}

			if claims.Role != role {
				response.Error(
					w,
					http.StatusForbidden,
					"FORBIDDEN",
					"You do not possess the required permissions to perform this action.",
					nil,
				)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
