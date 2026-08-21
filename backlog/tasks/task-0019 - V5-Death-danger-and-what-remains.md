---
id: TASK-0019
title: 'V5 - Death, danger, and what remains'
status: To Do
assignee: []
created_date: '2026-08-21 23:39'
labels:
  - vendor
  - m-0-build
milestone: m-0
dependencies:
  - TASK-0012
  - TASK-0014
documentation:
  - docs/design/demo-build-plan.md
  - docs/design/death-mechanics.md
  - docs/design/entity-implementation-comparison.md
priority: high
ordinal: 19000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
As a player, I want the walls and torches I built to be the reason my friends are still here in the morning, so that base-building carries the weight of protecting people rather than protecting loot.

**Scope boundary.** **Preconditions, verified before any implementation** (rulings R-4, R-5): whether POI re-claim has natural lag after `releaseAllPois()`, and where the zombie-siege trigger can be Mixin-suppressed plus whether a 3-villager cast satisfies the (version-dependent, inconsistently documented) village-eligibility thresholds at all — **suppress regardless of eligibility**. **Admitted, no work needed:** hostile-mob night targeting, falls, drowning, fire/lava, player direct and indirect kills — all inherited free, all legible as neglect rather than luck. **Suppressed:** zombie sieges — an event that can overwhelm well-built defences on a nightly dice roll teaches the player that walls do not help, which is the fragility posture inverted. **One conversion-cancel Mixin** (permadeath; conversion is ruled equivalent to death in v1). **Remains, authored because free is invisible:** a mod-placed grave marker at the death location (or nearest safe buildable surface) with **no villager agency required** — a 2-3 survivor cast cannot be relied on to volunteer, and the memorial gesture cannot be optional; a **belongings bundle** capturing the hidden inventory *before* vanilla destroys it, placed at the grave as an ordinary `roles: ["storage"]` thing named for its owner; an **optional** job-board entry ("tend Tam's grave") a survivor may take up or ignore, riding V4's existing mechanism. **Grief period** per ruling R-3: bed and job-site held unclaimed for one in-game day/night cycle (~20 real minutes at 1x), **exposed as config, not a constant**. **Token discipline:** a new body token for the grave or converted mob; the dead villager's token is retired and never reissued.

**Done proves.** On a dev server: a villager killed by a zombie leaves a named grave at the death site with a belongings chest beside it; their bed stays unclaimed for the configured grief period; **no siege ever fires**; a witnessing villager receives ordinary `sighting` percepts (not a magic death broadcast) and an absent one receives a `change_report` with `change: "gone"` on return, plus a `sighting` of the grave; the dead body's token is never reissued.

**Depends on.** V2, V3.

**Escalation trigger (named, not a tier bump).** This is the one task standing on unverified engine surface (R-4, R-5, and decision-0002's flagged-thin `GossipManager` genericity finding). If the siege suppression point is not where death mechanics section 1 assumes, or if suppression turns out to need more than a targeted injection, that is an architecture question outside the implementing tier's scope: **stop and escalate** rather than growing the Mixin surface past decision-0002's committed bound.

**Design check — micromanagement.** This design *adds nothing* to villager self-preservation, on purpose: villagers cannot starve (there is no starvation loop at all), panic and flee autonomously, and seek shelter on their own schedule. No feeding UI, no escort quest, no babysitting surface. The siege suppression is the one addition, and it **removes** an arbitrary-death vector rather than adding a management surface.

**Design check — politeness-policing.** No engine guardrail on friendly fire. The embodied player has real physical agency including the capacity to do real harm, and adding a guardrail would be a fiction about who the player is. What the death *means* rides the memory channel (M7), not a judgment layer.

**References.** docs/design/demo-build-plan.md section 3.3 (V5) and its rulings R-3, R-4, R-5 are the plan of record. Ratified surfaces consumed: docs/design/death-mechanics.md (admitted/suppressed causes, grave + belongings + POI grief period, memory carry, section 6.2's open items), decision-0002 + docs/design/entity-implementation-comparison.md (the bounded Mixin budget, the GossipManager genericity caveat), docs/design/body-protocol-v0.md (change_report delivery restriction, token discipline), docs/design/kithcraft-brief.md (the micromanagement and politeness-policing spell-breakers).

**Suggested tier: `sonnet` with a named escalation trigger (next sweep's runbook decides).**
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Preconditions verified before implementation: whether POI re-claim lags naturally after releaseAllPois() (R-4), and where the zombie-siege trigger can be Mixin-suppressed plus whether a 3-villager cast qualifies at all (R-5)
- [ ] #2 Sieges are suppressed regardless of eligibility and no siege ever fires on a dev server run
- [ ] #3 One conversion-cancel Mixin makes zombie conversion equivalent to death, and the total Mixin surface stays inside decision-0002's committed bound
- [ ] #4 A villager killed by a zombie leaves a mod-placed named grave at the death site (or nearest safe buildable surface) with no villager agency required
- [ ] #5 A belongings bundle captures the hidden inventory before vanilla destroys it and is placed at the grave as an ordinary roles:[storage] thing named for its owner
- [ ] #6 An optional 'tend the grave' job-board entry rides V4's mechanism and a survivor may take it up or ignore it
- [ ] #7 The dead villager's bed and job site stay unclaimed for the configured grief period (default one in-game cycle per R-3), exposed as config rather than a constant
- [ ] #8 A witnessing villager receives ordinary sighting percepts (no magic death broadcast) and an absent one receives a change_report with change:'gone' on return plus a sighting of the grave
- [ ] #9 The dead body's token is retired and never reissued; the grave or converted mob gets a new body token
- [ ] #10 Design check (micromanagement): nothing is added to villager self-preservation - no feeding UI, no escort, no vigilance surface
- [ ] #11 Design check (politeness-policing): no engine guardrail on friendly fire
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Tests pass
- [ ] #2 Docs and wiki are updated and pass freshness tests
- [ ] #3 Spec and Backlog are in sync
<!-- DOD:END -->
