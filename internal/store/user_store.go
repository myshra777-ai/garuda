// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/myshra777-ai/garuda/internal/auth"
)

func (s *PostgresStore) CreateUser(ctx context.Context, user *auth.User) error {
	_, err := s.pool.Exec(ctx,
		"INSERT INTO users (id, email, password_hash, full_name, role, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		user.ID, user.Email, user.PasswordHash, user.FullName, user.Role, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (s *PostgresStore) GetUserByEmail(ctx context.Context, email string) (*auth.User, error) {
	var user auth.User
	err := s.pool.QueryRow(ctx,
		"SELECT id, email, password_hash, full_name, role, created_at, updated_at FROM users WHERE email = $1",
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, auth.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}
	return &user, nil
}
