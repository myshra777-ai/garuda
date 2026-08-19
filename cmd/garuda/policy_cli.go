// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var rememberCmd = &cobra.Command{
	Use:   "remember <statement>",
	Short: "Remember a policy (lock intent)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		statement := args[0]
		scopeDomain, _ := cmd.Flags().GetString("scope-domain")
		scopeSystem, _ := cmd.Flags().GetString("scope-system")
		validTo, _ := cmd.Flags().GetString("valid-to")

		payload := map[string]interface{}{
			"statement":    statement,
			"scope_domain": scopeDomain,
			"scope_system": scopeSystem,
		}
		if validTo != "" {
			t, _ := time.Parse(time.RFC3339, validTo)
			payload["valid_to"] = t
		}

		postEndpoint("/api/v1/policies", payload)
	},
}

var policiesCmd = &cobra.Command{
	Use:   "policies",
	Short: "List active policies",
	Run: func(cmd *cobra.Command, args []string) {
		domain, _ := cmd.Flags().GetString("domain")
		system, _ := cmd.Flags().GetString("system")
		url := fmt.Sprintf("/api/v1/policies?domain=%s&system=%s", domain, system)
		fetchEndpoint(url)
	},
}

var supersedePolicyCmd = &cobra.Command{
	Use:   "supersede <policy_id> <new_statement>",
	Short: "Supersede a policy with a new one",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		policyID := args[0]
		newStatement := args[1]
		scopeDomain, _ := cmd.Flags().GetString("scope-domain")
		scopeSystem, _ := cmd.Flags().GetString("scope-system")
		validTo, _ := cmd.Flags().GetString("valid-to")

		payload := map[string]interface{}{
			"statement":    newStatement,
			"scope_domain": scopeDomain,
			"scope_system": scopeSystem,
		}
		if validTo != "" {
			t, _ := time.Parse(time.RFC3339, validTo)
			payload["valid_to"] = t
		}

		postEndpoint("/api/v1/policies/"+policyID+"/supersede", payload)
	},
}

func init() {
	rememberCmd.Flags().String("scope-domain", "general", "Domain scope")
	rememberCmd.Flags().String("scope-system", "cli", "System scope")
	rememberCmd.Flags().String("valid-to", "", "Expiration (RFC3339)")

	policiesCmd.Flags().String("domain", "", "Filter by domain")
	policiesCmd.Flags().String("system", "", "Filter by system")

	supersedePolicyCmd.Flags().String("scope-domain", "general", "Domain scope")
	supersedePolicyCmd.Flags().String("scope-system", "cli", "System scope")
	supersedePolicyCmd.Flags().String("valid-to", "", "Expiration (RFC3339)")

	rootCmd.AddCommand(rememberCmd)
	rootCmd.AddCommand(policiesCmd)
	rootCmd.AddCommand(supersedePolicyCmd)
}
