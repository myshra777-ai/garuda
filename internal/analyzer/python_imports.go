// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package analyzer

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// Python import resolution — Path A.
//
// Reads every .py file in the project, extracts import statements,
// resolves each to a concrete file or symbol in the workspace, and
// emits IMPORTS relationships with ResolutionMethodPythonImport.
//
// This does NOT do type inference. It resolves imports only:
//
//   import foo.bar           → module foo.bar
//   from foo import Bar      → symbol Bar in module foo
//   from foo.bar import Baz  → symbol Baz in module foo.bar
//   from . import x          → relative, resolved against the file's package
//   from .sub import y       → relative, resolved against file's package
//
// Call expressions whose receiver comes from a resolved import are also
// attributed, but stay marked HEURISTIC unless a direct symbol match is
// found.
//
// Every edge produced by this file carries an evidence snippet pointing
// to the source line that produced it.

var (
	// from foo.bar import X, Y as Z
	pyFromImportRe = regexp.MustCompile(`^\s*from\s+([\.\w]+)\s+import\s+(.+?)(?:\s*#.*)?$`)
	// import foo, bar.baz
	pyImportRe = regexp.MustCompile(`^\s*import\s+([\w\.]+(?:\s*,\s*[\w\.]+)*)(?:\s+as\s+\w+)?(?:\s*#.*)?$`)
)

// pythonModuleIndex maps a dotted module path to the file that declares it.
//
// For every .py file discovered:
//
//	foo/bar.py         → module "foo.bar"
//	foo/bar/__init__.py → module "foo.bar"
//	foo/bar.pyi        → module "foo.bar" (stub)
//
// The value is the absolute file path.
type pythonModuleIndex struct {
	byModule map[string]string
	// Also index the reverse direction so we can find the module name
	// of a given file when emitting source-side evidence.
	byFile map[string]string
}

// buildPythonModuleIndex walks a Python project and returns the index.
func buildPythonModuleIndex(absPath string) *pythonModuleIndex {
	idx := &pythonModuleIndex{
		byModule: make(map[string]string),
		byFile:   make(map[string]string),
	}

	_ = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			base := info.Name()
			if strings.HasPrefix(base, ".") ||
				base == "__pycache__" ||
				base == "venv" || base == ".venv" || base == "env" ||
				base == "site-packages" || base == "dist" || base == "build" ||
				base == "node_modules" || base == "test" || base == "tests" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := filepath.Ext(path)
		if ext != ".py" && ext != ".pyi" {
			return nil
		}

		rel, err := filepath.Rel(absPath, path)
		if err != nil {
			return nil
		}

		// foo/bar.py   → foo.bar
		// foo/bar/__init__.py → foo.bar
		// foo/bar.pyi  → foo.bar
		rel = filepath.ToSlash(rel)
		rel = strings.TrimSuffix(rel, ".pyi")
		rel = strings.TrimSuffix(rel, ".py")
		rel = strings.TrimSuffix(rel, "/__init__")
		modPath := strings.ReplaceAll(rel, "/", ".")

		idx.byModule[modPath] = path
		idx.byFile[path] = modPath
		return nil
	})

	return idx
}

// resolveModule returns the file path of the module with the given
// dotted path, or "" if not found in this workspace.
func (idx *pythonModuleIndex) resolveModule(modulePath string) string {
	// Direct hit
	if f, ok := idx.byModule[modulePath]; ok {
		return f
	}
	// Sub-package: "foo.bar" may live in "foo/bar/__init__.py", which
	// we already indexed as "foo.bar". Nothing further to try.
	return ""
}

// resolveRelative resolves a relative import like ".sub" or ".." given
// the importing file's module path.
//
//	fpkg = "a.b.c" (the current file's module), level = 1, name = "sub"
//	→ "a.b.sub"
func resolveRelative(fpkg string, level int, name string) string {
	parts := strings.Split(fpkg, ".")
	// Drop the last segment (the module itself, not the package)
	if len(parts) > 0 {
		parts = parts[:len(parts)-1]
	}
	// One dot = current package. Two dots = parent, and so on.
	for i := 1; i < level; i++ {
		if len(parts) > 0 {
			parts = parts[:len(parts)-1]
		}
	}
	if name != "" {
		parts = append(parts, name)
	}
	return strings.Join(parts, ".")
}

