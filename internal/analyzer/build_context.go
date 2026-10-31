// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package analyzer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// BuildContext captures the exact compiler, platform, build-tag, dependency-manifest,
// and git working-tree environment used to construct a semantic graph and evaluate policies.
//
// Why this matters (Theo Valmis / Reproducible Policy Gate Invariant):
// Two valid builds of the same repository (e.g., local editor on darwin/arm64 with CGO_ENABLED=1
// vs CI on linux/amd64 with CGO_ENABLED=0 or -tags=integration,purego) can expose different
// AST files, conditional imports, and call edges. Embedding BuildContext into every graph
// snapshot and Merkle-anchored policy receipt guarantees that any reviewer or CI runner can
// deterministically reproduce a verdict when CI and a local editor disagree.
type BuildContext struct {
	Language                string   `json:"language"`                              // "go", "python", "typescript", "polyglot"
	TargetOS                string   `json:"target_os"`                             // GOOS / sys.platform (e.g. "linux", "darwin", "windows")
	TargetArch              string   `json:"target_arch"`                           // GOARCH / machine arch (e.g. "amd64", "arm64")
	TargetPlatform          string   `json:"target_platform"`                       // "<target_os>/<target_arch>"
	CGOEnabled              bool     `json:"cgo_enabled"`                           // Go CGO_ENABLED state
	BuildTags               []string `json:"build_tags"`                            // Sorted build tags / conditional compilation flags
	CompilerVersion         string   `json:"compiler_version,omitempty"`            // e.g. "go1.24", "python3", "typescript-5.x"
	CommitSHA               string   `json:"commit_sha"`                            // Exact 40-char git commit SHA (or "uncommitted")
	CommitShort             string   `json:"commit_short,omitempty"`                // 8-char short commit SHA (with "-dirty" if applicable)
	Branch                  string   `json:"branch,omitempty"`                      // Active git branch name
	WorktreeDirty           bool     `json:"worktree_dirty"`                        // True if uncommitted changes existed during graph build
	DirtyFilesCount         int      `json:"dirty_files_count,omitempty"`           // Count of uncommitted modified/untracked files
	DirtyFilesHash          string   `json:"dirty_files_hash,omitempty"`            // SHA-256 over uncommitted git status & file contents
	ConfigManifestHash      string   `json:"config_manifest_hash,omitempty"`        // SHA-256 of go.mod+go.sum / pyproject.toml / tsconfig.json
	GraphFingerprint        string   `json:"graph_fingerprint,omitempty"`           // Deterministic SHA-256 of AST graph under this build context
	ReceiptHash             string   `json:"receipt_hash,omitempty"`                // Canonical SHA-256 digest of this BuildContext
	ActiveFilesCount        int      `json:"active_files_count,omitempty"`          // Source files included under active build constraints
	SkippedByBuildTagsCount int      `json:"skipped_by_build_tags_count,omitempty"` // Source files excluded by //go:build or OS/Arch suffixes
	SkippedFilesSample      []string `json:"skipped_files_sample,omitempty"`        // Sample of excluded files explaining conditional edge diffs
	ReproduceCommand        string   `json:"reproduce_command"`                     // Exact 1-line CLI command to reproduce graph & verdict
}

// BuildTargetSpec specifies an explicit target platform and tag set for analysis or matrix checks.
type BuildTargetSpec struct {
	TargetOS   string   `json:"target_os"`
	TargetArch string   `json:"target_arch"`
	CGOEnabled *bool    `json:"cgo_enabled,omitempty"`
	BuildTags  []string `json:"build_tags,omitempty"`
}

// EnvironmentDriftReport summarizes differences between two BuildContexts (e.g., CI vs Local Editor).
type EnvironmentDriftReport struct {
	DriftDetected    bool          `json:"drift_detected"`
	Severity         string        `json:"severity"` // "NONE" | "INFO" | "WARNING" | "CRITICAL"
	Summary          string        `json:"summary"`
	Discrepancies    []string      `json:"discrepancies"`
	GraphContext     *BuildContext `json:"graph_context,omitempty"`
	CurrentContext   *BuildContext `json:"current_context,omitempty"`
	ReproduceCommand string        `json:"reproduce_command,omitempty"`
}

