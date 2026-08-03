package types

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// TemporalMetadata holds explicit time-to-live and effective applicability constraints.
type TemporalMetadata struct {
	EffectiveFrom *time.Time `json:"effective_from,omitempty"`
	EffectiveTo   *time.Time `json:"effective_to,omitempty"`
	TTLSeconds    int64      `json:"ttl_seconds,omitempty"`
}

// Scope defines the applicability boundary of a decision.
type Scope struct {
	Domain string `json:"domain"`
	System string `json:"system"`
	Team   string `json:"team,omitempty"`
	Env    string `json:"env,omitempty"`
}

// DecisionStatus captures the lifecycle state of a governance decision.
type DecisionStatus string

const (
	StatusDraft       DecisionStatus = "draft"
	StatusReview      DecisionStatus = "review"
	StatusApproved    DecisionStatus = "approved"
	StatusCanonical   DecisionStatus = "canonical"
	StatusSuperseded  DecisionStatus = "superseded"
	StatusArchived    DecisionStatus = "archived"
	StatusDeprecated  DecisionStatus = "deprecated"
	StatusActive      DecisionStatus = "active"
	StatusQuarantined DecisionStatus = "quarantined"
)

func (s DecisionStatus) String() string {
	return string(s)
}

// EvidenceHash is a content-addressable identifier for evidence blocks.
type EvidenceHash [32]byte

func (h EvidenceHash) String() string {
	return hex.EncodeToString(h[:])
}

