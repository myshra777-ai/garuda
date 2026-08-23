# 🦅 Garuda Operator & Developer Playbook

A practical guide for bootstrapping, operating, and integrating Garuda into development workflows, IDEs, and CI pipelines.

---

## 1. Quickstart (60-Second Setup)

### Automated Bootstrap
```bash
# 1. Run universal installer
chmod +x install.sh && ./install.sh

# 2. Launch background unified daemon
garuda dev &
Manual Build from Source
Bash
git clone [https://github.com/myshra777-ai/garuda.git](https://github.com/myshra777-ai/garuda.git)
cd garuda
go build -o bin/garuda ./cmd/garuda
sudo cp bin/garuda /usr/local/bin/garuda
garuda init
2. Daily Developer Workflows
Static Codebase Analysis & Ledger Anchoring
Bash
# Analyze Go packages and persist AST claims to Merkle ledger
garuda analyze ./cmd ./internal -s --workspace uuid-ws

# Inspect overall platform health and active ledger height
garuda status
Architecture Inspection & Dead Code Audits
Bash
# View architectural hubs, interface consumers, and cross-repo bridges
garuda summary

# Audit dead code, duplicate functions, and Go standard library modernizations
garuda ponytail

# Trace blast radius of a specific symbol across all indexed repositories
garuda impact PostgresStore
Epistemic Grounding Benchmark
Bash
# Run the GAP-20 empirical comparison against ungrounded LLM inference
garuda bench
3. IDE Integration (VS Code / Cursor)
Build and install the extension:

Bash
cd vscode-extension
npm install && npm run compile
mkdir -p ~/.cursor/extensions/myshra777-ai.garuda-vscode-0.1.0
cp -r package.json out ~/.cursor/extensions/myshra777-ai.garuda-vscode-0.1.0/
Reload IDE window (Ctrl+Shift+P → Developer: Reload Window).

Features:

Inline Squiggles: Shows ARCH_DRIFT_001 warnings on unapproved runtime calls.

Blast Radius Hover: Hover over any Go struct/interface to view caller graphs.

Status Bar: Displays live block height $(shield) Garuda: #898 (10 Violations).

4. OpenTelemetry Telemetry Pipeline
Stream production runtime spans directly into Garuda using otel-collector-garuda.yaml:

YAML
exporters:
  garuda:
    endpoint: "http://localhost:8080/api/v1/telemetry/spans"
    tenant_id: "00000000-0000-0000-0000-000000000001"
    workspace_id: "532a8e33-975d-48a3-8f88-221cef52fec4"
Test span ingestion manually:

Bash
python3 scripts/test_otel_stream.py
5. Model Context Protocol (MCP) Configuration
Configure Cursor (.cursor/mcp.json) or Claude Desktop (~/.config/Claude/claude_desktop_config.json):

JSON
{
  "mcpServers": {
    "garuda": {
      "command": "/usr/local/bin/garuda",
      "args": ["mcp"],
      "env": {
        "DATABASE_URL": "postgres://test:test@localhost:5433/garuda_test?sslmode=disable"
      }
    }
  }
}
Available MCP Tools
get_runtime_state: Active block height, root hashes, and verification metrics.

get_contradictions: Active quarantined runtime violations.

get_verified_context: AST context filtered of unverified assumptions.

get_blast_radius: Recursive upstream/downstream dependency subgraphs.
