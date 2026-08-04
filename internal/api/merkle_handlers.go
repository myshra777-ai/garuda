package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

// HandleListMerkleSnapshots exposes the historical snapshot chain for a tenant.
func (s *Server) HandleListMerkleSnapshots(w http.ResponseWriter, r *http.Request) {
	tenantID, err := resolveTenantID(r, uuid.Nil)
	if err != nil {
		s.RespondWithError(w, http.StatusUnauthorized, "tenant_id required")
		return
	}

	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	snapshots, err := s.store.ListMerkleSnapshots(r.Context(), tenantID, limit)
	if err != nil {
		s.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"tenant_id": tenantID,
		"count":     len(snapshots),
		"snapshots": snapshots,
	})
}
