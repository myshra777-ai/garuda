// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package ingest

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/harvester"
	"github.com/myshra777-ai/garuda/internal/store"
)

// GitMiner extracts decisions from Git repositories
type GitMiner struct {
	repoPath string
	store    *store.PostgresStore
	tenantID uuid.UUID
}

// NewGitMiner creates a new Git miner
func NewGitMiner(repoPath string, store *store.PostgresStore, tenantID uuid.UUID) *GitMiner {
	return &GitMiner{
		repoPath: repoPath,
		store:    store,
		tenantID: tenantID,
	}
}

// Mine scans the repository for decisions
func (m *GitMiner) Mine(ctx context.Context) ([]*harvester.HarvestedDecision, error) {
	repo, err := git.PlainOpen(m.repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	head, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	commitIter, err := repo.Log(&git.LogOptions{
		From:  head.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get commit log: %w", err)
	}

	var decisions []*harvester.HarvestedDecision
	processed := 0

	err = commitIter.ForEach(func(c *object.Commit) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		processed++
		if processed > 100 { // Limit for MVP
			return nil
		}

		// Extract from commit message
		msg := c.Message
		if strings.Contains(strings.ToLower(msg), "decided:") ||
			strings.Contains(strings.ToLower(msg), "decision:") ||
			strings.Contains(strings.ToLower(msg), "adr:") {
			dec := m.extractFromCommit(c)
			if dec != nil {
				decisions = append(decisions, dec)
			}
		}

		// Check for ADR files in this commit
		if err := m.checkADRs(c, &decisions); err != nil {
			slog.Warn("failed to check ADRs in commit", "hash", c.Hash.String(), "error", err)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("commit iteration failed: %w", err)
	}

	slog.Info("Git mining completed", "repo", m.repoPath, "commits_processed", processed, "decisions_found", len(decisions))
	return decisions, nil
}

// extractFromCommit parses a commit for decisions
// extractFromCommit parses a commit for decisions
func (m *GitMiner) extractFromCommit(c *object.Commit) *harvester.HarvestedDecision {
	lines := strings.Split(c.Message, "\n")
	var decisionText string
	var confidence float64

	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "decided:") || strings.Contains(lower, "decision:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				decisionText = strings.TrimSpace(parts[1])
				confidence = 0.9
				break
			}
		}
		if strings.Contains(lower, "adr:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				decisionText = strings.TrimSpace(parts[1])
				confidence = 0.85
				break
			}
		}
	}

	if decisionText == "" {
		return nil
	}

	now := time.Now().UTC()
	hd := &harvester.HarvestedDecision{
		ID:                uuid.New(),
		TenantID:          m.tenantID,
		SourceType:        "git_commit",
		SourceID:          c.Hash.String(),
		SourceURL:         fmt.Sprintf("%s/commit/%s", m.repoPath, c.Hash.String()),
		RawText:           c.Message,
		ExtractedDecision: decisionText,
		Confidence:        confidence,
		HumanValidated:    false,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	slog.Debug("created harvested decision", "tenant_id", m.tenantID.String(), "source", hd.SourceID)

	return hd
}

// checkADRs scans a commit for ADR files (docs/adr/*.md)
func (m *GitMiner) checkADRs(c *object.Commit, decisions *[]*harvester.HarvestedDecision) error {
	tree, err := c.Tree()
	if err != nil {
		return err
	}

	adrFiles := []string{}
	walker := tree.Files()
	err = walker.ForEach(func(f *object.File) error {
		if strings.Contains(f.Name, "adr") && strings.HasSuffix(f.Name, ".md") {
			adrFiles = append(adrFiles, f.Name)
		}
		return nil
	})
	if err != nil {
		return err
	}

	for _, fname := range adrFiles {
		file, err := tree.File(fname)
		if err != nil {
			continue
		}
		content, err := file.Contents()
		if err != nil {
			continue
		}
		decisionText := parseADR(content)
		if decisionText != "" {
			now := time.Now().UTC()
			dec := &harvester.HarvestedDecision{
				ID:                uuid.New(),
				TenantID:          m.tenantID,
				SourceType:        "adr",
				SourceID:          fname,
				SourceURL:         fmt.Sprintf("%s/blob/%s/%s", m.repoPath, c.Hash.String(), fname),
				RawText:           content,
				ExtractedDecision: decisionText,
				Confidence:        0.95,
				HumanValidated:    false,
				CreatedAt:         now,
				UpdatedAt:         now,
			}

			slog.Debug("created harvested decision", "tenant_id", m.tenantID.String(), "source", dec.SourceID)

			*decisions = append(*decisions, dec)
		}
	}
	return nil
}

// parseADR extracts the decision from an ADR markdown file
func parseADR(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "# Decision:") ||
			strings.HasPrefix(trim, "## Decision") ||
			strings.HasPrefix(trim, "**Decision**:") {
			parts := strings.SplitN(trim, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
		if strings.Contains(strings.ToLower(trim), "decided") {
			return trim
		}
	}
	return ""
}
