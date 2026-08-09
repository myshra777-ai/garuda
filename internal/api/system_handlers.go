package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

// HandleSystemDiscover returns a machine‑readable index of all endpoints.
func (s *Server) HandleSystemDiscover(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	modules := []map[string]interface{}{
		{
			"name":        "System",
			"description": "Health, discovery, bootstrap, and sandbox.",
			"endpoints": []string{
				"GET /system/health",
				"GET /system/discover",
				"GET /system/bootstrap",
				"POST /sandbox",
			},
		},
		{
			"name":        "Decisions",
			"description": "Propose, verify, and trace decisions.",
			"endpoints": []string{
				"POST /decisions",
				"GET /decisions/{id}/lineage",
			},
		},
		{
			"name":        "Agents",
			"description": "Multi‑agent handoff, resume, and checkpointing.",
			"endpoints": []string{
				"POST /agents/handoff",
				"POST /agents/resume",
			},
		},
		{
			"name":        "Audit",
			"description": "Cryptographic audit and Merkle proof verification.",
			"endpoints": []string{
				"GET /audit/verify",
				"GET /audit/export",
			},
		},
		{
			"name":        "Budget",
			"description": "Token budget and metering.",
			"endpoints": []string{
				"GET /budget",
				"POST /budget/consume",
			},
		},
		{
			"name":        "Router",
			"description": "Pre‑flight model classification and routing.",
			"endpoints": []string{
				"POST /router/evaluate",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"version": "v1.0.0",
		"modules": modules,
	})
}

// HandleSystemBootstrap returns a single‑turn bootstrap manifest for agents.
func (s *Server) HandleSystemBootstrap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	tenantID, err := resolveTenantID(r, uuid.Nil)
	if err != nil || tenantID == uuid.Nil {
		tenantID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	}

	// Get Merkle root
	root, _ := s.store.GetLatestMerkleSnapshot(r.Context(), tenantID)
	rootHash := "genesis_root_00000000000000000000000000000000"
	blockHeight := int64(0)
	if root != nil {
		rootHash = root.RootHash
		blockHeight = root.BlockHeight
	}

	// Get budget
	budget, _ := s.store.GetTenantBudget(r.Context(), tenantID)

	// Define MCP tools
	mcpTools := []map[string]interface{}{
		{
			"name":        "garuda_propose_decision",
			"description": "Propose governance decision with automatic contradiction evaluation",
			"parameters": map[string]interface{}{
				"title":        "string",
				"scope_domain": "string",
				"scope_system": "string",
			},
		},
		{
			"name":        "garuda_handoff_task",
			"description": "Execute atomic, crash-safe task handoff to another agent",
			"parameters": map[string]interface{}{
				"task_id":         "uuid",
				"source_agent_id": "uuid",
				"target_agent_id": "uuid",
				"checkpoint_data": "object",
			},
		},
		{
			"name":        "garuda_resume_agent",
			"description": "Resume an agent from a checkpoint",
			"parameters": map[string]interface{}{
				"agent_id":      "uuid",
				"checkpoint_id": "uuid",
			},
		},
		{
			"name":        "garuda_get_lineage",
			"description": "Get the full lineage DAG for a task",
			"parameters": map[string]interface{}{
				"task_id": "uuid",
			},
		},
	}

	resp := map[string]interface{}{
		"system":             "Garuda Engine",
		"version":            "v1.0.0",
		"tenant_id":          tenantID.String(),
		"active_merkle_root": rootHash,
		"block_height":       blockHeight,
		"endpoints": map[string]string{
			"decisions_submit": "POST /decisions",
			"agent_handoff":    "POST /agents/handoff",
			"agent_resume":     "POST /agents/resume",
			"lineage_dag":      "GET /decisions/{id}/lineage",
			"audit_verify":     "GET /audit/verify",
		},
		"mcp_tools": mcpTools,
		"budget":    budget,
	}

	// Add Merkle headers
	w.Header().Set("X-Garuda-Merkle-Root", rootHash)
	w.Header().Set("X-Garuda-Block-Height", string(rune(blockHeight)))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// HandleSandbox provides an interactive playground for human onboarding.
func (s *Server) HandleSandbox(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Parse any JSON body
	var body interface{}
	json.NewDecoder(r.Body).Decode(&body)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Welcome to Garuda! Your request was received.",
		"received":      body,
		"real_endpoint": "/decisions",
		"tutorial":      "Try POST /decisions with {\"title\": \"Your decision\", \"scope_domain\": \"...\", \"scope_system\": \"...\"}",
		"documentation": "https://docs.garuda.dev",
	})
}
