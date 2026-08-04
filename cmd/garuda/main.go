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
	"syscall"
	"time"
)

const version = "v0.3.0-gas"

type ProposeRequest struct {
	Title       string `json:"title"`
	ScopeDomain string `json:"scope_domain"`
	ScopeSystem string `json:"scope_system"`
}

type ProposeResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Status      string `json:"status"`
	ScopeDomain string `json:"scope_domain"`
	ScopeSystem string `json:"scope_system"`
	MerkleHash  string `json:"merkle_hash,omitempty"`
	Message     string `json:"message,omitempty"`
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	cmd := os.Args[1]
	switch cmd {
	case "init":
		handleInit()
	case "up":
		handleUp()
	case "down":
		handleDown()
	case "status":
		handleStatus()
	case "propose":
		handlePropose(os.Args[2:])
	case "mcp":
		if len(os.Args) > 2 && os.Args[2] == "install" {
			handleMCPInstall()
		} else {
			fmt.Println("Usage: garuda mcp install")
		}
	case "dashboard":
		openDashboard()
	case "--version", "-v":
		fmt.Printf("Garuda Runtime Engine %s (%s/%s)\n", version, runtime.GOOS, runtime.GOARCH)
	default:
		if cmd[0] == '/' {
			fmt.Printf("💡 Slash command detected. Invoking MCP bridge for '%s'...\n", cmd)
		} else {
			fmt.Printf("Unknown command: %s\n\n", cmd)
			printHelp()
		}
	}
}

func printHelp() {
	fmt.Println(`
🛡️ Garuda — The Organizational Intelligence Runtime

Usage:
  garuda <command> [options]
  /garuda <action>  (Inside Cursor, Claude, or MCP clients)

Commands:
  init         Initialize local environment and generate docker-compose file
  up           Start complete Garuda stack (API, Worker, Dashboard)
  down         Stop background Garuda containers
  propose      Submit a new decision proposal (e.g. garuda propose "Enforce TLS 1.3" --scope-domain security --scope-system network)
  mcp install  Auto-inject Garuda MCP server into Claude Desktop & Cursor
  status       Inspect current Merkle root and daemon status
  dashboard    Open the Web Mission Control in browser
  --version    Display current binary release version
`)
}

