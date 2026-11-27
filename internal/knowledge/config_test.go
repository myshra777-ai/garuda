// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package knowledge

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestLoadWorkspaceConfig_Valid(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home directory")
	}
	p := writeFile(t, "workspace.yaml", `workspace: payments
tenant_id: 00000000-0000-0000-0000-000000000001
repositories:
  - path: ~/code/payments
    language: go
documents:
  - path: ~/code/payments/docs/adr
    type: adr
    watch: true
  - path: /abs/policy.md
`)
	cfg, err := LoadWorkspaceConfig(p)
	if err != nil {
		t.Fatalf("LoadWorkspaceConfig: %v", err)
	}
	if cfg.Workspace != "payments" || cfg.TenantID != "00000000-0000-0000-0000-000000000001" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
	if got, want := cfg.Repositories[0].Path, filepath.Join(home, "code/payments"); got != want {
		t.Fatalf("repository path = %q, want %q (~ expanded)", got, want)
	}
	if cfg.Repositories[0].Language != "go" {
		t.Fatalf("language = %q", cfg.Repositories[0].Language)
	}
	if len(cfg.Documents) != 2 || cfg.Documents[0].Type != "adr" || !cfg.Documents[0].Watch {
		t.Fatalf("documents not parsed: %+v", cfg.Documents)
	}
	if cfg.Documents[1].Path != "/abs/policy.md" {
		t.Fatalf("absolute path changed: %q", cfg.Documents[1].Path)
	}
}

func TestLoadWorkspaceConfig_RequiredFields(t *testing.T) {
	cases := map[string]string{
		"workspace": "tenant_id: t1\n",
		"tenant_id": "workspace: w1\n",
	}
	for field, body := range cases {
		_, err := LoadWorkspaceConfig(writeFile(t, "workspace.yaml", body))
		if err == nil || !strings.Contains(err.Error(), field) {
			t.Errorf("missing %s: expected an error naming it, got %v", field, err)
		}
	}
}

func TestLoadWorkspaceConfig_Errors(t *testing.T) {
	if _, err := LoadWorkspaceConfig(filepath.Join(t.TempDir(), "missing.yaml")); err == nil {
		t.Error("expected an error for a missing file")
	}
	if _, err := LoadWorkspaceConfig(writeFile(t, "bad.yaml", "workspace: [unclosed\n")); err == nil {
		t.Error("expected an error for invalid YAML")
	}
}

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home directory")
	}
	cases := map[string]string{
		"~":          home,
		"~/docs/adr": filepath.Join(home, "docs/adr"),
		"/abs/path":  "/abs/path",
		"rel/path":   "rel/path",
		"~other/x":   "~other/x",
		"":           "",
	}
	for in, want := range cases {
		if got := ExpandHome(in); got != want {
			t.Errorf("ExpandHome(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestResolveDocPaths(t *testing.T) {
	root := t.TempDir()
	mk := func(rel string) string {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		return p
	}
	adr := mk("docs/adr/0001.md")
	txt := mk("docs/adr/notes.TXT")
	docx := mk("docs/design/spec.docx")
	mk("docs/adr/diagram.png")            // unsupported extension
	mk("docs/.hidden/secret.md")          // hidden directory
	mk("docs/node_modules/pkg/readme.md") // dependency directory
	mk("docs/vendor/lib/readme.md")       // vendored directory
	single := mk("policy/payments.rst")

	cfg := &WorkspaceConfig{Documents: []DocConfig{
		{Path: filepath.Join(root, "docs")},
		{Path: single},
		{Path: filepath.Join(root, "does-not-exist")}, // skipped, not fatal
	}}
	got, err := cfg.ResolveDocPaths()
	if err != nil {
		t.Fatalf("ResolveDocPaths: %v", err)
	}
	sort.Strings(got)
	want := []string{adr, txt, docx, single}
	sort.Strings(want)
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestEverySupportedDocExtensionCanBeNormalized(t *testing.T) {
	// ResolveDocPaths and NormalizeDocument must agree: anything discovered
	// for ingestion is either normalized or refused with a clear reason.
	for _, ext := range []string{".md", ".markdown", ".txt", ".rst"} {
		if !isSupportedDocExt(ext) {
			t.Fatalf("%s should be discoverable", ext)
		}
		if _, err := NormalizeDocument(writeFile(t, "doc"+ext, "x")); err != nil {
			t.Fatalf("%s is discoverable but NormalizeDocument failed: %v", ext, err)
		}
	}
	if isSupportedDocExt(".png") || isSupportedDocExt("") {
		t.Fatal("unsupported extensions must not be discovered")
	}
}

func TestFindWorkspaceConfig_WalksUp(t *testing.T) {
	root := t.TempDir()
	cfgPath := filepath.Join(root, ".garuda", "workspace.yaml")
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cfgPath, []byte("workspace: w\ntenant_id: t\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	deep := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(deep)

	got, err := FindWorkspaceConfig()
	if err != nil {
		t.Fatalf("FindWorkspaceConfig: %v", err)
	}
	gotReal, _ := filepath.EvalSymlinks(got)
	wantReal, _ := filepath.EvalSymlinks(cfgPath)
	if gotReal != wantReal {
		t.Fatalf("found %q, want %q", got, cfgPath)
	}
}
