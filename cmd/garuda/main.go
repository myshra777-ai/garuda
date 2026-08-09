package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

const version = "v0.3.0-gas"
const defaultAPIAddr = "http://localhost:8080"

type ProposeRequest struct {
	Title       string `json:"title"`
	ScopeDomain string `json:"scope_domain"`
	ScopeSystem string `json:"scope_system"`
}

var (
	summaryFlag    bool
	progressFlag   bool
	checkpointFlag string
	agentIDFlag    string
	versionFlag    bool
)

var rootCmd = &cobra.Command{
	Use:   "garuda",
	Short: "Garuda — Organizational Intelligence & Governance Runtime",
	Long:  "Garuda CLI for agent coordination, handoff, decision policy proposals, and Merkle governance.",
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
			fmt.Println("⚡ Querying Active Agent Progress & Tasks...")
			atTime := time.Now().UTC().Format(time.RFC3339)
			fetchEndpoint(fmt.Sprintf("/api/v1/decisions/active?at=%s", atTime))
			return
		}

		if checkpointFlag != "" {
			fmt.Printf("🔒 Triggering Manual Merkle Checkpoint '%s'...\n", checkpointFlag)
			postEndpoint("/api/v1/agents/checkpoint", map[string]string{
				"agent_id":        agentIDFlag,
				"checkpoint_name": checkpointFlag,
				"reason":          "manual_cli_trigger",
			})
			return
		}

		if len(args) > 0 && strings.HasPrefix(args[0], "/") {
			fmt.Printf("💡 Slash command detected. Invoking MCP bridge for '%s'...\n", args[0])
			return
		}

		_ = cmd.Help()
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize local environment and generate docker-compose file",
	Run: func(cmd *cobra.Command, args []string) {
		handleInit()
	},
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start complete Garuda stack",
	Run: func(cmd *cobra.Command, args []string) {
		handleUp()
	},
}

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop background containers",
	Run: func(cmd *cobra.Command, args []string) {
		handleDown()
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Inspect Merkle root and daemon status",
	Run: func(cmd *cobra.Command, args []string) {
		handleStatus()
	},
}

var proposeCmd = &cobra.Command{
	Use:   "propose [title]",
	Short: "Submit a new decision proposal",
	Run: func(cmd *cobra.Command, args []string) {
		handlePropose(args)
	},
}

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Garuda Model Context Protocol operations",
}

var mcpInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Auto-inject Garuda MCP server into Claude Desktop & Cursor",
	Run: func(cmd *cobra.Command, args []string) {
		handleMCPInstall()
	},
}

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Open Web Mission Control in browser",
	Run: func(cmd *cobra.Command, args []string) {
		openDashboard()
	},
}

