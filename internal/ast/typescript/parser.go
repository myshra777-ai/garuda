// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

// Package typescript extracts semantic entities and relationships from
// TypeScript source using tree-sitter. It mirrors the shape of the
// Python parser: a Module per file, containing Classes, Interfaces,
// Functions, and Imports.
package typescript

import (
	"context"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/typescript/typescript"

	"github.com/myshra777-ai/garuda/internal/ast/treesitter"
)

// Module is the parsed representation of a single TypeScript file.
type Module struct {
	Path       string
	Module     string // derived from file path
	Classes    []Class
	Interfaces []Interface
	Functions  []Function
	Imports    []Import
}

type Class struct {
	Name       string
	Extends    string   // base class
	Implements []string // implemented interfaces
	Decorators []string
	LineStart  int
	LineEnd    int
	IsExported bool
	IsAbstract bool
	Methods    []Function
	Properties []Property
}

type Interface struct {
	Name       string
	Extends    []string
	LineStart  int
	LineEnd    int
	IsExported bool
	Properties []Property
}

type Function struct {
	Name       string
	Params     []string
	ReturnType string
	Decorators []string
	IsAsync    bool
	IsArrow    bool
	Receiver   string // enclosing class/interface if method
	LineStart  int
	LineEnd    int
	IsExported bool
}

type Property struct {
	Name       string
	Type       string
	IsOptional bool
	IsReadonly bool
}

type Import struct {
	Source    string   // module path
	Names     []string // named imports
	Default   string   // default import name
	Namespace string   // * as X
	Line      int
}

// Parser wraps a tree-sitter parser configured for the TypeScript grammar.
type Parser struct {
	inner *treesitter.Parser
}

// NewParser constructs a TypeScript parser.
func NewParser() *Parser {
	return &Parser{
		inner: treesitter.Wrap(treesitter.LanguageTypeScript, typescript.GetLanguage()),
	}
}

// Close releases parser resources.
func (p *Parser) Close() {
	if p != nil && p.inner != nil {
		p.inner.Close()
	}
}

