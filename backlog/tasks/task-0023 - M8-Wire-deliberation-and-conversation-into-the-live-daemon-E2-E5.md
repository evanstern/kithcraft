---
id: TASK-0023
title: M8 - Wire deliberation and conversation into the live daemon (E2-E5)
status: To Do
assignee: []
created_date: '2026-09-01 02:13'
updated_date: '2026-09-01 02:13'
labels:
  - mind
  - m-0-build
milestone: m-0
dependencies:
  - TASK-0016
  - TASK-0017
  - TASK-0021
priority: high
ordinal: 23000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
As a villager, I want my deliberation and my voice wired into the body I actually inhabit, so that the evening's beats — taking a job, building beside the player, talking at dusk — happen live and not only in tests.

**Why this task exists (operator ruling 2026-08-31, TASK-0022 checkpoint 5).** I1's runtime deliberately wired only E1 (genesis) and E6 (consolidation) — its plan recorded "deliberation/converse wiring beyond what the ACs need stays for I2's findings." TASK-0022's staging pass (run-kit.md §0, watch-list #1) found the consequence: mind/deliberate (E2/E3) and mind/converse (E4/E5) are merged, tested packages with NO caller in cmd/minddaemon, so coverage-map beats 4-6 (job-board claim, build-alongside, dusk conversation) cannot fire in a live run and A-4/A-5/A-6 + the E4 latency measurement would read zero. The operator chose wire-first over run-with-gaps.

**Scope boundary.** Assembly of merged, tested parts into Runtime — the judgment calls prior phases deliberately deferred are now settled by the written surfaces they cited: E2 triggers on schedule transitions and open choices arriving as percepts (routing §2.2's trigger list); E3 on text/origin:read (M5's TriggerE3, already written); the deliberation loop's Vendor/Proposer contract (M5's recorded Proposer/ErrDone convention) wired to the live session and the real prompt assembly (persona + window — M5's WindowItem snapshot fed from the daemon's stores); the §5.5 urgency interrupt registered on the live ingest path; dusk exchange driven off the pair-formation signal percepts (M6's Slot/Pool pregen + Exchange over the live session; body-to-persona identity binding — the empty-ConsolidationStablePrefix ponytail I1 left — closes here); E5 ambient pool refilled per cycle and served via speak intents; E4 FirstTokenLatency surfaced into session-report.log (TASK-0022 watch-list #6). Known adjacent gaps stay OUT of scope unless glue-sized en route (judged and recorded per precedent): the mod-side reconnect identity gap and heartbeat admissibility remain refactor-triage items.

**Done proves.** Against the fake vendor through the REAL daemon binary (not package tests): a board text percept yields a live claim-or-decline intent with an authored reason; an urgent percept mid-deliberation cancels and coalesces per §5.5; a scripted dusk pair signal produces a pre-generated opening and a multi-turn exchange with FirstTokenLatency in the session report; the ambient pool serves and refreshes. On a dev server with the daemon attached: at least one live E2/E3 deliberation and one dusk exchange observed end-to-end (honest not-observed records where the substrate's known timing questions bite).

**Depends on.** TASK-0016, TASK-0017, TASK-0021 (all merged). **Blocks TASK-0022's Phase 2** — the operator's evening follows this task's merge.

**References.** docs/design/llm-routing-and-budget.md (§2.2 triggers, §5.5 interrupt, §5.2 pre-generation), specs/016/017/021's recorded deferrals and conventions, specs/022-the-evening/run-kit.md §0 + watch-list (#1, #6), docs/design/demo-build-plan.md §4 (beats 4-6).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Runtime composes E2 deliberations on schedule-transition/open-choice percept triggers and E3 on text/origin:read, through mind/deliberate's real loop with persona + K=10 window context from the daemon's stores
- [ ] #2 The §5.5 urgency interrupt is registered on the live ingest path: cancel, no own call, exactly one coalesced follow-up
- [ ] #3 Dusk exchange runs off the live pair-formation signal percepts with pre-generation, through mind/converse; body-to-persona identity binding closes I1's empty-prefix ponytail
- [ ] #4 The E5 ambient pool refills per cycle and serves via speak intents; specific-remark escalation reachable
- [ ] #5 E4 FirstTokenLatency is surfaced in session-report.log (watch-list #6 closed)
- [ ] #6 Fake-vendor-through-real-binary proofs: live claim/decline with authored reason, interrupt coalescing, pre-generated dusk opening, multi-turn exchange, pool serve/refresh
- [ ] #7 Dev-server observation: at least one live E2/E3 deliberation and one dusk exchange end-to-end, with honest not-observed records where known substrate timing questions bite
- [ ] #8 No new Mixins; no protocol extension; known adjacent gaps (reconnect identity, heartbeat admissibility) stay out of scope unless glue-sized, judged and recorded
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Tests pass
- [ ] #2 Docs and wiki are updated and pass freshness tests
- [ ] #3 Spec and Backlog are in sync
<!-- DOD:END -->
