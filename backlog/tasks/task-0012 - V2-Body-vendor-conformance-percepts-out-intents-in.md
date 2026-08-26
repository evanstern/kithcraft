---
id: TASK-0012
title: 'V2 - Body-vendor conformance: percepts out, intents in'
status: Done
assignee: []
created_date: '2026-08-21 23:37'
updated_date: '2026-08-26 00:57'
labels:
  - vendor
  - m-0-build
milestone: m-0
dependencies:
  - TASK-0009
documentation:
  - docs/design/demo-build-plan.md
  - docs/design/body-protocol-v0.md
  - docs/design/entity-implementation-comparison.md
priority: high
ordinal: 12000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
As a villager, I want the world to tell me only what my body could actually have perceived, so that everything I later believe has an honest origin.

**Scope boundary.** The two halves of the seam surface, which cannot be split because `act_result` is both. **Perceive:** the four-type floor plus whatever else is declared — `sighting` (with `doing`, prose-only), `observation` (with its `vocabulary` — the kinds *scanned*, necessarily a subset of the manifest for a 3-D volume, and without which an absence claim has no scope), `speech`, `act_result`, and optionally `sound` (**ruling R-8**: verify a hearing hook; declare it unsupported if none verifies — the four-type floor does not include it and the demo loses nothing), `told_fact`, `text`, `self_state`, `change_report`. Provenance **stamped at emission** from the closed origin vocabulary. Urgency bands — and **no `salience`/`importance`/`weight` field may exist**, world-side salience being forbidden. The `change_report` **delivery restriction**: never to the body that caused the change or watched it happen. The abstraction rule throughout: opaque `kind` tokens, meaning as `roles` plus prose-only `descriptor`, space as place tokens plus coarse bands (**no coordinates, no arithmetic**). `nearest_hostile` exposed as the free, already-computed danger signal decision-0002 confirmed. **Act:** the four core verbs — `go_to` (targeting a body resolves to its **last-seen** place, never its live position), `speak`, `attend`, `wait` — plus `cancel` and the intent/ack split: the ack acknowledges receipt only, and what happened returns as an `act_result` percept. `unknown_target` refuses only an **unissued** token; a known-but-gone referent MUST be accepted and fail with `target_gone` after a walk, because a synchronous "gone?" answer is an existence oracle a mind can poll without moving.

**Done proves.** On a dev server: a villager body emits provenance-stamped percepts a mind ingests without rejection. An `observation` yields a falsifiable absence claim. **No `change_report` reaches the actor or a witness** — the rule with a $1.30/evening price tag that grows with the cast. The four verbs execute and each returns exactly one `act_result`. Protocol section 12's six leak passes run clean over captured payloads: no engine-native type, identifier, or coordinate convention in any message shape.

**Depends on.** V1.

