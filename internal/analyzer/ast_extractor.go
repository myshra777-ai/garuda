package analyzer

import (
	"fmt"
	"go/ast"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/tools/go/packages"
)

// Extract extracts entities and relationships from a Go repository
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
		return nil, fmt.Errorf("no Go packages found in %s", root)
	}

	result := &Result{
		Source:        root,
		AnalyzedAt:    time.Now().UTC(),
		Entities:      []Entity{},
		Relationships: []Relationship{},
		Stats:         Stats{},
	}

	entityMap := make(map[string]*Entity)

	for _, pkg := range pkgs {
		if strings.HasSuffix(pkg.PkgPath, "_test") {
			continue
		}
		result.Stats.Packages++

		for _, file := range pkg.Syntax {
			result.Stats.Files++

			// 1st Pass: Discover Structs, Interfaces, Standalone Functions
			ast.Inspect(file, func(n ast.Node) bool {
				if n == nil {
					return true
				}

				pos := pkg.Fset.Position(n.Pos())
				filename := filepath.Base(pos.Filename)

				switch x := n.(type) {
				case *ast.TypeSpec:
					if _, ok := x.Type.(*ast.StructType); ok {
						entity := extractStruct(pkg, x, filename)
						entityMap[entity.ID] = &entity
						result.Stats.Structs++
					}
					if _, ok := x.Type.(*ast.InterfaceType); ok {
						entity := extractInterface(pkg, x, filename)
						entityMap[entity.ID] = &entity
						result.Stats.Interfaces++
					}
				case *ast.FuncDecl:
					if x.Recv == nil { // Pure standalone function
						entity := extractFunction(pkg, x, filename)
						entityMap[entity.ID] = &entity
						result.Stats.Functions++
					}
				}
				return true
			})

			// 2nd Pass: Attach struct receiver methods
			ast.Inspect(file, func(n ast.Node) bool {
				fd, ok := n.(*ast.FuncDecl)
				if !ok || fd.Recv == nil || len(fd.Recv.List) == 0 {
					return true
				}

				recvType := exprToString(fd.Recv.List[0].Type)
				recvType = strings.TrimPrefix(recvType, "*") // Strip pointer prefix if present

				entityID := fmt.Sprintf("%s.%s", pkg.PkgPath, recvType)
				if structEntity, exists := entityMap[entityID]; exists {
					structEntity.Methods = append(structEntity.Methods, Method{
						Name:       fd.Name.Name,
						Signature:  exprToString(fd.Type),
						IsExported: fd.Name.IsExported(),
					})
				}
				return true
			})
		}
	}

	// Build Relationships (Package Imports)
	for _, pkg := range pkgs {
		for _, imp := range pkg.Imports {
			result.Relationships = append(result.Relationships, Relationship{
				From: pkg.PkgPath,
				To:   imp.PkgPath,
				Type: "imports",
			})
			result.Stats.Imports++
		}
	}

	// Transfer entities from map to result slice
	for _, e := range entityMap {
		result.Entities = append(result.Entities, *e)
		result.Stats.TotalFields += len(e.Fields)
	}

	result.Fingerprint = generateFingerprint(result)
	return result, nil
}

func extractStruct(pkg *packages.Package, ts *ast.TypeSpec, filename string) Entity {
	e := Entity{
		Package:    pkg.PkgPath,
		File:       filename,
		Kind:       "struct",
		Name:       ts.Name.Name,
		IsExported: ast.IsExported(ts.Name.Name),
		Comments:   extractComments(ts.Doc),
		Fields:     []Field{},
		Methods:    []Method{},
	}
	e.ID = fmt.Sprintf("%s.%s", pkg.PkgPath, ts.Name.Name)

	if st, ok := ts.Type.(*ast.StructType); ok {
		for _, field := range st.Fields.List {
			if len(field.Names) == 0 { // Embedded type
				f := Field{
					Name:      exprToString(field.Type),
					Type:      exprToString(field.Type),
					IsPointer: isPointer(field.Type),
				}
				e.Fields = append(e.Fields, f)
				continue
			}
			for _, name := range field.Names {
				f := Field{
					Name:      name.Name,
					Type:      exprToString(field.Type),
					Comment:   extractComments(field.Doc),
					IsPointer: isPointer(field.Type),
					IsSlice:   isSlice(field.Type),
				}
				if field.Tag != nil {
					tag := field.Tag.Value
					f.Tag = strings.Trim(tag, "`")
					f.JSONTag = parseTag(tag, "json")
					f.DBTag = parseTag(tag, "db")
					f.ValidateTag = parseTag(tag, "validate")
				}
				e.Fields = append(e.Fields, f)
			}
		}
	}
	return e
}

func extractInterface(pkg *packages.Package, ts *ast.TypeSpec, filename string) Entity {
	e := Entity{
		Package:    pkg.PkgPath,
		File:       filename,
		Kind:       "interface",
		Name:       ts.Name.Name,
		IsExported: ast.IsExported(ts.Name.Name),
		Comments:   extractComments(ts.Doc),
		Methods:    []Method{},
	}
	e.ID = fmt.Sprintf("%s.%s", pkg.PkgPath, ts.Name.Name)

	if iface, ok := ts.Type.(*ast.InterfaceType); ok {
		for _, method := range iface.Methods.List {
			if len(method.Names) == 0 {
				continue
			}
			for _, name := range method.Names {
				e.Methods = append(e.Methods, Method{
					Name:       name.Name,
					Signature:  exprToString(method.Type),
					IsExported: ast.IsExported(name.Name),
				})
			}
		}
	}
	return e
}

func extractFunction(pkg *packages.Package, fd *ast.FuncDecl, filename string) Entity {
	e := Entity{
		Package:    pkg.PkgPath,
		File:       filename,
		Kind:       "func",
		Name:       fd.Name.Name,
		IsExported: fd.Name.IsExported(),
		Comments:   extractComments(fd.Doc),
	}
	e.ID = fmt.Sprintf("%s.%s", pkg.PkgPath, fd.Name.Name)
	return e
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
	case *ast.FuncType:
		return "func(...)"
	default:
		return fmt.Sprintf("%T", expr)
	}
}

func isPointer(expr ast.Expr) bool {
	_, ok := expr.(*ast.StarExpr)
	return ok
}

func isSlice(expr ast.Expr) bool {
	if arr, ok := expr.(*ast.ArrayType); ok {
		return arr.Len == nil
	}
	return false
}

func parseTag(tagValue, key string) string {
	if tagValue == "" {
		return ""
	}
	tag := strings.Trim(tagValue, "`")
	parts := strings.Fields(tag)
	for _, part := range parts {
		if strings.HasPrefix(part, key+":") {
			kv := strings.SplitN(part, ":", 2)
			if len(kv) == 2 {
				return strings.Trim(kv[1], `"`)
			}
		}
	}
	return ""
}

func extractComments(cg *ast.CommentGroup) string {
	if cg == nil {
		return ""
	}
	var lines []string
	for _, c := range cg.List {
		lines = append(lines, strings.TrimPrefix(c.Text, "// "))
	}
	return strings.Join(lines, "\n")
}
