# Implementation Plan: Body-vendor conformance — percepts out, intents in

**Branch**: `task-0012-vendor-conformance` | **Date**: 2026-08-25 | **Spec**: specs/012-vendor-conformance/spec.md

## Summary

The Fabric mod becomes a conformant body vendor: percept emission (four-type floor +
declared extras) with provenance stamped at emission and the §4.10 delivery
restriction, the act surface's four core verbs with the intent/ack/act_result split
and §5.6 target resolution, R-8's hearing-hook verification, and §12's leak passes
over captured payloads.

## Technical Context

**Language/Version**: Java 21 / Gradle + Fabric Loom (the V1 toolchain: MC 26.2,
Loader 0.19.3, API 0.158.0).
**Primary Dependencies**: What V1 established — Gson for parse, the custom canonical
writer, WireClient, TokenRegistry. Gradle test stack as in V1 (JUnit).
**Storage**: TokenRegistry (V1's SavedData) extended only as token issuance demands.
**Testing**: Unit tests for shape/stamping/restriction logic; dev-server observation
records for the live proofs (documented `gradle runServer` observations where no
automated harness exists — the runbook's V-task gate line).
**Constraints**: decision-0002's bounded Mixin surface — percept hooks prefer plain
API (events, `Brain<E>` reads) over new Mixins; any Mixin addition is enumerated in
the PR. AR-1..AR-6 govern every emitted shape.
**Scale/Scope**: 3–6 villager bodies on one vendor connection.

## Constitution Check

The constitution at `.specify/memory/constitution.md` is an **unfilled template**
(stated plainly; house precedent specs 004–011). Checked against grounding docs:

- body-protocol-v0.md §2–§5: every emitted/consumed shape implements the contract
  verbatim; no protocol semantics altered — PASS.
- decision-0002 + entity-implementation-comparison.md: vanilla `Brain<E>` owns
  doing; the mod emits and executes, never plans; Mixin surface bounded — PASS.
- demo-build-plan R-8: verification inside this task, declare-or-unsupported — PASS.
- One-task-one-PR; V3/V4/V5 scope excluded (spec FR-009) — PASS.

**Post-design re-check**: structure below adds no scope beyond FR-001..FR-009 — PASS.

## Project Structure

### Documentation (this feature)

```text
specs/012-vendor-conformance/
├── README.md  spec.md  plan.md  tasks.md  (this cycle)
└── research/
    └── r8-hearing-hook.md   # R-8 verification record (hook found or not, evidence)
```

### Source Code (repository root)

```text
mod/src/main/java/dev/kithcraft/mod/
├── percept/
│   ├── PerceptEmitter.java     # composes+stamps percepts, urgency, seq, shedding
│   ├── Provenance.java         # closed origin vocabulary; stamped at emission
│   ├── Sightings.java          # sighting/observation sweeps; vocabulary scoping
│   ├── ChangeReports.java      # §4.10 delivery restriction (actor/witness exclusion)
│   └── SelfState.java          # felt-origin condition bands
├── act/
│   ├── IntentHandler.java      # decode → ack (receipt only) → execute → one act_result
│   ├── TargetResolution.java   # last-seen resolution; unknown_target vs target_gone
│   └── Verbs.java              # go_to / speak / attend / wait / cancel via Brain<E>/handlers
└── src/test/java/dev/kithcraft/mod/
    ├── percept/*Test.java      # stamping, vocabulary scope, restriction, no-salience-field
    ├── act/*Test.java          # ack-then-result, §5.6 cases, one-result-per-intent
    └── LeakPassTest.java       # §12's six passes over captured/constructed payloads
```

**Structure Decision**: two new packages mirroring the two halves of the surface;
`act_result` bridges them (emitted by `percept` machinery, triggered by `act`
completion) — which is why the card says the halves cannot be split. Verb execution
delegates to vanilla `Brain<E>`/mod handlers per decision-0002; this task adds no
new Mixins unless R-8's hook or a verb demands one, and enumerates any it adds.

## Complexity Tracking

No violations. Deliberate simplifications: percept sweeps use simple periodic
scanning with a dedup window (ponytail: ceiling is demo-scale cast and chunk range;
optimize sweep scheduling only if the dev server shows tick-budget pressure).
Capture-for-leak-pass is a test-side payload recorder, not a production tap.
