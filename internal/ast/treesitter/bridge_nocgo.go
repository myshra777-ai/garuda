//go:build !cgo

// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

// Package treesitter provides a pure-Go no-CGO compilation unit when built with
// CGO_ENABLED=0 (for example, on Windows, macOS, or Linux systems without GCC/Clang).
// In !cgo mode, internal/ast/typescript uses its pure-Go structural AST parser.
package treesitter

// Language identifies a grammar name.
type Language string

const (
	LanguageTypeScript Language = "typescript"
	LanguageTSX        Language = "tsx"
)
