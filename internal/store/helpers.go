// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package store

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DetectModulePath reads go.mod and extracts the module path
func DetectModulePath(repoPath string) (string, error) {
	modPath := filepath.Join(repoPath, "go.mod")
	f, err := os.Open(modPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if modPath, ok := parseModuleDirective(line); ok {
			return modPath, nil
		}
	}
	return "", fmt.Errorf("module path not found in go.mod")
}

// parseModuleDirective extracts the module path from a go.mod "module" line.
// It strips trailing "// comments" (valid in go.mod, e.g. `module x/y // v1.9.8`)
// and optional surrounding quotes, which a naive TrimPrefix leaves in the path.
func parseModuleDirective(line string) (string, bool) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "module") {
		return "", false
	}
	rest := line[len("module"):]
	if rest == "" || (rest[0] != ' ' && rest[0] != '\t') {
		return "", false
	}
	if i := strings.Index(rest, "//"); i >= 0 {
		rest = rest[:i]
	}
	rest = strings.TrimSpace(rest)
	if len(rest) >= 2 && (rest[0] == '"' || rest[0] == '`') && rest[len(rest)-1] == rest[0] {
		rest = rest[1 : len(rest)-1]
	}
	if rest == "" {
		return "", false
	}
	return rest, true
}
