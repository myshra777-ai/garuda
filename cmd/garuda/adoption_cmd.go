// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/myshra777-ai/garuda/internal/analyzer"
	"github.com/spf13/cobra"
)

// StarterTemplate defines a pre-built architectural or financial policy template.
type StarterTemplate struct {
	Key         string
	Filename    string
	Title       string
	Category    string
	Description string
	YAMLContent string
}

var starterTemplates = []StarterTemplate{
	{
		Key:         "hexagonal",
		Filename:    "hexagonal.yaml",
		Title:       "Hexagonal / Ports & Adapters Isolation",
		Category:    "Architecture",
		Description: "Domain core & use-cases must never import infrastructure adapters (SQL, HTTP, Redis, gRPC) directly.",
		YAMLContent: `# Garuda Starter Policy Template: Hexagonal / Ports & Adapters
id: tpl-hexagonal-ports-adapters
version: v1
title: "Hexagonal Architecture: Domain core must not import infrastructure adapters"
description: |
  Enforces strict Ports & Adapters isolation. Domain entities and use-case ports
  must never depend directly on database drivers, HTTP routers, or external SDKs.
priority: 200
language: any
authority: "architecture-board@company.com"
scope:
  domain: core
  pattern: ".*domain.*|.*core.*|.*usecase.*"
when:
  - type: claim_exists
    params:
      claim_type: IMPORTS
      from_name_pattern: ".*(domain|core|usecase).*"
      to_name_pattern: ".*(infra|adapter|postgres|mysql|redis|http|grpc|database/sql).*"
then:
  decision: BLOCK
  reason: |
    Domain/Core layer imported an infrastructure adapter directly. Define an outbound Port
    interface in the domain package and inject the adapter implementation from cmd/ or internal/wiring.
`,
	},
	{
		Key:         "layered",
		Filename:    "layered.yaml",
		Title:       "Strict Layered Architecture (No Upward Dependencies)",
		Category:    "Architecture",
		Description: "Persistence/store layers must never import service or API/transport packages.",
		YAMLContent: `# Garuda Starter Policy Template: Strict Layered Architecture
id: tpl-layered-no-upward-deps
version: v1
title: "Layered Architecture: Lower layers must never import upper layers"
description: |
  Prevents upward dependency inversions (store/data importing service/api, or
  service importing api/transport).
priority: 190
language: any
authority: "architecture-board@company.com"
scope:
  domain: architecture
  pattern: ".*store.*|.*repo.*|.*data.*"
when:
  - type: claim_exists
    params:
      claim_type: IMPORTS
      from_name_pattern: ".*(store|repository|repo|dal).*"
      to_name_pattern: ".*(api|handler|controller|transport|ui|payment).*"
then:
  decision: BLOCK
  reason: |
    Upward dependency violation: persistence/store layer imported service/transport symbols.
    Move shared DTOs or interfaces into a domain/types package.
`,
	},
	{
		Key:         "microservices",
		Filename:    "microservices.yaml",
		Title:       "Microservices Boundary (No Cross-Service DB Access)",
		Category:    "Architecture",
		Description: "Prevents one service from importing or querying another service's internal persistence store.",
		YAMLContent: `# Garuda Starter Policy Template: Microservices Database Boundary
id: tpl-microservices-db-isolation
version: v1
title: "Microservices Boundary: No cross-service internal store or table access"
description: |
  Each microservice owns its state schema exclusively. Cross-service communication
  must traverse versioned API/gRPC/event contracts, never another service's store package.
priority: 210
language: any
authority: "platform-governance@company.com"
scope:
  domain: microservices
  pattern: ".*service.*"
when:
  - type: claim_exists
    params:
      claim_type: CALLS
      from_name_pattern: ".*(order|catalog|shipping|notification).*"
      to_name_pattern: ".*(payment|ledger|billing).*Store.*"
then:
  decision: BLOCK
  reason: |
    Cross-service database boundary violation detected. Access foreign service state
    via its public API/gRPC client rather than calling its internal Store directly.
`,
	},
	{
		Key:         "no-circular-imports",
		Filename:    "no-circular-imports.yaml",
		Title:       "Zero Circular Package Imports",
		Category:    "Hygiene",
		Description: "Blocks cyclic dependencies between internal packages to keep the dependency DAG acyclic.",
		YAMLContent: `# Garuda Starter Policy Template: Zero Circular Imports
id: tpl-no-circular-imports
version: v1
title: "Package Hygiene: Zero circular package or module dependencies"
description: |
  Detects and blocks cyclic package dependencies that degrade build determinism
  and amplify change blast radius.
priority: 180
language: any
authority: "platform-governance@company.com"
scope:
  domain: topology
  pattern: ".*"
when:
  - type: claim_exists
    params:
      claim_type: IMPORTS
      from_name_pattern: ".*internal/store.*"
      to_name_pattern: ".*internal/policy.*"
then:
  decision: BLOCK
  reason: |
    Circular dependency hazard detected between internal packages. Extract shared
    value types into internal/types to keep the dependency DAG acyclic.
`,
	},
	{
		Key:         "financial-no-direct-sql",
		Filename:    "financial-no-direct-sql.yaml",
		Title:       "Financial: Payment Code Must Not Import database/sql Directly",
		Category:    "Financial / FinOps",
		Description: "Forces payment, refund, and settlement paths through an audited Repository interface.",
		YAMLContent: `# Garuda Starter Policy Template: Financial SQL Isolation
id: tpl-financial-no-direct-sql
version: v1
title: "Financial Guardrail: Payment & billing code must not import database/sql directly"
description: |
  All payment, refund, settlement, and ledger mutations must route through an
  audited, idempotency-enforced Repository interface rather than raw SQL handles.
priority: 250
language: go
authority: "finops-security@company.com"
scope:
  domain: payments
  pattern: ".*Payment.*|.*Refund.*|.*Charge.*|.*Settlement.*|.*Payout.*"
when:
  - type: claim_exists
    params:
      claim_type: IMPORTS
      from_name_pattern: ".*(Payment|Refund|Charge|Settlement|Payout|payment|refund|billing).*"
      to_name_pattern: "^(database/sql|github.com/jackc/pgx.*)$"
then:
  decision: BLOCK
  reason: |
    Direct SQL driver import in financial/payment path violates ADR-0001. Route all
    monetary mutations through the audited PaymentRepository / LedgerStore interface.
`,
	},
	{
		Key:         "financial-decimal-precision",
		Filename:    "financial-decimal-precision.yaml",
		Title:       "Financial: Monetary Calculations Must Use Decimal (Never float64)",
		Category:    "Financial / FinOps",
		Description: "Blocks IEEE-754 float64 arithmetic in currency, fee, tax, and loss-ratio calculations.",
		YAMLContent: `# Garuda Starter Policy Template: Financial Decimal Precision
id: tpl-financial-decimal-precision
version: v1
title: "Financial Guardrail: Monetary & fee calculations must use Decimal or integer minor units, never float64"
description: |
  IEEE-754 floating-point arithmetic introduces rounding drift in currency, fee,
  tax, and loss-ratio calculations. Requires fixed-point decimal or int64 minor units (cents/paise).
priority: 260
language: any
authority: "finops-security@company.com"
scope:
  domain: finance
  pattern: ".*Amount.*|.*Price.*|.*Fee.*|.*Tax.*|.*Settlement.*|.*LossRatio.*|.*Claim.*"
when:
  - type: claim_exists
    params:
      claim_type: CALLS
      from_name_pattern: ".*(Calculate|Compute|Settle|Charge|Refund).*"
      to_name_pattern: ".*Float64.*"
then:
  decision: BLOCK
  reason: |
    Floating-point (float64) arithmetic detected in financial calculation path.
    Use shopspring/decimal, Python decimal.Decimal, or int64 minor units (cents/paise).
`,
	},
	{
		Key:         "zero-trust-auth",
		Filename:    "zero-trust-auth.yaml",
		Title:       "Zero-Trust API Tenant & Workspace Scope Guard",
		Category:    "Security",
		Description: "Requires exported API handlers to verify tenant_id and workspace_id scope before store access.",
		YAMLContent: `# Garuda Starter Policy Template: Zero-Trust Auth & Tenant Scope
id: tpl-zero-trust-tenant-scope
version: v1
title: "Zero-Trust API Guardrail: Exported handlers must enforce tenant & workspace scope"
description: |
  Every multi-tenant API handler must extract and validate tenant_id and workspace_id
  before querying or mutating persistence state (Invariant I-11).
priority: 240
language: any
authority: "security-architecture@company.com"
scope:
  domain: api
  pattern: "^Handle.*"
when:
  - type: claim_exists
    params:
      claim_type: CALLS
      from_name_pattern: "^Handle.*"
      to_name_pattern: ".*DirectUnscopedQuery.*"
then:
  decision: BLOCK
  reason: |
    API handler bypassed tenant/workspace scope enforcement (Invariant I-11).
    Call workspaceScopeFromRequest(r) before invoking store methods.
`,
	},
	{
		Key:         "ai-agent-safety",
		Filename:    "ai-agent-safety.yaml",
		Title:       "Autonomous Coding Agent Safety Gate (Talon Fail-Closed)",
		Category:    "AI Governance",
		Description: "Freezes SEVERE criticality Merkle, auth, and security surfaces against unapproved AI edits.",
		YAMLContent: `# Garuda Starter Policy Template: Autonomous AI Agent Safety Gate
id: tpl-ai-agent-safety-gate
version: v1
title: "AI Agent Harness: Autonomous agents must not modify cryptographic ledger or auth signing keys"
description: |
  Freezes mission-critical cryptographic, Merkle sealing, and JWT verification
  surfaces against autonomous agent modification without explicit human CLI co-sign.
priority: 300
language: any
authority: "staff-security@company.com"
scope:
  domain: security
  pattern: ".*AppendLeafAndSealTx.*|.*VerifyInclusionProof.*|.*JWTConfig.*"
when:
  - type: claim_exists
    params:
      claim_type: CALLS
      from_name_pattern: ".*AgentAutoPatch.*"
      to_name_pattern: ".*(AppendLeafAndSealTx|VerifyInclusionProof|JWTConfig).*"
then:
  decision: BLOCK
  reason: |
    Autonomous agent attempted to modify a frozen SEVERE criticality security/Merkle surface.
    Human architect CLI co-sign (GARUDA_MCP_APPROVER) is required.
`,
	},
	{
		Key:         "pii-gdpr-hipaa",
		Filename:    "pii-gdpr-hipaa.yaml",
		Title:       "Regulated PII / PHI Data Lineage Guard (GDPR / HIPAA / DPDP)",
		Category:    "Privacy",
		Description: "Prevents customer PII/PHI structs from flowing into raw stdout/file log sinks.",
		YAMLContent: `# Garuda Starter Policy Template: Regulated PII / PHI Data Lineage
id: tpl-pii-gdpr-hipaa-sink-guard
version: v1
title: "Privacy Guardrail: Customer PII/PHI records must never flow into raw log sinks"
description: |
  Prevents accidental leakage of SSN, PAN, Aadhaar, card numbers, or patient health
  fields into unredacted stdout/file loggers.
priority: 230
language: any
authority: "privacy-compliance@company.com"
scope:
  domain: privacy
  pattern: ".*Customer.*|.*Patient.*|.*Identity.*|.*KYC.*"
when:
  - type: claim_exists
    params:
      claim_type: CALLS
      from_name_pattern: ".*(KYC|Customer|Patient|Identity).*"
      to_name_pattern: "^(fmt\\.Print|log\\.Print).*RawPII.*"
then:
  decision: BLOCK
  reason: |
    Unredacted PII/PHI struct passed to raw logger. Route through internal/redact.SanitizePII
    before emitting telemetry or logs.
`,
	},
	{
		Key:         "idempotent-mutations",
		Filename:    "idempotent-mutations.yaml",
		Title:       "Idempotent Payment & State Mutations",
		Category:    "Reliability",
		Description: "Requires external payment, refund, and transfer mutations to verify an idempotency key.",
		YAMLContent: `# Garuda Starter Policy Template: Idempotent State Mutations
id: tpl-idempotent-state-mutations
version: v1
title: "Reliability Guardrail: External payment & state mutations must enforce idempotency keys"
description: |
  Prevents duplicate charges or double-ledger entries during network retries or
  multi-agent handoff replays (ADR-0001).
priority: 220
language: any
authority: "sre-platform@company.com"
scope:
  pattern: ".*(ExecuteCharge|ExecuteRefund|TransferFunds|ProcessInstantCheckout|RecordChargeTx|AppendLeafAndSealTx|SaveTask|ExecuteHandoff).*"
when:
  - type: claim_exists
    params:
      claim_type: CALLS
      from_name_pattern: ".*(ExecuteCharge|ExecuteRefund|TransferFunds).*"
      to_name_pattern: ".*DirectBankAPI.*"
then:
  decision: REVIEW
  reason: |
    External state mutation must verify an idempotency key via IdempotencyStore
    before invoking downstream banking/payment gateways.
`,
	},
}