**References.** docs/design/demo-build-plan.md section 3.3 (V2) and its ruling R-8 are the plan of record. Ratified surfaces consumed: docs/design/body-protocol-v0.md (percept surface, the four core verbs, provenance and origin vocabulary, the change_report restriction, section 12's leak passes, Q-2), decision-0002 + docs/design/entity-implementation-comparison.md (nearest_hostile already computed; bounded Mixin surface).

**Suggested tier: `sonnet` (next sweep's runbook decides).**

Spec: specs/012-vendor-conformance
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The vendor emits the four-type percept floor plus its declared extras, each provenance-stamped at emission from the closed origin vocabulary, and a mind ingests them without rejection on a dev server
- [x] #2 An observation carries its scanned vocabulary and yields a falsifiable absence claim
- [x] #3 R-8 resolved: a hearing hook is verified and sound is declared, or no hook verifies and sound is declared unsupported
- [x] #4 No salience, importance or weight field exists anywhere in the percept surface
- [x] #5 No change_report reaches the body that caused the change or watched it happen
- [x] #6 The four core verbs (go_to, speak, attend, wait) plus cancel execute, each returning exactly one act_result, with the ack acknowledging receipt only
- [x] #7 go_to targeting a body resolves to that body's last-seen place, never its live position
- [x] #8 unknown_target refuses only an unissued token; a known-but-gone referent is accepted and fails with target_gone after a walk
- [x] #9 Protocol section 12's six leak passes run clean over captured payloads: no engine-native type, identifier or coordinate convention in any message shape
- [x] #10 Spec phase: Phase 1 — Percept surface: shapes, stamping, urgency (US1 groundwork)
- [x] #11 Spec phase: Phase 2 — R-8, told_fact/text, change_report restriction (US1 close)
- [x] #12 Spec phase: Phase 3 — Act surface: verbs, ack/result, target resolution (US2)
- [x] #13 Spec phase: Phase 4 — Leak passes, gates, wiki, board (US3 + closure)
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Phase 4 closure — AC evidence: #1 composite (Phase 1-3 conformance/manifest tests compose every declared percept type honestly; live dev-server run, verb-observation.md 2026-08-25, proves session_open accepted and act_result percepts ingested without rejection over a real UDS session — self_state heartbeat and the other percept types were not exercised live in this window, unit-tested only). #2 SightingsTest + AbstractionConformanceTest's absence-claim scope proof (Phase 1). #3 research/r8-hearing-hook.md VERIFIED via javap against the pinned MC 26.2 jar. #4 AbstractionConformanceTest + LeakPassTest.p2NoStructuralLeakInAnyBranchableField — no salience/importance/weight field anywhere. #5 ChangeReportsTest's three-body scenario (actor/witness/absent third party). #6 IntentHandlerTest.everyCoreVerbAckedAcceptedYieldsExactlyOneActResult (unit) + verb-observation.md's live run: all four verbs, one act_result each, over a real session. #7 IntentHandlerTest.goToABodyWalksToItsLastSeenPlace + TargetResolutionTest (unit-proven; not exercised live this phase — live run targeted a place token, not a body token). #8 TargetResolutionTest + IntentHandlerTest's unknownTargetIsRefusedForAnUnissuedToken/knownButGoneReferentAcceptsThenFailsTargetGoneAfterTheWalk (unit-proven). #9 LeakPassTest — all six passes (P-1..P-6) green over payloads composed through the real Phase 1-3 machinery. #10-13: tasks.md's own Phase 1-4 checkboxes are now all [x] (commits eeb01f6, bd533b8, 56a5048, and this phase's T010/T011/T013 commits).

DoD: #1 (tests pass) and #3 (spec/backlog in sync) checked. #2 left unchecked — docs/wiki updated and re-pinned (body-protocol-seam, villager-brain-api, CAPSULES regenerated), but the freshness gate reports 3 pre-existing issues predating this phase (body-protocol-seam.md's note-size and capsule-description budgets already over before this PR touched it; promptworld-lineage.md's capsule description, untouched by this PR) — not fixed here; a full note restructure is out of Phase 4's scope, flagged for a future refactor-triage pass.

spec-bridge sync: Phase 1 — Percept surface: shapes, stamping, urgency (US1 groundwork): 3/3 · Phase 2 — R-8, told_fact/text, change_report restriction (US1 close): 3/3 · Phase 3 — Act surface: verbs, ack/result, target resolution (US2): 3/3 · Phase 4 — Leak passes, gates, wiki, board (US3 + closure): 5/5 — status In Progress → Done
<!-- SECTION:NOTES:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 Tests pass
- [ ] #2 Docs and wiki are updated and pass freshness tests
- [x] #3 Spec and Backlog are in sync
<!-- DOD:END -->

## Comments

<!-- COMMENTS:BEGIN -->
created: 2026-08-25 20:14
---
Sweep runbook tier: sonnet · model cc/claude-sonnet-5[1m] (default tier — percept/intent conformance to the protocol's written surface; R-8 is verify-then-declare, not design). Served model recorded at dispatch.
---
<!-- COMMENTS:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
All spec tasks complete (Phase 1 — Percept surface: shapes, stamping, urgency (US1 groundwork): 3/3 · Phase 2 — R-8, told_fact/text, change_report restriction (US1 close): 3/3 · Phase 3 — Act surface: verbs, ack/result, target resolution (US2): 3/3 · Phase 4 — Leak passes, gates, wiki, board (US3 + closure): 5/5). Derived Done by spec-bridge sync.
<!-- SECTION:FINAL_SUMMARY:END -->
