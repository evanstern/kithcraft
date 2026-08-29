---
id: TASK-0016
title: 'M5 - Deliberation and the job-board decision (E2, E3)'
status: Done
assignee: []
created_date: '2026-08-21 23:39'
updated_date: '2026-08-29 02:03'
labels:
  - mind
  - m-0-build
milestone: m-0
dependencies:
  - TASK-0010
  - TASK-0011
documentation:
  - docs/design/demo-build-plan.md
  - docs/design/llm-routing-and-budget.md
  - docs/design/body-protocol-v0.md
  - docs/design/kithcraft-brief.md
priority: high
ordinal: 16000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
<!-- SECTION:DESCRIPTION:BEGIN -->
As a villager, I want work orders to arrive on top of a life I am already living, so that taking one — or not — is a choice I made rather than a command I executed.

**Scope boundary.** Port `toolloop`'s bounded-loop shape — *a tool call is a REQUEST; an event is the FACT; the gate decides* — onto `intent`/`intent_ack`/`act_result`, the one-to-one map decision-0003 identified. Verb vocabulary from the runtime manifest, **not** from a compiled-in list. E2 routine deliberation at schedule transitions and open choices (~8/villager/cycle); E3 job-board deliberation on a `text` percept with `origin: read`, carrying its own context shape (board contents, other villagers' claims, standing relationship to the player, current commitments). The **urgency interrupt** exactly as routing section 5.5 states it: an `urgent` percept **cancels the in-flight deliberation**, **does not itself trigger a model call**, and **enqueues one deliberation** whose context includes it — because the body's reflex has already run. Memory window K=10 situated: top K-2 by recency-decayed weight plus **2 seeded serendipity picks from the older half** (the thing that stops a villager's context collapsing onto its five loudest days). Structured output, so an intent is a value rather than a text to interpret.

**Done proves.** Against the fake vendor: a scripted board posting yields a claim-or-decline intent carrying an **authored `reason`** (section 5.2 requires the mind to have a why). A decline is reachable and reads as this persona's decline, not a generic refusal. An `urgent` percept mid-deliberation cancels the call and produces exactly one follow-up deliberation — not three, and not zero. No intent names a target by description ("the nearest bed"); every target is a token the mind was given.

**Depends on.** M2, M4.

