---
id: TASK-0006
title: Build plan for the "one real evening" demo
status: To Do
assignee: []
created_date: '2026-08-19 18:38'
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
Once the mod stack, body protocol, entity implementation, and mind daemon decisions are ratified, author the build plan for the v1 demo defined in the brief: three villagers on a survival server with the player — names, generated personas and endogenous desires, schedules (wake/work/socialize/sleep), persistent memory; player posts a simple blueprint on the diegetic job-board book; one villager builds it while the player builds alongside; at dusk the villagers talk to each other about the day, the work, and the player. This task produces the plan (a Spec Kit spec sliced into buildable deliverable tasks on the board, likely runnable as a pdlc:sweep), not the implementation. Design against the spell-breakers: tedious player interactions, required micromanagement, politeness-simulator offense-taking.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A build plan exists that decomposes the demo into deliverable tasks on the board, each mapping to one PR
- [ ] #2 The plan covers all demo beats: personas/desires generation, schedules, persistent memory, job-board book, blueprint build, dusk conversation, night danger
- [ ] #3 The plan honors the two load-bearing constraints (loneliness-cure thesis, minds-are-others) and names the spell-breakers as design checks
- [ ] #4 Operator has signed off on the plan
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Tests pass
- [ ] #2 Docs and wiki are updated and pass freshness tests
- [ ] #3 Spec and Backlog are in sync
<!-- DOD:END -->
