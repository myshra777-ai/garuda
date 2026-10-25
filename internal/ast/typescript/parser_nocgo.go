//go:build !cgo

// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

// Package typescript provides a 100% pure-Go TypeScript AST & call-graph extractor
// when compiled with CGO_ENABLED=0 (no GCC/Clang/MinGW/Xcode required).
package typescript

import (
	"context"
	"regexp"
	"strings"
)

// Parser is a zero-CGO structural TypeScript parser.
type Parser struct{}

// NewParser constructs a pure-Go TypeScript parser.
func NewParser() *Parser {
	return &Parser{}
}

// Close is a no-op for the pure-Go parser.
func (p *Parser) Close() {}

var (
	reTSImport       = regexp.MustCompile(`(?m)^\s*import\s+(.+?)\s+from\s+['"]([^'"]+)['"]`)
	reTSClass        = regexp.MustCompile(`^\s*(export\s+)?(abstract\s+)?class\s+([A-Za-z_$][A-Za-z0-9_$]*)(?:\s+extends\s+([A-Za-z0-9_$.<>]+))?(?:\s+implements\s+([^{]+))?\s*\{`)
	reTSInterface    = regexp.MustCompile(`^\s*(export\s+)?interface\s+([A-Za-z_$][A-Za-z0-9_$]*)(?:\s+extends\s+([^{]+))?\s*\{`)
	reTSFunc         = regexp.MustCompile(`^\s*(export\s+)?(async\s+)?function\s+([A-Za-z_$][A-Za-z0-9_$]*)\s*(?:<[^>]*>)?\s*\(([^)]*)\)\s*(?::\s*([^{]+))?\s*\{`)
	reTSArrow        = regexp.MustCompile(`^\s*(export\s+)?(?:const|let|var)\s+([A-Za-z_$][A-Za-z0-9_$]*)\s*(?::[^=]+)?=\s*(async\s+)?\(([^)]*)\)\s*(?::\s*([^=]+))?=>`)
	reTSMethod       = regexp.MustCompile(`^\s*(?:(?:public|private|protected|readonly|static|override|abstract)\s+)*(async\s+)?([A-Za-z_$][A-Za-z0-9_$]*)\s*(?:<[^>]*>)?\s*\(([^)]*)\)\s*(?::\s*([^{;]+))?\s*\{`)
	reTSProperty     = regexp.MustCompile(`^\s*(?:(?:public|private|protected)\s+)?(readonly\s+)?([A-Za-z_$][A-Za-z0-9_$]*)(\?)?\s*:\s*([^;=]+)\s*[;=]?`)
	reTSNewAssign    = regexp.MustCompile(`(?:(?:const|let|var)\s+)?((?:this\.)?[A-Za-z_$][A-Za-z0-9_$]*)\s*(?::[^=]+)?=\s*new\s+([A-Za-z_$][A-Za-z0-9_$.]*)\s*\(`)
	reTSNewExpr      = regexp.MustCompile(`\bnew\s+([A-Za-z_$][A-Za-z0-9_$.]*)\s*\(`)
	reTSCallExpr     = regexp.MustCompile(`\b((?:this\.)?[A-Za-z_$][A-Za-z0-9_$]*(?:\.[A-Za-z_$][A-Za-z0-9_$]*)*)\s*\(`)
)

