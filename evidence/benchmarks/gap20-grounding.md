# GAP-20 Epistemic Grounding Benchmark

The GAP-20 benchmark evaluates LLM refactoring accuracy on high-centrality Go symbols, comparing unassisted frontier models against Garuda MCP-grounded generation.

> **Scope:** These are observed results from the GAP-20 controlled grounding benchmark on the symbols and cases in that harness. They are not a universal hallucination rate, a production guarantee, or a measurement of every relationship in arbitrary repositories. The current evidence index and verification boundaries are in [EVIDENCE.md](../../EVIDENCE.md).

---

## Benchmark Results Matrix

| Metric Dimension | Naive LLM (Unassisted) | Garuda MCP Grounded | Delta / Gain |
| :--- | :---: | :---: | :---: |
| **Average Symbol Precision** | 40.0% | **100.0%** | **+150.0%** |
| **Upstream Caller Recall** | 20.0% | **100.0%** | **+80.0%** |
| **Downstream Dep Recall** | 33.0% | **100.0%** | **+67.0%** |
| **Observed hallucinations in benchmark cases** | 66.7% | **0.0%** | No wrong signature/receiver result observed in this benchmark |
| **Quarantine Isolation Rate** | 0.0% | **100.0%** | **+100.0%** |
| **Context Overhead (Tokens)** | 4,850 tokens | **620 tokens** | **-87.2% Compression** |

---

## Benchmark Findings

1. **Observed grounding result:** In the benchmarked cases, compiler-verified AST subgraphs prevented the measured method-signature and receiver mistakes. This does not establish a zero hallucination rate for other tasks or repositories.
2. **Token efficiency in this comparison:** Context overhead dropped by **87.2%** (from 4,850 tokens of raw file dumps to 620 tokens of structured semantic context). Treat this as a task- and prompt-specific result, not a universal token-saving guarantee.
3. **Contradiction handling in this benchmark:** The tested Garuda-grounded path surfaced the benchmark's quarantined endpoints. The result does not show that all runtime violations are discoverable or that every agent will obey injected context.
