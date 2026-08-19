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
)

func Extract(root string) (*Result, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	fmt.Printf("DEBUG: Walking directory %s\n", absRoot)

	var goFiles []string
	err = filepath.Walk(absRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			name := info.Name()
			if name == "vendor" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			goFiles = append(goFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}
	if len(goFiles) == 0 {
		return nil, fmt.Errorf("no Go files found in '%s'", root)
	}
	fmt.Printf("DEBUG: Found %d Go files\n", len(goFiles))

	commit := getCommit(root)
	analyzerVersion := "go-ast-v1"

	entityMap := make(map[string]Entity)
	relationList := []Relationship{}

	getPackageName := func(filename string) string {
		fset := token.NewFileSet()
		content, err := os.ReadFile(filename)
		if err != nil {
			return "main"
		}
		fileAst, err := parser.ParseFile(fset, filename, content, parser.PackageClauseOnly)
		if err != nil {
			return "main"
		}
		if fileAst.Name != nil {
			return fileAst.Name.Name
		}
		return "main"
	}

	getEntityID := func(pkg, name string, kind EntityKind) string {
		if pkg == "" {
			pkg = "main"
		}
		return pkg + "." + name
	}

	addEntity := func(e Entity) {
		if _, exists := entityMap[e.ID]; !exists {
			entityMap[e.ID] = e
		}
	}

	addRelation := func(from, to string, relType string, line int, file string) {
		if from == "" || to == "" {
			fmt.Printf("DEBUG: addRelation skipped (from=%q, to=%q)\n", from, to)
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
		fmt.Printf("DEBUG: addRelation %s %s -> %s\n", relType, from, to)
	}

	for _, filename := range goFiles {
		pkgName := getPackageName(filename)
		pkgPath := pkgName
		if pkgPath == "" {
			pkgPath = "main"
		}

		fmt.Printf("DEBUG: Analyzing file %s (package: %s)\n", filename, pkgName)

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
		addRelation(pkgEntityID, fileID, string(RelContains), 0, filename)

		fset := token.NewFileSet()
		content, err := os.ReadFile(filename)
		if err != nil {
			fmt.Printf("DEBUG: Failed to read file: %v\n", err)
			continue
		}
		fileAst, err := parser.ParseFile(fset, filename, content, parser.AllErrors)
		if err != nil {
			fmt.Printf("DEBUG: Failed to parse file: %v\n", err)
			continue
		}

		importedNames := make(map[string]bool)
		var currentFuncName string
		var currentFuncPkg string

		ast.Inspect(fileAst, func(n ast.Node) bool {
			if n == nil {
				return true
			}

			switch x := n.(type) {
			case *ast.FuncDecl:
				funcName := x.Name.Name
				funcPkg := pkgPath
				var funcKind EntityKind = KindFunction

				startPos := fset.Position(x.Pos())
				endPos := fset.Position(x.End())

				fmt.Printf("DEBUG: Found function %s at line %d-%d in package %s\n", funcName, startPos.Line, endPos.Line, funcPkg)

				if x.Recv != nil {
					funcKind = KindMethod
					funcID := getEntityID(funcPkg, funcName, funcKind)
					addEntity(Entity{
						ID:        funcID,
						Kind:      funcKind,
						Name:      funcName,
						Package:   funcPkg,
						File:      filename,
						Line:      startPos.Line,
						LineStart: startPos.Line,
						LineEnd:   endPos.Line,
						Exported:  x.Name.IsExported(),
						Signature: exprToString(x.Type),
					})
					addRelation(fileID, funcID, string(RelContains), startPos.Line, filename)
					if recv := x.Recv; recv != nil && len(recv.List) > 0 {
						recvType := exprToString(recv.List[0].Type)
						recvID := getEntityID(funcPkg, recvType, KindStruct)
						addRelation(funcID, recvID, string(RelReferences), startPos.Line, filename)
					}
				} else {
					funcID := getEntityID(funcPkg, funcName, funcKind)
					addEntity(Entity{
						ID:        funcID,
						Kind:      funcKind,
						Name:      funcName,
						Package:   funcPkg,
						File:      filename,
						Line:      startPos.Line,
						LineStart: startPos.Line,
						LineEnd:   endPos.Line,
						Exported:  x.Name.IsExported(),
						Signature: exprToString(x.Type),
					})
					addRelation(fileID, funcID, string(RelContains), startPos.Line, filename)
				}
				currentFuncName = funcName
				currentFuncPkg = funcPkg

			case *ast.ImportSpec:
				importPath := strings.Trim(x.Path.Value, `"`)
				if importPath == "" {
					return true
				}
				line := fset.Position(x.Pos()).Line
				fmt.Printf("DEBUG: Direct ImportSpec: %s at line %d\n", importPath, line)
				name := importPath
				if x.Name != nil {
					name = x.Name.Name
				} else {
					name = filepath.Base(importPath)
				}
				importedNames[name] = true
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
				addRelation(pkgEntityID, extPkgID, string(RelImports), line, filename)

			case *ast.GenDecl:
				if x.Tok == token.IMPORT {
					fmt.Printf("DEBUG: Found import block in %s\n", filename)
					for _, spec := range x.Specs {
						if impSpec, ok := spec.(*ast.ImportSpec); ok {
							importPath := strings.Trim(impSpec.Path.Value, `"`)
							if importPath == "" {
								continue
							}
							line := fset.Position(impSpec.Pos()).Line
							fmt.Printf("DEBUG: Import path: %s at line %d\n", importPath, line)
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
							addRelation(pkgEntityID, extPkgID, string(RelImports), line, filename)
						}
					}
				}

			case *ast.TypeSpec:
				startPos := fset.Position(x.Pos())
				endPos := fset.Position(x.End())

				switch t := x.Type.(type) {
				case *ast.StructType:
					structName := x.Name.Name
					structID := getEntityID(pkgPath, structName, KindStruct)
					addEntity(Entity{
						ID:        structID,
						Kind:      KindStruct,
						Name:      structName,
						Package:   pkgPath,
						File:      filename,
						Line:      startPos.Line,
						LineStart: startPos.Line,
						LineEnd:   endPos.Line,
						Exported:  x.Name.IsExported(),
					})
					addRelation(fileID, structID, string(RelDefines), startPos.Line, filename)

					for _, field := range t.Fields.List {
						if field.Type != nil {
							fieldType := exprToString(field.Type)
							fieldLine := fset.Position(field.Pos()).Line
							if fieldType != "" {
								fieldID := getEntityID(pkgPath, fieldType, KindExternal)
								addRelation(structID, fieldID, string(RelReferences), fieldLine, filename)
							}
						}
					}
				case *ast.InterfaceType:
					ifaceName := x.Name.Name
					ifaceID := getEntityID(pkgPath, ifaceName, KindInterface)
					addEntity(Entity{
						ID:        ifaceID,
						Kind:      KindInterface,
						Name:      ifaceName,
						Package:   pkgPath,
						File:      filename,
						Line:      startPos.Line,
						LineStart: startPos.Line,
						LineEnd:   endPos.Line,
						Exported:  x.Name.IsExported(),
					})
					addRelation(fileID, ifaceID, string(RelDefines), startPos.Line, filename)
				}

			case *ast.CallExpr:
				if currentFuncName == "" {
					return true
				}
				calledName := resolveCallTarget(x.Fun, importedNames)
				if calledName == "" {
					return true
				}
				line := fset.Position(x.Pos()).Line
				fmt.Printf("DEBUG: CallExpr in function %s -> %s at line %d\n", currentFuncName, calledName, line)
				calledID := getEntityID(currentFuncPkg, calledName, KindExternal)
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
				callerID := getEntityID(currentFuncPkg, currentFuncName, KindFunction)
				addRelation(callerID, calledID, string(RelCalls), line, filename)
			}
			return true
		})
	}

	entities := make([]Entity, 0, len(entityMap))
	for _, e := range entityMap {
		entities = append(entities, e)
	}

	stats := Stats{
		Files:       len(goFiles),
		Packages:    countPackages(entities),
		Structs:     countByKind(entities, KindStruct),
		Interfaces:  countByKind(entities, KindInterface),
		Functions:   countByKind(entities, KindFunction),
		Imports:     countImports(relationList),
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

func countPackages(entities []Entity) int {
	m := make(map[string]bool)
	for _, e := range entities {
		if e.Kind == KindPackage || e.Kind == KindFile {
			m[e.Package] = true
		}
	}
	return len(m)
}

func countImports(rels []Relationship) int {
	c := 0
	for _, r := range rels {
		if r.Type == string(RelImports) {
			c++
		}
	}
	return c
}

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

func resolveCallTarget(expr ast.Expr, importedNames map[string]bool) string {
	switch fun := expr.(type) {
	case *ast.Ident:
		return fun.Name
	case *ast.SelectorExpr:
		base := resolveCallTarget(fun.X, importedNames)
		if base == "" {
			return fun.Sel.Name
		}
		if importedNames[base] {
			return base + "." + fun.Sel.Name
		}
		return fun.Sel.Name
	case *ast.IndexExpr:
		return resolveCallTarget(fun.X, importedNames)
	default:
		return exprToString(expr)
	}
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
