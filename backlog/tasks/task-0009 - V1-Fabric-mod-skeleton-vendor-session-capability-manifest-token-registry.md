---
id: TASK-0009
title: 'V1 - Fabric mod skeleton, vendor session, capability manifest, token registry'
status: In Progress
assignee: []
created_date: '2026-08-21 23:36'
updated_date: '2026-08-25 19:21'
labels:
  - vendor
  - m-0-build
milestone: m-0
dependencies:
  - TASK-0007
documentation:
  - docs/design/demo-build-plan.md
  - docs/design/entity-implementation-comparison.md
  - docs/design/body-protocol-v0.md
priority: high
ordinal: 9000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
As a future implementer, I want a mod that can hold a protocol session and hand out tokens that still mean the same thing after a restart, so that the mind's memories survive the world they were formed in.

**Scope boundary.** The mod jar and its dev-server setup against the target Minecraft version — and **re-verify the Yarn mappings and version-dependent facts here**, since `villager-brain-api`'s symbol names were checked at yarn-1.21.3+build.1 and routing's A-2 daylight arithmetic is flagged "verify against the target version before building." The transport client per S1. The `session_open` handshake: `time_unit` (a declared unit, **never ticks**), continuity, capabilities — declared `percept_types` (at minimum the four-type floor: `act_result`, `observation`, `sighting`, `speech`), origins, verbs with target shapes, `salient_kinds` in role-annotated form, bearings, distance bands. The **token registry** — `body`/`place`/`thing_id`/`kind` -> referent, persisted across sessions: the vendor's hardest obligation, and tokens are never reused. No client jar (decision-0002).

**Done proves.** On a dev server: the mod loads, a villager's session opens against the daemon (or a stub mind) and the handshake round-trips S1's golden vectors. **The manifest is identical for every body and does not vary with world state** — the L-7 test, because a vendor populating `salient_kinds` from what is nearby has made `session_open` a "what is around me" query and defeated SI-1 before the first percept. Tokens issued before a server restart still resolve to the same referents after it.

**Depends on.** S1 (the wire and its golden vectors).

**References.** docs/design/demo-build-plan.md section 3.3 (V1) is the plan of record. Ratified surfaces consumed: decision-0001 (Fabric server-side mod), decision-0002 + docs/design/entity-implementation-comparison.md (augmented vanilla VillagerEntity, bounded ~4-injection Mixin surface, no client jar), docs/design/body-protocol-v0.md (session_open, capability manifest, token discipline, SI-1).

