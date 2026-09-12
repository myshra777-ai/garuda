// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrInvalidCredentials is returned by SignIn for any authentication
// failure. Callers must not distinguish between "email not found" and
// "wrong password" — both return this error.
var ErrInvalidCredentials = errors.New("invalid credentials")

// AuthService handles user authentication.
type AuthService struct {
	store UserStore
	jwt   *JWTConfig
}

// NewAuthService creates a new auth service.
func NewAuthService(store UserStore, jwt *JWTConfig) *AuthService {
	return &AuthService{
		store: store,
		jwt:   jwt,
	}
}

// SignUp registers a new user.
func (s *AuthService) SignUp(ctx context.Context, email, password, fullName string) (*User, string, error) {
	if email == "" || password == "" {
		return nil, "", errors.New("email and password are required")
	}
	if len(password) < 8 {
		return nil, "", errors.New("password must be at least 8 characters")
	}

	if _, err := s.store.GetUserByEmail(ctx, email); err == nil {
		return nil, "", ErrUserExists
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, "", err
	}

	now := time.Now().UTC()
	user := &User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: hash,
		FullName:     fullName,
		Role:         "user",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.store.CreateUser(ctx, user); err != nil {
		return nil, "", err
	}

	token, err := s.jwt.GenerateToken(user.Email, uuid.Nil)
	if err != nil {
		return nil, "", err
	}
	return user, token, nil
}

// SignIn verifies credentials and returns the user on success.
//
// Both "user not found" and "wrong password" return ErrInvalidCredentials
// to prevent email enumeration via timing or error shape.
func (s *AuthService) SignIn(ctx context.Context, email, password string) (*User, error) {
	user, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		// Still perform a bcrypt compare against a dummy hash to
		// equalize timing between "unknown email" and "wrong password".
		_ = CheckPassword(password, "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy")
		return nil, ErrInvalidCredentials
	}
	if err := CheckPassword(password, user.PasswordHash); err != nil {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

// ValidateToken validates a JWT token and returns the actor.
func (s *AuthService) ValidateToken(ctx context.Context, token string) (string, error) {
	actor, _, err := s.jwt.ValidateToken(token)
	return actor, err
}
