# Tasks: Dusk conversation and the ambient pool (E4, E5)

**Spec dir**: `specs/017-dusk-conversation` · **Branch**: `task-0017-dusk-conversation`

## Phase 1 — The dusk exchange (US1)

- [x] T001 `mind/converse/`: two-mind exchange loop — alternating turns via the
      E4 registry config (asserted: Sonnet 5, streaming, effort low, thinking
      off, cached prefix, ~300 max_tokens — card AC #1); speak intents out
      through the seam surface
      (`mind/converse/converse.go`, `TestExchangeUsesE4Config`). Deviation:
      the intent schema (body-protocol-v0.md §5.2) has no separate content
      field for a speak's text, so it rides inside `target` alongside `type`
      (`{"type":"none","text":...}`) — the only place the schema leaves for
      it; noted in `Speaker.speak`'s doc comment. Target is `type: "none"`
      (spoken into earshot at the gathering place) rather than a directed
      `body` target, since a directed target would need a pre-issued token
      (fakevendor's `targetUnissued` check) this phase has no reason to set
      up.
- [x] T002 Per-turn context assembly: transcript-so-far + interlocutor slice
      (own slice, from beliefs) + memory window, rendered in the variable
      suffix; stable prefix byte-untouched (test)
      (`Speaker.assemble`, `TestSpeakerAssemble_StablePrefixByteIdentical`).
- [x] T003 Termination condition: closing-marker convention detected; safety
      bound is config, asserted never to fire in scripted tests (card AC #7
      no-mid-sentence-cap clause)
      (`ClosingMarker`/`DefaultMaxTurns`, `TestExchange_SafetyBoundNeverFires`).
- [x] T004 Scripted dusk exchange against the fake vendor + mock model: multi-
      turn, about the day/work/player; first-token latency instrumented and
      under the test ceiling (card ACs #2, #3); vet + tests green
      (`mind/converse/fakevendor_test.go`,
      `TestDuskExchange_AgainstFakeVendor` — `go vet ./...` and
      `go test -race ./...` both green in `mind/`).

## Phase 2 — Pre-generation (US2)

- [x] T005 `Slot`/`Pool` (`mind/converse/pregen.go`): a per-pair slot keyed
      `PairKey{PairID, Day}`; `Pool.Begin` fills off V3's pair-formation
      signal (modeled as the call the daemon wiring will make — no seam
      ingest wiring built this phase, per plan.md decision 3); `Exchange`
      gained `Config.OpeningSlot`/`OpeningWait` so a's opening turn serves
      from an already-filled slot without a new E4 call (card AC #4)
      (`TestSlot_ServeAtConvergence`, `TestPool_BeginTakeServesKeyedByPairDay`,
      `TestExchange_OpeningSlotServesWithoutNewCall`).
- [x] T006 Abort-discard: `Slot.Discard`/`Pool.Discard` drop a slot unspoken
      even if its fill is still in flight (`TestSlot_AbortDiscard`,
      `TestPool_DiscardDropsUnspoken`). Live-stream fallback: an unfilled
      slot (TASK-0014's measured 1.82-4.96 s signal lead vs nominal ~10 s —
      first-class, not an edge case) falls back to `Exchange`'s normal live
      E4 call for the opening turn, ceiling unaffected
      (`TestExchange_OpeningSlotFallsBackWhenUnfilled`). Race, both orders:
      `Slot.Take` is a one-shot claim (mutex-guarded `spoken` flag) so a
      fill that completes after convergence already fell back is discarded,
      and a fill that completes before convergence serves cleanly
      (`TestSlot_LiveFallbackDiscardsLateFill`,
      `TestSlot_ConcurrentFillAndTakeRace` — 25 trials under `-race`).
      Deviation: `OpeningWait` defaults to a non-blocking check (0), not a
      timed wait — spending any of the ceiling waiting on a slot that may
      already be known-unfilled would eat into the budget pre-generation
      exists to protect; the parameter exists for a caller that wants to
      wait a bounded amount instead, but Phase 2 exercises the zero-wait
      default only.

## Phase 3 — The ambient pool (US3)

- [x] T007 `AmbientPool` (`mind/converse/pool.go`, named distinctly from
      pregen.go's `Pool` per Phase 2's naming note): `Refill` runs one
      batched E5 call/villager/day (asserted Haiku 4.5, non-streaming —
      `TestE5UsesHaiku45`) producing ~8 persona-flavoured lines
      (`TestAmbientPool_RefillProducesLines`); `Serve` is a mutex-guarded
      map lookup, measured < 200 ms, and removes each line as it's served
      so nothing repeats within a cycle
      (`TestAmbientPool_ServeUnderBudgetNoRepeat`); daily refresh keyed by
      `day int64` (world_time arithmetic, M2 convention — computing it is
      the daemon wiring's job, not this package's, same posture as
      pregen.go's `PairKey.Day`): `Serve` on a day that doesn't match the
      pool's last `Refill` reports pool-empty rather than serving
      yesterday's still-unspent lines (`TestAmbientPool_DayRolloverClearsYesterday`,
      card AC #5).
- [x] T008 `IsTargeted`/`Escalate` (`mind/converse/pool.go`): a non-empty
      trigger subject routes to `Escalate`, a live E5 call (same Haiku 4.5
      tier as the pool, made live instead of batched) that never touches
      `AmbientPool` (`TestIsTargeted`, `TestEscalate_LiveCallBypassesPool`,
      card AC #6). `StallLines`/`Stall`: the pool-empty stall-line policy
      (spec.md Edge Cases) — `Stall` takes no `*llm.Client`, so it cannot
      make a model call by construction, not by caller discipline; picks
      deterministically from four short, non-committal lines
      (`TestStall_NeverCallsModel`, `TestStall_Deterministic`; card AC #7's
      stall clause). Deviation: no anti-repetition or usage-rate tracking
      on `Stall` itself — "used rarely, never a prefix tic" is a caller
      discipline this phase documents (`Stall`'s doc comment: "a
      substitute for a turn where Serve had nothing", not a default
      reply), not a mechanism enforced here; T009 (Phase 4) is where the
      tedium checks get consolidated across the whole class.

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
