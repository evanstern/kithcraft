---
id: TASK-0008
title: >-
  M1 - Mind daemon skeleton: process, vendor port, boundary decode, session
  lifecycle
status: Done
assignee: []
created_date: '2026-08-21 23:36'
updated_date: '2026-08-25 19:10'
labels:
  - mind
  - m-0-build
milestone: m-0
dependencies:
  - TASK-0007
documentation:
  - docs/design/demo-build-plan.md
  - docs/design/body-protocol-v0.md
  - docs/design/llm-routing-and-budget.md
priority: high
ordinal: 8000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
As a future implementer, I want a daemon that can hold a session and refuse a malformed percept before it touches any state, so that every later mind component is built on a boundary that cannot be talked past.

**Scope boundary.** Its own Go module and `go.mod`, outside the mod's Gradle build. The vendor port as an interface **declared at the consumer**. The boundary-decode component decision-0003 schedules explicitly, in harness H-1's order: presence-checked decode -> validate -> mutate, never interleaved. V-1..V-6: ignore unknown fields; fall back on unknown enum values; retain but never interpret an unknown `percept_type`; refuse an unknown verb; **a missing required field is malformed, not defaulted**; unrecognized or absent `origin` classifies secondhand. Session lifecycle: `session_open` handshake and manifest ingest, continuity handling (`body_continuous`, and a gap reported as a gap — never backfilled), `session_close`. Percept ingest: dedup by `percept_id`, `seq` gap detection. Intent bookkeeping: the pending set, `supersedes`, matching `act_result` to `intent_id`, `cancel`. A minimal in-test double sufficient to drive this; S2 grows it into the conformance harness.

**Done proves.** The daemon binary starts, opens a session against the double, ingests a scripted percept stream with duplicates and a `seq` gap, and emits intents. A percept missing `provenance` mutates nothing. An `origin` value from a future minor version classifies secondhand. Restarting the daemon mid-session re-opens with continuity and does not invent what happened in the gap.

**Depends on.** S1 (the wire and its golden vectors).

**References.** docs/design/demo-build-plan.md section 3.2 (M1) is the plan of record. Ratified surfaces consumed: docs/design/body-protocol-v0.md (SI-1..SI-5, V-1..V-6, session lifecycle, intent/ack split), decision-0003 + docs/design/llm-routing-and-budget.md (Go daemon rebuilt behind the seam; boundary decode scheduled explicitly; the decomposition splits at the seam).

**Suggested tier: `sonnet` (next sweep's runbook decides).**

Spec: specs/008-mind-daemon-skeleton
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The daemon is its own Go module with its own go.mod, outside the mod's Gradle build, and the vendor port is an interface declared at the consumer
- [x] #2 Boundary decode runs presence-checked decode, then validate, then mutate, never interleaved, and satisfies V-1..V-6
- [x] #3 A percept missing a required field (e.g. provenance) is treated as malformed and mutates no state
- [x] #4 An unrecognized or absent origin value classifies the percept as secondhand
- [x] #5 The daemon opens a session against a test double, ingests a scripted percept stream with duplicates and a seq gap, and emits intents
- [x] #6 Restarting the daemon mid-session re-opens with continuity and reports the gap as a gap rather than backfilling it
- [x] #7 Spec phase: Phase 1 — Module, wire codec, vector proof (US1 groundwork)
- [x] #8 Spec phase: Phase 2 — Vendor port, listener, session lifecycle (US2)
- [x] #9 Spec phase: Phase 3 — Ingest, intents, the double, end-to-end (US3)
- [x] #10 Spec phase: Phase 4 — Closure: gates, wiki, board
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Claimed by sweep-0007-0022 orchestrator 2026-08-22 on branch task-0008-mind-daemon-skeleton (worktree .worktrees/task-0008). Tier: sonnet (default tier per runbook — greenfield against the written contract; boundary decode, V-1..V-6, session lifecycle all specified). Model ID: cc/claude-sonnet-5[1m]. Served model recorded at dispatch. Wire inputs now ratified: decision-0004 (UDS, mind listens / vendor dials), docs/design/seam-wire-v0.md, seam/vectors/ (17).

Phase 1 done (72b3c09, sonnet verified — claude-sonnet-5, ~138k subagent tokens): mind/ module exists (kithcraft/mind, go1.26.4); wire codec (frame.go 1MiB cap, canonical.go C-1..C-10 hand-rolled writer, decode.go tolerant V-1/V-2/V-3/V-5 half) proven against all 17 vectors — census + roundtrip + framing refusals + non-canonical acceptance, orchestrator re-ran go vet + go test green. C-4 dup-key detection deliberately skipped (no vector pins it; ponytail comment names the trigger).

Phase 2 done (6b2ee3d, sonnet verified — claude-sonnet-5, ~179k subagent tokens): seam.Conn port declared at the consumer with NewWireConn adapter; UDS listener with decision-0004 liveness-probe unlink; fail-closed negotiation, per-body multiplexing with byte-identical capabilities check; continuity by body token with named no-backfill restart test (previous_session not required, zero fabricated messages). Orchestrator re-ran vet+test green (3 packages). Deviations recorded: sun_path length workaround in listener tests; §1.3 last-open-wins cross-connection rule ponytailed out of phase scope.

Phase 3 done (4f6e819, sonnet verified — claude-sonnet-5, ~186k subagent tokens): ingest.go (validate→mutate, V-6 pure classifier, reconnect-scoped dedup, seq-gap shed accounting — extended to ALL vendor→mind messages sharing the body counter so an interleaved intent_ack isn't misattributed), intents.go (pending set, supersedes, act_result matching, cancel, V-4 verb refusal), seamtest double, both e2e tests (AC #5 scripted stream + AC #6 restart/no-backfill through the real listener). Orchestrator re-ran vet+test green incl. -race per agent. Card ACs #1-#6 all demonstrated by named tests — checked.

Phase 4 (closure, T013-T016) done: go vet + go test ./... green across mind/ (wire, seam, cmd/minddaemon, seamtest); scope check clean (diff stays within mind/, specs/008-*, backlog/, sweep runbook); body-protocol-seam.md and overview.md re-verified and amended (mind/ is no longer contract-only — first real implementation landed), sourced on mind/wire/canonical.go + mind/seam/session.go, re-pinned to f51b933 (verified resolvable); INDEX.md's stale pre-code framing corrected; CAPSULES.md regenerated. tasks.md T013-T016 ticked. Did not check card DoD items or change status per dispatch scope.

spec-bridge sync: Phase 1 — Module, wire codec, vector proof (US1 groundwork): 5/5 · Phase 2 — Vendor port, listener, session lifecycle (US2): 3/3 · Phase 3 — Ingest, intents, the double, end-to-end (US3): 4/4 · Phase 4 — Closure: gates, wiki, board: 4/4 — status In Progress → Done
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
All spec tasks complete (Phase 1 — Module, wire codec, vector proof (US1 groundwork): 5/5 · Phase 2 — Vendor port, listener, session lifecycle (US2): 3/3 · Phase 3 — Ingest, intents, the double, end-to-end (US3): 4/4 · Phase 4 — Closure: gates, wiki, board: 4/4). Derived Done by spec-bridge sync.
<!-- SECTION:FINAL_SUMMARY:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Tests pass
- [ ] #2 Docs and wiki are updated and pass freshness tests
- [ ] #3 Spec and Backlog are in sync
<!-- DOD:END -->