// Parse parses source bytes and returns a structured Module.
func (p *Parser) Parse(ctx context.Context, src []byte, path, moduleName string) (*Module, error) {
	parsed, err := p.inner.Parse(ctx, src)
	if err != nil {
		return nil, err
	}
	defer parsed.Tree.Close()

	m := &Module{Path: path, Module: moduleName}

	parsed.Walk(func(n *sitter.Node) bool {
		switch n.Type() {
		case "class_declaration", "abstract_class_declaration":
			cls := extractClass(n, parsed)
			m.Classes = append(m.Classes, cls)
			return false // don't descend; we already extracted methods
		case "interface_declaration":
			iface := extractInterface(n, parsed)
			m.Interfaces = append(m.Interfaces, iface)
			return false
		case "function_declaration":
			fn := extractFunction(n, parsed, "")
			fn.IsExported = isExportedNode(n)
			m.Functions = append(m.Functions, fn)
			return false
		case "import_statement":
			imp := extractImport(n, parsed)
			m.Imports = append(m.Imports, imp)
			return false
		}
		return true
	})

	// Handle exported declarations: export statement wraps the inner decl.
	// We catch these on the second pass so we don't double-count.
	parsed.Walk(func(n *sitter.Node) bool {
		if n.Type() != "export_statement" {
			return true
		}
		// export_statement contains either:
		//   - a declaration (class/interface/function/lexical)
		//   - a re-export (export { x } from 'y')
		for i := 0; i < int(n.NamedChildCount()); i++ {
			child := n.NamedChild(i)
			if child == nil {
				continue
			}
			switch child.Type() {
			case "class_declaration", "abstract_class_declaration":
				markExported(&m.Classes, child, parsed)
			case "interface_declaration":
				markExportedInterface(&m.Interfaces, child, parsed)
			case "function_declaration":
				markExportedFunction(&m.Functions, child, parsed)
			case "lexical_declaration", "variable_declaration":
				// export const foo = () => ... — treat arrow functions as functions
				extractExportedArrow(child, parsed, m)
			}
		}
		return true
	})

	return m, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Extractors
// ─────────────────────────────────────────────────────────────────────────────

func extractClass(n *sitter.Node, f *treesitter.ParsedFile) Class {
	cls := Class{
		LineStart:  treesitter.NodeLine(n),
		LineEnd:    treesitter.NodeEndLine(n),
		IsAbstract: n.Type() == "abstract_class_declaration",
	}

	if nameNode := treesitter.ChildByField(n, "name"); nameNode != nil {
		cls.Name = f.NodeText(nameNode)
	}

	// Heritage: class_heritage contains extends_clause and implements_clause
	if heritage := treesitter.FindChildByType(n, "class_heritage"); heritage != nil {
		for i := 0; i < int(heritage.NamedChildCount()); i++ {
			clause := heritage.NamedChild(i)
			if clause == nil {
				continue
			}
			switch clause.Type() {
			case "extends_clause":
				if t := firstTypeName(clause, f); t != "" {
					cls.Extends = t
				}
			case "implements_clause":
				for j := 0; j < int(clause.NamedChildCount()); j++ {
					c := clause.NamedChild(j)
					if c != nil {
						cls.Implements = append(cls.Implements, f.NodeText(c))
					}
				}
			}
		}
	}

	// Methods and properties live in class_body
	if body := treesitter.ChildByField(n, "body"); body != nil {
		for i := 0; i < int(body.NamedChildCount()); i++ {
			member := body.NamedChild(i)
			if member == nil {
				continue
			}
			switch member.Type() {
			case "method_definition":
				fn := extractMethod(member, f, cls.Name)
				cls.Methods = append(cls.Methods, fn)
			case "public_field_definition", "property_definition":
				cls.Properties = append(cls.Properties, extractProperty(member, f))
			}
		}
	}

	return cls
}

func extractInterface(n *sitter.Node, f *treesitter.ParsedFile) Interface {
	iface := Interface{
		LineStart: treesitter.NodeLine(n),
		LineEnd:   treesitter.NodeEndLine(n),
	}
	if nameNode := treesitter.ChildByField(n, "name"); nameNode != nil {
		iface.Name = f.NodeText(nameNode)
	}
	// extends_clause for interfaces
	for i := 0; i < int(n.NamedChildCount()); i++ {
		c := n.NamedChild(i)
		if c != nil && c.Type() == "extends_clause" {
			for j := 0; j < int(c.NamedChildCount()); j++ {
				tc := c.NamedChild(j)
				if tc != nil {
					iface.Extends = append(iface.Extends, f.NodeText(tc))
				}
			}
		}
	}
	// Properties: property_signature nodes inside interface_body
	if body := treesitter.ChildByField(n, "body"); body != nil {
		for i := 0; i < int(body.NamedChildCount()); i++ {
			p := body.NamedChild(i)
			if p != nil && p.Type() == "property_signature" {
				iface.Properties = append(iface.Properties, extractProperty(p, f))
			}
		}
	}
	return iface
}

func extractFunction(n *sitter.Node, f *treesitter.ParsedFile, receiver string) Function {
	fn := Function{
		LineStart: treesitter.NodeLine(n),
		LineEnd:   treesitter.NodeEndLine(n),
		Receiver:  receiver,
	}
	if nameNode := treesitter.ChildByField(n, "name"); nameNode != nil {
		fn.Name = f.NodeText(nameNode)
	}
	if params := treesitter.ChildByField(n, "parameters"); params != nil {
		fn.Params = extractParamNames(params, f)
	}
	if rt := treesitter.ChildByField(n, "return_type"); rt != nil {
		fn.ReturnType = strings.TrimSpace(f.NodeText(rt))
	}
	fn.IsAsync = hasChildOfType(n, "async")
	return fn
}

func extractMethod(n *sitter.Node, f *treesitter.ParsedFile, receiver string) Function {
	fn := Function{
		LineStart: treesitter.NodeLine(n),
		LineEnd:   treesitter.NodeEndLine(n),
		Receiver:  receiver,
	}
	if nameNode := treesitter.ChildByField(n, "name"); nameNode != nil {
		fn.Name = f.NodeText(nameNode)
	}
	if params := treesitter.ChildByField(n, "parameters"); params != nil {
		fn.Params = extractParamNames(params, f)
	}
	if rt := treesitter.ChildByField(n, "return_type"); rt != nil {
		fn.ReturnType = strings.TrimSpace(f.NodeText(rt))
	}
	fn.IsAsync = hasChildOfType(n, "async")
	return fn
}

func extractProperty(n *sitter.Node, f *treesitter.ParsedFile) Property {
	p := Property{}
	if nameNode := treesitter.ChildByField(n, "name"); nameNode != nil {
		p.Name = f.NodeText(nameNode)
	}
	if ta := treesitter.ChildByField(n, "type"); ta != nil {
		p.Type = strings.TrimSpace(f.NodeText(ta))
	}
	p.IsOptional = hasChildOfType(n, "?")
	p.IsReadonly = hasChildOfType(n, "readonly")
	return p
}

func extractImport(n *sitter.Node, f *treesitter.ParsedFile) Import {
	imp := Import{Line: treesitter.NodeLine(n)}
	if src := treesitter.ChildByField(n, "source"); src != nil {
		imp.Source = strings.Trim(f.NodeText(src), `"'`)
	}
	// import_clause contains: default identifier, namespace_import, named_imports
	if clause := treesitter.FindChildByType(n, "import_clause"); clause != nil {
		for i := 0; i < int(clause.NamedChildCount()); i++ {
			c := clause.NamedChild(i)
			if c == nil {
				continue
			}
			switch c.Type() {
			case "identifier":
				imp.Default = f.NodeText(c)
			case "namespace_import":
				imp.Namespace = f.NodeText(c)
			case "named_imports":
				for j := 0; j < int(c.NamedChildCount()); j++ {
					spec := c.NamedChild(j)
					if spec == nil {
						continue
					}
					if nameNode := treesitter.ChildByField(spec, "name"); nameNode != nil {
						imp.Names = append(imp.Names, f.NodeText(nameNode))
					}
				}
			}
		}
	}
	return imp
}

func extractExportedArrow(n *sitter.Node, f *treesitter.ParsedFile, m *Module) {
	// export const foo = () => {...} — extract as function
	for i := 0; i < int(n.NamedChildCount()); i++ {
		decl := n.NamedChild(i)
		if decl == nil || decl.Type() != "variable_declarator" {
			continue
		}
		nameNode := treesitter.ChildByField(decl, "name")
		valueNode := treesitter.ChildByField(decl, "value")
		if nameNode == nil || valueNode == nil {
			continue
		}
		if valueNode.Type() == "arrow_function" || valueNode.Type() == "function" {
			fn := Function{
				Name:       f.NodeText(nameNode),
				IsArrow:    valueNode.Type() == "arrow_function",
				IsExported: true,
				LineStart:  treesitter.NodeLine(decl),
				LineEnd:    treesitter.NodeEndLine(decl),
			}
			if params := treesitter.ChildByField(valueNode, "parameters"); params != nil {
				fn.Params = extractParamNames(params, f)
			}
			if rt := treesitter.ChildByField(valueNode, "return_type"); rt != nil {
				fn.ReturnType = strings.TrimSpace(f.NodeText(rt))
			}
			m.Functions = append(m.Functions, fn)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func extractParamNames(params *sitter.Node, f *treesitter.ParsedFile) []string {
	var names []string
	for i := 0; i < int(params.NamedChildCount()); i++ {
		p := params.NamedChild(i)
		if p == nil {
			continue
		}
		if nameNode := treesitter.ChildByField(p, "pattern"); nameNode != nil {
			names = append(names, f.NodeText(nameNode))
		} else if p.Type() == "required_parameter" || p.Type() == "optional_parameter" {
			if nameNode := treesitter.ChildByField(p, "pattern"); nameNode != nil {
				names = append(names, f.NodeText(nameNode))
			}
		} else {
			names = append(names, f.NodeText(p))
		}
	}
	return names
}

func firstTypeName(node *sitter.Node, f *treesitter.ParsedFile) string {
	for i := 0; i < int(node.NamedChildCount()); i++ {
		c := node.NamedChild(i)
		if c == nil {
			continue
		}
		switch c.Type() {
		case "type_identifier", "identifier", "generic_type", "nested_type_identifier":
			return f.NodeText(c)
		}
	}
	return ""
}

func hasChildOfType(n *sitter.Node, typ string) bool {
	for i := 0; i < int(n.ChildCount()); i++ {
		c := n.Child(i)
		if c != nil && (c.Type() == typ || c.Type() == "?") {
			if typ == "?" && c.Type() == "?" {
				return true
			}
			if typ != "?" && c.Type() == typ {
				return true
			}
		}
	}
	return false
}

func isExportedNode(n *sitter.Node) bool {
	parent := n.Parent()
	return parent != nil && parent.Type() == "export_statement"
}

func markExported(classes *[]Class, node *sitter.Node, f *treesitter.ParsedFile) {
	if name := treesitter.ChildByField(node, "name"); name != nil {
		n := f.NodeText(name)
		for i := range *classes {
			if (*classes)[i].Name == n {
				(*classes)[i].IsExported = true
				return
			}
		}
	}
}

func markExportedInterface(ifaces *[]Interface, node *sitter.Node, f *treesitter.ParsedFile) {
	if name := treesitter.ChildByField(node, "name"); name != nil {
		n := f.NodeText(name)
		for i := range *ifaces {
			if (*ifaces)[i].Name == n {
				(*ifaces)[i].IsExported = true
				return
			}
		}
	}
}

func markExportedFunction(fns *[]Function, node *sitter.Node, f *treesitter.ParsedFile) {
	if name := treesitter.ChildByField(node, "name"); name != nil {
		n := f.NodeText(name)
		for i := range *fns {
			if (*fns)[i].Name == n {
				(*fns)[i].IsExported = true
				return
			}
		}
	}
}
