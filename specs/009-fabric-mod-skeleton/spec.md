# Feature Specification: Fabric mod skeleton

**Feature Branch**: `task-0009-fabric-mod-skeleton` · **Spec dir**: `specs/009-fabric-mod-skeleton`

**Created**: 2026-08-22 · **Status**: Draft

**Input**: TASK-0009 / V1 — Fabric mod skeleton, vendor session, capability manifest,
token registry. Consumes decision-0001 (Fabric server-side), decision-0002 (augmented
vanilla villager, no client jar), decision-0004 (UDS, vendor dials), seam-wire-v0.md,
seam/vectors/.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A mod that can hold a protocol session (Priority: P1)

As a future implementer, I want the mod to dial the daemon and complete the
`session_open` handshake per the pinned wire, so that every later vendor surface
(V2/V3) transmits over a proven channel.

**Why this priority**: The transport client and handshake are what every other V-task
assumes exists.

**Independent Test**: On a dev server the mod loads, a villager's session opens
against the daemon or a stub mind, and the handshake round-trips S1's golden vectors
(card AC #1).

**Acceptance Scenarios**:

1. **Given** a running mind daemon (or stub) listening on the configured UDS path,
   **When** the server starts with the mod, **Then** the mod dials, sends
   `session_open` per seam-wire-v0.md (canonical bytes), and completes the handshake.
2. **Given** the 17 golden vectors, **When** run through the mod's codec, **Then**
   every vector round-trips (decode → structural equality; encode → byte equality).

### User Story 2 - A manifest that describes the vendor, not the world (Priority: P1)

As a future implementer, I want the capability manifest identical for every body and
independent of world state, so that SI-1 is not defeated at `session_open` (L-7).

**Independent Test**: Card AC #3 — open sessions for multiple bodies in different
world states; byte-compare the `capabilities` objects: identical.

**Acceptance Scenarios**:

1. **Given** any two bodies anywhere in any world, **Then** their manifests are
   byte-identical (also enforced daemon-side per seam-wire-v0.md, but the vendor must
   be correct, not merely caught).
2. **Given** the manifest, **Then** it declares at minimum the four-type floor
   (act_result, observation, sighting, speech), origins, verbs with target shapes,
   role-annotated salient_kinds, bearings, distance bands (card AC #2), and
   `time_unit` is a declared unit, never raw ticks (card AC #4).

### User Story 3 - Tokens that outlive the session (Priority: P1)

As a future implementer, I want the token registry persisted across server restarts
with tokens never reused, so that a mind's memories keep meaning what they meant.

**Independent Test**: Card AC #5 — issue tokens, restart the dev server, re-resolve:
same referents; retired tokens never reappear.

**Acceptance Scenarios**:

1. **Given** tokens issued before a restart, **When** the server restarts, **Then**
   each token resolves to the same referent (body/place/thing/kind).
2. **Given** any token ever issued, **Then** it is never reissued for a different
   referent, across all sessions and restarts.

### Edge Cases

- Daemon not running at server start / daemon dies mid-session: the mod's dial-retry
  posture per decision-0004's narrowing effects (vendor re-dials with backoff; bodies
  keep living engine-side while disconnected — full behavior is V3's, but the client
  must not crash the server).
- Version re-verification (card AC #6): villager-brain-api's symbols were checked at
  yarn-1.21.3+build.1; the target Minecraft/Fabric/Yarn versions are re-verified with
  the evidence rule (URL + accessed date) before the toolchain is pinned.
- No client jar is produced (card AC #7, decision-0002).

## Requirements *(mandatory)*

- **FR-001**: A Fabric server-side mod project (Gradle + fabric-loom) against the
  re-verified target version; loads on a dev server.
- **FR-002**: A transport client per decision-0004/seam-wire-v0.md: UDS dial,
  length-prefixed canonical-JSON codec proven against all 17 vectors (JVM-side).
- **FR-003**: `session_open` per protocol §6.2: time_unit declared (never ticks),
  continuity fields, the full capability manifest — identical for every body (L-7).
- **FR-004**: The token registry: body/place/thing_id/kind → referent, persisted
  across server restarts, tokens never reused.
- **FR-005**: Per the operator's 2026-08-22 ruling: this PR (the one that introduces
  Gradle) replaces seam/java-roundtrip's hand-rolled JSON parsing with a
  library-based harness; the canonical writer stays custom only if the chosen library
  provably cannot emit C-1..C-10 form (verified against the vectors, recorded).
- **FR-006**: Scope: no percept emission beyond the handshake, no intent execution
  (V2), no brain/schedule work (V3), no client jar.

## Success Criteria *(mandatory)*

- **SC-001**: `gradle build` green; mod loads on a dev server without error.
- **SC-002**: JVM vector suite 17/17 against the mod's real codec.
- **SC-003**: L-7 manifest byte-identity demonstrated by a named test.
- **SC-004**: Token persistence demonstrated across a real server restart.
- **SC-005**: Version facts re-verified and cited (evidence rule).

## Assumptions

- The operator's environment provides a JDK (26.0.2 verified present in TASK-0007);
  Gradle arrives via the wrapper this task commits.
- A stub mind (trivial listener honoring the handshake) is in scope for dev-server
  proof if TASK-0008's daemon is not yet merged — the two lane-1 tasks must not
  depend on each other.
