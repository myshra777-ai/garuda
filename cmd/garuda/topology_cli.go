package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var topologyRecommendCmd = &cobra.Command{
	Use:   "recommend <goal>",
	Short: "Recommend a topology for a goal",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		goal := strings.Join(args, " ")
		domain, _ := cmd.Flags().GetString("domain")
		system, _ := cmd.Flags().GetString("system")
		budget, _ := cmd.Flags().GetInt64("budget")

		payload := map[string]interface{}{
			"goal":              goal,
			"scope_domain":      domain,
			"scope_system":      system,
			"max_budget_tokens": budget,
		}
		body, _ := json.Marshal(payload)

		authToken := getAuthToken()
		req, _ := http.NewRequest("POST", defaultAPIAddr+"/api/v1/topology/recommend", strings.NewReader(string(body)))
		req.Header.Set("Content-Type", "application/json")
		if authToken != "" {
			req.Header.Set("Authorization", "Bearer "+authToken)
		}
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		fmt.Println(string(respBody))
	},
}

var topologyExecuteCmd = &cobra.Command{
	Use:   "execute --topology <id>",
	Short: "Execute a topology",
	Run: func(cmd *cobra.Command, args []string) {
		topologyID, _ := cmd.Flags().GetString("topology")
		if topologyID == "" {
			fmt.Println("❌ --topology flag is required")
			return
		}
		authToken := getAuthToken()
		req, _ := http.NewRequest("POST", defaultAPIAddr+"/api/v1/topology/"+topologyID+"/execute", nil)
		if authToken != "" {
			req.Header.Set("Authorization", "Bearer "+authToken)
		}
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		fmt.Println(string(respBody))
	},
}

func init() {
	topologyRecommendCmd.Flags().String("domain", "general", "Scope domain")
	topologyRecommendCmd.Flags().String("system", "default", "Scope system")
	topologyRecommendCmd.Flags().Int64("budget", 50000, "Max token budget")

	topologyExecuteCmd.Flags().String("topology", "", "Topology ID")

	rootCmd.AddCommand(topologyRecommendCmd)
	rootCmd.AddCommand(topologyExecuteCmd)
}
