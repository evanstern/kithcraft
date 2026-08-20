---
id: TASK-0001
title: 'Decide mod stack: Fabric vs Paper/Citizens vs hybrid'
status: In Progress
assignee: []
created_date: '2026-08-19 18:36'
updated_date: '2026-08-20 14:50'
labels:
  - design-decision
milestone: m-0
dependencies: []
documentation:
  - docs/design/kithcraft-brief.md
priority: high
ordinal: 1000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
As the Kithcraft team, we want the mod stack chosen (Fabric vs Paper/Citizens2 vs hybrid) with evidence, so that every downstream decision — entity implementation, body vendor, build plan — rests on a ratified foundation instead of an assumption.

Context: Kithcraft (promptworld II) is a Minecraft server mod giving the player LLM villagers as company (see docs/design/kithcraft-brief.md — ratified, do not relitigate). Evaluate the options against the ratified decisions: villager-shaped smarter NPCs (not bot clients), small cast, real-time only, world-agnostic body protocol with the mod as first body vendor. Re-verify prior art (Citizens2, CraftAgent, AI_NPC, Fabric villager brain API) for maintenance status, target MC version support, and licenses before relying on any of it — the brief's links were verified 2026-08-19 and the space moves fast.

Spec: specs/001-mod-stack-decision
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Comparison of Fabric, Paper/Citizens2, and hybrid options is written up with evidence (versions, maintenance status, licenses re-verified, dated)
- [ ] #2 A recommendation with rationale is recorded as a Backlog decision record and ratified by the operator
- [x] #3 The recommendation explicitly addresses the body-protocol seam (mod as swappable body vendor) and the villager-shaped-not-bot-client constraint
- [x] #4 Spec phase: Phase 1 — Re-verify prior art
- [x] #5 Spec phase: Phase 2 — Comparison document
- [x] #6 Spec phase: Phase 3 — Recommendation & decision record
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Claimed by sweep (runbook docs/design/task-0001-mod-stack-runbook.md, signed-off 2026-08-19). Spec dir: specs/001-mod-stack-decision. Tier: sonnet (default) · model cc/claude-sonnet-5[1m] — rubric: research/comparison executed against a written spec; the judgment call (ratification) is an operator checkpoint. Served model to be verified from first dispatch transcript.

Phases 1-3 complete on branch (f99016d, 73b3ddd, c07a2ea), all served-model-verified cc/claude-sonnet-5[1m]. Comparison: docs/design/mod-stack-comparison.md. Decision record decision-0001 (Fabric server-side mod, PROPOSED). AC #2 awaits operator ratification.
<!-- SECTION:NOTES:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Tests pass
- [ ] #2 Docs and wiki are updated and pass freshness tests
- [ ] #3 Spec and Backlog are in sync
<!-- DOD:END -->
