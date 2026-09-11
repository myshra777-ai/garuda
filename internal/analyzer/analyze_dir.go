// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// AnalyzeDirectory is the single entry point for extracting a semantic
// *Result from a directory path. It replaces the legacy AST-only analyzer
// with the full workspace analyzer: cross-module type-checking, honest
// resolution classification (GO_TYPES / IMPORT_RESOLUTION / AST_EXACT /
// HEURISTIC), and evidence capture per edge.
//
// Callers:
//   - cmd/garuda/ci.go            — CI gate analysis
//   - internal/selfdescribe       — product self-description
//   - analyzer_test.go            — package smoke test
//
// The old AST-only extractor (internal/analyzer/ast_extractor.go) has been
// removed. Any relationship it emitted carried Confidence=1.0 unconditionally,
// which violated Law 7 (explicit uncertainty). The workspace analyzer is now
// the only extraction path that writes to the semantic core.

package analyzer

import (
	"context"
	"fmt"
)

// AnalyzeDirectory runs the full workspace analysis on a directory path.
//
// It is the replacement for the legacy Analyze / Extract functions. Any
// caller that needs a *Result from a directory should use this function
// and pass a context.
func AnalyzeDirectory(ctx context.Context, path string) (*Result, error) {
	ws, err := DiscoverWorkspace(path)
	if err != nil {
		return nil, fmt.Errorf("discover workspace: %w", err)
	}
	return AnalyzeWorkspaceWithOptions(ctx, ws, WorkspaceAnalysisOptions{
		Cache: NewMemoryPackageCache(),
	})
}

// Analyze is a backward-compatible wrapper around AnalyzeDirectory.
//
// Deprecated: use AnalyzeDirectory and pass a context explicitly. Retained
// only so that internal tests and out-of-tree consumers do not break during
// the deprecation window.
func Analyze(root string) (*Result, error) {
	return AnalyzeDirectory(context.Background(), root)
}

// Extract is a backward-compatible alias for Analyze.
//
// Deprecated: use AnalyzeDirectory.
func Extract(root string) (*Result, error) {
	return AnalyzeDirectory(context.Background(), root)
}
