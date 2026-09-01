# Tasks: Wire deliberation and conversation into the live daemon (M8)

**Spec dir**: `specs/023-daemon-wiring` · **Branch**: `task-0023-daemon-wiring`

## Phase 1 — Deliberation live (US1 + US2)

- [ ] T001 The live Vendor: wire-envelope implementation over `seam.Conn`
      promoted from TASK-0016's evening-test adapter shape; intents out on
      the body's own session, act_results resolved back through the loop
- [ ] T002 E2/E3 composition in Runtime: trigger classification over existing
      percepts (§2.2 list + TriggerE3), per-body deliberation loop (one at a
      time + trigger queue), persona + K=10 window context (store→WindowItem
      snapshot adapter — the smallest exported read that serves it); ErrDone
      convention honored; nil-client no-op with log (card AC #1)
- [ ] T003 §5.5 interrupt on live ingest: IsUrgent → Interrupt registered on
      the per-body loop; cancel + no-own-call + exactly-one coalesced
      follow-up, proven through the daemon path (card AC #2)

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
