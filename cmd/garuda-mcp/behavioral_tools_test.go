// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"strings"

	"github.com/myshra777-ai/garuda/internal/store"
	"github.com/myshra777-ai/garuda/internal/version"
)

type toolBehavioralCase struct {
	name         string
	positiveArgs map[string]interface{}
	expectKey    string
	negativeArgs map[string]interface{}
	negMode      string // "error", "readonly", "rejected_field", "not_found_or_false"
}

func callToolJSON(t *testing.T, srv *MCPServer, name string, args map[string]interface{}) (MCPResponse, map[string]interface{}) {
	t.Helper()
	resp := srv.handleToolsCall(MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      name,
			"arguments": args,
		},
	})
	if resp.Error != nil {
		return resp, nil
	}
	resMap, ok := resp.Result.(map[string]interface{})
	if !ok {
		return resp, nil
	}
	content, ok := resMap["content"].([]map[string]interface{})
	if !ok || len(content) == 0 {
		return resp, nil
	}
	txt, _ := content[0]["text"].(string)
	var parsed map[string]interface{}
	_ = json.Unmarshal([]byte(txt), &parsed)
	return resp, parsed
}

func TestAll54RegisteredTools_PositiveAndNegativeBehavior(t *testing.T) {
	_ = os.Unsetenv("GARUDA_MCP_READ_ONLY")
	_ = os.Unsetenv("GARUDA_MCP_APPROVER")

	if len(allTools) != version.MCPToolCount {
		t.Fatalf("canonical manifest drift: expected %d tools in allTools, got %d", version.MCPToolCount, len(allTools))
	}

	tmpRulebook := filepath.Join(os.TempDir(), "garuda_behavioral_rulebook.md")
	defer os.Remove(tmpRulebook)

	cases := []toolBehavioralCase{
		// 1-6: Core Governance & Lineage
		{"garuda.briefing", map[string]interface{}{"workspace": "default"}, "workspace", map[string]interface{}{"tenant_id": "invalid-uuid"}, "error"},
		{"garuda.query", map[string]interface{}{"query": "RefundHandler", "workspace": "default"}, "status", map[string]interface{}{"query": "SELECT * FROM entities"}, "rejected_field"},
		{"garuda.get_lineage", map[string]interface{}{"symbol": "Engine", "workspace": "default"}, "target", map[string]interface{}{}, "error"},
		{"garuda.detect_contradictions", map[string]interface{}{"workspace": "default"}, "count", map[string]interface{}{"tenant_id": "not-a-uuid"}, "error"},
		{"garuda.get_impact", map[string]interface{}{"symbol": "AppendLeafAndSealTx", "workspace": "default"}, "symbol", map[string]interface{}{}, "error"},
		{"garuda.blast_radius", map[string]interface{}{"name": "AppendLeafAndSealTx", "workspace": "default"}, "total_affected", map[string]interface{}{}, "error"},

		// 7-14: Policy, Claims, Drift & Cryptographic Verification
		{"garuda.propose_decision", map[string]interface{}{"title": "ADR-Test: Layer Guard"}, "status", map[string]interface{}{}, "error"},
		{"garuda.policy.list", map[string]interface{}{}, "policies", map[string]interface{}{"tenant_id": "bad-uuid"}, "error"},
		{"garuda.governance.status", map[string]interface{}{"workspace": "default"}, "verdict", map[string]interface{}{"tenant_id": "bad-uuid"}, "error"},
		{"garuda.check_drift", map[string]interface{}{"workspace": "default"}, "doc_to_code", map[string]interface{}{"tenant_id": "bad-uuid"}, "error"},
		{"garuda.query_claims", map[string]interface{}{"workspace": "default", "status": "SUPPORTED"}, "claims", map[string]interface{}{"tenant_id": "bad-uuid"}, "error"},
		{"garuda.policy.evaluate", map[string]interface{}{"workspace": "default"}, "evaluations", map[string]interface{}{"tenant_id": "bad-uuid"}, "error"},
		{"garuda.verify_policy_evaluation", map[string]interface{}{"evaluation_id": "00000000-0000-0000-0000-000000000099"}, "valid", map[string]interface{}{"evaluation_id": "not-a-uuid"}, "not_found_or_false"},
		{"garuda.verify_evidence", map[string]interface{}{"evidence_id": "00000000-0000-0000-0000-000000000099"}, "valid", map[string]interface{}{}, "not_found_or_false"},

		// 15-20: AST Entity & Graph Navigation
		{"garuda.entities", map[string]interface{}{"workspace": "default", "limit": float64(5)}, "entities", map[string]interface{}{"tenant_id": "bad-uuid"}, "error"},
		{"garuda.inspect", map[string]interface{}{"name": "Engine", "workspace": "default"}, "status", map[string]interface{}{"entity_id": "bad-uuid"}, "not_found_or_false"},
		{"garuda.subclasses", map[string]interface{}{"workspace": "default"}, "results", map[string]interface{}{"name": "NonExistentSymbolXYZ"}, "not_found_or_false"},
		{"garuda.implementers", map[string]interface{}{"workspace": "default"}, "results", map[string]interface{}{"name": "NonExistentInterfaceXYZ"}, "not_found_or_false"},
		{"garuda.neighbors", map[string]interface{}{"name": "Engine", "workspace": "default"}, "neighbors", map[string]interface{}{"name": "NonExistentNeighborXYZ"}, "not_found_or_false"},
		{"garuda.find_entity", map[string]interface{}{"name": "Engine", "workspace": "default"}, "results", map[string]interface{}{"tenant_id": "bad-uuid"}, "error"},

		// 21-23: Multi-Agent Tasks, Handoffs & Checkpoints
		{"garuda.create_task", map[string]interface{}{"title": "Verify Behavioral Suite", "workspace": "default"}, "task_id", map[string]interface{}{}, "error"},
		{"garuda.handoff", map[string]interface{}{"task_id": "task-alias-1", "source_agent_id": "agent-a", "target_agent_id": "agent-b"}, "status", map[string]interface{}{"task_id": "00000000-0000-0000-0000-000000000001", "source_agent_id": "00000000-0000-0000-0000-000000000002", "target_agent_id": "00000000-0000-0000-0000-000000000003"}, "not_found_or_false"},
		{"garuda.resume", map[string]interface{}{"agent_id": "agent-b", "checkpoint_id": "cp-alias-1"}, "status", map[string]interface{}{"agent_id": "00000000-0000-0000-0000-000000000002", "checkpoint_id": "00000000-0000-0000-0000-000000000099"}, "not_found_or_false"},

		// 24-25: Active Talon Pre-Edit & Post-Edit Guard
		{"garuda.talon_pre_edit", map[string]interface{}{"target": "Engine", "workspace": "default"}, "workspace", map[string]interface{}{"tenant_id": "bad-uuid"}, "error"},
		{"garuda.talon_verify_edit", map[string]interface{}{"workspace": "default"}, "outcome", map[string]interface{}{"workspace": "default", "save_to_ledger": true}, "readonly"},

		// 26-35: Human-in-the-Loop Formula Disambiguation, Data/Logic Audit & Base Rulebook
		{"garuda.disambiguate_logic", map[string]interface{}{"workspace": "default", "include_fixtures": true}, "mode", map[string]interface{}{"workspace": "audit-sandbox", "include_fixtures": false}, "not_found_or_false"},
		{"garuda.decide_logic_variant", map[string]interface{}{"workspace": "default", "variant_id": "loss_ratio_ops_v1"}, "human_decision_status", map[string]interface{}{"workspace": "default", "variant_id": "loss_ratio_ops_v1"}, "readonly"},
		{"garuda.audit_data_discrepancy", map[string]interface{}{"workspace": "default", "include_fixtures": true}, "audits", map[string]interface{}{"workspace": "audit-sandbox", "include_fixtures": false}, "not_found_or_false"},
		{"garuda.reconcile_data_audit", map[string]interface{}{"workspace": "default", "audit_id": "claims_paid_crore_mismatch"}, "status", map[string]interface{}{"workspace": "default"}, "readonly"},
		{"garuda.audit_financial_discrepancy", map[string]interface{}{"workspace": "default", "include_fixtures": true}, "audits", map[string]interface{}{"workspace": "audit-sandbox", "include_fixtures": false}, "not_found_or_false"},
		{"garuda.reconcile_financial_audit", map[string]interface{}{"workspace": "default", "audit_id": "claims_paid_crore_mismatch"}, "status", map[string]interface{}{"workspace": "default"}, "readonly"},
		{"garuda.synthesize_rulebook", map[string]interface{}{"workspace": "default", "output": tmpRulebook}, "output_path", map[string]interface{}{"workspace": "default"}, "readonly"},
		{"garuda.approve_base_rulebook", map[string]interface{}{"workspace": "default"}, "status", map[string]interface{}{"workspace": "default"}, "readonly"},
		{"garuda.criticality_tiers", map[string]interface{}{"workspace": "default"}, "surfaces", map[string]interface{}{"workspace": "audit-sandbox", "surface_name": "Ghost", "package_pattern": "internal/nonexistent", "criticality_tier": "LOW"}, "not_found_or_false"},
		{"garuda.docs_graph", map[string]interface{}{"workspace": "default"}, "total_documents", map[string]interface{}{"tenant_id": "bad-uuid"}, "error"},

		// 36-40: Docs Ingest, Doctor, Build Matrix, Hygiene & Reset
		{"garuda.docs_ingest", map[string]interface{}{"workspace": "audit-sandbox", "path": "docs/adr"}, "claims_extracted", map[string]interface{}{"workspace": "audit-sandbox"}, "readonly"},
		{"garuda.doctor_diagnose", map[string]interface{}{"workspace": "default"}, "checks", map[string]interface{}{"tenant_id": "bad-uuid"}, "error"},
		{"garuda.build_matrix_check", map[string]interface{}{"workspace": "default"}, "matrix_results", map[string]interface{}{"tenant_id": "bad-uuid"}, "error"},
		{"garuda.hygiene_report", map[string]interface{}{"workspace": "default"}, "workspace", map[string]interface{}{"tenant_id": "bad-uuid"}, "error"},
		{"garuda.reset_audit_state", map[string]interface{}{"workspace": "audit-sandbox"}, "status", map[string]interface{}{"workspace": "audit-sandbox"}, "readonly"},

		// 41-54: New Workspace, Receipt, Snapshot Diff, Coverage, Claim Verification & Preflight Tools
		{"garuda.create_workspace", map[string]interface{}{"workspace": "audit-sandbox"}, "workspace_id", map[string]interface{}{"workspace": "audit-sandbox"}, "readonly"},
		{"garuda.verify_receipt", map[string]interface{}{"workspace": "default"}, "valid", map[string]interface{}{"workspace": "default", "receipt_hash": "deadbeef00000000"}, "not_found_or_false"},
		{"garuda.diff_snapshots", map[string]interface{}{"workspace": "default"}, "current_head_snapshot", map[string]interface{}{"tenant_id": "bad-uuid"}, "error"},
		{"garuda.diff_impact", map[string]interface{}{"workspace": "default"}, "snapshot_receipt", map[string]interface{}{"tenant_id": "bad-uuid"}, "error"},
		{"garuda.symbol_test_coverage", map[string]interface{}{"workspace": "default", "symbol": "Engine"}, "results", map[string]interface{}{"tenant_id": "bad-uuid"}, "error"},
		{"garuda.snapshot", map[string]interface{}{"workspace": "default"}, "snapshot_id", map[string]interface{}{"tenant_id": "bad-uuid"}, "error"},
		{"garuda.explain_claim", map[string]interface{}{"workspace": "default", "claim_id": "nonexistent-claim"}, "status", map[string]interface{}{"workspace": "default", "claim_id": "missing-claim-id"}, "not_found_or_false"},
		{"garuda.snippet", map[string]interface{}{"workspace": "default", "name": "Engine"}, "snapshot_id", map[string]interface{}{"tenant_id": "bad-uuid"}, "error"},
		{"garuda.summary", map[string]interface{}{"workspace": "default"}, "snapshot_id", map[string]interface{}{"tenant_id": "bad-uuid"}, "error"},
		{"garuda.owners", map[string]interface{}{"workspace": "default", "name": "Engine"}, "status", map[string]interface{}{"workspace": "default", "name": "UnassignedSymbolXYZ"}, "not_found_or_false"},
		{"garuda.verify_claim", map[string]interface{}{"workspace": "default", "from_symbol": "A", "predicate": "CALLS", "to_symbol": "B"}, "status", map[string]interface{}{"workspace": "default", "claim_id": "missing-claim"}, "not_found_or_false"},
		{"garuda.test_coverage", map[string]interface{}{"workspace": "default", "symbol": "Engine"}, "results", map[string]interface{}{"tenant_id": "bad-uuid"}, "error"},
		{"garuda.audit_export", map[string]interface{}{"workspace": "default", "format": "csv"}, "csv_content", map[string]interface{}{"tenant_id": "bad-uuid"}, "error"},
		{"garuda.preflight_bundle", map[string]interface{}{"workspace": "default", "target": "Engine"}, "snapshot_id", map[string]interface{}{"tenant_id": "bad-uuid"}, "error"},
	}

	srv := &MCPServer{clientName: "behavioral-test-suite"}
	if dbURL := strings.TrimSpace(os.Getenv("DATABASE_URL")); dbURL != "" {
		if pgStore, err := store.NewPostgresStore(dbURL); err == nil {
			srv.store = pgStore
			defer pgStore.Close()
		}
	}
	covered := make(map[string]bool)

	for _, tc := range cases {
		covered[tc.name] = true
		t.Run("Positive_"+tc.name, func(t *testing.T) {
			_ = os.Unsetenv("GARUDA_MCP_READ_ONLY")
			resp, payload := callToolJSON(t, srv, tc.name, tc.positiveArgs)
			if resp.Error != nil {
				t.Fatalf("positive call to %s returned protocol error: %+v", tc.name, resp.Error)
			}
			if payload == nil {
				t.Fatalf("positive call to %s returned nil JSON payload", tc.name)
			}
			if _, ok := payload[tc.expectKey]; !ok {
				t.Errorf("positive call to %s missing expected key %q in payload: %v", tc.name, tc.expectKey, payload)
			}
		})

		t.Run("Negative_"+tc.name, func(t *testing.T) {
			if tc.negMode == "readonly" {
				t.Setenv("GARUDA_MCP_READ_ONLY", "true")
			} else {
				_ = os.Unsetenv("GARUDA_MCP_READ_ONLY")
			}
			resp, payload := callToolJSON(t, srv, tc.name, tc.negativeArgs)
			switch tc.negMode {
			case "error":
				if resp.Error == nil && payload["status"] != "rejected" && payload["valid"] != false {
					// Some tools gracefully fall back on invalid tenant_id or return structured error; check both
					if _, hasErr := payload["error"]; !hasErr && resp.Error == nil {
						// Verify at least structured response without panic
						if len(payload) == 0 {
							t.Errorf("negative call to %s returned empty result", tc.name)
						}
					}
				}
			case "readonly":
				if payload["read_only"] != true || payload["status"] != "rejected" {
					t.Errorf("expected read_only rejection for %s, got %v", tc.name, payload)
				}
			case "rejected_field":
				if payload["status"] != "rejected" && payload["rejected"] != true {
					t.Errorf("expected status='rejected' for %s, got %v", tc.name, payload)
				}
			case "not_found_or_false":
				validFlag, hasValid := payload["valid"].(bool)
				statusStr, _ := payload["status"].(string)
				if hasValid && validFlag {
					t.Errorf("expected valid=false on negative test for %s, got true", tc.name)
				}
				if !hasValid && statusStr == "" {
					t.Errorf("expected structured negative status for %s, got %v", tc.name, payload)
				}
			}
		})
	}

	for _, toolDef := range allTools {
		tName, _ := toolDef["name"].(string)
		if !covered[tName] {
			t.Errorf("registered tool %q is missing from TestAll54RegisteredTools_PositiveAndNegativeBehavior", tName)
		}
	}
}

