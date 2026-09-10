package knowledge

import (
	"time"

	"github.com/google/uuid"
)

// Modality represents normative strength (RFC 2119).
type Modality string

const (
	ModalityMust      Modality = "MUST"
	ModalityMustNot   Modality = "MUST_NOT"
	ModalityShould    Modality = "SHOULD"
	ModalityShouldNot Modality = "SHOULD_NOT"
	ModalityMay       Modality = "MAY"
)

// Predicate defines the structural or runtime relationship asserted.
type Predicate string

const (
	PredicateCalls      Predicate = "CALLS"
	PredicateImplements Predicate = "IMPLEMENTS"
	PredicateEnforces   Predicate = "ENFORCES"
	PredicateExposes    Predicate = "EXPOSES"
	PredicateIdempotent Predicate = "IDEMPOTENT"
	PredicateDependsOn  Predicate = "DEPENDS_ON"
)

// ProvenanceClass tracks how the claim was obtained.
type ProvenanceClass string

const (
	ProvenanceExtracted ProvenanceClass = "EXTRACTED" // Direct deterministic extraction
	ProvenanceDeclared  ProvenanceClass = "DECLARED"  // Explicit author annotation
	ProvenanceInferred  ProvenanceClass = "INFERRED"  // Heuristic or model candidate
)

// VerificationStatus tracks the claim's verification state against code/runtime.
type VerificationStatus string

const (
	StatusSupported    VerificationStatus = "SUPPORTED"
	StatusUnverified   VerificationStatus = "UNVERIFIED"
	StatusContradicted VerificationStatus = "CONTRADICTED"
)

// ClaimIR represents an atomic, verifiable intent assertion extracted from text.
type ClaimIR struct {
	ID               uuid.UUID          `json:"id"`
	Workspace        string             `json:"workspace"`
	TenantID         uuid.UUID          `json:"tenant_id"`
	DocumentPath     string             `json:"document_path"`
	DocumentTitle    string             `json:"document_title"`
	DocumentType     string             `json:"document_type"`
	SectionTitle     string             `json:"section_title"`
	LineStart        int                `json:"line_start"`
	LineEnd          int                `json:"line_end"`
	RawStatement     string             `json:"raw_statement"`
	Subject          string             `json:"subject"`
	Modality         Modality           `json:"modality"`
	Predicate        Predicate          `json:"predicate"`
	Object           string             `json:"object"`
	Provenance       ProvenanceClass    `json:"provenance_class"`
	Confidence       float64            `json:"confidence"`
	Status           VerificationStatus `json:"status"`
	MatchedEntityID  *uuid.UUID         `json:"matched_entity_id,omitempty"`
	ContradictionMsg string             `json:"contradiction_reason,omitempty"`
	CreatedAt        time.Time          `json:"created_at"`
}
