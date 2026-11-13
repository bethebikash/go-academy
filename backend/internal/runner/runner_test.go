package runner_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"go-academy/backend/internal/runner"
)

func TestProcessRunner_Lifecycle(t *testing.T) {
	r := runner.NewProcessRunner(5 * time.Second)

	// 1. Correct Solution
	validCode := `package main

func TwoSum(nums []int, target int) []int {
	seen := make(map[int]int, len(nums))
	for i, num := range nums {
		diff := target - num
		if j, ok := seen[diff]; ok {
			return []int{j, i}
		}
		seen[num] = i
	}
	return nil
}
`
	ctx1, cancel1 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel1()

	res, err := r.Execute(ctx1, runner.ExecutionRequest{
		ChallengeSlug: "two-sum-in-go",
		Code:          validCode,
		TestCases: []runner.TestCaseInput{
			{
				ID:             uuid.New(),
				Title:          "Sample 1",
				Input:          "[2, 7, 11, 15], 9",
				ExpectedOutput: "[0, 1]",
				Hidden:         false,
			},
		},
		Timeout: 4 * time.Second,
	})
	if err != nil {
		t.Fatalf("expected no error running valid code, got: %v", err)
	}
	if res.Status != runner.StatusPassed {
		t.Errorf("expected status 'passed', got %s. Output: %s", res.Status, res.CompilerOutput)
	}
	if res.PassedTests < 1 {
		t.Errorf("expected at least 1 passed test, got %d", res.PassedTests)
	}

	// 2. Incorrect Solution
	failingCode := `package main

func TwoSum(nums []int, target int) []int {
	return []int{0, 0}
}
`
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()

	failRes, err := r.Execute(ctx2, runner.ExecutionRequest{
		ChallengeSlug: "two-sum-in-go",
		Code:          failingCode,
		TestCases: []runner.TestCaseInput{
			{
				ID:             uuid.New(),
				Title:          "Sample 1",
				Input:          "[2, 7, 11, 15], 9",
				ExpectedOutput: "[0, 1]",
				Hidden:         false,
			},
		},
		Timeout: 4 * time.Second,
	})
	if err != nil {
		t.Fatalf("expected no execution error for failing logic, got: %v", err)
	}
	if failRes.Status != runner.StatusFailed {
		t.Errorf("expected status 'failed', got %s", failRes.Status)
	}

	// 3. Compile Error
	syntaxErrorCode := `package main

func TwoSum(nums []int, target int) []int {
	THIS IS INVALID GO CODE
}
`
	ctx3, cancel3 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel3()

	compileRes, err := r.Execute(ctx3, runner.ExecutionRequest{
		ChallengeSlug: "two-sum-in-go",
		Code:          syntaxErrorCode,
		TestCases:     nil,
		Timeout:       4 * time.Second,
	})
	if err != nil {
		t.Fatalf("unexpected execution driver error: %v", err)
	}
	if compileRes.Status != runner.StatusCompileError {
		t.Errorf("expected status 'compile_error', got %s", compileRes.Status)
	}
	if compileRes.CompilerOutput == "" {
		t.Errorf("expected non-empty compiler output")
	}
}

func TestProcessRunner_TimeoutEnforcement(t *testing.T) {
	r := runner.NewProcessRunner(1 * time.Second)

	infiniteLoopCode := `package main

import "time"

func TwoSum(nums []int, target int) []int {
	for {
		time.Sleep(10 * time.Millisecond)
	}
}
`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	res, err := r.Execute(ctx, runner.ExecutionRequest{
		ChallengeSlug: "two-sum-in-go",
		Code:          infiniteLoopCode,
		TestCases: []runner.TestCaseInput{
			{
				ID:             uuid.New(),
				Title:          "Sample 1",
				Input:          "loop",
				ExpectedOutput: "done",
			},
		},
		Timeout: 500 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("unexpected driver error: %v", err)
	}
	if res.Status != runner.StatusTimedOut {
		t.Errorf("expected status 'timed_out', got %s", res.Status)
	}
}
