// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package engine

import (
	"encoding/json"
	"math"
)

// EstimateOperationTokens computes a conservative token estimate prior to execution.
func EstimateOperationTokens(operation string, payload interface{}) int {
	baseCost := 200 // Base overhead for DB / JSON serialization

	switch operation {
	case "mcp.garuda.propose_decision", "api.propose_decision":
		baseCost += 500
	case "mcp.garuda.query", "api.query":
		baseCost += 300
	case "mcp.garuda.detect_contradictions":
		baseCost += 800
	case "agent.checkpoint":
		baseCost += 400
	}

	if payload == nil {
		return baseCost
	}

	// Dynamic payload estimation: ~1 token per 4 characters
	bytes, err := json.Marshal(payload)
	if err != nil {
		return baseCost
	}

	payloadTokens := int(math.Ceil(float64(len(bytes)) / 4.0))
	return baseCost + payloadTokens
}
