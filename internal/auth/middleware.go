package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// Agent represents an authenticated agent.
type Agent struct {
	ID         uuid.UUID
	CustomerID uuid.UUID
	Scopes     []string
}

// AuthMiddleware validates API keys for agents.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Extract API key from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing Authorization header", http.StatusUnauthorized)
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "invalid Authorization header format", http.StatusUnauthorized)
			return
		}
		apiKey := parts[1]

		// 2. Validate API key (Phase 2: query database)
		// For now, we accept a hardcoded key for testing.
		if apiKey != "garuda_test_key_123" {
			http.Error(w, "invalid API key", http.StatusUnauthorized)
			return
		}

		// 3. Set agent context
		agent := &Agent{
			ID:         uuid.New(),
			CustomerID: uuid.MustParse("00000000-0000-0000-0000-000000000001"),
			Scopes:     []string{"read", "write"},
		}
		ctx := context.WithValue(r.Context(), "agent", agent)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetAgent retrieves the agent from the context.
func GetAgent(ctx context.Context) *Agent {
	agent, ok := ctx.Value("agent").(*Agent)
	if !ok {
		return nil
	}
	return agent
}
