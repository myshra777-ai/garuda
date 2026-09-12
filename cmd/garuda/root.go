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
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/tenant"
	"github.com/spf13/cobra"
)

const version = "v0.4.0-multi-repo"
const defaultAPIAddr = "http://localhost:8080"

var (
	summaryFlag    bool
	progressFlag   bool
	checkpointFlag string
	agentIDFlag    string
	versionFlag    bool
)

var rootCmd = &cobra.Command{
	Use:   "garuda",
	Short: "Garuda Organizational Intelligence & Governance Runtime",
	Long: `Garuda is a multi-repository intelligence platform that builds a semantic Company Brain.
It analyzes code, extracts schemas, detects dependencies, and maintains cryptographic audit trails.`,
	Run: func(cmd *cobra.Command, args []string) {
		if versionFlag {
			fmt.Printf("Garuda Runtime Engine %s (%s/%s)\n", version, runtime.GOOS, runtime.GOARCH)
			return
		}

		if summaryFlag {
			fmt.Println("📊 Querying Executive Summary Metrics...")
			fetchEndpoint("/api/v1/dashboard/stats")
			return
		}

		if progressFlag {
			fmt.Println("🔄 Querying Active Agent Progress & Tasks...")
			atTime := time.Now().UTC().Format(time.RFC3339)
			fetchEndpoint(fmt.Sprintf("/api/v1/decisions/active?at=%s", atTime))
			return
		}

		if checkpointFlag != "" {
			fmt.Printf("🛡️ Triggering Manual Merkle Checkpoint '%s'...\n", checkpointFlag)
			postEndpoint("/api/v1/agents/checkpoint", map[string]string{
				"agent_id":        agentIDFlag,
				"checkpoint_name": checkpointFlag,
				"reason":          "manual_cli_trigger",
			})
			return
		}

		if len(args) > 0 && strings.HasPrefix(args[0], "/") {
			fmt.Printf("⚡ Slash command detected. Invoking MCP bridge for '%s'...\n", args[0])
			return
		}

		_ = cmd.Help()
	},
}

func init() {
	rootCmd.Flags().BoolVar(&summaryFlag, "summary", false, "Fetch executive metrics")
	rootCmd.Flags().BoolVar(&progressFlag, "progress", false, "Fetch active decisions")
	rootCmd.Flags().StringVar(&checkpointFlag, "checkpoint", "", "Create checkpoint")
	rootCmd.Flags().StringVar(&agentIDFlag, "agent", "cli-operator", "Agent ID context")
	rootCmd.Flags().BoolVar(&versionFlag, "version", false, "Display runtime version")

	if statsCmd != nil {
		rootCmd.AddCommand(statsCmd)
	}
}

func getDBURL() string {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		fmt.Fprintln(os.Stderr, "⚠️ WARNING: DATABASE_URL not set. Using default local dev string.")
		url = "postgres://test:test@localhost:5433/garuda_test?sslmode=disable"
	}
	return url
}

// getTenantIDString resolves the tenant for CLI commands.
//
// Behavior change from the previous implementation: an unset
// GARUDA_TENANT_ID no longer aborts the process. It resolves to the
// canonical tenant, matching the behavior of the API, the MCP server,
// and the twelve CLI commands that already hardcoded the canonical
// value. A malformed GARUDA_TENANT_ID still aborts — that is a
// configuration error, not a default.
//
// See internal/tenant for the resolution contract.
func getTenantIDString() string {
	id, err := tenant.Resolve()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ FATAL: %v\n", err)
		os.Exit(1)
	}
	return id.String()
}

func getTenantID() uuid.UUID {
	id, err := tenant.Resolve()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ FATAL: %v\n", err)
		os.Exit(1)
	}
	return id
}

func getAuthToken() string {
	if token := os.Getenv("GARUDA_API_KEY"); token != "" {
		return token
	}

	fmt.Fprintln(os.Stderr, "⚠️ WARNING: GARUDA_API_KEY not set. Using debug token endpoint (insecure).")
	req, err := http.NewRequest("GET", defaultAPIAddr+"/debug/token?actor=cli-operator&tenant_id=00000000-0000-0000-0000-000000000001", nil)
	if err != nil {
		return ""
	}
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	var tokenData struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenData); err != nil {
		return ""
	}
	return tokenData.Token
}

func sanitizeGitURL(raw string) string {
	if !strings.Contains(raw, "@") || !strings.HasPrefix(raw, "http") {
		return raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	u.User = nil
	return u.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func openFile(filename string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", filename)
	case "darwin":
		cmd = exec.Command("open", filename)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", filename)
	default:
		fmt.Printf("⚠️ Unsupported OS. Please open %s manually.\n", filename)
		return
	}
	if err := cmd.Start(); err != nil {
		fmt.Printf("❌ Failed to open browser: %v\n", err)
	}
}

func fetchEndpoint(path string) {
	authToken := getAuthToken()
	req, err := http.NewRequest("GET", defaultAPIAddr+path, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Request construction failed: %v\n", err)
		os.Exit(1)
	}
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error connecting to Garuda API: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, body, "", "  "); err == nil {
		fmt.Println(prettyJSON.String())
	} else {
		fmt.Println(string(body))
	}
}

func postEndpoint(path string, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to marshal payload: %v\n", err)
		os.Exit(1)
	}

	authToken := getAuthToken()
	req, err := http.NewRequest("POST", defaultAPIAddr+path, bytes.NewBuffer(data))
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Request construction failed: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/json")
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error connecting to Garuda API: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, body, "", "  "); err == nil {
		fmt.Println(prettyJSON.String())
	} else {
		fmt.Println(string(body))
	}
}
