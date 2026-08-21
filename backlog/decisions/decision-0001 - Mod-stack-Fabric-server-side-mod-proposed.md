---
id: decision-0001
title: 'Mod stack: Fabric server-side mod'
date: '2026-08-20 14:48'
status: accepted
---
## Context

Spec 001 (TASK-0001) evaluated three mod-stack options for Kithcraft against the ratified
brief (`docs/design/kithcraft-brief.md`): Fabric server-side mod, Paper plugin + Citizens2,
and a hybrid. Full evidence: `specs/001-mod-stack-decision/research/prior-art.md` (re-verified
2026-08-20) and `docs/design/mod-stack-comparison.md` (comparison + recommendation).

**Status: ACCEPTED — ratified by the operator 2026-08-20** (PR #1 review; merge 71baeb8). In force.

## Decision

Recommend **Option A — Fabric server-side mod**, driving vanilla `Villager` entities (or a
custom entity on the same brain substrate, per TASK-0003) directly through the engine's
brain/schedule/POI/memory API.

Rationale (full detail in `docs/design/mod-stack-comparison.md` §Recommendation):
- **Body-protocol seam:** a Fabric mod is a thin server-side surface (tick loop + Mixin
  hooks) with no requirement that mind logic live inside it — the seam stays clean.
  Option B carries the same seam quality but adds a load-bearing license risk (below).
- **Villager-shaped, not bot clients:** Fabric mods manipulate real server entities
  directly — no fake-player client anywhere in the loop, the strongest fit of the three
  options.
- **Village fiction (beds, workstations, schedules):** Option A runs on the actual vanilla
  villager brain, where reading/using `Schedule`/`Activity`/`MemoryModuleType`/POI machinery
  is plain API access (prior-art.md §6). Option B's Citizens2 substrate is a separate
  goals/behavior-tree system layered over an impersonated entity, not the vanilla brain —
  riding the village fiction there means fighting Citizens2's own abstraction.
- **Dependency health decides against Option B:** Citizens2 is OSL-3.0, a copyleft license
  not on the FSF/OSI GPL-compatible list — tight integration would pull body-vendor code
  under copyleft terms. Citizens2 also lags current MC builds (last free-repo commit
  2026-02-16, user-reported breakage on 1.21.10/1.21.11).
- **Option C (hybrid) rejected:** no prior art demonstrates the combination working
  end-to-end; it accumulates both options' risks (Citizens2's license flag, Fabric's Mixin
  burden) rather than averaging them, for a v1 "one real evening" demo (3–6 NPCs) that
  doesn't need either option's dependency set stretched.

## Consequences

- TASK-0002 (body protocol) and TASK-0006 (demo build plan) proceed against a Fabric base.
- TASK-0003 (entity implementation: custom entity vs augmented vanilla villager) is
  narrowed to choosing between two sub-options **within** Fabric — both remain open; this
  decision forecloses neither. It rules out any Citizens2-substrate framing for that choice.
- Accepted risk: no off-the-shelf LLM-NPC framework to build on. CraftAgent (nearest Fabric
  prior art) is a documented dead end (stalled ~7.5 months, pinned to MC 1.20.1/1.21.8); its
  more active sibling SecondBrain is design reference only, not a dependency. The LLM-NPC
  integration layer is written from scratch against the raw Fabric brain API.
- Accepted risk: extending the brain substrate (new activities, memory types, POI types)
  requires Mixin/accessor code the project owns directly — standard Fabric practice, not a
  blocker, but real engineering surface area.
- Ratification: ratified by the operator 2026-08-20 at PR #1 review (merge 71baeb8). The
  decision is settled fact for TASK-0002/0003/0006. (Status flipped by direct edit: the
  backlog CLI's decision command supports only `create`, no edit verb.)
