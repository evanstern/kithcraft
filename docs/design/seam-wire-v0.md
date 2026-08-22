# Seam wire v0 — the framing and serialization of the body protocol

**Status:** draft contract, v0. Spec `specs/007-seam-transport/`, board TASK-0007.
**Authority:** `docs/design/body-protocol-v0.md` defines the message shapes and left the wire an
open question (Q-1, §8/§11). `backlog/decisions/decision-0004` closes it. This document is the
spec-002 successor: it defines the **wire form** of every shape the protocol already binds.
Wiki grounding: [[body-protocol-seam]].

**What this binds.** How the mind daemon and a body vendor connect, how a protocol message
becomes bytes, how those bytes are delimited on a stream, and what ordering, backpressure,
versioning, and equality mean at that layer.

**What this does not bind.** Any message semantics. Every shape, field, enum, and MUST in
`body-protocol-v0.md` is unchanged and unrestated here. Where this document appears to say
something about meaning, it is citing; where the protocol is silent on a *wire* consequence,
this document decides it and says so.

**How to read.** `MUST` / `MUST NOT` / `SHOULD` / `MAY` are load-bearing where they appear.
Unmarked prose is rationale. Section numbering follows `specs/007-seam-transport/data-model.md`.
Verified platform facts are cited to `specs/007-seam-transport/research.md` rather than
re-asserted here.

---

## 1. The wire and its connection story

### 1.1 The socket

The seam is a **Unix domain socket** — `AF_UNIX`, `SOCK_STREAM` — at a filesystem path
(decision-0004).

- The path MUST come from configuration on both sides, with a default under the daemon's own
  state directory. It MUST NOT be derived from a working directory.
- The path MUST be at most **103 bytes** of UTF-8. `sun_path` is a 104-byte field on the demo
  host and ~108 on Linux, and the last byte is the terminator (verified, research.md). Both
  sides MUST validate the configured path's length at startup and **fail loudly**; neither may
  truncate, and neither may discover the problem at connect time.
- `SOCK_SEQPACKET` is not available on the demo host (`EPROTONOSUPPORT`, verified in
  research.md). Message boundaries are therefore §2's work, not the wire's — this is the whole
  reason this document has a framing section at all.
- Access control is filesystem permission on the socket's containing directory, and it is the
  listener's obligation (decision-0004, M1). It is not restated as a wire rule because it is
  not carried on the wire.

Go reaches this with `net.Listen`/`net.Dial` and network `"unix"`; the JVM reaches it with
`java.nio.channels` and `StandardProtocolFamily.UNIX` — not `java.net.Socket`, which has no
UNIX family (both verified, research.md). Neither side takes a dependency.

### 1.2 Who listens, who dials

**The mind daemon listens. The body vendor dials.** The vendor never accepts a connection and
opens no socket of its own (decision-0004; this is load-bearing for SI-1, not a preference).

The listener accepts in a loop for the daemon's whole life. A vendor disconnect is normal; a
reconnect is a fresh accepted connection.

### 1.3 Connection, session, body — the hierarchy (T-3)

> **One connection per vendor. One session per connection. N bodies multiplex over that
> session, disambiguated by the envelope's `body`.**

This is the reading the protocol already implies, and the wire adopts it rather than inventing
an alternative:

- `seq` is monotonic per **`(session, body, direction)`** (§2.1). The `body` component is only
  meaningful if one session carries several bodies.
- §6.2 requires `salient_kinds` to be "identical for every body **in a session**" — a
  constraint that exists only because a session has more than one body in it.
- §2.1 states outright that a vendor MAY multiplex several bodies over one connection and that
  `body` disambiguates.

**Why not one connection per body.** It would make the connection the session boundary, leaving
`body` in the envelope redundant and the `(session, body)` seq tuple degenerate — the protocol
would have specified two fields it never needed. It also buys an isolation it cannot deliver:
the events that actually drop a stream here are process death and a framing desync, both of
which are whole-process events. Per-body connections would give the *appearance* of independent
failure while the real failure modes stayed global, at the cost of N accept loops, N codecs, and
N sets of buffers on both sides.

**Consequently:**

- The **session is the connection.** It begins when the vendor's dial succeeds and ends when the
  connection closes, in either direction, for any reason.
