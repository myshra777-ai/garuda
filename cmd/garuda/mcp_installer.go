// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type MCPConfig struct {
	MCPServers map[string]MCPServer `json:"mcpServers"`
}

type MCPServer struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

func injectMCPConfig(clientName, configPath string) {
	// Create parent directory if missing
	_ = os.MkdirAll(filepath.Dir(configPath), 0755)

	var cfg MCPConfig
	cfg.MCPServers = make(map[string]MCPServer)

	// Read existing config if present
	data, err := os.ReadFile(configPath)
	if err == nil {
		_ = json.Unmarshal(data, &cfg)
	}

	// Get current garuda executable path
	execPath, err := os.Executable()
	if err != nil {
		execPath = "garuda"
	}

	// Add Garuda MCP entry
	cfg.MCPServers["garuda"] = MCPServer{
		Command: execPath,
		Args:    []string{"mcp", "serve"},
	}

	updatedJSON, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		fmt.Printf("❌ Failed to format JSON for %s: %v\n", clientName, err)
		return
	}

	if err := os.WriteFile(configPath, updatedJSON, 0644); err != nil {
		fmt.Printf("❌ Failed to write config to %s: %v\n", clientName, err)
		return
	}

	fmt.Printf("✅ Garuda MCP Server auto-injected into %s!\n", clientName)
}
