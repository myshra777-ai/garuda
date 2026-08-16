package analyzer

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/tools/go/packages"
)

// Extract analyses a Go repository and returns entities and relationships.
func Extract(root string) (*Result, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo |
			packages.NeedModule |
			packages.NeedImports,
		Dir: root,
	}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no Go packages found")
	}

	commit := getCommit(root)
	analyzerVersion := "go-ast-v1"

	entityMap := make(map[string]Entity)
	relationList := []Relationship{}

	getEntityID := func(pkg, name string, kind string) string {
		return pkg + "." + name
	}

	addEntity := func(e Entity) {
		if _, exists := entityMap[e.ID]; !exists {
			entityMap[e.ID] = e
		}
	}

	addRelation := func(from, to string, relType RelationshipType, line int, file string) {
		if from == "" || to == "" {
			return
		}
		relationList = append(relationList, Relationship{
			From:       from,
			To:         to,
			Type:       relType,
			Confidence: 1.0,
			Evidence: Evidence{
				File:     file,
				Line:     line,
				Commit:   commit,
				Analyzer: analyzerVersion,
			},
		})
	}

	for _, pkg := range pkgs {
		if strings.HasSuffix(pkg.PkgPath, "_test") {
			continue
		}
		pkgPath := pkg.PkgPath
		pkgEntityID := pkgPath
		if _, exists := entityMap[pkgEntityID]; !exists {
			entityMap[pkgEntityID] = Entity{
				ID:       pkgEntityID,
				Kind:     KindPackage,
				Name:     pkgPath,
				Package:  pkgPath,
				File:     "",
				Exported: true,
			}
		}

		// Iterate over compiled Go files (not syntax, to get correct filenames)
		for _, filename := range pkg.CompiledGoFiles {
			// Skip test files if desired
			if strings.HasSuffix(filename, "_test.go") {
				continue
			}
			fileID := filename
			if _, exists := entityMap[fileID]; !exists {
				entityMap[fileID] = Entity{
					ID:       fileID,
					Kind:     KindFile,
					Name:     filepath.Base(filename),
					Package:  pkgPath,
					File:     filename,
					Exported: true,
				}
			}
			addRelation(pkgEntityID, fileID, RelContains, 0, filename)

			fset := token.NewFileSet()
			content, err := os.ReadFile(filename)
			if err != nil {
				continue
			}
			fileAst, err := parser.ParseFile(fset, filename, content, parser.AllErrors)
			if err != nil {
				continue
			}

			var currentFuncName, currentFuncPkg string

			ast.Inspect(fileAst, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.FuncDecl:
					funcName := x.Name.Name
					funcPkg := pkgPath
					funcKind := KindFunction
					funcID := getEntityID(funcPkg, funcName, string(funcKind))
					if x.Recv != nil {
						funcKind = KindMethod
						funcID = getEntityID(funcPkg, funcName, string(KindMethod))
						addEntity(Entity{
							ID:        funcID,
							Kind:      KindMethod,
							Name:      funcName,
							Package:   funcPkg,
							File:      filename,
							Line:      fset.Position(x.Pos()).Line,
							Exported:  x.Name.IsExported(),
							Signature: exprToString(x.Type),
						})
						addRelation(fileID, funcID, RelContains, fset.Position(x.Pos()).Line, filename)
						if recv := x.Recv; recv != nil && len(recv.List) > 0 {
							recvType := exprToString(recv.List[0].Type)
							recvID := getEntityID(funcPkg, recvType, string(KindStruct))
							addRelation(funcID, recvID, RelReferences, fset.Position(x.Pos()).Line, filename)
						}
					} else {
						addEntity(Entity{
							ID:        funcID,
							Kind:      KindFunction,
							Name:      funcName,
							Package:   funcPkg,
							File:      filename,
							Line:      fset.Position(x.Pos()).Line,
							Exported:  x.Name.IsExported(),
							Signature: exprToString(x.Type),
						})
						addRelation(fileID, funcID, RelContains, fset.Position(x.Pos()).Line, filename)
					}
					currentFuncName = funcName
					currentFuncPkg = funcPkg

				case *ast.TypeSpec:
					switch t := x.Type.(type) {
					case *ast.StructType:
						structName := x.Name.Name
						structID := getEntityID(pkgPath, structName, string(KindStruct))
						addEntity(Entity{
							ID:       structID,
							Kind:     KindStruct,
							Name:     structName,
							Package:  pkgPath,
							File:     filename,
							Line:     fset.Position(x.Pos()).Line,
							Exported: x.Name.IsExported(),
						})
						addRelation(fileID, structID, RelDefines, fset.Position(x.Pos()).Line, filename)
						for _, field := range t.Fields.List {
							if field.Type != nil {
								fieldType := exprToString(field.Type)
								if fieldType != "" {
									fieldID := getEntityID(pkgPath, fieldType, string(KindExternal))
									addRelation(structID, fieldID, RelReferences, fset.Position(field.Pos()).Line, filename)
								}
							}
						}
					case *ast.InterfaceType:
						ifaceName := x.Name.Name
						ifaceID := getEntityID(pkgPath, ifaceName, string(KindInterface))
						addEntity(Entity{
							ID:       ifaceID,
							Kind:     KindInterface,
							Name:     ifaceName,
							Package:  pkgPath,
							File:     filename,
							Line:     fset.Position(x.Pos()).Line,
							Exported: x.Name.IsExported(),
						})
						addRelation(fileID, ifaceID, RelDefines, fset.Position(x.Pos()).Line, filename)
					}

				case *ast.ImportSpec:
					importPath := strings.Trim(x.Path.Value, `"`)
					extPkgID := importPath
					if _, exists := entityMap[extPkgID]; !exists {
						entityMap[extPkgID] = Entity{
							ID:       extPkgID,
							Kind:     KindExternal,
							Name:     importPath,
							Package:  importPath,
							File:     filename,
							Exported: true,
						}
					}
					addRelation(pkgEntityID, extPkgID, RelImports, fset.Position(x.Pos()).Line, filename)

				case *ast.CallExpr:
					if currentFuncName == "" {
						return true
					}
					var calledName string
					switch fun := x.Fun.(type) {
					case *ast.Ident:
						calledName = fun.Name
					case *ast.SelectorExpr:
						calledName = exprToString(fun)
					default:
						calledName = exprToString(fun)
					}
					if calledName != "" {
						calledID := getEntityID(currentFuncPkg, calledName, string(KindFunction))
						if _, exists := entityMap[calledID]; !exists {
							entityMap[calledID] = Entity{
								ID:       calledID,
								Kind:     KindExternal,
								Name:     calledName,
								Package:  currentFuncPkg,
								File:     filename,
								Exported: false,
							}
						}
						callerID := getEntityID(currentFuncPkg, currentFuncName, string(KindFunction))
						addRelation(callerID, calledID, RelCalls, fset.Position(x.Pos()).Line, filename)
					}
				}
				return true
			})
		}
	}

	entities := make([]Entity, 0, len(entityMap))
	for _, e := range entityMap {
		entities = append(entities, e)
	}

	stats := Stats{
		Files:       len(entityMap),
		Packages:    len(pkgs),
		Structs:     countByKind(entities, KindStruct),
		Interfaces:  countByKind(entities, KindInterface),
		Functions:   countByKind(entities, KindFunction),
		Imports:     0,
		TotalFields: 0,
	}

	result := &Result{
		Entities:      entities,
		Relationships: relationList,
		AnalyzedAt:    time.Now().UTC(),
		Package:       root,
		Source:        root,
		Stats:         stats,
	}
	result.Fingerprint = generateFingerprint(result)
	return result, nil
}

// -----------------------------------------------------------------------------
// Helpers (unchanged)
// -----------------------------------------------------------------------------

func countByKind(entities []Entity, kind EntityKind) int {
	c := 0
	for _, e := range entities {
		if e.Kind == kind {
			c++
		}
	}
	return c
}

func getCommit(root string) string {
	cmd := exec.Command("git", "-C", root, "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + exprToString(t.X)
	case *ast.ArrayType:
		return "[]" + exprToString(t.Elt)
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	case *ast.MapType:
		return "map[" + exprToString(t.Key) + "]" + exprToString(t.Value)
	case *ast.ChanType:
		return "chan " + exprToString(t.Value)
	case *ast.ParenExpr:
		return "(" + exprToString(t.X) + ")"
	default:
		return fmt.Sprintf("%T", expr)
	}
}
