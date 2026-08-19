---
id: TASK-0004
title: Decide mind daemon language and LLM routing/budget for a small cast at 1x
status: To Do
assignee: []
created_date: '2026-08-19 18:37'
labels:
  - design-decision
  - architecture
milestone: m-0
dependencies:
  - TASK-0002
documentation:
  - docs/design/kithcraft-brief.md
priority: medium
ordinal: 4000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Two coupled open questions from the brief. (a) Mind daemon language: whether promptworld I's Go daemon survives behind the body-protocol seam or the mind layer is rebuilt fresh — an implementation detail behind the seam, but one that must be decided to start building. (b) LLM routing and budget for 3–6 villagers at 1x real time: which calls go to which model tier, expected call cadence (the reflex/planner split means scripted competence for doing, LLM only for choosing and relating), cost envelope for an evening of play, and latency posture (a villager taking 20s to decide is a person mulling — real time deletes the governor/speed-ladder machinery). Doctrine reference: promptworld I docs/wiki (event-sourced memory, reflex/planner split, salience+consolidation, persona firewall).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A decision on mind daemon language / reuse-vs-rebuild is recorded as a Backlog decision record with rationale
- [ ] #2 An LLM routing sketch exists: which villager cognition events call an LLM, at what tier, at what expected cadence for a 3-villager evening
- [ ] #3 A cost envelope estimate for the one-real-evening demo is written down with its assumptions
- [ ] #4 Operator has ratified the decisions
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Tests pass
- [ ] #2 Docs and wiki are updated and pass freshness tests
- [ ] #3 Spec and Backlog are in sync
<!-- DOD:END -->
