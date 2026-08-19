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
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Analyze is an exported alias for Extract.
func Analyze(root string) (*Result, error) {
	return Extract(root)
}

// Extract analyzes a Go repository and returns a semantic Result conforming to model.go.
func Extract(root string) (*Result, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	fset := token.NewFileSet()
	var entities []Entity
	var relationships []Relationship
	var filesCount int
	pkgMap := make(map[string]bool)

	structMethods := make(map[string][]string)
	interfaceMethods := make(map[string][]string)
	relDedup := make(map[string]bool)

	addRelationship := func(rel Relationship) {
		dedupKey := fmt.Sprintf("%s|%s|%s", rel.From, rel.To, rel.Type)
		if !relDedup[dedupKey] {
			relDedup[dedupKey] = true
			relationships = append(relationships, rel)
		}
	}

	err = filepath.Walk(absRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		filesCount++

		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(absRoot, path)
		pkgName := node.Name.Name
		pkgMap[pkgName] = true

		entities = append(entities, Entity{
			ID:        fmt.Sprintf("%s:%s", pkgName, relPath),
			Name:      filepath.Base(relPath),
			Kind:      KindFile,
			Package:   pkgName,
			File:      relPath,
			Line:      fset.Position(node.Pos()).Line,
			LineStart: fset.Position(node.Pos()).Line,
			LineEnd:   fset.Position(node.End()).Line,
		})

		ast.Inspect(node, func(n ast.Node) bool {
			switch decl := n.(type) {
			case *ast.GenDecl:
				for _, spec := range decl.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok {
						lineStart := fset.Position(ts.Pos()).Line
						lineEnd := fset.Position(ts.End()).Line

						switch t := ts.Type.(type) {
						case *ast.StructType:
							var fields []Field
							if t.Fields != nil {
								for _, f := range t.Fields.List {
									fType := exprToString(f.Type)
									tagVal := ""
									if f.Tag != nil {
										tagVal = f.Tag.Value
									}

									// Anonymous embedded field
									if len(f.Names) == 0 {
										cleanEmb := cleanTypeName(fType)
										if cleanEmb != "" && !isBuiltin(cleanEmb) {
											addRelationship(Relationship{
												From:           ts.Name.Name,
												To:             cleanEmb,
												Type:           string(RelEmbeds),
												Confidence:     1.0,
												EpistemicClass: "OBSERVATION",
												Evidence: Evidence{
													File:      relPath,
													Line:      fset.Position(f.Pos()).Line,
													LineStart: fset.Position(f.Pos()).Line,
													LineEnd:   fset.Position(f.End()).Line,
													Analyzer:  "ast",
												},
											})
										}
									}

									for _, fn := range f.Names {
										fields = append(fields, Field{
											Name:      fn.Name,
											Type:      fType,
											Tag:       tagVal,
											LineStart: fset.Position(f.Pos()).Line,
											LineEnd:   fset.Position(f.End()).Line,
										})
									}
								}
							}

							entities = append(entities, Entity{
								ID:        fmt.Sprintf("%s:struct:%s", pkgName, ts.Name.Name),
								Name:      ts.Name.Name,
								Kind:      KindStruct,
								Package:   pkgName,
								File:      relPath,
								Exported:  ts.Name.IsExported(),
								Line:      lineStart,
								LineStart: lineStart,
								LineEnd:   lineEnd,
								Fields:    fields,
							})
							if _, exists := structMethods[ts.Name.Name]; !exists {
								structMethods[ts.Name.Name] = []string{}
							}

						case *ast.InterfaceType:
							var methods []string
							var methodEntities []Method
							if t.Methods != nil {
								for _, m := range t.Methods.List {
									if len(m.Names) > 0 {
										mName := m.Names[0].Name
										methods = append(methods, mName)
										methodEntities = append(methodEntities, Method{
											Name:       mName,
											Signature:  exprToString(m.Type),
											IsExported: m.Names[0].IsExported(),
											LineStart:  fset.Position(m.Pos()).Line,
											LineEnd:    fset.Position(m.End()).Line,
										})
									}
								}
							}
							interfaceMethods[ts.Name.Name] = methods

							entities = append(entities, Entity{
								ID:        fmt.Sprintf("%s:interface:%s", pkgName, ts.Name.Name),
								Name:      ts.Name.Name,
								Kind:      KindInterface,
								Package:   pkgName,
								File:      relPath,
								Exported:  ts.Name.IsExported(),
								Line:      lineStart,
								LineStart: lineStart,
								LineEnd:   lineEnd,
								Methods:   methodEntities,
							})

						default:
							kind := KindType
							if ts.Assign.IsValid() {
								kind = KindAlias
							}
							entities = append(entities, Entity{
								ID:        fmt.Sprintf("%s:%s:%s", pkgName, kind, ts.Name.Name),
								Name:      ts.Name.Name,
								Kind:      kind,
								Package:   pkgName,
								File:      relPath,
								Exported:  ts.Name.IsExported(),
								Line:      lineStart,
								LineStart: lineStart,
								LineEnd:   lineEnd,
							})
						}
					}
				}

			case *ast.FuncDecl:
				lineStart := fset.Position(decl.Pos()).Line
				lineEnd := fset.Position(decl.End()).Line
				funcName := decl.Name.Name

				typeParams := make(map[string]bool)
				if decl.Type != nil && decl.Type.TypeParams != nil {
					for _, tp := range decl.Type.TypeParams.List {
						for _, name := range tp.Names {
							typeParams[name.Name] = true
						}
					}
				}

				if decl.Recv != nil && len(decl.Recv.List) > 0 {
					recvType := exprToString(decl.Recv.List[0].Type)
					cleanRecv := strings.TrimPrefix(recvType, "*")

					entities = append(entities, Entity{
						ID:           fmt.Sprintf("%s:method:%s.%s", pkgName, cleanRecv, funcName),
						Name:         funcName,
						Kind:         KindMethod,
						Package:      pkgName,
						ReceiverType: recvType,
						File:         relPath,
						Exported:     decl.Name.IsExported(),
						Line:         lineStart,
						LineStart:    lineStart,
						LineEnd:      lineEnd,
					})

					structMethods[cleanRecv] = append(structMethods[cleanRecv], funcName)

					addRelationship(Relationship{
						From:           cleanRecv,
						To:             funcName,
						Type:           string(RelDefines),
						Confidence:     1.0,
						EpistemicClass: "OBSERVATION",
						Evidence: Evidence{
							File:      relPath,
							Line:      lineStart,
							LineStart: lineStart,
							LineEnd:   lineEnd,
							Analyzer:  "ast",
						},
					})
				} else {
					entities = append(entities, Entity{
						ID:        fmt.Sprintf("%s:func:%s", pkgName, funcName),
						Name:      funcName,
						Kind:      KindFunction,
						Package:   pkgName,
						File:      relPath,
						Exported:  decl.Name.IsExported(),
						Line:      lineStart,
						LineStart: lineStart,
						LineEnd:   lineEnd,
					})
				}

				if decl.Type != nil {
					if decl.Type.Params != nil {
						for _, p := range decl.Type.Params.List {
							tName := cleanTypeName(exprToString(p.Type))
							if tName != "" && !isBuiltin(tName) && !typeParams[tName] {
								addRelationship(Relationship{
									From:           funcName,
									To:             tName,
									Type:           string(RelReferences),
									Confidence:     1.0,
									EpistemicClass: "OBSERVATION",
									Evidence: Evidence{
										File:      relPath,
										Line:      fset.Position(p.Pos()).Line,
										LineStart: fset.Position(p.Pos()).Line,
										LineEnd:   fset.Position(p.End()).Line,
										Analyzer:  "ast",
									},
								})
							}
						}
					}
					if decl.Type.Results != nil {
						for _, r := range decl.Type.Results.List {
							tName := cleanTypeName(exprToString(r.Type))
							if tName != "" && !isBuiltin(tName) && !typeParams[tName] {
								addRelationship(Relationship{
									From:           funcName,
									To:             tName,
									Type:           string(RelReferences),
									Confidence:     1.0,
									EpistemicClass: "OBSERVATION",
									Evidence: Evidence{
										File:      relPath,
										Line:      fset.Position(r.Pos()).Line,
										LineStart: fset.Position(r.Pos()).Line,
										LineEnd:   fset.Position(r.End()).Line,
										Analyzer:  "ast",
									},
								})
							}
						}
					}
				}

				if decl.Body != nil {
					ast.Inspect(decl.Body, func(inner ast.Node) bool {
						if compLit, ok := inner.(*ast.CompositeLit); ok {
							tName := cleanTypeName(exprToString(compLit.Type))
							if tName != "" && !isBuiltin(tName) && !typeParams[tName] {
								addRelationship(Relationship{
									From:           funcName,
									To:             tName,
									Type:           string(RelReferences),
									Confidence:     1.0,
									EpistemicClass: "OBSERVATION",
									Evidence: Evidence{
										File:      relPath,
										Line:      fset.Position(compLit.Pos()).Line,
										LineStart: fset.Position(compLit.Pos()).Line,
										LineEnd:   fset.Position(compLit.End()).Line,
										Analyzer:  "ast",
									},
								})
							}
						}
						return true
					})
				}
			}
			return true
		})
		return nil
	})

	if err != nil {
		return nil, err
	}

	for iface, ifaceMethodsList := range interfaceMethods {
		if len(ifaceMethodsList) == 0 {
			continue
		}
		for strct, strctMethodsList := range structMethods {
			if structImplements(strctMethodsList, ifaceMethodsList) {
				addRelationship(Relationship{
					From:           strct,
					To:             iface,
					Type:           string(RelImplements),
					Confidence:     1.0,
					EpistemicClass: "OBSERVATION",
					Evidence: Evidence{
						Analyzer: "ast",
					},
				})
			}
		}
	}

	var structCount, ifaceCount, funcCount, totalFields int
	for _, e := range entities {
		switch e.Kind {
		case KindStruct:
			structCount++
			totalFields += len(e.Fields)
		case KindInterface:
			ifaceCount++
		case KindFunction:
			funcCount++
		}
	}

	result := &Result{
		Entities:      entities,
		Relationships: relationships,
		AnalyzedAt:    time.Now().UTC(),
		Stats: Stats{
			Files:       filesCount,
			Packages:    len(pkgMap),
			Structs:     structCount,
			Interfaces:  ifaceCount,
			Functions:   funcCount,
			TotalFields: totalFields,
		},
	}

	return result, nil
}

