# Tasks: Deliberation and the job-board decision (E2, E3)

**Spec dir**: `specs/016-deliberation` · **Branch**: `task-0016-deliberation`

## Phase 1 — The bounded loop (US1)

- [ ] T001 `mind/deliberate/`: loop skeleton — compose intent via `seam.Pending`,
      REQUEST/FACT/gate mapping onto intent/intent_ack/act_result; bounded
      iterations (config with default); structured-output decode of the model's
      intent value (A-9)
- [ ] T002 Verb vocabulary read from the session manifest — no compiled-in verb
      list in the deliberation path; structural test proves it (card AC #2)
- [ ] T003 Token-only targets: descriptive targets rejected before compose; every
      composed target traceable to a token the mind was given (card AC #7)
- [ ] T004 Loop treats only act_result as fact — the fact handed to M2's
      admission gate derives from act_result, never from intent emission
      (card AC #1); tests green (`go vet` + `go test ./...`)

## Phase 2 — The job-board decision (US2)

- [ ] T005 E3 trigger on `text` percept with `origin: read`; E3 context struct
      (board contents, other claims, player relationship, commitments) rendered
      via `mind/prompt` variable suffix (stable prefix untouched)
- [ ] T006 Claim-or-decline intent with non-empty authored `reason` (card AC #3);
      decline reachable, reason cites the villager's own wants/commitments/
      relationship (card AC #4)
- [ ] T007 Politeness-policing structural check: no compliance gate, cooldown, or
      player-conduct-keyed refusal path exists (card AC #9)

## Phase 3 — Interrupt and window (US3 + US4)

- [ ] T008 §5.5 interrupt: urgent percept cancels in-flight call (RT-2 context
      cancellation), fires no call of its own, enqueues exactly one follow-up
      deliberation whose context includes it; multiple urgents coalesce to one;
      cancel/complete race tested both orders (card AC #5)
- [ ] T009 K=10 situated window: top K−2 by salience-halved-per-day decayed
      weight + 2 seeded serendipity picks from the older half; deterministic
      under seed; graceful under-K degradation (card AC #6)

## Phase 4 — Design checks, gates, and closure

- [ ] T010 Scripted-evening micromanagement check against the fake vendor +
      scripted model: postings get worked without re-posting; refusals carry
      persona-grounded reasons (card AC #8)
- [ ] T011 Full gates: `go vet` + `go test ./...` green; scope clean
- [ ] T012 Wiki re-ground: notes whose sources this branch touches re-verified
      honestly (overview at minimum — mind gains a deliberation loop); CAPSULES
      regenerated if any description changed; freshness probe green
- [ ] T013 Card ACs ticked with citing proofs; board/spec synced at PR time
