# Garuda launch channels: GitHub Packages, Releases and the Marketplace Action

How Garuda is published on GitHub, what each workflow does, and what you still do by hand. The step-by-step release commands live in [LISTING_GUIDE.md](LISTING_GUIDE.md#cutting-a-release); this page explains the channels.

Everything here works on a **free GitHub account**. No organization or LLC is needed.

---

## 1. Channels

| Channel | What users get | Identifier | Built by |
| :--- | :--- | :--- | :--- |
| **GitHub Releases** | Static binaries (Linux, macOS, Windows, FreeBSD, Android/Termux), `.deb`, `SHA256SUMS.txt`, build-provenance attestations | `https://github.com/myshra777-ai/garuda/releases` | [`release.yml`](../.github/workflows/release.yml) |
| **GitHub Packages (`ghcr.io`), API** | Multi-arch container (`linux/amd64`, `linux/arm64`) with all four binaries | `ghcr.io/myshra777-ai/garuda:<version>` | [`release-packages.yml`](../.github/workflows/release-packages.yml) |
| **GitHub Packages (`ghcr.io`), MCP server** | The `garuda-mcp` image the MCP Registry points at | `ghcr.io/myshra777-ai/garuda-mcp:<version>` | [`publish-mcp-image.yml`](../.github/workflows/publish-mcp-image.yml) |
| **GitHub Marketplace (Action)** | A pull-request and push gate: diff-scoped blast-radius checks, Talon policy verification, a deterministic `receipt_hash` in `$GITHUB_STEP_SUMMARY` | `myshra777-ai/garuda@v1.9.11` | [`action.yml`](../action.yml) |
| **Official MCP Registry** | Listing for `io.github.myshra777-ai/garuda` | [`server.json`](../server.json) | [`publish-mcp-registry.yml`](../.github/workflows/publish-mcp-registry.yml) |

The three tag-triggered workflows start when you push a `vX.Y.Z` tag. The registry workflow starts when you publish the draft release.

---

## 2. GitHub Packages (`ghcr.io`)

- **Authentication.** The workflows use the built-in `GITHUB_TOKEN` with `packages: write`; no registry secret is needed.
- **Repository link.** The images carry the `org.opencontainers.image.source` label, so the package appears in the repository sidebar under **Packages**.
- **Provenance.** `release-packages.yml` attests the image with `actions/attest-build-provenance`.
- **Visibility.** A package created by a workflow starts **private**. Make it public once: GitHub → your profile → **Packages** → the package → **Package settings** → **Change visibility** → **Public**. Do this for `garuda-mcp` before the registry workflow runs, because the registry must be able to pull the image without signing in.
- **Repository settings.** If a workflow cannot push, check **Settings → Actions → General → Workflow permissions**. The workflows ask for the permissions they need, so *Read repository contents* is enough as the default.

Pull the API image:

```bash
docker pull ghcr.io/myshra777-ai/garuda:1.9.11
```

Image tags use the bare version (`1.9.10`) and `latest`. Release archives use the `v` prefix (`garuda-v1.9.11-linux-amd64.tar.gz`).

---

## 3. GitHub Marketplace (the Action)

### 3.1 What GitHub requires

- [x] `action.yml` at the repository root, the only action metadata file there.
- [x] A unique action `name` ("Garuda Talon Architecture & Data-Logic Gate"), a `description`, and `branding` (icon `shield`, colour `green`).
- [x] A public repository with an open-source licence (Apache-2.0).
- [ ] Two-factor authentication on the publishing account and acceptance of the *GitHub Marketplace Developer Agreement* (done in the browser, once).

`python3 scripts/sync_docs_from_manifest.py --check` verifies that `action.yml` still fails closed and carries the current version.

### 3.2 Publishing

1. Open the draft release that `release.yml` created (**Releases**), or draft one by hand for a tag.
2. Tick **Publish this Action to the GitHub Marketplace** and accept the agreement if GitHub asks.
3. Choose a primary category (**Code quality**) and, optionally, a secondary one (**Continuous integration** or **Security**).
4. Click **Publish release**.

GitHub then lists the Action under `https://github.com/marketplace/actions/` with a name derived from `name:` in `action.yml`.

### 3.3 Copy-paste snippet for users

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
        uses: myshra777-ai/garuda@v1.9.11
        with:
          workspace: default
          policy-dir: policies
          fail-on-block: "true"
          matrix: "true"
```

Pin to a release tag (or a commit SHA) in your own workflows. The sync script keeps the tag in this snippet equal to the current release.

---

## 4. Other listings

- **MCP Registry, Smithery, Glama and plugin marketplaces:** see [LISTING_GUIDE.md](LISTING_GUIDE.md). The MCP server exposes `54` tools (`41` read-only, `13` mutating); `GARUDA_MCP_READ_ONLY` and `GARUDA_MCP_APPROVER` control what an agent may change.
- **VS Code Marketplace and Open VSX:** [`vscode-extension/`](../vscode-extension/) holds the TypeScript sources and the compiled `out/` folder, but the repository does not track an extension `package.json` or `tsconfig.json`, so it cannot be packaged from a fresh clone yet. Add them (they were hidden by an old `*.json` ignore rule; check `git status` on your machine), then `npx @vscode/vsce package` and publish to both marketplaces. Open VSX is what Cursor and Windsurf read.
