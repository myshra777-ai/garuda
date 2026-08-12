package analyzer

import "github.com/google/uuid"

// Provenance captures the repository context of an analysis.
type Provenance struct {
	WorkspaceID uuid.UUID `json:"workspace_id,omitempty"`
	RepoID      uuid.UUID `json:"repo_id,omitempty"`
	CommitSHA   string    `json:"commit_sha,omitempty"`
}

// RevisionSummary is the lightweight payload stored in decision_revisions.canonical_json.
// It includes the SHA‑256 hash of the full AST payload, ensuring non‑repudiation.
type RevisionSummary struct {
	Fingerprint string      `json:"fingerprint"`
	Source      string      `json:"source"`
	Stats       Stats       `json:"stats"`
	PayloadHash string      `json:"payload_hash"` // hex of SHA‑256 of full AST JSON
	Provenance  *Provenance `json:"provenance,omitempty"`
}
type Result = AnalysisResult
