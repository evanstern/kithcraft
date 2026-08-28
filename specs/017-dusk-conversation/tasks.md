# Tasks: Dusk conversation and the ambient pool (E4, E5)

**Spec dir**: `specs/017-dusk-conversation` · **Branch**: `task-0017-dusk-conversation`

## Phase 1 — The dusk exchange (US1)

- [ ] T001 `mind/converse/`: two-mind exchange loop — alternating turns via the
      E4 registry config (asserted: Sonnet 5, streaming, effort low, thinking
      off, cached prefix, ~300 max_tokens — card AC #1); speak intents out
      through the seam surface
- [ ] T002 Per-turn context assembly: transcript-so-far + interlocutor slice
      (own slice, from beliefs) + memory window, rendered in the variable
      suffix; stable prefix byte-untouched (test)
- [ ] T003 Termination condition: closing-marker convention detected; safety
      bound is config, asserted never to fire in scripted tests (card AC #7
      no-mid-sentence-cap clause)
- [ ] T004 Scripted dusk exchange against the fake vendor + mock model: multi-
      turn, about the day/work/player; first-token latency instrumented and
      under the test ceiling (card ACs #2, #3); vet + tests green

## Phase 2 — Pre-generation (US2)

- [ ] T005 Pair slot keyed (pairID, day): fill on V3's pair-formation signal,
      serve at convergence without a new call (card AC #4)
- [ ] T006 Abort-discard (signal fired, meeting aborted → slot dropped unspoken)
      and live-stream fallback (slot unfilled at convergence → ceiling still
      held via streaming; late fill discarded, at-most-one opening spoken)

## Phase 3 — The ambient pool (US3)

- [ ] T007 E5 pool: one batched Haiku 4.5 call per villager per in-game day,
      ~8 persona-flavoured lines; serve < 200 ms measured; no intra-cycle
      repeat; daily refresh on world_time rollover (card AC #5)
- [ ] T008 Specific-remark escalation: targeted trigger → live Haiku call, not
      the pool (card AC #6); pool-empty stall-line policy sparing (card AC #7)

## Phase 4 — Spell-breaker checks, gates, and closure

- [ ] T009 Tedium checks consolidated: no repeats within cycle, natural end,
      stall-line sparing (card AC #7 complete)
- [ ] T010 Politeness-policing: no lecture/moralize/conduct-gate template in
      E4/E5 assembly — structural check on prompt text + scripted content check
      (card AC #8)
- [ ] T011 Full gates: `go vet` + `go test ./...` green; scope clean
- [ ] T012 Wiki re-ground: touched-source notes re-verified honestly (overview
      at minimum); CAPSULES regenerated if descriptions changed; freshness green
- [ ] T013 Card ACs ticked with citing proofs; board/spec synced at PR time
