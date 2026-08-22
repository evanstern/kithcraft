# Implementation Plan: Seam transport decision and wire pinning

**Branch**: `task-0007-seam-transport` | **Date**: 2026-08-21 | **Spec**: specs/007-seam-transport/spec.md

**Input**: Feature specification from `specs/007-seam-transport/spec.md`

## Summary

Close protocol open question Q-1: choose the mind↔vendor wire among {UDS, TCP, stdio},
answering body-protocol-v0 §8's T-1..T-7 one by one; pin the choice with a
framing/serialization spec (`docs/design/`, spec-002 successor) and language-neutral
golden vectors that a trivial Go encoder and a trivial Java decoder each round-trip.
The transport *choice itself* is the task's design work — this plan structures the work
and deliberately does not pre-decide it (the sweep runbook escalates implementation to
the opus tier for exactly that judgment; ratification is the operator's at PR review).

## Technical Context

**Language/Version**: Deliverables are documents + fixtures; round-trip harnesses in
Go (per decision-0003, the daemon language) and Java (per decision-0001, the mod
language). Go toolchain and JDK versions are recorded by the implementer at
harness-creation time (first code in the repo — no versions exist yet to match).

**Primary Dependencies**: None beyond each language's stdlib for the trivial
encoder/decoder. Adding a serialization library is a decision the implementer must
justify inside the decision record, not an assumption.

**Storage**: Fixture files in-repo (language-neutral; location chosen in Phase 1 —
see data-model.md).

**Testing**: `go test` for the Go round-trip; the Java round-trip runs standalone
(single-file, `java` invocation) unless the implementer records a reason to introduce
the Gradle scaffold early — V1 owns the mod's build, and this task must not preempt it.

**Target Platform**: The demo host (macOS dev machine); nothing platform-specific may
enter the *contract* — UDS availability on the deployment platform is a T-criterion
input, not an output.

**Project Type**: Contract/decision task — documents, fixtures, two minimal test
harnesses.

**Performance Goals**: None binding at this layer. Latency ceilings (E4 < 3 s first
token) live above the transport; the decision record must only show the chosen wire
does not threaten them.

**Constraints**: Real wires only (decision-0003, one-way). T-1..T-7 verbatim from
protocol §8. No engine-native leak in spec or vectors (§12 discipline). Nothing built
beyond the five artifacts (spec FR-007).

**Scale/Scope**: 3–6 sessions (one per villager), one host, message rates per routing
doc's cadence model — small; the constraint is correctness and restart independence
(T-4), not throughput.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

The constitution at `.specify/memory/constitution.md` is an **unfilled template** —
stated plainly per house precedent (specs 004–006 did the same). This plan is checked
against the grounding docs instead:

- `docs/design/kithcraft-brief.md` (ratified): no player-facing surface here; the seam
  serves the minds-are-others constraint — PASS (the wire choice must keep SI-1
  structurally enforceable; T-1 encodes this).
- `docs/design/body-protocol-v0.md` (accepted): T-1..T-7 are the gate — the decision
  record answers each explicitly; the framing spec may not alter message semantics —
  PASS by construction.
- decision-0003 + `llm-routing-and-budget.md`: real wires only; the decomposition
  splits at the seam — this task builds no daemon and no mod — PASS (spec FR-007).
- Project CLAUDE.md / one-task-one-PR: five artifacts, one PR, operator ratifies the
  decision at review — PASS.

**Post-Phase-1 re-check (2026-08-21):** design artifacts (research.md, data-model.md,
contracts/, quickstart.md) introduce no new surface beyond the five artifacts and
pre-decide no T-criterion — still PASS.

## Project Structure

### Documentation (this feature)

```text
specs/007-seam-transport/
├── README.md            # claim stub (superseded by these artifacts)
├── spec.md
├── plan.md              # this file
├── research.md          # Phase 0: candidate profiles, T-matrix skeleton
├── data-model.md        # Phase 1: artifact/entity shapes
├── quickstart.md        # Phase 1: how to run both round-trips
├── contracts/
│   └── vectors.md       # Phase 1: the vector-set contract (what must exist)
├── checklists/requirements.md
└── tasks.md             # Phase 2 (/speckit-tasks)
```

### Source Code (repository root)

```text
docs/design/
└── seam-wire-v0.md          # the framing/serialization spec (spec-002 successor)

backlog/decisions/           # decision record via `backlog decision create` (CLI-owned)

seam/
├── vectors/                 # golden vectors: language-neutral fixture files
│   └── *.json + *.bin as the framing spec dictates
├── go-roundtrip/            # trivial Go encoder + test (own tiny go.mod; M1 replaces)
└── java-roundtrip/          # trivial Java decoder + test (single-file, no Gradle)
```

**Structure Decision**: A top-level `seam/` directory holds what both sides consume —
the vectors are the contract's shared physical artifact, so they live at the root
rather than inside either implementation's future tree (`mind/`, `mod/` — neither
exists yet). The round-trip harnesses sit beside the vectors they prove and are
explicitly throwaway-grade: M1/V1 build the real transport layers against the same
fixtures.

## Complexity Tracking

No constitution violations to justify (constitution unfilled; grounding-doc gates
pass). One deliberate simplification: the Java harness avoids Gradle — V1 owns the
build scaffold, and a second build system in this PR would be scope creep the spec's
FR-007 forbids.
