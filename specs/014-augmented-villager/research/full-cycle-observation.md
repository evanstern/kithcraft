# T011 — full-cycle unattended observation (FR-007, card AC #1/#2), and T009's re-attempt

**Scope**: Phase 4. One continuous `./gradlew runServer` run, three named cast villagers, no
player connected, spanning at least one full day-length (24000 ticks / ~20 real minutes) of
elapsed game time, with a criteria checklist for wake/work/socialize/sleep/breeding/golems.
The same run also re-attempts T009 (inherited, unticked from Phase 3) since it passes through
whatever the world's current day/night position is.

## How it was run

`./gradlew runServer` in the background, world state reused as-is from the prior session
(cast already seeded — idempotence confirmed: no re-spawn log lines, `CastSeeder.seedIfNeeded`
no-ops as designed), `run/stdin.fifo` for console commands (no interactive TTY), `pause-when-
empty-seconds=-1`, no player ever connected. No `kithcraft.sock` present, so the mind is
permanently stalled at dial for the whole run (this run doubles as T010's live proof — see
`research/body-keeps-moving.md`). Server boot tick ≈ 29020; run ended (clean `stop`) after
**24m42s wall-clock**, tick ≈ 53400+ — comfortably past one full 24000-tick day-length span
starting from wherever the persisted world's day-position happened to be.

## Criteria checklist

| Criterion | Outcome | Evidence |
|---|---|---|
| **Wake** | **Not observed** | See "What was NOT observed" below. |
| **Work** (or tolerated wander fallback) | **Not observed** | See below — neither work-site pathing nor wander occurred. |
| **Dusk socialize** | **Not observed** | No `[dusk] pairing signal` log line; see T009 section below. |
| **Sleep in claimed beds** | **Not observed** | Villager Y-position (72) never matched claimed bed Y (67) at any sampled point. |
| **Zero breeding** | **Confirmed** | `execute if entity @e[type=minecraft:villager,tag=baby] run say BABY_FOUND` — never fired, sampled ~17 times across the run. |
| **Zero golem summons** | **Confirmed** | `execute if entity @e[type=minecraft:iron_golem] run say GOLEM_FOUND` — never fired, sampled ~17 times across the run. |
| **No player action required** | **Confirmed** | No player connected the entire run; only read-only `/data get`/`/time query`/`/execute if` diagnostic queries were issued — none of which affect villager behavior. |
| **Identity survives restart** (card AC #6, this run's restart) | **Confirmed** | All three villagers present immediately post-restart with correct `profession`/`type`: Aldric `armorer`/`plains`, Petra `farmer`/`desert`, Yenna `fisherman`/`taiga` — matches `Cast.MEMBERS` exactly. |

## What was NOT observed, and the honest finding behind it

Repeated `/data get entity <name> Pos` samples for all three villagers, taken roughly every
60–90s across the full 24m42s run (18 samples per villager), returned **bit-for-bit identical
floating-point positions every single time** — e.g. Aldric's position was
`[32.04675475590949d, 72.0d, 19.41952349530544d]` at the first sample and remained exactly
that string at the last sample, ~24 minutes later. This was cross-checked against a wild
chicken in the same world, which showed the same frozen-position pattern over a shorter
window, and against `Motion`, which was small but non-zero (e.g. `[0.0654, -0.0784, 0.0293]`)
at the instant sampled — the entity's physics tick is running (not a dead/no-AI entity;
`NoAI` tag is absent), but nothing ever displaced it to a new integer/fractional position
across ~29,000 ticks of elapsed time.

Cross-referencing `Brain` memory dumps: each villager's `home` (claimed bed) and
`meeting_point` GlobalPos are both at Y=67, while every villager's own entity position stayed
at Y=72 the entire run — a 5-block vertical gap that never closed. **The villagers were never
at their bed and never at the meeting point at any sampled instant.** Combined with the zero
displacement above, the honest read is: **no schedule-driven relocation (work-site pathing,
wander, dusk convergence, or sleep) was observed this run** — only continuous idle-in-place
ticking (small physics jitter, brain memories present and queryable, no exceptions).

**This is a materially different, and more concerning, finding than the "not enough real
time" root cause `research/pair-observation.md` recorded for Phase 3.** This run's window
(24m42s, ~24400 ticks) was deliberately long enough to cross a full day-length span
irrespective of the persisted world's starting day-position — yet no activity transition
manifested as villager movement at all. Root-cause hypotheses, **not chased further this
session** (an unbounded dive into which data-driven schedule resource governs 26.2's actual
day-boundaries, or why `JOB_SITE` is absent for at least Aldric — confirmed via a live `Brain`
dump this session, consistent with `research/schedule-observation.md`'s already-flagged-open
JOB_SITE finding — is exactly the open-ended investigation this project's own precedent
(`pair-observation.md`, `schedule-observation.md`) flags back rather than guesses through):

