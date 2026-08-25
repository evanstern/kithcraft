---
name: body-protocol-seam
description: The anti-corner architecture move — a world-agnostic body protocol (perceive / act / remember) between mind daemon and world, with the Minecraft mod as the first body vendor. Why minds never couple to Minecraft, what the seam buys, the v0 contract (five seam invariants, three surfaces, four core verbs), and the wire beneath it — decision-0004's Unix domain socket, length-prefixed canonical JSON framing, and seventeen golden vectors two harnesses agree on. Load for any protocol, transport, or architecture work.
kind: pattern
sources:
  - docs/design/kithcraft-brief.md
  - docs/design/mod-stack-comparison.md
  - docs/design/body-protocol-v0.md
  - docs/design/seam-wire-v0.md
  - backlog/decisions/decision-0004 - Seam-transport-UDS-AF_UNIX-SOCK_STREAM-mind-listens-and-vendor-dials.md
  - mind/wire/canonical.go
  - mind/seam/session.go
  - mod/src/main/java/dev/kithcraft/mod/wire/CanonicalJson.java
  - mod/src/main/java/dev/kithcraft/mod/session/Handshake.java
  - mind/memory/log.go
  - mind/memory/beliefs.go
  - mind/memory/provenance.go
  - mind/memory/admission.go
verified_against: 4ee4efe074a060e59122106ee839a3590262bce6
---

# Body-protocol seam

The brief's single architecture posture: the mind daemon speaks a world-agnostic protocol
of **perceive / act / remember**, and the Minecraft mod is merely the first *body vendor*
implementing it. Minds never couple to Minecraft. This is the deliberate anti-corner
move — the one interface the project refuses to compromise early.

**The seam is now a contract with a wire under it.** TASK-0002 drafted protocol v0 in
`docs/design/body-protocol-v0.md` — the message shapes, and the contract for anything
field-level. TASK-0007 added the wire beneath them: decision-0004 chose the transport and
`docs/design/seam-wire-v0.md` defines the framing, with `seam/vectors/` pinning it in bytes.
Those documents, not this note, are the contract; what follows is the summary.

## How it works

The seam's guaranteed consequences (from the brief):

- Whether promptworld I's Go mind daemon survives or the mind layer is rebuilt fresh is
  an implementation detail *behind* the seam, not an early commitment.
- A future owned world engine (V2) is a second body vendor, not a rewrite.
- Minds are testable without booting Minecraft.

The seam also did real work in the mod-stack decision: the comparison
(`docs/design/mod-stack-comparison.md`) found both Fabric and Paper keep the seam equally
clean — a mod/plugin is a thin server-side surface with no requirement that mind logic
live in-process — so the seam alone didn't discriminate; dependency health did. What the
seam *did* rule on: Citizens2's OSL-3.0 copyleft posed a license-entanglement risk to a
tightly-coupled body-vendor implementation, while Fabric (Apache-2.0) plus vanilla engine
API leaves the body vendor freely licensable.

## What v0 binds

**Five seam invariants** — break one and it is a different protocol, not a revision:
the mind has **no read access** to the world (percepts are the only path in; no query, no
lookup); every percept is **provenance-stamped at emission** from a closed vocabulary;
percepts are **self-contained when they cross** (the mind cannot call back); the body
reports what *is* while the mind owns what was expected and what to keep; **durable memory
belongs to the mind**.

**Three surfaces.** *Perceive* — a one-directional push of typed percepts (sighting,
observation, sound, speech, told_fact, text, act_result, self_state, change_report), each
carrying provenance and an urgency band. *Act* — an intent/result split: the mind sends
intents, the vendor acks receipt only, and what actually happened returns as an
`act_result` **percept**, because an act that returned world state would be a query in
disguise. Four core verbs every vendor must implement: `go_to`, `speak`, `attend`, `wait`.
*Remember* — mostly a statement of what does **not** cross: the vendor persists bodies,
the world, and the token registry; it never persists, ranks, or weights a mind's memories.

**The mechanism that makes epistemic hygiene checkable:** origins are a closed vocabulary
(`acted`/`saw`/`heard`/`felt`/`told`/`read`) and `direct_perception(origin)` is a pure
function of that value alone, defined in the protocol so mind and vendor cannot disagree.
It never reads prose — which is the only thing that stops a language model writing "I saw
it myself" into a belief. An unknown or absent origin classifies **secondhand, never
direct**, which is what makes additive versioning safe.

