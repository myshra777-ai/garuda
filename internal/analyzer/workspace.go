//Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package analyzer

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

// ModuleDescriptor records the identity, filesystem location, and Go version of a Go module.
type ModuleDescriptor struct {
	ModulePath string `json:"module_path"`
	RootPath   string `json:"root_path"`
	GoVersion  string `json:"go_version"`
}

// WorkspaceContext indexes all modules and resolvable package import paths across the workspace.
type WorkspaceContext struct {
	RootPath     string                       `json:"root_path"`
	IsGoWork     bool                         `json:"is_go_work"`
	Modules      map[string]*ModuleDescriptor `json:"modules"`       // Key: ModulePath -> ModuleDescriptor
	PackageRoots map[string]string            `json:"package_roots"` // Key: Fully-Qualified Import Path -> Directory Path
}

// DiscoverWorkspace scans the given root directory for `go.work` or standalone `go.mod` files
// and builds a unified index mapping every package import path to its filesystem directory.
func DiscoverWorkspace(rootDir string) (*WorkspaceContext, error) {
	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve workspace absolute path: %w", err)
	}

	info, err := os.Stat(absRoot)
	if err != nil {
		return nil, fmt.Errorf("workspace root does not exist: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("workspace root is not a directory: %s", absRoot)
	}

	ws := &WorkspaceContext{
		RootPath:     absRoot,
		Modules:      make(map[string]*ModuleDescriptor),
		PackageRoots: make(map[string]string),
	}

	workFilePath := filepath.Join(absRoot, "go.work")
	if data, err := os.ReadFile(workFilePath); err == nil {
		ws.IsGoWork = true
		if err := parseGoWork(absRoot, data, ws); err != nil {
			return nil, fmt.Errorf("failed to parse go.work: %w", err)
		}
	} else {
		if err := scanForGoMods(absRoot, ws); err != nil {
			return nil, fmt.Errorf("failed to scan go.mod files: %w", err)
		}
	}

	if len(ws.Modules) == 0 {
		return nil, fmt.Errorf("no Go modules found in workspace root: %s", absRoot)
	}

	if err := indexWorkspacePackages(ws); err != nil {
		return nil, fmt.Errorf("failed to index workspace packages: %w", err)
	}

	return ws, nil
}

// parseGoWork extracts all active module directories declared in `go.work`.
func parseGoWork(workDir string, data []byte, ws *WorkspaceContext) error {
	f, err := modfile.ParseWork("go.work", data, nil)
	if err != nil {
		return err
	}

	for _, use := range f.Use {
		modRelPath := filepath.FromSlash(use.Path)
		modDir := filepath.Join(workDir, modRelPath)

		modFilePath := filepath.Join(modDir, "go.mod")
		modBytes, err := os.ReadFile(modFilePath)
		if err != nil {
			continue // Skip unreadable or missing nested module directories
		}

		mf, err := modfile.Parse("go.mod", modBytes, nil)
		if err != nil || mf.Module == nil {
			continue
		}

		goVersion := ""
		if mf.Go != nil {
			goVersion = mf.Go.Version
		}

		ws.Modules[mf.Module.Mod.Path] = &ModuleDescriptor{
			ModulePath: mf.Module.Mod.Path,
			RootPath:   modDir,
			GoVersion:  goVersion,
		}
	}
	return nil
}

// scanForGoMods recursively walks directories to locate all `go.mod` declarations.
func scanForGoMods(rootDir string, ws *WorkspaceContext) error {
	return filepath.WalkDir(rootDir, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}

		name := d.Name()
		if isIgnoredDir(name) && filePath != rootDir {
			return filepath.SkipDir
		}

		modFilePath := filepath.Join(filePath, "go.mod")
		modBytes, err := os.ReadFile(modFilePath)
		if err != nil {
			return nil
		}

		mf, err := modfile.Parse("go.mod", modBytes, nil)
		if err != nil || mf.Module == nil {
			return nil
		}

		goVersion := ""
		if mf.Go != nil {
			goVersion = mf.Go.Version
		}

		ws.Modules[mf.Module.Mod.Path] = &ModuleDescriptor{
			ModulePath: mf.Module.Mod.Path,
			RootPath:   filePath,
			GoVersion:  goVersion,
		}
		return nil
	})
}

// indexWorkspacePackages walks each discovered module root to map import paths to filesystem folders.
func indexWorkspacePackages(ws *WorkspaceContext) error {
	for modPath, modDesc := range ws.Modules {
		err := filepath.WalkDir(modDesc.RootPath, func(dirPath string, d fs.DirEntry, err error) error {
			if err != nil || !d.IsDir() {
				return nil
			}

			name := d.Name()
			if isIgnoredDir(name) && dirPath != modDesc.RootPath {
				return filepath.SkipDir
			}

			// Exclude sub-modules embedded inside this directory tree (they are indexed by their own descriptor)
			if dirPath != modDesc.RootPath {
				if _, err := os.Stat(filepath.Join(dirPath, "go.mod")); err == nil {
					return filepath.SkipDir
				}
			}

			// Only index directories that contain at least one .go file
			if !hasGoSourceFiles(dirPath) {
				return nil
			}

			rel, err := filepath.Rel(modDesc.RootPath, dirPath)
			if err != nil {
				return nil
			}

			var pkgImportPath string
			if rel == "." {
				pkgImportPath = modPath
			} else {
				pkgImportPath = path.Join(modPath, filepath.ToSlash(rel))
			}

			ws.PackageRoots[pkgImportPath] = dirPath
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func hasGoSourceFiles(dirPath string) bool {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
			return true
		}
	}
	return false
}

func isIgnoredDir(name string) bool {
	if name == ".git" || name == "vendor" || name == "node_modules" || name == "testdata" {
		return true
	}
	if len(name) > 1 && strings.HasPrefix(name, ".") {
		return true
	}
	return false
}
