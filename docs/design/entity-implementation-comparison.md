# Entity implementation comparison — custom Fabric entity vs augmented vanilla villager

**Spec:** 003-entity-implementation · **Phases:** 2 (comparison) + 3 (recommendation,
proposed — pending operator ratification via `decision-0002`). Per [[mod-stack-decision]],
both options sit on the same vanilla Fabric brain/Mixin substrate ([[villager-brain-api]]):
"custom entity" means a Fabric entity class wired into `Brain<E>`, never a Citizens2-authored
NPC. All citations below are carried from
`specs/003-entity-implementation/research/engine-behavior.md` (Phase 1), checked live
**2026-08-21** unless a source is version-pinned, in which case the mapping build is stated
inline.

## R1.1 — Village fiction reuse (beds, workstations, schedules, POI)

**Augmented vanilla `VillagerEntity`:** inherits the entire mechanic for free — job-site
claim search/pathing/claim-radius, the 3-slot POI claim map (bed/gathering/job site), bed
pathing and sleep/wake rules, and the fixed profession-keyed `Schedule`. All plain-API
`Brain<E>` get/set per [[villager-brain-api]] once inside the existing task lists.
https://minecraft.wiki/w/Village_mechanics — accessed 2026-08-21.
https://minecraft.wiki/w/Villager (§Profession, §Sleeping, §Schedule) — accessed 2026-08-21.

**Custom entity:** inherits none of it. Must independently implement POI claim/pathing
logic and either register/reuse the same `MemoryModuleType`s and `Sensor`s (e.g.
`SecondaryPoiSensor`) via Mixin-exposed accessors, or Mixin-copy vanilla's
`Schedule`/`ScheduleBuilder`. The claim/pathing/sleep `Task<E>` implementations themselves
are only free if the custom entity extends `VillagerEntity`/`AbstractVillagerEntity` (or
re-invokes its Mixin-exposed brain-init methods) — merely wiring identical memory/sensor
*names* on a fresh class does not pull in the vanilla task behavior.
`specs/003-entity-implementation/research/engine-behavior.md` §2, sourced to
https://minecraft.wiki/w/Village_mechanics and https://minecraft.wiki/w/Villager — accessed
2026-08-21.

**Mixin surface owned:** augmented — none beyond what any custom `Activity`/memory
addition already requires per [[villager-brain-api]]. Custom entity — full reimplementation
or Mixin-copy of POI claim, sleep pathing, and `Schedule` wiring.

## R1.2 — Real night danger (hostile targeting, panic, raids)

**Augmented vanilla `VillagerEntity`:** inherits both halves for free. Zombie/husk/
drowned/zombie-villager "Active Target - Merchant" goals target `VillagerEntity`/
`WanderingTraderEntity` by Java class check (35–52.5-block radius, see-through-walls,
target-locks until death/attack/player-proximity), and the brain-based
`VillagerHostilesSensor` independently populates the villager's own `nearest_hostile`
memory (fixed hostile-type list, 8–15 block ranges) driving panic/flee. Raid behavior
(flee to nearest house with door+bed, bell-ring warning, Glowing on illagers) is likewise
inherited. https://minecraft.wiki/w/Mob_AI — accessed 2026-08-21.
https://minecraft.wiki/w/Zombie — accessed 2026-08-21.
https://minecraft.wiki/w/Villager — accessed 2026-08-21.

**Custom entity:** invisible to attacker goals by default, because the goal-based targeting
is a hardcoded class check, not a registry/tag lookup — zombies etc. will not target a
custom entity unless the mod Mixins its own `NearestAttackableTargetGoal` into hostile mob
goal selectors, or the entity extends `AbstractVillagerEntity`/`VillagerEntity`. Self-panic
sensing (`VillagerHostilesSensor`) would also need to be wired independently or reused —
**flagged as thin evidence**: whether that sensor is generic `Brain<E>` machinery or
hard-wired to `VillagerEntity` internals was not independently verified against source in
Phase 1 (open risk on the custom-entity side, not a settled "must reimplement").
`specs/003-entity-implementation/research/engine-behavior.md` §1, sourced to
https://minecraft.wiki/w/Mob_AI — accessed 2026-08-21.

