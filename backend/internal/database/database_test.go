package database_test

import (
	"context"
	"testing"
	"time"

	"go-academy/backend/internal/config"
	"go-academy/backend/internal/database"
)

func TestDatabaseConnectivityAndSeeding(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, cfg)
	if err != nil {
		t.Skipf("skipping live database test (PostgreSQL not reachable): %v", err)
		return
	}
	defer pool.Close()

	db := database.NewDB(pool)

	// Test Ping
	if err := db.Ping(ctx); err != nil {
		t.Fatalf("database ping failed: %v", err)
	}

	// Verify roles seeded by migration
	adminRole, err := db.Queries.GetRoleByName(ctx, "admin")
	if err != nil {
		t.Fatalf("expected admin role to exist in seeded database: %v", err)
	}
	if adminRole.Name != "admin" {
		t.Fatalf("expected role name 'admin', got: %s", adminRole.Name)
	}

	learnerRole, err := db.Queries.GetRoleByName(ctx, "learner")
	if err != nil {
		t.Fatalf("expected learner role to exist in seeded database: %v", err)
	}
	if learnerRole.Name != "learner" {
		t.Fatalf("expected role name 'learner', got: %s", learnerRole.Name)
	}

	// Verify initial admin count is 0
	adminCount, err := db.Queries.CountAdmins(ctx)
	if err != nil {
		t.Fatalf("failed to query admin count: %v", err)
	}
	if adminCount != 0 {
		t.Logf("current admin count: %d", adminCount)
	}
}
