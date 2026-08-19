// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package auth

import (
	"context"
	"errors"

	"github.com/myshra777-ai/garuda/internal/types"
)

// AuthService handles user authentication and session management.
type AuthService struct {
	store types.DecisionStore
	jwt   *JWTConfig
}

// NewAuthService creates a new auth service.
func NewAuthService(store types.DecisionStore, jwt *JWTConfig) *AuthService {
	return &AuthService{
		store: store,
		jwt:   jwt,
	}
}

// SignUp registers a new user.
func (s *AuthService) SignUp(ctx context.Context, email, password, fullName string) (*User, string, error) {
	return nil, "", errors.New("not implemented")
}

// SignIn authenticates a user and returns a JWT token.
func (s *AuthService) SignIn(ctx context.Context, email, password string) (*User, string, error) {
	return nil, "", errors.New("not implemented")
}

// ValidateToken validates a JWT token and returns the actor.
func (s *AuthService) ValidateToken(ctx context.Context, token string) (string, error) {
	return "", nil
}
