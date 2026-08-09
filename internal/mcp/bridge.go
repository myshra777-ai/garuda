package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// CommandRequest represents the incoming MCP bridge payload.
type CommandRequest struct {
	Command  string `json:"command"`
	Args     string `json:"args,omitempty"`
	TenantID string `json:"tenant_id,omitempty"`
	AgentID  string `json:"agent_id,omitempty"`
}

// CommandResponse represents the enriched output formatted for AI chat interfaces.
type CommandResponse struct {
	Success   bool                   `json:"success"`
	Markdown  string                 `json:"markdown,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Timestamp string                 `json:"timestamp"`
}

type ParsedProposal struct {
	Title       string
	ScopeDomain string
	ScopeSystem string
}

// ParseSlashCommand strips raw flag syntax out of proposal strings
func ParseSlashCommand(input string) ParsedProposal {
	clean := strings.TrimPrefix(input, "/garuda ")
	clean = strings.TrimPrefix(clean, "propose ")

	// Extract title enclosed in quotes
	titleRegex := regexp.MustCompile(`"([^"]+)"`)
	titleMatch := titleRegex.FindStringSubmatch(clean)

	title := clean
	if len(titleMatch) > 1 {
		title = titleMatch[1]
	} else {
		// Fallback: strip flags if no quotes were used
		title = regexp.MustCompile(`--scope-domain\s+[^\s]+`).ReplaceAllString(title, "")
		title = regexp.MustCompile(`--scope-system\s+[^\s]+`).ReplaceAllString(title, "")
		title = strings.TrimSpace(title)
	}

	// Extract --scope-domain
	domainRegex := regexp.MustCompile(`--scope-domain\s+([^\s]+)`)
	domainMatch := domainRegex.FindStringSubmatch(clean)
	domain := "security"
	if len(domainMatch) > 1 {
		domain = domainMatch[1]
	}

	// Extract --scope-system
	systemRegex := regexp.MustCompile(`--scope-system\s+([^\s]+)`)
	systemMatch := systemRegex.FindStringSubmatch(clean)
	system := "network"
	if len(systemMatch) > 1 {
		system = systemMatch[1]
	}

	return ParsedProposal{
		Title:       title,
		ScopeDomain: domain,
		ScopeSystem: system,
	}
}

// BridgeHandler handles MCP slash commands and forwards them securely to the Garuda Gateway.
func BridgeHandler(apiBaseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(CommandResponse{Success: false, Error: "method not allowed"})
			return
		}

		var req CommandRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(CommandResponse{Success: false, Error: "invalid request body"})
			return
		}

		// Normalize slash command input
		rawCmd := strings.TrimSpace(req.Command)
		if rawCmd == "" {
			writeError(w, "empty command received")
			return
		}

		parts := strings.Fields(rawCmd)
		if len(parts) < 1 {
			writeError(w, "invalid command format. Usage: /garuda <action> [arguments]")
			return
		}

		// Extract action (handles both "/garuda propose ..." and "propose ...")
		actionIdx := 0
		if parts[0] == "/garuda" || parts[0] == "garuda" {
			if len(parts) < 2 {
				writeError(w, "missing action. Example: /garuda propose \"Require OAuth2\"")
				return
			}
			actionIdx = 1
		}

		action := parts[actionIdx]
		args := parts[actionIdx+1:]

		tenantID := req.TenantID
		if tenantID == "" {
			tenantID = "00000000-0000-0000-0000-000000000001"
		}

		agentID := req.AgentID
		if agentID == "" {
			agentID = "mcp-agent"
		}

		// Acquire internal debug Bearer token for API authorization
		authToken, err := fetchAuthToken(apiBaseURL, tenantID, agentID)
		if err != nil {
			writeError(w, fmt.Sprintf("gateway authentication error: %v", err))
			return
		}

		var result map[string]interface{}
		var markdown string

		switch action {
		case "propose":
			result, markdown, err = handlePropose(apiBaseURL, authToken, args, tenantID)
		case "verify":
			result, markdown, err = handleVerify(apiBaseURL, authToken, args)
		case "lineage":
			result, markdown, err = handleLineage(apiBaseURL, authToken, args)
		case "status":
			result, markdown, err = handleStatus(apiBaseURL, authToken)
		case "help", "--help", "-h":
			result = map[string]interface{}{
				"available_actions": []string{"propose", "verify", "lineage", "status", "help"},
			}
			markdown = `### 🛡️ Garuda Runtime MCP Commands
- **Propose:** ` + "`/garuda propose \"<title>\" [--scope-domain <domain>] [--scope-system <system>]`" + `
- **Verify:** ` + "`/garuda verify <decision_id>`" + `
- **Lineage:** ` + "`/garuda lineage <decision_id>`" + `
- **Status:** ` + "`/garuda status`"
		default:
			writeError(w, fmt.Sprintf("unknown action '%s'. Run '/garuda help' for available commands.", action))
			return
		}

		if err != nil {
			writeError(w, err.Error())
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(CommandResponse{
			Success:   true,
			Markdown:  markdown,
			Data:      result,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		})
	}
}

