# T009 — dev-server live-wiring observation (card AC #7/#8)

**Scope**: Phase 3. `docs/design/demo-runbook.md` followed (Order A, daemon first), rehearsal
mode throughout — **`ANTHROPIC_API_KEY` was not present in this session's environment**
(checked before boot: `[ -n "$ANTHROPIC_API_KEY" ]` false), so per this dispatch's own
instruction the operator's key was not sanctioned for this run and **zero live model calls**
were made or expected (`-genesis=false`, stub personas pre-seeded exactly like
`specs/021-demo-config/research/bringup-observation.md`'s own precedent). Confirmed
zero-call at teardown: `session-report.log`'s E1-E6 rows are all `calls=0`.

The honest result: bring-up, session attach, and the mod's own live percept emission were all
confirmed working; **no E2/E3 trigger and no dusk exchange were observed reaching the daemon
this session** — not from timing alone, but from two structural findings this observation
surfaced (§3, §4), one of which (§3) is new and worth an operator's attention before Phase 4.

## 0. Setup

```
cd mind && go build -o minddaemon ./cmd/minddaemon
cd mod && ./gradlew build -x test
mkdir -p mind/run mod/run && echo eula=true > mod/run/eula.txt
```

Stub personas pre-seeded by hand (`mind/run/persona/{Aldric,Petra,Yenna}.json`, 0444,
`cast_id`/`name`/`profession`/`biome_variant` matching `persona.DemoCast()`) — no
`ANTHROPIC_API_KEY`, so a real genesis cannot run; this is the documented rehearsal fallback
(demo-runbook.md §4), re-bind via `Load`, zero model calls.

`docs/wiki/villager-brain-api.md`'s chunk-ticket trap did not need any manual action this
run: `CastSeeder.keepChunksLoaded` re-covers the cast's *actual current* footprint every
boot automatically (`KithcraftMod.onServerTickUnsafe`, confirmed by reading the call site —
no `/forceload` command was ever issued by this observation).

**Runbook gotcha reconfirmed** (bringup-observation.md's own §0 finding, hit again here): a
fresh `mod/run/server.properties` defaults `pause-when-empty-seconds=60`, which paused the
whole tick loop ~60s after boot with no player connected — `"Server empty for 60 seconds,
pausing"` in `mod/run/server.log`. Fixed per demo-runbook.md §2 (`pause-when-empty-seconds=-1`)
and the mod server restarted; world/cast persisted across the restart (same villager UUID
`4726341b-...` re-attached, confirming idempotent re-attach same as bringup-observation.md's
own second-boot section).

## 1. Bring-up — confirmed

Order A (daemon first): `mind/minddaemon -socket "$SOCK" -rundir "$(pwd)/mind/run"
-genesis=false` → `minddaemon: listening on .../mod/run/kithcraft.sock`, then
`./gradlew runServer`. Mod log:

```
[live] token registry live entries at boot: {}
[live] attach scan at tick 1200: 3 villager(s), ... entities
[dusk] Aldric body token: b-11
[dusk] Petra body token: b-12
[dusk] Yenna body token: b-13
[live] session_open sent: session=s-live-... body=b-14 place=pl-15
[live] attached to villager 4726341b-13db-4aa7-b993-281cb4b7f4f0
```

Confirmed via `/data get entity 4726341b-... CustomName` → `"Aldric"`: the one
live-attached villager (`BodySession`'s single body, see §3) is Aldric.

## 2. What was observed

| Item | Outcome |
|---|---|
| Both artifacts boot independently, either order | Confirmed (Order A here; Order B already confirmed in specs/021's own observation) |
| Session attaches, cast bound (rehearsal) | Confirmed — daemon re-binds all three stub personas via `Load`, zero calls |
| World/cast persist across a mod-server restart | Confirmed — same villager UUID re-attached |
| `self_state` heartbeat emitted mod-side | Confirmed by reading `BodySession.tick` — called every server tick, fires immediately on attach (`lastHeartbeat` starts at `Long.MIN_VALUE`) and every 100 ticks (~5s) after; no wire error in either server log across the whole session |
| Zero live model calls | Confirmed — `session-report.log`'s E1-E6 rows all `calls=0 cancelled=0 ...=0` |
| Clean shutdown produces the session report | Confirmed — SIGTERM to the daemon after a graceful mod `stop` wrote `session-report.log` with a per-villager E6/E4 section for both boot lifetimes' body tokens (`b-6`, `b-14`) |

## 3. New finding: the live body token is never a CastID — persona binding cannot occur on the live-attached body

`runtime.go`'s own top-of-file comment (T004, Phase 2) names "a body token that literally IS
a loaded CastID" as one of only two real binding signals the wire carries, calling it "the
demo/dev-server convention DemoCast's ids already give bodies." This session checked that
claim against the actual live mod and it does not hold:

- `BodySession.open` (`mod/.../live/BodySession.java`) mints the live session's body token via
  `ground.issueBody("a villager", villager.blockPosition())` — an opaque per-boot token
  (`b-14` this run, `b-6` the run before), with `"a villager"` as its registry referent, not
  a cast name. Confirmed live: `[live] session_open sent: ... body=b-14 ...`, and the boot
  registry dump shows `b-6=a villager` (generic) sitting alongside `b-3=Aldric`, `b-4=Petra`,
  `b-5=Yenna` (`DuskPairing`'s own, *separate* token family, whose registry referent IS the
  cast name — see §4).
- No code path in the mod sets a `BodySession`'s body token, or anything else the
  `session_open` message carries, to a cast member's name (`grep`-verified: only
  `LiveDeathHandling` ever calls `villager.getName()`, for an unrelated grief-hold label).
- `HandleSessionOpen` (`runtime.go`) binds a persona with exactly `rt.Personas[body]` — a
  literal map lookup keyed by CastID. For the one live-attached body this session, that is
  `rt.Personas["b-14"]`, which is never populated (the map only has `"Aldric"`/`"Petra"`/
  `"Yenna"` keys). **The live-attached body's own persona can never bind via this path.**

The other binding path (`conversation.go`'s `bindPersonaIfUnbound`, keyed off a pairing
signal's `otherName`) is unaffected by this — it keys off the display name directly, not the
body token — but it only ever binds the *other* pair member's `bodyStore`, and per §4 no
live pairing signal is ever sent at all, so it never fires live either.

**Judgment (card AC #8): not glue-sized, not fixed in this dispatch.** Closing this needs a
real decision — e.g. the mod passing the matched `Cast.Member`'s name into `BodySession.open`
so `session_open`'s body token (or a session_open extension) carries it, or the daemon
learning a name↔token mapping some other way — either is a design call, not an
obviously-correct one-liner, and FR-008 reserves that judgment for an operator checkpoint,
not a Phase 3 implementer. Flagged here for Phase 4/the next task, not silently patched.

This also means T008's daemon-level tests (which script `body="Aldric"` etc. directly, per
their own doc comments) are proving the WIRING correctly, but on an idealized body-token
convention the live mod does not currently produce — worth knowing when citing those tests
as "through the real daemon binary": real binary, yes; real mod-issued body tokens, no.

## 4. Dusk exchange — not observed, and provably cannot be with the current mod build

`DuskPairing`'s own class javadoc (`mod/.../brain/DuskPairing.java`) already says so,
unprompted by anything this session did:

> ponytail: firing logs the composed `sighting` content (T009's dev-server evidence) but
> does not yet send it over a live `WireClient` session — `BodySession`'s own doc already
> flags this mod's live wiring as single-body only.

Confirmed structurally: `KithcraftMod`'s `attached` field is a single `BodySession`, not a
collection (`private BodySession attached;`), and `BodySession`'s own class doc: "Single-body
scope deliberately... not the multi-body session manager V3 needs." A dusk pair needs BOTH
members' own live sessions to send their half of the pairing signal (conversation.go's own
package doc: "DuskPairing fires this TWICE per approaching pair, once per perceiver"); with
only one villager ever live-attached, the second signal can never be sent over the wire at
all, regardless of how long this session waited. No pairing signal fired in this session's
window anyway (`grep -c "\[dusk\] pairing signal"` → 0 across both server logs — expected,
`CycleTicks=24000` at 20 ticks/s is a ~20-minute day, far past this observation's budget), so
this finding rests on reading the code, not on a timeout.

**This is a pre-existing, already-documented ceiling** (the class doc cites its own earlier
"T009," a different task's dev-server evidence, not this one) — not something TASK-0023
introduced or regressed. Recorded here because FR-007 asks this observation to try, and
because it's the concrete reason the "live first/second-signal ordering" and
"`pairConvergeTimeout` 30s calibration" questions this dispatch asked about **could not be
checked live this session** — there is no live second signal to order or time out.

## 5. E2/E3 live trigger — not observed reaching the daemon

`self_state`'s heartbeat is confirmed emitted mod-side (§2) and `triggerE2` (deliberation.go)
returns `true` unconditionally for any `self_state` percept — so if one reaches
`HandlePercept`, `runDeliberation` prints `"deliberation trigger for body %q but no
ANTHROPIC_API_KEY — skipped (rehearsal mode)"` to the daemon's stderr immediately (rehearsal
mode's honest skip, not a crash).

Across three bounded checks (~30s apart, ~75s total, spanning well past the 5s heartbeat
interval, after `pause-when-empty-seconds=-1` was confirmed in effect) the daemon log showed
**nothing beyond the startup `listening on ...` line**, and `mind/run/villagers/b-14.jsonl`
stayed at 0 bytes throughout. This reconfirms, with a longer and cleaner window,
`specs/021-demo-config/research/bringup-observation.md` §5's exact open question: "whether
the heartbeat is even reaching the daemon at all... was not distinguished within this
observation's budget." No wire error appeared in either server log, so the connection itself
was healthy; whether the heartbeat is silently dropped somewhere in the seam layer or never
actually written to the socket despite the mod's own log believing it emitted is **still not
root-caused** — packet-level capture on the UDS socket would be needed to go further, out of
this dispatch's budget, matching the prior observation's own deferral.

**A board-read (lectern) E3 trigger was not attempted**: `BoardVisit`'s AI only fires during
`Activity.MEET`, the same real-tick-only substrate dependency as the dusk exchange (plan.md's
Risks note); with §3's binding gap meaning the trigger would skip on "no bound persona" even
if it fired, and §5's heartbeat question already open, forcing a board visit within this
session's bounded-check budget was not attempted as not worth the time — an honest scope
choice, not an oversight.

## 6. FirstTokenLatency in session-report.log — not observed live (already proven daemon-level)

No dusk exchange ran (§4), so no `FirstTokenLatency` row was produced this session —
`session-report.log`'s per-villager section reads `(no dusk exchange turns)` for both `b-6`
and `b-14`. This requirement is fully proven at the daemon level instead, through the real
listener/wire, by T008's `TestDuskExchange_PairSignalConvergesAndRecordsLatency` (both
speakers' latencies asserted present in the report text) — see `tasks.md` T008/T007.

## 7. Adjacent-gap judgments (card AC #8)

- **§3 (body-token-vs-CastID binding)**: new finding this session, not glue-sized, not fixed
  here — flagged for the operator/Phase 4.
- **§4 (DuskPairing never sends live)**: pre-existing, already documented in the mod's own
  code before this session ran — reconfirmed, not new, out of scope by the same class doc's
  own citation.
- **§5 (self_state reaching the daemon)**: pre-existing open question
  (specs/021-demo-config), reconfirmed with more evidence (a longer window, pause fixed) but
  still not root-caused — out of scope for the same reason the prior observation deferred it.
- No Mixins added, no protocol extension made, no code changed by this observation (it is a
  read-and-log session; the only writes were the stub persona fixtures and the two run
  directories, both gitignored).

## Gates

`go build -o minddaemon ./cmd/minddaemon` and `./gradlew build -x test` both green before
this run. All processes (daemon, gradle/server JVM, the stdin-fifo holder) were stopped
cleanly at the end of this session (mod `stop` console command, then `SIGTERM` to the
daemon); `git status` clean of this observation's artifacts (the built `minddaemon` binary
was removed; `mind/run/` and `mod/run/` are both gitignored).
