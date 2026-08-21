---
name: root-guard
description: The root-read-only enforcement hook — .claude/hooks/root-guard-hook.mjs wired as PreToolUse on Bash and Write/Edit, blocking direct modification of the root checkout; changes reach main only by merge from .worktrees/ branches. The backlog-commit exception, the quote-state scanner, and the no-rebase rule. Load when a write is blocked or when planning where work lands.
kind: component
sources:
  - .claude/hooks/root-guard-hook.mjs
  - .claude/hooks/shell-scan.mjs
  - .claude/settings.json
verified_against: 50c3def435dd9326d38e51118f08944815cbe80c
---

# Root guard

`.claude/hooks/root-guard-hook.mjs` enforces the worktree doctrine at the harness level:
**nothing is modified in the root checkout directly** — every change is authored on a
branch in a worktree under `.worktrees/` and reaches main only by merge (PR merge or
manual `git merge --no-ff` at root). Planted from pdlc (TASK-101 / spec 051); registered
in `.claude/settings.json` as PreToolUse hooks.

## How it works

Two modes, selected by argv:

- **`pre-bash`** (PreToolUse on Bash): finds every command-position git invocation in
  the command string — using the quote-state scanner in `shell-scan.mjs`, which replaced
  the upstream's quote-blind regexes (those shattered commit messages containing `)` or
  newlines into false pathspecs; see spec 051 "The defect") — resolves each invocation's
  effective directory (cwd, `cd` segments, `-C` options), and blocks
  root-mutating git operations in jurisdiction.
- **`pre-write`** (PreToolUse on Write|Edit|NotebookEdit): blocks writes to root paths
  outright, including `backlog/` (the CLI via Bash is the sanctioned editor).

**The one ratified exception** (TASK-161): board-sync commits scoped entirely to
`backlog/` are allowed at root via pre-bash `git commit` — the board is the plan of
record and the concurrent-session mutex. The exception does not extend to pre-write.

**Rebases are forbidden everywhere in this repo** — history moves are merges.

There is deliberately **no bypass flag**: emergencies go through the operator editing
the hook config, visibly. When blocked, the remedy is the doctrine: author in a
worktree, land by merge. In practice this puts the project in the sweep's no-main-push
mode ([[pdlc-process]]): bookkeeping rides task branches and wrap-up PRs.

## Connections

Shapes every landing path in [[pdlc-process]]; the worktree discipline it enforces is
also the operator's standing global instruction; TASK-0001's sweep ran entirely under
it (bookkeeping via branch commits, close via wrap-up PR #2).

## Operational notes

Node ≥ 18, ESM, zero npm dependencies. Behavior is ported verbatim from promptworld's
hook; only parsing was hardened. Hook errors surface in the tool result with the
blocked path and the remedy text.
