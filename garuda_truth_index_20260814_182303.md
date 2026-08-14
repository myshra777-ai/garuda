# 🧠 Garuda Truth Index

**Generated:** Fri Aug 14 06:23:03 PM IST 2026
**Project:** /home/rohit/garuda
**Go Version:** go version go1.25.0 linux/amd64

## 1. Build & Test Status

- Build: ✅ PASS
- Tests (short): # github.com/myshra777-ai/garuda/cmd/garuda
cmd/garuda/main.go:573:27: fmt.Printf format %s has arg report.FingerprintDiff.Match of wrong type bool
❌ FAIL

## 2. Database Schema

✅ Database reachable at postgres://test:test@localhost:5433/garuda_test?sslmode=disable

### Tables
  -  agent_checkpoints
  -  agents
  -  assumptions
  -  audit_events
  -  budget_ledger
  -  checkpoints
  -  contradictions
  -  decision_revisions
  -  decisions
  -  evidence_blocks
  -  evidence_store
  -  facts
  -  handoffs
  -  harvested_decisions
  -  lineage_edges
  -  merkle_roots
  -  merkle_snapshots
  -  migrations
  -  milestones
  -  policies
  -  policy_violations
  -  repositories
  -  task_manifests
  -  tasks
  -  telemetry_events
  -  tenant_budgets
  -  topologies
  -  topology_audit
  -  topology_handoffs
  -  topology_tasks
  -  workspaces

### Table: `decisions`
```sql
                           Table "public.decisions"
      Column       |           Type           | Collation | Nullable | Default 
-------------------+--------------------------+-----------+----------+---------
 id                | uuid                     |           | not null | 
 tenant_id         | uuid                     |           | not null | 
 title             | text                     |           | not null | 
 statement         | text                     |           |          | 
 status            | text                     |           | not null | 
 domain            | text                     |           | not null | 
 system            | text                     |           | not null | 
 owner             | text                     |           | not null | 
 fingerprint       | text                     |           | not null | 
 created_at        | timestamp with time zone |           | not null | 
 updated_at        | timestamp with time zone |           | not null | 
 evidence_ids      | bytea[]                  |           |          | 
 temporal_metadata | jsonb                    |           |          | 
 scope             | jsonb                    |           |          | 
 approved_at       | timestamp with time zone |           |          | 
Indexes:
    "decisions_pkey" PRIMARY KEY, btree (tenant_id, id)
    "idx_decisions_tenant" btree (tenant_id)
Referenced by:
    TABLE "decision_revisions" CONSTRAINT "decision_revisions_tenant_id_decision_id_fkey" FOREIGN KEY (tenant_id, decision_id) REFERENCES decisions(tenant_id, id) ON DELETE RESTRICT

```

### Table: `decision_revisions`
```sql
                              Table "public.decision_revisions"
         Column         |           Type           | Collation | Nullable |      Default      
------------------------+--------------------------+-----------+----------+-------------------
 id                     | uuid                     |           | not null | gen_random_uuid()
 decision_id            | uuid                     |           | not null | 
 tenant_id              | uuid                     |           | not null | 
 revision_number        | integer                  |           | not null | 
 canonical_json         | jsonb                    |           | not null | 
 decision_hash          | bytea                    |           | not null | 
 previous_revision_hash | bytea                    |           | not null | 
 created_at             | timestamp with time zone |           | not null | 
Indexes:
    "decision_revisions_pkey" PRIMARY KEY, btree (id)
    "decision_revisions_tenant_id_decision_id_revision_number_key" UNIQUE CONSTRAINT, btree (tenant_id, decision_id, revision_number)
    "idx_revisions_created" btree (tenant_id, created_at DESC)
    "idx_revisions_decision" btree (tenant_id, decision_id, revision_number)
    "idx_revisions_hash" btree (decision_hash)
Foreign-key constraints:
    "decision_revisions_tenant_id_decision_id_fkey" FOREIGN KEY (tenant_id, decision_id) REFERENCES decisions(tenant_id, id) ON DELETE RESTRICT

```

