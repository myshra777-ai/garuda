// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

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
	FromID           string
	ToID             string
	ClaimType        string // IMPORTS | CALLS | INHERITS
	Confidence       float64
	ResolutionStatus string // RESOLVED | AMBIGUOUS
	ResolutionMethod string // PYTHON_IMPORT | AST_EXACT | HEURISTIC
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
	// Maps class name → module name
	classModule := make(map[string]string)
	// Maps qualified name (module.Class.method or module.fn) → entity ID, for CALLS resolution
	qualifiedIDs := make(map[string]string)
	// Maps qualified function/method name → declared ReturnType (e.g. "OrderRepository")
	funcReturnTypes := make(map[string]string)
	// Maps qualified class name (module.Class) → list of raw base class names
	classRawBases := make(map[string][]string)

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
		// Support `src/` layout by also indexing the stripped module path
		if strings.HasPrefix(moduleName, "src.") {
			stripped := strings.TrimPrefix(moduleName, "src.")
			if _, exists := moduleIDs[stripped]; !exists {
				moduleIDs[stripped] = modID
			}
		}
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
			if strings.HasPrefix(moduleName, "src.") {
				qualifiedIDs[strings.TrimPrefix(moduleName, "src.")+"."+cls.Name] = classID
			}
			classIDs[cls.Name] = classID
			classModule[cls.Name] = moduleName
			classRawBases[moduleName+"."+cls.Name] = cls.Bases
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
				if mth.Name == "__init__" && mth.LineStart == cls.LineStart {
					// Synthetic class-attribute holder
					continue
				}
				methodID := deterministicID("python", "method", moduleName+"."+cls.Name+"."+mth.Name, filePath, mth.LineStart)
				qMethod := moduleName + "." + cls.Name + "." + mth.Name
				qualifiedIDs[qMethod] = methodID
				if strings.HasPrefix(moduleName, "src.") {
					qualifiedIDs[strings.TrimPrefix(moduleName, "src.")+"."+cls.Name+"."+mth.Name] = methodID
				}
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
					Language:   "python",
				})
				result.Stats.Methods++
			}
		}

		// Module-level functions
		for _, fn := range m.Functions {
			fnID := deterministicID("python", "function", moduleName+"."+fn.Name, filePath, fn.LineStart)
			qFn := moduleName + "." + fn.Name
			qualifiedIDs[qFn] = fnID
			if strings.HasPrefix(moduleName, "src.") {
				qualifiedIDs[strings.TrimPrefix(moduleName, "src.")+"."+fn.Name] = fnID
			}
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

	// Pass 1.5: Propagate __init__.py re-exports (e.g., `from .repo import OrderRepository` in `db/__init__.py`)
	for _, m := range parsedModules {
		for _, imp := range m.Imports {
			targetMod := resolvePythonModuleName(m.Module, imp.Module, moduleIDs)
			for localAlias, origName := range imp.NameAliases {
				srcQ := targetMod + "." + origName
				dstQ := m.Module + "." + localAlias
				if id, ok := qualifiedIDs[srcQ]; ok {
					if _, exists := qualifiedIDs[dstQ]; !exists {
						qualifiedIDs[dstQ] = id
					}
					// Also copy method IDs for re-exported classes
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

	// Pass 2: extract relations
	for filePath, m := range parsedModules {
		moduleName := m.Module
		moduleID := moduleIDs[moduleName]

		importedModules := make(map[string]string)
		importedSymbols := make(map[string]string)

		for _, imp := range m.Imports {
			targetModule := resolvePythonModuleName(moduleName, imp.Module, moduleIDs)
			if targetID, ok := moduleIDs[targetModule]; ok {
				result.Relationships = append(result.Relationships, RelationRecord{
					FromID:           moduleID,
					ToID:             targetID,
					ClaimType:        "IMPORTS",
					Confidence:       1.0,
					ResolutionStatus: "RESOLVED",
					ResolutionMethod: "PYTHON_IMPORT",
					FilePath:         filePath,
					Line:             imp.Line,
				})
			}

			if len(imp.Names) == 0 {
				localMod := imp.Module
				if imp.Alias != "" {
					localMod = imp.Alias
				}
				importedModules[localMod] = targetModule
			} else {
				for localAlias, origName := range imp.NameAliases {
					subMod := targetModule + "." + origName
					if _, isSubMod := moduleIDs[subMod]; isSubMod {
						importedModules[localAlias] = subMod
					} else {
						importedSymbols[localAlias] = targetModule + "." + origName
					}
				}
				for _, origName := range imp.Names {
					if _, exists := importedSymbols[origName]; !exists {
						subMod := targetModule + "." + origName
						if _, isSubMod := moduleIDs[subMod]; isSubMod {
							importedModules[origName] = subMod
						} else {
							importedSymbols[origName] = targetModule + "." + origName
						}
					}
				}
			}
		}

		// Collect class-level attribute types from __init__ and class body annotations,
		// resolving function return types if `self.repo = get_repo()` was used.
		resolveTypeBinding := func(callerClass, rawRHS string) string {
			if rawRHS == "" {
				return ""
			}
			if strings.HasPrefix(rawRHS, "self.") && callerClass != "" {
				mName := strings.TrimPrefix(rawRHS, "self.")
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
		classResolvedBases := make(map[string][]string) // cls.Name → qualified base class names (e.g. "db.BaseRepo")
		for _, cls := range m.Classes {
			ft := make(map[string]string)
			for _, mth := range cls.Methods {
				for k, v := range mth.LocalTypes {
					if strings.HasPrefix(k, "self.") || strings.HasPrefix(k, "cls.") {
						normKey := "self." + strings.TrimPrefix(strings.TrimPrefix(k, "self."), "cls.")
						ft[normKey] = resolveTypeBinding(cls.Name, v)
					}
				}
			}
			classFieldTypes[cls.Name] = ft

			// INHERITS relations: class → base class
			classID := qualifiedIDs[moduleName+"."+cls.Name]
			for _, base := range cls.Bases {
				if qSym, ok := importedSymbols[base]; ok {
					classResolvedBases[cls.Name] = append(classResolvedBases[cls.Name], qSym)
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
						continue
					}
				}
				if dotIdx := strings.IndexByte(base, '.'); dotIdx > 0 {
					modAlias := base[:dotIdx]
					bName := base[dotIdx+1:]
					if targetMod, ok := importedModules[modAlias]; ok {
						qBase := targetMod + "." + bName
						classResolvedBases[cls.Name] = append(classResolvedBases[cls.Name], qBase)
						if baseID, ok := qualifiedIDs[qBase]; ok {
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
							continue
						}
					}
				}
				if baseID, ok := qualifiedIDs[moduleName+"."+base]; ok {
					classResolvedBases[cls.Name] = append(classResolvedBases[cls.Name], moduleName+"."+base)
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
				} else if baseID, ok := classIDs[base]; ok {
					if cMod, ok := classModule[base]; ok {
						classResolvedBases[cls.Name] = append(classResolvedBases[cls.Name], cMod+"."+base)
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
				modAlias := typeName[:dotT]
				clsName := typeName[dotT+1:]
				if targetMod, ok := importedModules[modAlias]; ok {
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

		resolvePyTarget := func(callerClass string, localTypes map[string]string, rawCallee string) (string, float64, string) {
			// 0. Chained call: `self.get_repo().fetch_order` or `get_repo().fetch_order`
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
				for _, qBase := range classResolvedBases[callerClass] {
					if tid, ok := qualifiedIDs[qBase+"."+mName]; ok {
						return tid, 0.95, "AST_EXACT"
					}
				}
			}

			// 2. Direct self.<method>() or cls.<method>() inside a class
			if (strings.HasPrefix(rawCallee, "self.") || strings.HasPrefix(rawCallee, "cls.")) && callerClass != "" {
				rest := strings.TrimPrefix(strings.TrimPrefix(rawCallee, "self."), "cls.")
				if !strings.Contains(rest, ".") {
					if tid, ok := qualifiedIDs[moduleName+"."+callerClass+"."+rest]; ok {
						return tid, 0.95, "AST_EXACT"
					}
					// Check inherited methods on base classes
					for _, qBase := range classResolvedBases[callerClass] {
						if tid, ok := qualifiedIDs[qBase+"."+rest]; ok {
							return tid, 0.95, "AST_EXACT"
						}
					}
				} else {
					parts := strings.SplitN(rest, ".", 2)
					attrKey := "self." + parts[0]
					mName := parts[1]
					typeName := resolveTypeBinding(callerClass, localTypes[attrKey])
					if typeName == "" && classFieldTypes[callerClass] != nil {
						typeName = classFieldTypes[callerClass][attrKey]
					}
					if tid := resolveMethodOnType(typeName, mName); tid != "" {
						return tid, 0.95, "AST_EXACT"
					}
				}
			}

			// 3. Dotted call: `alias.func()`, `Class.method()`, or `local_var.method()`
			if dotIdx := strings.IndexByte(rawCallee, '.'); dotIdx > 0 {
				prefix := rawCallee[:dotIdx]
				member := rawCallee[dotIdx+1:]

				if rawVarType, ok := localTypes[prefix]; ok {
					varType := resolveTypeBinding(callerClass, rawVarType)
					if tid := resolveMethodOnType(varType, member); tid != "" {
						return tid, 0.95, "AST_EXACT"
					}
				}

				if targetMod, ok := importedModules[prefix]; ok {
					if tid, ok := qualifiedIDs[targetMod+"."+member]; ok {
						return tid, 0.95, "AST_EXACT"
					}
					if tid, ok := qualifiedIDs[targetMod+"."+member+".__init__"]; ok {
						return tid, 0.95, "AST_EXACT"
					}
				}

				if targetSym, ok := importedSymbols[prefix]; ok {
					if tid, ok := qualifiedIDs[targetSym+"."+member]; ok {
						return tid, 0.95, "AST_EXACT"
					}
				}
			}

			// 4. Bare call to an explicitly imported symbol or alias: `from pkg import fn as f; f()`
			if targetSym, ok := importedSymbols[rawCallee]; ok {
				if tid, ok := qualifiedIDs[targetSym]; ok {
					return tid, 0.95, "AST_EXACT"
				}
				if tid, ok := qualifiedIDs[targetSym+".__init__"]; ok {
					return tid, 0.95, "AST_EXACT"
				}
			}

			// 5. Same-module function or class
			if tid, ok := qualifiedIDs[moduleName+"."+rawCallee]; ok {
				return tid, 0.95, "AST_EXACT"
			}
			if tid, ok := qualifiedIDs[rawCallee]; ok {
				return tid, 0.95, "AST_EXACT"
			}

			// 6. Fallback scan across imported modules
			for _, imp := range m.Imports {
				targetMod := resolvePythonModuleName(moduleName, imp.Module, moduleIDs)
				if tid, ok := qualifiedIDs[targetMod+"."+rawCallee]; ok {
					return tid, 0.90, "PYTHON_IMPORT"
				}
			}
			return "", 0, ""
		}

		for _, fn := range m.Functions {
			fromID := qualifiedIDs[moduleName+"."+fn.Name]
			seenCalls := make(map[string]bool)
			for idx, callee := range fn.Calls {
				callLine := fn.LineStart
				if idx < len(fn.CallSites) && fn.CallSites[idx].Line > 0 {
					callLine = fn.CallSites[idx].Line
				}
				if targetID, conf, method := resolvePyTarget("", fn.LocalTypes, callee); targetID != "" && !seenCalls[targetID] {
					seenCalls[targetID] = true
					result.Relationships = append(result.Relationships, RelationRecord{
						FromID:           fromID,
						ToID:             targetID,
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
				if mth.Name == "__init__" && mth.LineStart == cls.LineStart {
					continue
				}
				fromID := qualifiedIDs[moduleName+"."+cls.Name+"."+mth.Name]
				seenCalls := make(map[string]bool)
				for idx, callee := range mth.Calls {
					callLine := mth.LineStart
					if idx < len(mth.CallSites) && mth.CallSites[idx].Line > 0 {
						callLine = mth.CallSites[idx].Line
					}
					if targetID, conf, method := resolvePyTarget(cls.Name, mth.LocalTypes, callee); targetID != "" && !seenCalls[targetID] {
						seenCalls[targetID] = true
						result.Relationships = append(result.Relationships, RelationRecord{
							FromID:           fromID,
							ToID:             targetID,
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

	result.Fingerprint = computeFingerprint(result)
	return result, nil
}

func resolvePythonModuleName(currentModule, rawImport string, moduleIDs map[string]string) string {
	if strings.HasPrefix(rawImport, ".") {
		level := 0
		for level < len(rawImport) && rawImport[level] == '.' {
			level++
		}
		rest := rawImport[level:]
		parts := strings.Split(currentModule, ".")
		if len(parts) > 0 {
			parts = parts[:len(parts)-1]
		}
		for i := 1; i < level && len(parts) > 0; i++ {
			parts = parts[:len(parts)-1]
		}
		if rest != "" {
			parts = append(parts, rest)
		}
		candidate := strings.Join(parts, ".")
		if _, ok := moduleIDs[candidate]; ok {
			return candidate
		}
	}
	if _, ok := moduleIDs[rawImport]; ok {
		return rawImport
	}
	if _, ok := moduleIDs["src."+rawImport]; ok {
		return "src." + rawImport
	}
	for mod := range moduleIDs {
		if strings.HasSuffix(mod, "."+rawImport) {
			return mod
		}
	}
	return rawImport
}

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
