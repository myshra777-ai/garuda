# 🦅 Garuda Epistemic Intelligence: Visual Walkthrough & System Tour

This guide details the complete developer and operational visual interface of Garuda across the IDE extension, live epistemic visualizer, and daemon telemetry pipelines.

---

## 1. IDE Extension & Inline Epistemic Diagnostics

Garuda actively correlates Go AST compiler truth with runtime execution traces, surfacing violations and blast radiuses directly in the editor.

### Inline Contradiction Squiggles (`ARCH_DRIFT_001`)
When code attempts an unapproved runtime call or touches quarantined endpoints, Garuda flags the exact file and line with red squiggles and reports them in the VS Code / Cursor Problems tab.

<div align="center">
  <img src="../assets/screenshots/image_74eba3.png" alt="Garuda Inline Diagnostic Squiggles and Problems Panel" width="850" />
  <p><em>Real-time detection of quarantined runtime drift (ARCH_DRIFT_001) across multiple service handlers.</em></p>
</div>

---

### Real-Time Blast Radius Hover
Hovering over any Go struct, method, or interface displays upstream callers, downstream dependencies, and direct visualizer links across multi-repository boundaries.

<div align="center">
  <img src="../assets/screenshots/image_748dca.png" alt="Garuda Blast Radius Hover Context" width="850" />
  <p><em>Cross-module symbol hover displaying incoming callers, outgoing dependencies, and interface implementations.</em></p>
</div>

---

### Dual-Root Merkle Status Bar & Sidebar
The status bar and activity bar sidebar continuously poll the background verification daemon to display active Merkle block height and quarantined contradiction counts.

<div align="center">
  <img src="../assets/screenshots/image_749946.png" alt="Garuda Live Status Bar Telemetry" width="600" />
  <p><em>Live cryptographic ledger status (#898) with real-time contradiction alerts.</em></p>
</div>

---

## 2. D3 Interactive Epistemic Truth Graph

The unified daemon (`garuda dev`) serves an interactive dark-mode topology at `http://localhost:8080/graph` or directly via the command palette (`Garuda: Open Graph Visualizer`).

<div align="center">
  <img src="../assets/screenshots/image_750a24.png" alt="Garuda Force-Directed Epistemic Graph" width="900" />
  <p><em>Interactive topology: Purple nodes represent Go repositories, green nodes are verified callers, and red dashed edges indicate quarantined drift.</em></p>
</div>

---

## 3. Telemetry Stream & CI Verification

Runtime OpenTelemetry spans are continuously ingested, mapped, and audited against the cryptographic ledger:

* **Live OTel Ingestion:** Trace spans are streamed via `garudaexporter` to `http://localhost:8080/api/v1/telemetry/spans`.
* **Zero-Hallucination Grounding:** AI agents query compiler-verified facts via Model Context Protocol tools (`get_verified_context`, `get_blast_radius`).
* **Automated CI Gating:** Pull requests introducing unauthorized architectural drift or breaking changes are blocked automatically by `.github/workflows/garuda-gate.yml`.

---

[⬅️ Back to Main README](../README.md) · [📖 Architecture Documentation](ARCHITECTURE.md) · [📊 GAP-20 Benchmarks](../BENCHMARKS.md)