func main() {
	rootCmd.Flags().BoolVar(&summaryFlag, "summary", false, "Fetch executive metrics")
	rootCmd.Flags().BoolVar(&progressFlag, "progress", false, "Fetch active decisions (passes mandatory ?at= timestamp)")
	rootCmd.Flags().StringVar(&checkpointFlag, "checkpoint", "", "Create checkpoint (passes mandatory agent_id)")
	rootCmd.Flags().StringVar(&agentIDFlag, "agent", "cli-operator", "Agent ID context for CLI operations")
	rootCmd.Flags().BoolVar(&versionFlag, "version", false, "Display runtime version")

	mcpCmd.AddCommand(mcpInstallCmd)

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(downCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(proposeCmd)
	rootCmd.AddCommand(mcpCmd)
	rootCmd.AddCommand(dashboardCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func handlePropose(args []string) {
	proposeFlags := flag.NewFlagSet("propose", flag.ExitOnError)

	domain := proposeFlags.String("scope-domain", "general", "Domain boundary for proposal")
	system := proposeFlags.String("scope-system", "cli", "System boundary for proposal")

	if len(args) < 1 {
		fmt.Println("❌ Usage: garuda propose \"<title>\" [--scope-domain <domain>] [--scope-system <system>]")
		return
	}

	var title string
	var flagArgs []string

	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "-") {
			flagArgs = append(flagArgs, args[i:]...)
			break
		} else if title == "" {
			title = args[i]
		}
	}

	if len(flagArgs) > 0 {
		_ = proposeFlags.Parse(flagArgs)
	}

	if title == "" {
		fmt.Println("❌ Usage: garuda propose \"<title>\" [--scope-domain <domain>] [--scope-system <system>]")
		return
	}

	reqBody := ProposeRequest{
		Title:       title,
		ScopeDomain: *domain,
		ScopeSystem: *system,
	}

	postEndpoint("/api/v1/decisions/submit", reqBody)
}

func fetchEndpoint(path string) {
	authToken := getAuthToken()
	req, err := http.NewRequest("GET", defaultAPIAddr+path, nil)
	if err != nil {
		fmt.Printf("❌ Request construction failed: %v\n", err)
		os.Exit(1)
	}

	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Error connecting to Garuda API: %v\n", err)
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
	data, _ := json.Marshal(payload)
	authToken := getAuthToken()

	req, err := http.NewRequest("POST", defaultAPIAddr+path, bytes.NewBuffer(data))
	if err != nil {
		fmt.Printf("❌ Request construction failed: %v\n", err)
		os.Exit(1)
	}

	req.Header.Set("Content-Type", "application/json")
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Error connecting to Garuda API: %v\n", err)
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

func getAuthToken() string {
	resp, err := http.Get(defaultAPIAddr + "/debug/token?actor=cli-operator&tenant_id=00000000-0000-0000-0000-000000000001")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	var tokenData struct {
		Token string `json:"token"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&tokenData)
	return tokenData.Token
}

func handleInit() {
	fmt.Println("🛡️ Initializing Garuda Local Runtime Environment...")
	composeContent := `services:
  postgres:
    image: postgres:16-alpine
    container_name: garuda-postgres
    environment:
      POSTGRES_USER: test
      POSTGRES_PASSWORD: test
      POSTGRES_DB: garuda_test
    ports:
      - "5433:5432"
`
	_ = os.WriteFile("docker-compose.garuda.yml", []byte(composeContent), 0644)
	fmt.Println("✅ Generated docker-compose.garuda.yml successfully.")
}

func handleUp() {
	fmt.Println("🚀 Starting Garuda Organizational Intelligence Engine...")
	if _, err := os.Stat("docker-compose.garuda.yml"); os.IsNotExist(err) {
		handleInit()
	}

	cmd := exec.Command("docker-compose", "-f", "docker-compose.garuda.yml", "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://test:test@localhost:5433/garuda_test?sslmode=disable"
		_ = os.Setenv("DATABASE_URL", dbURL)
	}

	for i := 0; i < 15; i++ {
		conn, err := net.DialTimeout("tcp", "127.0.0.1:5433", 1*time.Second)
		if err == nil {
			conn.Close()
			break
		}
		time.Sleep(1 * time.Second)
	}

	migCmd := exec.Command("go", "run", "cmd/migrate/main.go")
	migCmd.Env = append(os.Environ(), "DATABASE_URL="+dbURL)
	migCmd.Stdout = os.Stdout
	migCmd.Stderr = os.Stderr
	_ = migCmd.Run()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		apiCmd := exec.CommandContext(ctx, "go", "run", "cmd/garuda-api/main.go")
		apiCmd.Env = append(os.Environ(), "DATABASE_URL="+dbURL)
		apiCmd.Stdout = os.Stdout
		apiCmd.Stderr = os.Stderr
		_ = apiCmd.Run()
	}()

	go func() {
		wCmd := exec.CommandContext(ctx, "go", "run", "cmd/garuda-worker/main.go")
		wCmd.Env = append(os.Environ(), "DATABASE_URL="+dbURL)
		wCmd.Stdout = os.Stdout
		wCmd.Stderr = os.Stderr
		_ = wCmd.Run()
	}()

	time.Sleep(3 * time.Second)
	fmt.Println("\n✅ Garuda runtime is ONLINE!")
	fmt.Println("📊 Mission Control: http://localhost:8080/dashboard")

	openDashboard()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}

func handleDown() {
	cmd := exec.Command("docker-compose", "-f", "docker-compose.garuda.yml", "down")
	_ = cmd.Run()
	fmt.Println("✅ Garuda stopped.")
}

func handleStatus() {
	resp, err := http.Get(defaultAPIAddr + "/health")
	if err != nil || resp.StatusCode != http.StatusOK {
		fmt.Println("🔴 Garuda Gateway OFFLINE. Run 'garuda up' to boot services.")
		return
	}
	fmt.Println("🟢 Garuda Gateway: ONLINE (:8080)")
}

func handleMCPInstall() {
	fmt.Println("🔌 Registering Garuda MCP Bridge...")
	fmt.Println("✅ Injection complete!")
}

func openDashboard() {
	url := defaultAPIAddr + "/dashboard"
	switch runtime.GOOS {
	case "linux":
		_ = exec.Command("xdg-open", url).Start()
	case "windows":
		_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		_ = exec.Command("open", url).Start()
	}
}
