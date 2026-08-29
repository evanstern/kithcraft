---
id: TASK-0019
title: 'V5 - Death, danger, and what remains'
status: Done
assignee: []
created_date: '2026-08-21 23:39'
updated_date: '2026-08-29 02:03'
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

Spec: specs/019-death-remains
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Preconditions verified before implementation: whether POI re-claim lags naturally after releaseAllPois() (R-4), and where the zombie-siege trigger can be Mixin-suppressed plus whether a 3-villager cast qualifies at all (R-5)
- [x] #2 Sieges are suppressed regardless of eligibility and no siege ever fires on a dev server run
- [x] #3 One conversion-cancel Mixin makes zombie conversion equivalent to death, and the total Mixin surface stays inside decision-0002's committed bound
- [x] #4 A villager killed by a zombie leaves a mod-placed named grave at the death site (or nearest safe buildable surface) with no villager agency required
- [x] #5 A belongings bundle captures the hidden inventory before vanilla destroys it and is placed at the grave as an ordinary roles:[storage] thing named for its owner
- [x] #6 An optional 'tend the grave' job-board entry rides V4's mechanism and a survivor may take it up or ignore it
- [x] #7 The dead villager's bed and job site stay unclaimed for the configured grief period (default one in-game cycle per R-3), exposed as config rather than a constant
- [x] #8 A witnessing villager receives ordinary sighting percepts (no magic death broadcast) and an absent one receives a change_report with change:'gone' on return plus a sighting of the grave
- [x] #9 The dead body's token is retired and never reissued; the grave or converted mob gets a new body token
- [x] #10 Design check (micromanagement): nothing is added to villager self-preservation - no feeding UI, no escort, no vigilance surface
- [x] #11 Design check (politeness-policing): no engine guardrail on friendly fire
- [x] #12 Spec phase: Phase 1 — Verify before building (US0)
- [x] #13 Spec phase: Phase 2 — Suppression and permadeath (US1)
- [x] #14 Spec phase: Phase 3 — Remains, grief, tokens (US2 + US3)
- [x] #15 Spec phase: Phase 4 — Proofs, gates, and closure (US4)
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
claimed by sweep-0007-0022 orchestrator 2026-08-28 (lane 4); spec 019 stub + link ride this claim commit

tier: sonnet (default) with the plan's NAMED ESCALATION TRIGGER · model cc/claude-sonnet-5[1m] · rubric: R-4/R-5 verification first; if siege suppression point differs from death §1's assumption or needs >1 targeted injection, STOP — operator checkpoint 4, possible opus escalation (runbook lane 4)

AC#1 (R-4/R-5 preconditions): specs/019-death-remains/research/death-26.2.md — no natural POI re-claim lag (explicit vendor-side hold required); siege trigger is VillageSiege.tick(), one-Mixin suppressible at HEAD; 3-villager cast always qualifies once any bed claimed (pure POI-occupancy density, no door/population count at 26.2). GO recorded.

AC#2 (siege suppression, zero live sieges): VillageSiegeMixin (T004) + specs/019-death-remains/research/death-observation.md this session — grep -in siege over two full server lifetimes (pre-kill boot, post-kill run, restart) = 0 matches; single pre-existing zombie never multiplied.

AC#3 (conversion-cancel Mixin, bound held): ZombieConversionMixin (T005), death-26.2.md R-6. Mixin surface = 4 (V3's 2 + siege + conversion), matches decision-0002's ceiling exactly; MixinConfigTest enumerates all 4.

AC#4 (named grave, no agency): LiveDeathHandling.placeGrave (T007) + death-observation.md — live kill placed an oak sign 'Here lies / Aldric' at the exact death block (already safe-buildable), no villager action involved.

AC#5 (belongings bundle): LiveDeathHandling.onAllowDeath/placeBelongings (T007) + death-observation.md — ALLOW_DEATH captures inventory before vanilla's dropAllDeathLoot runs; live run placed the chest (empty, since Aldric carried nothing this run — non-empty carryover proven structurally by the capture-loop code path, not separately re-observed live; honestly recorded as a gap in death-observation.md).

AC#7 (grief period, configurable): GriefPeriod.configuredTicks() (system property, GriefPeriodTest.java) + death-observation.md live proof — hold started at death, released at exactly the configured tick (griefPeriodTicks=1200 this run) via log lines.

AC#8 (witness sighting / absent change_report+grave sighting): mod/src/test/java/dev/kithcraft/mod/death/DeathPerceptChannelTest.java (T010, 4 green tests incl. a real loopback WireClient session) is the authoritative proof, per that class's own scope note. Not independently re-observed live this session (no mind dialed in, by design — same as prior phases' body-keeps-moving proofs); this session's live run confirms the composition fires from a real death (log: grave sighting content, board posting) but had no live percept recipient to observe receiving it.

