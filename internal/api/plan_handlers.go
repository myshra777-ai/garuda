package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/types"
)

// HandleGetPlan returns a structured plan for a given scope.
func (s *Server) HandleGetPlan(w http.ResponseWriter, r *http.Request) {
	tenantID, err := resolveTenantID(r, uuid.Nil)
	if err != nil {
		s.RespondWithError(w, http.StatusUnauthorized, "tenant_id required")
		return
	}

	// Parse query parameters
	scopeDomain := r.URL.Query().Get("domain")
	scopeSystem := r.URL.Query().Get("system")
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}
	atStr := r.URL.Query().Get("at")
	var at *time.Time
	if atStr != "" {
		t, err := time.Parse(time.RFC3339, atStr)
		if err == nil {
			at = &t
		}
	}
	statuses := r.URL.Query()["status"]

	req := &types.PlanRequest{
		ScopeDomain: scopeDomain,
		ScopeSystem: scopeSystem,
		At:          at,
		Statuses:    statuses,
	}

	plan, err := s.store.GetPlan(r.Context(), tenantID, req)
	if err != nil {
		s.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if format == "markdown" {
		md := renderPlanMarkdown(plan)
		w.Header().Set("Content-Type", "text/markdown")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(md))
		return
	}

	// Default JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(plan)
}

// renderPlanMarkdown converts a types.PlanResult to Markdown.
func renderPlanMarkdown(p *types.PlanResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# 📋 Plan for Scope %s/%s\n\n", p.Scope.Domain, p.Scope.System))
	sb.WriteString(fmt.Sprintf("*Generated at: %s*\n\n", p.GeneratedAt.Format(time.RFC3339)))

	// Decisions
	sb.WriteString("## ✅ Canonical Decisions\n\n")
	if len(p.Decisions) == 0 {
		sb.WriteString("*No active decisions.*\n\n")
	} else {
		for _, d := range p.Decisions {
			sb.WriteString(fmt.Sprintf("- **%s** (ID: `%s`) – Status: `%s`\n", d.Title, d.ID.String(), d.Status))
			if d.ValidTo != nil {
				sb.WriteString(fmt.Sprintf("  *Valid until: %s*\n", d.ValidTo.Format(time.RFC3339)))
			}
		}
		sb.WriteString("\n")
	}

	// Tasks
	sb.WriteString("## 📌 Tasks\n\n")
	if len(p.Tasks) == 0 {
		sb.WriteString("*No tasks in progress.*\n\n")
	} else {
		for _, t := range p.Tasks {
			owner := "Unassigned"
			if t.OwnerAgentID != nil {
				owner = t.OwnerAgentID.String()
			}
			sb.WriteString(fmt.Sprintf("- **%s** (ID: `%s`) – Owner: `%s`, Status: `%s`\n", t.Title, t.ID.String(), owner, t.Status))
		}
		sb.WriteString("\n")
	}

	// Handoffs
	sb.WriteString("## 🔄 Recent Handoffs\n\n")
	if len(p.Handoffs) == 0 {
		sb.WriteString("*No handoffs recorded.*\n\n")
	} else {
		for _, h := range p.Handoffs {
			sb.WriteString(fmt.Sprintf("- **Handoff %s**: `%s` → `%s` (Task: %s)\n", h.ID.String(), h.SourceAgentID, h.TargetAgentID, h.TaskID))
		}
		sb.WriteString("\n")
	}

	// Milestones
	sb.WriteString("## 🎯 Milestones\n\n")
	if len(p.Milestones) == 0 {
		sb.WriteString("*No milestones defined.*\n\n")
	} else {
		for _, m := range p.Milestones {
			status := "⏳ Pending"
			if m.Status == "completed" {
				status = "✅ Completed"
			}
			sb.WriteString(fmt.Sprintf("- **%s** – %s\n", m.Title, status))
			if m.DueDate != nil {
				sb.WriteString(fmt.Sprintf("  *Due: %s*\n", m.DueDate.Format(time.RFC3339)))
			}
		}
		sb.WriteString("\n")
	}

	// Dependencies
	sb.WriteString("## 🔗 Dependencies\n\n")
	if len(p.Dependencies) == 0 {
		sb.WriteString("*No dependencies.*\n\n")
	} else {
		for _, e := range p.Dependencies {
			sb.WriteString(fmt.Sprintf("- `%s` → `%s` (type: %s)\n", e.SourceTaskID, e.TargetTaskID, e.EdgeType))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
