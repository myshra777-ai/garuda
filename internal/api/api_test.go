package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/techtaytor/garuda/internal/auth"
	"github.com/techtaytor/garuda/internal/engine"
	"github.com/techtaytor/garuda/internal/lineage"
	"github.com/techtaytor/garuda/internal/types"
)

func setupTestServer(t *testing.T) (*Server, *auth.JWTConfig, *registry.Registry, http.Handler) {
	reg := registry.NewRegistry(100)
	graph := lineage.NewGraph(100)
	lineageEngine := engine.NewLineageEngine(reg, graph)
	contradictionEngine := engine.NewContradictionEngine(reg, graph)

	jwtConfig, err := auth.NewJWTConfig("garuda-test", "garuda-api-test", 15*time.Minute)
	if err != nil {
		t.Fatalf("failed to create JWT config: %v", err)
	}

	// Create in-memory user store for testing
	userStore := auth.NewMemoryUserStore()

	// Create auth service with in-memory store
	authService := auth.NewAuthService(reg, userStore, jwtConfig)

	server := NewServer(reg, contradictionEngine, lineageEngine, jwtConfig, authService)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/decisions", server.HandleSubmitDecision)
	mux.HandleFunc("GET /v1/decisions/{id}/lineage", server.HandleGetDecisionLineage)
	mux.HandleFunc("GET /v1/decisions", server.HandleListDecisions)

	rateLimiter := NewRateLimiter(100, time.Minute, 1000)
	handler := WithRecovery(
		WithLogging(
			WithRequestID(
				WithAuth(jwtConfig)(
					WithRateLimit(rateLimiter)(
						WithCORS([]string{"*"})(mux),
					),
				),
			),
		),
	)

	return server, jwtConfig, reg, handler
}

func generateTestToken(t *testing.T, jwtConfig *auth.JWTConfig, actor string) string {
	token, err := jwtConfig.GenerateToken(actor)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	return token
}

func TestHandleSubmitDecision_Valid(t *testing.T) {
	_, jwtConfig, _, handler := setupTestServer(t)
	token := generateTestToken(t, jwtConfig, "alice@company.com")

	rec := registry.DecisionRecord{
		ID:         "D-001",
		Decision:   "Use PostgreSQL for financial records",
		Scope:      registry.Scope{Domain: "infrastructure", System: "database"},
		Status:     registry.StatusDraft,
		Owner:      "alice@company.com",
		Approvers:  []string{"bob@company.com"},
		Confidence: 0.9,
	}
	body, _ := json.Marshal(rec)

	req := httptest.NewRequest(http.MethodPost, "/v1/decisions", bytes.NewReader(body))
	req = req.WithContext(context.Background())
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rr.Code)
	}
}

func TestHandleSubmitDecision_Conflict(t *testing.T) {
	_, jwtConfig, reg, handler := setupTestServer(t)
	token := generateTestToken(t, jwtConfig, "alice@company.com")

	// First, create a canonical decision
	rec1 := registry.DecisionRecord{
		ID:         "D-001",
		Decision:   "Use PostgreSQL for financial records",
		Scope:      registry.Scope{Domain: "infrastructure", System: "database"},
		Status:     registry.StatusDraft,
		Owner:      "alice@company.com",
		Approvers:  []string{"bob@company.com"},
		Confidence: 0.9,
	}
	if err := reg.Append(&rec1, "alice@company.com"); err != nil {
		t.Fatalf("failed to append initial decision: %v", err)
	}

	// Legal transition path: DRAFT → REVIEW → APPROVED → CANONICAL
	if err := reg.Transition("D-001", registry.StatusReview, "alice@company.com"); err != nil {
		t.Fatalf("failed to transition to REVIEW: %v", err)
	}
	if err := reg.Transition("D-001", registry.StatusApproved, "alice@company.com"); err != nil {
		t.Fatalf("failed to transition to APPROVED: %v", err)
	}
	if err := reg.Transition("D-001", registry.StatusCanonical, "alice@company.com"); err != nil {
		t.Fatalf("failed to transition to CANONICAL: %v", err)
	}

	// Now try to create a conflicting decision
	rec2 := registry.DecisionRecord{
		ID:         "D-002",
		Decision:   "Use MongoDB for financial records",
		Scope:      registry.Scope{Domain: "infrastructure", System: "database"},
		Status:     registry.StatusDraft,
		Owner:      "alice@company.com",
		Approvers:  []string{"bob@company.com"},
		Confidence: 0.9,
	}
	body, _ := json.Marshal(rec2)

	req := httptest.NewRequest(http.MethodPost, "/v1/decisions", bytes.NewReader(body))
	req = req.WithContext(context.Background())
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", rr.Code)
	}
}

func TestHandleGetDecisionLineage(t *testing.T) {
	_, jwtConfig, reg, handler := setupTestServer(t)
	token := generateTestToken(t, jwtConfig, "alice@company.com")

	// Create a decision
	rec := registry.DecisionRecord{
		ID:         "D-001",
		Decision:   "Use PostgreSQL",
		Scope:      registry.Scope{Domain: "infrastructure"},
		Status:     registry.StatusDraft,
		Owner:      "alice@company.com",
		Approvers:  []string{"bob@company.com"},
		Confidence: 0.9,
	}
	if err := reg.Append(&rec, "alice@company.com"); err != nil {
		t.Fatalf("failed to append decision: %v", err)
	}

	// Transition to CANONICAL
	if err := reg.Transition("D-001", registry.StatusReview, "alice@company.com"); err != nil {
		t.Fatalf("failed to transition to REVIEW: %v", err)
	}
	if err := reg.Transition("D-001", registry.StatusApproved, "alice@company.com"); err != nil {
		t.Fatalf("failed to transition to APPROVED: %v", err)
	}
	if err := reg.Transition("D-001", registry.StatusCanonical, "alice@company.com"); err != nil {
		t.Fatalf("failed to transition to CANONICAL: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/decisions/D-001/lineage", nil)
	req = req.WithContext(context.Background())
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestHandleListDecisions(t *testing.T) {
	_, jwtConfig, reg, handler := setupTestServer(t)
	token := generateTestToken(t, jwtConfig, "alice@company.com")

	// Create two decisions with different scopes
	rec1 := registry.DecisionRecord{
		ID:         "D-001",
		Decision:   "Use PostgreSQL",
		Scope:      registry.Scope{Domain: "infrastructure", System: "database"},
		Status:     registry.StatusDraft,
		Owner:      "alice@company.com",
		Approvers:  []string{"bob@company.com"},
		Confidence: 0.9,
	}
	rec2 := registry.DecisionRecord{
		ID:         "D-002",
		Decision:   "Use AWS for hosting",
		Scope:      registry.Scope{Domain: "infrastructure", System: "hosting"},
		Status:     registry.StatusDraft,
		Owner:      "alice@company.com",
		Approvers:  []string{"bob@company.com"},
		Confidence: 0.9,
	}
	if err := reg.Append(&rec1, "alice@company.com"); err != nil {
		t.Fatalf("failed to append rec1: %v", err)
	}
	if err := reg.Append(&rec2, "alice@company.com"); err != nil {
		t.Fatalf("failed to append rec2: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/decisions?system=database", nil)
	req = req.WithContext(context.Background())
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var decisions []*registry.DecisionRecord
	if err := json.NewDecoder(rr.Body).Decode(&decisions); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(decisions) != 1 {
		t.Errorf("expected 1 decision, got %d", len(decisions))
	}
	if decisions[0].ID != "D-001" {
		t.Errorf("expected D-001, got %s", decisions[0].ID)
	}
}
