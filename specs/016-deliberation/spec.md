# Feature Specification: Deliberation and the job-board decision (E2, E3)

**Feature Branch**: `task-0016-deliberation` · **Spec dir**: `specs/016-deliberation`

**Created**: 2026-08-28 · **Status**: Draft

**Input**: TASK-0016 / M5 — port `toolloop`'s bounded-loop shape onto
`intent`/`intent_ack`/`act_result`; E2 routine deliberation and E3 job-board
deliberation; the §5.5 urgency interrupt; the K=10 situated memory window.
Consumes decision-0003 + docs/design/llm-routing-and-budget.md (§2.3 context
shapes, §5.5 interrupt, §6 daemon demands), docs/design/body-protocol-v0.md
(intent surface, verbs from the runtime manifest, tokens-not-descriptions),
docs/design/kithcraft-brief.md (#6 reluctance; micromanagement and
politeness-policing spell-breakers). Plan of record: demo-build-plan.md §3.2 M5.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A choice, not a command (Priority: P1)

As a villager, I want work orders to arrive on top of a life I am already living,
so that taking one — or not — is a choice I made rather than a command I executed.

**Why this priority**: The bounded loop is the task's spine: a tool call is a
REQUEST; an event (`act_result`) is the FACT; the gate decides. Everything else
(interrupt, window, E3 context) hangs off this loop.

**Independent Test**: Against the fake vendor, a deliberation produces an intent,
receives `intent_ack` then `act_result`, and the loop treats only the
`act_result` as fact (card AC #1).

**Acceptance Scenarios**:

1. **Given** a deliberation emitting an intent, **When** the vendor acks and later
   resolves it, **Then** the mind's record of what happened derives from the
   `act_result`, never from the intent having been sent (AC #1).
2. **Given** the runtime manifest's verb list, **When** the deliberation offers
   verbs to the model, **Then** the offered set is read from the session manifest —
   no compiled-in verb list exists anywhere in the deliberation path (AC #2).
3. **Given** a model response naming a target, **Then** every target in the
   composed intent is a token the mind was given via percepts/manifest; a
   descriptive target ("the nearest bed") is rejected before compose (AC #7).
4. **Given** structured output (A-9: E2/E3 parse a value), **Then** an intent is
   decoded as a value, never scraped from prose.

### User Story 2 - The board posting and the authored why (Priority: P1)

As a player, I want a posted job to be answered by a villager's own yes or no with
their reason, so that the board feels read rather than executed.

**Independent Test**: A scripted board posting (a `text` percept with
`origin: read`) yields a claim-or-decline intent carrying an authored `reason`
(card AC #3); a decline is reachable and reads as this persona's (AC #4).

**Acceptance Scenarios**:

1. **Given** a `text` percept with `origin: read` carrying board contents,
   **When** E3 deliberation runs, **Then** its context carries the E3 shape:
   board contents, other villagers' claims, standing relationship to the player,
   current commitments (routing §2.3).
2. **Given** the resulting intent, **Then** it is claim-or-decline and carries a
   non-empty authored `reason` (§5.2: the mind must have a why) (AC #3).
3. **Given** a persona whose commitments conflict with the posting, **Then** a
   decline is reachable and its reason cites the villager's own wants,
   commitments, or relationships — not a generic refusal (AC #4).
4. **Given** a scripted evening of postings (micromanagement check), **Then**
   work gets done without re-posting, and the refusals that occur are legible as
   this villager's (AC #8).
5. **Given** the whole deliberation path (politeness-policing check), **Then** no
   compliance gate, cooldown, or player-conduct-keyed refusal mechanism exists;
   refusal grounds are only the villager's wants/commitments/relationship (AC #9).

### User Story 3 - The interrupt that doesn't panic (Priority: P2)

As a villager, I want a sudden danger to cancel what I was mulling and fold into
my next thought, so that my mind doesn't stack three panic calls while my body
has already fled.

**Independent Test**: An `urgent` percept mid-deliberation cancels the in-flight
call, triggers no model call of its own, and enqueues exactly one follow-up
deliberation whose context includes it — not three, and not zero (card AC #5).

**Acceptance Scenarios**:

1. **Given** an in-flight E2 call, **When** an `urgent` percept arrives, **Then**
   the call is cancelled via the client's cancellation primitive (RT-2).
2. **Given** the cancellation, **Then** no model call fires from the urgent
   percept itself; exactly one deliberation is enqueued, its context including
   the urgent percept (§5.5's middle clause — the body's reflex already ran).
3. **Given** several urgent percepts before the follow-up runs, **Then** still
   exactly one follow-up deliberation is enqueued (coalesced), each urgent
   percept present in its context.

### User Story 4 - A memory window that doesn't collapse (Priority: P2)

As a villager, I want my deliberation context to remember odd old days, not just
my five loudest ones, so that who I am stays wider than what just happened.

**Independent Test**: With a store of weighted memories, the selected window is
exactly top K−2 (K=10) by recency-decayed weight plus 2 seeded serendipity picks
from the older half (card AC #6).

**Acceptance Scenarios**:

1. **Given** memories with weights and ages, **Then** decay is salience halved
   per day of age (`world_time` arithmetic), top K−2 selected by decayed weight.
2. **Given** the older half of the store, **Then** 2 serendipity picks are drawn
   from it with a seeded (deterministic, per-villager) source — reproducible in
   tests.
3. **Given** fewer than K memories, **Then** selection degrades gracefully
   (no duplicates, no panic; all available memories, serendipity only when an
   older half exists).

### Edge Cases

- Manifest offers a verb the model never uses: fine. Model names a verb outside
  the manifest: refused at compose (existing `Pending` behaviour) and surfaced as
  a deliberation failure, not a crash.
- `act_result` for an unknown/expired intent: already ignored by seam bookkeeping.
- Cancellation racing normal completion: at-most-one of cancel/complete wins;
  either way exactly one follow-up deliberation results from the urgent percept.
- Empty board posting / unparseable board text: E3 still answers (decline with
  reason) rather than erroring silently.

## Requirements *(mandatory)*

- **FR-001**: A bounded deliberation loop maps request/fact/gate onto
  intent/intent_ack/act_result (card AC #1). Loop iterations are bounded; the
  bound is configuration with a default, not a magic number.
- **FR-002**: Verb vocabulary is read from the runtime session manifest (AC #2).
- **FR-003**: E3 fires on `text` percepts with `origin: read`; its context shape
  is board contents, other claims, player relationship, commitments (AC #3).
- **FR-004**: Every intent carries an authored `reason`; targets are tokens the
  mind was given (ACs #3, #7).
- **FR-005**: The §5.5 interrupt: cancel in-flight, no own call, exactly one
  enqueued follow-up carrying the urgent percept (AC #5).
- **FR-006**: K=10 situated window: top K−2 recency-decayed + 2 seeded
  serendipity picks from the older half (AC #6).
- **FR-007**: Persona-grounded declines; no compliance/cooldown/conduct
  mechanism (ACs #4, #8, #9) — verified structurally (no such code path) and
  behaviourally (scripted-evening test).
- **FR-008**: E2 triggers are schedule transitions and open choices; deliberation
  uses the E2/E3 class configs from `mind/llm` (structured output per A-9).

## Success Criteria *(mandatory)*

- All nine card ACs demonstrated by named tests against the fake vendor (no live
  API calls; the model client is mocked/scripted as in M4's tests).
- `go vet` + `go test ./...` green in `mind/`.
- Wiki notes whose sources this PR touches re-verified honestly; CAPSULES
  regenerated if descriptions change; board/spec in sync at PR time.
