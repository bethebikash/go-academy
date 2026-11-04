package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all platform runtime configuration settings.
type Config struct {
	AppEnv             string
	Port               string
	AppName            string
	LogLevel           string
	FrontendURL        string
	DatabaseURL        string
	DBMaxOpenConns     int
	DBMaxIdleConns     int
	DBMaxIdleTime      time.Duration
	JWTSecret          string
	JWTAccessExpiry    time.Duration
	RefreshTokenExpiry time.Duration
	CookieSecure       bool
	CookieDomain       string
	RateLimitLogin     int
	RateLimitRegister  int
}

// Load reads and validates configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		AppEnv:             getEnv("APP_ENV", "development"),
		Port:               getEnv("APP_PORT", "8080"),
		AppName:            getEnv("APP_NAME", "Go Academy"),
		LogLevel:           getEnv("LOG_LEVEL", "info"),
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:5173"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://goacademy:goacademy_password@localhost:5432/goacademy_dev?sslmode=disable"),
		DBMaxOpenConns:     getEnvInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:     getEnvInt("DB_MAX_IDLE_CONNS", 5),
		DBMaxIdleTime:      getEnvDuration("DB_MAX_IDLE_TIME", 15*time.Minute),
		JWTSecret:          getEnv("JWT_SECRET", "development_jwt_secret_change_in_production_min_32_bytes!"),
		JWTAccessExpiry:    getEnvDuration("JWT_ACCESS_EXPIRY", 15*time.Minute),
		RefreshTokenExpiry: getEnvDuration("REFRESH_TOKEN_EXPIRY", 7*24*time.Hour),
		CookieSecure:       getEnvBool("COOKIE_SECURE", false),
		CookieDomain:       getEnv("COOKIE_DOMAIN", "localhost"),
		RateLimitLogin:     getEnvInt("RATE_LIMIT_LOGIN", 5),
		RateLimitRegister:  getEnvInt("RATE_LIMIT_REGISTER", 5),
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// Validate checks critical configuration requirements.
func (c *Config) Validate() error {
	if c.AppEnv == "production" {
		if len(c.JWTSecret) < 32 {
			return fmt.Errorf("JWT_SECRET must be at least 32 characters long in production")
		}
		if !c.CookieSecure {
			return fmt.Errorf("COOKIE_SECURE must be true in production")
		}
	}
	return nil
}

// IsProduction returns true if current environment is production.
func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			return boolVal
		}
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		if dur, err := time.ParseDuration(val); err == nil {
			return dur
		}
	}
	return defaultVal
}
