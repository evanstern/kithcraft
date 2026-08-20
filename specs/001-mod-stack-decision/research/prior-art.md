# Prior-art re-verification — spec 001 Phase 1

All facts below were checked live on **2026-08-20** (accessed date for every citation
unless noted otherwise). This supersedes the 2026-08-19 brief citations for currency,
not correctness — no contradictions were found, only version/date drift.

## 1. Fabric (Loader + API)

- **Current stable Fabric Loader:** 0.19.3.
  https://github.com/FabricMC/fabric-loader/releases — accessed 2026-08-20.
- **Current Fabric API:** 0.158.0+26.2, built against Minecraft 26.2 (the latest MC
  release line as of this check).
  https://github.com/FabricMC/fabric-api/releases — accessed 2026-08-20.
- **Target Minecraft version support:** Fabric API ships per-MC-version builds; 26.2 is
  current, with parallel builds still published for 26.1.x and the 1.21.x line (e.g.
  1.21.11, 1.21.10) per CurseForge's version index.
  https://www.curseforge.com/minecraft/mc-mods/fabric-api — accessed 2026-08-20.
  Official notes for the current MC release: https://fabricmc.net/2026/06/15/262.html —
  accessed 2026-08-20.
- **Maintenance status:** actively maintained — Fabric API 0.158.0 shipped for the 26.2
  line, and Fabric Loader 0.19.3 is the current stable loader; both are FabricMC-org
  repos with continuous release cadence (multiple releases/month observed in the
  releases feed). No lapse in maintenance.
- **License:** Apache License 2.0, both fabric-loader and fabric-api.
  https://github.com/FabricMC/fabric-loader/blob/master/LICENSE — accessed 2026-08-20.
  https://github.com/FabricMC/fabric-api/blob/master/LICENSE — accessed 2026-08-20.

## 2. Paper

- **Current version:** Paper 26.2, build #112.
  https://papermc.io/downloads/paper — accessed 2026-08-20.
