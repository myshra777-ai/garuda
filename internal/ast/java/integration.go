// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package java

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

type EntityRecord struct {
	ID         string
	Name       string
	Kind       string // class | interface | method | module
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
	ResolutionMethod string // JAVA_IMPORT | AST_EXACT | HEURISTIC
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
	Methods    int
	Imports    int
}

// AnalyzeDirectory parses all Java source files (.java) under rootPath and resolves
// cross-package imports (including static and wildcard imports), class inheritance,
// interface implementations, Spring/Jakarta field & constructor DI, generic return-type
// propagation, and method/static/super calls.
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
				base == "target" || base == "build" || base == "out" ||
				base == "bin" || base == "node_modules" || base == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) == ".java" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)

	moduleIDs := make(map[string]string)         // FQN class/file module -> entityID
	classIDs := make(map[string]string)          // FQN or simple class name -> entityID
	ifaceIDs := make(map[string]string)          // FQN or simple interface name -> entityID
	qualifiedMethods := make(map[string]string)  // "pkg.Class.method" or "Class.method" -> method entityID
	methodReturnTypes := make(map[string]string) // "pkg.Class.method" or "Class.method" -> unwrapped return type
	classExtends := make(map[string]string)      // simple/FQN class -> base class simple name
	ifaceImpls := make(map[string][]string)      // simple interface name -> implementing simple class names
	pkgClasses := make(map[string][]string)      // packageName -> list of simple class/interface names

	ctx := context.Background()
	parsedModules := make(map[string]*Module)

	for _, filePath := range files {
		src, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}
		fallbackPkg := deriveJavaPackage(rootPath, filePath)
		m, err := parser.Parse(ctx, src, filePath, fallbackPkg)
		if err != nil {
			continue
		}
		baseFileName := strings.TrimSuffix(filepath.Base(filePath), ".java")
		if m.PackageName != "" {
			m.Module = m.PackageName + "." + baseFileName
		} else {
			m.Module = baseFileName
		}
		parsedModules[filePath] = m
		result.Stats.Files++

		modID := deterministicID("java", "module", m.Module, filePath, 0)
		moduleIDs[m.Module] = modID
		result.Entities = append(result.Entities, EntityRecord{
			ID:         modID,
			Name:       m.Module,
			Kind:       "module",
			Package:    m.PackageName,
			FilePath:   filePath,
			LineStart:  1,
			LineEnd:    strings.Count(string(src), "\n") + 1,
			IsExported: true,
			Language:   "java",
			Hash:       hashOf(src),
		})

		for _, cls := range m.Classes {
			fqnClass := cls.Name
			if m.PackageName != "" {
				fqnClass = m.PackageName + "." + cls.Name
			}
			cID := deterministicID("java", "class", fqnClass, filePath, cls.LineStart)
			classIDs[fqnClass] = cID
			classIDs[cls.Name] = cID
			moduleIDs[fqnClass] = cID
			pkgClasses[m.PackageName] = append(pkgClasses[m.PackageName], cls.Name)

			if cls.Extends != "" {
				classExtends[cls.Name] = cls.Extends
				classExtends[fqnClass] = cls.Extends
			}
			for _, impl := range cls.Implements {
				ifaceImpls[impl] = append(ifaceImpls[impl], cls.Name)
			}

			result.Entities = append(result.Entities, EntityRecord{
				ID:         cID,
				Name:       cls.Name,
				Kind:       "class",
				Package:    m.PackageName,
				FilePath:   filePath,
				LineStart:  cls.LineStart,
				LineEnd:    cls.LineEnd,
				IsExported: cls.IsPublic,
				Signature:  classSignature(cls),
				Language:   "java",
			})
			result.Stats.Classes++

			for _, mth := range cls.Methods {
				fqnMethod := fqnClass + "." + mth.Name
				mID := deterministicID("java", "method", fqnMethod, filePath, mth.LineStart)
				qualifiedMethods[fqnMethod] = mID
				qualifiedMethods[cls.Name+"."+mth.Name] = mID
				if mth.ReturnType != "" {
					methodReturnTypes[fqnMethod] = mth.ReturnType
					methodReturnTypes[cls.Name+"."+mth.Name] = mth.ReturnType
				}
				result.Entities = append(result.Entities, EntityRecord{
					ID:         mID,
					Name:       cls.Name + "." + mth.Name,
					Kind:       "method",
					Package:    m.PackageName,
					FilePath:   filePath,
					LineStart:  mth.LineStart,
					LineEnd:    mth.LineEnd,
					IsExported: mth.IsPublic,
					Signature:  methodSignature(mth),
					Language:   "java",
				})
				result.Stats.Methods++
			}
		}

		for _, iface := range m.Interfaces {
			fqnIface := iface.Name
			if m.PackageName != "" {
				fqnIface = m.PackageName + "." + iface.Name
			}
			iID := deterministicID("java", "interface", fqnIface, filePath, iface.LineStart)
			ifaceIDs[fqnIface] = iID
			ifaceIDs[iface.Name] = iID
			moduleIDs[fqnIface] = iID
			pkgClasses[m.PackageName] = append(pkgClasses[m.PackageName], iface.Name)

			result.Entities = append(result.Entities, EntityRecord{
				ID:         iID,
				Name:       iface.Name,
				Kind:       "interface",
				Package:    m.PackageName,
				FilePath:   filePath,
				LineStart:  iface.LineStart,
				LineEnd:    iface.LineEnd,
				IsExported: iface.IsPublic,
				Signature:  "interface " + iface.Name,
				Language:   "java",
			})
			result.Stats.Interfaces++

			for _, mth := range iface.Methods {
				fqnMethod := fqnIface + "." + mth.Name
				mID := deterministicID("java", "method", fqnMethod, filePath, mth.LineStart)
				qualifiedMethods[fqnMethod] = mID
				qualifiedMethods[iface.Name+"."+mth.Name] = mID
				if mth.ReturnType != "" {
					methodReturnTypes[fqnMethod] = mth.ReturnType
					methodReturnTypes[iface.Name+"."+mth.Name] = mth.ReturnType
				}
				result.Entities = append(result.Entities, EntityRecord{
					ID:         mID,
					Name:       iface.Name + "." + mth.Name,
					Kind:       "method",
					Package:    m.PackageName,
					FilePath:   filePath,
					LineStart:  mth.LineStart,
					LineEnd:    mth.LineEnd,
					IsExported: true,
					Signature:  methodSignature(mth),
					Language:   "java",
				})
				result.Stats.Methods++
			}
		}
	}

	// Pass 2: Resolve Relationships (IMPORTS, INHERITS, IMPLEMENTS, CALLS)
	seenRels := make(map[string]bool)
	addRel := func(rel RelationRecord) {
		if rel.FromID == "" || rel.ToID == "" || rel.FromID == rel.ToID {
			return
		}
		key := rel.FromID + "|" + rel.ToID + "|" + rel.ClaimType
		if seenRels[key] {
			return
		}
		seenRels[key] = true
		result.Relationships = append(result.Relationships, rel)
	}

	for _, filePath := range files {
		m, ok := parsedModules[filePath]
		if !ok {
			continue
		}
		fromModID := moduleIDs[m.Module]
		importedTypes := make(map[string]string) // simpleName -> FQN
		staticImports := make(map[string]string) // staticMethodName -> target method entityID

		// Same-package classes are automatically in scope in Java
		for _, samePkgCls := range pkgClasses[m.PackageName] {
			if m.PackageName != "" {
				importedTypes[samePkgCls] = m.PackageName + "." + samePkgCls
			} else {
				importedTypes[samePkgCls] = samePkgCls
			}
		}

		for _, imp := range m.Imports {
			result.Stats.Imports++
			line := imp.Line
			if line <= 0 {
				line = 1
			}
			if imp.IsStatic {
				if imp.IsWildcard {
					// import static com.acme.AuditLogger.*;
					simpleCls := imp.ClassName
					if dot := strings.LastIndex(simpleCls, "."); dot >= 0 {
						simpleCls = simpleCls[dot+1:]
					}
					if targetID := lookupTypeID(imp.ClassName, simpleCls, classIDs, ifaceIDs, moduleIDs); targetID != "" {
						addRel(RelationRecord{
							FromID:           fromModID,
							ToID:             targetID,
							ClaimType:        "IMPORTS",
							Confidence:       1.0,
							ResolutionStatus: "RESOLVED",
							ResolutionMethod: "JAVA_IMPORT",
							FilePath:         filePath,
							Line:             line,
						})
					}
					for qKey, mID := range qualifiedMethods {
						if strings.HasPrefix(qKey, imp.ClassName+".") || strings.HasPrefix(qKey, simpleCls+".") {
							parts := strings.Split(qKey, ".")
							mName := parts[len(parts)-1]
							staticImports[mName] = mID
						}
					}
				} else {
					// import static com.acme.AuditLogger.recordAudit;
					simpleCls := imp.ClassName
					if dot := strings.LastIndex(simpleCls, "."); dot >= 0 {
						simpleCls = simpleCls[dot+1:]
					}
					if mID, ok := qualifiedMethods[imp.Path]; ok {
						staticImports[imp.SimpleName] = mID
					} else if mID, ok := qualifiedMethods[simpleCls+"."+imp.SimpleName]; ok {
						staticImports[imp.SimpleName] = mID
					}
					if targetID := lookupTypeID(imp.ClassName, simpleCls, classIDs, ifaceIDs, moduleIDs); targetID != "" {
						addRel(RelationRecord{
							FromID:           fromModID,
							ToID:             targetID,
							ClaimType:        "IMPORTS",
							Confidence:       1.0,
							ResolutionStatus: "RESOLVED",
							ResolutionMethod: "JAVA_IMPORT",
							FilePath:         filePath,
							Line:             line,
						})
					}
				}
				continue
			}

			if imp.IsWildcard {
				pkgPrefix := strings.TrimSuffix(imp.Path, ".*")
				for _, clsName := range pkgClasses[pkgPrefix] {
					fqn := pkgPrefix + "." + clsName
					importedTypes[clsName] = fqn
					if targetID := lookupTypeID(fqn, clsName, classIDs, ifaceIDs, moduleIDs); targetID != "" {
						addRel(RelationRecord{
							FromID:           fromModID,
							ToID:             targetID,
							ClaimType:        "IMPORTS",
							Confidence:       0.95,
							ResolutionStatus: "RESOLVED",
							ResolutionMethod: "JAVA_IMPORT",
							FilePath:         filePath,
							Line:             line,
						})
					}
				}
				continue
			}

			importedTypes[imp.SimpleName] = imp.Path
			if targetID := lookupTypeID(imp.Path, imp.SimpleName, classIDs, ifaceIDs, moduleIDs); targetID != "" {
				addRel(RelationRecord{
					FromID:           fromModID,
					ToID:             targetID,
					ClaimType:        "IMPORTS",
					Confidence:       1.0,
					ResolutionStatus: "RESOLVED",
					ResolutionMethod: "JAVA_IMPORT",
					FilePath:         filePath,
					Line:             line,
				})
			}
		}

		// Resolve Class Inheritance, Interface Implementations, and Method Calls
		for _, cls := range m.Classes {
			fqnClass := cls.Name
			if m.PackageName != "" {
				fqnClass = m.PackageName + "." + cls.Name
			}
			cID := classIDs[fqnClass]

			if cls.Extends != "" {
				baseFQN := importedTypes[cls.Extends]
				if baseID := lookupTypeID(baseFQN, cls.Extends, classIDs, ifaceIDs, moduleIDs); baseID != "" {
					addRel(RelationRecord{
						FromID:           cID,
						ToID:             baseID,
						ClaimType:        "INHERITS",
						Confidence:       1.0,
						ResolutionStatus: "RESOLVED",
						ResolutionMethod: "AST_EXACT",
						FilePath:         filePath,
						Line:             cls.LineStart,
					})
				}
			}

			for _, impl := range cls.Implements {
				implFQN := importedTypes[impl]
				if implID := lookupTypeID(implFQN, impl, ifaceIDs, classIDs, moduleIDs); implID != "" {
					addRel(RelationRecord{
						FromID:           cID,
						ToID:             implID,
						ClaimType:        "IMPLEMENTS",
						Confidence:       1.0,
						ResolutionStatus: "RESOLVED",
						ResolutionMethod: "AST_EXACT",
						FilePath:         filePath,
						Line:             cls.LineStart,
					})
				}
			}

			// Build class field map
			classFieldTypes := make(map[string]string)
			for _, f := range cls.Fields {
				if f.Type != "" {
					classFieldTypes[f.Name] = f.Type
					classFieldTypes["this."+f.Name] = f.Type
				}
			}

			for _, mth := range cls.Methods {
				fqnMethod := fqnClass + "." + mth.Name
				callerID := qualifiedMethods[fqnMethod]
				if callerID == "" {
					continue
				}

				localTypes := make(map[string]string)
				for k, v := range classFieldTypes {
					localTypes[k] = v
				}
				for k, v := range mth.LocalTypes {
					localTypes[k] = v
				}

				// Resolve `@return:...` return-type propagation entries
				for varName, typeRef := range localTypes {
					if strings.HasPrefix(typeRef, "@return:") {
						calleeExpr := strings.TrimPrefix(typeRef, "@return:")
						if rType := resolveReturnType(calleeExpr, cls.Name, fqnClass, localTypes, methodReturnTypes); rType != "" {
							localTypes[varName] = rType
						}
					}
				}

				for _, site := range mth.CallSites {
					targetID, conf, method := resolveJavaCallTarget(
						site.Target,
						cls.Name,
						fqnClass,
						localTypes,
						importedTypes,
						staticImports,
						classExtends,
						ifaceImpls,
						qualifiedMethods,
						methodReturnTypes,
						classIDs,
					)
					if targetID != "" && targetID != callerID {
						addRel(RelationRecord{
							FromID:           callerID,
							ToID:             targetID,
							ClaimType:        "CALLS",
							Confidence:       conf,
							ResolutionStatus: "RESOLVED",
							ResolutionMethod: method,
							FilePath:         filePath,
							Line:             site.Line,
						})
					}
				}
			}
		}
	}

	result.Fingerprint = computeFingerprint(result)
	return result, nil
}

