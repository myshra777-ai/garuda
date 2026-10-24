// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package policy

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

func ParseFile(path string) (*Policy, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read policy file %s: %w", path, err)
	}

	var p Policy
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, "", fmt.Errorf("invalid yaml in %s: %w", path, err)
	}

	if err := ValidatePolicy(&p); err != nil {
		return nil, "", fmt.Errorf("validation failed in %s: %w", path, err)
	}

	hash := sha256.Sum256(data)
	return &p, hex.EncodeToString(hash[:]), nil
}

type LoadedPolicy struct {
	Policy  *Policy
	Content string
	Hash    string
	Path    string
}

// ParseDirectory loads and validates all .yaml/.yml policies in dir.
//
// Fail-Closed Invariant: If any .yaml or .yml file in the policy directory is
// unreadable, contains invalid YAML, or fails schema/predicate validation,
// ParseDirectory returns an error immediately instead of silently skipping the
// broken policy.
func ParseDirectory(dir string) ([]LoadedPolicy, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to access policy directory %s: %w", dir, err)
	}
	if !info.IsDir() {
		p, h, err := ParseFile(dir)
		if err != nil {
			return nil, err
		}
		raw, _ := os.ReadFile(dir)
		return []LoadedPolicy{{Policy: p, Content: string(raw), Hash: h, Path: dir}}, nil
	}

	var loaded []LoadedPolicy
	err = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}

		p, h, parseErr := ParseFile(path)
		if parseErr != nil {
			return fmt.Errorf("fail-closed policy load error in %s: %w", path, parseErr)
		}
		raw, readErr := os.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("failed to read policy file %s: %w", path, readErr)
		}
		loaded = append(loaded, LoadedPolicy{
			Policy:  p,
			Content: string(raw),
			Hash:    h,
			Path:    path,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Deterministic lexical ordering by policy name
	sort.Slice(loaded, func(i, j int) bool {
		return loaded[i].Policy.Name < loaded[j].Policy.Name
	})

	return loaded, nil
}

func ValidatePolicy(p *Policy) error {
	if p.APIVersion == "" {
		return fmt.Errorf("missing apiVersion")
	}
	if p.Kind != "GarudaPolicy" {
		return fmt.Errorf("invalid kind %q, expected GarudaPolicy", p.Kind)
	}
	if p.Name == "" {
		return fmt.Errorf("missing policy name")
	}
	switch p.Severity {
	case SeverityInfo, SeverityWarn, SeverityError, SeverityCritical:
	default:
		return fmt.Errorf("invalid severity %q", p.Severity)
	}
	switch p.Action {
	case ActionAllow, ActionWarn, ActionBlock, ActionRequireApproval:
	default:
		return fmt.Errorf("invalid action %q", p.Action)
	}
	if p.Rule.Predicate == "" {
		return fmt.Errorf("missing rule.predicate")
	}
	return nil
}
