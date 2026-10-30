// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/myshra777-ai/garuda/internal/diff"
	"github.com/myshra777-ai/garuda/internal/talon"
)

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
	wsName := s.resolveMCPWorkspaceName(args, workspaceID)
	policyDir := resolveMCPPolicyDir("")

	contract, err := talon.CompileContract(context.Background(), s.store.Pool(), tenantID, workspaceID, wsName, policyDir)
	if err != nil {
		return s.errorResponse(req.ID, -32000, err.Error())
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"description":  "Garuda Talon Active Architectural & ADR Contract for workspace " + wsName,
			"workspace":    wsName,
			"workspace_id": workspaceID.String(),
			"tenant_id":    tenantID.String(),
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
	resArgs := map[string]interface{}{}
	if ws, ok := req.Params["workspace"].(string); ok && strings.TrimSpace(ws) != "" {
		resArgs["workspace"] = strings.TrimSpace(ws)
	}
	if parsedURL, uErr := url.Parse(uri); uErr == nil && parsedURL.RawQuery != "" {
		if qWs := strings.TrimSpace(parsedURL.Query().Get("workspace")); qWs != "" {
			resArgs["workspace"] = qWs
		}
		if qTid := strings.TrimSpace(parsedURL.Query().Get("tenant_id")); qTid != "" {
			resArgs["tenant_id"] = qTid
		}
	}

	tenantID, workspaceID, err := s.resolveTenantAndWorkspace(resArgs)
	if err != nil {
		return s.errorResponse(req.ID, -32000, err.Error())
	}
	wsName := s.resolveMCPWorkspaceName(resArgs, workspaceID)
	policyDir := resolveMCPPolicyDir("")

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
			"workspace":    wsName,
			"workspace_id": workspaceID.String(),
			"tenant_id":    tenantID.String(),
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

// handleTalonPreEdit is the active pre-edit MCP tool (`garuda.talon_pre_edit`) (BUG-16 & CLASS-10).
func (s *MCPServer) handleTalonPreEdit(args map[string]interface{}) (interface{}, error) {
	tenantID, workspaceID, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		return nil, err
	}
	wsName := s.resolveMCPWorkspaceName(args, workspaceID)
	target, _ := args["target"].(string)
	if target == "" {
		target, _ = args["symbol"].(string)
	}
	if target == "" {
		target, _ = args["file"].(string)
	}

	policyDir := resolveMCPPolicyDir("")

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
	for i := range brief.MatchedEntities {
		brief.MatchedEntities[i].FilePath = sanitizeMCPFilePath(brief.MatchedEntities[i].FilePath)
	}
	return brief, nil
}

// handleTalonVerifyEdit is the active post-edit MCP tool (`garuda.talon_verify_edit`) (BUG-17 & CLASS-10).
// Runs in dry-run mode (zero Merkle block writes) unless `save_to_ledger: true` or `save_ledger: true`
// is explicitly passed, and declares `side_effects` in its response.
func (s *MCPServer) handleTalonVerifyEdit(args map[string]interface{}) (interface{}, error) {
	saveLedger := false
	if explicitSave, ok := args["save_to_ledger"].(bool); ok {
		saveLedger = explicitSave
	} else if explicitSaveLegacy, ok := args["save_ledger"].(bool); ok {
		saveLedger = explicitSaveLegacy
	}

	if saveLedger && ReadOnlyMode() {
		return ReadOnlyRejection("garuda.talon_verify_edit"), nil
	}

	tenantID, workspaceID, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		return nil, err
	}
	wsName := s.resolveMCPWorkspaceName(args, workspaceID)

	repoDir, _ := args["repo_dir"].(string)
	if repoDir == "" || repoDir == "." {
		repoDir = resolveMCPRepoRoot()
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

	_, deltas, dErr := diff.CollectGitDeltas(absRepo, staged, diffRev, explicitFiles)
	hasDirtyDeltas := dErr == nil && len(deltas) > 0

	if hasDirtyDeltas && !ReadOnlyMode() {
		if self, exErr := os.Executable(); exErr == nil {
			garudaBin := filepath.Join(filepath.Dir(self), "garuda")
			if _, stErr := os.Stat(garudaBin); stErr == nil {
				_ = exec.Command(garudaBin, "analyze", absRepo, "--workspace", wsName, "--incremental", "--save").Run()
			}
		}
	}

	policyDir := resolveMCPPolicyDir(filepath.Join(absRepo, "policies"))
	s.markOriented()

	verdict, err := talon.RunGuard(context.Background(), s.store, tenantID, workspaceID, talon.GuardOptions{
		RepoDir:       absRepo,
		WorkspaceName: wsName,
		PolicyDir:     policyDir,
		Staged:        staged,
		DiffRev:       diffRev,
		ExplicitFiles: explicitFiles,
		NetNewOnly:    true,
		SaveLedger:    saveLedger,
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

// injectActiveContract tracks session orientation and attaches declarative
// Talon contract metadata on the initial call when GARUDA_TALON_ENFORCE=true
// (without imperative prompt-injection directives, satisfying both BUG-21 and mcp_verify.py).
func (s *MCPServer) injectActiveContract(toolName string, result interface{}) interface{} {
	if toolName == "garuda.briefing" || toolName == "garuda.talon_pre_edit" || toolName == "garuda.talon_verify_edit" {
		s.markOriented()
		return result
	}
	s.sessionMu.Lock()
	alreadyOriented := s.oriented
	s.oriented = true
	s.sessionMu.Unlock()

	if !alreadyOriented && strings.EqualFold(strings.TrimSpace(os.Getenv("GARUDA_TALON_ENFORCE")), "true") {
		if m, ok := result.(map[string]interface{}); ok {
			m["talon_enforcement"] = "enabled"
			m["talon_active_contract"] = ".garuda/TALON_CONTRACT.md"
		}
	}
	return result
}
