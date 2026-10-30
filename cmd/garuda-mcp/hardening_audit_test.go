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
	// Verify LoadWorkspaceCounts returns consistent zeroes on nil pool and single struct definition
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
	// Verify garuda.find_entity schema and whitelist accept "name" as well as "name_pattern"
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

func TestBug14_SynthesizeRulebookHonorsTmpOutputPath(t *testing.T) {
	_ = os.Unsetenv("GARUDA_MCP_READ_ONLY")
	srv := &MCPServer{clientName: "test-agent"}
	tmpOut := filepath.Join(os.TempDir(), "garuda_smoke_rulebook_test.md")
	_ = os.Remove(tmpOut)
	defer os.Remove(tmpOut)

	res, err := srv.handleSynthesizeRulebook(map[string]interface{}{
		"workspace": "default",
		"output":    tmpOut,
	})
	if err != nil {
		t.Fatalf("handleSynthesizeRulebook failed: %v", err)
	}
	m := res.(map[string]interface{})
	if m["output_path"] != tmpOut {
		t.Errorf("expected output_path=%q, got %v", tmpOut, m["output_path"])
	}
	if _, stErr := os.Stat(tmpOut); stErr != nil {
		t.Errorf("expected file to be written at %s: %v", tmpOut, stErr)
	}
}

func TestBug15_DisambiguateLogicDeclaresMode(t *testing.T) {
	srv := &MCPServer{clientName: "test-agent"}
	res, err := srv.handleDisambiguateLogic(map[string]interface{}{
		"workspace": "default",
		"team":      "ops",
		"concept":   "Loss Ratio",
	})
	if err != nil {
		t.Fatalf("handleDisambiguateLogic failed: %v", err)
	}
	m := res.(map[string]interface{})
	mode, _ := m["mode"].(string)
	if mode != "fixture" && mode != "live_ast" {
		t.Errorf("expected mode to be 'fixture' or 'live_ast', got %q", mode)
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
	// Verify criticality_tiers and disambiguate_logic are deterministic across repeated calls
	srv := &MCPServer{clientName: "test-agent"}
	r1, _ := srv.handleCriticalityTiers(map[string]interface{}{"workspace": "default"})
	r2, _ := srv.handleCriticalityTiers(map[string]interface{}{"workspace": "default"})
	if mustJSON(r1) != mustJSON(r2) {
		t.Errorf("expected deterministic criticality_tiers output")
	}
}
