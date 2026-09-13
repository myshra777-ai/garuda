// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/myshra777-ai/garuda/internal/auth"
)

// ErrEmailExists is returned by SignupUser when the email address is
// already registered. The users_email_key constraint makes the check
// atomic: two concurrent signups for the same email both pass the
// pre-check, both attempt the insert, and one gets SQLSTATE 23505.
// The handler maps this to a user-facing message.
var ErrEmailExists = errors.New("email already registered")

// SignupUser creates a user, a personal tenant, a default workspace,
// and the two membership rows that make the new user the owner of
// both, all in one transaction.
//
// This is the write path for POST /signup. It is intentionally a
// single method rather than a sequence of calls, because a partial
// failure — user created but tenant insert failed — would leave a
// row in users with no corresponding tenant_members row, and the
// middleware in Session C would refuse to let that user log in.
//
// The personal tenant model: each signup creates a new tenant named
// "<email>'s tenant" with the new user as owner. When invite flows
// land in a later session, additional users can join an existing
// tenant as members. A user who joins a second tenant gets a second
// tenant_members row; the tenant is resolved per request, not stored
// in the session.
//
// The default workspace: named "default", owned by the new user. A
// customer who wants distinct workspaces per team creates more of
// them via the CLI or the API, each scoped to their tenant.
func (s *PostgresStore) SignupUser(
	ctx context.Context,
	email, passwordHash, fullName string,
) (*auth.User, *Workspace, error) {
	userID := uuid.New()
	tenantID := uuid.New()
	workspaceID := uuid.New()
	tenantName := email + "'s tenant"
	now := time.Now().UTC()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("begin signup transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. users
	if _, err := tx.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'user', $5, $5)
	`, userID, email, passwordHash, fullName, now); err != nil {
		if isDuplicateKey(err) {
			return nil, nil, ErrEmailExists
		}
		return nil, nil, fmt.Errorf("insert user: %w", err)
	}

	// 2. tenants — the personal tenant for this signup.
	if _, err := tx.Exec(ctx, `
		INSERT INTO tenants (id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $3)
	`, tenantID, tenantName, now); err != nil {
		return nil, nil, fmt.Errorf("insert tenant: %w", err)
	}

	// 3. workspaces — the default workspace inside the new tenant.
	if _, err := tx.Exec(ctx, `
		INSERT INTO workspaces (id, tenant_id, name, root_path, is_go_work, description, created_at, updated_at)
		VALUES ($1, $2, 'default', '.', false, '', $3, $3)
	`, workspaceID, tenantID, now); err != nil {
		return nil, nil, fmt.Errorf("insert workspace: %w", err)
	}

	// 4. tenant_members — the new user is owner of their tenant.
	if _, err := tx.Exec(ctx, `
		INSERT INTO tenant_members (user_id, tenant_id, role, created_at)
		VALUES ($1, $2, 'owner', $3)
	`, userID, tenantID, now); err != nil {
		return nil, nil, fmt.Errorf("insert tenant member: %w", err)
	}

	// 5. workspace_members — the new user is owner of their workspace.
	if _, err := tx.Exec(ctx, `
		INSERT INTO workspace_members (user_id, workspace_id, role, created_at)
		VALUES ($1, $2, 'owner', $3)
	`, userID, workspaceID, now); err != nil {
		return nil, nil, fmt.Errorf("insert workspace member: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, fmt.Errorf("commit signup transaction: %w", err)
	}

	user := &auth.User{
		ID:           userID.String(),
		Email:        email,
		PasswordHash: passwordHash,
		FullName:     fullName,
		Role:         "user",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	ws := &Workspace{
		ID:          workspaceID,
		TenantID:    tenantID,
		Name:        "default",
		RootPath:    ".",
		IsGoWork:    false,
		Description: "",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return user, ws, nil
}

// isDuplicateKey returns true for a Postgres unique-constraint
// violation (SQLSTATE 23505). Used by SignupUser to convert the
// users_email_key conflict into ErrEmailExists, and by any future
// caller that inserts into a table with a unique constraint.
func isDuplicateKey(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