// ResolveBuildContext constructs a normalized, deterministic BuildContext for the given
// repository directory, language, and optional BuildTargetSpec overrides.
func ResolveBuildContext(repoDir, language, workspaceName string, spec *BuildTargetSpec) *BuildContext {
	absDir := repoDir
	if abs, err := filepath.Abs(repoDir); err == nil {
		absDir = abs
	}
	if language == "" {
		language = detectRepoLanguage(absDir)
	}
	if workspaceName == "" {
		workspaceName = strings.TrimSpace(os.Getenv("GARUDA_WORKSPACE"))
		if workspaceName == "" {
			workspaceName = "default"
		}
	}

	targetOS := strings.TrimSpace(os.Getenv("GARUDA_TARGET_OS"))
	if targetOS == "" {
		targetOS = strings.TrimSpace(os.Getenv("GOOS"))
	}
	if spec != nil && strings.TrimSpace(spec.TargetOS) != "" {
		targetOS = strings.TrimSpace(spec.TargetOS)
	}
	if targetOS == "" {
		targetOS = runtime.GOOS
	}

	targetArch := strings.TrimSpace(os.Getenv("GARUDA_TARGET_ARCH"))
	if targetArch == "" {
		targetArch = strings.TrimSpace(os.Getenv("GOARCH"))
	}
	if spec != nil && strings.TrimSpace(spec.TargetArch) != "" {
		targetArch = strings.TrimSpace(spec.TargetArch)
	}
	if targetArch == "" {
		targetArch = runtime.GOARCH
	}

	cgoEnabled := build.Default.CgoEnabled
	if envCGO := strings.TrimSpace(os.Getenv("CGO_ENABLED")); envCGO != "" {
		cgoEnabled = envCGO == "1" || strings.EqualFold(envCGO, "true")
	}
	if spec != nil && spec.CGOEnabled != nil {
		cgoEnabled = *spec.CGOEnabled
	}

	var rawTags []string
	if envTags := strings.TrimSpace(os.Getenv("GARUDA_BUILD_TAGS")); envTags != "" {
		rawTags = append(rawTags, splitBuildTags(envTags)...)
	} else if goFlags := strings.TrimSpace(os.Getenv("GOFLAGS")); goFlags != "" {
		rawTags = append(rawTags, extractTagsFromGOFLAGS(goFlags)...)
	}
	if spec != nil && len(spec.BuildTags) > 0 {
		rawTags = append([]string{}, spec.BuildTags...)
	}
	tags := normalizeBuildTags(rawTags)

	commitSHA, commitShort, branch, dirty, dirtyCount, dirtyHash := resolveGitState(absDir)
	configHash := computeConfigManifestHash(absDir, language)

	compilerVer := runtime.Version()
	switch language {
	case "python":
		compilerVer = "python-ast-v3"
		if len(tags) == 0 {
			tags = []string{"sys_platform=" + targetOS}
		}
	case "typescript":
		compilerVer = "typescript-treesitter-v5"
		if tsTags := extractTSConfigTags(absDir); len(tsTags) > 0 && len(tags) == 0 {
			tags = tsTags
		}
	}

	bc := &BuildContext{
		Language:           language,
		TargetOS:           targetOS,
		TargetArch:         targetArch,
		TargetPlatform:     targetOS + "/" + targetArch,
		CGOEnabled:         cgoEnabled,
		BuildTags:          tags,
		CompilerVersion:    compilerVer,
		CommitSHA:          commitSHA,
		CommitShort:        commitShort,
		Branch:             branch,
		WorktreeDirty:      dirty,
		DirtyFilesCount:    dirtyCount,
		DirtyFilesHash:     dirtyHash,
		ConfigManifestHash: configHash,
	}
	bc.ReproduceCommand = bc.FormatReproduceCommand(workspaceName)
	bc.ReceiptHash = bc.ComputeReceiptHash()
	return bc
}

// NewGoBuildContext returns a standard library go/build.Context configured to match
// the exact TargetOS, TargetArch, CGOEnabled, and BuildTags of this BuildContext.
func (bc *BuildContext) NewGoBuildContext() build.Context {
	bctx := build.Default
	if bc == nil {
		return bctx
	}
	if bc.TargetOS != "" {
		bctx.GOOS = bc.TargetOS
	}
	if bc.TargetArch != "" {
		bctx.GOARCH = bc.TargetArch
	}
	bctx.CgoEnabled = bc.CGOEnabled
	if len(bc.BuildTags) > 0 {
		bctx.BuildTags = append([]string{}, bc.BuildTags...)
	} else {
		bctx.BuildTags = nil
	}
	return bctx
}