func resolveReturnType(
	calleeExpr, clsName, fqnClass string,
	localTypes map[string]string,
	methodReturnTypes map[string]string,
) string {
	expr := strings.TrimPrefix(calleeExpr, "this.")
	if rt, ok := methodReturnTypes[fqnClass+"."+expr]; ok && rt != "" {
		return rt
	}
	if rt, ok := methodReturnTypes[clsName+"."+expr]; ok && rt != "" {
		return rt
	}
	if dot := strings.LastIndex(expr, "."); dot > 0 {
		recv := expr[:dot]
		mName := expr[dot+1:]
		if recvType := localTypes[recv]; recvType != "" {
			if rt, ok := methodReturnTypes[recvType+"."+mName]; ok && rt != "" {
				return rt
			}
		}
	}
	return ""
}

func resolveJavaCallTarget(
	rawTarget, clsName, fqnClass string,
	localTypes map[string]string,
	importedTypes map[string]string,
	staticImports map[string]string,
	classExtends map[string]string,
	ifaceImpls map[string][]string,
	qualifiedMethods map[string]string,
	methodReturnTypes map[string]string,
	classIDs map[string]string,
) (string, float64, string) {
	// 1. Chained call `@chain:this.getRepo.fetchOrder`
	if strings.HasPrefix(rawTarget, "@chain:") {
		chainExpr := strings.TrimPrefix(rawTarget, "@chain:")
		if dot := strings.LastIndex(chainExpr, "."); dot > 0 {
			firstCall := chainExpr[:dot]
			secondMethod := chainExpr[dot+1:]
			retType := resolveReturnType(firstCall, clsName, fqnClass, localTypes, methodReturnTypes)
			if retType != "" {
				if id := lookupMethodOnType(retType, secondMethod, importedTypes, ifaceImpls, qualifiedMethods); id != "" {
					return id, 0.95, "AST_EXACT"
				}
			}
		}
		return "", 0, ""
	}

	// 2. `super.method()` call via inheritance chain
	if strings.HasPrefix(rawTarget, "super.") {
		mName := strings.TrimPrefix(rawTarget, "super.")
		currBase := classExtends[clsName]
		for depth := 0; depth < 5 && currBase != ""; depth++ {
			if id := lookupMethodOnType(currBase, mName, importedTypes, ifaceImpls, qualifiedMethods); id != "" {
				return id, 0.95, "AST_EXACT"
			}
			currBase = classExtends[currBase]
		}
		return "", 0, ""
	}

	// 3. `this.field.method()` or `this.method()`
	if strings.HasPrefix(rawTarget, "this.") {
		rest := strings.TrimPrefix(rawTarget, "this.")
		if !strings.Contains(rest, ".") {
			// Same-class or inherited method `this.helper()`
			if id := lookupMethodOnType(clsName, rest, importedTypes, ifaceImpls, qualifiedMethods); id != "" {
				return id, 0.95, "AST_EXACT"
			}
			if base := classExtends[clsName]; base != "" {
				if id := lookupMethodOnType(base, rest, importedTypes, ifaceImpls, qualifiedMethods); id != "" {
					return id, 0.95, "AST_EXACT"
				}
			}
			return "", 0, ""
		}
		parts := strings.SplitN(rest, ".", 2)
		fieldName := parts[0]
		mName := parts[1]
		if fType := localTypes["this."+fieldName]; fType != "" {
			if id := lookupMethodOnType(fType, mName, importedTypes, ifaceImpls, qualifiedMethods); id != "" {
				return id, 0.95, "AST_EXACT"
			}
		}
		if fType := localTypes[fieldName]; fType != "" {
			if id := lookupMethodOnType(fType, mName, importedTypes, ifaceImpls, qualifiedMethods); id != "" {
				return id, 0.95, "AST_EXACT"
			}
		}
	}

	// 4. Dotted call `receiver.method()` or `ClassName.staticMethod()`
	if dot := strings.LastIndex(rawTarget, "."); dot > 0 {
		recv := rawTarget[:dot]
		mName := rawTarget[dot+1:]
		if rType := localTypes[recv]; rType != "" && !strings.HasPrefix(rType, "@") {
			if id := lookupMethodOnType(rType, mName, importedTypes, ifaceImpls, qualifiedMethods); id != "" {
				return id, 0.95, "AST_EXACT"
			}
		}
		// Static method call on imported or same-package class `AuditLogger.recordAudit(...)`
		if id := lookupMethodOnType(recv, mName, importedTypes, ifaceImpls, qualifiedMethods); id != "" {
			return id, 0.95, "AST_EXACT"
		}
		return "", 0, ""
	}

	// 5. Unqualified call: static import, constructor `new ClassName()`, or same-class/inherited method
	if mID, ok := staticImports[rawTarget]; ok && mID != "" {
		return mID, 0.95, "AST_EXACT"
	}
	// Constructor call `new OrderRepository()` where rawTarget is `OrderRepository`
	if id := lookupMethodOnType(rawTarget, rawTarget, importedTypes, ifaceImpls, qualifiedMethods); id != "" {
		return id, 0.95, "AST_EXACT"
	}
	if fqn, ok := importedTypes[rawTarget]; ok {
		if cID, ok := classIDs[fqn]; ok {
			return cID, 0.95, "AST_EXACT"
		}
	}
	if cID, ok := classIDs[rawTarget]; ok {
		return cID, 0.95, "AST_EXACT"
	}
	// Unqualified same-class or inherited method call `helperMethod()`
	if id := lookupMethodOnType(clsName, rawTarget, importedTypes, ifaceImpls, qualifiedMethods); id != "" {
		return id, 0.95, "AST_EXACT"
	}
	if base := classExtends[clsName]; base != "" {
		if id := lookupMethodOnType(base, rawTarget, importedTypes, ifaceImpls, qualifiedMethods); id != "" {
			return id, 0.95, "AST_EXACT"
		}
	}

	return "", 0, ""
}

