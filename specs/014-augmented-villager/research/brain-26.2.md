# Brain<E> surface re-derivation against MC 26.2 (T001 / FR-001)

**Evidence rule**: every claim below is a direct `javap -public`/`javap -p` read against the
pinned jar, checked 2026-08-27. Jar: `~/.gradle/caches/fabric-loom/minecraftMaven/net/
minecraft/minecraft-merged-deobf/26.2/minecraft-merged-deobf-26.2.jar` (`net.minecraft.world.
entity.ai.Brain`, mappings: Mojang official, unobfuscated — see `specs/009-fabric-mod-
skeleton/research/versions.md`). Commands used are inlined per section so this is
independently re-runnable.

## 1. Brain<E> full public surface

```
javap -public -classpath minecraft-merged-deobf-26.2.jar net.minecraft.world.entity.ai.Brain
```

Confirms `docs/wiki/villager-brain-api.md`'s already-verified memory-module renames
unchanged, and closes every UNVERIFIED row:

- **Schedule is gone as a type.** No `Schedule` or `ScheduleBuilder` class exists anywhere in
  the jar (`jar tf | grep -i schedule` → only `net.minecraft.world.entity.schedule.Activity`
  and its package-info). `Brain.setSchedule` survives as a **name** but takes
  `EnvironmentAttribute<Activity>` — a generic environment-attribute value, the same
  mechanism vanilla now uses for fog color, sky light, ambient sounds, etc.
  (`net.minecraft.world.attribute.EnvironmentAttribute`/`EnvironmentAttributes`). There is no
  1:1 "schedule successor type" — a custom `Activity`-valued `EnvironmentAttribute` (built via
  `EnvironmentAttribute.builder(AttributeType<Activity>)`, populated with `Environment
  AttributeLayer.TimeBased`/`.Positional` layers — both classes exist in the jar) is the 26.2
  shape of "what activity is active when." **Full construction of a working schedule
  attribute is not derived here** — Phase 2 (T004) scope — this section only establishes the
  shape exists and is public API, no Mixin needed to *read or set* it.
