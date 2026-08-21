# Plan 006 — Demo build plan

**Constitution status:** `.specify/memory/constitution.md` is an unfilled template —
no ratified constitution exists. Per PDLC doctrine this plan is checked against the
project's grounding docs instead: `docs/design/kithcraft-brief.md` (ratified),
`CLAUDE.md`, and `docs/wiki/` (load [[v1-demo]], [[villager-brain-api]],
[[body-protocol-seam]], [[pdlc-process]] just-in-time), plus the four ratified
decision surfaces (decision-0001..0003, death design, routing/budget doc).

## Approach

Planning task; no code. Two phases (a decomposition pass, then the board landing +
sign-off prep — fewer, bigger phases than specs 001–005 because the deliverable is
one coherent plan, not three separable research/design products):

1. **Decompose the demo.** From the demo beats (spec R2) and the ratified surfaces,
   derive the task graph: mind-side tasks (daemon skeleton, persona genesis, memory,
   conversation), vendor-side tasks (Fabric mod skeleton, brain/schedule wiring,
   job-board book, death/danger Mixins), seam tasks (transport decision, protocol
   conformance + fake-vendor harness), and integration tasks (the evening itself,
   demo config). For each: user story, scope boundary ("done proves"), dependencies,
   spell-breaker checks where the risk lives, and a suggested tier per the model-tier
   rubric. Output: `docs/design/demo-build-plan.md`.
2. **Land it on the board + sign-off prep.** Create the tasks via the `backlog` CLI
   (collision-checking ids against fresh origin/main), dependencies recorded,
   milestone m-0, each description opening with its user story. Cross-check coverage
   beat-by-beat against [[v1-demo]] and the plan doc; state suggested lanes for the
   next sweep. Operator signs off at PR review.

## Structure

- `docs/design/demo-build-plan.md` — Phase 1 deliverable (the plan of record's
  narrative: graph, rationale, lanes, checks).
- `backlog/tasks/task-00NN…` — Phase 2: the created tasks (CLI only, committed on
  this branch).
- Wiki: [[v1-demo]]'s Operational notes list the open questions feeding the demo —
  all resolved once this lands; the note is re-verified in this PR. CAPSULES
  regenerated if any description changes.

## Risks / notes

- **Decomposition granularity** is the judgment call: too-big tasks recreate the
  undifferentiated ambition; too-small tasks violate "a PR exists only where it
  carries a real approval decision". The one-PR test (coherent, reviewable
  deliverable) is the razor.
- Task ids: next free id derived from the board at claim time; concurrent sessions
  could take ids — collision-check before create, renumber on conflict.
- The plan will suggest tiers for future tasks, but tier assignment at the next
  sweep's runbook (with operator escalation checkpoints) remains authoritative.
