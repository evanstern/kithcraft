# Tasks: Deliberation and the job-board decision (E2, E3)

**Spec dir**: `specs/016-deliberation` · **Branch**: `task-0016-deliberation`

## Phase 1 — The bounded loop (US1)

- [x] T001 `mind/deliberate/`: loop skeleton — compose intent via `seam.Pending`,
      REQUEST/FACT/gate mapping onto intent/intent_ack/act_result; bounded
      iterations (config with default); structured-output decode of the model's
      intent value (A-9)
      — `loop.go`: `Loop.Run` ports toolloop's *shape* (round → REQUEST →
      await FACT → hand to sink → repeat, bounded) onto intent/act_result;
      decode reuses `llm.ParseIntent` (no new Intent type). Note: the spec
      does not pin exact multi-round semantics (when does a deliberation
      propose a *second* intent?) — resolved here as: a `Proposer` signals
      `ErrDone` when it has nothing further, mirroring toolloop's
      `model_done`; `DefaultMaxIterations=5` is a placeholder bound
      (`ponytail:` comment in loop.go) pending a real multi-round E2/E3
      trigger, which is out of this phase's scope per plan.md's Risks note.
- [x] T002 Verb vocabulary read from the session manifest — no compiled-in verb
      list in the deliberation path; structural test proves it (card AC #2)
      — `manifest.go`'s `ManifestVerbs` is the only place a verb name is
      read; `manifest_test.go`'s `TestNoCompiledInVerbVocabulary` greps
      every non-test source file in the package for a hardcoded core-verb
      literal (fails the build if one is added), plus a behavioural test
      proving two different manifests yield two different accepted sets.
- [x] T003 Token-only targets: descriptive targets rejected before compose; every
      composed target traceable to a token the mind was given (card AC #7)
      — `tokens.go`'s `Tokens`/`Observe`/`ValidateTarget`; a bare string or
      an unseen place/thing/body token is rejected in `Loop.Run` before
      `seam.Pending.Compose` is ever called.
- [x] T004 Loop treats only act_result as fact — the fact handed to M2's
      admission gate derives from act_result, never from intent emission
      (card AC #1); tests green (`go vet` + `go test ./...`)
      — `Loop.Deliver` is the *only* call site that invokes `Config.OnFact`;
      `TestLoop_TreatsOnlyActResultAsFact` asserts no fact fires between
      send and delivery, and `TestLoop_FactWiresIntoAdmissionGate` wires a
      real `memory.Gate.Decide` as the sink end-to-end. `go vet ./...` and
      `go test ./...` both green in `mind/` (16 new tests in
      `mind/deliberate`, whole-module suite unaffected).

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
