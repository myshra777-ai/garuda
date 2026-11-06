// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package analyzer

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
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

// AnalyzeTypeScriptWorkspace converts a TypeScript Result to analyzer.Result
// with explicit ResolutionStatus, ResolutionMethod, and source file/line Evidence.
func AnalyzeTypeScriptWorkspace(ctx context.Context, absPath string) (*Result, error) {
	tsResult, err := typescript.AnalyzeDirectory(absPath)
	if err != nil {
		return nil, err
	}

	buildCtx := ResolveBuildContext(absPath, "typescript", "", nil)
	buildCtx.ActiveFilesCount = tsResult.Stats.Files
	buildCtx.GraphFingerprint = tsResult.Fingerprint
	buildCtx.ReceiptHash = buildCtx.ComputeReceiptHash()

	result := &Result{
		Fingerprint:  tsResult.Fingerprint,
		AnalyzedAt:   time.Now().UTC(),
		Source:       absPath,
		BuildContext: buildCtx,
		Stats: Stats{
			Files:            tsResult.Stats.Files,
			Packages:         countUniquePackagesTS(tsResult.Entities),
			Structs:          tsResult.Stats.Classes,
			Interfaces:       tsResult.Stats.Interfaces,
			Functions:        tsResult.Stats.Functions + tsResult.Stats.Methods,
			Imports:          tsResult.Stats.Imports,
			AnalysisComplete: true,
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

	seenEdges := make(map[string]bool)
	for _, r := range tsResult.Relationships {
		resStatus := r.ResolutionStatus
		if resStatus == "" {
			resStatus = "RESOLVED"
		}
		resMethod := r.ResolutionMethod
		if resMethod == "" {
			if r.ClaimType == "IMPORTS" {
				resMethod = "TS_IMPORT"
			} else if r.Confidence >= 0.90 {
				resMethod = "AST_EXACT"
			} else {
				resMethod = "HEURISTIC"
			}
		}
		evFile := r.FilePath
		if rel, relErr := filepath.Rel(absPath, r.FilePath); relErr == nil && !strings.HasPrefix(rel, "..") {
			evFile = rel
		}
		evLine := r.Line
		if evLine <= 0 {
			evLine = 1
		}
		edgeKey := r.FromID + "|" + r.ToID + "|" + r.ClaimType
		seenEdges[edgeKey] = true

		result.Relationships = append(result.Relationships, Relationship{
			From:             r.FromID,
			To:               r.ToID,
			Type:             r.ClaimType,
			Confidence:       r.Confidence,
			ResolutionStatus: resStatus,
			ResolutionMethod: resMethod,
			EpistemicClass:   "OBSERVATION",
			Evidence: Evidence{
				File:      evFile,
				Line:      evLine,
				LineStart: evLine,
				LineEnd:   evLine,
				Commit:    buildCtx.CommitSHA,
				Analyzer:  "typescript/ast-type-resolver",
			},
		})
	}

	// Path A supplemental import resolution
	fileIndex := make(map[string]uuid.UUID)
	for _, e := range result.Entities {
		if e.Kind == KindFile || e.Kind == KindPackage {
			if id, err := uuid.Parse(e.ID); err == nil {
				fileIndex[e.Name] = id
			}
		}
	}
	if len(fileIndex) > 0 {
		if importRels, ierr := ResolveTypeScriptImports(absPath, fileIndex); ierr == nil {
			for _, ir := range importRels {
				fromID := ir.From
				if id, ok := fileIndex[ir.From]; ok {
					fromID = id.String()
				}
				toID := ir.To
				if id, ok := fileIndex[ir.To]; ok {
					toID = id.String()
				}
				key := fromID + "|" + toID + "|" + ir.Type
				if !seenEdges[key] {
					seenEdges[key] = true
					result.Relationships = append(result.Relationships, ir)
				}
			}
		}
	}
	return result, nil
}

func mapTSKindToEntityKind(kind string) EntityKind {
	switch kind {
	case "class":
		return KindClass
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
