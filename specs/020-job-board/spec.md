# Feature Specification: The job-board book and the blueprint build (V4)

**Feature Branch**: `task-0020-job-board` · **Spec dir**: `specs/020-job-board`

**Created**: 2026-08-28 · **Status**: Draft

**Input**: TASK-0020 / V4 — the demo's centrepiece, kept whole: the diegetic
board (book/lectern), the read-channel crossing (Q-6, no protocol extension),
claims visible to other villagers, and the thinnest possible engine-side build
execution (interruptible, resumable). Consumes kithcraft-brief.md (#7 diegetic
orders; tedium + micromanagement spell-breakers; loneliness-cure constraint),
body-protocol-v0.md (text/origin:read, Q-6 thin target, AR-4 token
resolution), decision-0002 (engine-side on the vanilla substrate), and M5's
merged claim behaviour (mind/deliberate, PR #22 — the one cross-lane
dependency, real not stubbed). Plan of record: demo-build-plan.md §3.3 V4.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - One gesture posts an order (Priority: P1)

As a player, I want to write a blueprint into a book on a lectern and have that
be the whole interface, so that giving work feels like asking a person.

**Independent Test**: A player-written book on the board lectern becomes a
posting; reading it emits a `text` percept with `origin: "read"` carrying the
blueprint as text over the seam — no protocol extension anywhere (card ACs #1,
#2, #3).

**Acceptance Scenarios**:

1. **Given** a lectern designated as the board, **When** the player writes a
   book and places it, **Then** the posting exists with no form, syntax, or
   phrasing requirement — free text is a valid blueprint (tedium check,
   card AC #7).
2. **Given** a villager near the board on its schedule, **When** it reads,
   **Then** a `text` percept with `origin: "read"` crosses the seam carrying
   the book's text; the wire payload uses only existing v0 shapes (structural
   test: no new percept/target fields) (card ACs #2, #3).
3. **Given** claims already made, **Then** the board's readable content
   includes other villagers' visible claims (card AC #4's read half).

### User Story 2 - The claim is the villager's (Priority: P1)

As a villager, I want taking a posted job to be my own deliberation's outcome,
so that my work is a favor, not a command executed.

**Independent Test**: M5's claim intent (merged, real) drives the claim; the
claim becomes visible on the board to other villagers; no code path forces a
claim (card ACs #4, #9).

**Acceptance Scenarios**:

1. **Given** the read percept reaching an attached mind, **When** the mind's E3
   deliberation returns a claim intent (M5's real machinery — the fake vendor
   test drives it with mind/deliberate's loop, not a stub), **Then** the claim
   registers engine-side and appears in the board's readable content.
2. **Given** the whole mod surface, **Then** no command, API, or code path
   exists by which the player forces a claim (structural absence test,
   card AC #9).
3. **Given** a claim intent naming the posting by token, **Then** AR-4 holds:
   the mind names a token, the engine resolves it.

### User Story 3 - Built beside the player, without supervision (Priority: P1)

As a player, I want the claimed build to proceed block by block while I build
alongside, so that the work is company, not a cutscene.

**Independent Test**: A claimed blueprint is built block by block engine-side
with material sourcing; once claimed it needs no re-issuing, supervising, or
hand-feeding (card ACs #5, #8).

**Acceptance Scenarios**:

1. **Given** a claimed posting, **When** the work period arrives, **Then** the
   villager builds block by block at the site (engine-side placement per
   decision-0002; deliberately the thinnest build system that can stand
   beside the player).
2. **Given** materials, **Then** sourcing is engine-side and self-directed (no
   player hand-feeding path; micromanagement check, card AC #8).
3. **Given** the blueprint text, **Then** the engine's interpretation is
   deliberately simple (a small vocabulary of inferable shapes/materials from
   free text — the fidelity floor the demo sheds first per the card); a
   posting the engine cannot interpret still supports claim-and-decline
   (the mind's reason can say why).

### User Story 4 - Interrupted at dusk, resumed at work (Priority: P2)

As a villager, I want dusk to end my workday mid-wall, so that my life is not
suspended for a job.

**Independent Test**: Interrupting at dusk (schedule transition) or danger
leaves a partial build that resumes the next work period (card AC #6).

**Acceptance Scenarios**:

1. **Given** an in-progress build at a schedule transition, **Then** building
   stops, the partial state persists, and the villager follows its schedule.
2. **Given** the next work period, **Then** the build resumes from the partial
   state without re-claiming or player action.
3. **Given** danger (panic), **Then** the same interrupt/resume shape holds.

### Edge Cases

- Two villagers read the same posting: first accepted claim wins engine-side;
  the loser's next read shows the claim (argument material, not a race bug).
- Claimed villager dies: V5's machinery already covers remains; the posting
  reverts to unclaimed on the next board read cycle (claim holder gone).
- Board book removed/edited mid-build: the claimed build continues from its
  captured blueprint snapshot; new reads see the new content.
- Unreachable build site: claim stands, build attempts pause and retry within
  the work period; recorded honestly if observed live.

## Requirements *(mandatory)*

- **FR-001**: Diegetic board (book/lectern), free-text blueprint, no syntax.
- **FR-002**: Read emits text/origin:read with the blueprint as text; zero
  protocol extension (structural test).
- **FR-003**: Claims visible to other villagers through board content.
- **FR-004**: Claim driven by M5's real deliberation machinery in tests (fake
  vendor + mind/deliberate loop), never a stub claim.
- **FR-005**: No force-claim path (structural absence).
- **FR-006**: Engine-side block placement + material sourcing against the
  blueprint; interruptible by schedule/danger; resumable; partial state
  persisted.
- **FR-007**: No new Mixins unless the substrate genuinely requires one — the
  bound is decision-0002's ceiling, already at 4: if a Mixin is needed, STOP
  and surface (escalation, runbook checkpoint 6).
- **FR-008**: Dev-server proof per the runbook's gate: the full beat observed
  live (post → read → claim → build → dusk interrupt → resume), recorded in a
  research observation doc; stub mind acceptable for the seam, M5's loop for
  the claim logic in unit/integration tests.

## Success Criteria *(mandatory)*

- All nine card ACs demonstrated: gradle build + test green; live observation
  doc for the full beat; honest not-observed records where live proof falls
  short.
- Wiki notes whose sources this PR touches re-verified honestly; CAPSULES
  regenerated if descriptions change; board/spec in sync at PR time.
