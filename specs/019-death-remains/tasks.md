# Tasks: Death, danger, and what remains (V5)

**Spec dir**: `specs/019-death-remains` · **Branch**: `task-0019-death-remains`

## Phase 1 — Verify before building (US0)

- [x] T001 R-4 verified: does POI re-claim have natural lag after
      `releaseAllPois()` at 26.2? Evidence at the brain-26.2.md standard
      (javap + decompiled source, cited), recorded in research/death-26.2.md
      — **no engine-side cooldown**; `releaseAllPois`/`releasePoi`/
      `PoiManager.release` run synchronously with no delay field anywhere;
      the only lag is another villager's own `AcquirePoi` scan cadence
      (~1-2s) or pathfind-retry backoff (2-22s, capped 20s) — grief period
      must be an explicit vendor-side hold, not natural lag.
- [x] T002 R-5 verified: where the zombie-siege trigger sits at 26.2, whether
      one targeted injection suppresses it, and whether a 3-villager cast
      meets eligibility at all; same evidence standard, same findings doc
      (card AC #1) — trigger is `VillageSiege.tick(ServerLevel, boolean)`
      (sole per-tick entry, one class); one `@Inject(HEAD, cancellable) →
      ci.cancel()` suppresses it fully; eligibility is pure POI-occupancy
      density (`PoiManager.isVillageCenter`) with no door/population count in
      26.2 — a 3-villager cast always qualifies once any bed/job-site POI is
      claimed.
- [x] T003 STOP/GO recorded: if the suppression point is not where death §1
      assumes, or suppression needs >1 targeted injection — STOP, surface to
      operator (runbook checkpoint 4); otherwise GO recorded with the planned
      injection point — **GO**. Suppression point matches death §1's
      assumption; exactly one targeted injection needed; total Mixin surface
      (V3's 2 + siege + conversion-cancel) lands at 4, decision-0002's ~4
      ceiling. Planned injection:
      `dev.kithcraft.mod.mixin.VillageSiegeMixin` on `VillageSiege.tick`,
      `@At("HEAD")`, unconditional `ci.cancel()`.

## Phase 2 — Suppression and permadeath (US1)

- [x] T004 Siege-suppression Mixin at the verified point; suppressed regardless
      of eligibility; MixinConfigTest enumeration updated (card AC #2) —
      `dev.kithcraft.mod.mixin.VillageSiegeMixin` on `VillageSiege.tick`,
      `@At("HEAD")`, unconditional `ci.cancel()`, exactly as death-26.2.md R-5
      planned; `kithcraft.mixins.json` lists it with a `kithcraftPurposes`
      citation. (Live zero-siege dev-server observation is Phase 4/T011, not
      this box — this is the Mixin half of card AC #2.)
- [x] T005 Conversion-cancel Mixin: conversion terminal-equivalent, routed
      through the death path; total Mixin surface within decision-0002's bound
      (card AC #3) — `dev.kithcraft.mod.mixin.ZombieConversionMixin` on
      `Zombie.convertVillagerToZombieVillager`, `@At("HEAD")`, cancellable,
      `cir.setReturnValue(false)`. Conversion point verified fresh this phase
      (death-26.2.md R-6, `javap -p -c` on `Zombie`/`ZombieVillager`,
      2026-08-28): `Zombie.killedEntity` rolls conversion odds and calls this
      one method, whose entire body is `Villager.convertTo(...)` — cancelling
      at HEAD (offset 0, nothing runs first) prevents the entity substitution
      outright; the victim's own `die()` (POIs already released beforehand,
      per R-4) falls through to its ordinary loot/removal path unchanged, so
      no explicit "route to death" code was needed — the normal path is what
      remains once the abnormal one is removed. Total Mixin surface: V3's 2 +
      siege (1) + this (1) = **4**, decision-0002's ceiling exactly.
- [x] T006 Structural absence checks: no self-preservation surface added
      (card AC #10), no friendly-fire guardrail (card AC #11); gradle green —
      `mod/src/test/java/dev/kithcraft/mod/death/StructuralAbsenceTest.java`,
      grep-style regression scan over `mod/src/main/java` for
      feed(ing)/escort(ing)/vigilan(t|ce) and friendly[- ]?fire; both absent,
      both tests pass. `MixinConfigTest`'s bound raised 3→4 (still enumerated,
      still purpose-cited per name — decision-0002's ceiling, not an open
      bound). `./gradlew build test`: **113 tests, 0 failures, 0 errors.**

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