// Parse parses TypeScript source bytes into a structured Module without requiring CGO.
func (p *Parser) Parse(_ context.Context, src []byte, path, moduleName string) (*Module, error) {
	text := string(src)
	lines := strings.Split(text, "\n")
	m := &Module{Path: path, Module: moduleName}

	// 1. Extract Imports
	for idx, rawLine := range lines {
		lineNum := idx + 1
		trimmed := strings.TrimSpace(rawLine)
		if !strings.HasPrefix(trimmed, "import ") {
			continue
		}
		if im := reTSImport.FindStringSubmatch(trimmed); im != nil {
			imp := parseTSImportClause(im[1], im[2], lineNum)
			m.Imports = append(m.Imports, imp)
		}
	}

	// 2. Scan lines for Classes, Interfaces, Functions, and Exported Arrows
	for i := 0; i < len(lines); i++ {
		line := stripTSLineComment(lines[i])
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Class declaration
		if cm := reTSClass.FindStringSubmatch(line); cm != nil {
			endIdx := findMatchingBraceLine(lines, i)
			cls := Class{
				Name:       strings.TrimSpace(cm[3]),
				Extends:    strings.TrimSpace(cm[4]),
				IsExported: strings.TrimSpace(cm[1]) != "",
				IsAbstract: strings.TrimSpace(cm[2]) != "",
				LineStart:  i + 1,
				LineEnd:    endIdx + 1,
			}
			if strings.TrimSpace(cm[5]) != "" {
				for _, impl := range strings.Split(cm[5], ",") {
					if v := strings.TrimSpace(impl); v != "" {
						cls.Implements = append(cls.Implements, v)
					}
				}
			}
			parseTSClassBody(lines, i+1, endIdx, &cls)
			m.Classes = append(m.Classes, cls)
			i = endIdx
			continue
		}

		// Interface declaration
		if im := reTSInterface.FindStringSubmatch(line); im != nil {
			endIdx := findMatchingBraceLine(lines, i)
			iface := Interface{
				Name:       strings.TrimSpace(im[2]),
				IsExported: strings.TrimSpace(im[1]) != "",
				LineStart:  i + 1,
				LineEnd:    endIdx + 1,
			}
			if strings.TrimSpace(im[3]) != "" {
				for _, ext := range strings.Split(im[3], ",") {
					if v := strings.TrimSpace(ext); v != "" {
						iface.Extends = append(iface.Extends, v)
					}
				}
			}
			for j := i + 1; j < endIdx && j < len(lines); j++ {
				pLine := strings.TrimSpace(stripTSLineComment(lines[j]))
				if pLine == "" || pLine == "}" {
					continue
				}
				if pm := reTSProperty.FindStringSubmatch(pLine); pm != nil && !strings.Contains(pLine, "(") {
					iface.Properties = append(iface.Properties, Property{
						Name:       strings.TrimSpace(pm[2]),
						Type:       strings.TrimSpace(pm[4]),
						IsOptional: strings.TrimSpace(pm[3]) == "?",
						IsReadonly: strings.TrimSpace(pm[1]) != "",
					})
				}
			}
			m.Interfaces = append(m.Interfaces, iface)
			i = endIdx
			continue
		}

		// Top-level Function declaration
		if fm := reTSFunc.FindStringSubmatch(line); fm != nil {
			endIdx := findMatchingBraceLine(lines, i)
			localTypes := make(map[string]string)
			params := parseTSParams(fm[4], localTypes)
			fn := Function{
				Name:       strings.TrimSpace(fm[3]),
				Params:     params,
				ReturnType: strings.TrimSpace(fm[5]),
				IsAsync:    strings.TrimSpace(fm[2]) != "",
				IsExported: strings.TrimSpace(fm[1]) != "",
				LineStart:  i + 1,
				LineEnd:    endIdx + 1,
				LocalTypes: localTypes,
			}
			fn.Calls = extractTSBodyCalls(lines[i:endIdx+1], fn.Name, fn.LocalTypes)
			m.Functions = append(m.Functions, fn)
			i = endIdx
			continue
		}

		// Exported Arrow Function
		if am := reTSArrow.FindStringSubmatch(line); am != nil && strings.TrimSpace(am[1]) != "" {
			endIdx := findMatchingBraceLine(lines, i)
			localTypes := make(map[string]string)
			params := parseTSParams(am[4], localTypes)
			fn := Function{
				Name:       strings.TrimSpace(am[2]),
				Params:     params,
				ReturnType: strings.TrimSpace(am[5]),
				IsAsync:    strings.TrimSpace(am[3]) != "",
				IsArrow:    true,
				IsExported: true,
				LineStart:  i + 1,
				LineEnd:    endIdx + 1,
				LocalTypes: localTypes,
			}
			fn.Calls = extractTSBodyCalls(lines[i:endIdx+1], fn.Name, fn.LocalTypes)
			m.Functions = append(m.Functions, fn)
			i = endIdx
			continue
		}
	}

	return m, nil
}

func parseTSClassBody(lines []string, startIdx, endIdx int, cls *Class) {
	for i := startIdx; i < endIdx && i < len(lines); i++ {
		line := stripTSLineComment(lines[i])
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || trimmed == "}" {
			continue
		}
		if mm := reTSMethod.FindStringSubmatch(line); mm != nil {
			mName := strings.TrimSpace(mm[2])
			if isTSControlKeyword(mName) {
				continue
			}
			mEnd := findMatchingBraceLine(lines, i)
			localTypes := make(map[string]string)
			params := parseTSParams(mm[3], localTypes)
			fn := Function{
				Name:       mName,
				Params:     params,
				ReturnType: strings.TrimSpace(mm[4]),
				IsAsync:    strings.TrimSpace(mm[1]) != "",
				Receiver:   cls.Name,
				LineStart:  i + 1,
				LineEnd:    mEnd + 1,
				LocalTypes: localTypes,
			}
			fn.Calls = extractTSBodyCalls(lines[i:mEnd+1], mName, fn.LocalTypes)
			cls.Methods = append(cls.Methods, fn)
			i = mEnd
			continue
		}
		if pm := reTSProperty.FindStringSubmatch(line); pm != nil && !strings.Contains(line, "(") {
			cls.Properties = append(cls.Properties, Property{
				Name:       strings.TrimSpace(pm[2]),
				Type:       strings.TrimSpace(pm[4]),
				IsOptional: strings.TrimSpace(pm[3]) == "?",
				IsReadonly: strings.TrimSpace(pm[1]) != "",
			})
		}
	}
}

