---
id: TASK-0003
title: 'Decide entity implementation: custom entity vs augmented vanilla villager'
status: In Progress
assignee: []
created_date: '2026-08-19 18:37'
updated_date: '2026-08-21 19:49'
labels:
  - design-decision
milestone: m-0
dependencies:
  - TASK-0001
documentation:
  - docs/design/kithcraft-brief.md
priority: high
ordinal: 3000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
As a player, I want my villagers to be real inhabitants of the world — sleeping in beds, working at stations, threatened by the same night I am — so that protecting them with walls and torches feels like protecting friends, not managing props.

Context: ratified — villager-shaped, smarter, riding the existing village fiction (beds, workstations, schedules). Decide whether Kithcraft villagers are custom entities or augmented vanilla villagers (e.g. via Fabric's villager brain API: activity/schedule injection, points of interest, memory modules). Depends on the mod stack choice (TASK-0001). Weigh: schedule substrate reuse, control over behavior, mob interactions (night danger must threaten villagers for base-building to be emotionally load-bearing), permadeath, rendering/skin flexibility, and multiplayer visibility (a friend dropping in must see them as real entities).

Spec: specs/003-entity-implementation
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Custom entity vs augmented vanilla villager trade-offs are written up against the ratified constraints (village fiction reuse, real night danger, permadeath, drop-in multiplayer)
- [ ] #2 A recommendation with rationale is recorded as a Backlog decision record and ratified by the operator
- [ ] #3 Spec phase: Phase 1 — Engine behavior evidence
- [ ] #4 Spec phase: Phase 2 — Comparison document
- [ ] #5 Spec phase: Phase 3 — Recommendation & decision record
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Dispatch tier: sonnet (cc/claude-sonnet-5[1m], no fallback declared) — default tier: trade-off analysis against ratified constraints, same shape as TASK-0001's comparison which sonnet served (verified 2026-08-20). Served model to be re-verified from Phase 1 transcript.
<!-- SECTION:NOTES:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Tests pass
- [ ] #2 Docs and wiki are updated and pass freshness tests
- [ ] #3 Spec and Backlog are in sync
<!-- DOD:END -->
