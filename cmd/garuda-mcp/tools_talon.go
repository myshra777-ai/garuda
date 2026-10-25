// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/myshra777-ai/garuda/internal/diff"
	"github.com/myshra777-ai/garuda/internal/talon"
)

// handlePromptsList exposes the active Garuda Talon architectural contract prompt
// so MCP hosts (Cursor, Claude Desktop, Claude Code) automatically inject it.
func (s *MCPServer) handlePromptsList(req MCPRequest) MCPResponse {
	return MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"prompts": []map[string]interface{}{
				{
					"name":        "garuda.architectural_contract",
					"description": "Mandatory Garuda Talon architectural contract, ADR rules, and active policy guardrails for this workspace.",
					"arguments": []map[string]interface{}{
						{
							"name":        "workspace",
							"description": "Optional workspace name",
							"required":    false,
						},
						{
							"name":        "target",
							"description": "Optional target file or symbol being edited",
							"required":    false,
						},
					},
				},
			},
		},
	}
}

func (s *MCPServer) handlePromptsGet(req MCPRequest) MCPResponse {
	args, _ := req.Params["arguments"].(map[string]interface{})
	if args == nil {
		args = map[string]interface{}{}
	}
	tenantID, workspaceID, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		return s.errorResponse(req.ID, -32000, err.Error())
	}
	wsName, _ := args["workspace"].(string)
	if wsName == "" {
		wsName = os.Getenv("GARUDA_WORKSPACE")
	}
	if wsName == "" {
		wsName = workspaceID.String()
	}

	policyDir := os.Getenv("GARUDA_POLICY_DIR")
	if policyDir == "" {
		policyDir = "./policies"
	}

	contract, err := talon.CompileContract(context.Background(), s.store.Pool(), tenantID, workspaceID, wsName, policyDir)
	if err != nil {
		return s.errorResponse(req.ID, -32000, err.Error())
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"description": "Garuda Talon Active Architectural & ADR Contract for workspace " + wsName,
			"messages": []map[string]interface{}{
				{
					"role": "user",
					"content": map[string]interface{}{
						"type": "text",
						"text": contract.Markdown,
					},
				},
			},
		},
	}
}

// handleResourcesList exposes the live workspace contract and ADR guardrails as MCP resources.
func (s *MCPServer) handleResourcesList(req MCPRequest) MCPResponse {
	return MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"resources": []map[string]interface{}{
				{
					"uri":         "garuda://workspace/contract",
					"name":        "Garuda Talon Active Architectural Contract",
					"description": "Compiled Markdown contract of all active YAML policies, ADR document_claims, and high-blast-radius hubs.",
					"mimeType":    "text/markdown",
				},
				{
					"uri":         "garuda://workspace/adr-guardrails",
					"name":        "Garuda Workspace ADR & Policy Guardrails (JSON)",
					"description": "Machine-readable JSON of all active ADR MUST/MUST_NOT claims and declarative policies.",
					"mimeType":    "application/json",
				},
			},
		},
	}
}

func (s *MCPServer) handleResourcesRead(req MCPRequest) MCPResponse {
	uri, _ := req.Params["uri"].(string)
	tenantID, workspaceID, err := s.resolveTenantAndWorkspace(map[string]interface{}{})
	if err != nil {
		return s.errorResponse(req.ID, -32000, err.Error())
	}
	wsName := os.Getenv("GARUDA_WORKSPACE")
	if wsName == "" {
		wsName = workspaceID.String()
	}
	policyDir := os.Getenv("GARUDA_POLICY_DIR")
	if policyDir == "" {
		policyDir = "./policies"
	}

	contract, err := talon.CompileContract(context.Background(), s.store.Pool(), tenantID, workspaceID, wsName, policyDir)
	if err != nil {
		return s.errorResponse(req.ID, -32000, err.Error())
	}

	mimeType := "text/markdown"
	text := contract.Markdown
	if strings.Contains(uri, "adr-guardrails") {
		mimeType = "application/json"
		text = mustJSON(contract)
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"contents": []map[string]interface{}{
				{
					"uri":      uri,
					"mimeType": mimeType,
					"text":     text,
				},
			},
		},
	}
}

// handleTalonPreEdit is the active pre-edit MCP tool (`garuda.talon_pre_edit`).
// It marks the session as oriented and returns the target's blast radius,
// callers, ADR rules, and policy constraints before the agent modifies code.
func (s *MCPServer) handleTalonPreEdit(args map[string]interface{}) (interface{}, error) {
	tenantID, workspaceID, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		return nil, err
	}
	wsName, _ := args["workspace"].(string)
	if wsName == "" {
		wsName = os.Getenv("GARUDA_WORKSPACE")
	}
	if wsName == "" {
		wsName = workspaceID.String()
	}
	target, _ := args["target"].(string)
	if target == "" {
		target, _ = args["symbol"].(string)
	}
	if target == "" {
		target, _ = args["file"].(string)
	}

	policyDir := os.Getenv("GARUDA_POLICY_DIR")
	if policyDir == "" {
		policyDir = "./policies"
	}

	s.markOriented()

	brief, err := talon.BuildPreEditBriefing(
		context.Background(),
		s.store,
		tenantID,
		workspaceID,
		wsName,
		target,
		policyDir,
	)
	if err != nil {
		return nil, fmt.Errorf("build pre-edit briefing: %w", err)
	}
	return brief, nil
}

