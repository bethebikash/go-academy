package config_test

import (
	"os"
	"testing"

	"go-academy/backend/internal/config"
)

func TestConfigLoad(t *testing.T) {
	os.Setenv("APP_PORT", "9090")
	defer os.Unsetenv("APP_PORT")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected config to load successfully, got: %v", err)
	}

	if cfg.Port != "9090" {
		t.Fatalf("expected port 9090, got: %s", cfg.Port)
	}
}

func TestConfigValidation_Production(t *testing.T) {
	cfg := &config.Config{
		AppEnv:       "production",
		JWTSecret:    "short",
		CookieSecure: false,
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error in production for short JWTSecret and insecure cookies, got nil")
	}
}
