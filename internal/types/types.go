package types

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// EvidenceHash is a SHA-256 hash of a content block.
type EvidenceHash [32]byte

// Evidence represents a content‑addressable block that supports a fact or assumption.
type Evidence struct {
	Hash      EvidenceHash `json:"hash"`
	Content   string       `json:"content"`
	RefCount  int          `json:"ref_count"`
	CreatedAt time.Time    `json:"created_at"`
}

// Assumption is a claim that may be false, with confidence.
type Assumption struct {
	ID          uuid.UUID     `json:"id"`
	Statement   string        `json:"statement"`
	Confidence  float64       `json:"confidence"`
	EvidenceIDs []EvidenceHash `json:"evidence_ids"`
	CreatedAt   time.Time     `json:"created_at"`
}

// Fact is a claim believed to be true, with confidence and source.
type Fact struct {
	ID          uuid.UUID     `json:"id"`
	Statement   string        `json:"statement"`
	Confidence  float64       `json:"confidence"`
	EvidenceIDs []EvidenceHash `json:"evidence_ids"`
	CreatedAt   time.Time     `json:"created_at"`
}

// DecisionStatus defines the possible states of a decision.
type DecisionStatus string

const (
	StatusDraft     DecisionStatus = "draft"
	StatusApproved  DecisionStatus = "approved"
	StatusExecuted  DecisionStatus = "executed"
	StatusStale     DecisionStatus = "stale"
)

// Decision is the canonical governance object.
type Decision struct {
	ID               uuid.UUID            `json:"id"`
	TenantID         uuid.UUID            `json:"tenant_id"`
	Title            string               `json:"title"`
	Status           DecisionStatus       `json:"status"`
	Fingerprint      string               `json:"fingerprint"` // SHA-256 of normalized title + scope
	EvidenceIDs      []EvidenceHash       `json:"evidence_ids"`
	TemporalMetadata map[string]interface{} `json:"temporal_metadata"`
	CreatedAt        time.Time            `json:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at"`
}

// DecisionEdge represents a directed lineage link between two decisions.
type DecisionEdge struct {
	TenantID   uuid.UUID `json:"tenant_id"`
	FromID     uuid.UUID `json:"from_id"`
	ToID       uuid.UUID `json:"to_id"`
	EdgeType   string    `json:"edge_type"` // "depends_on", "supersedes"
	CreatedAt  time.Time `json:"created_at"`
}

// DecisionRevision represents an audit snapshot of a decision.
type DecisionRevision struct {
	ID             uuid.UUID    `json:"id"`
	TenantID       uuid.UUID    `json:"tenant_id"`
	DecisionID     uuid.UUID    `json:"decision_id"`
	RevisionNumber int          `json:"revision_number"`
	SnapshotJSON   []byte       `json:"snapshot_json"`
	CreatedAt      time.Time    `json:"created_at"`
}

// Contradiction represents a conflict between two decisions or facts.
type Contradiction struct {
	ID         uuid.UUID `json:"id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	NodeA      uuid.UUID `json:"node_a"`
	NodeB      uuid.UUID `json:"node_b"`
	AuthorityA string    `json:"authority_a"`
	AuthorityB string    `json:"authority_b"`
	Status     string    `json:"status"` // "active", "resolved", "quarantined"
	CreatedAt  time.Time `json:"created_at"`
}

// DecisionStore defines the interface for persisting decisions and evidence.
type DecisionStore interface {
	GetDecision(ctx context.Context, tenantID, decisionID uuid.UUID) (*Decision, error)
	SaveDecision(ctx context.Context, d *Decision) error
	GetDecisionRevisions(ctx context.Context, tenantID, decisionID uuid.UUID) ([]DecisionRevision, error)
	IngestEvidence(ctx context.Context, tenantID uuid.UUID, evidence []Evidence) error
	ConsumeBudget(ctx context.Context, tenantID uuid.UUID, tokens int) error
}
