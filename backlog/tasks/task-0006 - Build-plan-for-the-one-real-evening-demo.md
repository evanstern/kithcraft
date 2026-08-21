---
id: TASK-0006
title: Build plan for the "one real evening" demo
status: Done
assignee: []
created_date: '2026-08-19 18:38'
updated_date: '2026-08-21 23:50'
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
- [x] #1 A build plan exists that decomposes the demo into deliverable tasks on the board, each mapping to one PR
- [x] #2 The plan covers all demo beats: personas/desires generation, schedules, persistent memory, job-board book, blueprint build, dusk conversation, night danger
- [x] #3 The plan honors the two load-bearing constraints (loneliness-cure thesis, minds-are-others) and names the spell-breakers as design checks
- [x] #4 Operator has signed off on the plan
- [x] #5 Spec phase: Phase 1 — Decomposition
- [x] #6 Spec phase: Phase 2 — Board landing & sign-off prep
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Tier: opus (operator-escalated at runbook sign-off 2026-08-21 — demo decomposition is planning judgment the spec does not settle). Model ID: cc/claude-opus-5[1m], fallback cc/claude-opus-4-8[1m]. Opus tier already verified serving in this sweep (TASK-0004 all phases).

Phase 1 done (da7b43e, opus verified, ~159k tokens): docs/design/demo-build-plan.md — 16 tasks in 4 groups (seam S1-S2, mind M1-M7, vendor V1-V5, integration I1-I2), lanes 0-5 with S1 (transport) blocking everything, critical path 7 merges deep. Nine punted open items ruled in-plan (R-1..R-9). Tier floor sonnet ×16; proposed escalations for next sweep's sign-off: S1 → opus, V5 trigger-named. Spell-breakers attached as named checks on specific tasks.

Phase 2 done (a6bc694 + 1a8f547 + CAPSULES 5e4bce2, opus verified, ~137k tokens): 16 tasks created TASK-0007..0022, 1:1 with plan, milestone m-0, deps by real id, spaced ordinals; plan doc §8 coverage cross-check (all 9 beats trace); v1-demo re-verified + re-pinned to a6bc694. Card ACs 1-3 satisfied; AC 4 (operator sign-off) pending at PR review.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Build plan signed off (operator merged PR #10, merge 8ef4a5f, 2026-08-21). Deliverables: docs/design/demo-build-plan.md (16-task decomposition, lanes 0-5, coverage cross-check, R-1..R-9 rulings) and TASK-0007..0022 created on the board (milestone m-0, user stories, done-proves ACs, deps by id, spell-breaker checks, suggested tiers). v1-demo re-verified + re-pinned; CAPSULES regenerated. Escalation proposals for next sweep's runbook: TASK-0007 (transport) → opus; TASK-0019 conditional on its named trigger. Both phases served by verified cc/claude-opus-5[1m]. The next sweep's input is ready.
<!-- SECTION:FINAL_SUMMARY:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Tests pass
- [ ] #2 Docs and wiki are updated and pass freshness tests
- [ ] #3 Spec and Backlog are in sync
<!-- DOD:END -->