func lookupMethodOnType(
	typeName, methodName string,
	importedTypes map[string]string,
	ifaceImpls map[string][]string,
	qualifiedMethods map[string]string,
) string {
	// If typeName is an interface implemented by a concrete class in the workspace,
	// prefer resolving the concrete implementation method so call graphs connect through DI interfaces!
	if impls := ifaceImpls[typeName]; len(impls) > 0 {
		for _, implCls := range impls {
			if fqn, ok := importedTypes[implCls]; ok {
				if id, ok := qualifiedMethods[fqn+"."+methodName]; ok {
					return id
				}
			}
			if id, ok := qualifiedMethods[implCls+"."+methodName]; ok {
				return id
			}
		}
	}
	if fqn, ok := importedTypes[typeName]; ok {
		if id, ok := qualifiedMethods[fqn+"."+methodName]; ok {
			return id
		}
	}
	if id, ok := qualifiedMethods[typeName+"."+methodName]; ok {
		return id
	}
	return ""
}

func lookupTypeID(fqn, simple string, maps ...map[string]string) string {
	for _, m := range maps {
		if fqn != "" {
			if id, ok := m[fqn]; ok && id != "" {
				return id
			}
		}
		if simple != "" {
			if id, ok := m[simple]; ok && id != "" {
				return id
			}
		}
	}
	return ""
}