- **Task-list assignment: `addActivity`, one method, wider shape.** Confirmed:
  `public void addActivity(Activity, ImmutableList<? extends Pair<Integer, ? extends
  BehaviorControl<? super E>>>, Set<Pair<MemoryModuleType<?>, MemoryStatus>>,
  Set<MemoryModuleType<?>>)`. No `setTaskList` successor of that name exists at all (matches
  the wiki's carried-forward finding, now confirmed against the live symbol, not inferred).
  `addActivity` is called on an already-built `Brain<E>` obtained from `Villager.getBrain()`
  (public) — plain API, no Mixin required to *add* to an activity's task list.
- **`addActivity` is additive, not a replace.** There is no `removeActivity`/`clearActivity`
  overload on `Brain<E>` (full method list above is exhaustive). This matters for Phase 2's
  suppression design (see §4) — calling `addActivity` again for `Activity.CORE` cannot be
  used to *drop* an already-registered behavior; only to add more.
- **`hasActivity()` gone; `isActive(Activity)` is the surviving query** — confirmed present.
- Other brain lifecycle methods present and unchanged in shape: `setCoreActivities`,
  `getActiveActivities`, `useDefaultActivity`, `setActiveActivityIfPossible`,
  `updateActivityFromSchedule(EnvironmentAttributeSystem, long, Vec3)`,
  `setDefaultActivity`, `tick(ServerLevel, E)`, `stopAll`, `isBrainDead`.

## 2. Sensor-driven refresh — confirmed, unchanged in shape

```
javap -public -classpath minecraft-merged-deobf-26.2.jar net.minecraft.world.entity.ai.sensing.Sensor
javap -public -classpath minecraft-merged-deobf-26.2.jar net.minecraft.world.entity.ai.sensing.SensorType
```

`Sensor<E>` is an abstract class with `abstract Set<MemoryModuleType<?>> requires()` and a
`final void tick(ServerLevel, E)` — the wiki's "Sensor-based memory refresh" shape holds
unchanged. `SensorType<U extends Sensor<?>>` is a plain static-field registry-key holder
(same pattern as `MemoryModuleType`); relevant villager-adjacent constants confirmed present:
`NEAREST_LIVING_ENTITIES`, `NEAREST_PLAYERS`, `NEAREST_BED`, `VILLAGER_HOSTILES`,
`VILLAGER_BABIES`, `SECONDARY_POIS` (backing `SecondaryPoiSensor`, the 26.2 name for the
Yarn-mapped `SecondaryPointsOfInterestSensor`), `GOLEM_DETECTED`. All plain API to *read*
(sensors run automatically once wired into a `Brain.Provider`); a custom sensor is a subclass
of `Sensor<E>` (public, no restriction) registered into
`BuiltInRegistries.SENSOR_TYPE` (public `DefaultedRegistry`) — **plain API, no Mixin**.

## 3. MemoryModuleType / POI registration — plain API, not Mixin-required (revises the wiki's carried-forward claim)

```
javap -p -classpath minecraft-merged-deobf-26.2.jar net.minecraft.world.entity.ai.memory.MemoryModuleType
javap -public -classpath minecraft-merged-deobf-26.2.jar net.minecraft.world.entity.ai.village.poi.PoiType
javap -public -classpath minecraft-merged-deobf-26.2.jar net.minecraft.core.registries.BuiltInRegistries | grep -iE "ACTIVITY|MEMORY_MODULE|POINT_OF_INTEREST|SENSOR"
```

**Sharpest revision of the wiki's carried-forward "Mixin/accessor required to extend" claim**:
at 26.2, `MemoryModuleType<U>` has a **public constructor**
(`public MemoryModuleType(Optional<Codec<U>>)`) and `PoiType` has a **public constructor**
(`public PoiType(Set<BlockState>, int, int)`). Both are registered into public registries
(`BuiltInRegistries.MEMORY_MODULE_TYPE : DefaultedRegistry<MemoryModuleType<?>>`,
`BuiltInRegistries.POINT_OF_INTEREST_TYPE : Registry<PoiType>`) via the standard
`Registry.register(Registry<V>, String|Identifier|ResourceKey, T)` static helper (public).
**Registering a new custom `MemoryModuleType` or `PoiType` needs no Mixin/accessor at 26.2** —
construct + `Registry.register`, same as any other Fabric content registration. This is new
information the Yarn-mapped source this wiki page carried forward did not settle either way
(V3 was never blocked on it, since V1/V2 needed none of this — but it changes what Phase 2/3
should reach for first).

## 4. Activity registration — Mixin/accessor IS required (the one confirmed case)

```
javap -p -classpath minecraft-merged-deobf-26.2.jar net.minecraft.world.entity.schedule.Activity
```

Unlike `MemoryModuleType`/`PoiType`, `Activity`'s constructor is **private**
(`private Activity(String)`) and its own registration helper is **private static**
(`private static Activity register(String)`). `BuiltInRegistries.ACTIVITY` is a public
`Registry<Activity>`, but there is no public way to construct the `Activity` value to put
into it. **Minting a brand-new `Activity` constant at 26.2 requires Mixin/accessor injection**
(an `@Invoker` accessor onto the private constructor or the private `register` method) — this
is the one extension point in this file's scope that is genuinely Mixin-required, not merely
carried-forward assumption.

**Consequence for Phase 3 (dusk pair-formation Activity, FR-005) — flagged, not resolved
here**: vanilla already ships an `Activity.MEET` constant (confirmed above, §1), and
`net.minecraft.world.entity.ai.behavior.VillagerGoalPackages.getMeetPackage(float)` (public
static, confirmed below) is the vanilla "villagers socialize" task-list package already
associated with it. **Phase 3 may not need a new `Activity` value at all** — the dusk
pair-formation behavior could be added as additional `BehaviorControl`s on the *existing*
`Activity.MEET` via `Brain.addActivity(Activity.MEET, ...)` (plain API), sidestepping the
Mixin-required private-constructor path entirely. This is a plan-relevant finding, not a
Phase 3 decision — left for that phase's own design.

## 5. Villager's own brain wiring — confirms the ≤3-override Mixin surface, narrows two of the three targets

```
javap -public -classpath minecraft-merged-deobf-26.2.jar net.minecraft.world.entity.ai.behavior.VillagerGoalPackages
javap -p -classpath minecraft-merged-deobf-26.2.jar net.minecraft.world.entity.npc.villager.Villager
```

- `VillagerGoalPackages` is public with static `getCorePackage`/`getWorkPackage`/
  `getPlayPackage`/`getRestPackage`/`getMeetPackage`/`getIdlePackage`/`getPanicPackage`/
  `getPreRaidPackage`/`getRaidPackage`/`getHidePackage`, each returning the fixed
  `ImmutableList<Pair<Integer, BehaviorControl<? super Villager>>>` vanilla feeds to
  `Brain.addActivity`. A `VillagerMakeLove` (breeding) `BehaviorControl` class exists
  (`net.minecraft.world.entity.ai.behavior.VillagerMakeLove`) — confirmed present in the core
  package's behavior set by class existence; it is vanilla's breeding task.
- **The actual wiring call is private and non-overridable by composition**:
  `Villager.registerBrainGoals(Brain<Villager>)` is a **private** instance method, called
  from `makeBrain(Brain.Packed)` (protected, but villagers here are vanilla `Villager`
  instances per decision-0002, not a custom subclass — so `protected` does not help). Combined
  with §1's finding that `addActivity` is additive-only (no remove), **there is no plain-API
  way to omit a specific vanilla behavior from a stock package** — confirms FR-004's Mixin
  task-list overrides are structurally necessary (Mixin `@Inject`/`@Redirect` into
  `registerBrainGoals`, or accessor-based reconstruction), not a design choice that 26.2 makes
  obsolete.
- **Golem-summoning is not an independent task-list entry — it is called from inside
  `gossip()`.** Bytecode disassembly of `Villager` (`javap -p -c`) shows
  `public void gossip(ServerLevel, Villager, long)` invokes `spawnGolemIfNeeded(ServerLevel,
  long, int)` directly (`invokevirtual ... spawnGolemIfNeeded`) at the end of its body — golem
  summoning is a side effect of gossip exchange, not a separate task in any
  `VillagerGoalPackages` list. **Flagged for Phase 2/T005, not resolved here**: a single Mixin
  injection point on `gossip()` (or on whatever calls `gossip()`, not yet traced — it is not
  called from within `Villager` itself) may suppress both gossip and golem-summoning in one
  override, which would leave FR-004's "≤3 task-list overrides" budget at only two real
  overrides (breeding via `registerBrainGoals`, gossip-which-also-suppresses-golem) rather
  than three symmetric ones — consistent with the plan's own instruction to drop an override
  that turns out unnecessary rather than keep it for symmetry. T005 should re-verify the
  caller of `gossip()` before finalizing the Mixin config.

## 6. Cast seeding surface (T002) — fully plain API, no Mixin

```
javap -public -classpath minecraft-merged-deobf-26.2.jar net.minecraft.world.entity.npc.villager.VillagerProfession
javap -public -classpath minecraft-merged-deobf-26.2.jar net.minecraft.world.entity.npc.villager.VillagerType
javap -public -classpath minecraft-merged-deobf-26.2.jar net.minecraft.world.entity.npc.villager.VillagerData
javap -public -classpath minecraft-merged-deobf-26.2.jar net.minecraft.world.entity.Entity | grep -i customname
```

- Profession × biome variant: `new VillagerData(Holder<VillagerType>, Holder<VillagerProfession>,
  int level)` (public constructor) then `Villager.setVillagerData(VillagerData)` (public).
  `VillagerProfession`/`VillagerType` constants (`ARMORER`, `FARMER`, ... ; `DESERT`,
  `PLAINS`, `SNOW`, ...) are public static `ResourceKey`s resolved through the server's
  registry access — standard Fabric registry lookup, no Mixin.
- Nameplates: `Entity.setCustomName(Component)` + `Entity.setCustomNameVisible(boolean)` —
  both public, unchanged in shape from prior Minecraft versions.
- Identity persistence: no vanilla `PersistentState` class exists at 26.2 (confirmed already
  by `TokenRegistryData`'s own doc comment); the mod's `SavedData`/`SavedDataType` pattern
  (`dev.kithcraft.mod.tokens.TokenRegistryData`) is the precedent T002 follows directly — a
  pure core registry class (cast identity: name → profession/type/level) plus a thin
  `SavedData` adapter, exactly mirroring `TokenRegistry`/`TokenRegistryData`.

## Summary table — Mixin/accessor vs plain API at 26.2

| Extension point | 26.2 verdict | Evidence |
|---|---|---|
| Read/write memory modules | Plain API | §1 (unchanged from wiki's already-verified rows) |
| Task-list assignment (`addActivity`) | Plain API to *add*; Mixin to *omit* a vanilla entry | §1, §5 |
| Sensor-driven refresh (read) | Plain API | §2 |
| New custom `Sensor`/`SensorType` | Plain API (subclass + `Registry.register`) | §2 |
| New custom `MemoryModuleType` | Plain API (public ctor + `Registry.register`) | §3 |
| New custom `PoiType` | Plain API (public ctor + `Registry.register`) | §3 |
| New custom `Activity` value | **Mixin/accessor required** (private ctor + private `register`) | §4 |
| Suppressing breeding/gossip/golem in `registerBrainGoals` | **Mixin required** (private method, additive-only `addActivity`) | §5 |
| Cast seeding (profession/type/name/nameplate) | Plain API | §6 |

This closes FR-001: every symbol `docs/wiki/villager-brain-api.md` flagged UNVERIFIED is now
either confirmed-present-and-plain, confirmed-gone, or confirmed-Mixin-required, with dated
evidence above. Nothing here types brain/schedule/Mixin code (out of Phase 1 scope) — this is
the derivation Phase 2/3 build from.
