// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

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
	Region string `json:"region,omitempty"`
}

// DecisionStatus captures the lifecycle state of a governance decision.
type DecisionStatus string

// Add this struct at the end of the file
type EvaluationTest struct {
	ID   string
	Name string
}

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

	// Bitemporal fields (GAS Vol 009)
	ValidFrom time.Time  `json:"valid_from"`         // When this decision becomes effective
	ValidTo   *time.Time `json:"valid_to,omitempty"` // When this decision expires (NULL = indefinite)
}

// DecisionRevision snapshots an earlier version of a decision.

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

	// Topology methods
	SaveTopology(ctx context.Context, top *Topology) error
	GetTopology(ctx context.Context, id uuid.UUID) (*Topology, error)
	GetTasksByTopology(ctx context.Context, topologyID uuid.UUID) ([]*Task, error)
	UpdateTopologyStatus(ctx context.Context, id uuid.UUID, status TopologyStatus) error
	UpdateTask(ctx context.Context, task *Task) error
	UpdateTopologyTokens(ctx context.Context, id uuid.UUID, tokens int64) error

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

	// Merkle Snapshot Methods
	ListAllTenants(ctx context.Context) ([]uuid.UUID, error)
	GetLatestMerkleSnapshot(ctx context.Context, tenantID uuid.UUID) (*MerkleSnapshot, error)
	SaveMerkleSnapshot(ctx context.Context, snap *MerkleSnapshot) error
	ListMerkleSnapshots(ctx context.Context, tenantID uuid.UUID, limit int) ([]MerkleSnapshot, error)

	// Temporal Methods
	GetDecisionsActiveAt(ctx context.Context, tenantID uuid.UUID, at time.Time, scope Scope, statuses []DecisionStatus) ([]*Decision, error)
	GetDecisionHistory(ctx context.Context, tenantID, decisionID uuid.UUID) ([]*Decision, error)

	// Audit Trail Capabilities
	LogAuditEvent(ctx context.Context, tenantID uuid.UUID, eventType string, eventID uuid.UUID, actor string, payload interface{}) (*AuditEvent, error)
	VerifyAuditEvent(ctx context.Context, tenantID uuid.UUID, eventID uuid.UUID) (*AuditVerification, error)
	ListAuditEvents(ctx context.Context, tenantID uuid.UUID, since time.Time) ([]AuditEvent, error)

	// Plan assembling
	GetPlan(ctx context.Context, tenantID uuid.UUID, req *PlanRequest) (*PlanResult, error)

	// SubmitDecision atomically creates an immutable revision and returns result.
	SubmitDecision(ctx context.Context, req *SubmitDecisionRequest, actor, requestID string) (*SubmitDecisionResult, error)

	// Policy methods
	SavePolicy(ctx context.Context, p *Policy) error
	GetActivePolicies(ctx context.Context, tenantID uuid.UUID, scopeDomain, scopeSystem string) ([]*Policy, error)
	GetActivePoliciesByScope(ctx context.Context, tenantID uuid.UUID, scope Scope) ([]*Policy, error)
	SupersedePolicy(ctx context.Context, oldID, newID uuid.UUID) error
	LogPolicyViolation(ctx context.Context, v *PolicyViolation) error
}

// Agent represents an AI agent or worker.
type Agent struct {
	ID            uuid.UUID       `json:"id"`
	TenantID      uuid.UUID       `json:"tenant_id"`
	Name          string          `json:"name"`
	ModelType     string          `json:"model_type"`
	SessionID     string          `json:"session_id"`
	Status        string          `json:"status"` // idle, working, transitioning, paused, offline
	CurrentTaskID *uuid.UUID      `json:"current_task_id,omitempty"`
	LastHeartbeat time.Time       `json:"last_heartbeat"`
	Metadata      json.RawMessage `json:"metadata"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// Task represents a unit of work.
type Task struct {
	ID           uuid.UUID   `json:"id"`
	TenantID     uuid.UUID   `json:"tenant_id"`
	Title        string      `json:"title"`
	Description  string      `json:"description,omitempty"`
	Status       TaskStatus  `json:"status"` // pending, in_progress, paused, completed, abandoned
	Priority     int         `json:"priority"`
	OwnerAgentID *uuid.UUID  `json:"owner_agent_id,omitempty"`
	AssignedTo   *uuid.UUID  `json:"assigned_to,omitempty"`
	ParentTaskID *uuid.UUID  `json:"parent_task_id,omitempty"`
	ScopeDomain  string      `json:"scope_domain"`
	ScopeSystem  string      `json:"scope_system"`
	RequiredRole AgentRole   `json:"required_role,omitempty"`
	Scope        string      `json:"scope,omitempty"`
	TokenBudget  int64       `json:"token_budget,omitempty"`
	TokensUsed   int64       `json:"tokens_used,omitempty"`
	TopologyID   uuid.UUID   `json:"topology_id,omitempty"`
	SequenceNo   int         `json:"sequence_no,omitempty"`
	DependsOn    []uuid.UUID `json:"depends_on,omitempty"`
	Version      int         `json:"version"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	CompletedAt  *time.Time  `json:"completed_at,omitempty"`
}

