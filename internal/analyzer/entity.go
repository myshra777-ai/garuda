package analyzer

import "time"

// AnalysisResult is the top-level output
type AnalysisResult struct {
	Entities      []Entity       `json:"entities"`
	Relationships []Relationship `json:"relationships"`
	Fingerprint   string         `json:"fingerprint"`
	AnalyzedAt    time.Time      `json:"analyzed_at"`
	Package       string         `json:"package"`
	Source        string         `json:"source"` // root path
	Stats         Stats          `json:"stats"`
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

// Entity represents a discovered type
type Entity struct {
	ID         string   `json:"id"`
	Package    string   `json:"package"`
	File       string   `json:"file"`
	Kind       string   `json:"kind"`
	Name       string   `json:"name"`
	Fields     []Field  `json:"fields"`
	Methods    []Method `json:"methods"`
	Comments   string   `json:"comments"`
	IsExported bool     `json:"is_exported"`
}

// Field represents a struct field or interface method param
type Field struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	JSONTag     string `json:"json_tag,omitempty"`
	DBTag       string `json:"db_tag,omitempty"`
	ValidateTag string `json:"validate_tag,omitempty"`
	Comment     string `json:"comment,omitempty"`
	IsPointer   bool   `json:"is_pointer"`
	IsSlice     bool   `json:"is_slice"`
}

// Method represents a method
type Method struct {
	Name       string `json:"name"`
	Signature  string `json:"signature"`
	IsExported bool   `json:"is_exported"`
}

// Relationship represents a dependency
type Relationship struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"`
}
