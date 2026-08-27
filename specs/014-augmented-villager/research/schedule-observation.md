# T006 — dev-server observation: schedule wiring + suppression overrides

**Scope**: Phase 2 (T004–T006). Confirms the cast runs the vanilla day unattended (FR-003,
card AC #1's substrate) and the Mixin suppressions hold (FR-004, card AC #2), on
`./gradlew runServer` with `run/server.properties`'s `pause-when-empty-seconds=-1` (T009/T011
precedent, `specs/012-vendor-conformance/research/verb-observation.md`), commands piped
through `run/stdin.fifo`, no interactive TTY, no players connected throughout.

## What was observed, verbatim (this session's real transcript)

### 1. Chunks stay loaded and ticking with zero players (a T004 finding, not assumed)

First run, with no fix: three villagers spawn (`ENTITY_LOAD` fires, attach scan at tick 1
shows 3 villagers), then vanish entirely by tick 21 (`0 villager(s), 0 total entities`) and
never return. MC 26.2's chunk-ticket system does not keep any chunk loaded/ticking on its own
once the server has zero players connected — the same gap
`specs/012-vendor-conformance/research/verb-observation.md` flagged as V3's to solve
("keep-loaded chunk tickets"). Fixed in `CastSeeder.forceLoadCastChunks` via
`ServerLevel.setChunkForced` (the public API behind `/forceload`). After the fix, villagers
persist indefinitely (confirmed to tick 3400+ / ~170s real time in one run) and the attach
scan's distinct-entity-class list stays populated the whole time.

### 2. Identity stability: profession × biome variant survives, once two vanilla behaviors are accounted for

A code-spawned villager with a profession set directly (via `VillagerData`, not through
natural spawn/`finalizeSpawn`), 0 XP, and level 1 is exactly the condition vanilla's own
`ResetProfession` CORE behavior (confirmed by `javap -p -c`) treats as "claims a profession
but hasn't really worked it" — it fires once, ~1 second (20 ticks) after spawn, and resets
profession straight back to `NONE`. From there, `AssignProfessionFromJobSite` reassigns the
(now-unemployed) villager to whichever unclaimed job site it is nearest at that instant —
observed live as two of the three cast members' professions swapping between each other,
non-deterministically across runs (once Aldric↔none plus a rotation, once a clean Petra↔Yenna
swap). Root-caused via `javap -p -c` on `ResetProfession` (not assumed): the exact trigger is
`JOB_SITE absent && profession ∉ {NONE, NITWIT} && xp == 0 && level ≤ 1`. Fixed by seeding with
1 XP (`Villager.setVillagerXp(1)`, public), which short-circuits the trigger at its own
condition. Confirmed stable across three separate runs, checked at spawn, +5s, +15s, +30s,
and +55s real time — professions never drifted from `Cast.MEMBERS` once XP was seeded.

**Card AC #6** (profession × biome variant + nameplates, identity survives) holds for the
duration observed; full server-restart persistence (the "survives a restart" half) is
`CastData`'s SavedData round trip, already unit-tested (`CastTest`), not re-proven live here.

### 3. POI claims: bed (HOME) claims cleanly; job-site (JOB_SITE) claim is flaky — not fully closed

**HOME (bed)**: claims correctly and promptly via vanilla's own profession-independent CORE
behavior — confirmed live, e.g. `Aldric: {"minecraft:home": {value: {pos: [...], dimension:
"minecraft:overworld"}}}` within seconds of spawn, in every run. Vanilla claims the *nearest*
unclaimed bed, not necessarily the one placed for that specific member (beds are only 2 blocks
apart) — harmless, since bed identity doesn't matter functionally, only "a claimed bed exists."

