---
id: TASK-0005
title: 'Design death mechanics: what kills, what remains, how the living remember'
status: To Do
assignee: []
created_date: '2026-08-19 18:37'
updated_date: '2026-08-19 18:47'
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
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A death-mechanics design doc exists covering causes of death, physical remains, and how survivors' memories and conversations carry the dead
- [ ] #2 The design addresses the micromanagement spell-breaker: villagers are not so fragile that keeping them alive becomes babysitting
- [ ] #3 The design states what a permanently shrinking cast means for the v1 demo and accepts or scopes it
- [ ] #4 Operator has ratified the design
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Tests pass
- [ ] #2 Docs and wiki are updated and pass freshness tests
- [ ] #3 Spec and Backlog are in sync
<!-- DOD:END -->