- `session_open` and `session_close` are **per-body attach and detach events** inside that
  session. This is why both carry a `body` in the envelope and why §6.2 speaks of every body in
  a session.
- The **first** `session_open` on a connection establishes the session id (vendor-issued, §2.1),
  the `time_unit`, and the `capabilities` manifest. Every subsequent `session_open` on that
  connection MUST carry the same `session`, the same `time_unit`, and a byte-identical
  `capabilities` object. A receiver MUST refuse a differing manifest and close the connection
  with `session_close` / `reason: "error"`.
- A session id MUST NOT be reused across connections.
- A connection carries exactly one session for its lifetime. There is no session re-open on a
  live connection; a new session means a new connection.

The manifest-identity rule above is a wire rule with an epistemic payoff: §12's finding L-7 —
a manifest that described the body's surroundings rather than the vendor — becomes
**mechanically checkable** at the decoder, since two `session_open`s on one connection that
disagree are exactly the symptom a world-derived manifest would produce.

**One vendor, several connections.** If a vendor re-dials before the daemon has reaped a dead
connection, the daemon may briefly hold two. **Last open wins, per body:** a body attaching in a
newer session detaches from any older one, and subsequent messages for that body on the older
connection are refused as stale. Refusing the new connection instead would strand a live vendor
behind a corpse the kernel has not yet reaped.

### 1.4 Opening and closing

- The vendor speaks first. Its first frame on a connection is a `session_open` (decision-0004).
  A connection on which the mind speaks first is malformed; the mind MUST NOT send anything
  before it has read a `session_open`.
- **No half-close.** Neither side shuts down one direction of the stream while keeping the
  other. Close means close. A half-close dance would create a state in which a session is
  half-alive, which nothing in the protocol describes.
- **Orderly shutdown** is the vendor sending `session_close` (`reason: "shutdown"`, §6.3) for
  each attached body, then closing the connection. After sending `session_close` for a body, a
  sender MUST NOT send further messages naming that body.
- **The mind sends `session_close` only to refuse** — an unsupported version (§7.1) or a
  handshake it cannot accept — with `reason: "error"`. Routine session ending is the vendor's.
  A refusing side sends its frame, then closes after a bounded linger so the peer can read it
  rather than seeing a bare EOF.
- **A dropped connection ends the session.** The listener treats every still-attached body as
  though `session_close` with `reason: "error"` had arrived (decision-0004, M1), keeps all
  durable memory (SI-5), and waits for the next connection.

### 1.5 Restart and reconnect (T-4)

Neither side restarts the other. The recovery unit is the session, not the message, and §6.3
already specifies rejoin — the wire adds no resume, no replay buffer, and no message-level ack.

**When the mind vanishes.** The vendor's next write fails or its read returns EOF. It MUST
close its end, abandon the session, and re-dial. It MUST NOT block the world's main thread
doing any of this (decision-0004, V1).

**When the vendor vanishes.** The listener's read returns EOF or an error. It closes the
connection, ends the session per §1.4, and keeps listening. The daemon never dials except for
the stale-path liveness probe of decision-0004, and never exits because a vendor went away.

**Re-dial policy.** Bounded exponential backoff, retried indefinitely — the daemon may be down
for an hour and T-4 means that is survivable. At most one dial attempt in flight at a time. A
refused dial is expected, not an error. The **shape** is normative; the **constants** are
configuration, per the same discipline that keeps cadences and horizons out of the protocol
(§2.2, §4.11). A reasonable default, stated so nobody invents a wilder one: 250 ms initial,
doubling, capped at 10 s, with a little jitter.

**Continuity rides the reconnect, and is matched by `body`.** The new connection's
`session_open` for a body carries `continuity.previous_session`,
`previous_close_world_time`, and `body_continuous` (§6.2). The mind MUST match the returning
body by its **`body` token**, which is stable across sessions (§2.3) — **not** by
`previous_session`. `previous_session` is a consistency check a mind MAY use and MUST NOT
require: a mind that restarted has no memory of the previous session id but does have durable
memory keyed by body, and a mind that refused the rejoin for want of a session id it lost would
have made a mind restart require a vendor restart, which is T-4 failing in the one direction
T-4 is written to prevent.

`seq` counters restart at 0 in the new session (§3.2). The gap is a gap: the vendor MUST NOT
backfill it (§6.3), and the mind MUST NOT ask — there is no message with which it could.

