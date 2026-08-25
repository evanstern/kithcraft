# Phase 1 data model — seam transport artifacts

Three entities, all artifacts (no runtime state in this task).

## Decision record

Backlog decision (via `backlog decision create`, CLI-owned file under
`backlog/decisions/`). Fields the record must carry:

- Candidates: UDS, TCP (loopback), stdio — the closed set.
- The filled T-matrix: T-1..T-7 × 3 candidates, each cell an explicit answer with the
  property attributed to the wire or to the framing layer.
- The choice, with rationale; per rejected wire, the stated reason.
- Status: proposed → operator ratifies at PR review (merge is the ratification act).

## Framing spec — `docs/design/seam-wire-v0.md`

Sections the document must contain (the content is the implementer's design work):

1. The chosen wire and its connection story (who listens, who dials, T-3 sessions,
   T-4 restart/reconnect behavior).
2. Message delimiting/framing over the byte stream (T-5), and encoding of the
   protocol's JSON message shapes.
3. Ordering and `seq` semantics on this wire (T-2); dedup interaction (`percept_id`).
4. Backpressure: how `background` shedding (T-6) is realized; the never-split-an-
   `observation` rule at the framing layer.
5. Versioning on the wire: how protocol §7's additive/breaking story appears in
   framing (and in vectors).
6. Round-trip equality rule: byte-exact, or a declared canonical equivalence with
   justification.
7. Leak audit statement (§12 discipline applied to the wire layer).

## Golden vector

One fixture per protocol message shape (see contracts/vectors.md for the closed list).
Each vector pins:

- `name` — stable identifier (`percept_sighting`, `intent_go_to`, `session_open`, …).
- Decoded semantic form — the JSON message per body-protocol-v0.
- Exact wire bytes — encoded per the framing spec (hex/base64 representation chosen
  there).
- Direction(s) it must round-trip (encode, decode, or both).

Validation rules: complete per protocol requirements (a vector missing a required
field is itself malformed and belongs only in a declared error-vector set); no
engine-native token anywhere; language-neutral (no Go/Java-specific typing hints).

## Relationships

decision record → names the wire the framing spec assumes;
framing spec → defines the byte form every vector pins;
vectors → consumed by both round-trip harnesses (and later by M1/V1's real
implementations, unchanged).