// InstallStarterTemplates writes the requested template(s) into targetDir/policies.
// Accepts template keys like "hexagonal", "layered", "microservices", "financial", "all", or comma-separated lists.
func InstallStarterTemplates(targetDir, templateSpec string) ([]string, error) {
	if strings.TrimSpace(templateSpec) == "" {
		return nil, nil
	}
	policiesDir := filepath.Join(targetDir, "policies")
	templatesSubDir := filepath.Join(policiesDir, "templates")
	_ = os.MkdirAll(policiesDir, 0755)
	_ = os.MkdirAll(templatesSubDir, 0755)

	// Always seed all 10 reference templates into policies/templates/
	for _, tpl := range starterTemplates {
		_ = os.WriteFile(filepath.Join(templatesSubDir, tpl.Filename), []byte(tpl.YAMLContent), 0644)
	}

	requested := strings.Split(strings.ToLower(strings.TrimSpace(templateSpec)), ",")
	selectedKeys := make(map[string]bool)
	for _, r := range requested {
		r = strings.TrimSpace(r)
		switch r {
		case "all", "enterprise", "all-enterprise":
			for _, tpl := range starterTemplates {
				selectedKeys[tpl.Key] = true
			}
		case "financial", "fintech", "financial-regulated":
			selectedKeys["financial-no-direct-sql"] = true
			selectedKeys["financial-decimal-precision"] = true
			selectedKeys["idempotent-mutations"] = true
			selectedKeys["hexagonal"] = true
		case "circular", "no-circular":
			selectedKeys["no-circular-imports"] = true
		default:
			selectedKeys[r] = true
		}
	}

	var installed []string
	for _, tpl := range starterTemplates {
		if selectedKeys[tpl.Key] {
			dest := filepath.Join(policiesDir, tpl.Filename)
			if err := os.WriteFile(dest, []byte(tpl.YAMLContent), 0644); err != nil {
				return installed, err
			}
			installed = append(installed, dest)
		}
	}
	if len(installed) == 0 {
		return nil, fmt.Errorf("unknown template %q (valid: hexagonal, layered, microservices, no-circular-imports, financial, financial-no-direct-sql, financial-decimal-precision, zero-trust-auth, ai-agent-safety, pii-gdpr-hipaa, idempotent-mutations, all)", templateSpec)
	}
	return installed, nil
}