**Suggested tier: `sonnet` (next sweep's runbook decides).**

Spec: specs/009-fabric-mod-skeleton
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 On a dev server the mod loads and a villager's session opens against a daemon or stub mind, round-tripping S1's golden vectors
- [x] #2 The capability manifest declares at minimum the four-type percept floor (act_result, observation, sighting, speech), origins, verbs with target shapes, role-annotated salient_kinds, bearings and distance bands
- [x] #3 The manifest is identical for every body and does not vary with world state (the L-7 test)
- [x] #4 time_unit is a declared unit, never raw ticks
- [x] #5 The token registry persists across sessions: tokens issued before a server restart resolve to the same referents after it, and tokens are never reused
- [x] #6 Yarn mappings and version-dependent facts are re-verified against the target Minecraft version
- [x] #7 No client jar is produced
- [ ] #8 Spec phase: Phase 1 — Version re-verification and toolchain (US1 groundwork)
- [ ] #9 Spec phase: Phase 2 — Wire client and vector proof (US1)
- [ ] #10 Spec phase: Phase 3 — Manifest (L-7) and token registry (US2 + US3)
- [x] #11 Spec phase: Phase 4 — Harness replacement, gates, wiki, board
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Operator ruling 2026-08-22 (recorded in sweep-0007-0022 runbook, per-task artifacts section): when this task introduces Gradle, the same PR replaces seam/java-roundtrip's hand-rolled JSON parsing with a library-based harness. The hand-roll existed only to avoid introducing a build system before V1 owned that decision; once Gradle exists, that justification is void. The canonical writer stays custom only if the chosen library provably cannot emit C-1..C-10 canonical form (verify against the vectors, record the outcome).

Claimed by sweep-0007-0022 orchestrator 2026-08-22 on branch task-0009-fabric-mod-skeleton (worktree .worktrees/task-0009). Tier: sonnet (default tier per runbook — mod skeleton to the written handshake contract; Yarn-mapping re-verification is verification, not design). Model ID: cc/claude-sonnet-5[1m]. Served model recorded at dispatch. Wire inputs ratified: decision-0004 (UDS, vendor dials), seam-wire-v0.md, seam/vectors/. Carries the 2026-08-22 operator ruling: this task's Gradle PR replaces the hand-rolled java-roundtrip parser.

Phase 1 done (fa30968, sonnet verified — claude-sonnet-5, ~150k subagent tokens): versions re-verified with citations (MC 26.2 stable, Loader 0.19.3, Fabric API 0.158.0+26.2, Loom 1.17.19, Gradle 9.5.1 wrapper); mod/ scaffold builds green, runServer load observed (kithcraft 0.1.0, entrypoint log fired), single server jar — no client jar. FINDING for V3/TASK-0014: Yarn is discontinued (MC 26.1+ unobfuscated) — villager-brain-api.md's yarn-1.21.3 symbols must be re-checked against Mojang official names before brain/Mixin work; routing A-2 daylight flag carried forward. Orchestrator re-ran gradle build green.

Phase 2 done (fb48d1c, sonnet verified — claude-sonnet-5, ~209k subagent tokens): Gson chosen (already on classpath via Minecraft; no new dep) — C-1..C-10 emit check FAILED for Gson output (no sorted keys, no integer-only numerics), so CanonicalJson.java pairs Gson JsonReader parsing with a custom canonical writer, recorded per the operator ruling's carve-out. FrameCodec + WireClient (UDS via StandardProtocolFamily.UNIX) + Handshake/Continuity built; 26/26 Gradle tests green incl. the 17-vector suite and a stub-mind handshake round-trip. Orchestrator re-ran gradle test green.

Phase 3 done (62f375a, sonnet verified — claude-sonnet-5, ~190k subagent tokens): real static manifest in Handshake.java (4-type floor, 6 origins, 4 core verbs, 4 world-independent salient_kinds, bearings, distance bands, time_unit second) — L-7 proven two ways (byte-compare across bodies/world-states + zero-arg builder makes world state structurally not an input); TokenRegistry pure core + SavedData adapter (MC 26.2 has no PersistentState — SavedData/SavedDataType is the current API, confirmed against the unobfuscated jar); live restart observation recorded (b-0/pl-1/th-2 resolved identically post-restart). Orchestrator re-ran build+test green (32/32) post-merge-reconcile with main. Branch reconciled with post-#13 main; runbook conflict resolved per doctrine.

Phase 4 done (sonnet, claude-sonnet-5[1m], this dispatch): T010 replaced seam/java-roundtrip's hand-rolled JSON parser with Gson-based decode + hand-written canonical encode (same split as mod/wire/CanonicalJson.java, same C-1..C-10 emit-check verdict); shape chosen: a seamRoundTrip Gradle task in mod/build.gradle supplies Gson from sourceSets.main.runtimeClasspath, RoundTrip.java stays a single JEP-330 source-launch file (rationale in seam/java-roundtrip/README.md: avoids duplicating the wrapper jar or hand-rolling dependency resolution in a script). Harness green: "java-roundtrip: 91 passed, 0 failed, over 17 vectors". Mutation-check power demonstrated against a /tmp copy (tracked seam/vectors/ never touched, confirmed via git status): corrupted frame_hex byte -> red (2 failures: meaning+bytes); decoded-only drift (edited only the decoded field) -> red (1 failure: meaning only, bytes still passed since it re-encodes the actual frame, not the tampered field) - recorded in the README.

T011 gates: ./gradlew build+test 32/32 green, seamRoundTrip green; scope check: git diff --name-only origin/main...HEAD touches only mod/, seam/java-roundtrip/, specs/009-fabric-mod-skeleton/, backlog/ (incl. an unrelated TASK-0014 card edit), docs/design/sweep-0007-0022-runbook.md, docs/wiki/ - clean, nothing flagged.

T012 wiki: villager-brain-api.md marked UNVERIFIED against MC 26.2 (Yarn discontinued per versions.md), re-derivation deferred to V3/TASK-0014, not performed here; body-protocol-seam.md's First-implementations section extended with the mod/ vendor (CanonicalJson.java, Handshake.java added as sources); overview.md and INDEX.md's vendor-side-still-pre-code claims amended. All three re-pinned to cb5003531192512f992688cbf5a08d7b15799545 (resolves, confirmed via git cat-file -e before writing). CAPSULES.md regenerated (overview.md's description changed).

Card AC #11 (Phase 4) checked; AC #1 (orchestrator's full-handshake-against-daemon judgment) left unchecked, DoD and status untouched per dispatch scope.
<!-- SECTION:NOTES:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Tests pass
- [ ] #2 Docs and wiki are updated and pass freshness tests
- [ ] #3 Spec and Backlog are in sync
<!-- DOD:END -->
