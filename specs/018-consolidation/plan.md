# Implementation Plan: Nightly consolidation and the archived dead (E6)

**Spec dir**: `specs/018-consolidation` · **Branch**: `task-0018-consolidation`

**Constitution check**: `.specify/memory/constitution.md` is an unfilled template —
stated plainly per runbook gate. Planned against the grounding docs instead:
decision-0003 + llm-routing-and-budget.md, death-mechanics.md §3,
body-protocol-v0.md, kithcraft-brief.md, `docs/wiki/` ([[promptworld-lineage]]
— the consolidation machinery shape is a named port asset; [[body-protocol-seam]]).

## Where it lives

New package `mind/consolidate/` — ledger, digest cycle, death-carry weighting,
archival. Sits on `mind/memory` (admitted buffer, event log, `(tick, hash)`
identity — the log's `(world_time, hash)` pair per M2), `mind/llm` (E6 registry
config, mock client; the Digest shape M4 ponytailed for M7 gets completed here),
`mind/seam` (session refusal for archived minds, token retirement). Disjoint
from sibling in-flight packages (`mind/deliberate` TASK-0016, `mind/converse`
TASK-0017) except one seam: death-carry exposes a retrieval-weighting hook the
conversation context will consume — landed here as an exported selector so M6
can adopt it at merge, no cross-branch edit.

## Design decisions (settled surfaces restated)

1. **Machinery shape ported, not code** (decision-0003): once-per-night
   event-sourced ledger; ordinal `m1..mN` convention (the prompt convention IS
   the identity mechanism — accepted references mapped to `(world_time, hash)`);
   a transport failure lands NO marker. Digest completion detects over-limit as
   failure (I's day-20 truncation lesson — routing §6's measured lesson list).
2. **Trigger** — the sleep event arrives as a percept/event through the seam;
   all windows and retries computed on `world_time` (harness T-b). No wall
   clock anywhere in the package; no Batch API client path exists to misuse.
3. **Empty night lands a marker** — nothing-to-digest is a consolidated night;
   only failure withholds the marker (retry semantics stay crisp).
4. **Death carry** — a retrieval-frequency weighting function over conversation-
   context candidate selection: a death event's weight spikes for the next
   cycle, decays per cycle after, floors at normal-presence (never zero, never
   deleted — RM-7). Pure function over (store, now) — deterministic, testable.
5. **Archival (R-9)** — an archived-set keyed by mind identity: session opens
   refused for archived minds; body token retired into a never-reissue set
   (persisted with the log); the durable log stays readable via the existing
   store API. No process termination semantics — archival is state, not
   lifecycle.
6. **No new deps**; no live calls in tests (mock client, fake vendor).

## Risks / open items

- The Digest structured shape M4 ponytailed: completing it here may touch
  `mind/llm/structured.go` — a shared-file hotspot only if a sibling branch
  edits it too (neither 016 nor 017 should; noted for merge reconcile).
- Retrieval-frequency AC (#5) is statistical: made deterministic by seeding and
  by asserting on the weighting function's output distribution over fixed
  inputs, not on sampled draws.

## Phase map

Phase 1 — ledger + digest cycle + m1..mN mapping + no-marker-on-failure (US1+US2).
Phase 2 — death carry weighting (US3).
Phase 3 — archival, session refusal, token retirement (US4).
Phase 4 — gates, wiki re-ground, board close.
