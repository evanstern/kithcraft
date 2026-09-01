---
id: TASK-0023
title: M8 - Wire deliberation and conversation into the live daemon (E2-E5)
status: Done
assignee: []
created_date: '2026-09-01 02:13'
updated_date: '2026-09-01 03:47'
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

Spec: specs/023-daemon-wiring
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Runtime composes E2 deliberations on schedule-transition/open-choice percept triggers and E3 on text/origin:read, through mind/deliberate's real loop with persona + K=10 window context from the daemon's stores
- [x] #2 The §5.5 urgency interrupt is registered on the live ingest path: cancel, no own call, exactly one coalesced follow-up
- [x] #3 Dusk exchange runs off the live pair-formation signal percepts with pre-generation, through mind/converse; body-to-persona identity binding closes I1's empty-prefix ponytail
- [x] #4 The E5 ambient pool refills per cycle and serves via speak intents; specific-remark escalation reachable
- [x] #5 E4 FirstTokenLatency is surfaced in session-report.log (watch-list #6 closed)
- [x] #6 Fake-vendor-through-real-binary proofs: live claim/decline with authored reason, interrupt coalescing, pre-generated dusk opening, multi-turn exchange, pool serve/refresh
- [ ] #7 Dev-server observation: at least one live E2/E3 deliberation and one dusk exchange end-to-end, with honest not-observed records where known substrate timing questions bite
- [x] #8 No new Mixins; no protocol extension; known adjacent gaps (reconnect identity, heartbeat admissibility) stay out of scope unless glue-sized, judged and recorded
- [x] #9 Spec phase: Phase 1 — Deliberation live (US1 + US2)
- [x] #10 Spec phase: Phase 2 — Conversation live (US3 + US4)
- [x] #11 Spec phase: Phase 3 — Proofs (FR-006 + FR-007)
- [x] #12 Spec phase: Phase 4 — Gates and closure
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
claimed by sweep-0007-0022 orchestrator 2026-08-31 (lane 5 amendment — operator wire-first ruling at checkpoint 5); spec 023 stub + link ride this claim commit

tier: sonnet (default) · model cc/claude-sonnet-5[1m] · rubric: assembly of merged tested parts; the deferred judgment calls are settled by the written surfaces the card cites (M5's Proposer/ErrDone + WindowItem, M6's Slot/Pool + AmbientPool, I1's Runtime conventions) — no unsettled design (runbook lane 5 amendment)

AC #1 (E2/E3 composition) — TICKED. Proven by TestLiveE3Deliberation_PerceptInIntentOutWithAuthoredReason and TestDeliberation_NilClient_LogsAndSkips / TestDeliberation_NoBoundPersona_LogsAndSkips (mind/cmd/minddaemon/deliberation_test.go), through the real listener/wire, not a package double. Persona+K=10 window context via windowSnapshot/personaFor (deliberation.go, T002).

AC #2 (§5.5 interrupt) — TICKED. TestLiveInterrupt_CancelsInFlightAndProducesOneFollowUp (deliberation_test.go, T003): in-flight call to a gated fake model server cancelled, no intent sent for it, the urgent fires no call of its own, exactly one coalesced follow-up lands as a live intent — green under -race.

AC #3 (dusk exchange + persona binding) — TICKED, WITH A LIVE-BLOCKED CAVEAT. The daemon machinery is proven: TestDuskExchange_PairSignalConvergesAndRecordsLatency (T005/T008) — scripted pair signal through the real binary produces a pregen Fill, convergence, a multi-turn Exchange, alternating-body speak intents, both speakers' latencies recorded; TestConsolidationPrefixFor_CarriesPersonaText (T004) proves the ConsolidationStablePrefix ponytail is closed at the code level — the empty-prefix gap I1 left is genuinely fixed, not papered over. CAVEAT (T009 finding, full text in the dedicated operator note below): on the actual dev-server mod build, the one live-attached body's own persona cannot bind via the rt.Personas[body]==CastID path, because the live body token is an opaque per-boot value, never a cast name — so this AC's machinery is proven correct and wired, but not yet exercisable end-to-end on a live server without a follow-up decision. Judged satisfied by the daemon-level proof; the live gap is AC #7/#8's territory, not re-litigated here.

AC #4 (ambient pool) — TICKED. TestAmbientTrigger_ServeAndEscalate (T006/T008): day-crossing refill, a generic sighting served from the pool, a specific one escalated to a live call — all as live speak intents in order, -race green.

AC #5 (FirstTokenLatency in report) — TICKED. report.go's per-villager section (T007), asserted non-empty for both speakers by TestDuskExchange_PairSignalConvergesAndRecordsLatency's report check (T008).

AC #6 (fake-vendor-through-real-binary proof set) — TICKED. TestFR006_ProofSet (conversation_test.go, T008) names all four: (a) TestLiveE3Deliberation_PerceptInIntentOutWithAuthoredReason, (b) TestLiveInterrupt_CancelsInFlightAndProducesOneFollowUp, (c) TestDuskExchange_PairSignalConvergesAndRecordsLatency, (d) TestAmbientTrigger_ServeAndEscalate. All run through the real daemon binary/listener, -race clean.

AC #7 (dev-server live observation) — LEFT UNTICKED. Rehearsal mode this session (no ANTHROPIC_API_KEY sanctioned); bring-up, session attach, and mod-side self_state emission all confirmed live, but neither a live E2/E3 trigger nor a dusk exchange reached the daemon. Not a timing shrug — three structural findings, all cited in full in specs/023-daemon-wiring/research/live-wiring-observation.md §§3-5 and in the dedicated operator note below. Honest not-observed record per FR-007.

AC #8 (no new Mixins/protocol surface; adjacent-gap judgments) — TICKED. mod/ is untouched by this task (git diff origin/main --stat -- mod/ is empty) so no new Mixins; the only new Go-side surface is seam.Ingester.OnSessionOpen, a nil-by-default mind-side hook (same idiom as OnPercept/Archived) — no new percept, verb, manifest field, or wire change, documented in docs/wiki/body-protocol-seam.md's 2026-08-31 re-verification. Adjacent-gap judgments: the live persona-binding gap (T009 §3) is judged NOT glue-sized — it needs a real design decision (mod passes cast identity into session_open, or the daemon learns name<->token some other way), not a one-liner, so it is flagged for the operator rather than patched here. DuskPairing's never-sent-live signal (§4) and self_state daemon-reachability (§5) are reconfirmed pre-existing, out of scope per the same precedent specs/021 already set.

Phase 4 gates and re-ground (T010/T011): go vet clean; go test -count=1 -race ./... green across all mind/ packages (11 packages, mind/seamtest has no test files); mod/ untouched vs origin/main (gradle not run — nothing to check); scope confined to mind/ + specs/backlog/runbook bookkeeping. docs/wiki/overview.md and docs/wiki/body-protocol-seam.md re-verified/re-pinned (body-protocol-seam.md was STALE, mind/seam/ingest.go had changed under T002); docs/wiki/promptworld-lineage.md checked and left unchanged (its toolloop-port language already covers this work). CAPSULES.md regenerated. Freshness gate: exit 0, 11/11 notes fresh against their pinned sources. Commit 3d04c0d.

OPERATOR-FACING FINDINGS (Phase 4 sign-off / refactor-triage input) — three items from T009's dev-server observation, full text in specs/023-daemon-wiring/research/live-wiring-observation.md:

(1) NEW — DECISION NEEDED BEFORE THE EVENING. The live mod's BodySession.open mints an opaque per-boot body token (ground.issueBody, e.g. "b-14"), never a cast member's name, so HandleSessionOpen's rt.Personas[body] lookup can never bind the live-attached body's own persona — contradicts the "body token IS a CastID" convention T004's wiring assumed. Judged NOT glue-sized: closing it needs a real design decision (e.g. the mod passing the matched Cast.Member's name into BodySession.open / a session_open extension, or the daemon learning a name<->token mapping some other way), not a one-liner. Flagged for the operator, not patched in this dispatch. Until this is decided, no live E2/E3 deliberation or dusk exchange can run with a correctly-bound persona on a real server.