1. **Most likely — the persisted world's cast has no active pathing goal because `JOB_SITE`
   was never durably claimed** (confirmed absent from Aldric's live `Brain` dump this
   session — `{"minecraft:meeting_point": ..., "minecraft:home": ...}`, no `job_site` key),
   compounding `schedule-observation.md`'s already-open finding, and CORE-package random-
   stroll simply didn't roll during the sampled windows (a probabilistic per-tick behavior,
   not a guarantee within any given 60–90s slice).
2. **Possible — 26.2's `EnvironmentAttribute`-based schedule is not confirmed to wrap on a
   24000-tick cycle at all.** `research/pair-observation.md` already flagged that the actual
   day-boundary-to-`Activity` mapping is supplied by a reloadable data resource neither this
   session nor Phase 3's located; if that resource's boundaries are anchored differently than
   assumed (or the whole cast has settled into one `Activity` — e.g. `IDLE` — that this
   world's conditions never exit), a fixed-length real-time window doesn't guarantee crossing
   every phase the way vanilla's old cyclical day-time model did.
3. **Less likely — this specific persisted world's cast is in a genuinely stuck state**
   (not reproducible from a fresh cast) given three separate prior sessions
   (`schedule-observation.md`, `pair-observation.md`, this one) each observed *some* live
   villager movement or memory activity in this same world lineage at different times.

## T009 (inherited) — still not closed, same run

`grep -c "\[dusk\] pairing signal"` → **0** across the entire run. Consistent with the
"villagers never left position" finding above: `DuskPairing.tick()`'s own filter
(`Brain.isActive(Activity.MEET)`) never had a hit to report, because — per the Brain dumps —
neither villager's position ever moved toward the `meeting_point` (5+ blocks away the entire
run, well outside `ARRIVAL_RADIUS_BLOCKS`). This rules out `pair-observation.md`'s original
"already-within-arrival-radius, no-fire-by-design" alternate hypothesis for *this* run (they
were never close), and instead points at the same root cause as T011's finding above: the
`MEET` activity itself never appears to have engaged for either villager in this session's
window. **T009 remains unticked** — left honestly incomplete, same as Phase 3 left it, now
with an additional (and different) real-time budget invested and a materially different
finding recorded for whoever resumes it.

## Recommendation for whoever resumes T011/T009

Chase hypothesis 1 first (cheapest): confirm/claim `JOB_SITE` for all three cast members
durably (finish the work `schedule-observation.md` left open), then re-run this same
observation — if villagers still don't move, that isolates the cause to hypothesis 2 (the
schedule-attribute's actual cyclical behavior at 26.2), which needs the `EnvironmentAttribute`
data-resource research `pair-observation.md` already scoped out.

## Gates

`./gradlew build` green (111 tests, 0 failures/errors) both before this run (baseline) and
after (no production code changed by this phase's observation work — T010/T011/T009 are
observation-only tasks this session, per their own scope).

## 2026-08-27 — root cause of the zero-movement finding, found and fixed

Diagnosed on the exact persisted world this doc's run used (`mod/run/world`, cast already
seeded, reused unchanged), one continuous `./gradlew runServer` boot, no rebuild until after
the fix was confirmed live.

**Root cause.** `CastSeeder.forceLoadCastChunks` runs exactly once — inside `seedIfNeeded`,
before the `isSeeded()` early return — so it never re-runs on any subsequent boot, and its
footprint was computed from spawn-time positions only. `/forceload query` at this session's
boot confirmed only 2 chunks were still forced: `[1, 0]` and `[1, 1]`. All three cast members'
*actual* current positions (confirmed via `/data get entity <name> Pos`, all three at
`x≈32, z≈18–22`) sit in chunk `[2, 1]` — one chunk (Chebyshev distance 1) outside that forced
set. They got there by ordinary CORE-package wander after the original seeding (T006 already
documented this same wander live, up to ~14 blocks from spawn) and never wandered back.

