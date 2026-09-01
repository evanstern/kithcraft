# Feature Specification: Wire deliberation and conversation into the live daemon (M8)

**Feature Branch**: `task-0023-daemon-wiring` · **Spec dir**: `specs/023-daemon-wiring`

**Created**: 2026-08-31 · **Status**: Draft

**Input**: TASK-0023 / M8 — the operator's wire-first ruling at runbook
checkpoint 5: wire the merged, tested E2/E3 (mind/deliberate) and E4/E5
(mind/converse) packages into `cmd/minddaemon`'s Runtime so the evening's
beats 4–6 happen live. Provenance: specs/022-the-evening/run-kit.md §0 +
watch-list #1/#6; I1's recorded minimal-wiring scope (specs/021 plan decision
2). Consumes llm-routing-and-budget.md (§2.2 E2 triggers, §5.5 interrupt,
§5.2 pre-generation), M5's recorded conventions (Proposer/ErrDone, WindowItem
snapshot, TriggerE3), M6's (Slot/Pool pregen, Exchange, AmbientPool),
I1's (Runtime, HandlePercept, session report).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A villager that thinks in its body (Priority: P1)

As a villager, I want my deliberation wired into the body I inhabit, so that
taking a job or declining it happens in the world, not only in a test.

**Independent Test**: Against the fake vendor through the REAL daemon binary:
a board `text`/`origin:read` percept yields a live claim-or-decline intent
with an authored reason crossing the session (card ACs #1, #6).

**Acceptance Scenarios**:

1. **Given** the live ingest path, **When** a percept matches an E2 trigger
   (schedule transition or open choice, per routing §2.2's trigger list —
   arriving as the percepts the vendor already sends) or E3's
   (`TriggerE3`), **Then** Runtime composes a deliberation through
   `mind/deliberate.Loop` with real context: the villager's persona (M3,
   bound via I1), the K=10 window built from the daemon's own stores
   (`SelectWindow` over a snapshot — closing M5's wiring deferral), and E3's
   four-field context where applicable (card AC #1).
2. **Given** the loop's intent, **Then** it crosses the live session as a
   wire intent (the Vendor contract implemented over `seam.Conn`), the
   act_result resolves it, and the fact reaches the admission gate — the
   REQUEST/FACT/gate mapping live end-to-end.
3. **Given** a deliberation with nothing further, **Then** ErrDone ends it
   cleanly (M5's recorded convention).

### User Story 2 - The interrupt, live (Priority: P1)

As a villager, I want sudden danger to cancel what I was mulling in the real
daemon, so that §5.5 holds outside the test harness.

**Independent Test**: An `urgent` percept mid-deliberation through the live
ingest path cancels the in-flight call, fires no call of its own, and
enqueues exactly one coalesced follow-up (card ACs #2, #6).

**Acceptance Scenarios**:

1. **Given** an in-flight live deliberation, **When** an urgent percept
   arrives on the session, **Then** `Interrupt` (M5's state machine,
   registered on the ingest path) cancels it; multiple urgents coalesce to
   one follow-up whose context includes them all.

### User Story 3 - Dusk, spoken (Priority: P1)

As a player, I want the dusk conversation to actually happen on my server, so
that the demo's emotional payload is live.

**Independent Test**: A scripted pair-formation signal through the real
binary produces a pre-generated opening and a multi-turn exchange whose
FirstTokenLatency lands in session-report.log (card ACs #3, #5, #6).

**Acceptance Scenarios**:

1. **Given** the pair-formation sighting percepts (V3's signal, arriving on
   live sessions), **When** Runtime detects a pair forming, **Then** the
   pregen Slot fills for the designated opener (M6's Pool), and at
   convergence the Exchange runs between the two minds over their live
   sessions — speak intents out, transcript built from the spoken turns.
2. **Given** the exchange's turns, **Then** each speaker's context uses its
   own persona and interlocutor slice — which requires the body-to-persona
   identity binding I1 ponytailed (empty ConsolidationStablePrefix): Runtime
   now maps session body tokens to bound personas, and E6's prefix gains the
   persona text too (card AC #3).
3. **Given** measured turns, **Then** FirstTokenLatency values appear in the
   session-end report (closing watch-list #6) (card AC #5).

### User Story 4 - Ambient texture, served (Priority: P2)

As a player passing a villager, I want today's greeting from today's pool.

**Independent Test**: The AmbientPool refills per in-game cycle and serves
lines as speak intents; a specific remark escalates to a live call
(card AC #4).

**Acceptance Scenarios**:

1. **Given** a day rollover (the world_time crossing Runtime already
   tracks), **Then** each villager's pool refills (one batched E5 call).
2. **Given** an ambient trigger, **Then** a pool line goes out as a speak
   intent; targeted triggers escalate per M6's IsTargeted/Escalate.

### Edge Cases

- No API key (rehearsal mode): E2–E5 wiring is present but the nil-client
  path logs-and-skips exactly as E6 already does — the daemon never panics
  for lack of a key.
- A villager with no bound persona (stub cast): deliberation/conversation
  skip with a log line, never a crash; genesis-bound casts get the full path.
- Session drops mid-exchange: the exchange aborts cleanly; no partial turn
  is spoken after reconnect (at-most-once per M6's slot semantics).
- Archived mind: no deliberation, no conversation (M7's refusal already
  guards session_open; Runtime guards its own composition paths too).

## Requirements *(mandatory)*

- **FR-001**: E2/E3 composition wired per US1; triggers from existing
  percepts only — no protocol extension, no new percept types.
- **FR-002**: §5.5 interrupt registered on live ingest (US2).
- **FR-003**: Dusk exchange + pregen off the live signal; body-to-persona
  binding closes I1's ponytail; E6's prefix gains persona text (US3).
- **FR-004**: AmbientPool refill/serve/escalate wired (US4).
- **FR-005**: FirstTokenLatency in session-report.log (watch-list #6).
- **FR-006**: All fake-vendor proofs run through the REAL daemon binary
  (build + exec, or main's wiring exercised directly), not only package
  tests (card AC #6).
- **FR-007**: Dev-server observation: at least one live E2/E3 deliberation
  and one dusk exchange end-to-end; honest not-observed records where known
  substrate timing questions bite (card AC #7).
- **FR-008**: No new Mixins; no protocol extension; adjacent gaps (reconnect
  identity, heartbeat admissibility) out of scope unless glue-sized — judged
  and recorded (card AC #8).
- **FR-009**: No live API calls in tests; live-call proofs only in the
  dev-server observation where the operator's key is present, else honest
  stub records.

## Success Criteria *(mandatory)*

- All eight card ACs demonstrated (tests through the real binary + the
  observation doc); `go vet` + `go test` green; gradle green if mod touched.
- Wiki re-verified honestly (overview's not-closed addendum shrinks);
  CAPSULES regenerated if descriptions change; board/spec in sync at PR.
