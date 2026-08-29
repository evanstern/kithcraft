# Feature Specification: The evening — run it, measure it, check it against the brief (I2)

**Feature Branch**: `task-0022-the-evening` · **Spec dir**: `specs/022-the-evening`

**Created**: 2026-08-29 · **Status**: Draft

**Input**: TASK-0022 / I2 — the ~3-hour operator-run evening at 1x: walk the
beat checklist (plan §4's coverage map), walk the spell-breaker checklist
(§5.2), replace every A-n assumption with its measured value, and record which
dusk landed best. Consumes kithcraft-brief.md (v1-demo beats, three
spell-breakers), decision-0003 + llm-routing-and-budget.md (A-1..A-8, ~$20
ceiling, E4 < 3 s ceiling, E6 ~80/day assumption + ~150 upgrade trigger).
Plan of record: demo-build-plan.md §3.4 (I2), §4, §5.

**EXECUTION MODEL (runbook checkpoint 5)**: the run itself is OPERATOR-RUN —
~3 hours real time with the player present. The sweep prepares (checklists
staged, findings template ready, prerequisites verified) and STOPS. After the
operator's run, the write-up phase turns the recorded run into the findings
doc and closes the card.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The run is a decision about when (Priority: P1)

As an operator, I want everything staged so that running the evening is
sitting down with a checklist, not assembling one.

**Independent Test**: The run kit exists — the beat checklist, the
spell-breaker checklist, the A-n measurement sheet with capture instructions,
and the findings-doc template — each traceable to its plan-of-record section
(card ACs #2, #3, #4 staging halves).

**Acceptance Scenarios**:

1. **Given** plan §4's coverage map, **Then** the beat checklist enumerates
   all seven beats with, per beat, what to look for and which log/console
   evidence captures it.
2. **Given** §5.2, **Then** the spell-breaker checklist walks tedium,
   micromanagement, and politeness-policing with their concrete per-task
   checks and the "does the player start walking past?" test named.
3. **Given** A-1..A-8 and the ceilings, **Then** the measurement sheet maps
   each assumption to its capture source (session-report.log per-class
   counters, E6-input instrument, E4 latency instrumentation) and its
   threshold (~$20 ceiling, < 3 s, ~80/day + ~150 trigger).
4. **Given** I1's demo-runbook.md, **Then** prerequisites are verified
   current (build green at this branch, key needed for live genesis,
   dangerTuning documented as available for the recorded take).

### User Story 2 - The evening, run and recorded (Priority: P1) — OPERATOR

As a player, I want one real evening with my villagers, so that the demo's
claims are observations.

**Independent Test**: A recorded ~3-hour session at 1x, player present, three
villagers alive (card AC #1); every beat observed (card AC #2); no
spell-breaker fails (card AC #3).

**Acceptance Scenarios** (walked by the operator with the staged kit):

1. Personas/desires · schedules · persistent memory · the job-board book ·
   the blueprint build alongside · dusk conversation · night danger — each
   observed and marked with evidence.
2. Each spell-breaker walked; failures written down, never silent.
3. The best dusk noted for the highlight cut (card AC #6; R-1's recording
   framing — cut, not engineered).

### User Story 3 - Assumptions become numbers (Priority: P1)

As a future planner, I want every A-n annotated with its measured value, so
that the next budget is derived from reality.

**Independent Test**: The findings doc annotates calls/class, tokens/call,
total cost vs the ~$20 ceiling, E4 first-token latency vs 3 s, E6 input
tokens/villager-cycle vs ~80 (and the ~150 trigger), each with its measured
number beside the assumed one (card AC #4).

### User Story 4 - Failures get owners (Priority: P2)

As the board, I want every failed check written down with the task that owns
the fix, so that the run finds problems without failing silently.

**Independent Test**: The findings doc's failure section lists each failed or
unobserved check with an owning task (existing card or a named new one for
refactor-triage) (card AC #5). The known pre-run candidates are pre-listed so
the operator watches for them: the mod-side reconnect identity gap
(TASK-0021's flag), the live build-placement/Activity.WORK question
(TASK-0020's observation), the live signal lead 1.82–4.96 s (TASK-0014's
caveat), JOB_SITE claim flakiness (TASK-0014).

### Edge Cases

- A villager dies mid-run: the run continues — V5's machinery is itself a
  beat (night danger); "three villagers alive" (AC #1) reads per the brief as
  the cast surviving to the evening's end being the aspiration, a death being
  finding material, not run failure. Judged at write-up against the recording.
- The run aborts (crash, schedule conflict): partial findings are still
  findings; the run reschedules — the kit is reusable.
- Cost runs hot mid-evening: the session report is end-of-run; the operator
  may spot-check the run log; the ~$20 ceiling is ~4x baseline headroom.

## Requirements *(mandatory)*

- **FR-001**: The run kit (checklists + measurement sheet + findings
  template) staged under this spec dir, each item traceable to its
  plan-of-record section.
- **FR-002**: Prerequisites verified: build green, demo-runbook current, the
  known-issues watch list compiled from merged tasks' honest flags.
- **FR-003**: The run is operator-executed; the sweep stops at the staged kit
  (checkpoint 5). No agent simulates or substitutes for the evening.
- **FR-004**: Post-run: the findings doc (docs/design/, per the runbook's
  host-additions line) with every beat, every spell-breaker, every A-n
  annotated, every failure owned.
- **FR-005**: No code changes on this branch unless the operator's run
  requires a staging fix; findings that need code get owning tasks instead.

## Success Criteria *(mandatory)*

- Pre-run (the sweep's half): kit staged, prerequisites verified, watch list
  compiled — then STOP for the operator.
- Post-run (after the operator's evening): findings doc lands in the PR; all
  six card ACs resolved honestly; board/spec in sync.
