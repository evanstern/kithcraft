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

- [x] T004 Retrieval-frequency weighting over conversation-context selection:
      witnessed death spikes next cycle, decays per cycle, floors at present —
      never deleted (RM-7); deterministic distribution tests (card AC #5);
      exported selector hook for M6's context assembly
      — `DeathCarryWeight(deathCycle, now)` (deathcarry.go) mirrors
      `mind/memory/beliefs.go`'s `effectiveConfidence` read-time-arithmetic
      idiom (RM-5): pure function, no store mutation, no randomness. Spike is
      10.0 (promptworld I's old salience table, "witnessed death — 10★",
      death-mechanics.md §3, reused not re-derived), halving per elapsed
      cycle, floored at `NormalPresence` (1.0) and never below it (RM-7).
      `SelectionWeights(deaths []WitnessedDeath, now)` is the exported
      selector hook (plan.md's "landed here ... so M6 can adopt it at
      merge, no cross-branch edit") — death classification itself stays
      the caller's job (death-mechanics.md §3: a witnessed death is an
      ordinary `sighting`, no special percept type, so this package
      doesn't attempt detection). `TestDeathCarryWeight_SpikeNextCycle`,
      `_LowerTwoCyclesLater`, `_FloorsAtNormalPresence_NeverZero`, and
      `_Decreasing` are the deterministic distribution assertions (no
      sampled draws) covering card AC #5.

## Phase 3 — Archived, not terminated (US4)

- [x] T005 Archival per R-9: archived minds refuse session opens; durable log
      stays readable; body token retired into a persisted never-reissue set
      (card AC #6); death-before-own-consolidation edge covered
      — `Archive` (archive.go) is a shared, append-only JSONL registry
      (`archive.jsonl`, `ArchivePathFor(dir)`) mirroring `Ledger`'s
      replay-then-append shape (T001); `Archive.Archive` is idempotent by
      `MindID` (archival is one-way, one-time). `Ingester.Archived` (new
      field, ingest.go, mirroring the existing `OnPercept` hook idiom) is
      consulted in `HandleConnection` (session.go) both on the connection's
      first `session_open` and on every later multiplexed one, refusing
      fail-closed with the existing `session_close`/`reason:error` shape
      and `detail:"archived_mind"` (same refusal machinery as
      `unsupported_version`/`differing_*`). Archival never touches
      `memory.Log` or `Ledger` — the durable log and any already-
      consolidated nights stay exactly as replay left them, readable
      through their existing API (US4 AC #2); no process-termination
      semantics anywhere (plan.md design decision 5).
      — Deviation: mind identity and body token are treated as the same
      opaque string (`ArchiveRecord.MindID`/`BodyToken` both set from the
      caller's one identifier in every test and call site here) because
      `mind/seam` has no separate body→mind-identity resolution layer yet
      (the whole ingest skeleton is replaced by M5); `Archive.TokenRetired`
      is kept as a distinct method from `IsArchived` for when that
      resolution layer arrives (ponytail, noted at archive.go's header).
      — `TestSessionOpen_ArchivedMind_RefusedOnFirstOpen`,
      `_RefusedOnMultiplex`, and `_NotArchived_Unaffected` (session_test.go)
      cover the seam-side refusal; `TestArchive_LandsAndPersists`,
      `_IdempotentByMindID`, `_UnaffectedMindStaysOpen`, and
      `_DeathBeforeOwnConsolidation` (archive_test.go) cover archival
      itself and the villager-dies-before-its-own-consolidation edge case
      from spec.md's Edge Cases (archival wins; the ledger gains no second
      record; the log's unconsolidated tail survives unchanged and
      readable) — card AC #6.

## Phase 4 — Gates and closure

- [x] T006 Full gates: `go vet` + `go test ./...` green; scope clean
      — both run from `mind/` (the Go module root): `go vet ./...` clean,
      `go test -count=1 ./...` green across all nine packages
      (`cmd/minddaemon`, `consolidate`, `fakevendor`, `llm`, `memory`,
      `persona`, `prompt`, `seam`, `wire`); `git status --short` empty
      before this phase's own commits.
- [x] T007 Wiki re-ground: touched-source notes re-verified honestly
      ([[promptworld-lineage]] — consolidation port lands for real; overview);
      CAPSULES regenerated if descriptions changed; freshness green
      — [[body-protocol-seam]] was genuinely STALE (`mind/seam/session.go`
      changed, T005's archival hook); re-verified against the diff and
      amended with a new "First implementations" paragraph, plus
      `mind/seam/ingest.go` and `mind/consolidate/archive.go` added to its
      sources, then re-pinned. [[promptworld-lineage]] and [[overview]]
      had unchanged sources (RE-PIN-eligible by the mechanical rule) but
      were amended anyway per the phase-3 handoff note — the death-carry
      spike's reuse of I's salience-table number needed an honest
      re-verification against the "forbidden" ruling (no percept-level
      salience field is reinstated; the number is mind-side only), and
      both notes' "what exists" language was stale prose even though
      their sources hadn't moved — then re-pinned. [[v1-demo]]'s sources
      were unchanged and its prose needed no amendment: RE-PIN-only,
      left untouched (nothing to say). CAPSULES.md regenerated (both
      notes' descriptions changed). Freshness gate green for all four
      touched notes; [[villager-brain-api]] remains STALE but is
      pre-existing TASK-0014 debt already inherited from `main` before
      this branch cut — out of this phase's scope, left untouched.
- [x] T008 Card ACs ticked with citing proofs; board/spec synced at PR time
