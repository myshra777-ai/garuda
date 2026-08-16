package analyzer

import "time"

// EntityKind defines the type of a semantic entity.
type EntityKind string

const (
	KindRepository EntityKind = "repository"
	KindDirectory  EntityKind = "directory"
	KindFile       EntityKind = "file"
	KindPackage    EntityKind = "package"
	KindStruct     EntityKind = "struct"
	KindInterface  EntityKind = "interface"
	KindFunction   EntityKind = "function"
	KindMethod     EntityKind = "method"
	KindVariable   EntityKind = "variable"
	KindConstant   EntityKind = "constant"
	KindExternal   EntityKind = "external"
)

// RelationshipType defines the type of a relationship.
type RelationshipType string

const (
	RelContains   RelationshipType = "CONTAINS"
	RelDefines    RelationshipType = "DEFINES"
	RelImports    RelationshipType = "IMPORTS"
	RelCalls      RelationshipType = "CALLS"
	RelReferences RelationshipType = "REFERENCES"
	RelImplements RelationshipType = "IMPLEMENTS"
	RelEmbeds     RelationshipType = "EMBEDS"
	RelDependsOn  RelationshipType = "DEPENDS_ON"
)

// Entity represents a discovered type (struct, interface, or function)
type Entity struct {
	ID        string     `json:"id"`
	Kind      EntityKind `json:"kind"`
	Name      string     `json:"name"`
	Package   string     `json:"package"`
	File      string     `json:"file"`
	Line      int        `json:"line"`
	Exported  bool       `json:"exported"`
	Fields    []Field    `json:"fields,omitempty"`
	Methods   []Method   `json:"methods,omitempty"`
	Signature string     `json:"signature,omitempty"`
	Comments  string     `json:"comments,omitempty"`
}

// Field represents a struct field or interface method parameter
type Field struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Tag         string `json:"tag,omitempty"`
	JSONTag     string `json:"json_tag,omitempty"`
	DBTag       string `json:"db_tag,omitempty"`
	ValidateTag string `json:"validate_tag,omitempty"`
	Comment     string `json:"comment,omitempty"`
	IsPointer   bool   `json:"is_pointer"`
	IsSlice     bool   `json:"is_slice"`
}

// Method represents a method on a struct or interface
type Method struct {
	Name       string   `json:"name"`
	Signature  string   `json:"signature"`
	Parameters []string `json:"parameters,omitempty"`
	IsExported bool     `json:"is_exported"`
}

// Relationship represents a dependency
type Relationship struct {
	From       string           `json:"from"`
	To         string           `json:"to"`
	Type       RelationshipType `json:"type"`
	Confidence float64          `json:"confidence"`
	Evidence   Evidence         `json:"evidence"`
}

// Evidence anchors a relationship to source code and provenance.
type Evidence struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Commit   string `json:"commit"`
	Analyzer string `json:"analyzer"`
	Snapshot string `json:"snapshot,omitempty"`
}

// Stats holds summary counts
type Stats struct {
	Files       int `json:"files"`
	Packages    int `json:"packages"`
	Structs     int `json:"structs"`
	Interfaces  int `json:"interfaces"`
	Functions   int `json:"functions"`
	Imports     int `json:"imports"`
	TotalFields int `json:"total_fields"`
}

// Result is the top-level analysis output
type Result struct {
	Entities      []Entity       `json:"entities"`
	Relationships []Relationship `json:"relationships"`
	Fingerprint   string         `json:"fingerprint"`
	AnalyzedAt    time.Time      `json:"analyzed_at"`
	Package       string         `json:"package"`
	Source        string         `json:"source"`
	Stats         Stats          `json:"stats"`
}

// RevisionSummary is the lightweight summary stored in canonical_json
type RevisionSummary struct {
	Fingerprint string     `json:"fingerprint"`
	Stats       Stats      `json:"stats"`
	PayloadHash string     `json:"payload_hash"`
	Provenance  Provenance `json:"provenance,omitempty"`
	Source      string     `json:"source,omitempty"`
	Entities    int        `json:"entities"`
	Claims      int        `json:"claims"`
}

// Provenance tracks where the analysis came from
type Provenance struct {
	WorkspaceID  string `json:"workspace_id,omitempty"`
	RepositoryID string `json:"repository_id,omitempty"`
	CommitSHA    string `json:"commit_sha,omitempty"`
	SourcePath   string `json:"source_path,omitempty"`
}

// ----------------------------------------------------------------------------
// Diff-related types
// ----------------------------------------------------------------------------

// DiffReport holds the complete diff result
type DiffReport struct {
	StatsDiff         StatsDiff          `json:"stats_diff"`
	FingerprintDiff   FingerprintDiff    `json:"fingerprint_diff"`
	EntityDiffs       []EntityDiff       `json:"entity_diffs"`
	RelationshipDiffs []RelationshipDiff `json:"relationship_diffs"`
	Summary           DiffSummary        `json:"summary"`
}

// StatsDiff shows numeric changes
type StatsDiff struct {
	Files       int `json:"files"`
	Packages    int `json:"packages"`
	Structs     int `json:"structs"`
	Interfaces  int `json:"interfaces"`
	Functions   int `json:"functions"`
	Imports     int `json:"imports"`
	TotalFields int `json:"total_fields"`
}

// FingerprintDiff shows if fingerprints match
type FingerprintDiff struct {
	Before string `json:"before"`
	After  string `json:"after"`
	Match  bool   `json:"match"`
}

// EntityDiff represents a change to an entity
type EntityDiff struct {
	EntityID    string       `json:"entity_id"`
	Kind        EntityKind   `json:"kind"` // now EntityKind, not string
	Name        string       `json:"name"`
	File        string       `json:"file"`
	Status      string       `json:"status"`
	FieldsDiff  *FieldsDiff  `json:"fields_diff,omitempty"`
	MethodsDiff *MethodsDiff `json:"methods_diff,omitempty"`
	Impact      int          `json:"impact"`
}

// FieldsDiff shows field changes
type FieldsDiff struct {
	Added    []Field     `json:"added"`
	Removed  []Field     `json:"removed"`
	Modified []FieldDiff `json:"modified"`
}

// FieldDiff shows a field change (before/after)
type FieldDiff struct {
	Name   string `json:"name"`
	Before Field  `json:"before"`
	After  Field  `json:"after"`
}

// MethodsDiff shows method changes
type MethodsDiff struct {
	Added   []Method `json:"added"`
	Removed []Method `json:"removed"`
}

// RelationshipDiff represents a change to a relationship
type RelationshipDiff struct {
	From   string           `json:"from"`
	To     string           `json:"to"`
	Type   RelationshipType `json:"type"` // now RelationshipType
	Status string           `json:"status"`
}

// DiffSummary provides a high-level overview
type DiffSummary struct {
	BreakingChanges int `json:"breaking_changes"`
	Warnings        int `json:"warnings"`
	Additions       int `json:"additions"`
	Removals        int `json:"removals"`
	Modified        int `json:"modified"`
}
