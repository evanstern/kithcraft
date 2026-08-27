---
name: villager-brain-api
description: The vanilla Brain<E> substrate Fabric mods drive at MC 26.2 (Mojang official mappings, fully re-derived) — what is plain API access (MemoryModuleType read/write, Sensor-driven refresh, and as of this pass MemoryModuleType/PoiType/Sensor registration too) vs what genuinely requires Mixin/accessor injection (new Activity values; omitting a vanilla behavior from a stock task-list package). The engine surface TASK-0014's brain/schedule/cast work builds on. Load for entity/brain implementation work.
kind: component
sources:
  - specs/001-mod-stack-decision/research/prior-art.md
  - docs/design/mod-stack-comparison.md
  - specs/012-vendor-conformance/research/verb-observation.md
  - specs/014-augmented-villager/research/brain-26.2.md
verified_against: 2b3535131ed00533cffaabdc3749b6143dff0599
---

# Villager brain API (Fabric substrate)

The vanilla `Brain<E>` class is the Elder-Scrolls-schedule substrate that already exists
in-engine: activities, schedules, memory modules, and points of interest, all scriptable
from a Fabric mod without engine forking. Under [[mod-stack-decision]] this is the
substrate both of TASK-0003's entity sub-options sit on.

## How it works

**Fully re-derived against MC 26.2, 2026-08-27 (TASK-0014 Phase 1, T001)** — `javap
-public`/`javap -p` against the exact pinned jar (`~/.gradle/caches/fabric-loom/
minecraftMaven/net/minecraft/minecraft-merged-deobf/26.2/minecraft-merged-deobf-26.2.jar`)
closes every row TASK-0012 Phase 3/4 left UNVERIFIED. Full evidence, commands, and dates:
`specs/014-augmented-villager/research/brain-26.2.md`. Nothing below is carried forward as
a shape any more — every claim is a direct symbol read.

**Memory-module surface — renamed, not removed** (Yarn name → MC 26.2 name, `Brain<E>`) —
unchanged from the prior pass, still holds:

| Yarn (yarn-1.21.3+build.1) | MC 26.2 (Mojang official, verified) |
|---|---|
| `hasMemoryModule()` | `hasMemoryValue(MemoryModuleType<?>)` |
| `getOptionalMemory()` | `getMemory(MemoryModuleType<U>)` → `Optional<U>` |
| `remember()` | `setMemory(MemoryModuleType<U>, U)` / `setMemoryWithExpiry(..., long)` |
| `setMemory()` | `setMemory()` — name unchanged, now three overloads (incl. an `Optional`-accepting form) |
| `forget()` | `eraseMemory(MemoryModuleType<U>)` |
| `isMemoryInState()` | `checkMemory(MemoryModuleType<?>, MemoryStatus)` |
| `hasMemoryModuleWithValue()` | `isMemoryValue(MemoryModuleType<U>, U)` |

