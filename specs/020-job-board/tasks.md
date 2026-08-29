# Tasks: The job-board book and the blueprint build (V4)

**Spec dir**: `specs/020-job-board` · **Branch**: `task-0020-job-board`

## Phase 1 — The board and the read (US1)

- [ ] T001 `dev.kithcraft.mod.board/`: designated lectern + written-book
      posting; free text is a valid blueprint — no form, no syntax (card ACs
      #1, #7)
- [ ] T002 Read cycle on the schedule substrate (no new Activity, no Mixin):
      villager visits the board, read emits `text` percept with
      `origin: "read"` carrying the book text; structural test proves zero
      protocol extension (card ACs #2, #3)
- [ ] T003 Board content includes visible claims (card AC #4 read half);
      gradle green

## Phase 2 — The claim (US2)

- [ ] T004 Claim registration engine-side from a claim intent (AR-4
      token resolution; verb via manifest floor or declared extras — L-7
      holds); claim appears in board content (card AC #4)
- [ ] T005 M5-real claim driving in tests: mind/deliberate's E3 loop (merged)
      produces the claim intent against the fake vendor in a cross-language
      fixture or an M5-shaped scripted intent with the shape asserted —
      record which honestly (card AC #9 context)
- [ ] T006 Structural absence: no force-claim path anywhere (card AC #9);
      V5's tend-grave posting seam closed — GraveBoardEntry rides this board
      for real

## Phase 3 — The build (US3 + US4)

- [ ] T007 Blueprint interpretation: tiny deterministic free-text parser
      (shape, size, material, generous defaults); unparseable text still
      claims/declines cleanly; heavily ponytailed (fidelity floor)
- [ ] T008 Build execution: block-by-block placement on a tick budget in work
      periods; material sourcing by the simplest defensible rule (recorded);
      no re-issue/supervise/hand-feed path (card ACs #5, #8)
- [ ] T009 Interrupt/resume: schedule transition or danger stops placement,
      persists cursor (SavedData); next work period resumes without
      re-claiming (card AC #6)

## Phase 4 — The beat, gates, and closure

- [ ] T010 Live dev-server observation: post → read → claim → build → dusk
      interrupt → resume, recorded in research/board-observation.md with
      honest not-observed records (card ACs #2, #5, #6 live halves)
- [ ] T011 Full gates: gradle build + test green; scope clean
- [ ] T012 Wiki re-ground: touched-source notes re-verified honestly
      (overview — the mod gains the board/build surface; body-protocol-seam
      if seam-facing claims change; villager-brain-api if substrate facts
      grow); CAPSULES regenerated if descriptions changed; freshness green
- [ ] T013 Card ACs ticked with citing proofs; board/spec synced at PR time
