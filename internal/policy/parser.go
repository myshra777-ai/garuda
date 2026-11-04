// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParsedPolicy is the result of parsing one YAML file. The three
// values travel together so callers cannot lose the association
// between a policy and its source.
type ParsedPolicy struct {
	Policy     *Policy
	SourcePath string
	SourceHash string
}

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

// ParseFileAll loads every policy document in a .yaml/.yml file. Files with a
// single document behave exactly like ParseFile (same SourceHash). Files with
// several "---"-separated documents (as written by `garuda import`) yield one
// ParsedPolicy per document; yaml.Unmarshal alone would silently keep only the first.
func ParseFileAll(path string) ([]ParsedPolicy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read policy: %w", err)
	}
	docs := splitYAMLDocuments(data)
	if len(docs) <= 1 {
		p, h, err := ParseFile(path)
		if err != nil {
			return nil, err
		}
		return []ParsedPolicy{{Policy: p, SourcePath: path, SourceHash: h}}, nil
	}
	out := make([]ParsedPolicy, 0, len(docs))
	for i, d := range docs {
		var p Policy
		if err := yaml.Unmarshal(d, &p); err != nil {
			return nil, fmt.Errorf("parse policy %s (document %d of %d): %w", path, i+1, len(docs), err)
		}
		if err := validate(&p); err != nil {
			return nil, fmt.Errorf("validate policy %s (document %d of %d): %w", path, i+1, len(docs), err)
		}
		sum := sha256.Sum256(d)
		out = append(out, ParsedPolicy{Policy: &p, SourcePath: path, SourceHash: fmt.Sprintf("%x", sum)})
	}
	return out, nil
}

// ParseDirectory walks a directory and loads every .yaml/.yml policy it finds.
// Fail-Closed Invariant: If any .yaml/.yml file in the policy directory fails
// to parse or validate, ParseDirectory returns an error immediately instead of
// silently skipping the malformed policy.
func ParseDirectory(root string) ([]ParsedPolicy, error) {
	info, statErr := os.Stat(root)
	if statErr != nil {
		return nil, fmt.Errorf("access policy path %s: %w", root, statErr)
	}
	if !info.IsDir() {
		return ParseFileAll(root)
	}

	var parsed []ParsedPolicy

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			base := info.Name()
			if (strings.HasPrefix(base, ".") && path != root) || base == "vendor" || base == "node_modules" || (base == "templates" && filepath.Clean(path) != filepath.Clean(root)) {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}
		pps, err := ParseFileAll(path)
		if err != nil {
			return fmt.Errorf("fail-closed policy load error in %s: %w", path, err)
		}
		parsed = append(parsed, pps...)
		return nil
	})

	return parsed, err
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
		"language_matches",
		"transitive_claim_exists",
		"layer_violation",
		"cycle_exists",
		"blast_radius_exceeds":
		return true
	}
	return false
}
