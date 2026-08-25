# Feature Specification: Body-vendor conformance — percepts out, intents in

**Feature Branch**: `task-0012-vendor-conformance` · **Spec dir**: `specs/012-vendor-conformance`

**Created**: 2026-08-25 · **Status**: Draft

**Input**: TASK-0012 / V2 — the two halves of the seam surface in the Fabric mod:
percept emission with provenance stamped at emission, and the act surface's four
core verbs with the intent/ack/act_result split. Consumes body-protocol-v0.md
(§2–§5, §12), decision-0002 + entity-implementation-comparison.md, and demo-build-plan
ruling R-8.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A world that only reports what a body could perceive (Priority: P1)

As a villager, I want the world to tell me only what my body could actually have
perceived, so that everything I later believe has an honest origin.

**Why this priority**: The percept surface is what every mind belief rests on; a
dishonest emission poisons everything downstream.

**Independent Test**: On a dev server, a villager body emits provenance-stamped
percepts a mind (or stub) ingests without rejection (card AC #1).

**Acceptance Scenarios**:

1. **Given** the vendor's declared percept types, **Then** the four-type floor
   (`sighting`, `observation`, `speech`, `act_result`) plus declared extras
   (`told_fact`, `text`, `self_state`, `change_report`) are emitted, each with
   provenance stamped at emission from the closed origin vocabulary (§2.6–2.7).
2. **Given** an `observation`, **Then** it carries its `vocabulary` (the kinds
   scanned — a subset of the manifest for a 3-D volume) and yields a falsifiable
   absence claim (card AC #2, §4.3).
3. **Given** ruling R-8, **Then** a hearing hook is verified and `sound` declared,
   or no hook verifies and `sound` is declared unsupported — the outcome recorded
   either way (card AC #3).
4. **Given** any percept shape, **Then** no `salience`/`importance`/`weight` field
   exists anywhere (card AC #4, §2.8); urgency bands are the only body-side signal,
   and `nearest_hostile` rides the free already-computed danger signal
   (decision-0002).
5. **Given** a `change_report`, **Then** it never reaches the body that caused the
   change or watched it happen (card AC #5, §4.10 — the $1.30/evening rule).

### User Story 2 - Acts that are requests, results that are percepts (Priority: P1)

As a villager's mind, I want my intents acknowledged on receipt and their outcomes
returned as percepts, so that I never learn what happened except by perceiving it.

**Independent Test**: The four verbs execute on a dev server, each returning exactly
one `act_result` (card AC #6).

**Acceptance Scenarios**:

1. **Given** an intent for `go_to`, `speak`, `attend`, or `wait`, **Then** the ack
   acknowledges receipt only and exactly one `act_result` percept follows (§5.1–5.4).
2. **Given** `go_to` targeting a body, **Then** the target resolves to that body's
   **last-seen** place, never its live position (card AC #7, §5.6 — no existence
   oracle).
3. **Given** an intent naming an unissued token, **Then** it is refused
   `unknown_target`; **given** a known-but-gone referent, **Then** it is accepted
   and fails with `target_gone` after the walk (card AC #8, §5.6).
4. **Given** `cancel`, **Then** it cancels a pending intent per §5.7.

### User Story 3 - A seam nothing native leaks across (Priority: P2)

As a future second-vendor author, I want the wire clean of engine-native types, so
that the mind never couples to Minecraft.

**Independent Test**: Protocol §12's six leak passes run clean over captured
payloads (card AC #9).

**Acceptance Scenarios**:

1. **Given** captured session payloads, **Then** no engine-native type name,
   registry identifier, class name, coordinate convention, or unit appears in any
   field a mind branches on (AR-1..AR-6): opaque `kind` tokens, `roles` +
   prose-only `descriptor`, place tokens + coarse distance bands, declared time
   unit, condition/quantity bands.

### Edge Cases

- Provenance's `source` is the immediate teller, not the original observer; `hops`
  grows, `origin` stays (§2.6).
- `observed_at` for a `told_fact` is the teller's last-seen time; `null` = unknown.
- Percept `seq` gaps may only mean deliberate `background` shedding (§4.11) — the
  emitter's budget/shedding honors that contract.
- A `speak` act's `act_result` reports the saying, not the hearing — whether others
  heard arrives (if at all) as their percepts, never the speaker's confirmation.
- Dev-server proofs record how they were verified in the PR description (gate line:
  test output or a documented `gradle runServer` observation).

## Requirements *(mandatory)*

- **FR-001**: Percept emission for the four-type floor + declared extras, each shape
  per §4, provenance stamped at emission (§2.6–2.7), urgency bands (§2.8), no
  salience-shaped field anywhere.
- **FR-002**: `observation` with `vocabulary` scoping (§4.3); `sighting` with
  prose-only `doing`.
- **FR-003**: R-8 resolved by verification: hearing hook verified → `sound` emitted
  and declared in the manifest; unverified → declared unsupported, recorded.
- **FR-004**: `change_report` delivery restriction (§4.10): never to actor or
  witness.
- **FR-005**: The act surface: `intent` decode, `intent_ack` (receipt only), the
  four core verbs executing via the vanilla `Brain<E>`/mod handlers per
  decision-0002, exactly one `act_result` per intent, `cancel` (§5).
- **FR-006**: Target resolution per §5.6: last-seen place for bodies;
  `unknown_target` only for unissued tokens; `target_gone` after the walk for
  known-but-gone.
- **FR-007**: Abstraction rule AR-1..AR-6 throughout; `nearest_hostile` exposed as
  the danger signal.
- **FR-008**: A leak-pass check over captured payloads implementing §12's six
  passes, runnable as a test/task.
- **FR-009**: Scope: no schedules/cast/dusk pairing (V3), no death mechanics (V5),
  no job-board build (V4). Mixin surface stays within decision-0002's bound.

## Success Criteria *(mandatory)*

- **SC-001**: `gradle build` + `gradle test` green in `mod/`.
- **SC-002**: Every card AC (#1–#9) demonstrated by a named test or a recorded
  dev-server observation (per the code-gates line for V-tasks).
- **SC-003**: §12's six leak passes run clean over captured payloads.
- **SC-004**: The R-8 outcome is recorded with its evidence.

## Assumptions

- V1's transport client, canonical writer, handshake, and token registry (merged
  PR #14) are fixed inputs this task emits through.
- MC 26.2 / Loader 0.19.3 / Fabric API 0.158.0 toolchain as verified in TASK-0009;
  version-dependent findings recorded per the evidence rule.
- The mind side needs no changes: the daemon (M1) already ingests and acks; a stub
  mind suffices for dev-server proof.
