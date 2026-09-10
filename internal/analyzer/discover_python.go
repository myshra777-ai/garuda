// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package analyzer

import (
	"os"
	"path/filepath"
	"strings"
)

// DiscoverPythonWorkspace finds a Python project root under absPath.
// Returns the root path and a display name for the module.
func DiscoverPythonWorkspace(absPath string) (string, string, error) {
	// Strategy 1: pyproject.toml at root
	if fileExists(filepath.Join(absPath, "pyproject.toml")) {
		name := readProjectName(filepath.Join(absPath, "pyproject.toml"))
		return absPath, name, nil
	}

	// Strategy 2: setup.py at root
	if fileExists(filepath.Join(absPath, "setup.py")) {
		return absPath, filepath.Base(absPath), nil
	}

	// Strategy 3: Walk for .py files, group by common ancestor
	pyDirs := make(map[string]int)
	_ = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			base := info.Name()
			if strings.HasPrefix(base, ".") || base == "__pycache__" ||
				base == "venv" || base == ".venv" || base == "env" ||
				base == "node_modules" || base == "site-packages" ||
				base == "dist" || base == "build" || base == ".tox" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) == ".py" {
			pyDirs[filepath.Dir(path)]++
		}
		return nil
	})

	if len(pyDirs) == 0 {
		return "", "", os.ErrNotExist
	}

	// Pick the shallowest directory that has .py files
	var chosenRoot string
	for dir := range pyDirs {
		if chosenRoot == "" || len(dir) < len(chosenRoot) {
			chosenRoot = dir
		}
	}
	return chosenRoot, filepath.Base(chosenRoot), nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func readProjectName(pyprojectPath string) string {
	data, err := os.ReadFile(pyprojectPath)
	if err != nil {
		return filepath.Base(filepath.Dir(pyprojectPath))
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "name") && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				v := strings.TrimSpace(parts[1])
				v = strings.Trim(v, `"'`)
				if v != "" {
					return v
				}
			}
		}
	}
	return filepath.Base(filepath.Dir(pyprojectPath))
}
