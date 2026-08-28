---
id: TASK-0015
title: S2 - Fake body vendor and the protocol-rule harness
status: In Progress
assignee: []
created_date: '2026-08-21 23:38'
updated_date: '2026-08-27 21:45'
labels:
  - seam
  - m-0-build
milestone: m-0
dependencies:
  - TASK-0008
  - TASK-0010
documentation:
  - docs/design/demo-build-plan.md
  - docs/design/body-protocol-v0.md
priority: high
ordinal: 15000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
<!-- SECTION:DESCRIPTION:BEGIN -->
As a future implementer, I want the seam's unprovable rules turned into failing-on-violation tests, so that a rule that looks like a stylistic preference in prose cannot be deleted by the next refactor.

**Scope boundary.** `FakeVendor` per protocol section 10.1 — manifest, open/close, `emit`, `advance`, `.acts`, `resolve`, and the two deliberate misbehaviour switches (`strict`, `restrict_change_reports`). The six named tests: **H-1** (malformed rejected, never defaulted, before any state mutation), **H-2** (unknown or absent origin classifies secondhand), **H-3** (the classifier is pure — a percept whose prose says "I saw this myself" is still secondhand), **H-4** (`direct` never appears on the wire), **H-5** (`target_gone` is the only non-existence channel; an *unissued* token refuses at ack, a *known-but-gone* one must accept and fail after a walk), **H-6** (the 75% flood, reproduced on purpose: `flooded.memory_count > 3 x restricted.memory_count`). Plus section 10.5's scope discipline, enforced: no autonomous behaviour, **no read API for the mind**, no capability the real vendors lack.

**Done proves.** All six tests green, and each one red when its rule is lifted. H-6 prints the ratio. The fake vendor exposes no method by which a mind can learn world state without acting.

**Depends on.** M1 (the vendor port is declared at the consumer), M2 (H-6 counts memories; the canonical end-to-end asserts a belief's origin class).

**Design check — minds-are-others (constraint, structural).** Section 10.5's "no read API" is this constraint's structural defence: the fake vendor is the single most convenient place in the codebase to add `vendor.things_near(body)` for a test's benefit, and doing so builds an omniscience bug into the reference implementation.

**References.** docs/design/demo-build-plan.md section 3.1 (S2) is the plan of record. Ratified surfaces consumed: docs/design/body-protocol-v0.md (section 10.1 the fake vendor, H-1..H-6, section 10.2 the canonical end-to-end, section 10.5 scope discipline), decision-0003 + docs/design/llm-routing-and-budget.md (the harness as a first-class task: how mind work proceeds before the mod exists).

**Suggested tier: `sonnet` (next sweep's runbook decides).**
<!-- SECTION:DESCRIPTION:END -->

Spec: specs/015-fake-vendor-harness
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 FakeVendor implements protocol section 10.1: manifest, open/close, emit, advance, .acts, resolve, plus the strict and restrict_change_reports misbehaviour switches
- [x] #2 H-1 through H-6 are green, and each turns red when its rule is lifted
- [x] #3 H-6 reproduces the 75% flood and prints the ratio (flooded.memory_count > 3x restricted.memory_count)
- [x] #4 Design check (minds-are-others): the fake vendor exposes no read API and no method by which a mind can learn world state without acting
- [x] #5 The fake vendor has no autonomous behaviour and no capability the real vendors lack (section 10.5 scope discipline)
- [ ] #6 Spec phase: Phase 1 — FakeVendor shape and scope discipline (US2 + US3 groundwork)
- [ ] #7 Spec phase: Phase 2 — The cheap rules: H-1..H-4 (US1 groundwork)
- [ ] #8 Spec phase: Phase 3 — The structural rules: H-5 and H-6 (US1 close)
- [ ] #9 Spec phase: Phase 4 — Canonical end-to-end, gates, wiki, board (US2 close + closure)
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Sweep lane 3 claim (2026-08-27). Tier: sonnet · cc/claude-sonnet-5[1m] (default tier per runbook: protocol §10 specifies FakeVendor and H-1..H-6 in detail; execution against a written standard). Served model recorded at dispatch.

Dispatch record (2026-08-27): Phase 1 served by cc/claude-sonnet-5[1m] (VERIFIED from transcript). FakeVendor §10.1 shape landed with reflection surface-lock test; strict/restrict_change_reports flags are data-only until H-tests wire them (Phase 2/3).

AC #1: mind/fakevendor/fakevendor.go implements §10.1's full shape (manifest, open/close, emit, advance, .acts, resolve, strict, restrict_change_reports); proven by mind/fakevendor/fakevendor_test.go TestFakeVendor_Manifest_IsWireValid, TestFakeVendor_Advance_MovesTimeAndNothingElse, TestFakeVendor_DefaultIntentBehaviour_AckRecordWait, TestFakeVendor_Resolve_UnknownIntentID_LoudError, TestFakeVendor_Resolve_Twice_LoudError, TestFakeVendor_Emit_AfterClose_LoudError.

AC #2: H-1..H-6 green with mutation checks in mind/fakevendor/harness_test.go (H-1..H-4: TestH1_MissingProvenance_RejectedNeverDefaulted, TestH1_MissingProvenanceOrigin_RejectedNeverDefaulted, TestH2_UnknownOrigin_ClassifiesSecondhand, TestH2_AbsentOrigin_ClassifiesSecondhand, TestH3_ClassifierIgnoresProseHopsAndSource, TestH4_UnknownDirectFieldIgnored), mind/fakevendor/h5_test.go (H-5: TestH5_IssuedTokenGone_AcceptedThenFailsAfterAdvance, TestH5_UnissuedToken_RefusedUnknownTargetAtAck), mind/fakevendor/flood_test.go (H-6: TestH6_ChangeReportFlood_RestrictionCutsMemoryLoad) — each carries an inline mutation check proving red-when-lifted.

AC #3: mind/fakevendor/flood_test.go TestH6_ChangeReportFlood_RestrictionCutsMemoryLoad reproduces the flood scenario (3 bodies, restrict_change_reports off/on) and asserts flooded.memory_count > 3x restricted.memory_count, printing the ratio via t.Logf.

AC #4: mind/fakevendor/fakevendor_test.go TestFakeVendor_ExportedSurface_IsExactlySpec101 pins *FakeVendor's exported method and field set to exactly §10.1's shape via reflection — no read/query method can be added without failing this test.

AC #5: no autonomous behaviour proven by TestFakeVendor_Advance_MovesTimeAndNothingElse and TestFakeVendor_DefaultIntentBehaviour_AckRecordWait (advance/emit/intent do nothing beyond the script); no capability beyond the contract proven by TestFakeVendor_ExportedSurface_IsExactlySpec101 (mind/fakevendor/fakevendor_test.go).
<!-- SECTION:NOTES:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Tests pass
- [ ] #2 Docs and wiki are updated and pass freshness tests
- [ ] #3 Spec and Backlog are in sync
<!-- DOD:END -->
