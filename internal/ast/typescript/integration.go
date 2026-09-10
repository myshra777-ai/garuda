// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package typescript

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/uuid"
)

// EntityRecord is the analyzer-agnostic shape consumed by the store layer.
type EntityRecord struct {
	ID         string
	Name       string
	Kind       string // class | interface | function | method | import | module
	Package    string
	FilePath   string
	LineStart  int
	LineEnd    int
	IsExported bool
	Signature  string
	Language   string
	Hash       string
}

type RelationRecord struct {
	FromID     string
	ToID       string
	ClaimType  string // IMPORTS | INHERITS | IMPLEMENTS
	Confidence float64
}

type Result struct {
	Entities      []EntityRecord
	Relationships []RelationRecord
	Fingerprint   string
	Stats         Stats
}

type Stats struct {
	Files      int
	Classes    int
	Interfaces int
	Functions  int
	Methods    int
	Imports    int
}

// AnalyzeDirectory parses all TypeScript files under rootPath.
func AnalyzeDirectory(rootPath string) (*Result, error) {
	result := &Result{}
	parser := NewParser()
	defer parser.Close()

	var files []string
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			base := info.Name()
			if strings.HasPrefix(base, ".") ||
				base == "node_modules" || base == "dist" ||
				base == "build" || base == "out" || base == "coverage" ||
				base == ".next" || base == ".nuxt" || base == ".turbo" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := filepath.Ext(path)
		if ext == ".ts" || ext == ".tsx" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	moduleIDs := make(map[string]string)    // moduleName → moduleEntityID
	qualifiedIDs := make(map[string]string) // fullyQualified → entityID
	classIDs := make(map[string]string)     // bare class name → entityID
	ifaceIDs := make(map[string]string)     // bare interface name → entityID
	moduleByPath := make(map[string]string) // absolute file path → module name

	ctx := context.Background()
	parsedModules := make(map[string]*Module)

	for _, filePath := range files {
		src, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}
		moduleName := deriveModuleName(rootPath, filePath)
		m, err := parser.Parse(ctx, src, filePath, moduleName)
		if err != nil {
			continue
		}
		parsedModules[filePath] = m
		moduleByPath[filePath] = moduleName
		result.Stats.Files++

		modID := deterministicID("typescript", "module", moduleName, filePath, 0)
		moduleIDs[moduleName] = modID
		result.Entities = append(result.Entities, EntityRecord{
			ID:         modID,
			Name:       moduleName,
			Kind:       "module",
			Package:    moduleName,
			FilePath:   filePath,
			LineStart:  1,
			LineEnd:    strings.Count(string(src), "\n") + 1,
			IsExported: !strings.HasPrefix(filepath.Base(filePath), "_"),
			Language:   "typescript",
			Hash:       hashOf(src),
		})

		for _, cls := range m.Classes {
			classID := deterministicID("typescript", "class", moduleName+"."+cls.Name, filePath, cls.LineStart)
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
				Language:   "typescript",
			})
			result.Stats.Classes++

			for _, mth := range cls.Methods {
				methodID := deterministicID("typescript", "method", moduleName+"."+cls.Name+"."+mth.Name, filePath, mth.LineStart)
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
					Language:   "typescript",
				})
				result.Stats.Methods++
			}
		}

		for _, iface := range m.Interfaces {
			ifaceID := deterministicID("typescript", "interface", moduleName+"."+iface.Name, filePath, iface.LineStart)
			qualifiedIDs[moduleName+"."+iface.Name] = ifaceID
			ifaceIDs[iface.Name] = ifaceID
			result.Entities = append(result.Entities, EntityRecord{
				ID:         ifaceID,
				Name:       iface.Name,
				Kind:       "interface",
				Package:    moduleName,
				FilePath:   filePath,
				LineStart:  iface.LineStart,
				LineEnd:    iface.LineEnd,
				IsExported: iface.IsExported,
				Signature:  interfaceSignature(iface),
				Language:   "typescript",
			})
			result.Stats.Interfaces++
		}

		for _, fn := range m.Functions {
			fnID := deterministicID("typescript", "function", moduleName+"."+fn.Name, filePath, fn.LineStart)
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
				Language:   "typescript",
			})
			result.Stats.Functions++
		}

		for _, imp := range m.Imports {
			impID := deterministicID("typescript", "import", moduleName+"::"+imp.Source, filePath, imp.Line)
			result.Entities = append(result.Entities, EntityRecord{
				ID:        impID,
				Name:      imp.Source,
				Kind:      "import",
				Package:   moduleName,
				FilePath:  filePath,
				LineStart: imp.Line,
				LineEnd:   imp.Line,
				Language:  "typescript",
			})
			result.Stats.Imports++
		}
	}

	// Relations
	for filePath, m := range parsedModules {
		moduleName := m.Module
		moduleID := moduleIDs[moduleName]

		for _, imp := range m.Imports {
			var targetModule string

			if strings.HasPrefix(imp.Source, ".") {
				targetModule = resolveRelativeImport(filepath.Dir(filePath), imp.Source, moduleByPath)
			} else {
				targetModule = imp.Source
			}

			if targetModule != "" {
				if targetID, ok := moduleIDs[targetModule]; ok {
					result.Relationships = append(result.Relationships, RelationRecord{
						FromID: moduleID, ToID: targetID,
						ClaimType: "IMPORTS", Confidence: 1.0,
					})
				}
			}
		}

		for _, cls := range m.Classes {
			classID := qualifiedIDs[moduleName+"."+cls.Name]
			if cls.Extends != "" {
				if baseID, ok := classIDs[cls.Extends]; ok {
					result.Relationships = append(result.Relationships, RelationRecord{
						FromID: classID, ToID: baseID,
						ClaimType: "INHERITS", Confidence: 0.9,
					})
				}
			}
			for _, impl := range cls.Implements {
				if ifaceID, ok := ifaceIDs[impl]; ok {
					result.Relationships = append(result.Relationships, RelationRecord{
						FromID: classID, ToID: ifaceID,
						ClaimType: "IMPLEMENTS", Confidence: 1.0,
					})
				}
			}
		}
		for _, iface := range m.Interfaces {
			ifaceID := qualifiedIDs[moduleName+"."+iface.Name]
			for _, base := range iface.Extends {
				if baseID, ok := ifaceIDs[base]; ok {
					result.Relationships = append(result.Relationships, RelationRecord{
						FromID: ifaceID, ToID: baseID,
						ClaimType: "INHERITS", Confidence: 0.9,
					})
				}
			}
		}
	}

	// Determinism
	sort.Slice(result.Entities, func(i, j int) bool { return result.Entities[i].ID < result.Entities[j].ID })
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

	result.Fingerprint = computeFingerprint(result)
	return result, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

