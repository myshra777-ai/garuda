markdown
# Garuda

**The truth maintenance system for AI agents.**

Garuda gives AI agents shared memory, decision lineage, contradiction detection, and cost optimization—cutting cold starts from 8 minutes to 10ms and saving 40%+ on token costs.

## Quick Start

```bash
# Install
go install github.com/myshra777-ai/garuda@latest

# Run with Docker
docker run -p 8080:8080 myshra777-ai/garuda:latest

# Create your first decision
curl -X POST http://localhost:8080/v1/decision/propose \
  -H "Authorization: Bearer <token>" \
  -d '{"statement": "Use PostgreSQL for financial records"}'
Why Garuda?
Without Garuda	With Garuda
Agents re-read the same code 5x	Agents reuse previous context
8-minute cold starts	10ms warm starts
40%+ token waste	Save 40%+ tokens
Contradictory decisions	First‑class contradiction detection
Features
Decision lineage – trace why any decision was made

Assumption decay – auto‑mark stale decisions

Contradiction detection – reject conflicting agent actions

Content‑addressable evidence – deduplicate context blocks

Telemetry – anonymous usage data to improve the product (opt‑out available)

Documentation
Full docs: docs.garuda.dev

Telemetry
Garuda collects anonymous telemetry to improve the product. No sensitive data is ever collected.

To opt out, set:

bash
export GARUDA_TELEMETRY_ENABLED=false
License
Apache 2.0



Description: "Truth maintenance system for AI agents — shared memory, lineage, and contradiction detection"

Topics: ai-agents, go, truth-maintenance, decision-graph, llm
