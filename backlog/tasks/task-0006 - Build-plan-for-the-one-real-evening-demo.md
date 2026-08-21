---
id: TASK-0006
title: Build plan for the "one real evening" demo
status: In Progress
assignee: []
created_date: '2026-08-19 18:38'
updated_date: '2026-08-21 23:22'
labels:
  - planning
milestone: m-0
dependencies:
  - TASK-0001
  - TASK-0002
  - TASK-0003
  - TASK-0004
documentation:
  - docs/design/kithcraft-brief.md
priority: medium
ordinal: 6000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
As a player, I want to spend one real evening with three villagers — post a blueprint on the job board, build alongside the one who takes it, and overhear them talk about the day (and me) at dusk — so that a survival world finally has company in it.

Context: this task produces the build plan for that demo (a Spec Kit spec sliced into buildable deliverable tasks on the board, likely runnable as a pdlc:sweep), not the implementation. The demo per the brief: three villagers with names, generated personas and endogenous desires, schedules (wake/work/socialize/sleep), persistent memory; the diegetic job-board book; vanilla night danger making the player's walls and torches protect their friends. Requires the ratified decisions from TASK-0001 through TASK-0004. Design against the spell-breakers: tedious player interactions, required micromanagement, politeness-simulator offense-taking.

Spec: specs/006-evening-demo-build-plan
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A build plan exists that decomposes the demo into deliverable tasks on the board, each mapping to one PR
- [ ] #2 The plan covers all demo beats: personas/desires generation, schedules, persistent memory, job-board book, blueprint build, dusk conversation, night danger
- [ ] #3 The plan honors the two load-bearing constraints (loneliness-cure thesis, minds-are-others) and names the spell-breakers as design checks
- [ ] #4 Operator has signed off on the plan
- [ ] #5 Spec phase: Phase 1 — Decomposition
- [ ] #6 Spec phase: Phase 2 — Board landing & sign-off prep
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Tier: opus (operator-escalated at runbook sign-off 2026-08-21 — demo decomposition is planning judgment the spec does not settle). Model ID: cc/claude-opus-5[1m], fallback cc/claude-opus-4-8[1m]. Opus tier already verified serving in this sweep (TASK-0004 all phases).
<!-- SECTION:NOTES:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Tests pass
- [ ] #2 Docs and wiki are updated and pass freshness tests
- [ ] #3 Spec and Backlog are in sync
<!-- DOD:END -->
