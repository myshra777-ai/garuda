// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package typescript

import "strings"

// Module is the parsed representation of a single TypeScript file.
type Module struct {
	Path       string
	Module     string // derived from file path
	Classes    []Class
	Interfaces []Interface
	Functions  []Function
	Imports    []Import
	ReExports  []Import // export { X as Y } from './mod' or export * from './mod'
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

type CallSite struct {
	Target string
	Line   int
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
	Calls      []string
	CallSites  []CallSite
	LocalTypes map[string]string // variable/param name → class/type name
}

type Property struct {
	Name       string
	Type       string
	IsOptional bool
	IsReadonly bool
}

type Import struct {
	Source      string            // module path
	Names       []string          // named imports (exported names)
	NameAliases map[string]string // localAlias → exportedName
	Default     string            // default import name
	Namespace   string            // * as X (holds X) or "*" for `export * from`
	Line        int
}

// UnwrapTSTypeAnnotation strips Promise<T>, Observable<T>, Readonly<T>, Array<T>,
// T[], and union suffixes (`| null | undefined`) to yield the core type name T.
func UnwrapTSTypeAnnotation(raw string) string {
	t := strings.TrimSpace(raw)
	t = strings.TrimPrefix(t, ":")
	t = strings.TrimSpace(t)
	if t == "" || t == "void" || t == "any" || t == "unknown" || t == "never" {
		return ""
	}
	if pipe := strings.Index(t, "|"); pipe >= 0 {
		for _, part := range strings.Split(t, "|") {
			p := strings.TrimSpace(part)
			if p != "" && p != "null" && p != "undefined" && p != "void" {
				t = p
				break
			}
		}
	}
	wrappers := []string{"Promise<", "Observable<", "Readonly<", "Array<", "Partial<", "Required<"}
	for _, w := range wrappers {
		if strings.HasPrefix(t, w) && strings.HasSuffix(t, ">") {
			inner := t[len(w) : len(t)-1]
			return UnwrapTSTypeAnnotation(inner)
		}
	}
	t = strings.TrimSuffix(t, "[]")
	if lt := strings.Index(t, "<"); lt > 0 {
		t = t[:lt]
	}
	return strings.TrimSpace(t)
}
