---
name: pdlc-process
description: How work happens here — the PDLC loop (ground → spec → build → re-ground), the Backlog.md board as plan of record, Spec Kit specs bridged by spec-bridge, one-task-one-PR, sweeps with runbooks, and the gates that hold status to artifacts. Load when orchestrating, claiming, or closing work.
kind: pipeline
sources:
  - CLAUDE.md
  - AGENTS.md
  - backlog/config.yml
verified_against: 50c3def435dd9326d38e51118f08944815cbe80c
---

# PDLC process

Kithcraft develops under the praxisflux plugin suite's praxis development lifecycle:
ground the codebase → plan as specs → build → re-ground → teach/render. `CLAUDE.md`
carries the always-on doctrine; this note is its map, not its replacement.

## How it works

**Planes and their rules:**

- **The board** (Backlog.md, `backlog/`) is the plan of record. Statuses flow
  To Do → In Progress → Done; CLI-only, never hand-edit (AGENTS.md hardens this).
  Task descriptions open with a user story (project convention). The board's
  definition-of-done includes "Docs and wiki are updated and pass freshness tests" —
  this wiki is load-bearing for every task's DoD.
- **Specs** (Spec Kit, `specs/NNN-<slug>/`) drive features: spec.md + plan.md +
  tasks.md, all real before implementation dispatches. spec-bridge links a spec to its
  board task and its sync derives board status from spec artifacts — sync is the ONLY
  path that moves a linked task to Done.
- **One TASK, one PR** — a task maps 1:1 to a branch and PR; subtasks are commits.
  A PR exists only where it carries a real approval decision.
- **Artifact-grounded action** — nothing counts unless it lives in a file, commit, or
  tracker entry; decisions derive from artifacts and produce new ones.
- **Sweeps** (`/pdlc:sweep`) orchestrate task sets through the whole loop from a
  signed-off runbook (`docs/design/*-runbook.md`), dispatching implementation to tier
  agents ([[model-tiers]]) phase-by-phase, merging serially. TASK-0001's sweep
  (runbook `docs/design/task-0001-mod-stack-runbook.md`, status done) is the field
  example: 3 phases dispatched to sonnet, PR #1, ratified, synced Done.
- **Gates:** spec-bridge's Stop hook blocks board status exceeding spec artifacts;
  [[root-guard]] blocks root writes; this wiki's freshness gate runs as check scripts.
  When a gate blocks, produce the missing artifact — never argue or hand-edit around it.

**Execution mode note:** the root checkout is read-only ([[root-guard]]), so board and
runbook bookkeeping rides task branches (the sweep's no-main-push mode), with wrap-up
PRs for sweep-close.

## Connections

Model dispatch rules: [[model-tiers]]; root enforcement: [[root-guard]]; the process's
first full product: [[mod-stack-decision]]; overall map: [[overview]].

## Operational notes

Constitution (`.specify/memory/constitution.md`) is an unfilled template — plans state
that and check against grounding docs instead. Evidence rule: dependency claims carry a
URL + accessed date.
