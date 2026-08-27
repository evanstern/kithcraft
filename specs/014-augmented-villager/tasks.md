# Tasks: The augmented villager — brain, schedule, cast, dusk pair formation

**Input**: specs/014-augmented-villager/ (spec.md, plan.md)
**Prerequisites**: mod/ toolchain (V1, PR #14), percept/act surface (V2, PR #17),
decision-0002, demo-build-plan R-7, villager-brain-api.md (partial 26.2 verification)

**Organization**: Phases map 1:1 to phase-scoped dispatches.

## Phase 1 — Re-derive the 26.2 brain surface; cast seeding (US1 groundwork)

**Goal**: The 26.2 symbol surface is verified fact, and three named villagers exist.

**Independent test**: `gradle test` green; research/brain-26.2.md exists with cited
evidence; villagers visible on a dev server with nameplates.

- [x] T001 Re-derive the Brain<E> surface against MC 26.2 (FR-001): Activity
      registration, addActivity task-list assignment, Sensor refresh,
      MemoryModuleType/POI registration, ScheduleBuilder or its successor —
      javap/source evidence with dates in research/brain-26.2.md; note which
      extension points need Mixin/accessor vs plain API at 26.2
- [x] T002 Implement cast/ — seed three named villagers (profession × biome variant,
      nameplates, card AC #6), identity persisted across restart (SavedData
      pattern); unit tests for seeding idempotence and identity survival
- [x] T003 Amend docs/wiki/villager-brain-api.md: replace the UNVERIFIED carry-
      forward sections with the derived 26.2 facts (honest re-pin after amendment)

**Checkpoint**: the substrate is known, the cast exists.

## Phase 2 — Schedule wiring and the suppression overrides (US1 close)

**Goal**: The cast runs the vanilla day unattended, minus breeding/gossip/golems.

**Independent test**: `gradle test` green; overrides enumerated in the Mixin config;
dev-server observation shows wake/work/sleep transitions.

- [x] T004 Wire wake/work/socialize/sleep scheduling for the cast on the vanilla
      substrate per the Phase 1 derivation: POI bed/job-site claims, sleep pathing,
      door use — driven, not reimplemented (FR-003)
- [x] T005 Implement mixin/ — the ≤3 task-list overrides suppressing breeding,
      gossip, iron-golem summoning (FR-004, card AC #3); each enumerated in the
      Mixin config; a test asserts the config lists at most three and names them;
      drop any override 26.2 makes unnecessary via plain API
- [x] T006 Dev-server observation: schedule transitions occur unattended over a
      cycle segment; no breeding/golem events (recorded per the V-task gate)

**Checkpoint**: an unattended vanilla day, suppressions in place.

## Phase 3 — Dusk pair formation and the ~10 s signal (US2)

**Goal**: R-7 implemented: convergence plus the pre-arrival pairing signal.

**Independent test**: `gradle test` green — signal-timing unit test proves the
signal fires on predicted-arrival-minus-~10 s, not on arrival.

- [ ] T007 Implement brain/ dusk pair-formation Activity: pick pair, path both to
      the shared gathering place at dusk (FR-005); gathering place exposed as a
      place token via the existing token registry
- [ ] T008 Implement the pairing signal: emitted when predicted arrival is ~10 s
      out (path-length/speed estimate), carrying pair body tokens + place token in
      seam terms via existing percept types — no protocol extension; no-fire when
      pathing fails (spec edge case); unit tests for timing math and no-fire
- [ ] T009 Dev-server observation: two villagers converge at dusk; signal precedes
      arrival by ~10 s (logged timestamps recorded per the V-task gate, card AC #4)

**Checkpoint**: M6's pre-generation lever exists.

## Phase 4 — Body-keeps-moving, full-cycle proof, gates, wiki, board (US3 + closure)

**Goal**: The stalled-mind proof, the full unattended cycle, and every gate green.

**Independent test**: full `gradle build` + `gradle test` green; freshness probe
green; card ACs ticked with citations.

- [ ] T010 Body-keeps-moving (FR-006, card AC #5): audit that no schedule/activity
      code path awaits a mind response (structural — the seam surface is async
      already; assert no blocking call sites); dev-server observation with a stub
      mind that never responds: bodies continue full schedule
- [ ] T011 The full-cycle unattended observation (FR-007, card AC #1/#2): three
      villagers, one full day/night, no player action — criteria checklist recorded
      (wake, work, dusk socialize, sleep in claimed beds, zero breeding/golems)
- [ ] T012 Gates + wiki + board: gradle build/test green; villager-brain-api.md +
      any other touched notes re-verified and honestly re-pinned; CAPSULES
      regenerated if descriptions changed; freshness probe green; card ACs ticked
      with citing evidence (backlog CLI in-worktree); runbook log row updated

**Checkpoint**: V3 done — the village lives a day without anyone.
