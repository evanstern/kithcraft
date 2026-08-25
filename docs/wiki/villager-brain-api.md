---
name: villager-brain-api
description: The vanilla Brain<E> substrate Fabric mods drive — what is plain API access (MemoryModuleType read/write, Sensor-driven refresh) vs what requires Mixin/accessor injection (new Activities, MemoryModuleTypes, POI types), plus the MC 26.2 symbol renames TASK-0012 Phase 3/4 verified by javap. The engine surface TASK-0003's entity work builds on. Load for entity/brain implementation work.
kind: component
sources:
  - specs/001-mod-stack-decision/research/prior-art.md
  - docs/design/mod-stack-comparison.md
  - specs/012-vendor-conformance/research/verb-observation.md
verified_against: 4ac4264f8a16623fdc30d94347b4c7c1bee1ab23
---

# Villager brain API (Fabric substrate)

The vanilla `Brain<E>` class is the Elder-Scrolls-schedule substrate that already exists
in-engine: activities, schedules, memory modules, and points of interest, all scriptable
from a Fabric mod without engine forking. Under [[mod-stack-decision]] this is the
substrate both of TASK-0003's entity sub-options sit on.

## How it works

**Partially re-verified against MC 26.2, 2026-08-25 (TASK-0012 Phase 3/4)** — `javap
-public` against the exact pinned jar (`~/.gradle/caches/fabric-loom/minecraftMaven/net/
minecraft/minecraft-merged-deobf/26.2/minecraft-merged-deobf-26.2.jar`,
`net.minecraft.world.entity.ai.Brain`) confirmed the memory-module surface renamed and the
schedule/task-list surface is **gone**, not merely renamed. Everything else below this
section (Sensor-driven refresh, the Mixin-required extension list, POI/Activity
registration) is **still UNVERIFIED as of 2026-08-25 — re-derivation needed before use**,
carried forward only as a *shape*, not as verified-current symbol names. Full re-derivation
remains V3's scoped work (TASK-0014); nothing beyond the table below should be typed into
brain/schedule code before that pass.

**Memory-module surface — renamed, not removed** (Yarn name → MC 26.2 name, `Brain<E>`):

| Yarn (yarn-1.21.3+build.1) | MC 26.2 (Mojang official, verified) |
|---|---|
| `hasMemoryModule()` | `hasMemoryValue(MemoryModuleType<?>)` |
| `getOptionalMemory()` | `getMemory(MemoryModuleType<U>)` → `Optional<U>` |
| `remember()` | `setMemory(MemoryModuleType<U>, U)` / `setMemoryWithExpiry(..., long)` |
| `setMemory()` | `setMemory()` — name unchanged, now three overloads (incl. an `Optional`-accepting form) |
| `forget()` | `eraseMemory(MemoryModuleType<U>)` |
| `isMemoryInState()` | `checkMemory(MemoryModuleType<?>, MemoryStatus)` |
| `hasMemoryModuleWithValue()` | `isMemoryValue(MemoryModuleType<U>, U)` |

**Schedule/task-list surface — gone, not renamed.** No `getSchedule()` exists at all on MC
26.2's `Brain<E>` (confirmed absent from the full `javap -public` symbol list, not just
missed by a name guess). `setSchedule` survives as a name but its parameter type changed
from a `Schedule` to `net.minecraft.world.attribute.EnvironmentAttribute<Activity>` — a
structurally different shape, not a rename. `hasActivity()` is also gone; `isActive
(Activity)` is the closest surviving query. `setTaskList(Activity, ...)` has **no
successor of that name** — task-list assignment now goes through `addActivity(Activity,
ImmutableList<Pair<Integer, BehaviorControl<? super E>>>, Set<Pair<MemoryModuleType<?>,
MemoryStatus>>, Set<MemoryModuleType<?>>)`, a wider-shaped call. This is the sharpest
finding of the two passes: a mechanical Yarn→Mojang rename table would not have caught it,
because there is no 1:1 target to rename to.

**Still UNVERIFIED against MC 26.2 — carried forward as a *shape* only** (originally
verified against Yarn-mapped docs, `net.minecraft.entity.ai.brain.Brain`):

- `Sensor`-based memory refresh (e.g. `SecondaryPointsOfInterestSensor`), populating
  memory modules each tick.
- Mixin/accessor injection required to *extend*: registering new custom `Activity`
  values, new `MemoryModuleType`s, new `PointOfInterestType`s, and wiring a custom
  Activity into an entity's brain init and a `ScheduleBuilder`-modified `Schedule`. The
  pattern is demonstrated end-to-end (working Mixin code) in the Fabric community wiki's
  villager-activities tutorial (pre-26.2), Standard Fabric practice — not a blocker, but
  owned engineering surface, recorded as an accepted risk of the stack decision.

The net assessment from the research still holds at the *shape* level: the brief's claim
that "Fabric exposes activity/schedule injection into real villager brains" is confirmed —
injection is real but goes through Mixin for anything beyond what vanilla already defines.
TASK-0012 V2 needed none of this: `act/Verbs.java`'s live wiring (T011) drives movement and
look via `Mob.getNavigation()`/`Mob.getLookControl()` — plain `Mob` API, not `Brain<E>` at
all — so the schedule/task-list gap above did not block V2 and is flagged for V3 rather
than fixed here.

## Connections

Substrate chosen by [[mod-stack-decision]]; the "reflex" half of the reflex/planner split
([[promptworld-lineage]]) largely maps onto this machinery; the mind daemon reaches it
only through [[body-protocol-seam]]; health/citations in [[prior-art]]; the live `Mob`
API TASK-0012 V2 actually drives (sidestepping this note's still-open gaps) is documented
in [[body-protocol-seam]]'s "First implementations" section.

## Operational notes

The memory-module and schedule/task-list findings above are re-verified against MC 26.2
directly (`javap -public`, cited jar path above) — trust those rows. Everything else
(Sensor refresh, Mixin-required extension) is still carried forward from Yarn-mapped docs
(yarn-1.21.3+build.1) and stale: Yarn does not exist for MC 26.2 at all (mappings regime
changed, not just version drift — see `specs/009-fabric-mod-skeleton/research/
versions.md`'s "villager-brain-api.md re-check" section). Re-verify the remaining symbols
against Mojang's official 26.2 mappings before any Sensor/Mixin/POI code is written; this
is flagged for V3 (TASK-0014), not resolved here. Citations with URLs and dates:
`specs/001-mod-stack-decision/research/prior-art.md` §6.
