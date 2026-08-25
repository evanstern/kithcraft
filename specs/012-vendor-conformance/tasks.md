# Tasks: Body-vendor conformance — percepts out, intents in

**Input**: specs/012-vendor-conformance/ (spec.md, plan.md)
**Prerequisites**: mod/ toolchain + WireClient + TokenRegistry (V1, merged PR #14),
body-protocol-v0.md §2–§5/§12, decision-0002, demo-build-plan R-8

**Organization**: Phases map 1:1 to phase-scoped dispatches.

## Phase 1 — Percept surface: shapes, stamping, urgency (US1 groundwork)

**Goal**: The mod composes and emits every declared percept type, provenance stamped
at emission, honest to AR-1..AR-6.

**Independent test**: `gradle test` — shape/stamping tests green; no
salience-shaped field exists (a reflective/schema test proves it).

- [x] T001 Implement percept/Provenance.java (closed origin vocabulary, stamped at
      emission, source-is-immediate-teller, observed_at/received_at semantics) and
      percept/PerceptEmitter.java (envelope, urgency bands, seq, per-body streams
      over the V1 WireClient session)
- [x] T002 Implement percept/Sightings.java: sighting (prose-only doing) and
      observation with vocabulary scoping (subset of manifest, falsifiable absence
      claim); first-sighting and dedup/shedding per §4.11; nearest_hostile as the
      danger signal; percept/SelfState.java (felt origin, condition bands)
- [x] T003 Named tests: every emitted shape passes AR-1..AR-6 checks; provenance
      stamped correctly per type; card AC #4's no-salience-field proof; AC #2's
      absence-claim scope test

**Checkpoint**: everything the world says is honestly stamped and abstraction-clean.

## Phase 2 — R-8, told_fact/text, change_report restriction (US1 close)

**Goal**: The remaining percept types and the two verification-shaped obligations.

**Independent test**: `gradle test` green; R-8 record exists with evidence.

- [ ] T004 R-8: verify a hearing hook against MC 26.2 (research/r8-hearing-hook.md
      with evidence per the evidence rule); emit sound + declare in manifest if
      verified, else declare unsupported — record the outcome either way (card AC #3)
- [ ] T005 Implement told_fact and text emission (speech-adjacent channels; §4.6–4.7
      trust shapes) and percept/ChangeReports.java with the §4.10 delivery
      restriction: never to actor or witness (card AC #5) — named test with a
      three-body scenario (actor, witness, absent third party; only the third
      party receives)
- [ ] T006 Update the session manifest declaration (percept_types, origins,
      salient_kinds unchanged) to declare exactly what Phase 1–2 emit — L-7
      byte-identity must keep holding (re-run V1's ManifestTest)

**Checkpoint**: the emission half is contract-complete for the demo.

## Phase 3 — Act surface: verbs, ack/result, target resolution (US2)

**Goal**: Four verbs plus cancel execute; each intent yields receipt-ack then
exactly one act_result.

**Independent test**: `gradle test` for the decode/ack/result state machine; a
dev-server observation record for live verb execution.

- [ ] T007 Implement act/IntentHandler.java: intent decode (V-4 refuse unknown
      verb), intent_ack acknowledging receipt only, pending-intent bookkeeping,
      exactly one act_result per intent (card AC #6), cancel per §5.7
- [ ] T008 Implement act/TargetResolution.java per §5.6: body targets resolve to
      last-seen place never live position (card AC #7); unknown_target only for
      unissued tokens; known-but-gone accepted then target_gone after the walk
      (card AC #8) — named tests for all three cases
- [ ] T009 Implement act/Verbs.java: go_to/speak/attend/wait via vanilla Brain<E> +
      mod handlers per decision-0002 (enumerate any Mixin added); dev-server
      observation: each verb executed by a real villager body, act_results
      captured — recorded in the PR description

**Checkpoint**: the mind can act and only learns outcomes by perceiving them.

## Phase 4 — Leak passes, gates, wiki, board (US3 + closure)

- [ ] T010 Implement LeakPassTest: §12's six passes over captured payloads from the
      Phase 1–3 tests and the dev-server capture (card AC #9); fix any leak found
- [ ] T011 Dev-server end-to-end: villager body emits percepts a stub mind ingests
      without rejection (card AC #1) — recorded observation
- [ ] T012 gradle build + gradle test green; scope check: diff touches only mod/,
      specs/012-*, board files, runbook log row; Mixin additions (if any)
      enumerated against decision-0002's bound
- [ ] T013 Wiki: re-verify notes whose sources this PR touches (body-protocol-seam's
      mod/ sources; villager-brain-api if brain-surface facts were verified in
      passing) — amend honestly, re-pin, regenerate CAPSULES.md if descriptions
      changed
- [ ] T014 Tick this file, check card ACs now true (backlog CLI in-worktree), append
      phase-done note

## Dependencies

Phase 1 → 2 → 3 → 4 serial (stamping before restriction cases; percepts before
act_results; everything before the leak sweep).

## Implementation strategy

Emission-first: act_result is a percept, so the percept machinery must exist before
the act surface can complete its loop. R-8 verification early in Phase 2 so a
negative finding narrows scope rather than surprising Phase 4.