// Decision is the canonical governance component within the Garuda engine.
type Decision struct {
	ID               uuid.UUID        `json:"id"`
	TenantID         uuid.UUID        `json:"tenant_id"`
	Title            string           `json:"title"`
	Statement        string           `json:"statement,omitempty"`
	Status           DecisionStatus   `json:"status"`
	Scope            Scope            `json:"scope"`
	Owner            string           `json:"owner"`
	Confidence       float64          `json:"confidence"`
	Fingerprint      string           `json:"fingerprint,omitempty"`
	ParentID         *uuid.UUID       `json:"parent_id,omitempty"`
	EvidenceIDs      []EvidenceHash   `json:"evidence_ids,omitempty"`
	TemporalMetadata TemporalMetadata `json:"temporal_metadata,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
	ApprovedAt       *time.Time       `json:"approved_at,omitempty"`
	ScopeDomain      string           `json:"scope_domain"`
	ScopeSystem      string           `json:"scope_system"`
	MerkleHash       string           `json:"merkle_hash,omitempty"`
	ParentMerkleHash string           `json:"parent_merkle_hash,omitempty"`
}

// DecisionRevision snapshots an earlier version of a decision.
type DecisionRevision struct {
	ID             uuid.UUID `json:"id"`
	TenantID       uuid.UUID `json:"tenant_id"`
	DecisionID     uuid.UUID `json:"decision_id"`
	RevisionNumber int       `json:"revision_number"`
	SnapshotJSON   []byte    `json:"snapshot_json"`
	CreatedAt      time.Time `json:"created_at"`
}

// Contradiction represents conflicting governance decisions.
type Contradiction struct {
	ID                 uuid.UUID  `json:"id"`
	TenantID           uuid.UUID  `json:"tenant_id"`
	DecisionA          uuid.UUID  `json:"decision_a"`
	DecisionB          uuid.UUID  `json:"decision_b"`
	Severity           string     `json:"severity"`
	Quarantined        bool       `json:"quarantined"`
	Resolved           bool       `json:"resolved"`
	ResolutionStrategy string     `json:"resolution_strategy"` // human, auto_supersede
	CreatedAt          time.Time  `json:"created_at"`
	ResolvedAt         *time.Time `json:"resolved_at,omitempty"`
	AutoResolvedAt     *time.Time `json:"auto_resolved_at,omitempty"`
}

// Evidence represents a structured artifact ingested into the store.
type Evidence struct {
	Hash      EvidenceHash `json:"hash"`
	Content   string       `json:"content"`
	RefCount  int          `json:"ref_count"`
	CreatedAt time.Time    `json:"created_at"`
}

// Block represents a content-addressable evidence block.
type Block struct {
	Hash      EvidenceHash `json:"hash"`
	Content   string       `json:"content"`
	RefCount  int          `json:"ref_count"`
	CreatedAt time.Time    `json:"created_at"`
}

// TaskManifest records agent execution context and associated decisions.
type TaskManifest struct {
	TaskID         string         `json:"task_id"`
	CustomerID     string         `json:"customer_id"`
	CredentialRef  string         `json:"credential_ref,omitempty"`
	Title          string         `json:"title"`
	ScopeDomain    string         `json:"scope_domain"`
	ScopeSystem    string         `json:"scope_system"`
	Status         string         `json:"status"`
	ManifestBlocks []EvidenceHash `json:"manifest_blocks"`
	NormalizedIR   string         `json:"normalized_ir,omitempty"`
	IRVersion      int            `json:"ir_version"`
	DecisionIDs    []uuid.UUID    `json:"decision_ids"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// DecisionStore defines the multi-tenant contract for persisting governance objects.

// Checkpoint represents an agent's saved context state.
type Checkpoint struct {
	ID             uuid.UUID       `json:"id"`
	TenantID       uuid.UUID       `json:"tenant_id"`
	AgentID        string          `json:"agent_id"`
	TaskID         *uuid.UUID      `json:"task_id,omitempty"`
	CheckpointData json.RawMessage `json:"checkpoint_data"`
	Status         string          `json:"status"` // active, completed, transferred, expired
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	ExpiresAt      *time.Time      `json:"expires_at,omitempty"`
}

// DecisionStore defines the multi-tenant contract for persisting governance and context objects.
type DecisionStore interface {
	GetDecision(ctx context.Context, tenantID, decisionID uuid.UUID) (*Decision, error)
	SaveDecision(ctx context.Context, d *Decision) error
	GetDecisionsByScope(ctx context.Context, tenantID uuid.UUID, domain, system string) ([]*Decision, error)
	GetDecisionRevisions(ctx context.Context, tenantID, decisionID uuid.UUID) ([]DecisionRevision, error)
	ListDecisions(ctx context.Context, tenantID uuid.UUID, scope Scope, statuses []DecisionStatus) ([]*Decision, error)
	ListDecisionsByParent(ctx context.Context, tenantID, parentID uuid.UUID) ([]*Decision, error)
	ListContradictions(ctx context.Context, tenantID uuid.UUID, resolved bool) ([]Contradiction, error)
	GetContradiction(ctx context.Context, tenantID, id uuid.UUID) (*Contradiction, error)
	IngestEvidence(ctx context.Context, tenantID uuid.UUID, evidence []Evidence) error
	ConsumeBudget(ctx context.Context, tenantID uuid.UUID, tokens int) error

	// Checkpoint methods
	SaveCheckpoint(ctx context.Context, c *Checkpoint) error
	GetCheckpoint(ctx context.Context, tenantID, id uuid.UUID) (*Checkpoint, error)
	ListCheckpointsByAgent(ctx context.Context, tenantID uuid.UUID, agentID string, limit int) ([]*Checkpoint, error)

	// Budget & Metering methods
	GetTenantBudget(ctx context.Context, tenantID uuid.UUID) (*TenantBudget, error)
	PreflightCheckAndReserve(ctx context.Context, tenantID uuid.UUID, estimatedTokens int) error
	ConsumeBudgetDeduct(ctx context.Context, tenantID uuid.UUID, req BudgetConsumptionRequest) (*BudgetConsumptionResponse, error)

	// Quarantine methods
	QuarantineDecision(ctx context.Context, tenantID uuid.UUID, proposedID, conflictingID uuid.UUID, domain, system, reason string) (*Contradiction, error)
	ListUnresolvedContradictions(ctx context.Context, tenantID uuid.UUID) ([]Contradiction, error)
	ResolveContradiction(ctx context.Context, id uuid.UUID, strategy string) error

	// Merkle & Evidence Chain Methods
	GetMerkleRoot(ctx context.Context, tenantID uuid.UUID) (*MerkleRoot, error)
	AppendMerkleChain(ctx context.Context, tenantID uuid.UUID, decisionHash string) (*MerkleRoot, error)
	AddEvidenceBlock(ctx context.Context, tenantID, decisionID uuid.UUID, payload any) (*EvidenceBlock, error)
}