**Schedule/task-list surface — gone, not renamed.** No `Schedule`/`ScheduleBuilder` class
exists anywhere in the 26.2 jar at all (confirmed by a full jar symbol scan, not just a
missed name guess). `setSchedule` survives as a name but takes
`net.minecraft.world.attribute.EnvironmentAttribute<Activity>` — the same generic
environment-attribute mechanism vanilla now uses for fog color, sky light, ambient sound,
etc. — a structurally different shape, not a rename; a custom `Activity`-valued attribute
(via `EnvironmentAttribute.builder(AttributeType<Activity>)`, populated with a
`TimeBased`/`Positional` layer) is 26.2's schedule shape, though a full working
construction is Phase 2's job, not derived here. `hasActivity()` is gone; `isActive
(Activity)` is the closest surviving query. `setTaskList(Activity, ...)` has **no
successor of that name** — task-list assignment goes through `addActivity(Activity,
ImmutableList<Pair<Integer, BehaviorControl<? super E>>>, Set<Pair<MemoryModuleType<?>,
MemoryStatus>>, Set<MemoryModuleType<?>>)`, a wider-shaped call, confirmed the only such
method on `Brain<E>` — **additive only, no remove/replace counterpart**. This remains the
sharpest finding across both passes: a mechanical Yarn→Mojang rename table would not have
caught it, because there is no 1:1 target to rename to.

**Sensor-driven refresh — confirmed, unchanged in shape.** `Sensor<E>` (abstract,
`requires(): Set<MemoryModuleType<?>>`, a `final tick(ServerLevel, E)`) and `SensorType<U
extends Sensor<?>>` (plain static-field registry-key holder) both exist as expected;
villager-relevant constants confirmed present (`NEAREST_LIVING_ENTITIES`,
`NEAREST_PLAYERS`, `NEAREST_BED`, `VILLAGER_HOSTILES`, `VILLAGER_BABIES`, `SECONDARY_POIS`
— the 26.2 name for the Yarn-mapped `SecondaryPointsOfInterestSensor` is
`SecondaryPoiSensor`). A custom sensor is a plain subclass registered into
`BuiltInRegistries.SENSOR_TYPE` (public `DefaultedRegistry`) — plain API, no Mixin.

**Extension points — the blanket "Mixin/accessor required" claim from the Yarn-mapped pass
does not hold uniformly at 26.2; it splits per type:**

- **`MemoryModuleType` and `PoiType` — plain API, no Mixin.** Both have public constructors
  at 26.2 (`MemoryModuleType<U>(Optional<Codec<U>>)`, `PoiType(Set<BlockState>, int, int)`)
  and register into public registries (`BuiltInRegistries.MEMORY_MODULE_TYPE`,
  `BuiltInRegistries.POINT_OF_INTEREST_TYPE`) via the standard
  `Registry.register(Registry<V>, id, instance)` helper. This revises the prior carried-
  forward assumption — it was never load-bearing (V1/V2 needed neither), but Phase 2/3
  should reach for construct-and-register first, not for a Mixin.
- **New `Activity` values — Mixin/accessor genuinely required.** `Activity`'s constructor
  is private and its own `register(String)` helper is private static; `BuiltInRegistries.
  ACTIVITY` is a public registry with no public way to construct what goes into it. Minting
  a new `Activity` constant needs an accessor mixin onto the private constructor or
  register method. **Likely moot for Phase 3's dusk pair-formation Activity**: vanilla
  already ships `Activity.MEET` plus a public `VillagerGoalPackages.getMeetPackage(float)`
  task-list package for it — adding behaviors to the existing `Activity.MEET` via
  `Brain.addActivity(Activity.MEET, ...)` avoids minting a new Activity at all. Left for
  Phase 3 to decide, not resolved here.
- **Omitting a vanilla behavior from a stock task-list package (breeding/gossip/golem
  suppression, FR-004) — Mixin required.** `Villager.registerBrainGoals(Brain<Villager>)`,
  the method that actually calls `addActivity` with vanilla's packages, is **private**, and
  `addActivity` is additive-only (confirmed above) — there is no plain-API way to leave a
  specific vanilla `BehaviorControl` (e.g. `VillagerMakeLove`, the breeding task) out of a
  package once `registerBrainGoals` runs. This confirms FR-004's bounded Mixin surface is
  structurally necessary, not a design choice 26.2 makes obsolete. One further finding for
  Phase 2 to weigh: golem-summoning is not an independent task at all —
  `Villager.gossip(ServerLevel, Villager, long)` calls `spawnGolemIfNeeded(...)` directly
  from inside its own body (confirmed by bytecode disassembly), so a single Mixin
  injection on `gossip()` may suppress both gossip and golem-summoning, potentially leaving
  the "≤3 task-list overrides" budget at two real overrides rather than three.

The brief's claim that "Fabric exposes activity/schedule injection into real villager
brains" is confirmed at the symbol level, not just the shape level: injection is real,
free for memory modules/POI types/sensors, and Mixin-gated only for new `Activity` values
and for removing what vanilla already wired in. TASK-0012 V2 needed none of this:
`act/Verbs.java`'s live wiring (T011) drives movement and look via
`Mob.getNavigation()`/`Mob.getLookControl()` — plain `Mob` API, not `Brain<E>` at all — so
none of the above blocked V1/V2 and this pass is entirely TASK-0014's own scoped work.

**Cast seeding (T002, plain API, no Mixin)**: profession × biome variant via `new
VillagerData(Holder<VillagerType>, Holder<VillagerProfession>, int level)` +
`Villager.setVillagerData(...)`; nameplates via `Entity.setCustomName`/
`setCustomNameVisible`; identity persistence via the mod's `SavedData`/`SavedDataType`
pattern (same one `TokenRegistryData` established) — see `dev.kithcraft.mod.cast`.

## Connections

Substrate chosen by [[mod-stack-decision]]; the "reflex" half of the reflex/planner split
([[promptworld-lineage]]) largely maps onto this machinery; the mind daemon reaches it
only through [[body-protocol-seam]]; health/citations in [[prior-art]]; the live `Mob`
API TASK-0012 V2 actually drives (sidestepping this note's still-open gaps) is documented
in [[body-protocol-seam]]'s "First implementations" section.

## Operational notes

Every symbol on this page is now re-verified directly against the pinned MC 26.2 jar
(`javap -public`/`javap -p`, dated 2026-08-27) — nothing above is carried forward from the
Yarn-mapped docs (yarn-1.21.3+build.1) as a shape any more; Yarn does not exist for MC 26.2
at all (mappings regime changed, not just version drift — see `specs/009-fabric-mod-
skeleton/research/versions.md`'s "villager-brain-api.md re-check" section, TASK-0009, which
flagged this page for full re-derivation). Full command-by-command evidence:
`specs/014-augmented-villager/research/brain-26.2.md`. Citations for the pre-26.2 baseline
this page evolved from: `specs/001-mod-stack-decision/research/prior-art.md` §6.
