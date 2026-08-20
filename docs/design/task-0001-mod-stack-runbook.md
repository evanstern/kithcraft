# TASK-0001 mod-stack decision — sweep runbook (2026-08-19)

**You (the session reading this) are the ORCHESTRATOR** for the tasks below. Run each
through the host project's full PDLC — spec → link → worktree → delegated implementation →
PR → merge → re-ground — parallelizing within lanes, merging serially, treating merge
conflicts as routine. Direction is decided; do not re-litigate it:
`docs/design/kithcraft-brief.md` (ratified 2026-08-19) wins. Plan-of-record is the board;
this file carries only ordering, doctrine, and the log.

**Status:** signed-off · operator sign-off on lanes: 2026-08-19 (sonnet tier confirmed, no escalation)
<!-- Only the OPERATOR flips draft → signed-off (the author never pre-fills it). An
     executing session must refuse a runbook whose status it cannot verify. -->

## Read first (in this order)

1. `docs/design/kithcraft-brief.md` — the ratified design brief that produced this task.
2. Project `CLAUDE.md` — PDLC grounding, model-tier posture, task conventions.
3. `backlog task list --plain` — live state; other sessions move it while you work.
4. The task you're about to execute (`backlog task view TASK-0001 --plain`).

## State when this runbook was written (2026-08-19)

- **Done already:** none — this is the project's first sweep. Board seeded with
  TASK-0001..0006 under milestone m-0 "One real evening (v1 demo)".