// ============================================================================
// 1. garuda templates (List & Install Pre-Built Policy Templates — Item 2.2)
// ============================================================================

var templatesCmd = &cobra.Command{
	Use:   "templates [list|install <name>]",
	Short: "Browse and install Garuda's 10 pre-built architectural & financial policy templates",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) >= 2 && args[0] == "install" {
			installed, err := InstallStarterTemplates(".", args[1])
			if err != nil {
				return err
			}
			fmt.Printf("🦅 Installed %d starter policy template(s) into ./policies/:\n", len(installed))
			for _, p := range installed {
				fmt.Printf("  ✓ %s\n", p)
			}
			return nil
		}
		fmt.Println("╔══════════════════════════════════════════════════════════════════════════════════╗")
		fmt.Println("║              GARUDA STARTER POLICY TEMPLATES (10 PRE-BUILT RULES)                ║")
		fmt.Println("╚══════════════════════════════════════════════════════════════════════════════════╝")
		for i, tpl := range starterTemplates {
			fmt.Printf("  %02d. %-29s [%s]\n", i+1, tpl.Key, tpl.Category)
			fmt.Printf("      %s\n", tpl.Description)
		}
		fmt.Println("\n💡 Quick Install:")
		fmt.Println("   garuda init --template hexagonal")
		fmt.Println("   garuda init --template financial        # Installs SQL isolation + Decimal precision + Idempotency")
		fmt.Println("   garuda init --template all              # Installs all 10 enterprise & regulated guardrails")
		fmt.Println("   garuda templates install hexagonal")
		return nil
	},
}

// ============================================================================
// 2. Interactive 5-Question Architecture Wizard (garuda init --wizard / garuda wizard — Item 2.5)
// ============================================================================

var (
	wizardDefaultsFlag bool
	wizardTargetDir    string
)

var archWizardCmd = &cobra.Command{
	Use:   "wizard [target_dir]",
	Short: "Interactive 5-question architecture & agent governance onboarding wizard",
	RunE: func(cmd *cobra.Command, args []string) error {
		target := wizardTargetDir
		if len(args) > 0 && args[0] != "" {
			target = args[0]
		}
		return runInteractiveArchitectureWizard(target, wizardDefaultsFlag)
	},
}

func detectRepoLanguagesAndCriticalPaths(dir string) (string, []string) {
	langCounts := map[string]int{}
	critDirs := map[string]bool{}
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			base := strings.ToLower(d.Name())
			if base == ".git" || base == "node_modules" || base == "vendor" || base == ".garuda" {
				return filepath.SkipDir
			}
			for _, kw := range []string{"payment", "billing", "refund", "settlement", "ledger", "auth", "security", "store", "api", "talon"} {
				if strings.Contains(base, kw) {
					rel, rErr := filepath.Rel(dir, path)
					if rErr == nil && len(strings.Split(rel, string(os.PathSeparator))) <= 3 {
						critDirs[rel] = true
					}
				}
			}
			return nil
		}
		switch strings.ToLower(filepath.Ext(d.Name())) {
		case ".go":
			langCounts["Go"]++
		case ".py":
			langCounts["Python"]++
		case ".ts", ".tsx", ".js":
			langCounts["TypeScript/Node"]++
		case ".java", ".kt":
			langCounts["Java/JVM"]++
		case ".rs":
			langCounts["Rust"]++
		}
		return nil
	})
	detectedLang := "Go"
	maxC := 0
	for l, c := range langCounts {
		if c > maxC {
			maxC = c
			detectedLang = l
		}
	}
	if len(langCounts) > 1 {
		var parts []string
		for l := range langCounts {
			parts = append(parts, l)
		}
		sort.Strings(parts)
		detectedLang = "Polyglot (" + strings.Join(parts, " + ") + ")"
	}
	var paths []string
	for p := range critDirs {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	if len(paths) == 0 {
		paths = []string{"internal/payment", "internal/auth", "internal/store", "internal/api"}
	} else if len(paths) > 5 {
		paths = paths[:5]
	}
	return detectedLang, paths
}

func promptLine(reader *bufio.Reader, question, defaultVal string, nonInteractive bool) string {
	fmt.Printf("%s [%s]: ", question, defaultVal)
	if nonInteractive {
		fmt.Println(defaultVal)
		return defaultVal
	}
	line, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println(defaultVal)
		return defaultVal
	}
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return defaultVal
	}
	return trimmed
}

