package hasher_test

import (
	"testing"

	"go-academy/backend/internal/platform/hasher"
)

func TestHashAndVerifyPassword(t *testing.T) {
	password := "SecretP@ssw0rd!2026"

	// Use smaller params for fast testing
	testParams := &hasher.Params{
		Memory:      16 * 1024,
		Iterations:  1,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	}

	encoded, err := hasher.HashPasswordWithParams(password, testParams)
	if err != nil {
		t.Fatalf("expected no error hashing password, got: %v", err)
	}

	// Verify correct password
	match, err := hasher.VerifyPassword(password, encoded)
	if err != nil {
		t.Fatalf("expected no error verifying password, got: %v", err)
	}
	if !match {
		t.Fatalf("expected password to match, but it did not")
	}

	// Verify wrong password
	match, err = hasher.VerifyPassword("WrongPassword123", encoded)
	if err != nil {
		t.Fatalf("expected no error verifying wrong password, got: %v", err)
	}
	if match {
		t.Fatalf("expected wrong password NOT to match, but it did")
	}

	// Verify malformed hash
	_, err = hasher.VerifyPassword(password, "invalid$hash$string")
	if err == nil {
		t.Fatalf("expected error for malformed hash, got nil")
	}
}
