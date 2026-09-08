// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start complete Garuda stack (Postgres, API, Worker)",
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

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Open Web Mission Control in browser",
	Run: func(cmd *cobra.Command, args []string) {
		openDashboard()
	},
}

func init() {
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(downCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(dashboardCmd)
}

func handleInit() {
	fmt.Println("⚙️ Initializing Garuda Local Runtime Environment...")
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
    volumes:
      - garuda-pgdata:/var/lib/postgresql/data
volumes:
  garuda-pgdata:
`
	if err := os.WriteFile("docker-compose.garuda.yml", []byte(composeContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to write compose file: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Generated docker-compose.garuda.yml successfully.")
}

func handleUp() {
	fmt.Println("🚀 Starting Garuda Organizational Intelligence Engine...")
	if _, err := os.Stat("docker-compose.garuda.yml"); os.IsNotExist(err) {
		handleInit()
	}

	cmd := exec.Command("docker-compose", "-f", "docker-compose.garuda.yml", "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to start containers: %v\n", err)
		os.Exit(1)
	}

	dbURL := getDBURL()
	fmt.Println("⏳ Waiting for Postgres to be ready...")
	ready := false
	for i := 0; i < 30; i++ {
		conn, err := net.DialTimeout("tcp", "127.0.0.1:5433", 1*time.Second)
		if err == nil {
			conn.Close()
			ready = true
			break
		}
		time.Sleep(1 * time.Second)
	}

	if !ready {
		fmt.Fprintln(os.Stderr, "❌ Postgres did not become ready in 30 seconds.")
		os.Exit(1)
	}

	fmt.Println("🛠️ Running database migrations...")
	migCmd := exec.Command("go", "run", "cmd/migrate/main.go")
	migCmd.Env = append(os.Environ(), "DATABASE_URL="+dbURL)
	migCmd.Stdout = os.Stdout
	migCmd.Stderr = os.Stderr
	if err := migCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Migration failed: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	apiCmd := exec.CommandContext(ctx, "go", "run", "cmd/garuda-api/main.go")
	apiCmd.Env = append(os.Environ(), "DATABASE_URL="+dbURL)
	apiCmd.Stdout = os.Stdout
	apiCmd.Stderr = os.Stderr
	if err := apiCmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to start API: %v\n", err)
		os.Exit(1)
	}

	workerCmd := exec.CommandContext(ctx, "go", "run", "cmd/garuda-worker/main.go")
	workerCmd.Env = append(os.Environ(), "DATABASE_URL="+dbURL)
	workerCmd.Stdout = os.Stdout
	workerCmd.Stderr = os.Stderr
	if err := workerCmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to start Worker: %v\n", err)
		_ = apiCmd.Process.Kill()
		os.Exit(1)
	}

	fmt.Println("📡 Waiting for API to come online...")
	apiReady := false
	for i := 0; i < 20; i++ {
		resp, err := http.Get(defaultAPIAddr + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			apiReady = true
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	if !apiReady {
		fmt.Fprintln(os.Stderr, "⚠️ API health check timed out. It may still be starting.")
	}

	fmt.Println("\n✅ Garuda runtime is ONLINE!")
	fmt.Println("🌐 Mission Control: http://localhost:8080/dashboard")
	openDashboard()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n🛑 Shutting down gracefully...")
	cancel()
	time.Sleep(2 * time.Second)

	_ = apiCmd.Process.Kill()
	_ = workerCmd.Process.Kill()
	fmt.Println("✅ Garuda stopped.")
}

func handleDown() {
	cmd := exec.Command("docker-compose", "-f", "docker-compose.garuda.yml", "down")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error stopping containers: %v\n", err)
	}
	fmt.Println("✅ Garuda containers stopped.")
}

func handleStatus() {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(defaultAPIAddr + "/health")
	if err != nil {
		fmt.Println("❌ Garuda Gateway OFFLINE. Run 'garuda up' to boot services.")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("⚠️ Garuda Gateway UNHEALTHY (status %d).\n", resp.StatusCode)
		return
	}
	fmt.Println("✅ Garuda Gateway: ONLINE (:8080)")
}

func openDashboard() {
	url := defaultAPIAddr + "/dashboard"
	fmt.Printf("🌐 Opening %s...\n", url)
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		fmt.Printf("⚠️ Unsupported OS. Please open %s manually.\n", url)
		return
	}

	if err != nil {
		fmt.Printf("❌ Failed to open browser: %v. Visit %s manually.\n", err, url)
	}
}
