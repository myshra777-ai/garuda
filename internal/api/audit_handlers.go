package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// HandleExportAuditLogs handles GET /api/v1/audit/export?since=<RFC3339>
func (s *Server) HandleExportAuditLogs(w http.ResponseWriter, r *http.Request) {
	tenantID, err := resolveTenantID(r, uuid.Nil)
	if err != nil {
		s.RespondWithError(w, http.StatusUnauthorized, "tenant_id is required")
		return
	}

	sinceStr := r.URL.Query().Get("since")
	var since time.Time
	if sinceStr != "" {
		var parseErr error
		since, parseErr = time.Parse(time.RFC3339, sinceStr)
		if parseErr != nil {
			s.RespondWithError(w, http.StatusBadRequest, "invalid 'since' timestamp format, use RFC3339")
			return
		}
	} else {
		// Default to past 24 hours if 'since' is omitted
		since = time.Now().UTC().Add(-24 * time.Hour)
	}

	events, err := s.store.ListAuditEvents(r.Context(), tenantID, since)
	if err != nil {
		s.RespondWithError(w, http.StatusInternalServerError, "failed to query audit events: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"tenant_id": tenantID.String(),
		"count":     len(events),
		"since":     since.Format(time.RFC3339),
		"events":    events,
	})
}
