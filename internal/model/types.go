package model

import (
	"time"

	"github.com/google/uuid"
)

// ─── Epistemic Classification ───────────────────────────────────────────────

type EpistemicClass string

const (
	OBSERVED_CODE    EpistemicClass = "OBSERVED_CODE"
	OBSERVED_RUNTIME EpistemicClass = "OBSERVED_RUNTIME"
	OBSERVED_CONFIG  EpistemicClass = "OBSERVED_CONFIG"
	OBSERVED_DOC     EpistemicClass = "OBSERVED_DOC"
	INFERRED         EpistemicClass = "INFERRED"
	DECISION         EpistemicClass = "DECISION"
	POLICY           EpistemicClass = "POLICY"
	VERIFIED         EpistemicClass = "VERIFIED"
	CONFLICTED       EpistemicClass = "CONFLICTED"
)

type ClaimStatus string

const (
	ClaimProposed     ClaimStatus = "PROPOSED"
	ClaimVerified     ClaimStatus = "VERIFIED"
	ClaimContradicted ClaimStatus = "CONTRADICTED"
	ClaimDeprecated   ClaimStatus = "DEPRECATED"
)

// ─── Entity Identity ─────────────────────────────────────────────────────────

type EntityKind string

const (
	KindStruct    EntityKind = "STRUCT"
	KindInterface EntityKind = "INTERFACE"
	KindFunction  EntityKind = "FUNCTION"
	KindMethod    EntityKind = "METHOD"
	KindService   EntityKind = "SERVICE"
	KindAPI       EntityKind = "API"
	KindSchema    EntityKind = "SCHEMA"
	KindDatabase  EntityKind = "DATABASE"
	KindEvent     EntityKind = "EVENT"
	KindPackage   EntityKind = "PACKAGE"
)

type Entity struct {
	ID                  uuid.UUID  `json:"id"`
	TenantID            uuid.UUID  `json:"tenant_id"`
	RepositoryID        uuid.UUID  `json:"repository_id"`
	CanonicalURN        string     `json:"canonical_urn"` // go://tenant/repo/commit/package/name
	Name                string     `json:"name"`
	Kind                EntityKind `json:"kind"`
	Package             string     `json:"package"`
	OccurrenceID        string     `json:"occurrence_id"`        // commit-scoped symbol ID
	SemanticFingerprint string     `json:"semantic_fingerprint"` // normalized AST hash
	LineageID           string     `json:"lineage_id"`           // historical conceptual identity
	FirstSeenCommit     string     `json:"first_seen_commit"`
	LastSeenCommit      string     `json:"last_seen_commit"`
	LatestSnapshotID    uuid.UUID  `json:"latest_snapshot_id"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// ─── Provenance / Evidence ──────────────────────────────────────────────────

type Provenance struct {
	RepositoryID    uuid.UUID `json:"repository_id"`
	CommitSHA       string    `json:"commit_sha"`
	FilePath        string    `json:"file_path"`
	Symbol          string    `json:"symbol"`
	LineStart       int       `json:"line_start"`
	LineEnd         int       `json:"line_end"`
	Content         string    `json:"content"`      // short snippet
	ContentHash     string    `json:"content_hash"` // SHA256
	Language        string    `json:"language"`
	Analyzer        string    `json:"analyzer"`
	AnalyzerVersion string    `json:"analyzer_version"`
	CapturedAt      time.Time `json:"captured_at"`
}

// ─── Claim ──────────────────────────────────────────────────────────────────

type Claim struct {
	ID          uuid.UUID      `json:"id"`
	TenantID    uuid.UUID      `json:"tenant_id"`
	Subject     *Entity        `json:"subject,omitempty"`
	SubjectID   uuid.UUID      `json:"subject_entity_id"`
	Predicate   string         `json:"predicate"`
	Object      *Entity        `json:"object,omitempty"`
	ObjectID    uuid.UUID      `json:"object_entity_id"`
	Class       EpistemicClass `json:"class"`
	Confidence  float64        `json:"confidence"`
	Status      ClaimStatus    `json:"status"`
	EvidenceIDs []uuid.UUID    `json:"evidence_ids"`
	ValidFrom   time.Time      `json:"valid_from"`
	ValidTo     *time.Time     `json:"valid_to,omitempty"`
	SnapshotID  uuid.UUID      `json:"snapshot_id"`
	CreatedAt   time.Time      `json:"created_at"`
}

// ─── Repository Snapshot (Immutable Artifact) ──────────────────────────────

type RepositorySnapshot struct {
	ID                  uuid.UUID     `json:"id"`
	RepositoryID        uuid.UUID     `json:"repository_id"`
	CommitSHA           string        `json:"commit_sha"`
	AnalyzerName        string        `json:"analyzer_name"`
	AnalyzerVersion     string        `json:"analyzer_version"`
	SourceFingerprint   string        `json:"source_fingerprint"`
	SemanticFingerprint string        `json:"semantic_fingerprint"`
	SchemaVersion       int           `json:"schema_version"`
	Status              string        `json:"status"` // SUCCESS, PARTIAL, FAILED
	StartedAt           time.Time     `json:"started_at"`
	CompletedAt         *time.Time    `json:"completed_at,omitempty"`
	Entities            []*Entity     `json:"entities"`
	Claims              []*Claim      `json:"claims"`
	Evidence            []*Provenance `json:"evidence"`
	ErrorSummary        string        `json:"error_summary,omitempty"`
	CreatedAt           time.Time     `json:"created_at"`
}

// ─── Analysis Artifact (DB representation) ────────────────────────────────

type AnalysisArtifact struct {
	ID                  uuid.UUID           `json:"id"`
	RepositoryID        uuid.UUID           `json:"repository_id"`
	CommitSHA           string              `json:"commit_sha"`
	AnalyzerName        string              `json:"analyzer_name"`
	AnalyzerVersion     string              `json:"analyzer_version"`
	SourceFingerprint   string              `json:"source_fingerprint"`
	SemanticFingerprint string              `json:"semantic_fingerprint"`
	SchemaVersion       int                 `json:"schema_version"`
	Status              string              `json:"status"`
	StartedAt           time.Time           `json:"started_at"`
	CompletedAt         *time.Time          `json:"completed_at,omitempty"`
	Model               *RepositorySnapshot `json:"model"` // full snapshot
	ErrorSummary        string              `json:"error_summary,omitempty"`
	CreatedAt           time.Time           `json:"created_at"`
}
