---
id: TASK-0010
title: 'M2 - Event-sourced memory, the belief store, and the episodic admission gate'
status: Done
assignee: []
created_date: '2026-08-21 23:37'
updated_date: '2026-08-27 21:38'
labels:
  - mind
  - m-0-build
milestone: m-0
dependencies:
  - TASK-0008
documentation:
  - docs/design/demo-build-plan.md
  - docs/design/body-protocol-v0.md
  - docs/design/llm-routing-and-budget.md
priority: high
ordinal: 10000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
As a villager, I want to know only what I saw or was told, with provenance, so that what I say at dusk is honestly mine and not something I was handed.

**Scope boundary.** Reimplement event-sourced memory to the contract (~400 lines per decision-0003): append-only log, immutability **enforced in the schema, not in convention**, state as a reducer. Deliberately **not** carried: promptworld I's world-event vocabulary (the mind never sees it under SI-1) and its `log_format_version` migration chain (whose determinism-for-replay justification died with I). The private, provenance-stamped map — PM-1's spine, the reason two villagers see different worlds — kept distinct from the vendor's resolution index, which the mind never reads. RM-1..RM-7 reimplemented to the contract, porting nothing: witnessed claims, coerce-never-reject, secondhand never beating fresher firsthand, read-time confidence and freshness as arithmetic on `world_time`, and **only a correction, a death, or a witnessed removal deletes a fact — time alone never does**. The deterministic **episodic admission gate** (routing section 6.3): admit on urgency >= `notable`, any percept involving another body or the player, any `act_result` on an intent the mind authored a `reason` for, any `told_fact` or `text`, any first sighting of a `kind` or `place`; drop repeated `background` sightings of already-known things. Plus the one long-run instrument: **E6 input tokens per villager over time**, not dollars.

**Done proves.** The canonical end-to-end (protocol section 10.2) against the fake vendor, including its step 5: a mind told about the orchard cannot durably claim it *saw* apple trees there. The admission gate's instrument reports buffer size per villager-day. An attempt to mutate a logged event fails at the type level, not at review.

**Depends on.** M1.

**Design check — minds-are-others (constraint, structural).** The belief store has **no external write path**. Nothing outside the mind — not the vendor, not the player, not a debug command — may author a belief.

**References.** docs/design/demo-build-plan.md section 3.2 (M2) is the plan of record. Ratified surfaces consumed: docs/design/body-protocol-v0.md (SI-1, SI-5, PM-1, RM-1..RM-7, the canonical end-to-end in section 10.2), decision-0003 + docs/design/llm-routing-and-budget.md (event-sourced memory reimplemented not ported; the episodic admission gate in section 6.3; the E6-input-tokens instrument).

