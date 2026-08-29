# Tasks: The job-board book and the blueprint build (V4)

**Spec dir**: `specs/020-job-board` · **Branch**: `task-0020-job-board`

## Phase 1 — The board and the read (US1)

- [x] T001 `dev.kithcraft.mod.board/`: designated lectern + written-book
      posting; free text is a valid blueprint — no form, no syntax (card ACs
      #1, #7). `Board`/`BoardData`/`BoardSetup` (`mod/src/main/java/dev/kithcraft/mod/board/`).
      **Deviation from plan.md's open choice**: picked the fixed-config-position
      option (not "first-placed convention") — no block-place event hook needed,
      matching `ScheduleSetup`/`DuskPairing`'s existing fixed-offset-from-origin
      idiom. Lectern deliberately NOT registered as a vanilla POI (it's librarians'
      workstation; would make it claimable). Accepts either an unsigned
      "book and quill" or a signed written book — no signing required, matching
      "no form, no syntax". Proof: `BoardTest` (5 tests).
- [x] T002 Read cycle on the schedule substrate (no new Activity, no Mixin):
      villager visits the board, read emits `text` percept with
      `origin: "read"` carrying the book text; structural test proves zero
      protocol extension (card ACs #2, #3). `BoardVisit`/`BoardReadTracker`
      ride `Activity.MEET` exactly as `DuskPairing` already does (board placed
      inside its arrival radius) — no new Activity, no new Mixin (mixin count
      unchanged at 4, `MixinConfigTest` still green). Content/provenance
      composed through the pre-existing `Testimony`/`PerceptEmitter` only.
      **Deviation**: like `DuskPairing`, the read is composed and logged
      (dev-server evidence) rather than pushed over a live multi-body wire
      session — `BodySession` is single-body scoped; multi-body session
      multiplexing is out of Phase 1 scope (ponytailed in `BoardVisit`'s
      javadoc, same as `DuskPairing`'s own precedent). Proof:
      `BoardReadTrackerTest` (4 tests, dedup), `BoardReadProtocolTest`
      (2 tests, structural zero-extension proof — content keys and envelope
      keys asserted identical to §4.7's pre-existing shape; Phase 1 doesn't
      touch the act surface, so there is no target shape to check yet —
      Phase 2's claim-intent target is where that check belongs).
- [x] T003 Board content includes visible claims (card AC #4 read half);
      gradle green. `Board.recordClaim`/`claims()` append claim lines to
      `readableContent()`; Phase 2 is what actually calls `recordClaim` from
      real claim registration. `gradle test`: 138 tests, 0 failures (11 new:
      `BoardTest` 5, `BoardReadTrackerTest` 4, `BoardReadProtocolTest` 2).

## Phase 2 — The claim (US2)

- [x] T004 Claim registration engine-side from a claim intent (AR-4
      token resolution; verb via manifest floor or declared extras — L-7
      holds); claim appears in board content (card AC #4).
      **Verb mechanism: declared-extras, not the manifest floor** — none of the
      core four (`go_to`/`speak`/`attend`/`wait`) fits "take this posting", so
      `claim` rides §5.5's manifest-declared-extras half exactly like V2's
      `percept_types` floor+5 already does (plan.md decision 3).
      `Handshake.MANIFEST`'s `verbs` list gains one static entry
      (`{"verb":"claim","targets":["thing"]}`) — still a plain constant, L-7
      unaffected (`p5ManifestIsVendorInvariantNotWorldDescribing` still green).
      **AR-4 token resolution**: the claim intent's target is
      `{"type":"thing","thing_id":<the board's own thing token>}` — the SAME
      token every read percept already carries in `provenance.source` (T002's
      `BoardVisit`); no new token type minted, no new target field
      (`dev.kithcraft.mod.act.TargetResolution` untouched — the pre-existing
      `"thing"` case already parses it). `dev.kithcraft.mod.act.ClaimRegistry`
      is the new seam (mirrors `TargetResolution.Ground`/`Verbs.Actuator`);
      `dev.kithcraft.mod.board.BoardClaims` is its real, `Board`-backed
      implementation, threaded into `IntentHandler` via `BodySession`.
      `Board.tryClaim` (package-private — see T006) implements first-accepted-
      claim-wins and appends the claim line `Board.recordClaim` already carries
      (Phase 1's hook, now actually called). Proof: `BoardClaimTest` (2),
      `IntentHandlerTest`'s two new claim tests, `BoardTest`'s five new
      `tryClaim`/`postNotice` tests.
      **Structural zero-protocol-extension, extended per Phase 1's own note**:
      `BoardClaimProtocolTest` (3 tests) proves the claim target carries no key
      beyond `type`/`thing_id`, resolves through the unmodified
      `TargetResolution.resolve`, and that `claim`'s declared target set is
      exactly `["thing"]` — no new target type.
- [x] T005 M5-real claim driving in tests: mind/deliberate's E3 loop (merged)
      produces the claim intent against the fake vendor in a cross-language
      fixture or an M5-shaped scripted intent with the shape asserted —
      record which honestly (card AC #9 context).
      **Recorded honestly: M5-shaped scripted intent, not a generated
      cross-language fixture.** mind/deliberate has no fixture-export path —
      its own tests (`mind/deliberate/board_test.go`) script raw JSON inline;
      building a Go→Java fixture pipeline was judged out of this phase's
      scope (a design decision, not a mechanical one — flagging rather than
      guessing). Instead `BoardClaimTest.m5ShapedClaimIntent` composes the
      exact payload shape `mind/seam/intents.go`'s `Pending.Compose` produces
      (`intent_id, verb, target, reason, supersedes, not_after`) from an
      `Intent` decoded by `mind/llm/structured.go`'s `llm.ParseIntent`
      (`verb, target, reason, supersedes` — no other field exists to add
      one), and copies the verb, target, and reason text **verbatim** from
      `mind/deliberate/board_test.go`'s own
      `TestLoop_E3ClaimIntent_CarriesAuthoredReason` (`{"verb":"claim",
      "target":{"type":"thing","thing_id":"th-post-1"},"reason":"Building
      shelters is exactly my trade..."}`) so a future reader can diff this
      Java test against the real Go test directly. Drives the real engine
      pipeline end to end (`IntentHandler` → `BoardClaims` → `Board`), not a
      Board-only unit test, and asserts the mind's authored `reason` is
      echoed unaltered on the `act_result` (§5.2).
- [x] T006 Structural absence: no force-claim path anywhere (card AC #9);
      V5's tend-grave posting seam closed — GraveBoardEntry rides this board
      for real.
      **No force-claim path**: `dev.kithcraft.mod.death.StructuralAbsenceTest`
      gains `noForceClaimPath` (V5's own grep-over-source style, extended
      rather than duplicated into a new file) — checked as call-site
      cardinality rather than a forbidden-word grep: `Board.tryClaim` is
      called from exactly one place in the whole mod source (`BoardClaims`).
      `Board.tryClaim` itself is package-private, so this is also a
      compile-time guarantee, not only a regression test.
      **V5 seam closed for real**: `Board.postNotice` (new, delegates to the
      same `recordClaim` line-append `Board` already had) is called from
      `LiveDeathHandling.handleDeath` with `GraveBoardEntry#content()`'s text
      — a grave posting now actually lands on the persisted `Board`, not only
      a log line. Required threading `BoardData` into
      `LiveDeathHandling.register` (reordered ahead of it in
      `KithcraftMod.onServerStarted`, since board setup now has to exist
      first) and persisting `Board`'s new `claimedBy` field
      (`BoardData`'s codec gains `claimed_by`). Proof:
      `BoardTest.postNoticeRidesTheSameReadableContentAsAClaim`;
      `GraveBoardEntryTest` unchanged (the class's own `content()`/`take()`
      contract didn't change, only who calls it).

      **Deviation**: T005's own honest fixture-vs-scripted-intent call above
      is the one recorded deviation from the phase's stated menu of options.
      Everything else in T004/T006 landed as designed with no plan.md
      deviation.

      **Gates**: `./gradlew test` — 151 tests, 0 failures (13 new since
      Phase 1's 138: `BoardClaimProtocolTest` 3, `BoardClaimTest` 2,
      `BoardTest` +5, `IntentHandlerTest` +2, `StructuralAbsenceTest` +1).
      Mixin surface unchanged (no new Mixin this phase — `ClaimRegistry`/
      `BoardClaims`/`Board.tryClaim` are all plain-JDK/vanilla-API classes).

## Phase 3 — The build (US3 + US4)

- [x] T007 Blueprint interpretation: tiny deterministic free-text parser
      (shape, size, material, generous defaults); unparseable text still
      claims/declines cleanly; heavily ponytailed (fidelity floor).
      `dev.kithcraft.mod.build.Blueprint`/`BlueprintParser`
      (`mod/src/main/java/dev/kithcraft/mod/build/`): 4 shapes (wall/hut/
      tower/pen, `fence`/`house` as synonyms), 4 materials (wood/stone/brick/
      dirt, `plank`/`cobble`/`mud` as synonyms), earliest-keyword-in-text
      wins per field, size regex-extracted and clamped 2-10. Only `null`/
      blank text is unparseable (`Optional.empty()`) — any non-blank text,
      however unrecognized, still parses to the generous defaults; whether
      the mind then claims-and-declines a gibberish posting is M5's own
      judgment, not this parser's. Heavily ponytailed (see class javadoc):
      ceiling is "recognizable enough to watch rise beside the player," not
      a construction game. Proof: `BlueprintParserTest` (6 tests).
- [x] T008 Build execution: block-by-block placement on a tick budget in work
      periods; material sourcing by the simplest defensible rule (recorded);
      no re-issue/supervise/hand-feed path (card ACs #5, #8).
      `dev.kithcraft.mod.build.Placement` (pure offset math, 4 shape
      generators, deterministic order — `PlacementTest`, 5 tests) +
      `BuildEngine` (pure tick-budget stepper, 1 block/tick — ponytailed pace
      — `BuildEngineTest`) + `LiveBuildExecution` (MC glue: finds the
      claimed villager via the existing body-token lookup, reads
      `Activity.WORK`, applies `BuildEngine`'s offsets to the world via
      `level.setBlockAndUpdate`). Rides the schedule substrate exactly as
      instructed: an activity-check per tick (`isActive(Activity.WORK)`),
      no new Activity, no Mixin (mixin count unchanged at 4).
      **Material sourcing (plan.md decision 5): creative-style provision —
      a placed block simply appears, nothing drawn from any inventory.**
      Chosen over village-storage withdrawal as the simpler of the plan's
      two options — a real storage-container scan/decrement is a whole
      sub-feature the "thinnest possible build system" constraint doesn't
      ask for; ponytailed in `LiveBuildExecution`'s class javadoc with the
      upgrade path named. **No re-issue/supervise/hand-feed path**: the only
      driver of `BuildEngine.advance` in the whole mod is
      `LiveBuildExecution`'s own tick call — checked as call-site
      cardinality, `StructuralAbsenceTest#noManualBuildProgressPath`
      (T006's own style, extended into the same file).
      **Single global build site, not one per villager** (`BuildSetup`,
      fixed-offset-from-origin idiom matching `BoardSetup`/`DuskPairing`):
      recorded call per the dispatch's instruction — `Board` models exactly
      one posting at a time (Phase 2's finding), so at most one claim, and
      therefore at most one build, is ever in progress; a single site is
      sufficient, not a corner cut. The single-posting model was kept as-is;
      it did not block this phase.
- [x] T009 Interrupt/resume: schedule transition or danger stops placement,
      persists cursor (SavedData); next work period resumes without
      re-claiming (card AC #6).
      `dev.kithcraft.mod.build.BuildCursor` (pure, persisted resume point —
      `claimantBody`/`blueprintText` snapshot/`placed` index) + `BuildData`
      (SavedData adapter, V1's idiom exactly — `BoardData`/`CastData`'s own
      codec-wrapper pattern, untested by that same convention). **No
      separate interrupt-detection code**: `BuildEngine.advance` takes one
      boolean (`working`) computed from `isActive(Activity.WORK)`; a dusk
      schedule transition and a danger-driven PANIC both simply make that
      boolean false (PANIC outranks WORK on the same vanilla schedule
      substrate — villager-brain-api.md's addActivity-additive facts), so
      one mechanism covers both interrupt kinds with zero extra branching.
      **Both interrupt kinds tested** per the card, deliberately as
      mechanically-identical test methods
      (`scheduleTransitionInterruptsPlacementAndCursorPersistsUnchanged`,
      `dangerPanicInterruptsPlacementAndCursorPersistsUnchanged`) —
      documented in `BuildEngineTest`'s class javadoc as intentional: the
      sameness IS the design. Resume proven by
      `resumesFromTheCursorWithoutRestartingOrReclaiming` (continues at the
      exact next offset, no claim-related argument exists to pass).
      Blueprint-snapshot-at-claim-time (spec Edge Cases: "board book
      removed/edited mid-build continues from its captured snapshot")
      falls out of `BuildCursor` capturing `blueprintText` once, at
      `start()` — `LiveBuildExecution` only reads live board text for a
      brand-new cursor, never for an existing one.
      **Deviation recorded**: spec Edge Cases' "unreachable build site:
      claim stands, build attempts pause and retry" does not apply to this
      design — placement is a direct world write (`setBlockAndUpdate`), not
      a villager path to each block, so there is no pathing failure mode to
      pause/retry against; ponytailed in `LiveBuildExecution`'s class
      javadoc as an edge case this engine-side-instant-placement design
      structurally sidesteps, not one it handles.
      **Gates**: `./gradlew build test` — 168 tests, 0 failures (17 new
      since Phase 2's 151: `BlueprintParserTest` 6, `PlacementTest` 5,
      `BuildEngineTest` 5, `StructuralAbsenceTest` +1). Mixin surface
      unchanged (no new Mixin this phase — `build/` is all plain-JDK/
      vanilla-API classes, same posture as `board/`).

## Phase 4 — The beat, gates, and closure

- [x] T010 Live dev-server observation: post → read → claim → build → dusk
      interrupt → resume, recorded in research/board-observation.md with
      honest not-observed records (card ACs #2, #5, #6 live halves).
      **Post/read/claim confirmed live** (book persisted across two restarts;
      board reads fired repeatedly for Aldric and Petra; a claim completed
      live in the prior session and this session reconfirmed its persistence
      + first-accepted-wins across a restart). **Build placement/interrupt/
      resume NOT observed** — root cause identified, not just timed out:
      `LiveBuildExecution.findClaimant` resolves the claimant against
      `DuskPairing`'s own per-cast seat-token map, but a live claim is always
      submitted under `BodySession`'s separate generic single-attach token —
      two disjoint token namespaces that structurally never match, so build
      can never start through the current live wiring regardless of
      `Activity.WORK` timing. Full evidence, log excerpts, and the exact
      code citations: `specs/020-job-board/research/board-observation.md`.
- [ ] T011 Full gates: gradle build + test green; scope clean
- [ ] T012 Wiki re-ground: touched-source notes re-verified honestly
      (overview — the mod gains the board/build surface; body-protocol-seam
      if seam-facing claims change; villager-brain-api if substrate facts
      grow); CAPSULES regenerated if descriptions changed; freshness green
- [ ] T013 Card ACs ticked with citing proofs; board/spec synced at PR time
