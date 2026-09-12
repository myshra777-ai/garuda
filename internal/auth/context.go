// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package auth

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const (
	actorKey  contextKey = "actor"
	tenantKey contextKey = "tenant_id"
)

// ContextWithActorAndTenant stores authenticated request identity in context.
func ContextWithActorAndTenant(ctx context.Context, actor string, tenant interface{}) context.Context {
	ctx = context.WithValue(ctx, actorKey, actor)
	return context.WithValue(ctx, tenantKey, tenant)
}

// ActorFromContext returns the authenticated actor when present.
func ActorFromContext(ctx context.Context) (string, bool) {
	actor, ok := ctx.Value(actorKey).(string)
	return actor, ok && actor != ""
}

// TenantIDFromContext returns a tenant UUID stored as a UUID or UUID string.
func TenantIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	switch tenant := ctx.Value(tenantKey).(type) {
	case uuid.UUID:
		return tenant, tenant != uuid.Nil
	case string:
		id, err := uuid.Parse(tenant)
		return id, err == nil && id != uuid.Nil
	default:
		return uuid.Nil, false
	}
}

// WithActor stores the actor identity in the request context.
func WithActor(ctx context.Context, actor string) context.Context {
	return ContextWithActorAndTenant(ctx, actor, uuid.Nil)
}

// WithTenantID stores the tenant identity in the request context.
func WithTenantID(ctx context.Context, tenantID uuid.UUID) context.Context {
	return ContextWithActorAndTenant(ctx, "", tenantID)
}
