// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

// Package treesitter provides a thin shared wrapper around the smacker
// tree-sitter Go bindings. Language-specific extractors live in sibling
// packages (typescript, tsx) and consume the AST this package produces.
//
// Build requirement: CGO_ENABLED=1 and a working C toolchain (gcc/clang).
package treesitter

import (
	"context"
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
)

// Language identifies a tree-sitter grammar.
type Language string

const (
	LanguageTypeScript Language = "typescript"
	LanguageTSX        Language = "tsx"
)

// ParsedFile is the result of parsing a single source file.
type ParsedFile struct {
	Source   []byte
	Tree     *sitter.Tree
	Language Language
}

// Parser wraps a tree-sitter parser for a specific grammar.
type Parser struct {
	inner    *sitter.Parser
	language Language
}

// Wrap constructs a Parser around a grammar supplied by a language package.
// Language-specific packages use this to build their parser.
func Wrap(lang Language, grammar *sitter.Language) *Parser {
	p := sitter.NewParser()
	p.SetLanguage(grammar)
	return &Parser{inner: p, language: lang}
}

// Parse parses source bytes and returns a ParsedFile.
// The tree must be closed by the caller when done.
func (p *Parser) Parse(ctx context.Context, source []byte) (*ParsedFile, error) {
	tree, err := p.inner.ParseCtx(ctx, nil, source)
	if err != nil {
		return nil, fmt.Errorf("tree-sitter parse failed: %w", err)
	}
	return &ParsedFile{
		Source:   source,
		Tree:     tree,
		Language: p.language,
	}, nil
}

// Close releases the underlying parser resources.
func (p *Parser) Close() {
	if p.inner != nil {
		p.inner.Close()
	}
}

// Walk visits every named node in the tree in depth-first pre-order.
// The visitor returns true to descend into the node's children,
// false to skip the subtree.
func (f *ParsedFile) Walk(visit func(node *sitter.Node) bool) {
	if f == nil || f.Tree == nil {
		return
	}
	walkNode(f.Tree.RootNode(), visit)
}

func walkNode(node *sitter.Node, visit func(*sitter.Node) bool) {
	if node == nil {
		return
	}
	if !visit(node) {
		return
	}
	count := int(node.NamedChildCount())
	for i := 0; i < count; i++ {
		child := node.NamedChild(i)
		if child != nil {
			walkNode(child, visit)
		}
	}
}

// NodeText returns the source text for a node.
func (f *ParsedFile) NodeText(node *sitter.Node) string {
	if node == nil || f.Source == nil {
		return ""
	}
	return node.Content(f.Source)
}

// NodeLine returns the 1-based starting line of a node.
func NodeLine(node *sitter.Node) int {
	if node == nil {
		return 0
	}
	return int(node.StartPoint().Row) + 1
}

// NodeEndLine returns the 1-based ending line of a node.
func NodeEndLine(node *sitter.Node) int {
	if node == nil {
		return 0
	}
	return int(node.EndPoint().Row) + 1
}

// ChildByField returns the first named child with the given field name.
func ChildByField(node *sitter.Node, field string) *sitter.Node {
	if node == nil {
		return nil
	}
	return node.ChildByFieldName(field)
}

// FindChildByType returns the first named child whose type matches.
func FindChildByType(node *sitter.Node, typ string) *sitter.Node {
	if node == nil {
		return nil
	}
	for i := 0; i < int(node.NamedChildCount()); i++ {
		c := node.NamedChild(i)
		if c != nil && c.Type() == typ {
			return c
		}
	}
	return nil
}

// FindChildrenByType returns all named children whose type matches.
func FindChildrenByType(node *sitter.Node, typ string) []*sitter.Node {
	var out []*sitter.Node
	if node == nil {
		return out
	}
	for i := 0; i < int(node.NamedChildCount()); i++ {
		c := node.NamedChild(i)
		if c != nil && c.Type() == typ {
			out = append(out, c)
		}
	}
	return out
}
