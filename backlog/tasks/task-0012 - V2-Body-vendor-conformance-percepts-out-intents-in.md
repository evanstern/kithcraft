---
id: TASK-0012
title: 'V2 - Body-vendor conformance: percepts out, intents in'
status: In Progress
assignee: []
created_date: '2026-08-21 23:37'
updated_date: '2026-08-25 20:14'
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
- [ ] #1 The vendor emits the four-type percept floor plus its declared extras, each provenance-stamped at emission from the closed origin vocabulary, and a mind ingests them without rejection on a dev server
- [ ] #2 An observation carries its scanned vocabulary and yields a falsifiable absence claim
- [ ] #3 R-8 resolved: a hearing hook is verified and sound is declared, or no hook verifies and sound is declared unsupported
- [ ] #4 No salience, importance or weight field exists anywhere in the percept surface
- [ ] #5 No change_report reaches the body that caused the change or watched it happen
- [ ] #6 The four core verbs (go_to, speak, attend, wait) plus cancel execute, each returning exactly one act_result, with the ack acknowledging receipt only
- [ ] #7 go_to targeting a body resolves to that body's last-seen place, never its live position
- [ ] #8 unknown_target refuses only an unissued token; a known-but-gone referent is accepted and fails with target_gone after a walk
- [ ] #9 Protocol section 12's six leak passes run clean over captured payloads: no engine-native type, identifier or coordinate convention in any message shape
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Tests pass
- [ ] #2 Docs and wiki are updated and pass freshness tests
- [ ] #3 Spec and Backlog are in sync
<!-- DOD:END -->

## Comments

<!-- COMMENTS:BEGIN -->
created: 2026-08-25 20:14
---
Sweep runbook tier: sonnet · model cc/claude-sonnet-5[1m] (default tier — percept/intent conformance to the protocol's written surface; R-8 is verify-then-declare, not design). Served model recorded at dispatch.
---
<!-- COMMENTS:END -->