func runInteractiveArchitectureWizard(targetDir string, useDefaults bool) error {
	absDir, err := filepath.Abs(targetDir)
	if err != nil {
		absDir = targetDir
	}
	detectedLang, detectedPaths := detectRepoLanguagesAndCriticalPaths(absDir)
	defaultPathsStr := strings.Join(detectedPaths, ", ")

	// Auto-detect non-interactive stdin if piped/closed
	stat, stErr := os.Stdin.Stat()
	if stErr != nil || (stat.Mode()&os.ModeCharDevice) == 0 {
		useDefaults = true
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║        🦅 GARUDA 5-QUESTION ARCHITECTURE & AGENT GOVERNANCE WIZARD               ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════════════╝")
	fmt.Printf("  Target Repository: %s\n\n", absDir)

	q1 := promptLine(reader, "  [1/5] What language/framework? (Go / Python / TypeScript / Java / Polyglot)", detectedLang, useDefaults)
	q2 := promptLine(reader, "  [2/5] What is your architecture style? (hexagonal / layered / microservices / financial / all)", "financial", useDefaults)
	q3 := promptLine(reader, "  [3/5] Which code paths are SEVERE/HIGH criticality? (auto-detected)", defaultPathsStr, useDefaults)
	q4 := promptLine(reader, "  [4/5] Which autonomous coding agents do you run? (claude-code / cursor / codex / all)", "all", useDefaults)
	q5 := promptLine(reader, "  [5/5] Where should violations be reported? (cli / github-pr / slack / teams / all)", "cli,github-pr", useDefaults)

	fmt.Println("\n🔧 Provisioning Garuda Configuration, Starter Templates & Auto-Synthesized Draft Rulebook...")

	installed, tplErr := InstallStarterTemplates(absDir, q2)
	if tplErr != nil {
		installed, _ = InstallStarterTemplates(absDir, "hexagonal")
	}

	garudaDir := filepath.Join(absDir, ".garuda")
	_ = os.MkdirAll(garudaDir, 0755)

	configYAML := fmt.Sprintf(`# Generated by Garuda 5-Question Interactive Wizard
version: v2.0
workspace: %s
language: %q
architecture_style: %q
critical_paths:
  severe: [%s]
agents:
  enabled: %q
  fail_closed: true
  contract_path: ".garuda/TALON_CONTRACT.md"
reporting:
  sinks: %q
  sarif_output: "garuda.sarif"
  pr_comment_output: "garuda-pr-report.md"
`, filepath.Base(absDir), q1, q2, q3, q4, q5)

	cfgPath := filepath.Join(garudaDir, "config.yaml")
	_ = os.WriteFile(cfgPath, []byte(configYAML), 0644)

	draftRulebook := fmt.Sprintf(`# GARUDA AUTO-SYNTHESIZED BASE RULEBOOK (DRAFT)
- **Workspace**: `+"`%s`"+`
- **Detected Stack**: `+"`%s`"+`
- **Architecture Style**: `+"`%s`"+`
- **Criticality Frozen Paths (`+"`SEVERE / HIGH`"+`)**: `+"`%s`"+`
- **Active Agent Harnesses**: `+"`%s`"+`
- **Status**: `+"`DRAFT · AWAITING HUMAN APPROVAL`"+`

## 1. Reverse-Engineered Layer & Boundary Rules
1. **[BLOCK] `+"`tpl-hexagonal-ports-adapters`"+`**: Domain/Core packages MUST NOT import infrastructure adapters (`+"`database/sql`"+`, `+"`net/http`"+`, `+"`redis`"+`).
2. **[BLOCK] `+"`tpl-layered-no-upward-deps`"+`**: Persistence (`+"`store`"+` / `+"`repository`"+`) MUST NOT import service or transport (`+"`api`"+` / `+"`handler`"+`) packages.
3. **[BLOCK] `+"`tpl-financial-no-direct-sql`"+`**: Financial & payment paths (`+"`%s`"+`) MUST route through audited repository interfaces.
4. **[BLOCK] `+"`tpl-financial-decimal-precision`"+`**: Monetary, fee, tax, and loss-ratio calculations MUST use fixed-point Decimal or integer minor units (`+"`int64`"+`), never `+"`float64`"+`.

## 2. Human-in-the-Loop Approval Workflow
- Review visually in Dashboard (**Surface 04 — Base Rulebook Studio**) or approve from CLI:
  `+"```bash\n  garuda rulebook generate --approve --approver \"Staff Architect\"\n  ```"+`
`, filepath.Base(absDir), q1, q2, q3, q4, q3)

	draftPath := filepath.Join(garudaDir, "DRAFT_BASE_RULEBOOK.md")
	_ = os.WriteFile(draftPath, []byte(draftRulebook), 0644)

	fmt.Printf("  ✓ Wrote wizard config:        %s\n", cfgPath)
	fmt.Printf("  ✓ Installed %d policy files:   %s\n", len(installed), strings.Join(installed, ", "))
	fmt.Printf("  ✓ Synthesized draft rulebook: %s\n", draftPath)
	fmt.Println("\n🚀 Next Steps (First 3 Rules in 10 Minutes):")
	fmt.Println("   1. Review draft rulebook:   cat .garuda/DRAFT_BASE_RULEBOOK.md")
	fmt.Println("   2. Seal & approve rulebook: garuda rulebook generate --approve")
	fmt.Println("   3. Run diff-scoped gate:    garuda check --policy policies")
	return nil
}

// ============================================================================
// 3. garuda demo — 200-Line Sample Codebase with 3 Intentional Violations (Item 2.9)
// ============================================================================

var (
	demoDirFlag  string
	demoKeepFlag bool
	demoJSONFlag bool
)

var demoCmd = &cobra.Command{
	Use:   "demo",
	Short: "Spin up a 200-line sample fintech codebase with 3 intentional violations and run the full Garuda pipeline",
	Long: `Creates a self-contained ~200-line fintech/payment microservice (acme-pay) with 3 intentional
architectural & financial violations:
  1. [CRITICAL] Direct database/sql import in Payment Handler (violates tpl-financial-no-direct-sql)
  2. [CRITICAL] float64 currency math + 10x Crore/Lakh unit drift (violates tpl-financial-decimal-precision)
  3. [HIGH]     Upward layer dependency from internal/store -> internal/payment (violates tpl-layered-no-upward-deps)
Then runs AST extraction, Auto-Rulebook synthesis, Policy & Financial Audit gate, Talon Self-Repair, and Merkle sealing.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()
		demoDir := demoDirFlag
		if demoDir == "" {
			demoDir = filepath.Join(os.TempDir(), "garuda-demo-acme-pay")
		}
		_ = os.RemoveAll(demoDir)
		if err := os.MkdirAll(demoDir, 0755); err != nil {
			return err
		}
		if !demoKeepFlag {
			defer os.RemoveAll(demoDir)
		}

		// Write 200-line realistic sample codebase across 4 files
		files := map[string]string{
			"go.mod": "module github.com/acme-corp/acme-pay\n\ngo 1.24\n",
			"internal/domain/payment.go": `package domain

import "context"

// MoneyMinorUnits represents currency in exact integer paise/cents (ADR-0001).
type MoneyMinorUnits int64

type ChargeRequest struct {
	IdempotencyKey string
	CustomerID     string
	AmountPaise    MoneyMinorUnits
	Currency       string
}

type ChargeReceipt struct {
	TransactionID string
	SettledPaise  MoneyMinorUnits
	Status        string
}

// PaymentRepository is the outbound hexagonal port for audited ledger persistence.
type PaymentRepository interface {
	RecordChargeTx(ctx context.Context, req ChargeRequest) (*ChargeReceipt, error)
}
`,
			"internal/payment/checkout.go": `package payment

import (
	"context"
	"database/sql" // ❌ VIOLATION 1 [CRITICAL]: Payment handler imports database/sql directly!
	"fmt"

	"github.com/acme-corp/acme-pay/internal/domain"
)

type CheckoutHandler struct {
	rawDB *sql.DB
	repo  domain.PaymentRepository
}

func NewCheckoutHandler(db *sql.DB, repo domain.PaymentRepository) *CheckoutHandler {
	return &CheckoutHandler{rawDB: db, repo: repo}
}

// ProcessInstantCheckout bypasses PaymentRepository and executes raw SQL without idempotency lock.
func (h *CheckoutHandler) ProcessInstantCheckout(ctx context.Context, req domain.ChargeRequest) (*domain.ChargeReceipt, error) {
	if h.rawDB != nil {
		_, err := h.rawDB.ExecContext(ctx,
			"INSERT INTO raw_payments (customer_id, amount_paise) VALUES ($1, $2)",
			req.CustomerID, req.AmountPaise,
		)
		if err != nil {
			return nil, fmt.Errorf("raw sql payment insert failed: %w", err)
		}
	}
	feeFloat := CalculateSettlementFeeFloat64(float64(req.AmountPaise) / 100.0)
	settled := req.AmountPaise - domain.MoneyMinorUnits(feeFloat*100.0)
	return &domain.ChargeReceipt{
		TransactionID: "tx-" + req.IdempotencyKey,
		SettledPaise:  settled,
		Status:        "SETTLED_UNVERIFIED",
	}, nil
}
`,
			"internal/payment/settlement.go": `package payment

import "math"

// ❌ VIOLATION 2 [CRITICAL]: Uses IEEE-754 float64 for monetary fee & loss-ratio calculation!
// Also introduces a 10x Crore vs Lakh unit-scaling discrepancy between Ops and Finance pipelines.
func CalculateSettlementFeeFloat64(amountRupees float64) float64 {
	baseRate := 0.029
	gstRate := 0.18
	rawFee := amountRupees * baseRate * (1.0 + gstRate)
	return math.Round(rawFee*10000.0) / 10000.0
}

// ComputeOpsClaimsPaidCrore divides raw INR by 1e6 (10 Lakh) instead of 1e7 (1 Crore),
// causing Ops to report 48.0 Crore while Finance reports 4.8 Crore (10x unit mismatch).
func ComputeOpsClaimsPaidCrore(totalClaimsINR float64) float64 {
	return totalClaimsINR / 1_000_000.0 // BUG: 1e6 is 10 Lakh, not 1 Crore (1e7)!
}

func ComputeFinanceClaimsPaidCrore(totalClaimsINR float64) float64 {
	return totalClaimsINR / 10_000_000.0 // Correct: 1 Crore = 1e7 INR
}

// AuditContext DTO improperly referenced by lower persistence layer.
type SettlementAuditDTO struct {
	BatchID     string
	TotalCrores float64
}
`,
			"internal/store/ledger_repo.go": `package store

import (
	"context"
	"fmt"

	"github.com/acme-corp/acme-pay/internal/domain"
	"github.com/acme-corp/acme-pay/internal/payment" // ❌ VIOLATION 3 [HIGH]: Upward layer import (store -> payment)!
)

type PostgresLedgerRepo struct {
	dsn string
}

func NewPostgresLedgerRepo(dsn string) *PostgresLedgerRepo {
	return &PostgresLedgerRepo{dsn: dsn}
}

func (r *PostgresLedgerRepo) RecordChargeTx(ctx context.Context, req domain.ChargeRequest) (*domain.ChargeReceipt, error) {
	// Upward dependency inversion: store layer calls service/payment package directly
	dto := payment.SettlementAuditDTO{
		BatchID:     "batch-01",
		TotalCrores: payment.ComputeOpsClaimsPaidCrore(float64(req.AmountPaise) / 100.0),
	}
	if dto.TotalCrores < 0 {
		return nil, fmt.Errorf("invalid settlement batch")
	}
	return &domain.ChargeReceipt{
		TransactionID: "ledger-" + req.IdempotencyKey,
		SettledPaise:  req.AmountPaise,
		Status:        "RECORDED",
	}, nil
}
`,
		}

		totalLines := 0
		for rel, content := range files {
			full := filepath.Join(demoDir, rel)
			_ = os.MkdirAll(filepath.Dir(full), 0755)
			_ = os.WriteFile(full, []byte(content), 0644)
			totalLines += len(strings.Split(content, "\n"))
		}

		_, _ = InstallStarterTemplates(demoDir, "financial,layered")

		// Run the real Garuda Go analyzer on the generated sample. It needs no
		// database. If it cannot run (for example no Go toolchain is installed),
		// the demo says so instead of printing invented numbers.
		analysisStart := time.Now()
		analysisRes, analysisErr := analyzer.AnalyzeDirectory(context.Background(), demoDir)
		analysisMS := time.Since(analysisStart).Milliseconds()
		analysisEntities, analysisEdges := 0, 0
		if analysisErr == nil && analysisRes != nil {
			analysisEntities = len(analysisRes.Entities)
			analysisEdges = len(analysisRes.Relationships)
		}

		demoReport := map[string]interface{}{
			"demo_project":     "acme-pay (sample fintech microservice)",
			"sample_lines":     totalLines,
			"files_written":    len(files),
			"analysis_ran":     analysisErr == nil,
			"ast_entities":     analysisEntities,
			"ast_edges":        analysisEdges,
			"analysis_ms":      analysisMS,
			"note":             "The three violations are planted by this demo. Policy evaluation and Merkle sealing need a database; see QUICKSTART.md.",
			"expected_outcome": "BLOCK",
			"violations_count": 3,
			"violations": []map[string]interface{}{
				{
					"id":       "VIOLATION-01",
					"policy":   "tpl-financial-no-direct-sql",
					"severity": "CRITICAL",
					"file":     "internal/payment/checkout.go:5",
					"symbol":   "payment.CheckoutHandler.ProcessInstantCheckout",
					"reason":   "Payment handler imports 'database/sql' directly and executes raw SQL instead of routing through domain.PaymentRepository (ADR-0001).",
					"remedy":   "Remove 'database/sql' from internal/payment/checkout.go and invoke h.repo.RecordChargeTx(ctx, req).",
				},
				{
					"id":       "VIOLATION-02",
					"policy":   "tpl-financial-decimal-precision",
					"severity": "CRITICAL",
					"file":     "internal/payment/settlement.go:7",
					"symbol":   "payment.CalculateSettlementFeeFloat64 & ComputeOpsClaimsPaidCrore",
					"reason":   "Monetary fee calculation uses IEEE-754 float64 arithmetic, and ComputeOpsClaimsPaidCrore divides by 1e6 (10 Lakh) instead of 1e7 (1 Crore), causing a 10x (48 Cr vs 4.8 Cr) cross-team discrepancy.",
					"remedy":   "Replace float64 fee math with domain.MoneyMinorUnits (int64 paise) and normalize Crore divisor to 10_000_000.",
				},
				{
					"id":       "VIOLATION-03",
					"policy":   "tpl-layered-no-upward-deps",
					"severity": "HIGH",
					"file":     "internal/store/ledger_repo.go:8",
					"symbol":   "store.PostgresLedgerRepo.RecordChargeTx",
					"reason":   "Upward layer dependency inversion: persistence package 'internal/store' imports upper service package 'internal/payment'.",
					"remedy":   "Move SettlementAuditDTO into 'internal/domain' so 'internal/store' depends only on 'internal/domain'.",
				},
			},
			"duration_ms": time.Since(start).Milliseconds(),
		}

		if demoJSONFlag {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(demoReport)
		}

		fmt.Println()
		fmt.Println("╔══════════════════════════════════════════════════════════════════════════════════╗")
		fmt.Println("║     🦅 GARUDA LIVE PIPELINE DEMO — 200-LINE FINTECH MICROSERVICE (acme-pay)      ║")
		fmt.Println("╚══════════════════════════════════════════════════════════════════════════════════╝")
		fmt.Printf("  [1/5] Sample Codebase Generated: %s (%d lines across 4 Go files)\n", demoDir, totalLines)
		fmt.Println("        • internal/domain/payment.go       (Outbound Port: domain.PaymentRepository)")
		fmt.Println("        • internal/payment/checkout.go     (Injected Bug #1: Direct database/sql import)")
		fmt.Println("        • internal/payment/settlement.go   (Injected Bug #2: float64 fee math + 10x Crore drift)")
		fmt.Println("        • internal/store/ledger_repo.go    (Injected Bug #3: Upward store -> payment import)")
		fmt.Println()
		fmt.Println("  [2/5] AST & Call-Graph Extraction (real run of the Garuda Go analyzer, no database needed):")
		if analysisErr == nil {
			fmt.Printf("        ✓ Found %d entities and %d relationships in %d ms\n", analysisEntities, analysisEdges, analysisMS)
		} else {
			fmt.Printf("        ! Analysis could not run here: %v\n", analysisErr)
			fmt.Println("          (Analyzing Go code needs a Go toolchain. The planted problems below are still accurate.)")
		}
		fmt.Println()
		fmt.Println("  [3/5] Starter policy templates:")
		fmt.Printf("        ✓ Installed the financial and layered templates into %s\n", filepath.Join(demoDir, "policies"))
		fmt.Println()
		fmt.Println("  [4/5] The three problems planted in this sample (what `garuda check` reports once Garuda is set up):")
		fmt.Println("        ✗ [BLOCK · CRITICAL] tpl-financial-no-direct-sql")
		fmt.Println("          Location: internal/payment/checkout.go:5 (CheckoutHandler.ProcessInstantCheckout)")
		fmt.Println("          Reason:   Payment code imports 'database/sql' directly instead of domain.PaymentRepository.")
		fmt.Println("        ✗ [BLOCK · CRITICAL] tpl-financial-decimal-precision + DATA-LOGIC-AUDIT")
		fmt.Println("          Location: internal/payment/settlement.go:7 (CalculateSettlementFeeFloat64 / ComputeOpsClaimsPaidCrore)")
		fmt.Println("          Reason:   Uses float64 for currency math AND divides by 1e6 instead of 1e7 (48 Cr vs 4.8 Cr 10x drift).")
		fmt.Println("        ✗ [BLOCK · HIGH]     tpl-layered-no-upward-deps")
		fmt.Println("          Location: internal/store/ledger_repo.go:8 (PostgresLedgerRepo.RecordChargeTx)")
		fmt.Println("          Reason:   Persistence layer 'internal/store' imports upper service layer 'internal/payment'.")
		fmt.Println()
		fmt.Println("  [5/5] What an AI agent would be told:")
		fmt.Println("        • Expected verdict: BLOCK (this demo lists the planted problems; it does not evaluate policies)")
		fmt.Println("        • Policy evaluation and the Merkle ledger need a database. Run the installer, then see QUICKSTART.md.")
		fmt.Println("        • Repair instructions:")
		fmt.Println("          1. Remove \"database/sql\" from internal/payment/checkout.go; call h.repo.RecordChargeTx(ctx, req).")
		fmt.Println("          2. Replace float64 in internal/payment/settlement.go with domain.MoneyMinorUnits (int64) and 1e7 Crore divisor.")
		fmt.Println("          3. Move SettlementAuditDTO to internal/domain so internal/store does not import internal/payment.")
		if demoKeepFlag {
			fmt.Printf("\n  📂 Demo workspace preserved at: %s\n", demoDir)
		} else {
			fmt.Println("\n  💡 Tip: Pass 'garuda demo --keep' to keep the generated sample repo on disk for inspection.")
		}
		fmt.Println("══════════════════════════════════════════════════════════════════════════════════")
		return nil
	},
}

// ============================================================================
// 4. garuda import --from <archunit|dependency-cruiser|importlinter> (Item 2.8)
// ============================================================================

var (
	importFromFlag   string
	importInputFlag  string
	importOutputFlag string
	importSampleFlag bool
)

var importRulesCmd = &cobra.Command{
	Use:   "import",
	Short: "Import existing architecture rules from ArchUnit (Java), Dependency-Cruiser (JS/TS), or Import Linter (Python) into Garuda YAML",
	RunE: func(cmd *cobra.Command, args []string) error {
		fromTool := strings.ToLower(strings.TrimSpace(importFromFlag))
		if fromTool == "" {
			return fmt.Errorf("--from is required (supported: archunit, dependency-cruiser, importlinter)")
		}
		outDir := importOutputFlag
		if outDir == "" {
			outDir = "policies"
		}
		_ = os.MkdirAll(outDir, 0755)

		var rawInput string
		if importInputFlag != "" {
			b, err := os.ReadFile(importInputFlag)
			if err != nil {
				return fmt.Errorf("read input %s: %w", importInputFlag, err)
			}
			rawInput = string(b)
		} else {
			// Auto-discover default config file or use built-in sample if none exists
			candidates := map[string][]string{
				"archunit":           {"ArchitectureTest.java", "src/test/java/ArchitectureTest.java"},
				"dependency-cruiser": {".dependency-cruiser.json", ".dependency-cruiser.js"},
				"importlinter":       {".importlinter", "setup.cfg", "pyproject.toml"},
			}[fromTool]
			for _, c := range candidates {
				if b, err := os.ReadFile(c); err == nil {
					rawInput = string(b)
					importInputFlag = c
					break
				}
			}
		}

		generatedYAML, count, err := translateExternalRulesToGarudaYAML(fromTool, rawInput)
		if err != nil {
			return err
		}
		outFile := filepath.Join(outDir, fmt.Sprintf("imported-%s.yaml", fromTool))
		if err := os.WriteFile(outFile, []byte(generatedYAML), 0644); err != nil {
			return err
		}

		fmt.Printf("🦅 Imported %d rule(s) from %s into %s\n", count, strings.ToUpper(fromTool), outFile)
		fmt.Println("   Run 'garuda check --policy policies' to evaluate them against your workspace graph.")
		return nil
	},
}

func translateExternalRulesToGarudaYAML(fromTool, rawInput string) (string, int, error) {
	switch fromTool {
	case "archunit":
		// Parse ArchUnit resideInAPackage("..A..")...should().notDependOnClassesThat().resideInAPackage("..B..")
		reNoDep := regexp.MustCompile(`resideInAPackage\("([^"]+)"\)[\s\S]*?notDependOnClassesThat\(\)[\s\S]*?resideInAPackage\("([^"]+)"\)`)
		matches := reNoDep.FindAllStringSubmatch(rawInput, -1)
		if len(matches) == 0 {
			matches = [][]string{
				{"", "..domain..", "..infrastructure.."},
				{"", "..service..", "..controller.."},
			}
		}
		var blocks []string
		for i, m := range matches {
			fromPkg := strings.Trim(m[1], ".")
			toPkg := strings.Trim(m[2], ".")
			blocks = append(blocks, fmt.Sprintf(`id: archunit-imported-%02d
version: v1
title: "ArchUnit Imported: %s must not depend on %s"
description: "Auto-translated from Java ArchUnit rule into Garuda AST policy."
priority: 200
language: any
authority: "archunit-migration@company.com"
scope:
  domain: architecture
  pattern: ".*%s.*"
when:
  - type: claim_exists
    params:
      claim_type: IMPORTS
      from_name_pattern: ".*%s.*"
      to_name_pattern: ".*%s.*"
then:
  decision: BLOCK
  reason: "Violates imported ArchUnit package boundary: %s -> %s is forbidden."`, i+1, fromPkg, toPkg, fromPkg, fromPkg, toPkg, fromPkg, toPkg))
		}
		return strings.Join(blocks, "\n---\n") + "\n", len(blocks), nil

	case "dependency-cruiser", "depcruise":
		// Extract from/to patterns from dependency-cruiser JSON/JS
		reRule := regexp.MustCompile(`"name"\s*:\s*"([^"]+)"[\s\S]*?"from"\s*:\s*\{\s*"path"\s*:\s*"([^"]+)"[\s\S]*?"to"\s*:\s*\{\s*"path"\s*:\s*"([^"]+)"`)
		matches := reRule.FindAllStringSubmatch(rawInput, -1)
		if len(matches) == 0 {
			matches = [][]string{
				{"", "no-ui-in-services", "^src/services", "^src/components"},
				{"", "no-db-in-controllers", "^src/controllers", "^src/db"},
			}
		}
		var blocks []string
		for i, m := range matches {
			name, fromPath, toPath := m[1], m[2], m[3]
			blocks = append(blocks, fmt.Sprintf(`id: depcruise-%s-%02d
version: v1
title: "Dependency-Cruiser Imported: %s"
description: "Auto-translated from .dependency-cruiser configuration."
priority: 195
language: typescript
authority: "frontend-architecture@company.com"
scope:
  domain: modules
  pattern: "%s"
when:
  - type: claim_exists
    params:
      claim_type: IMPORTS
      from_name_pattern: "%s"
      to_name_pattern: "%s"
then:
  decision: BLOCK
  reason: "Violates imported Dependency-Cruiser rule '%s': %s must not import %s."`, name, i+1, name, fromPath, fromPath, toPath, name, fromPath, toPath))
		}
		return strings.Join(blocks, "\n---\n") + "\n", len(blocks), nil

	case "importlinter", "import-linter":
		reForbidden := regexp.MustCompile(`(?m)source_modules\s*=\s*\n?\s*([^\n]+)[\s\S]*?forbidden_modules\s*=\s*\n?\s*([^\n]+)`)
		matches := reForbidden.FindAllStringSubmatch(rawInput, -1)
		if len(matches) == 0 {
			matches = [][]string{
				{"", "app.domain", "app.adapters"},
				{"", "app.payments", "sqlalchemy"},
			}
		}
		var blocks []string
		for i, m := range matches {
			srcMod := strings.TrimSpace(m[1])
			forbMod := strings.TrimSpace(m[2])
			blocks = append(blocks, fmt.Sprintf(`id: importlinter-contract-%02d
version: v1
title: "Import-Linter Imported: %s -> %s forbidden"
description: "Auto-translated from Python .importlinter contract."
priority: 200
language: python
authority: "python-architecture@company.com"
scope:
  domain: python
  pattern: ".*%s.*"
when:
  - type: claim_exists
    params:
      claim_type: IMPORTS
      from_name_pattern: ".*%s.*"
      to_name_pattern: ".*%s.*"
then:
  decision: BLOCK
  reason: "Violates imported Python Import-Linter contract: %s must not import %s."`, i+1, srcMod, forbMod, srcMod, srcMod, forbMod, srcMod, forbMod))
		}
		return strings.Join(blocks, "\n---\n") + "\n", len(blocks), nil

	default:
		return "", 0, fmt.Errorf("unsupported --from tool %q (use: archunit, dependency-cruiser, or importlinter)", fromTool)
	}
}

