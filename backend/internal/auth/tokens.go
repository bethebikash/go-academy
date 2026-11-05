package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"go-academy/backend/internal/users"
)

var (
	// ErrInvalidToken indicates the JWT signature or claims are invalid.
	ErrInvalidToken = errors.New("invalid or expired authentication token")
	// ErrExpiredToken indicates the token has expired.
	ErrExpiredToken = errors.New("token has expired")
)

// Claims represents the JWT claims for Go Academy access tokens.
type Claims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Status   string `json:"status"`
	jwt.RegisteredClaims
}

// GenerateAccessToken generates a short-lived signed JWT access token.
func GenerateAccessToken(user *users.User, secret string, expiry time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		UserID:   user.ID.String(),
		Email:    user.Email,
		Username: user.Username,
		Role:     user.Role,
		Status:   user.Status,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}

	return tokenString, nil
}

// ValidateAccessToken parses and validates a JWT access token string.
func ValidateAccessToken(tokenStr, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// GenerateOpaqueToken generates 32 cryptographically secure random bytes
// and returns both the URL-safe base64 string and its SHA-256 hex digest.
func GenerateOpaqueToken() (raw string, hash string, err error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	raw = base64.RawURLEncoding.EncodeToString(bytes)
	hash = HashOpaqueToken(raw)
	return raw, hash, nil
}

// HashOpaqueToken calculates the SHA-256 hex string of a raw opaque token.
func HashOpaqueToken(raw string) string {
	digest := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(digest[:])
}

// Context key for claims
type contextKey string

const (
	claimsContextKey contextKey = "auth_claims"
)

// WithClaims injects Claims into a context.
func WithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, claimsContextKey, claims)
}

// GetClaims retrieves Claims from a context.
func GetClaims(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(*Claims)
	return claims, ok
}