**Design check — micromanagement.** *Reluctance is the product* (brief #6, routing E3), but reluctance is not non-compliance forever: a villager who never takes a posted job turns the board into a chore the player must keep re-issuing, which is the failure mode inverted rather than avoided. The check: across a scripted evening's postings, work gets done without the player re-posting, and the refusals that do occur are legible as *this* villager's.

**Design check — politeness-policing.** A refusal must be grounded in the villager's own wants, commitments or relationship — never in the player's conduct. There is no compliance gate, no cooldown, and no "you were rude to me so I won't work" mechanic anywhere in this task. The player can be a jerk; that costs them a relationship, not an API.

**References.** docs/design/demo-build-plan.md section 3.2 (M5) is the plan of record. Ratified surfaces consumed: decision-0003 + docs/design/llm-routing-and-budget.md (E2/E3 classes, the urgency interrupt in section 5.5, the K=10 situated window with serendipity picks, the toolloop-to-intent map), docs/design/body-protocol-v0.md (intent/intent_ack/act_result, verbs from the runtime manifest, tokens-not-descriptions, Q-6's read channel), docs/design/kithcraft-brief.md (#6 reluctance; the micromanagement and politeness-policing spell-breakers).

**Suggested tier: `sonnet` (next sweep's runbook decides).**
<!-- SECTION:DESCRIPTION:END -->

Spec: specs/016-deliberation
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The bounded deliberation loop maps onto intent/intent_ack/act_result: a tool call is a request, an act_result is the fact, and the gate decides
- [x] #2 The verb vocabulary is read from the runtime manifest, never a compiled-in list
- [x] #3 Against the fake vendor a scripted board posting yields a claim-or-decline intent carrying an authored reason
- [x] #4 A decline is reachable and reads as this persona's decline, not a generic refusal
- [x] #5 An urgent percept mid-deliberation cancels the in-flight call, triggers no model call of its own, and enqueues exactly one follow-up deliberation whose context includes it
- [x] #6 The K=10 situated memory window is top K-2 by recency-decayed weight plus 2 seeded serendipity picks from the older half
- [x] #7 No intent names a target by description: every target is a token the mind was given
- [x] #8 Design check (micromanagement): across a scripted evening's postings work gets done without the player re-posting, and refusals are legible as this villager's
- [x] #9 Design check (politeness-policing): refusals are grounded in the villager's wants, commitments or relationship, never the player's conduct; no compliance gate, cooldown or lockout exists
- [x] #10 Spec phase: Phase 1 — The bounded loop (US1)
- [x] #11 Spec phase: Phase 2 — The job-board decision (US2)
- [x] #12 Spec phase: Phase 3 — Interrupt and window (US3 + US4)
- [x] #13 Spec phase: Phase 4 — Design checks, gates, and closure
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
claimed by sweep-0007-0022 orchestrator 2026-08-28 (lane 4); spec 016 stub + link ride this claim commit

tier: sonnet (default) · model cc/claude-sonnet-5[1m] · rubric: toolloop shape ports one-to-one per decision-0003; urgency interrupt and K=10 window written in routing §5.5/§6 — execution against a written spec, no unsettled judgment calls (runbook lane 4)

AC #1 (bounded loop -> intent/intent_ack/act_result): mind/deliberate/loop_test.go TestLoop_TreatsOnlyActResultAsFact + TestLoop_FactWiresIntoAdmissionGate.

AC #2 (verbs from manifest, no compiled-in list): mind/deliberate/manifest_test.go TestNoCompiledInVerbVocabulary + loop_test.go TestLoop_UndeclaredVerb_RefusedAsDeliberationFailure.

AC #3 (claim-or-decline intent w/ authored reason): mind/deliberate/board_test.go TestLoop_E3ClaimIntent_CarriesAuthoredReason + evening_test.go TestEveningOfPostings_WorkGetsDoneWithoutRepostingOrPolicing.

AC #4 (decline reachable, persona-grounded): mind/deliberate/board_test.go TestLoop_E3DeclineIntent_ReachableWithPersonaGroundedReason.

AC #5 (urgency interrupt: cancel, no own call, one coalesced follow-up): mind/deliberate/interrupt_test.go TestInterrupt_CancelsInFlightCall_NoOwnModelCall, TestInterrupt_CompletionWinsRace_StillCoalescesFollowup, TestInterrupt_CoalescesMultipleUrgentsIntoOneEnqueue, TestInterrupt_DrainOpensNewCoalescingWindow.

AC #6 (K=10 window, top K-2 decayed + 2 seeded serendipity from older half): mind/deliberate/window_test.go TestSelectWindow_TopKMinus2ByDecayedWeight, TestSelectWindow_SerendipityFromOlderHalf, TestSelectWindow_Deterministic, TestSelectWindow_GracefulUnderK, TestSelectWindow_DecayHalvesPerDayOfAge.

AC #7 (no descriptive targets, token-only): mind/deliberate/loop_test.go TestLoop_DescriptiveTarget_RejectedBeforeCompose + TestLoop_UnknownTokenTarget_RejectedBeforeCompose.

AC #8 (design check, micromanagement): mind/deliberate/evening_test.go TestEveningOfPostings_WorkGetsDoneWithoutRepostingOrPolicing — 3 scripted postings against the real FakeVendor over the real wire; exactly 1 intent per posting (no re-posting), no two reasons repeated, decline reason names a prior commitment.

AC #9 (design check, no politeness-policing): mind/deliberate/politeness_test.go TestNoPolitenessPolicingInDeliberationPath — structural grep proves no compliance/cooldown/conduct-keyed refusal path exists anywhere in the package.

AC #10-13 (spec phases 1-4): specs/016-deliberation/tasks.md T001-T013 all checked with citing proofs; go vet + go test -count=1 ./... green in mind/; wiki re-grounded (overview.md, promptworld-lineage.md re-pinned, CAPSULES regenerated, freshness gate green).
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
M5 delivered via PR #22 (merge 27be6ae, merge commit, pins preserved). mind/deliberate/: the bounded deliberation loop porting toolloop's REQUEST/FACT/gate shape onto intent/intent_ack/act_result — OnFact fires only on act_result delivery, wired to M2's admission gate end-to-end; verb vocabulary read solely from the session manifest (source-scan guard); token-only targets and non-empty authored reasons rejected before compose; E3 job-board deliberation on text/origin:read with the four-field §2.3 context in the variable suffix (stable prefix byte-identical); the §5.5 urgency interrupt as a mutex'd state machine that structurally cannot fire its own call and coalesces to exactly one follow-up; the K=10 situated window (top K−2 salience-halved-per-day + 2 seeded older-half picks). Scripted-evening design check: three postings, one intent each, no re-posting, persona-grounded decline. 33 tests -race green. Spec-bridge derivation: 4 phases 13/13, Done-eligible. ~692k subagent tokens across 4 sonnet dispatches (cc/claude-sonnet-5[1m], verified per dispatch).
<!-- SECTION:FINAL_SUMMARY:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Tests pass
- [ ] #2 Docs and wiki are updated and pass freshness tests
- [ ] #3 Spec and Backlog are in sync
<!-- DOD:END -->
