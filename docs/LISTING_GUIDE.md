# Listing Garuda in plugin and MCP directories

None of these need a company or LLC; a GitHub account is enough. Items marked **(verify)** are from memory of each directory's process. Check the directory's current page before submitting.

## What is already in this repo

| File | Used by |
| :--- | :--- |
| [`.claude-plugin/plugin.json`](../.claude-plugin/plugin.json), [`marketplace.json`](../.claude-plugin/marketplace.json) | Claude Code plugin marketplace (this repo *is* the marketplace) |
| [`.codex-plugin/plugin.json`](../.codex-plugin/plugin.json), [`.agents/plugins/marketplace.json`](../.agents/plugins/marketplace.json) | OpenAI Codex plugins |
| [`.cursor-plugin/plugin.json`](../.cursor-plugin/plugin.json) | Cursor plugins (**verify** schema) |
| [`.mcp.json`](../.mcp.json) | MCP config shared by the plugins |
| [`server.json`](../server.json) + [`Dockerfile.mcp`](../Dockerfile.mcp) + [`publish-mcp-image.yml`](../.github/workflows/publish-mcp-image.yml) | Official MCP Registry |
| [`smithery.yaml`](../smithery.yaml) | Smithery |
| [`glama.json`](../glama.json) | Glama |
| [`skills/`](../skills/) | All of the above (6 skills) |
| [`action.yml`](../action.yml) | GitHub Marketplace (Action) |

## Test locally first

```bash
claude plugin validate .          # Claude Code manifests
claude --plugin-dir .             # load the plugin from this folder
codex plugin marketplace add ./   # Codex, from a local checkout
```

## Steps, in order of effort

1. **Make the repo findable.** Set the GitHub *About* description and add topics: `mcp`, `mcp-server`, `claude-code`, `claude-code-plugin`, `codex`, `ai-agents`, `architecture`, `governance`, `merkle-tree`.
2. **Cut a release.** `git tag v1.9.9 && git push --tags` runs the release workflow and the MCP-image workflow. After the first run, set the new `garuda-mcp` package to **public** under GitHub → Packages.
3. **Official MCP Registry.** Install `mcp-publisher`, then `mcp-publisher login github` and `mcp-publisher publish` in the repo root. The server name `io.github.myshra777-ai/garuda` requires you to be signed in as that GitHub account. The image label in `Dockerfile.mcp` proves ownership **(verify: OCI label name)**. Many aggregators (PulseMCP and others) read from this registry.
4. **Smithery**: connect the repo at smithery.ai/new. It reads `smithery.yaml`.
5. **Glama**: add the server at glama.ai/mcp/servers, then claim it with `glama.json`.
6. **Claude Code**: users can already run `/plugin marketplace add myshra777-ai/garuda`. For Anthropic's official directory, use the submission form linked from the Claude Code plugin docs **(verify)**.
7. **Codex**: users add it with `codex plugin marketplace add myshra777-ai/garuda`. The in-app directory is curated by OpenAI **(verify current submission route)**.
8. **GitHub Marketplace (Action)**: Releases → *Publish this Action to the GitHub Marketplace*. `action.yml` already has branding. Needs a unique action name and 2FA.
9. **Community lists**: open PRs to `punkpeye/awesome-mcp-servers` and `modelcontextprotocol/servers` (community section), and submit to mcp.so and cursor.directory **(verify each)**.

## Known gaps

- `Dockerfile.mcp` and the publish workflow have not been built in CI yet. Run the workflow once with *workflow_dispatch* and fix any failure before publishing `server.json`.
- The MCP server needs a PostgreSQL database. Directory users will not have one; the README tells them to run `./setup.sh` first. A zero-database "read-only demo mode" would make directory listings much easier to try.
- The dashboard's About page still says "23 MCP tools"; the server exposes 54 (see `capabilities.json`).
