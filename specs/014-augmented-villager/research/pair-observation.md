# T009 — dev-server observation: dusk pair formation + the pre-arrival signal

**Scope**: Phase 3 (T007–T009). Attempts to observe two cast villagers converge at the
shared gathering place and the pairing signal fire ~10s before arrival (FR-005, card AC #4).

## Substrate fact found this session: MC 26.2's villager schedule ignores `/time set`

`/time set <n>` (and `/time query daytime`, which errors — `Can't find element
'minecraft:daytime' of type 'minecraft:timeline'`, a separate 26.2 rename) only writes the
level's cyclical day-time field. Villager scheduling does **not** read that field any more.
Confirmed by `javap -p -c` on `Villager.class`'s `registerBrainGoals`:

```
aload_1
aload_0 ; level()
invokevirtual Level.getGameTime:()J      <-- monotonic tick count, NOT day-time
invokevirtual Brain.updateActivityFromSchedule:(...)V
```

and by direct correlation with this mod's own log: `KithcraftMod.onServerTickUnsafe`'s
`worldTime = server.overworld().getGameTime()` (used for `[live] attach scan at tick N`) kept
climbing monotonically across every `/time set` jump issued this session — never reset, never
affected. `Brain.setSchedule` now takes a generic `EnvironmentAttribute<Activity>`
(`EnvironmentAttributes.VILLAGER_ACTIVITY`, confirmed via `javap` on
`EnvironmentAttributes.class`: registered with `.defaultValue(Activity.IDLE)` and no
in-code layers — the actual time-window-to-Activity mapping is data-driven, not a
`bipush`/`sipush` constant in Java code, unlike the old `ScheduleBuilder`). **Practical
consequence**: the only way to move a cast villager through WORK → MEET → REST in this dev
harness is to let real server ticks elapse (20/s, no shortcut found this session); jumping
day-time has no effect on which `Activity` is active. Filed here because it is exactly the
kind of 26.2-substrate fact `research/brain-26.2.md` exists to carry, and the next person
driving a schedule-dependent observation needs it before spending time on `/time set` as this
session initially did.

## A real bug found and fixed this session: `DuskPairing` wired with zero villagers on boot

`KithcraftMod.onServerStarted` originally called `findCastVillagers(level)` (an
`getEntitiesOfClass` scan) synchronously, in the same handler that seeds the cast. Live log
evidence, every boot, restart after restart:

```
[18:25:15] [live] attach scan at tick 26115: 0 villager(s), 4 total entities: [...Chicken]
[18:25:15] [live] ENTITY_LOAD: ...Villager at BlockPos{x=32, y=72, z=18}
[18:25:15] [live] ENTITY_LOAD: ...Villager at BlockPos{x=32, y=72, z=22}
[18:25:15] [live] ENTITY_LOAD: ...Villager at BlockPos{x=32, y=72, z=19}
[18:25:16] [live] attach scan at tick 26135: 3 villager(s), ...
```

The persisted cast's `ENTITY_LOAD` fires a few ticks *after* `SERVER_STARTED`, matching the
same "not loaded yet when this fires" class of race `ScheduleSetup`'s and `CastSeeder`'s own
javadocs already found and fixed for chunk force-loading and POI registration — just not
extended to this mod's own boot sequence for `DuskPairing`. Consequence: `DuskPairing.setUp`
ran with an empty `castVillagers` list every single restart, so `seats` was permanently empty
and the pairing signal could never fire regardless of world time.

**Fix** (`KithcraftMod.java`): `onServerStarted` now only stashes the token store + origin;
`DuskPairing.setUp` is deferred into `onServerTickUnsafe`, retried every 20 ticks (mirroring
the existing attach-loop throttle) until `findCastVillagers` returns a non-empty list.
Confirmed live after the fix: villagers load at tick+20 as before, and by the very next
throttled check the villager Brain carries `minecraft:meeting_point` —

```
[18:27:23] Yenna has the following entity data: {..., Brain: {memories: {
  "minecraft:meeting_point": {value: {pos: [I; 26, 67, 14], dimension: "minecraft:overworld"}},
  "minecraft:home": {value: {pos: [I; 20, 67, 12], dimension: "minecraft:overworld"}}}},
  ..., Motion: [0.271, -0.078, -0.112], VillagerData: {profession: "minecraft:fisherman", ...}}
```

confirming `DuskPairing.setUp` ran successfully once villagers were actually present, and the
Brain is actively ticking (nonzero `Motion`).

## What was NOT observed: the pairing signal itself

Across this session's live runs (several restarts, ~20+ cumulative real minutes, plus one
`/time set` sweep across the full 0–24000 range that — per the finding above — had no effect
on scheduling), `world.getGameTime() % 24000` never reached a value late enough in the day for
`Activity.MEET` to engage during the observation window: the final run ended around tick
28655 (≈ day-position 4655, i.e. still WORK), short of the old vanilla schedule's ≈9000 MEET
boundary reference point (26.2's actual boundary is data-driven and not independently
re-derived this session — see substrate fact above). No `[dusk] pairing signal` log line
ever appeared; `DuskPairing.tick`'s `Activity.MEET` filter (`seats.stream().filter(s ->
s.villager().getBrain().isActive(Activity.MEET))`) never had a hit to report in this
session's real-time budget.

**Root-cause hypothesis, in order of likelihood:**
1. **Most likely — simply not enough elapsed real time.** With no day-time shortcut
   available (see above), reaching MEET requires waiting out real ticks from whatever
   `getGameTime() % 24000` position the world starts at; this session's runs never had a long
   enough uninterrupted window (each restart also resets the "day position" forward only by
   real elapsed seconds, not backward — a run that starts late in the day-length can't cheaply
   reach the next MEET window without ~20 real minutes).
2. **Less likely — the code path itself.** Ruled out to the extent observable: the fix above
   is confirmed live (`MEETING_POINT` correctly set on a real villager's Brain), so the
   MEET-side wiring (`Brain.getBrain().isActive(Activity.MEET)`, `PathNavigation`-based ETA in
   `DuskPairing.eta`) is untested live but has no further known blocker; `PairingSignalTest`'s
   9 unit tests independently cover the timing math and no-fire edge cases the spec's
   independent test names (predicted-arrival-minus-~10s, never on arrival, no-fire on pathing
   failure).

**Not chased further this session** — an unbounded wait to reach a specific data-driven
schedule window that this session could not shortcut is exactly the kind of open-ended
investigation that gets flagged back rather than guessed through. `T009` is left **unticked**
in `tasks.md`; the honest state is: fixture wiring confirmed live and fixed, signal math
unit-tested, but the live end-to-end "signal precedes arrival" observation itself did not
complete within this session.

## Recommendation for whoever resumes T009

Either (a) run the dev server for a genuinely long unattended real-time window (≥20 minutes
from a fresh boot, so `getGameTime() % 24000` is guaranteed to cross the MEET boundary once),
or (b) spend the research time this session didn't: `javap`/inspect
`EnvironmentAttributeLayer$TimeBased` and whatever data resource populates
`EnvironmentAttributes.VILLAGER_ACTIVITY`'s actual layers at runtime (this session confirmed
the Java static initializer registers the attribute with no layers baked in — the real
day-boundaries are supplied by a reloadable data resource this session did not locate) to find
a legitimate way to fast-forward to MEET, then repeat this observation.

## Gates

`./gradlew build` and `./gradlew test` both green after the fix above (see commit for exact
SHA). `PairingSignalTest`'s 9 cases (timing math, no-fire edge cases, content shape) pass
unmodified from the inherited code.

## 2026-08-27 (later) — T009 CLOSED, in the verification re-run

`research/full-cycle-observation.md`'s dated "verification re-run" section is the full record
(same session, same run — T009 and T011 share the one continuous ~29m50s boot). Summary here
for T009 specifically: `[dusk] pairing signal` fired **10 times** across all three pair
combinations (Aldric+Petra, Aldric+Yenna, Petra+Yenna) between world tick 83088 and 85092
(real time 19:42:38–19:44:18), with etas of 1.82s–4.96s — inside `PairingSignal.LEAD_SECONDS`
(10s), always `> 0` (never at/after arrival). Emission preceded the sampled confirmation of
actual arrival-radius convergence (Aldric+Yenna both within 4 blocks of the bell by 19:43:11,
~33s after the first fire). This closes the "signal precedes arrival" observation this
session's earlier root-cause work (chunk-boundary fix, commit 22e93e1) unblocked — Phase 3's
original finding (fixture wiring correct, live firing unobserved for lack of elapsed real
time/reachable MEET window) and Phase 4's first attempt (villagers frozen, no movement at all)
are both superseded by this result. **T009 is now ticked in tasks.md.**

Honest note carried forward: the observed lead times (1.8–4.96s) never approached the nominal
~10s ceiling — see `full-cycle-observation.md`'s note on `StrollAroundPoi`/`SocializeAtBell`'s
short recomputed hops producing a per-tick eta that reflects the current hop, not the full
remaining approach. Worth a closer look if a future task cares about the lead time being
closer to a full 10s rather than merely `≤10s`; out of scope for this closure.
