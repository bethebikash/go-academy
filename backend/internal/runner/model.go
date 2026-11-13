package runner

import (
	"time"

	"github.com/google/uuid"
)

// Execution Status constants
const (
	StatusPassed       = "passed"
	StatusFailed       = "failed"
	StatusCompileError = "compile_error"
	StatusTimedOut     = "timed_out"
	StatusError        = "error"
)

// TestCaseInput represents an evaluation test case passed into the runner
type TestCaseInput struct {
	ID             uuid.UUID `json:"id"`
	Title          string    `json:"title"`
	Input          string    `json:"input"`
	ExpectedOutput string    `json:"expected_output"`
	Hidden         bool      `json:"hidden"`
}

// ExecutionRequest represents a request to run untrusted code against test cases
type ExecutionRequest struct {
	ChallengeSlug string          `json:"challenge_slug"`
	Code          string          `json:"code"`
	TestCases     []TestCaseInput `json:"test_cases"`
	Timeout       time.Duration   `json:"timeout"`
}

// TestCaseOutcome represents the outcome of executing a single test case
type TestCaseOutcome struct {
	TestCaseID      uuid.UUID `json:"test_case_id"`
	Title           string    `json:"title"`
	Passed          bool      `json:"passed"`
	Hidden          bool      `json:"hidden"`
	Input           string    `json:"input"`
	ExpectedOutput  string    `json:"expected_output"`
	ActualOutput    string    `json:"actual_output"`
	ExecutionTimeMs int64     `json:"execution_time_ms"`
	ErrorMessage    string    `json:"error_message,omitempty"`
}

// ExecutionResult encapsulates the aggregate test execution results
type ExecutionResult struct {
	Status          string            `json:"status"` // "passed", "failed", "compile_error", "timed_out", "error"
	TotalTests      int               `json:"total_tests"`
	PassedTests     int               `json:"passed_tests"`
	ExecutionTimeMs int64             `json:"execution_time_ms"`
	CompilerOutput  string            `json:"compiler_output,omitempty"`
	TestOutcomes    []TestCaseOutcome `json:"test_outcomes"`
}