**Mixin surface owned:** augmented — none; this is the constraint area where augmentation
carries essentially zero owned surface. Custom entity — Mixin into every relevant hostile's
goal selector (or extend a vanilla class family, which reopens the fiction-reuse tradeoff
above), plus an unresolved-cost hostile-sensing wire-up.

## R1.3 — Permadeath (death handling, despawn, what must be suppressed)

**Despawn:** villagers are passive mobs; passive mobs do not despawn from distance/time in
Java Edition (the 1/800-per-tick random despawn applies to monster/ambient/aquatic
categories, not passive). https://minecraft.wiki/w/Mob_spawning — accessed 2026-08-21.
https://minecraft.wiki/w/Tutorial:Mob_switch — accessed 2026-08-21. This holds for
**either option** as long as the custom entity is not registered under a different
`MobCategory`; the Fabric entity-creation tutorial shows `MobCategory` is an explicit
registration choice, so this is a configuration risk for the custom-entity option, not an
inherent one. https://wiki.fabricmc.net/tutorial:entity — accessed 2026-08-21. Both
options share the Y ≤ -128 instant-despawn edge case.

**Zombie-villager conversion (the fiction-breaking behavior):** on `VillagerEntity`, a
killing blow from a zombie/husk/drowned/zombie villager converts rather than kills, at
0/50/100% by difficulty — this is a fork inside vanilla's damage-handling path specific to
`VillagerEntity`. https://minecraft.wiki/w/Villager (§Attacking) — accessed 2026-08-21.
**Augmented vanilla** must Mixin-cancel this conversion check (or force it to always
resolve to true death) to make death final. **Custom entity** never enters this code path
at all if it doesn't extend `VillagerEntity`/`AbstractVillagerEntity` — the behavior simply
doesn't exist, no suppression needed. Curing (Weakness + golden apple, ~3–5 min,
Nausea + trade restoration) is downstream of conversion and becomes moot wherever
conversion is absent/suppressed. https://minecraft.wiki/w/Zombie_Villager (§Curing) —
accessed 2026-08-21.

**Mixin surface owned:** augmented — one targeted Mixin injection into the
conversion/`onDeath` dispatch path. Custom entity — none (moot by construction), traded
against the targeting cost in R1.2 above (a custom entity that never enters
`VillagerEntity`'s conversion path also never enters its being-targeted path without
separate Mixin work).

## R1.4 — Drop-in multiplayer (what a vanilla client sees)

This is **the single most load-bearing and most one-sided finding** in the evidence base.

**Augmented vanilla `VillagerEntity`:** zero client-side requirement. The entity remains
`minecraft:villager` on the wire; a fully vanilla client renders it exactly as any other
villager, no mod, no crash risk, no registry mismatch.

**Custom entity:** requires every connecting client to run the same mod (or at minimum the
matching client-side renderer/entity-registration jar). Evidence chain:
- New entity types need client-side `EntityRenderers.register` in a
  `ClientModInitializer`; a non-`LivingEntity` type additionally needs a custom spawn
  packet handler — the network representation is not automatic.
  https://wiki.fabricmc.net/tutorial:entity — accessed 2026-08-21.
- Fabric's registry-sync mechanism is designed to **kick** vanilla clients that don't
  understand the sync packet on a real (non-dev-environment) server.
  https://github.com/FabricMC/fabric-api/issues/894 — accessed 2026-08-21.
- A historical bug report documents the failure mode directly: mismatched entity registry
  IDs between a Fabric server and a vanilla/mismatched client caused **client-side
  crashes**. https://github.com/FabricMC/fabric-api/issues/135 — accessed 2026-08-21.
