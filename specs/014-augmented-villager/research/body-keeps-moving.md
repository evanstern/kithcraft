# T010 — body-keeps-moving: structural audit + stalled-mind live proof

**Scope**: Phase 4 (FR-006, card AC #5). Two parts: (a) a structural audit that no
schedule/activity/brain code path blocks awaiting a mind response, (b) a live dev-server
proof with the mind stalled.

## Structural audit

Grep across `mod/src/main/java/dev/kithcraft/mod/brain/` and `.../cast/` for any reference to
`WireClient`, `BodySession`, or blocking I/O:

```
$ grep -rln "WireClient\|BodySession" mod/src/main/java/dev/kithcraft/mod/
mod/src/main/java/dev/kithcraft/mod/brain/DuskPairing.java   (javadoc mention only)
mod/src/main/java/dev/kithcraft/mod/cast/CastSeeder.java     (javadoc mention only)
```

Neither `ScheduleSetup`, `DuskPairing`, `Cast`, `CastData`, nor `CastSeeder` — the entire
brain/schedule/cast surface driving the three named villagers — imports or calls into
`WireClient`/`BodySession`/`IntentHandler` anywhere in executable code. The cast's
wake/work/socialize/sleep is driven entirely by vanilla `Brain<E>` machinery plus this mod's
one-time fixture/memory setup (`ScheduleSetup.placeFixtures`, `DuskPairing.setUp`) and a
per-tick, purely local `DuskPairing.tick()` (reads `Brain.isActive`/`PathNavigation`, no I/O).
**The mind daemon and the cast's scheduled bodies are structurally disjoint** in this
codebase's current state — the schedule cannot block on the mind because it never talks to
it at all.

The one place the mind IS consulted (`dev.kithcraft.mod.live.BodySession`, T011's single-body
live-actuator wiring, unrelated to the cast) was also audited:

- `KithcraftMod.onServerTickUnsafe` calls `attached.tick(worldTime)` unconditionally every
  tick, wrapped in the `[live] onServerTick threw` catch-all. `BodySession.tick` calls
  `intents.tick(worldTime)` (local state-machine advance, no I/O read) and conditionally
  `emitter.emit(...)` (a **write** to the socket — see below).
- `IntentHandler.tick`/`.handle` never read from the socket; they only produce envelopes that
  `BodySession`/`PerceptEmitter` write out. No method in `act/` or `live/` calls
  `WireClient.receive()`.
- `WireClient.receive()` (the one blocking call in this codebase) is only ever invoked from
  `BodySession.readLoop`, run on its own **daemon thread** (`new Thread(..., "kithcraft-live-
  reader")`), never on the server tick thread. A mind that never sends anything leaves that
  one reader thread parked in a blocking read — harmless, since the tick thread never joins or
  waits on it.
- `BodySession.open` itself is called from the tick thread but only on an *attach* attempt
  (`KithcraftMod.onServerTickUnsafe`), and `WireClient.dial()` (a **connect**, not a blocking
  read) either succeeds immediately or throws `IOException` immediately if nothing is
  listening — no retry loop runs inline; the retry is the tick-throttled attach loop, not a
  blocking wait.

**Conclusion**: no code path in this codebase awaits a mind response before letting the body
proceed. The cast's schedule doesn't touch the mind at all; the one live-actuator session that
does touch it only ever *writes* from the tick thread and *reads* from an isolated daemon
thread.

## Live proof: the mind never connects (chosen over connect-then-silent)

**Which variant the code supports, and why this one was run.** The dispatch allows either "a
stub mind that connects and goes silent" or "no mind at all, if the mod tolerates absent
connection." The mod tolerates absent connection by construction — `BodySession.open` throws
on a failed dial, `KithcraftMod.onServerTickUnsafe` catches nothing there (the call sites are
already inside a `try`/catch at the tick-loop level) and simply leaves `attached == null`,
retried every 20 ticks forever (`lastAttachAttempt` throttle). This is also the *stronger* form
of "stalled": the connection never even opens, for the whole run, rather than opening once and
then going quiet. Running the connect-then-silent variant as well would have required a second
full ~20-minute cycle under this session's time budget (the dispatch names one run for
T009+T010+T011 together); the reader-thread isolation argument above covers that variant
structurally without needing a second live run.

**How it was run.** `./gradlew runServer` in the background (`mod/run/`, `pause-when-empty-
seconds=-1`, `eula=true`, `online-mode=false`, world reused from the prior session's persisted
state — cast already seeded, `run/stdin.fifo` for console commands, no `kithcraft.socket`
property set so `BodySession.open` looks for `kithcraft.sock` relative to the run dir, which
does not exist). Total run: 24m42s wall-clock (`BUILD SUCCESSFUL in 24m 42s`), world tick
advanced from ~29020 to ~53400+ (a full 24000-tick day-length span; also T011's run).

**What was observed, verbatim:**

- `[live] mind dial failed, will retry: java.net.SocketException: No such file or directory`
  — **1,474 occurrences**, one per throttled attach attempt (~every 20 ticks), across the
  entire 24m42s run. The dial never once succeeded.
- `[live] attach scan at tick N: 3 villager(s), ...` — **1,475 occurrences**, confirming the
  attach loop itself kept running normally the whole time, never wedged waiting on the failed
  dial.
- **Zero** `ERROR`-level log lines and zero uncaught exceptions of any kind for the entire run
  (`grep -c "ERROR\]"` → 0). The `[live] onServerTick threw` safety net (added in TASK-0012)
  never fired once.
- The server's own tick counter (via `time query gametime`, sampled repeatedly) advanced
  monotonically at the expected ~20 ticks/real-second throughout — no stall, no slowdown,
  matching 24m42s real time to ~24000+ game ticks elapsed.
- All three cast villagers (`Aldric`, `Petra`, `Yenna`) remained present and their `Brain`
  memories (`meeting_point`, `home`) queryable via `/data get entity` the entire run — the
  cast's own schedule/brain machinery (independent of T011's finding below) was never
  observed to stop functioning.

**What this proves.** With the mind permanently, unrecoverably stalled at the connection step
for the full run: the server tick loop never blocked, never errored, never slowed; the cast's
brain-driven bodies remained queryable and alive throughout (T011's own full-cycle
observation, run concurrently in this same session, is the direct evidence for their
schedule/activity state); and the retry loop itself proves resilience rather than a hang.
**Card AC #5 holds**: a stalled mind does not convert a 20-second thought into a 20-second (or
24-minute) freeze — nothing here ever froze.

## Gates

`./gradlew build` (compile + full test suite, `--rerun-tasks` to force a real run rather than
an up-to-date short-circuit) green: 111 tests, 0 failures, 0 errors, across all existing
suites (`CastTest`, `MixinConfigTest`, `PairingSignalTest`, `IntentHandlerTest`, etc.) —
unmodified, since T010 required no production code change (the structural property already
held from Phase 1–3's own design, per the audit above).
