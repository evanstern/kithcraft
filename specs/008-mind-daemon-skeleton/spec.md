# Feature Specification: Mind daemon skeleton

**Feature Branch**: `task-0008-mind-daemon-skeleton` · **Spec dir**: `specs/008-mind-daemon-skeleton`

**Created**: 2026-08-22 · **Status**: Draft

**Input**: TASK-0008 / M1 — Go daemon skeleton: process, vendor port, boundary decode,
session lifecycle. Consumes ratified decision-0004 (UDS, mind listens / vendor dials),
docs/design/seam-wire-v0.md, and seam/vectors/.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A boundary that cannot be talked past (Priority: P1)

As a future implementer, I want a daemon that refuses a malformed percept before it
touches any state, so that every later mind component is built on a hardened boundary.

**Why this priority**: Decision-0003 schedules boundary decode explicitly so V-5's cost
is paid once, up front. Every M-task builds on this.

**Independent Test**: Feed the daemon malformed and future-versioned inputs; observe
refusal-before-mutation and tolerant-reader fallbacks per V-1..V-6.

**Acceptance Scenarios**:

1. **Given** a percept missing `provenance`, **When** it arrives, **Then** it is
   rejected as malformed and no state mutates (card AC #3).
2. **Given** a percept whose `origin` is an unrecognized value from a future minor
   version, **When** decoded, **Then** it decodes successfully and classifies as
   secondhand (card AC #4).
3. **Given** any input, **Then** processing is presence-checked decode → validate →
   mutate, never interleaved (card AC #2), satisfying V-1..V-6: ignore unknown fields;
   fall back on unknown enum values; retain but never interpret an unknown
   `percept_type`; refuse an unknown verb; a missing required field is malformed, not
   defaulted; unrecognized or absent `origin` → secondhand.

### User Story 2 - A session that survives its own restart (Priority: P1)

As a future implementer, I want the daemon to hold sessions per the wire spec and
re-open with continuity, so that a mind restart never requires a vendor restart (T-4).

**Independent Test**: Open a session against the in-test double, kill the daemon
mid-session, restart, reconnect; the gap is reported as a gap, never backfilled.

**Acceptance Scenarios**:

1. **Given** the double dials the daemon's UDS listener, **When** `session_open`
   arrives, **Then** the handshake completes per seam-wire-v0.md §1 (version check
   fails closed; manifest ingested; per-body sessions multiplexed on one connection;
   byte-identical `capabilities` required across `session_open`s on a connection).
2. **Given** a daemon restart mid-session, **When** the double reconnects and re-opens
   with continuity, **Then** the daemon resumes without inventing what happened in the
   gap (card AC #6).

### User Story 3 - Percepts in, intents out, bookkeeping honest (Priority: P2)

As a future implementer, I want percept ingest and intent bookkeeping working against a
scripted double, so that M2/M4/M5 have a live skeleton to build into.

**Independent Test**: Card AC #5 — the daemon starts, opens a session against the
double, ingests a scripted stream containing duplicates and a `seq` gap, and emits
intents.

**Acceptance Scenarios**:

1. **Given** duplicate `percept_id`s across a reconnect, **Then** dedup drops them;
   **given** a `seq` gap within a session, **Then** it is recorded as shed-count
   evidence (seam-wire-v0.md: a gap can only mean deliberate `background` shedding).
2. **Given** an emitted intent, **Then** the pending set tracks it; `supersedes`
   replaces; `act_result` matches by `intent_id`; `cancel` works.

### Edge Cases

- The wire frames refused at the framing layer (oversize, short read, invalid UTF-8)
  per seam-wire-v0.md — connection-fatal, distinct from message-level refusal.
- The canonical encoder must be hand-rolled: Go `encoding/json` over-escapes beyond
  C-7 (recorded TASK-0007 finding).
- All 17 golden vectors must round-trip through THIS daemon's codec (the real one,
  replacing the throwaway harness's proof).

## Requirements *(mandatory)*

- **FR-001**: The daemon is its own Go module (`mind/`), outside any Gradle build; the
  vendor port is an interface declared at the consumer (card AC #1).
- **FR-002**: Boundary decode implements V-1..V-6 in decode → validate → mutate order.
- **FR-003**: Transport per decision-0004 + seam-wire-v0.md: UDS listener (unlink
  stale path before bind), length-prefixed canonical-JSON framing, 1 MiB cap,
  fail-closed version negotiation.
- **FR-004**: Session lifecycle per protocol §6: `session_open`/`session_close`,
  continuity (gap reported as gap), manifest ingest.
- **FR-005**: Percept ingest: `percept_id` dedup (reconnect scope), `seq` gap
  detection (shed accounting).
- **FR-006**: Intent bookkeeping: pending set, `supersedes`, `act_result` matching,
  `cancel`.
- **FR-007**: A minimal in-test double drives all of the above (S2 grows it later);
  the daemon's codec round-trips all 17 vectors in `seam/vectors/`.
- **FR-008**: Scope: no memory store (M2), no LLM client (M4), no deliberation (M5).

## Success Criteria *(mandatory)*

- **SC-001**: `go vet` + `go test ./...` green in the module; vector suite 17/17.
- **SC-002**: Every card AC (#1–#6) demonstrated by a named test.
- **SC-003**: Malformed-input tests prove zero mutation before validation completes.
- **SC-004**: Restart test proves continuity without backfill.

## Assumptions

- seam-wire-v0.md and decision-0004 are fixed inputs; any contradiction found is
  surfaced to the orchestrator, not resolved unilaterally.
- Go stdlib only unless a dependency is justified in plan.md terms.
