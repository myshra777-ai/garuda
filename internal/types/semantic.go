// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"time"

	"github.com/google/uuid"
)

// Predicate constants define typed graph relationship semantics.
const (
	PredicateCalls          = "CALLS"
	PredicateCallsInterface = "CALLS_INTERFACE"
	PredicateImplements     = "IMPLEMENTS"
	PredicateImports        = "IMPORTS"
	PredicateEmbeds         = "EMBEDS"
	PredicateReferences     = "REFERENCES"
	PredicateDependsOn      = "DEPENDS_ON"
)

// EpistemicClass separates observed AST facts from derived inferences (Law 1 & Law 3).
type EpistemicClass string

const (
	EpistemicClassObservation EpistemicClass = "OBSERVATION"
	EpistemicClassInference   EpistemicClass = "INFERENCE"
	EpistemicClassDecision    EpistemicClass = "DECISION"
	EpistemicClassPolicy      EpistemicClass = "POLICY"
)

// ResolutionStatus enforces explicit uncertainty (Law 7).
type ResolutionStatus string

const (
	ResolutionStatusResolved   ResolutionStatus = "RESOLVED"
	ResolutionStatusAmbiguous  ResolutionStatus = "AMBIGUOUS"
	ResolutionStatusUnresolved ResolutionStatus = "UNRESOLVED"
	ResolutionStatusInferred   ResolutionStatus = "INFERRED"
)

// ResolutionMethod records the semantic authority used for resolution.
type ResolutionMethod string

const (
	ResolutionMethodGoTypes          ResolutionMethod = "GO_TYPES"
	ResolutionMethodImportResolution ResolutionMethod = "IMPORT_RESOLUTION"
	ResolutionMethodASTExact         ResolutionMethod = "AST_EXACT"
	ResolutionMethodHeuristic        ResolutionMethod = "HEURISTIC"
	ResolutionMethodGraphDerived     ResolutionMethod = "GRAPH_DERIVED"
)

// EntityKind defines semantic categorization of code nodes.
type EntityKind string

const (
	EntityKindStruct    EntityKind = "struct"
	EntityKindInterface EntityKind = "interface"
	EntityKindMethod    EntityKind = "method"
	EntityKindFunction  EntityKind = "function"
	EntityKindTypeAlias EntityKind = "type_alias"
	EntityKindVariable  EntityKind = "variable"
	EntityKindConstant  EntityKind = "constant"
)

// Severity defines risk levels for diffs and downstream impacts.
type Severity string

const (
	SeverityBreaking    Severity = "BREAKING"
	SeverityNonBreaking Severity = "NON_BREAKING"
	SeverityCritical    Severity = "CRITICAL"
	SeverityHigh        Severity = "HIGH"
	SeverityMedium      Severity = "MEDIUM"
	SeverityLow         Severity = "LOW"
)

// DiffClassification categorizes change impact.
type DiffClassification string

const (
	ClassificationBreaking    DiffClassification = "BREAKING_CHANGE"
	ClassificationNonBreaking DiffClassification = "NON_BREAKING_CHANGE"
)

// Entity represents an AST and semantic node with canonical identity and evidence.
type Entity struct {
	ID               uuid.UUID        `json:"id"`
	CanonicalID      string           `json:"canonical_id"`
	TenantID         uuid.UUID        `json:"tenant_id"`
	WorkspaceID      uuid.UUID        `json:"workspace_id"`
	RepositoryID     uuid.UUID        `json:"repository_id"`
	Name             string           `json:"name"`
	QualifiedName    string           `json:"qualified_name"`
	Kind             EntityKind       `json:"kind"`
	ReceiverType     string           `json:"receiver_type,omitempty"`
	Package          string           `json:"package,omitempty"`
	ModulePath       string           `json:"module_path,omitempty"`
	PackagePath      string           `json:"package_path"`
	FilePath         string           `json:"file_path"`
	File             string           `json:"file,omitempty"`
	Line             int              `json:"line,omitempty"`
	LineStart        int              `json:"line_start"`
	LineEnd          int              `json:"line_end"`
	Exported         bool             `json:"exported"`
	Signature        string           `json:"signature,omitempty"`
	Fields           []string         `json:"fields,omitempty"`
	Methods          []string         `json:"methods,omitempty"`
	Comments         string           `json:"comments,omitempty"`
	ContentSnippet   string           `json:"content_snippet,omitempty"`
	EvidenceHash     string           `json:"evidence_hash,omitempty"`
	SourceHash       string           `json:"source_hash,omitempty"`
	Status           string           `json:"status,omitempty"`
	ResolutionStatus ResolutionStatus `json:"resolution_status,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
}

// Relationship represents a typed, evidence-backed edge between entities.
type Relationship struct {
	ID               uuid.UUID        `json:"id"`
	TenantID         uuid.UUID        `json:"tenant_id"`
	WorkspaceID      uuid.UUID        `json:"workspace_id"`
	RepositoryID     uuid.UUID        `json:"repository_id"`
	SourceID         uuid.UUID        `json:"source_id"`
	SourceName       string           `json:"source_name"`
	TargetID         uuid.UUID        `json:"target_id"`
	TargetName       string           `json:"target_name"`
	Predicate        string           `json:"predicate"`
	Type             string           `json:"type,omitempty"`
	Confidence       float64          `json:"confidence"`
	ResolutionStatus ResolutionStatus `json:"resolution_status,omitempty"`
	ResolutionMethod ResolutionMethod `json:"resolution_method,omitempty"`
	EpistemicClass   EpistemicClass   `json:"epistemic_class"`
	EvidenceHash     string           `json:"evidence_hash,omitempty"`
	LineStart        int              `json:"line_start,omitempty"`
	LineEnd          int              `json:"line_end,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
}