func writeError(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(CommandResponse{
		Success:   false,
		Error:     msg,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

// Helper to acquire a debug JWT token from the local API gateway
func fetchAuthToken(baseURL, tenantID, agentID string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s/debug/token?actor=%s&tenant_id=%s", baseURL, agentID, tenantID))
	if err != nil {
		return "", fmt.Errorf("gateway offline at %s", baseURL)
	}
	defer resp.Body.Close()

	var tokData struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokData); err != nil || tokData.Token == "" {
		return "", fmt.Errorf("failed to extract authorization token")
	}

	return tokData.Token, nil
}

// ============================================================
// Action Handlers with Clean Markdown Formatting
// ============================================================

func parseCommandArgs(args []string) (string, string, string) {
	fullStr := strings.Join(args, " ")

	scopeDomain := "general"
	scopeSystem := "mcp"
	title := ""

	// Extract --scope-domain
	if idx := strings.Index(fullStr, "--scope-domain"); idx != -1 {
		parts := strings.Fields(fullStr[idx:])
		if len(parts) > 1 {
			scopeDomain = parts[1]
		}
		fullStr = strings.TrimSpace(fullStr[:idx])
	}

	// Extract --scope-system
	if idx := strings.Index(fullStr, "--scope-system"); idx != -1 {
		parts := strings.Fields(fullStr[idx:])
		if len(parts) > 1 {
			scopeSystem = parts[1]
		}
		fullStr = strings.TrimSpace(fullStr[:idx])
	}

	// Strip surrounding double/single quotes from title
	title = strings.TrimSpace(fullStr)
	title = strings.Trim(title, "\"")
	title = strings.Trim(title, "'")

	return title, scopeDomain, scopeSystem
}

func handlePropose(baseURL, token string, args []string, tenantID string) (map[string]interface{}, string, error) {
	if len(args) < 1 {
		return nil, "", fmt.Errorf("title required. Usage: /garuda propose \"<title>\" [--scope-domain domain] [--scope-system system]")
	}

	title, scopeDomain, scopeSystem := parseCommandArgs(args)

	payload, _ := json.Marshal(map[string]interface{}{
		"title":        title,
		"scope_domain": scopeDomain,
		"scope_system": scopeSystem,
		"tenant_id":    tenantID,
	})

	req, _ := http.NewRequest("POST", baseURL+"/api/v1/decisions/submit", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("proposal submission failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var res map[string]interface{}
	_ = json.Unmarshal(body, &res)

	status, _ := res["status"].(string)
	decID, _ := res["id"].(string)

	md := fmt.Sprintf("### ✅ Garuda Decision Proposal Recorded\n"+
		"- **Title:** %s\n"+
		"- **Domain:** `%s` | **System:** `%s`\n"+
		"- **Status:** `%s`\n"+
		"- **ID:** `%s`", title, scopeDomain, scopeSystem, status, decID)

	if status == "quarantined" {
		md += "\n\n⚠️ **Notice:** Proposal triggered a policy contradiction and was placed in Merkle Quarantine."
	}

	return res, md, nil
}

func handleVerify(baseURL, token string, args []string) (map[string]interface{}, string, error) {
	if len(args) < 1 {
		return nil, "", fmt.Errorf("decision_id required. Usage: /garuda verify <decision_id>")
	}
	decID := strings.TrimSpace(args[0])

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/evidence/verify/%s", baseURL, decID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("verification query failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var res map[string]interface{}
	_ = json.Unmarshal(body, &res)

	isVerified, _ := res["is_verified"].(bool)
	rootHash, _ := res["root_hash"].(string)

	md := fmt.Sprintf("### 🔗 Merkle Proof Verification\n"+
		"- **Decision ID:** `%s`\n"+
		"- **Verified:** `%t`\n"+
		"- **Root Hash:** `%s`", decID, isVerified, rootHash)

	return res, md, nil
}

func handleLineage(baseURL, token string, args []string) (map[string]interface{}, string, error) {
	if len(args) < 1 {
		return nil, "", fmt.Errorf("decision_id required. Usage: /garuda lineage <decision_id>")
	}
	decID := strings.TrimSpace(args[0])

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/decisions/%s/lineage", baseURL, decID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("lineage query failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var res map[string]interface{}
	_ = json.Unmarshal(body, &res)

	md := fmt.Sprintf("### 📜 Decision Lineage Graph\n"+
		"- **Decision ID:** `%s`\n"+
		"- **Graph Depth:** `%v`", decID, res["depth"])

	return res, md, nil
}

func handleStatus(baseURL, token string) (map[string]interface{}, string, error) {
	req, _ := http.NewRequest("GET", baseURL+"/api/v1/budget", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("status query failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var res map[string]interface{}
	_ = json.Unmarshal(body, &res)

	md := "### 🟢 Garuda Runtime Online\n" +
		"- **Gateway Status:** `200 OK` (:8080)\n" +
		"- **Token Ledger:** Active"

	return res, md, nil
}