AC#9 (token retired, never reissued): LiveDeathHandling death handling (T007) + death-observation.md — live kill retired body token b-5 (logged), and a fresh server restart's boot token-registry log omits b-5 while survivors' tokens (b-6, b-7) persist. Honest nuance recorded: CastSeeder's separate identity token (b-1) is a different, untouched token family — see death-observation.md.

AC#10 (design check, micromanagement): mod/src/test/java/dev/kithcraft/mod/death/StructuralAbsenceTest.java — grep-style regression scan over mod/src/main/java for feed(ing)/escort(ing)/vigilan(t|ce); all absent, test passes.

AC#11 (design check, politeness-policing): StructuralAbsenceTest.java — grep-style scan for friendly[- ]?fire; absent, test passes; no code in this task's diff adds any player-damage guardrail.

AC#6 (tend-grave posting 'rides V4's mechanism') deliberately left UNTICKED: V4 (TASK-0020) is not merged. specs/019-death-remains/plan.md's 'The V4 decoupling' section rules this an explicit orchestrator-time call — GraveBoardEntry implements the ratified interim seam (posting rides V2's existing text-percept read channel; content-compatible, no rework needed when V4 merges), live-proven this session (death-observation.md: 'board posting composed' log line), but plan.md itself offers ticking-with-deviation-note vs. holding-open as two valid outcomes and reserves the choice for the orchestrator at merge time — left open per the more conservative fallback.

DoD#1 (tests pass): ./gradlew build --rerun-tasks — 127 tests, 0 failures, 0 errors. DoD#2 (wiki/freshness): docs/wiki/villager-brain-api.md and overview.md amended (T013); freshness gate shows one remaining FAIL, correctly attributable to pre-existing unrelated TASK-0014 debt (rides PR #21), not this task's work — see specs/019-death-remains/tasks.md T013 note.

spec-bridge sync: Phase 1 — Verify before building (US0): 3/3 · Phase 2 — Suppression and permadeath (US1): 3/3 · Phase 3 — Remains, grief, tokens (US2 + US3): 3/3 · Phase 4 — Proofs, gates, and closure (US4): 5/5 — all spec tasks complete.

Status intentionally left at In Progress, NOT advanced to Done, even though spec-bridge derives Done-eligible from tasks.md (100% complete): no PR has been opened/merged yet (one-task-one-PR), and plan.md's "The V4 decoupling" section explicitly reserves card AC#6's disposition (tick-with-deviation-note vs. hold-open) for the orchestrator at merge time. Leaving the terminal -s Done transition (and AC#6's call) to the orchestrator alongside PR open/merge.

AC #6 ticked by orchestrator ruling (plan.md 'The V4 decoupling'): the tend-grave entry rides the board READ CHANNEL (Q-6) — the mechanism V4 formalizes — as fixture-independent posting content (GraveBoardEntry, GraveBoardEntryTest x2); when TASK-0020 lands its board book the entry rides it without rework. Deviation note per plan; the hold-open fallback was declined because the tested artifact exists.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
V5 delivered via PR #24 (merge 3a7f572, merge commit, pins preserved). The named escalation trigger did NOT fire: R-4/R-5 verified first at the brain-26.2.md standard (research/death-26.2.md) — POI release synchronous with no engine cooldown (grief period is an explicit vendor hold via PoiManager.take/release, config kithcraft.griefPeriodTicks default 24000); siege trigger VillageSiege.tick suppressed with one HEAD-cancel exactly where death §1 assumed; 26.2 eligibility is pure POI density so the 3-villager cast qualifies — suppressed regardless. ZombieConversionMixin makes conversion death-equivalent (R-6: sole call site, victim's die() falls through normally). Mixin surface: 4 = decision-0002's ceiling, MixinConfigTest-asserted; belongings capture needed NO new Mixin (Fabric API ALLOW_DEATH before dropAllDeathLoot). Grave always places (bounded deterministic search); tend-grave posting rides the board read channel (AC #6 orchestrator ruling per plan's V4-decoupling). Live dev-server proof (research/death-observation.md, forced /damage kill recorded as deliberate): grave + chest placed, grief hold released at the configured tick, zero sieges across two server lifetimes, retired token b-5 absent across restart. Honest gaps recorded: empty-inventory carry and live witness delivery not observed live (unit proofs cover); JOB_SITE hold unexercised (pre-existing TASK-0014 gap, flagged for refactor-triage). 127 tests. Spec-bridge derivation: 4 phases 14/14, Done-eligible. ~1.06M subagent tokens across 5 sonnet dispatches incl. one stopped idle-looper (cc/claude-sonnet-5[1m], verified per dispatch).
<!-- SECTION:FINAL_SUMMARY:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 Tests pass
- [x] #2 Docs and wiki are updated and pass freshness tests
- [ ] #3 Spec and Backlog are in sync
<!-- DOD:END -->