- A parallel, better-documented Mojang registry-sync case (biome/trim/dimension-type/
  damage-type/chat-type) confirms the general pattern — unrecognized registry content
  either hard-crashes the client or (post-1.20.5-era fix, for those six named registries
  only) disconnects cleanly; no equivalent official fix was found for custom
  `ENTITY_TYPE` specifically. https://mojira.dev/MC-267103 — accessed 2026-08-21.

**Evidence gap, stated rather than rounded off:** the precise failure UX for a custom
entity type specifically (clean kick vs. crash) is not confirmed by a single current-version
authoritative source — the two GitHub issues are 2019–2020 vintage and establish design
intent (kick) and a historical bug (crash), not a current-version confirmed outcome. What
is not in doubt across every source found: a vanilla client cannot use a custom entity
without the mod, in some failure mode. A friend joining vanilla-client, as the brief's
"drop-in multiplayer" language could be read to promise, is **only satisfied by the
augmented-vanilla option**.

**Mixin surface owned:** augmented — none (this constraint is free). Custom entity —
none at the Mixin layer, but a hard mod-distribution requirement outside the Mixin surface
entirely (client jar must ship to every player who joins).

## R1.5 — Behavior control (trading, gossip, breeding, restocking)

| Behavior | Vanilla implementation | Augmented villager disable point | Custom entity disable point |
|---|---|---|---|
| Trading UI | `VillagerEntity.interactMob` → `beginTradeWith`. Confirmed across Yarn builds (yarn-1.21+build.2). https://maven.fabricmc.net/docs/yarn-1.21+build.2/net/minecraft/entity/passive/VillagerEntity.html — accessed 2026-08-21. A shipped mod achieves this via plain interaction-event interception. https://www.curseforge.com/minecraft/mc-mods/disable-villager-trades — accessed 2026-08-21. | Plain event API — Fabric interaction-event callback (`ActionResult.FAIL`), no Mixin required. | Moot — never implemented. |
| Breeding | Willingness-driven, brain/schedule-integrated `Task`, not a classic `Goal`. https://minecraft.wiki/w/Villager (§Breeding) — accessed 2026-08-21. | Mixin access to brain init required to remove the task from its Activity's task list — no official "remove task" call exists per [[villager-brain-api]]. | Moot if brain init is authored from scratch and the task is never added. |
| Gossip | `GossipManager`/reputation memories updated on trade/attack/kill/cure, 20-min decay, drives price and iron-golem hostility. https://minecraft.wiki/w/Villager (§Reputation) — accessed 2026-08-21. | No plain-API single point found; Mixin-blocking gossip-update call sites or overriding price calc required. | Moot if `GossipManager` is never attached — **flagged as thin evidence**: whether it's generic `Brain<E>` state or hard-wired to `VillagerEntity` was not independently re-checked against source. |
| Iron-golem summoning | Villager-initiated `Task`, gated by gossip/panic sensor thresholds (5+ gossiping or 3+ panicking within 10 blocks) plus cooldown/visibility checks. https://minecraft.wiki/w/Iron_Golem (§Summoning) — accessed 2026-08-21. | Same Mixin task-list-removal surface as breeding. | Moot if the task is never wired in. |
| Trade restocking | `Task<VillagerEntity>` run during Work activity, up to twice/day. https://minecraft.wiki/w/Trading (§Restocking) — accessed 2026-08-21. | Mixin removal from the Work activity's task list if full suppression (not just UI-block) is required. | Moot if never wired in. **Flagged as body-protocol-seam interaction** — see below. |

**Net read:** trading UI is the one behavior with a genuine Mixin-free disable point on
*either* option. Breeding, gossip, and golem-summoning are brain `Task`-list members:
suppressing them on **augmented vanilla** costs Mixin-level task-list-override access for
each one; on **custom entity** all three are eliminated for free by never registering the
corresponding tasks in a from-scratch brain init. This is a direct tradeoff, not a
one-sided result: augmentation buys more free village-fiction machinery (R1.1) at the cost
of more actively-suppressed unwanted behavior here; a custom entity buys freedom from
unwanted behavior at the cost of having to rebuild the wanted machinery in R1.1.

