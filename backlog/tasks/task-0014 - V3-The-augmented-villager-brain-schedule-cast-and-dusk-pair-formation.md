---
id: TASK-0014
title: 'V3 - The augmented villager: brain, schedule, cast, and dusk pair formation'
status: In Progress
assignee: []
created_date: '2026-08-21 23:38'
updated_date: '2026-08-27 20:40'
labels:
  - vendor
  - m-0-build
milestone: m-0
dependencies:
  - TASK-0009
documentation:
  - docs/design/demo-build-plan.md
  - docs/design/entity-implementation-comparison.md
  - docs/design/kithcraft-brief.md
priority: high
ordinal: 14000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
<!-- SECTION:DESCRIPTION:BEGIN -->
As a player, I want three named neighbours who live a full day without me, so that the village is somewhere I arrived rather than something I have to run.

**Scope boundary.** `Schedule` get/set for wake/work/socialize/sleep on the vanilla `Brain<E>` substrate; `Activity` registration and task-list assignment; memory modules; POI bed and job-site claim, sleep pathing, door use — all inherited free per decision-0002. The bounded Mixin surface, **enumerated and no larger**: up to three task-list overrides suppressing breeding, gossip and iron-golem summoning (the conversion-cancel injection belongs to V5). Cast setup: three named villagers distinguished by profession x biome variant plus nameplates. The dusk **pair-formation Activity** implementing ruling **R-7**: villagers path to a shared gathering place at dusk, and the pairing signal is emitted **~10 s ahead of arrival** so M6 can pre-generate the opening turn. And the rule that makes the whole latency posture survivable: **the scheduled activity keeps the body busy while the mind thinks** — a villager standing motionless awaiting a response has converted a 20-second thought into a 20-second bug, and no tier change fixes it.

**Done proves.** On a dev server, three named villagers run a **full day/night cycle unattended**: wake, work, socialize at dusk, and sleep in their claimed beds. No breeding occurs, no gossip-driven golem is summoned, and no player action is required at any point. Two villagers converge on the gathering place at dusk and the pairing signal precedes arrival by ~10 s. With a deliberately stalled mind, bodies keep moving.

**Depends on.** V1.

**Design check — micromanagement.** "A full cycle unattended" is not a nice-to-have in the scope boundary; it *is* the spell-breaker check, made testable. The moment keeping a villager fed, escorted, or on-task requires the player, the demo has shipped the failure mode the brief names.

**References.** docs/design/demo-build-plan.md section 3.3 (V3) and its ruling R-7 are the plan of record. Ratified surfaces consumed: decision-0002 + docs/design/entity-implementation-comparison.md (augmented vanilla VillagerEntity, the bounded ~4-injection Mixin budget, schedules/POI/pathing inherited free, profession x biome variant), docs/design/kithcraft-brief.md (the micromanagement spell-breaker), docs/design/llm-routing-and-budget.md (section 5.2 lever 2, pre-generation ahead of arrival).

**Suggested tier: `sonnet` (next sweep's runbook decides).**
<!-- SECTION:DESCRIPTION:END -->

Spec: specs/014-augmented-villager
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Three named villagers run a full day/night cycle unattended on a dev server: wake, work, socialize at dusk, and sleep in their claimed beds, with no player action required at any point
- [ ] #2 No breeding occurs and no gossip-driven iron golem is summoned
- [ ] #3 The Mixin surface is enumerated and no larger than three task-list overrides (conversion-cancel belongs to V5), staying inside decision-0002's committed bound
- [ ] #4 Two villagers converge on a shared gathering place at dusk and the pair-formation signal precedes arrival by ~10 s (ruling R-7)
- [ ] #5 With a deliberately stalled mind, bodies keep moving: the scheduled activity keeps the body busy while the mind thinks
- [ ] #6 The three villagers are distinguished by profession x biome variant plus nameplates
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Upstream finding from TASK-0009 Phase 1 (2026-08-25, specs/009-fabric-mod-skeleton/research/versions.md): target is MC 26.2, which is UNOBFUSCATED — Yarn discontinued after 1.21.11. villager-brain-api.md's symbol names (checked at yarn-1.21.3+build.1) must be re-verified against Mojang official names before this task's brain/Mixin work; routing A-2's daylight arithmetic also still unverified at 26.2.

Sweep lane 3 claim (2026-08-27). Tier: sonnet · cc/claude-sonnet-5[1m] (default tier per runbook: vanilla Brain<E> substrate work mapped in villager-brain-api; Mixin surface enumerated and bounded by decision-0002; the 26.2 symbol re-derivation is verification, not design). Served model recorded at dispatch.
<!-- SECTION:NOTES:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Tests pass
- [ ] #2 Docs and wiki are updated and pass freshness tests
- [ ] #3 Spec and Backlog are in sync
<!-- DOD:END -->
