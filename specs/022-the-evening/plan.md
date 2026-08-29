# Implementation Plan: The evening (I2)

**Spec dir**: `specs/022-the-evening` · **Branch**: `task-0022-the-evening`

**Constitution check**: `.specify/memory/constitution.md` is an unfilled template —
stated plainly per runbook gate. Planned against the grounding docs:
kithcraft-brief.md, demo-build-plan.md §§3.4/4/5, llm-routing-and-budget.md
(A-1..A-8, ceilings), demo-runbook.md (I1's bring-up), `docs/wiki/`
([[v1-demo]], [[overview]]).

## Execution model

Two halves with an operator checkpoint between them (runbook checkpoint 5):

1. **Staging (sweep, Phase 1)** — author the run kit as documents; verify
   prerequisites; compile the watch list from the merged tasks' honest flags.
   Pure documentation work at this branch; no code.
2. **The evening (OPERATOR, Phase 2)** — ~3 h with the player. The sweep
   stops; the operator schedules and runs it with the kit, capturing logs,
   the session report, and the recording.
3. **Write-up (sweep, Phase 3, post-run)** — findings doc from the operator's
   captured run; A-n annotations from session-report.log; failures carded;
   gates/wiki/board close. Dispatched only after the operator hands back the
   run artifacts.

## Design decisions

1. Kit lives in `specs/022-the-evening/` (run-kit.md holding all three
   checklists + the measurement sheet); the post-run findings doc lands as
   `docs/design/evening-findings.md` per the runbook host-additions line.
2. The watch list is compiled from the board's own honest flags (TASK-0014
   signal lead + JOB_SITE; TASK-0020 build-placement/WORK timing; TASK-0021
   reconnect identity + heartbeat admissibility) — the run watches for known
   issues rather than rediscovering them.
3. Live genesis: the operator's key; personas may pre-exist from TASK-0013's
   live run — demo-runbook.md's re-bind path covers both.
4. No A-n is measured by an agent in a simulated run — the numbers come from
   the operator's real session only (FR-003).

## Phase map

Phase 1 — stage the run kit, verify prerequisites, compile the watch list (sweep).
Phase 2 — the evening (OPERATOR — checkpoint 5; the sweep is stopped here).
Phase 3 — findings doc, A-n annotations, failure carding, gates, wiki, close (sweep, post-run).
