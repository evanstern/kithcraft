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

- [x] T004 Body-to-persona binding in Runtime (closing I1's empty-prefix
      ponytail); E6 prefix gains persona text; E2/E4 prefixes bound the same
      way (card AC #3 binding half).
      Investigated what the wire actually carries before choosing a
      mechanism (runtime.go's top-of-file comment records the finding):
      `mod/.../session/Handshake.java`'s MANIFEST is a fixed,
      world-independent constant — identical capabilities for every body,
      by that class's own doc — so session_open carries no cast identity at
      all, and `self_state` (`mod/.../percept/SelfState.java`) carries none
      either. Two real signals exist: (1) a body token that literally IS a
      loaded CastID — the demo/dev-server convention Phase 1's placeholder
      already assumed, promoted here to the real mechanism, bound once at
      attach (`runtime.go`'s `HandleSessionOpen`, guarded by `hasPersona` so
      a reconnect never downgrades an established binding); (2) the dusk
      pairing-signal sighting's `thing.descriptor`
      (`mod/.../brain/PairingSignal.java`'s `content(otherBody, otherName)`
      literally carries the OTHER body's cast display name) —
      `conversation.go`'s `bindPersonaIfUnbound` learns from it
      opportunistically, T005's own wiring. `personaFor` (deliberation.go)
      now reads the cached binding instead of the map directly; E6's
      `runNight` (runtime.go) builds `consolidationPrefixFor` from it,
      closing I1's empty-`ConsolidationStablePrefix` ponytail; E4's
      `exchangeSpeaker` (conversation.go) reuses the SAME `stablePrefixFor`
      E2/E3 already used ("Phase 1's stablePrefixFor grows up"). Proven by
      `TestConsolidationPrefixFor_CarriesPersonaText` (pure-function, E6's
      half) and by `TestDuskExchange_PairSignalConvergesAndRecordsLatency`
      failing closed if either speaker's binding were missing (E4's half) —
      both `mind/cmd/minddaemon/conversation_test.go`. Deviation from
      Phase-1 test scaffolding: `startDeliberationDaemon`
      (deliberation_test.go) mutated `rt.Personas` AFTER the listener
      goroutine started, which Phase 1's lazy per-call map read tolerated
      but T004's bind-at-attach caching turned into a real data race
      (caught by `-race`); fixed by taking the whole `map[string]persona.
      Persona` up front, matching the "populated once at startup, before
      serving begins" invariant `LoadOrGenesisCast`'s own doc already
      stated.
- [x] T005 Dusk exchange off the live pair signal: Slot fill on signal
      percepts, Exchange at convergence between the two live sessions (speak
      intents out per side), abort-discard + live fallback semantics held
      under live timing (card AC #3).
      `mind/cmd/minddaemon/conversation.go`: `isPairSignal` classifies the
      exact shape `PairingSignal.content`/`Sightings.sightingContent`
      compose (a `k:person` sighting doing "walking to the gathering
      place") — recorded bounded heuristic (package doc): DuskPairing fires
      this TWICE per approaching pair (once per perceiver, each naming the
      OTHER member), so this build treats the FIRST signal for a
      (pairID,day) as starting the pregen Fill for its sender (the
      designated opener — always live-attached by construction) and the
      SECOND, from the other side, AS convergence itself — the wire carries
      no separate "arrived"/"met" percept yet. `handlePairSignal` fills M6's
      `Pool.Begin` on the first signal and runs `runExchange` (in-process,
      each speaker's own `Out`/`Session` per plan.md's Risks note) on the
      second; `pendingPair`/`abortPair` implement abort-discard via a
      `pairConvergeTimeout` (30s, well past PairingSignal.LEAD_SECONDS=10s;
      var not const so tests can shorten it) with identity-checked timers
      so a superseded pair's stale timer can never touch a different
      pending pair sharing its key (a documented first-signal race:
      last-write-wins on `rt.pairs[key]`, the loser's Fill/timer are inert).
      `Config.OpeningWait` stays at its zero default (TASK-0014's own
      short-lead finding, already the package's precedent) — the live
      fallback fires whenever pregen isn't ready, which is expected and
      "first-class" per V3's measured 1.82-4.96s lead. Proven by
      `TestDuskExchange_PairSignalConvergesAndRecordsLatency` (convergence
      → live exchange → spoken intent on the opener's OWN session,
      `-race` green) and `TestDuskExchange_UnconfirmedPairAbortsAfterTimeout`
      (abort-discard: no leaked pending pair, never spoken).
- [x] T006 AmbientPool wired: refill on the world_time day crossing Runtime
      tracks, serve via speak intents, IsTargeted/Escalate reachable
      (card AC #4).
      `refillAmbient` (conversation.go) rides the SAME `consolidate.
      SleepTriggered` check `runNight` already reacts to (runtime.go's
      `HandlePercept`), one batched E5 call per villager per crossing.
      Ambient trigger chosen as directed: the smallest live trigger this
      build has for "a player passing a villager" — any `sighting` of a
      `k:person` thing that ISN'T T005's exact pairing-signal shape
      (`handleAmbientTrigger`). Reuses the wire's own optional-field
      semantics for targeting rather than inventing one: `Sightings.
      sightingContent`'s `doing` MAY be null (mod's own doc) — absent/empty
      is the generic case (`converse.IsTargeted("")` false → `AmbientPool.
      Serve`, falling back to `Stall` when the pool has nothing); present
      is "about something specific" (`IsTargeted` true → `converse.
      Escalate`). `speakLine` composes the speak intent via the SAME
      `wireVendor` T001 already wired. Proven by
      `TestAmbientTrigger_ServeAndEscalate`: day-crossing refill, a generic
      sighting served from the pool, a specific one escalated — both as
      live speak intents in order, `-race` green.
- [x] T007 FirstTokenLatency surfaced into session-report.log (watch-list #6
      closed) (card AC #5).
      `bodyStore.turnLatencies` (runtime.go) collects every dusk-exchange
      turn's `converse.Turn.FirstTokenLatency`, filed under its OWN
      speaker's body (`conversation.go`'s `recordLatencies`) since a shared
      Exchange's turns alternate between two different bodies. `report.go`
      adds a per-villager section (raw values, arrival order — smallest
      useful form, mirroring the existing E6-instrument section's own
      per-body/sorted style). Proven by
      `TestDuskExchange_PairSignalConvergesAndRecordsLatency`'s report
      assertion (the spoken villager's line is present and non-empty).

## Phase 3 — Proofs (FR-006 + FR-007)

- [x] T008 Real-binary fake-vendor proofs: board text percept → live
      claim/decline with authored reason; urgent mid-deliberation →
      coalesced follow-up; scripted pair signal → pre-generated opening +
      multi-turn exchange with latency in the report; pool serve/refresh
      (card AC #6)
      Three of four proofs already existed from Phases 1-2 (deliberation.go/
      conversation.go's own T00x notes); the one gap was the dusk-exchange
      proof only ever exercising ONE turn (the reply's own ClosingMarker
      closed the exchange after the opener) where spec.md US3 scenario 1
      names a "multi-turn exchange" explicitly. Closed by extending
      `sseTextServer` (conversation_test.go) to answer calls in sequence
      (textMessageServer's own per-call-index cycling, applied to SSE) and
      rewriting `TestDuskExchange_PairSignalConvergesAndRecordsLatency` to
      script the ClosingMarker onto the SECOND call only, with a 50ms pause
      between the two pairing signals biasing deterministically toward the
      pregen-served path (documented in the test's own doc comment: the
      live-fallback path is equally valid per T005's package doc, but
      racing it here would make the test non-deterministic, not prove
      anything this proof needs) — now asserts exactly 2 model hits (pregen
      fill + Petra's one live call, no wasted fallback), two alternating-
      body intents, and FirstTokenLatency recorded for BOTH speakers in the
      session report. 5x repeated `-run` green, `-race` clean. FR-006's SET
      is enumerated as a single named, greppable entry point —
      `TestFR006_ProofSet` (conversation_test.go, doc-only/instant-skip) —
      citing all four proofs by name: (a)
      `TestLiveE3Deliberation_PerceptInIntentOutWithAuthoredReason`
      (deliberation_test.go — decline itself proven at the mind/deliberate
      package level, the wire mechanism is verb-agnostic, not duplicated
      here); (b) `TestLiveInterrupt_CancelsInFlightAndProducesOneFollowUp`
      (deliberation_test.go); (c)
      `TestDuskExchange_PairSignalConvergesAndRecordsLatency` (this file,
      above); (d) `TestAmbientTrigger_ServeAndEscalate` (this file — day-
      crossing refill already proven there, no separate gap).
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
