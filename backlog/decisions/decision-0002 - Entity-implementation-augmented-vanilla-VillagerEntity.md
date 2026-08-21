---
id: decision-0002
title: 'Entity implementation: augmented vanilla VillagerEntity'
date: '2026-08-21 20:05'
status: proposed
---
## Context

Spec 003 (TASK-0003) evaluated two entity-implementation shapes for Kithcraft villagers,
both sitting on the Fabric substrate fixed by decision-0001: augmenting vanilla
`VillagerEntity` (activity/schedule/memory injection on the stock entity) vs. a custom
Fabric entity wired into the same `Brain<E>` machinery. Full evidence:
`specs/003-entity-implementation/research/engine-behavior.md` (checked live 2026-08-21) and
`docs/design/entity-implementation-comparison.md` (comparison + recommendation).

**Status: PROPOSED — pending operator ratification at the PR checkpoint.** Not yet in
force.

## Decision

Recommend **augmented vanilla `VillagerEntity`** over a custom Fabric entity.

Rationale (full detail in `docs/design/entity-implementation-comparison.md` §Recommendation):

- **Village fiction reuse:** augmented inherits POI claim, sleep pathing, and `Schedule`
  for free; a custom entity gets none of it and must reimplement or Mixin-copy the lot.
- **Real night danger — the load-bearing finding:** hostile-mob targeting
  (`NearestAttackableTargetGoal` equivalents in zombie/husk/drowned goal selectors) is a
  **hardcoded Java class check against `VillagerEntity`/`AbstractVillagerEntity`, not a
  registry or tag lookup**. A custom entity that doesn't extend that hierarchy is invisible
  to attacker AI by default — the mechanic that makes walls-protect-friends real would have
  to be rebuilt with Mixins into every hostile's goal selector. Augmented inherits both
  being-targeted and self-panic-sensing for free.
- **Drop-in multiplayer — the single most one-sided finding:** augmented is
  `minecraft:villager` on the wire, zero client-side requirement. A custom entity requires
  every connecting client to run the matching mod jar; Fabric's registry-sync design intent
  is to **kick** clients that don't recognize the entity type, and a historical bug report
  documents an outright **client-side crash** from mismatched entity registry IDs. "A friend
  can drop in" is satisfied by exactly one of the two options.
- **Permadeath:** augmented owns one targeted Mixin (cancel the zombie-villager conversion
  fork in the death-dispatch path); custom entity never enters that code path so needs none
  — the one area where custom entity is cheaper, and the smallest cost differential in the
  comparison (0 vs. 1 well-understood injection).
- **Behavior control — genuine trade-off, not one-sided:** augmented must actively suppress
  breeding/gossip/golem-summoning (Mixin task-list overrides); custom entity gets them moot
  by omission. But augmented's total owned Mixin surface across every constraint area is at
  most ~4 enumerable injection points; custom entity's corresponding surface (full
  POI/pathing/sleep reconstruction, a Mixin into every hostile's goal selector, independent
  hostile-sensing) is larger and, per two flagged-thin findings, not even fully known in
  size.
- **Rendering/skin flexibility — the thin point for this recommendation:** no confirmed
  server-only path exists to give a vanilla-rendered villager an arbitrary custom skin;
  distinctiveness is bounded to 7 biome variants × profession. Custom entity has full
  rendering control in principle, but only for clients running the mod — inaccessible to
  exactly the vanilla clients the multiplayer constraint cares about.

**What this gives up (named, not rounded off):**
1. No confirmed server-only path to per-individual custom skins — the only lever found
   (resource pack) is client-side and re-opens the "no client mod" trade this decision is
   otherwise winning on.
2. `VillagerHostilesSensor` and `GossipManager` genericity (reusable `Brain<E>` machinery
   vs. hard-wired to `VillagerEntity`) was not independently re-verified against source —
   flagged thin; this only makes the custom-entity alternative's true cost larger or equal,
   never smaller, so it does not change the recommendation.

**Mitigation posture:** neither give-up blocks the v1 demo (3–6 named villagers, one
evening). (1) is accepted as-is for v1 — cast distinctiveness is met via
profession/biome-variant assignment plus names/dialogue; a resource-pack skin spike is a
later, separately-scoped option if per-individual skins become load-bearing. (2) costs
nothing on the chosen (augmented) path and only matters if a future task proposes
reopening the custom-entity path.

## Consequences

- TASK-0006 (demo build plan) is narrowed: entity work is `VillagerEntity` augmentation
  against the existing brain substrate ([[villager-brain-api]]), not ground-up
  brain/POI/schedule reconstruction. The Mixin surface for the demo's cast is committed and
  bounded: at most one conversion-cancel injection (permadeath) plus up to three
  task-list-override injections (breeding/gossip/golem-summoning) — a sizeable, enumerable
  slice of work, not open-ended.
- No client-side mod work is needed for the demo — "a friend can drop in" is satisfied by
  the entity choice alone; TASK-0006 does not plan, build, or distribute a client jar.
- The cast's appearance without a client mod is bounded to profession × 7 biome variants;
  TASK-0006 plans distinguishable villagers through profession/biome assignment and
  in-fiction means (names, dialogue, job-board role), not free custom skins.
- The `nearest_hostile` memory is confirmed available as a free, already-computed percept
  on the chosen path — TASK-0002/TASK-0006 can treat "danger nearby" as a cheap signal to
  the mind daemon with no open sensor-reuse question to resolve first.
- Ratification: pending. This decision is proposed by TASK-0003 and becomes settled fact
  for TASK-0002/0006 only once the operator ratifies at the PR checkpoint (per
  decision-0001's precedent, status flips by direct edit — the backlog CLI's decision
  command supports only `create`, no edit verb).