// ComputeReceiptHash returns a canonical SHA-256 hex digest over the build-context parameters.
func (bc *BuildContext) ComputeReceiptHash() string {
	if bc == nil {
		return ""
	}
	tagsJoined := strings.Join(normalizeBuildTags(bc.BuildTags), ",")
	canonical := fmt.Sprintf(
		"v1|lang=%s|os=%s|arch=%s|cgo=%t|tags=%s|commit=%s|dirty=%t|dirty_hash=%s|cfg=%s|fp=%s",
		bc.Language,
		bc.TargetOS,
		bc.TargetArch,
		bc.CGOEnabled,
		tagsJoined,
		bc.CommitSHA,
		bc.WorktreeDirty,
		bc.DirtyFilesHash,
		bc.ConfigManifestHash,
		bc.GraphFingerprint,
	)
	sum := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(sum[:])
}

// FormatReproduceCommand constructs a deterministic, copy-pasteable 1-line CLI command
// that any engineer or CI runner can run to reproduce the exact graph & policy verdict.
func (bc *BuildContext) FormatReproduceCommand(workspaceName string) string {
	if bc == nil {
		return "garuda check --workspace default"
	}
	if workspaceName == "" {
		workspaceName = "default"
	}
	cgoVal := "0"
	if bc.CGOEnabled {
		cgoVal = "1"
	}
	var parts []string
	if bc.Language == "go" || bc.Language == "" {
		parts = append(parts, fmt.Sprintf("GOOS=%s GOARCH=%s CGO_ENABLED=%s", bc.TargetOS, bc.TargetArch, cgoVal))
	}
	parts = append(parts, "garuda check")
	if bc.TargetOS != "" && bc.TargetArch != "" {
		parts = append(parts, fmt.Sprintf("--target %s/%s", bc.TargetOS, bc.TargetArch))
	}
	if bc.Language == "go" || bc.Language == "" {
		parts = append(parts, fmt.Sprintf("--cgo=%t", bc.CGOEnabled))
	}
	if len(bc.BuildTags) > 0 {
		parts = append(parts, fmt.Sprintf("--tags %s", strings.Join(bc.BuildTags, ",")))
	}
	if bc.CommitShort != "" && bc.CommitShort != "uncommitted" {
		cleanSHA := strings.TrimSuffix(bc.CommitShort, "-dirty")
		parts = append(parts, fmt.Sprintf("--commit %s", cleanSHA))
	}
	if bc.WorktreeDirty {
		parts = append(parts, "--worktree-dirty")
		if bc.DirtyFilesHash != "" {
			dh := bc.DirtyFilesHash
			if len(dh) > 12 {
				dh = dh[:12]
			}
			parts = append(parts, fmt.Sprintf("--dirty-tree-hash %s", dh))
		}
	}
	parts = append(parts, fmt.Sprintf("--workspace %s", workspaceName))
	return strings.Join(parts, " ")
}

// SummaryLine returns a compact human-readable badge string for CLI and commit trailers.
func (bc *BuildContext) SummaryLine() string {
	if bc == nil {
		return "unknown-build-context"
	}
	tagsStr := "none"
	if len(bc.BuildTags) > 0 {
		tagsStr = strings.Join(bc.BuildTags, ",")
	}
	cgoStr := "off"
	if bc.CGOEnabled {
		cgoStr = "on"
	}
	commitDisplay := bc.CommitShort
	if commitDisplay == "" {
		commitDisplay = bc.CommitSHA
		if len(commitDisplay) > 12 {
			commitDisplay = commitDisplay[:12]
		}
	}
	fpShort := bc.GraphFingerprint
	if len(fpShort) > 12 {
		fpShort = fpShort[:12]
	}
	if fpShort == "" {
		fpShort = "n/a"
	}
	return fmt.Sprintf("lang=%s platform=%s/%s cgo=%s tags=[%s] commit=%s graph_fp=%s",
		bc.Language, bc.TargetOS, bc.TargetArch, cgoStr, tagsStr, commitDisplay, fpShort)
}

