// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package typescript

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
	Calls      []string
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
	Namespace   string            // * as X (holds X)
	Line        int
}