// Snapshot is a deterministic semantic snapshot of a codebase commit.
type Snapshot struct {
	Fingerprint   string         `json:"fingerprint"`
	CommitSHA     string         `json:"commit_sha"`
	Entities      []Entity       `json:"entities"`
	Relationships []Relationship `json:"relationships"`
}

// SemanticDiffChange represents a granular AST modification between snapshots.
type SemanticDiffChange struct {
	ChangeType   string   `json:"change_type"`
	Severity     Severity `json:"severity"`
	TargetEntity string   `json:"target_entity"`
	FieldName    string   `json:"field_name,omitempty"`
	Description  string   `json:"description,omitempty"`
}

// DiffResult captures overall breaking change status and itemized diffs.
type DiffResult struct {
	Classification DiffClassification   `json:"classification"`
	BreakingCount  int                  `json:"breaking_count"`
	Changes        []SemanticDiffChange `json:"changes"`
}

// ImpactedEntity models downstream nodes reached via dependency traversal.
type ImpactedEntity struct {
	QualifiedName string   `json:"qualified_name"`
	Depth         int      `json:"depth"`
	Severity      Severity `json:"severity"`
	Relationship  string   `json:"relationship"`
	IsDirect      bool     `json:"is_direct"`
}

// ImpactReport contains blast radius calculations for a target mutation.
type ImpactReport struct {
	TargetMutation   string           `json:"target_mutation"`
	ImpactedEntities []ImpactedEntity `json:"impacted_entities"`
}

// AnalysisStatus represents the execution lifecycle of a semantic analysis run.
type AnalysisStatus string

const (
	AnalysisStatusPending    AnalysisStatus = "PENDING"
	AnalysisStatusRunning    AnalysisStatus = "RUNNING"
	AnalysisStatusSucceeded  AnalysisStatus = "SUCCEEDED"
	AnalysisStatusFailed     AnalysisStatus = "FAILED"
	AnalysisStatusSuperseded AnalysisStatus = "SUPERSEDED"
)

// AnalysisManifest records provenance, versions, and integrity summaries for a snapshot.
type AnalysisManifest struct {
	AnalysisID              uuid.UUID      `json:"analysis_id"`
	TenantID                uuid.UUID      `json:"tenant_id"`
	WorkspaceID             uuid.UUID      `json:"workspace_id"`
	RepositoryID            uuid.UUID      `json:"repository_id"`
	CommitSHA               string         `json:"commit_sha"`
	AnalyzerVersion         string         `json:"analyzer_version"`
	SchemaVersion           string         `json:"schema_version"`
	IdentityVersion         string         `json:"identity_version"`
	CanonicalizationVersion string         `json:"canonicalization_version"`
	ConfigHash              string         `json:"config_hash,omitempty"`
	EntityCount             int            `json:"entity_count"`
	RelationshipCount       int            `json:"relationship_count"`
	EvidenceCount           int            `json:"evidence_count"`
	SnapshotHash            string         `json:"snapshot_hash"`
	Status                  AnalysisStatus `json:"status"`
	StartedAt               time.Time      `json:"started_at"`
	CompletedAt             *time.Time     `json:"completed_at,omitempty"`
}