// LineageEdge represents an edge in the lineage DAG.
type LineageEdge struct {
	SourceTaskID uuid.UUID  `json:"source_task_id"`
	TargetTaskID uuid.UUID  `json:"target_task_id"`
	EdgeType     string     `json:"edge_type"` // handoff, depends_on, supersedes
	HandoffID    *uuid.UUID `json:"handoff_id,omitempty"`
	Depth        int        `json:"depth,omitempty"`
}

// AuditEvent represents a logged state change with Merkle chain linkage.
type AuditEvent struct {
	ID          uuid.UUID       `json:"id"`
	TenantID    uuid.UUID       `json:"tenant_id"`
	EventType   string          `json:"event_type"`
	EventID     uuid.UUID       `json:"event_id"`
	Actor       string          `json:"actor"`
	Payload     interface{}     `json:"payload"`
	EventHash   string          `json:"event_hash"`
	MerkleProof json.RawMessage `json:"merkle_proof,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

// AuditVerification returns the Merkle proof for an event.
type AuditVerification struct {
	EventID     uuid.UUID `json:"event_id"`
	EventHash   string    `json:"event_hash"`
	RootHash    string    `json:"root_hash"`
	BlockHeight int64     `json:"block_height"`
	IsVerified  bool      `json:"is_verified"`
}

// Milestone represents a project milestone.
type Milestone struct {
	ID          uuid.UUID  `json:"id"`
	TenantID    uuid.UUID  `json:"tenant_id"`
	TaskID      *uuid.UUID `json:"task_id,omitempty"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Status      string     `json:"status"` // pending, completed, superseded
	DueDate     *time.Time `json:"due_date,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// HandoffRecord represents a handoff between agents.
type HandoffRecord struct {
	ID            uuid.UUID  `json:"id"`
	TenantID      uuid.UUID  `json:"tenant_id"`
	TaskID        uuid.UUID  `json:"task_id"`
	SourceAgentID uuid.UUID  `json:"source_agent_id"`
	TargetAgentID uuid.UUID  `json:"target_agent_id"`
	Reason        string     `json:"reason"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
}

// PlanRequest defines the filters for a plan query.
type PlanRequest struct {
	ScopeDomain string     `json:"scope_domain"`
	ScopeSystem string     `json:"scope_system"`
	At          *time.Time `json:"at"`
	Statuses    []string   `json:"statuses"`
}

// PlanResult encapsulates the full plan for a scope.
type PlanResult struct {
	TenantID     uuid.UUID        `json:"tenant_id"`
	Scope        Scope            `json:"scope"`
	Decisions    []*Decision      `json:"decisions"`
	Tasks        []*Task          `json:"tasks"`
	Handoffs     []*HandoffRecord `json:"handoffs"`
	Milestones   []*Milestone     `json:"milestones"`
	Dependencies []LineageEdge    `json:"dependencies"`
	GeneratedAt  time.Time        `json:"generated_at"`
}

// DecisionID is a UUID for the decision identity.
type DecisionID = uuid.UUID

// DecisionContent is the immutable content (defined in canonical package).
// We keep a copy here for convenience.

// DecisionRevision represents an immutable revision.
type DecisionRevision struct {
	ID                   uuid.UUID `json:"id"`
	TenantID             uuid.UUID `json:"tenant_id"`
	DecisionID           uuid.UUID `json:"decision_id"`
	RevisionNumber       int       `json:"revision_number"`
	ContentHash          []byte    `json:"content_hash"` // SHA-256 of canonical content
	PreviousRevisionHash []byte    `json:"previous_revision_hash"`
	Actor                string    `json:"actor"`      // From auth context
	RequestID            string    `json:"request_id"` // For correlation
	CreatedAt            time.Time `json:"created_at"`
	// Metadata (not hashed)
}

// SubmitDecisionRequest is used for creating a new immutable revision.
type SubmitDecisionRequest struct {
	TenantID       uuid.UUID  `json:"tenant_id"`
	DecisionID     uuid.UUID  `json:"decision_id"`
	Title          string     `json:"title"`
	Statement      string     `json:"statement"`
	Scope          Scope      `json:"scope"`
	Owner          string     `json:"owner"`
	Confidence     float64    `json:"confidence"`
	Evidence       []Evidence `json:"evidence,omitempty"` // changed from []EvidenceHash
	ParentID       *uuid.UUID `json:"parent_id,omitempty"`
	IdempotencyKey string     `json:"idempotency_key,omitempty"`
}

// Actor is NOT in the request; it comes from auth context.
// IdempotencyKey allows safe retries.

// SubmitDecisionResult returned after atomic commit.
type SubmitDecisionResult struct {
	DecisionID     uuid.UUID `json:"decision_id"`
	RevisionID     uuid.UUID `json:"revision_id"`
	RevisionNumber int       `json:"revision_number"`
	ContentHash    []byte    `json:"content_hash"`
	MerkleRoot     []byte    `json:"merkle_root"`
}