// resolveRelativeImport turns "./foo" or "../bar/baz" into the canonical
// module name used by moduleByPath, or returns "" if unresolved.
func resolveRelativeImport(currentDir, importPath string, moduleByPath map[string]string) string {
	for _, ext := range []string{".tsx", ".ts", ".jsx", ".js"} {
		importPath = strings.TrimSuffix(importPath, ext)
	}
	joined := filepath.Join(currentDir, importPath)
	joined = filepath.Clean(joined)

	if mod, ok := moduleByPath[joined+".ts"]; ok {
		return mod
	}
	if mod, ok := moduleByPath[joined+".tsx"]; ok {
		return mod
	}
	if mod, ok := moduleByPath[filepath.Join(joined, "index.ts")]; ok {
		return mod
	}
	if mod, ok := moduleByPath[filepath.Join(joined, "index.tsx")]; ok {
		return mod
	}

	return ""
}

func deriveModuleName(root, filePath string) string {
	rel, err := filepath.Rel(root, filePath)
	if err != nil {
		rel = filePath
	}
	for _, ext := range []string{".tsx", ".ts", ".jsx", ".js"} {
		rel = strings.TrimSuffix(rel, ext)
	}
	rel = strings.ReplaceAll(rel, string(filepath.Separator), ".")
	rel = strings.TrimSuffix(rel, ".index")
	if rel == "index" || rel == "" {
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

func classSignature(c Class) string {
	parts := []string{}
	if c.IsAbstract {
		parts = append(parts, "abstract ")
	}
	sig := "class " + c.Name
	if c.Extends != "" {
		sig += " extends " + c.Extends
	}
	if len(c.Implements) > 0 {
		sig += " implements " + strings.Join(c.Implements, ", ")
	}
	return strings.Join(parts, "") + sig
}

func interfaceSignature(i Interface) string {
	sig := "interface " + i.Name
	if len(i.Extends) > 0 {
		sig += " extends " + strings.Join(i.Extends, ", ")
	}
	return sig
}

func functionSignature(f Function) string {
	prefix := ""
	if f.IsAsync {
		prefix = "async "
	}
	sig := prefix + "function " + f.Name + "(" + strings.Join(f.Params, ", ") + ")"
	if f.ReturnType != "" {
		sig += ": " + f.ReturnType
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