**Mixin surface owned:** augmented — up to 3 task-list-override Mixins (breeding, gossip,
golem-summoning) plus the conversion-suppression Mixin from R1.3, none needed if a design
choice tolerates any of those behaviors. Custom entity — zero for this area (all moot by
omission), but this credit is exactly offset by the reimplementation cost already counted
in R1.1.

## R1.6 — Rendering/skin flexibility (distinguishable individuals)

**Augmented vanilla `VillagerEntity`:** skins are determined by profession × 7 biome
variants, not freely assignable per-individual. **No evidence found of a server-only way to
give a vanilla-rendered villager an arbitrary custom skin** — true per-individual
distinctiveness would need a resource pack, which is client-side and therefore a weaker "no
client mod" claim than the entity-existence claim in R1.4. **Flagged as an open question,
not asserted either way** — not independently confirmed possible or impossible by direct
source. `specs/003-entity-implementation/research/engine-behavior.md` §5, sourced to the
profession/biome-variant mechanic in https://minecraft.wiki/w/Villager (§Profession) —
accessed 2026-08-21.

**Custom entity:** full rendering/model/texture control exists, but strictly requires the
client-side mod jar (same requirement established in R1.4) — the "no client mod" framing
does not apply to this option's rendering at all; it only describes what a modded client
*could* render once installed. https://wiki.fabricmc.net/tutorial:entity — accessed
2026-08-21.

**Mixin surface owned:** augmented — none identified (the constraint is bounded by biome/
profession variants, not a Mixin problem). Custom entity — none at the Mixin layer either;
the cost here is entirely the same client-distribution requirement as R1.4, not additional
engineering surface.

## Per-option total owned Mixin surface

**Augmented vanilla `VillagerEntity`:**
- R1.2 (targeting/panic): none.
- R1.3 (permadeath): 1 Mixin injection (conversion/`onDeath` dispatch cancel).
- R1.5 (behavior control): up to 3 Mixin task-list overrides (breeding, gossip,
  golem-summoning), scoped to whichever of those the design actually wants suppressed.
- R1.1, R1.4, R1.6: none beyond the standard "new Activity/MemoryModuleType/POI type"
  Mixin surface already accepted as a stack-decision risk in [[mod-stack-decision]].
- **Total: small and enumerable** — at most ~4 targeted Mixin injection points beyond the
  substrate-level surface every brain extension already requires.

**Custom Fabric entity:**
- R1.1 (village fiction): full POI/pathing/sleep reimplementation or Mixin-copy, plus
  `Schedule`/`ScheduleBuilder` wiring from scratch.
- R1.2 (targeting/panic): Mixin into every relevant hostile's goal selector (or a class
  hierarchy choice that reopens R1.1's tradeoff), plus unresolved-cost hostile-sensing.
- R1.3 (permadeath): none (moot).
- R1.4/R1.6: none at the Mixin layer, but a standing client-jar distribution requirement.
- R1.5 (behavior control): none (moot by omission).
- **Total: larger and less enumerable** — the village-fiction and targeting reimplementation
  costs are open-ended reconstructions of vanilla machinery rather than discrete injection
  points, and two sub-findings (hostile-sensing reuse, `GossipManager` coupling) are
  explicitly flagged as unverified, so the true size of this surface carries more
  uncertainty than the augmented option's.

## Interactions with the body-protocol seam (TASK-0002, flagged only — not decided)

- Trade restocking (R1.5) and the trading UI both touch whatever the mind daemon decides
  "trading" means for Kithcraft villagers (a percept, a body action, or something disabled
  outright) — TASK-0002's lane owns that shape; this document only records where the
  vanilla behavior lives and how each entity option would suppress or keep it.
