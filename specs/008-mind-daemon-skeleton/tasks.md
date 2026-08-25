# Tasks: Mind daemon skeleton

**Input**: specs/008-mind-daemon-skeleton/ (spec.md, plan.md)
**Prerequisites**: decision-0004 (accepted), docs/design/seam-wire-v0.md, seam/vectors/

**Organization**: Phases map 1:1 to phase-scoped dispatches. US1 = boundary (spec User
Story 1), US2 = sessions (US2), US3 = ingest/intents (US3).

## Phase 1 — Module, wire codec, vector proof (US1 groundwork)

**Goal**: `mind/` module exists; framing + canonical codec round-trips all 17 vectors.

**Independent test**: `cd mind && go test ./wire/...` — vector suite 17/17 green.

- [x] T001 Create mind/go.mod (module kithcraft/mind, go directive matching toolchain)
      and cmd/minddaemon/main.go stub that parses a --socket flag and exits cleanly
- [x] T002 Implement mind/wire/frame.go: 4-byte BE length-prefix read/write, 1 MiB
      pre-allocation cap, connection-fatal error taxonomy per seam-wire-v0.md §2
- [x] T003 Implement mind/wire/canonical.go: hand-rolled canonical JSON writer
      (C-1..C-10; sorted keys, minimal escaping, literal UTF-8) — encoding/json is
      known non-conformant for output (TASK-0007 finding)
- [x] T004 Implement mind/wire/decode.go: tolerant presence-checked decode — V-1
      (ignore unknown fields), V-2 (unknown enum decodes, flagged for fallback), V-3
      (unknown percept_type retained uninterpreted), V-5 (missing required field →
      malformed, never defaulted)
- [x] T005 Write mind/wire/vectors_test.go: every vector in seam/vectors/ decodes,
      re-encodes byte-exactly (census + roundtrip + refusal behavior, mirroring the
      TASK-0007 harness obligations against THIS codec)

**Checkpoint**: the daemon's codec is proven against the pinned wire.

## Phase 2 — Vendor port, listener, session lifecycle (US2)

**Goal**: The daemon listens on UDS, completes the handshake, survives restart with
honest continuity.

**Independent test**: `go test ./seam/...` session tests green, including the
kill-and-reconnect continuity test.

- [ ] T006 Declare the vendor port interface in mind/seam/port.go (at the consumer)
      and implement the UDS listener in cmd/minddaemon: unlink stale path before bind
      (with the decision-0004 liveness probe), accept loop, per-connection framing
- [ ] T007 Implement mind/seam/session.go: fail-closed version negotiation
      (unsupported_version close), manifest ingest, per-body session multiplexing on
      one connection, byte-identical capabilities check across session_opens,
      session_close semantics
- [ ] T008 Implement continuity per protocol §6.3: reconnect with body-token matching
      (seam-wire-v0.md §1.5), gap reported as gap — a named test restarts the daemon
      mid-session against the double and asserts no backfill

**Checkpoint**: T-4 restart independence is demonstrated, not asserted.

## Phase 3 — Ingest, intents, the double, end-to-end (US3)

**Goal**: Scripted percept stream in, intents out, bookkeeping honest.

**Independent test**: card AC #5's end-to-end named test passes.

- [ ] T009 Implement mind/seam/ingest.go: validate → mutate ordering with V-4 (refuse
      unknown verb at the intent boundary), V-6 (unrecognized/absent origin →
      secondhand classification hook), percept_id dedup (reconnect scope), seq-gap
      shed accounting; named tests prove zero mutation on malformed input (AC #3)
      and future-origin secondhand classification (AC #4)
- [ ] T010 Implement mind/seam/intents.go: pending set, supersedes replacement,
      act_result matching by intent_id, cancel; named tests for each
- [ ] T011 Implement mind/seamtest/double.go: dials the daemon over UDS (or in-process
      net.Pipe behind the same port — T-7), scripts a percept stream with duplicates
      and a seq gap, records emitted intents
- [ ] T012 End-to-end named test (AC #5): daemon starts, session opens against the
      double, scripted stream ingested (dupes dropped, gap accounted), intents
      emitted and acked; plus the AC #6 restart variant

## Phase 4 — Closure: gates, wiki, board

- [ ] T013 Run go vet ./... and go test ./... across the module; fix to green
- [ ] T014 Scope check: git diff origin/main...HEAD touches only mind/, specs/008-*,
      board files, and the runbook log row
- [ ] T015 Wiki: re-verify body-protocol-seam.md (its prose describes the seam as
      contract-only; the first real implementation exists now) — amend honestly,
      grow sources if the note's claims now rest on mind/ code, re-pin, regenerate
      CAPSULES.md if descriptions changed; add an overview.md re-check ("no code
      exists yet" claims are now false — amend)
- [ ] T016 Tick this file, check card ACs now true (backlog task edit TASK-0008
      --check-ac <n>), append phase-done note

## Dependencies

Phase 1 → 2 → 3 → 4 serial (codec before sessions; sessions before ingest e2e).
Within Phase 3: T009/T010 parallel, T011 after the port exists, T012 last.

## Implementation strategy

The codec-first order means the pinned wire is load-bearing from the first commit:
nothing above it can drift from the vectors without a red test.
