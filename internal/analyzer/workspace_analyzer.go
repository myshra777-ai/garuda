// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package analyzer

import (
	"context"
	"crypto/sha256"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/uuid"
)

// WorkspaceAnalysisOptions configures workspace analysis parameters.
type WorkspaceAnalysisOptions struct {
	TenantID uuid.UUID
	Cache    PackageCache
}

// AnalyzeWorkspace orchestrates analysis with default options.
func AnalyzeWorkspace(ctx context.Context, ws *WorkspaceContext) (*Result, error) {
	return AnalyzeWorkspaceWithOptions(ctx, ws, WorkspaceAnalysisOptions{})
}

// AnalyzeWorkspaceWithOptions orchestrates full AST parsing, cross-module type checking,
// and semantic relationship extraction with incremental caching support.
func AnalyzeWorkspaceWithOptions(ctx context.Context, ws *WorkspaceContext, opts WorkspaceAnalysisOptions) (*Result, error) {
	if ws == nil || len(ws.Modules) == 0 {
		return nil, fmt.Errorf("cannot analyze empty workspace context")
	}

	fset := token.NewFileSet()
	imp := NewMultiModuleImporter(fset, ws)

	result := &Result{
		Source:        ws.RootPath,
		Entities:      make([]Entity, 0),
		Relationships: make([]Relationship, 0),
		Stats:         Stats{},
	}

	packageFiles := make(map[string][]*ast.File)
	packageHashes := make(map[string][]byte)
	repoHasher := sha256.New()

	var sortedPkgPaths []string
	for pkgPath := range ws.PackageRoots {
		sortedPkgPaths = append(sortedPkgPaths, pkgPath)
	}
	sort.Strings(sortedPkgPaths)

	var packagesToParse []string

	for _, pkgPath := range sortedPkgPaths {
		dirPath := ws.PackageRoots[pkgPath]
		treeHash, err := ComputePackageTreeHash(dirPath)
		if err != nil {
			return nil, fmt.Errorf("failed to compute tree hash for %s: %w", pkgPath, err)
		}
		packageHashes[pkgPath] = treeHash

		repoHasher.Write(fmt.Appendf(nil, "%s:%x:", pkgPath, treeHash))

		var cacheHit bool
		if opts.Cache != nil {
			cached, hit, err := opts.Cache.GetPackage(ctx, opts.TenantID, pkgPath, treeHash)
			if err == nil && hit && cached != nil {
				result.Entities = append(result.Entities, cached.Entities...)
				result.Relationships = append(result.Relationships, cached.Relationships...)
				cacheHit = true
			}
		}

		if !cacheHit {
			packagesToParse = append(packagesToParse, pkgPath)
		}
	}

	result.Fingerprint = fmt.Sprintf("%x", repoHasher.Sum(nil))

	for _, pkgPath := range packagesToParse {
		dirPath := ws.PackageRoots[pkgPath]
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read package dir %s: %w", dirPath, err)
		}

		var goFileNames []string
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") && !strings.HasSuffix(entry.Name(), "_test.go") {
				goFileNames = append(goFileNames, entry.Name())
			}
		}
		sort.Strings(goFileNames)

		var parsedFiles []*ast.File
		for _, name := range goFileNames {
			filePath := filepath.Join(dirPath, name)
			content, err := os.ReadFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read source file %s: %w", filePath, err)
			}

			fileAst, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
			if err != nil {
				return nil, fmt.Errorf("failed to parse file %s: %w", filePath, err)
			}
			parsedFiles = append(parsedFiles, fileAst)
			result.Stats.Files++
		}

		if len(parsedFiles) > 0 {
			packageFiles[pkgPath] = parsedFiles
		}
	}

	typeErrors := 0
	var typeDiagnostics []string
	const maxRecordedDiagnostics = 25

	type checkedPkg struct {
		pkgPath  string
		files    []*ast.File
		pkg      *types.Package
		info     *types.Info
		entities []Entity
	}
	var checkedPkgs []checkedPkg

	// Pass 1: Type-check all packages and collect all workspace entities first
	// so cross-module interface satisfaction sees every interface regardless of lexical package order.
	for _, pkgPath := range packagesToParse {
		files, exists := packageFiles[pkgPath]
		if !exists || len(files) == 0 {
			continue
		}

		pkg := types.NewPackage(pkgPath, files[0].Name.Name)
		conf := types.Config{
			Importer: imp,
			Error: func(err error) {
				typeErrors++
				if len(typeDiagnostics) < maxRecordedDiagnostics {
					typeDiagnostics = append(typeDiagnostics, err.Error())
				}
			},
		}

		info := &types.Info{
			Types: make(map[ast.Expr]types.TypeAndValue),
			Defs:  make(map[*ast.Ident]types.Object),
			Uses:  make(map[*ast.Ident]types.Object),
		}

		checker := types.NewChecker(&conf, fset, pkg, info)
		_ = checker.Files(files)

		pkgEntities := extractWorkspaceEntities(fset, pkgPath, files, pkg, info)
		result.Entities = append(result.Entities, pkgEntities...)

		checkedPkgs = append(checkedPkgs, checkedPkg{
			pkgPath:  pkgPath,
			files:    files,
			pkg:      pkg,
			info:     info,
			entities: pkgEntities,
		})
	}

	// Pass 2: Extract relationships across all packages now that result.Entities is complete
	for _, cp := range checkedPkgs {
		pkgRels := extractWorkspaceRelationships(fset, cp.pkgPath, cp.files, cp.pkg, cp.info, result.Entities)
		result.Relationships = append(result.Relationships, pkgRels...)

		if opts.Cache != nil {
			_ = opts.Cache.PutPackage(ctx, opts.TenantID, cp.pkgPath, packageHashes[cp.pkgPath], CachedPackageData{
				Entities:      cp.entities,
				Relationships: pkgRels,
			})
		}
	}

	// Pass 3: Cross-package structural interface satisfaction fallback (deduplicated against existing edges)
	implRels := deriveWorkspaceInterfaceImplementations(result.Entities, result.Relationships)
	result.Relationships = append(result.Relationships, implRels...)

	result.Stats.Packages = len(sortedPkgPaths)
	for _, e := range result.Entities {
		switch e.Kind {
		case KindStruct:
			result.Stats.Structs++
		case KindInterface:
			result.Stats.Interfaces++
		case KindFunction, KindMethod:
			result.Stats.Functions++
		}
	}
	result.Stats.TypeErrors = typeErrors
	result.Stats.AnalysisComplete = (typeErrors == 0)
	result.TypeDiagnostics = typeDiagnostics

	return result, nil
}

