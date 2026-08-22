// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package analyzer

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// MultiModuleImporter implements types.Importer and types.ImporterFrom.
// It resolves package imports across multiple modules indexed in WorkspaceContext,
// falling back to stdlib/installed packages via default toolchain importer.
type MultiModuleImporter struct {
	mu         sync.RWMutex
	fset       *token.FileSet
	ws         *WorkspaceContext
	packages   map[string]*types.Package
	inProgress map[string]bool
	fallback   types.Importer
}

// NewMultiModuleImporter creates an authoritative cross-module type importer.
func NewMultiModuleImporter(fset *token.FileSet, ws *WorkspaceContext) *MultiModuleImporter {
	return &MultiModuleImporter{
		fset:       fset,
		ws:         ws,
		packages:   make(map[string]*types.Package),
		inProgress: make(map[string]bool),
		fallback:   importer.Default(),
	}
}

// Import resolves a package import path.
func (imp *MultiModuleImporter) Import(path string) (*types.Package, error) {
	return imp.ImportFrom(path, "", 0)
}

// ImportFrom resolves a package import path with contextual directory awareness.
func (imp *MultiModuleImporter) ImportFrom(path, srcDir string, mode types.ImportMode) (*types.Package, error) {
	imp.mu.Lock()

	// 1. Check if package is already fully resolved in cache
	if pkg, ok := imp.packages[path]; ok && pkg.Complete() {
		imp.mu.Unlock()
		return pkg, nil
	}

	// 2. Cycle detection guard
	if imp.inProgress[path] {
		imp.mu.Unlock()
		return nil, fmt.Errorf("import cycle detected while loading package: %s", path)
	}

	// 3. Locate physical directory across workspace modules
	dirPath, found := imp.ws.PackageRoots[path]
	if !found {
		imp.mu.Unlock()
		// Fallback to stdlib / installed packages
		return imp.fallback.Import(path)
	}

	imp.inProgress[path] = true
	imp.mu.Unlock()

	defer func() {
		imp.mu.Lock()
		delete(imp.inProgress, path)
		imp.mu.Unlock()
	}()

	// 4. Parse non-test Go source files in the target workspace directory
	files, pkgName, err := imp.parseWorkspacePackageFiles(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse workspace package %s at %s: %w", path, dirPath, err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("package %s at %s contains no buildable Go source files", path, dirPath)
	}

	// 5. Type-check package with multi-module importer
	pkg := types.NewPackage(path, pkgName)
	conf := types.Config{
		Importer: imp,
		Error:    func(err error) {}, // Non-fatal collector for robust AST resolution
	}

	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}

	checker := types.NewChecker(&conf, imp.fset, pkg, info)
	if err := checker.Files(files); err != nil {
		// Non-fatal: type errors during cross-module checking still produce partial types.Package
	}

	imp.mu.Lock()
	imp.packages[path] = pkg
	imp.mu.Unlock()

	return pkg, nil
}

// parseWorkspacePackageFiles reads and parses all non-test .go files in the directory.
func (imp *MultiModuleImporter) parseWorkspacePackageFiles(dirPath string) ([]*ast.File, string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, "", err
	}

	var files []*ast.File
	var detectedPkgName string

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}

		filePath := filepath.Join(dirPath, entry.Name())
		src, err := os.ReadFile(filePath)
		if err != nil {
			return nil, "", err
		}

		fileAst, err := parser.ParseFile(imp.fset, filePath, src, parser.ParseComments)
		if err != nil {
			return nil, "", err
		}

		if detectedPkgName == "" {
			detectedPkgName = fileAst.Name.Name
		} else if fileAst.Name.Name != detectedPkgName {
			// Skip documentation or mismatching package names in the same directory
			continue
		}

		files = append(files, fileAst)
	}

	return files, detectedPkgName, nil
}