Two rules that look like fields and are actually invariants: `observation.vocabulary` (the
closed set of kinds scanned — without it an absence claim has no scope) and the
`change_report` **delivery restriction** (never deliver a change report to the body that
caused it or watched it happen — promptworld I measured that bug at **75% of all memories
formed**).

World concepts cross abstracted: opaque `kind` tokens a mind may compare but never parse,
meaning carried as `roles` plus prose-only `descriptor`, space as opaque place tokens plus
coarse bands (no coordinates, no arithmetic), time as integers in a vendor-declared unit
(never ticks).

## What the wire binds

**The transport is decided.** The protocol deliberately left Q-1 open — the shapes are data,
not a wire format — and TASK-0007 closed it: decision-0004 (proposed) chooses a **Unix domain
socket** (`AF_UNIX`, `SOCK_STREAM`) at a configured path, with the **mind daemon listening and
the body vendor dialing**. Four of the seven T-criteria did not discriminate at all; the choice
lives in T-4 (restart independence) and T-3 (sessions), which eliminate stdio, and between UDS
and loopback TCP in UDS's filesystem-permission access control and its absence of a listening
port. Listening is the mind's because the vendor never accepts a connection — a body that
opened a socket would be an inbound surface SI-1 has no rule for.

`docs/design/seam-wire-v0.md` is the framing spec, the spec-002 successor: **length-prefixed
canonical JSON**, a four-byte big-endian byte count and nothing else in the header, one frame
per message. The header is bare on purpose — a type tag could contradict the envelope, and a
wall-clock stamp would hand a mind a second clock to difference against coarse bands, one step
from the arithmetic AR-4 forbids. Sending is canonical (sorted keys, no whitespace, integers
only, literal UTF-8) while receiving accepts any conforming JSON, which is what makes
byte-exact round-trip vectors affordable rather than a burden. The connection **is** the
session; N bodies multiplex over it; a drop ends the session and the vendor re-dials with
backoff, carrying §6.3's continuity report — there is no resume, no replay buffer, and no
message-level ack, because each would be T-1's forbidden pull channel wearing a different hat.

`seam/vectors/` holds seventeen golden vectors pinning every percept type, intent shape, and
session message in **both** decoded form and exact wire bytes; two harnesses sharing no code
(`seam/go-roundtrip`, `seam/java-roundtrip`) agree on all of them. That agreement, not the
prose, is what makes the wire real — the Go and Java sides now have something to be wrong
against before either exists.

The doc proves the seam two ways: a second-vendor sketch (a trivial text world implementing
the whole surface, which exposed that optional percept channels were being assumed rather
than declared) and a specified in-memory **fake vendor** — scripted percepts in, asserted
acts out — that tests the rules no compiler can, including reproducing the 75% flood on
demand by lifting the delivery restriction.

## First implementations

The seam stopped being contract-only in this branch: TASK-0008 (M1) landed the mind side
in `mind/` — a Go module outside the Gradle build implementing the pieces above. The
framing and canonical-JSON writer that `seam-wire-v0.md` §2.4 specifies are hand-rolled in
`mind/wire/canonical.go` (the `encoding/json` non-conformance noted above is why); the UDS
listener, handshake, byte-identical-capabilities check, and continuity-on-restart from
this note's "what the wire binds" section are implemented in `mind/seam/session.go`. All
seventeen `seam/vectors/` golden vectors round-trip through this codec (a third proof,
alongside the Go and Java throwaway harnesses this note already described). Scope stays
bounded: no memory store, LLM client, or deliberation live here — those are M2/M4/M5.

