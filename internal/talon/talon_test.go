// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package talon

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderContractAndInstallHooks(t *testing.T) {
	tmp := t.TempDir()
	// Create a fake .git/hooks directory to verify git hooks installation
	if err := os.MkdirAll(filepath.Join(tmp, ".git", "hooks"), 0755); err != nil {
		t.Fatal(err)
	}

	contract := &WorkspaceContract{
		WorkspaceName: "payments-core",
		WorkspaceID:   "00000000-0000-0000-0000-000000000099",
		CompiledAt:    "2026-10-25T12:00:00Z",
		EntityCount:   142,
		ClaimCount:    380,
		Policies: []PolicyGuardrail{
			{
				ID:         "no-handler-to-db",
				Title:      "HTTP Handlers Must Not Call DB Layer Directly",
				Decision:   "BLOCK",
				Reason:     "Route all persistence through service layer",
				Predicates: []string{"layer_violation"},
			},
		},
		ADRGuardrails: []ADRGuardrail{
			{
				Document:  "ADR-004.md",
				Subject:   "CheckoutHandler",
				Modality:  "MUST_NOT",
				Predicate: "CALLS",
				Object:    "OrderRepository",
				Status:    "SUPPORTED",
			},
		},
		TopHubs: []HubSummary{
			{
				ID:           "hub-1",
				Name:         "OrderService",
				Kind:         "struct",
				Package:      "service",
				FilePath:     "internal/service/order.go",
				InboundEdges: 48,
			},
		},
	}
	contract.Markdown = RenderContractMarkdown(contract)

	if !strings.Contains(contract.Markdown, "no-handler-to-db") {
		t.Errorf("expected policy ID in contract markdown")
	}
	if !strings.Contains(contract.Markdown, "CheckoutHandler") {
		t.Errorf("expected ADR subject in contract markdown")
	}

	summary, err := InstallHooks(
		tmp,
		"/usr/local/bin/garuda",
		"/usr/local/bin/garuda-mcp",
		"postgres://test:test@localhost:5433/garuda_test?sslmode=disable",
		"payments-core",
		contract,
	)
	if err != nil {
		t.Fatalf("InstallHooks failed: %v", err)
	}

	for _, path := range []string{
		summary.ContractPath,
		summary.CursorRulePath,
		summary.CursorMCPPath,
		summary.ClaudeMDPath,
		summary.ClaudeSettings,
		summary.AgentsMDPath,
		summary.WindsurfRulesPath,
		summary.GitPreCommitPath,
		summary.GitCommitMsgPath,
	} {
		if path == "" {
			t.Errorf("expected non-empty installed path in summary: %+v", summary)
			continue
		}
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected file %s to exist: %v", path, err)
		}
	}

	verdict := &GuardVerdict{
		Workspace:              "payments-core",
		Outcome:                "BLOCK",
		NetNewViolationsCount:  1,
		ADRContradictionsCount: 1,
		TouchedSymbols:         []string{"api.CheckoutHandler (handler.go:42)"},
		Violations:             []string{"[no-handler-to-db:BLOCK] layer violation (CALLS): api.CheckoutHandler -> store.OrderRepository"},
		SelfRepairInstructions: []string{"Route CheckoutHandler through OrderService instead of calling store.OrderRepository directly."},
	}
	prompt := BuildSelfRepairPrompt(verdict)
	if !strings.Contains(prompt, "GARUDA TALON ENFORCEMENT GATE: CHANGE BLOCKED") {
		t.Errorf("unexpected self-repair prompt header: %s", prompt)
	}
}
