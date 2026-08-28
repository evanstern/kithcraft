# Death surface re-derivation against MC 26.2 (Phase 1 / T001-T003)

**Evidence rule**: every claim below is a direct `javap -public`/`javap -p`/`javap -p -c` read
against the pinned jar, checked 2026-08-28. Jar: `~/.gradle/caches/fabric-loom/minecraftMaven/
net/minecraft/minecraft-merged-deobf/26.2/minecraft-merged-deobf-26.2.jar` (Mojang official
mappings, unobfuscated — same jar `specs/014-augmented-villager/research/brain-26.2.md` used).
Commands are inlined per section so this is independently re-runnable. This closes the two
open items death-mechanics.md §6.2 left for this task (items 1 and 4) and the two unverified
flags `specs/005-death-mechanics/research/death-surface.md` carried forward (§2 note, footnote
2).

## R-4: does POI re-claim have natural lag after `releaseAllPois()`?

```
javap -p -c -classpath minecraft-merged-deobf-26.2.jar net.minecraft.world.entity.npc.villager.Villager
javap -p -classpath minecraft-merged-deobf-26.2.jar net.minecraft.world.entity.ai.village.poi.PoiManager
javap -p -classpath minecraft-merged-deobf-26.2.jar net.minecraft.world.entity.ai.village.poi.PoiRecord
javap -p -c -classpath minecraft-merged-deobf-26.2.jar net.minecraft.world.entity.ai.behavior.AcquirePoi
javap -p -c -classpath minecraft-merged-deobf-26.2.jar 'net.minecraft.world.entity.ai.behavior.AcquirePoi$JitteredLinearRetry'
```

**Finding: no engine-imposed cooldown on the release/re-claim path itself. The only lag is
another villager's own scan cadence, and it is short (seconds, not the grief period's
day/night cycle).**

- `Villager.die(DamageSource)` calls `releaseAllPois()` synchronously (bytecode offset 37,
  before the `super.die()` call at 41) — not deferred to next tick. `releaseAllPois()` (private)
  calls `releasePoi(MemoryModuleType)` four times, once each for `HOME`, `JOB_SITE`,
  `POTENTIAL_JOB_SITE`, `MEETING_POINT`.
- `releasePoi` reads the villager's own memory for that type, and — if present — calls
  `PoiManager.getServer().getLevel(dimension).getPoiManager()` then (via
  `lambda$releasePoi$0`) resolves the POI's type and, when the memory's freed type matches
  `POI_MEMORIES`'s registered predicate for that `MemoryModuleType`, releases the ticket. This
  chain runs synchronously inside `die()` — no `server.execute(...)`-deferred scheduling, no
  tick delay.
- `PoiManager.release(BlockPos)` → `PoiSection.release(BlockPos)` → `PoiRecord.releaseTicket()`
  (`private int freeTickets; protected boolean releaseTicket()` — increments the in-memory
  ticket count directly). **No cooldown field, no timestamp gate anywhere in `PoiRecord`,
  `PoiSection`, or `PoiManager.release`** — the freed ticket is visible to any subsequent
  `PoiManager` query (e.g. `findAllClosestFirstWithType(..., Occupancy.HAS_SPACE)`) the instant
  `die()` returns.
- **The only real "lag" is the searching villager's own `AcquirePoi` behavior cadence**, not
  anything the freed POI itself imposes:
  - `AcquirePoi`'s trigger (`lambda$create$3`) only re-scans when its `MutableLong` next-attempt
    timestamp has passed: first scan is scheduled `random.nextInt(20)` ticks out, then each
    subsequent empty-handed tick reschedules `+20 + random.nextInt(20)` ticks (so roughly every
    **20-40 game ticks, 1-2 real seconds**, while the memory stays `VALUE_ABSENT`) before it
    calls `PoiManager.findAllClosestFirstWithType(..., SCAN_RANGE=48, Occupancy.HAS_SPACE)`
    again.
  - If a target is found but pathfinding to it fails, `JitteredLinearRetry` backs off instead:
    `markAttempt` sets the next retry `currentDelay = min(400, currentDelay + 40 + rand(40))`
    ticks out (so a failing path retries every **2-22 real seconds**, capped at 400
    ticks/20s — `MIN_INTERVAL_INCREASE=40`, `MAX_INTERVAL_INCREASE=80`,
    `MAX_RETRY_PATHFINDING_INTERVAL=400`, all confirmed constants). This only fires on
    pathfinding failure, not on every re-claim.
- **Consequence for death-mechanics.md §2's grief-period design**: the "instant reassignment"
  risk §2 worried about is real and, if anything, faster than "instant" suggests — a freed
  bed/job-site is re-claimable by any other qualifying villager within about 1-2 seconds of
  `die()` returning, not after any built-in delay. There is no free lunch here: **the grief
  period must be an explicit vendor-side hold (a Mixin-visible reservation or an
  occupancy-blocking marker), not something read off natural engine lag** — confirms plan.md's
  own hedge ("natural lag may carry part of it; the config is the guarantee") the wrong way:
  natural lag carries none of it in practice. T008 should not budget on natural lag at all.

