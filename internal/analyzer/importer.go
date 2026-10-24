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

// MultiModuleImporter resolves imports across multiple modules in a workspace
// (including vendor/ directories) and falls back to the standard library
// importer for external packages.
type MultiModuleImporter struct {
	fset        *token.FileSet
	modules     []ModuleInfo
	cache       map[string]*types.Package
	inProgress  map[string]bool
	defaultImp  types.Importer
	mu          sync.Mutex
	Unresolved  []string
}

// NewMultiModuleImporter creates a new importer for the given workspace modules.
func NewMultiModuleImporter(fset *token.FileSet, modules []ModuleInfo) *MultiModuleImporter {
	return &MultiModuleImporter{
		fset:       fset,
		modules:    modules,
		cache:      make(map[string]*types.Package),
		inProgress: make(map[string]bool),
		defaultImp: importer.Default(),
	}
}

// Import implements types.Importer.
func (m *MultiModuleImporter) Import(path string) (*types.Package, error) {
	return m.ImportFrom(path, "", 0)
}

// ImportFrom implements types.ImporterFrom.
func (m *MultiModuleImporter) ImportFrom(path, dir string, mode types.ImportMode) (*types.Package, error) {
	m.mu.Lock()
	if pkg, ok := m.cache[path]; ok {
		m.mu.Unlock()
		return pkg, nil
	}
	if m.inProgress[path] {
		m.mu.Unlock()
		return nil, fmt.Errorf("import cycle detected on %s", path)
	}
	m.inProgress[path] = true
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		delete(m.inProgress, path)
		m.mu.Unlock()
	}()

	// 1. Check if the import path belongs to one of the workspace modules
	for _, mod := range m.modules {
		if path == mod.ModulePath || strings.HasPrefix(path, mod.ModulePath+"/") {
			relPath := strings.TrimPrefix(path, mod.ModulePath)
			relPath = strings.TrimPrefix(relPath, "/")
			targetDir := filepath.Join(mod.RootDir, filepath.FromSlash(relPath))

			if info, err := os.Stat(targetDir); err == nil && info.IsDir() {
				pkg, err := m.loadPackageFromDir(path, targetDir)
				if err == nil && pkg != nil {
					m.mu.Lock()
					m.cache[path] = pkg
					m.mu.Unlock()
					return pkg, nil
				}
			}
		}

		// 2. Check module-local vendor/ directory if present
		vendorDir := filepath.Join(mod.RootDir, "vendor", filepath.FromSlash(path))
		if info, err := os.Stat(vendorDir); err == nil && info.IsDir() {
			pkg, err := m.loadPackageFromDir(path, vendorDir)
			if err == nil && pkg != nil {
				m.mu.Lock()
				m.cache[path] = pkg
				m.mu.Unlock()
				return pkg, nil
			}
		}
	}

	// 3. Fallback to standard library / default Go importer
	if impFrom, ok := m.defaultImp.(types.ImporterFrom); ok {
		pkg, err := impFrom.ImportFrom(path, dir, mode)
		if err == nil {
			m.mu.Lock()
			m.cache[path] = pkg
			m.mu.Unlock()
			return pkg, nil
		}
	}

	pkg, err := m.defaultImp.Import(path)
	if err == nil {
		m.mu.Lock()
		m.cache[path] = pkg
		m.mu.Unlock()
		return pkg, nil
	}

	m.mu.Lock()
	m.Unresolved = append(m.Unresolved, path)
	m.mu.Unlock()
	return nil, err
}

func (m *MultiModuleImporter) loadPackageFromDir(importPath, dir string) (*types.Package, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []*ast.File
	var pkgName string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		fullPath := filepath.Join(dir, name)
		f, err := parser.ParseFile(m.fset, fullPath, nil, parser.ParseComments)
		if err != nil {
			continue
		}

		if pkgName == "" {
			pkgName = f.Name.Name
		} else if f.Name.Name != pkgName {
			continue
		}

		files = append(files, f)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no buildable Go files in %s", dir)
	}

	conf := types.Config{
		Importer:                 m,
		DisableUnusedImportCheck: true,
		Error:                    func(err error) {},
	}

	info := &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
	}

	pkg, _ := conf.Check(importPath, m.fset, files, info)
	if pkg == nil {
		return nil, fmt.Errorf("failed to type-check package %s", importPath)
	}

	return pkg, nil
}
