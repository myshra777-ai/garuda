# GARUDA AUTO-SYNTHESIZED BASE RULEBOOK (DRAFT FOR HUMAN APPROVAL)

- **Workspace**: `default`
- **Status**: `APPROVED_BASE_RULEBOOK`
- **Generated At**: `2026-10-29T07:18:01Z`

## 1. Surface Criticality Tiers
- **[HIGH_SEVERE] Core Homepage, Search Bar & Search Results Redirect**: Mission-critical (142 callers). Freeze against automated AI edits.
- **[MEDIUM] Non-Blocking Features (Wishlist Loader, Promo Badges)**: 19 callers. Trade-offs permitted with human review.
- **[LOW] Minor UI Box Alignment & Telemetry Helpers**: 4 callers. Safe for routine changes.

## 2. Multi-Team Formula Weights (`Loss Ratio`)
- **Ops Team (`ops.CalculateLossRatio`)**: Weight `0.70` (LOW severity, least architectural harm).
- **Marketing Team (`marketing.ComputeCampaignLossRatio`)**: Weight `0.10` (MEDIUM severity).
- **Management Team (`management.EvaluateEnterpriseLossRatioGate`)**: Weight `0.10` (HIGH_SEVERE — frozen).

## 3. Financial & Data Audit Invariants
- **Crore Normalization (`SYNTH-AUDIT-005`)**: All Claims Paid aggregations must normalize raw INR by `1e7` (preventing 48 Crore vs 4.8 Crore 10x drift).

## 4. Discovered AST Hubs in Workspace `default` (4 Repos, 3774 Entities, 11511 Claims)
- **[HIGH_SEVERE] `sqlmodel.main.SQLModel` (class)**: `171` inbound callers in `/home/rohit/garuda-validation/py/sqlmodel/sqlmodel/main.py`.
- **[HIGH_SEVERE] `.Pool` (method)**: `68` inbound callers in `/home/rohit/garuda/internal/store/postgres.go`.
- **[HIGH_SEVERE] `.Command` (struct)**: `49` inbound callers in `/home/rohit/garuda-validation/go/cobra/command.go`.
- **[HIGH_SEVERE] `.Valid` (method)**: `40` inbound callers in `/home/rohit/garuda/internal/api/generated.go`.
- **[HIGH_SEVERE] `.Name` (method)**: `36` inbound callers in `/home/rohit/garuda-validation/go/cobra/command.go`.

- **Approved By**: `Rohit Mishra` at `2026-10-29T07:18:01Z`
- **RFC 6962 Merkle Leaf Hash**: `cb60d3f1d2d8fcaed9818071c899e12fae99a0270af041a0fc426689fe39c53a`