func deriveJavaPackage(rootPath, filePath string) string {
	rel, err := filepath.Rel(rootPath, filepath.Dir(filePath))
	if err != nil || rel == "." {
		return "default"
	}
	slash := filepath.ToSlash(rel)
	for _, prefix := range []string{"src/main/java/", "src/test/java/", "src/java/", "src/"} {
		if idx := strings.Index(slash, prefix); idx >= 0 {
			slash = slash[idx+len(prefix):]
			break
		}
	}
	slash = strings.Trim(slash, "/")
	if slash == "" {
		return "default"
	}
	return strings.ReplaceAll(slash, "/", ".")
}

func deterministicID(lang, kind, qualifiedName, filePath string, line int) string {
	seed := fmt.Sprintf("%s:%s:%s:%s:%d", lang, kind, qualifiedName, filePath, line)
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(seed)).String()
}

func hashOf(b []byte) string {
	h := sha256.Sum256(b)
	return fmt.Sprintf("%x", h[:8])
}

func classSignature(cls Class) string {
	sig := cls.Kind + " " + cls.Name
	if cls.Extends != "" {
		sig += " extends " + cls.Extends
	}
	if len(cls.Implements) > 0 {
		sig += " implements " + strings.Join(cls.Implements, ", ")
	}
	return sig
}

func methodSignature(m Method) string {
	ret := m.ReturnType
	if ret == "" {
		ret = "void"
	}
	return fmt.Sprintf("%s(%s): %s", m.Name, strings.Join(m.Params, ", "), ret)
}

func computeFingerprint(r *Result) string {
	var parts []string
	for _, e := range r.Entities {
		parts = append(parts, e.ID+":"+e.Hash)
	}
	sort.Strings(parts)
	h := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return fmt.Sprintf("%x", h[:])
}
