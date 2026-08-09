package api

import (
	"encoding/json"
	"net/http"
	"time"
)

// RespondWithProblemDetails sends a detailed error response using generated types.
func (s *Server) RespondWithProblemDetails(w http.ResponseWriter, status int, msg string, opts ...func(*ProblemDetails)) {
	now := time.Now().UTC()
	pd := &ProblemDetails{
		Code:      status, // value, not pointer
		Message:   msg,    // value, not pointer
		Timestamp: now,    // value, not pointer
	}
	for _, opt := range opts {
		opt(pd)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(pd)
}

// WithRemediation adds a remediation block.
func WithRemediation(action string, patch []map[string]interface{}, altEndpoint string) func(*ProblemDetails) {
	return func(pd *ProblemDetails) {
		if pd.Remediation == nil {
			pd.Remediation = &struct {
				Action              *string        `json:"action,omitempty"`
				AlternativeEndpoint *string        `json:"alternative_endpoint,omitempty"`
				SuggestedPatch      *[]interface{} `json:"suggested_patch,omitempty"`
			}{}
		}
		pd.Remediation.Action = &action
		if altEndpoint != "" {
			pd.Remediation.AlternativeEndpoint = &altEndpoint
		}
		if len(patch) > 0 {
			patchSlice := make([]interface{}, len(patch))
			for i, p := range patch {
				patchSlice[i] = p
			}
			pd.Remediation.SuggestedPatch = &patchSlice
		}
	}
}

// WithDetails adds extra context.
func WithDetails(details map[string]interface{}) func(*ProblemDetails) {
	return func(pd *ProblemDetails) {
		if pd.Details == nil {
			pd.Details = &map[string]interface{}{}
		}
		for k, v := range details {
			(*pd.Details)[k] = v
		}
	}
}

// WithErrorDomain adds an error domain.
func WithErrorDomain(domain string) func(*ProblemDetails) {
	return func(pd *ProblemDetails) {
		pd.ErrorDomain = &domain
	}
}

// RespondWithError is a convenience wrapper that forwards to RespondWithProblemDetails.
func (s *Server) RespondWithError(w http.ResponseWriter, code int, msg string) {
	s.RespondWithProblemDetails(w, code, msg)
}
