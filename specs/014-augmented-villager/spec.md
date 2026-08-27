# Feature Specification: The augmented villager — brain, schedule, cast, dusk pair formation

**Feature Branch**: `task-0014-augmented-villager` · **Spec dir**: `specs/014-augmented-villager`

**Created**: 2026-08-27 · **Status**: Draft

**Input**: TASK-0014 / V3 — the vanilla `Brain<E>` substrate driven from the Fabric mod:
schedules, activities, memory modules, POI claims inherited free per decision-0002; the
bounded Mixin surface (three task-list overrides, no more); the three-villager cast;
the dusk pair-formation Activity implementing ruling R-7; and the body-keeps-moving
rule. Consumes decision-0002 + entity-implementation-comparison.md, demo-build-plan
§3.3 (V3) + R-7, kithcraft-brief (micromanagement spell-breaker),
llm-routing-and-budget §5.2 lever 2, and docs/wiki/villager-brain-api.md (partially
verified at MC 26.2 — full re-derivation is THIS task's scoped work).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Neighbours who live a full day without me (Priority: P1)

As a player, I want three named neighbours who live a full day without me, so that the
village is somewhere I arrived rather than something I have to run.

**Why this priority**: "A full cycle unattended" IS the micromanagement spell-breaker
check, made testable (card design check). The moment keeping a villager fed, escorted,
or on-task requires the player, the demo has shipped the failure mode the brief names.

**Independent Test**: On a dev server, three named villagers run a full day/night
cycle unattended: wake, work, socialize at dusk, sleep in claimed beds — no player
action at any point (card AC #1).

**Acceptance Scenarios**:

1. **Given** a dev server with the cast seeded, **When** a full day/night cycle
   elapses with no player action, **Then** each villager wakes, works its job site,
   socializes at dusk, and sleeps in its claimed bed (card AC #1).
2. **Given** the suppression overrides, **Then** no breeding occurs and no
   gossip-driven iron golem is summoned across the cycle (card AC #2).
3. **Given** the cast, **Then** the three villagers are distinguished by profession ×
   biome variant plus nameplates (card AC #6, decision-0002 pairing).

### User Story 2 - Dusk pair formation with the pre-generation signal (Priority: P1)

As the mind daemon's conversation engine (M6), I want the pairing signal ~10 s before
the villagers actually meet, so that the opening turn can be pre-generated and the
latency posture survives (routing §5.2 lever 2).

**Independent Test**: Two villagers converge on the gathering place at dusk; the
pair-formation signal precedes arrival by ~10 s (card AC #4, ruling R-7).

**Acceptance Scenarios**:

1. **Given** dusk, **When** the socialize activity begins, **Then** two villagers
   path to the shared gathering place, and the pairing signal is emitted when arrival
   is predicted ~10 s out — not on arrival (R-7).
2. **Given** the signal fires, **Then** it identifies the pair and the place in seam
   terms (tokens, not coordinates) so M6 can consume it without a protocol extension.

### User Story 3 - Bodies keep moving while minds think (Priority: P1)

As a player watching a villager, I want the body busy while the mind thinks, so that a
20-second thought never looks like a 20-second freeze.

**Independent Test**: With a deliberately stalled mind (stub that never responds),
bodies continue their scheduled activity — observed on the dev server (card AC #5).

**Acceptance Scenarios**:

1. **Given** a mind connection that never resolves, **Then** the scheduled activity
   continues driving the body: pathing, work animations, schedule transitions all
   proceed (card AC #5).
2. **Given** the mind reconnects later, **Then** the body's schedule state is
   whatever vanilla scheduling says it is — the mind never became load-bearing for
   body liveness.

### Edge Cases

- Villager cannot path to the gathering place (blocked route): the signal must not
  fire for a pair that will not arrive; no pair forms that dusk, bodies continue.
- Bed or job-site POI missing/occupied: vanilla fallback behaviour stands; the cycle
  test tolerates vanilla wander-instead-of-work but not player intervention.
- Server restart mid-cycle: cast identity (names, profession × variant) persists;
  re-seeding does not duplicate villagers.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Re-derive the `Brain<E>` symbol surface against MC 26.2 (Mojang
  official mappings) BEFORE any brain/schedule code: Sensor-driven refresh, Activity
  registration, task-list assignment (`addActivity` family), `MemoryModuleType`
  registration, POI types, ScheduleBuilder — recorded with evidence (javap or source
  citations, dated) in `research/brain-26.2.md`. villager-brain-api.md carries the
  verified memory-module rows and flags everything else UNVERIFIED; this closes it.
- **FR-002**: Three named villagers seeded with distinct profession × biome variant
  and visible nameplates (card AC #6); identity survives a server restart.
- **FR-003**: Wake/work/socialize/sleep scheduling on the vanilla substrate: POI bed
  and job-site claims, sleep pathing, door use — inherited vanilla behaviour, driven
  not reimplemented (decision-0002).
- **FR-004**: The Mixin surface is enumerated and no larger than three task-list
  overrides suppressing breeding, gossip, and iron-golem summoning; conversion-cancel
  explicitly excluded (belongs to V5). Each override is listed in the PR and in the
  Mixin config (card AC #3). Prefer plain API where 26.2 allows it — an override that
  turns out unnecessary is dropped, not kept for symmetry.
- **FR-005**: A dusk pair-formation Activity (R-7): villagers path to a shared
  gathering place at dusk; the pairing signal is emitted ~10 s ahead of predicted
  arrival, carrying pair + place in seam-compatible terms (card AC #4).
- **FR-006**: The body-keeps-moving rule: no code path blocks a scheduled activity on
  a mind response; a stalled mind leaves bodies fully animated by their schedule
  (card AC #5).
- **FR-007**: Full-cycle unattended proof on the dev server, recorded per the
  runbook's V-task gate (test output or documented `gradle runServer` observation)
  (card AC #1, #2).
- **FR-008**: Out of scope: conversation content (M6), job-board build (V4), death
  and grief (V5), conversion-cancel injection (V5), any protocol extension.

### Key Entities

- **Cast member**: named villager = profession × biome variant + nameplate + persona
  binding slot (persona itself is M3's).
- **Gathering place**: the shared dusk destination — a place the pair-formation
  Activity paths to; exposed to the seam as a place token, not coordinates.
- **Pairing signal**: the ~10 s-ahead notification consumed by M6 for
  pre-generation; carries the pair's body tokens and the place token.

## Success Criteria *(mandatory)*

- **SC-001**: Card AC #1 — full unattended day/night cycle, three villagers, dev
  server, no player action.
- **SC-002**: Card AC #2 — zero breeding events, zero gossip-driven golem summons.
- **SC-003**: Card AC #3 — Mixin surface ≤ 3 task-list overrides, enumerated.
- **SC-004**: Card AC #4 — dusk convergence with signal ~10 s pre-arrival.
- **SC-005**: Card AC #5 — stalled-mind test: bodies keep moving.
- **SC-006**: Card AC #6 — cast distinguished by profession × variant + nameplates.
