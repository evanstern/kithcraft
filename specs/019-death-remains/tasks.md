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

- [x] T007 Death handler: named grave at death site or nearest safe buildable
      surface, no villager agency; belongings captured before vanilla
      destruction into a roles:["storage"] thing named for its owner; new body
      token for the grave; dead token retired never-reissued (card ACs #4, #5,
      #9) — `dev.kithcraft.mod.death.GravePlacement` (pure, unit-tested
      bounded/deterministic search) + `dev.kithcraft.mod.live.LiveDeathHandling`
      (live wiring, both death kinds uniformly per R-6 — no conversion branch).
      **Capture hook needed NO new Mixin**: Fabric API's own
      `ServerLivingEntityEvents.ALLOW_DEATH` (already-shipped, already-Mixin'd
      inside the `fabric-entity-events-v1` module this mod already depends on
      transitively via `fabric-api`) fires at the head of
      `LivingEntity.die()`, before `dropAllDeathLoot` — verified by reading
      `villager.getInventory()` (public, `AbstractVillager implements
      InventoryCarrier`) there and copying every non-empty `ItemStack` before
      vanilla destroys them. `AFTER_DEATH` does the rest: token retirement
      (`DuskPairing.bodyTokenFor(UUID)`, added this phase, → `tokens.retire`),
      grave placement (vanilla `Blocks.OAK_SIGN` — no new block type — named
      via `SignBlockEntity.updateText`), a NEW body+place token pair for the
      grave (`TokenRegistryData.issue`), and the belongings bundle (vanilla
      `Blocks.CHEST`, captured items via `ChestBlockEntity.setItem`). Mixin
      surface stays at 4 — `kithcraft.mixins.json` untouched.
      **Deviation (card AC #5's edge case, spec's own "recorded as a deviation
      either way")**: chosen "empty chest, not omitted" — an empty capture
      still places a real (empty) chest, so the grave's world footprint is
      uniform regardless of what the villager was carrying.
      **Deviation/ponytail**: the belongings chest is placed at a fixed
      `gravePos.offset(1,0,0)` with no safety check of its own (only the
      grave site itself runs `GravePlacement`'s search) — correct for the
      common case, a documented ceiling for the rare case where that one
      offset happens to be unsafe (e.g. also over a lava edge); a real fix
      would run the same search from the grave site outward for the bundle
      too, not needed for this cast size/scope.
- [x] T008 Grief period: bed + job-site held unclaimed for configured period
      (default one cycle per R-3), config not constant, informed by R-4's
      finding (card AC #7) — `dev.kithcraft.mod.death.GriefPeriod` (pure,
      unit-tested: `configuredTicks()` reads `kithcraft.griefPeriodTicks`
      system property, default 24000, same config-not-constant idiom
      `BodySession`'s socket path already uses). Live half in
      `LiveDeathHandling.holdGrief`: since R-4 found `Villager.releasePoi`
      frees the `PoiManager` ticket but never erases the `HOME`/`JOB_SITE`
      brain memory it reads, both `GlobalPos`es are still readable in
      `AFTER_DEATH` — re-claims the just-freed ticket via `PoiManager.take`
      (an explicit vendor-side hold, exactly what R-4 said was needed, no
      natural lag to lean on) and releases it back via `PoiManager.release`
      once `GriefPeriod.Hold.isHeld(tick)` goes false, checked every server
      tick. **Ponytail**: `take()` reclaims exactly one ticket — correct for
      a bed (max 1 ticket) and this cast's job-site POI types; a POI type
      with more than one free slot would only have one slot held, not the
      whole record sealed. Ceiling: fine for the 3-villager cast this design
      targets; a full seal would need iterating `PoiManager.getInSquare` at
      that position instead of one `take` call.
- [x] T009 Tend-grave posting through the board read channel (plan's V4
      decoupling seam); takeable or ignorable (card AC #6; deviation note if
      the orchestrator holds AC #6 for V4 merge) —
      `dev.kithcraft.mod.death.GraveBoardEntry` (pure, unit-tested): composes
      §4.7 `text` content (`Testimony.textContent`, V2's existing "read from
      an artifact" percept — no new percept type) plus a mutable `taken`
      flag a survivor may or may not ever set. No deviation judged necessary:
      the content is genuinely fixture-independent (a `Place` + a string),
      exactly plan.md's "the seam is the posting content, not the book
      block" — when V4's board book merges, this content rides it unchanged.
      Composed and logged from `LiveDeathHandling.handleDeath` alongside the
      grave/belongings percepts; live delivery over a `WireClient` session is
      T010's job (Phase 4), same split `DuskPairing`'s `PairingSignal`
      already established (composed+logged this phase, wired to a live
      session later).

## Phase 4 — Proofs, gates, and closure (US4)

- [x] T010 Percept-channel proofs: witness gets ordinary sightings (no death
      percept type); absent villager gets change_report change:"gone" on
      return + grave sighting (card AC #8) —
      `mod/src/test/java/dev/kithcraft/mod/death/DeathPerceptChannelTest.java`,
      4 tests: structural absence (no death-shaped string in
      `Handshake.MANIFEST`'s `percept_types`), witness sighting composition
      (`origin:"saw"`, `doing:"dying"`), the §4.10 restriction resolving to
      exactly the absent body (no acting villager — a zombie holds no body
      token) plus its grave sighting, and all three envelopes pushed and
      received over a real loopback UDS `WireClient` session against a stub
      mind listener (`HandshakeWireClientTest`'s harness, reused). All 4
      green.
- [ ] T011 Dev-server observation: zombie-kill → grave + bundle + grief hold +
      zero sieges over the window; recorded per the runbook's dev-server-proofs
      gate (card ACs #2, #4 live halves)
- [ ] T012 Full gates: gradle build + test green; scope clean
- [ ] T013 Wiki re-ground: touched-source notes re-verified honestly
      ([[villager-brain-api]] — Mixin surface grows; overview); CAPSULES
      regenerated if descriptions changed; freshness green
- [ ] T014 Card ACs ticked with citing proofs; board/spec synced at PR time
