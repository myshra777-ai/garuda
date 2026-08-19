// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/myshra777-ai/garuda/internal/pool"
	"github.com/spf13/cobra"
)

var keyStatusCmd = &cobra.Command{
	Use:   "key status",
	Short: "Display key pool health and status",
	Run: func(cmd *cobra.Command, args []string) {
		// Create a temporary pool to check status
		p := pool.NewKeyPool()
		p.LoadFromEnv()

		if len(p.Slots) == 0 {
			fmt.Println("❌ No API keys configured. Run 'garuda init' first.")
			os.Exit(1)
		}

		// Run health checks
		p.HealthCheckAll(cmd.Context())

		status := p.GetStatus()
		pretty, _ := json.MarshalIndent(status, "", "  ")
		fmt.Println(string(pretty))
	},
}

func init() {
	rootCmd.AddCommand(keyStatusCmd)
}
