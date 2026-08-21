---
name: mod-stack-decision
description: decision-0001 (ACCEPTED 2026-08-20) — Fabric server-side mod over Paper/Citizens2 and hybrid. The rationale (license health, vanilla-brain directness, village-fiction fit), the accepted risks (no off-the-shelf LLM-NPC framework; owned Mixin surface), and what it narrows for TASK-0003. Load before any stack-dependent work.
kind: concept
sources:
  - backlog/decisions/decision-0001 - Mod-stack-Fabric-server-side-mod-proposed.md
  - docs/design/mod-stack-comparison.md
verified_against: 50c3def435dd9326d38e51118f08944815cbe80c
---

# Mod-stack decision (decision-0001, accepted)

Kithcraft builds as a **Fabric server-side mod**, driving vanilla `Villager` entities (or
a custom entity on the same brain substrate) through the engine's brain API. Ratified by
the operator 2026-08-20 (PR #1 review, merge 71baeb8); recorded as decision-0001 with
full evidence in `docs/design/mod-stack-comparison.md`. In force for TASK-0002/0003/0006.

## How it works

The rationale, compressed from the comparison's §Recommendation:

- **Body-protocol seam:** Fabric and Paper are equally thin seams; dependency health
  discriminated. Citizens2 (Option B's only viable NPC substrate) is OSL-3.0 — copyleft,
  not GPL-compatible — risking license entanglement for a tightly-coupled body vendor.
  Fabric Loader/API are Apache-2.0 and the villager brain is vanilla engine API, no
  third-party dependency at all.
- **Villager-shaped, not bot clients:** strongest fit of the three — real server entities
  on the actual vanilla villager brain, no intermediate NPC-behavior abstraction.
  Citizens2 layers its own goals/behavior-tree system over an impersonated entity.
- **Village fiction — the deciding factor:** vanilla `Schedule`/`Activity`/
  `MemoryModuleType`/POI machinery is plain API access on the real brain
  ([[villager-brain-api]]); Citizens2 would mean bypassing its behavior system or
  reimplementing schedules.
- **Hybrid rejected:** no prior art demonstrates the combination; it accumulates both
  options' risks rather than averaging them, for a 3–6 NPC demo that needs neither.

**Accepted risks:** no off-the-shelf LLM-NPC framework (CraftAgent is a documented dead
end; SecondBrain is design reference only) — the integration layer is written from
scratch; extending the brain substrate requires Mixin/accessor code the project owns.

**What it narrows for TASK-0003:** custom-entity vs augmented-vanilla-villager stays
fully open, but both sub-options now sit on the vanilla Fabric brain/Mixin substrate — a
"custom entity" means a Fabric entity class wired into `Brain<E>`, never a
Citizens2-authored NPC.

## Connections

Evidence base: [[prior-art]]; substrate detail: [[villager-brain-api]]; constraints it
was judged against: [[design-brief]], [[body-protocol-seam]]; produced by the process in
[[pdlc-process]] (spec 001, TASK-0001, PR #1).

## Operational notes

The decision record's status lives in its frontmatter (`status: accepted`). The backlog
CLI has no decision-edit verb — the proposed→accepted flip was a sanctioned direct edit
recorded in the file itself.
