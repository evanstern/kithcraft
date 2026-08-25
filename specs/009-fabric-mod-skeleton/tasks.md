# Tasks: Fabric mod skeleton

**Input**: specs/009-fabric-mod-skeleton/ (spec.md, plan.md)
**Prerequisites**: decision-0004 (accepted), seam-wire-v0.md, seam/vectors/,
operator ruling 2026-08-22 (harness replacement)

**Organization**: Phases map 1:1 to phase-scoped dispatches. US1 = session (spec User
Story 1), US2 = manifest (US2), US3 = tokens (US3).

## Phase 1 — Version re-verification and toolchain (US1 groundwork)

**Goal**: Target versions re-verified with citations; the Gradle/loom project builds
and loads on a dev server.

**Independent test**: `cd mod && ./gradlew build` green; `./gradlew runServer` loads
the mod without error (observation recorded).

- [x] T001 Re-verify target versions with the evidence rule (URL + accessed date):
      current stable Minecraft, matching Fabric loader/API and Yarn build, loom
      version; confirm-or-bump villager-brain-api's yarn-1.21.3+build.1 symbols and
      routing A-2's daylight arithmetic flag; record in
      specs/009-fabric-mod-skeleton/research/versions.md
- [x] T002 Scaffold mod/ — Gradle wrapper, settings.gradle, build.gradle
      (fabric-loom), gradle.properties pinning the T001 versions, fabric.mod.json
      (server entrypoint only), KithcraftMod entrypoint stub; `./gradlew build` green
- [x] T003 Verify dev-server load: `./gradlew runServer` observation recorded (mod id
      in the log, no client jar produced — card AC #7 evidence)

**Checkpoint**: the toolchain exists; JVM code can be written and run.

## Phase 2 — Wire client and vector proof (US1)

**Goal**: The mod's codec round-trips all 17 vectors; the UDS client dials and
completes the handshake against a stub mind.

**Independent test**: JVM vector suite 17/17 via `./gradlew test`; handshake
round-trip observed against the stub listener.

- [ ] T004 Choose the JSON library (T001 records candidates) and run the C-1..C-10
      emit check against the vectors; implement mod/wire/CanonicalJson.java — thin
      wrapper if the library conforms, custom writer if not (record which, per the
      operator ruling's carve-out)
- [ ] T005 Implement mod/wire/FrameCodec.java (4-byte BE length prefix, 1 MiB cap,
      connection-fatal taxonomy) and the vector suite as Gradle tests: census +
      decode/meaning/bytes/validation per vector, mirroring the TASK-0007 harness
      obligations against THIS codec
- [ ] T006 Implement mod/wire/WireClient.java: UDS dial
      (StandardProtocolFamily.UNIX), minimal re-dial backoff, send/receive framing;
      mod/session/Handshake.java + Continuity.java building session_open per
      protocol §6.2 / seam-wire-v0.md §1; prove against an in-test stub mind
      listener (java, no daemon dependency)

**Checkpoint**: the vendor side of the wire is proven against the same pinned bytes
as the mind side.

## Phase 3 — Manifest (L-7) and token registry (US2 + US3)

**Goal**: The manifest is world-independent; tokens survive restarts.

**Independent test**: L-7 byte-identity test green; token persistence proven across a
real `runServer` restart (observation recorded).

- [ ] T007 [US2] Implement the static capability manifest: four-type floor plus
      declared extras, origins, verbs with target shapes, role-annotated
      salient_kinds, bearings, distance bands, time_unit "second" (never ticks);
      named test byte-compares manifests across two bodies in different world states
      (L-7, card AC #3)
- [ ] T008 [US3] Implement mod/tokens/TokenRegistry.java: issue/resolve/retire for
      body/place/thing_id/kind, backed by the world save's persistent mod data;
      tokens never reused (monotonic issuance survives restart)
- [ ] T009 [US3] Prove persistence: unit tests for issue/resolve/retire + a recorded
      dev-server restart observation — tokens issued pre-restart resolve to the same
      referents post-restart (card AC #5)

**Checkpoint**: the vendor's hardest obligation (tokens) and its easiest-to-lose
invariant (L-7) are both tested.

## Phase 4 — Harness replacement, gates, wiki, board

**Goal**: The operator's ruling executed; gates green; grounding honest.

- [ ] T010 Replace seam/java-roundtrip's hand-rolled parsing per the ruling: rebuild
      the harness on the chosen library (keep the census + 4-checks-per-vector + 6
      framing/asymmetry obligations and the mutation-check power; shape — Gradle
      test module vs single-file-over-library — recorded with rationale); delete the
      hand-rolled parser code; harness green over all 17 vectors
- [ ] T011 Run all gates: ./gradlew build + test green; scope check (diff touches
      only mod/, seam/java-roundtrip/, specs/009-*, board, runbook row)
- [ ] T012 Wiki: re-verify villager-brain-api.md against T001's version findings
      (amend if symbols moved), body-protocol-seam.md (first real vendor exists),
      overview.md ("no code" claims); honest re-pins; regenerate CAPSULES.md if
      descriptions changed
- [ ] T013 Tick this file, check card ACs now true (--check-ac), append phase-done
      note

## Dependencies

Phase 1 → 2 → 3 → 4 serial. Within Phase 3: T007 ∥ T008, T009 after T008. T010
depends on T004's library choice (same library).

## Implementation strategy

Toolchain-first because everything else compiles against it; vectors immediately
after so no vendor code can drift from the pinned wire; the manifest and registry are
the two card obligations with named structural tests; the harness replacement lands
last, once the library's conformance is already proven by the mod's own suite.