**JOB_SITE**: placing the job-site block alone never produced a claim for an already-employed
villager, even after 30+ real seconds standing directly on it — traced to `AssignProfessionFromJobSite`
being the *only* vanilla path that ever sets `JOB_SITE`, and it only runs for
`profession == NONE`. Setting `JOB_SITE` directly on the Brain (`Brain.setMemory`, public)
was the next attempt; that memory kept disappearing, traced further to `ValidateNearbyPoi`
(CORE package) erasing it whenever `PoiManager.exists(pos, matchingType)` comes back false —
i.e. the block existing physically is not the same as the `PoiManager` having registered it,
and a `setBlockAndUpdate` issued synchronously during `onServerStarted` is not guaranteed to
have reached the `PoiManager` by the time `ValidateNearbyPoi` first runs. `ScheduleSetup` now
also calls `PoiManager.add` explicitly (via `PoiTypes.forState`) to close that race. The last
observed run after this fix shows real progress — a `SecondaryPoiSensor`-driven
`potential_job_site` memory appearing for the first time (proof at least one POI is now
correctly registered) — but the *own* `JOB_SITE` claim for the matching-profession villager
still did not reliably persist in the observation window.

**This is left open, not silently dropped.** The spec's own edge case explicitly tolerates
this: *"Bed or job-site POI missing/occupied: vanilla fallback behaviour stands; the cycle
test tolerates vanilla wander-instead-of-work but not player intervention."* A villager
without a persisted `JOB_SITE` still runs its full CORE/IDLE/MEET/REST schedule (wake, wander,
socialize, sleep) — it just doesn't path to work a specific block, which is the tolerated
degraded mode, not a card-AC-blocking failure. Chasing this further is exactly the kind of
open-ended vanilla-POI-internals investigation flagged back to the dispatcher rather than
guessed through indefinitely; `dev.kithcraft.mod.brain.ScheduleSetup`'s javadoc carries the
full chain of findings (chunk timing → ResetProfession → ValidateNearbyPoi/PoiManager) for
whoever picks this back up.

### 4. Wandering/movement: brains are ticking, not frozen

Villager positions were queried repeatedly across ~90+ seconds of real time in one run;
positions changed meaningfully between checks (e.g. one villager moved from spawn at
`z≈13` to `z≈27` over that window) — the brain is actively driving pathing, not stalled.
This is independent of the JOB_SITE finding above: CORE-package behaviors (wandering, look-at,
door interaction) run regardless of whether a job site is held.

### 5. Suppression overrides hold: zero breeding, zero golems, across every run

Across all runs in this session (several minutes of cumulative real time, 1000+ ticks per
run, three villagers wandering and re-clustering near each other and their shared fixtures —
exactly the proximity conditions that would trigger vanilla breeding/gossip if unsuppressed):

- `execute if entity @e[type=minecraft:iron_golem] run say GOLEM_FOUND` — never fired.
- `execute if entity @e[type=minecraft:villager,tag=baby] run say BABY_FOUND` — never fired.
- The attach-scan's distinct-entity-class log (fires every ~1s while unattached) never listed
  `IronGolem` or a baby `Villager` across any run.

Confirms `VillagerGoalPackagesMixin` (breeding, via `getIdlePackage`) and
`VillagerGossipMixin` (gossip + golem, via `gossip()` HEAD-cancel) hold under live conditions,
not just the unit-level Mixin-config check (`MixinConfigTest`).

## What was NOT observed

- A full day/night cycle (24000 ticks / 20 real minutes) — this pass used `/time set` jumps
  (dawn → noon → dusk → night) to exercise schedule transitions within a practical dev-loop
  window, per T006's "cycle segment" scope; the full unattended 24000-tick cycle is T011's
  (Phase 4) card AC #1/#2 closure, not this phase's.
- Dusk pair formation / the ~10s pairing signal — Phase 3 (US2), not built yet.
- A reliable JOB_SITE claim for an already-employed villager — see §3 above; left open with
  full findings for whoever resumes it, not claimed as done.
- Door-use — no walls/doors exist in this fixture layout (`ScheduleSetup`'s own ponytail note);
  not exercised.

## Gates

`./gradlew build` (compile + `processResources` incl. `kithcraft.mixins.json` + full test
suite) green throughout this phase's final state — `MixinConfigTest` (Mixin config ≤3,
purposes named) and all pre-existing tests (`CastTest` et al.) pass unmodified.
