// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package policy

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseFile loads a single .yaml or .yml policy file.
func ParseFile(path string) (*Policy, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("read policy: %w", err)
	}

	var p Policy
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, "", fmt.Errorf("parse policy %s: %w", path, err)
	}

	if err := validate(&p); err != nil {
		return nil, "", fmt.Errorf("validate policy %s: %w", path, err)
	}

	sum := sha256.Sum256(data)
	return &p, fmt.Sprintf("%x", sum), nil
}

// ParseDirectory walks a directory and loads every .yaml/.yml policy it finds.
// Returns the policies and their source hashes keyed by absolute path.
func ParseDirectory(root string) ([]*Policy, map[string]string, error) {
	var policies []*Policy
	hashes := make(map[string]string)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			base := info.Name()
			if strings.HasPrefix(base, ".") || base == "vendor" || base == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}
		p, h, err := ParseFile(path)
		if err != nil {
			return nil // skip malformed policies; they're reported elsewhere
		}
		policies = append(policies, p)
		hashes[path] = h
		return nil
	})

	return policies, hashes, err
}

func validate(p *Policy) error {
	if p.ID == "" {
		return fmt.Errorf("policy id is required")
	}
	if p.Version == "" {
		p.Version = "v1"
	}
	if p.Title == "" {
		return fmt.Errorf("policy title is required")
	}
	if len(p.When) == 0 {
		return fmt.Errorf("policy must have at least one 'when' predicate")
	}
	switch p.Then.Decision {
	case DecisionAllow, DecisionWarn, DecisionReview, DecisionBlock:
	default:
		return fmt.Errorf("invalid decision %q (must be ALLOW, WARN, REVIEW, or BLOCK)", p.Then.Decision)
	}
	if p.Then.Reason == "" {
		return fmt.Errorf("policy must provide a reason")
	}
	if p.Priority == 0 {
		p.Priority = 100
	}
	if p.Language == "" {
		p.Language = "any"
	}
	for i, pred := range p.When {
		if pred.Type == "" {
			return fmt.Errorf("when[%d]: predicate type is required", i)
		}
		if !isSupportedPredicate(pred.Type) {
			return fmt.Errorf("when[%d]: unsupported predicate %q", i, pred.Type)
		}
	}
	return nil
}

func isSupportedPredicate(t string) bool {
	switch t {
	case "entity_exists",
		"claim_exists",
		"contradiction_exists",
		"verification_missing",
		"language_matches":
		return true
	}
	return false
}
