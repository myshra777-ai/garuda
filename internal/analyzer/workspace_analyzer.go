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

	// 1. Compute tree hash per package and evaluate cache hits
	var packagesToParse []string

	for _, pkgPath := range sortedPkgPaths {
		dirPath := ws.PackageRoots[pkgPath]
		treeHash, err := ComputePackageTreeHash(dirPath)
		if err != nil {
			return nil, fmt.Errorf("failed to compute tree hash for %s: %w", pkgPath, err)
		}
		packageHashes[pkgPath] = treeHash

		// Incorporate into repository fingerprint
		repoHasher.Write(fmt.Appendf(nil, "%s:%x:", pkgPath, treeHash))

		// Check cache
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

	// 2. Parse only packages that missed cache or need type checking
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

	// 3. Type-check and extract entities + relationships for parsed packages
	for _, pkgPath := range packagesToParse {
		files, exists := packageFiles[pkgPath]
		if !exists || len(files) == 0 {
			continue
		}

		pkg := types.NewPackage(pkgPath, files[0].Name.Name)
		conf := types.Config{
			Importer: imp,
			Error:    func(err error) {},
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

		pkgRels := extractWorkspaceRelationships(pkgPath, files, pkg, info, result.Entities)
		result.Relationships = append(result.Relationships, pkgRels...)

		// Populate cache for newly analyzed package
		if opts.Cache != nil {
			_ = opts.Cache.PutPackage(ctx, opts.TenantID, pkgPath, packageHashes[pkgPath], CachedPackageData{
				Entities:      pkgEntities,
				Relationships: pkgRels,
			})
		}
	}

	// 4. Update summary stats
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

	return result, nil
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
							ID:       canonicalID.String(),
							Name:     name,
							Kind:     KindStruct,
							Package:  pkgPath,
							File:     fset.Position(typeSpec.Pos()).Filename,
							Line:     fset.Position(typeSpec.Pos()).Line,
							Exported: ast.IsExported(name),
							Fields:   fields,
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
							ID:       canonicalID.String(),
							Name:     name,
							Kind:     KindInterface,
							Package:  pkgPath,
							File:     fset.Position(typeSpec.Pos()).Filename,
							Line:     fset.Position(typeSpec.Pos()).Line,
							Exported: ast.IsExported(name),
							Methods:  methods,
						})

					default:
						// Non-struct, non-interface type declaration.
						//   type A = B  → alias      (Assign != 0, '=' present)
						//   type A B    → defined    (Assign == 0, no '=')
						kind := KindType
						if typeSpec.Assign != 0 {
							kind = KindAlias
						}
						entities = append(entities, Entity{
							ID:       canonicalID.String(),
							Name:     name,
							Kind:     kind,
							Package:  pkgPath,
							File:     fset.Position(typeSpec.Pos()).Filename,
							Line:     fset.Position(typeSpec.Pos()).Line,
							Exported: ast.IsExported(name),
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

				entities = append(entities, Entity{
					ID:        canonicalID.String(),
					Name:      funcName,
					Kind:      kind,
					Package:   pkgPath,
					Signature: sig,
					File:      fset.Position(d.Pos()).Filename,
					Line:      fset.Position(d.Pos()).Line,
					Exported:  ast.IsExported(funcName),
				})
			}
		}
	}

	return entities
}

func extractWorkspaceRelationships(pkgPath string, files []*ast.File, pkg *types.Package, info *types.Info, allEntities []Entity) []Relationship {
	var rels []Relationship

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
			})
		}
	}

	// ─── 2. CALLS edges via info.Uses on selector expressions ───
	for _, file := range files {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if obj, exists := info.Uses[sel.Sel]; exists && obj != nil {
					if obj.Pkg() != nil {
						rels = append(rels, Relationship{
							From:             pkgPath,
							To:               fmt.Sprintf("%s.%s", obj.Pkg().Path(), obj.Name()),
							Type:             string(RelCalls),
							Confidence:       1.0,
							ResolutionStatus: "RESOLVED",
							ResolutionMethod: "GO_TYPES",
							EpistemicClass:   "OBSERVATION",
						})
					}
				}
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

			// 3a. DEFINES: type → each declared method
			for i := 0; i < namedType.NumMethods(); i++ {
				m := namedType.Method(i)
				rels = append(rels, Relationship{
					From:             fmt.Sprintf("%s.%s", pkgPath, name),
					To:               fmt.Sprintf("%s.%s", pkgPath, m.Name()),
					Type:             string(RelDefines),
					Confidence:       0.90,
					ResolutionStatus: "RESOLVED",
					ResolutionMethod: "AST_EXACT",
					EpistemicClass:   "OBSERVATION",
				})
			}

			// 3b. EMBEDS: outer struct → embedded field types
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
					rels = append(rels, Relationship{
						From:             fmt.Sprintf("%s.%s", pkgPath, name),
						To:               fmt.Sprintf("%s.%s", inner.Obj().Pkg().Path(), inner.Obj().Name()),
						Type:             string(RelEmbeds),
						Confidence:       0.90,
						ResolutionStatus: "RESOLVED",
						ResolutionMethod: "AST_EXACT",
						EpistemicClass:   "OBSERVATION",
					})
				}
			}

			// 3c. IMPLEMENTS: struct → interface via types.Implements
			for _, target := range allEntities {
				if target.Kind != KindInterface {
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
					})
				}
			}
		}
	}

	// ─── 4. REFERENCES via function signatures (go/types) ───
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
			seen := map[string]bool{}
			for i := 0; i < sig.Params().Len(); i++ {
				emitReferencesForType(pkgPath, name, sig.Params().At(i).Type(), seen, &rels)
			}
			for i := 0; i < sig.Results().Len(); i++ {
				emitReferencesForType(pkgPath, name, sig.Results().At(i).Type(), seen, &rels)
			}
		}
	}

	// ─── 5. REFERENCES via composite literals in function bodies ───
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
				})
				return true
			})
			return true
		})
	}

	// ─── 6. Deduplicate edges by (From, To, Type) ───
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

// emitReferencesForType unwraps pointer/slice/map/chan types and emits a
// REFERENCES edge from funcName to each named type it uncovers.
func emitReferencesForType(pkgPath, funcName string, t types.Type, seen map[string]bool, rels *[]Relationship) {
	switch tt := t.(type) {
	case *types.Pointer:
		emitReferencesForType(pkgPath, funcName, tt.Elem(), seen, rels)
		return
	case *types.Slice:
		emitReferencesForType(pkgPath, funcName, tt.Elem(), seen, rels)
		return
	case *types.Chan:
		emitReferencesForType(pkgPath, funcName, tt.Elem(), seen, rels)
		return
	case *types.Map:
		emitReferencesForType(pkgPath, funcName, tt.Key(), seen, rels)
		emitReferencesForType(pkgPath, funcName, tt.Elem(), seen, rels)
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
		})

	case *types.Alias:
		// Go represents `type A = B` as *types.Alias, not *types.Named.
		// The alias is a distinct named declaration in source even though
		// its underlying type is transparent to the type-checker. Emit a
		// REFERENCES edge to preserve the source-level name.
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
		})
	}
}

// extractCompositeLitTypeName returns the fully-qualified name of the type
// being constructed in a composite literal, or "" if the type cannot be
// resolved (e.g., an anonymous struct literal).
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
