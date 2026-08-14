package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Generate a structured plan from Garuda's truth graph",
	Run: func(cmd *cobra.Command, args []string) {
		scopeDomain, _ := cmd.Flags().GetString("scope-domain")
		scopeSystem, _ := cmd.Flags().GetString("scope-system")
		format, _ := cmd.Flags().GetString("format")
		at, _ := cmd.Flags().GetString("at")
		statuses, _ := cmd.Flags().GetStringSlice("status")

		endpoint := fmt.Sprintf("%s/api/v1/plan?domain=%s&system=%s&format=%s", defaultAPIAddr, scopeDomain, scopeSystem, format)
		if at != "" {
			endpoint += "&at=" + at
		}
		for _, s := range statuses {
			endpoint += "&status=" + s
		}

		authToken := getAuthToken()
		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			fmt.Printf("❌ Request error: %v\n", err)
			os.Exit(1)
		}
		if authToken != "" {
			req.Header.Set("Authorization", "Bearer "+authToken)
		}

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("❌ Failed to fetch plan: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("❌ API error (%d): %s\n", resp.StatusCode, string(body))
			os.Exit(1)
		}

		if format == "markdown" {
			fmt.Println(string(body))
		} else {
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, body, "", "  "); err == nil {
				fmt.Println(prettyJSON.String())
			} else {
				fmt.Println(string(body))
			}
		}
	},
}

func init() {
	planCmd.Flags().String("scope-domain", "", "Filter by domain (e.g., 'infrastructure')")
	planCmd.Flags().String("scope-system", "", "Filter by system (e.g., 'database')")
	planCmd.Flags().String("format", "markdown", "Output format: json or markdown")
	planCmd.Flags().String("at", "", "Point-in-time (RFC3339)")
	planCmd.Flags().StringSlice("status", []string{}, "Filter by status (canonical, in_progress, etc.)")
	rootCmd.AddCommand(planCmd)
}
