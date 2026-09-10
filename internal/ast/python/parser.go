// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package python

import (
	"regexp"
	"strings"
)

// ─────────────────────────────────────────────────────────────────────────────
// Public types
// ─────────────────────────────────────────────────────────────────────────────

type Module struct {
	Path      string
	Module    string // dotted module path derived from file
	Classes   []Class
	Functions []Function
	Imports   []Import
}

type Class struct {
	Name       string
	Bases      []string
	Decorators []string
	Docstring  string
	LineStart  int
	LineEnd    int
	IsExported bool
	Methods    []Function
}

type Function struct {
	Name       string
	Params     []string
	ReturnType string
	Decorators []string
	IsAsync    bool
	Receiver   string // enclosing class name if method
	Docstring  string
	LineStart  int
	LineEnd    int
	IsExported bool
}

type Import struct {
	Module string
	Names  []string // for `from X import a, b`
	Alias  string   // for `import X as Y`
	Line   int
}

// analysisFrame tracks an open block (class or function) during parsing.
type analysisFrame struct {
	kind     string // "class" or "def"
	indent   int
	classIdx int // index into m.Classes, -1 if none
	funcIdx  int // index into m.Functions or m.Classes[classIdx].Methods
}

// ─────────────────────────────────────────────────────────────────────────────
// Regexes — applied only to logical, string-stripped lines
// ─────────────────────────────────────────────────────────────────────────────

var (
	reClass     = regexp.MustCompile(`^class\s+([A-Za-z_][A-Za-z0-9_]*)\s*(?:\(([^)]*)\))?\s*:`)
	reDef       = regexp.MustCompile(`^(?:async\s+)?def\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(([^)]*)\)\s*(?:->\s*([^:]+))?\s*:`)
	reImport    = regexp.MustCompile(`^import\s+(.+)$`)
	reFromImpt  = regexp.MustCompile(`^from\s+([A-Za-z0-9_.]+)\s+import\s+(.+)$`)
	reDecor     = regexp.MustCompile(`^@([A-Za-z_][A-Za-z0-9_.]*(?:\([^)]*\))?)`)
	reDocstring = regexp.MustCompile(`^\s*(?:"""|''')`)
)

// ─────────────────────────────────────────────────────────────────────────────
// Parse
// ─────────────────────────────────────────────────────────────────────────────

// Parse reads Python source and returns a structural module.
func Parse(src, path, moduleName string) (*Module, error) {
	lines := stripAndClean(src)
	m := &Module{Path: path, Module: moduleName}

	// Block stack — tracks (kind, indent, classIdx, funcIdx)
	type frame struct {
		kind     string // "class" or "def"
		indent   int
		classIdx int // index into m.Classes, -1 if none
		funcIdx  int // index into m.Functions or m.Classes[classIdx].Methods
	}

	var stack []analysisFrame

	lastDecorators := []string{}

	for i := 0; i < len(lines); i++ {
		ln := lines[i]
		if strings.TrimSpace(ln.Text) == "" {
			continue
		}

		indent := ln.Indent
		text := ln.Text

		// Pop stack: any block with indent >= current is closed
		for len(stack) > 0 && indent <= stack[len(stack)-1].indent {
			closeBlock(m, &stack, lines, i)
		}
		// Decorator — accumulate for next def/class
		if dm := reDecor.FindStringSubmatch(text); dm != nil {
			lastDecorators = append(lastDecorators, dm[1])
			continue
		}

		// Class
		if cm := reClass.FindStringSubmatch(text); cm != nil {
			cls := Class{
				Name:       cm[1],
				Bases:      parseBases(cm[2]),
				Decorators: lastDecorators,
				LineStart:  ln.LineNum,
				IsExported: !strings.HasPrefix(cm[1], "_"),
			}
			lastDecorators = nil
			m.Classes = append(m.Classes, cls)
			stack = append(stack, analysisFrame{kind: "class", indent: indent, classIdx: len(m.Classes) - 1, funcIdx: -1})
			continue
		}

		// Function / method
		if dm := reDef.FindStringSubmatch(text); dm != nil {
			fn := Function{
				Name:       dm[1],
				Params:     parseParams(dm[2]),
				ReturnType: strings.TrimSpace(dm[3]),
				Decorators: lastDecorators,
				IsAsync:    strings.HasPrefix(strings.TrimSpace(text), "async"),
				LineStart:  ln.LineNum,
				IsExported: !strings.HasPrefix(dm[1], "_"),
			}
			lastDecorators = nil

			// Attach to enclosing class if stack top is a class
			if len(stack) > 0 && stack[len(stack)-1].kind == "class" {
				ci := stack[len(stack)-1].classIdx
				fn.Receiver = m.Classes[ci].Name
				m.Classes[ci].Methods = append(m.Classes[ci].Methods, fn)
				stack = append(stack, analysisFrame{kind: "def", indent: indent, classIdx: ci, funcIdx: len(m.Classes[ci].Methods) - 1})
			} else {
				m.Functions = append(m.Functions, fn)
				stack = append(stack, analysisFrame{kind: "def", indent: indent, classIdx: -1, funcIdx: len(m.Functions) - 1})
			}
			continue
		}

		// Imports — only at module level (indent == 0)
		if indent == 0 {
			if im := reImport.FindStringSubmatch(text); im != nil {
				for _, spec := range splitTopLevel(im[1], ',') {
					spec = strings.TrimSpace(spec)
					if spec == "" {
						continue
					}
					imp := Import{Line: ln.LineNum}
					if asIdx := strings.Index(spec, " as "); asIdx >= 0 {
						imp.Module = strings.TrimSpace(spec[:asIdx])
						imp.Alias = strings.TrimSpace(spec[asIdx+4:])
					} else {
						imp.Module = spec
					}
					m.Imports = append(m.Imports, imp)
				}
				continue
			}
			if fm := reFromImpt.FindStringSubmatch(text); fm != nil {
				mod := fm[1]
				names := []string{}
				for _, n := range splitTopLevel(fm[2], ',') {
					n = strings.TrimSpace(n)
					if n == "" {
						continue
					}
					names = append(names, n)
				}
				m.Imports = append(m.Imports, Import{Module: mod, Names: names, Line: ln.LineNum})
				continue
			}
		}
	}

	// Close any remaining open blocks
	for len(stack) > 0 {
		closeBlock(m, &stack, lines, len(lines))
	}

	// Attach docstrings
	attachDocstrings(m, lines)

	return m, nil
}

