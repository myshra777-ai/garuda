# Listing Garuda in plugin and MCP directories

None of these need a company or LLC; a GitHub account is enough. Items marked **(verify)** are from memory of each directory's process. Check the directory's current page before submitting.

## What is already in this repo

| File | Used by |
| :--- | :--- |
| [`.claude-plugin/plugin.json`](../.claude-plugin/plugin.json), [`marketplace.json`](../.claude-plugin/marketplace.json) | Claude Code plugin marketplace (this repo *is* the marketplace) |
| [`.codex-plugin/plugin.json`](../.codex-plugin/plugin.json), [`.agents/plugins/marketplace.json`](../.agents/plugins/marketplace.json) | OpenAI Codex plugins |
| [`.cursor-plugin/plugin.json`](../.cursor-plugin/plugin.json) | Cursor plugins (**verify** schema) |
| [`.mcp.json`](../.mcp.json) | MCP config shared by the plugins |
| [`server.json`](../server.json) + [`Dockerfile.mcp`](../Dockerfile.mcp) + [`publish-mcp-image.yml`](../.github/workflows/publish-mcp-image.yml) + [`publish-mcp-registry.yml`](../.github/workflows/publish-mcp-registry.yml) | Official MCP Registry |
| [`release.yml`](../.github/workflows/release.yml), [`release-packages.yml`](../.github/workflows/release-packages.yml) | Draft GitHub release (binaries, checksums, provenance) and the API container image |
| [`smithery.yaml`](../smithery.yaml) | Smithery |
| [`glama.json`](../glama.json) | Glama |
| [`skills/`](../skills/) | All of the above (6 skills) |
| [`action.yml`](../action.yml) | GitHub Marketplace (Action) |

## Check the listing files locally first

```bash
python3 scripts/check_marketplace_manifests.py   # every listing file is tracked, not ignored, and at the current version
python3 scripts/sync_docs_from_manifest.py --check
claude plugin validate .          # Claude Code manifests
claude --plugin-dir .             # load the plugin from this folder
codex plugin marketplace add ./   # Codex, from a local checkout
```

## Cutting a release

`capabilities.json` is the single source of truth for the version. For the next release (example: `v1.9.11`):

```bash
python3 scripts/sync_docs_from_manifest.py --set v1.9.11 --date 2026-12-01   # rewrites every version string
python3 scripts/sync_docs_from_manifest.py --check                            # fix anything it still reports by hand
# write the v1.9.11 entry in CHANGELOG.md and replace RELEASE_NOTES.md
git commit -am "Release v1.9.11" && git push
git tag v1.9.11 && git push origin v1.9.11
```

Pushing the tag starts three workflows:

| Workflow | What it does |
| :--- | :--- |
| `release.yml` | Checks that `version.txt` equals the tag, that the docs and manifests are in sync, builds and tests the Go code, cross-compiles the binaries, writes `SHA256SUMS.txt` and build-provenance attestations, then creates a **draft** GitHub release from `RELEASE_NOTES.md`. |
| `publish-mcp-image.yml` | Pushes `ghcr.io/myshra777-ai/garuda-mcp:<version>` (the image the MCP Registry points at). |
| `release-packages.yml` | Pushes the API image `ghcr.io/myshra777-ai/garuda:<version>` and `:latest`. |

Then, by hand, because GitHub does not let a workflow publish these for you:

1. **Make the `garuda-mcp` package public** (first release only): GitHub → your profile → Packages → `garuda-mcp` → Package settings → Change visibility → Public. The registry workflow refuses to run while the image cannot be pulled anonymously.
2. **Review the draft release** under Releases, then **Publish release**. For the first release also tick *Publish this Action to the GitHub Marketplace* (needs a unique action name and 2FA on your account).
3. Publishing the release starts `publish-mcp-registry.yml`, which signs in to the MCP Registry with GitHub OIDC (no stored secret) and publishes `server.json`.

If a workflow fails, fix the cause and re-run it from the Actions tab (every one has *Run workflow*). `release.yml` also accepts an existing tag as input.

## Other listings, in order of effort

1. **Make the repo findable.** Set the GitHub *About* description and add topics: `mcp`, `mcp-server`, `claude-code`, `claude-code-plugin`, `codex`, `ai-agents`, `architecture`, `governance`, `merkle-tree`.
2. **Official MCP Registry.** Automated by `publish-mcp-registry.yml` (see above). To publish from your own machine instead: install `mcp-publisher`, then `mcp-publisher login github` and `mcp-publisher publish` in the repo root. The server name `io.github.myshra777-ai/garuda` requires you to be signed in as that GitHub account, and the `io.modelcontextprotocol.server.name` label in `Dockerfile.mcp` proves you own the image. Many aggregators (PulseMCP and others) read from this registry.
3. **Smithery**: connect the repo at smithery.ai/new. It reads `smithery.yaml`.
4. **Glama**: add the server at glama.ai/mcp/servers, then claim it with `glama.json`.
5. **Claude Code**: users can already run `/plugin marketplace add myshra777-ai/garuda`. For Anthropic's official directory, use the submission form linked from the Claude Code plugin docs **(verify)**.
6. **Codex**: users add it with `codex plugin marketplace add myshra777-ai/garuda`. The in-app directory is curated by OpenAI **(verify current submission route)**.
7. **GitHub Marketplace (Action)**: ticked when you publish the release (step 2 above). `action.yml` already has branding.
8. **Community lists**: open PRs to `punkpeye/awesome-mcp-servers` and `modelcontextprotocol/servers` (community section), and submit to mcp.so and cursor.directory **(verify each)**.

## Known gaps

- None of the release, image or registry workflows has run on GitHub yet. Run `publish-mcp-image.yml` once with *workflow_dispatch* and fix any failure before relying on the tag flow; `server.json` is only accepted by the registry once that image is public.
- A workflow run by `GITHUB_TOKEN` does not trigger other workflows. That is why the registry publish starts from the release you publish by hand, not from `release.yml` itself.
- The MCP server needs a PostgreSQL database. Directory users will not have one; the README tells them to run `./setup.sh` first. A zero-database "read-only demo mode" would make directory listings much easier to try.
