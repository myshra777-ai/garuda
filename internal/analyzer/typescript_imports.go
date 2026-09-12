// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package analyzer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// TypeScript import resolution — Path A.
//
// Reads every .ts/.tsx file, extracts import/require statements, and
// resolves each to a concrete file in the workspace. Emits IMPORTS
// relationships with ResolutionMethodTSImport.
//
// Resolution handles:
//
//   import { X } from './relative'        → relative path resolution
//   import X from '../parent'             → parent path resolution
//   import X from 'package-name'          → node_modules or tsconfig path
//   require('module')                     → CommonJS require
//
// The node_modules lookup is limited to files that were analyzed as
// part of the workspace; external packages remain as-is.

var (
	// import ... from 'path'
	tsImportFromRe = regexp.MustCompile(`^\s*import\s+.*?from\s+['"]([^'"]+)['"]`)
	// import 'path'  (bare side-effect import)
	tsImportBareRe = regexp.MustCompile(`^\s*import\s+['"]([^'"]+)['"]`)
	// const X = require('path')
	tsRequireRe = regexp.MustCompile(`require\s*\(\s*['"]([^'"]+)['"]\s*\)`)
)

// tsconfig holds the subset of tsconfig.json we care about.
type tsconfig struct {
	CompilerOptions struct {
		BaseURL string              `json:"baseUrl"`
		Paths   map[string][]string `json:"paths"`
	} `json:"compilerOptions"`
}

// tsModuleIndex maps a relative module path (from a source file) to the
// concrete file that provides it.
type tsModuleIndex struct {
	// byCleanPath: canonical "./foo/bar" from project root → file
	byCleanPath map[string]string
	// baseURL and paths from tsconfig.json, if present
	tsconfig *tsconfig
	rootDir  string
}

func buildTSModuleIndex(absPath string) *tsModuleIndex {
	idx := &tsModuleIndex{
		byCleanPath: make(map[string]string),
		rootDir:     absPath,
	}

	// Load tsconfig.json if present
	tsconfigPath := filepath.Join(absPath, "tsconfig.json")
	if data, err := os.ReadFile(tsconfigPath); err == nil {
		var cfg tsconfig
		if err := json.Unmarshal(data, &cfg); err == nil {
			idx.tsconfig = &cfg
		}
	}

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
		if ext != ".ts" && ext != ".tsx" && ext != ".js" && ext != ".jsx" {
			return nil
		}
		rel, _ := filepath.Rel(absPath, path)
		rel = filepath.ToSlash(rel)
		// Strip extension so "foo.ts" and "foo.tsx" and "foo/index.ts"
		// all resolve from the same path
		noExt := rel
		noExt = strings.TrimSuffix(noExt, ".tsx")
		noExt = strings.TrimSuffix(noExt, ".ts")
		noExt = strings.TrimSuffix(noExt, ".jsx")
		noExt = strings.TrimSuffix(noExt, ".js")
		noExt = strings.TrimSuffix(noExt, "/index")
		idx.byCleanPath[noExt] = path
		return nil
	})
	return idx
}

