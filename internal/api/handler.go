package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/techtaytor/garuda/internal/auth"
	"github.com/techtaytor/garuda/internal/types"
)

// APIError represents the formal, sanitized external error schema.
type APIError struct {
	Code      int       `json:"code"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// Server holds the injected execution capabilities for the network layer.
type Server struct {
	store       types.DecisionStore
	authService *auth.AuthService
	jwtConfig   *auth.JWTConfig
}

// NewServer builds a network isolation frame around core business capabilities.
func NewServer(store types.DecisionStore, authService *auth.AuthService, jwtConfig *auth.JWTConfig) *Server {
	return &Server{
		store:       store,
		authService: authService,
		jwtConfig:   jwtConfig,
	}
}

// RespondWithError writes a standardized, sanitized JSON error frame to the wire.
func (s *Server) RespondWithError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(APIError{
		Code:      code,
		Message:   msg,
		Timestamp: time.Now().UTC(),
	})
}

// HandleHealth returns service health status.
func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
	})
}

// HandleDebugToken generates a JWT for testing (DO NOT USE IN PRODUCTION).
func (s *Server) HandleDebugToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	actor := r.URL.Query().Get("actor")
	if actor == "" {
		s.RespondWithError(w, http.StatusBadRequest, "missing actor parameter")
		return
	}
	token, err := s.jwtConfig.GenerateToken(actor)
	if err != nil {
		s.RespondWithError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
}
