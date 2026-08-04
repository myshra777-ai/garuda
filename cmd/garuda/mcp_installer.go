package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type MCPConfig struct {
	MCPServers map[string]MCPServer `json:"mcpServers"`
}

type MCPServer struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

func handleMCPInstall() {
	fmt.Println("🔌 Detecting installed AI desktop clients...")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("❌ Unable to locate user home directory: %v\n", err)
		return
	}

	// Target config paths across OS
	var claudeConfigPath string
	switch runtime.GOOS {
	case "darwin":
		claudeConfigPath = filepath.Join(homeDir, "Library", "Application Support", "Claude", "claude_desktop_config.json")
	case "windows":
		claudeConfigPath = filepath.Join(homeDir, "AppData", "Roaming", "Claude", "claude_desktop_config.json")
	case "linux":
		claudeConfigPath = filepath.Join(homeDir, ".config", "Claude", "claude_desktop_config.json")
	}

	if claudeConfigPath != "" {
		injectMCPConfig("Claude Desktop", claudeConfigPath)
	}
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
		fmt.Printf("❌ Failed to write config to %s: %v\n", configPath, err)
		return
	}

	fmt.Printf("✅ Garuda MCP Server auto-injected into %s!\n", clientName)
}
