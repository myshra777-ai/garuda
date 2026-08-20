// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package canonical

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/types"
)

// IdentityVersion defines the active scheme for entity and relationship hashing.
const IdentityVersion = "go-canonical-v1"

// GarudaEntityNamespace is the fixed UUIDv5 namespace for all deterministic semantic entities.
var GarudaEntityNamespace = uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

// EntityKeySpec encapsulates all dimensions required to compute a stable semantic identity.
type EntityKeySpec struct {
	Kind         types.EntityKind `json:"kind"`
	PackagePath  string           `json:"package_path"`
	ReceiverType string           `json:"receiver_type,omitempty"`
	Name         string           `json:"name"`
}

// NormalizeReceiver strips redundant parentheses or spaces while preserving pointer semantics.
func NormalizeReceiver(recv string) string {
	recv = strings.TrimSpace(recv)
	recv = strings.TrimPrefix(recv, "(")
	recv = strings.TrimSuffix(recv, ")")
	return strings.TrimSpace(recv)
}

// BuildEntityCanonicalID creates an authoritative string key for an AST entity.
// Format: <version>:<kind>:<package_path>:[<normalized_receiver>:]<name>
func BuildEntityCanonicalID(spec EntityKeySpec) string {
	pkg := strings.TrimSpace(spec.PackagePath)
	name := strings.TrimSpace(spec.Name)
	kind := strings.ToLower(strings.TrimSpace(string(spec.Kind)))

	if spec.ReceiverType != "" {
		recv := NormalizeReceiver(spec.ReceiverType)
		return fmt.Sprintf("%s:%s:%s:(%s).%s", IdentityVersion, kind, pkg, recv, name)
	}

	return fmt.Sprintf("%s:%s:%s:%s", IdentityVersion, kind, pkg, name)
}

// GenerateEntityUUID derives a deterministic UUIDv5 from the canonical string ID.
func GenerateEntityUUID(spec EntityKeySpec) (uuid.UUID, string) {
	canonicalID := BuildEntityCanonicalID(spec)
	entityUUID := uuid.NewSHA1(GarudaEntityNamespace, []byte(canonicalID))
	return entityUUID, canonicalID
}

// RelationshipKeySpec defines the endpoints and predicate for an edge identity.
type RelationshipKeySpec struct {
	SourceID  uuid.UUID `json:"source_id"`
	Predicate string    `json:"predicate"`
	TargetID  uuid.UUID `json:"target_id"`
	Qualifier string    `json:"qualifier,omitempty"`
}

// BuildRelationshipCanonicalID formats a deterministic key for semantic graph edges.
func BuildRelationshipCanonicalID(spec RelationshipKeySpec) string {
	pred := strings.ToUpper(strings.TrimSpace(spec.Predicate))
	if spec.Qualifier != "" {
		return fmt.Sprintf("%s:%s:%s:%s:%s", IdentityVersion, spec.SourceID, pred, spec.TargetID, spec.Qualifier)
	}
	return fmt.Sprintf("%s:%s:%s:%s", IdentityVersion, spec.SourceID, pred, spec.TargetID)
}

// GenerateRelationshipUUID derives a deterministic UUIDv5 for a relationship edge.
func GenerateRelationshipUUID(spec RelationshipKeySpec) uuid.UUID {
	canonicalID := BuildRelationshipCanonicalID(spec)
	return uuid.NewSHA1(GarudaEntityNamespace, []byte(canonicalID))
}