func handlePropose(args []string) {
	proposeFlags := flag.NewFlagSet("propose", flag.ExitOnError)

	domain := proposeFlags.String("scope-domain", "general", "Domain boundary for the decision proposal (e.g. security, billing)")
	system := proposeFlags.String("scope-system", "cli", "System boundary for the decision proposal (e.g. auth, api)")

	// Allow usage: garuda propose "Enforce TLS 1.3" --scope-domain security --scope-system network
	if len(args) < 1 {
		fmt.Println("❌ Usage: garuda propose \"<title>\" [--scope-domain <domain>] [--scope-system <system>]")
		return
	}

	title := args[0]
	// Parse remaining flags starting after the positional title argument
	if len(args) > 1 {
		if err := proposeFlags.Parse(args[1:]); err != nil {
			fmt.Printf("❌ Invalid flags: %v\n", err)
			return
		}
	}

	reqBody := ProposeRequest{
		Title:       title,
		ScopeDomain: *domain,
		ScopeSystem: *system,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Printf("❌ Failed to marshal request: %v\n", err)
		return
	}

	// Fetch debug token for authenticating CLI requests
	tokenResp, err := http.Get("http://localhost:8080/debug/token?actor=cli-operator&tenant_id=00000000-0000-0000-0000-000000000001")
	if err != nil {
		fmt.Println("🔴 Garuda API Gateway is OFFLINE. Run 'garuda up' first.")
		return
	}
	defer tokenResp.Body.Close()

	var tokenData struct {
		Token string `json:"token"`
	}
	_ = json.NewDecoder(tokenResp.Body).Decode(&tokenData)

	// Send proposal to API gateway
	req, err := http.NewRequest("POST", "http://localhost:8080/api/v1/decisions/submit", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Printf("❌ Failed to build request: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if tokenData.Token != "" {
		req.Header.Set("Authorization", "Bearer "+tokenData.Token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		fmt.Printf("⚠️ Proposal Submission Result (%d):\n%s\n", resp.StatusCode, string(bodyBytes))
		return
	}

	var res ProposeResponse
	_ = json.Unmarshal(bodyBytes, &res)

	fmt.Println("\n✅ Decision Proposal Submitted Successfully!")
	fmt.Printf("   📌 Title:        %s\n", title)
	fmt.Printf("   🏷️  Scope Domain: %s\n", *domain)
	fmt.Printf("   ⚙️  Scope System: %s\n", *system)
	if res.ID != "" {
		fmt.Printf("   🆔 Decision ID:   %s\n", res.ID)
	}
	if res.Status != "" {
		fmt.Printf("   🚦 Status:        %s\n", res.Status)
	}
	if res.MerkleHash != "" {
		fmt.Printf("   🔗 Merkle Leaf:   %s\n", res.MerkleHash)
	}
	fmt.Println()
}

func handleInit() {
	fmt.Println("🛡️ Initializing Garuda Local Runtime...")

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
	err := os.WriteFile("docker-compose.garuda.yml", []byte(composeContent), 0644)
	if err != nil {
		fmt.Printf("❌ Failed to create compose file: %v\n", err)
		return
	}

	fmt.Println("✅ Generated docker-compose.garuda.yml successfully.")
}

func handleUp() {
	fmt.Println("🚀 Starting Garuda Organizational Intelligence Engine...")

	// 1. Auto-init if compose file is missing
	if _, err := os.Stat("docker-compose.garuda.yml"); os.IsNotExist(err) {
		handleInit()
	}

	// 2. Start containers
	cmd := exec.Command("docker-compose", "-f", "docker-compose.garuda.yml", "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Docker compose failed: %v\n", err)
		return
	}

	// 3. Set DATABASE_URL
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://test:test@localhost:5433/garuda_test?sslmode=disable"
		_ = os.Setenv("DATABASE_URL", dbURL)
	}

	// 4. Wait for PostgreSQL container to accept TCP connections
	fmt.Println("⏳ Waiting for PostgreSQL container to become ready...")
	for i := 0; i < 15; i++ {
		conn, err := net.DialTimeout("tcp", "127.0.0.1:5433", 1*time.Second)
		if err == nil {
			conn.Close()
			break
		}
		time.Sleep(1 * time.Second)
	}

	// 5. Run database migrations
	fmt.Println("📦 Running database schema migrations...")
	migCmd := exec.Command("go", "run", "cmd/migrate/main.go")
	migCmd.Env = append(os.Environ(), "DATABASE_URL="+dbURL)
	migCmd.Stdout = os.Stdout
	migCmd.Stderr = os.Stderr
	if err := migCmd.Run(); err != nil {
		fmt.Printf("❌ Migration failed: %v\n", err)
		return
	}

	// 6. Concurrently boot API Gateway and Worker
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Println("🌐 Starting API Gateway on http://localhost:8080...")
	go func() {
		apiCmd := exec.CommandContext(ctx, "go", "run", "cmd/garuda-api/main.go")
		apiCmd.Env = append(os.Environ(), "DATABASE_URL="+dbURL)
		apiCmd.Stdout = os.Stdout
		apiCmd.Stderr = os.Stderr
		if err := apiCmd.Run(); err != nil {
			fmt.Printf("❌ API server error: %v\n", err)
		}
	}()

	fmt.Println("⏳ Launching Merkle Root Snapshot Worker...")
	go func() {
		wCmd := exec.CommandContext(ctx, "go", "run", "cmd/garuda-worker/main.go")
		wCmd.Env = append(os.Environ(), "DATABASE_URL="+dbURL)
		wCmd.Stdout = os.Stdout
		wCmd.Stderr = os.Stderr
		if err := wCmd.Run(); err != nil {
			fmt.Printf("❌ Snapshot worker error: %v\n", err)
		}
	}()

	time.Sleep(3 * time.Second)
	fmt.Println("\n✅ Garuda runtime is ONLINE!")
	fmt.Println("📊 Mission Control: http://localhost:8080/dashboard")
	fmt.Println("Press Ctrl+C to stop services.")

	openDashboard()

	// 7. Keep main thread alive until interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	fmt.Println("\n🛑 Shutting down Garuda services gracefully...")
}

func handleDown() {
	fmt.Println("🛑 Stopping Garuda background containers...")
	cmd := exec.Command("docker-compose", "-f", "docker-compose.garuda.yml", "down")
	_ = cmd.Run()
	fmt.Println("✅ Garuda stopped.")
}

func handleStatus() {
	fmt.Println("🔍 Querying Garuda Runtime Health...")
	resp, err := http.Get("http://localhost:8080/api/v1/evidence/snapshots?limit=1")
	if err != nil || resp.StatusCode != http.StatusOK {
		fmt.Println("🔴 Garuda Gateway OFFLINE. Run './garuda up' to boot services.")
		return
	}
	fmt.Println("🟢 Garuda Gateway: ONLINE (:8080)")
}

func openDashboard() {
	url := "http://localhost:8080/dashboard"
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	}
	if err != nil {
		fmt.Printf("Open your browser at: %s\n", url)
	}
}
