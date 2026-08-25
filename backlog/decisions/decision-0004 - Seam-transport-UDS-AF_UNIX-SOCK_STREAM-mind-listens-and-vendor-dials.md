---
id: decision-0004
title: 'Seam transport: UDS (AF_UNIX SOCK_STREAM), mind listens and vendor dials'
date: '2026-08-22 03:10'
status: proposed
---
## Context

`docs/design/body-protocol-v0.md` §8 left the mind↔vendor wire an **open question (Q-1)** with
seven constraints (T-1…T-7) and no answer; §11 carries it forward. decision-0003 narrowed the
option set **one way and deliberately**: a Go daemon against a JVM mod forecloses the in-process
option, so Q-1 is a choice among **real wires only — UDS, TCP (loopback), stdio** — and not a
choice between a wire and a method call. The narrowing is itself an SI-1 move: separate
processes make a breach structurally impossible rather than merely forbidden.

TASK-0007 (S1) closes Q-1. This decision is its first deliverable; the framing spec
(`docs/design/seam-wire-v0.md`) and the golden vectors under `seam/vectors/` are the other two
and are written against this choice. Every m-0 task encodes against it: TASK-0008 (M1, the Go
daemon) and TASK-0009 (V1, the Fabric mod) both depend on TASK-0007 and cannot write a line of
connection code until the wire is named.

Full analysis, including the criterion-by-criterion walk and the verified platform facts:
`specs/007-seam-transport/research.md`. Sits on decision-0001 (Fabric server-side mod),
decision-0002 (augmented vanilla `VillagerEntity`), and decision-0003 (Go mind daemon), and
behind the accepted `docs/design/body-protocol-v0.md`.

