package runner

import (
	"bufio"
	"encoding/json"
	"strings"

	"github.com/google/uuid"
)

type rawTestResult struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Passed bool   `json:"passed"`
	Actual string `json:"actual"`
	Error  string `json:"error,omitempty"`
	TimeMs int64  `json:"time_ms"`
}

// ParseExecutionOutput inspects raw stdout/stderr from the sandbox execution.
func ParseExecutionOutput(output string, req ExecutionRequest, totalDurationMs int64, exitCode int) *ExecutionResult {
	var outcomes []TestCaseOutcome
	var compilerErrors []string
	var totalMs int64

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "[TEST_RESULT]") {
			parts := strings.Split(line, "[TEST_RESULT]")
			if len(parts) >= 2 {
				jsonStr := strings.TrimSpace(parts[1])
				var raw rawTestResult
				if err := json.Unmarshal([]byte(jsonStr), &raw); err == nil {
					tcID, _ := uuid.Parse(raw.ID)
					outcomes = append(outcomes, TestCaseOutcome{
						TestCaseID:      tcID,
						Title:           raw.Title,
						Passed:          raw.Passed,
						ActualOutput:    raw.Actual,
						ExecutionTimeMs: raw.TimeMs,
						ErrorMessage:    raw.Error,
					})
					totalMs += raw.TimeMs
				}
			}
		} else if strings.Contains(line, "[build failed]") ||
			strings.Contains(line, ": undefined:") ||
			strings.Contains(line, ": syntax error:") ||
			strings.Contains(line, ": cannot use") ||
			strings.Contains(line, "compile:") {
			compilerErrors = append(compilerErrors, line)
		}
	}

	// Match outcomes with original test cases if available
	for i := range outcomes {
		if i < len(req.TestCases) {
			if outcomes[i].TestCaseID == uuid.Nil {
				outcomes[i].TestCaseID = req.TestCases[i].ID
			}
			outcomes[i].Hidden = req.TestCases[i].Hidden
			outcomes[i].Input = req.TestCases[i].Input
			outcomes[i].ExpectedOutput = req.TestCases[i].ExpectedOutput
		}
	}

	// If no outcomes parsed but build failed
	if len(outcomes) == 0 && (len(compilerErrors) > 0 || exitCode != 0) {
		compileOutput := strings.Join(compilerErrors, "\n")
		if compileOutput == "" {
			// Capture truncated output as compiler error
			lines := strings.Split(output, "\n")
			if len(lines) > 10 {
				lines = lines[:10]
			}
			compileOutput = strings.Join(lines, "\n")
		}
		return &ExecutionResult{
			Status:          StatusCompileError,
			TotalTests:      len(req.TestCases),
			PassedTests:     0,
			ExecutionTimeMs: totalDurationMs,
			CompilerOutput:  compileOutput,
			TestOutcomes:    nil,
		}
	}

	passCount := 0
	for _, o := range outcomes {
		if o.Passed {
			passCount++
		}
	}

	status := StatusPassed
	if passCount < len(outcomes) || len(outcomes) == 0 {
		status = StatusFailed
	}

	if totalMs == 0 {
		totalMs = totalDurationMs
	}

	return &ExecutionResult{
		Status:          status,
		TotalTests:      len(outcomes),
		PassedTests:     passCount,
		ExecutionTimeMs: totalMs,
		TestOutcomes:    outcomes,
	}
}
