# Implementation Plan: Fabric mod skeleton

**Branch**: `task-0009-fabric-mod-skeleton` | **Date**: 2026-08-22 | **Spec**: specs/009-fabric-mod-skeleton/spec.md

## Summary

First Java in the repo: the Fabric server-side mod skeleton. Gradle/fabric-loom
toolchain against a re-verified target version, the UDS transport client with a
vector-proven codec, the `session_open` handshake with the L-7-safe manifest, and the
persisted token registry. Also executes the operator's ruling: the hand-rolled
java-roundtrip harness is replaced by a library-based one in this same PR.

## Technical Context

**Language/Version**: Java (JDK 26.0.2 verified on host); target Minecraft/Fabric/Yarn
versions re-verified in Phase 1 with citations (villager-brain-api was checked at
yarn-1.21.3+build.1 — confirm or bump).
**Primary Dependencies**: fabric-loader + fabric-api (versions cited at
re-verification), Gradle via committed wrapper, fabric-loom. A JSON library for the
harness replacement (FR-005) — chosen in Phase 1 with the C-1..C-10 emit check
recorded; JDK 16+ UDS via `StandardProtocolFamily.UNIX` SocketChannel (verified in
TASK-0007) needs no dependency.
**Storage**: Token registry persisted in the world save's mod data (Fabric's
PersistentState or equivalent — exact mechanism verified against the target version).
**Testing**: JUnit via Gradle for codec/manifest/registry units; dev-server
(`gradle runServer`) observations recorded in the PR for load + handshake proofs.
**Target Platform**: Server-side only; no client jar.
**Project Type**: Fabric mod (Gradle project at `mod/`).
**Constraints**: SI-1 — the manifest must not read world state (L-7); tokens are
opaque and never reused; canonical bytes on the wire.
**Scale/Scope**: Skeleton only — handshake + registry; V2/V3 build the surfaces.

## Constitution Check

Constitution at `.specify/memory/constitution.md` is an **unfilled template** (stated
plainly, house precedent). Checked against grounding docs:

- decision-0001/0002: Fabric, server-side, no client jar — the project shape — PASS.
- decision-0004 + seam-wire-v0.md: vendor dials, canonical framing — consumed as
  fixed inputs — PASS.
- body-protocol-v0.md §6.2: manifest contents and L-7 — mapped to FR-003 — PASS.
- Operator ruling 2026-08-22 (runbook + card note): harness replacement rides this
  PR — FR-005 — PASS.
- One-task-one-PR; no V2/V3 scope — PASS (FR-006).

**Post-design re-check**: structure adds nothing beyond FR-001..FR-006 — PASS.

## Project Structure

### Documentation (this feature)

```text
specs/009-fabric-mod-skeleton/
├── README.md  spec.md  plan.md  tasks.md
└── research/versions.md   # Phase 1 output: re-verified version facts, cited
```

### Source Code (repository root)

```text
mod/
├── gradle/ gradlew gradlew.bat settings.gradle build.gradle gradle.properties
├── src/main/resources/fabric.mod.json
├── src/main/java/dev/kithcraft/mod/
│   ├── KithcraftMod.java          # entrypoint: config, registry load, client start
│   ├── wire/                      # framing + canonical codec (seam-wire-v0.md)
│   │   ├── FrameCodec.java        # length prefix, caps, connection-fatal errors
│   │   ├── CanonicalJson.java     # canonical writer (custom only if the library
│   │   │                          #   fails the C-1..C-10 emit check — recorded)
│   │   └── WireClient.java        # UDS dial, backoff re-dial, session plumbing
│   ├── session/
│   │   ├── Handshake.java         # session_open builder; L-7-safe static manifest
│   │   └── Continuity.java        # previous-session fields, body-token matching
│   └── tokens/
│       └── TokenRegistry.java     # issue/resolve/retire; PersistentState-backed
└── src/test/java/…                # vector suite, L-7 byte-identity, registry tests

seam/java-roundtrip/               # REPLACED per operator ruling: library-based
                                   # harness (Gradle test module or kept single-file
                                   # over the chosen library — Phase 4 decides shape,
                                   # obligations unchanged: census + 4 checks/vector)
```

**Structure Decision**: `mod/` mirrors `mind/` as the vendor-side artifact root
(decision-0003: the decomposition splits at the seam). The wire package is
deliberately parallel to the Go daemon's — same two governing documents, same layer
split. The manifest is a static constant assembled from declared capabilities — the
L-7 test byte-compares it across bodies/world states, and its content never touches a
world query API by construction.

## Complexity Tracking

No violations. Deliberate simplifications: dial-retry is minimal backoff (full
disconnected-body behavior is V3's); the stub mind for dev-server proof is a trivial
handshake listener, not a daemon.
