# Tasks: Nightly consolidation and the archived dead (E6)

**Spec dir**: `specs/018-consolidation` · **Branch**: `task-0018-consolidation`

## Phase 1 — The nightly digest (US1 + US2)

- [x] T001 `mind/consolidate/`: nightly ledger — sleep-event trigger, windows
      and retry arithmetic on `world_time` only (no wall clock in the package);
      E6 from the registry (Opus 5), asserted in a test; no Batch API path
      (card AC #1)
      — `RunNight(ctx, log, ledger, digester, stable, worldTime)` is the
      trigger entry point (caller supplies `worldTime` from the sleep event);
      `TestE6IsOpus5AndOffline` asserts the registry; `TestNoBatchAPIPath`
      greps the package's own non-test sources for the SDK's batch-service
      identifier so the absence stays checked, not just stated.
- [x] T002 Digest cycle: admitted buffer → ordinal `m1..mN` prompt → accepted
      references mapped back to durable `(world_time, hash)` pairs; digest
      stored event-sourced; consolidated window excluded from the next buffer
      (card AC #2); no formativeness scoring pass anywhere (card AC #4)
      — `renderBuffer`/`mapReferences` in cycle.go; `Ledger` (ledger.go) is
      the event-sourced store (append-only JSONL + `Watermark()` reducer,
      M2's shape — plain `encoding/json` rather than `mind/wire`'s canonical
      encoder, since a night record carries no wire-protocol identity to
      verify, only mind/percept content does). `llm.Digest` widened
      (mind/llm/structured.go) with a `References []string` field completing
      M4's ponytail. `TestRunNight_OrdinalMappingRoundTrip` and
      `TestRunNight_ConsolidatedWindowExcludedNextNight` cover AC #2; no
      scoring-pass code exists anywhere in the package (AC #4 is an absence,
      documented at cycle.go's header).
- [x] T003 No-marker-on-failure: transport failure, cancellation, and
      over-limit detection all land no marker, keep the buffer intact, and
      retry on the next trigger; empty night lands a marker; multi-night
      accumulation covered (card AC #3)
      — covered by `TestRunNight_TransportFailureLandsNoMarker`,
      `TestRunNight_CancellationLandsNoMarker`,
      `TestRunNight_OverLimitLandsNoMarker` (truncation detected via SDK
      `StopReasonMaxTokens`, proven at the client boundary by
      `TestClientDigester_TruncationDetected`),
      `TestRunNight_EmptyNightLandsMarker`, and
      `TestRunNight_MultiNightAccumulationAfterFailures`.

## Phase 2 — The death carry (US3)

- [ ] T004 Retrieval-frequency weighting over conversation-context selection:
      witnessed death spikes next cycle, decays per cycle, floors at present —
      never deleted (RM-7); deterministic distribution tests (card AC #5);
      exported selector hook for M6's context assembly

## Phase 3 — Archived, not terminated (US4)

- [ ] T005 Archival per R-9: archived minds refuse session opens; durable log
      stays readable; body token retired into a persisted never-reissue set
      (card AC #6); death-before-own-consolidation edge covered

## Phase 4 — Gates and closure

- [ ] T006 Full gates: `go vet` + `go test ./...` green; scope clean
- [ ] T007 Wiki re-ground: touched-source notes re-verified honestly
      ([[promptworld-lineage]] — consolidation port lands for real; overview);
      CAPSULES regenerated if descriptions changed; freshness green
- [ ] T008 Card ACs ticked with citing proofs; board/spec synced at PR time
