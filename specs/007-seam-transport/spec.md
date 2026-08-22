# Feature Specification: Seam transport decision and wire pinning

**Feature Branch**: `task-0007-seam-transport` · **Spec dir**: `specs/007-seam-transport`

**Created**: 2026-08-21

**Status**: Draft

**Input**: TASK-0007 / S1 — decide the mind↔vendor transport among real wires only
(UDS, TCP, stdio) per decision-0003's narrowing, weighed against body-protocol-v0
§8's T-1..T-7; pin the wire with a framing spec and golden vectors.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The transport is decided, with the reasoning shown (Priority: P1)

As a future implementer, I want the wire chosen with T-1..T-7 each answered explicitly,
so that neither the Go nor the Java side re-litigates the transport when its first
inconvenient consequence appears.

**Why this priority**: Every other m-0 task encodes against this choice; an implicit or
partially-argued choice invites re-decision mid-sweep.

**Independent Test**: Read the decision record: each of T-1..T-7 has a named answer for
the chosen wire, the rejected wires each have a stated reason, and the record is a
Backlog decision awaiting operator ratification.

**Acceptance Scenarios**:

1. **Given** the three candidate wires (UDS, TCP, stdio), **When** the decision record is
   read, **Then** each of T-1..T-7 is answered one by one for the chosen wire and each
   rejected wire carries a reason (maps to card AC #1).
2. **Given** decision-0003's one-way narrowing, **When** any in-process transport is
   considered, **Then** it is out of scope by construction — only real wires appear.

---

### User Story 2 - The framing is pinned by a spec both sides can implement (Priority: P1)

As a future implementer, I want a framing/serialization spec as the spec-002 successor,
so that message boundaries, encoding, and the session lifecycle on the wire are one
document, not two implementations' guesses.

**Why this priority**: The protocol doc (spec-002's output) deliberately deferred Q-1;
this document closes it. M1 and V1 both read it before writing a line.

**Independent Test**: The framing spec exists under `docs/design/`, states message
delimiting, encoding, connection lifecycle, and how T-2 ordering and T-6 backpressure
ride the chosen wire, without leaking any engine-native convention.

**Acceptance Scenarios**:

1. **Given** the accepted body-protocol-v0 message shapes, **When** the framing spec is
   applied, **Then** every §-defined message has exactly one wire representation (maps to
   card AC #2).

---

### User Story 3 - Golden vectors both languages round-trip (Priority: P1)

As a future implementer, I want language-neutral fixture files — one per percept type,
one per intent shape, plus the `session_open` handshake — so that the Go and Java sides
prove agreement against the same bytes before either exists in earnest.

**Why this priority**: This is the card's "built independently without discovering they
disagree at first contact" — the vectors are the contact, moved to now.

**Independent Test**: A trivial Go encoder and a trivial Java decoder each round-trip
every vector (encode → bytes match fixture; decode fixture → semantic equality); both
run in CI-runnable form (maps to card ACs #3, #4).

**Acceptance Scenarios**:

1. **Given** the vector fixtures, **When** the trivial Go encoder runs, **Then** its
   output round-trips every fixture byte-exactly or by the spec's declared equivalence.
2. **Given** the same fixtures, **When** the trivial Java decoder runs, **Then** every
   fixture decodes and re-encodes to equality under the same rule.
3. **Given** the fixture set, **When** audited, **Then** there is one vector per percept
   type, one per intent shape, and one for `session_open` — none missing, none extra
   beyond declared error/edge vectors.

---

### Edge Cases

- A percept type or intent shape added to the protocol later: the framing spec must state
  how vectors are extended (additive versioning per protocol §9) without invalidating
  existing ones.
- A vector that both implementations round-trip but interpret differently: vectors pin
  bytes AND the decoded semantic form, not bytes alone.
- Scope creep: nothing beyond the decision record, the framing spec, and the vectors is
  built (card AC #5) — no daemon, no mod, no reusable transport library.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The transport MUST be chosen from exactly {UDS, TCP, stdio}, with T-1..T-7
  each answered explicitly for the choice and a stated reason per rejected wire.
- **FR-002**: The choice MUST land as a Backlog decision record (proposed → operator
  ratifies at PR review).
- **FR-003**: A framing/serialization spec MUST exist under `docs/design/` as the
  spec-002 successor: message delimiting, encoding, session lifecycle on the wire,
  ordering (T-2), backpressure semantics (T-6), and the versioning story's wire form.
- **FR-004**: Golden vectors MUST exist as language-neutral fixture files: one per
  percept type defined in body-protocol-v0, one per intent shape, plus the
  `session_open` handshake; each pins both bytes and expected decoded form.
- **FR-005**: A trivial Go encoder MUST round-trip every vector; a trivial Java decoder
  MUST round-trip every vector; both runnable as tests.
- **FR-006**: No engine-native type, identifier, or coordinate convention may appear in
  the framing spec or any vector (protocol §12 discipline applies to the wire too).
- **FR-007**: Nothing beyond FR-001..FR-006's artifacts is built.

### Key Entities

- **Decision record**: the Backlog decision — candidates, T-1..T-7 answers, choice,
  status proposed until operator ratification.
- **Framing spec**: `docs/design/` document defining the wire form of every protocol
  message.
- **Golden vector**: a fixture file pairing exact wire bytes with the expected decoded
  message, language-neutral, consumed by both implementations' tests.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of T-1..T-7 have explicit answers in the decision record.
- **SC-002**: 100% of protocol percept types and intent shapes have a golden vector, plus
  `session_open`.
- **SC-003**: Both round-trip suites pass over every vector with zero mismatches.
- **SC-004**: Zero engine-native leaks in spec or vectors under a §12-style audit.
- **SC-005**: The diff contains nothing outside the decision record, the framing spec,
  the vectors, and the two trivial round-trip harnesses.

## Assumptions

- Body-protocol-v0 (accepted) is the message-shape source of truth; this task adds the
  wire form and changes no message semantics.
- "Trivial" encoder/decoder means the minimum code that proves round-trip — not M1's or
  V1's production transport layer (those consume this task's output).
- The operator ratifies the decision at PR review (checkpoint 2 of the sweep runbook);
  merge is the ratification act, per house precedent from decisions 0001–0003.
- Vector fixtures live where both build systems can reach them (a top-level fixtures
  directory; exact path is a plan-time choice).