func parseTSParams(rawParams string, localTypes map[string]string) []string {
	var names []string
	if strings.TrimSpace(rawParams) == "" {
		return names
	}
	for _, part := range strings.Split(rawParams, ",") {
		p := strings.TrimSpace(part)
		if p == "" {
			continue
		}
		for _, mod := range []string{"public ", "private ", "protected ", "readonly "} {
			p = strings.TrimSpace(strings.TrimPrefix(p, mod))
		}
		if eq := strings.Index(p, "="); eq >= 0 {
			p = strings.TrimSpace(p[:eq])
		}
		if colon := strings.Index(p, ":"); colon >= 0 {
			pName := strings.TrimSuffix(strings.TrimSpace(p[:colon]), "?")
			tName := strings.TrimSpace(p[colon+1:])
			if pName != "" {
				names = append(names, pName)
				if tName != "" && localTypes != nil {
					localTypes[pName] = tName
					localTypes["this."+pName] = tName
				}
			}
		} else {
			pName := strings.TrimSuffix(strings.TrimSpace(p), "?")
			if pName != "" {
				names = append(names, pName)
			}
		}
	}
	return names
}

func extractTSBodyCalls(bodyLines []string, selfFuncName string, localTypes map[string]string) []string {
	var calls []string
	for idx, raw := range bodyLines {
		line := stripTSLineComment(raw)
		for _, nam := range reTSNewAssign.FindAllStringSubmatch(line, -1) {
			lhs := strings.TrimSpace(nam[1])
			ctor := strings.TrimSpace(nam[2])
			if lhs != "" && ctor != "" && localTypes != nil {
				localTypes[lhs] = ctor
			}
		}
		for _, nm := range reTSNewExpr.FindAllStringSubmatch(line, -1) {
			ctor := strings.TrimSpace(nm[1])
			if ctor != "" {
				calls = append(calls, ctor)
			}
		}
		for _, cm := range reTSCallExpr.FindAllStringSubmatch(line, -1) {
			callee := strings.TrimSpace(cm[1])
			if callee == "" || isTSControlKeyword(callee) {
				continue
			}
			// Skip declaration header on first line
			if idx == 0 && callee == selfFuncName {
				continue
			}
			// Avoid duplicating `new Foo()` as both new_expression and call_expression
			if strings.Contains(line, "new "+callee+"(") {
				continue
			}
			calls = append(calls, callee)
		}
	}
	return calls
}

func parseTSImportClause(clause, source string, line int) Import {
	imp := Import{
		Source:      strings.TrimSpace(source),
		NameAliases: make(map[string]string),
		Line:        line,
	}
	c := strings.TrimSpace(clause)
	if strings.HasPrefix(c, "* as ") {
		imp.Namespace = strings.TrimSpace(strings.TrimPrefix(c, "* as "))
		return imp
	}
	if bStart := strings.Index(c, "{"); bStart >= 0 {
		prefix := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(c[:bStart]), ","))
		if prefix != "" {
			imp.Default = prefix
		}
		bEnd := strings.LastIndex(c, "}")
		if bEnd > bStart {
			inner := c[bStart+1 : bEnd]
			for _, spec := range strings.Split(inner, ",") {
				s := strings.TrimSpace(spec)
				s = strings.TrimSpace(strings.TrimPrefix(s, "type "))
				if s == "" {
					continue
				}
				if asIdx := strings.Index(s, " as "); asIdx >= 0 {
					orig := strings.TrimSpace(s[:asIdx])
					al := strings.TrimSpace(s[asIdx+4:])
					if orig != "" {
						imp.Names = append(imp.Names, orig)
						if al != "" {
							imp.NameAliases[al] = orig
						}
					}
				} else {
					imp.Names = append(imp.Names, s)
					imp.NameAliases[s] = s
				}
			}
		}
		return imp
	}
	imp.Default = strings.TrimSpace(strings.TrimPrefix(c, "type "))
	return imp
}

func findMatchingBraceLine(lines []string, startIdx int) int {
	depth := 0
	started := false
	for i := startIdx; i < len(lines); i++ {
		line := stripTSLineComment(lines[i])
		for j := 0; j < len(line); j++ {
			switch line[j] {
			case '{':
				depth++
				started = true
			case '}':
				depth--
				if started && depth <= 0 {
					return i
				}
			}
		}
	}
	return startIdx
}

func stripTSLineComment(s string) string {
	inSingle, inDouble, inBacktick := false, false, false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '\\' && i+1 < len(s) {
			i++
			continue
		}
		switch {
		case c == '\'' && !inDouble && !inBacktick:
			inSingle = !inSingle
		case c == '"' && !inSingle && !inBacktick:
			inDouble = !inDouble
		case c == '`' && !inSingle && !inDouble:
			inBacktick = !inBacktick
		case c == '/' && i+1 < len(s) && s[i+1] == '/' && !inSingle && !inDouble && !inBacktick:
			return s[:i]
		}
	}
	return s
}

func isTSControlKeyword(w string) bool {
	switch w {
	case "if", "for", "while", "switch", "catch", "function", "return", "typeof", "instanceof", "super":
		return true
	default:
		return false
	}
}