// deriveWorkspaceInterfaceImplementations computes structural IMPLEMENTS edges
// across workspace structs and interfaces when not already emitted by go/types.
func deriveWorkspaceInterfaceImplementations(allEntities []Entity, existing []Relationship) []Relationship {
	var interfaces []Entity
	var structs []Entity
	methodsByStruct := make(map[string]map[string]bool)

	for _, e := range allEntities {
		switch e.Kind {
		case KindInterface:
			if len(e.Methods) > 0 {
				interfaces = append(interfaces, e)
			}
		case KindStruct:
			structs = append(structs, e)
		case KindMethod:
			recvClean := strings.TrimPrefix(e.ReceiverType, "*")
			if recvClean != "" {
				key := e.Package + "." + recvClean
				if methodsByStruct[key] == nil {
					methodsByStruct[key] = make(map[string]bool)
				}
				methodsByStruct[key][e.Name] = true
			}
		}
	}

	seen := make(map[string]bool)
	for _, r := range existing {
		if r.Type == string(RelImplements) {
			seen[r.From+"|"+r.To] = true
		}
	}

	var rels []Relationship
	for _, st := range structs {
		stKey := st.Package + "." + st.Name
		stMethods := methodsByStruct[stKey]
		if len(stMethods) == 0 {
			continue
		}
		for _, iface := range interfaces {
			if st.Name == iface.Name && st.Package == iface.Package {
				continue
			}
			allMatch := true
			for _, m := range iface.Methods {
				if !stMethods[m.Name] {
					allMatch = false
					break
				}
			}
			if allMatch {
				fromQ := stKey
				toQ := iface.Package + "." + iface.Name
				edgeKey := fromQ + "|" + toQ
				if seen[edgeKey] {
					continue
				}
				seen[edgeKey] = true
				rels = append(rels, Relationship{
					From:             fromQ,
					To:               toQ,
					Type:             string(RelImplements),
					Confidence:       1.0,
					ResolutionStatus: "RESOLVED",
					ResolutionMethod: "GO_TYPES",
					EpistemicClass:   "OBSERVATION",
					Evidence: Evidence{
						File:      st.File,
						Line:      st.Line,
						LineStart: st.LineStart,
						LineEnd:   st.LineEnd,
						Analyzer:  "workspace_analyzer",
					},
				})
			}
		}
	}
	return rels
}

