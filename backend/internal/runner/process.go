package runner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// ProcessRunner runs code inside an isolated workspace directory using host Go toolchain.
type ProcessRunner struct {
	timeout time.Duration
}

// NewProcessRunner constructs a new ProcessRunner.
func NewProcessRunner(timeout time.Duration) *ProcessRunner {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &ProcessRunner{timeout: timeout}
}

func (p *ProcessRunner) Execute(ctx context.Context, req ExecutionRequest) (*ExecutionResult, error) {
	tmpDir, err := os.MkdirTemp("", "goacademy-proc-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create sandbox workspace: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	_ = os.Chmod(tmpDir, 0700)

	// Write files
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(GenerateGoMod()), 0600); err != nil {
		return nil, fmt.Errorf("failed to write go.mod: %w", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "solution.go"), []byte(req.Code), 0600); err != nil {
		return nil, fmt.Errorf("failed to write solution.go: %w", err)
	}
	harness := GenerateTestHarness(req)
	if err := os.WriteFile(filepath.Join(tmpDir, "solution_test.go"), []byte(harness), 0600); err != nil {
		return nil, fmt.Errorf("failed to write solution_test.go: %w", err)
	}

	timeout := req.Timeout
	if timeout <= 0 || timeout > p.timeout {
		timeout = p.timeout
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "go", "test", "-v", "-timeout", "5s", ".")
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(),
		"GOCACHE=/tmp/gocache",
		"CGO_ENABLED=0",
	)

	startTime := time.Now()
	combinedOutput, cmdErr := cmd.CombinedOutput()
	durationMs := time.Since(startTime).Milliseconds()

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
