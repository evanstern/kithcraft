# Tasks: Demo configuration and the two run targets (I1)

**Spec dir**: `specs/021-demo-config` · **Branch**: `task-0021-demo-config`

## Phase 1 — The daemon runtime loop (US1 machinery + US2 machinery)

- [x] T001 `cmd/minddaemon` real runtime: open M2 log + M7 ledger/archive per
      villager; persona load-or-genesis at start (M3's Load/Genesis; resume
      partial genesis; clear failure when key absent); listener serving
      sessions with the Archive hook consulted on session_open (closes
      TASK-0018's named daemon-wiring deferral)

      Done: `mind/cmd/minddaemon/runtime.go` (new) — `Runtime` assembles
      already-tested packages: `NewRuntime` opens the shared `consolidate.
      Archive` under `<rundir>/villagers` eagerly (Client/Digester wired
      from `ANTHROPIC_API_KEY` alone, nil together in its absence — the
      zero-call rehearsal path); `LoadOrGenesisCast` calls `persona.Load`
      per demo cast id (`persona.DemoCast()`, promoted to an exported
      function in `mind/persona/persona.go` so main.go and the package's
      own tests share one source of truth — `persona_external_test.go`'s
      want-list extended for it, read-only precedent per Validate's own
      TASK-0013 addition), genesis-resumes only the missing ids via
      `persona.Genesis`+`WriteOnce`, and fails loudly naming the missing
      ids when `ANTHROPIC_API_KEY` is unset and genesis is actually needed
      — no partial cast is ever written. `bodyOrOpen` lazily opens each
      body's `memory.Log`/`consolidate.Ledger` (keyed by the session's own
      body token — mind identity == body token, matching `consolidate/
      archive.go`'s existing ponytail convention; persona CastID and body
      token are deliberately disjoint namespaces in this build, see T003).
      `HandlePercept` (wired as `Ingester.OnPercept` in `main.go`) runs
      `memory.Gate.Decide` and appends admitted percepts to that body's log
      — real memory admission, not the skeleton's intent-emitting hook
      (which stays test-only in e2e_test.go; M5 still owns deliberation).
      `main.go` wires `ing.Archived = rt.Archive.IsArchived` (closes
      TASK-0018's named deferral) and a `-rundir` flag (ponytail: further
      config surface — env vars, R-3/R-6 knobs, a genesis on/off switch —
      is T004/Phase 2's job, not duplicated here).
      Tests: `mind/cmd/minddaemon/runtime_test.go` (new) —
      `TestNewRuntime_OpensArchiveUnderRunDir`,
      `TestLoadOrGenesisCast_NoKeyAndMissingPersonas_ClearFailure` (zero
      network calls: no httptest server even started),
      `TestLoadOrGenesisCast_ResumesPartialGenesis` (two of three personas
      pre-seeded; asserts exactly 1 E1 hit and that the pre-seeded persona
      was re-bound via Load, not touched by genesis).

- [x] T002 Sleep-trigger wiring: RunNight invoked off the session's sleep
      signal on world_time; no-marker retry semantics reachable end-to-end
      (mock-model test through the daemon loop)

      Done: no wire-level "sleep" message exists (body-protocol-v0.md has
      no such percept_type, and FR-006 forbids adding one), so the sleep
      signal is derived purely from world_time already carried on every
      seam message — `consolidate.SleepTriggered(prev, cur)` (new,
      `mind/consolidate/cycle.go`) plus `consolidate.CycleTicks = 24000`
      (R-1's ruling: the kept vanilla daylight cycle, a constant not a
      knob per plan.md design decision 4). `Runtime.HandlePercept` tracks
      each body's last-observed world_time and calls `consolidate.RunNight`
      when a cycle boundary crosses, with an empty `prompt.
      ConsolidationStablePrefix{}` (ponytail: persona text belongs once a
      body-to-persona identity layer exists — M5 — the empty prefix still
      exercises the real E6 call/parse/ledger path end to end). A nil
      Digester (no key) logs and skips rather than panicking.
      Tests: `TestSleepTriggered` (`mind/consolidate/consolidate_test.go`,
      table-driven boundary-crossing unit proof) and
      `TestSleepTriggerEndToEnd_NoMarkerOnFailureThenRetrySucceeds`
      (`mind/cmd/minddaemon/runtime_test.go`) — a real listener/Ingester/
      HandlePercept path, scripted Digester (no SDK, no network): the
      first cycle crossing's Digest call fails and lands no ledger record
      (buffer preserved), the second crossing's Digest call succeeds and
      lands one record whose window covers both crossings — the
      no-marker-on-failure retry semantics proven end-to-end through the
      daemon loop, not by calling RunNight directly.

- [x] T003 TASK-0020's token-namespace finding: fix findClaimant/BodySession
      token mismatch here if it is glue-sized, else record the explicit
      hand-off to I2 with the observation cited — orchestrator-visible either
      way

      Call: FIXED here — glue-sized, exactly the option board-observation.md
      §4 itself named ("findClaimant also consulting the live attach-token
      registry, or claims carrying the seat token"). Root cause: `Live
      BuildExecution.tick`/`BoardVisit.tick` both take a single `Function
      <UUID, Optional<String>>` body-token lookup, but `KithcraftMod` only
      ever wired `duskPairing::bodyTokenFor` (the dusk-gathering SEAT-token
      namespace) into it — a live claim's body instead carries `BodySession`
      's own `ground.issueBody` token, a disjoint namespace `findClaimant`'s
      `equals` check could never match.
      Fix: `mod/src/main/java/dev/kithcraft/mod/live/BodyTokenLookups.java`
      (new) — a pure `combine(primary, secondary)` over `Function<UUID,
      Optional<String>>`, no Minecraft type involved. `BodySession` now
      exposes `villagerId()`/`body()` (the UUID it attached to and its own
      body token). `KithcraftMod.bodyTokenLookup()` (new private method)
      combines `duskPairing::bodyTokenFor` with a lookup over the single
      live-attached `BodySession`, and both `boardVisit.tick`/
      `buildExecution.tick` call sites now consume that one combined
      lookup instead of `duskPairing::bodyTokenFor` alone. No new Mixin, no
      protocol change — three files touched, all existing surfaces.
      Test: `mod/src/test/java/dev/kithcraft/mod/live/
      BodyTokenLookupsTest.java` (new, plain JUnit, no Minecraft bootstrap)
      — proves the fallback fires exactly when the seat-token namespace
      has no answer, including the named bug shape verbatim: a UUID the
      seat-token lookup never registered is still found through the
      live-attach lookup.
      Scope note: this fixes the LOOKUP gap only — whether a live build
      then actually places blocks still also needs `Activity.WORK` timing
      to cooperate (board-observation.md §4's other open question) and a
      live re-observation is Phase 3/T007's job, not re-derived here.

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
