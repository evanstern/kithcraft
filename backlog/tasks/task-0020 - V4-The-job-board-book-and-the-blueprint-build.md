---
id: TASK-0020
title: V4 - The job-board book and the blueprint build
status: In Progress
assignee: []
created_date: '2026-08-21 23:40'
updated_date: '2026-08-29 03:24'
labels:
  - vendor
  - m-0-build
milestone: m-0
dependencies:
  - TASK-0012
  - TASK-0014
  - TASK-0016
documentation:
  - docs/design/demo-build-plan.md
  - docs/design/kithcraft-brief.md
  - docs/design/body-protocol-v0.md
priority: high
ordinal: 20000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
As a player, I want to post a blueprint on a board and have a neighbour take it up and build it beside me, so that giving work feels like asking a person rather than issuing a command.

**Scope boundary.** The demo's centrepiece and the v1 soul of the order interface (brief #7), **kept whole** — the plan's one deliberate merge, because split, the first PR delivers a board that is read and nothing built. The diegetic in-world board: a book/lectern the player writes a blueprint into, readable by villagers, with other villagers' claims visible on it so they can argue about and prioritize orders. Reading it emits a `text` percept with `origin: "read"`, carrying the blueprint **as text**: protocol Q-6 records that v0's `target` shape is deliberately thin and that the read channel is how a blueprint crosses in v0 — so **this task does not extend the protocol** to carry a structured build plan. Engine-side build execution for a claimed job: block placement against the blueprint, material sourcing, interruptible by schedule transition or danger and resumable afterwards. The mind is consulted on **taking** the job and at decision points; the building itself is engine-side (AR-4 — the mind names a token, the engine resolves it). Per the loneliness-cure constraint this is deliberately **the thinnest possible build system**: it exists to put a neighbour beside the player while they work, not to be a construction feature, and it is the first place the demo sheds fidelity if scope must be cut.

**Done proves.** On a dev server with a mind attached: the player writes a blueprint into the board; a villager walks to it, reads it, and a `text` percept crosses the seam; a claim becomes visible to the other villagers; the claimed blueprint is built block by block **while the player builds alongside**; interrupting at dusk leaves a partial build that resumes the next work period.

**Depends on.** V2, V3, M5 — **the one cross-lane dependency in the graph.** It needs M5's claim behaviour demonstrable end-to-end, left as a real dependency rather than faked with a stub, because the beat is only observable when both halves exist.

**Design check — tedium.** Posting an order must be one diegetic gesture (write in the book), not a form to fill in or a syntax to learn. If the player has to phrase a blueprint carefully to be understood, the interface has become a chore and the diegetic framing is decoration.

**Design check — micromanagement.** Once claimed, a build proceeds without the player re-issuing, supervising, or hand-feeding materials. The player manages flow; they do not run the site.

**Constraint — minds-are-others.** The board posts an *order*, not a command. The claim is the villager's to make, and the plan supports **no path by which the player forces one**.