This is not "unloaded" — it's worse to diagnose. Confirmed via `javap` on the pinned 26.2 jar
(`ChunkMap.FORCED_TICKET_LEVEL = ChunkLevel.byStatus(FullChunkStatus.ENTITY_TICKING)`,
`ChunkLevel`'s `ENTITY_TICKING_LEVEL`/`BLOCK_TICKING_LEVEL`/`FULL_CHUNK_LEVEL` triplet): a
forced chunk's ticket level propagates outward by Chebyshev distance, so the ring of chunks
one step beyond any forced chunk sits at `BLOCK_TICKING` — loaded, blocks/physics tick, entity
NBT is queryable and even jitters (small `Motion` values), but `ENTITY_TICKING` (the level
Brain/goal-selector AI needs) never applies. A villager that wanders one chunk too far
freezes permanently — present, queryable, physically jittering, but its Brain never advances
again: no wander, no schedule transition, no `[dusk] pairing signal`, ever, restart after
restart, indistinguishable from "hasn't decided to move yet" without this cross-check. This
also fully explains why `schedule-observation.md` (T006, same session as the original seed)
DID see real wandering: it queried shortly after the fresh seed, before any villager had
wandered past the tight, un-padded original footprint.

**How it was confirmed, live, no rebuild.** With the server already up on this persisted
world: `/data get entity Aldric Pos` → `[32.04675475590949d, 72.0d, 19.41952349530544d]`,
bit-identical to this doc's original 24-minute-old sample from a prior session — proving the
freeze is a permanent structural trap, not "insufficient elapsed time." `/forceload query` →
only `[1,0]`/`[1,1]` forced. Issued `/forceload add 0 -16 47 31` (covers `[2,1]` plus margin,
a plain console command, no code change) — within the same tick, log lines confirm new
`ENTITY_LOAD`s (a `Cow`, more `Pig`s, `FallingBlockEntity`s) as the previously-`BLOCK_TICKING`
area came up to full entity ticking. ~75 seconds later: `/data get entity Aldric Pos` →
`[22.36676251416939d, 67.5625d, 12.664268232989444d]` — moved ~13 blocks, landed at Y=67.56,
matching Aldric's claimed-bed Y (67) exactly — and `/data get entity Aldric Brain` now carries
a brand-new `minecraft:last_slept: 64429L` memory that was absent before the fix. Full
schedule-driven REST (sleep), not physics jitter, resumed within under two minutes of the
chunk actually reaching `ENTITY_TICKING`.

**Fix applied** (plain API, no Mixin, mirrors the existing `DuskPairing`-deferred-setup idiom
already in `KithcraftMod`): `CastSeeder.keepChunksLoaded(ServerLevel, List<Villager>)` computes
a bounding box from the cast's *actual current* block positions (not a stale spawn-time
origin) with a 3-chunk margin on every side — comfortably past the ~14-block wander T006
observed — and force-loads it. `KithcraftMod.onServerTickUnsafe` calls it once per boot, in
the same throttled-until-found block that already locates the cast for `DuskPairing.setUp`
(so it re-covers wherever the cast has actually wandered to, every restart, not just where it
originally spawned). `CastSeeder.forceLoadCastChunks` (the original, spawn-time-only call)
is unchanged and still runs once at first seed.

**Verification.** Live console commands only, within the one boot described above (no rebuild
needed to prove the mechanism); `./gradlew build` green after the code fix (20 test-result
files, 0 failures/errors) as the gate, not re-run live against a fresh 24-minute unattended
window — that full-cycle re-run (closing T011/T009 for real) is the next session's job, now
unblocked. This session did not chase whether the still-open `JOB_SITE` question
(`schedule-observation.md` §3) is affected by the same chunk-boundary mechanism — worth
checking first thing next session, since a villager whose job-site claim races
`ValidateNearbyPoi` while sitting in a `BLOCK_TICKING` chunk would silently fail that race
forever, same shape of bug.

Gate: `./gradlew build` green, 20 test-result files / 0 failures / 0 errors, after this
session's `CastSeeder.java`/`KithcraftMod.java` changes.

## 2026-08-27 (later) — verification re-run: T011/T009 CLOSED

Fresh `./gradlew build` (rerun with `--rerun-tasks`, not just cache-hit): **green, 111 tests,
0 failures/errors**. Same persisted world as the fix session (`mod/run/world`, cast already
seeded, `keepChunksLoaded` fix from commit 22e93e1 in place, no rebuild of the world). One
continuous `./gradlew runServer` boot, `run/stdin.fifo` (held open via a background `sleep`
writer so console reads never hit EOF between commands), no player ever connected. Sampled
every ~55–70s real time via `/data get entity @e[type=minecraft:villager,name=<X>,limit=1]
Pos` (`/data get entity <name>` alone errors — needs a selector, not a bare name),
`/time query gametime`, and the same `execute if ... run say BABY_FOUND`/`GOLEM_FOUND` probes
as prior sessions. Run: `BUILD SUCCESSFUL in 29m 49s`, world tick 66879 → ~102000+ — comfortably
past one full 24000-tick day-length, and started at a lucky day-position (≈19765, late night)
that let this single window cross wake → work → dusk → sleep → wake again.

