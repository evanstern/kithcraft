# Feature Specification: Dusk conversation and the ambient pool (E4, E5)

**Feature Branch**: `task-0017-dusk-conversation` · **Spec dir**: `specs/017-dusk-conversation`

**Created**: 2026-08-28 · **Status**: Draft

**Input**: TASK-0017 / M6 — E4 conversation turns under the design's only hard
latency ceiling (< 3 s to first token), pre-generation off V3's pair-formation
signal, the interlocutor-model context slice; E5 ambient pool on Haiku 4.5,
< 200 ms serving, daily refresh, specific-remark escalation. Consumes
decision-0003 + docs/design/llm-routing-and-budget.md (§2 E4/E5 rows, §2.3
context shapes, §5.2 latency budget and lever 2 pre-generation, §5.3 pool),
docs/design/body-protocol-v0.md (speak → speech in earshot),
docs/design/kithcraft-brief.md (dusk-conversation beat; tedium and
politeness-policing spell-breakers). Plan of record: demo-build-plan.md §3.2 M6.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Dusk talk that opens instantly and ends naturally (Priority: P1)

As a player, I want to overhear my neighbours talking about the day, the work,
and me, so that the base I built stops being a place and starts being a household.

**Why this priority**: The dusk exchange is the demo's emotional payload — the
class the whole project exists to make good (routing E4 row).

**Independent Test**: Against the fake vendor with a scripted/mocked model, a
dusk exchange between two minds produces a multi-turn conversation about the
day, the work, and the player; measured first-token latency < 3 s; the
conversation reaches a termination condition, never a mid-sentence turn cap
(card ACs #2, #3, #7).

**Acceptance Scenarios**:

1. **Given** E4's class config, **Then** it is Sonnet 5, streaming on,
   `effort: low`, thinking off, cached prefix, `max_tokens` ~300 (card AC #1) —
   asserted against the ratified registry in `mind/llm/classes.go`.
2. **Given** two minds paired at dusk, **When** the exchange runs, **Then**
   turns alternate, each turn's context carries the E4 shape (transcript so far,
   interlocutor model — who this is, what I think of them, shared history — as
   its own slice, memory window), and the content references the day, the work,
   and the player (scripted; card AC #2).
3. **Given** first-token latency measured per turn (mock client with injected
   delay in tests; wall-clock instrumentation in the real path), **Then** the
   measurement exists and the test ceiling holds < 3 s (card AC #3).
4. **Given** a conversation, **Then** it ends by a termination condition (a
   model-signaled close or a natural-end heuristic), not a turn cap leaving two
   villagers mid-sentence (card AC #7).

### User Story 2 - The opening turn is already written (Priority: P1)

As a villager walking to the fire, I want my opening line ready when I arrive,
so that the scene opens instantly.

**Independent Test**: Consuming V3's pair-formation signal (the ~10 s-ahead
sighting percepts, R-7), the opening turn is generated during the walk and
served from the pre-generation slot at scene open (card AC #4).

**Acceptance Scenarios**:

1. **Given** the pair-formation signal for (A, B), **When** it arrives, **Then**
   A's opening turn generation starts immediately (E4 call), keyed to the pair.
2. **Given** the pair converges, **Then** the opening turn serves from the slot
   without a new call; **Given** the pairing never converges (signal fired but
   meeting aborted), **Then** the slot is discarded without being spoken.
3. **Given** V3's measured signal lead (1.82–4.96 s live, vs nominal ~10 s —
   TASK-0014 finding), **Then** pre-generation still pays: if the turn isn't
   ready at convergence, the scene falls back to a live streamed call (the
   ceiling still holds — this is the shorter-lead check M6 was asked to make).

### User Story 3 - Ambient texture that never repeats itself into a spell-breaker (Priority: P2)

As a player passing a villager at work, I want a greeting or a grumble that
sounds like them today, so that they don't read as a state machine.

**Independent Test**: One batched Haiku 4.5 call per villager per in-game day
yields ~8 persona-flavoured lines; the pool serves in < 200 ms measured; no line
repeats within a cycle; the pool refreshes daily (card AC #5).

**Acceptance Scenarios**:

1. **Given** E5's class config, **Then** it is Haiku 4.5, one batched call per
   villager per day, ~8 lines out (card AC #5).
2. **Given** pool serving, **Then** measured service time < 200 ms and a line
   already served this cycle is not served again; **Given** the pool exhausts,
   **Then** the sparing stall-line policy applies (used rarely, never as a
   prefix tic) (card AC #7).
3. **Given** a new in-game day, **Then** the pool refreshes (new batched call);
   yesterday's lines are gone.
4. **Given** a remark about something specific (a targeted prompt, not a
   passing greeting), **Then** it escalates to a live Haiku call instead of
   drawing from the pool (card AC #6).

### Edge Cases

- Pool empty (call failed / not yet generated): stall-line policy + retry next
  natural trigger; never a blocking call on the < 200 ms path.
- Pre-generation call still in flight at convergence: fall back to live
  streaming; discard the late result (at-most-one opening turn spoken).
- Both minds pre-generate: only the designated opener's slot is spoken first;
  the other's context treats the spoken line as transcript.
- Politeness-policing (card AC #8): resentment, grumbling, and to-the-face
  complaints are allowed content; no lecture/moralize/conduct-gating template
  anywhere in the E4/E5 prompt assembly — structural check on prompt text plus
  scripted-content check.

## Requirements *(mandatory)*

- **FR-001**: E4 runs with the ratified config (Sonnet 5, streaming, effort low,
  thinking off, cached prefix, ~300 max_tokens) — from the registry, not ad hoc.
- **FR-002**: Multi-turn dusk exchange between two minds with per-turn context
  (transcript, interlocutor slice, memory window) and a termination condition.
- **FR-003**: First-token latency is measured (instrumented, not assumed) and
  the test ceiling holds.
- **FR-004**: Pre-generation consumes the pair-formation signal; slot serve,
  abort-discard, and live fallback all covered.
- **FR-005**: E5 pool: one batched call/villager/day, ~8 lines, < 200 ms serve,
  no intra-cycle repeats, daily refresh, specific-remark escalation.
- **FR-006**: No lecture/moralize/conduct-gate mechanism in either class's
  assembly (card AC #8).
- **FR-007**: No live API calls in tests; the model client is mocked/scripted.

## Success Criteria *(mandatory)*

- All eight card ACs demonstrated by named tests (`go vet` + `go test ./...`
  green in `mind/`).
- Latency claims carry measurements, not assertions.
- Wiki notes whose sources this PR touches re-verified honestly; CAPSULES
  regenerated if descriptions change; board/spec in sync at PR time.
