# TASK-0004..0006 mind-routing, death design & demo plan — sweep runbook (2026-08-21)

**You (the session reading this) are the ORCHESTRATOR** for the tasks below. Run each
through the host project's full PDLC — spec → link → worktree → delegated implementation →
PR → merge → re-ground — parallelizing within lanes, merging serially, treating merge
conflicts as routine. Direction is decided; do not re-litigate it:
`docs/design/kithcraft-brief.md` (ratified 2026-08-19), decision-0001 (Fabric, accepted
2026-08-20), the body protocol v0 (PR #6, contract accepted 2026-08-21), and decision-0002
(augmented vanilla villager, accepted 2026-08-21) win. Plan-of-record is the board; this
file carries only ordering, doctrine, and the log.

**Status:** signed-off · operator sign-off on lanes: 2026-08-21
<!-- Sign-off ruling (2026-08-21): lanes approved as authored; BOTH proposed opus
     escalations TAKEN by the operator — TASK-0004 and TASK-0006 run opus
     (cc/claude-opus-5[1m]); TASK-0005 stays sonnet. The escalation checkpoints the
     tier rubric requires are this ruling. Ratification gates (checkpoints 2–4)
     remain the operator's at each PR. -->
<!-- Only the OPERATOR flips draft → signed-off (the author never pre-fills it). An
     executing session must refuse a runbook whose status it cannot verify. -->
<!-- This draft rides the TASK-0004 branch per the operator's 2026-08-21 no-runbook-PR
     ruling: runbook commits ride task branches / the board track, never their own PR. -->

## Read first (in this order)

1. `docs/design/kithcraft-brief.md` — the ratified design brief that produced these tasks.
2. `docs/wiki/CAPSULES.md` — corpus rollup; then just-in-time full notes:
   [[promptworld-lineage]] + [[body-protocol-seam]] + [[model-tiers]] for TASK-0004,
   [[design-brief]] + [[v1-demo]] for TASK-0005,
   [[v1-demo]] + [[villager-brain-api]] + [[pdlc-process]] for TASK-0006.
3. Project `CLAUDE.md` — PDLC grounding, model-tier posture, task conventions.
4. `backlog task list --plain` — live state; other sessions move it while you work.
5. The task you're about to execute (`backlog task view TASK-<n> --plain`).

## State when this runbook was written (2026-08-21)

- **Done already:** TASK-0001 (Fabric, decision-0001), TASK-0002 (body protocol v0,
  PR #6), TASK-0003 (entity decision, decision-0002, PR #5). Wiki 11 notes, freshness
  green at cc72b67.
- **In flight in other sessions (do not duplicate; expect their merges):** none known.
- **Paused — untouched:** none.
- **Queued (this runbook's scope):** TASK-0004, TASK-0005 (lane 1, parallel);
  TASK-0006 (lane 2, after lane 1 merges).

## Execution lanes (dependency-ordered; parallelize within a lane)

Rule of thumb: DEVELOP in parallel, MERGE serially. TASK-0004 and TASK-0005 are
parallel-safe: 0004 is mind-layer architecture behind the seam (deps: TASK-0002, Done);
0005 is gameplay design with no deps. Their spec dirs and primary design docs are
disjoint; shared surfaces are the board dir and this file (see hotspots). TASK-0006
formally depends on TASK-0001..0004 and consumes lane 1's ratified decisions — it is the
lane-2 task and also the sweep's tail: its output is the NEXT sweep's input.

**Lane 1 — start immediately, in parallel:**

- **TASK-0004 (opus · model `cc/claude-opus-5[1m]`, fallback `cc/claude-opus-4-8[1m]` —
  ESCALATED by operator at sign-off 2026-08-21: mind-daemon language
  reuse-vs-rebuild and LLM routing/budget are coupled architecture decisions the spec
  cannot pre-settle — design work per the rubric, directly analogous to TASK-0002's
  escalation. The final human gate — ratifying the decisions (card AC #4) — remains the
  operator's regardless of tier)** — decide mind daemon language + LLM routing/budget
  sketch for 3–6 villagers at 1x. Spec 004. **Contract-shaped: its ratified
  language/routing decisions are inputs TASK-0006's plan consumes** — it merges first
  if both lane-1 PRs are ready together.
- **TASK-0005 (sonnet · model `cc/claude-sonnet-5[1m]`, no fallback declared for this
  tier — default tier: the posture (permadeath real, graves/memories/stories) is
  already ratified in the brief; this pass fleshes mechanics out within it, analogous
  to TASK-0003's comparison which ran sonnet; the judgment call — ratifying the design
  (card AC #4) — is an operator checkpoint regardless of tier)** — death-mechanics
  design doc: causes, remains, how survivors remember; micromanagement spell-breaker
  addressed; shrinking-cast consequence stated. Spec 005.

**Lane 2 — after both lane-1 PRs are merged and their decisions ratified:**

- **TASK-0006 (opus · model `cc/claude-opus-5[1m]`, fallback `cc/claude-opus-4-8[1m]` —
  ESCALATED by operator at sign-off 2026-08-21: decomposing the demo into
  a dependency-ordered set of one-PR deliverable tasks is planning judgment the spec
  does not settle — "thinking is Opus-tier" is the posture's own line; a wrong
  decomposition taxes every task of the next sweep. The plan itself is
  operator-signed-off (card AC #4) regardless of tier)** — build plan for the "one
  real evening" demo: Spec Kit spec sliced into buildable board tasks, likely the next
  sweep's input. Spec 006. Waits on 0004 (hard dep) and 0005 (its ratified death
  design is demo texture the plan must account for — cheap to wait, one lane).

Tiers and their model IDs come from **`.claude/model-tiers.json`** — `tiers.mjs --root .
--check` exited 0 on 2026-08-21 before these lanes were authored (all three agent
definitions `unchanged`; no regeneration, so no session restart required). Both opus
escalations were offered and TAKEN at sign-off — ruling recorded at the status line;
dispatches name `opus-implementer` (0004, 0006) and `sonnet-implementer` (0005).

Record the model tier + explicit model ID + rubric justification on each board task at
dispatch, including which model actually served (**verify from the first dispatch's
transcript before any sibling dispatch** — one wrong pin is a rounding error; a lane of
them is the budget).

## Per-PR gates this project enforces (enumerated — implementers cannot miss these)

- **Merge-drift gate: absent.** No `scripts/check-merge-drift.mjs` in this repo; raw
  git discipline stands: `git fetch origin && git pull --ff-only` at root before
  claim/worktree/merge; verify merged (`gh api repos/{owner}/{repo}/pulls/{n} --jq
  .merged`) before deleting branches/worktrees.
- **Root-guard hook** (`.claude/hooks/root-guard-hook.mjs`, PreToolUse on Bash and
  Write/Edit): the root checkout is read-only — all work rides worktree branches under
  `.worktrees/`; board/runbook bookkeeping rides the task branch (no-main-push mode);
  sweep-close lands via a wrap-up PR. The one exception: board-sync commits scoped
  entirely to `backlog/` may commit at root via Bash.
- **No runbook PRs (operator ruling 2026-08-21):** runbook commits — draft, sign-off,
  log rows — ride task branches or the wrap-up PR, never a dedicated PR.
- **Rebases are forbidden repo-wide** (root-guard doctrine) — every reconcile is a
  merge of `origin/main` into the branch; **PRs land as merge commits, never squash**
  (branches carry wiki re-pins; squash rewrites the hashes those pins reference).
- **Wiki freshness probe after every history move, unconditionally:** for each note in
  `docs/wiki/`, `git log <verified_against>..HEAD -- <sources>` must be empty or the
  note re-verified. Honest re-pins only — classify every staled note against
  `git diff <old-pin>..<merge-commit> -- <sources>` as RE-PIN-ONLY or NEEDS-REVIEW
  (amend prose BEFORE bumping); the merge commit is the re-pin target, never the
  justification.
- **CAPSULES.md is derived state:** regenerate via grounding-wiki `capsules.mjs`
  whenever any note's `description:` changes; never hand-edit.
- **Evidence rule:** any claim about an external dependency's version, maintenance,
  license, or pricing carries a URL + accessed date (established TASK-0001; binds
  TASK-0004's model-pricing/cost-envelope claims).
- **Board via CLI only,** from a checkout the CLI can write (the task worktree in this
  mode); add specific task files to git, never `backlog/` wholesale.

## Per-task artifacts required before PR

Per-TASK obligations. **No PR opens for a task until each line below checks true for
it.** The sweep's Output gate re-checks the first two lines at the end.

- [ ] `specs/004-mind-daemon-routing/` (TASK-0004), `specs/005-death-mechanics/`
      (TASK-0005), and `specs/006-evening-demo-build-plan/` (TASK-0006) each carry a
      real `spec.md` (problem + requirements mapped to the card's ACs), `plan.md` (the
      constitution at `.specify/memory/constitution.md` is an unfilled template —
      plan.md must state that plainly and plan against the grounding docs: the brief,
      CLAUDE.md, `docs/wiki/`), and `tasks.md` (phased checkboxes the bridge derives
      from), committed on the task's branch. A claim stub reserves the number; it
      satisfies nothing here.
- [ ] The card carries its Spec marker from the claim commit (`spec-bridge:link`
      against the stub), and phase ACs are seeded from tasks.md (link update mode)
      before implementation dispatch.
- **Escape lines (operator-signed only):** none.
- **Host additions:**
  - TASK-0004: the language/reuse decision lands as a Backlog decision record
    (`backlog decision create …`) in the same PR — card AC #1; the routing sketch and
    cost envelope (with assumptions) land under `docs/design/` in the same PR. Cost
    claims follow the evidence rule. Operator ratification (AC #4) follows PR review,
    before spec-bridge:sync moves the card Done.
  - TASK-0005: the death-mechanics design doc lands under `docs/design/` in the same
    PR; it must name the micromanagement spell-breaker check (AC #2) and the
    shrinking-cast stance (AC #3). Operator ratification (AC #4) follows PR review,
    before sync moves the card Done.
  - TASK-0006: the build plan lands under `docs/design/` in the same PR; the
    decomposed deliverable tasks are created on the board via the `backlog` CLI on the
    task's branch (each task one-PR-shaped, user-story-first per project convention,
    named design checks against the spell-breakers). Operator sign-off (AC #4) follows
    PR review, before sync moves the card Done.
  - All: wiki notes whose sources the PR touches are re-verified in that PR (board
    DoD: "Docs and wiki are updated and pass freshness tests").

## Concurrency & conflict doctrine

- **Hotspots:** `backlog/tasks/*` (all lanes flip cards and tick ACs),
  `docs/design/sweep-0004-0006-runbook.md` (all lanes log rows — later merge takes
  main's side and re-adds its row), `docs/wiki/CAPSULES.md` (regenerate after merge if
  any note description changed — never hand-merge), `docs/wiki/promptworld-lineage.md`
  (0004 primary; 0005/0006 must not edit it), `backlog/tasks/` new-task files
  (TASK-0006 creates cards — collision-check ids against fresh `origin/main` before
  creating).
- **Paused tasks are not live lanes** — none currently; if a `paused` label appears
  mid-sweep, never claim, rebase, or clean that task's branches/worktrees.
- Reconcile: **merge `origin/main` into the branch** (repo-wide rebase ban); take
  main's side for anything not deliberately changed. Treat every branch as
  pin-carrying once its wiki re-verifications land — merge-commit PRs only.
- After every history move: re-run gates AND the freshness probe unconditionally —
  pins reference `docs/design/*` sources outside the wiki, so a wiki-untouched diff
  can still be stale.
- Lane-1 PRs merge serially with a reconcile between: after the first merges, the
  second branch merges fresh `origin/main` in, re-runs the probe, then merges.
  **TASK-0004 merges first when both are ready** (contract-shaped); otherwise
  smallest-first.
- **Claim before work:** the FIRST task-scoped commit of each branch claims it — board
  card → In Progress, spec-number stub dir, spec-bridge link — pushed immediately
  (`git push -u origin <branch>`); never force-push a claim. Check `specs/` on fresh
  `origin/main` for number collisions before claiming. (On the TASK-0004 branch this
  runbook's draft precedes the claim — sanctioned by the no-runbook-PR ruling; the
  claim is the next commit after sign-off.)
- **A rejected push means you lost the race:** fetch, re-read board and `specs/`.
  Another holder → STOP the lane, surface to operator. Unrelated rejection with
  task+number still free → merge `origin/main` in and re-push (plain push).
- Verify a PR is merged (`gh api … --jq .merged`) before deleting its branch/worktree;
  never delete+recreate a closed PR's head.

## Operator checkpoints (do not proceed silently)

1. **Sign-off on these lanes** — pending. Includes ruling on the two PROPOSED opus
   escalations (TASK-0004, TASK-0006).
2. **TASK-0004 ratification (card AC #4):** operator ratifies the language decision
   record and routing/budget sketch. Surface the PR and stop before merging.
3. **TASK-0005 ratification (card AC #4):** operator ratifies the death design.
   Surface the PR and stop before merging.
4. **TASK-0006 plan sign-off (card AC #4):** operator signs off the build plan and its
   decomposed board tasks. Surface the PR and stop before merging.
5. Tier escalations; lane amendments (amend this file, note why, tell the operator).

## Done means

TASK-0004, TASK-0005, TASK-0006 Done on the board via one merged PR each;
`specs/004-*/`, `specs/005-*/`, `specs/006-*/` each carry real spec.md + plan.md +
tasks.md; all three cards still carry their Spec markers; decision record(s) for 0004
exist and are operator-ratified; the death design and the demo build plan exist and are
operator-ratified/signed-off; the demo's deliverable tasks exist on the board; wiki
freshness probe green over all notes; no stale sweep worktrees; this file's log
complete and status flipped to done (via wrap-up PR).

## Execution log

Multi-phase dispatch stays visible in `notes` — one slot, never a second table: while
a task is in flight its row carries the phases dispatched/completed (e.g.
`phases: 1-2 done, 3 dispatched`), updated at each dispatch boundary, so a resuming
session can see where within the task the last one stopped; the closing note on merge
replaces or absorbs it. `tokens/cost` carries best-effort actuals from the
harness/transcript, so future runbook authoring budgets against real numbers.
(Prior-sweep actuals for budgeting: a 3-phase opus task ran ~397k subagent tokens; a
3-phase sonnet task ~350k.)

| date | task | PR | merge | tokens/cost (best-effort) | notes |
|------|------|----|-------|---------------------------|-------|
| 2026-08-21 | TASK-0004 | — | — | — | in flight: claimed (35f2175), spec cycle committed (41268d6), phase ACs seeded (793d251), tier note (07d8a85). phases: 1 done (58a733e, served model VERIFIED cc/claude-opus-5[1m] from report, ~239k tokens; headline: promptworld I daemon is a co-process of the dead engine — 151 sim.* refs, 62% of non-test lines die — no daemon survives the seam in any language; portable assets: toolloop, persona, tool registry, llm layer minus cognition; rebuild-JVM added as 4th candidate), 2 dispatched. |
| 2026-08-21 | TASK-0005 | — | — | — | in flight: claimed (30d7702), spec cycle committed (72e4683), phase ACs seeded (ae62b5b), tier note (e88d029). phases: 1 done (dda2a19, sonnet verified, ~141k tokens; no villager hunger death, zombie kill usually converts, death drops nothing, POIs release, murder gossip), 2 done (4fe6f9d, sonnet verified, ~139k tokens; night danger admitted, conversion = death, sieges suppressed, grave + belongings + POI grief-period, zero protocol extension), 3 dispatched. |
