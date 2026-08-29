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

- [x] T005 E3 trigger on `text` percept with `origin: read`; E3 context struct
      (board contents, other claims, player relationship, commitments) rendered
      via `mind/prompt` variable suffix (stable prefix untouched)
      — `e3.go`: `TriggerE3` checks `percept_type == "text"` and
      `provenance.origin == "read"` (§4.7); `E3Context.VariableContext()`
      renders the four §2.3 fields through `prompt.VariableContext`, never
      touching `prompt.DeliberationStablePrefix`. `e3_test.go`'s
      `TestE3Context_StablePrefixByteUntouched` assembles the same stable
      prefix against two E3Contexts with entirely different board/claims/
      relationship/commitments content and asserts `Assembled.Stable` is
      byte-identical (mirrors `prompt.TestAssemble_StablePrefixByteIdentical`).
- [x] T006 Claim-or-decline intent with non-empty authored `reason` (card AC #3);
      decline reachable, reason cites the villager's own wants/commitments/
      relationship (card AC #4)
      — `board_test.go` scripts a `Class: llm.E3` `Loop` with `claim`/
      `decline` manifest verbs and proves the model's authored reason
      reaches `Vendor.SendIntent` unaltered for both verbs (decline
      reachable, no special-cased rejection). Deviation from the "no
      changes to Phase 1 files" note: `loop.go`'s `Run` gained a
      three-line check — an intent with an empty `Reason` is now rejected
      before compose, the same before-compose posture `Tokens.ValidateTarget`
      already has for targets (`llm.ParseIntent` only requires a non-empty
      verb, leaving `Reason` unchecked). This generalizes FR-004 ("every
      intent carries an authored reason") across all classes rather than
      gating it to E3 only, since the requirement is stated generally and
      every Phase 1 scripted intent already carried a reason (verified
      against `loop_test.go` before making the change — no existing test
      broke). `TestLoop_EmptyReasonIntent_RejectedBeforeCompose` covers it.
- [x] T007 Politeness-policing structural check: no compliance gate, cooldown, or
      player-conduct-keyed refusal path exists (card AC #9)
      — `politeness_test.go`'s `TestNoPolitenessPolicingInDeliberationPath`,
      in `TestNoCompiledInVerbVocabulary`'s style (T002): greps every
      non-test `.go` file in `mind/deliberate` for
      compliance/cooldown/conduct/politeness-shaped identifiers. Scope is
      this package only, matching T002's own precedent for "the
      deliberation path" this PR owns.

## Phase 3 — Interrupt and window (US3 + US4)

- [x] T008 §5.5 interrupt: urgent percept cancels in-flight call (RT-2 context
      cancellation), fires no call of its own, enqueues exactly one follow-up
      deliberation whose context includes it; multiple urgents coalesce to one;
      cancel/complete race tested both orders (card AC #5)
      — `interrupt.go`: `IsUrgent` reads §2.8's top-level `urgency` field
      (not `percept_type`); `Interrupt` is a mutex'd state machine
      (`Register`/`Urgent`/`Drain`) holding no `Proposer`/`Vendor`
      reference, so no model call can fire from `Urgent` by construction.
      `Urgent` cancels the registered `context.CancelFunc` (a no-op if the
      call already completed — Go's own contract) and fires `onEnqueue`
      at most once per coalescing window; `Drain` atomically hands back
      every urgent buffered since the last `Drain` and reopens the
      window, so any urgent arriving between the enqueue signal and the
      caller actually starting the follow-up still lands in that same
      follow-up's context (US3 AC #3's coalescing scenario). No changes
      to `loop.go` — `Run`'s existing `case <-ctx.Done(): return
      res, ctx.Err()` (Phase 1) is the only cancellation hook needed.
      `interrupt_test.go` covers both race orders named in this box: cancel
      wins (`TestInterrupt_CancelsInFlightCall_NoOwnModelCall` — a
      genuinely blocked `propose` is cancelled, `Run` returns wrapping
      `context.Canceled`) and completion wins
      (`TestInterrupt_CompletionWinsRace_StillCoalescesFollowup` — `Deliver`
      resolves the round first, `Run` finishes normally, and the urgent's
      later `cancel()` call is exercised as a no-op) — both under `go test
      -race`, plus `TestInterrupt_CoalescesMultipleUrgentsIntoOneEnqueue`
      (5 concurrent `Urgent` calls, `onEnqueue` fires exactly once,
      `Drain` returns all 5) and `TestInterrupt_DrainOpensNewCoalescingWindow`.
- [x] T009 K=10 situated window: top K−2 by salience-halved-per-day decayed
      weight + 2 seeded serendipity picks from the older half; deterministic
      under seed; graceful under-K degradation (card AC #6)
      — `window.go`'s `SelectWindow(Snapshot, now, seed)` is the pure
      function plan.md decision 3 calls for. Deviation: the protocol
      explicitly forbids a `salience`/`weight`/`importance` field on any
      percept (body-protocol-v0.md §2.8), so `WindowItem`/`Snapshot` are a
      new, self-contained mind-side type this file owns rather than a
      selector over `memory.Store`/`memory.Log` directly (neither exposes
      enumeration or a salience concept today) — wiring a real store into
      a `[]WindowItem` snapshot is later-phase work, the same posture
      tasks.md's Phase 1 note takes toward live trigger wiring. Decay
      mirrors `mind/memory/beliefs.go`'s `effectiveConfidence` shape
      (`salience * 0.5^(age/dayLength)`, future-dated `ObservedAt` clamped
      to age 0) but halves per configurable `DayLength` (mind
      configuration, `memory.Instrument`'s `dayLength` posture) rather
      than per a `Config`-keyed half-life. "Older half" is the whole
      snapshot's oldest half by age (not just the unpicked remainder);
      serendipity candidates exclude anything already in the top K−2, so
      duplicates are impossible by construction and serendipity silently
      yields fewer than 2 (or 0) once the store is small enough that nothing
      unpicked remains — no flag needed. `window_test.go`: 8 tests including
      `TestSelectWindow_TopKMinus2ByDecayedWeight`,
      `TestSelectWindow_SerendipityFromOlderHalf`,
      `TestSelectWindow_Deterministic` (same seed twice → identical
      selection), `TestSelectWindow_GracefulUnderK`,
      `TestSelectWindow_PartialSerendipity_NoPanic` (9 items → exactly 1
      of 2 serendipity slots fills), `TestSelectWindow_Empty`, and
      `TestSelectWindow_DecayHalvesPerDayOfAge` (a controlled 3-item
      weight tie/flip that only holds under exact per-day halving, not
      just age-monotonic ordering).

## Phase 4 — Design checks, gates, and closure

- [x] T010 Scripted-evening micromanagement check against the fake vendor +
      scripted model: postings get worked without re-posting; refusals carry
      persona-grounded reasons (card AC #8)
      — `evening_test.go`: `TestEveningOfPostings_WorkGetsDoneWithoutRepostingOrPolicing`
      drives three board postings through a real `mind/fakevendor.FakeVendor` over
      the real `net.Pipe`-backed `seam.Conn` wire (a `wireVendor` adapter stamps the
      envelope exactly as `loop.go`'s `Vendor` doc describes a real wiring — Phases
      1-3's tests used `VendorFunc` doubles directly, no wire in between). Each
      posting is answered with exactly one intent (`FakeVendor.Acts()` == 3 for 3
      postings — no re-posting needed) and, once resolved, the villager's Loop
      completes without a second round. No two of the two claims and one decline
      share a reason string; the decline's is grounded in a named prior commitment
      ("I promised Mira...to mind the stall"), not a canned refusal.
- [x] T011 Full gates: `go vet` + `go test ./...` green; scope clean
      — both green across the whole `mind/` module (10 packages); the evening test
      also passes under `-race -count=20`. `git status --short` showed only
      `evening_test.go` before that commit.
- [x] T012 Wiki re-ground: notes whose sources this branch touches re-verified
      honestly (overview at minimum — mind gains a deliberation loop); CAPSULES
      regenerated if any description changed; freshness probe green
      — `overview.md` and `promptworld-lineage.md` amended (mind/deliberate/ now
      exists; toolloop's REQUEST/FACT/gate shape graduated from source material to
      a real port) and re-pinned to this branch's Phase 4 HEAD; CAPSULES regenerated
      for both changed descriptions; the freshness gate script
      (`grounding-wiki/0.57.0/gates/freshness.mjs`) exits 0. `body-protocol-seam.md`
      and `v1-demo.md` checked (both grep-matched "mind/") — no cited source touched
      and no claim invalidated (deliberation sits above the seam, not a new
      wire/vendor landing either note tracks) — left untouched.
- [x] T013 Card ACs ticked with citing proofs; board/spec synced at PR time