// handleTalonVerifyEdit is the active post-edit MCP tool (`garuda.talon_verify_edit`).
// It runs incremental AST re-indexing on dirty files (if available), executes
// diff-scoped policy & ADR evaluation, anchors the decision to the RFC 6962
// Merkle ledger, and returns a GuardVerdict with self-repair instructions on BLOCK.
func (s *MCPServer) handleTalonVerifyEdit(args map[string]interface{}) (interface{}, error) {
	tenantID, workspaceID, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		return nil, err
	}
	wsName, _ := args["workspace"].(string)
	if wsName == "" {
		wsName = os.Getenv("GARUDA_WORKSPACE")
	}
	if wsName == "" {
		wsName = workspaceID.String()
	}

	repoDir, _ := args["repo_dir"].(string)
	if repoDir == "" {
		repoDir = os.Getenv("GARUDA_REPO_ROOT")
	}
	if repoDir == "" {
		repoDir = "."
	}
	absRepo, _ := filepath.Abs(repoDir)

	staged, _ := args["staged"].(bool)
	diffRev, _ := args["diff"].(string)
	var explicitFiles []string
	if rawFiles, ok := args["files"].(string); ok && rawFiles != "" {
		for _, f := range strings.Split(rawFiles, ",") {
			if trimmed := strings.TrimSpace(f); trimmed != "" {
				explicitFiles = append(explicitFiles, trimmed)
			}
		}
	}

	// Best-effort incremental AST sync via sibling `garuda` binary when dirty files exist
	if _, deltas, dErr := diff.CollectGitDeltas(absRepo, staged, diffRev, explicitFiles); dErr == nil && len(deltas) > 0 {
		if self, exErr := os.Executable(); exErr == nil {
			garudaBin := filepath.Join(filepath.Dir(self), "garuda")
			if _, stErr := os.Stat(garudaBin); stErr == nil {
				_ = exec.Command(garudaBin, "analyze", absRepo, "--workspace", wsName, "--incremental", "--save").Run()
			}
		}
	}

	policyDir := os.Getenv("GARUDA_POLICY_DIR")
	if policyDir == "" {
		policyDir = filepath.Join(absRepo, "policies")
	}

	s.markOriented()

	verdict, err := talon.RunGuard(context.Background(), s.store, tenantID, workspaceID, talon.GuardOptions{
		RepoDir:       absRepo,
		WorkspaceName: wsName,
		PolicyDir:     policyDir,
		Staged:        staged,
		DiffRev:       diffRev,
		ExplicitFiles: explicitFiles,
		NetNewOnly:    true,
		SaveLedger:    true,
	})
	if err != nil {
		return nil, fmt.Errorf("talon verify edit failed: %w", err)
	}
	return verdict, nil
}

func (s *MCPServer) markOriented() {
	s.sessionMu.Lock()
	s.oriented = true
	s.sessionMu.Unlock()
}

// injectActiveContract attaches the workspace's active Talon enforcement contract
// summary into MCP tool responses whenever GARUDA_TALON_ENFORCE is enabled (default true),
// so AI agents are continuously bound by ADR rules and policies even if they skip briefing.
func (s *MCPServer) injectActiveContract(toolName string, result interface{}) interface{} {
	if os.Getenv("GARUDA_TALON_ENFORCE") == "false" || os.Getenv("GARUDA_TALON_ENFORCE") == "0" {
		return result
	}
	if toolName == "garuda.briefing" || toolName == "garuda.talon_pre_edit" {
		s.markOriented()
	}

	m, ok := result.(map[string]interface{})
	if !ok {
		return result
	}

	s.sessionMu.RLock()
	wasOriented := s.oriented
	s.sessionMu.RUnlock()

	if !wasOriented {
		// Auto-orient the agent on its very first MCP tool call by attaching active ADR & policy rules
		tenantID, workspaceID, err := s.resolveTenantAndWorkspace(map[string]interface{}{})
		if err == nil {
			wsName := os.Getenv("GARUDA_WORKSPACE")
			if wsName == "" {
				wsName = workspaceID.String()
			}
			policyDir := os.Getenv("GARUDA_POLICY_DIR")
			if policyDir == "" {
				policyDir = "./policies"
			}
			if contract, cErr := talon.CompileContract(context.Background(), s.store.Pool(), tenantID, workspaceID, wsName, policyDir); cErr == nil {
				m["talon_active_contract"] = map[string]interface{}{
					"enforcement":        "ACTIVE_FAIL_CLOSED",
					"mandatory_next_step": "Call garuda.talon_pre_edit before modifying files, and garuda.talon_verify_edit immediately after editing.",
					"adr_guardrails":     contract.ADRGuardrails,
					"active_policies":    contract.Policies,
				}
				s.markOriented()
				return m
			}
		}
	}

	m["talon_enforcement"] = "ACTIVE (call garuda.talon_verify_edit after any code edit)"
	return m
}
