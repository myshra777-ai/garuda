// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package auth

import (
	"context"

	"github.com/google/uuid"
)

// WithActor stores the actor identity in the request context.
func WithActor(ctx context.Context, actor string) context.Context {
	return ContextWithActorAndTenant(ctx, actor, uuid.Nil)
}

// WithTenantID stores the tenant identity in the request context.
func WithTenantID(ctx context.Context, tenantID uuid.UUID) context.Context {
	return ContextWithActorAndTenant(ctx, "", tenantID)
}