type evidenceIndex struct {
	positions map[types.Object]Evidence
}

func buildEvidenceIndex(fset *token.FileSet, files []*ast.File, info *types.Info) *evidenceIndex {
	idx := &evidenceIndex{
		positions: make(map[types.Object]Evidence),
	}
	if info == nil || info.Defs == nil {
		return idx
	}
	for ident, obj := range info.Defs {
		if obj == nil || ident == nil {
			continue
		}
		pos := fset.Position(ident.Pos())
		end := fset.Position(ident.End())
		idx.positions[obj] = Evidence{
			File:      pos.Filename,
			Line:      pos.Line,
			LineStart: pos.Line,
			LineEnd:   end.Line,
			Analyzer:  "workspace_analyzer",
		}
	}
	return idx
}

func (idx *evidenceIndex) lookup(obj types.Object) Evidence {
	if idx == nil || obj == nil {
		return Evidence{}
	}
	if ev, ok := idx.positions[obj]; ok {
		return ev
	}
	return Evidence{}
}

func evidenceAt(fset *token.FileSet, pos, end token.Pos) Evidence {
	start := fset.Position(pos)
	finish := fset.Position(end)
	return Evidence{
		File:      start.Filename,
		Line:      start.Line,
		LineStart: start.Line,
		LineEnd:   finish.Line,
		Analyzer:  "workspace_analyzer",
	}
}

func extractWorkspaceEntities(fset *token.FileSet, pkgPath string, files []*ast.File, pkg *types.Package, info *types.Info) []Entity {
	var entities []Entity

	for _, file := range files {
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}

					name := typeSpec.Name.Name
					canonicalID := uuid.NewSHA1(uuid.NameSpaceURL, fmt.Appendf(nil, "%s:%s", pkgPath, name))
					startPos := fset.Position(typeSpec.Pos())
					endPos := fset.Position(typeSpec.End())

					switch t := typeSpec.Type.(type) {
					case *ast.StructType:
						var fields []Field
						if t.Fields != nil {
							for _, f := range t.Fields.List {
								fieldType := types.ExprString(f.Type)
								if len(f.Names) == 0 {
									fields = append(fields, Field{
										Name: fieldType,
										Type: fieldType,
									})
								} else {
									for _, fieldName := range f.Names {
										fields = append(fields, Field{
											Name: fieldName.Name,
											Type: fieldType,
										})
									}
								}
							}
						}

						entities = append(entities, Entity{
							ID:        canonicalID.String(),
							Name:      name,
							Kind:      KindStruct,
							Package:   pkgPath,
							File:      startPos.Filename,
							Line:      startPos.Line,
							LineStart: startPos.Line,
							LineEnd:   endPos.Line,
							Exported:  ast.IsExported(name),
							Fields:    fields,
						})

					case *ast.InterfaceType:
						var methods []Method
						if t.Methods != nil {
							for _, m := range t.Methods.List {
								if len(m.Names) > 0 {
									methods = append(methods, Method{
										Name:      m.Names[0].Name,
										Signature: types.ExprString(m.Type),
									})
								}
							}
						}

						entities = append(entities, Entity{
							ID:        canonicalID.String(),
							Name:      name,
							Kind:      KindInterface,
							Package:   pkgPath,
							File:      startPos.Filename,
							Line:      startPos.Line,
							LineStart: startPos.Line,
							LineEnd:   endPos.Line,
							Exported:  ast.IsExported(name),
							Methods:   methods,
						})

					default:
						kind := KindType
						if typeSpec.Assign != 0 {
							kind = KindAlias
						}
						entities = append(entities, Entity{
							ID:        canonicalID.String(),
							Name:      name,
							Kind:      kind,
							Package:   pkgPath,
							File:      startPos.Filename,
							Line:      startPos.Line,
							LineStart: startPos.Line,
							LineEnd:   endPos.Line,
							Exported:  ast.IsExported(name),
						})
					}
				}

			case *ast.FuncDecl:
				funcName := d.Name.Name
				var receiver string
				kind := KindFunction

				if d.Recv != nil && len(d.Recv.List) > 0 {
					kind = KindMethod
					receiver = types.ExprString(d.Recv.List[0].Type)
				}

				canonicalID := uuid.NewSHA1(uuid.NameSpaceURL, fmt.Appendf(nil, "%s:%s:%s", pkgPath, receiver, funcName))
				sig := types.ExprString(d.Type)
				startPos := fset.Position(d.Pos())
				endPos := fset.Position(d.End())

				entities = append(entities, Entity{
					ID:           canonicalID.String(),
					Name:         funcName,
					Kind:         kind,
					Package:      pkgPath,
					ReceiverType: receiver,
					Signature:    sig,
					File:         startPos.Filename,
					Line:         startPos.Line,
					LineStart:    startPos.Line,
					LineEnd:      endPos.Line,
					Exported:     ast.IsExported(funcName),
				})
			}
		}
	}

	return entities
}