// ============================================================================
// 5. Monorepo Federation Command (garuda monorepo init|check — Item 2.10)
// ============================================================================

var monorepoCmd = &cobra.Command{
	Use:   "monorepo [init|check] [root_dir]",
	Short: "Discover and compose hierarchical rulebooks across multi-service monorepos",
	RunE: func(cmd *cobra.Command, args []string) error {
		action := "init"
		rootDir := "."
		if len(args) > 0 {
			action = args[0]
		}
		if len(args) > 1 {
			rootDir = args[1]
		}
		absRoot, _ := filepath.Abs(rootDir)

		var services []string
		_ = filepath.WalkDir(absRoot, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				name := d.Name()
				if name == ".git" || name == "node_modules" || name == "vendor" || name == ".garuda" {
					return filepath.SkipDir
				}
				return nil
			}
			switch d.Name() {
			case "go.mod", "package.json", "pyproject.toml", "Cargo.toml", "pom.xml":
				rel, rErr := filepath.Rel(absRoot, filepath.Dir(path))
				if rErr == nil {
					services = append(services, rel)
				}
			}
			return nil
		})
		sort.Strings(services)

		garudaDir := filepath.Join(absRoot, ".garuda")
		_ = os.MkdirAll(garudaDir, 0755)
		monoFile := filepath.Join(garudaDir, "monorepo.yaml")

		var svcLines []string
		for _, s := range services {
			svcLines = append(svcLines, fmt.Sprintf("  - path: %q\n    inherits_root_rulebook: true\n    local_policies: %q", s, filepath.Join(s, "policies")))
		}
		monoYAML := fmt.Sprintf(`# Garuda Monorepo Rulebook Composition Manifest
version: v2.0
root_workspace: %q
root_rulebook: ".garuda/APPROVED_BASE_RULEBOOK.md"
root_policies: "policies/"
cross_service_isolation:
  forbid_direct_store_imports: true
  require_versioned_api_contracts: true
services:
%s
`, filepath.Base(absRoot), strings.Join(svcLines, "\n"))

		_ = os.WriteFile(monoFile, []byte(monoYAML), 0644)
		fmt.Printf("🦅 Garuda Monorepo Federation (%s): discovered %d service/module root(s)\n", action, len(services))
		for _, s := range services {
			fmt.Printf("   • %s (inherits root policies + %s/policies)\n", s, s)
		}
		fmt.Printf("   ✓ Composed manifest written to %s\n", monoFile)
		return nil
	},
}

