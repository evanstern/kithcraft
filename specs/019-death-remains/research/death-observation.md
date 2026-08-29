# T011 — dev-server death observation (card ACs #2, #4, #7, #8, #9 live halves)

**Scope**: Phase 4. One `./gradlew runServer` boot against the persisted `mod/run/world`
(cast already seeded — Aldric/Petra/Yenna — from prior sessions), no player connected,
console commands via `run/stdin.fifo` (held open with a background `sleep 999999 >
run/stdin.fifo` writer so reads never hit EOF between commands, the same idiom prior
phases' research docs used). `-Dkithcraft.griefPeriodTicks=1200` (60s at 20 ticks/s)
passed to the forked server JVM via `JAVA_TOOL_OPTIONS` (loom's default `runServer` task
does not forward `-D` flags from the `gradlew` invocation itself; `JAVA_TOOL_OPTIONS` is
read by the `java` launcher directly regardless of who forked it — confirmed live by the
JVM's own `Picked up JAVA_TOOL_OPTIONS: -Dkithcraft.griefPeriodTicks=1200` banner in the
console log), so hold+release was observable in-session instead of waiting a full
24000-tick day/night cycle.

## The kill was forced, not waited for

No random zombie kill occurred or was waited on. After boot, `/data get entity
@e[type=minecraft:villager,name=Aldric,limit=1] Pos` located Aldric, then
`/damage @e[type=minecraft:villager,name=Aldric,limit=1] 100 minecraft:mob_attack` killed
him on the first attempt — `Applied 100.0 damage to Aldric`, followed immediately by
`Villager ... died, message: '%1$s was slain by %2$s'`. This is an honest forced kill: the
`mob_attack` damage type is the same category a real zombie's melee hit uses, and
`ServerLivingEntityEvents.ALLOW_DEATH`/`AFTER_DEATH` (`LiveDeathHandling`'s hooks) fire
identically regardless of the specific attacker per that class's own doc ("no separate
conversion branch" — R-6, `death-26.2.md`) — a forced `/damage` kill exercises the exact
same code path a live zombie kill would. This was chosen over summon-and-wait after the
prior session's two stalled attempts; `/damage`'s syntax (`<target> <amount>
[<damage type>]`) worked on the first try, no error-message discovery needed.

## Checklist — what was observed

| Criterion | Outcome | Evidence |
|---|---|---|
| **Named grave placed at/near death site** | **Confirmed** | Death at `x=2.50, y=68.00, z=62.25`. `/data get block 2 68 62` → `{id: "minecraft:sign", front_text: {messages: ["Here lies", "Aldric", "", ""]}}` — an oak sign at the death block itself (already a safe buildable surface, no search needed). |
| **Belongings chest beside the grave** | **Confirmed (empty)** | `/data get block 3 68 62` → `{id: "minecraft:chest", Items: []}`. Log: `belongings thing=Thing[... descriptor=Aldric's things ...] (0 item(s))` — Aldric's inventory was empty at death, so an empty chest was placed per `LiveDeathHandling.placeBelongings`'s documented "empty, not omitted" choice. This proves placement, not item-carryover (no fixture had given Aldric an inventory this session) — see "not observed" below. |
| **Grief hold started** | **Confirmed, `home` only** | `[death] grief hold started: Aldric's minecraft:home at BlockPos{x=15, y=71, z=78} until tick 4317`. Only one hold fired (not two) — Aldric's `Brain` had no `JOB_SITE` memory to release, consistent with TASK-0014's already-documented JOB_SITE-claim gap (`docs/wiki/villager-brain-api.md`, `full-cycle-observation.md`'s "work (tolerated wander)" row); `holdGrief` iterates `{HOME, JOB_SITE}` and each is independently `ifPresent`-gated, so this is the existing gap surfacing here, not a bug in this task's own code. |
| **Grief hold released after the configured window** | **Confirmed** | `[death] grief hold released: Aldric's minecraft:home at BlockPos{x=15, y=71, z=78} (tick 4317)` — released at exactly the tick the start log promised (`until tick 4317`), ~1200 ticks (60s) after the hold began, matching `-Dkithcraft.griefPeriodTicks=1200`. |
| **No siege ever fires, whole session** | **Confirmed** | `grep -in siege` over both console logs (pre-kill boot + post-kill run + restart) → 0 matches, across two full server lifetimes. A live zombie-count probe (`execute as @e[type=minecraft:zombie] run say ZOMBIE_HERE`) echoed exactly once, confirming the world's single pre-existing zombie never multiplied — no siege spawn-drip occurred. |
| **Retired token logged; never reissued** | **Confirmed** | `[death] retired body token b-5 for Aldric (never reissued)` at the moment of death. Restarting the server fresh, the boot registry log read: `{b-1=Aldric, b-2=Petra, b-3=Yenna, pl-0=..., pl-4=..., b-6=Yenna, b-8=grave of Aldric, b-7=Petra, pl-9=Aldric's grave}` — **`b-5` is absent**; `b-6`/`b-7` (Yenna's and Petra's own `DuskPairing` body tokens) persist. See the honest nuance below. |
| **Board posting composed (V4 seam)** | **Confirmed, at the composition boundary this task owns** | `[death] board posting composed: {text=Tend Aldric's grave., attributed_to=null}` — `GraveBoardEntry`'s own doc records V4 (the job-board fixture, TASK-0020) is not merged, so this correctly rides the existing text-percept read channel (§4.7) rather than a not-yet-real book block; nothing more was expected to be observed here this session. |
| **Grave sighting content composed** | **Confirmed** | `[death] grave sighting content: Thing[kind=k:grave, roles=[readable], descriptor=Aldric's grave, body=null, count=1]` — matches T010's percept-channel proofs (unit-tested; this is the same composition, now seen fired from a real death). |

## Honest nuance: two token families, only one retired

The boot registry after restart shows `b-1=Aldric` **still present** even though Aldric is
dead. This is not a bug in card AC #9 — `b-1` is `CastSeeder`'s original cast-identity
token, issued once at world-seed time and never touched by `LiveDeathHandling` (which only
knows about `DuskPairing`'s per-boot body token, `b-5`, retrieved via
`pairing.bodyTokenFor(villager.getUUID())`). The task's own scope (`LiveDeathHandling`'s
class doc, T007-T009) retires the token the death-handling/pairing machinery actually
tracks; `b-1` belongs to a separate, pre-existing cast-seeding concern this task's spec
never asked it to touch. Recorded plainly rather than glossed over — a future task
touching `CastSeeder`'s identity tokens should know `b-1`-style tokens are a second,
untouched token family.

## What was NOT observed

- **A real live zombie's own melee kill.** The kill was forced via `/damage ...
  minecraft:mob_attack` (see above) rather than waited on. The prior session's two
  passive-wait attempts did not produce one in reasonable real time; forcing was the
  deliberate, explicit choice this session made instead of waiting a third time.
- **Belongings carryover of a non-empty inventory.** Aldric's inventory was empty at
  death (no fixture in this world gave him hidden items), so the belongings chest placed
  was empty. `LiveDeathHandling.placeBelongings`'s item-copy loop is exercised by
  `ALLOW_DEATH`'s capture step either way (the loop ran, found zero non-empty stacks), but
  a *non-empty* capture was not observed live this session — covered structurally, not
  empirically, by this run alone (out of this session's scope to fabricate; a future
  session could `/give` Aldric an item pre-kill to close this specific gap if wanted).
- **A witnessing villager's live sighting percept of the grave, or an absent villager's
  `change_report`.** T010's unit tests already prove this composition deterministically
  (`DeathPerceptChannelTest`, 4 green tests including a real loopback `WireClient`
  session); this session did not additionally re-observe it live because no mind was
  dialed in (`kithcraft.sock` absent by design, same as prior phases' "body-keeps-moving"
  proofs) — there was no live percept *recipient* to observe receiving anything. This
  matches this task's own class-doc scope note: "Multiplexing percepts across more than
  one attached body is `BodySession`'s already-flagged single-body ceiling... T010 is
  where percept-channel delivery gets its live proof" — T010 is exactly where that proof
  lives, and it is unit-level, not this dev-server session's job to duplicate live.
- **The "tend the grave" posting actually taken up by a survivor.** `GraveBoardEntry.take()`
  exists and is callable but nothing in this session called it — no survivor agency was
  exercised or expected to be (card AC #6 explicitly requires it be optional/ignorable).
- **JOB_SITE grief hold.** Not observed for the reason given in the checklist above
  (Aldric had no JOB_SITE memory to release) — an existing, already-documented gap
  (`full-cycle-observation.md`), not new to this task.

## Gates

`./gradlew compileJava` green before the run (confirms the observability-logging commit
compiles). Full `./gradlew build` (all tests) is T012, run separately after this
observation and recorded there.
