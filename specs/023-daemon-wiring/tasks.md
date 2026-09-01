# Tasks: Wire deliberation and conversation into the live daemon (M8)

**Spec dir**: `specs/023-daemon-wiring` · **Branch**: `task-0023-daemon-wiring`

## Phase 1 — Deliberation live (US1 + US2)

- [x] T001 The live Vendor: wire-envelope implementation over `seam.Conn`
      promoted from TASK-0016's evening-test adapter shape; intents out on
      the body's own session, act_results resolved back through the loop.
      `mind/cmd/minddaemon/vendor.go` (`wireVendor`, reading conn/session/seq
      live off `*bodyStore` under `rt.mu` so a reconnect is picked up rather
      than captured stale); `Loop.Deliver` wired in `runtime.go`'s
      `HandlePercept` for every `act_result` percept. Proven by
      `TestLiveE3Deliberation_PerceptInIntentOutWithAuthoredReason`
      (`mind/cmd/minddaemon/deliberation_test.go`), through the real
      listener/wire, not a package double.
- [x] T002 E2/E3 composition in Runtime: trigger classification over existing
      percepts (§2.2 list + TriggerE3), per-body deliberation loop (one at a
      time + trigger queue), persona + K=10 window context (store→WindowItem
      snapshot adapter — the smallest exported read that serves it); ErrDone
      convention honored; nil-client no-op with log (card AC #1).
      `mind/cmd/minddaemon/deliberation.go`: `triggerE2` (self_state/notable
      sighting/completed act_result, ponytailed against §2.2's list),
      `classifyAndTrigger`/`startDeliberationLoop`/`runDeliberation` (one
      goroutine + buffered trigger channel per body, design decision 2);
      `windowSnapshot` maps `memory.Log.Events()` — already the smallest
      exported read, closing M5's noted deviation without a new export —
      into `[]deliberate.WindowItem` (uniform placeholder Salience,
      ponytailed against routing §1.3's own no-scoring-pass-in-v1 note);
      `personaFor` is Phase 1's placeholder body-token==CastID binding
      (Phase 2/T004 owns the real one); `singleRoundProposer` exercises
      ErrDone deterministically on round 2. Needed one small additive glue
      change outside `cmd/minddaemon`: `seam.Ingester.OnSessionOpen` (new
      hook, `mind/seam/ingest.go`+`session.go`) — capabilities arrive only on
      session_open, never on a percept, and loop.go's own Vendor doc
      requires `Config.Verbs` come from the body's actual declared set,
      so Runtime needed a way to learn them; proven at the seam layer by
      `TestOnSessionOpen_FiresForFirstFrameAndMultiplex`. Nil-client and
      no-bound-persona skip paths proven by
      `TestDeliberation_NilClient_LogsAndSkips` /
      `TestDeliberation_NoBoundPersona_LogsAndSkips`. E3 half of card AC #1
      proven by `TestLiveE3Deliberation_PerceptInIntentOutWithAuthoredReason`.
- [x] T003 §5.5 interrupt on live ingest: IsUrgent → Interrupt registered on
      the per-body loop; cancel + no-own-call + exactly-one coalesced
      follow-up, proven through the daemon path (card AC #2).
      `classifyAndTrigger` routes `deliberate.IsUrgent` percepts to
      `bs.interrupt.Urgent` before any E2/E3 check; `runDeliberation`
      registers/clears the in-flight call's `cancel` func and Drains
      coalesced urgents into the next deliberation's context. Proven by
      `TestLiveInterrupt_CancelsInFlightAndProducesOneFollowUp`
      (`deliberation_test.go`) through the real daemon path: an in-flight
      call to a gated fake model server is cancelled (no intent sent for
      it), the urgent itself fires no call, and exactly one follow-up
      lands as a live intent — green under `-race`. Multi-urgent
      coalescing into one enqueue is already proven deterministically at
      the package level
      (`mind/deliberate/interrupt_test.go`'s
      `TestInterrupt_CoalescesMultipleUrgentsIntoOneEnqueue`); racing two
      urgent percepts over the wire against this body's own async Drain
      timing would be a timing bet, not a proof, so the daemon-level test
      scopes to what T003 actually adds — the live-ingest wiring — and is
      documented as such in the test's doc comment.

## Phase 2 — Conversation live (US3 + US4)

- [ ] T004 Body-to-persona binding in Runtime (closing I1's empty-prefix
      ponytail); E6 prefix gains persona text; E2/E4 prefixes bound the same
      way (card AC #3 binding half)
- [ ] T005 Dusk exchange off the live pair signal: Slot fill on signal
      percepts, Exchange at convergence between the two live sessions (speak
      intents out per side), abort-discard + live fallback semantics held
      under live timing (card AC #3)
- [ ] T006 AmbientPool wired: refill on the world_time day crossing Runtime
      tracks, serve via speak intents, IsTargeted/Escalate reachable
      (card AC #4)
- [ ] T007 FirstTokenLatency surfaced into session-report.log (watch-list #6
      closed) (card AC #5)

## Phase 3 — Proofs (FR-006 + FR-007)

- [ ] T008 Real-binary fake-vendor proofs: board text percept → live
      claim/decline with authored reason; urgent mid-deliberation →
      coalesced follow-up; scripted pair signal → pre-generated opening +
      multi-turn exchange with latency in the report; pool serve/refresh
      (card AC #6)
- [ ] T009 Dev-server observation (research/live-wiring-observation.md): at
      least one live E2/E3 deliberation and one dusk exchange end-to-end;
      bounded checks, honest not-observed records where substrate timing
      bites (card AC #7); adjacent-gap judgments recorded (card AC #8)

## Phase 4 — Gates and closure

- [ ] T010 Full gates: go vet + go test -count=1 ./... green; gradle green
      if mod touched; scope clean; no new Mixins/protocol surface
      (card AC #8)
- [ ] T011 Wiki re-ground: overview's not-closed addendum shrinks honestly
      (E2–E5 now wired; what remains for the evening); body-protocol-seam if
      seam-facing claims change; CAPSULES if descriptions change; freshness
      green
- [ ] T012 Card ACs ticked with citing proofs; board/spec synced at PR time
