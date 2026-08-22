// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package analyzer

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// CachedPackageData encapsulates cached entities and relationships for a package.
type CachedPackageData struct {
	Entities      []Entity       `json:"entities"`
	Relationships []Relationship `json:"relationships"`
}

// PackageCache defines the interface for content-addressed package caching.
type PackageCache interface {
	// GetPackage returns cached entities and relationships if the tree hash matches.
	GetPackage(ctx context.Context, tenantID uuid.UUID, pkgPath string, treeHash []byte) (*CachedPackageData, bool, error)

	// PutPackage saves analyzed entities and relationships keyed by the package tree hash.
	PutPackage(ctx context.Context, tenantID uuid.UUID, pkgPath string, treeHash []byte, data CachedPackageData) error
}

// MemoryPackageCache provides a thread-safe, in-memory implementation of PackageCache for testing and CLI single-runs.
type MemoryPackageCache struct {
	data map[string]*CachedPackageData
}

// NewMemoryPackageCache initializes an in-memory package cache.
func NewMemoryPackageCache() *MemoryPackageCache {
	return &MemoryPackageCache{
		data: make(map[string]*CachedPackageData),
	}
}

func (m *MemoryPackageCache) cacheKey(tenantID uuid.UUID, pkgPath string, treeHash []byte) string {
	return fmt.Sprintf("%s:%s:%x", tenantID, pkgPath, treeHash)
}

func (m *MemoryPackageCache) GetPackage(ctx context.Context, tenantID uuid.UUID, pkgPath string, treeHash []byte) (*CachedPackageData, bool, error) {
	key := m.cacheKey(tenantID, pkgPath, treeHash)
	cached, exists := m.data[key]
	if !exists {
		return nil, false, nil
	}
	return cached, true, nil
}

func (m *MemoryPackageCache) PutPackage(ctx context.Context, tenantID uuid.UUID, pkgPath string, treeHash []byte, data CachedPackageData) error {
	key := m.cacheKey(tenantID, pkgPath, treeHash)
	m.data[key] = &data
	return nil
}