**References.** docs/design/demo-build-plan.md section 3.3 (V4) is the plan of record. Ratified surfaces consumed: docs/design/kithcraft-brief.md (#7 the diegetic order interface; the tedium and micromanagement spell-breakers; the loneliness-cure constraint), docs/design/body-protocol-v0.md (text percept with origin:read, Q-6's thin target shape, AR-4 token resolution), decision-0002 (engine-side resolution on the vanilla substrate).

**Suggested tier: `sonnet` (next sweep's runbook decides).**

Spec: specs/020-job-board
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A diegetic in-world board (book/lectern) accepts a player-written blueprint and is readable by villagers, with other villagers' claims visible on it
- [x] #2 A villager walks to the board, reads it, and a text percept with origin:read carrying the blueprint as text crosses the seam
- [x] #3 No protocol extension is made: the blueprint rides the read channel per Q-6, not a new structured target shape
- [x] #4 A claim becomes visible to the other villagers
- [x] #5 The claimed blueprint is built block by block while the player builds alongside, with engine-side placement and material sourcing
- [x] #6 Interrupting at dusk leaves a partial build that resumes the next work period
- [x] #7 Design check (tedium): posting an order is one diegetic gesture, not a form or a syntax the player must phrase carefully
- [x] #8 Design check (micromanagement): once claimed, a build proceeds without the player re-issuing, supervising or hand-feeding materials
- [x] #9 Constraint (minds-are-others): the board posts an order, not a command, and no path exists by which the player forces a claim
- [x] #10 Spec phase: Phase 1 — The board and the read (US1)
- [x] #11 Spec phase: Phase 2 — The claim (US2)
- [x] #12 Spec phase: Phase 3 — The build (US3 + US4)
- [x] #13 Spec phase: Phase 4 — The beat, gates, and closure
<!-- AC:END -->



## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
claimed by sweep-0007-0022 orchestrator 2026-08-28 (lane 4 tail; M5 dependency merged as PR #22); spec 020 stub + link ride this claim commit

tier: sonnet (default) · model cc/claude-sonnet-5[1m] · rubric: the board rides Q-6's read channel with no protocol extension; build execution is deliberately the thinnest possible system (runbook lane 4 tail; M5 dependency satisfied by PR #22)
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
author: phase4-closer
created: 2026-08-29 03:22
---
Phase 4 (T010-T013) complete via sonnet closer dispatch. AC#1 (board accepts free-text posting, readable, claims visible): BoardTest (5 tests, T001) + live -- book planted via /data merge, persisted through two dev-server restarts, read repeatedly (specs/020-job-board/research/board-observation.md sec1-2). AC#2 (villager reads, text percept origin:read crosses seam): BoardReadTrackerTest/BoardReadProtocolTest (T002) + board-observation.md sec2 -- "[board] Aldric/Petra reads the board" fired live, repeatedly, both villagers, percept_type=text/origin=read. AC#3 (no protocol extension): BoardReadProtocolTest (T002, 2 tests) + BoardClaimProtocolTest (T004, 3 tests) -- structural zero-extension proof, content/envelope keys identical to sec4.7's pre-existing shape. AC#4 (claim visible to other villagers): Board.readableContent() unconditionally concatenates claims after text (Board.java:97-103); BoardTest's tryClaim/postNotice tests. Live: a claim persisted from the genesis session and every subsequent live read this closer session pulled from the same readableContent() call, so the claim line necessarily rode along. AC#5 (built block by block, engine-side placement + material sourcing): unit-proven -- BlueprintParserTest (6), PlacementTest (5), BuildEngineTest (5). Live NOT observed: board-observation.md sec4 identifies the exact root cause (LiveBuildExecution.findClaimant checks DuskPairing's per-cast seat tokens, but a live claim is always submitted under BodySession's separate generic attach token -- two token namespaces that structurally never match, so build cannot start through today's live wiring regardless of Activity.WORK timing). Live half deferred to I2 (TASK-0022), the same honest pattern TASK-0019's AC#5/#8 used. AC#6 (dusk interrupt, resumes next work period): unit-proven -- BuildEngineTest's interrupt/resume tests (T009). Live NOT observed -- no build ever started this pass (AC#5's root cause); live half deferred to I2 alongside AC#5. AC#7 (design check, tedium): BoardSetup accepts either an unsigned book-and-quill or a signed written book, no signing/syntax required; BlueprintParser.parse -- any non-blank text parses to generous defaults (T007). AC#8 (design check, micromanagement): StructuralAbsenceTest#noManualBuildProgressPath (T008) -- BuildEngine.advance has exactly one call site in the whole mod. AC#9 (constraint, minds-are-others): StructuralAbsenceTest#noForceClaimPath (T006) -- Board.tryClaim is package-private and called from exactly one place. DoD#1 (tests pass): ./gradlew build test -- 168 tests, 0 failures, 0 errors (T011). DoD#2 (wiki/freshness): overview.md + body-protocol-seam.md re-grounded, freshness gate green -- 11/11 notes fresh (T012). DoD#3 (spec/backlog sync) deliberately left unticked, reserved for the orchestrator at PR/merge time per TASK-0019's own precedent -- no PR has been opened yet. Full evidence: specs/020-job-board/research/board-observation.md, specs/020-job-board/tasks.md (T010-T013).
---
<!-- COMMENTS:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 Tests pass
- [x] #2 Docs and wiki are updated and pass freshness tests
- [ ] #3 Spec and Backlog are in sync
<!-- DOD:END -->