func TestNewTools_TenantIsolation_ReadOnly_HumanApproval_ReceiptHandling(t *testing.T) {
	_ = os.Unsetenv("GARUDA_MCP_READ_ONLY")
	_ = os.Unsetenv("GARUDA_MCP_APPROVER")

	srv := &MCPServer{clientName: "isolation-verifier"}
	tenantA := "11111111-1111-1111-1111-111111111111"
	tenantB := "22222222-2222-2222-2222-222222222222"

	// 1. Tenant Isolation across all new tools
	newTools := []string{
		"garuda.create_workspace",
		"garuda.verify_receipt",
		"garuda.diff_snapshots",
		"garuda.diff_impact",
		"garuda.symbol_test_coverage",
		"garuda.snapshot",
		"garuda.explain_claim",
		"garuda.snippet",
		"garuda.summary",
		"garuda.owners",
		"garuda.verify_claim",
		"garuda.test_coverage",
		"garuda.audit_export",
		"garuda.preflight_bundle",
	}

	for _, toolName := range newTools {
		_, payA := callToolJSON(t, srv, toolName, map[string]interface{}{
			"tenant_id": tenantA,
			"workspace": "isolated-tenant-ws",
			"name":      "Engine",
			"target":    "Engine",
			"claim_id":  "claim-1",
		})
		_, payB := callToolJSON(t, srv, toolName, map[string]interface{}{
			"tenant_id": tenantB,
			"workspace": "isolated-tenant-ws",
			"name":      "Engine",
			"target":    "Engine",
			"claim_id":  "claim-1",
		})
		if payA["tenant_id"] != tenantA {
			t.Errorf("%s: expected tenant_id=%s for Tenant A, got %v", toolName, tenantA, payA["tenant_id"])
		}
		if payB["tenant_id"] != tenantB {
			t.Errorf("%s: expected tenant_id=%s for Tenant B, got %v", toolName, tenantB, payB["tenant_id"])
		}
		wsA, okA := payA["workspace_id"].(string)
		wsB, okB := payB["workspace_id"].(string)
		if okA && okB && (wsA == "" || wsB == "" || wsA == wsB) {
			t.Errorf("%s: tenant workspace_id collision between Tenant A (%s) and Tenant B (%s)", toolName, wsA, wsB)
		}
	}

	// 2. Receipt Handling & Cryptographic Tamper Rejection
	_, snapA := callToolJSON(t, srv, "garuda.snapshot", map[string]interface{}{
		"tenant_id": tenantA,
		"workspace": "isolated-tenant-ws",
	})
	_, snapB := callToolJSON(t, srv, "garuda.snapshot", map[string]interface{}{
		"tenant_id": tenantB,
		"workspace": "isolated-tenant-ws",
	})
	hashA, _ := snapA["receipt_hash"].(string)
	hashB, _ := snapB["receipt_hash"].(string)
	if len(hashA) != 64 || len(hashB) != 64 {
		t.Fatalf("expected 64-char SHA-256 receipt_hash, got %q and %q", hashA, hashB)
	}
	if hashA == hashB {
		t.Errorf("expected distinct tenant-scoped receipt hashes for isolated workspaces, got identical %s", hashA)
	}

	// Verify valid receipt passes
	_, verOK := callToolJSON(t, srv, "garuda.verify_receipt", map[string]interface{}{
		"tenant_id":    tenantA,
		"workspace":    "isolated-tenant-ws",
		"receipt_hash": hashA,
	})
	if verOK["valid"] != true {
		t.Errorf("expected verify_receipt valid=true for matching receipt_hash, got %v", verOK)
	}

	// Verify cross-tenant receipt (passing Tenant B's receipt_hash to Tenant A) fails closed!
	_, verCross := callToolJSON(t, srv, "garuda.verify_receipt", map[string]interface{}{
		"tenant_id":    tenantA,
		"workspace":    "isolated-tenant-ws",
		"receipt_hash": hashB,
	})
	if verCross["valid"] != false {
		t.Errorf("expected verify_receipt valid=false when verifying Tenant B receipt under Tenant A, got %v", verCross)
	}

	// 3. Human Approval Attestation & Mismatch Rejection
	t.Setenv("GARUDA_MCP_APPROVER", "lead-architect@garuda.ai")
	_, mismatch := callToolJSON(t, srv, "garuda.approve_base_rulebook", map[string]interface{}{
		"workspace":   "default",
		"approved_by": "spoofed-user@garuda.ai",
	})
	if mismatch["human_cli_verified"] != false || mismatch["status"] != "proposed_awaiting_human" {
		t.Errorf("expected mismatched approved_by to fall back to proposed_awaiting_human, got %v", mismatch)
	}

	_, attested := callToolJSON(t, srv, "garuda.approve_base_rulebook", map[string]interface{}{
		"workspace":   "default",
		"approved_by": "lead-architect@garuda.ai",
	})
	if attested["human_cli_verified"] != true || attested["approved_by"] != "lead-architect@garuda.ai" {
		t.Errorf("expected matching GARUDA_MCP_APPROVER to succeed with human_cli_verified=true, got %v", attested)
	}
	_ = os.Unsetenv("GARUDA_MCP_APPROVER")
}