## R-5: where does the zombie-siege trigger sit, is it one-Mixin suppressible, and does a 3-villager cast meet eligibility?

### 5a. Trigger location and construction site

```
jar tf minecraft-merged-deobf-26.2.jar | grep -i siege
javap -p -c -classpath minecraft-merged-deobf-26.2.jar net.minecraft.world.entity.ai.village.VillageSiege
javap -p -c -classpath minecraft-merged-deobf-26.2.jar net.minecraft.server.MinecraftServer   # grep VillageSiege
```

- `net.minecraft.world.entity.ai.village.VillageSiege implements CustomSpawner` is the entire
  siege mechanism — one class, no package split. `MinecraftServer.createLevels()` bytecode
  shows the overworld's custom-spawner list built as a single 5-element
  `ImmutableList.of(new PhantomSpawner(), new PatrolSpawner(), new CatSpawner(),
  new VillageSiege(), new WanderingTraderSpawner(savedDataStorage))`, passed into the overworld
  `ServerLevel`'s constructor — confirmed by the `new #640 // class .../VillageSiege` +
  `invokespecial <init>:()V` pair at MinecraftServer bytecode offsets 80/84, inside the
  `LevelStem.OVERWORLD`-keyed branch only (sieges are overworld-only, matches vanilla).
- `VillageSiege.tick(ServerLevel, boolean)` is the single per-tick entry point
  `ServerLevel`'s custom-spawner loop calls every tick (the `CustomSpawner` interface contract).
  Its own bytecode: bail early if not night/no-forced-tick; else check a clock-based time
  marker `ClockTimeMarkers.ROLL_VILLAGE_SIEGE` (26.2's environment-attribute clock system,
  replacing the old hardcoded tick-18000 check death-surface.md §2 cited — same "schedule is
  gone as a type" restructure `brain-26.2.md` §1 found); when that marker fires, roll
  `random.nextInt(10) == 0` (the 10% nightly chance) to flip `siegeState` to `SIEGE_TONIGHT`;
  once in that state, `tryToSetupSiege` (village-eligibility + spawn-point search) then
  `trySpawn` (spawns one zombie per tick, 20 total, via `zombiesToSpawn`/`nextSpawnTime`
  countdown) run each subsequent call until `zombiesToSpawn` hits 0 (`siegeState = SIEGE_DONE`).

**Suppression point: `VillageSiege.tick(ServerLevel, boolean)`, `@Inject(at = @At("HEAD"),
cancellable = true)`, unconditional `ci.cancel()`.** One method, one class, same shape as the
already-landed `VillagerGossipMixin` (`@Inject(method = "gossip", at = @At("HEAD"),
cancellable = true)`). Cancelling at HEAD stops the state machine from ever advancing past
`SIEGE_DONE`'s default — no roll, no setup, no spawn, regardless of night/day, clock marker, or
eligibility. This is the death-mechanics.md §1 "recommend Mixin-suppress the trigger" ruling,
confirmed at a concrete, single injection point. **One targeted injection, not more.**

*(An alternative suppression point exists — `@Redirect`/`@ModifyArg` on the
`ImmutableList.of(...)` call in `MinecraftServer.createLevels()` to drop `VillageSiege` from
the constructed spawner list — but it targets a constructor-argument list on a much larger,
more volatile method (`createLevels()` is deep JVM/level-bootstrap plumbing) for no suppression
benefit over cancelling `tick()` directly. `tick()` HEAD is strictly the safer, smaller target
and is what T004 should use.)*

### 5b. Village-eligibility thresholds — does a 3-villager cast ever qualify?

```
javap -p -c -classpath minecraft-merged-deobf-26.2.jar net.minecraft.world.entity.ai.village.VillageSiege   # tryToSetupSiege
javap -p -c -classpath minecraft-merged-deobf-26.2.jar net.minecraft.server.level.ServerLevel   # isVillage/isCloseToVillage/sectionsToVillage
javap -p -c -classpath minecraft-merged-deobf-26.2.jar net.minecraft.world.entity.ai.village.poi.PoiManager   # sectionsToVillage/isVillageCenter + lambdas
javap -p -constants -classpath minecraft-merged-deobf-26.2.jar net.minecraft.world.entity.ai.village.poi.PoiManager
```

- `tryToSetupSiege(ServerLevel)` iterates every non-spectator online player and, for each,
  calls `ServerLevel.isVillage(player.blockPosition())` — **no per-player count, no door count,
  no population count anywhere in `VillageSiege` itself.** It only additionally excludes the
  `#minecraft:without_zombie_sieges` biome tag.
- `ServerLevel.isVillage(BlockPos)` → `isCloseToVillage(pos, 1)` → `sectionsToVillage(SectionPos)
  <= 1` → `PoiManager.sectionsToVillage(SectionPos)` → `DistanceTracker.getLevel(sectionKey)`, a
  **flood-fill graph over 16-block sections** (same mechanism family as light propagation:
  `PoiManager$DistanceTracker extends SectionTracker`). `DistanceTracker.getLevelFromSource`
  seeds level **0** for any section where `PoiManager.isVillageCenter(sectionKey)` is true, and
  the flood propagates outward capped at level **`MAX_VILLAGE_DISTANCE = 6`** sections
  (confirmed `-constants`).
- `isVillageCenter(long)` (private) resolves to
  `lambda$isVillageCenter$0`: `section.getRecords(poiType -> poiType.is(PoiTypeTags.VILLAGE),
  Occupancy.IS_OCCUPIED).findAny().isPresent()` — **true the instant a single POI record in
  that section, of a type tagged `#minecraft:village`, is OCCUPIED (i.e., one villager has
  claimed it as home/job-site).** `data/minecraft/tags/point_of_interest_type/village.json`
  (present in the jar) is exactly the tag beds and most job-site POI types carry.
- **This is a strictly different, and strictly lower, bar than the door/population-count
  folklore `death-surface.md` §2 and its footnote 2 flagged as unverified.** At 26.2 there is
  no door count and no population count anywhere in the eligibility path — it's pure POI
  occupancy density. **A 3-villager cast qualifies as soon as any one of the three has claimed
  a bed (or any `village`-tagged job-site POI)** — which, per the ratified wake/work/sleep
  schedule ([[v1-demo]]), every cast member does by design within the first night. **A
  3-villager household absolutely can and, on any populated bed, will trigger the 10%-nightly
  siege roll** — the "maybe it never qualifies" hope death-mechanics.md §1 flagged as an open
  question is resolved: it always qualifies, once any bed is claimed. This *reinforces* plan.md
  design decision 2 ("suppress regardless of eligibility") rather than making it moot — the
  eligibility gate was never going to save a small cast from this, so the Mixin is doing real
  work, not covering a case that couldn't have fired.

## Findings summary

| Question | 26.2 verdict | Evidence |
|---|---|---|
| POI re-claim natural lag after `releaseAllPois()`? | **None on the release/manager side** (synchronous, no cooldown field). Only lag is the *next villager's* own `AcquirePoi` scan cadence (~1-2s while memory absent) or `JitteredLinearRetry` backoff on pathfind failure (2-22s, capped 20s) — both far short of a day/night cycle. Grief period must be an explicit vendor-side hold. | R-4, `Villager.die/releaseAllPois/releasePoi`, `PoiManager.release`, `PoiRecord.releaseTicket`, `AcquirePoi`/`JitteredLinearRetry` bytecode + constants |
| Where does the siege trigger sit? | `VillageSiege implements CustomSpawner`, constructed once in `MinecraftServer.createLevels()`'s overworld spawner list; `VillageSiege.tick(ServerLevel, boolean)` is the sole per-tick entry point (clock-marker gate → 10% nightly roll → eligibility/setup → 20-zombie spawn drip). | R-5a, `VillageSiege`/`MinecraftServer` bytecode |
| Suppressible with one targeted Mixin? | **Yes** — `@Inject(method="tick", at=@At("HEAD"), cancellable=true)` on `VillageSiege`, unconditional `ci.cancel()`. Same shape/size as the landed `VillagerGossipMixin`. | R-5a |
| Does a 3-villager cast meet eligibility at all? | **Yes, unconditionally, once any bed/job-site POI is occupied** — 26.2 eligibility is pure POI-occupancy-density flood-fill (`PoiManager.isVillageCenter`/`DistanceTracker`), no door/population count exists in this version's code path at all. | R-5b, `ServerLevel`/`PoiManager` bytecode + constants |

## STOP/GO

**GO.**

- The suppression point (`VillageSiege.tick`, HEAD, unconditional cancel) is exactly the shape
  death-mechanics.md §1 assumed ("recommend Mixin-suppress the trigger") — a single class, a
  single per-tick method, no state machine to partially unwind. Not a mismatch.
- Suppression needs **one** targeted injection, not more than one — the escalation trigger's
  second clause does not fire either.
- Eligibility resolves cleanly (3-villager cast always qualifies once a bed is claimed), which
  removes the one place this could have gone sideways (a finding of "sieges need >1 injection
  because eligibility is checked in multiple places" — it isn't; it's checked once, in
  `tryToSetupSiege`, itself gated by the single `tick()` entry this Mixin cancels before it's
  ever reached).
- Mixin budget check: V3 landed 2 (`VillagerGoalPackagesMixin`, `VillagerGossipMixin`). Phase 2
  adds this siege-suppression Mixin (1) plus the already-scoped conversion-cancel Mixin (1,
  R1.3's committed budget line in `docs/design/entity-implementation-comparison.md`) — total
  **4**, landing exactly at decision-0002's "~4 targeted Mixin injection points" ceiling, not
  past it.

**Planned injection point for T004**: `dev.kithcraft.mod.mixin.VillageSiegeMixin`, `@Mixin(VillageSiege.class)`, `@Inject(method = "tick", at = @At("HEAD"), cancellable = true)`, `ci.cancel()` — no arguments consumed, no eligibility replication needed in the Mixin itself since cancelling before any of `tryToSetupSiege`/`trySpawn` runs is sufficient regardless of what they would have found.
