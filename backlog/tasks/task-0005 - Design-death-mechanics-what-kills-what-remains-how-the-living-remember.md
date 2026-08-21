---
id: TASK-0005
title: 'Design death mechanics: what kills, what remains, how the living remember'
status: In Progress
assignee: []
created_date: '2026-08-19 18:37'
updated_date: '2026-08-21 21:38'
labels:
  - design-decision
  - gameplay
milestone: m-0
dependencies: []
documentation:
  - docs/design/kithcraft-brief.md
priority: medium
ordinal: 5000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
As a player, I want a villager's death to be permanent and to leave real traces — a grave, belongings, survivors who remember and tell stories — so that loss stings and the friends I keep alive matter more.

Context: ratified posture — permadeath is real; graves, persisting memories of the dead, and stories told about them are the intended texture (mechanics were deferred to this design pass). Resolve the brief's open question: what can kill a villager (night danger, hunger, falls?), what remains in the world (grave, belongings), and how surviving villagers remember and talk about the dead (memory entries, dusk conversation material). Constraint from the spell-breakers list: micromanagement required to keep villagers alive is a failure mode — death must be possible without turning play into babysitting. Replenishment is punted in v1 (no spawning/wanderers), so death permanently shrinks the cast; the design must account for that.

Spec: specs/005-death-mechanics
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A death-mechanics design doc exists covering causes of death, physical remains, and how survivors' memories and conversations carry the dead
- [ ] #2 The design addresses the micromanagement spell-breaker: villagers are not so fragile that keeping them alive becomes babysitting
- [ ] #3 The design states what a permanently shrinking cast means for the v1 demo and accepts or scopes it
- [x] #4 Operator has ratified the design
- [ ] #5 Spec phase: Phase 1 — Death surface evidence
- [ ] #6 Spec phase: Phase 2 — Design document
- [ ] #7 Spec phase: Phase 3 — Ratification prep
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Tier: sonnet (default tier — design fleshes out ratified posture, judgment gate is operator ratification at AC #4). Model ID: cc/claude-sonnet-5[1m]. Served model to be verified from first dispatch transcript.

Phase 1 done (dda2a19): death-surface evidence committed. Served model VERIFIED: cc/claude-sonnet-5[1m] (from implementer report, 2026-08-21). Key facts: no villager hunger death; zombie kill usually converts (50%/100%) rather than kills; death drops nothing (inventory lost), POIs released, murder gossip broadcast.
<!-- SECTION:NOTES:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Tests pass
- [ ] #2 Docs and wiki are updated and pass freshness tests
- [ ] #3 Spec and Backlog are in sync
<!-- DOD:END -->