- **In flight in other sessions (do not duplicate; expect their merges):** none known.
- **Paused — untouched:** none.
- **Queued (this runbook's scope):** TASK-0001 only.

## Execution lanes (dependency-ordered; parallelize within a lane)

Single-task sweep; one lane.

**Lane 1 — start immediately:**
- **TASK-0001 (sonnet · model `cc/claude-sonnet-5[1m]`, no fallback declared for this
  tier — default tier: the spec settles the evaluation criteria and the comparison is
  research executed against them; the actual judgment call, ratifying the
  recommendation, is an operator checkpoint (AC #2) regardless of tier)** — evaluate
  Fabric vs Paper/Citizens2 vs hybrid, re-verify prior art, write the comparison +
  recommendation, record a Backlog decision record. This task's output is a decision
  document, not code; its "contract" (the ratified stack choice) unblocks TASK-0002
  and TASK-0003.

Tiers and their model IDs come from **`.claude/model-tiers.json`** — `tiers.mjs --root .
--check` exited 0 on 2026-08-19 before these lanes were authored (all three agent
definitions `unchanged`; no regeneration, so no session restart required). Escalation to
opus (`cc/claude-opus-5[1m]`, fallback `cc/claude-opus-4-8[1m]`) is available if the
operator prefers the comparison itself run at the thinking tier — that is an operator
checkpoint recorded before dispatch.

Record the model tier + explicit model ID + rubric justification on the board task at
dispatch, including which model actually served (verify from the first dispatch's
transcript before any sibling dispatch).

## Per-PR gates this project enforces (enumerated — implementers cannot miss these)

- **Merge-drift gate: absent.** No `scripts/check-merge-drift.mjs` in this repo; raw
  git discipline stands: `git fetch origin && git pull --ff-only` at root before
  claim/worktree/merge; verify merged (`gh api repos/{owner}/{repo}/pulls/{n} --jq
  .merged`) before deleting branches/worktrees.
- **No wiki yet:** `docs/wiki/` does not exist (project pre-dates its first
  `/grounding-wiki:wiki-build`). The board DoD line "Docs and wiki are updated and pass
  freshness tests" is satisfied for this task by the docs the task itself produces;
  there are no pins to re-verify and no freshness probe to run. First wiki build is
  future work, out of this sweep's scope.
- **Root-guard hook** (`.claude/hooks/root-guard-hook.mjs`) enforces worktree
  discipline on Bash/Write — work on the task branch happens in the worktree only.
- **Prior-art re-verification is evidence-dated:** every claim about a dependency's
  version/maintenance/license in the comparison must carry the date it was checked and
  a URL (task AC #1).

## Per-task artifacts required before PR

Per-TASK obligations. **No PR opens for a task until each line below checks true for
it.** The sweep's Output gate re-checks the first two lines at the end.

- [ ] `specs/001-mod-stack-decision/` carries a real `spec.md` (problem + requirements
      mapped to the card's ACs), `plan.md` (the constitution at
      `.specify/memory/constitution.md` is an unfilled template — plan.md must state
      that plainly and plan against the grounding docs: the brief + CLAUDE.md), and
      `tasks.md` (phased checkboxes the bridge derives from), committed on the task's
      branch. A claim stub reserves the number; it satisfies nothing here.
- [ ] The card carries its Spec marker from the claim commit (`spec-bridge:link`
      against the stub), and phase ACs are seeded from tasks.md (link update mode)
      before implementation dispatch.
- **Escape lines (operator-signed only):** none.
- **Host additions:** the decision itself lands as a Backlog decision record
  (`backlog decision create …`) in the same PR — the card's AC #2 names it; a
  recommendation living only in the comparison doc does not satisfy it. Operator
  ratification of the recommendation is a checkpoint after PR review, before the task
  is synced Done.

## Concurrency & conflict doctrine

- **Hotspots:** `backlog/tasks/*` (any concurrent session touching the board),
  `docs/design/*` (this runbook's own log updates). Low risk — this is currently the
  only known session.
- Reconcile by what the branch carries: no wiki exists, so branches are **pin-free**
  and rebase onto fresh `origin/main`; take main's side for anything not deliberately
  changed.
- **Claim before work:** the FIRST commit of the task branch claims it — board card →
  In Progress, spec dir stub, spec-bridge link — pushed immediately
  (`git push -u origin <branch>`); never force-push a claim. A rejected push means the
  race was lost: fetch, re-read board and `specs/`; another holder → STOP the lane and
  surface to operator.
- Verify a PR merged before deleting its branch/worktree; never delete+recreate a
  closed PR's head.

## Operator checkpoints (do not proceed silently)

1. **Sign-off on this runbook's lane** (pending — required before execution).
2. **Tier escalation** if the operator wants the comparison run at opus instead of
   the sonnet default (see Lane 1 note).
3. **Ratifying the recommendation** (task AC #2): the sweep produces the comparison,
   recommendation, and decision record; the operator ratifies before spec-bridge:sync
   moves the card to Done. This is the task's built-in one-way door — TASK-0002/0003
   build on the ratified choice.

## Done means

TASK-0001 Done on the board via one merged PR; `specs/001-mod-stack-decision/` carries
real spec.md + plan.md + tasks.md; the card still carries its Spec marker; the
comparison doc and Backlog decision record exist; the recommendation is
operator-ratified; no stale worktrees; this file's log complete and status flipped to
done.

## Execution log

| date | task | PR | merge | tokens/cost (best-effort) | notes |
|------|------|----|-------|---------------------------|-------|
| 2026-08-19 | TASK-0001 | — | — | — | in flight: claimed (main aae1b84), spec cycle committed on branch (b591b7d, pushed), phase ACs seeded (main d77d20e). Phase 1 dispatch BLOCKED: agent registry predates this session's tier agents ("sonnet-implementer not found") — session restart required before any dispatch. Resume: verify branch task-0001-mod-stack-decision + worktree .worktrees/task-0001 intact, then dispatch Phase 1 per tasks.md at sonnet (cc/claude-sonnet-5[1m]); verify served model from first transcript. |
| 2026-08-20 | TASK-0001 | — | — | — | resumed: registry has tier agents, tiers --check green (all unchanged), worktree+branch intact, origin/main merged in (569ee7c; spec.md add/add resolved to branch's real spec). Root is read-only (root-guard) so runbook/board bookkeeping rides this branch. phases: 1 done (f99016d, served model VERIFIED cc/claude-sonnet-5[1m] from transcript, ~114k tokens; AI_NPC license unverifiable — documented dead end), 2 dispatched. |
