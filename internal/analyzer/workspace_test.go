// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package analyzer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverWorkspace_SingleModule(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Create root go.mod
	goModContent := `module github.com/example/single

go 1.22
`
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	// 2. Create root package Go file
	rootPkgFile := `package single

type Engine struct {
	ID string
}
`
	if err := os.WriteFile(filepath.Join(tempDir, "engine.go"), []byte(rootPkgFile), 0644); err != nil {
		t.Fatalf("failed to create engine.go: %v", err)
	}

	// 3. Create subpackage `pkg/types`
	subPkgDir := filepath.Join(tempDir, "pkg", "types")
	if err := os.MkdirAll(subPkgDir, 0755); err != nil {
		t.Fatalf("failed to create subpackage dir: %v", err)
	}
	subPkgFile := `package types

type Request struct {
	Query string
}
`
	if err := os.WriteFile(filepath.Join(subPkgDir, "types.go"), []byte(subPkgFile), 0644); err != nil {
		t.Fatalf("failed to create types.go: %v", err)
	}

	// 4. Create empty directory (should not be indexed)
	if err := os.MkdirAll(filepath.Join(tempDir, "docs"), 0755); err != nil {
		t.Fatalf("failed to create docs dir: %v", err)
	}

	// Run discovery
	ws, err := DiscoverWorkspace(tempDir)
	if err != nil {
		t.Fatalf("DiscoverWorkspace failed: %v", err)
	}

	if ws.IsGoWork {
		t.Errorf("expected IsGoWork to be false, got true")
	}

	if len(ws.Modules) != 1 {
		t.Fatalf("expected 1 module, got %d", len(ws.Modules))
	}

	mod, ok := ws.Modules["github.com/example/single"]
	if !ok {
		t.Fatalf("module github.com/example/single not found in context")
	}
	if mod.GoVersion != "1.22" {
		t.Errorf("expected GoVersion '1.22', got %q", mod.GoVersion)
	}

	// Check package roots
	expectedPkgs := map[string]string{
		"github.com/example/single":           tempDir,
		"github.com/example/single/pkg/types": subPkgDir,
	}

	if len(ws.PackageRoots) != len(expectedPkgs) {
		t.Fatalf("expected %d indexed packages, got %d: %+v", len(expectedPkgs), len(ws.PackageRoots), ws.PackageRoots)
	}

	for pkgPath, expectedDir := range expectedPkgs {
		dir, exists := ws.PackageRoots[pkgPath]
		if !exists {
			t.Errorf("missing package path %q from index", pkgPath)
			continue
		}
		if dir != expectedDir {
			t.Errorf("package %q path mismatch: expected %q, got %q", pkgPath, expectedDir, dir)
		}
	}
}

func TestDiscoverWorkspace_GoWorkMultiModule(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Create go.work
	goWorkContent := `go 1.22

use (
	./core
	./services/api
)
`
	if err := os.WriteFile(filepath.Join(tempDir, "go.work"), []byte(goWorkContent), 0644); err != nil {
		t.Fatalf("failed to create go.work: %v", err)
	}

	// 2. Create core module
	coreDir := filepath.Join(tempDir, "core")
	if err := os.MkdirAll(coreDir, 0755); err != nil {
		t.Fatalf("failed to create core dir: %v", err)
	}
	coreGoMod := `module github.com/example/core

go 1.22
`
	if err := os.WriteFile(filepath.Join(coreDir, "go.mod"), []byte(coreGoMod), 0644); err != nil {
		t.Fatalf("failed to write core go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(coreDir, "core.go"), []byte("package core\n"), 0644); err != nil {
		t.Fatalf("failed to write core.go: %v", err)
	}

	// 3. Create services/api module with nested package
	apiDir := filepath.Join(tempDir, "services", "api")
	apiPkgDir := filepath.Join(apiDir, "handler")
	if err := os.MkdirAll(apiPkgDir, 0755); err != nil {
		t.Fatalf("failed to create api handler dir: %v", err)
	}
	apiGoMod := `module github.com/example/api

go 1.22
`
	if err := os.WriteFile(filepath.Join(apiDir, "go.mod"), []byte(apiGoMod), 0644); err != nil {
		t.Fatalf("failed to write api go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(apiDir, "main.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatalf("failed to write api main.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(apiPkgDir, "handler.go"), []byte("package handler\n"), 0644); err != nil {
		t.Fatalf("failed to write handler.go: %v", err)
	}

	// Run discovery
	ws, err := DiscoverWorkspace(tempDir)
	if err != nil {
		t.Fatalf("DiscoverWorkspace failed: %v", err)
	}

	if !ws.IsGoWork {
		t.Errorf("expected IsGoWork to be true")
	}

	if len(ws.Modules) != 2 {
		t.Fatalf("expected 2 modules, got %d", len(ws.Modules))
	}

	if _, ok := ws.Modules["github.com/example/core"]; !ok {
		t.Errorf("missing core module from workspace index")
	}
	if _, ok := ws.Modules["github.com/example/api"]; !ok {
		t.Errorf("missing api module from workspace index")
	}

	expectedPkgs := map[string]string{
		"github.com/example/core":        coreDir,
		"github.com/example/api":         apiDir,
		"github.com/example/api/handler": apiPkgDir,
	}

	for pkgPath, expectedPath := range expectedPkgs {
		dir, exists := ws.PackageRoots[pkgPath]
		if !exists {
			t.Errorf("missing package path %q", pkgPath)
			continue
		}
		if dir != expectedPath {
			t.Errorf("package %q path mismatch: expected %q, got %q", pkgPath, expectedPath, dir)
		}
	}
}

func TestDiscoverWorkspace_IgnoredDirectories(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Create root go.mod
	goModContent := `module github.com/example/ignored

go 1.22
`
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "root.go"), []byte("package ignored\n"), 0644); err != nil {
		t.Fatalf("failed to write root.go: %v", err)
	}

	// 2. Create vendor and .hidden directories containing .go files
	vendorDir := filepath.Join(tempDir, "vendor", "lib")
	hiddenDir := filepath.Join(tempDir, ".internal_cache", "store")
	testdataDir := filepath.Join(tempDir, "testdata", "fixture")

	for _, d := range []string{vendorDir, hiddenDir, testdataDir} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatalf("failed to create dir %s: %v", d, err)
		}
		if err := os.WriteFile(filepath.Join(d, "file.go"), []byte("package lib\n"), 0644); err != nil {
			t.Fatalf("failed to write dummy go file: %v", err)
		}
	}

	ws, err := DiscoverWorkspace(tempDir)
	if err != nil {
		t.Fatalf("DiscoverWorkspace failed: %v", err)
	}

	if len(ws.PackageRoots) != 1 {
		t.Fatalf("expected exactly 1 indexed package, got %d: %+v", len(ws.PackageRoots), ws.PackageRoots)
	}

	if _, ok := ws.PackageRoots["github.com/example/ignored"]; !ok {
		t.Errorf("missing root package from index")
	}
}

func TestDiscoverWorkspace_NoGoModulesError(t *testing.T) {
	tempDir := t.TempDir()

	_, err := DiscoverWorkspace(tempDir)
	if err == nil {
		t.Fatalf("expected error when no go.mod/go.work present, got nil")
	}
}
