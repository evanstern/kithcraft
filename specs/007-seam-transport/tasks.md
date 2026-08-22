# Tasks: Seam transport decision and wire pinning

**Input**: Design documents from `specs/007-seam-transport/`
**Prerequisites**: plan.md, research.md, data-model.md, contracts/vectors.md, quickstart.md

**Organization**: Phases map 1:1 to the sweep's phase-scoped dispatches — one fresh
implementer per phase; handoff is this file's tick-state plus the branch's commits.
Story labels: US1 = the decision (spec User Story 1), US2 = the framing spec (US2),
US3 = vectors + round-trips (US3).

## Phase 1 — The decision and the T-matrix (US1)

**Goal**: The wire is chosen with T-1..T-7 answered one by one; the record exists as a
proposed Backlog decision.

**Independent test**: Read the decision record — every cell of the 7×3 matrix filled,
each rejected wire reasoned, status proposed.

- [x] T001 [US1] Fill the T-matrix: answer T-1..T-7 (body-protocol-v0 §8, verbatim
      criteria) for each of UDS, TCP (loopback), stdio, attributing each property to
      the wire or to the framing layer, in specs/007-seam-transport/research.md
      (replace the `?` skeleton)
- [x] T002 [US1] Choose the wire and write the rationale + per-rejected-wire reasons
      into the T-matrix analysis in specs/007-seam-transport/research.md
- [x] T003 [US1] Create the Backlog decision record via
      `backlog decision create` (status proposed, operator ratifies at PR) carrying
      the choice, the filled matrix, and the rationale — file lands under
      backlog/decisions/
- [x] T004 [US1] Record the decision's narrowing effects for M1 (TASK-0008) and V1
      (TASK-0009) — connection story each side must implement — as a section of the
      decision record

**Checkpoint**: The decision is made and recorded; framing can begin.

## Phase 2 — The framing/serialization spec (US2)

**Goal**: `docs/design/seam-wire-v0.md` exists as the spec-002 successor with all seven
sections per data-model.md.

**Independent test**: Every protocol message shape has exactly one wire representation;
T-2/T-5/T-6 mechanics are specified; leak audit statement present.

- [x] T005 [US2] Write docs/design/seam-wire-v0.md sections 1–2: connection story
      (listen/dial, T-3 sessions, T-4 restart/reconnect) and message
      delimiting/encoding (T-5) for the chosen wire
- [x] T006 [US2] Write docs/design/seam-wire-v0.md sections 3–4: ordering/`seq` on
      this wire (T-2, dedup interaction) and backpressure (T-6 `background` shedding;
      never-split-an-`observation` at the framing layer)
- [x] T007 [US2] Write docs/design/seam-wire-v0.md sections 5–7: wire form of the
      protocol §7 versioning story, the round-trip equality rule (byte-exact or a
      justified canonical equivalence), and the §12-style leak audit statement

**Checkpoint**: The wire form is fully specified; vectors can be authored against it.

## Phase 3 — Golden vectors and both round-trips (US3)

**Goal**: `seam/vectors/` holds the closed vector set; trivial Go encoder and Java
decoder each round-trip every vector.

**Independent test**: `cd seam/go-roundtrip && go test ./...` green;
`java RoundTrip.java ../vectors` green; vector census matches contracts/vectors.md
exactly.

- [x] T008 [US3] Author the nine percept-type vectors per contracts/vectors.md
      (percept_sighting … percept_change_report) in seam/vectors/, each pinning
      decoded form + exact wire bytes per seam-wire-v0.md
- [x] T009 [P] [US3] Author the intent-shape vectors (intent, intent_ack, cancel), the
      session vectors (session_open with full manifest, session_close), and the three
      declared error/edge vectors (err_missing_provenance, err_unknown_origin,
      intent_ack_refused) in seam/vectors/
- [x] T010 [US3] Write the trivial Go round-trip harness in seam/go-roundtrip/ (own
      minimal go.mod + one test file): decode → re-encode → equality for every
      non-error vector; pinned refusal behavior for error vectors
- [x] T011 [P] [US3] Write the trivial Java round-trip harness in seam/java-roundtrip/
      (single file, plain `java`, no Gradle) with the same obligations and no shared
      code with the Go harness
- [x] T012 [US3] Run both harnesses over the full vector set; fix vectors or framing
      spec until both are green with zero mismatches; record the run in
      specs/007-seam-transport/quickstart.md's expected-output terms

**Checkpoint**: The wire is pinned by executable agreement.

## Phase 4 — Closure: leak audit, wiki, board

**Goal**: Deliverables audited, grounding re-verified, card synced.

- [x] T013 Run the §12-method leak sweep over seam-wire-v0.md and every vector (no
      engine-native type/identifier/coordinate convention); record findings in
      docs/design/seam-wire-v0.md's audit section
- [x] T014 Verify FR-007 scope: `git diff origin/main...HEAD` contains nothing outside
      the decision record, seam-wire-v0.md, seam/, and the specs/007 + board paper
      trail
- [x] T015 Re-verify wiki notes whose prose this PR touches — at minimum
      docs/wiki/body-protocol-seam.md (Q-1 is now closed; sources grow by
      docs/design/seam-wire-v0.md) — amend prose, re-pin honestly, regenerate
      docs/wiki/CAPSULES.md if any description changed
- [x] T016 Tick this file's boxes as completed, check the card's ACs that are now
      true via `backlog task edit TASK-0007 --check-ac <n>`, and append a phase-done
      note to the board card

## Dependencies

- Phase 1 → Phase 2 → Phase 3 → Phase 4, strictly serial across phases (the framing
  needs the wire; vectors need the framing; audit needs everything).
- Within Phase 3: T008/T009 parallel; T010/T011 parallel after vectors exist; T012
  after all four.

## Implementation strategy

Strictly incremental: the decision is the MVP (US1 alone is a ratifiable deliverable);
framing and vectors make it binding. Nothing beyond FR-001..FR-006's artifacts is
built — the harnesses are throwaway-grade proof, replaced by M1/V1's real transports.
