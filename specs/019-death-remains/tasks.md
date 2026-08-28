# Tasks: Death, danger, and what remains (V5)

**Spec dir**: `specs/019-death-remains` · **Branch**: `task-0019-death-remains`

## Phase 1 — Verify before building (US0)

- [ ] T001 R-4 verified: does POI re-claim have natural lag after
      `releaseAllPois()` at 26.2? Evidence at the brain-26.2.md standard
      (javap + decompiled source, cited), recorded in research/death-26.2.md
- [ ] T002 R-5 verified: where the zombie-siege trigger sits at 26.2, whether
      one targeted injection suppresses it, and whether a 3-villager cast
      meets eligibility at all; same evidence standard, same findings doc
      (card AC #1)
- [ ] T003 STOP/GO recorded: if the suppression point is not where death §1
      assumes, or suppression needs >1 targeted injection — STOP, surface to
      operator (runbook checkpoint 4); otherwise GO recorded with the planned
      injection point

## Phase 2 — Suppression and permadeath (US1)

- [ ] T004 Siege-suppression Mixin at the verified point; suppressed regardless
      of eligibility; MixinConfigTest enumeration updated (card AC #2)
- [ ] T005 Conversion-cancel Mixin: conversion terminal-equivalent, routed
      through the death path; total Mixin surface within decision-0002's bound
      (card AC #3)
- [ ] T006 Structural absence checks: no self-preservation surface added
      (card AC #10), no friendly-fire guardrail (card AC #11); gradle green

## Phase 3 — Remains, grief, tokens (US2 + US3)

- [ ] T007 Death handler: named grave at death site or nearest safe buildable
      surface, no villager agency; belongings captured before vanilla
      destruction into a roles:["storage"] thing named for its owner; new body
      token for the grave; dead token retired never-reissued (card ACs #4, #5,
      #9)
- [ ] T008 Grief period: bed + job-site held unclaimed for configured period
      (default one cycle per R-3), config not constant, informed by R-4's
      finding (card AC #7)
- [ ] T009 Tend-grave posting through the board read channel (plan's V4
      decoupling seam); takeable or ignorable (card AC #6; deviation note if
      the orchestrator holds AC #6 for V4 merge)

## Phase 4 — Proofs, gates, and closure (US4)

- [ ] T010 Percept-channel proofs: witness gets ordinary sightings (no death
      percept type); absent villager gets change_report change:"gone" on
      return + grave sighting (card AC #8)
- [ ] T011 Dev-server observation: zombie-kill → grave + bundle + grief hold +
      zero sieges over the window; recorded per the runbook's dev-server-proofs
      gate (card ACs #2, #4 live halves)
- [ ] T012 Full gates: gradle build + test green; scope clean
- [ ] T013 Wiki re-ground: touched-source notes re-verified honestly
      ([[villager-brain-api]] — Mixin surface grows; overview); CAPSULES
      regenerated if descriptions changed; freshness green
- [ ] T014 Card ACs ticked with citing proofs; board/spec synced at PR time
