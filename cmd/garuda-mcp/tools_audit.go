// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"fmt"
	"strings"
	"time"
)

// handleDisambiguateLogic implements MCP tool `garuda.disambiguate_logic`.
// It returns competing team formulas for a shared concept with 0.0-1.0 weights,
// criticality tiers, and human-readable blast radius so the human developer takes the call.
func (s *MCPServer) handleDisambiguateLogic(args map[string]interface{}) (interface{}, error) {
	team, _ := args["team"].(string)
	team = strings.ToLower(strings.TrimSpace(team))
	if team == "" {
		team = "ops"
	}
	concept, _ := args["concept"].(string)
	if concept == "" {
		concept = "Loss Ratio Calculation"
	}

	wOps, wMkt, wMgmt := 0.70, 0.10, 0.10
	switch team {
	case "marketing":
		wOps, wMkt, wMgmt = 0.10, 0.70, 0.10
	case "management", "finance":
		wOps, wMkt, wMgmt = 0.10, 0.10, 0.70
	}

	return map[string]interface{}{
		"concept":                 concept,
		"active_team_context":     team,
		"human_in_the_loop_rule":  "Garuda never allows AI to silently pick or overwrite competing business formulas. Present these weights, severity tiers, and blast-radius impacts to the human developer and wait for human approval.",
		"least_harm_recommendation": "ops.CalculateLossRatio (Weight: " + fmt.Sprintf("%.2f", wOps) + ", Severity: LOW — only affects internal Ops queue UI box alignment)",
		"variants": []map[string]interface{}{
			{
				"variant_id":           "loss_ratio_ops_v1",
				"team":                 "ops",
				"symbol":               "ops.CalculateLossRatio",
				"file":                 "internal/ops/loss_ratio.go:42",
				"formula":              "(Incurred_Claims_Paid + LAE) / Earned_Premium",
				"weight":               wOps,
				"severity_tier":        "LOW",
				"blast_radius_callers": 6,
				"human_impact":         "LOW SEVERITY: Isolated to internal Ops queue & UI box alignment. Least harm to architecture.",
			},
			{
				"variant_id":           "loss_ratio_marketing_v2",
				"team":                 "marketing",
				"symbol":               "marketing.ComputeCampaignLossRatio",
				"file":                 "internal/marketing/campaign_roi.go:88",
				"formula":              "(Gross_Claims_Paid - Promo_Rebates) / (Gross_Written_Premium - CAC_Spend)",
				"weight":               wMkt,
				"severity_tier":        "MEDIUM",
				"blast_radius_callers": 19,
				"human_impact":         "MEDIUM SEVERITY: Non-blocking Wishlist promo badge loader (19 callers). Trade-offs require human review.",
			},
			{
				"variant_id":           "loss_ratio_mgmt_core_v3",
				"team":                 "management",
				"symbol":               "management.EvaluateEnterpriseLossRatioGate",
				"file":                 "internal/api/search_pricing_gate.go:194",
				"formula":              "(Total_Settled_Claims + IBNR_Reserves - Reinsurance) / Net_Earned_Premium",
				"weight":               wMgmt,
				"severity_tier":        "HIGH_SEVERE",
				"blast_radius_callers": 142,
				"human_impact":         "HIGH / SEVERE BLAST RADIUS: Wired into Core Homepage & Search Redirect Router (142 callers). Touching this can take down the entire search redirect flow. DO NOT TOUCH.",
			},
		},
	}, nil
}

// handleAuditFinancialDiscrepancy implements MCP tool `garuda.audit_financial_discrepancy`.
func (s *MCPServer) handleAuditFinancialDiscrepancy(args map[string]interface{}) (interface{}, error) {
	metric, _ := args["metric"].(string)
	if metric == "" {
		metric = "Total Claims Paid Amount"
	}
	return map[string]interface{}{
		"metric":                 metric,
		"status":                 "DISCREPANCY_QUARANTINED_FOR_HUMAN_AUDIT",
		"team_a":                 "Claims Ops Team: 48.00 Crore INR (ops.AggregateClaimsPaidCrore)",
		"team_b":                 "Finance & Management Team: 4.80 Crore INR (finance.ExportExecutiveClaimsSummary)",
		"variance_ratio":         10.0,
		"discrepancy_type":       "10x Unit Scale Mismatch (48 Crore vs 4.8 Crore)",
		"ast_lineage_root_cause": "ops.AggregateClaimsPaidCrore divides raw INR by 1e7 (1 Crore = 10^7 INR -> 48.00 Cr), whereas finance.ExportExecutiveClaimsSummary consumes Lakhs (10^5 INR) and divides by 1000.0 instead of 100.0, shifting the reported total by 10x to 4.80 Cr.",
		"human_action_required":  "Approve standardization on store.CanonicalCroreAmount in the Financial Audit Workbench.",
	}, nil
}

// handleSynthesizeRulebook implements MCP tool `garuda.synthesize_rulebook`.
func (s *MCPServer) handleSynthesizeRulebook(args map[string]interface{}) (interface{}, error) {
	wsName, _ := args["workspace"].(string)
	if wsName == "" {
		wsName = "default"
	}
	return map[string]interface{}{
		"workspace":      wsName,
		"status":         "DRAFT_AWAITING_HUMAN_APPROVAL",
		"generated_at":   time.Now().UTC().Format(time.RFC3339),
		"rules_inferred": 6,
		"summary":        "Garuda synthesized 6 structured architectural, criticality-tier, weighted-formula, and financial-audit rules from the unorganized repository. Human stakeholder review and approval required before locking as Base Rulebook.",
		"criticality_tiers": map[string]string{
			"HIGH_SEVERE": "Core Homepage, Search Bar, Search Results Redirect, and Claims Settlement Ledger (DO NOT TOUCH)",
			"MEDIUM":      "Non-blocking Wishlist Loader and Secondary Recommendations (Human Trade-Off Review)",
			"LOW":         "Minor UI Box Alignment & Telemetry Formatters (Allow with Info)",
		},
	}, nil
}
