// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package analyzer

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/myshra777-ai/garuda/internal/ast/typescript"
)

// IsTypeScriptProject returns true if absPath looks like a TypeScript project.
func IsTypeScriptProject(absPath string) bool {
	if _, err := os.Stat(filepath.Join(absPath, "tsconfig.json")); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(absPath, "package.json")); err == nil {
		tsCount := 0
		_ = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				base := info.Name()
				if strings.HasPrefix(base, ".") || base == "node_modules" || base == "dist" || base == "build" {
					return filepath.SkipDir
				}
				return nil
			}
			ext := filepath.Ext(path)
			if ext == ".ts" || ext == ".tsx" {
				tsCount++
			}
			return nil
		})
		return tsCount > 0
	}
	return false
}

// AnalyzeTypeScriptWorkspace converts a TypeScript Result to analyzer.Result.
func AnalyzeTypeScriptWorkspace(ctx context.Context, absPath string) (*Result, error) {
	tsResult, err := typescript.AnalyzeDirectory(absPath)
	if err != nil {
		return nil, err
	}

	result := &Result{
		Fingerprint: tsResult.Fingerprint,
		Stats: Stats{
			Files:      tsResult.Stats.Files,
			Packages:   countUniquePackagesTS(tsResult.Entities),
			Structs:    tsResult.Stats.Classes,
			Interfaces: tsResult.Stats.Interfaces,
			Functions:  tsResult.Stats.Functions + tsResult.Stats.Methods,
			Imports:    tsResult.Stats.Imports,
		},
	}

	for _, e := range tsResult.Entities {
		result.Entities = append(result.Entities, Entity{
			ID:          e.ID,
			Kind:        mapTSKindToEntityKind(e.Kind),
			Name:        e.Name,
			Package:     e.Package,
			ModulePath:  e.Package,
			PackagePath: e.Package,
			File:        e.FilePath,
			Line:        e.LineStart,
			LineStart:   e.LineStart,
			LineEnd:     e.LineEnd,
			Exported:    e.IsExported,
			Language:    "typescript",
			Signature:   e.Signature,
		})
	}

	for _, r := range tsResult.Relationships {
		result.Relationships = append(result.Relationships, Relationship{
			From:       r.FromID,
			To:         r.ToID,
			Type:       r.ClaimType,
			Confidence: r.Confidence,
			Evidence:   Evidence{Analyzer: "typescript/tree-sitter"},
		})
	}

	return result, nil
}

func mapTSKindToEntityKind(kind string) EntityKind {
	switch kind {
	case "class":
		return KindStruct
	case "interface":
		return KindInterface
	case "function":
		return KindFunction
	case "method":
		return KindMethod
	case "module":
		return KindPackage
	case "import":
		return KindExternal
	default:
		return EntityKind(kind)
	}
}

func countUniquePackagesTS(entities []typescript.EntityRecord) int {
	seen := map[string]bool{}
	for _, e := range entities {
		if e.Package != "" {
			seen[e.Package] = true
		}
	}
	return len(seen)
}
