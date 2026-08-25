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
