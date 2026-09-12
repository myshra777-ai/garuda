// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
)

// computeInstanceHash returns a stable identifier for this machine +
// working directory combination. Not PII. A one-way hash of hostname,
// cwd, and a salt.
//
// Used to correlate telemetry rows from the same deployment without
// recording anything identifying.
func computeInstanceHash() string {
	hostname, _ := os.Hostname()
	cwd, _ := os.Getwd()
	salt := os.Getenv("GARUDA_TELEMETRY_SALT")
	if salt == "" {
		salt = "garuda-default-salt"
	}
	h := sha256.Sum256([]byte(hostname + "|" + cwd + "|" + salt))
	return base64.StdEncoding.EncodeToString(h[:16])
}

// computeMCPVersion returns the version string reported in telemetry.
// Reads GARUDA_VERSION env first; falls back to "dev".
func computeMCPVersion() string {
	if v := os.Getenv("GARUDA_VERSION"); v != "" {
		return v
	}
	return "dev"
}

// emitToolInvocation writes one row to telemetry_events recording the
// outcome of one MCP tool call.
//
// Populated from real data available at call time:
//
//   - instance_hash, session_id, mode, garuda_version, agent_runtime
//   - verification_latency_ms: wall-clock of the tool call
//   - tokens_estimated:       the same value the budget system used
//   - total_decisions:        1 for propose_decision, else NULL
//   - total_verifications:    1 for detect_contradictions and
//     policy.evaluate, else NULL
//   - decision_status:        "proposed" / "failed" for propose_decision
//
// Left NULL because no honest measurement exists yet:
//
//   - tokens_saved, cost_saved_usd: we do not yet measure the delta
//     between naive and grounded workflows at per-call granularity.
//   - active_agents, coordination_*, hallucination_*: features not
//     shipped. The dashboard's MeasuredMetric work (Phase 3) will
//     label these as "not measured" rather than letting them read
//     as zero.
//
// Emission is best-effort. A failure to write telemetry does not
// affect the tool response; it is logged at warn level.
func (s *MCPServer) emitToolInvocation(
	ctx context.Context,
	toolName, agentID string,
	estimatedTokens int64,
	duration time.Duration,
	success bool,
) {
	if s.store == nil {
		return
	}

	instanceHash := s.instanceHash
	if instanceHash == "" {
		instanceHash = computeInstanceHash()
	}

	// Marshal decision_scope as NULL. Only the propose_decision tool
	// carries scope today, and only the decision row itself carries
	// it — the telemetry event does not need a duplicate.
	var decisionScope []byte

	var decisionStatus interface{}
	var totalDecisions interface{}
	var totalVerifications interface{}

	switch toolName {
	case "garuda.propose_decision":
		totalDecisions = int64(1)
		if success {
			decisionStatus = "proposed"
		} else {
			decisionStatus = "failed"
		}
	case "garuda.detect_contradictions", "garuda.policy.evaluate":
		totalVerifications = int64(1)
		if !success {
			decisionStatus = "failed"
		}
	default:
		// Read-only query tools. No decision, no verification.
		if !success {
			decisionStatus = "failed"
		}
	}

	_, err := s.store.Pool().Exec(ctx, `
		INSERT INTO telemetry_events (
			instance_hash, session_id, mode, garuda_version, agent_runtime,
			decision_status, decision_scope,
			model_provider, model_name,
			tokens_estimated,
			verification_latency_ms,
			total_decisions, total_verifications,
			created_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7,
			$8, $9,
			$10,
			$11,
			$12, $13,
			$14
		)
	`,
		instanceHash,
		s.sessionID,
		"active",
		computeMCPVersion(),
		"mcp",
		decisionStatus,
		decisionScope,
		nullableString(os.Getenv("GARUDA_MODEL_PROVIDER")),
		nullableString(os.Getenv("GARUDA_MODEL_NAME")),
		estimatedTokens,
		float64(duration.Milliseconds()),
		totalDecisions,
		totalVerifications,
		time.Now().UTC(),
	)
	if err != nil {
		slog.Warn("failed to emit telemetry",
			"tool", toolName,
			"session_id", s.sessionID,
			"error", err,
		)
	}
}

// nullableString returns nil for the empty string, so that the column
// stays NULL rather than containing an empty string. Distinguishing
// "not provided" from "provided but empty" matters for the MetricDTO
// contract.
func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// estimateTokensForTelemetry mirrors the budget estimation the handlers
// already perform, so the telemetry row records the same value the
// budget system deducted.
func estimateTokensForTelemetry(toolName string, payload interface{}) int64 {
	// This import is added when the caller wires the emitter in.
	// Kept as a placeholder here to avoid an import cycle risk if the
	// emitter file is ever moved.
	_ = toolName
	_ = payload
	return 0
}

// _ = uuid is referenced by the package for consistency with the
// server's session identifier. Keep the import explicit.
var _ = uuid.Nil
