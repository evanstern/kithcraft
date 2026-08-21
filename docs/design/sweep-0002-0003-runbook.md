# TASK-0002 + TASK-0003 body-protocol & entity-decision — sweep runbook (2026-08-21)

**You (the session reading this) are the ORCHESTRATOR** for the tasks below. Run each
through the host project's full PDLC — spec → link → worktree → delegated implementation →
PR → merge → re-ground — parallelizing within lanes, merging serially, treating merge
conflicts as routine. Direction is decided; do not re-litigate it:
`docs/design/kithcraft-brief.md` (ratified 2026-08-19) and decision-0001 (Fabric,
accepted 2026-08-20) win. Plan-of-record is the board; this file carries only ordering,
doctrine, and the log.

**Status:** signed-off · operator sign-off on lanes: 2026-08-21
<!-- Sign-off ruling (2026-08-21): lanes approved as authored; TASK-0002 tier ESCALATED
     to opus by the operator (the protocol is the project's one-way door — buy the
     thinking tier for it). TASK-0003 stays sonnet. Checkpoint 1 satisfied; the
     escalation checkpoint for opus dispatch is recorded here per the tier rubric. -->
<!-- Only the OPERATOR flips draft → signed-off (the author never pre-fills it). An
     executing session must refuse a runbook whose status it cannot verify. -->

## Read first (in this order)

1. `docs/design/kithcraft-brief.md` — the ratified design brief that produced these tasks.
2. `docs/wiki/CAPSULES.md` — corpus rollup; then just-in-time full notes:
   [[body-protocol-seam]] + [[promptworld-lineage]] for TASK-0002,
   [[villager-brain-api]] + [[mod-stack-decision]] for TASK-0003.
3. Project `CLAUDE.md` — PDLC grounding, model-tier posture, task conventions.
4. `backlog task list --plain` — live state; other sessions move it while you work.
5. The task you're about to execute (`backlog task view TASK-<n> --plain`).

## State when this runbook was written (2026-08-21)

