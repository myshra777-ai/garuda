// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package analyzer

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// AnalyzeWorkspace performs multi-package semantic analysis across an entire workspace.
//
// Type-Checking Completeness Tracking:
// Instead of silently discarding go/types errors, AnalyzeWorkspace counts all
// type-checking diagnostics in result.Stats.TypeErrors, records up to 25
// diagnostic messages in result.TypeDiagnostics, and sets
// result.Stats.AnalysisComplete = (typeErrors == 0).
func (a *Analyzer) AnalyzeWorkspace(root string) (*Result, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve workspace root: %w", err)
	}

	modules, err := DiscoverModules(absRoot)
	if err != nil || len(modules) == 0 {
		modules = []ModuleInfo{{
			ModulePath: "workspace",
			RootDir:    absRoot,
		}}
	}

	fset := token.NewFileSet()
	imp := NewMultiModuleImporter(fset, modules)

	pkgDirs := make(map[string]string) // dir -> modulePath
	err = filepath.WalkDir(absRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "vendor" || name == "node_modules" || name == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(d.Name(), ".go") && !strings.HasSuffix(d.Name(), "_test.go") {
			dir := filepath.Dir(path)
			modPath := "workspace"
			for _, m := range modules {
				if strings.HasPrefix(dir, m.RootDir) {
					modPath = m.ModulePath
					break
				}
			}
			pkgDirs[dir] = modPath
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var sortedDirs []string
	for dir := range pkgDirs {
		sortedDirs = append(sortedDirs, dir)
	}
	sort.Strings(sortedDirs)

	result := &Result{
		AnalyzedAt:    time.Now().UTC(),
		Source:        absRoot,
		Entities:      []Entity{},
		Relationships: []Relationship{},
	}

	repoName := filepath.Base(absRoot)
	repoID := fmt.Sprintf("repo:%s", repoName)
	result.Entities = append(result.Entities, Entity{
		ID:       repoID,
		Kind:     KindRepository,
		Name:     repoName,
		Package:  repoName,
		File:     absRoot,
		Exported: true,
		Language: "go",
	})

	entityMap := make(map[string]Entity)
	entityMap[repoID] = result.Entities[0]

	var allPkgSuites []struct {
		importPath string
		files      []*ast.File
		info       *types.Info
		typesPkg   *types.Package
	}

	typeErrors := 0
	var typeDiagnostics []string
	const maxRecordedDiagnostics = 25

	for _, dir := range sortedDirs {
		modPath := pkgDirs[dir]
		var modRoot string
		for _, m := range modules {
			if m.ModulePath == modPath {
				modRoot = m.RootDir
				break
			}
		}

		relDir, _ := filepath.Rel(modRoot, dir)
		importPath := modPath
		if relDir != "." && relDir != "" {
			importPath = filepath.ToSlash(filepath.Join(modPath, relDir))
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		var files []*ast.File
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
				continue
			}
			fullPath := filepath.Join(dir, entry.Name())
			f, err := parser.ParseFile(fset, fullPath, nil, parser.ParseComments)
			if err != nil {
				typeErrors++
				if len(typeDiagnostics) < maxRecordedDiagnostics {
					typeDiagnostics = append(typeDiagnostics, fmt.Sprintf("parse: %s: %v", fullPath, err))
				}
				continue
			}
			files = append(files, f)
			result.Stats.Files++
		}

		if len(files) == 0 {
			continue
		}

		result.Stats.Packages++
		pkgName := files[0].Name.Name

		info := &types.Info{
			Types:      make(map[ast.Expr]types.TypeAndValue),
			Defs:       make(map[*ast.Ident]types.Object),
			Uses:       make(map[*ast.Ident]types.Object),
			Selections: make(map[*ast.SelectorExpr]*types.Selection),
		}

		conf := types.Config{
			Importer:                 imp,
			DisableUnusedImportCheck: true,
			Error: func(err error) {
				typeErrors++
				if len(typeDiagnostics) < maxRecordedDiagnostics {
					typeDiagnostics = append(typeDiagnostics, err.Error())
				}
			},
		}
		typesPkg, _ := conf.Check(importPath, fset, files, info)

		allPkgSuites = append(allPkgSuites, struct {
			importPath string
			files      []*ast.File
			info       *types.Info
			typesPkg   *types.Package
		}{
			importPath: importPath,
			files:      files,
			info:       info,
			typesPkg:   typesPkg,
		})

		pkgID := fmt.Sprintf("pkg:%s", importPath)
		pkgEntity := Entity{
			ID:          pkgID,
			Kind:        KindPackage,
			Name:        pkgName,
			Package:     pkgName,
			PackagePath: importPath,
			ModulePath:  modPath,
			File:        dir,
			Exported:    true,
			Language:    "go",
		}
		result.Entities = append(result.Entities, pkgEntity)
		entityMap[pkgID] = pkgEntity

		result.Relationships = append(result.Relationships, Relationship{
			From:             repoID,
			To:               pkgID,
			Type:             string(RelContains),
			Confidence:       1.0,
			ResolutionStatus: "RESOLVED",
			ResolutionMethod: "WORKSPACE_WALK",
			EpistemicClass:   "STATIC_FACT",
		})

		// Extract package-level entities
		for _, f := range files {
			filePos := fset.Position(f.Pos())
			relFile, _ := filepath.Rel(absRoot, filePos.Filename)

			// Imports
			for _, impSpec := range f.Imports {
				result.Stats.Imports++
				impPath := strings.Trim(impSpec.Path.Value, `"`)
				targetPkgID := fmt.Sprintf("pkg:%s", impPath)
				result.Relationships = append(result.Relationships, Relationship{
					From:             pkgID,
					To:               targetPkgID,
					Type:             string(RelImports),
					Confidence:       1.0,
					ResolutionStatus: "RESOLVED",
					ResolutionMethod: "AST_IMPORT",
					EpistemicClass:   "STATIC_FACT",
					Evidence: Evidence{
						File: relFile,
						Line: fset.Position(impSpec.Pos()).Line,
					},
				})
			}

			fileEntities := a.extractFileEntities(fset, f, pkgName, importPath, modPath, relFile)
			for _, e := range fileEntities {
				result.Entities = append(result.Entities, e)
				entityMap[e.ID] = e

				switch e.Kind {
				case KindStruct:
					result.Stats.Structs++
					result.Stats.TotalFields += len(e.Fields)
				case KindInterface:
					result.Stats.Interfaces++
				case KindFunction, KindMethod:
					result.Stats.Functions++
				}

				result.Relationships = append(result.Relationships, Relationship{
					From:             pkgID,
					To:               e.ID,
					Type:             string(RelContains),
					Confidence:       1.0,
					ResolutionStatus: "RESOLVED",
					ResolutionMethod: "AST_DECL",
					EpistemicClass:   "STATIC_FACT",
					Evidence: Evidence{
						File: e.File,
						Line: e.LineStart,
					},
				})

				// Struct embedding and field reference edges
				if e.Kind == KindStruct {
					for _, field := range e.Fields {
						cleanType := strings.TrimLeft(field.Type, "*[]")
						if cleanType == "" || isBuiltinGoType(cleanType) {
							continue
						}
						targetID := resolveTypeTargetID(importPath, cleanType, entityMap)
						if targetID != "" {
							relType := RelReferences
							if field.Name == cleanType || strings.HasSuffix(cleanType, "."+field.Name) {
								relType = RelEmbeds
							}
							result.Relationships = append(result.Relationships, Relationship{
								From:             e.ID,
								To:               targetID,
								Type:             string(relType),
								Confidence:       0.95,
								ResolutionStatus: "RESOLVED",
								ResolutionMethod: "STRUCT_FIELD_TYPE",
								EpistemicClass:   "STATIC_FACT",
								Evidence: Evidence{
									File: e.File,
									Line: field.LineStart,
								},
							})
						}
					}
				}

				// Method -> Receiver structDEFINES/CONTAINS link
				if e.Kind == KindMethod && e.ReceiverType != "" {
					recvClean := strings.TrimPrefix(e.ReceiverType, "*")
					recvID := fmt.Sprintf("%s.%s", importPath, recvClean)
					if _, exists := entityMap[recvID]; exists {
						result.Relationships = append(result.Relationships, Relationship{
							From:             recvID,
							To:               e.ID,
							Type:             string(RelDefines),
							Confidence:       1.0,
							ResolutionStatus: "RESOLVED",
							ResolutionMethod: "RECEIVER_BINDING",
							EpistemicClass:   "STATIC_FACT",
							Evidence: Evidence{
								File: e.File,
								Line: e.LineStart,
							},
						})
					}
				}
			}
		}
	}

	// Second pass: Extract CALLS and IMPLEMENTS across all packages
	for _, suite := range allPkgSuites {
		for _, f := range suite.files {
			filePos := fset.Position(f.Pos())
			relFile, _ := filepath.Rel(absRoot, filePos.Filename)
			callRels := a.extractCallsFromFile(fset, f, suite.importPath, relFile, suite.info, entityMap)
			result.Relationships = append(result.Relationships, callRels...)
		}
	}

	// Interface satisfaction (IMPLEMENTS) check using go/types
	implRels := extractImplementsRelationships(allPkgSuites, entityMap)
	result.Relationships = append(result.Relationships, implRels...)

	result.Stats.TypeErrors = typeErrors
	result.Stats.AnalysisComplete = (typeErrors == 0)
	result.TypeDiagnostics = typeDiagnostics
	result.Fingerprint = computeFingerprint(result)
	return result, nil
}
