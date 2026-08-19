// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package merkle

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// HashDecision computes deterministic SHA-256 digest of decision contents.
func HashDecision(decisionID uuid.UUID, title, status, scopeDomain, scopeSystem, owner string, evidenceIDs []string) string {
	payload := map[string]interface{}{
		"id":           decisionID.String(),
		"title":        title,
		"status":       status,
		"scope_domain": scopeDomain,
		"scope_system": scopeSystem,
		"owner":        owner,
		"evidence_ids": strings.Join(evidenceIDs, ","),
	}

	bytes, _ := json.Marshal(payload)
	hash := sha256.Sum256(bytes)
	return hex.EncodeToString(hash[:])
}

// ChainHash computes SHA256(parentHash + decisionHash)
func ChainHash(parentHash, decisionHash string) string {
	combined := parentHash + decisionHash
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}

// VerifyChain verify whether leaf hash accurately chains into parent hash
func VerifyChain(parentHash, decisionHash, expectedRoot string) bool {
	computed := ChainHash(parentHash, decisionHash)
	return computed == expectedRoot
}

// SnapshotHash computes a deterministic SHA-256 digest for a snapshot:
// SHA256(tenant_id | root_hash | block_height | parent_hash | epoch_unix)
func SnapshotHash(tenantID uuid.UUID, rootHash string, blockHeight int64, parentHash string, epochUnix int64) string {
	data := fmt.Sprintf("%s|%s|%d|%s|%d", tenantID.String(), rootHash, blockHeight, parentHash, epochUnix)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
