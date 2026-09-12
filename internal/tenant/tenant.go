// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

// Package tenant is the single source of truth for the canonical tenant
// UUID used throughout Garuda.
//
// Before this package, the constant was hardcoded in 22 files across
// cmd/, internal/api/, internal/mcp/, internal/store/, and
// internal/exporter/. Three different resolution behaviors coexisted:
// some callers required GARUDA_TENANT_ID and exited on empty, some fell
// back to the canonical value, and most ignored the env var entirely.
//
// The canonical tenant is the only tenant that exists in the current
// schema. Multi-tenancy is planned; when it arrives, this package is
// where the resolution logic changes.
package tenant

import (
	"fmt"
	"os"

	"github.com/google/uuid"
)

// CanonicalIDStr is the string form of the canonical tenant UUID. Every
// string-typed site references this; no other file should contain the
// literal.
const CanonicalIDStr = "00000000-0000-0000-0000-000000000001"

// CanonicalID is the parsed form of CanonicalIDStr. Parsed once at
// package load; a malformed constant panics at process start, not on
// every call.
var CanonicalID = uuid.MustParse(CanonicalIDStr)

// Resolve returns the tenant UUID for the current process.
//
//	GARUDA_TENANT_ID unset      -> CanonicalID
//	GARUDA_TENANT_ID set        -> the parsed value
//	GARUDA_TENANT_ID malformed  -> error; the caller decides how to fail
//
// An unset env var is the common case; the canonical tenant is the
// default. A malformed env var is a configuration error, not a default:
// silently falling back would write data under the wrong tenant.
func Resolve() (uuid.UUID, error) {
	s := os.Getenv("GARUDA_TENANT_ID")
	if s == "" {
		return CanonicalID, nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, fmt.Errorf("GARUDA_TENANT_ID is not a valid UUID: %w", err)
	}
	return id, nil
}

// MustResolve is Resolve with panic semantics. Intended for package-level
// variable initialization and any path where a resolved tenant is
// required before a caller can return an error.
func MustResolve() uuid.UUID {
	id, err := Resolve()
	if err != nil {
		panic(err)
	}
	return id
}
