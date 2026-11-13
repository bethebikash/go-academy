package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"golang.org/x/time/rate"

	"go-academy/backend/internal/auth"
	"go-academy/backend/internal/config"
	"go-academy/backend/internal/database"
	"go-academy/backend/internal/health"
	"go-academy/backend/internal/invitations"
	appmiddleware "go-academy/backend/internal/middleware"
	"go-academy/backend/internal/platform/response"
	"go-academy/backend/internal/users"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Configure structured logger
	var logHandler slog.Handler
	logLevel := slog.LevelInfo
	if cfg.LogLevel == "debug" {
		logLevel = slog.LevelDebug
	}

	if cfg.IsProduction() {
		logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	} else {
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	}
	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	logger.Info("starting application",
		slog.String("app_name", cfg.AppName),
		slog.String("environment", cfg.AppEnv),
		slog.String("port", cfg.Port),
	)

	// Apply database migrations
	if err := database.RunMigrations(cfg.DatabaseURL, "migrations", logger); err != nil {
		logger.Warn("migration runner notice", slog.String("detail", err.Error()))
	}

	// Initialize database connection pool
	ctxInit, cancelInit := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelInit()

	pool, err := database.NewPool(ctxInit, cfg)
	if err != nil {
		logger.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	db := database.NewDB(pool)
	defer db.Close()

	logger.Info("database connection pool established",
		slog.Int("max_conns", cfg.DBMaxOpenConns),
		slog.Int("min_conns", cfg.DBMaxIdleConns),
	)

	// Domain Services
	userService := users.NewService(db)
	authService := auth.NewService(db, cfg, userService)
	invitationService := invitations.NewService(db, cfg, userService)

	// HTTP Handlers
	healthHandler := health.NewHandler(db)
	authHandler := auth.NewHandler(authService, userService, cfg)
	invitationHandler := invitations.NewHandler(invitationService, cfg)
	userHandler := users.NewHandler(userService)

	// Rate Limiters
	loginLimiter := appmiddleware.NewIPRateLimiter(rate.Every(time.Minute/time.Duration(cfg.RateLimitLogin)), cfg.RateLimitLogin)
	registerLimiter := appmiddleware.NewIPRateLimiter(rate.Every(time.Minute/time.Duration(cfg.RateLimitRegister)), cfg.RateLimitRegister)

	// Router setup
	r := chi.NewRouter()

	// Base middlewares
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.RequestID)
	r.Use(appmiddleware.RequestLogger(logger))
	r.Use(appmiddleware.Recoverer(logger))

	// CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.FrontendURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Request-ID"},
		ExposedHeaders:   []string{"Link", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Enforce 1MB max request body limit
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB
			next.ServeHTTP(w, r)
		})
	})

	// Global JWT parsing middleware
	r.Use(appmiddleware.Authenticate(cfg.JWTSecret))

	// Health routes with live database ping check
	r.Get("/health/live", healthHandler.Live)
	r.Get("/health/ready", healthHandler.Ready)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			response.JSON(w, http.StatusOK, map[string]string{
				"message": "Go Academy API v1 operational",
			}, nil)
		})

		// Authentication routes
		r.Route("/auth", func(r chi.Router) {
			r.With(loginLimiter.Limit).Post("/login", authHandler.Login)
			r.With(registerLimiter.Limit).Post("/register", invitationHandler.Register)
			r.Get("/invitations/validate", invitationHandler.Validate)
			r.Post("/refresh", authHandler.Refresh)
			r.Post("/logout", authHandler.Logout)
			r.With(appmiddleware.RequireAuth).Get("/me", authHandler.Me)
		})

		// Admin routes (Protected by authentication + admin role)
		r.Route("/admin", func(r chi.Router) {
			r.Use(appmiddleware.RequireAuth)
			r.Use(appmiddleware.RequireRole(users.RoleAdmin))

			r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
				response.JSON(w, http.StatusOK, map[string]string{
					"message": "Admin authorization verified",
				}, nil)
			})

			// Admin invitation management
			r.Route("/invitations", func(r chi.Router) {
				r.Post("/", invitationHandler.Create)
				r.Get("/", invitationHandler.List)
				r.Delete("/{id}", invitationHandler.Revoke)
			})

			// Admin user management
			r.Route("/users", func(r chi.Router) {
				r.Get("/", userHandler.List)
				r.Patch("/{id}/status", userHandler.UpdateStatus)
			})
		})
	})

	// Server configuration
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Server shutdown channel
	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		logger.Info("shutting down server", slog.String("signal", s.String()))

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		shutdownError <- server.Shutdown(ctx)
	}()

	logger.Info("server listening", slog.String("addr", server.Addr))
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server failed to start", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if err := <-shutdownError; err != nil {
		logger.Error("graceful shutdown failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("server stopped cleanly")
}
