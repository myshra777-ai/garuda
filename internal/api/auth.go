package api

import (
	"encoding/json"
	"net/http"

	"github.com/myshra777-ai/garuda/internal/auth"
)

// SignUpRequest represents the signup request body.
type SignUpRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

// SignInRequest represents the login request body.
type SignInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents the authentication response.
type AuthResponse struct {
	Token string     `json:"token"`
	User  *auth.User `json:"user"`
}

// HandleSignUp handles user registration.
func (s *Server) HandleSignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req SignUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		s.RespondWithError(w, http.StatusBadRequest, "email and password are required")
		return
	}
	if len(req.Password) < 8 {
		s.RespondWithError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	// Call auth service
	user, token, err := s.authService.SignUp(r.Context(), req.Email, req.Password, req.FullName)
	if err != nil {
		s.RespondWithError(w, http.StatusConflict, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(AuthResponse{
		Token: token,
		User:  user,
	})
}

// HandleSignIn handles user login.
func (s *Server) HandleSignIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req SignInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		s.RespondWithError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	user, token, err := s.authService.SignIn(r.Context(), req.Email, req.Password)
	if err != nil {
		s.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(AuthResponse{
		Token: token,
		User:  user,
	})
}

// HandleSignOut handles user logout (client-side token discard).
func (s *Server) HandleSignOut(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// No server-side action needed for JWT. Client discards token.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "signed out successfully",
	})
}