func extractWorkspaceRelationships(fset *token.FileSet, pkgPath string, files []*ast.File, pkg *types.Package, info *types.Info, allEntities []Entity) []Relationship {
	var rels []Relationship

	idx := buildEvidenceIndex(fset, files, info)

	// ─── 1. IMPORTS edges ───
	for _, file := range files {
		for _, imp := range file.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			rels = append(rels, Relationship{
				From:             pkgPath,
				To:               importPath,
				Type:             string(RelImports),
				Confidence:       0.95,
				ResolutionStatus: "RESOLVED",
				ResolutionMethod: "IMPORT_RESOLUTION",
				EpistemicClass:   "OBSERVATION",
				Evidence:         evidenceAt(fset, imp.Pos(), imp.End()),
			})
		}
	}

	// ─── 2. CALLS edges ───
	for _, file := range files {
		var currentCaller string

		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.FuncDecl:
				if node.Recv != nil && len(node.Recv.List) > 0 {
					recv := types.ExprString(node.Recv.List[0].Type)
					currentCaller = fmt.Sprintf("%s.%s.%s", pkgPath, recv, node.Name.Name)
				} else {
					currentCaller = fmt.Sprintf("%s.%s", pkgPath, node.Name.Name)
				}
				return true

			case *ast.CallExpr:
				if currentCaller == "" {
					return true
				}
				switch fun := node.Fun.(type) {
				case *ast.SelectorExpr:
					obj, exists := info.Uses[fun.Sel]
					if !exists || obj == nil || obj.Pkg() == nil {
						return true
					}
					rels = append(rels, Relationship{
						From:             currentCaller,
						To:               fmt.Sprintf("%s.%s", obj.Pkg().Path(), obj.Name()),
						Type:             string(RelCalls),
						Confidence:       1.0,
						ResolutionStatus: "RESOLVED",
						ResolutionMethod: "GO_TYPES",
						EpistemicClass:   "OBSERVATION",
						Evidence:         evidenceAt(fset, node.Pos(), node.End()),
					})
				case *ast.Ident:
					obj, exists := info.Uses[fun]
					if !exists || obj == nil || obj.Pkg() == nil {
						return true
					}
					if _, isFunc := obj.(*types.Func); !isFunc {
						return true
					}
					rels = append(rels, Relationship{
						From:             currentCaller,
						To:               fmt.Sprintf("%s.%s", obj.Pkg().Path(), obj.Name()),
						Type:             string(RelCalls),
						Confidence:       1.0,
						ResolutionStatus: "RESOLVED",
						ResolutionMethod: "GO_TYPES",
						EpistemicClass:   "OBSERVATION",
						Evidence:         evidenceAt(fset, node.Pos(), node.End()),
					})
				}
				return true
			}
			return true
		})
	}

	// ─── 3. Type-level edges via go/types ───
	if pkg != nil && pkg.Scope() != nil {
		scope := pkg.Scope()
		for _, name := range scope.Names() {
			obj := scope.Lookup(name)
			if obj == nil {
				continue
			}
			namedType, ok := obj.Type().(*types.Named)
			if !ok {
				continue
			}

			srcEvidence := idx.lookup(namedType.Obj())
			if srcEvidence.File == "" {
				srcEvidence = Evidence{Analyzer: "workspace_analyzer"}
			}

			// 3a. DEFINES
			for i := 0; i < namedType.NumMethods(); i++ {
				m := namedType.Method(i)
				methodEv := idx.lookup(m)
				if methodEv.File == "" {
					methodEv = srcEvidence
				}
				rels = append(rels, Relationship{
					From:             fmt.Sprintf("%s.%s", pkgPath, name),
					To:               fmt.Sprintf("%s.%s", pkgPath, m.Name()),
					Type:             string(RelDefines),
					Confidence:       0.90,
					ResolutionStatus: "RESOLVED",
					ResolutionMethod: "AST_EXACT",
					EpistemicClass:   "OBSERVATION",
					Evidence:         methodEv,
				})
			}

			// 3b. EMBEDS
			if st, ok := namedType.Underlying().(*types.Struct); ok {
				for i := 0; i < st.NumFields(); i++ {
					f := st.Field(i)
					if !f.Embedded() {
						continue
					}
					ft := f.Type()
					if ptr, ok := ft.(*types.Pointer); ok {
						ft = ptr.Elem()
					}
					inner, ok := ft.(*types.Named)
					if !ok || inner.Obj().Pkg() == nil {
						continue
					}
					fieldEv := idx.lookup(f)
					if fieldEv.File == "" {
						fieldEv = srcEvidence
					}
					rels = append(rels, Relationship{
						From:             fmt.Sprintf("%s.%s", pkgPath, name),
						To:               fmt.Sprintf("%s.%s", inner.Obj().Pkg().Path(), inner.Obj().Name()),
						Type:             string(RelEmbeds),
						Confidence:       0.90,
						ResolutionStatus: "RESOLVED",
						ResolutionMethod: "AST_EXACT",
						EpistemicClass:   "OBSERVATION",
						Evidence:         fieldEv,
					})
				}
			}

			// 3c. IMPLEMENTS
			for _, target := range allEntities {
				if target.Kind != KindInterface {
					continue
				}
				if name == target.Name && pkgPath == target.Package {
					continue
				}
				var ifaceNamed *types.Named
				if target.Package == "" || target.Package == pkgPath {
					if tn, ok := scope.Lookup(target.Name).(*types.TypeName); ok {
						ifaceNamed, _ = tn.Type().(*types.Named)
					}
				} else {
					for _, imp := range pkg.Imports() {
						if imp.Path() == target.Package {
							if tn, ok := imp.Scope().Lookup(target.Name).(*types.TypeName); ok {
								ifaceNamed, _ = tn.Type().(*types.Named)
							}
							break
						}
					}
				}
				if ifaceNamed == nil {
					continue
				}
				iface, ok := ifaceNamed.Underlying().(*types.Interface)
				if !ok {
					continue
				}
				iface.Complete()

				if types.Implements(namedType, iface) || types.Implements(types.NewPointer(namedType), iface) {
					rels = append(rels, Relationship{
						From:             fmt.Sprintf("%s.%s", pkgPath, name),
						To:               fmt.Sprintf("%s.%s", target.Package, target.Name),
						Type:             string(RelImplements),
						Confidence:       1.0,
						ResolutionStatus: "RESOLVED",
						ResolutionMethod: "GO_TYPES",
						EpistemicClass:   "OBSERVATION",
						Evidence:         srcEvidence,
					})
				}
			}
		}
	}

	// ─── 4. REFERENCES via function signatures ───
	if pkg != nil && pkg.Scope() != nil {
		scope := pkg.Scope()
		for _, name := range scope.Names() {
			obj := scope.Lookup(name)
			if obj == nil {
				continue
			}
			fn, ok := obj.(*types.Func)
			if !ok {
				continue
			}
			sig, _ := fn.Type().(*types.Signature)
			if sig == nil {
				continue
			}
			fnEvidence := idx.lookup(fn)
			if fnEvidence.File == "" {
				fnEvidence = Evidence{Analyzer: "workspace_analyzer"}
			}
			seen := map[string]bool{}
			for i := 0; i < sig.Params().Len(); i++ {
				emitReferencesForType(pkgPath, name, sig.Params().At(i).Type(), fnEvidence, seen, &rels)
			}
			for i := 0; i < sig.Results().Len(); i++ {
				emitReferencesForType(pkgPath, name, sig.Results().At(i).Type(), fnEvidence, seen, &rels)
			}
		}
	}

	// ─── 5. REFERENCES via composite literals ───
	for _, file := range files {
		ast.Inspect(file, func(n ast.Node) bool {
			fd, ok := n.(*ast.FuncDecl)
			if !ok || fd.Body == nil {
				return true
			}
			funcName := fd.Name.Name
			ast.Inspect(fd.Body, func(inner ast.Node) bool {
				cl, ok := inner.(*ast.CompositeLit)
				if !ok || cl.Type == nil {
					return true
				}
				qname := extractCompositeLitTypeName(cl.Type, info)
				if qname == "" {
					return true
				}
				rels = append(rels, Relationship{
					From:             fmt.Sprintf("%s.%s", pkgPath, funcName),
					To:               qname,
					Type:             string(RelReferences),
					Confidence:       0.90,
					ResolutionStatus: "RESOLVED",
					ResolutionMethod: "AST_EXACT",
					EpistemicClass:   "OBSERVATION",
					Evidence:         evidenceAt(fset, cl.Pos(), cl.End()),
				})
				return true
			})
			return true
		})
	}

	// ─── 6. Deduplicate by (From, To, Type) ───
	seen := map[string]bool{}
	deduped := rels[:0]
	for _, r := range rels {
		key := r.From + "|" + r.To + "|" + r.Type
		if seen[key] {
			continue
		}
		seen[key] = true
		deduped = append(deduped, r)
	}
	return deduped
}

