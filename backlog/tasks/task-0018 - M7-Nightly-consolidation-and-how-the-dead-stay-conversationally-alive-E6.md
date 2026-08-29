---
id: TASK-0018
title: 'M7 - Nightly consolidation, and how the dead stay conversationally alive (E6)'
status: Done
assignee: []
created_date: '2026-08-21 23:39'
updated_date: '2026-08-29 02:03'
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
- [x] #1 E6 runs on Opus 5, triggered by the sleep event and timed against world_time rather than a wall clock, inside the sleep window and not on the Batch API
- [x] #2 A scripted day's admitted buffer consolidates into a digest whose references resolve back to durable (tick, hash) pairs via the ordinal m1..mN prompt convention
- [x] #3 A consolidation that fails mid-call lands no marker and is retried on the next attempt
- [x] #4 v1 runs no formativeness scoring pass: the admission gate decides eligibility and E6 decides what mattered
- [x] #5 A witnessed death is retrieved at high frequency in the following cycle's conversation context, at lower frequency two cycles later, and is still present rather than deleted well after that
- [x] #6 Ruling R-9 holds: a dead villager's mind is archived not terminated, its log is readable, no session opens for it, and its body token is retired and never reissued
- [x] #7 Spec phase: Phase 1 — The nightly digest (US1 + US2)
- [x] #8 Spec phase: Phase 2 — The death carry (US3)
- [x] #9 Spec phase: Phase 3 — Archived, not terminated (US4)
- [x] #10 Spec phase: Phase 4 — Gates and closure
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
claimed by sweep-0007-0022 orchestrator 2026-08-28 (lane 4); spec 018 stub + link ride this claim commit

tier: sonnet (default) · model cc/claude-sonnet-5[1m] · rubric: consolidation ports the measured machinery shape (ordinal convention, no-marker-on-failure); R-9 is ruled (runbook lane 4)

AC#1 (E6/Opus5/sleep-event/world_time/sleep-window/no-Batch-API): TestE6IsOpus5AndOffline + TestNoBatchAPIPath (mind/consolidate/consolidate_test.go).

AC#2 (m1..mN ordinal convention, references resolve to (tick,hash)): TestRunNight_OrdinalMappingRoundTrip + TestRunNight_ConsolidatedWindowExcludedNextNight (consolidate_test.go).

AC#3 (no-marker-on-failure, retried next attempt): TestRunNight_TransportFailureLandsNoMarker, _CancellationLandsNoMarker, _OverLimitLandsNoMarker (truncation proven at the client boundary by TestClientDigester_TruncationDetected), _EmptyNightLandsMarker, _MultiNightAccumulationAfterFailures (consolidate_test.go, client_digester_test.go).

AC#4 (no formativeness scoring pass): structural absence — no scoring code anywhere in mind/consolidate, documented at cycle.go's header; the admission gate (mind/memory.Gate) already decided eligibility before RunNight ever sees an event.

AC#5 (death-carry frequency: high next cycle, lower two cycles later, still present): TestDeathCarryWeight_SpikeNextCycle, _LowerTwoCyclesLater, _FloorsAtNormalPresence_NeverZero, _Decreasing (deathcarry_test.go) — deterministic distribution assertions on the curve itself, per plan.md's risk note. Honest caveat: this package exports SelectionWeights as the retrieval-weighting hook only; conversation-context assembly's actual consumption of it is M6/mind/converse (TASK-0017, sibling in-flight branch), landing at merge, not in this branch.

AC#6 (R-9: archived not terminated, log readable, no session, token retired never reissued): TestSessionOpen_ArchivedMind_RefusedOnFirstOpen, _RefusedOnMultiplex, _NotArchived_Unaffected (mind/seam/session_test.go); TestArchive_LandsAndPersists, _IdempotentByMindID, _UnaffectedMindStaysOpen, _DeathBeforeOwnConsolidation (mind/consolidate/archive_test.go).

AC#7-9 (spec phases 1-3): T001-T003, T004, T005 all landed and tested (see above).

AC#10 (spec phase 4, gates and closure): go vet + go test -count=1 ./... green from mind/ (all 9 packages); scope clean; wiki re-ground done (freshness gate green for all touched notes; see phase-4 commit).

Deferred, flagged for refactor-triage: daemon-level wiring of Ledger/Archive/RunNight into a real long-running mind daemon process (a scheduler invoking RunNight on the sleep event, Archive.IsArchived actually plumbed into a live Ingester at daemon startup) is explicitly out of scope for this task per plan.md — everything here is proven against the fake vendor and scripted/mocked clients, correct but not yet wired into cmd/minddaemon's runtime. Naming it here so it doesn't silently vanish before a sweep picks it up.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
M7 delivered via PR #23 (merge 5dd2c48, merge commit, pins preserved). mind/consolidate/: E6 nightly digest on Opus 5 per the registry — sleep-event trigger with all arithmetic on world_time (no wall clock, no Batch API path, both structurally checked); append-only JSONL Ledger with Watermark reducer; the m1..mN ordinal convention round-tripped to durable (world_time, hash) pairs; no-marker-on-failure across transport failure, cancellation, and over-limit (StopReasonMaxTokens — I's truncation lesson); empty night lands a marker; multi-night accumulation. Death carry: pure retrieval-frequency weighting (spike 10.0 citing I's salience band, halves per cycle, floors at 1.0 — RM-7 never deletes), exported SelectionWeights hook now consumable by mind/converse post-merge. R-9 archival: archived minds refuse session opens fail-closed, log stays readable, token retired never-reissued persisted. Deferred + named for refactor-triage: daemon-level wiring of Ledger/Archive/RunNight into cmd/minddaemon. Spec-bridge derivation: 4 phases 8/8, Done-eligible. ~747k subagent tokens across 4 sonnet dispatches (cc/claude-sonnet-5[1m], verified per dispatch).
<!-- SECTION:FINAL_SUMMARY:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 Tests pass
- [x] #2 Docs and wiki are updated and pass freshness tests
- [x] #3 Spec and Backlog are in sync
<!-- DOD:END -->
