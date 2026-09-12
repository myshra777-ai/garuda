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

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/ast/python"
)

// IsPythonProject returns true if absPath looks like a Python project.
// Markers: pyproject.toml, setup.py, requirements.txt, Pipfile.
// Fallback: no go.mod but at least one .py file exists.
func IsPythonProject(absPath string) bool {
	markers := []string{"pyproject.toml", "setup.py", "requirements.txt", "Pipfile"}
	for _, m := range markers {
		if _, err := os.Stat(filepath.Join(absPath, m)); err == nil {
			return true
		}
	}

	// If a go.mod exists, treat as Go — Python projects don't ship go.mod.
	if _, err := os.Stat(filepath.Join(absPath, "go.mod")); err == nil {
		return false
	}

	// Count .py files, skipping common vendored dirs.
	pyCount := 0
	_ = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			base := info.Name()
			if strings.HasPrefix(base, ".") ||
				base == "venv" || base == ".venv" || base == "env" ||
				base == "__pycache__" || base == "node_modules" ||
				base == "site-packages" || base == "dist" || base == "build" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".py") {
			pyCount++
		}
		return nil
	})

	return pyCount > 0
}

// AnalyzePythonWorkspace runs the Python parser and converts the result
// into the standard analyzer.Result shape so downstream persistence works.
func AnalyzePythonWorkspace(ctx context.Context, absPath string) (*Result, error) {
	pyResult, err := python.AnalyzeDirectory(absPath, "", "")
	if err != nil {
		return nil, err
	}

	result := &Result{
		Fingerprint: pyResult.Fingerprint,
		Stats: Stats{
			Files:      pyResult.Stats.Files,
			Packages:   countUniquePackages(pyResult.Entities),
			Structs:    pyResult.Stats.Classes,
			Interfaces: 0,
			Functions:  pyResult.Stats.Functions + pyResult.Stats.Methods,
			Imports:    pyResult.Stats.Imports,
		},
	}

	for _, e := range pyResult.Entities {
		result.Entities = append(result.Entities, Entity{
			ID:          e.ID,
			Kind:        mapPythonKindToEntityKind(e.Kind),
			Name:        e.Name,
			Package:     e.Package,
			ModulePath:  e.Package,
			PackagePath: e.Package,
			File:        e.FilePath,
			Line:        e.LineStart,
			LineStart:   e.LineStart,
			LineEnd:     e.LineEnd,
			Exported:    e.IsExported,
			Language:    "python", // NEW
			Signature:   e.Signature,
		})
	}

	for _, r := range pyResult.Relationships {
		result.Relationships = append(result.Relationships, Relationship{
			From:       r.FromID,
			To:         r.ToID,
			Type:       r.ClaimType,
			Confidence: r.Confidence,
			Evidence: Evidence{
				Analyzer: "python/structural",
			},
		})
	}
	// Path A: import resolution.
	//
	// Build a name → UUID index from the entities we just extracted,
	// then walk the source tree to resolve every import statement.
	// Edges carry ResolutionMethodPythonImport.
	//
	// Skip if there are no package entities to correlate against.
	{
		pkgIndex := make(map[string]uuid.UUID)
		for _, e := range result.Entities {
			if e.Kind == KindPackage {
				pkgIndex[e.Name] = uuid.MustParse(e.ID)
			}
		}
		if len(pkgIndex) > 0 {
			if importRels, ierr := ResolvePythonImports(absPath, pkgIndex); ierr == nil {
				result.Relationships = append(result.Relationships, importRels...)
			}
		}
	}
	return result, nil
}

// mapPythonKindToEntityKind converts Python entity kinds to analyzer's typed enum.
func mapPythonKindToEntityKind(kind string) EntityKind {
	switch kind {
	case "class":
		return KindClass
	case "function":
		return KindFunction
	case "method":
		return KindMethod
	case "module":
		return KindPackage
	case "import":
		return KindExternal
	case "variable":
		return KindVariable
	case "constant":
		return KindConstant
	case "type":
		return KindType
	default:
		// Preserve unknown kinds as a string-typed EntityKind.
		return EntityKind(kind)
	}
}

func countUniquePackages(entities []python.EntityRecord) int {
	seen := map[string]bool{}
	for _, e := range entities {
		if e.Package != "" {
			seen[e.Package] = true
		}
	}
	return len(seen)
}