- **Done already:** TASK-0001 (Fabric decision, PR #1, decision-0001 accepted). First
  wiki build merged (PR #3, 11 notes pinned at 50c3def).
- **In flight in other sessions (do not duplicate; expect their merges):** none known.
- **Paused — untouched:** none.
- **Queued (this runbook's scope):** TASK-0002, TASK-0003 — one lane, parallel
  development, serial merges.

## Execution lanes (dependency-ordered; parallelize within a lane)

Rule of thumb: DEVELOP in parallel, MERGE serially. The two tasks are
parallel-safe by design: the body protocol (TASK-0002) is deliberately world-agnostic —
the anti-corner move forbids it coupling to the entity choice — and the entity decision
(TASK-0003) is vendor-internal, behind the seam. Both depend only on TASK-0001 (Done).
Their spec dirs, design docs, and primary wiki notes are disjoint; shared surfaces are
the board dir and this file (see hotspots).

**Lane 1 — start immediately, in parallel:**

- **TASK-0002 (opus · model `cc/claude-opus-5[1m]`, fallback `cc/claude-opus-4-8[1m]` —
  ESCALATED by operator at sign-off 2026-08-21: protocol drafting is design work on the
  project's one-way door (the seam), a judgment call the spec does not settle; the
  escalation checkpoint the tier rubric requires is the sign-off ruling recorded at the
  status line. The final human gate — accepting the contract — remains the operator's
  PR review)** — draft body protocol v0
  (perceive/act/remember message shapes, versioning story, perception model with
  provenance, fake/test vendor spec). Spec 002. **Contract-shaped: its merged protocol
  doc is the interface every future mind/vendor task consumes** — it merges first if
  both PRs are ready together.
- **TASK-0003 (sonnet · model `cc/claude-sonnet-5[1m]`, no fallback declared for this
  tier — default tier: trade-off analysis against ratified constraints, directly
  analogous to TASK-0001's comparison which ran sonnet; the judgment call — ratifying
  the recommendation — is an operator checkpoint (AC #2) regardless of tier)** —
  custom entity vs augmented vanilla villager writeup + recommendation + Backlog
  decision record. Spec 003.

Tiers and their model IDs come from **`.claude/model-tiers.json`** — `tiers.mjs --root .
--check` exited 0 on 2026-08-21 before these lanes were authored (all three agent
definitions `unchanged`; no regeneration, so no session restart required). The opus
escalation for TASK-0002 was offered and TAKEN at sign-off — ruling recorded at the
status line; dispatch names `opus-implementer`.

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
- **Rebases are forbidden repo-wide** (root-guard doctrine) — every reconcile is a
  merge of `origin/main` into the branch, for pin-free and pin-carrying branches
  alike; **PRs land as merge commits, never squash** (both branches carry wiki
  re-pins; squash rewrites the hashes those pins reference).
- **Wiki freshness probe after every history move, unconditionally:** for each note in
  `docs/wiki/`, `git log <verified_against>..HEAD -- <sources>` must be empty or the
  note re-verified. Honest re-pins only — classify every staled note against
  `git diff <old-pin>..<merge-commit> -- <sources>` as RE-PIN-ONLY or NEEDS-REVIEW
  (amend prose BEFORE bumping); the merge commit is the re-pin target, never the
  justification.
- **CAPSULES.md is derived state:** regenerate via grounding-wiki `capsules.mjs`
  whenever any note's `description:` changes; never hand-edit.
- **Evidence rule:** any claim about an external dependency's version, maintenance, or
  license carries a URL + accessed date (established TASK-0001; binds TASK-0003's
  entity comparison).
- **Board via CLI only,** from a checkout the CLI can write (the task worktree in this
  mode); add specific task files to git, never `backlog/` wholesale.

## Per-task artifacts required before PR

Per-TASK obligations. **No PR opens for a task until each line below checks true for
it.** The sweep's Output gate re-checks the first two lines at the end.

- [ ] `specs/002-body-protocol-v0/` (TASK-0002) and `specs/003-entity-implementation/`
      (TASK-0003) each carry a real `spec.md` (problem + requirements mapped to the
      card's ACs), `plan.md` (the constitution at `.specify/memory/constitution.md` is
      an unfilled template — plan.md must state that plainly and plan against the
      grounding docs: the brief, CLAUDE.md, `docs/wiki/`), and `tasks.md` (phased
      checkboxes the bridge derives from), committed on the task's branch. A claim
      stub reserves the number; it satisfies nothing here.
- [ ] The card carries its Spec marker from the claim commit (`spec-bridge:link`
      against the stub), and phase ACs are seeded from tasks.md (link update mode)
      before implementation dispatch.
- **Escape lines (operator-signed only):** none.
- **Host additions:**
  - TASK-0002: the protocol doc lands under `docs/design/` (or a location the spec
    names) in the same PR; [[body-protocol-seam]]'s sources grow to include it and the
    note is re-verified in the same PR (its own Operational notes demand this — the
    seam moves from posture to contract).
  - TASK-0003: the recommendation lands as a Backlog decision record
    (`backlog decision create …`) in the same PR — card AC #2; a recommendation living
    only in the comparison doc does not satisfy it. Operator ratification follows PR
    review, before spec-bridge:sync moves the card Done.
  - Both: wiki notes whose sources the PR touches are re-verified in that PR
    (board DoD: "Docs and wiki are updated and pass freshness tests").

## Concurrency & conflict doctrine

- **Hotspots:** `backlog/tasks/*` (both lanes flip cards and tick ACs),
  `docs/design/sweep-0002-0003-runbook.md` (both lanes log rows — second merge takes
  main's side and re-adds its row), `docs/wiki/CAPSULES.md` (regenerated if either
  task changes a note description — regenerate after merge, don't hand-merge),
  `docs/wiki/body-protocol-seam.md` (0002 primary; 0003 must not edit it).
- **Paused tasks are not live lanes** — none currently; if a `paused` label appears
  mid-sweep, never claim, rebase, or clean that task's branches/worktrees.
- Reconcile: **merge `origin/main` into the branch** (repo-wide rebase ban — see
  gates); take main's side for anything not deliberately changed. Both branches are
  pin-carrying once their wiki re-verifications land, so merge-commit PRs only.
- After every history move: re-run gates AND the freshness probe unconditionally —
  pins reference `docs/design/*` sources outside the wiki, so a wiki-untouched diff
  can still be stale.
- The two PRs merge serially with a reconcile between: after the first merges, the
  second branch merges fresh `origin/main` in, re-runs the probe, and only then merges.
  **TASK-0002 merges first when both are ready** (contract-shaped); otherwise
  smallest-first.
- **Claim before work:** the FIRST commit of each task branch claims it — board card →
  In Progress, spec-number stub dir, spec-bridge link — pushed immediately
  (`git push -u origin <branch>`); never force-push a claim. Check `specs/` on fresh
  `origin/main` for number collisions before claiming.
- **A rejected push means you lost the race:** fetch, re-read board and `specs/`.
  Another holder → STOP the lane, surface to operator. Unrelated rejection with
  task+number still free → merge `origin/main` in and re-push (plain push).
- Verify a PR is merged (`gh api … --jq .merged`) before deleting its branch/worktree;
  never delete+recreate a closed PR's head.

## Operator checkpoints (do not proceed silently)

1. **Sign-off on these lanes** — SATISFIED 2026-08-21: lanes approved, TASK-0002
   escalated to opus (ruling at the status line).
2. **TASK-0002 protocol acceptance:** PR review is the human gate on the contract —
   the seam is the project's one-way door. Surface the PR and stop before merging.
3. **TASK-0003 ratification (card AC #2):** operator ratifies the entity
   recommendation (decision record → accepted) before spec-bridge:sync moves the card
   Done. Surface the PR and stop before merging.
4. Tier escalations; lane amendments (amend this file, note why, tell the operator).

## Done means

TASK-0002 and TASK-0003 Done on the board via one merged PR each;
`specs/002-*/` and `specs/003-*/` each carry real spec.md + plan.md + tasks.md; both
cards still carry their Spec markers; the protocol doc exists and
[[body-protocol-seam]] cites it; the entity decision record exists and is
operator-ratified; wiki freshness probe green over all notes; no stale sweep
worktrees; this file's log complete and status flipped to done (via wrap-up PR).

## Execution log

Multi-phase dispatch stays visible in `notes` — one slot, never a second table: while
a task is in flight its row carries the phases dispatched/completed (e.g.
`phases: 1-2 done, 3 dispatched`), updated at each dispatch boundary, so a resuming
session can see where within the task the last one stopped; the closing note on merge
replaces or absorbs it. `tokens/cost` carries best-effort actuals from the
harness/transcript, so future runbook authoring budgets against real numbers.

| date | task | PR | merge | tokens/cost (best-effort) | notes |
|------|------|----|-------|---------------------------|-------|
| 2026-08-21 | TASK-0002 | — | — | — | in flight: claimed (340c3a2), spec cycle committed (34cb2df), phase ACs seeded (494f823). phases: 1 dispatching (opus-implementer, cc/claude-opus-5[1m] — operator-escalated at sign-off; served model to be verified from first transcript). |
| 2026-08-21 | TASK-0003 | — | — | — | in flight: claimed (2fe6f9c), spec cycle committed (b1b1439), phase ACs seeded (92f8c49). phases: 1 dispatching (sonnet-implementer, cc/claude-sonnet-5[1m]; sonnet pin field-verified on this host in TASK-0001 sweep — still re-verify from first transcript). |
