package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const MaxOutputBytes = 64 * 1024 // 64 KB bounded output

// DockerRunner runs code inside an isolated non-root Docker container.
type DockerRunner struct {
	image   string
	timeout time.Duration
}

// NewDockerRunner creates a new DockerRunner with the specified image.
func NewDockerRunner(image string, timeout time.Duration) *DockerRunner {
	if image == "" {
		image = "golang:alpine"
	}
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &DockerRunner{
		image:   image,
		timeout: timeout,
	}
}

func (d *DockerRunner) Execute(ctx context.Context, req ExecutionRequest) (*ExecutionResult, error) {
	// Create isolated workspace
	tmpDir, err := os.MkdirTemp("", "goacademy-run-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create sandbox workspace: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	_ = os.Chmod(tmpDir, 0777)

	// Write files
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(GenerateGoMod()), 0666); err != nil {
		return nil, fmt.Errorf("failed to write go.mod: %w", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "solution.go"), []byte(req.Code), 0666); err != nil {
		return nil, fmt.Errorf("failed to write solution.go: %w", err)
	}
	harness := GenerateTestHarness(req)
	if err := os.WriteFile(filepath.Join(tmpDir, "solution_test.go"), []byte(harness), 0666); err != nil {
		return nil, fmt.Errorf("failed to write solution_test.go: %w", err)
	}

	timeout := req.Timeout
	if timeout <= 0 || timeout > d.timeout {
		timeout = d.timeout
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "docker", "run", "--rm",
		"--network", "none",
		"--memory", "256m",
		"--cpus", "1.0",
		"--pids-limit", "64",
		"-v", fmt.Sprintf("%s:/workspace", tmpDir),
		"-w", "/workspace",
		"-e", "GOCACHE=/tmp/gocache",
		"-e", "CGO_ENABLED=0",
		d.image,
		"go", "test", "-v", "-timeout", "5s", ".",
	)

	var buf bytes.Buffer
	limitedWriter := io.LimitReader(&buf, MaxOutputBytes)
	_ = limitedWriter

	startTime := time.Now()
	combinedOutput, cmdErr := cmd.CombinedOutput()
	durationMs := time.Since(startTime).Milliseconds()

	// Truncate to MaxOutputBytes if needed
	if len(combinedOutput) > MaxOutputBytes {
		combinedOutput = append(combinedOutput[:MaxOutputBytes], []byte("\n... [output truncated at 64KB]")...)
	}
	outputStr := string(combinedOutput)

	if errors.Is(execCtx.Err(), context.DeadlineExceeded) {
		return &ExecutionResult{
			Status:          StatusTimedOut,
			TotalTests:      len(req.TestCases),
			PassedTests:     0,
			ExecutionTimeMs: durationMs,
			CompilerOutput:  "Execution timed out (exceeded 5s wall-clock limit)",
			TestOutcomes:    nil,
		}, nil
	}

	exitCode := 0
	if cmdErr != nil {
		var exitErr *exec.ExitError
		if errors.As(cmdErr, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	result := ParseExecutionOutput(outputStr, req, durationMs, exitCode)
	return result, nil
}
