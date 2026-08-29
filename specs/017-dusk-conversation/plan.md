# Implementation Plan: Dusk conversation and the ambient pool (E4, E5)

**Spec dir**: `specs/017-dusk-conversation` · **Branch**: `task-0017-dusk-conversation`

**Constitution check**: `.specify/memory/constitution.md` is an unfilled template —
stated plainly per runbook gate. Planned instead against the grounding docs:
kithcraft-brief.md, decision-0003 + llm-routing-and-budget.md, body-protocol-v0.md,
project CLAUDE.md, and `docs/wiki/` ([[body-protocol-seam]], [[overview]]).

## Where it lives

New package `mind/converse/` — dusk exchange, pre-generation slot, ambient pool.
Sits on `mind/llm` (E4/E5 class configs, streaming client, mock), `mind/prompt`
(stable prefix + variable suffix; new interlocutor slice), `mind/memory` (window),
`mind/seam` (speak intents out). No wire/protocol changes. TASK-0016's
`mind/deliberate` (in flight on a sibling branch) is NOT a dependency — the two
packages are disjoint; expect a clean merge either order.

## Design decisions (settled surfaces restated)

1. **E4 config comes from the registry** (`mind/llm/classes.go`, ratified in
   TASK-0011): Sonnet 5, streaming, effort low, thinking off, cache breakpoints,
   max_tokens ~300. The exchange code reads the registry; it never re-declares
   params. E4 is prose (StructuredOutput false) — turns are text.
2. **Termination** — natural end: the prompt instructs a closing marker on the
   final turn (a convention the mind detects), plus a generous safety bound that
   is a config value, with the safety-bound path asserting it never fires in the
   scripted tests. The card forbids a *turn cap as the mechanism*; a safety bound
   that provably doesn't drive endings is engineering hygiene.
3. **Pre-generation** — a per-pair slot keyed (pairID, day): fill on signal,
   serve-or-discard on converge/abort, live-stream fallback if unfilled at
   convergence (covers V3's measured 1.82–4.96 s lead honestly).
4. **Interlocutor model** — its own context slice (who this is, what I think of
   them, shared history) assembled from the belief store; rendered in the
   variable suffix (stable prefix untouched — caching lever intact).
5. **Pool** — per-villager ring of ~8 lines with served-set tracking; refresh
   keyed to in-game day rollover; serve is a map lookup (the < 200 ms budget is
   trivially met in-process — measured anyway, per FR-003 discipline).
   Specific-remark escalation is a predicate on the trigger (targeted prompt
   context present → live Haiku call).
6. **Latency measurement** — first-token wall-clock instrumented in the client
   streaming path (M4's AccountedStream is the natural hook); tests inject mock
   delays to prove the measurement works, not to prove Anthropic is fast.
7. **No new deps** — stdlib + existing modules only. No live calls in tests.

## Risks / open items

- The real < 3 s ceiling is provable only with live calls (I2's evening run
  measures it for real); tests prove the instrumentation and the config posture.
  Stated honestly in test names/comments.
- Day-rollover source: world_time from percepts (M2 arithmetic convention), not
  wall clock.

## Phase map

Phase 1 — the exchange: turns, context slices, termination (US1).
Phase 2 — pre-generation slot off the pair signal (US2).
Phase 3 — ambient pool + escalation (US3).
Phase 4 — spell-breaker checks, gates, wiki re-ground, board close.
