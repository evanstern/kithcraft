# Tasks: Demo configuration and the two run targets (I1)

**Spec dir**: `specs/021-demo-config` · **Branch**: `task-0021-demo-config`

## Phase 1 — The daemon runtime loop (US1 machinery + US2 machinery)

- [ ] T001 `cmd/minddaemon` real runtime: open M2 log + M7 ledger/archive per
      villager; persona load-or-genesis at start (M3's Load/Genesis; resume
      partial genesis; clear failure when key absent); listener serving
      sessions with the Archive hook consulted on session_open (closes
      TASK-0018's named daemon-wiring deferral)
- [ ] T002 Sleep-trigger wiring: RunNight invoked off the session's sleep
      signal on world_time; no-marker retry semantics reachable end-to-end
      (mock-model test through the daemon loop)
- [ ] T003 TASK-0020's token-namespace finding: fix findClaimant/BodySession
      token mismatch here if it is glue-sized, else record the explicit
      hand-off to I2 with the observation cited — orchestrator-visible either
      way

## Phase 2 — Knobs and the report (US3 + US4)

- [ ] T004 Config surface: daemon flags/env (socket, run dir, genesis mode);
      mod danger-tuning knob (R-6) present + OFF by default; grief knob (R-3)
      verified config; config-not-constant audit test (card ACs #4, #6)
- [ ] T005 Session-end report: per-class call/token counters (M4 Accounting)
      + E6-input-tokens instrument (M2), emitted unconditionally on shutdown
      and appended to a run log (card AC #5); zero-call path tested

## Phase 3 — The documented sequence and the live proof (US1 + US2 live)

- [ ] T006 docs/design/demo-runbook.md: prerequisites (JDK/Gradle, Go, key),
      the one ordered command sequence, both start orders documented, R-1
      recorded as a ruling (nine cycles, 27 consolidations), rehearsal
      (zero-call) path documented (card ACs #1, #3)
- [ ] T007 Live bring-up observation (research/bringup-observation.md):
      sequence followed verbatim, three personas seeded/bound (live genesis
      or persisted re-bind, recorded which); daemon killed + restarted
      mid-session — memories survive, gap reported not backfilled (card
      AC #2); honest not-observed records where live proof falls short

## Phase 4 — Gates and closure

- [ ] T008 Full gates: go vet + go test ./... green; gradle build + test
      green; scope clean
- [ ] T009 Wiki re-ground: touched-source notes re-verified honestly
      (overview — the daemon becomes runnable-for-real; body-protocol-seam
      if session wiring changes seam claims; promptworld-lineage if the
      daemon-wiring deferral note needs closing); CAPSULES regenerated if
      descriptions changed; freshness green
- [ ] T010 Card ACs ticked with citing proofs; board/spec synced at PR time