TASK-0009 (V1) then landed the **first real vendor**: the Fabric mod in `mod/`. Its wire
client mirrors the mind's split — Gson's `JsonReader` for decode, a hand-written canonical
writer for encode, in `mod/src/main/java/dev/kithcraft/mod/wire/CanonicalJson.java`, for
the identical reason `mind/wire/canonical.go` hand-rolls its writer: no library on either
side can be coaxed into emitting C-1..C-10 form. `mod/.../wire/FrameCodec.java` implements
the 4-byte length-prefix framing; `mod/.../wire/WireClient.java` dials the UDS path
(`StandardProtocolFamily.UNIX`, the vendor side of decision-0004's "mind listens, vendor
dials"); `mod/.../session/Handshake.java` builds `session_open` — the static capability
manifest (four-type floor, six origins, four core verbs, world-independent
`salient_kinds`, bearings, distance bands, `time_unit: "second"`) proven byte-identical
across bodies and world states (L-7), which is the vendor half of this note's "what v0
binds" epistemic-hygiene mechanism actually existing in code. `mod/.../tokens/
TokenRegistry.java` persists the body/place/thing_id/kind token registry across server
restarts (SavedData-backed; tokens never reused). All seventeen `seam/vectors/` golden
vectors round-trip through the mod's own codec too — the fourth independent proof, and the
first from the vendor side rather than a throwaway harness. Scope stays bounded here as
well: no percept emission beyond the handshake, no intent execution, no brain/schedule
work — those are V2/V3 (TASK-0014 and beyond).

Both harnesses this note described (`seam/go-roundtrip`, `seam/java-roundtrip`) still
exist as throwaway, spec-facing proof, independent of either real implementation; the Java
one was rebuilt on the same Gson-decode/custom-encode split once TASK-0009 introduced
Gradle (the only reason it had stayed hand-rolled), per the operator's 2026-08-22 ruling —
see `seam/java-roundtrip/README.md`.

TASK-0010 (M2) then landed the mind's own durable memory in `mind/memory/`, the concrete
code behind this note's "durable memory belongs to the mind" invariant and its PM-1
epistemic-hygiene mechanism. `log.go` is the append-only event log: type-level
immutability (`Event`'s fields are unexported, `Log.Append` the only writer) and state as
a reducer replaying the log reproduces. `beliefs.go` is the private, provenance-stamped
belief store (PM-1) — distinct from any vendor resolution index, no external write path —
implementing RM-4..RM-7 as read-time arithmetic on `world_time`: confidence and freshness
are computed at read time and never stored, and only a correction, a death, or a witnessed
removal deletes a belief. `provenance.go` implements this note's `direct_perception`
classifier (§2.7, pure over `origin` alone) and the RM-2/RM-3 citation-resolution gate,
which coerces a model-authored claim down to what its cited percepts support and never
rejects it. `admission.go` implements routing §6.3's deterministic episodic admission
gate between the percept inbox and the log. Scope stays bounded once more: no model call
anywhere in this package (routing §1.2) — a model authoring belief claims is M4/M5's work,
not this one's.

## Connections

Mandated by [[design-brief]]; evaluated as constraint R3 in [[mod-stack-decision]];
first vendor substrate is [[villager-brain-api]]; the reflex/planner doctrine it carries,
and the promptworld I machinery it deliberately does not inherit, are in
[[promptworld-lineage]]. The doctrine port behind v0's cited rule IDs (`EH-n`/`PM-n`/`F-n`)
is `specs/002-body-protocol-v0/research/doctrine-port.md`; the transport analysis behind
decision-0004 — the filled T-matrix and the verified platform facts — is
`specs/007-seam-transport/research.md`. The wire's Go-daemon side is TASK-0008 (M1) and its
Fabric-mod side TASK-0009 (V1); both were blocked on this note's Q-1 and have now landed
([[v1-demo]]).

## Operational notes

The protocol doc is the contract; this note is a map to it. Any change to the message
shapes, the origin vocabulary, or the seam invariants lands there first — and a change to
an invariant or to `DIRECT_ORIGINS` membership is a **major** version bump, not a revision.
Wire-form changes land in `seam-wire-v0.md` instead, and a frame-format change is a protocol
**MAJOR** bump: the frame layer carries no version of its own because nothing in a frame is
readable until the frame format is already assumed.

**Vectors are append-only within a MAJOR.** A MINOR bump adds files to `seam/vectors/`; it
never edits or deletes an existing one, which is what keeps every existing vector valid
byte-for-byte across additive changes — and what turns the retained old vectors into the
older-sender fixtures that prove receivers tolerate unknown fields. Any commit adding a vector
re-runs `body-protocol-v0.md` §12's six leak passes over the added file (`seam-wire-v0.md` §7.3
states the obligation); the vector set is clean today because its author had read both audits,
which is a property that lapses the moment someone who hasn't appends to it.

**Still open:** hearing's engine hook in the first vendor (the protocol's §11 Q-2). Transport
(Q-1) is closed — but decision-0004 is **proposed**, not accepted: the operator ratifies at the
TASK-0007 PR, and merge is the ratification act per the decision-0001/0002/0003 precedent.
