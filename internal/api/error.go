package api

import (
	"encoding/json"
	"net/http"
	"time"
)

// RespondWithProblemDetails sends a detailed error response using functional options.
func (s *Server) RespondWithProblemDetails(w http.ResponseWriter, status int, msg string, opts ...func(*ProblemDetails)) {
	now := time.Now().UTC()
	pd := &ProblemDetails{
		Code:      status,
		Message:   msg,
		Timestamp: now,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(pd)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(pd)
}

// WithRemediation adds a remediation block to ProblemDetails.
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

// WithDetails adds extra contextual metadata.
func WithDetails(details map[string]interface{}) func(*ProblemDetails) {
	return func(pd *ProblemDetails) {
		if pd.Details == nil {
			m := make(map[string]interface{})
			pd.Details = &m
		}
		for k, v := range details {
			(*pd.Details)[k] = v
		}
	}
}

// WithErrorDomain adds an error domain category.
func WithErrorDomain(domain string) func(*ProblemDetails) {
	return func(pd *ProblemDetails) {
		pd.ErrorDomain = &domain
	}
}

// RespondWithError is a convenience wrapper for RespondWithProblemDetails.
// It accepts optional request IDs for correlation without exposing raw internal error details.
func (s *Server) RespondWithError(w http.ResponseWriter, code int, msg string, requestID ...string) {
	var opts []func(*ProblemDetails)
	if len(requestID) > 0 && requestID[0] != "" {
		opts = append(opts, WithDetails(map[string]interface{}{"request_id": requestID[0]}))
	}
	s.RespondWithProblemDetails(w, code, msg, opts...)
}