(2) PRE-EXISTING, RECONFIRMED. DuskPairing's pairing-signal sighting is composed and logged mod-side but never sent over a live WireClient session — KithcraftMod's single-BodySession scope is a structural ceiling, already documented in the mod's own class javadoc before this session ran. A dusk pair needs both members' own live sessions to send their half of the signal; with only one villager ever live-attached today, the second signal can never reach the wire regardless of wait time. Refactor-triage candidate, not new debt from TASK-0023.

(3) PRE-EXISTING, RECONFIRMED. Whether self_state heartbeats reach the daemon at all is still not root-caused (specs/021-demo-config's own prior open question) — more evidence gathered this session (a longer, cleaner observation window, the pause-when-empty gotcha fixed), still not root-caused; packet-level capture on the UDS socket would be needed to go further. Refactor-triage candidate, not new debt from TASK-0023.

Item (1) is the one item requiring an operator decision before the evening's live run; (2) and (3) are pre-existing and best routed to refactor-triage.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
M8 delivered via PR #27 (merge 92edf49, merge commit, pins preserved). E2-E5 wired into cmd/minddaemon: live Vendor promoting the evening-test envelope contract (reconnect-safe); per-body deliberation loops with §2.2/TriggerE3 classification, K=10 window over the existing Log.Events() export, ErrDone honored; §5.5 interrupt on live ingest proven through the real listener under -race; persona binding at attach with E6/E2/E4 prefixes gaining real persona text (I1's ponytail closed); dusk exchange off the pair-signal shape with pregen + convergence-on-second-signal + 30s timeout; ambient pool refill/serve/escalate; FirstTokenLatency per villager in the session report (watch-list #6 closed). FR-006 proof set consolidated (TestFR006_ProofSet, all four beats through the real binary, multi-turn dusk). One additive seam hook (Ingester.OnSessionOpen, mind-side not wire). AC #7 honestly UNTICKED: the dev-server observation (rehearsal, zero calls) found live persona binding STRUCTURALLY blocked — BodySession mints opaque per-boot tokens, never CastIDs — the operator's pre-evening decision; DuskPairing signal-over-live-session and self_state admissibility reconfirmed as refactor-triage items. Spec-bridge derivation: 4 phases 12/12, Done-eligible with the one honest AC gap recorded. ~1.07M subagent tokens across 4 sonnet dispatches (cc/claude-sonnet-5[1m], verified per dispatch).
<!-- SECTION:FINAL_SUMMARY:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 Tests pass
- [x] #2 Docs and wiki are updated and pass freshness tests
- [x] #3 Spec and Backlog are in sync
<!-- DOD:END -->