**Suggested tier: `sonnet` (next sweep's runbook decides).**

Spec: specs/010-event-sourced-memory
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Event-sourced memory is append-only with immutability enforced in the schema, not by convention: an attempt to mutate a logged event fails at the type level
- [x] #2 The private, provenance-stamped map is kept distinct from the vendor's resolution index, which the mind never reads
- [x] #3 RM-1..RM-7 hold: witnessed claims, coerce-never-reject, secondhand never beats fresher firsthand, confidence and freshness computed at read time from world_time, and only a correction, a death or a witnessed removal deletes a fact
- [x] #4 The deterministic episodic admission gate admits per routing section 6.3 and drops repeated background sightings of already-known things
- [x] #5 The canonical end-to-end (protocol section 10.2) passes against the fake vendor, including step 5: a mind told about the orchard cannot durably claim it saw apple trees there
- [x] #6 The E6-input-tokens-per-villager instrument reports admitted buffer size per villager-day
- [x] #7 Design check (minds-are-others): the belief store has no external write path from the vendor, the player or a debug command
- [x] #8 Spec phase: Phase 1 — The log and the reducer (US1)
- [x] #9 Spec phase: Phase 2 — Beliefs, provenance, the RM rules (US2)
- [x] #10 Spec phase: Phase 3 — Admission gate, instrument, end-to-end (US3)
- [x] #11 Spec phase: Phase 4 — Closure: gates, wiki, board
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Phase 4 (T011-T013) closure. AC evidence, each a named test in mind/memory/: AC#1 (mutation impossible at type level) - TestEvent_ImmutableAtTypeLevel, TestEvent_NoMutatingMethods, TestLog_ReplayReproducesStateByteForByte. AC#2 (private map distinct from vendor index, no vendor read path) - TestStore_NoReadPathToVendorState. AC#3 (RM-1..RM-7) - TestRM1_DirectPerceptionGatesOnOriginAlone, TestRM2RM3_CitationGateCoercesNeverRejects, TestRM4_SecondhandNeverOverwritesFresherOrEqualFirsthand, TestRM5_StoredConfidenceNeverMutatesEffectiveDecaysAtReadTime, TestRM6_FreshnessIsPerKindReadTimeArithmetic, TestRM7_DeletionOnlyViaCorrectionDeathOrWitnessedRemoval. AC#4 (admission gate + drop rule) - TestGate_AdmitsOnUrgencyAtLeastNotable, TestGate_AdmitsPerceptInvolvingOtherBodyOrPlayer, TestGate_AdmitsActResultWithAuthoredReason, TestGate_AdmitsToldFactAndText, TestGate_AdmitsFirstSightingOfKindOrPlace, TestGate_DropsRepeatedBackgroundSightingOfKnownThing, TestGate_Deterministic_SameStreamSameAdmissions. AC#6 (E6 instrument, buffer size per villager-day) - TestInstrument_BucketsByVillagerDay. AC#7 (no external write path) - TestStore_NoExternalWritePathBeyondIngestAPI.

AC#5 left unticked deliberately: its wording is 'against the fake vendor'. e2e_test.go's TestEndToEnd_ProtocolSection10_2_ToldCannotBecomeWitnessed proves step 5's epistemic assertion (told-about-orchard cannot durably become witnessed) mechanically, driven directly against this package's own API per FR-007/FR-006 scope, not against a scripted double session through a fake vendor - there is no memory wiring into the seam yet (that lands in M5), and the fake vendor proper is S2/TASK-0015's deliverable. Honesty over completion: the epistemic core is proven, the AC's literal wording is not yet true, so it stays open until S2/TASK-0015 and M5 land the missing pieces.

Phases 1-4 summary: P1 (T001-T003) append-only JSONL event log, unexported Event fields, (world_time,hash) identity, replay-verifies-hash-on-read. P2 (T004-T006) provenance.go's direct_perception classifier + RM-2/RM-3 coerce-never-reject citation gate; beliefs.go's private PM-1 store, RM-4 upsert rule, RM-5/RM-6 read-time confidence/freshness arithmetic, RM-7 closed-vocabulary retraction. P3 (T007-T010) admission.go's routing-section-6.3 gate (5 admit rules + drop rule, deterministic, no model call), instrument.go's per-villager-day buffer-size instrument, e2e_test.go's protocol-10.2 walkthrough proving step 5. P4 (T011-T013): go vet + go test green (kithcraft/mind/... all packages), diff scope confirmed clean (mind/, specs/010-*, backlog/, docs/design/sweep-0007-0022-runbook.md only), body-protocol-seam.md and overview.md amended and re-pinned to reflect M2's landing, CAPSULES.md regenerated for overview's changed description, 6 of 7 card ACs ticked on cited tests, AC#5 left honestly open pending S2/TASK-0015's fake vendor.

spec-bridge sync: Phase 1 — The log and the reducer (US1): 3/3 · Phase 2 — Beliefs, provenance, the RM rules (US2): 3/3 · Phase 3 — Admission gate, instrument, end-to-end (US3): 4/4 · Phase 4 — Closure: gates, wiki, board: 3/3 — status In Progress → Done

AC #5 closed by TASK-0015 T008: the section-10.2 canonical end-to-end now runs against the fake vendor (mind/fakevendor/e2e_test.go), completing the deliberate carry from PR #15.
<!-- SECTION:NOTES:END -->

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
Sweep runbook tier: sonnet · model cc/claude-sonnet-5[1m] (default tier — reimplementation to a written contract: RM-1..RM-7, admission gate per routing §6.3; judgment calls settled by decision-0003). Served model recorded at dispatch.
---
<!-- COMMENTS:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
All spec tasks complete (Phase 1 — The log and the reducer (US1): 3/3 · Phase 2 — Beliefs, provenance, the RM rules (US2): 3/3 · Phase 3 — Admission gate, instrument, end-to-end (US3): 4/4 · Phase 4 — Closure: gates, wiki, board: 3/3). Derived Done by spec-bridge sync.
<!-- SECTION:FINAL_SUMMARY:END -->
