// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWTGenerationAndValidation(t *testing.T) {
	config, err := NewJWTConfig("garuda", "garuda-api", 15*time.Minute)
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	actor := "alice@company.com"
	token, err := config.GenerateToken(actor, uuid.Nil)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	validatedActor, _, err := config.ValidateToken(token)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}
	if validatedActor != actor {
		t.Errorf("expected actor %s, got %s", actor, validatedActor)
	}
}

func TestRefreshToken(t *testing.T) {
	config, err := NewJWTConfig("garuda", "garuda-api", 7*24*time.Hour)
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	actor := "alice@company.com"
	refreshToken, err := config.GenerateRefreshToken(actor, uuid.Nil)
	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}

	validatedActor, _, err := config.ValidateRefreshToken(refreshToken)
	if err != nil {
		t.Fatalf("failed to validate refresh token: %v", err)
	}
	if validatedActor != actor {
		t.Errorf("expected actor %s, got %s", actor, validatedActor)
	}
}

func TestInvalidToken(t *testing.T) {
	config, err := NewJWTConfig("garuda", "garuda-api", 15*time.Minute)
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	// Tampered token
	invalidToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

	_, _, err = config.ValidateToken(invalidToken)
	if err == nil {
		t.Error("expected error for invalid token, got nil")
	}
}