---

## 2. Framing and encoding (T-5)

### 2.1 The frame

Every protocol message crosses as exactly one frame:

```
+--------+--------+--------+--------+--------------------------------+
|             length (uint32, BE)   |  body: length bytes of UTF-8    |
+--------+--------+--------+--------+--------------------------------+
```

- **`length`** — unsigned 32-bit integer, **big-endian**, counting **only** the body bytes that
  follow. It does not include itself. `binary.BigEndian` in Go and a default-order
  `ByteBuffer` on the JVM are each a single call; nothing needs negotiating.
- **body** — the message, serialized as JSON (§2.4), encoded UTF-8, **no BOM**.
- No trailing delimiter, no padding, no alignment. The next frame's length word begins at the
  byte after the previous body's last byte.
- A `length` of `0` is malformed: an empty body is not a JSON value.

**Why a length prefix and not a delimiter.** Newline-delimited JSON was the alternative and it
is genuinely cheaper to read by hand. It loses on three counts, all of them about boundaries
being explicit rather than emergent:

1. **The boundary stops depending on the payload's escaping discipline.** With a delimiter, a
   frame's extent is a property of what the encoder did *inside* the JSON. With a prefix, it is
   stated outright.
2. **Size can be enforced before allocation.** A receiver reads four bytes, compares against the
   cap (§2.5), and refuses without ever buffering the body. A delimiter forces buffering to the
   cap before the question can even be asked.
3. **No line-terminator or charset ambiguity in either runtime.** The JVM's line-oriented
   readers treat `\r`, `\n`, and `\r\n` alike and decode a charset while doing it — a frame
   ending `}\r\n` would read cleanly and re-encode as `}\n`, a silent byte difference under §6's
   equality rule. A length prefix has no such reading.

### 2.2 The header carries a length and nothing else

**No timestamp, no message-type tag, no checksum, no flags, no version.** Each of those would
be wire-level state that can disagree with the message it wraps, which is §2.7's argument
against putting `direct` on the wire, one layer down:

- A **type tag** would duplicate the envelope's `message` field and could contradict it.
- A **checksum** re-solves what the stream already guarantees (§3.1) and would invite treating a
  mismatch as recoverable, which it is not.
- A **version** in the header cannot be read before the frame format is already assumed; §5
  handles versioning where it belongs.
- A **wall-clock timestamp** is the one that matters most: `world_time` is the protocol's only
  clock (§2.2) and it is vendor-declared in a vendor-declared unit. A header timestamp would put
  a second, real-time clock on the seam that the mind could difference against `world_time` —
  and a mind holding a real-time metric alongside coarse distance bands is a mind one step from
  the spatial arithmetic AR-4 forbids. See §7, P-6.

### 2.3 Reading a frame

Both sides MUST read exactly, in two steps, looping until each step is satisfied:

1. Read **4 bytes**. A short read is not a frame boundary — continue reading until 4 bytes are
   in hand or the stream ends.
2. Read exactly `length` bytes. Again, loop; a short read is normal on a stream and means
   nothing.

A stream that ends **between** frames is an orderly disconnect. A stream that ends **inside**
one is a truncated frame: the receiver MUST discard it (never parse a partial body) and treat
the connection as failed. There is no resynchronization — a stream whose framing is in doubt is
a stream whose every subsequent boundary is in doubt — and there is no need for one, because the
session boundary is the recovery unit (§1.5).

### 2.4 The body: canonical JSON

The body is a JSON text per RFC 8259 carrying exactly one message envelope (§2.1 of the
protocol). One frame, one message.

**The rule is asymmetric, and deliberately so:**

> **Senders MUST emit the canonical form below. Receivers MUST accept any conforming JSON
> encoding of the same value.**

The receiving half is just V-1's robustness applied to serialization: member order is not
semantic in JSON, so a peer that emits it differently interoperates fine and nothing breaks.
The sending half is what makes the golden vectors *the wire* rather than a test-only dialect —
without it, a fixture's byte column would pin something no implementation is obliged to produce.
Determinism costs a sender nothing and buys reproducible captures that diff across a session.

**Canonical form:**

