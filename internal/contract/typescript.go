// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package contract

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

type TypeExporter struct {
	fset *token.FileSet
}

func NewTypeExporter() *TypeExporter {
	return &TypeExporter{
		fset: token.NewFileSet(),
	}
}

// GenerateTypeScript parses Go source and produces TypeScript interface contracts.
func (e *TypeExporter) GenerateTypeScript(sourceCode string) (string, error) {
	node, err := parser.ParseFile(e.fset, "", sourceCode, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("failed to parse Go source: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString("/**\n * Garuda Verified Interface Contract\n")
	buf.WriteString(" * Deterministically derived from Go AST\n */\n\n")

	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			buf.WriteString(fmt.Sprintf("export interface %s {\n", typeSpec.Name.Name))
			for _, field := range structType.Fields.List {
				fieldName := ""
				if len(field.Names) > 0 {
					fieldName = field.Names[0].Name
				} else {
					continue // anonymous/embedded
				}

				// Check json tag if present
				jsonName := fieldName
				if field.Tag != nil {
					tagVal := strings.Trim(field.Tag.Value, "`")
					for _, tag := range strings.Split(tagVal, " ") {
						if strings.HasPrefix(tag, `json:"`) {
							parts := strings.Split(strings.Trim(tag[6:], `"`), ",")
							if parts[0] != "" && parts[0] != "-" {
								jsonName = parts[0]
							}
						}
					}
				}

				tsType := e.mapGoTypeToTS(field.Type)
				buf.WriteString(fmt.Sprintf("  %s: %s;\n", jsonName, tsType))
			}
			buf.WriteString("}\n\n")
		}
	}

	return buf.String(), nil
}

func (e *TypeExporter) mapGoTypeToTS(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		switch t.Name {
		case "string":
			return "string"
		case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64":
			return "number"
		case "bool":
			return "boolean"
		default:
			return t.Name
		}
	case *ast.ArrayType:
		elemType := e.mapGoTypeToTS(t.Elt)
		return elemType + "[]"
	case *ast.MapType:
		keyType := e.mapGoTypeToTS(t.Key)
		valType := e.mapGoTypeToTS(t.Value)
		return fmt.Sprintf("Record<%s, %s>", keyType, valType)
	case *ast.StarExpr:
		return e.mapGoTypeToTS(t.X) + " | null"
	default:
		return "any"
	}
}
