// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

// Package java provides a zero-CGO, zero-JVM Java AST, Spring/Jakarta DI,
// inheritance, interface implementation, and type-directed call-graph analyzer.
package java

import (
	"context"
	"regexp"
	"strings"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Close() {}

var (
	reJavaPackage     = regexp.MustCompile(`^\s*package\s+([A-Za-z_][A-Za-z0-9_.]*)\s*;`)
	reJavaImport      = regexp.MustCompile(`^\s*import\s+(static\s+)?([A-Za-z_][A-Za-z0-9_.*]*)\s*;`)
	reJavaAnnotation  = regexp.MustCompile(`^\s*@([A-Za-z_][A-Za-z0-9_.]*)`)
	reJavaClass       = regexp.MustCompile(`^\s*(?:(?:public|protected|private|abstract|final|static|sealed|non-sealed)\s+)*(class|record|enum)\s+([A-Za-z_$][A-Za-z0-9_$]*)(?:<[^>]*>)?(?:\s*\(([^)]*)\))?(?:\s+extends\s+([A-Za-z0-9_$.<>]+))?(?:\s+implements\s+([^{]+))?\s*\{`)
	reJavaInterface   = regexp.MustCompile(`^\s*(?:(?:public|protected|private|abstract|sealed|non-sealed)\s+)*interface\s+([A-Za-z_$][A-Za-z0-9_$]*)(?:<[^>]*>)?(?:\s+extends\s+([^{]+))?\s*\{`)
	reJavaCtor        = regexp.MustCompile(`^\s*(?:(?:public|protected|private)\s+)?([A-Za-z_$][A-Za-z0-9_$]*)\s*\(([^)]*)\)\s*(?:throws\s+[^{]+)?\{`)
	reJavaMethod      = regexp.MustCompile(`^\s*(?:(?:public|protected|private|static|final|synchronized|abstract|default|native)\s+)*(?:<[^>]+>\s+)?([A-Za-z_$][A-Za-z0-9_$.<>\[\],\s?]*?)\s+([A-Za-z_$][A-Za-z0-9_$]*)\s*\(([^)]*)\)\s*(?:throws\s+[^{;]+)?(\{|;)`)
	reJavaField       = regexp.MustCompile(`^\s*(?:(?:public|protected|private|static|final|transient|volatile)\s+)+([A-Za-z_$][A-Za-z0-9_$.<>\[\]]*)\s+([A-Za-z_$][A-Za-z0-9_$]*)\s*(?:=\s*[^;]+)?;`)
	reJavaNewAssign   = regexp.MustCompile(`(?:(?:final\s+)?([A-Za-z_$][A-Za-z0-9_$.<>]*)\s+)?((?:this\.)?[A-Za-z_$][A-Za-z0-9_$]*)\s*=\s*new\s+([A-Za-z_$][A-Za-z0-9_$.]*)\s*(?:<[^>]*>)?\s*\(`)
	reJavaCallAssign  = regexp.MustCompile(`(?:(?:final\s+)?([A-Za-z_$][A-Za-z0-9_$.<>]*)\s+)?((?:this\.)?[A-Za-z_$][A-Za-z0-9_$]*)\s*=\s*((?:this\.|super\.)?[A-Za-z_$][A-Za-z0-9_$.]*)\s*\(`)
	reJavaTypedLocal  = regexp.MustCompile(`^\s*(?:final\s+)?([A-Z][A-Za-z0-9_$.<>]*)\s+([a-z_][A-Za-z0-9_$]*)\s*[;=]`)
	reJavaFieldAssign = regexp.MustCompile(`^\s*this\.([A-Za-z_$][A-Za-z0-9_$]*)\s*=\s*([A-Za-z_$][A-Za-z0-9_$]*)\s*;`)
	reJavaChainCall   = regexp.MustCompile(`((?:this\.|super\.)?[A-Za-z_$][A-Za-z0-9_$.]*)\s*\([^)]*\)\s*\.\s*([A-Za-z_$][A-Za-z0-9_$]*)\s*\(`)
	reJavaNewExpr     = regexp.MustCompile(`\bnew\s+([A-Za-z_$][A-Za-z0-9_$.]*)\s*(?:<[^>]*>)?\s*\(`)
	reJavaCallExpr    = regexp.MustCompile(`\b((?:this\.|super\.)?[A-Za-z_$][A-Za-z0-9_$]*(?:\.[A-Za-z_$][A-Za-z0-9_$]*)*)\s*\(`)
)

// Parse parses Java source bytes into a structured Module.
func (p *Parser) Parse(_ context.Context, src []byte, path, fallbackPkg string) (*Module, error) {
	lines := strings.Split(string(src), "\n")
	m := &Module{
		Path:        path,
		PackageName: fallbackPkg,
	}

	// 1. Extract package and import statements
	for idx, rawLine := range lines {
		lineNum := idx + 1
		trimmed := strings.TrimSpace(stripJavaComment(rawLine))
		if trimmed == "" {
			continue
		}
		if pm := reJavaPackage.FindStringSubmatch(trimmed); pm != nil {
			m.PackageName = strings.TrimSpace(pm[1])
			continue
		}
		if im := reJavaImport.FindStringSubmatch(trimmed); im != nil {
			isStatic := strings.TrimSpace(im[1]) != ""
			impPath := strings.TrimSpace(im[2])
			isWild := strings.HasSuffix(impPath, ".*")
			simple := impPath
			className := ""
			if isWild {
				simple = "*"
				if isStatic {
					className = strings.TrimSuffix(impPath, ".*")
				}
			} else if dot := strings.LastIndex(impPath, "."); dot >= 0 {
				simple = impPath[dot+1:]
				if isStatic {
					className = impPath[:dot]
				}
			}
			m.Imports = append(m.Imports, Import{
				Path:       impPath,
				SimpleName: simple,
				ClassName:  className,
				IsStatic:   isStatic,
				IsWildcard: isWild,
				Line:       lineNum,
			})
		}
	}

	// 2. Scan for classes, records, enums, and interfaces
	var pendingAnnotations []string
	for i := 0; i < len(lines); i++ {
		line := stripJavaComment(lines[i])
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if am := reJavaAnnotation.FindStringSubmatch(trimmed); am != nil && !strings.Contains(trimmed, "class ") && !strings.Contains(trimmed, "interface ") && !strings.Contains(trimmed, "record ") {
			pendingAnnotations = append(pendingAnnotations, strings.TrimSpace(am[1]))
			continue
		}

		// Class / Record / Enum declaration
		if cm := reJavaClass.FindStringSubmatch(line); cm != nil {
			endIdx := findMatchingBrace(lines, i)
			kind := strings.TrimSpace(cm[1])
			clsName := strings.TrimSpace(cm[2])
			cls := Class{
				Name:        clsName,
				Kind:        kind,
				Extends:     UnwrapJavaTypeAnnotation(cm[4]),
				Annotations: pendingAnnotations,
				LineStart:   i + 1,
				LineEnd:     endIdx + 1,
				IsPublic:    strings.Contains(line, "public "),
				IsAbstract:  strings.Contains(line, "abstract "),
			}
			pendingAnnotations = nil

			// Java 16+ record header components act as final fields
			if kind == "record" && strings.TrimSpace(cm[3]) != "" {
				localMap := make(map[string]string)
				parseJavaParams(cm[3], localMap)
				for fName, fType := range localMap {
					cls.Fields = append(cls.Fields, Field{
						Name:    fName,
						Type:    fType,
						IsFinal: true,
					})
				}
			}

			if strings.TrimSpace(cm[5]) != "" {
				for _, impl := range strings.Split(cm[5], ",") {
					if v := UnwrapJavaTypeAnnotation(impl); v != "" {
						cls.Implements = append(cls.Implements, v)
					}
				}
			}

			parseJavaClassBody(lines, i+1, endIdx, &cls)
			m.Classes = append(m.Classes, cls)
			i = endIdx
			continue
		}

		// Interface declaration
		if im := reJavaInterface.FindStringSubmatch(line); im != nil {
			endIdx := findMatchingBrace(lines, i)
			iface := Interface{
				Name:        strings.TrimSpace(im[1]),
				Annotations: pendingAnnotations,
				LineStart:   i + 1,
				LineEnd:     endIdx + 1,
				IsPublic:    strings.Contains(line, "public "),
			}
			pendingAnnotations = nil
			if strings.TrimSpace(im[2]) != "" {
				for _, ext := range strings.Split(im[2], ",") {
					if v := UnwrapJavaTypeAnnotation(ext); v != "" {
						iface.Extends = append(iface.Extends, v)
					}
				}
			}
			parseJavaInterfaceBody(lines, i+1, endIdx, &iface)
			m.Interfaces = append(m.Interfaces, iface)
			i = endIdx
			continue
		}
	}

	return m, nil
}

func parseJavaClassBody(lines []string, startIdx, endIdx int, cls *Class) {
	var memberAnns []string
	fieldTypes := make(map[string]string)
	for _, f := range cls.Fields {
		if f.Type != "" {
			fieldTypes[f.Name] = f.Type
		}
	}

	// Pass 1: collect all declared fields so methods & constructors know field types
	for i := startIdx; i < endIdx && i < len(lines); i++ {
		line := stripJavaComment(lines[i])
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || trimmed == "}" {
			continue
		}
		if am := reJavaAnnotation.FindStringSubmatch(trimmed); am != nil && !strings.Contains(trimmed, "(") && !strings.HasSuffix(trimmed, ";") {
			memberAnns = append(memberAnns, strings.TrimSpace(am[1]))
			continue
		}
		if fm := reJavaField.FindStringSubmatch(line); fm != nil && !strings.Contains(line, "(") {
			fType := UnwrapJavaTypeAnnotation(fm[1])
			fName := strings.TrimSpace(fm[2])
			if fName != "" && !isJavaKeyword(fName) {
				cls.Fields = append(cls.Fields, Field{
					Name:        fName,
					Type:        fType,
					Annotations: memberAnns,
					IsFinal:     strings.Contains(line, "final "),
					IsStatic:    strings.Contains(line, "static "),
				})
				if fType != "" {
					fieldTypes[fName] = fType
				}
			}
			memberAnns = nil
		}
	}

	// Pass 2: parse constructors and methods
	memberAnns = nil
	for i := startIdx; i < endIdx && i < len(lines); i++ {
		line := stripJavaComment(lines[i])
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || trimmed == "}" {
			continue
		}
		if am := reJavaAnnotation.FindStringSubmatch(trimmed); am != nil && !strings.HasSuffix(trimmed, "{") && !strings.HasSuffix(trimmed, ";") {
			memberAnns = append(memberAnns, strings.TrimSpace(am[1]))
			continue
		}

		// Constructor matching class name
		if ctm := reJavaCtor.FindStringSubmatch(line); ctm != nil && strings.TrimSpace(ctm[1]) == cls.Name {
			mEnd := findMatchingBrace(lines, i)
			localTypes := make(map[string]string)
			params := parseJavaParams(ctm[2], localTypes)
			// Propagate constructor parameter types assigned via `this.field = param;`
			for k := i + 1; k < mEnd && k < len(lines); k++ {
				if fam := reJavaFieldAssign.FindStringSubmatch(stripJavaComment(lines[k])); fam != nil {
					fName := strings.TrimSpace(fam[1])
					pName := strings.TrimSpace(fam[2])
					if pType, ok := localTypes[pName]; ok && pType != "" {
						fieldTypes[fName] = pType
						cls.Fields = append(cls.Fields, Field{Name: fName, Type: pType, IsFinal: true})
					}
				}
			}
			for fName, fType := range fieldTypes {
				localTypes["this."+fName] = fType
				if _, exists := localTypes[fName]; !exists {
					localTypes[fName] = fType
				}
			}
			mth := Method{
				Name:          cls.Name,
				Params:        params,
				Annotations:   memberAnns,
				IsPublic:      !strings.Contains(line, "private "),
				IsConstructor: true,
				Receiver:      cls.Name,
				LineStart:     i + 1,
				LineEnd:       mEnd + 1,
				LocalTypes:    localTypes,
			}
			mth.Calls, mth.CallSites = extractJavaBodyCalls(lines[i:mEnd+1], i+1, cls.Name, localTypes)
			cls.Methods = append(cls.Methods, mth)
			memberAnns = nil
			i = mEnd
			continue
		}

		// Standard method
		if mm := reJavaMethod.FindStringSubmatch(line); mm != nil {
			retRaw := strings.TrimSpace(mm[1])
			mName := strings.TrimSpace(mm[2])
			if isJavaKeyword(mName) || isJavaKeyword(retRaw) || mName == cls.Name {
				continue
			}
			hasBody := strings.TrimSpace(mm[4]) == "{"
			mEnd := i
			if hasBody {
				mEnd = findMatchingBrace(lines, i)
			}
			localTypes := make(map[string]string)
			for fName, fType := range fieldTypes {
				localTypes["this."+fName] = fType
				localTypes[fName] = fType
			}
			params := parseJavaParams(mm[3], localTypes)
			mth := Method{
				Name:        mName,
				Params:      params,
				ReturnType:  UnwrapJavaTypeAnnotation(retRaw),
				Annotations: memberAnns,
				IsPublic:    !strings.Contains(line, "private "),
				IsStatic:    strings.Contains(line, "static "),
				Receiver:    cls.Name,
				LineStart:   i + 1,
				LineEnd:     mEnd + 1,
				LocalTypes:  localTypes,
			}
			if hasBody {
				mth.Calls, mth.CallSites = extractJavaBodyCalls(lines[i:mEnd+1], i+1, mName, localTypes)
			}
			cls.Methods = append(cls.Methods, mth)
			memberAnns = nil
			i = mEnd
			continue
		}
	}
}

func parseJavaInterfaceBody(lines []string, startIdx, endIdx int, iface *Interface) {
	for i := startIdx; i < endIdx && i < len(lines); i++ {
		line := stripJavaComment(lines[i])
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || trimmed == "}" || strings.HasPrefix(trimmed, "@") {
			continue
		}
		if mm := reJavaMethod.FindStringSubmatch(line); mm != nil {
			mName := strings.TrimSpace(mm[2])
			if isJavaKeyword(mName) {
				continue
			}
			localTypes := make(map[string]string)
			params := parseJavaParams(mm[3], localTypes)
			iface.Methods = append(iface.Methods, Method{
				Name:       mName,
				Params:     params,
				ReturnType: UnwrapJavaTypeAnnotation(mm[1]),
				IsPublic:   true,
				Receiver:   iface.Name,
				LineStart:  i + 1,
				LineEnd:    i + 1,
				LocalTypes: localTypes,
			})
		}
	}
}

func parseJavaParams(rawParams string, localTypes map[string]string) []string {
	var names []string
	if strings.TrimSpace(rawParams) == "" {
		return names
	}
	for _, part := range splitTopLevelCommas(rawParams) {
		p := strings.TrimSpace(part)
		if p == "" {
			continue
		}
		// Strip leading annotations like @RequestBody, @PathVariable("id"), final
		for strings.HasPrefix(p, "@") {
			if idx := strings.IndexAny(p, " \t)"); idx > 0 {
				if p[idx] == ')' {
					p = strings.TrimSpace(p[idx+1:])
				} else {
					p = strings.TrimSpace(p[idx:])
				}
			} else {
				break
			}
		}
		p = strings.TrimSpace(strings.TrimPrefix(p, "final "))
		fields := strings.Fields(p)
		if len(fields) >= 2 {
			pName := strings.TrimSpace(fields[len(fields)-1])
			typePart := strings.Join(fields[:len(fields)-1], " ")
			tName := UnwrapJavaTypeAnnotation(typePart)
			if pName != "" {
				names = append(names, pName)
				if tName != "" && localTypes != nil {
					localTypes[pName] = tName
				}
			}
		}
	}
	return names
}

func splitTopLevelCommas(s string) []string {
	var parts []string
	depth := 0
	start := 0
	for i, r := range s {
		switch r {
		case '<', '(':
			depth++
		case '>', ')':
			if depth > 0 {
				depth--
			}
		case ',':
			if depth == 0 {
				parts = append(parts, s[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, s[start:])
	return parts
}

func extractJavaBodyCalls(bodyLines []string, startLine int, selfName string, localTypes map[string]string) ([]string, []CallSite) {
	seen := make(map[string]bool)
	var calls []string
	var sites []CallSite

	addCall := func(target string, line int) {
		target = strings.TrimSpace(target)
		if target == "" || target == selfName || isJavaKeyword(target) {
			return
		}
		if !seen[target] {
			seen[target] = true
			calls = append(calls, target)
			sites = append(sites, CallSite{Target: target, Line: line})
		}
	}

	for offset, rawLine := range bodyLines {
		lineNo := startLine + offset
		line := stripJavaComment(rawLine)
		if offset == 0 {
			if brace := strings.Index(line, "{"); brace >= 0 {
				line = line[brace+1:]
			} else {
				continue
			}
		}
		if strings.TrimSpace(line) == "" {
			continue
		}

		// 1. Track `Type varName = new ClassName(...)` or `var varName = new ClassName(...)`
		for _, nm := range reJavaNewAssign.FindAllStringSubmatch(line, -1) {
			declType := UnwrapJavaTypeAnnotation(nm[1])
			lhs := strings.TrimSpace(nm[2])
			ctorType := UnwrapJavaTypeAnnotation(nm[3])
			if localTypes != nil && lhs != "" {
				if ctorType != "" {
					localTypes[lhs] = ctorType
				} else if declType != "" {
					localTypes[lhs] = declType
				}
			}
			if ctorType != "" {
				addCall(ctorType, lineNo)
			}
		}

		// 2. Track `Type varName = this.getRepo(...)` or `var varName = getRepo(...)`
		for _, cm := range reJavaCallAssign.FindAllStringSubmatch(line, -1) {
			declType := UnwrapJavaTypeAnnotation(cm[1])
			lhs := strings.TrimSpace(cm[2])
			callee := strings.TrimSpace(cm[3])
			if localTypes != nil && lhs != "" && callee != "new" {
				if declType != "" {
					localTypes[lhs] = declType
				} else {
					localTypes[lhs] = "@return:" + callee
				}
			}
		}

		// 3. Track explicit typed local declarations `OrderRepository repo;`
		if tm := reJavaTypedLocal.FindStringSubmatch(line); tm != nil && localTypes != nil {
			tName := UnwrapJavaTypeAnnotation(tm[1])
			vName := strings.TrimSpace(tm[2])
			if tName != "" && vName != "" && localTypes[vName] == "" {
				localTypes[vName] = tName
			}
		}

		// 4. Track chained calls `this.getRepo().fetchOrder(id)`
		for _, ch := range reJavaChainCall.FindAllStringSubmatch(line, -1) {
			firstCall := strings.TrimSpace(ch[1])
			chainedMethod := strings.TrimSpace(ch[2])
			if firstCall != "" && chainedMethod != "" {
				addCall(firstCall, lineNo)
				addCall("@chain:"+firstCall+"."+chainedMethod, lineNo)
			}
		}

		// 5. Track standalone `new ClassName(...)`
		for _, ne := range reJavaNewExpr.FindAllStringSubmatch(line, -1) {
			if cName := UnwrapJavaTypeAnnotation(ne[1]); cName != "" {
				addCall(cName, lineNo)
			}
		}

		// 6. Track standard method/function calls
		for _, ce := range reJavaCallExpr.FindAllStringSubmatch(line, -1) {
			callee := strings.TrimSpace(ce[1])
			if callee == "" || isJavaKeyword(callee) {
				continue
			}
			addCall(callee, lineNo)
		}
	}

	return calls, sites
}

func findMatchingBrace(lines []string, startIdx int) int {
	depth := 0
	started := false
	for i := startIdx; i < len(lines); i++ {
		line := stripJavaComment(lines[i])
		for _, ch := range line {
			if ch == '{' {
				depth++
				started = true
			} else if ch == '}' {
				depth--
				if started && depth <= 0 {
					return i
				}
			}
		}
	}
	return startIdx
}

func stripJavaComment(line string) string {
	if idx := strings.Index(line, "//"); idx >= 0 {
		return line[:idx]
	}
	return line
}

func isJavaKeyword(s string) bool {
	switch s {
	case "if", "for", "while", "switch", "catch", "synchronized", "return", "throw", "new", "super", "this", "assert", "try", "else", "case", "default":
		return true
	}
	return false
}
