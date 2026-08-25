# Implementation Plan: Event-sourced memory, belief store, admission gate

**Branch**: `task-0010-event-sourced-memory` | **Date**: 2026-08-25 | **Spec**: specs/010-event-sourced-memory/spec.md

## Summary

The mind's remember surface: an append-only event log (immutability in the types),
belief-store state as a reducer over it, RM-1..RM-7 as mechanical rules with named
tests, the deterministic admission gate of routing §6.3, and the
E6-input-tokens-per-villager-day instrument. Proven by the protocol §10.2 canonical
end-to-end against the existing in-test double.

## Technical Context

**Language/Version**: Go (module `kithcraft/mind`, go1.26.4 — the M1 module).
**Primary Dependencies**: Go stdlib only. Durability lands as an append-only log
file (JSON lines through the existing canonical writer); no SQLite dependency —
decision-0003's trigger suggestion is a pattern, and an unexported-struct +
append-only-writer design gives the same schema-not-convention guarantee at the type
level, which is what the card's AC #1 actually demands.
**Storage**: Append-only per-villager log + in-memory reduced state; replay on open.
**Testing**: `go test`; the §10.2 end-to-end drives the store through `seamtest`.
**Constraints**: SI-1 (no world-read path), PM-1 (private map), minds-are-others
(no external write path — package API exposes ingest-side entry points only).
**Scale/Scope**: 3–6 villagers, ~80 admitted memories/villager-day (A-2 ballpark).

## Constitution Check

The constitution at `.specify/memory/constitution.md` is an **unfilled template**
(stated plainly; house precedent specs 004–009). Checked against grounding docs:

- body-protocol-v0.md §6.4: RM-1..RM-7 implemented as written, each named in a test;
  the §2.7 classifier is a pure function of `origin` — PASS.
- Routing §6.3/§1.2: the admission gate is deterministic, no model call — PASS.
- decision-0003: reimplement-to-contract (~400 lines), port nothing — PASS.
- One-task-one-PR; no M4/M5/M7 scope — PASS (spec FR-007).

**Post-design re-check**: structure below adds no scope beyond FR-001..FR-007 — PASS.

## Project Structure

### Documentation (this feature)

```text
specs/010-event-sourced-memory/
├── README.md  spec.md  plan.md  tasks.md  (this cycle)
└── (no research.md — the contract and routing rules are ratified inputs)
```

### Source Code (repository root)

```text
mind/
├── memory/
│   ├── log.go            # append-only event log: unexported event struct, append+replay only
│   ├── log_test.go       # immutability-at-type-level proof; reducer replay test
│   ├── beliefs.go        # belief store: reducer over the log; PM-1 private map
│   ├── beliefs_test.go   # RM-1..RM-7 named tests
│   ├── provenance.go     # §2.7 classifier; RM-2/RM-3 citation gate (coerce, count)
│   ├── admission.go      # §6.3 deterministic gate + first-sighting tracking
│   ├── admission_test.go # every admit rule + the drop rule
│   └── instrument.go     # E6-input-tokens / buffer-size per villager-day
└── memory/e2e_test.go    # protocol §10.2 canonical end-to-end via seamtest double
```

**Structure Decision**: one new package `mind/memory`, consumed by later M-tasks.
The belief store's write path is unexported behind ingest-facing functions
(`Ingest(percept)`, `Consolidate(...)` reserved for M7) — minds-are-others enforced
by package visibility: nothing outside `memory` can construct or apply a raw belief
write. The admission gate sits in the same package because its "first sighting"
rule reads belief state.

## Complexity Tracking

No violations. Deliberate simplifications: log durability is a flat JSONL file per
villager (ponytail: ceiling is single-process access and evening-scale volume;
upgrade to SQLite with no_update/no_delete triggers when multi-process access or
crash-consistency demands it — the reducer interface hides the swap). Confidence
half-life constants are mind configuration with demo defaults, not tuned values.
