// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package python

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/uuid"
)

// EntityRecord mirrors the Go analyzer's shape so both languages land
// in the same semantic graph schema.
type EntityRecord struct {
	ID         string
	Name       string
	Kind       string // class | function | method | import | module
	Package    string // dotted module path
	FilePath   string
	LineStart  int
	LineEnd    int
	IsExported bool
	Signature  string
	Language   string // "python"
	Hash       string
}

type RelationRecord struct {
	FromID     string
	ToID       string
	ClaimType  string // IMPORTS | CALLS | INHERITS
	Confidence float64
}

type Result struct {
	Entities      []EntityRecord
	Relationships []RelationRecord
	Fingerprint   string
	Stats         Stats
}

type Stats struct {
	Files     int
	Classes   int
	Functions int
	Methods   int
	Imports   int
}

// AnalyzeDirectory walks a Python workspace and returns the full result.
func AnalyzeDirectory(rootPath, repoID, workspaceID string) (*Result, error) {
	result := &Result{}
	var moduleFiles []string

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			base := info.Name()
			if strings.HasPrefix(base, ".") || base == "__pycache__" ||
				base == "venv" || base == ".venv" || base == "env" ||
				base == "node_modules" || base == "site-packages" ||
				base == "dist" || base == "build" || base == ".tox" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) == ".py" {
			moduleFiles = append(moduleFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Maps module name → module-level entity ID, for IMPORTS resolution
	moduleIDs := make(map[string]string)
	// Maps class name → entity ID, for INHERITS resolution
	classIDs := make(map[string]string)
	// Maps qualified name (module.Class.method) → entity ID, for CALLS resolution
	qualifiedIDs := make(map[string]string)

	// Pass 1: extract entities
	parsedModules := make(map[string]*Module) // path → module
	for _, filePath := range moduleFiles {
		src, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}
		moduleName := deriveModuleName(rootPath, filePath)
		m, err := Parse(string(src), filePath, moduleName)
		if err != nil {
			continue
		}
		parsedModules[filePath] = m
		result.Stats.Files++

		// Module entity
		modID := deterministicID("python", "module", moduleName, filePath, 0)
		moduleIDs[moduleName] = modID
		result.Entities = append(result.Entities, EntityRecord{
			ID:         modID,
			Name:       moduleName,
			Kind:       "module",
			Package:    moduleName,
			FilePath:   filePath,
			LineStart:  1,
			LineEnd:    countLines(src),
			IsExported: !strings.HasPrefix(filepath.Base(filePath), "_"),
			Signature:  "",
			Language:   "python",
			Hash:       hashOf(src),
		})

		// Classes
		for _, cls := range m.Classes {
			classID := deterministicID("python", "class", moduleName+"."+cls.Name, filePath, cls.LineStart)
			qualifiedIDs[moduleName+"."+cls.Name] = classID
			classIDs[cls.Name] = classID
			result.Entities = append(result.Entities, EntityRecord{
				ID:         classID,
				Name:       cls.Name,
				Kind:       "class",
				Package:    moduleName,
				FilePath:   filePath,
				LineStart:  cls.LineStart,
				LineEnd:    cls.LineEnd,
				IsExported: cls.IsExported,
				Signature:  classSignature(cls),
				Language:   "python",
				Hash:       "",
			})
			result.Stats.Classes++

			// Methods
			for _, mth := range cls.Methods {
				methodID := deterministicID("python", "method", moduleName+"."+cls.Name+"."+mth.Name, filePath, mth.LineStart)
				qualifiedIDs[moduleName+"."+cls.Name+"."+mth.Name] = methodID
				result.Entities = append(result.Entities, EntityRecord{
					ID:         methodID,
					Name:       mth.Name,
					Kind:       "method",
					Package:    moduleName,
					FilePath:   filePath,
					LineStart:  mth.LineStart,
					LineEnd:    mth.LineEnd,
					IsExported: mth.IsExported,
					Signature:  functionSignature(mth),
					Language:   "python",
				})
				result.Stats.Methods++
			}
		}

		// Module-level functions
		for _, fn := range m.Functions {
			fnID := deterministicID("python", "function", moduleName+"."+fn.Name, filePath, fn.LineStart)
			qualifiedIDs[moduleName+"."+fn.Name] = fnID
			result.Entities = append(result.Entities, EntityRecord{
				ID:         fnID,
				Name:       fn.Name,
				Kind:       "function",
				Package:    moduleName,
				FilePath:   filePath,
				LineStart:  fn.LineStart,
				LineEnd:    fn.LineEnd,
				IsExported: fn.IsExported,
				Signature:  functionSignature(fn),
				Language:   "python",
			})
			result.Stats.Functions++
		}

		// Imports — as entities too, so they show up in the graph
		for _, imp := range m.Imports {
			impID := deterministicID("python", "import", moduleName+"::"+imp.Module, filePath, imp.Line)
			result.Entities = append(result.Entities, EntityRecord{
				ID:         impID,
				Name:       imp.Module,
				Kind:       "import",
				Package:    moduleName,
				FilePath:   filePath,
				LineStart:  imp.Line,
				LineEnd:    imp.Line,
				IsExported: false,
				Language:   "python",
			})
			result.Stats.Imports++
		}
	}

	// Pass 2: extract relations
	for filePath, m := range parsedModules {
		moduleName := m.Module
		moduleID := moduleIDs[moduleName]

		// IMPORTS relations: module → target module
		for _, imp := range m.Imports {
			targetModule := imp.Module
			if targetID, ok := moduleIDs[targetModule]; ok {
				result.Relationships = append(result.Relationships, RelationRecord{
					FromID:     moduleID,
					ToID:       targetID,
					ClaimType:  "IMPORTS",
					Confidence: 1.0,
				})
			}
			// External imports — record as edges to a stub? For now, skip.
		}

		// INHERITS relations: class → base class
		for _, cls := range m.Classes {
			classID := qualifiedIDs[moduleName+"."+cls.Name]
			for _, base := range cls.Bases {
				// Try to resolve fully qualified or bare class name
				if baseID, ok := qualifiedIDs[base]; ok {
					result.Relationships = append(result.Relationships, RelationRecord{
						FromID:     classID,
						ToID:       baseID,
						ClaimType:  "INHERITS",
						Confidence: 1.0,
					})
				} else if baseID, ok := classIDs[base]; ok {
					result.Relationships = append(result.Relationships, RelationRecord{
						FromID:     classID,
						ToID:       baseID,
						ClaimType:  "INHERITS",
						Confidence: 0.85,
					})
				}
			}
		}

		_ = filePath
	}

	// Determinism: sort by stable keys
	sort.Slice(result.Entities, func(i, j int) bool {
		return result.Entities[i].ID < result.Entities[j].ID
	})
	sort.Slice(result.Relationships, func(i, j int) bool {
		a, b := result.Relationships[i], result.Relationships[j]
		if a.FromID != b.FromID {
			return a.FromID < b.FromID
		}
		if a.ToID != b.ToID {
			return a.ToID < b.ToID
		}
		return a.ClaimType < b.ClaimType
	})

	// Fingerprint: SHA-256 of sorted entity IDs + relation keys
	result.Fingerprint = computeFingerprint(result)
	return result, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func deriveModuleName(root, filePath string) string {
	rel, err := filepath.Rel(root, filePath)
	if err != nil {
		rel = filePath
	}
	rel = strings.TrimSuffix(rel, ".py")
	rel = strings.ReplaceAll(rel, string(filepath.Separator), ".")
	rel = strings.TrimSuffix(rel, ".__init__")
	if rel == "__init__" || rel == "" {
		return filepath.Base(root)
	}
	return rel
}

func deterministicID(lang, kind, qualified, file string, line int) string {
	data := fmt.Sprintf("%s|%s|%s|%s|%d", lang, kind, qualified, file, line)
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte(data)).String()
}

func hashOf(data []byte) string {
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h[:16])
}

func countLines(data []byte) int {
	return strings.Count(string(data), "\n") + 1
}

func classSignature(c Class) string {
	if len(c.Bases) == 0 {
		return "class " + c.Name
	}
	return "class " + c.Name + "(" + strings.Join(c.Bases, ", ") + ")"
}

func functionSignature(f Function) string {
	prefix := ""
	if f.IsAsync {
		prefix = "async "
	}
	sig := prefix + "def " + f.Name + "(" + strings.Join(f.Params, ", ") + ")"
	if f.ReturnType != "" {
		sig += " -> " + f.ReturnType
	}
	return sig
}

func computeFingerprint(r *Result) string {
	var b strings.Builder
	for _, e := range r.Entities {
		b.WriteString(e.ID)
		b.WriteByte('\n')
	}
	for _, rel := range r.Relationships {
		b.WriteString(rel.FromID)
		b.WriteByte('|')
		b.WriteString(rel.ToID)
		b.WriteByte('|')
		b.WriteString(rel.ClaimType)
		b.WriteByte('\n')
	}
	h := sha256.Sum256([]byte(b.String()))
	return fmt.Sprintf("%x", h[:16])
}