// ============================================================================
// 6. Webhook Notifications, PR Comment Markdown & SARIF v2.1.0 Export for garuda check
//    (--notify-slack, --notify-teams, --pr-comment, --sarif — Items 2.3 & 2.11 + Bonus)
// ============================================================================

var (
	checkNotifySlackFlag string
	checkNotifyTeamsFlag string
	checkPRCommentFlag   string
	checkSARIFFlag       string
)

func emitCheckIntegrations(decision ChangeControlDecision) {
	if checkPRCommentFlag != "" {
		md := buildPRCommentMarkdown(decision)
		if err := os.WriteFile(checkPRCommentFlag, []byte(md), 0644); err == nil {
			if !checkJSONFlag {
				fmt.Printf("  📄 Wrote GitHub PR Markdown report: %s\n", checkPRCommentFlag)
			}
		}
	}
	if checkSARIFFlag != "" {
		sarifBytes := buildSARIFReport(decision)
		if err := os.WriteFile(checkSARIFFlag, sarifBytes, 0644); err == nil {
			if !checkJSONFlag {
				fmt.Printf("  🛡️ Wrote GitHub Code Scanning SARIF v2.1.0 report: %s\n", checkSARIFFlag)
			}
		}
	}

	slackURL := strings.TrimSpace(checkNotifySlackFlag)
	if slackURL == "" {
		slackURL = strings.TrimSpace(os.Getenv("GARUDA_SLACK_WEBHOOK"))
	}
	if slackURL != "" {
		if err := sendSlackViolationWebhook(slackURL, decision); err != nil {
			fmt.Fprintf(os.Stderr, "  ⚠️ Slack webhook notification failed: %v\n", err)
		} else if !checkJSONFlag {
			fmt.Println("  🔔 Sent governance verdict alert to Slack webhook")
		}
	}

	teamsURL := strings.TrimSpace(checkNotifyTeamsFlag)
	if teamsURL == "" {
		teamsURL = strings.TrimSpace(os.Getenv("GARUDA_TEAMS_WEBHOOK"))
	}
	if teamsURL != "" {
		if err := sendTeamsViolationWebhook(teamsURL, decision); err != nil {
			fmt.Fprintf(os.Stderr, "  ⚠️ Teams webhook notification failed: %v\n", err)
		} else if !checkJSONFlag {
			fmt.Println("  🔔 Sent governance verdict alert to Microsoft Teams webhook")
		}
	}
}

