# Phase 0 research — seam transport

Plan-level unknowns resolved here. The wire choice itself is NOT resolved here — it is
the task's deliverable, made by the dispatched implementer inside the decision record
and ratified by the operator. This file bounds the decision space and settles the
plumbing questions so the implementer spends judgment only where the spec left it.

## Decision: candidate set is closed — UDS, TCP (loopback), stdio

- **Rationale**: decision-0003 narrowed one-way to real wires; the card restates it.
- **Alternatives considered**: in-process/FFI — foreclosed deliberately (an SI-1 breach
  must be structurally impossible); anything brokered (message queue, HTTP server
  framework) — fails T-7's "process-separable but not process-required" spirit by
  adding a third process and is outside the card's closed set.

## The T-matrix the decision record must fill (from protocol §8, verbatim criteria)

| | T-1 push | T-2 order/seq | T-3 sessions | T-4 restart indep. | T-5 message-oriented | T-6 backpressure | T-7 separable-not-required |
|---|---|---|---|---|---|---|---|
| UDS | ? | ? | ? | ? | ? | ? | ? |
| TCP (loopback) | ? | ? | ? | ? | ? | ? | ? |
| stdio | ? | ? | ? | ? | ? | ? | ? |

Known asymmetries the record must engage honestly (not a pre-decision; these are the
questions, surfaced): stdio couples process lifecycles (T-4 pressure: whose child is
the daemon? a vendor restart tears the pipe); UDS vs TCP differ mainly in namespace
(filesystem path vs port), permissions, and Windows/portability posture; both are
byte streams, so T-5/T-6 land on the *framing* layer above either. The framing spec
therefore does most of T-2/T-5/T-6's work regardless of wire — the record must say
which properties come from the wire and which from the framing.

## Decision: fixture home is a top-level `seam/vectors/`

- **Rationale**: consumed by both languages' builds; neither `mind/` nor `mod/` exists
  yet and the vectors must not live inside either (the contract outlives both).
- **Alternatives considered**: `specs/007-*/fixtures/` — wrong lifetime (specs are
  per-feature paper trail; vectors are living contract); `docs/` — not
  test-consumable convention.

## Decision: vector fixture form is JSON-with-expected-bytes

Each vector pins (a) the decoded semantic form and (b) the exact wire bytes (hex or
base64 alongside, as the framing spec dictates once framing is chosen). Round-trip
equality is defined by the framing spec (byte-exact preferred; a declared canonical
equivalence only if the chosen serialization makes byte-exactness unreasonable — the
record must justify any such loosening).

- **Rationale**: spec edge case — "vectors pin bytes AND the decoded form, not bytes
  alone"; two implementations agreeing on bytes but not meaning is the failure mode.

## Decision: harness shapes

- Go: `seam/go-roundtrip/` with its own minimal `go.mod` + one test file. M1 later
  starts the real daemon module; this one is throwaway-grade proof.
- Java: `seam/java-roundtrip/` single-file program runnable with plain `java`
  (JDK ≥ 17 assumed present for Fabric work anyway); no Gradle — V1 owns the build
  scaffold.
- **Alternatives considered**: shared Gradle now — rejected, preempts V1 and violates
  FR-007's nothing-else-is-built.

## Decision: the framing spec is `docs/design/seam-wire-v0.md`

- **Rationale**: house convention (peer of body-protocol-v0.md); named as the spec-002
  successor by the card; versioned v0 to match the protocol's own versioning story.
