// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/store"
)

// fakeHandoffStore embeds fakeDecisionStore (defined in
// decision_handlers_test.go, same package) to satisfy the full
// types.DecisionStore interface, and adds ExecuteHandoffTransaction so
// the type assertion in HandleHandoff succeeds. Only the handoff path
// is exercised; every other method is a stub from the embedded fake.
type fakeHandoffStore struct {
	*fakeDecisionStore
	handoffErr error
}

func (f *fakeHandoffStore) ExecuteHandoffTransaction(ctx context.Context, req *store.HandoffRequest) (*store.HandoffResponse, error) {
	if f.handoffErr != nil {
		return nil, f.handoffErr
	}
	return &store.HandoffResponse{
		HandoffID:    uuid.New(),
		CheckpointID: uuid.New(),
		TaskID:       req.TaskID,
		Status:       "completed",
	}, nil
}

// TestHandleHandoff_ErrorMapping asserts that the HTTP handoff
// handler surfaces typed precondition errors verbatim, and collapses
// every other store error into a generic message.
//
// Before the fix, every store error was collapsed into "handoff
// execution failed". A caller could not distinguish "target agent is
// offline" (change the target, retry) from "source does not own this
// task" (fix the arguments) from a wrapped internal error (retry
// later).
//
// The fourth case is the guard: it proves the fix does not
// accidentally leak the raw wrapped error text — which carries
// Postgres table and constraint names — for non-typed failures.
//
// Assertion strategy: substring match on the raw response body, not
// a decode-and-read-JSON-field. The response is ProblemDetails (see
// internal/api/error.go). Asserting on substring avoids coupling the
// test to the JSON schema in case ProblemDetails is extended.
func TestHandleHandoff_ErrorMapping(t *testing.T) {
	cases := []struct {
		name        string
		storeErr    error
		wantMessage string
		forbidden   string
	}{
		{
			name:        "source transitioning surfaces store message",
			storeErr:    store.ErrSourceTransitioning,
			wantMessage: "source agent is already in transition",
		},
		{
			name:        "target offline surfaces store message",
			storeErr:    store.ErrTargetOffline,
			wantMessage: "target agent is offline",
		},
		{
			name:        "source does not own surfaces store message",
			storeErr:    store.ErrSourceDoesNotOwn,
			wantMessage: "source agent does not own this task",
		},
		{
			name:        "wrapped internal error stays collapsed",
			storeErr:    errors.New("failed to commit handoff transaction: connection refused"),
			wantMessage: "handoff execution failed",
			forbidden:   "connection refused",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fake := &fakeHandoffStore{
				fakeDecisionStore: &fakeDecisionStore{},
				handoffErr:        tc.storeErr,
			}
			srv := &Server{store: fake}

			body := []byte(`{
				"task_id":         "77770007-0007-4007-8007-000000000007",
				"source_agent_id": "eeee0005-0005-4005-8005-000000000005",
				"target_agent_id": "ffff0006-0006-4006-8006-000000000006",
				"reason":          "b33 test"
			}`)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/handoff", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
			req = req.WithContext(context.WithValue(req.Context(), TenantIDKey, tenantID))

			rec := httptest.NewRecorder()
			srv.HandleHandoff(rec, req)

			if rec.Code != http.StatusConflict {
				t.Fatalf("status = %d, want 409 (body=%s)", rec.Code, rec.Body.String())
			}
			respBody := rec.Body.String()

			if !strings.Contains(respBody, tc.wantMessage) {
				t.Errorf("body does not contain %q\nbody: %s", tc.wantMessage, respBody)
			}
			if tc.forbidden != "" && strings.Contains(respBody, tc.forbidden) {
				t.Errorf("body leaks forbidden substring %q\nbody: %s", tc.forbidden, respBody)
			}
		})
	}
}
