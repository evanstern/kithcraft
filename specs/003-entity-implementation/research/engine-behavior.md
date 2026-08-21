# Engine-behavior evidence — spec 003 Phase 1

All facts below were checked live on **2026-08-21** (accessed date for every citation unless
noted otherwise). Version sensitivity: minecraft.wiki mechanics pages are not tied to a
specific game version in their citation metadata and describe current Java Edition behavior;
where a source is version-pinned (Yarn class docs), the mapping build is stated. This builds
on — does not duplicate — [[villager-brain-api]] (Brain<E> substrate) and spec 001's
`prior-art.md` (Fabric/Paper/Citizens2 dependency health). Per [[mod-stack-decision]], both
sub-options here sit on the same vanilla Fabric brain/Mixin substrate: "custom entity" means
a Fabric entity class wired into `Brain<E>`, not a bot client or a Citizens2 NPC.

## 1. Hostile-mob targeting of villagers

**Targeting is by Java class/type check against `VillagerEntity`/`AbstractVillagerEntity`
(Merchant), not a generic tag.** The zombie goal list includes an "Active Target - Merchant"
goal (priority 3) that targets `VillagerEntity` or `WanderingTraderEntity` specifically;
"Active Target - Iron Golem" (priority 3) is a separate goal.
https://minecraft.wiki/w/Mob_AI — accessed 2026-08-21.

- Zombies (and husks, drowned, zombie villagers) seek out and attack villagers within a
  35–52.5-block radius (Java, depends on regional difficulty) and can see them through walls;
  once a zombie has targeted a villager it ignores the player and other villagers until the
  target dies, the zombie is attacked, or the player gets close.
  https://minecraft.wiki/w/Zombie — accessed 2026-08-21.