// DetectEnvironmentDrift compares the BuildContext recorded when the graph/receipt was created
// against the current machine's BuildContext (e.g., CI vs Local Editor) and identifies why
// two runs might expose different edges or verdicts.
func DetectEnvironmentDrift(recorded, current *BuildContext, workspaceName string) *EnvironmentDriftReport {
	report := &EnvironmentDriftReport{
		DriftDetected:  false,
		Severity:       "NONE",
		Summary:        "Build environment matches graph receipt (0 platform, tag, or commit discrepancies).",
		Discrepancies:  []string{},
		GraphContext:   recorded,
		CurrentContext: current,
	}
	if recorded != nil {
		report.ReproduceCommand = recorded.FormatReproduceCommand(workspaceName)
	} else if current != nil {
		report.ReproduceCommand = current.FormatReproduceCommand(workspaceName)
	}
	if recorded == nil || current == nil {
		return report
	}

	if recorded.TargetOS != "" && current.TargetOS != "" && recorded.TargetOS != current.TargetOS {
		report.DriftDetected = true
		report.Discrepancies = append(report.Discrepancies,
			fmt.Sprintf("Target OS mismatch: receipt built for GOOS=%s, current environment is GOOS=%s (OS-specific files like *_%s.go expose different edges).",
				recorded.TargetOS, current.TargetOS, recorded.TargetOS))
	}
	if recorded.TargetArch != "" && current.TargetArch != "" && recorded.TargetArch != current.TargetArch {
		report.DriftDetected = true
		report.Discrepancies = append(report.Discrepancies,
			fmt.Sprintf("Target Arch mismatch: receipt built for GOARCH=%s, current environment is GOARCH=%s.",
				recorded.TargetArch, current.TargetArch))
	}
	if recorded.Language == "go" && recorded.CGOEnabled != current.CGOEnabled {
		report.DriftDetected = true
		report.Discrepancies = append(report.Discrepancies,
			fmt.Sprintf("CGO mismatch: receipt built with CGO_ENABLED=%t, current environment has CGO_ENABLED=%t (//go:build cgo vs !cgo branches differ).",
				recorded.CGOEnabled, current.CGOEnabled))
	}

	recTags := strings.Join(normalizeBuildTags(recorded.BuildTags), ",")
	curTags := strings.Join(normalizeBuildTags(current.BuildTags), ",")
	if recTags != curTags {
		report.DriftDetected = true
		report.Discrepancies = append(report.Discrepancies,
			fmt.Sprintf("Build tags mismatch: receipt used tags=[%s], current environment has tags=[%s].",
				emptyIfNone(recTags), emptyIfNone(curTags)))
	}

	if recorded.CommitSHA != "" && current.CommitSHA != "" &&
		recorded.CommitSHA != "uncommitted" && current.CommitSHA != "uncommitted" &&
		recorded.CommitSHA != current.CommitSHA {
		report.DriftDetected = true
		report.Discrepancies = append(report.Discrepancies,
			fmt.Sprintf("Git Commit SHA mismatch: receipt constructed @ %s, current HEAD is @ %s.",
				shortSHA(recorded.CommitSHA), shortSHA(current.CommitSHA)))
	}

	if recorded.WorktreeDirty != current.WorktreeDirty ||
		(recorded.WorktreeDirty && current.WorktreeDirty && recorded.DirtyFilesHash != "" && current.DirtyFilesHash != "" && recorded.DirtyFilesHash != current.DirtyFilesHash) {
		report.DriftDetected = true
		report.Discrepancies = append(report.Discrepancies,
			fmt.Sprintf("Working-tree state mismatch: receipt dirty=%t (hash=%s), current dirty=%t (hash=%s).",
				recorded.WorktreeDirty, shortSHA(recorded.DirtyFilesHash), current.WorktreeDirty, shortSHA(current.DirtyFilesHash)))
	}

	if recorded.ConfigManifestHash != "" && current.ConfigManifestHash != "" && recorded.ConfigManifestHash != current.ConfigManifestHash {
		report.DriftDetected = true
		report.Discrepancies = append(report.Discrepancies,
			fmt.Sprintf("Dependency/compiler manifest drift: go.mod/pyproject.toml/tsconfig.json hash changed (%s vs %s).",
				shortSHA(recorded.ConfigManifestHash), shortSHA(current.ConfigManifestHash)))
	}

	if report.DriftDetected {
		report.Severity = "WARNING"
		if recorded.TargetOS != current.TargetOS || recTags != curTags || recorded.CGOEnabled != current.CGOEnabled {
			report.Severity = "CRITICAL"
		}
		report.Summary = fmt.Sprintf(
			"Environment drift detected (%d difference(s) between receipt and active environment). Run `%s` to reproduce exact graph and verdict.",
			len(report.Discrepancies), report.ReproduceCommand,
		)
	}

	return report
}