func structImplements(structMethods, ifaceMethods []string) bool {
	methodMap := make(map[string]bool)
	for _, m := range structMethods {
		methodMap[m] = true
	}
	for _, m := range ifaceMethods {
		if !methodMap[m] {
			return false
		}
	}
	return true
}

func cleanTypeName(name string) string {
	name = strings.TrimPrefix(name, "*")
	name = strings.TrimPrefix(name, "[]")
	name = strings.TrimPrefix(name, "...")
	if idx := strings.Index(name, "["); idx != -1 {
		name = name[:idx]
	}
	return name
}

func exprToString(expr ast.Expr) string {
	if expr == nil {
		return ""
	}
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + exprToString(t.X)
	case *ast.ArrayType:
		return "[]" + exprToString(t.Elt)
	case *ast.Ellipsis:
		return "..." + exprToString(t.Elt)
	case *ast.IndexExpr:
		return exprToString(t.X)
	case *ast.IndexListExpr:
		return exprToString(t.X)
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	default:
		return ""
	}
}

func isBuiltin(name string) bool {
	switch strings.TrimPrefix(name, "*") {
	case "string", "int", "int64", "int32", "int16", "int8",
		"uint", "uint64", "uint32", "uint16", "uint8", "uintptr",
		"bool", "float64", "float32", "byte", "rune", "error", "any":
		return true
	default:
		return false
	}
}
