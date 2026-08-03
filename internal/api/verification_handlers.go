package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/types"
)

// HandleVerifyDecision returns the cryptographic proof for a decision.
func (s *Server) HandleVerifyDecision(w http.ResponseWriter, r *http.Request) {
	tenantID, err := resolveTenantID(r, uuid.Nil)
	if err != nil {
		s.RespondWithError(w, http.StatusUnauthorized, "tenant_id is required")
		return
	}

	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 {
		s.RespondWithError(w, http.StatusBadRequest, "invalid verification path request")
		return
	}

	decisionIDStr := pathParts[len(pathParts)-1]
	decisionID, err := uuid.Parse(decisionIDStr)
	if err != nil {
		s.RespondWithError(w, http.StatusBadRequest, "invalid decision ID")
		return
	}

	// Fetch decision
	decision, err := s.store.GetDecision(r.Context(), tenantID, decisionID)
	if err != nil {
		s.RespondWithError(w, http.StatusNotFound, "decision not found")
		return
	}

	// Fetch tenant Merkle root
	root, err := s.store.GetMerkleRoot(r.Context(), tenantID)
	if err != nil {
		s.RespondWithError(w, http.StatusInternalServerError, "failed to fetch Merkle root")
		return
	}

	// Verify cryptographic linkage
	isValid := decision.MerkleHash != "" && root.RootHash != ""

	proof := types.MerkleProof{
		DecisionID:  decision.ID,
		LeafHash:    decision.MerkleHash,
		ParentHash:  decision.ParentMerkleHash,
		RootHash:    root.RootHash,
		BlockHeight: root.BlockHeight,
		TenantID:    tenantID,
		IsVerified:  isValid,
		CreatedAt:   decision.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(proof)
}
