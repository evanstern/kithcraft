# Implementation Plan: The augmented villager — brain, schedule, cast, dusk pair formation

**Branch**: `task-0014-augmented-villager` | **Date**: 2026-08-27 | **Spec**: specs/014-augmented-villager/spec.md

## Summary

The mod becomes a village: three named vanilla villagers (augmented, not custom
entities — decision-0002) run wake/work/socialize/sleep on the vanilla `Brain<E>`
substrate; a bounded Mixin surface (≤3 task-list overrides) suppresses breeding,
gossip, and golem summoning; a custom dusk Activity implements R-7's pair formation
with the ~10 s-ahead signal; and no scheduled behaviour ever blocks on the mind.

## Technical Context

**Language/Version**: Java 21 / Gradle + Fabric Loom (V1 toolchain: MC 26.2, Loader
0.19.3, API 0.158.0). MC 26.2 is UNOBFUSCATED (Mojang official names; Yarn
discontinued) — every symbol beyond villager-brain-api.md's verified memory-module
rows must be re-derived first (FR-001).
**Primary Dependencies**: What V1/V2 established — the mod tree under
`mod/src/main/java/dev/kithcraft/mod/`, WireClient, TokenRegistry, PerceptEmitter.
Fabric API events where they cover a need before any new Mixin.
**Storage**: Cast identity via the mod's existing SavedData pattern (V1's
TokenRegistry precedent); no new persistence design.
**Testing**: Unit tests for signal timing logic, override presence, cast seeding;
dev-server observation records for the live proofs (runbook V-task gate). The
full-cycle test is a documented `gradle runServer` observation with the criteria
checklist — no automated overnight harness exists yet and building one is out of
scope (thinnest thing that proves the card).
**Constraints**: decision-0002's Mixin budget: V3 owns at most three task-list
overrides; conversion-cancel is V5's. Body-keeps-moving is structural: the mind is
consulted via the seam's async surface only; no schedule code awaits a response.
**Scale/Scope**: exactly three villagers, one dev server.

## Constitution Check

The constitution at `.specify/memory/constitution.md` is an **unfilled template**
(stated plainly; house precedent specs 004–012). Checked against grounding docs:

- decision-0002 + entity-implementation-comparison.md: augmented vanilla villager,
  schedules/POI/pathing inherited free, bounded ~4-injection Mixin budget of which
  V3 takes ≤3 — PASS.
- demo-build-plan R-7: pairing signal ~10 s ahead of arrival, consumed by M6 —
  implemented as specified, no early-fire redesign — PASS.
- kithcraft-brief micromanagement spell-breaker: the full-cycle-unattended proof is
  the check, run as specified — PASS.
- body-protocol-v0: no protocol extension; the pairing signal reaches the mind via
  existing percept types (sighting/self_state at the gathering place), and place
  identity is tokens not coordinates — PASS.
- One-task-one-PR; V4/V5/M6 scope excluded (spec FR-008) — PASS.

**Post-design re-check**: structure below adds no scope beyond FR-001..FR-008 — PASS.

## Project Structure

```
mod/src/main/java/dev/kithcraft/mod/
  cast/        # cast seeding, identity persistence, nameplates (new)
  brain/       # schedule wiring, dusk pair-formation Activity, signal (new)
  mixin/       # the ≤3 task-list overrides (new; enumerated in fabric mod json)
mod/src/test/java/dev/kithcraft/mod/
  cast/ brain/ # unit tests
specs/014-augmented-villager/research/brain-26.2.md   # FR-001 evidence
```

## Phase ordering rationale

FR-001 (symbol re-derivation) gates everything — villager-brain-api.md's sharpest
finding is that `setTaskList` has NO successor of that name, so no brain code is
typed before the 26.2 surface is derived. Cast + schedule next (US1's substrate),
then the Mixin overrides (US1 close), then pair formation (US2) which needs the
schedule in place, then the live proofs (US1+US3 closure) which need everything.
