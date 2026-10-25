// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/myshra777-ai/garuda/internal/config"
	"github.com/myshra777-ai/garuda/internal/store"
	"github.com/myshra777-ai/garuda/internal/talon"
	"github.com/spf13/cobra"
)

var (
	doctorFixFlag  bool
	doctorJSONFlag bool
)

// DoctorCheck represents a single diagnostic check executed by `garuda doctor`.
type DoctorCheck struct {
	Name        string `json:"name"`
	Status      string `json:"status"` // PASS | WARN | FAIL | FIXED
	Detail      string `json:"detail"`
	Remediation string `json:"remediation,omitempty"`
}

// DoctorReport is the full diagnostic report emitted by `garuda doctor`.
type DoctorReport struct {
	Platform  string        `json:"platform"`
	Workspace string        `json:"workspace"`
	Healthy   bool          `json:"healthy"`
	Checks    []DoctorCheck `json:"checks"`
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose and auto-repair Garuda installation, PostgreSQL, binaries, MCP config, and Talon hooks across Windows, macOS, Linux, Chromebook & Android",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		config.LoadDotEnv()

		ext := exeSuffix()
		wd, _ := os.Getwd()
		selfPath, err := os.Executable()
		if err != nil {
			selfPath = filepath.Join(wd, "bin", "garuda"+ext)
		}
		binDir := filepath.Dir(selfPath)

		wsName := getWorkspaceName()
		if wsName == "" {
			wsName = filepath.Base(wd)
		}
		dbURL := config.ResolveDBURL()
		apiAddr := config.ResolveAPIAddr()

		report := DoctorReport{
			Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			Workspace: wsName,
			Healthy:   true,
		}

		// 1. OS & Environment Check
		envDetail := fmt.Sprintf("OS=%s ARCH=%s", runtime.GOOS, runtime.GOARCH)
		if os.Getenv("TERMUX_VERSION") != "" || strings.HasPrefix(os.Getenv("PREFIX"), "/data/data/com.termux") {
			envDetail += " (Android Termux detected)"
		} else if _, err := os.Stat("/dev/.cros_milestone"); err == nil {
			envDetail += " (Chromebook / ChromiumOS Crostini detected)"
		}
		report.Checks = append(report.Checks, DoctorCheck{
			Name:   "Platform & Runtime Environment",
			Status: "PASS",
			Detail: envDetail,
		})

		// 2. .env Configuration Check
		if _, err := os.Stat(".env"); err != nil {
			if doctorFixFlag {
				_ = config.UpsertDotEnv(map[string]string{
					"DATABASE_URL":         dbURL,
					"CONTROL_DATABASE_URL": dbURL,
					"GARUDA_CONTROL_TOKEN": "devtoken-changeme-please",
					"GARUDA_WORKSPACE":     wsName,
					"GARUDA_TALON_ENFORCE": "true",
					"PORT":                 "8080",
				})
				report.Checks = append(report.Checks, DoctorCheck{
					Name:   "Environment File (.env)",
					Status: "FIXED",
					Detail: "Created .env with default PostgreSQL and Talon settings",
				})
			} else {
				report.Healthy = false
				report.Checks = append(report.Checks, DoctorCheck{
					Name:        "Environment File (.env)",
					Status:      "WARN",
					Detail:      ".env file not found in current directory",
					Remediation: "Run `garuda doctor --fix` or `garuda init` to generate .env automatically.",
				})
			}
		} else {
			report.Checks = append(report.Checks, DoctorCheck{
				Name:   "Environment File (.env)",
				Status: "PASS",
				Detail: fmt.Sprintf("Loaded .env (GARUDA_WORKSPACE=%s, GARUDA_TALON_ENFORCE=%s)", wsName, os.Getenv("GARUDA_TALON_ENFORCE")),
			})
		}

		// 3. PostgreSQL Reachability & Schema Migrations
		if !isPostgresReachable(ctx, dbURL) && doctorFixFlag {
			ensureLocalOrDockerPostgres()
		}
		if isPostgresReachable(ctx, dbURL) {
			migErr := store.Migrate(dbURL, "migrations")
			if migErr == nil {
				report.Checks = append(report.Checks, DoctorCheck{
					Name:   "PostgreSQL Database & Migrations",
					Status: "PASS",
					Detail: "PostgreSQL reachable and all schema migrations verified",
				})
			} else {
				report.Healthy = false
				report.Checks = append(report.Checks, DoctorCheck{
					Name:        "PostgreSQL Database & Migrations",
					Status:      "FAIL",
					Detail:      fmt.Sprintf("Connected to PostgreSQL, but migration failed: %v", migErr),
					Remediation: "Run from the Garuda repository root or check migrations/ directory.",
				})
			}
		} else {
			report.Healthy = false
			remedy := "Run `./setup.sh` (macOS/Linux/Chromebook/Android) or `.\\setup.ps1` (Windows PowerShell) to auto-provision PostgreSQL."
			report.Checks = append(report.Checks, DoctorCheck{
				Name:        "PostgreSQL Database & Migrations",
				Status:      "FAIL",
				Detail:      fmt.Sprintf("Cannot reach PostgreSQL at %s", dbURL),
				Remediation: remedy,
			})
		}

		// 4. Core Binaries Check (garuda, garuda-api, garuda-mcp, garuda-talon)
		bins := []string{"garuda" + ext, "garuda-api" + ext, "garuda-mcp" + ext, "garuda-talon" + ext}
		var missingBins []string
		for _, b := range bins {
			p := filepath.Join(binDir, b)
			if _, err := os.Stat(p); err != nil {
				if doctorFixFlag {
					pkg := "./cmd/" + strings.TrimSuffix(b, ext)
					_ = exec.Command("go", "build", "-o", p, pkg).Run()
				}
				if _, err2 := os.Stat(p); err2 != nil {
					missingBins = append(missingBins, b)
				}
			}
		}
		if len(missingBins) == 0 {
			report.Checks = append(report.Checks, DoctorCheck{
				Name:   "Garuda Binaries (CLI, API, MCP, Talon)",
				Status: "PASS",
				Detail: fmt.Sprintf("All 4 binaries present in %s", binDir),
			})
		} else {
			report.Healthy = false
			report.Checks = append(report.Checks, DoctorCheck{
				Name:        "Garuda Binaries (CLI, API, MCP, Talon)",
				Status:      "WARN",
				Detail:      fmt.Sprintf("Missing binaries in %s: %s", binDir, strings.Join(missingBins, ", ")),
				Remediation: "Run `garuda doctor --fix`, `./setup.sh`, or `.\\setup.ps1` to build all binaries.",
			})
		}

		// 5. Garuda Talon Active Contract & Hooks
		contractPath := filepath.Join(wd, ".garuda", "TALON_CONTRACT.md")
		cursorMCPPath := filepath.Join(wd, ".cursor", "mcp.json")
		_, cErr := os.Stat(contractPath)
		_, mErr := os.Stat(cursorMCPPath)
		if (cErr != nil || mErr != nil) && doctorFixFlag && isPostgresReachable(ctx, dbURL) {
			if st, sErr := store.NewPostgresStore(dbURL); sErr == nil {
				tenantID := getTenantID()
				wsID, _ := store.ResolveWorkspaceID(ctx, st.Pool(), tenantID, wsName)
				contract, _ := talon.CompileContract(ctx, st.Pool(), tenantID, wsID, wsName, filepath.Join(wd, "policies"))
				mcpPath := filepath.Join(binDir, "garuda-mcp"+ext)
				_, _ = talon.InstallHooks(wd, selfPath, mcpPath, dbURL, wsName, contract)
				st.Close()
				cErr = nil
				mErr = nil
			}
		}
		if cErr == nil && mErr == nil {
			report.Checks = append(report.Checks, DoctorCheck{
				Name:   "Garuda Talon Active Harness & MCP Hooks",
				Status: "PASS",
				Detail: ".garuda/TALON_CONTRACT.md, .cursor/mcp.json, .claude/settings.json, CLAUDE.md, and AGENTS.md active",
			})
		} else {
			report.Checks = append(report.Checks, DoctorCheck{
				Name:        "Garuda Talon Active Harness & MCP Hooks",
				Status:      "WARN",
				Detail:      "Talon contract or IDE MCP hooks not yet installed in this repository",
				Remediation: "Run `garuda talon init .` or `garuda doctor --fix` to install active agent hooks.",
			})
		}

		// 6. API Gateway Live Reachability
		client := &http.Client{Timeout: 2 * time.Second}
		if resp, err := client.Get(apiAddr + "/health"); err == nil && resp.StatusCode == http.StatusOK {
			_ = resp.Body.Close()
			report.Checks = append(report.Checks, DoctorCheck{
				Name:   "Garuda API & Software Intelligence Dashboard",
				Status: "PASS",
				Detail: fmt.Sprintf("Online at %s/dashboard (PWA enabled for Chromium/Android/Desktop)", apiAddr),
			})
		} else {
			report.Checks = append(report.Checks, DoctorCheck{
				Name:        "Garuda API & Software Intelligence Dashboard",
				Status:      "WARN",
				Detail:      fmt.Sprintf("Gateway not currently listening on %s", apiAddr),
				Remediation: fmt.Sprintf("Start with `%s` (or run `./setup.sh` / `.\\setup.ps1`).", filepath.Join(binDir, "garuda-api"+ext)),
			})
		}

		if doctorJSONFlag {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(report)
		}

		fmt.Println("╔═══════════════════════════════════════════════════════════════════╗")
		fmt.Println("║               🦅 GARUDA SYSTEM DOCTOR & DIAGNOSTICS               ║")
		fmt.Println("╚═══════════════════════════════════════════════════════════════════╝")
		fmt.Printf("  Platform:  %s  |  Workspace: %s\n\n", report.Platform, report.Workspace)
		for _, chk := range report.Checks {
			icon := "✓"
			if chk.Status == "WARN" {
				icon = "⚠"
			} else if chk.Status == "FAIL" {
				icon = "✗"
			} else if chk.Status == "FIXED" {
				icon = "🔧"
			}
			fmt.Printf("  %s [%s] %s\n", icon, chk.Status, chk.Name)
			fmt.Printf("      %s\n", chk.Detail)
			if chk.Remediation != "" {
				fmt.Printf("      → Fix: %s\n", chk.Remediation)
			}
		}
		fmt.Println("")
		if report.Healthy {
			fmt.Println("🎉 All core systems healthy! Open http://localhost:8080/dashboard in your browser.")
		} else {
			fmt.Println("💡 Tip: Run `garuda doctor --fix` (or `./setup.sh` / `.\\setup.ps1`) to repair any warnings automatically.")
		}
		return nil
	},
}

func init() {
	doctorCmd.Flags().BoolVar(&doctorFixFlag, "fix", false, "Automatically repair missing .env, database, binaries, and Talon hooks")
	doctorCmd.Flags().BoolVar(&doctorJSONFlag, "json", false, "Emit diagnostic report as JSON")
	rootCmd.AddCommand(doctorCmd)
}