- The `nearest_hostile` memory populated by `VillagerHostilesSensor` (R1.2) is a cheap,
  already-computed percept either option could expose to the mind daemon as a "danger
  nearby" signal without extra engine work on the augmented-vanilla path; on the
  custom-entity path this signal's cost depends on the unresolved question of whether the
  sensor is reusable as-is (flagged in R1.2) — worth surfacing to 0002 as a low-cost
  percept source on at least one option, not deciding its wire format here.

## Areas where evidence is thin or version-dependent (carried from Phase 1, not rounded off)

1. Whether `VillagerHostilesSensor` and `GossipManager` are generic `Brain<E>`/`Sensor`
   machinery reusable by any brain-driven entity, or hard-wired to `VillagerEntity`
   specifically, was not independently re-verified against Yarn-mapped source — only
   inferred from wiki mechanics descriptions. Treated here as an open risk on the
   custom-entity side (R1.2, R1.5), not a settled "must reimplement."
2. The exact client failure UX (kick vs. crash) for an unrecognized custom `ENTITY_TYPE`
   specifically (R1.4) was not found in a single authoritative current-version source; the
   underlying conclusion (a vanilla client cannot use a custom entity without the mod) is
   not in doubt across all sources found.
3. Per-individual custom skins on an augmented vanilla `VillagerEntity` without any
   client-side asset changes (R1.6) were not confirmed possible or impossible by direct
   source — open question, not asserted either way.

## Recommendation (Phase 3)

**Recommend: augmented vanilla `VillagerEntity`.**

### Rationale, mapped to the ratified constraints

- **Village fiction reuse:** augmented inherits the entire mechanic — POI claim, sleep
  pathing, `Schedule` — for free (R1.1); custom entity must reimplement or Mixin-copy all
  of it, and only the memory/sensor *names* are shareable without also either extending
  `VillagerEntity` or re-invoking its Mixin-exposed brain-init methods. This is the
  constraint the brief states most literally ("villager-shaped... riding the village
  fiction," [[design-brief]] #3) and augmented satisfies it by construction, not by effort.
- **Real night danger:** the load-bearing finding is that hostile-mob targeting
  (`NearestAttackableTargetGoal` equivalents in zombie/husk/drowned goal selectors) is a
  **hardcoded Java class check against `VillagerEntity`/`AbstractVillagerEntity`, not a
  registry or tag lookup** (R1.2, `engine-behavior.md` §1). A custom entity that doesn't
  extend that hierarchy is invisible to attacker AI by default — the single mechanic that
  makes walls-protect-friends real would have to be rebuilt with Mixins into every hostile's
  goal selector, and would still need its own hostile-sensing wire-up (flagged thin).
  Augmented gets both halves — being targeted and self-panic-sensing — free.
- **Permadeath:** augmented owns exactly one targeted Mixin (cancel the zombie-villager
  conversion fork in `onDeath`/damage dispatch, R1.3); custom entity never enters that code
  path so needs none. This is the one constraint area where custom entity is cheaper, but
  the cost differential (0 vs. 1 well-understood injection point) is the smallest in the
  whole comparison.
- **Drop-in multiplayer:** the single most one-sided finding in the evidence base (R1.4).
  Augmented is `minecraft:villager` on the wire — zero client-side requirement, a vanilla
  client renders it exactly as any other villager. A custom entity requires every
  connecting client to run the matching mod jar or renderer registration; Fabric's own
  registry-sync design intent is to **kick** clients that don't recognize the entity type,
  and a historical bug report documents an outright **client-side crash** from mismatched
  entity registry IDs. Read literally, "a friend can drop in" ([[design-brief]] #9,
  [[v1-demo]]) is satisfied by exactly one of the two options.
- **Behavior control:** a genuine trade-off, not one-sided (R1.5) — augmented must
  actively suppress breeding/gossip/golem-summoning (each a Mixin task-list override) where
  custom entity gets them moot by never wiring the tasks in. But the suppression surface is
  small and enumerable: **augmented's total owned Mixin surface across every constraint area
  is at most ~4 targeted injection points** (1 conversion-cancel + up to 3 task-list
  overrides), all beyond a substrate-level surface already accepted as a stack-decision risk
  in [[mod-stack-decision]]. Custom entity's corresponding surface — full POI/pathing/sleep
  reconstruction, a Mixin into every hostile's goal selector, hostile-sensing rewiring — is
  larger and, per the two flagged-thin findings (`VillagerHostilesSensor` and
  `GossipManager` reusability), not even fully known in size.
- **Rendering/skin flexibility:** the one constraint where evidence for augmented is thin
  rather than clean — no confirmed server-only way to give a vanilla-rendered villager an
  arbitrary custom skin (R1.6); distinctiveness is bounded by 7 biome variants × profession.
  Custom entity has full rendering control in principle, but only for clients running the
  mod — for the "no client mod" framing this constraint is actually decided by R1.4, not by
  R1.6 itself, since a custom entity's rendering flexibility is inaccessible to exactly the
  vanilla clients the multiplayer constraint cares about.

### What this gives up

Recommending augmented vanilla is not a clean sweep — two things stay named rather than
rounded off:

1. **No confirmed server-only path to per-individual custom skins** (R1.6). If the cast
   needs to read as visually distinct individuals beyond profession/biome variation, the
   only lever found is a resource pack, which is client-side and re-opens exactly the "no
   client mod" trade this recommendation is otherwise winning on.
2. **`VillagerHostilesSensor` and `GossipManager` genericity is unverified** (flagged thin
   in both R1.2 and R1.5) — this doesn't change the recommendation (it only makes the
   custom-entity alternative's true cost *larger or equal*, never smaller, since these are
   findings about what custom entity would additionally need to reimplement if the sensors
   turn out to be `VillagerEntity`-internal), but it means the size of the road not taken is
   understated here, not overstated.