// ParseTargetFlag parses a "--target linux/amd64" flag into (os, arch).
func ParseTargetFlag(target string) (string, string) {
	target = strings.TrimSpace(target)
	if target == "" {
		return "", ""
	}
	parts := strings.SplitN(target, "/", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return strings.TrimSpace(parts[0]), ""
}

// SplitBuildTagsFlag splits a comma- or space-delimited build tags string into sorted unique tags.
func SplitBuildTagsFlag(raw string) []string {
	return normalizeBuildTags(splitBuildTags(raw))
}

func splitBuildTags(raw string) []string {
	var out []string
	for _, f := range strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ' ' || r == ';'
	}) {
		if t := strings.TrimSpace(f); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func normalizeBuildTags(tags []string) []string {
	seen := make(map[string]bool, len(tags))
	var out []string
	for _, t := range tags {
		t = strings.TrimSpace(t)
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}

func extractTagsFromGOFLAGS(goFlags string) []string {
	var tags []string
	parts := strings.Fields(goFlags)
	for i := 0; i < len(parts); i++ {
		p := parts[i]
		if strings.HasPrefix(p, "-tags=") || strings.HasPrefix(p, "--tags=") {
			val := strings.SplitN(p, "=", 2)[1]
			tags = append(tags, splitBuildTags(strings.Trim(val, `"'`))...)
		} else if (p == "-tags" || p == "--tags") && i+1 < len(parts) {
			tags = append(tags, splitBuildTags(strings.Trim(parts[i+1], `"'`))...)
			i++
		}
	}
	return tags
}

func detectRepoLanguage(absDir string) string {
	if _, err := os.Stat(filepath.Join(absDir, "go.mod")); err == nil {
		return "go"
	}
	if IsTypeScriptProject(absDir) {
		return "typescript"
	}
	if IsPythonProject(absDir) {
		return "python"
	}
	return "go"
}

func resolveGitState(absDir string) (commitSHA, commitShort, branch string, dirty bool, dirtyCount int, dirtyHash string) {
	commitSHA = "uncommitted"
	commitShort = "uncommitted"
	branch = "main"

	if out, err := exec.Command("git", "-C", absDir, "rev-parse", "HEAD").Output(); err == nil {
		sha := strings.TrimSpace(string(out))
		if len(sha) >= 7 {
			commitSHA = sha
			if len(sha) >= 12 {
				commitShort = sha[:12]
			} else {
				commitShort = sha
			}
		}
	}

	if out, err := exec.Command("git", "-C", absDir, "rev-parse", "--abbrev-ref", "HEAD").Output(); err == nil {
		if b := strings.TrimSpace(string(out)); b != "" {
			branch = b
		}
	}

	if out, err := exec.Command("git", "-C", absDir, "status", "--porcelain").Output(); err == nil {
		statusText := strings.TrimSpace(string(out))
		if statusText != "" {
			dirty = true
			lines := strings.Split(statusText, "\n")
			dirtyCount = len(lines)
			h := sha256.New()
			h.Write([]byte(statusText))
			for _, line := range lines {
				if len(line) > 3 {
					relPath := strings.TrimSpace(line[3:])
					if idx := strings.Index(relPath, " -> "); idx >= 0 {
						relPath = strings.TrimSpace(relPath[idx+4:])
					}
					if content, rErr := os.ReadFile(filepath.Join(absDir, relPath)); rErr == nil {
						fh := sha256.Sum256(content)
						h.Write(fh[:])
					}
				}
			}
			dirtyHash = hex.EncodeToString(h.Sum(nil))[:16]
			if commitShort != "uncommitted" && !strings.HasSuffix(commitShort, "-dirty") {
				commitShort = commitShort + "-dirty"
			}
		}
	}
	return
}

func computeConfigManifestHash(absDir, language string) string {
	var candidates []string
	switch language {
	case "python":
		candidates = []string{"pyproject.toml", "requirements.txt", "setup.py", "Pipfile.lock"}
	case "typescript":
		candidates = []string{"tsconfig.json", "package.json", "package-lock.json", "pnpm-lock.yaml"}
	default:
		candidates = []string{"go.mod", "go.sum", ".garudarules"}
	}
	h := sha256.New()
	matched := 0
	for _, name := range candidates {
		if data, err := os.ReadFile(filepath.Join(absDir, name)); err == nil {
			h.Write([]byte(name + ":"))
			h.Write(data)
			matched++
		}
	}
	if matched == 0 {
		return ""
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func extractTSConfigTags(absDir string) []string {
	data, err := os.ReadFile(filepath.Join(absDir, "tsconfig.json"))
	if err != nil {
		return nil
	}
	text := string(data)
	var tags []string
	for _, marker := range []string{"ES2022", "ES2023", "ESNext", "NodeNext", "bundler", "commonjs", "strict"} {
		if strings.Contains(text, marker) {
			tags = append(tags, "ts:"+marker)
		}
	}
	return normalizeBuildTags(tags)
}

func emptyIfNone(s string) string {
	if s == "" {
		return "none"
	}
	return s
}

func shortSHA(s string) string {
	if len(s) > 12 {
		return s[:12]
	}
	if s == "" {
		return "n/a"
	}
	return s
}
