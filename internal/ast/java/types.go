// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package java

import "strings"

// Module is the parsed representation of a single Java source file (.java).
type Module struct {
	Path        string
	PackageName string // from `package com.acme.payments;` or derived from src/main/java
	Module      string // file-scoped module identifier (e.g. com.acme.payments.RefundService)
	Classes     []Class
	Interfaces  []Interface
	Imports     []Import
}

type Class struct {
	Name        string
	Kind        string // class | record | enum
	Extends     string
	Implements  []string
	Annotations []string
	LineStart   int
	LineEnd     int
	IsPublic    bool
	IsAbstract  bool
	Methods     []Method
	Fields      []Field
}

type Interface struct {
	Name        string
	Extends     []string
	Annotations []string
	LineStart   int
	LineEnd     int
	IsPublic    bool
	Methods     []Method
}

type CallSite struct {
	Target string
	Line   int
}

type Method struct {
	Name          string
	Params        []string
	ReturnType    string
	Annotations   []string
	IsPublic      bool
	IsStatic      bool
	IsConstructor bool
	Receiver      string // enclosing class/interface name
	LineStart     int
	LineEnd       int
	Calls         []string
	CallSites     []CallSite
	LocalTypes    map[string]string // variable/param name -> unwrapped Java type name
}

type Field struct {
	Name        string
	Type        string
	Annotations []string
	IsFinal     bool
	IsStatic    bool
}

type Import struct {
	Path       string // e.g. "com.acme.db.OrderRepository" or "com.acme.db.*"
	SimpleName string // e.g. "OrderRepository" or "*" or static member name
	ClassName  string // for static imports: enclosing class FQN
	IsStatic   bool
	IsWildcard bool
	Line       int
}

// UnwrapJavaTypeAnnotation strips Java generic wrappers such as Optional<T>,
// CompletableFuture<T>, ResponseEntity<T>, Mono<T>, Flux<T>, List<T>, Set<T>,
// array suffixes T[], and varargs T... to yield the core reference type T.
func UnwrapJavaTypeAnnotation(raw string) string {
	t := strings.TrimSpace(raw)
	if t == "" {
		return ""
	}
	for _, mod := range []string{"final ", "@NonNull ", "@Nonnull ", "@Nullable ", "@Valid "} {
		t = strings.TrimSpace(strings.TrimPrefix(t, mod))
	}
	t = strings.TrimSuffix(t, "...")
	t = strings.TrimSuffix(t, "[]")
	t = strings.TrimSpace(t)

	primitives := map[string]bool{
		"void": true, "boolean": true, "byte": true, "short": true,
		"int": true, "long": true, "float": true, "double": true, "char": true,
		"var": true, "Object": true, "?": true,
	}
	if primitives[t] {
		return ""
	}

	wrappers := []string{
		"Optional<", "CompletableFuture<", "Future<", "ResponseEntity<",
		"Mono<", "Flux<", "List<", "ArrayList<", "Set<", "HashSet<",
		"Collection<", "Iterable<", "Stream<", "Supplier<", "Provider<", "ObjectProvider<",
	}
	for _, w := range wrappers {
		if strings.HasPrefix(t, w) && strings.HasSuffix(t, ">") {
			inner := strings.TrimSpace(t[len(w) : len(t)-1])
			if strings.HasPrefix(inner, "? extends ") {
				inner = strings.TrimSpace(strings.TrimPrefix(inner, "? extends "))
			} else if strings.HasPrefix(inner, "? super ") {
				inner = strings.TrimSpace(strings.TrimPrefix(inner, "? super "))
			}
			return UnwrapJavaTypeAnnotation(inner)
		}
	}

	if lt := strings.Index(t, "<"); lt > 0 {
		t = t[:lt]
	}
	if dot := strings.LastIndex(t, "."); dot >= 0 {
		t = t[dot+1:]
	}
	t = strings.TrimSpace(t)
	if primitives[t] {
		return ""
	}
	return t
}