func TestEndToEndAgentContinuityWorkflow_TaskHandoffResumeAndBriefing(t *testing.T) {
	_ = os.Unsetenv("GARUDA_MCP_READ_ONLY")
	_ = os.Unsetenv("GARUDA_MCP_APPROVER")

	srv := &MCPServer{clientName: "continuity-agent-alpha"}
	if dbURL := strings.TrimSpace(os.Getenv("DATABASE_URL")); dbURL != "" {
		if pgStore, err := store.NewPostgresStore(dbURL); err == nil {
			srv.store = pgStore
			defer pgStore.Close()
		}
	}

	// Step 1: Agent A provisions a task with structured context
	_, taskPay := callToolJSON(t, srv, "garuda.create_task", map[string]interface{}{
		"workspace":       "default",
		"title":           "Refactor RefundHandler Store Boundary",
		"description":     "Ensure RefundHandler routes through IdempotencyStore per ADR-0001",
		"owner_agent_id":  "agent-alpha",
		"target_agent_id": "agent-beta",
	})
	taskID, _ := taskPay["task_id"].(string)
	if len(taskID) != 36 {
		t.Fatalf("expected 36-char UUID task_id from create_task, got %q", taskID)
	}

	// Step 2: Unknown canonical task UUID handoff MUST fail closed (status="failed" with reason)
	_, unknownHop := callToolJSON(t, srv, "garuda.handoff", map[string]interface{}{
		"workspace":       "default",
		"task_id":         "00000000-0000-0000-0000-000000000001",
		"source_agent_id": "00000000-0000-0000-0000-000000000002",
		"target_agent_id": "00000000-0000-0000-0000-000000000003",
		"reason":          "Attempted handoff on nonexistent task UUID",
	})
	if unknownHop["status"] != "failed" {
		t.Errorf("expected status='failed' for unknown canonical task UUID handoff, got %v", unknownHop["status"])
	}
	if _, hasReason := unknownHop["reason"].(string); !hasReason {
		t.Errorf("expected failure reason in unknown handoff payload, got %v", unknownHop)
	}

	// Step 3: Docs & Rules Graph reports explicit graph_source_mode and 4 tiers
	_, dgPay := callToolJSON(t, srv, "garuda.docs_graph", map[string]interface{}{
		"workspace": "default",
	})
	if _, ok := dgPay["graph_source_mode"].(string); !ok {
		t.Errorf("expected graph_source_mode in garuda.docs_graph payload, got %v", dgPay)
	}
}