| # | Rule |
|---|---|
| **C-1** | UTF-8, no BOM. |
| **C-2** | No insignificant whitespace anywhere: no spaces after `:` or `,`, no newlines, no indentation. |
| **C-3** | Object members in **ascending order of the key's UTF-8 bytes**. Every key in v0 is ASCII lowercase with underscores, so byte order, code-point order, and ASCII order coincide; the rule is stated by bytes so it stays unambiguous if a later version adds a non-ASCII key. |
| **C-4** | **No duplicate keys.** A sender MUST NOT emit one; a receiver encountering one MUST treat the message as malformed (V-5) rather than applying a last-wins rule. |
| **C-5** | **Array order is semantic and is never touched by the codec.** Where the protocol requires an order — §4.3's `present` is "in a stable sorted order" — that is the vendor's obligation at emission, not the serializer's. A codec that sorted arrays would silently satisfy a rule the vendor had broken. |
| **C-6** | **Numbers are integers**, emitted with no fraction, no exponent, no leading `+`, no leading zeros, and `-` only for negatives. Every number in v0 is an integer (`seq`, `world_time`, `observed_at`, `received_at`, `hops`, `count`, `not_after`). A value MUST be representable as a signed 64-bit integer; a receiver MUST refuse one outside that range as malformed rather than silently losing precision to a double. A future version that introduces a non-integer number MUST pin its own form (§5). |
| **C-7** | **Strings** escape exactly what RFC 8259 requires and nothing more: `"` → `\"`, `\` → `\\`, and U+0000–U+001F via the two-character escapes `\b` `\t` `\n` `\f` `\r` where they exist, otherwise `\u00xx` with **lowercase** hex. `/` is **not** escaped. Non-ASCII characters are emitted as **literal UTF-8**, never as `\u` escapes and never as surrogate pairs — a non-BMP character is a four-byte UTF-8 sequence on the wire. |
| **C-8** | **A lone surrogate MUST NOT be emitted**, and a receiver MUST refuse a message containing one as malformed. UTF-16-native runtimes can hold an unpaired surrogate in a string that no valid UTF-8 can represent; a codec that transcoded it to a replacement character would silently alter an utterance or a descriptor. |
| **C-9** | **A required field is always present** — carrying `null` where its declared type permits (`source`, `place`, `observed_at`, …). **An optional field that is not carried is omitted, never emitted as `null`.** The protocol's own tables distinguish required-and-nullable (`✔` with `\| null`) from optional (`—`); the wire keeps the distinction so V-5's "a missing required field is malformed" stays a decidable test. Receivers, per the asymmetry above, MUST tolerate an explicit `null` for an optional field. |
| **C-10** | Literals are lowercase `true` / `false` / `null` (RFC 8259 requires this; stated so no implementation "helps"). |

### 2.5 Size posture

**Maximum frame body: 1 MiB (1 048 576 bytes).** Both sides MUST enforce it. A receiver whose
length word exceeds the cap MUST close the connection without reading the body; a sender MUST
NOT emit such a frame.

The cap bounds a **corrupt length word**, not content. The largest plausible message is an
`observation` over a vendor's `salient_kinds` with a long `present` list, which is orders of
magnitude below this — and §4.3 already requires the scanned vocabulary to be closed and small,
so a message approaching the cap is a scope bug in the vendor, not congestion. It is therefore
**not** a shedding input: an oversize frame is refused, never trimmed, and never partially sent
(§4.11 forbids a partial `observation`, and §4.2 forbids a partial anything).

The cap is a wire safety limit and MUST NOT be read as a percept budget. §4.11 keeps cadence,
per-beat caps, and staggering off the protocol entirely; nothing here relaxes that.

---

## 3. Ordering and `seq` (T-2)

### 3.1 What the stream gives

Within one connection, a `SOCK_STREAM` socket delivers bytes **in order, exactly once, without
loss**; a stream that cannot do so is torn down rather than degraded (research.md, T-2). Framing
carries that up intact: **frames arrive in the order they were written, whole, exactly once, for
as long as the connection lives.**

So the wire needs no message-level ack, no retransmit layer, and no reordering buffer, and
neither side may build one.

### 3.2 What `seq` adds

Nothing about ordering — the stream already settled that. On this wire `seq` is three things: a
**gap detector** (§3.3), a correlation handle for logs and captures, and the field that would
let a future non-stream transport reorder without a protocol change.

Wire rules for the counter:

- One counter per **`(session, body, direction)`**, per §2.1.
- The **vendor→mind** counter for a body starts at **0** with that body's `session_open`
  (matching §6.2's example).
- The **mind→vendor** counter for a body starts at **0** with the mind's first message naming
  that body on that connection.
- Every message sent increments its counter by exactly 1.
- Counters restart at 0 in a new session (§1.5). They are per-session and carry no meaning
  across one.
- A received `seq` that is **not greater than** the last for its `(session, body, direction)`
  is malformed: the receiver refuses that message and records it. It does not close the
  connection — a single sender bug should not end N bodies' sessions — unless it can no longer
  maintain frame sync (§2.3), which is a different failure.

### 3.3 What a gap means, and what the receiver does

> **`seq` is assigned when a message is admitted to the outbound queue, before any shedding
> decision. A shed percept's number is spent and never reused.**

That one rule is what makes §2.1's "gaps mean loss" true on a wire that cannot lose anything.
Because the stream neither reorders nor drops, **a gap has exactly two possible meanings**:

1. The vendor deliberately shed that many `background` percepts (§4.11, and §4 below); or
2. the session broke — which is not a gap in a run of `seq` at all, but a counter restarting
   from 0 in a new session.

Shedding therefore becomes observable **without a new message shape**: the count of what was
shed is the size of the gap. No `shed_count` field and no diagnostic message is added, because
adding one would be a protocol change and this document changes no shapes.

On receiving a gap, a receiver:

- **MUST NOT** request retransmission. There is no such message, and inventing one would be
  precisely T-1's forbidden "get percepts since seq N" (§7, P-4).
- **MUST NOT** treat it as an error, close the connection, or refuse the message that follows.
- **MAY** record it. A gap in the vendor→mind direction is a count of texture the mind was never
  going to act on, by design.
- A gap in the **mind→vendor** direction is a sender bug: minds have nothing to shed (intents
  carry no urgency and §4.11 licenses shedding only for `background` percepts). A vendor
  observing one records it and continues.

### 3.4 Dedup, and the one job `percept_id` has

Within a session, `percept_id` has **no** dedup job: the stream cannot duplicate. Neither side
may build a general at-least-once dedup cache (decision-0004).

Its one job spans a **reconnect**. The wire requires no retransmission — a percept that failed
to write before a drop is simply lost, and the gap is a gap (§1.5). But **if** a vendor chooses
to re-emit such a percept in the new session, it **MUST reuse the original `percept_id`**; a
vendor composing a fresh percept about the same fact MUST NOT. Because `percept_id` is unique
within a session (§4.1), an id reappearing **across** a session boundary is unambiguously a
retransmission, and that is the whole signal.

A mind's dedup window therefore needs only to span a reconnect, scoped per body. Its size is
mind configuration — never a protocol constant, per the same rule that keeps horizons off the
wire (§2.2).

---

## 4. Backpressure (T-6)

### 4.1 The only signal the wire gives

A stalled socket is **undifferentiated**: a blocking write does not return, and a non-blocking
one reports `EWOULDBLOCK`/`EAGAIN` — on the JVM, a `write()` that consumes fewer bytes than
offered, or none. The socket says *the peer is not draining*. It does not and cannot say
anything about urgency, which is exactly what §4.11 requires it to discriminate on
(research.md, T-6).

Every part of §4.11 is therefore this layer's work.

### 4.2 One writer, whole frames

> **Exactly one thread or goroutine writes to a connection, for that connection's whole life.
> A frame is serialized to a complete byte buffer — length word and body together — before any
> of it is written, and the writer loops until the entire buffer is written or the connection
> fails.**

Concurrent writers on a byte stream interleave partial frames. That is precisely the split the
protocol forbids, and no wire feature prevents it — single-writer discipline does.

Note what this makes unnecessary: **the never-split-an-`observation` rule is not implemented as
a rule about observations.** The framing layer never splits *any* frame, which subsumes §4.11's
requirement. An observation-specific rule would be a rule that could be forgotten for one code
path; a universal one cannot.

**If a write fails mid-frame the sender MUST close the connection.** A receiver cannot
distinguish a truncated frame from a short one, so there is nothing to resynchronize to
(§2.3), and the reconnect path (§1.5) is the recovery.

**The producing path MUST NOT write to the socket.** Percepts are produced by the world; the
vendor's producer hands a complete message to a per-body outbound queue and returns. The writer
drains the queue. Writing from the producing path would put the socket's undifferentiated stall
directly into the world's main thread, which decision-0004 forbids and which no amount of
policy above it could then repair.

### 4.3 The queue, and where shedding happens

> **Shedding happens at enqueue. Never at write.**

By the time bytes are moving, the message is committed; the decision must be made while it is
still a message with an `urgency` on it.

- Each body has a **bounded** outbound queue. Bound in messages, and size it so §4.11's shedding
  is routine and the §4.3-below escalation is genuinely rare.
- **Only `background` percepts are shed.** A vendor MAY drop them when the queue is at its bound
  and the writer is not draining, and SHOULD prefer dropping the **oldest** background entries
  first — they are the stalest and therefore the least worth delivering. Which ones is vendor
  policy; *only background* is not.
- **`urgent` is never shed** (§4.11).
- **`notable` is not shed here either.** §4.11 licenses shedding `background` and forbids
  shedding `urgent`, leaving `notable` unaddressed; this wire declines to invent a licence the
  protocol did not grant. A queue that cannot admit a `notable`-or-higher message after shedding
  every `background` entry it holds is **failing, not congested** — and the sender MUST close
  the connection, ending the session. The three available responses are to drop something the
  protocol protects, to block the world, or to end the session; only the third neither lies to
  the mind nor stalls the world, and §6.3's continuity path exists precisely to recover from it.
- **Messages with no `urgency` are never shed.** `session_open`, `session_close`, `intent`,
  `intent_ack`, and `cancel` carry no urgency field; only percepts do. A non-percept that cannot
  be enqueued is the failing case above.
- An **`observation` is one queue entry and one frame.** There is no representation in which
  half of one exists — at the queue layer it is shed whole or not at all, and at the frame layer
  §4.2 guarantees the rest. §4.11's requirement is structurally satisfied at both.
- **A shed message's `seq` is already spent** (§3.3). That is how the mind learns anything was
  shed.
- **Minds do not shed.** A mind MUST NOT drop a queued message: it would open a `seq` gap in a
  direction where a gap has no licensed meaning (§3.3), and §5.4's one-accepted-intent-one-result
  contract counts only what arrived.

### 4.4 The read side

**A receiver MUST drain the socket continuously and hand messages off through its own queue. It
MUST NOT do work of unbounded duration on the read path.**

This costs one sentence and prevents the failure this whole section is about. The mind's
per-percept work is a language model call measured in seconds. A mind that decodes and *thinks*
in the same loop stops draining for the duration, the vendor's socket buffer fills, and the
vendor sheds — so the mind's thinking time is converted into the vendor discarding percepts the
mind wanted. The queue that decouples them is the fix, and it belongs on the read side where the
slowness is.

---

## 5. Versioning on the wire

### 5.1 The frame layer is not separately versioned

The frame format has no version field and gets none. It cannot: nothing in a frame can be read
until the frame format is already assumed, so a header version would be unverifiable
self-assertion. The frame layer is versioned by the protocol **MAJOR** carried inside the first
frame — a change to the frame format is necessarily a MAJOR bump (§7.3), and a receiver MUST
NOT attempt to sniff a frame format.

### 5.2 Negotiation is the handshake, and it fails closed

§7.1 already specifies the posture; the wire pins the mechanics:

- The vendor's first frame is `session_open`, carrying `protocol` (§1.4). The mind MUST read and
  check `protocol` **before acting on anything else in the message**.
- On an unsupported version the mind replies `session_close` with `reason: "error"` and
  `detail: "unsupported_version"` (§7.1's exact string), then closes after a bounded linger
  (§1.4) so the vendor learns why rather than seeing a bare EOF.
- **That reply's envelope**, which the protocol does not spell out because it had no wire to
  spell it out on: the mind **echoes** the `session` and `body` from the frame it refused, uses
  `seq: 0` (its first message in that direction), and echoes the `world_time` it just received.
  A mind has no clock of its own — `world_time` is vendor-declared in a vendor-declared unit
  (§2.2) — so echoing is the only honest value available to it.
- **After the handshake, receivers compare only MAJOR.** A message whose MAJOR differs from the
  session's is malformed. A differing MINOR is accepted; that is exactly what V-1 and V-2 are
  for.
- There is no downgrade negotiation, no capability probe, and no re-handshake on a live
  connection.

### 5.3 Additive changes (MINOR) — and why existing vectors survive them

A new optional field, percept type, origin, verb, or open-vocabulary token (§7.2) reaches the
wire as a new member slotting into its canonical sorted position (C-3). Nothing is removed, so:

> **Vectors are append-only within a MAJOR.**

- A MINOR bump **adds** vectors for the new shapes. It MUST NOT edit an existing vector to carry
  the new field, and MUST NOT delete one.
- Every existing vector therefore stays valid **byte-for-byte** across a MINOR bump. This is the
  spec's stated edge case ("how vectors are extended without invalidating existing ones")
  answered by a rule rather than by care.
- The retained old vectors then do real work: they *are* the older-sender fixtures, so the
  vector set proves V-1 (receivers ignore unknown fields) and V-6 (an unknown origin classifies
  secondhand) mechanically, using nothing but the bytes an earlier version already pinned.

A **MAJOR** bump gets its own vector set. The previous major's vectors are retained beside it,
not deleted — they document what the old wire was, which is the only record of it that exists.

---

## 6. Round-trip equality

> **Byte-exact, over the whole frame — length word included.**

For every non-error vector, both harnesses assert two things, and both are required:

1. **Bytes.** Decode the fixture's frame, re-encode it in canonical form (§2.4), and compare the
   resulting bytes to the fixture's — **including the four-byte length prefix**, so a wrong
   length is caught by the same assertion rather than by an implementation's own arithmetic.
2. **Meaning.** The decoder's in-memory value equals the fixture's declared decoded form,
   compared **structurally** — objects as unordered maps, arrays in order, numbers by value.

The second assertion is not redundant with the first. Research recorded the failure mode
plainly: two implementations agreeing on bytes but not on meaning. A decoder that mangled a
field and an encoder that mangled it back would pass a byte comparison alone.

**Why byte-exact is adoptable rather than a burden.** Canonical form (§2.4) removes every degree
of freedom that would otherwise make byte-exactness unreasonable — member order, whitespace,
number spelling, escaping. What remains is fully determined, so there is nothing left to loosen
*to*. A canonical-form equivalence would have been the answer if senders had been left free to
order members as they liked; because §2.4 constrains senders instead, the stronger rule is
available for the same effort. And the asymmetry of §2.4 keeps the cost off production: a
sender that drifts from canonical form still interoperates, because every receiver accepts any
conforming JSON — it fails its vectors, which is where a drift belongs.

**Error vectors are exempt from the round-trip and pin a behavior instead** (per
`contracts/vectors.md`): a decoder's refusal or acceptance is the assertion, not a byte
comparison.

**Fixture byte representation: lowercase hexadecimal, unseparated, covering the entire frame**
including the length prefix. Hex over base64 because it is byte-aligned — two characters is one
byte, so the offset of a mismatch is computable by hand from a failing diff, which base64 does
not permit. The fixture also carries the decoded JSON, per `data-model.md`; the file layout is
Phase 3's.

---

## 7. Leak audit

**Performed** over this document, applying §12's six passes unchanged. Two things were hunted:
engine-native concepts reaching the wire (AR-1…AR-6), and **query-shaped affordances** — any way
a mind could pull world state rather than receive it (SI-1). §12's closing lesson set the
priority: every genuine leak in the protocol was query-shaped, not engine-shaped, so P-4 got the
most attention here.

### 7.1 Findings

| # | Finding | Pass | Verdict |
|---|---|---|---|
| **W-1** | The socket path is operator-visible configuration and could be spelled with a vendor's name. | P-1 | **No change needed.** The path is never carried in any message and cannot reach a mind's state; it is not seam data. Recorded because it is the only vendor-nameable string this layer has, and a future reader may expect a rule where none is required. |
| **W-2** | The frame header — the only new structure this document adds. | P-1, P-2 | **Verified clean, and deliberately bare.** It carries a byte count and nothing else. §2.2 states the reasoning; a type tag, checksum, flag word, or version would each be wire-level state able to disagree with the message it wraps. |
| **W-3** | A wall-clock timestamp in the frame header. | P-6 | **Considered and refused** (§2.2). `world_time` is the protocol's only clock and is vendor-declared in a vendor-declared unit (AR-5). A real-time header stamp would give a mind a metric to difference against coarse bands, one step from the spatial arithmetic AR-4 forbids. Wall-clock *arrival* time is unavoidable on any transport — the in-process fake vendor has it too — but the wire declines to hand it over as data. |
| **W-4** | A frame-level acknowledgement or flow-control credit message. | P-4 | **Considered and refused.** It would be a reply-shaped message beyond the protocol's own `intent_ack` — the obligation T-1 places on framing (research.md) — and would put a synchronous channel back to the vendor behind every percept, which is the shape L-6 was written about. Backpressure is a kernel buffer and a queue (§4), never a message. |
| **W-5** | A "resume from `seq` N" or retransmission-request message. | P-4 | **Considered and refused.** This is verbatim T-1's forbidden "get percepts since seq N": a pull channel for percepts, and a mind able to demand history it did not receive. Recorded with its reason so a later implementer meets the argument rather than rediscovering the idea as an optimization. §3.4 gives the whole recovery story instead. |
| **W-6** | A keepalive / ping-pong frame. | P-4 | **Not added — and out of this document's jurisdiction.** A dead peer EOFs immediately; a hung-but-alive peer surfaces through the write path (§4.1). A ping would be a **new message shape**, which is a protocol addition under §7.2 and not a framing decision. Recorded so the absence reads as a ruling rather than an oversight. |
| **W-7** | The listener's stale-path liveness probe (dial before unlink, decision-0004). | P-4 | **Verified clean.** It is the only place either side dials to learn something, and what it learns is whether *another daemon* holds the path — the mind's own liveness, never world state. It happens before any session exists. |
| **W-8** | The `session_open` manifest under the multi-body session model (§1.3). | P-5 | **Strengthened, not weakened.** Requiring a byte-identical `capabilities` object across every `session_open` on a connection makes §12's L-7 — a manifest describing the body's surroundings rather than the vendor — mechanically checkable at the decoder, since two disagreeing manifests are exactly the symptom a world-derived manifest produces. |
| **W-9** | Free-text fields added by this layer. | P-3 | **None.** The wire adds no field of any kind to any message. The one string literal it pins, `"unsupported_version"`, is quoted from §7.1. |
| **W-10** | The 1 MiB frame cap as an inferable world quantity. | P-2, P-6 | **No change needed.** It is a constant of the transport, identical for every message and every body, and carries no information about the world. §2.5 states it is not a percept budget and MUST NOT be used as a shedding input, closing the one path by which it could have become one. |
| **W-11** | `seq` gaps as a channel (§3.3). | P-2, P-4 | **Verified clean.** A gap conveys a count of shed `background` percepts and nothing else — no identity, no place, no kind. It is a diagnostic derived from an existing field, not a new one, and §3.3 forbids the receiver from acting on it beyond recording it. |

### 7.2 Result

Eleven findings: **four deliberate refusals** (W-3, W-4, W-5, W-6), **one strengthening** (W-8),
**six verified clean** (W-1, W-2, W-7, W-9, W-10, W-11), and **no fixes required** — this
document adds no message, no field, and no enum value, so there was nothing to repair.

**No engine-native type, identifier, coordinate convention, unit, or scale appears anywhere in
this document's normative text.** The frame carries a byte count and UTF-8; everything above it
is the protocol's, unchanged.

**The pattern worth carrying.** Every refusal above was a *convenience*: an ack that would have
made flow control legible, a resume that would have made a reconnect lossless, a ping that would
have made liveness explicit. Each is the shape §12 warned about — an affordance that arrives
looking like a kindness to implementers and leaves with SI-1 or T-1 in its pocket. The framing
layer is the easiest place in the system to add one, because at this depth it does not look like
a protocol change at all. **Re-run all six passes whenever a frame-layer field is proposed** —
and treat "it is only framing" as the reason to look harder, not as a reason to skip.

---

**Phase 2 complete.** Q-1 is closed by decision-0004 plus this document; every shape in
`docs/design/body-protocol-v0.md` now has exactly one wire representation. The golden vectors
(`seam/vectors/`) pin it in bytes.
