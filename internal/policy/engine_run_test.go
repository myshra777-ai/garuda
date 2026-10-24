// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package policy

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestApplyEvaluationOutcome_BlockedByFirstBlockingEvaluation(t *testing.T) {
	res := &RunResult{
		WorkspaceID:   uuid.New(),
		WorkspaceName: "ws-test",
		EvaluatedAt:   time.Now().UTC(),
		FinalDecision: DecisionAllow,
		TotalPolicies: 3,
	}

	evAllow := &PolicyEvaluation{
		ID:         uuid.New(),
		PolicyName: "policy-allow",
		Decision:   DecisionAllow,
	}
	evBlock1 := &PolicyEvaluation{
		ID:         uuid.New(),
		PolicyName: "policy-block-1",
		Decision:   DecisionBlock,
	}
	evBlock2 := &PolicyEvaluation{
		ID:         uuid.New(),
		PolicyName: "policy-block-2",
		Decision:   DecisionBlock,
	}

	noOverride := func(ctx context.Context, policyName string, now time.Time) (*uuid.UUID, bool) {
		return nil, false
	}

	applyEvaluationOutcome(res, evAllow, noOverride, context.Background())
	if res.BlockedBy != nil {
		t.Fatalf("expected BlockedBy to be nil after ALLOW, got %v", res.BlockedBy)
	}

	applyEvaluationOutcome(res, evBlock1, noOverride, context.Background())
	if res.FinalDecision != DecisionBlock {
		t.Fatalf("expected FinalDecision=BLOCK, got %s", res.FinalDecision)
	}
	if res.BlockedBy == nil || *res.BlockedBy != evBlock1.ID {
		t.Fatalf("expected BlockedBy=%s, got %v", evBlock1.ID, res.BlockedBy)
	}

	applyEvaluationOutcome(res, evBlock2, noOverride, context.Background())
	if res.BlockedCount != 2 {
		t.Fatalf("expected BlockedCount=2, got %d", res.BlockedCount)
	}
	if *res.BlockedBy != evBlock1.ID {
		t.Fatalf("expected BlockedBy to remain first blocking evaluation %s, got %s", evBlock1.ID, *res.BlockedBy)
	}
}

func TestApplyEvaluationOutcome_AcceptedDecisionOverridesBlock(t *testing.T) {
	res := &RunResult{
		WorkspaceID:   uuid.New(),
		WorkspaceName: "ws-override",
		EvaluatedAt:   time.Now().UTC(),
		FinalDecision: DecisionAllow,
		TotalPolicies: 1,
	}

	decisionID := uuid.New()
	evBlock := &PolicyEvaluation{
		ID:              uuid.New(),
		PolicyName:      "no-cross-domain-imports",
		Decision:        DecisionBlock,
		ViolationCount:  2,
		EvidencePayload: json.RawMessage(`{"violation_count":2}`),
	}

	withOverride := func(ctx context.Context, policyName string, now time.Time) (*uuid.UUID, bool) {
		if policyName == "no-cross-domain-imports" {
			return &decisionID, true
		}
		return nil, false
	}

	applyEvaluationOutcome(res, evBlock, withOverride, context.Background())

	if evBlock.Decision != DecisionOverridden {
		t.Fatalf("expected evaluation decision to become OVERRIDDEN, got %s", evBlock.Decision)
	}
	if evBlock.OverriddenByDecisionID == nil || *evBlock.OverriddenByDecisionID != decisionID {
		t.Fatalf("expected OverriddenByDecisionID=%s, got %v", decisionID, evBlock.OverriddenByDecisionID)
	}
	if res.FinalDecision != DecisionAllow {
		t.Fatalf("expected FinalDecision to remain ALLOW after override, got %s", res.FinalDecision)
	}
	if res.BlockedCount != 0 {
		t.Fatalf("expected BlockedCount=0 after override, got %d", res.BlockedCount)
	}
	if res.PassedCount != 1 {
		t.Fatalf("expected PassedCount=1 after override, got %d", res.PassedCount)
	}
	if res.BlockedBy != nil {
		t.Fatalf("expected BlockedBy=nil after override, got %v", res.BlockedBy)
	}

	var payload map[string]any
	if err := json.Unmarshal(evBlock.EvidencePayload, &payload); err != nil {
		t.Fatalf("failed to unmarshal updated evidence payload: %v", err)
	}
	if payload["overridden_by_decision_id"] != decisionID.String() {
		t.Fatalf("expected evidence_payload.overridden_by_decision_id=%s, got %v", decisionID, payload["overridden_by_decision_id"])
	}
}

func TestApplyEvaluationOutcome_ExpiredDecisionDoesNotOverride(t *testing.T) {
	res := &RunResult{
		WorkspaceID:   uuid.New(),
		WorkspaceName: "ws-expired",
		EvaluatedAt:   time.Now().UTC(),
		FinalDecision: DecisionAllow,
		TotalPolicies: 1,
	}

	evBlock := &PolicyEvaluation{
		ID:              uuid.New(),
		PolicyName:      "no-cross-domain-imports",
		Decision:        DecisionBlock,
		ViolationCount:  1,
		EvidencePayload: json.RawMessage(`{"violation_count":1}`),
	}

	expiredLookup := func(ctx context.Context, policyName string, now time.Time) (*uuid.UUID, bool) {
		return nil, false
	}

	applyEvaluationOutcome(res, evBlock, expiredLookup, context.Background())

	if evBlock.Decision != DecisionBlock {
		t.Fatalf("expected evaluation decision to stay BLOCK when override expired, got %s", evBlock.Decision)
	}
	if res.FinalDecision != DecisionBlock {
		t.Fatalf("expected FinalDecision=BLOCK when override expired, got %s", res.FinalDecision)
	}
	if res.BlockedBy == nil || *res.BlockedBy != evBlock.ID {
		t.Fatalf("expected BlockedBy=%s, got %v", evBlock.ID, res.BlockedBy)
	}
}

func TestParseDirectory_FailsClosedOnMalformedPolicy(t *testing.T) {
	tmpDir := t.TempDir()
	validYAML := `apiVersion: garuda.io/v1alpha1
kind: GarudaPolicy
name: valid-policy
version: "1.0.0"
severity: ERROR
action: BLOCK
rule:
  predicate: max_entity_fan_in
  threshold: 10
`
	if err := os.WriteFile(filepath.Join(tmpDir, "01-valid.yaml"), []byte(validYAML), 0644); err != nil {
		t.Fatalf("failed to write valid policy: %v", err)
	}
	// Write a broken YAML policy file in the same directory
	brokenYAML := `apiVersion: garuda.io/v1alpha1
kind: GarudaPolicy
name: broken-policy
severity: INVALID_SEVERITY
action: BLOCK
`
	if err := os.WriteFile(filepath.Join(tmpDir, "02-broken.yaml"), []byte(brokenYAML), 0644); err != nil {
		t.Fatalf("failed to write broken policy: %v", err)
	}

	_, err := ParseDirectory(tmpDir)
	if err == nil {
		t.Fatalf("expected ParseDirectory to fail closed on malformed policy file, but got nil error")
	}
}
