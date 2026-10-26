// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package analyzer

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestTwoBuildsExposeDifferentEdges proves Theo Valmis's invariant:
// Two valid builds of the exact same repository (e.g., different build tags or
// different target platforms) expose different call/import edges and produce
// distinct, reproducible BuildContext receipts and graph fingerprints.
func TestTwoBuildsExposeDifferentEdges(t *testing.T) {
	tmp := t.TempDir()

	goMod := "module example.com/multibuild\n\ngo 1.22\n"
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	pkgDir := filepath.Join(tmp, "internal", "payment")
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Shared gateway interface
	commonGo := `package payment

type StoreGateway struct{}
func (s *StoreGateway) RecordSafe() {}

type DirectSQLDriver struct{}
func (d *DirectSQLDriver) ExecRawBypass() {}
`
	// Safe default build (!legacy_bypass)
	safeGo := `//go:build !legacy_bypass

package payment

func ProcessCharge() {
	gw := &StoreGateway{}
	gw.RecordSafe()
}
`
	// Conditional build tag (legacy_bypass) that exposes a forbidden edge to DirectSQLDriver.ExecRawBypass
	bypassGo := `//go:build legacy_bypass

package payment

func ProcessCharge() {
	drv := &DirectSQLDriver{}
	drv.ExecRawBypass()
}
`
	// Windows-only platform file
	windowsGo := `//go:build windows

package payment

func PlatformAuditHook() {
	drv := &DirectSQLDriver{}
	drv.ExecRawBypass()
}
`
	for name, body := range map[string]string{
		"common.go":         commonGo,
		"charge_safe.go":    safeGo,
		"charge_bypass.go":  bypassGo,
		"audit_windows.go":  windowsGo,
	} {
		if err := os.WriteFile(filepath.Join(pkgDir, name), []byte(body), 0644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	ws, err := DiscoverWorkspace(tmp)
	if err != nil {
		t.Fatalf("DiscoverWorkspace: %v", err)
	}

	cgoOff := false
	// Build 1: linux/amd64, default tags
	resDefault, err := AnalyzeWorkspaceWithOptions(context.Background(), ws, WorkspaceAnalysisOptions{
		WorkspaceName: "default",
		BuildTarget: &BuildTargetSpec{
			TargetOS:   "linux",
			TargetArch: "amd64",
			CGOEnabled: &cgoOff,
			BuildTags:  nil,
		},
	})
	if err != nil {
		t.Fatalf("Analyze default: %v", err)
	}

	// Build 2: linux/amd64, -tags=legacy_bypass
	resBypass, err := AnalyzeWorkspaceWithOptions(context.Background(), ws, WorkspaceAnalysisOptions{
		WorkspaceName: "default",
		BuildTarget: &BuildTargetSpec{
			TargetOS:   "linux",
			TargetArch: "amd64",
			CGOEnabled: &cgoOff,
			BuildTags:  []string{"legacy_bypass"},
		},
	})
	if err != nil {
		t.Fatalf("Analyze legacy_bypass: %v", err)
	}

	// Build 3: windows/amd64, default tags
	resWindows, err := AnalyzeWorkspaceWithOptions(context.Background(), ws, WorkspaceAnalysisOptions{
		WorkspaceName: "default",
		BuildTarget: &BuildTargetSpec{
			TargetOS:   "windows",
			TargetArch: "amd64",
			CGOEnabled: &cgoOff,
			BuildTags:  nil,
		},
	})
	if err != nil {
		t.Fatalf("Analyze windows: %v", err)
	}

	hasEdgeTo := func(r *Result, targetSubstr string) bool {
		for _, rel := range r.Relationships {
			if strings.Contains(rel.To, targetSubstr) {
				return true
			}
		}
		return false
	}

	if hasEdgeTo(resDefault, "ExecRawBypass") {
		t.Errorf("default linux build should NOT expose ExecRawBypass call edge")
	}
	if !hasEdgeTo(resBypass, "ExecRawBypass") {
		t.Errorf("legacy_bypass build MUST expose ExecRawBypass call edge")
	}
	if !hasEdgeTo(resWindows, "ExecRawBypass") {
		t.Errorf("windows build MUST expose ExecRawBypass call edge via audit_windows.go")
	}

	if resDefault.Fingerprint == resBypass.Fingerprint {
		t.Errorf("expected distinct graph fingerprints between default and legacy_bypass builds, both got %s", resDefault.Fingerprint)
	}
	if resDefault.BuildContext.ReceiptHash == resBypass.BuildContext.ReceiptHash {
		t.Errorf("expected distinct ReceiptHash between default and legacy_bypass builds")
	}
	if resDefault.BuildContext.SkippedByBuildTagsCount < 2 {
		t.Errorf("expected at least 2 skipped files in linux default build (charge_bypass.go & audit_windows.go), got %d",
			resDefault.BuildContext.SkippedByBuildTagsCount)
	}

	drift := DetectEnvironmentDrift(resBypass.BuildContext, resDefault.BuildContext, "default")
	if !drift.DriftDetected {
		t.Fatalf("expected DetectEnvironmentDrift to flag build tag discrepancy")
	}
	if !strings.Contains(drift.ReproduceCommand, "--tags legacy_bypass") {
		t.Errorf("expected ReproduceCommand to include --tags legacy_bypass, got %q", drift.ReproduceCommand)
	}
}
