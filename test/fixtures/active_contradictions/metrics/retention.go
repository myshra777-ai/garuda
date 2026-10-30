// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package metrics

// Ops team formula: (Active Users 30d) / Total Registered Users
// LOW blast radius: Internal SLA monitoring dashboard only (4 callers).
func ComputeOpsRetention(active30d, registered int) float64 {
	if registered == 0 {
		return 0.0
	}
	return float64(active30d) / float64(registered)
}

// Marketing team formula: (Active Users 30d) / (Total Registered - Churned Trials)
// MEDIUM blast radius: CAC & campaign attribution loaders (19 callers).
func ComputeMarketingRetention(active30d, registered, churnedTrials int) float64 {
	net := registered - churnedTrials
	if net <= 0 {
		return 0.0
	}
	return float64(active30d) / float64(net)
}

// Management & Solvency Engine formula:
// HIGH_SEVERE blast radius: Core Homepage & Search Results Redirect Router (142 callers).
func EvaluateEnterpriseSolvencyGate(active30d, registered int, solvencyMargin float64) bool {
	rate := ComputeOpsRetention(active30d, registered)
	return rate >= solvencyMargin
}

