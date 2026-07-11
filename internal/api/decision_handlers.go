package api

import (
	"net/http"

	"github.com/google/uuid"
)

// ProposeDecisionRequest defines the payload for proposing a new decision.
type ProposeDecisionRequest struct {
	Statement string `json:"statement"`
}

// ProposeDecisionResponse defines the response for a decision proposal.
type ProposeDecisionResponse struct {
	ID     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}

// HandleProposeDecision creates a new decision.
func (s *Server) HandleProposeDecision(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Implement decision creation logic here (will be added later)
	w.WriteHeader(http.StatusNotImplemented)
}
