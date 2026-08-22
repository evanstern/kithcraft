# Phase 0 research — seam transport

Plan-level unknowns resolved here. The wire choice itself was NOT resolved here — it is
the task's deliverable. This file bounded the decision space and settled the plumbing
questions so the implementer spent judgment only where the spec left it.

**Phase 1 (2026-08-22) filled the T-matrix below and made the choice: UDS (`AF_UNIX`
`SOCK_STREAM`), the mind listening and the vendor dialling.** The record of record is the
Backlog decision `decision-0004`; the analysis it rests on is this file's T-matrix section.

## Decision: candidate set is closed — UDS, TCP (loopback), stdio

- **Rationale**: decision-0003 narrowed one-way to real wires; the card restates it.
- **Alternatives considered**: in-process/FFI — foreclosed deliberately (an SI-1 breach
  must be structurally impossible); anything brokered (message queue, HTTP server
  framework) — fails T-7's "process-separable but not process-required" spirit by
  adding a third process and is outside the card's closed set.

## The T-matrix (from protocol §8, verbatim criteria) — filled

Legend for the attribution tag in each cell: **(w)** the property comes from the wire
itself; **(f)** the wire is neutral and the property is the framing layer's to provide;
**(a)** the property comes from architecture above both (the port, the process topology).

| | T-1 push | T-2 order/seq | T-3 sessions | T-4 restart indep. | T-5 message-oriented | T-6 backpressure | T-7 separable-not-required |
|---|---|---|---|---|---|---|---|
| **UDS** (`AF_UNIX`, `SOCK_STREAM`) | **Yes (w)** — full-duplex stream, either side writes unsolicited; no request/response idiom to use against the grain | **Yes (w)** — in-order, lossless, no duplication within a connection; `seq` gaps therefore only ever mean deliberate shedding or a session break **(f)** | **Yes (w)** — connections are long-lived by nature and outlive any message; N bodies multiplex over one connection via the envelope's `body` **(f)** | **Yes (w)** — the path is a stable rendezvous name independent of both processes' lifetimes; whichever side restarts, the other re-dials or keeps listening. Wart: the listener must `unlink` a stale path before `bind` | **No (w) / Yes (f)** — a byte stream, so framing supplies message boundaries; the schema half (JSON an engineless fake can emit) is encoding, not wire | **Signal only (w) / policy (f)** — kernel send buffer blocks or returns `EWOULDBLOCK`, which is undifferentiated stall; selective `background` shedding and whole-`observation` frames are the framing's job | **Yes (a)** — the vendor port is an interface declared at the consumer; a wire is one implementation of it, an in-process fake is another. The wire contributes only that its framing is defined over an abstract stream **(f)** |
| **TCP (loopback)** | **Yes (w)** — identical to UDS | **Yes (w)** — identical to UDS (loopback does not reorder; a stream that cannot deliver in order dies instead) **(f)** for `seq` semantics | **Yes (w)** — identical to UDS | **Yes (w)** — the port is a stable rendezvous name; same shape as UDS. Wart: listener rebind needs `SO_REUSEADDR` against `TIME_WAIT` rather than an unlink | **No (w) / Yes (f)** — identical to UDS | **Signal only (w) / policy (f)** — identical to UDS, plus a TCP send window that adds nothing at this message rate | **Yes (a)** — identical to UDS |
| **stdio** (inherited pipes) | **Yes (w)** — a pipe pair is unidirectional each way but the pair is full-duplex; nothing forces pull | **Yes (w)** — a pipe is ordered and lossless; same `seq` conclusion **(f)** | **Partial (w)** — one channel only, and its lifetime **is** the process pair's lifetime; a session can never outlive either end, and a second connection needs more inherited fds handed out at spawn | **No (w)** — **the structural failure.** A pipe has no name: it exists only because a parent created it at spawn. The child can be restarted (by the parent); the **parent cannot be restarted without destroying the channel**, and the child has nothing to re-dial. Restart independence would require a supervisor process that is neither mind nor vendor | **No (w) / Yes (f)** — identical to the others | **Signal only (w) / policy (f)** — pipe buffer blocks the writer; same undifferentiated stall, less tunable | **Yes (a)** — the port argument is unchanged; stdio neither helps nor hurts here |

### The criteria, one by one

**T-1 (push, not pull).** All three are full-duplex byte channels on which either party may
write at any time; none has a request/response idiom that percepts would have to be pushed
*against*. This criterion does not discriminate — but it does place an obligation on the
framing: the wire form must not introduce a poll, a "get percepts since seq N", or any
reply-shaped message beyond the protocol's own `intent_ack`. T-1 is satisfied by all three
wires and can only be lost by the layer above them.

**T-2 (ordered per body, or reorderable by `seq`).** All three deliver bytes in order,
without loss or duplication, *for as long as the connection lives* — and none of them can
lose a message mid-stream without tearing the connection outright. The consequence is worth
stating because it simplifies the framing spec: **a `seq` gap can never mean "the wire
reordered or dropped it"**. It means either the vendor deliberately shed a `background`
percept (§4.11) or the session broke. `percept_id` dedup therefore has exactly one job —
covering a retransmission across a *reconnect* — and no job at all within a session. No wire
here needs a message-level ack or retransmit layer.

