---
name: villager-brain-api
description: The vanilla Brain<E> substrate Fabric mods drive — what is plain API access (Schedule get/set, Activity task lists, MemoryModuleType read/write, Sensor-driven refresh) vs what requires Mixin/accessor injection (new Activities, MemoryModuleTypes, POI types). The engine surface TASK-0003's entity work builds on. Load for entity/brain implementation work.
kind: component
sources:
  - specs/001-mod-stack-decision/research/prior-art.md
  - docs/design/mod-stack-comparison.md
verified_against: cb5003531192512f992688cbf5a08d7b15799545
---

# Villager brain API (Fabric substrate)

The vanilla `Brain<E>` class is the Elder-Scrolls-schedule substrate that already exists
in-engine: activities, schedules, memory modules, and points of interest, all scriptable
from a Fabric mod without engine forking. Under [[mod-stack-decision]] this is the
substrate both of TASK-0003's entity sub-options sit on.

## How it works

**UNVERIFIED as of 2026-08-25 — re-derivation needed before use.** Every symbol below was
checked at `yarn-1.21.3+build.1`. TASK-0009's version re-verification
(`specs/009-fabric-mod-skeleton/research/versions.md`) found Yarn **discontinued**: MC
26.1+ ships unobfuscated with Mojang's official names directly, and Fabric confirms names
moved (not just mapping style) in that transition, e.g. `ItemGroupEvents` →
`CreativeModeTabEvents`. The target here (MC 26.2) is two major lines past Yarn's last
release. No symbol below has been checked against Mojang's 26.2 official names — this
section's claims are carried forward as a *shape* (what's plain API vs what needs Mixin),
not as verified-current symbol names. **Re-deriving the exact 26.2 names is V3's scoped
work** (TASK-0014); nothing here should be typed into Mixin/brain code before that pass.

**Plain API access — no Mixin needed to *use* an existing brain** (originally verified
against Yarn-mapped docs, `net.minecraft.entity.ai.brain.Brain`; names below are Yarn's,
not yet re-checked against MC 26.2's official mappings):

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

API names above are Yarn mappings (checked at yarn-1.21.3+build.1) and are now stale:
Yarn does not exist for the target MC 26.2 at all (mappings regime changed, not just
version drift — see `specs/009-fabric-mod-skeleton/research/versions.md`'s
"villager-brain-api.md re-check" section, 2026-08-25). Re-verify every symbol against
Mojang's official 26.2 mappings before any brain/schedule/Mixin code is written; this is
flagged for V3 (TASK-0014), not resolved here. Citations with URLs and dates:
`specs/001-mod-stack-decision/research/prior-art.md` §6.
