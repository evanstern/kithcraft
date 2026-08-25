# Golden vectors — the seam wire, pinned in bytes

These files are the executable form of `docs/design/seam-wire-v0.md`. Each one pins a single
protocol message three ways at once: what it **means**, what it **weighs on the wire**, and what
a receiver **must do** with it. Two harnesses that share no code — `../go-roundtrip` and
`../java-roundtrip` — check every file against all three, and it is that agreement, not the
prose, that makes the wire real.

The set is **closed**. `specs/007-seam-transport/contracts/vectors.md` lists exactly seventeen
vectors: nine percept types, three intent shapes, two session-lifecycle messages, and three
declared error/edge cases. Both harnesses fail on a missing file *and* on an extra one, so the
census cannot drift in either direction.

## The fixture format

One JSON file per vector, named for the vector it carries (`percept_sighting.json` holds
`percept_sighting` — both harnesses check that the name and the filename agree). JSON, and not
a bespoke format, because the harnesses must read it in two languages without sharing a line of
code, and because the whole point of the file is to sit next to the JSON message it pins.

| field | required | meaning |
|---|---|---|
| `name` | ✔ | Stable identifier. Must equal the filename's stem. |
| `direction` | ✔ | `vendor_to_mind` or `mind_to_vendor` — which way this message travels. |
| `expect` | ✔ | What a receiver must do: `roundtrip`, `decode_ok`, or `refuse` (below). |
| `refusal` | on `refuse` | The exact refusal string a validating receiver must produce. |
| `decoded` | ✔ | The message itself, as ordinary readable JSON per `body-protocol-v0.md`. |
| `frame_hex` | ✔ | The canonical wire bytes: lowercase unseparated hex of the **whole frame**, four-byte big-endian length prefix included. |
| `note` | — | Prose, for the vectors whose point is not obvious from their shape. |

**`decoded` is authored for humans; `frame_hex` is derived.** The indentation, member order, and
line breaks in `decoded` carry no meaning — the harnesses compare it *structurally* (objects as
unordered maps, arrays in order, numbers by value) and it is deliberately formatted to be read
and reviewed rather than to look like the wire. The wire form is `frame_hex` and only
`frame_hex`; it is what canonical encoding (§2.4, C-1..C-10) produces from `decoded`, and
regenerating it is a mechanical act (see below), never an editorial one.

**Hex over base64** because it is byte-aligned: two characters is one byte, so a failing diff
reports the offset of the first bad byte and you can find it by counting. Both harnesses print
exactly that on a mismatch.

## What each `expect` value obliges

- **`roundtrip`** — the full contract of `seam-wire-v0.md` §6, and what fourteen of the
  seventeen vectors carry. Decode `frame_hex`; the decoded value must match `decoded`
  structurally; re-encoding it canonically must reproduce `frame_hex` **byte for byte, length
  word included**. Both assertions are required and neither implies the other: a decoder that
  mangled a field and an encoder that mangled it back would pass the byte check alone, which is
  the exact failure mode the research pass recorded.

- **`decode_ok`** — decoding must **succeed**, and the message must pass required-field
  validation, but the vector makes no claim about what the value means. `err_unknown_origin` is
  the only one: an origin outside v0's vocabulary is what a MINOR-newer sender looks like from
  here, so refusing it at the codec would make every additive change a breaking one. Classifying
  it — EH-2b's *unrecognized or absent origin is secondhand, never direct* — is the mind's job,
  above the wire. (These vectors round-trip too; the pinned behavior is the acceptance.)

- **`refuse`** — validation must **reject** the message, with the exact string in `refusal`.
  `err_missing_provenance` is the only one. Its frame is perfectly well-formed and canonical:
  the refusal is a message-validity ruling, not a framing error, and the distinction matters
  because a decoder that supplied a default provenance would be minting first-hand-looking
  evidence out of a vendor bug.

Note that `intent_ack_refused` is an `roundtrip` vector despite living in the error group. It
carries a *refusal* (`unknown_target`) but is not itself malformed — it is listed as an edge
vector because that particular refusal is the one §5.3 most warns about: `unknown_target` means
the vendor never issued the token, never that the referent stopped existing. That discovery
belongs to `act_result` / `target_gone` alone, and an ack that answered it would be a synchronous
existence oracle bolted to the act surface.

## Regenerating `frame_hex`

The Go harness can rewrite the byte column from the decoded form:

```sh
cd ../go-roundtrip && go test ./... -update
```

It rewrites only the `frame_hex` line, leaving the hand-authored `decoded` byte-identical. This
is an **authoring** convenience, never a way to make a failure go away: the check that means
anything is the Java harness, which has no `-update` mode and only ever verifies. If the two
disagree, one of the implementations is wrong — regenerating from the one that happens to be
running would simply pick a winner without finding out which.

## Extending the set

Vectors are **append-only within a protocol MAJOR** (§5.3 of the framing spec). A MINOR bump adds
files for the new shapes; it must not edit an existing vector to carry a newly added field, and
must not delete one. That rule is what keeps every existing vector valid byte-for-byte across an
additive change — and it is what turns the retained old vectors into the older-sender fixtures
that prove receivers tolerate unknown fields, using nothing but bytes an earlier version already
pinned.

A MAJOR bump gets its own set, kept beside this one rather than replacing it. The old vectors are
the only surviving record of what the old wire was.

## Content discipline

Every value here is world-flavored and engine-free, per the protocol's §3 abstraction rule:
opaque tokens (`b-…`, `pl-…`, `th-…`, `k:…`), meaning carried as `roles` plus prose
`descriptor`, space as place tokens and coarse bands, time as integers in the session's declared
unit. No engine-native type name, registry identifier, coordinate, or native scale appears in
any vector — and a vector that introduced one would be pinning the leak into the wire, where
every later implementation would inherit it.
