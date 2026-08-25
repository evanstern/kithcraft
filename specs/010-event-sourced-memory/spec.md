# Feature Specification: Event-sourced memory, belief store, admission gate

**Feature Branch**: `task-0010-event-sourced-memory` · **Spec dir**: `specs/010-event-sourced-memory`

**Created**: 2026-08-25 · **Status**: Draft

**Input**: TASK-0010 / M2 — event-sourced memory reimplemented to the contract, the
private provenance-stamped belief store, the deterministic episodic admission gate
(routing §6.3), and the E6-input-tokens instrument. Consumes body-protocol-v0.md
(SI-1, SI-5, PM-1, RM-1..RM-7, §10.2) and decision-0003 + llm-routing-and-budget.md.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A memory that cannot be rewritten (Priority: P1)

As a villager, I want my memory to be an append-only record of what actually reached
me, so that nothing I know can be silently edited after the fact.

**Why this priority**: Immutability enforced in the schema — not convention — is the
card's AC #1 and the foundation every other rule stands on.

**Independent Test**: Attempt to mutate a logged event; the attempt fails at the type
level (compile error or schema-level ABORT), not at review.

**Acceptance Scenarios**:

1. **Given** a logged memory event, **When** code attempts to modify it, **Then** the
   attempt is impossible by construction (card AC #1).
2. **Given** the event log, **Then** belief-store state is a reducer over the log —
   replaying the log reproduces the state.

### User Story 2 - Beliefs with honest provenance (Priority: P1)

As a villager, I want every belief I hold to carry what kind of evidence backs it, so
that what I say at dusk is honestly mine.

**Independent Test**: The canonical end-to-end (protocol §10.2) against the in-test
double/fake vendor, including step 5: a mind told about the orchard cannot durably
claim it *saw* apple trees there.

**Acceptance Scenarios**:

1. **Given** a model-authored belief citing percepts, **When** the deterministic gate
   resolves citations, **Then** provenance is **coerced** down to what the evidence
   supports — witnessed→told→inferred — never rejected (RM-2, RM-3), and coercion
   count is recorded.
2. **Given** a told fact about something the mind has fresher firsthand knowledge of,
   **Then** the told fact does not overwrite it (RM-4).
3. **Given** a stored belief aging, **Then** stored confidence never changes;
   effective confidence and freshness are computed at read time from `world_time`
   arithmetic (RM-5, RM-6); below the floor a belief stops driving behaviour but
   remains revisable.
4. **Given** any fact, **Then** it is deleted only by a correction, a death, or a
   witnessed removal — time alone never deletes (RM-7); staleness hides, never deletes.
5. **Given** the private provenance-stamped map (PM-1), **Then** it is distinct from
   any vendor resolution index and the mind has no read path to the vendor's (card
   AC #2, SI-1).

### User Story 3 - A gate that makes the expensive thought affordable (Priority: P2)

As an operator, I want a deterministic filter deciding which percepts are eligible to
become memories, so that E6's input does not grow without bound.

**Independent Test**: Scripted percept streams exercise every admission rule and the
drop rule; the instrument reports admitted buffer size per villager-day.

**Acceptance Scenarios**:

1. **Given** routing §6.3's rules, **Then** the gate admits: urgency ≥ `notable`; any
   percept involving another body or the player; any `act_result` on an intent the
   mind authored a `reason` for; any `told_fact` or `text`; any first sighting of a
   `kind` or `place` (card AC #4).
2. **Given** a repeated `background` sighting of an already-known thing, **Then** it
   is dropped from the episodic buffer — which is not belief deletion and does not
   violate RM-7.
3. **Given** an evening of admitted percepts, **Then** the E6-input-tokens instrument
   reports buffer size per villager-day (card AC #6).
4. **Given** the gate, **Then** it is deterministic — no model call anywhere in it
   (routing §1.2).

### Edge Cases

- Design check (minds-are-others, card AC #7): the belief store has **no external
  write path** — not the vendor, not the player, not a debug command. The only write
  path is the mind's own ingest/consolidation flow.
- `observed_at: null` means maximally stale (§2.6) — freshness arithmetic must treat
  it as such, not as fresh or as an error.
- `hops` never promotes or demotes origin; the classifier reads only `origin` (§2.7).
- An `observation`'s absence claim is scoped by its `vocabulary` — remembering the
  scope is part of remembering the claim (§4.3).

## Requirements *(mandatory)*

- **FR-001**: An append-only memory event log with immutability enforced in the
  schema/types, state as a reducer over the log. NOT carried from promptworld I: its
  world-event vocabulary and `log_format_version` migration chain.
- **FR-002**: A private, provenance-stamped belief store (PM-1) distinct from any
  vendor resolution index; no external write path (card AC #7).
- **FR-003**: RM-1..RM-7 reimplemented to the contract, each with a named test:
  direct-perception gating via the §2.7 classifier, citation resolution with
  coerce-never-reject, secondhand-never-beats-fresher-firsthand, read-time
  confidence/freshness as `world_time` arithmetic, no silent forgetting.
- **FR-004**: The deterministic episodic admission gate per routing §6.3, with the
  drop rule for repeated `background` sightings of known things.
- **FR-005**: The E6-input-tokens instrument: admitted buffer size per villager-day.
- **FR-006**: The canonical end-to-end (protocol §10.2) runs against the existing
  in-test double (mind/seamtest), including step 5's epistemic assertion.
- **FR-007**: Scope: no LLM calls (M4), no deliberation (M5), no consolidation prompt
  (M7) — this task builds the store and gate those consume.

## Success Criteria *(mandatory)*

- **SC-001**: `go vet` + `go test ./...` green in the `mind/` module.
- **SC-002**: Every card AC (#1–#7) demonstrated by a named test.
- **SC-003**: The §10.2 end-to-end passes, step 5 included.
- **SC-004**: Mutation of a logged event is impossible at the type level.

## Assumptions

- body-protocol-v0.md and routing §6.3 are fixed inputs; contradictions surface to
  the orchestrator, never resolved unilaterally.
- Go stdlib only unless a dependency is justified in plan.md terms (decision-0003's
  SQLite-triggers suggestion is a *pattern* reference; an in-Go schema-enforced log
  satisfies "schema, not convention" if mutation is a compile-time impossibility).
