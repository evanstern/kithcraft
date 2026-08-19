---
id: TASK-0001
title: 'Decide mod stack: Fabric vs Paper/Citizens vs hybrid'
status: To Do
assignee: []
created_date: '2026-08-19 18:36'
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
Kithcraft (promptworld II) is a Minecraft server mod giving the player LLM villagers as company (see docs/design/kithcraft-brief.md — ratified, do not relitigate). The first open question is the mod stack. Evaluate Fabric vs Paper/Citizens2 vs a hybrid against the ratified decisions: villager-shaped smarter NPCs (not bot clients), small cast, real-time only, world-agnostic body protocol with the mod as first body vendor. Re-verify prior art (Citizens2, CraftAgent, AI_NPC, Fabric villager brain API) for maintenance status, target MC version support, and licenses before relying on any of it — the brief's links were verified 2026-08-19 and the space moves fast.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Comparison of Fabric, Paper/Citizens2, and hybrid options is written up with evidence (versions, maintenance status, licenses re-verified, dated)
- [ ] #2 A recommendation with rationale is recorded as a Backlog decision record and ratified by the operator
- [ ] #3 The recommendation explicitly addresses the body-protocol seam (mod as swappable body vendor) and the villager-shaped-not-bot-client constraint
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Tests pass
- [ ] #2 Docs and wiki are updated and pass freshness tests
- [ ] #3 Spec and Backlog are in sync
<!-- DOD:END -->