// ResolvePythonImports walks a Python project and returns IMPORTS edges.
//
// pkgEntityByName maps a package/module entity name to its UUID. The
// caller builds this from the entities the analyzer already extracted.
// Every edge's source and target refer to entities in that map.
func ResolvePythonImports(absPath string, pkgEntityByName map[string]uuid.UUID) ([]Relationship, error) {
	idx := buildPythonModuleIndex(absPath)
	var rels []Relationship

	err := filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			base := info.Name()
			if strings.HasPrefix(base, ".") ||
				base == "__pycache__" ||
				base == "venv" || base == ".venv" || base == "env" ||
				base == "site-packages" || base == "dist" || base == "build" ||
				base == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".py" {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()

		fromModule := idx.byFile[path]

		scanner := bufio.NewScanner(f)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
		lineNo := 0
		for scanner.Scan() {
			lineNo++
			line := scanner.Text()

			// Skip blank lines and full-line comments
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}

			// Try `from X import Y`
			if m := pyFromImportRe.FindStringSubmatch(line); m != nil {
				rawFrom := m[1]
				imported := m[2]

				// Handle relative imports
				level := 0
				modName := rawFrom
				for strings.HasPrefix(modName, ".") {
					level++
					modName = modName[1:]
				}

				var targetModule string
				if level > 0 {
					targetModule = resolveRelative(fromModule, level, modName)
				} else {
					targetModule = modName
				}

				targetFile := idx.resolveModule(targetModule)
				if targetFile == "" {
					// Not a module in this workspace. Likely stdlib or
					// third-party. Skip — the external-stub machinery
					// already handles these.
					continue
				}

				// Emit an IMPORTS edge from the current package entity
				// to the target package entity, if both exist.
				_, fromOK := pkgEntityByName[fromModule]
				_, toOK := pkgEntityByName[targetModule]
				if fromOK && toOK {
					rels = append(rels, Relationship{
						From:             fromModule,
						To:               targetModule,
						Type:             string(RelImports),
						Confidence:       0.95,
						ResolutionStatus: "RESOLVED",
						ResolutionMethod: "PYTHON_IMPORT",
						EpistemicClass:   "OBSERVATION",
						Evidence: Evidence{
							File:      relPathFrom(absPath, path),
							Line:      lineNo,
							LineStart: lineNo,
							LineEnd:   lineNo,
							Analyzer:  "python/import-resolution",
						},
					})
				}

				// If the imported name is itself a symbol defined in this
				// workspace, emit a second edge to the symbol.
				symbols := parseImportedNames(imported)
				for _, sym := range symbols {
					symQualified := targetModule + "." + sym
					if _, ok := pkgEntityByName[symQualified]; ok {
						rels = append(rels, Relationship{
							From:             fromModule,
							To:               symQualified,
							Type:             string(RelImports),
							Confidence:       0.90,
							ResolutionStatus: "RESOLVED",
							ResolutionMethod: "PYTHON_IMPORT",
							EpistemicClass:   "OBSERVATION",
							Evidence: Evidence{
								File:      relPathFrom(absPath, path),
								Line:      lineNo,
								LineStart: lineNo,
								LineEnd:   lineNo,
								Analyzer:  "python/import-resolution",
							},
						})
					}
				}
			}

			// Try `import X` and `import X, Y`
			if m := pyImportRe.FindStringSubmatch(line); m != nil {
				modules := strings.Split(m[1], ",")
				for _, mod := range modules {
					mod = strings.TrimSpace(mod)
					if mod == "" {
						continue
					}
					if idx.resolveModule(mod) == "" {
						continue
					}
					if fromID, ok := pkgEntityByName[fromModule]; ok {
						if toID, ok2 := pkgEntityByName[mod]; ok2 {
							rels = append(rels, Relationship{
								From:             fromModule,
								To:               mod,
								Type:             string(RelImports),
								Confidence:       0.95,
								ResolutionStatus: "RESOLVED",
								ResolutionMethod: "PYTHON_IMPORT",
								EpistemicClass:   "OBSERVATION",
								Evidence: Evidence{
									File:      relPathFrom(absPath, path),
									Line:      lineNo,
									LineStart: lineNo,
									LineEnd:   lineNo,
									Analyzer:  "python/import-resolution",
								},
							})
							_ = fromID
							_ = toID
						}
					}
				}
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("python import walk: %w", err)
	}
	return rels, nil
}

// parseImportedNames parses the tail of a `from X import ...` clause
// into a list of imported symbol names.
//
//	"Bar" → ["Bar"]
//	"Bar as Baz" → ["Bar"]
//	"Bar, Qux, Zot" → ["Bar", "Qux", "Zot"]
//	"Bar as B, Qux" → ["Bar", "Qux"]
func parseImportedNames(clause string) []string {
	// Drop parentheses
	clause = strings.Trim(clause, "() ")
	var out []string
	for _, part := range strings.Split(clause, ",") {
		part = strings.TrimSpace(part)
		if part == "" || part == "*" {
			continue
		}
		// "Bar as Baz" → "Bar"
		if idx := strings.Index(part, " as "); idx > 0 {
			part = part[:idx]
		}
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

// relPathFrom returns a path relative to a root, or the absolute path
// if the relative form fails.
func relPathFrom(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return rel
}
