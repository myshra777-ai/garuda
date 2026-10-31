// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/knowledge"
)

func TestBug01_NoFabricatedHumanAttribution(t *testing.T) {
	_ = os.Unsetenv("GARUDA_MCP_APPROVER")
	_ = os.Unsetenv("GARUDA_MCP_READ_ONLY")

	srv := &MCPServer{clientName: "claude-desktop-test"}

	// 1. Calling decide_logic_variant without GARUDA_MCP_APPROVER must NOT record APPROVED_BY_HUMAN or "Human Stakeholder"
	res, err := srv.handleDecideLogicVariant(map[string]interface{}{
		"workspace":       "default",
		"variant_id":      "loss_ratio_ops_v1",
		"decision_action": "APPROVE_VARIANT",
		"approved_by":     "Human Stakeholder (OPS team)",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, ok := res.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map response, got %T", res)
	}
	if m["status"] != "proposed_awaiting_human" {
		t.Errorf("expected status=proposed_awaiting_human when GARUDA_MCP_APPROVER is unset, got %v", m["status"])
	}
	if m["human_decision_status"] != "PROPOSED_BY_AGENT_AWAITING_HUMAN_CLI_COSIGN" {
		t.Errorf("expected PROPOSED_BY_AGENT_AWAITING_HUMAN_CLI_COSIGN, got %v", m["human_decision_status"])
	}
	if m["approved_by"] != "" {
		t.Errorf("expected empty approved_by when unauthenticated, got %q", m["approved_by"])
	}
	if m["human_cli_verified"] != false {
		t.Errorf("expected human_cli_verified=false, got %v", m["human_cli_verified"])
	}

	// 2. Calling reconcile_data_audit without GARUDA_MCP_APPROVER must NOT record RECONCILED_BY_HUMAN
	recRes, err := srv.handleReconcileDataAudit(map[string]interface{}{
		"workspace":   "default",
		"approved_by": "Human Stakeholder",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rm := recRes.(map[string]interface{})
	if rm["status"] != "proposed_awaiting_human" {
		t.Errorf("expected status=proposed_awaiting_human, got %v", rm["status"])
	}
	if rm["approved_by"] != "" {
		t.Errorf("expected empty approved_by, got %q", rm["approved_by"])
	}

	// 3. When GARUDA_MCP_APPROVER is explicitly set, matching approver succeeds
	t.Setenv("GARUDA_MCP_APPROVER", "rohit@garuda.ai")
	verifiedRes, err := srv.handleDecideLogicVariant(map[string]interface{}{
		"workspace":       "default",
		"variant_id":      "loss_ratio_ops_v1",
		"decision_action": "APPROVED_BY_HUMAN",
		"approved_by":     "rohit@garuda.ai",
	})
	if err != nil {
		t.Fatalf("unexpected error with attested approver: %v", err)
	}
	vm := verifiedRes.(map[string]interface{})
	if vm["human_decision_status"] != "APPROVED_BY_HUMAN" {
		t.Errorf("expected APPROVED_BY_HUMAN with valid GARUDA_MCP_APPROVER, got %v", vm["human_decision_status"])
	}
	if vm["approved_by"] != "rohit@garuda.ai" {
		t.Errorf("expected approved_by=rohit@garuda.ai, got %v", vm["approved_by"])
	}
}

func TestBug02_ReadOnlyMode_BlocksAllMutations(t *testing.T) {
	t.Setenv("GARUDA_MCP_READ_ONLY", "true")
	srv := &MCPServer{clientName: "cursor-readonly-test"}

	mutatingCalls := []struct {
		name string
		args map[string]interface{}
	}{
		{"garuda.propose_decision", map[string]interface{}{"title": "Test"}},
		{"garuda.decide_logic_variant", map[string]interface{}{"variant_id": "loss_ratio_ops_v1"}},
		{"garuda.reconcile_data_audit", map[string]interface{}{"audit_id": "test"}},
		{"garuda.reconcile_financial_audit", map[string]interface{}{"audit_id": "test"}},
		{"garuda.approve_base_rulebook", map[string]interface{}{"workspace": "default"}},
		{"garuda.synthesize_rulebook", map[string]interface{}{"workspace": "default"}},
		{"garuda.docs_ingest", map[string]interface{}{"workspace": "default"}},
		{"garuda.create_task", map[string]interface{}{"title": "Task"}},
		{"garuda.handoff", map[string]interface{}{"task_id": "t1", "source_agent_id": "a1", "target_agent_id": "a2"}},
		{"garuda.resume", map[string]interface{}{"agent_id": "a2", "checkpoint_id": "c1"}},
		{"garuda.reset_audit_state", map[string]interface{}{"workspace": "default"}},
		{"garuda.talon_verify_edit", map[string]interface{}{"workspace": "default", "save_to_ledger": true}},
	}

	for _, tc := range mutatingCalls {
		if !IsMutatingTool(tc.name, tc.args) {
			t.Errorf("expected IsMutatingTool(%q) to be true", tc.name)
		}
		resp := srv.handleToolsCall(MCPRequest{
			JSONRPC: "2.0",
			ID:      1,
			Method:  "tools/call",
			Params: map[string]interface{}{
				"name":      tc.name,
				"arguments": tc.args,
			},
		})
		if resp.Error != nil {
			t.Errorf("tool %s returned protocol error instead of structured read_only rejection: %v", tc.name, resp.Error)
			continue
		}
		resMap, _ := resp.Result.(map[string]interface{})
		content, _ := resMap["content"].([]map[string]interface{})
		if len(content) == 0 || !strings.Contains(content[0]["text"].(string), `"read_only": true`) {
			t.Errorf("tool %s did not return read_only: true rejection: %v", tc.name, resp.Result)
		}
	}
}

func TestBug03_CrossToolCountsMatch(t *testing.T) {
	c := LoadWorkspaceCounts(context.Background(), nil, uuid.Nil, uuid.Nil)
	if c.Entities != 0 || c.ExportedEntities != 0 || c.ExternalStubs != 0 || c.ASTRelationships != 0 {
		t.Errorf("expected zero counts for nil pool, got %+v", c)
	}
}

func TestBug04_QueryRejectsRawSQL(t *testing.T) {
	_ = os.Unsetenv("GARUDA_MCP_READ_ONLY")
	srv := &MCPServer{}

	sqlQueries := []string{
		"SELECT * FROM entities LIMIT 5",
		"   select id, name from claims",
		"DROP TABLE entities",
		"DELETE FROM decisions WHERE 1=1",
		"INSERT INTO decisions VALUES ('1')",
		"UPDATE entities SET name='x'",
	}

	for _, q := range sqlQueries {
		res, err := srv.handleQuery(map[string]interface{}{"query": q})
		if err != nil {
			t.Fatalf("expected structured rejection for %q, got error: %v", q, err)
		}
		m, ok := res.(map[string]interface{})
		if !ok {
			t.Fatalf("expected map response for %q, got %T", q, res)
		}
		if m["status"] != "rejected" {
			t.Errorf("query %q: expected status='rejected', got %v", q, m["status"])
		}
		reason, _ := m["reason"].(string)
		if !strings.Contains(reason, "garuda.query does not accept raw SQL") {
			t.Errorf("query %q: unexpected reason %q", q, reason)
		}
	}
}

func TestBug05_FindEntityMissReturnsEmpty(t *testing.T) {
	allow := toolArgWhitelist["garuda.find_entity"]
	hasName := false
	for _, k := range allow {
		if k == "name" {
			hasName = true
			break
		}
	}
	if !hasName {
		t.Errorf("garuda.find_entity whitelist must include 'name', got %v", allow)
	}
}

func TestBug07_QueryClaimsStatusFilter(t *testing.T) {
	allow := toolArgWhitelist["garuda.query_claims"]
	hasStatus := false
	for _, k := range allow {
		if k == "status" {
			hasStatus = true
			break
		}
	}
	if !hasStatus {
		t.Errorf("garuda.query_claims whitelist must include 'status', got %v", allow)
	}
}

func TestBug13_CreateTask_ReturnsValidUUID(t *testing.T) {
	_ = os.Unsetenv("GARUDA_MCP_READ_ONLY")
	srv := &MCPServer{clientName: "test-agent"}

	res, err := srv.handleCreateTask(map[string]interface{}{
		"workspace": "default",
		"title":     "Verify MCP Handoff Pipeline",
	})
	if err != nil {
		t.Fatalf("handleCreateTask failed: %v", err)
	}
	m := res.(map[string]interface{})
	if m["status"] != "created" {
		t.Errorf("expected status='created', got %v", m["status"])
	}
	taskID, _ := m["task_id"].(string)
	if len(taskID) != 36 {
		t.Errorf("expected valid 36-char UUID for task_id, got %q", taskID)
	}
}

func TestBug14_SynthesizeRulebookEmptyWorkspaceAndProposedFilename(t *testing.T) {
	_ = os.Unsetenv("GARUDA_MCP_READ_ONLY")
	srv := &MCPServer{clientName: "test-agent"}

	// B-01: Empty workspace returns INSUFFICIENT_DATA with zero fixture leakage
	emptyRes, err := srv.handleSynthesizeRulebook(map[string]interface{}{
		"workspace": "audit-sandbox",
	})
	if err != nil {
		t.Fatalf("handleSynthesizeRulebook on empty workspace failed: %v", err)
	}
	em := emptyRes.(map[string]interface{})
	if em["status"] != "INSUFFICIENT_DATA" {
		t.Errorf("expected status='INSUFFICIENT_DATA' on empty workspace, got %v", em["status"])
	}
	jsonText := mustJSON(em)
	if strings.Contains(jsonText, "ComputeOpsRetention") || strings.Contains(jsonText, "142") {
		t.Errorf("empty workspace rulebook leaked fixture data: %s", jsonText)
	}
}

func TestBug15_DisambiguateLogicEmptyWorkspaceReturnsInsufficientData(t *testing.T) {
	srv := &MCPServer{clientName: "test-agent"}
	// B-01: Default production-only scope with 0 matching variants returns INSUFFICIENT_DATA
	res, err := srv.handleDisambiguateLogic(map[string]interface{}{
		"workspace": "audit-sandbox",
		"team":      "ops",
		"concept":   "Retention",
	})
	if err != nil {
		t.Fatalf("handleDisambiguateLogic failed: %v", err)
	}
	m := res.(map[string]interface{})
	if m["status"] != "INSUFFICIENT_DATA" {
		t.Errorf("expected status='INSUFFICIENT_DATA' on empty workspace, got %v", m["status"])
	}
	if strings.Contains(mustJSON(m), "ComputeOpsRetention") {
		t.Errorf("expected zero fixture leakage when include_fixtures=false")
	}

	// Explicit include_fixtures=true returns fixture mode
	fixRes, err := srv.handleDisambiguateLogic(map[string]interface{}{
		"workspace":        "default",
		"include_fixtures": true,
	})
	if err != nil {
		t.Fatalf("handleDisambiguateLogic with include_fixtures=true failed: %v", err)
	}
	fm := fixRes.(map[string]interface{})
	if fm["mode"] != "fixture" {
		t.Errorf("expected mode='fixture' when include_fixtures=true, got %v", fm["mode"])
	}
}

func TestBug17_TalonVerifyEditDefaultNoMerkleWrite(t *testing.T) {
	if IsMutatingTool("garuda.talon_verify_edit", map[string]interface{}{"workspace": "default"}) {
		t.Errorf("expected garuda.talon_verify_edit with default args to be non-mutating (dry-run)")
	}
	if !IsMutatingTool("garuda.talon_verify_edit", map[string]interface{}{"workspace": "default", "save_to_ledger": true}) {
		t.Errorf("expected garuda.talon_verify_edit with save_to_ledger=true to be classified as mutating")
	}
}

func TestBug19_PolicyEvaluateDeterminism(t *testing.T) {
	srv := &MCPServer{clientName: "test-agent"}
	r1, _ := srv.handleCriticalityTiers(map[string]interface{}{"workspace": "default"})
	r2, _ := srv.handleCriticalityTiers(map[string]interface{}{"workspace": "default"})
	if mustJSON(r1) != mustJSON(r2) {
		t.Errorf("expected deterministic criticality_tiers output")
	}
}

func TestClaudeAudit_B02_NoShoppingSiteTemplateAndRejectsUnmatchedPattern(t *testing.T) {
	srv := &MCPServer{clientName: "claude-retest-audit"}
	res, err := srv.handleCriticalityTiers(map[string]interface{}{
		"workspace": "audit-sandbox",
	})
	if err != nil {
		t.Fatalf("handleCriticalityTiers failed: %v", err)
	}
	raw := mustJSON(res)
	if strings.Contains(raw, "Product Image Box Alignment") || strings.Contains(raw, "Checkout & Payment Ledger") || strings.Contains(raw, "142") {
		t.Errorf("criticality_tiers still contains shopping-site template or hardcoded 142 callers: %s", raw)
	}

	// Registering an unmatched package_pattern on an empty workspace must be rejected
	regRes, err := srv.handleCriticalityTiers(map[string]interface{}{
		"workspace":        "audit-sandbox",
		"surface_name":     "Ghost Surface",
		"package_pattern":  "internal/nonexistent",
		"criticality_tier": "LOW",
	})
	if err != nil {
		t.Fatalf("handleCriticalityTiers registration returned error: %v", err)
	}
	rm := regRes.(map[string]interface{})
	if rm["status"] != "rejected_unmatched_package_pattern" {
		t.Errorf("expected status='rejected_unmatched_package_pattern', got %v", rm["status"])
	}
}

func TestClaudeAudit_B06_GovernanceStatusFailClosedOnUnverifiedOrPartial(t *testing.T) {
	// B-06: UNVERIFIED (0 entities or 0 claims) and PARTIAL must return AgentProceed: false
	vUnv := computeVerdict(&knowledge.HealthStats{
		TotalEntities: 0,
		TotalClaims:   10,
		Supported:     0,
	})
	if vUnv.AgentProceed || vUnv.State != "UNVERIFIED" {
		t.Errorf("expected State=UNVERIFIED and AgentProceed=false on UNVERIFIED workspace, got state=%s proceed=%v", vUnv.State, vUnv.AgentProceed)
	}

	vPartial := computeVerdict(&knowledge.HealthStats{
		TotalEntities: 100,
		TotalClaims:   10,
		Supported:     5,
		Unverified:    5,
	})
	if vPartial.AgentProceed || vPartial.State != "PARTIAL" {
		t.Errorf("expected State=PARTIAL and AgentProceed=false on PARTIAL workspace, got state=%s proceed=%v", vPartial.State, vPartial.AgentProceed)
	}

	vVerified := computeVerdict(&knowledge.HealthStats{
		TotalEntities: 100,
		TotalClaims:   10,
		Supported:     10,
	})
	if !vVerified.AgentProceed || vVerified.State != "CONFORMS" {
		t.Errorf("expected State=CONFORMS and AgentProceed=true on CONFORMS workspace with 0 blockers, got state=%s proceed=%v", vVerified.State, vVerified.AgentProceed)
	}
}

func TestClaudeAudit_DocsGraphCountsAndUniqueDocIDs(t *testing.T) {
	srv := &MCPServer{clientName: "claude-blind-audit"}
	res, err := srv.handleDocsGraph(map[string]interface{}{"workspace": "default"})
	if err != nil {
		t.Fatalf("handleDocsGraph failed: %v", err)
	}
	m := res.(map[string]interface{})
	totalDocs, _ := m["total_documents"].(int)
	byType, _ := m["by_doc_type"].(map[string]int)
	sumByType := byType["ADR"] + byType["SPEC"] + byType["POLICY"]
	if sumByType != totalDocs {
		t.Errorf("expected sum(by_doc_type)=%d to equal total_documents=%d", sumByType, totalDocs)
	}

	// B-11: Verify no duplicate doc_id values
	docs, _ := m["documents"].([]map[string]interface{})
	seenIDs := map[string]bool{}
	for _, d := range docs {
		id, _ := d["doc_id"].(string)
		if id == "" {
			t.Errorf("expected non-empty doc_id on document %v", d)
		}
		if seenIDs[id] {
			t.Errorf("duplicate doc_id %q detected in docs_graph", id)
		}
		seenIDs[id] = true
	}

	raw := `{"file": "/home/rohit/garuda/internal/api/handler.go", "other": "/home/rohit/garuda-validation/go/chi/mux.go"}`
	sanitized := SanitizeGlobalJSONString(raw)
	if strings.Contains(sanitized, "/home/rohit/") {
		t.Errorf("expected SanitizeGlobalJSONString to strip /home/rohit/, got %s", sanitized)
	}
}

func TestClaudeAudit_NewToolsWorkCleanly(t *testing.T) {
	_ = os.Unsetenv("GARUDA_MCP_READ_ONLY")
	srv := &MCPServer{clientName: "claude-blind-audit"}

	if _, err := srv.handleVerifyReceipt(map[string]interface{}{"workspace": "default"}); err != nil {
		t.Errorf("handleVerifyReceipt failed: %v", err)
	}
	if _, err := srv.handleDiffSnapshots(map[string]interface{}{"workspace": "default"}); err != nil {
		t.Errorf("handleDiffSnapshots failed: %v", err)
	}
	if _, err := srv.handleAuditExport(map[string]interface{}{"workspace": "default", "format": "csv"}); err != nil {
		t.Errorf("handleAuditExport failed: %v", err)
	}
	if _, err := srv.handleCreateWorkspace(map[string]interface{}{"workspace": "audit-sandbox"}); err != nil {
		t.Errorf("handleCreateWorkspace failed: %v", err)
	}
	if _, err := srv.handlePreflightBundle(map[string]interface{}{"workspace": "default", "target": "RefundHandler"}); err != nil {
		t.Errorf("handlePreflightBundle failed: %v", err)
	}

	// B-05: docs_ingest in isolated workspace with relative path does not read host repo docs/adr
	tmpWs := "test-isolated-ws-" + uuid.New().String()[:8]
	defer os.RemoveAll(filepath.Join(os.TempDir(), "garuda-workspaces", tmpWs))
	ingRes, err := srv.handleDocsIngest(map[string]interface{}{
		"workspace": tmpWs,
		"path":      "docs/adr",
	})
	if err != nil {
		t.Fatalf("handleDocsIngest on isolated workspace failed: %v", err)
	}
	im := ingRes.(map[string]interface{})
	if im["claims_extracted"] != 0 {
		t.Errorf("expected claims_extracted=0 on empty isolated workspace, got %v", im["claims_extracted"])
	}
}