**Status: PROPOSED — pending operator ratification at the TASK-0007 PR checkpoint.** Not yet
in force. Per decision-0001/0002/0003 precedent, merge is the ratification act and status flips
by direct edit (the backlog CLI's decision command supports only `create`, no edit verb).

### Candidates — the closed set

**UDS** (`AF_UNIX`, `SOCK_STREAM`), **TCP (loopback)**, **stdio** (inherited pipes). Nothing
else was in scope: in-process/FFI is foreclosed by decision-0003, and anything brokered (a
message queue, an HTTP server framework) adds a third process and fails T-7's
"process-separable but not process-required" spirit.

## Decision

**The seam runs over a Unix domain socket — `AF_UNIX`, `SOCK_STREAM` — at a configured
filesystem path. The mind daemon listens; the body vendor dials.** A dropped connection ends
the session; the vendor re-dials with backoff and opens a **new** session carrying §6.3's
continuity report. Framing is defined over an abstract byte stream, so the same codec runs
over a socket, over an in-memory pipe, and over a byte array in a vector harness.

### The T-matrix

Attribution tag per cell: **(w)** the property comes from the wire itself; **(f)** the wire is
neutral and the property is the framing layer's to provide; **(a)** it comes from architecture
above both (the vendor port, the process topology).

| | T-1 push | T-2 order/seq | T-3 sessions | T-4 restart indep. | T-5 message-oriented | T-6 backpressure | T-7 separable-not-required |
|---|---|---|---|---|---|---|---|
| **UDS** (`AF_UNIX`, `SOCK_STREAM`) | **Yes (w)** — full-duplex stream, either side writes unsolicited; no request/response idiom to use against the grain | **Yes (w)** — in-order, lossless, no duplication within a connection; a `seq` gap therefore only ever means deliberate shedding or a session break **(f)** | **Yes (w)** — connections are long-lived by nature and outlive any message; N bodies multiplex over one connection via the envelope's `body` **(f)** | **Yes (w)** — the path is a stable rendezvous name independent of both processes' lifetimes; whichever side restarts, the other re-dials or keeps listening. Wart: a stale path must be unlinked before `bind` | **No (w) / Yes (f)** — a byte stream, so framing supplies message boundaries; the schema half (JSON an engineless fake can emit) is encoding, not wire | **Signal only (w) / policy (f)** — kernel send buffer blocks or returns `EWOULDBLOCK`, an undifferentiated stall; selective `background` shedding and whole-`observation` frames are the framing's job | **Yes (a)** — the vendor port is an interface declared at the consumer; a wire is one implementation, an in-process fake another. The wire contributes only that framing is defined over an abstract stream **(f)** |
| **TCP (loopback)** | **Yes (w)** — identical to UDS | **Yes (w)** — identical to UDS; loopback does not reorder, and a stream that cannot deliver in order dies instead **(f)** for `seq` semantics | **Yes (w)** — identical to UDS | **Yes (w)** — the port is a stable rendezvous name; same shape as UDS. Wart: rebind needs `SO_REUSEADDR` against `TIME_WAIT` rather than an unlink | **No (w) / Yes (f)** — identical to UDS | **Signal only (w) / policy (f)** — identical to UDS, plus a send window that adds nothing at this message rate | **Yes (a)** — identical to UDS |
| **stdio** (inherited pipes) | **Yes (w)** — a pipe pair is unidirectional each way but the pair is full-duplex; nothing forces pull | **Yes (w)** — a pipe is ordered and lossless; same `seq` conclusion **(f)** | **Partial (w)** — one channel only, and its lifetime **is** the process pair's lifetime; a session can never outlive either end, and a second connection needs more fds handed out at spawn | **No (w)** — **the structural failure.** A pipe has no name: it exists only because a parent created it at spawn. The parent cannot be restarted without destroying the channel, and the survivor has nothing to re-dial. Independence would require a supervisor process that is neither mind nor vendor | **No (w) / Yes (f)** — identical to the others | **Signal only (w) / policy (f)** — pipe buffer blocks the writer; same undifferentiated stall, less tunable | **Yes (a)** — unchanged; stdio neither helps nor hurts here |

**What the matrix says as a whole:** four of the seven criteria (T-1, T-2, T-5, T-6) do not
discriminate at all, and T-7 is settled by architecture rather than by transport. The decision
lives in **T-4 and T-3** — which eliminate stdio — and, between the two survivors, in a
property T-1…T-7 does not name and SI-1 does: **who else on the host can open the channel.**

### The criteria, one by one

**T-1 (push, not pull).** All three are full-duplex byte channels on which either party may
write at any time; none has a request/response idiom percepts would have to be pushed
*against*. Not discriminating — but it places an obligation on framing: the wire form must
introduce no poll, no "get percepts since seq N", and no reply-shaped message beyond the
protocol's own `intent_ack`. T-1 can only be lost above the wire.

**T-2 (ordered per body, or reorderable by `seq`).** All three deliver bytes in order, without
loss or duplication, for as long as the connection lives, and none can lose a message
mid-stream without tearing the connection outright. The consequence simplifies the framing
spec: **a `seq` gap can never mean "the wire reordered or dropped it."** It means the vendor
deliberately shed a `background` percept (§4.11), or the session broke. `percept_id` dedup
therefore has exactly one job — covering retransmission across a *reconnect* — and none within
a session. No wire here needs a message-level ack or retransmit layer.

**T-3 (long-lived sessions with explicit open/close).** UDS and TCP hold connections open
indefinitely and can carry several protocol sessions (one per body, or one connection per
vendor with the envelope's `body` disambiguating, §2.1). stdio admits exactly one peer over a
channel minted at spawn that cannot be re-established. Session *lifecycle* is framing work on
all three; session *capacity* is where stdio is structurally poorer.

**T-4 (mind restart must not require a vendor restart, and vice versa).** The criterion that
decides the shortlist. UDS and TCP each provide a **stable rendezvous name** — a filesystem
path or a port — that outlives both processes, so restart independence is symmetric. stdio has
no name: a pipe is a capability a parent creates at spawn, so the topology must be
parent/child and the parent's restart destroys the channel. Under decision-0003 the two
artifacts are a Minecraft server (operator-started, heavy) and a Go daemon; making either the
other's child is wrong in both directions — a daemon that dies with the server fails T-4
outright, and a daemon that spawns the Minecraft server is not how a Minecraft server is run.
Recovering T-4 for stdio needs a third supervisor process: the brokered shape already outside
the closed set.

**T-5 (message-oriented, schema producible without an engine).** Two halves, neither a wire
property here. *Message orientation*: all three are byte streams, so framing supplies
boundaries. The one wire that would have given boundaries for free — `AF_UNIX` with
`SOCK_SEQPACKET` — **is not available on the demo host**: verified on this machine
(macOS 26.5.1, arm64), `socket(AF_UNIX, SOCK_SEQPACKET)` fails with `EPROTONOSUPPORT`
(errno 43) while `SOCK_STREAM` and `SOCK_DGRAM` both succeed. That removes the only way a wire
choice could have shortened the framing spec. *The schema half* — the protocol's JSON shapes,
emittable by an engineless fake per §10 — is encoding, and wire-independent.

**T-6 (shed `background` only; never split an `observation`).** All three offer the same
primitive: a bounded kernel buffer that blocks the writer or returns `EWOULDBLOCK` when the
reader stops draining. That signal is **undifferentiated** — a stalled socket stalls urgent
percepts exactly as hard as background ones, which §4.11 forbids. The wire contributes only
*the signal that the peer is not draining*; the policy (a vendor-side per-session queue
dropping `background` at enqueue when the socket is not draining, never dropping `urgent`,
emitting each frame whole or not at all so an `observation` is never split) lives above it. No
candidate is better or worse.

**T-7 (process-separable but not process-required).** Already satisfied by architecture:
decision-0003 fixes the vendor port as an interface *declared at the consumer*, so the real
transport and the in-memory `FakeVendor` (§10.1) are two implementations of one port and the
fake never touches a wire. The wire's only obligation is not to make framing untestable
without a socket — avoided by defining framing over an abstract byte stream
(`io.Reader`/`io.Writer`, `InputStream`/`OutputStream`), so one codec serves a real socket, an
in-memory pipe, and a byte array in a vector harness. True for all three candidates.

### Rationale for UDS

**stdio is eliminated on T-4 and T-3, structurally.** Not a close call, and no amount of
framing rescues it.

**UDS beats TCP loopback on three things, none of them performance.**

1. **Reachability is the whole argument (SI-1).** A loopback TCP port is dialable by *any*
   process on the host, under any user account. The seam carries the percept stream — the only
   channel by which ground truth ever leaves the vendor — and accepts intents that move a
   body. A channel any local process can open is a channel any local process can read the
   world through and act through. A UDS path is a filesystem object: access is the containing
   directory's mode, so "who may speak on the seam" becomes an ownership question with a
   `chmod` answer instead of an honour system. This is decision-0003's own reasoning — *make
   the breach structurally impossible rather than merely forbidden* — applied one layer down.
2. **A TCP bind has a silent-exposure failure mode; a path has none.** Binding `0.0.0.0`
   instead of `127.0.0.1` is a one-character mistake that publishes the seam to the network
   with no error, no log line, and no behavioural difference on the demo host. There is no
   analogous typo for a socket path.
3. **The port namespace is global and shared; the path namespace is ours.** A fixed port
   collides with whatever else on the operator's machine wants it; an ephemeral port forces a
   discovery mechanism the design does not otherwise need. A path under the daemon's own state
   directory collides only with the daemon.

**Direction: the mind listens, the vendor dials.**

- **A second vendor is a promise of the brief, and only one direction keeps it cheap.** "V2 is
  a second body vendor, not a rewrite." A mind that *accepts* connections gains a second
  vendor by having it dial the same path. A mind that dialled would need to be told every
  vendor's address — configuration that grows with the thing the seam exists to make free.
- **It puts the accept-side asymmetry on the safer side of SI-1.** Whoever listens accepts
  whatever dials. If the *vendor* listened, an unauthorized dialer would become a mind: it
  would receive the percept stream and could inject intents. Because the *mind* listens, an
  unauthorized dialer can only become a vendor — it can feed a mind lies, which the epistemics
  already treat as a hostile input class (provenance, RM-1…RM-7), but it cannot read the
  world. The vendor never accepts a connection at all, so the surface for "who can read the
  world" is empty rather than merely guarded.
- **Connection direction then matches protocol direction.** The vendor issues the session id
  (§2.1) and speaks first, with `session_open` and its manifest (§6.2). Dial-then-speak keeps
  one direction of causation instead of two.

**Reconnect posture falls out of the protocol, not the wire.** T-4 names the session boundary
as the recovery unit and §6.3 already specifies rejoin, so a dropped connection ends the
session and the vendor re-dials and opens a **new** one carrying `continuity.previous_session`,
`previous_close_world_time`, and `body_continuous`. The wire needs no resume-mid-stream, no
message-level acks, and no replay buffer — which is why a plain reliable stream suffices and
why nothing here reopens a settled part of the contract.

**Latency was not used as a discriminator.** `llm-routing-and-budget.md` §7 item 4 records that
nothing in the routing posture excludes any candidate transport: E4's <3 s first-token ceiling
is mind-internal and the seam adds only the `speak` intent hop. At this project's message rate
the gap between UDS and loopback TCP is far below anything the design can perceive; choosing on
measured throughput would have been choosing on noise.

### Alternatives rejected, each with its stated reason

- **stdio (inherited pipes)** — **fails T-4 structurally.** A pipe has no name, so the channel
  exists only as a parent/child capability minted at spawn: the parent's restart destroys it
  and the survivor has nothing to reconnect to. Neither direction of the required topology is
  tenable (a daemon that dies with the Minecraft server violates T-4 outright; a daemon that
  spawns the Minecraft server is not how the server is run), and restoring independence needs
  a third supervisor process — the brokered shape already outside the closed set. Also weakest
  on T-3: exactly one peer, and a session that cannot outlive either process.
- **TCP (loopback)** — **functionally adequate, rejected on reachability.** It satisfies
  T-1…T-7 exactly as well as UDS; this is a preference between two working options, not a
  defect. It loses because its rendezvous name is a globally-dialable port: any local process
  under any account can open the seam, read the percept stream, and inject intents, and a
  `0.0.0.0`-instead-of-`127.0.0.1` bind exposes it off-host silently. UDS replaces that honour
  system with filesystem permissions. Secondary: port-namespace collision versus a path we own.
- **In-process / FFI** — out of scope by construction (decision-0003's one-way narrowing); an
  SI-1 breach must be structurally impossible.
- **Anything brokered** (message queue, HTTP server framework) — outside the card's closed set
  and against T-7's spirit: it adds a third process the seam does not need.

### Costs of the choice, recorded rather than discovered later

- **Stale socket file.** The listener must `unlink` a leftover path before `bind` or fail with
  `EADDRINUSE` — and must not unlink a path a *live* daemon holds: dial first, unlink only
  when the dial is refused. (Go's `net` unlinks on listener close when it created the file,
  which covers the graceful path but not a crash.)
- **Path length.** `sun_path` is 104 bytes on this host (verified: macOS SDK `sys/un.h`,
  `char sun_path[104]`) and ~108 on Linux. The path must be short and configured, never
  derived from a deep working directory; the framing spec pins this and the daemon must fail
  loudly rather than truncate.
- **Java needs the NIO API, not the legacy one.** Verified on this host (JDK 26.0.2):
  `ServerSocketChannel.open(StandardProtocolFamily.UNIX)` and `SocketChannel.open(...)` work,
  while `java.net.Socket` is TCP-only with no UNIX-family constructor. Unix-domain socket
  channels are JDK 16+ (JEP 380) and the Fabric target is well above that, so this costs V1 an
  API choice and no dependency. Go's side is `net.Dial`/`net.Listen` with network `"unix"` —
  stdlib, no dependency (toolchain verified present: go1.26.4 darwin/arm64).
- **Windows portability is deferred, not solved.** The demo host is macOS and no Windows
  requirement exists. Rather than make a version-specific claim about Windows `AF_UNIX`
  support in either runtime, the mitigation is structural: framing is defined over an abstract
  byte stream, so if a Windows target ever appears, swapping the wire is a change to the
  listen/dial call on each side and touches neither the framing spec nor a single vector.

## Consequences

- **Seam Q-1 is closed.** `docs/design/body-protocol-v0.md` §8 and §11's Q-1, and
  `docs/wiki/body-protocol-seam.md`'s "transport is deliberately undecided", are superseded by
  this record plus `docs/design/seam-wire-v0.md` (the framing spec, TASK-0007 Phase 2).
- **Framing does all of T-2/T-5/T-6's work.** The wire supplies ordering and a stall signal and
  nothing else; message boundaries, selective `background` shedding, and the
  never-split-an-`observation` rule are the framing spec's obligations, on any of the three
  candidates. `SOCK_SEQPACKET`'s unavailability on this host is why no wire choice could have
  shortened that document.
- **`percept_id` dedup covers reconnect only.** Because the wire cannot reorder or silently
  drop within a connection, dedup has no within-session job. Both implementations should say so
  rather than build a general at-least-once machine.
- **No new dependency on either side.** Go stdlib `net` with network `"unix"`; Java
  `java.nio.channels` with `StandardProtocolFamily.UNIX`.
- **Startup ordering is soft, not hard.** The daemon must be listening before the vendor's
  first successful dial, but dial-with-backoff means launch order is a convenience rather than
  a requirement. This is the operational face of decision-0003's "two run targets and an
  ordering."
- **Ratification: pending.** Proposed by TASK-0007; becomes settled fact for TASK-0008 and
  TASK-0009 only once the operator ratifies at the PR checkpoint.

## Narrowing effects — what M1 (TASK-0008) and V1 (TASK-0009) must each implement

The connection story, split by side, so neither implementation has to re-derive it. Both sides
consume this section plus `docs/design/seam-wire-v0.md`; nothing below is optional and nothing
below is more than the connection story (message shapes stay the protocol's, framing stays the
framing spec's).

### M1 — the Go mind daemon (TASK-0008) — **the listener**

- **Listens.** `net.Listen("unix", <configured path>)` at a path from configuration (with a
  default under the daemon's own state directory), not a path derived from the working
  directory. Reject a path too long for `sun_path` at startup with a clear error — never
  truncate.
- **Owns the socket file's lifecycle.** On startup, if the path exists: attempt a dial; if the
  dial succeeds, another daemon is live — exit rather than steal the seam. If the dial is
  refused, the file is stale — unlink it and bind. On graceful shutdown, close the listener
  (which removes the file) and best-effort unlink on signal paths.
- **Sets access by permissions, not by hope.** The socket's containing directory is
  owner-only; this is the mechanism that makes the reachability argument real rather than
  rhetorical, so it is M1's obligation and not a deployment note.
- **Accepts repeatedly, and forever.** `Accept` runs in a loop for the daemon's whole life: a
  vendor disconnect is normal, not terminal, and a reconnect is a fresh accepted connection.
  The daemon never dials and never exits because a vendor went away.
- **Treats a dropped connection as a session end, per §6.3.** Close out the session's state as
  though `session_close` arrived (reason `error` where none did), keep all durable memory
  (SI-5: memory is the mind's and survives the vendor), and wait for the next connection.
- **Accepts a new session on reconnect and honours continuity.** The next `session_open` will
  carry `previous_session`, `previous_close_world_time`, and `body_continuous`; M1 must join on
  it and **must not** backfill the gap. The gap is a gap.
- **Reads framing off an abstract stream.** The codec takes `io.Reader`/`io.Writer` so the same
  code path serves the socket, `net.Pipe`, and the golden-vector harness. The `net.Listener` is
  the only UDS-aware code in the daemon; nothing above it knows what a socket is.
- **Does not build a resume-mid-stream or replay layer.** The session boundary is the recovery
  unit. Dedup by `percept_id` guards reconnect only.
- **Acceptance check this pins:** "restart the daemon mid-session and the villagers keep their
  memories" (decision-0003) is now testable exactly — kill the daemon, the vendor's writes
  fail, the vendor re-dials, a new session opens with continuity, memory is intact.

### V1 — the Fabric mod (TASK-0009) — **the dialler**

- **Dials.** `SocketChannel.open(UnixDomainSocketAddress.of(path))` via
  `java.nio.channels` — **not** `java.net.Socket`, which has no UNIX family. The path comes
  from mod configuration and must match the daemon's.
- **Dials with backoff, and keeps dialling.** The Minecraft server may start before the daemon.
  A refused dial is expected, not an error: retry with bounded backoff and never block the
  server thread doing it. Server startup must not fail because the mind is not up yet.
- **Speaks first.** On a successful dial the vendor sends `session_open` with the full manifest
  (§6.2) — including the four-type percept floor, origins, verbs with target shapes,
  role-annotated `salient_kinds`, bearings, distance bands — and the vendor issues the session
  id.
- **Re-dials on drop and opens a NEW session with continuity.** Never attempts to resume the
  old session. The new `session_open` carries `continuity.previous_session`,
  `previous_close_world_time`, and honest `body_continuous`. It must not report what happened
  during the gap (§6.3); a returning body may discover changes through ordinary
  `change_report` percepts, stamped at resume time.
- **Never listens.** The mod opens no port and no socket of its own. This is load-bearing for
  SI-1, not a stylistic preference: it means there is no surface at all by which another
  process could attach to the vendor and read the world.
- **Owns the T-6 policy on the write side.** A per-session outbound queue that drops
  `background` percepts at enqueue when the socket is not draining, never drops `urgent`, and
  writes each frame whole or not at all so an `observation` is never split. The socket only
  reports that the peer is not draining; the discrimination is V1's.
- **Writes framing to an abstract stream.** The codec takes `OutputStream`/`InputStream` (or
  buffers) so it is testable against the golden vectors without a socket; the `SocketChannel`
  is the only UDS-aware code in the mod.
- **Sends `session_close` on orderly shutdown** (`reason: "shutdown"`) so the daemon's session
  end is clean rather than inferred from a dropped connection.