**Mitigation posture:** neither give-up blocks the demo (TASK-0006, 3–6 named villagers,
one evening). (1) is accepted as-is for v1 — the brief's cast-distinctiveness need can be
met with profession/biome-variant assignment plus names/dialogue rather than skins; if
true per-individual skins become load-bearing later, the next step is a scoped spike into
resource-pack-based skin assignment (client-side, separately scoped, not blocking server
logic) rather than a bare re-litigation of this decision. (2) is accepted because it only
matters if a future task revisits the custom-entity path; it costs nothing on the augmented
path chosen here, so no mitigation action is needed unless that reversal is proposed.

### Narrowing effects on TASK-0006 (demo build plan)

- **Entity work is `VillagerEntity` augmentation, not a new entity class** — TASK-0006 can
  plan Mixin/accessor work directly against the existing brain substrate
  ([[villager-brain-api]]) rather than budgeting for ground-up brain/POI/schedule
  reconstruction.
- **The Mixin surface for the demo's 3–6 villagers is now committed and small:** at most one
  conversion-cancel injection (permadeath, R1.3) plus up to three task-list-override
  injections (breeding/gossip/golem-summoning, R1.5) — the demo build plan can size this as
  a bounded, enumerable slice of work, not an open-ended one.
- **No client-side mod work is needed for the demo at all** — the "friend can drop in"
  texture ([[v1-demo]]) is satisfied by the entity choice alone; TASK-0006 does not need to
  plan, build, or distribute a client jar.
- **The cast's appearance without a client mod is bounded to profession × 7 biome variants**
  — TASK-0006 should plan named/distinguishable villagers through profession assignment,
  biome-variant selection, and in-fiction means (names, dialogue, job-board role) rather than
  assuming free custom skins; a resource-pack skin path stays a later, separately-scoped
  option if needed, not a demo dependency.
- **The `nearest_hostile` memory (R1.2) is confirmed available as a free, already-computed
  percept** on the augmented path — TASK-0006 (and TASK-0002's protocol shape) can treat
  "danger nearby" as a cheap signal to expose to the mind daemon without additional engine
  work, with no open sensor-reuse question to resolve first (that question only existed on
  the custom-entity path, which is not the chosen one).