| Criterion | Outcome | Evidence |
|---|---|---|
| **Wake** | **Confirmed**, twice | `last_woken: 74099` (all three) — villagers stationary at bed Y=67.5625 through tick 73489 (19:34:38), first seen displaced toward the meeting point by tick 74711 (19:35:39). Second wake: stable sleep through tick 97810 (19:54:54), a Y=67.0 "just risen" cluster position at tick 99033 (19:55:55), dispersing by tick 100255. |
| **Work (tolerated wander)** | **Wander confirmed; true job-site work NOT** | Widely scattered positions (x/z spanning ~10–52, Y 63–80) across both day windows (tick ~74711–83000 and ~99033–101477+). `Brain` dump mid-run (tick ~89000s) shows `{meeting_point, home, last_slept, last_woken}` — **no `job_site` key** for any of the three. The fix does not touch `ScheduleSetup`/POI claiming, so this is the same pre-existing JOB_SITE-claim gap `schedule-observation.md` already flagged, unaffected by this session's chunk-boundary fix. Falls under the spec's own tolerated fallback ("vanilla wander-instead-of-work"), not a blocker. |
| **Dusk socialize** | **Confirmed** | Positions converge on the bell/`meeting_point` (`[26, 67, 14]`) starting ≈tick 81276 (Aldric+Petra within ~1 block of each other, still short of the bell); by tick 83720 (19:43:11) Aldric `[26.87, 69, 16.39]` and Yenna `[27.49, 69, 16.52]` are both within `ARRIVAL_RADIUS_BLOCKS` (4) of the bell. |
| **Pairing signal precedes arrival (T009, FR-005, card AC #4)** | **Confirmed** | `[dusk] pairing signal` fired **10 times** across all three pair combinations between tick 83088 (19:42:38) and tick 85092 (19:44:18) — etas 1.82s–4.96s, all in `(0, LEAD_SECONDS=10]`. Emission (19:42:38, first fires) precedes the sampled arrival-radius confirmation (19:43:11, ≈33s later) — signal strictly before arrival, never at/after it. Honest note: observed lead times (1.8–4.96s) were well inside the 10s cap but shorter than its nominal value in every instance this run — consistent with `DuskPairing`'s own documented finding that vanilla's `StrollAroundPoi`/`SocializeAtBell` moves in short recomputed hops, so a given tick's eta reflects the current hop's remaining distance, not the full remaining approach; this is an already-known, already-flagged characteristic (see the class javadoc), not a new defect, and does not violate the "never fires on/after arrival" requirement. |
| **Sleep in claimed beds** | **Confirmed** | After dusk, positions settle to each villager's own claimed `home` BlockPos exactly — Aldric `[22.369…, 67.5625, 12.666…]` (home `[22,67,12]`), Petra `[24.35, 67.5625, 12.649…]` (home `[24,67,12]`), Yenna `[20.363, 67.5625, 12.661…]` (home `[20,67,12]`) — and stay **bit-identical across 5 consecutive samples** spanning tick 92921→97810 (~4 real minutes), confirming a genuine sleep hold, not a pass-through. `last_slept`: Aldric 86349, Petra 86165, Yenna 86185. |
| **Zero breeding** | **Confirmed** | `BABY_FOUND` never fired across 24 sampled checks over the whole run. |
| **Zero golem summons** | **Confirmed** | `GOLEM_FOUND` never fired across 24 sampled checks over the whole run. |
| **No player action required** | **Confirmed** | No player connected; only read-only `/data get`/`/time query`/`/execute if` diagnostics issued. |

**T011 and T009 are both closed by this run.** The chunk-boundary fix (commit 22e93e1) is the
root-cause resolution: the prior "zero movement" finding is gone, and a full wake → wander →
dusk-converge → pairing-signal → sleep cycle plays out unattended, twice over in this single
window (one full cycle, plus wake into a second). The still-open JOB_SITE gap
(`schedule-observation.md` §3) is unaffected by this fix and remains out of this task's scope,
per the spec's own tolerated-wander carve-out.

Gate: `./gradlew build --rerun-tasks` green, 111 tests / 0 failures / 0 errors, immediately
before this run (no production code changed this session — verification only).
