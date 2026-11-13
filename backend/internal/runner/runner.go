package runner

import (
	"context"
	"os/exec"
	"time"

	"go-academy/backend/internal/config"
)

// Runner represents the interface for sandbox code execution.
type Runner interface {
	Execute(ctx context.Context, req ExecutionRequest) (*ExecutionResult, error)
}

// New constructs a Runner based on configuration and environment capability.
func New(cfg *config.Config) Runner {
	if cfg.RunnerMode == "process" {
		return NewProcessRunner(cfg.RunnerTimeout)
	}

	if cfg.RunnerMode == "docker" {
		return NewDockerRunner(cfg.RunnerImage, cfg.RunnerTimeout)
	}

	// Auto mode: check if docker is accessible
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := exec.CommandContext(ctx, "docker", "version").Run(); err == nil {
		return NewDockerRunner(cfg.RunnerImage, cfg.RunnerTimeout)
	}

	return NewProcessRunner(cfg.RunnerTimeout)
}
