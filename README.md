

<p align="center">
  <img src="assets/garuda-logo.png" width="96" alt="Garuda Logo">
</p>

<h1 align="center">Garuda</h1>

<p align="center">
  <strong>Prevent AI coding agents from compromising your system architecture.</strong><br>
  Deterministic CI/CD gates, verified MCP codebase context, and automated architectural drift prevention.
</p>

<p align="center">
  <a href="https://github.com/myshra777-ai/garuda/releases/tag/v0.2.0"><img src="https://img.shields.io/badge/release-v0.2.0-blue.svg?style=flat-square" alt="Release"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache_2.0-green.svg?style=flat-square" alt="License"></a>
  <a href="PLAYBOOK.md"><img src="https://img.shields.io/badge/Quickstart-Read%20the%20Playbook-orange.svg?style=flat-square" alt="Quickstart"></a>
  <a href="EVIDENCE.md"><img src="https://img.shields.io/badge/Benchmarks-View%20Evidence-purple.svg?style=flat-square" alt="Benchmarks"></a>
</p>

---

## 💡 Why Garuda?

Autonomous coding agents (Cursor, Claude, Devin, Codex) generate code faster than humans can review it. While individual functions may look correct, multi-file changes often introduce **architectural drift**:
- Presentation layers silently importing database drivers.
- Circular dependencies and undesirable cross-package coupling.
- Changes that pass unit tests but violate core system invariants.

Traditional linters analyze files in isolation. **Garuda gives AI agents a verified, compiler-backed system graph *before* they write code, and acts as a deterministic policy gate *during* CI/CD.**

---

## 🚀 Key Features

- 🛡️ **CI/CD Enforcement Gate:** Validate architectural constraints on every PR. Use `garuda policy evaluate --fail-on-block` to automatically intercept breaking changes.
- 🧠 **Verified MCP Context:** Integrate with **Cursor, Claude Desktop, or Codex CLI**. Agents query Garuda for exact symbol and call graphs instead of guessing package dependencies.
- ⚡ **Local Drift Prevention:** The VS Code extension reports structural violations in real-time, surfacing `ARCH_DRIFT` warnings before you even commit.
- 🔒 **Cryptographic Provenance:** Every policy validation produces a Merkle-backed receipt, creating a verifiable audit history of structural modifications.

---

## 🏗️ Architecture Overview

```mermaid
flowchart LR
    A[Repository Code + Policies] --> B[Garuda: Semantic Ontology Engine]
    B --> C[MCP Server: Verified Context for AI Agents]
    B --> D[IDE / CLI: Immediate Impact Analysis]
    B --> E[CI/CD: Deterministic Policy Evaluation]
    E --> F{Merge Gate}
    F -->|Clean| G[Pass & Generate Provenance Receipt]
    F -->|Violation| H[Block PR: Zero Bypass]
```
*Garuda complements your existing test suites and linters by adding a dedicated structural safety layer.*

---

## ⚡ Quickstart

### Prerequisites
- Go 1.25+
- PostgreSQL 14+

### 1. Build from Source
```bash
git clone https://github.com/myshra777-ai/garuda.git
cd garuda
mkdir -p bin
go build -o bin/garuda ./cmd/garuda
go build -o bin/garuda-mcp ./cmd/garuda-mcp
```
*(See the [Playbook](PLAYBOOK.md) for automated database configuration and migration setup).*

### 2. Connect to Cursor or Claude Desktop
Add Garuda to your agent configuration (`claude_desktop_config.json` or `.cursor/mcp.json`). **Note:** Ensure `DATABASE_URL` and `GARUDA_WORKSPACE` match your environment.

```json
{
  "mcpServers": {
    "garuda": {
      "command": "/absolute/path/to/garuda/bin/garuda-mcp",
      "env": {
        "DATABASE_URL": "postgres://user:pass@localhost:5432/garuda?sslmode=disable",
        "GARUDA_WORKSPACE": "my-workspace"
      }
    }
  }
}
```
Now ask your agent: *"Query Garuda to find all downstream callers of ProcessOrder before editing this package."*

### 3. Enforce in CI/CD
Add a blocking check to your GitHub Actions workflow:

```yaml
- name: Evaluate Architecture Policy
  env:
    DATABASE_URL: ${{ secrets.DATABASE_URL }}
    GARUDA_WORKSPACE: my-workspace
  run: ./bin/garuda policy evaluate . --policies ./policies --save --fail-on-block
```

---

## 🌐 Language Support & Scope

| Language | Analysis Engine | Status |
| :--- | :--- | :--- |
| **Go** | Compiler-backed AST & Type Resolution (Tier 5) | 🟢 Production-ready |
| **Python** | Structural Abstract Syntax Tree | 🟡 Beta |
| **TypeScript** | Structural Abstract Syntax Tree | 🟡 Beta |

> **Engineering Note:** A passing check verifies compliance against your configured policies and evidence. `UNVERIFIED` indicates insufficient graph evidence, not a negative assertion. Read our [Trust Invariants](docs/invariants.md) for full formal guarantees.

---

## 📚 Documentation

Don't guess. We keep our documentation explicit and evidence-backed.

- 📘 **[Playbook & Setup](PLAYBOOK.md)** — Step-by-step database setup, workspace commands, and daily workflows.
- 📙 **[System Specifications](docs/SPECS.md)** — Complete MCP surface, CLI reference, and CI/CD policy syntax.
- 📊 **[Verification Evidence](EVIDENCE.md)** — Benchmarks across 14 production Go repositories and controlled contradiction tests.
- 🛡️ **[Security Policy](SECURITY.md)** · **[Contributing](CONTRIBUTING.md)**

---
<p align="center">
  <strong>Understand the system. Verify the change. Keep AI aligned.</strong>
</p>