### Table: `evidence_store`
```sql
                     Table "public.evidence_store"
   Column   |           Type           | Collation | Nullable | Default 
------------+--------------------------+-----------+----------+---------
 tenant_id  | uuid                     |           | not null | 
 block_hash | bytea                    |           | not null | 
 content    | jsonb                    |           | not null | 
 ref_count  | integer                  |           | not null | 1
 created_at | timestamp with time zone |           | not null | 
Indexes:
    "evidence_store_pkey" PRIMARY KEY, btree (tenant_id, block_hash)
    "idx_evidence_tenant" btree (tenant_id)

```

### Table: `merkle_roots`
```sql
                      Table "public.merkle_roots"
   Column   |           Type           | Collation | Nullable | Default 
------------+--------------------------+-----------+----------+---------
 tenant_id  | uuid                     |           | not null | 
 root_hash  | bytea                    |           | not null | 
 updated_at | timestamp with time zone |           | not null | 
Indexes:
    "merkle_roots_pkey" PRIMARY KEY, btree (tenant_id)

```

### Table: `audit_events`
```sql
                                       Table "public.audit_events"
   Column    |           Type           | Collation | Nullable |                 Default                  
-------------+--------------------------+-----------+----------+------------------------------------------
 id          | integer                  |           | not null | nextval('audit_events_id_seq'::regclass)
 decision_id | text                     |           | not null | 
 actor       | text                     |           | not null | 
 old_status  | text                     |           |          | 
 new_status  | text                     |           | not null | 
 reason      | text                     |           |          | 
 timestamp   | timestamp with time zone |           | not null | 
 signature   | bytea                    |           |          | 
Indexes:
    "audit_events_pkey" PRIMARY KEY, btree (id)
    "idx_audit_actor" btree (actor)
    "idx_audit_decision" btree (decision_id)
    "idx_audit_timestamp" btree ("timestamp")

```

### Table: `workspaces`
```sql
                        Table "public.workspaces"
   Column    |           Type           | Collation | Nullable | Default 
-------------+--------------------------+-----------+----------+---------
 id          | uuid                     |           | not null | 
 tenant_id   | uuid                     |           | not null | 
 name        | text                     |           | not null | 
 description | text                     |           |          | 
 created_at  | timestamp with time zone |           | not null | 
 updated_at  | timestamp with time zone |           | not null | 
Indexes:
    "workspaces_pkey" PRIMARY KEY, btree (id)
    "workspaces_tenant_id_name_key" UNIQUE CONSTRAINT, btree (tenant_id, name)
Referenced by:
    TABLE "repositories" CONSTRAINT "repositories_workspace_id_fkey" FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE

```

### Table: `repositories`
```sql
                             Table "public.repositories"
      Column      |           Type           | Collation | Nullable |     Default     
------------------+--------------------------+-----------+----------+-----------------
 id               | uuid                     |           | not null | 
 workspace_id     | uuid                     |           | not null | 
 provider         | text                     |           | not null | 
 url              | text                     |           | not null | 
 default_branch   | text                     |           | not null | 'main'::text
 language         | text                     |           |          | 
 current_commit   | text                     |           |          | 
 enabled          | boolean                  |           | not null | true
 analysis_status  | text                     |           | not null | 'pending'::text
 last_analyzed_at | timestamp with time zone |           |          | 
 created_at       | timestamp with time zone |           | not null | 
 updated_at       | timestamp with time zone |           | not null | 
Indexes:
    "repositories_pkey" PRIMARY KEY, btree (id)
    "idx_repositories_status" btree (analysis_status)
    "idx_repositories_workspace" btree (workspace_id)
    "repositories_workspace_id_url_key" UNIQUE CONSTRAINT, btree (workspace_id, url)
Foreign-key constraints:
    "repositories_workspace_id_fkey" FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE

```

