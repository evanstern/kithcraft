# Vector-set contract — what must exist in `seam/vectors/`

The closed list, derived from body-protocol-v0. The set is complete when every row
below has a fixture; extra vectors beyond declared error/edge vectors are scope creep.

## Percept types (§4.2–§4.10) — one vector each

| vector | protocol section |
|---|---|
| `percept_sighting` | §4.2 |
| `percept_observation` | §4.3 (must include its `vocabulary`) |
| `percept_sound` | §4.4 (optional percept type, still a defined shape — pinned) |
| `percept_speech` | §4.5 |
| `percept_told_fact` | §4.6 |
| `percept_text` | §4.7 (carries `origin: read` — V4's blueprint channel) |
| `percept_act_result` | §4.8 / §5.4 |
| `percept_self_state` | §4.9 |
| `percept_change_report` | §4.10 |

## Intent shapes (§5) — one vector each

| vector | protocol section |
|---|---|
| `intent` | §5.2 (a `go_to` with `reason` and `supersedes` exercised) |
| `intent_ack` | §5.3 (one accepted; refusal covered in error vectors) |
| `cancel` | §5.7 |

## Session lifecycle (§6)

| vector | protocol section |
|---|---|
| `session_open` | §6.2 — the full handshake with manifest, `time_unit`, continuity |
| `session_close` | §6.2/§6.3 counterpart (in-scope: the handshake is a pair) |

## Declared error/edge vectors (bounded — these and no more)

| vector | proves |
|---|---|
| `err_missing_provenance` | a percept missing `provenance` is malformed (EH-2a / V-5) — decoders must refuse, never default |
| `err_unknown_origin` | an unrecognized `origin` decodes (V-2 fallback) — classification to secondhand is the *mind's* job, not the decoder's; the vector pins decode-success |
| `intent_ack_refused` | the `unknown_target` refusal shape (§5.3/§5.6) |

## Round-trip obligations

- Go harness: for every non-error vector, decode fixture → re-encode → equality per
  the framing spec's rule; for error vectors, decode behaves as pinned.
- Java harness: same obligations, same fixtures, no shared code with the Go harness.
