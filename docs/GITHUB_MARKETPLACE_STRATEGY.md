# Garuda Launch Strategy: GitHub Packages (`ghcr.io`) & GitHub Marketplace (`v1.9.9`)

This document outlines the technical packaging, CI/CD automation, verification gates, and listing workflow to launch **Garuda `v1.9.9`** across **GitHub Packages** and the **GitHub Marketplace**.

---

## 1. Two-Track Distribution Architecture

Garuda reaches engineering teams and AI agent workflows through two complementary GitHub channels:

| Channel | Artifact Type | Target Identifier | Primary Use Case |
| :--- | :--- | :--- | :--- |
| **Track A — GitHub Packages (`ghcr.io`)** | Multi-Arch OCI Container (`linux/amd64`, `linux/arm64`) + `.deb`/`.tar.gz`/`.zip` Release Assets | `ghcr.io/myshra777-ai/garuda:v1.9.9` | Zero-install Docker/Kubernetes deployment, devcontainers, air-gapped enterprise runners, and hosted MCP server containers. |
| **Track B — GitHub Marketplace (GitHub Action)** | Composite GitHub Action (`action.yml`) | `myshra777-ai/garuda@v1.9.9` | 3-line PR & push gate that runs diff-scoped AST blast-radius checks, Talon policy verification, and posts deterministic `receipt_hash` summaries to `$GITHUB_STEP_SUMMARY`. |

---

## 2. Track A — Publishing to GitHub Packages (`ghcr.io`)

### 2.1 What Is Already Wired in This Repository
- **Workflow**: `.github/workflows/release-packages.yml`
- **Authentication**: Uses the built-in `${{ secrets.GITHUB_TOKEN }}` with `packages: write`, `contents: write`, and `attestations: write` permissions—zero external registry secrets required.
- **OCI Metadata Labels**: Automatically binds `org.opencontainers.image.source` to `https://github.com/myshra777-ai/garuda` so the package card appears directly on the main repository sidebar under **Packages**.

### 2.2 Step-by-Step Commands to Publish `v1.9.9` to GitHub Packages

1. **Enable Workflow Read/Write & Package Permissions**:
   - In your GitHub repository, navigate to **Settings → Actions → General → Workflow permissions**.
   - Select **Read and write permissions** and check **Allow GitHub Actions to create and approve pull requests**.
2. **Tag and Push `v1.9.9`**:
   ```bash
   git add -A
   git commit -m "release: Garuda v1.9.9 — 54-tool behavioral suite, canonical manifest, progressive dashboard & Marketplace action"
   git tag -a v1.9.9 -m "Garuda v1.9.9 Universal Release"
   git push origin main --tags
   ```
3. **Set Package Visibility to Public**:
   - Once `.github/workflows/release-packages.yml` completes (~2 minutes), click **Packages** on your repository homepage (`ghcr.io/myshra777-ai/garuda`).
   - Open **Package settings → Danger Zone → Change visibility → Public** so any user or CI runner can run:
     ```bash
     docker pull ghcr.io/myshra777-ai/garuda:v1.9.9
     ```

---

## 3. Track B — Launching on the GitHub Marketplace

### 3.1 Repository Readiness Checklist (100% Complete in `v1.9.9`)
GitHub Marketplace enforces strict validation on repositories publishing an Action:
- [x] **Root `action.yml` Present**: Located at `/action.yml` with unique `name` (`"Garuda Talon Architecture & Data-Logic Gate"`), `description`, `author`, `inputs`, `outputs`, and `branding` (`icon: "shield"`, `color: "green"`).
- [x] **Public Repository with Open-Source License**: `LICENSE` (`Apache-2.0`).
- [x] **Canonical Version & Capability Parity**: Verified by `python3 scripts/sync_docs_from_manifest.py --check`.
- [x] **54/54 Behavioral & Security Tests Passing**: Verified by `go test ./cmd/garuda-mcp/...`.

### 3.2 Step-by-Step GitHub Marketplace Publishing Flow

1. **Accept the GitHub Marketplace Developer Agreement**:
   - Navigate to your repository on GitHub (`https://github.com/myshra777-ai/garuda`).
   - Notice the banner at the top of the repo when `action.yml` is committed: *"You can publish this Action to the GitHub Marketplace"* — click **Draft a release**.
2. **Configure the Marketplace Listing inside the Release Editor**:
   - Check **[x] Publish this Action to the GitHub Marketplace**.
   - **Primary Category**: Select **`Code quality`**.
   - **Secondary Category**: Select **`Continuous integration`** (or **`Security`**).
3. **Select the Tag & Release Title**:
   - **Tag**: `v1.9.9` (create new tag on publish or select existing `v1.9.9` tag).
   - **Release Title**: `Garuda v1.9.9 — Epistemic Architecture Gate, Talon AI Harness & Data/Logic Audit Engine`
   - **Release Body**: Copy the contents of `RELEASE_NOTES.md`.
4. **Click "Publish release"**:
   - Your Action is immediately live on `https://github.com/marketplace/actions/garuda-talon-architecture-data-logic-gate`!

### 3.3 Customer Copy-Paste Snippet for GitHub Marketplace Users

Once published, any repository can add Garuda to `.github/workflows/garuda.yml` in 12 lines:

```yaml
name: Garuda Architecture & Logic Gate
on: [pull_request, push]

jobs:
  garuda-gate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - name: Run Garuda Talon Gate
        uses: myshra777-ai/garuda@v1.9.9
        with:
          workspace: default
          policy-dir: policies
          fail-on-block: "true"
          matrix: "true"
```

---

## 4. Post-Launch Ecosystem Expansion (MCP Registry & VS Code Marketplace)

1. **Official Model Context Protocol (MCP) Registry & Smithery / Glama Listing**:
   - Submit `cmd/garuda-mcp` (`54` tools: `41` read-only, `13` mutating, `GARUDA_MCP_READ_ONLY` + `GARUDA_MCP_APPROVER` controls) with the canonical metadata from `capabilities.json`.
2. **VS Code Marketplace & OpenVSX (`vscode-extension/`)**:
   - Package `vscode-extension/` via `npx @vscode/vsce package` and publish to both the Visual Studio Marketplace (for VS Code) and **OpenVSX** (which powers **Cursor** and **Windsurf** extension marketplaces natively).
