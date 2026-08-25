# Tasks: Event-sourced memory, belief store, admission gate

**Input**: specs/010-event-sourced-memory/ (spec.md, plan.md)
**Prerequisites**: mind/ module (M1, merged), body-protocol-v0.md §2.6–2.7/§6.4/§10.2,
llm-routing-and-budget.md §6.3

**Organization**: Phases map 1:1 to phase-scoped dispatches.

## Phase 1 — The log and the reducer (US1)

**Goal**: Append-only event log with type-level immutability; state as a reducer.

**Independent test**: `go test ./memory/...` — immutability and replay tests green.

- [x] T001 Create mind/memory/log.go: memory-event shape (unexported fields, no
      setters), append-only writer over the canonical JSON encoder, JSONL file per
      villager, replay reader
- [x] T002 Named test proving mutation is impossible at the type level (card AC #1)
      and that replaying the log reproduces reduced state byte-for-byte
- [x] T003 Wire durable identity as `(world_time, hash)` pairs (the inherited E6
      identity convention M7 will address memories by)

**Checkpoint**: the log cannot lie and cannot be edited.

## Phase 2 — Beliefs, provenance, the RM rules (US2)

**Goal**: The private belief store with RM-1..RM-7 mechanical and named.

**Independent test**: `go test ./memory/...` — one named test per RM rule.

- [x] T004 Implement mind/memory/provenance.go: DIRECT_ORIGINS classifier (§2.7,
      pure function of origin), RM-2/RM-3 citation-resolution gate — coerce
      witnessed→told→inferred, never reject, count coercions
- [x] T005 Implement mind/memory/beliefs.go: reducer over the log; PM-1 private
      provenance-stamped map; RM-4 upsert rule (secondhand never beats fresher
      firsthand); RM-5/RM-6 read-time confidence and freshness as world_time
      arithmetic (observed_at null = maximally stale); RM-7 deletion only via
      correction / death / witnessed removal
- [x] T006 Named tests RM-1 through RM-7 (card AC #3), plus the AC #2 test that the
      store is distinct from any vendor index and AC #7's no-external-write-path
      check (package API surface assertion)

**Checkpoint**: every remember-surface rule is a test, not a comment.

## Phase 3 — Admission gate, instrument, end-to-end (US3)

**Goal**: The §6.3 gate and the instrument, proven by the canonical end-to-end.

**Independent test**: the §10.2 end-to-end named test passes, step 5 included.

- [ ] T007 Implement mind/memory/admission.go: admit on urgency ≥ notable / other
      body or player involved / act_result with authored reason / told_fact or text /
      first sighting of a kind or place; drop repeated background sightings of known
      things (card AC #4); deterministic — no model hook anywhere
- [ ] T008 Implement mind/memory/instrument.go: admitted buffer size per
      villager-day (card AC #6), reported at session end
- [ ] T009 Admission-gate named tests: one per admit rule, the drop rule, and a
      determinism check (same stream → same admissions)
- [ ] T010 The canonical end-to-end (protocol §10.2) against the seamtest double,
      including step 5: told-about-orchard cannot durably claim saw (card AC #5)

## Phase 4 — Closure: gates, wiki, board

- [ ] T011 go vet ./... + go test ./... green across the module; scope check: diff
      touches only mind/, specs/010-*, board files, runbook log row
- [ ] T012 Wiki: re-verify notes whose sources this PR touches (body-protocol-seam
      lists mind/ sources; overview's daemon description) — amend honestly, re-pin,
      regenerate CAPSULES.md if descriptions changed
- [ ] T013 Tick this file, check card ACs now true (backlog CLI in-worktree), append
      phase-done note

## Dependencies

Phase 1 → 2 → 3 → 4 serial (log before beliefs; beliefs before gate's
first-sighting rule; everything before e2e).

## Implementation strategy

Log-first: type-level immutability is the foundation claim; every later rule is a
reducer or a read-time function over an unfalsifiable record.