**T-3 (long-lived sessions with explicit open/close).** UDS and TCP both hold connections
open indefinitely and can carry several protocol sessions (one per body, or one connection
per vendor with the envelope's `body` disambiguating, per §2.1). stdio is the outlier: the
channel is minted at spawn, cannot be re-established, and admits exactly one peer. Session
*lifecycle* (`session_open`/`session_close`) is framing work on all three; session *capacity*
is where stdio is structurally poorer.

**T-4 (mind restart must not require a vendor restart, and vice versa).** This is the
criterion that decides the shortlist. UDS and TCP each provide a **stable rendezvous name**
— a filesystem path or a port — that outlives both processes, so restart independence is
symmetric: whichever side restarts, the other either keeps listening or re-dials. stdio has
no name. A pipe is a capability created by a parent at spawn, so the topology must be
parent/child, and the parent's restart destroys the channel with nothing for the survivor to
reconnect to. Under decision-0003 the two artifacts are a Minecraft server (operator-started,
heavy) and a Go daemon; making either the other's child is wrong in both directions — a
daemon that dies with the server fails T-4 outright, and a daemon that spawns the Minecraft
server is not how a Minecraft server is run. Recovering T-4 for stdio requires a third
supervisor process, which is the brokered shape research already rejected as outside the
closed set. **stdio fails T-4 structurally, not incidentally.**

**T-5 (message-oriented, with a schema the fake vendor can produce without an engine).**
Two halves, and neither is a wire property here. *Message orientation*: all three candidates
are byte streams, so framing must supply boundaries. The one wire that would have given
message boundaries for free — `AF_UNIX` with `SOCK_SEQPACKET` — **is not available on the
demo host**: verified on this machine (macOS 26.5.1, arm64), `socket(AF_UNIX, SOCK_SEQPACKET)`
fails with `EPROTONOSUPPORT` (errno 43) while `SOCK_STREAM` and `SOCK_DGRAM` both succeed.
That result removes the only reason a wire choice could have shortened the framing spec, and
it is why framing does the same amount of work whichever wire wins. *The schema half* is
about encoding (the protocol's JSON shapes, emittable by an engineless fake per §10) and is
wire-independent.

**T-6 (backpressure that sheds `background` only, and never splits an `observation`).** All
three offer the same primitive: a bounded kernel buffer that blocks the writer or returns
`EWOULDBLOCK` when the reader stops draining. That signal is **undifferentiated** — a stalled
socket stalls urgent percepts exactly as hard as background ones, which is precisely what
§4.11 forbids. So the wire contributes only the *signal that the peer is not draining*; the
policy (a vendor-side per-session queue that drops `background` at enqueue when the socket is
not draining, never drops `urgent`, and emits a frame whole or not at all so an `observation`
is never split) lives above it. No candidate is better or worse at T-6.

**T-7 (process-separable but not process-required).** The in-process fake vendor is already
satisfied by architecture rather than by transport: decision-0003 fixes the vendor port as a
Go interface *declared at the consumer*, so the real transport and the in-memory `FakeVendor`
(§10.1) are two implementations of the same port and the fake never touches a wire. The one
thing the wire choice must not do is make framing untestable without a socket — which is
avoided by defining framing over an abstract byte stream (`io.Reader`/`io.Writer`,
`InputStream`/`OutputStream`), so the same codec runs over `net.Pipe`, over a byte array in a
vector harness, and over the real socket unchanged. True for all three candidates.

## Decision: the wire is **UDS** — `AF_UNIX`, `SOCK_STREAM`, mind listens, vendor dials

Recorded as Backlog **decision-0004** (status `proposed`; the operator ratifies at the
TASK-0007 PR checkpoint, per decision-0001/0002/0003 precedent).

**stdio is eliminated on T-4 and T-3, and the elimination is structural.** Everything above.
It is not a close call and no amount of framing rescues it.

**UDS beats TCP loopback on three things, none of them performance.**

1. **Reachability is the whole argument (SI-1).** A loopback TCP port is dialable by *any*
   process on the host, under any user account. The seam carries the percept stream — the
   only channel by which ground truth ever leaves the vendor — and accepts intents that move
   a body. A channel any local process can open is a channel any local process can read the
   world through and act through. A UDS path is a filesystem object: access is the
   containing directory's mode, so "who may speak on the seam" is an ownership question with
   a `chmod` answer rather than an honour system. This is the same reasoning decision-0003
   used to forbid in-process co-location — *make the breach structurally impossible rather
   than merely forbidden* — applied one layer down, and it is the reason to prefer UDS even
   though both wires are functionally identical.
2. **A TCP bind has a silent-exposure failure mode; a path has none.** Binding `0.0.0.0`
   instead of `127.0.0.1` is a one-character mistake that publishes the seam to the network
   and produces no error, no log line, and no behavioural difference on the demo host. There
   is no analogous typo for a socket path.
3. **The port namespace is global and shared; the path namespace is ours.** A fixed port
   collides with whatever else on the operator's machine wants it, and an ephemeral port
   forces a discovery mechanism the design does not otherwise need. A path under the
   daemon's own state directory collides only with the daemon.

**Direction: the mind listens, the vendor dials.** Three reasons, and the first is the one
that generalizes.

- **A second vendor is a promise of the brief, and only one direction keeps it cheap.** "V2
  is a second body vendor, not a rewrite." A mind that *accepts* vendor connections gains a
  second vendor by having it dial the same path. A mind that dials would need to be told
  every vendor's address — configuration that grows with the thing the seam exists to make
  free.
- **It puts the accept-side asymmetry on the safer side of SI-1.** Whoever listens accepts
  whatever dials. If the *vendor* listened, an unauthorized dialer would become a mind — it
  would receive the percept stream (world state) and could inject intents. Because the
  *mind* listens, an unauthorized dialer can only become a vendor: it can feed a mind lies,
  which the epistemics already treat as a hostile input class (provenance, RM-1…RM-7), but
  it cannot read the world. The vendor never accepts a connection at all, so the surface for
  "who can read the world" is empty rather than merely guarded.
- **Connection direction then matches protocol direction.** The vendor issues the session id
  (§2.1) and speaks first with `session_open` and its manifest (§6.2). Dial-then-speak keeps
  one direction of causation instead of two.

**Reconnect posture falls out of the protocol, not out of the wire.** T-4 names the session
boundary as the recovery unit and §6.3 already specifies rejoin: a dropped connection ends
the session, and the vendor re-dials and opens a **new** session carrying
`continuity.previous_session`, `previous_close_world_time`, and `body_continuous`. So the
wire needs no resume-mid-stream, no message-level acks, and no replay buffer — which is why a
plain reliable stream is sufficient and why nothing in this choice reopens a settled part of
the contract. The vendor re-dials with backoff; the daemon simply keeps listening. Startup
ordering (daemon before mod) is already an accepted consequence of decision-0003's two run
targets, and dial-with-retry makes it soft rather than hard.

**Latency is not a discriminator and was not used as one.** `llm-routing-and-budget.md` §7
item 4 records that nothing in the routing posture excludes any candidate transport — E4's
<3 s first-token ceiling is mind-internal and the seam adds only the `speak` intent hop. At
this project's message rate the difference between UDS and loopback TCP is far below anything
the design can perceive. Choosing on measured throughput here would have been choosing on
noise.

### Rejected wires, each with its stated reason

| Wire | Reason rejected |
|---|---|
| **stdio (inherited pipes)** | **Fails T-4 structurally.** A pipe has no name, so the channel exists only as a parent/child capability minted at spawn: the parent's restart destroys it and the survivor has nothing to reconnect to. Neither direction of the required topology is tenable (a daemon that dies with the Minecraft server violates T-4 outright; a daemon that spawns the Minecraft server is not how the server is run), and restoring independence needs a third supervisor process — the brokered shape already outside the closed set. Also weakest on T-3: exactly one peer, and a session that cannot outlive either process. |
| **TCP (loopback)** | **Functionally adequate, rejected on reachability.** It satisfies T-1…T-7 exactly as well as UDS — this is a preference between two working options, not a defect. It loses because its rendezvous name is a globally-dialable port: any local process under any account can open the seam, read the percept stream, and inject intents, and a `0.0.0.0`-instead-of-`127.0.0.1` bind exposes it off-host silently. UDS replaces that honour system with filesystem permissions. Secondary: port-namespace collision versus a path we own. |

### Costs of the choice, recorded rather than discovered later

- **Stale socket file.** The listener must `unlink` a leftover path before `bind` or fail with
  `EADDRINUSE`. The daemon must also not unlink a path a *live* daemon holds — dial first,
  and only unlink when the dial is refused. (Go's `net` unlinks on listener close by default
  when it created the file, which covers the graceful path but not a crash.)
- **Path length.** `sun_path` is 104 bytes on this host (verified: macOS SDK `sys/un.h`,
  `char sun_path[104]`) and ~108 on Linux. The socket path must therefore be short and
  configured, not derived from a deep working directory; the framing spec pins this and the
  daemon must fail loudly rather than truncate.
- **Java requires the NIO API, not the legacy one.** Verified on this host (JDK 26.0.2):
  `ServerSocketChannel.open(StandardProtocolFamily.UNIX)` / `SocketChannel.open(...)` work,
  while `java.net.Socket` is TCP-only and has no UNIX-family constructor. Unix-domain socket
  channels are JDK 16+ (JEP 380); the Fabric target is well above that, so this costs V1 an
  API choice and no dependency. Go's side is `net.Dial`/`net.Listen` with network `"unix"` —
  stdlib, no dependency (verified toolchain present: go1.26.4 darwin/arm64).
- **Windows portability is deferred, not solved.** The demo host is macOS (plan.md's target
  platform) and no Windows requirement exists. Rather than make a version-specific claim
  about Windows `AF_UNIX` support in either runtime, the mitigation is structural: framing is
  defined over an abstract byte stream, so if a Windows target ever appears, swapping the
  wire is a change to the listen/dial call on each side and touches neither the framing spec
  nor a single vector.

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