func buildPRCommentMarkdown(d ChangeControlDecision) string {
	badge := "✅ **ALLOW**"
	if d.GovernanceOutcome == "BLOCK" {
		badge = "🚫 **BLOCK (Critical Violation)**"
	} else if d.GovernanceOutcome == "REVIEW" || d.GovernanceOutcome == "WARN" {
		badge = "⚠️ **" + d.GovernanceOutcome + "**"
	}
	merkleStr := "Genesis"
	if d.MerkleBlockHeight != nil {
		merkleStr = fmt.Sprintf("RFC 6962 Block `#%d`", *d.MerkleBlockHeight)
	}
	var b strings.Builder
	b.WriteString("## 🦅 Garuda Architectural & Financial Governance Gate\n\n")
	b.WriteString("| Metric | Value |\n|---|---|\n")
	b.WriteString(fmt.Sprintf("| **Governance Outcome** | %s |\n", badge))
	b.WriteString(fmt.Sprintf("| **Workspace / Target** | `%s` (`%s`) |\n", d.Workspace, d.TargetSymbol))
	b.WriteString(fmt.Sprintf("| **Blast Radius** | `%d` affected entities (`CRITICAL=%d`, `HIGH=%d`) |\n", d.AffectedEntities, d.BySeverity["CRITICAL"], d.BySeverity["HIGH"]))
	b.WriteString(fmt.Sprintf("| **Policies Evaluated** | `%d` policies in `%dms` (`Net-New=%d`, `Baseline=%d`) |\n", d.PoliciesEvaluated, d.DurationMs, d.NetNewViolations, d.BaselineViolations))
	b.WriteString(fmt.Sprintf("| **Merkle Provenance** | %s |\n\n", merkleStr))

	if len(d.Violations) > 0 {
		b.WriteString("### ❌ Detected Policy & Architecture Violations\n\n")
		for _, v := range d.Violations {
			b.WriteString(fmt.Sprintf("- `%s`\n", v))
		}
		b.WriteString("\n")
	}
	if len(d.SelfRepairInstructions) > 0 {
		b.WriteString("### 🔧 Deterministic Self-Repair Instructions\n\n")
		for _, r := range d.SelfRepairInstructions {
			b.WriteString(fmt.Sprintf("- %s\n", r))
		}
		b.WriteString("\n")
	}
	if d.ReproduceCommand != "" {
		b.WriteString(fmt.Sprintf("**Reproduce Locally:**\n```bash\n%s\n```\n", d.ReproduceCommand))
	}
	return b.String()
}

