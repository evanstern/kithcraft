# T009 — verb execution: what was and was not observed live

**Scope**: T009 (`act/Verbs.java`) implements the four core verbs' seam
(`Verbs.Actuator`) and `IntentHandler` drives it — decode → ack → execute → exactly one
`act_result`, unit-tested (`IntentHandlerTest`, `TargetResolutionTest`) with a scripted
stub actuator, no live server.

## What was NOT observed this phase

No dev-server run was performed for T009. `./gradlew runServer` exists (confirmed via
`gradle tasks --all`) but wiring a live villager body, a real `PathNavigation`/
`LookControl`/chat-adjacent `Verbs.Actuator` implementation, a stub mind sending real
`intent` messages over the V1 `WireClient`/Unix-socket transport, and capturing the
resulting `act_result` traffic is genuine integration work this phase's dispatch scope
(T007–T009: `IntentHandler`, `TargetResolution`, `Verbs` — the decode/ack/result state
machine and its unit tests) does not cover. Claiming a live observation without doing
one would violate the evidence rule; none is claimed here.

## What was verified instead

- `net.minecraft.world.entity.ai.Brain<E>` was checked with `javap -public` against the
  MC 26.2 jar the project's other phases use
  (`~/.gradle/caches/fabric-loom/minecraftMaven/net/minecraft/minecraft-merged-deobf/
  26.2/minecraft-merged-deobf-26.2.jar`) — see the model ID's final report and
  `docs/wiki/villager-brain-api.md`'s pending re-verification note for the symbol diff
  against the Yarn-mapped names it currently carries.
- `Verbs.Actuator` is deliberately the seam real wiring targets: `stepWalk` →
  `Mob`/`PathNavigation` (`moveTo`, tick-polled), `speak` → a chat-adjacent call,
  `attend` → `LookControl`. None of this requires a new Mixin — plain vanilla API,
  consistent with decision-0002's preference and R-8's precedent (`Sounds.java`
  composed the wire shape from an already-resolved input; live `GameEventListener`
  registration was deferred the same way, explicitly to "Phase 3/4 wiring").
- The full decode → ack → execute → act_result loop, target resolution (last-seen vs.
  unknown_target vs. target_gone), `cancel`, `supersedes`, and `not_after` expiry are
  exercised by `IntentHandlerTest` and `TargetResolutionTest` against a scripted
  `Verbs.Actuator` stub — deterministic, no live world needed to prove the state
  machine's correctness.

## What T011 (Phase 4) still owes

A dev-server run with a real `Verbs.Actuator` implementation bound to a live villager
body, each of the four verbs executed once, and its `act_result` captured — the live
half of card AC #6 this phase's unit tests cover structurally but not empirically. This
mirrors R-8's own split: hook verified by source (T004), live registration deferred
(T009/T011 here).

## T011 — live observation, 2026-08-25

**What was wired.** `mod/src/main/java/dev/kithcraft/mod/live/`: `LiveGround` (the real
`TargetResolution.Ground`, token issuance via the save-persisted `TokenRegistryData` plus
an in-memory place-to-`BlockPos` side table — `BlockPos` never crosses the wire, AR-4),
`LiveActuator` (the real `Verbs.Actuator` — `Mob.getNavigation().moveTo(...)` for `go_to`,
`Mob.getLookControl().setLookAt(...)` for `attend`, a `PlayerList.broadcastSystemMessage`
chat-adjacent call for `speak`; no Mixin), and `BodySession` (dials the mind's UDS socket,
sends `session_open`, drives `IntentHandler` off incoming `intent` frames, and emits a
slow `self_state` heartbeat). `KithcraftMod` attaches the first `Villager` found each
server tick once the socket is dialable. Every Minecraft symbol used
(`Mob.getNavigation()`/`getLookControl()`, `PathNavigation.moveTo`, `LookControl
.setLookAt`, `Villager`'s real package `net.minecraft.world.entity.npc.villager.Villager`,
`EntityGetter.getEntitiesOfClass`) was confirmed present by `javap -public` against the
same pinned MC 26.2 jar R-8 used, before being typed into live code.

