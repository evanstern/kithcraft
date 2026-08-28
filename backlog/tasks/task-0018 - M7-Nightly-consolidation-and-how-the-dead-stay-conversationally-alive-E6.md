---
id: TASK-0018
title: 'M7 - Nightly consolidation, and how the dead stay conversationally alive (E6)'
status: In Progress
assignee: []
created_date: '2026-08-21 23:39'
updated_date: '2026-08-28 19:24'
labels:
  - mind
  - m-0-build
milestone: m-0
dependencies:
  - TASK-0010
  - TASK-0011
documentation:
  - docs/design/demo-build-plan.md
  - docs/design/llm-routing-and-budget.md
  - docs/design/death-mechanics.md
priority: high
ordinal: 18000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
As a villager, I want to wake up having kept what mattered about yesterday — including who is no longer here — so that a day with the player accumulates into a history instead of evaporating.

**Scope boundary.** E6 on **Opus 5**, triggered by the sleep event and timed against `world_time`, **never a wall clock** (harness T-b). Port the machinery *shape* and its measured lessons: the nightly ledger, the ordinal `m1..mN` prompt convention (memories have no stable IDs and slice indexes are unstable, so the prompt convention *is* the identity mechanism, with accepted references mapped back to durable `(tick, hash)` pairs), and the rule that **a transport failure lands no marker** — the night is retried, not lost, and a villager waking with yesterday undigested is recoverable and invisible. Runs inside the ~5.8-minute sleep window; **not** on the Batch API, where a digest arriving after the villager has woken is a correctness problem rather than a latency one. v1 has **no formativeness scoring pass** — the admission gate decides eligibility, E6 decides what mattered. The death carry (death mechanics section 3): a recent death surfaces disproportionately in dusk conversation, then **fades in retrieval frequency** — never silently deleted (RM-7). Implements ruling **R-9**: a dead villager's mind is **archived, not terminated** — it opens no new session and its body token is retired and never reissued, but its durable log survives, because survivors' memories cite it and "stories told about them" is the texture permadeath exists to produce.

**Done proves.** Against the fake vendor: a scripted day's admitted buffer consolidates into a digest whose references resolve back to `(tick, hash)` pairs. A consolidation failed mid-call leaves **no marker** and is retried on the next attempt. A witnessed death is retrieved at high frequency in the following cycle's conversation context and at lower frequency two cycles later, and is still present — not deleted — well after that. A dead villager's log is readable; no session opens for it.

**Depends on.** M2, M4.

**References.** docs/design/demo-build-plan.md section 3.2 (M7) and its ruling R-9 are the plan of record. Ratified surfaces consumed: decision-0003 + docs/design/llm-routing-and-budget.md (E6 on Opus 5, the sleep-window trigger, no formativeness scoring pass in v1, the no-marker-on-failure rule, harness T-b), docs/design/death-mechanics.md (section 3 memory carry, token discipline), docs/design/body-protocol-v0.md (RM-7: time alone never deletes a fact), docs/design/kithcraft-brief.md (#4 stories told about them).

**Suggested tier: `sonnet` (next sweep's runbook decides).**

Spec: specs/018-consolidation
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 E6 runs on Opus 5, triggered by the sleep event and timed against world_time rather than a wall clock, inside the sleep window and not on the Batch API
- [ ] #2 A scripted day's admitted buffer consolidates into a digest whose references resolve back to durable (tick, hash) pairs via the ordinal m1..mN prompt convention
- [ ] #3 A consolidation that fails mid-call lands no marker and is retried on the next attempt
- [ ] #4 v1 runs no formativeness scoring pass: the admission gate decides eligibility and E6 decides what mattered
- [ ] #5 A witnessed death is retrieved at high frequency in the following cycle's conversation context, at lower frequency two cycles later, and is still present rather than deleted well after that
- [ ] #6 Ruling R-9 holds: a dead villager's mind is archived not terminated, its log is readable, no session opens for it, and its body token is retired and never reissued
- [ ] #7 Spec phase: Phase 1 — The nightly digest (US1 + US2)
- [ ] #8 Spec phase: Phase 2 — The death carry (US3)
- [ ] #9 Spec phase: Phase 3 — Archived, not terminated (US4)
- [ ] #10 Spec phase: Phase 4 — Gates and closure
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
claimed by sweep-0007-0022 orchestrator 2026-08-28 (lane 4); spec 018 stub + link ride this claim commit

tier: sonnet (default) · model cc/claude-sonnet-5[1m] · rubric: consolidation ports the measured machinery shape (ordinal convention, no-marker-on-failure); R-9 is ruled (runbook lane 4)
<!-- SECTION:NOTES:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Tests pass
- [ ] #2 Docs and wiki are updated and pass freshness tests
- [ ] #3 Spec and Backlog are in sync
<!-- DOD:END -->