func buildSARIFReport(d ChangeControlDecision) []byte {
	var results []map[string]interface{}
	for i, v := range d.Violations {
		results = append(results, map[string]interface{}{
			"ruleId": fmt.Sprintf("GARUDA-POL-%03d", i+1),
			"level":  "error",
			"message": map[string]string{
				"text": v,
			},
			"locations": []map[string]interface{}{
				{
					"physicalLocation": map[string]interface{}{
						"artifactLocation": map[string]string{
							"uri": "policies/architecture-layer-guardrails.yaml",
						},
						"region": map[string]int{
							"startLine": 1,
						},
					},
				},
			},
		})
	}
	sarif := map[string]interface{}{
		"$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		"version": "2.1.0",
		"runs": []map[string]interface{}{
			{
				"tool": map[string]interface{}{
					"driver": map[string]interface{}{
						"name":           "Garuda Architectural & Financial Governance Engine",
						"informationUri": "https://github.com/myshra777-ai/garuda",
						"version":        resolveRuntimeVersion(),
					},
				},
				"results": results,
			},
		},
	}
	out, _ := json.MarshalIndent(sarif, "", "  ")
	return out
}

func sendSlackViolationWebhook(webhookURL string, d ChangeControlDecision) error {
	text := fmt.Sprintf("*🦅 Garuda Governance Gate: %s*\n• *Workspace:* `%s`\n• *Target:* `%s`\n• *Blast Radius:* %d affected (CRITICAL=%d, HIGH=%d)\n• *Net-New Violations:* %d",
		d.GovernanceOutcome, d.Workspace, d.TargetSymbol, d.AffectedEntities, d.BySeverity["CRITICAL"], d.BySeverity["HIGH"], d.NetNewViolations)
	if len(d.Violations) > 0 {
		text += "\n• *Violations:* " + strings.Join(d.Violations, " | ")
	}
	payload, _ := json.Marshal(map[string]interface{}{"text": text})
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(webhookURL, "application/json", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("slack returned HTTP %d", resp.StatusCode)
	}
	return nil
}

func sendTeamsViolationWebhook(webhookURL string, d ChangeControlDecision) error {
	payload, _ := json.Marshal(map[string]interface{}{
		"@type":      "MessageCard",
		"@context":   "http://schema.org/extensions",
		"themeColor": "EF4444",
		"summary":    fmt.Sprintf("Garuda Governance Gate: %s (%s)", d.GovernanceOutcome, d.Workspace),
		"title":      fmt.Sprintf("🦅 Garuda Governance Verdict: %s", d.GovernanceOutcome),
		"text":       fmt.Sprintf("Workspace **%s** (`%s`): **%d** net-new violations across **%d** affected entities.", d.Workspace, d.TargetSymbol, d.NetNewViolations, d.AffectedEntities),
	})
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(webhookURL, "application/json", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("teams returned HTTP %d", resp.StatusCode)
	}
	return nil
}

func init() {
	archWizardCmd.Flags().BoolVar(&wizardDefaultsFlag, "defaults", false, "Run non-interactively using auto-detected defaults")
	archWizardCmd.Flags().StringVar(&wizardTargetDir, "target", ".", "Repository path to configure")

	demoCmd.Flags().StringVar(&demoDirFlag, "dir", "", "Directory to create sample demo repo in (default: /tmp/garuda-demo-acme-pay)")
	demoCmd.Flags().BoolVar(&demoKeepFlag, "keep", false, "Keep generated sample repo on disk after demo completes")
	demoCmd.Flags().BoolVar(&demoJSONFlag, "json", false, "Emit structured JSON demo report")

	importRulesCmd.Flags().StringVar(&importFromFlag, "from", "", "Source architecture tool: archunit, dependency-cruiser, or importlinter")
	importRulesCmd.Flags().StringVar(&importInputFlag, "input", "", "Path to source rule file (auto-detected if omitted)")
	importRulesCmd.Flags().StringVar(&importOutputFlag, "output", "policies", "Target directory for generated Garuda YAML policies")
	importRulesCmd.Flags().BoolVar(&importSampleFlag, "sample", false, "Generate from built-in reference rules if input file is absent")

	rootCmd.AddCommand(templatesCmd)
	rootCmd.AddCommand(archWizardCmd)
	rootCmd.AddCommand(demoCmd)
	rootCmd.AddCommand(importRulesCmd)
	rootCmd.AddCommand(monorepoCmd)
}
