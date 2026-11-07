// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package analyzer

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/myshra777-ai/garuda/internal/ast/java"
)

// IsJavaProject returns true if absPath looks like a Java project (Maven or Gradle)
// containing at least one .java source file.
func IsJavaProject(absPath string) bool {
	markers := []string{
		"pom.xml",
		"build.gradle",
		"build.gradle.kts",
		"settings.gradle",
		"settings.gradle.kts",
		"mvnw",
		"gradlew",
	}
	hasMarker := false
	for _, m := range markers {
		if _, err := os.Stat(filepath.Join(absPath, m)); err == nil {
			hasMarker = true
			break
		}
	}
	if !hasMarker {
		return false
	}

	javaCount := 0
	_ = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			base := info.Name()
			if strings.HasPrefix(base, ".") || base == "target" || base == "build" || base == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) == ".java" {
			javaCount++
		}
		return nil
	})
	return javaCount > 0
}

// AnalyzeJavaWorkspace runs the Java AST & DI type-annotation resolver and converts
// its output into a canonical analyzer.Result with explicit ResolutionStatus,
// ResolutionMethod (JAVA_IMPORT / AST_EXACT), and file/line Evidence.
func AnalyzeJavaWorkspace(ctx context.Context, absPath string) (*Result, error) {
	jResult, err := java.AnalyzeDirectory(absPath)
	if err != nil {
		return nil, err
	}

	buildCtx := ResolveBuildContext(absPath, "java", "", nil)
	buildCtx.ActiveFilesCount = jResult.Stats.Files
	buildCtx.GraphFingerprint = jResult.Fingerprint
	buildCtx.ReceiptHash = buildCtx.ComputeReceiptHash()

	result := &Result{
		Fingerprint:  jResult.Fingerprint,
		AnalyzedAt:   time.Now().UTC(),
		Source:       absPath,
		BuildContext: buildCtx,
		Stats: Stats{
			Files:            jResult.Stats.Files,
			Packages:         countUniquePackagesJava(jResult.Entities),
			Structs:          jResult.Stats.Classes,
			Interfaces:       jResult.Stats.Interfaces,
			Functions:        jResult.Stats.Methods,
			Imports:          jResult.Stats.Imports,
			AnalysisComplete: true,
		},
	}

	for _, e := range jResult.Entities {
		result.Entities = append(result.Entities, Entity{
			ID:          e.ID,
			Kind:        mapJavaKindToEntityKind(e.Kind),
			Name:        e.Name,
			Package:     e.Package,
			ModulePath:  e.Package,
			PackagePath: e.Package,
			File:        e.FilePath,
			Line:        e.LineStart,
			LineStart:   e.LineStart,
			LineEnd:     e.LineEnd,
			Exported:    e.IsExported,
			Language:    "java",
			Signature:   e.Signature,
		})
	}

	for _, r := range jResult.Relationships {
		resStatus := r.ResolutionStatus
		if resStatus == "" {
			resStatus = "RESOLVED"
		}
		resMethod := r.ResolutionMethod
		if resMethod == "" {
			if r.ClaimType == "IMPORTS" {
				resMethod = "JAVA_IMPORT"
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
				Analyzer:  "java/ast-type-resolver",
			},
		})
	}

	return result, nil
}

func mapJavaKindToEntityKind(kind string) EntityKind {
	switch kind {
	case "class":
		return KindStruct
	case "interface":
		return KindInterface
	case "method":
		return KindMethod
	case "module":
		return KindPackage
	default:
		return KindFunction
	}
}

func countUniquePackagesJava(entities []java.EntityRecord) int {
	pkgs := make(map[string]bool)
	for _, e := range entities {
		if e.Package != "" {
			pkgs[e.Package] = true
		}
	}
	return len(pkgs)
}