func emitReferencesForType(pkgPath, funcName string, t types.Type, evidence Evidence, seen map[string]bool, rels *[]Relationship) {
	switch tt := t.(type) {
	case *types.Pointer:
		emitReferencesForType(pkgPath, funcName, tt.Elem(), evidence, seen, rels)
		return
	case *types.Slice:
		emitReferencesForType(pkgPath, funcName, tt.Elem(), evidence, seen, rels)
		return
	case *types.Chan:
		emitReferencesForType(pkgPath, funcName, tt.Elem(), evidence, seen, rels)
		return
	case *types.Map:
		emitReferencesForType(pkgPath, funcName, tt.Key(), evidence, seen, rels)
		emitReferencesForType(pkgPath, funcName, tt.Elem(), evidence, seen, rels)
		return
	case *types.Named:
		if tt.Obj().Pkg() == nil {
			return
		}
		qname := tt.Obj().Pkg().Path() + "." + tt.Obj().Name()
		if seen[qname] {
			return
		}
		seen[qname] = true
		*rels = append(*rels, Relationship{
			From:             fmt.Sprintf("%s.%s", pkgPath, funcName),
			To:               qname,
			Type:             string(RelReferences),
			Confidence:       1.0,
			ResolutionStatus: "RESOLVED",
			ResolutionMethod: "GO_TYPES",
			EpistemicClass:   "OBSERVATION",
			Evidence:         evidence,
		})

	case *types.Alias:
		if tt.Obj().Pkg() == nil {
			return
		}
		qname := tt.Obj().Pkg().Path() + "." + tt.Obj().Name()
		if seen[qname] {
			return
		}
		seen[qname] = true
		*rels = append(*rels, Relationship{
			From:             fmt.Sprintf("%s.%s", pkgPath, funcName),
			To:               qname,
			Type:             string(RelReferences),
			Confidence:       1.0,
			ResolutionStatus: "RESOLVED",
			ResolutionMethod: "GO_TYPES",
			EpistemicClass:   "OBSERVATION",
			Evidence:         evidence,
		})
	}
}

func extractCompositeLitTypeName(expr ast.Expr, info *types.Info) string {
	tv, ok := info.Types[expr]
	if !ok || tv.Type == nil {
		return ""
	}
	t := tv.Type
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	named, ok := t.(*types.Named)
	if !ok || named.Obj().Pkg() == nil {
		return ""
	}
	return named.Obj().Pkg().Path() + "." + named.Obj().Name()
}