- **Target Minecraft version support:** tracks the current MC release line (26.2); the
  downloads/build-explorer service (fill.papermc.io) serves builds per-MC-version back
  through legacy lines.
  https://docs.papermc.io/misc/downloads-service/ — accessed 2026-08-20.
  Java requirement: 26.1+ needs Java 25 (per PaperMC's version/Java support table).
  https://docs.papermc.io/paper/getting-started/ — accessed 2026-08-20.
- **Maintenance status:** actively maintained — current build (#112) is fresh against
  the 26.2 MC line; PaperMC also runs Hangar (its own plugin repository) and a
  first-party downloads API, both live.
- **License:** GNU General Public License v3 (inherited from Spigot/Bukkit/CraftBukkit
  lineage); some individual contributors' code is dual-released under MIT — see the
  per-author list in LICENSE.md. Net effect: **GPLv3** governs the distributable
  artifact.
  https://github.com/PaperMC/Paper/blob/main/LICENSE.md — accessed 2026-08-20.

## 3. Citizens2

- **Current version:** 2.0.43 (free/GitHub-CI builds track the same line; premium
  Spigot builds are labeled e.g. "26.2 BUILD #2").
  https://ci.citizensnpcs.co/job/citizens2/ — accessed 2026-08-20.
  https://www.spigotmc.org/resources/citizens.13811/ — accessed 2026-08-20.
- **Target Minecraft version support:** native major version 1.21; tested major
  versions listed as 1.21, 26.1, 26.2 on the Spigot resource page.
  https://www.spigotmc.org/resources/citizens.13811/ — accessed 2026-08-20.
- **Maintenance status:** actively maintained but with real compatibility lag —
  the free GitHub repo's most recent commit found was 2026-02-16 ("Manually bump CB
  version," CitizensDev/Citizens2@6506c3e), about six months before this check; the
  GitHub repo publishes no formal Releases (CI + Spigot premium builds are the
  distribution channel instead). User reviews on the Spigot page as recently as this
  check report breakage on 1.21.10/1.21.11 requiring reloads or waiting for a fix, with
  the maintainer confirming Paper-stable-build compatibility issues are being worked.
  Net: actively developed, single-maintainer-led (fullwall), with a documented history
  of trailing new MC releases by some weeks.
  https://github.com/CitizensDev/Citizens2/commit/6506c3eff7440f11b4f68743f3fc025f902a1522 — accessed 2026-08-20.
  https://github.com/CitizensDev/Citizens2/releases (no releases published) — accessed 2026-08-20.
- **License:** Open Software License 3.0 (OSL-3.0).
  https://github.com/CitizensDev/Citizens2/blob/master/LICENSE — accessed 2026-08-20.
  **Note:** OSL-3.0 is a copyleft license with network-use provisions (similar in
  spirit to AGPL) and is *not* on the FSF/OSI list of licenses considered GPL-compatible
  for combined works — any code statically/tightly integrated with Citizens2 under
  OSL-3.0 needs its own license review before combining with a differently-licensed
  codebase. Flagged for the comparison doc.

## 4. CraftAgent (prskid1000/CraftAgent)

- **Current version / activity:** repo created 2025-12-29, last push 2026-01-06 — about
  seven and a half months stale as of this check (2026-08-20). Single contributor
  (prskid1000), 1 star, 1 fork, 1 open issue.
  https://github.com/prskid1000/CraftAgent — accessed 2026-08-20.
- **Target Minecraft version support:** README states requirements as Minecraft
  1.20.1 or 1.21.8, Fabric Loader 0.17.3+, Java 17/21 depending on MC version — **not**
  updated to the current 26.x line.
  https://github.com/prskid1000/CraftAgent (README, "Requirements" section) — accessed
  2026-08-20.
- **Maintenance status:** effectively stalled. No commits in ~7.5 months, no MC-version
  bump past 1.21.8, single maintainer, minimal community uptake (1 star). This is a
  **documented dead end as load-bearing prior art** — usable as a design reference
  (its feature set: SQLite conversation memory, world perception, action handlers,
  Brigadier command execution, web dashboard) but not as a dependency to build on
  without a fork/revival.
- **License:** GNU Lesser General Public License v3.0 (LGPL-3.0).
  https://github.com/prskid1000/CraftAgent — accessed 2026-08-20 (GitHub license
  detection + README "This project is licensed under LGPL-3.0.").

  **Adjacent, more actively maintained sibling project found during this check:**
  sailex428/SecondBrain (formerly/aka "ai-npc") — same category (Fabric mod, LLM-driven
  NPCs), created 2024-10-09, **last push 2026-03-31**, 46 releases, latest v3.1.6
  (2026-03-25), LGPL-3.0. Materially more active than CraftAgent and worth flagging to
  the comparison doc as a live alternative in the same design space.
  https://github.com/sailex428/SecondBrain — accessed 2026-08-20.

## 5. AI_NPC

- **Current version:** 2.0.0, "First Public" release, published 2026-07-23. Supported
  platform range listed as Paper 1.18.2–26.2.
  https://hangar.papermc.io/NNNNTX/AI_NPC/versions/2.0.0 — accessed 2026-08-20.
- **Target Minecraft version support:** 1.18.2 through 26.2 (per Hangar's platform
  compatibility field), i.e. current as of this check.
  https://hangar.papermc.io/NNNNTX/AI_NPC — accessed 2026-08-20.
- **Maintenance status:** very new — this is the *first public release* (2.0.0, one
  month old at check time), 34 downloads. Too early to assess a maintenance cadence;
  no track record of sustained updates yet.
- **License: could not be verified.** The Hangar listing page does not surface a
  license field in its rendered content, and no linked public source-code repository
  was found from the Hangar page during this check (searches for a GitHub/GitLab repo
  under the author handle "NNNNTX" plus "AI_NPC" did not surface one distinct from the
  Hangar listing itself). **Documented dead end:** license status is unresolved and
  must be confirmed directly from the plugin author (or its distributed jar's
  plugin.yml/LICENSE, if obtained) before AI_NPC could be relied on as a dependency.
  https://hangar.papermc.io/NNNNTX/AI_NPC — accessed 2026-08-20.

## 6. Fabric villager brain API surface

Confirmed via Yarn-mapped Mojang API docs (mirrors vanilla, exposed to Fabric mods) and
a community Fabric wiki tutorial demonstrating the pattern end-to-end.

- **What's available, natively in the vanilla `Brain<E>` class** (Fabric mods read/call
  this directly; no Mixin needed to *use* an existing brain):
  - `Schedule` get/set (`getSchedule()`/`setSchedule()`) — the activity timetable.
  - `Activity` queries (`hasActivity()`) and per-activity task-list assignment
    (`setTaskList(Activity, ...)` overloads, several arities including
    required/forgetting memory sets).
  - `MemoryModuleType` read/write: `hasMemoryModule()`, `getOptionalMemory()`,
    `remember()`, `setMemory()`, `forget()`, `isMemoryInState()`,
    `hasMemoryModuleWithValue()`.
  - `Sensor`-based memory refresh (e.g. `SecondaryPointsOfInterestSensor`, POI-related
    sensors that populate memory modules each tick).
  https://maven.fabricmc.net/docs/yarn-1.21.3+build.1/net/minecraft/entity/ai/brain/Brain.html — accessed 2026-08-20.