- The brain-based `VillagerHostilesSensor` independently scans `visible_mobs` for Drowned
  (8 blocks), Evoker (12), Husk (8), Illusioner (12), Pillager (15), Ravager (12), Vex (8),
  Vindicator (10), Zoglin (10), Zombie (8), Zombie Villager (8) and stores the nearest as the
  `nearest_hostile` memory — this is what drives the villager's own panic/flee behavior
  (distinct from the attacker's goal-based targeting above). This sensor list is a fixed set
  of vanilla entity types.
  https://minecraft.wiki/w/Mob_AI — accessed 2026-08-21.
- Raids: villagers flee illagers to the nearest house with a door and a bed; a villager rings
  the bell to warn the village before hiding (Java); ringing gives nearby illagers Glowing.
  https://minecraft.wiki/w/Villager — accessed 2026-08-21.

**What this means for the two options:** goal-based attacker AI (`NearestAttackableTargetGoal<VillagerEntity>` equivalents baked into zombie/husk/drowned goal selectors) is hardcoded to
the vanilla `VillagerEntity`/`AbstractVillagerEntity` class hierarchy — it is **not** a
registry/tag lookup. A custom Fabric entity that does not extend `VillagerEntity` (or at
least `AbstractVillagerEntity`/`MerchantEntity`) is **invisible to these attacker goals
by default**: zombies will not add it as a target unless the mod adds its own
`NearestAttackableTargetGoal` mixed into hostile mobs' goal selectors, or the custom entity
subclasses a type the vanilla goals already check. The `VillagerHostilesSensor` that drives
the villager's *own* panic is independent of the class question — a custom entity would need
its own sensor (or reuse of `VillagerHostilesSensor`, which is written to run against the
villager's `visible_mobs` memory, not tied to `VillagerEntity` internally as far as verified
here — **not independently re-verified against source**, flagged as thin evidence). Augmenting
vanilla `VillagerEntity` inherits both halves (being-targeted and self-panic-sensing) for
free; a custom entity built as a fresh class must Mixin into every hostile's goal selector
(or extend `AbstractVillagerEntity`) to be attacked at all, and separately wire its own
hostile-sensing.

## 2. Village fiction mechanics (beds/sleep, workstations, POI, schedules)

- **POI model:** beds (`minecraft:home`), bells (`minecraft:meeting`), and job-site blocks are
  the three POI types villagers claim; each villager gets exactly 3 claim slots (bed,
  gathering site, job site) in the village's POI map.
  https://minecraft.wiki/w/Village_mechanics — accessed 2026-08-21.
- **Job-site claiming (Java):** unemployed villagers search a 48-block sphere for an unclaimed
  job site, pathfind to it with a provisional (temporary) claim, and fully claim it within a
  2-block radius; claim emits green particles; profession is set/changed by the claimed block.
  https://minecraft.wiki/w/Villager (§Profession) — accessed 2026-08-21.
- **Sleep:** villagers head home before sunset, target a spot beside their claimed bed, and
  sleep at night; they wake on food thrown at them, being pushed out, bed destruction, being
  attacked, or a bell ring, and try to resume sleep after interruption.
  https://minecraft.wiki/w/Villager (§Sleeping) — accessed 2026-08-21.
- **Schedule:** villagers follow a fixed tick-based `Schedule` keyed to employment/age status
  (Work/Wander/Gather/Sleep phases at specific tick marks) — this is the vanilla
  `Schedule`/`Activity` machinery [[villager-brain-api]] already confirmed as plain
  get/set API on `Brain<E>`.
  https://minecraft.wiki/w/Villager (§Schedule) — accessed 2026-08-21.

**Free vs must-wire per option:** vanilla `VillagerEntity` ships with the job-site claim logic,
the POI-claim bookkeeping (3-slot map), the sleep/bed pathing, the fixed profession-based
`Schedule`s, and the sensors that populate the memories those schedules' task lists read
(`SecondaryPoiSensor`, etc.) — all free by inheritance/augmentation. A custom Fabric entity
that does not extend `VillagerEntity` gets **none of this**: it must independently implement
POI claim/pathing logic, register (or reuse, via Mixin-exposed accessors) the same
`MemoryModuleType`s and `Sensor`s, and either build a `Schedule`/`ScheduleBuilder` from
scratch or Mixin-copy vanilla's, per [[villager-brain-api]]'s existing finding that new
Activities/MemoryModuleTypes/POI types require Mixin/accessor injection regardless of the
entity class. **The delta specific to entity choice** is the claim/pathing/sleep *tasks*
themselves (the `Task<E>` implementations bound into `VillagerEntity`'s activity task lists)
— those are only "free" if the custom entity extends `VillagerEntity` (or its `AbstractVillagerEntity`/`Brain`-init methods are Mixin-exposed and re-invoked on the custom class), not merely by wiring the same memory/sensor names.

## 3. Death/despawn/permadeath

- **Despawning:** villagers are passive mobs. Passive mobs spawn on a much rarer cycle and,
  per the mob-switch tutorial, "Passive mobs in Java Edition don't despawn, but count towards
  the mob cap" — the distance-based random despawn chance (1/800 per tick beyond 32 blocks)
  documented on the mob-spawning page applies to monster/ambient/aquatic mobs, explicitly
  excluding passives by category.
  https://minecraft.wiki/w/Mob_spawning — accessed 2026-08-21.
  https://minecraft.wiki/w/Tutorial:Mob_switch — accessed 2026-08-21.
  **Practical read for the brief's permadeath goal:** villagers already will not randomly
  vanish from distance/time — "permadeath" in this project is a design constraint about death
  being final and mattering, not a despawn-prevention problem the engine causes. (One
  exception found: all mobs, including persistent ones, despawn instantly at Y ≤ -128, and
  hostile-mob-category despawn rules apply if a villager mob is ever miscategorized — n/a for
  a `VillagerEntity` subclass, worth re-checking if a custom entity is registered under a
  different `MobCategory`, since the Fabric entity tutorial shows entities register with an
  explicit `MobCategory` at creation.
  https://wiki.fabricmc.net/tutorial:entity — accessed 2026-08-21.)
- **Zombie-villager conversion on death (fiction-breaking behavior #1):** killing blow from a
  zombie/husk/drowned/zombie villager (and zombified piglin, Java only) converts the villager
  instead of killing it outright, at 0% (Easy) / 50% (Normal) / 100% (Hard) chance.
  https://minecraft.wiki/w/Villager (§Attacking) — accessed 2026-08-21.
  This is a fork inside vanilla's damage-handling path on `VillagerEntity` (effectively
  `LivingEntity.onDeath`/damage-source dispatch checking attacker type and difficulty before
  actually removing the entity). Suppression point: on vanilla `VillagerEntity`, this must be
  Mixin-cancelled (inject into the conversion check/`onDeath` path) or the fork must be forced
  to always resolve to true death; a custom entity that does not extend `VillagerEntity` (or
  `AbstractVillagerEntity`) simply never enters vanilla's conversion code path at all — the
  behavior doesn't exist unless the mod's own damage handling reimplements it.
- **Curing (the reverse path):** requires Weakness + golden apple, ~3–5 min timer, ends in
  Nausea + profession/trade restoration. Only reachable if conversion (above) is not
  suppressed; curing itself only exists in the zombie-villager code path, so suppressing
  conversion makes curing moot rather than requiring separate suppression.
  https://minecraft.wiki/w/Zombie_Villager (§Curing) — accessed 2026-08-21.
- **Trade restocking (fiction-breaking behavior #2, softer):** working at the job site
  restocks exhausted trades up to twice/day; this is arguably not fiction-breaking by itself
  (it's the trading system continuing to function) but ties into whatever the mind daemon
  decides trading should mean for Kithcraft — flagged as an interaction with the body-protocol
  seam (TASK-0002's lane), not decided here. Suppression point if trading is disabled
  wholesale: same `interactMob`/`beginTradeWith` override noted in §4 below; restocking itself
  is a `Task<VillagerEntity>` run during the Work activity and would need Mixin removal from
  the activity's task list if trading-adjacent behavior must be fully absent rather than just
  UI-inaccessible.
  https://minecraft.wiki/w/Trading (§Restocking) — accessed 2026-08-21.

## 4. Unwanted vanilla behaviors: disable points and Mixin surface

| Behavior | Where implemented (vanilla) | Disable point per option |
|---|---|---|
| Trading UI | `VillagerEntity.interactMob(PlayerEntity, Hand)` opens the trade screen via `beginTradeWith`. Confirmed method signature present across multiple Yarn builds (e.g. yarn-1.21+build.2). https://maven.fabricmc.net/docs/yarn-1.21+build.2/net/minecraft/entity/passive/VillagerEntity.html — accessed 2026-08-21. A shipped Fabric mod ("Disable Villager trades") achieves exactly this by intercepting `interactMob` to stop the trade screen while leaving raids/Hero-of-the-Village untouched, confirming the surface is a single reachable interception point. https://www.curseforge.com/minecraft/mc-mods/disable-villager-trades — accessed 2026-08-21. | **Augmented villager:** plain event API — Fabric's player-block/entity interaction events (`UseEntityCallback` or Mixin-cancel `interactMob`) cancel the trade UI without touching anything else; no Mixin strictly required if an `ActionResult.FAIL`-returning event callback runs first. **Custom entity:** moot — a fresh entity class simply never implements `interactMob` to open a trade screen; nothing to disable. |
| Breeding | Villager breeding is driven by "willingness" (food-derived) and executes when two willing villagers can path to an unclaimed bed; described as brain/schedule-integrated behavior, not a classic `Goal`. https://minecraft.wiki/w/Villager (§Breeding) — accessed 2026-08-21. | **Augmented villager:** must Mixin-cancel/no-op the breeding `Task`'s condition or strip it from the activity task list (setTaskList override) — [[villager-brain-api]] confirms task-list assignment is plain API, but *removing* a vanilla-registered task from a specific Activity requires re-calling `setTaskList` with the filtered list, which typically means Mixin access to brain init since vanilla doesn't expose an official "remove task" call. **Custom entity:** moot if brain init is authored from scratch and breeding tasks are never added. |
| Gossip | `GossipManager`/reputation memories, updated on trade/attack/kill/cure, decaying every 20 min, drives trade pricing and iron-golem hostility toward the player. https://minecraft.wiki/w/Villager (§Reputation) — accessed 2026-08-21. | **Augmented villager:** gossip is memory-module + sensor driven; suppressing "unwanted" gossip effects (e.g. price swings) means Mixin-blocking the gossip-update call sites (on trade/attack/kill) or overriding price calculation — no plain-API single point found. **Custom entity:** moot if `GossipManager` is never attached/initialized (it's part of `VillagerEntity`'s own data, not generic `Brain<E>` state as far as verified here — not independently re-checked against source, flagged as thin). |
| Iron-golem summoning | Villagers summon golems when gossiping (5+ villagers within 10 blocks) or panicking (3+ within 10 blocks), subject to sleep-recency/golem-visibility/cooldown checks — this is villager-initiated behavior, not player-triggered. https://minecraft.wiki/w/Iron_Golem (§Summoning) — accessed 2026-08-21. | **Augmented villager:** this is a `Task` gated by sensors/memories similar to breeding — same Mixin-removal-from-task-list surface. **Custom entity:** moot if the summoning task is never wired in. |

**Net Mixin-surface read:** trading UI is the one behavior with a genuine plain-event
disable point regardless of entity choice (Fabric interaction-event API, no Mixin needed).
Breeding, gossip, and golem-summoning are all brain `Task`-list members; suppressing them on
an **augmented vanilla villager** requires Mixin-level access to override/replace the task
lists vanilla wires into `VillagerEntity`'s brain init (consistent with [[villager-brain-api]]'s
finding that anything beyond *using* the existing brain needs Mixin). On a **custom entity**,
all three are eliminated for free simply by never registering the corresponding tasks in a
from-scratch brain init — the entity-choice axis directly trades "more free village-fiction
machinery" (augmented) against "fewer unwanted behaviors to actively suppress" (custom),
which the comparison doc (Phase 2) needs to weigh explicitly.

## 5. Client-side visibility for a vanilla (unmodded) client

**A vanilla client cannot render, and in current Fabric practice cannot even safely receive,
a custom entity type from a server-only Fabric mod.** Evidence:

- Fabric's own entity-creation tutorial states new entity types are registered under
  `ENTITY_TYPE`/`BuiltInRegistries.ENTITY_TYPE` and require a **client-side** renderer
  registration (`EntityRenderers.register` in a `ClientModInitializer`) to display at all; it
  further notes "If your entity does not extend `LivingEntity` you have to create your own
  spawn packet handler" — i.e. even the network representation of a non-`LivingEntity` custom
  type is not automatic.
  https://wiki.fabricmc.net/tutorial:entity — accessed 2026-08-21.
- Fabric's registry-sync mechanism is explicitly documented (maintainer discussion on the
  Fabric API tracker) as intending to **kick vanilla clients that don't understand the
  registry sync packet** on a "real" (non-dev-environment) modded server — i.e. Fabric's
  design goal is to prevent, not gracefully degrade, connections from clients lacking the
  server's registered content.
  https://github.com/FabricMC/fabric-api/issues/894 — accessed 2026-08-21.
- A related historical bug report documents the failure mode directly: a Fabric server with
  custom/shifted entity registry IDs and a vanilla (or mismatched) client caused **client-side
  crashes** because entity IDs sent over the network didn't match the client's registry.
  https://github.com/FabricMC/fabric-api/issues/135 — accessed 2026-08-21.
- A parallel, better-documented case (Mojang's own network **registry** packets — biome/trim/
  dimension-type/damage-type/chat-type, a related but distinct sync mechanism from entity
  types) confirms the general pattern: a vanilla client fed content referencing registry
  entries it doesn't have either **hard-crashes** or, after a later 1.20.5-era fix,
  **disconnects gracefully** — never "renders as something generic." No equivalent official
  fix was found specifically for custom *entity type* registries (the fix cited is for the six
  named data-driven registries, not `ENTITY_TYPE`).
  https://mojira.dev/MC-267103 — accessed 2026-08-21.

**Net assessment — this is the single most load-bearing finding for the "drop-in
multiplayer" constraint:** a **custom Fabric entity requires every connecting client to run
the same mod (or at minimum the same client-side renderer/entity-registration jar)** — a
friend on a vanilla client either cannot join at all (best case: clean kick via registry-sync
mismatch) or crashes/disconnects on first sight of the entity (worse case, version- and
Fabric-API-behavior-dependent). This directly contradicts an unstated assumption the brief's
"drop-in multiplayer" framing could invite (that a friend joining vanilla-client "just sees
them") for the **custom-entity** option. An **augmented vanilla `VillagerEntity`** carries
**zero client-side requirement**: it is still `minecraft:villager` on the wire, so a fully
vanilla client renders it exactly as any other villager, with no mod, no crash risk, no
registry mismatch — the server-side brain/behavior changes are invisible to (and don't need
cooperation from) the network protocol.

**Skin/appearance flexibility, per option (evidence thinner here — flagged explicitly):**
- Augmented `VillagerEntity`: vanilla villager skins are determined by profession and biome
  type (7 biome variants × profession-driven clothing), not freely assignable per-individual;
  distinguishing named villagers as individuals without a client mod is constrained to
  whatever profession/biome combinations vanilla ships — **no evidence found of a
  server-only way to give a vanilla-rendered villager an arbitrary custom skin**; this would
  need a resource pack (client-side) for true per-individual distinctiveness, which is a
  weaker "no client mod" claim than the entity-existence claim above. Not independently
  re-verified beyond the profession/biome-variant mechanic already documented in §2's sources.
- Custom entity: full rendering/model/texture control exists but strictly requires the
  client-side mod jar per the tutorial above — the "no client mod" framing does not apply to
  this option's rendering at all, only to whether a modded client *could* render it distinctly
  once installed.

## Interactions with the body-protocol seam (flagged, not decided)

- Trade restocking (§3) and trading UI (§4) both touch whatever the mind daemon eventually
  decides "trading" means for Kithcraft villagers (a percept the mind reacts to, a body action
  it can request, or something disabled entirely) — TASK-0002's lane owns that shape; this
  doc only records where the vanilla behavior lives and how each entity option would
  suppress/keep it.
- The `nearest_hostile` memory populated by `VillagerHostilesSensor` (§1) is a cheap,
  already-computed percept either option could expose to the mind daemon as a "danger nearby"
  signal without extra engine work — worth flagging to 0002 as a low-cost percept source,
  not deciding its wire format here.

## Areas where evidence is thin or version-dependent (do not round off)

1. Whether `VillagerHostilesSensor` and `GossipManager` are generic `Brain<E>`/`Sensor`
   machinery reusable by any brain-driven entity, or are hard-wired into `VillagerEntity`
   specifically, was **not independently re-verified against Yarn-mapped source** in this
   pass — only inferred from wiki mechanics descriptions. This matters for exactly how much a
   custom entity can share vs. must reimplement; flagged for the comparison doc to treat as an
   open risk on the custom-entity side rather than a settled "must reimplement."
2. Exact client-crash-vs-clean-disconnect behavior for an unrecognized `ENTITY_TYPE` specifically
   (as opposed to the six documented data-driven registries in MC-267103) was not found as a
   single authoritative current-version source — the two GitHub issues cited are from 2019–2020
   and describe the design intent (kick) and a historical bug (crash), not a current-version
   confirmed outcome. The practical conclusion (a vanilla client cannot use a custom entity
   without the mod) is not in doubt across all sources found; the precise failure UX (kick vs.
   crash) is.
3. Per-individual custom skins on an augmented vanilla `VillagerEntity` without any client-side
   asset changes were not confirmed possible or impossible by direct source — flagged as an
   open question for the comparison doc, not asserted either way.
