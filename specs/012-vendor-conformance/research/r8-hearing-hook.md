# R-8 — hearing hook verification (T004)

**Evidence rule**: source-level verification against the toolchain on disk, cited by file and
symbol. Checked 2026-08-25 against the exact jars `versions.md` (specs/009) pinned for this
project: `~/.gradle/caches/fabric-loom/minecraftMaven/net/minecraft/minecraft-merged-deobf/
26.2/minecraft-merged-deobf-26.2.jar` (Mojang official mappings, unobfuscated — no Yarn exists
for 26.2) and `~/.gradle/caches/modules-2/.../fabric-api-0.158.0+26.2.jar`.

## Outcome: VERIFIED

MC 26.2 ships a public, native "hearing" substrate — the same mechanism vanilla uses for sculk
sensors and the Warden — under `net.minecraft.world.level.gameevent`. Symbols confirmed by
`javap -public` against the jar above:

- `GameEvent` (public final class, `Holder<GameEvent>` constants incl. `STEP`, `ENTITY_ACTION`,
  `SHRIEK`, `PROJECTILE_SHOOT`, etc.) — the closed set of "something happened here" events
  vanilla already fires for ordinary world activity.
- `GameEventListener` (public interface: `getListenerSource()`, `getListenerRadius()`,
  `handleGameEvent(ServerLevel, Holder<GameEvent>, GameEvent.Context, Vec3)`) — the callback a
  mod implements to be notified.
- `DynamicGameEventListener<T extends GameEventListener>` (public class: `add(ServerLevel)`,
  `remove(ServerLevel)`, `move(ServerLevel)`) — the entity-attached registration helper; it
  tracks the listener across chunk-section boundaries as its `PositionSource` moves, which is
  exactly what an entity (as opposed to a fixed block) needs.
- `EntityPositionSource` (public class implementing `PositionSource`, constructor
  `(Entity, float)`) — a `PositionSource` that follows a live entity, e.g. a villager body.
- `ServerLevel#gameEvent(Holder<GameEvent>, Vec3, GameEvent.Context)` (public method) — the
  entry point vanilla itself calls to broadcast a sound-worthy event; nothing about it is
  private or internal.
- `VibrationSystem.User` (public interface: `getListenerRadius()`, `getPositionSource()`,
  `canReceiveVibration(...)`, `onReceiveVibration(ServerLevel, BlockPos, Holder<GameEvent>,
  Entity, Entity, float)`) — the higher-level convenience shape vanilla's Warden uses for
  exactly the "hear a direction/distance, not the cause" pattern §4.4 specifies. Available as
  an alternative to a bare `GameEventListener` if the richer travel-time/suspicion bookkeeping
  is ever wanted; not required for the minimal shape this task emits.

All constructors and methods above are `public`. Attaching a listener to a villager body is:
construct a `GameEventListener` (or use `EntityPositionSource` + a listener implementation),
wrap it in `new DynamicGameEventListener<>(listener)`, call `.add(serverLevel)`. No Mixin is
required — this is plain API, inside decision-0002's preference for events/reads over new
Mixins.

## What this resolves for §4.4

- `sound` is declared in the manifest (T006) and its content/provenance composed by
  `percept/Sounds.java`.
- Live registration of a `GameEventListener` onto an actual villager entity (turning a fired
  vanilla `GameEvent` into an outbound `sound` percept end-to-end) is **Phase 3/4 wiring**, once
  a live villager body exists to attach the listener to — consistent with `PerceptEmitter`'s
  existing ponytail note that session/world wiring lands when a body is actually connected.
  Phase 2's `Sounds.java` composes the wire shape from an already-resolved `(sound_kind,
  bearing, distance, descriptor)` so that wiring, whenever it lands, has a tested landing spot.

## Not used: `fabric-sound-api-v1`

Fabric API 0.158.0+26.2 bundles `fabric-sound-api-v1` (confirmed present in the fabric-api
jar's `META-INF/jars/`), but its contents (`FabricSoundInstance`, a client-side
`SoundEngineMixin`) are for **playing** custom sounds client-side, not for detecting that a
sound occurred — irrelevant to R-8's question. The vanilla `gameevent` package above is the
relevant hook.