// resolveImport returns the file that a TS import specifier refers to,
// or "" if the specifier does not resolve within the workspace.
func (idx *tsModuleIndex) resolveImport(fromFile, specifier string) string {
	// Relative import
	if strings.HasPrefix(specifier, "./") || strings.HasPrefix(specifier, "../") {
		base := filepath.Dir(fromFile)
		full := filepath.Join(base, specifier)
		rel, _ := filepath.Rel(idx.rootDir, full)
		rel = filepath.ToSlash(rel)
		rel = strings.TrimSuffix(rel, ".tsx")
		rel = strings.TrimSuffix(rel, ".ts")
		rel = strings.TrimSuffix(rel, ".jsx")
		rel = strings.TrimSuffix(rel, ".js")
		if f, ok := idx.byCleanPath[rel]; ok {
			return f
		}
		// Try with /index appended
		if f, ok := idx.byCleanPath[rel+"/index"]; ok {
			return f
		}
		return ""
	}

	// tsconfig paths mapping
	if idx.tsconfig != nil && idx.tsconfig.CompilerOptions.BaseURL != "" {
		baseURL := idx.tsconfig.CompilerOptions.BaseURL
		for pattern, targets := range idx.tsconfig.CompilerOptions.Paths {
			// Support single-wildcard patterns like "@app/*"
			if strings.Contains(pattern, "*") {
				prefix := strings.TrimSuffix(pattern, "*")
				if strings.HasPrefix(specifier, prefix) {
					rest := strings.TrimPrefix(specifier, prefix)
					for _, target := range targets {
						resolved := strings.ReplaceAll(target, "*", rest)
						candidate := filepath.ToSlash(filepath.Join(baseURL, resolved))
						if f, ok := idx.byCleanPath[candidate]; ok {
							return f
						}
					}
				}
			} else if pattern == specifier {
				for _, target := range targets {
					candidate := filepath.ToSlash(filepath.Join(baseURL, target))
					if f, ok := idx.byCleanPath[candidate]; ok {
						return f
					}
				}
			}
		}
	}

	// Non-relative — likely a package. Not in this workspace index.
	return ""
}

// ResolveTypeScriptImports walks a TypeScript project and returns
// IMPORTS edges with TS_IMPORT resolution method.
func ResolveTypeScriptImports(absPath string, fileEntityByPath map[string]uuid.UUID) ([]Relationship, error) {
	idx := buildTSModuleIndex(absPath)
	var rels []Relationship

	err := filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
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
		if ext != ".ts" && ext != ".tsx" && ext != ".js" && ext != ".jsx" {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()

		relPath := relPathFrom(absPath, path)
		scanner := bufio.NewScanner(f)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
		lineNo := 0
		for scanner.Scan() {
			lineNo++
			line := scanner.Text()
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "*") {
				continue
			}

			var specifiers []string
			if m := tsImportFromRe.FindStringSubmatch(line); m != nil {
				specifiers = append(specifiers, m[1])
			}
			if m := tsImportBareRe.FindStringSubmatch(line); m != nil {
				specifiers = append(specifiers, m[1])
			}
			for _, m := range tsRequireRe.FindAllStringSubmatch(line, -1) {
				if len(m) > 1 {
					specifiers = append(specifiers, m[1])
				}
			}

			for _, spec := range specifiers {
				targetFile := idx.resolveImport(path, spec)
				if targetFile == "" {
					continue
				}
				targetRel := relPathFrom(absPath, targetFile)
				rels = append(rels, Relationship{
					From:             tsEntityNameFromPath(relPath),
					To:               tsEntityNameFromPath(targetRel),
					Type:             string(RelImports),
					Confidence:       0.95,
					ResolutionStatus: "RESOLVED",
					ResolutionMethod: "TS_IMPORT",
					EpistemicClass:   "OBSERVATION",
					Evidence: Evidence{
						File:      relPath,
						Line:      lineNo,
						LineStart: lineNo,
						LineEnd:   lineNo,
						Analyzer:  "typescript/import-resolution",
					},
				})
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("typescript import walk: %w", err)
	}
	return rels, nil
}

// tsEntityNameFromPath converts a relative file path to the hierarchical
// dotted name the TypeScript analyzer uses for entity names.
//
//	src/client.ts         → src.client
//	packages/client/x.tsx → packages.client.x
//	foo/index.ts          → foo
//
// This matches the naming convention in internal/ast/typescript/integration.go.
func tsEntityNameFromPath(relPath string) string {
	rel := filepath.ToSlash(relPath)
	rel = strings.TrimSuffix(rel, ".tsx")
	rel = strings.TrimSuffix(rel, ".ts")
	rel = strings.TrimSuffix(rel, ".jsx")
	rel = strings.TrimSuffix(rel, ".js")
	rel = strings.TrimSuffix(rel, "/index")
	return strings.ReplaceAll(rel, "/", ".")
}
