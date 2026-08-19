// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package parser

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/goccy/go-yaml"
	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/types"
)

// RuleSet represents the structure of a .garudarules file
type RuleSet struct {
	Version string      `yaml:"version" json:"version" toml:"version"`
	Rules   []RuleEntry `yaml:"rules" json:"rules" toml:"rules"`
}

// RuleEntry is a single rule definition
type RuleEntry struct {
	ID          string                 `yaml:"id" json:"id" toml:"id"`
	Name        string                 `yaml:"name" json:"name" toml:"name"`
	Description string                 `yaml:"description" json:"description" toml:"description"`
	Action      string                 `yaml:"action" json:"action" toml:"action"` // ALLOW, DENY, WARN
	Scope       map[string]string      `yaml:"scope" json:"scope" toml:"scope"`
	Condition   string                 `yaml:"condition" json:"condition" toml:"condition"`
	Severity    string                 `yaml:"severity" json:"severity" toml:"severity"`
	Enforcement string                 `yaml:"enforcement" json:"enforcement" toml:"enforcement"` // hard, shadow, audit
	Metadata    map[string]interface{} `yaml:"metadata" json:"metadata" toml:"metadata"`
}

// Parser handles loading and validating .garudarules files
type Parser struct {
	strict bool // strict mode: fail on unknown fields
}

// NewParser creates a new parser instance
func NewParser(strict bool) *Parser {
	return &Parser{strict: strict}
}

// ParseFile loads and parses a .garudarules file (YAML or TOML)
func (p *Parser) ParseFile(path string) (*RuleSet, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}
	return p.ParseBytes(data)
}

// ParseBytes parses raw bytes (auto-detects YAML/TOML)
func (p *Parser) ParseBytes(data []byte) (*RuleSet, error) {
	// Trim BOM and whitespace
	trimmed := bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
	trimmed = bytes.TrimSpace(trimmed)

	// Detect by first non-space character
	if len(trimmed) == 0 {
		return nil, errors.New("empty file")
	}

	firstChar := trimmed[0]

	// TOML usually starts with '#', '[' or a key
	// YAML starts with a letter, '-', or '---'
	if firstChar == '[' || firstChar == '#' {
		return p.parseTOML(trimmed)
	}
	return p.parseYAML(trimmed)
}

func (p *Parser) parseYAML(data []byte) (*RuleSet, error) {
	var rs RuleSet
	if err := yaml.Unmarshal(data, &rs); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	if err := p.validate(rs); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	return &rs, nil
}

func (p *Parser) parseTOML(data []byte) (*RuleSet, error) {
	var rs RuleSet
	if err := toml.Unmarshal(data, &rs); err != nil {
		return nil, fmt.Errorf("failed to parse TOML: %w", err)
	}
	if err := p.validate(rs); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	return &rs, nil
}

// validate ensures all rules have required fields and valid values
func (p *Parser) validate(rs RuleSet) error {
	if rs.Version == "" {
		return errors.New("missing 'version' field")
	}
	if len(rs.Rules) == 0 {
		return errors.New("no rules defined")
	}
	for _, rule := range rs.Rules {
		if rule.ID == "" {
			return fmt.Errorf("rule missing 'id'")
		}
		if rule.Name == "" {
			return fmt.Errorf("rule %s missing 'name'", rule.ID)
		}
		if rule.Condition == "" {
			return fmt.Errorf("rule %s missing 'condition'", rule.ID)
		}
		if rule.Action == "" {
			rule.Action = "ALLOW" // default
		}
		action := strings.ToUpper(rule.Action)
		if action != "ALLOW" && action != "DENY" && action != "WARN" {
			return fmt.Errorf("rule %s has invalid action '%s' (must be ALLOW, DENY, or WARN)", rule.ID, rule.Action)
		}
		if rule.Enforcement == "" {
			rule.Enforcement = "shadow" // default
		}
		enf := strings.ToLower(rule.Enforcement)
		if enf != "hard" && enf != "shadow" && enf != "audit" {
			return fmt.Errorf("rule %s has invalid enforcement '%s' (must be hard, shadow, or audit)", rule.ID, rule.Enforcement)
		}
	}
	return nil
}

// ConvertToPolicies converts a RuleSet into internal types.Policy objects
// ConvertToPolicies converts a RuleSet into internal types.Policy objects
func ConvertToPolicies(tenantID string, rs *RuleSet) ([]*types.Policy, error) {
	var policies []*types.Policy
	tid, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	for _, rule := range rs.Rules {
		// Map action to status
		status := "active"
		if strings.ToUpper(rule.Action) == "WARN" {
			status = "warning"
		} else if strings.ToUpper(rule.Action) == "ALLOW" {
			status = "passive"
		}

		// Condition JSON and metadata
		cond := map[string]interface{}{
			"condition": rule.Condition,
			"type":      inferConditionType(rule.Condition),
		}
		condJSON, _ := json.Marshal(cond)

		ruleID, err := uuid.Parse(rule.ID)
		if err != nil {
			// Generate a UUID from the string if parsing fails
			ruleID = uuid.NewSHA1(uuid.NameSpaceOID, []byte(rule.ID))
		}

		// Map rule into the canonical Policy struct. Keep original rule
		// name/description/severity in Metadata for observability.
		meta := map[string]interface{}{}
		for k, v := range rule.Metadata {
			meta[k] = v
		}
		meta["name"] = rule.Name
		meta["description"] = rule.Description
		meta["severity"] = rule.Severity
		meta["enforcement"] = strings.ToLower(rule.Enforcement)
		meta["condition"] = string(condJSON)

		pol := &types.Policy{
			ID:          ruleID,
			TenantID:    tid,
			Statement:   rule.Condition,
			ScopeDomain: rule.Scope["domain"],
			ScopeSystem: rule.Scope["system"],
			Actor:       "system",
			Status:      status,
			ValidFrom:   time.Now().UTC(),
			Metadata:    meta,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}
		policies = append(policies, pol)
	}
	return policies, nil
}

// inferConditionType guesses if it's a regex or statement
func inferConditionType(cond string) string {
	if strings.HasPrefix(cond, "regex:") || strings.HasPrefix(cond, "REGEX:") {
		return "regex"
	}
	if strings.HasPrefix(cond, "statement:") || strings.HasPrefix(cond, "STATEMENT:") {
		return "statement"
	}
	return "unknown"
}