### Table: `harvested_decisions`
```sql
                            Table "public.harvested_decisions"
       Column       |           Type           | Collation | Nullable |      Default      
--------------------+--------------------------+-----------+----------+-------------------
 id                 | uuid                     |           | not null | gen_random_uuid()
 tenant_id          | uuid                     |           | not null | 
 source_type        | text                     |           | not null | 
 source_id          | text                     |           | not null | 
 source_url         | text                     |           |          | 
 raw_text           | text                     |           | not null | 
 extracted_decision | text                     |           |          | 
 confidence         | double precision         |           |          | 0.7
 human_validated    | boolean                  |           |          | false
 decision_id        | uuid                     |           |          | 
 metadata           | jsonb                    |           |          | 
 created_at         | timestamp with time zone |           | not null | now()
 updated_at         | timestamp with time zone |           | not null | now()
Indexes:
    "harvested_decisions_pkey" PRIMARY KEY, btree (id)
    "idx_harvested_confidence" btree (tenant_id, confidence)
    "idx_harvested_created" btree (tenant_id, created_at DESC)
    "idx_harvested_decision" btree (tenant_id, decision_id)
    "idx_harvested_source" btree (tenant_id, source_type, source_id)
    "idx_harvested_validated" btree (tenant_id, human_validated)
Check constraints:
    "harvested_decisions_confidence_check" CHECK (confidence >= 0::double precision AND confidence <= 1::double precision)
    "harvested_decisions_source_type_check" CHECK (source_type = ANY (ARRAY['git_commit'::text, 'adr'::text, 'slack'::text, 'github_pr'::text, 'email'::text]))
Policies:
    POLICY "harvest_tenant_isolation"
      USING ((tenant_id = (current_setting('app.current_tenant_id'::text))::uuid))

```

### Table: `contradictions`
```sql
                             Table "public.contradictions"
       Column        |           Type           | Collation | Nullable |    Default    
---------------------+--------------------------+-----------+----------+---------------
 id                  | uuid                     |           | not null | 
 tenant_id           | uuid                     |           | not null | 
 decision_a          | uuid                     |           | not null | 
 decision_b          | uuid                     |           | not null | 
 severity            | text                     |           | not null | 
 resolved            | boolean                  |           | not null | false
 created_at          | timestamp with time zone |           | not null | now()
 resolved_at         | timestamp with time zone |           |          | 
 quarantined         | boolean                  |           |          | true
 resolution_strategy | text                     |           |          | 'human'::text
 auto_resolved_at    | timestamp with time zone |           |          | 
Indexes:
    "contradictions_pkey" PRIMARY KEY, btree (id)
    "idx_contradictions_resolved" btree (tenant_id, resolved)
    "idx_contradictions_tenant" btree (tenant_id)
    "idx_contradictions_unresolved" btree (tenant_id, resolved, quarantined)

```

### Migration Status
```sql
No migrations table found
```

## 3. Go Core Types

### Decision
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

### DecisionRevision
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

### Contradiction
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

### Evidence
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

### Scope
type Scope struct {
	Domain string `json:"domain"`
	System string `json:"system"`
	Team   string `json:"team,omitempty"`
	Env    string `json:"env,omitempty"`
	Region string `json:"region,omitempty"`
}

// DecisionStatus captures the lifecycle state of a governance decision.
type DecisionStatus string


### Analyze Result
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
    Fingerprint string    `json:"fingerprint"`
    Stats       Stats     `json:"stats"`
    PayloadHash string    `json:"payload_hash"`
    Provenance  Provenance `json:"provenance,omitempty"`
    Source string `json:"source,omitempty"`
}

// Provenance tracks where the analysis came from
type Provenance struct {
    WorkspaceID  string `json:"workspace_id,omitempty"`
    RepositoryID string `json:"repository_id,omitempty"`
    CommitSHA    string `json:"commit_sha,omitempty"`
    SourcePath   string `json:"source_path,omitempty"`
}

## 4. Store Methods

