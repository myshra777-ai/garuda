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

	// 1. Resolve IMPLEMENTS edges via go/types method sets
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

			for _, target := range allEntities {
				if target.Kind != KindInterface {
					continue
				}

				ptrType := types.NewPointer(namedType)
				for _, candidate := range []types.Type{namedType, ptrType} {
					for _, ifaceMethod := range target.Methods {
						m, _, _ := types.LookupFieldOrMethod(candidate, true, pkg, ifaceMethod.Name)
						if m != nil {
							rels = append(rels, Relationship{
								From:       fmt.Sprintf("%s.%s", pkgPath, name),
								To:         fmt.Sprintf("%s.%s", target.Package, target.Name),
								Type:       string(RelImplements),
								Confidence: 1.0,
							})
							break
						}
					}
				}
			}
		}
	}

	// 2. Resolve CALLS and IMPORTS edges from AST
	for _, file := range files {
		for _, imp := range file.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			rels = append(rels, Relationship{
				From:       pkgPath,
				To:         importPath,
				Type:       string(RelImports),
				Confidence: 1.0,
			})
		}

		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if obj, exists := info.Uses[sel.Sel]; exists && obj != nil {
					if obj.Pkg() != nil {
						rels = append(rels, Relationship{
							From:       pkgPath,
							To:         fmt.Sprintf("%s.%s", obj.Pkg().Path(), obj.Name()),
							Type:       string(RelCalls),
							Confidence: 1.0,
						})
					}
				}
			}
			return true
		})
	}

	return rels
}
