---
name: villager-brain-api
description: The vanilla Brain<E> substrate Fabric mods drive — what is plain API access (Schedule get/set, Activity task lists, MemoryModuleType read/write, Sensor-driven refresh) vs what requires Mixin/accessor injection (new Activities, MemoryModuleTypes, POI types). The engine surface TASK-0003's entity work builds on. Load for entity/brain implementation work.
kind: component
sources:
  - specs/001-mod-stack-decision/research/prior-art.md
  - docs/design/mod-stack-comparison.md
verified_against: 50c3def435dd9326d38e51118f08944815cbe80c
---

# Villager brain API (Fabric substrate)

The vanilla `Brain<E>` class is the Elder-Scrolls-schedule substrate that already exists
in-engine: activities, schedules, memory modules, and points of interest, all scriptable
from a Fabric mod without engine forking. Under [[mod-stack-decision]] this is the
substrate both of TASK-0003's entity sub-options sit on.

## How it works

**Plain API access — no Mixin needed to *use* an existing brain** (verified against
Yarn-mapped docs, `net.minecraft.entity.ai.brain.Brain`):

- `Schedule` get/set (`getSchedule()` / `setSchedule()`) — the activity timetable.
- `Activity` queries (`hasActivity()`) and per-activity task-list assignment
  (`setTaskList(Activity, ...)` overloads, including required/forgetting memory sets).
- `MemoryModuleType` read/write: `hasMemoryModule()`, `getOptionalMemory()`,
  `remember()`, `setMemory()`, `forget()`, `isMemoryInState()`,
  `hasMemoryModuleWithValue()`.
- `Sensor`-based memory refresh (e.g. `SecondaryPointsOfInterestSensor`), populating
  memory modules each tick.

**Mixin/accessor injection required to *extend*:** registering new custom `Activity`
values, new `MemoryModuleType`s, new `PointOfInterestType`s, and wiring a custom
Activity into an entity's brain init and a `ScheduleBuilder`-modified `Schedule`. The
pattern is demonstrated end-to-end (working Mixin code) in the Fabric community wiki's
villager-activities tutorial. Standard Fabric practice — not a blocker, but owned
engineering surface, recorded as an accepted risk of the stack decision.

The net assessment from the research: the brief's claim that "Fabric exposes
activity/schedule injection into real villager brains" is confirmed — injection is real
but goes through Mixin for anything beyond what vanilla already defines.

## Connections

Substrate chosen by [[mod-stack-decision]]; the "reflex" half of the reflex/planner split
([[promptworld-lineage]]) largely maps onto this machinery; the mind daemon reaches it
only through [[body-protocol-seam]]; health/citations in [[prior-art]].

## Operational notes

API names above are Yarn mappings (checked at yarn-1.21.3+build.1); exact symbol names
can shift across MC versions — re-verify mappings against the target MC version when
implementation starts. Citations with URLs and dates:
`specs/001-mod-stack-decision/research/prior-art.md` §6.
