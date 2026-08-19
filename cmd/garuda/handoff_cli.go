// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

var handoffCmd = &cobra.Command{
	Use:   "handoff <task_id> <source_agent_id> <target_agent_id>",
	Short: "Initiate an atomic task handoff between two agents",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		taskID := args[0]
		sourceAgentID := args[1]
		targetAgentID := args[2]
		reason, _ := cmd.Flags().GetString("reason")

		endpoint := "http://localhost:8080/api/v1/agents/handoff"
		payload := map[string]interface{}{
			"task_id":         taskID,
			"source_agent_id": sourceAgentID,
			"target_agent_id": targetAgentID,
			"reason":          reason,
			"checkpoint_data": map[string]string{"CLI_EXEC": "garuda_cli_handoff"},
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Garuda-Token", os.Getenv("GARUDA_TOKEN"))

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("❌ Handoff failed: %v\n", err)
			return
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusOK {
			fmt.Printf("✅ Task Handoff Successful!\n%s\n", string(respBody))
		} else {
			fmt.Printf("❌ Error (%d): %s\n", resp.StatusCode, string(respBody))
		}
	},
}

var lineageCmd = &cobra.Command{
	Use:   "lineage <task_id>",
	Short: "Query the lineage DAG for a given task",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		taskID := args[0]
		endpoint := fmt.Sprintf("http://localhost:8080/api/v1/tasks/%s/lineage", taskID)

		req, _ := http.NewRequest("GET", endpoint, nil)
		req.Header.Set("X-Garuda-Token", os.Getenv("GARUDA_TOKEN"))

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("❌ Lineage query failed: %v\n", err)
			return
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)
		fmt.Printf("📊 Lineage DAG Output:\n%s\n", string(respBody))
	},
}

func init() {
	handoffCmd.Flags().String("reason", "manual CLI handoff", "Reason for handoff")
	rootCmd.AddCommand(handoffCmd)
	rootCmd.AddCommand(lineageCmd)
}
