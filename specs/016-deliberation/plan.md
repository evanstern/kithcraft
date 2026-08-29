# Implementation Plan: Deliberation and the job-board decision (E2, E3)

**Spec dir**: `specs/016-deliberation` · **Branch**: `task-0016-deliberation`

**Constitution check**: `.specify/memory/constitution.md` is an unfilled template —
stated plainly per runbook gate. This plan is checked instead against the grounding
docs: kithcraft-brief.md, the five settled surfaces (decision-0001..0004 +
body-protocol-v0 + llm-routing-and-budget + entity-implementation-comparison),
project CLAUDE.md, and `docs/wiki/` ([[body-protocol-seam]],
[[promptworld-lineage]], [[overview]]).

## Where it lives

New package `mind/deliberate/` — the loop, triggers, interrupt, and window sit
above `mind/seam` (Pending, manifest verbs), `mind/memory` (beliefs, admission),
`mind/llm` (client, classes, structured output), and `mind/prompt` (assembly).
No changes to the wire or protocol surface: deliberation is pure mind-side
composition of already-merged parts. `mind/fakevendor` + the M4 mock client
drive every test; no live API calls.

## Design decisions (settled by the ratified surfaces; restated, not re-made)

1. **Loop shape** — port toolloop's *shape* only (promptworld code does not
   survive the seam, decision-0003): compose intent (REQUEST) → vendor acks →
   `act_result` (FACT) → gate decides what enters memory (M2's admission gate,
   already merged). The loop owns no memory writes beyond handing facts to the
   admission gate.
2. **Interrupt** — `context.Context` cancellation on the in-flight SDK call
   (RT-2, already in M4's client). A pending-urgent buffer, drained into exactly
   one enqueued deliberation; draining is idempotent under multiple urgents.
3. **Window** — selector over M2's belief/event store: decay = salience halved
   per day (`world_time` arithmetic, RM-5/RM-6 style), top K−2, plus 2 picks
   seeded per-villager (deterministic PRNG seeded from persona identity) from
   the older half. Pure function over (store snapshot, now, seed) → testable
   without time mocking.
4. **E3 context shape** — a struct assembled from board text percept + claims +
   relationship + commitments, rendered into the variable suffix by
   `mind/prompt` (stable prefix untouched — §4.3 caching lever intact).
5. **Reluctance** — personas make declines reachable; the scripted-evening test
   uses scripted model responses (the mock client), so the design checks verify
   the *plumbing* (reasons carried, no conduct gate, work completes) rather than
   model behaviour — the honest boundary for a no-live-calls task.
6. **No new deps** — stdlib + existing modules only.

## Risks / open items

- The exact trigger wiring to a live schedule (V3's world) is out of scope: E2
  triggers arrive as percepts/events through the seam; the demo run (I2) is
  where end-to-end firing is observed. Tests script the triggers.
- Coalescing semantics under cancel/complete race: settle in code with a mutex'd
  state machine; test both orders.

## Phase map

Phase 1 — the bounded loop + manifest verbs + token-only targets (US1).
Phase 2 — E3 job-board deliberation + authored reason + persona declines (US2).
Phase 3 — urgency interrupt (US3) + K=10 window (US4).
Phase 4 — scripted-evening design checks, gates, wiki re-ground, board close.
