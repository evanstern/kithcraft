---
id: TASK-0003
title: 'Decide entity implementation: custom entity vs augmented vanilla villager'
status: To Do
assignee: []
created_date: '2026-08-19 18:37'
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
Ratified: villager-shaped, smarter — riding the existing village fiction (beds, workstations, schedules). Decide whether Kithcraft villagers are custom entities or augmented vanilla villagers (e.g. via Fabric's villager brain API: activity/schedule injection, points of interest, memory modules). Depends on the mod stack choice (TASK-0001). Weigh: schedule substrate reuse, control over behavior, mob interactions (night danger must threaten villagers for base-building to be emotionally load-bearing), permadeath, rendering/skin flexibility, and multiplayer visibility (a friend dropping in must see them as real entities).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Custom entity vs augmented vanilla villager trade-offs are written up against the ratified constraints (village fiction reuse, real night danger, permadeath, drop-in multiplayer)
- [ ] #2 A recommendation with rationale is recorded as a Backlog decision record and ratified by the operator
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Tests pass
- [ ] #2 Docs and wiki are updated and pass freshness tests
- [ ] #3 Spec and Backlog are in sync
<!-- DOD:END -->