func closeBlock(m *Module, stack *[]analysisFrame, lines []cleanLine, endIdx int) {
	if len(*stack) == 0 {
		return
	}
	f := (*stack)[len(*stack)-1]
	*stack = (*stack)[:len(*stack)-1]

	endLine := 0
	if endIdx > 0 && endIdx <= len(lines) {
		endLine = lines[endIdx-1].LineNum
	} else if len(lines) > 0 {
		endLine = lines[len(lines)-1].LineNum
	}

	if f.kind == "class" && f.classIdx >= 0 && f.classIdx < len(m.Classes) {
		m.Classes[f.classIdx].LineEnd = endLine
	}
	if f.kind == "def" {
		if f.classIdx >= 0 && f.classIdx < len(m.Classes) {
			methods := m.Classes[f.classIdx].Methods
			if f.funcIdx >= 0 && f.funcIdx < len(methods) {
				m.Classes[f.classIdx].Methods[f.funcIdx].LineEnd = endLine
			}
		} else if f.funcIdx >= 0 && f.funcIdx < len(m.Functions) {
			m.Functions[f.funcIdx].LineEnd = endLine
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Cleaner: strip comments, handle multi-line strings, implicit continuations
// ─────────────────────────────────────────────────────────────────────────────

type cleanLine struct {
	Text    string
	Indent  int
	LineNum int
	Blank   bool
}

// stripAndClean produces a slice of logical lines with comments and strings removed
// from structural analysis, but preserving indentation for block detection.
func stripAndClean(src string) []cleanLine {
	rawLines := strings.Split(src, "\n")
	out := make([]cleanLine, 0, len(rawLines))

	inTripleQuote := false
	var tripleDelim string
	bracketDepth := 0
	var buffer strings.Builder
	bufferLineNum := 0
	bufferIndent := 0

	for i, raw := range rawLines {
		lineNum := i + 1
		line := strings.TrimRight(raw, " \t\r")

		// If we're in a triple-quoted string, check for end
		if inTripleQuote {
			if idx := strings.Index(line, tripleDelim); idx >= 0 {
				inTripleQuote = false
				// Append remainder to buffer (the code after the closing delimiter)
				remainder := line[idx+3:]
				buffer.WriteString(" ")
				buffer.WriteString(stripComment(remainder))
			}
			continue
		}

		// Check if this line starts a triple-quoted string
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, `"""`) || strings.HasPrefix(trimmed, `'''`) {
			delim := `"""`
			if strings.HasPrefix(trimmed, `'''`) {
				delim = `'''`
			}
			// Check if it opens AND closes on the same line
			rest := trimmed[3:]
			if !strings.Contains(rest, delim) {
				inTripleQuote = true
				tripleDelim = delim
			}
			// Either way, we don't care about the content for structural purposes
			continue
		}

		// Strip comment (but respect quotes)
		stripped := stripComment(line)
		if strings.TrimSpace(stripped) == "" {
			continue
		}

		// Compute logical line: if inside brackets, accumulate
		if bracketDepth > 0 {
			buffer.WriteString(" ")
			buffer.WriteString(strings.TrimSpace(stripped))
			bracketDepth += bracketDelta(stripped)
			if bracketDepth <= 0 {
				out = append(out, cleanLine{
					Text:    strings.TrimSpace(buffer.String()),
					Indent:  bufferIndent,
					LineNum: bufferLineNum,
				})
				buffer.Reset()
				bracketDepth = 0
			}
			continue
		}

		// Compute indent from leading whitespace (tabs = 4 spaces)
		indent := 0
		for _, c := range line {
			if c == ' ' {
				indent++
			} else if c == '\t' {
				indent += 4
			} else {
				break
			}
		}

		delta := bracketDelta(stripped)
		if delta > 0 {
			buffer.Reset()
			buffer.WriteString(strings.TrimSpace(stripped))
			bufferIndent = indent
			bufferLineNum = lineNum
			bracketDepth = delta
			continue
		}

		out = append(out, cleanLine{
			Text:    strings.TrimSpace(stripped),
			Indent:  indent,
			LineNum: lineNum,
		})
	}

	return out
}

// stripComment removes a trailing `# comment` while respecting string literals.
func stripComment(line string) string {
	inSingle, inDouble := false, false
	for i := 0; i < len(line); i++ {
		c := line[i]
		if c == '\\' && i+1 < len(line) {
			i++
			continue
		}
		switch {
		case c == '\'' && !inDouble:
			inSingle = !inSingle
		case c == '"' && !inSingle:
			inDouble = !inDouble
		case c == '#' && !inSingle && !inDouble:
			return line[:i]
		}
	}
	return line
}

// bracketDelta returns net open brackets in a line.
func bracketDelta(line string) int {
	depth := 0
	inSingle, inDouble := false, false
	for i := 0; i < len(line); i++ {
		c := line[i]
		if c == '\\' && i+1 < len(line) {
			i++
			continue
		}
		switch {
		case c == '\'' && !inDouble:
			inSingle = !inSingle
		case c == '"' && !inSingle:
			inDouble = !inDouble
		case inSingle || inDouble:
			continue
		case c == '(' || c == '[' || c == '{':
			depth++
		case c == ')' || c == ']' || c == '}':
			depth--
		}
	}
	return depth
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func parseBases(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	var bases []string
	for _, b := range splitTopLevel(s, ',') {
		b = strings.TrimSpace(b)
		if b == "" || b == "metaclass=type" {
			continue
		}
		// Drop keyword args like metaclass=...
		if strings.Contains(b, "=") {
			continue
		}
		bases = append(bases, b)
	}
	return bases
}

func parseParams(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	var params []string
	for _, p := range splitTopLevel(s, ',') {
		p = strings.TrimSpace(p)
		if p == "" || p == "*" || p == "/" {
			continue
		}
		// Name = default → keep name only
		if eq := strings.Index(p, "="); eq >= 0 {
			p = strings.TrimSpace(p[:eq])
		}
		// Annotations: name: Type → keep name
		if colon := strings.Index(p, ":"); colon >= 0 {
			p = strings.TrimSpace(p[:colon])
		}
		// *args / **kwargs → strip leading *
		p = strings.TrimLeft(p, "*")
		params = append(params, p)
	}
	return params
}

// splitTopLevel splits s by sep, ignoring separators inside brackets or strings.
func splitTopLevel(s string, sep rune) []string {
	var parts []string
	depth := 0
	inSingle, inDouble := false, false
	var buf strings.Builder
	for i := 0; i < len(s); i++ {
		c := rune(s[i])
		if c == '\\' && i+1 < len(s) {
			buf.WriteRune(c)
			i++
			buf.WriteByte(s[i])
			continue
		}
		switch {
		case c == '\'' && !inDouble:
			inSingle = !inSingle
		case c == '"' && !inSingle:
			inDouble = !inDouble
		case !inSingle && !inDouble && (c == '(' || c == '[' || c == '{'):
			depth++
		case !inSingle && !inDouble && (c == ')' || c == ']' || c == '}'):
			depth--
		case c == sep && depth == 0 && !inSingle && !inDouble:
			parts = append(parts, buf.String())
			buf.Reset()
			continue
		}
		buf.WriteRune(c)
	}
	if buf.Len() > 0 {
		parts = append(parts, buf.String())
	}
	return parts
}

// attachDocstrings walks a module's classes and functions, looking for
// triple-quoted strings immediately after the def/class line.
func attachDocstrings(m *Module, lines []cleanLine) {
	// Not critical for v1 — left as a future enhancement.
	// The scanner already strips docstrings cleanly.
	_ = m
	_ = lines
}
