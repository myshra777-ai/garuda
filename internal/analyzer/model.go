package analyzer

import "time"

// Entity represents a discovered type (struct, interface, or function)
type Entity struct {
	ID         string   `json:"id"` // package + name
	Package    string   `json:"package"`
	File       string   `json:"file"`
	Kind       string   `json:"kind"` // "struct", "interface", "function", "type"
	Name       string   `json:"name"`
	Fields     []Field  `json:"fields,omitempty"`
	Methods    []Method `json:"methods,omitempty"`
	Signature  string   `json:"signature,omitempty"` // for functions
	Comments   string   `json:"comments,omitempty"`
	IsExported bool     `json:"is_exported"`
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
	From    string `json:"from"` // entity ID
	To      string `json:"to"`
	Type    string `json:"type"` // "imports", "calls", "embeds", "references"
	Package string `json:"package,omitempty"`
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