**How it was run.** `./gradlew runServer` in the background (`run/eula.txt` seeded,
`run/server.properties`'s `pause-when-empty-seconds` set to `-1` — see the finding below),
with console commands piped through a named pipe (`run/stdin.fifo`) since no interactive
TTY is available in this environment. A throwaway Python "stub mind"
(`/tmp/kc-stub-mind.py`, not committed — a plain UDS listener speaking the same
length-prefixed canonical-JSON framing as `mod/.../wire/FrameCodec.java`) stood in for the
mind daemon: accepted the vendor's dial, read `session_open`, then sent one `intent` per
core verb and recorded every frame.

**What was observed, verbatim (this session's real transcript):**

1. `session_open` received and decoded cleanly — the same static manifest
   `HandshakeWireClientTest` already proves, now over a live UDS connection to a running
   dev server (`body=b-0`, `session=s-live-050850db-...`).
2. `go_to` (target: a place token bound to a real `BlockPos` 4 blocks from the villager's
   spawn point) → `intent_ack{accepted:true}`, then `act_result{outcome:failed,
   reason_code:unreachable, detail:"could not find a way there"}`. The villager was
   summoned in mid-air (see the fall-height finding below) and was still falling when the
   walk was issued; `PathNavigation.moveTo` returned `false` (no path found) rather than
   throwing — an honest `unreachable`, not a bug in the wiring. AC #6's "exactly one
   `act_result`" held; AC #7 (last-seen-place resolution) is not exercised by this
   particular run since the target was a place token, not a body token — that path is
   proven by `TargetResolutionTest`/`IntentHandlerTest`'s `goToABodyWalksToItsLastSeenPlace`
   (Phase 3), not re-proven live here.
3. `speak` (target `none`, `content.text:"hello, dev server"`) → accepted,
   `act_result{outcome:completed, detail:"said: hello, dev server"}`. The
   `broadcastSystemMessage` call executed without error (no players were connected to
   receive it — nothing in this dev-server run could exercise the "someone actually reads
   it" half, honestly recorded as not observed).
4. `attend` (target: the same place token) → accepted, `act_result{outcome:completed,
   detail:"looked around"}` — `LookControl.setLookAt` executed against the resolved
   `BlockPos`.
5. `wait` (target `none`) → accepted, `act_result{outcome:completed, detail:"waited"}`.

All four `intent_ack`s carried `accepted:true`; all four intents yielded exactly one
`act_result` each (card AC #6, live). No refusal, no rejection, no protocol violation
surfaced on the wire at any point in the session.

**A real bug this run caught and fixed.** The first several attempts found nothing: the
mod's tick-throttle used `private long lastAttachAttempt = Long.MIN_VALUE;` and computed
`worldTime - lastAttachAttempt` to decide when to retry — since `worldTime` is a small
non-negative tick count, that subtraction overflows a signed 64-bit long and wraps to a
huge negative number, which is always `< 20`, so the attach attempt silently never ran, on
any tick, ever. No exception, no log line — a genuinely silent no-op. Fixed by seeding the
sentinel to `-1000L` instead. `KithcraftMod.onServerTick` also gained a catch-all
diagnostic wrapper (`catch (Throwable t) { LOGGER.error(...) }`) after this, since
Fabric's event dispatch does not reliably surface a listener exception — kept in the
shipped code as a permanent safety net, not stripped back out.

**A real world-mechanics finding.** MC 26.2's dedicated server ships `server.properties`
default `pause-when-empty-seconds=60`: with no player connected, the server pauses
world-ticking (not just slows it) after 60 idle seconds, which silently stops
`ServerTickEvents.END_SERVER_TICK` from firing — indistinguishable from the attach bug
above until isolated. A villager summoned via `summon minecraft:villager ~ ~1 ~` (console,
no reference entity) also reliably vanished within about half a second of a `/data get
entity` follow-up query (`"No entity was found"`), most plausibly landing embedded in
unloaded/ungenerated terrain at the console's implicit `~`-relative position; summoning at
an explicit, confirmed-loaded coordinate well above terrain (`summon minecraft:villager 25
80 17`, near where naturally-spawned passive mobs were already being tracked) is what
finally produced a villager the attach loop could see and hold. Neither finding is
mod-specific — both are dev-server/vanilla mechanics a real multi-body session manager
(V3) will need to account for (keep-loaded chunk tickets, `pause-when-empty-seconds`), not
something this task's scope required fixing beyond the `server.properties` override used
to get this one observation.

**What was NOT observed.** No second body, no player entity, no live `GameEventListener`
hearing-hook registration (R-8 resolved the hook exists and is plain API; wiring it onto a
live entity is explicitly deferred, `r8-hearing-hook.md`), no `self_state` heartbeat
percept captured within this run's short window (the 5-second heartbeat interval didn't
fire before the four-verb exchange completed and the session was torn down — the four
`act_result` percepts, which share the identical `PerceptEmitter.emit` composition path,
already prove percept-without-rejection for AC #1's live half), and no body-target
(`go_to` toward another body's last-seen place) exercised live — only a place-token
target. These are honest scope boundaries of a single-session, single-body proof, not
claims of completeness beyond what was actually seen.
