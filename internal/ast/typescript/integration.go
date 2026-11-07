// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package typescript

import (
	"context"
	"crypto/sha256"
	"encoding/json"
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
	FromID           string
	ToID             string
	ClaimType        string // IMPORTS | INHERITS | IMPLEMENTS | CALLS
	Confidence       float64
	ResolutionStatus string // RESOLVED | AMBIGUOUS
	ResolutionMethod string // TS_IMPORT | AST_EXACT | HEURISTIC
	FilePath         string
	Line             int
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

type tsConfigPaths struct {
	CompilerOptions struct {
		BaseURL string              `json:"baseUrl"`
		Paths   map[string][]string `json:"paths"`
	} `json:"compilerOptions"`
}

// AnalyzeDirectory parses all TypeScript files under rootPath and resolves
// cross-file imports (including tsconfig.json paths & barrel index.ts re-exports),
// interface implementations, class inheritance, constructor/field types,
// return-type propagation, and method/function calls.
func AnalyzeDirectory(rootPath string) (*Result, error) {
	result := &Result{}
	parser := NewParser()
	defer parser.Close()

	var tsCfg *tsConfigPaths
	if rawCfg, err := os.ReadFile(filepath.Join(rootPath, "tsconfig.json")); err == nil {
		var parsedCfg tsConfigPaths
		if json.Unmarshal(rawCfg, &parsedCfg) == nil {
			tsCfg = &parsedCfg
		}
	}

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

	moduleIDs := make(map[string]string)       // moduleName → moduleEntityID
	qualifiedIDs := make(map[string]string)    // fullyQualified → entityID
	classIDs := make(map[string]string)        // bare class name → entityID
	classModule := make(map[string]string)     // bare class name → moduleName
	ifaceIDs := make(map[string]string)        // bare interface name → entityID
	moduleByPath := make(map[string]string)    // absolute file path → module name
	funcReturnTypes := make(map[string]string) // qualified fn/method → unwrapped ReturnType

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
			classModule[cls.Name] = moduleName
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
				qMethod := moduleName + "." + cls.Name + "." + mth.Name
				qualifiedIDs[qMethod] = methodID
				if mth.ReturnType != "" {
					funcReturnTypes[qMethod] = mth.ReturnType
					funcReturnTypes[cls.Name+"."+mth.Name] = mth.ReturnType
				}
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
			qFn := moduleName + "." + fn.Name
			qualifiedIDs[qFn] = fnID
			if fn.ReturnType != "" {
				funcReturnTypes[qFn] = fn.ReturnType
				funcReturnTypes[fn.Name] = fn.ReturnType
			}
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

	// Pass 1.5: Propagate barrel re-exports (`export * from './repo'`, `export { OrderRepo as Repo } from './repo'`)
	for filePath, m := range parsedModules {
		for _, reExp := range m.ReExports {
			targetMod := resolveTSImportTarget(rootPath, filepath.Dir(filePath), reExp.Source, moduleByPath, moduleIDs, tsCfg)
			if targetMod == "" {
				continue
			}
			if modID, ok := moduleIDs[m.Module]; ok {
				if targetModID, ok2 := moduleIDs[targetMod]; ok2 {
					result.Relationships = append(result.Relationships, RelationRecord{
						FromID:           modID,
						ToID:             targetModID,
						ClaimType:        "IMPORTS",
						Confidence:       1.0,
						ResolutionStatus: "RESOLVED",
						ResolutionMethod: "TS_IMPORT",
						FilePath:         filePath,
						Line:             reExp.Line,
					})
				}
			}
			if reExp.Namespace == "*" {
				prefix := targetMod + "."
				for k, id := range qualifiedIDs {
					if strings.HasPrefix(k, prefix) {
						suffix := strings.TrimPrefix(k, prefix)
						dstKey := m.Module + "." + suffix
						if _, exists := qualifiedIDs[dstKey]; !exists {
							qualifiedIDs[dstKey] = id
						}
					}
				}
			} else {
				for localAlias, origName := range reExp.NameAliases {
					srcQ := targetMod + "." + origName
					dstQ := m.Module + "." + localAlias
					if id, ok := qualifiedIDs[srcQ]; ok {
						qualifiedIDs[dstQ] = id
						prefix := srcQ + "."
						for k, mID := range qualifiedIDs {
							if strings.HasPrefix(k, prefix) {
								mSuffix := strings.TrimPrefix(k, prefix)
								qualifiedIDs[dstQ+"."+mSuffix] = mID
							}
						}
					}
				}
			}
		}
	}

	// Pass 2: Extract relationships
	for filePath, m := range parsedModules {
		moduleName := m.Module
		moduleID := moduleIDs[moduleName]

		importedNamespaces := make(map[string]string) // localNS → targetModule
		importedSymbols := make(map[string]string)    // localAlias → targetModule + "." + exportedName

		for _, imp := range m.Imports {
			targetModule := resolveTSImportTarget(rootPath, filepath.Dir(filePath), imp.Source, moduleByPath, moduleIDs, tsCfg)

			if targetModule != "" {
				if targetID, ok := moduleIDs[targetModule]; ok {
					result.Relationships = append(result.Relationships, RelationRecord{
						FromID:           moduleID,
						ToID:             targetID,
						ClaimType:        "IMPORTS",
						Confidence:       1.0,
						ResolutionStatus: "RESOLVED",
						ResolutionMethod: "TS_IMPORT",
						FilePath:         filePath,
						Line:             imp.Line,
					})
				}
				if imp.Namespace != "" {
					importedNamespaces[imp.Namespace] = targetModule
				}
				if imp.Default != "" {
					importedNamespaces[imp.Default] = targetModule
					importedSymbols[imp.Default] = targetModule + "." + imp.Default
				}
				for localAlias, origName := range imp.NameAliases {
					importedSymbols[localAlias] = targetModule + "." + origName
				}
				for _, origName := range imp.Names {
					if _, exists := importedSymbols[origName]; !exists {
						importedSymbols[origName] = targetModule + "." + origName
					}
				}
			}
		}

		resolveTypeBinding := func(callerClass, rawRHS string) string {
			if rawRHS == "" {
				return ""
			}
			if strings.HasPrefix(rawRHS, "this.") && callerClass != "" {
				mName := strings.TrimPrefix(rawRHS, "this.")
				if rt, ok := funcReturnTypes[moduleName+"."+callerClass+"."+mName]; ok && rt != "" {
					return rt
				}
			}
			if qSym, ok := importedSymbols[rawRHS]; ok {
				if rt, hasRT := funcReturnTypes[qSym]; hasRT && rt != "" {
					return rt
				}
			}
			if rt, ok := funcReturnTypes[moduleName+"."+rawRHS]; ok && rt != "" {
				return rt
			}
			if rt, ok := funcReturnTypes[rawRHS]; ok && rt != "" {
				return rt
			}
			return rawRHS
		}

		classFieldTypes := make(map[string]map[string]string)
		classResolvedExtends := make(map[string]string) // cls.Name → qualified base class (e.g. "db.BaseRepo")
		for _, cls := range m.Classes {
			ft := make(map[string]string)
			for _, prop := range cls.Properties {
				if prop.Name != "" && prop.Type != "" {
					ft["this."+prop.Name] = prop.Type
				}
			}
			for _, mth := range cls.Methods {
				for k, v := range mth.LocalTypes {
					if strings.HasPrefix(k, "this.") {
						ft[k] = resolveTypeBinding(cls.Name, v)
					}
				}
			}
			classFieldTypes[cls.Name] = ft

			classID := qualifiedIDs[moduleName+"."+cls.Name]
			if cls.Extends != "" {
				if qSym, ok := importedSymbols[cls.Extends]; ok {
					classResolvedExtends[cls.Name] = qSym
					if baseID, ok := qualifiedIDs[qSym]; ok {
						result.Relationships = append(result.Relationships, RelationRecord{
							FromID:           classID,
							ToID:             baseID,
							ClaimType:        "INHERITS",
							Confidence:       0.95,
							ResolutionStatus: "RESOLVED",
							ResolutionMethod: "AST_EXACT",
							FilePath:         filePath,
							Line:             cls.LineStart,
						})
					}
				} else if baseID, ok := qualifiedIDs[moduleName+"."+cls.Extends]; ok {
					classResolvedExtends[cls.Name] = moduleName + "." + cls.Extends
					result.Relationships = append(result.Relationships, RelationRecord{
						FromID:           classID,
						ToID:             baseID,
						ClaimType:        "INHERITS",
						Confidence:       1.0,
						ResolutionStatus: "RESOLVED",
						ResolutionMethod: "AST_EXACT",
						FilePath:         filePath,
						Line:             cls.LineStart,
					})
				} else if baseID, ok := classIDs[cls.Extends]; ok {
					if cMod, ok := classModule[cls.Extends]; ok {
						classResolvedExtends[cls.Name] = cMod + "." + cls.Extends
					}
					result.Relationships = append(result.Relationships, RelationRecord{
						FromID:           classID,
						ToID:             baseID,
						ClaimType:        "INHERITS",
						Confidence:       0.90,
						ResolutionStatus: "RESOLVED",
						ResolutionMethod: "AST_EXACT",
						FilePath:         filePath,
						Line:             cls.LineStart,
					})
				}
			}
			for _, impl := range cls.Implements {
				if qSym, ok := importedSymbols[impl]; ok {
					if ifaceID, ok := qualifiedIDs[qSym]; ok {
						result.Relationships = append(result.Relationships, RelationRecord{
							FromID:           classID,
							ToID:             ifaceID,
							ClaimType:        "IMPLEMENTS",
							Confidence:       1.0,
							ResolutionStatus: "RESOLVED",
							ResolutionMethod: "AST_EXACT",
							FilePath:         filePath,
							Line:             cls.LineStart,
						})
						continue
					}
				}
				if ifaceID, ok := ifaceIDs[impl]; ok {
					result.Relationships = append(result.Relationships, RelationRecord{
						FromID:           classID,
						ToID:             ifaceID,
						ClaimType:        "IMPLEMENTS",
						Confidence:       1.0,
						ResolutionStatus: "RESOLVED",
						ResolutionMethod: "AST_EXACT",
						FilePath:         filePath,
						Line:             cls.LineStart,
					})
				}
			}
		}
		for _, iface := range m.Interfaces {
			ifaceID := qualifiedIDs[moduleName+"."+iface.Name]
			for _, base := range iface.Extends {
				if qSym, ok := importedSymbols[base]; ok {
					if baseID, ok := qualifiedIDs[qSym]; ok {
						result.Relationships = append(result.Relationships, RelationRecord{
							FromID:           ifaceID,
							ToID:             baseID,
							ClaimType:        "INHERITS",
							Confidence:       0.95,
							ResolutionStatus: "RESOLVED",
							ResolutionMethod: "AST_EXACT",
							FilePath:         filePath,
							Line:             iface.LineStart,
						})
						continue
					}
				}
				if baseID, ok := ifaceIDs[base]; ok {
					result.Relationships = append(result.Relationships, RelationRecord{
						FromID:           ifaceID,
						ToID:             baseID,
						ClaimType:        "INHERITS",
						Confidence:       0.90,
						ResolutionStatus: "RESOLVED",
						ResolutionMethod: "AST_EXACT",
						FilePath:         filePath,
						Line:             iface.LineStart,
					})
				}
			}
		}

		resolveMethodOnType := func(typeName, mName string) string {
			if typeName == "" || mName == "" {
				return ""
			}
			if qClass, ok := importedSymbols[typeName]; ok {
				if tid, ok := qualifiedIDs[qClass+"."+mName]; ok {
					return tid
				}
			}
			if dotT := strings.IndexByte(typeName, '.'); dotT > 0 {
				nsAlias := typeName[:dotT]
				clsName := typeName[dotT+1:]
				if targetMod, ok := importedNamespaces[nsAlias]; ok {
					if tid, ok := qualifiedIDs[targetMod+"."+clsName+"."+mName]; ok {
						return tid
					}
				}
			}
			if tid, ok := qualifiedIDs[moduleName+"."+typeName+"."+mName]; ok {
				return tid
			}
			if tid, ok := qualifiedIDs[typeName+"."+mName]; ok {
				return tid
			}
			if cMod, ok := classModule[typeName]; ok {
				if tid, ok := qualifiedIDs[cMod+"."+typeName+"."+mName]; ok {
					return tid
				}
			}
			return ""
		}

		resolveTSTarget := func(callerClass string, localTypes map[string]string, rawCallee string) (string, float64, string) {
			// 0. Chained call: `this.getRepo().fetchOrder` or `getRepo().fetchOrder`
			if strings.Contains(rawCallee, "().") {
				parts := strings.SplitN(rawCallee, "().", 2)
				producer := parts[0]
				mName := parts[1]
				retType := resolveTypeBinding(callerClass, producer)
				if tid := resolveMethodOnType(retType, mName); tid != "" {
					return tid, 0.95, "AST_EXACT"
				}
			}

			// 1. super.<method>() inside a subclass
			if strings.HasPrefix(rawCallee, "super.") && callerClass != "" {
				mName := strings.TrimPrefix(rawCallee, "super.")
				if qBase := classResolvedExtends[callerClass]; qBase != "" {
					if tid, ok := qualifiedIDs[qBase+"."+mName]; ok {
						return tid, 0.95, "AST_EXACT"
					}
				}
			}

			// 2. Direct this.<method>() or this.<field>.<method>()
			if strings.HasPrefix(rawCallee, "this.") && callerClass != "" {
				rest := strings.TrimPrefix(rawCallee, "this.")
				if !strings.Contains(rest, ".") {
					if tid, ok := qualifiedIDs[moduleName+"."+callerClass+"."+rest]; ok {
						return tid, 0.95, "AST_EXACT"
					}
					if qBase := classResolvedExtends[callerClass]; qBase != "" {
						if tid, ok := qualifiedIDs[qBase+"."+rest]; ok {
							return tid, 0.95, "AST_EXACT"
						}
					}
				} else {
					parts := strings.SplitN(rest, ".", 2)
					fieldKey := "this." + parts[0]
					mName := parts[1]
					typeName := resolveTypeBinding(callerClass, localTypes[fieldKey])
					if typeName == "" && classFieldTypes[callerClass] != nil {
						typeName = classFieldTypes[callerClass][fieldKey]
					}
					if tid := resolveMethodOnType(typeName, mName); tid != "" {
						return tid, 0.95, "AST_EXACT"
					}
				}
			}

			// 3. Dotted call: `ns.fn()`, `ClassAlias.staticFn()`, or `localInstance.method()`
			if dotIdx := strings.IndexByte(rawCallee, '.'); dotIdx > 0 {
				prefix := rawCallee[:dotIdx]
				member := rawCallee[dotIdx+1:]

				if rawVarType, ok := localTypes[prefix]; ok {
					varType := resolveTypeBinding(callerClass, rawVarType)
					if tid := resolveMethodOnType(varType, member); tid != "" {
						return tid, 0.95, "AST_EXACT"
					}
				}

				if targetMod, ok := importedNamespaces[prefix]; ok {
					if tid, ok := qualifiedIDs[targetMod+"."+member]; ok {
						return tid, 0.95, "AST_EXACT"
					}
					if tid, ok := qualifiedIDs[targetMod+"."+member+".constructor"]; ok {
						return tid, 0.95, "AST_EXACT"
					}
				}

				if targetSym, ok := importedSymbols[prefix]; ok {
					if tid, ok := qualifiedIDs[targetSym+"."+member]; ok {
						return tid, 0.95, "AST_EXACT"
					}
				}
			}

			// 4. Named import or alias call: `import { foo as bar } from './mod'; bar()`
			if targetSym, ok := importedSymbols[rawCallee]; ok {
				if tid, ok := qualifiedIDs[targetSym]; ok {
					return tid, 0.95, "AST_EXACT"
				}
				if tid, ok := qualifiedIDs[targetSym+".constructor"]; ok {
					return tid, 0.95, "AST_EXACT"
				}
			}

			// 5. Same-module function/class
			if tid, ok := qualifiedIDs[moduleName+"."+rawCallee]; ok {
				return tid, 0.95, "AST_EXACT"
			}
			if tid, ok := qualifiedIDs[rawCallee]; ok {
				return tid, 0.95, "AST_EXACT"
			}
			return "", 0, ""
		}

		for _, fn := range m.Functions {
			fromID := qualifiedIDs[moduleName+"."+fn.Name]
			seen := make(map[string]bool)
			for idx, callee := range fn.Calls {
				callLine := fn.LineStart
				if idx < len(fn.CallSites) && fn.CallSites[idx].Line > 0 {
					callLine = fn.CallSites[idx].Line
				}
				if tid, conf, method := resolveTSTarget("", fn.LocalTypes, callee); tid != "" && !seen[tid] {
					seen[tid] = true
					result.Relationships = append(result.Relationships, RelationRecord{
						FromID:           fromID,
						ToID:             tid,
						ClaimType:        "CALLS",
						Confidence:       conf,
						ResolutionStatus: "RESOLVED",
						ResolutionMethod: method,
						FilePath:         filePath,
						Line:             callLine,
					})
				}
			}
		}
		for _, cls := range m.Classes {
			for _, mth := range cls.Methods {
				fromID := qualifiedIDs[moduleName+"."+cls.Name+"."+mth.Name]
				seen := make(map[string]bool)
				for idx, callee := range mth.Calls {
					callLine := mth.LineStart
					if idx < len(mth.CallSites) && mth.CallSites[idx].Line > 0 {
						callLine = mth.CallSites[idx].Line
					}
					if tid, conf, method := resolveTSTarget(cls.Name, mth.LocalTypes, callee); tid != "" && !seen[tid] {
						seen[tid] = true
						result.Relationships = append(result.Relationships, RelationRecord{
							FromID:           fromID,
							ToID:             tid,
							ClaimType:        "CALLS",
							Confidence:       conf,
							ResolutionStatus: "RESOLVED",
							ResolutionMethod: method,
							FilePath:         filePath,
							Line:             callLine,
						})
					}
				}
			}
		}
	}

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

func resolveTSImportTarget(rootPath, currentDir, importPath string, moduleByPath, moduleIDs map[string]string, tsCfg *tsConfigPaths) string {
	if strings.HasPrefix(importPath, ".") {
		return resolveRelativeImport(currentDir, importPath, moduleByPath)
	}
	if tsCfg != nil {
		baseDir := rootPath
		if tsCfg.CompilerOptions.BaseURL != "" {
			baseDir = filepath.Join(rootPath, tsCfg.CompilerOptions.BaseURL)
		}
		for pattern, targets := range tsCfg.CompilerOptions.Paths {
			if strings.Contains(pattern, "*") {
				prefix := strings.TrimSuffix(pattern, "*")
				if strings.HasPrefix(importPath, prefix) {
					rest := strings.TrimPrefix(importPath, prefix)
					for _, t := range targets {
						resolved := strings.ReplaceAll(t, "*", rest)
						if mod := resolveRelativeImport(baseDir, "./"+resolved, moduleByPath); mod != "" {
							return mod
						}
					}
				}
			} else if pattern == importPath {
				for _, t := range targets {
					if mod := resolveRelativeImport(baseDir, "./"+t, moduleByPath); mod != "" {
						return mod
					}
				}
			}
		}
		if tsCfg.CompilerOptions.BaseURL != "" {
			if mod := resolveRelativeImport(baseDir, "./"+importPath, moduleByPath); mod != "" {
				return mod
			}
		}
	}
	// Fallback check against workspace modules
	dotted := strings.ReplaceAll(strings.TrimPrefix(importPath, "@/"), "/", ".")
	if _, ok := moduleIDs[dotted]; ok {
		return dotted
	}
	if _, ok := moduleIDs["src."+dotted]; ok {
		return "src." + dotted
	}
	return importPath
}

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