- **What requires Mixin/accessor injection to *extend* (not just use)**: registering
  new custom `Activity` values, new `MemoryModuleType`s, new `PointOfInterestType`s, and
  wiring a custom `Activity` into a specific entity's brain init and into a
  `ScheduleBuilder`-modified `Schedule` — all demonstrated end-to-end (with working
  Mixin/accessor code) in the Fabric community wiki's villager-activities tutorial,
  built against a real mod (ReligiousVillagersMinecraftMod).
  https://wiki.fabricmc.net/tutorial:villager_activities — accessed 2026-08-20.
- **Net assessment:** the brain/schedule/memory/POI substrate is real, documented, and
  scriptable from a Fabric mod without engine forking — reading/using existing
  schedules and memories is plain API access; adding *new* activities, memory types, or
  POI types requires Mixin accessors (standard Fabric practice, not a blocker, but not
  zero-friction either). This confirms the brief's claim that "Fabric exposes
  activity/schedule injection into real villager brains" — injection is real but goes
  through Mixin for anything beyond using what vanilla already defines.

## Summary table

| Dependency | Version (as of 2026-08-20) | MC support | Maintenance | License | Verified? |
|---|---|---|---|---|---|
| Fabric Loader | 0.19.3 | current (26.2 line) | active | Apache-2.0 | Yes |
| Fabric API | 0.158.0+26.2 | 26.2 (+ parallel 26.1.x/1.21.x builds) | active | Apache-2.0 | Yes |
| Paper | 26.2 build #112 | 26.2 | active | GPLv3 (+ opt-in MIT per-contributor) | Yes |
| Citizens2 | 2.0.43 | native 1.21, tested through 26.2 | active but lagging (last free-repo commit 2026-02-16; user-reported breakage on newest MC builds) | OSL-3.0 (copyleft, GPL-incompatible) | Yes, with caveats |
| CraftAgent | last push 2026-01-06 | 1.20.1 / 1.21.8 (not current) | **stalled** (~7.5 months, single contributor) | LGPL-3.0 | Yes — documented dead end |
| AI_NPC | 2.0.0 (2026-07-23) | 1.18.2–26.2 | too new to assess | **unverified** — no license found | Partial — documented dead end on license |
| Fabric villager brain API | n/a (engine API) | current (Yarn mapped through 1.21.3+; pattern stable across versions checked) | n/a (vanilla engine surface) | n/a | Yes |
