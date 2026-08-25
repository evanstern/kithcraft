# Implementation Plan: Mind daemon skeleton

**Branch**: `task-0008-mind-daemon-skeleton` | **Date**: 2026-08-22 | **Spec**: specs/008-mind-daemon-skeleton/spec.md

## Summary

First real Go in the repo: the mind daemon's process skeleton. UDS listener +
length-prefixed canonical-JSON codec (per decision-0004 / seam-wire-v0.md, proven
against the 17 golden vectors), boundary decode in decode→validate→mutate order
(V-1..V-6), session lifecycle with honest continuity, percept ingest with
dedup/gap accounting, intent bookkeeping, and the in-test double that drives it all.

## Technical Context

**Language/Version**: Go (go1.26.4 verified on host during TASK-0007).
**Primary Dependencies**: Go stdlib only (`net` for UDS, hand-rolled canonical JSON
writer per the TASK-0007 finding that `encoding/json` violates C-7 on output;
`encoding/json` remains fine for *decoding*, which is tolerant by design).
**Storage**: None — in-memory session state only. M2 owns durable memory.
**Testing**: `go test`; the vector suite reads `../seam/vectors/` (path fixed at repo
root).
**Target Platform**: The demo host (macOS, UDS verified available).
**Project Type**: Long-running daemon binary + library packages.
**Constraints**: SI-1 — the daemon has no world-read path by construction; the vendor
port interface is declared in the consuming package, implementations (real UDS server,
in-test double) satisfy it.
**Scale/Scope**: 3–6 bodies over one vendor connection; correctness over throughput.

## Constitution Check

The constitution at `.specify/memory/constitution.md` is an **unfilled template**
(stated plainly, house precedent specs 004–007). Checked against grounding docs
instead:

- body-protocol-v0.md: V-1..V-6, SI-1..SI-5, §6 session lifecycle — the spec maps
  each to a requirement; no protocol semantics altered — PASS.
- decision-0004 + seam-wire-v0.md: wire and framing consumed as fixed inputs; the
  narrowing-effects section (mind listens, unlink-before-bind, fail-closed
  negotiation) is implemented, not re-decided — PASS.
- One-task-one-PR; nothing from M2/M4/M5 scope — PASS (spec FR-008).

**Post-design re-check**: structure below adds no scope beyond FR-001..FR-008 — PASS.

## Project Structure

### Documentation (this feature)

```text
specs/008-mind-daemon-skeleton/
├── README.md  spec.md  plan.md  tasks.md  (this cycle)
└── (no research.md — no plan-level unknowns: wire, framing, and vectors are ratified
    inputs; data-model/contracts live in body-protocol-v0.md and seam-wire-v0.md,
    which this task implements rather than re-derives)
```

### Source Code (repository root)

```text
mind/
├── go.mod                    # module kithcraft/mind
├── cmd/minddaemon/main.go    # binary: flags (socket path), listener loop, shutdown
├── wire/                     # framing + canonical codec (seam-wire-v0.md)
│   ├── frame.go              # length-prefix read/write, caps, connection-fatal errors
│   ├── canonical.go          # hand-rolled canonical JSON writer (C-1..C-10)
│   ├── decode.go             # tolerant reader: presence-checked decode (V-1..V-6 half)
│   └── vectors_test.go       # all 17 golden vectors round-trip through THIS codec
├── seam/                     # protocol layer
│   ├── port.go               # the vendor port interface — declared at the consumer
│   ├── session.go            # lifecycle, continuity, manifest ingest, negotiation
│   ├── ingest.go             # validate → mutate; dedup; seq-gap shed accounting
│   ├── intents.go            # pending set, supersedes, act_result matching, cancel
│   └── *_test.go             # V-1..V-6 named tests, restart/continuity test
└── seamtest/
    └── double.go             # minimal in-test double (dials, scripts percepts,
                              # records intents) — S2 grows this into FakeVendor
```

**Structure Decision**: `mind/` is the daemon's module root per decision-0003 (own
module, outside Gradle). `wire` vs `seam` split mirrors the two governing documents
(seam-wire-v0.md vs body-protocol-v0.md): framing errors are connection-fatal, message
errors are session-level — different layers, different packages. The port interface
lives in `seam` (the consumer), not in the double's package — card AC #1's
declared-at-the-consumer rule, which is what lets S2's FakeVendor and the real vendor
both satisfy it without the daemon importing either.

## Complexity Tracking

No violations. Deliberate simplifications: in-memory state only (M2 owns durability);
the double scripts percepts but has no world model (S2's job); single vendor
connection assumed (the wire spec's model).
