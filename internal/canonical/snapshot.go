// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package canonical

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"time"

	"github.com/myshra777-ai/garuda/internal/types"
)

// CanonicalizationVersion identifies the deterministic snapshot serialization schema.
const CanonicalizationVersion = "canonical-snapshot-v1"

// CanonicalizeSnapshot sorts entities, relationships, fields, and methods into a deterministic byte stream.
func CanonicalizeSnapshot(snap types.Snapshot) ([]byte, string, error) {
	// Clone and normalize to avoid mutating input slices and strip runtime timestamps
	entities := make([]types.Entity, len(snap.Entities))
	for i, e := range snap.Entities {
		entities[i] = e
		entities[i].CreatedAt = time.Time{} // Strip wall-clock time for content hashing
	}

	relationships := make([]types.Relationship, len(snap.Relationships))
	for i, r := range snap.Relationships {
		relationships[i] = r
		relationships[i].CreatedAt = time.Time{} // Strip wall-clock time for content hashing
	}

	// 1. Sort Entities deterministically by CanonicalID, then Kind, then Name
	sort.Slice(entities, func(i, j int) bool {
		if entities[i].CanonicalID != entities[j].CanonicalID {
			return entities[i].CanonicalID < entities[j].CanonicalID
		}
		if entities[i].Kind != entities[j].Kind {
			return entities[i].Kind < entities[j].Kind
		}
		return entities[i].Name < entities[j].Name
	})

	// 2. Sort internal entity slices (Fields, Methods)
	for i := range entities {
		if len(entities[i].Fields) > 1 {
			sort.Strings(entities[i].Fields)
		}
		if len(entities[i].Methods) > 1 {
			sort.Strings(entities[i].Methods)
		}
	}

	// 3. Sort Relationships by (SourceID, Predicate, TargetID)
	sort.Slice(relationships, func(i, j int) bool {
		if relationships[i].SourceID != relationships[j].SourceID {
			return relationships[i].SourceID.String() < relationships[j].SourceID.String()
		}
		if relationships[i].Predicate != relationships[j].Predicate {
			return relationships[i].Predicate < relationships[j].Predicate
		}
		return relationships[i].TargetID.String() < relationships[j].TargetID.String()
	})

	canonicalPayload := struct {
		Version       string               `json:"canonical_version"`
		CommitSHA     string               `json:"commit_sha"`
		Entities      []types.Entity       `json:"entities"`
		Relationships []types.Relationship `json:"relationships"`
	}{
		Version:       CanonicalizationVersion,
		CommitSHA:     snap.CommitSHA,
		Entities:      entities,
		Relationships: relationships,
	}

	data, err := json.Marshal(canonicalPayload)
	if err != nil {
		return nil, "", err
	}

	hashBytes := sha256.Sum256(data)
	snapshotHash := hex.EncodeToString(hashBytes[:])

	return data, snapshotHash, nil
}