### Decision Store Methods
  - func (s *PostgresStore) SaveDecision(ctx context.Context, d *types.Decision) error {
  - func (s *PostgresStore) GetDecision(ctx context.Context, tenantID, decisionID uuid.UUID) (*types.Decision, error) {
  - func (s *PostgresStore) GetDecisionRevisions(ctx context.Context, tenantID, decisionID uuid.UUID) ([]types.DecisionRevision, error) {
  - func (s *PostgresStore) GetDecisionsByScope(ctx context.Context, tenantID uuid.UUID, domain, system string) ([]*types.Decision, error) {
  - func (s *PostgresStore) ListDecisionsByParent(ctx context.Context, tenantID, parentID uuid.UUID) ([]*types.Decision, error) {

### Revision Store Methods
  - func (s *PostgresStore) SubmitDecision(
  - func (s *PostgresStore) getDecisionByidempotencyKey(ctx context.Context, tenantID uuid.UUID, key string) (*types.SubmitDecisionResult, error) {

### Contradiction Store Methods
  - func (s *PostgresStore) QuarantineDecision(ctx context.Context, tenantID uuid.UUID,
  - func (s *PostgresStore) ListUnresolvedContradictions(ctx context.Context, tenantID uuid.UUID) ([]types.Contradiction, error) {
  - func (s *PostgresStore) ResolveContradiction(ctx context.Context, id uuid.UUID, strategy string) error {
  - func (s *PostgresStore) GetContradiction(ctx context.Context, tenantID, id uuid.UUID) (*types.Contradiction, error) {
  - func (s *PostgresStore) ListContradictions(ctx context.Context, tenantID uuid.UUID, resolved bool) ([]types.Contradiction, error) {

### Workspace Store Methods
  - func (s *PostgresStore) CreateWorkspace(ctx context.Context, tenantIDStr, name, description string) (*Workspace, error) {
  - func (s *PostgresStore) ListWorkspaces(ctx context.Context, tenantIDStr string) ([]*Workspace, error) {
  - func (s *PostgresStore) AddRepository(ctx context.Context, workspaceID uuid.UUID, provider, url, defaultBranch, language string) (*Repository, error) {
  - func (s *PostgresStore) ListRepositories(ctx context.Context, workspaceID uuid.UUID) ([]*Repository, error) {
  - func (s *PostgresStore) UpdateRepositorySyncStatus(ctx context.Context, tenantIDStr string, repoID uuid.UUID, commitSHA, status string) error {
  - func (s *PostgresStore) SyncWorkspace(ctx context.Context, workspaceID uuid.UUID, syncFunc func(repo *Repository, path string) error) error {

### Analysis Store Methods
  - func (s *PostgresStore) SaveAnalysisDecision(ctx context.Context, tenantIDStr string, result *analyzer.Result) (string, int, error) {

## 5. API Handlers

### Registered Routes

### Handler Functions
  - internal/api/audit_handlers.go:func (s *Server) HandleExportAuditLogs(w http.ResponseWriter, r *http.Request) {
  - internal/api/auth.go:func (s *Server) HandleSignUp(w http.ResponseWriter, r *http.Request) {
  - internal/api/auth.go:func (s *Server) HandleSignIn(w http.ResponseWriter, r *http.Request) {
  - internal/api/auth.go:func (s *Server) HandleSignOut(w http.ResponseWriter, r *http.Request) {
  - internal/api/budget_handlers.go:func (s *Server) HandleGetBudget(w http.ResponseWriter, r *http.Request) {
  - internal/api/budget_handlers.go:func (s *Server) HandleConsumeBudget(w http.ResponseWriter, r *http.Request) {
  - internal/api/checkpoint_handlers.go:func (s *Server) HandleAgentCheckpoint(w http.ResponseWriter, r *http.Request) {
  - internal/api/checkpoint_handlers.go:func (s *Server) HandleGetAgentCheckpoint(w http.ResponseWriter, r *http.Request) {
  - internal/api/checkpoint_handlers.go:func (s *Server) HandleAgentResume(w http.ResponseWriter, r *http.Request) {
  - internal/api/checkpoint_handlers.go:func (s *Server) HandleAgentHandoff(w http.ResponseWriter, r *http.Request) {
  - internal/api/dashboard_handlers.go:func (s *Server) HandleDashboard(w http.ResponseWriter, r *http.Request) {
  - internal/api/dashboard_handlers.go:func (s *Server) HandleDashboardStats(w http.ResponseWriter, r *http.Request) {
  - internal/api/dashboard_handlers.go:func (s *Server) HandleLiveEvents(w http.ResponseWriter, r *http.Request) {
  - internal/api/decision_handlers.go:func (s *Server) HandleProposeDecision(w http.ResponseWriter, r *http.Request) {
  - internal/api/decision_handlers.go:func (s *Server) HandleDecisionLineage(w http.ResponseWriter, r *http.Request) {
  - internal/api/decision_handlers.go:func (s *Server) HandleSystemPromptContext(w http.ResponseWriter, r *http.Request) {
  - internal/api/decision_handlers.go:func (s *Server) HandleEvaluateRoute(w http.ResponseWriter, r *http.Request) {
  - internal/api/decision_handlers.go:func (s *Server) HandleAgentWarmup(w http.ResponseWriter, r *http.Request) {
  - internal/api/decision_handlers.go:func (s *Server) HandleAuditVerify(w http.ResponseWriter, r *http.Request) {
  - internal/api/handler.go:func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
  - internal/api/handler.go:func (s *Server) HandleDebugToken(w http.ResponseWriter, r *http.Request) {
  - internal/api/handler.go:func (s *Server) HandleVerifyAuditLog(w http.ResponseWriter, r *http.Request) {
  - internal/api/handoff_handlers.go:func (s *Server) HandleHandoff(w http.ResponseWriter, r *http.Request) {
  - internal/api/handoff_handlers.go:func (s *Server) HandleResume(w http.ResponseWriter, r *http.Request) {
  - internal/api/handoff_handlers.go:func (s *Server) HandleGetLineage(w http.ResponseWriter, r *http.Request) {
  - internal/api/merkle_handlers.go:func (s *Server) HandleListMerkleSnapshots(w http.ResponseWriter, r *http.Request) {
  - internal/api/plan_handlers.go:func (s *Server) HandleGetPlan(w http.ResponseWriter, r *http.Request) {
  - internal/api/policy_handlers.go:func (s *Server) HandleRememberPolicy(w http.ResponseWriter, r *http.Request) {
  - internal/api/policy_handlers.go:func (s *Server) HandleListPolicies(w http.ResponseWriter, r *http.Request) {
  - internal/api/policy_handlers.go:func (s *Server) HandleSupersedePolicy(w http.ResponseWriter, r *http.Request) {
  - internal/api/swagger.go:func (s *Server) HandleOpenAPISpec(w http.ResponseWriter, r *http.Request) {
  - internal/api/swagger_ui.go:func (s *Server) HandleSwaggerUI(w http.ResponseWriter, r *http.Request) {
  - internal/api/system_handlers.go:func (s *Server) HandleSystemDiscover(w http.ResponseWriter, r *http.Request) {
  - internal/api/system_handlers.go:func (s *Server) HandleSystemBootstrap(w http.ResponseWriter, r *http.Request) {
  - internal/api/system_handlers.go:func (s *Server) HandleSandbox(w http.ResponseWriter, r *http.Request) {
  - internal/api/temporal_handlers.go:func (s *Server) HandleDecisionsActiveAt(w http.ResponseWriter, r *http.Request) {
  - internal/api/temporal_handlers.go:func (s *Server) HandleDecisionHistory(w http.ResponseWriter, r *http.Request) {
  - internal/api/topology_handlers.go:func (s *Server) HandleTopologyRecommend(w http.ResponseWriter, r *http.Request) {
  - internal/api/topology_handlers.go:func (s *Server) HandleTopologyExecute(w http.ResponseWriter, r *http.Request) {
  - internal/api/topology_handlers.go:func (s *Server) HandleTopologyStatus(w http.ResponseWriter, r *http.Request) {
  - internal/api/verification_handlers.go:func (s *Server) HandleVerifyDecision(w http.ResponseWriter, r *http.Request) {

## 6. CLI Commands

### Root Commands
  - 	rootCmd.AddCommand(initCmd)
  - 	rootCmd.AddCommand(upCmd)
  - 	rootCmd.AddCommand(downCmd)
  - 	rootCmd.AddCommand(statusCmd)
  - 	rootCmd.AddCommand(proposeCmd)
  - 	rootCmd.AddCommand(verifyCmd)
  - 	rootCmd.AddCommand(explainCmd)
  - 	rootCmd.AddCommand(mcpCmd)
  - 	rootCmd.AddCommand(dashboardCmd)
  - 	rootCmd.AddCommand(ingestCmd)
  - 	rootCmd.AddCommand(analyzeCmd)
  - 	rootCmd.AddCommand(diffCmd)
  - 	rootCmd.AddCommand(workspaceCmd)
  - 	rootCmd.AddCommand(repoCmd)
  - 	rootCmd.AddCommand(inspectCmd)

### Command Definitions
  - var rootCmd = &cobra.Command{
  - var analyzeCmd = &cobra.Command{
  - var diffCmd = &cobra.Command{
  - var workspaceCmd = &cobra.Command{
  - var workspaceCreateCmd = &cobra.Command{
  - var workspaceListCmd = &cobra.Command{
  - var workspaceDeleteCmd = &cobra.Command{
  - var workspaceSyncCmd = &cobra.Command{
  - var repoCmd = &cobra.Command{
  - var repoAddCmd = &cobra.Command{
  - var repoListCmd = &cobra.Command{
  - var repoRemoveCmd = &cobra.Command{
  - var repoEnableCmd = &cobra.Command{
  - var repoDisableCmd = &cobra.Command{
  - var initCmd = &cobra.Command{
  - var upCmd = &cobra.Command{
  - var downCmd = &cobra.Command{
  - var statusCmd = &cobra.Command{
  - var proposeCmd = &cobra.Command{
  - var verifyCmd = &cobra.Command{
  - var explainCmd = &cobra.Command{
  - var mcpCmd = &cobra.Command{
  - var mcpInstallCmd = &cobra.Command{
  - var ingestCmd = &cobra.Command{
  - var dashboardCmd = &cobra.Command{
  - var inspectCmd = &cobra.Command{

## 7. Migration Files

### Applied Migrations
  - migrations/000_create_extensions.sql
  - migrations/001_cas_blocks.sql
  - migrations/001_create_decisions_table.sql
  - migrations/002_create_audit_events_table.sql
  - migrations/002_task_manifests.sql
  - migrations/003_create_users_table.sql
  - migrations/003_rename_cas_blocks_to_evidence.sql
  - migrations/004_facts_assumptions.sql
  - migrations/005_contradictions.sql
  - migrations/005_decision_revisions.sql
  - migrations/007_add_tenant_id_to_decisions.sql
  - migrations/007_tenant_composite_key.sql
  - migrations/008_contradictions_schema.sql
  - migrations/009_agent_checkpoints.sql
  - migrations/010_tenant_budgets.sql
  - migrations/011_contradiction_quarantine.sql
  - migrations/012_merkle_verification.sql
  - migrations/013_add_scope_columns.sql
  - migrations/013_bitemporal_validity.sql
  - migrations/014_merkle_snapshots.sql
  - migrations/015_bitemporal_validity.sql
  - migrations/016_merkle_snapshots.sql
  - migrations/017_agent_handoff_lineage.sql
  - migrations/019_telemetry.sql
  - migrations/020_plan_support.sql
  - migrations/021_policy_engine.sql
  - migrations/022_topology.sql
  - migrations/023_harvested_decisions.sql
  - migrations/024_immutable_revisions.sql
  - migrations/025_fix_foreign_keys.sql
  - migrations/026_workspaces.sql
  - migrations/027_hardening.sql
  - migrations/028_semantic_core.sql
  - migrations/029_fix_merkle_root_text.sql
  - migrations/030_add_tenant_id_to_repositories.sql
  - migrations/030_fix_schema_gaps.sql

### Migration Status (from DB)
  - Migration table not found

## 8. Key Findings

### Critical Schema Checks
  - decisions.tenant_id: ✅ EXISTS
  - decision_revisions.decision_hash: ✅ EXISTS
  - evidence_store.tenant_id: ✅ EXISTS
  - merkle_roots table: ❌ MISSING

## 9. Git Status

  - 165f091 docs: add truth index snapshot
  - 2768c28 fix: migration 027 idempotent & merkle root scan to []byte
  - 8d9872e feat: complete Phase 1 – hardened foundation with multi-tenant evidence store, workspace/repo management, and cryptographic audit chain
  - 35cadf1 decision: use PostgreSQL for production
  - fb05419 fix: remove duplicate topology route

---
✅ Truth Index generated: garuda_truth_index_20260814_182303.md

📌 Next: Review this file to identify schema/code mismatches.
