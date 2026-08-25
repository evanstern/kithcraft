# Tasks: Model client, prompt assembly, tier routing, instrumentation

**Input**: specs/011-model-client-routing/ (spec.md, plan.md)
**Prerequisites**: mind/ module (M1, merged), decision-0003,
llm-routing-and-budget.md §1.3/§2.3/§4.3/§5/§6.1

**Organization**: Phases map 1:1 to phase-scoped dispatches.

## Phase 1 — Class registry and prompt assembly (US2 groundwork + US1 data)

**Goal**: The six-class table and the assembler exist; byte-identity is provable
before any SDK code.

**Independent test**: `go test ./prompt/... ./llm/...` — assembly and registry
tests green (no SDK yet in the loop).

- [x] T001 Implement mind/llm/classes.go: E1..E6 registry as data — model ID,
      latency posture flags (streaming/effort/thinking), max_tokens (E6: 4096 vs
      1200 expected), cache policy (E2/E3/E4 cached, E1/E5/E6 not), structured
      output on/off — mirroring routing §1.3/§5
- [x] T002 Implement mind/prompt/shapes.go + assemble.go: typed stable-prefix
      inputs per class (no time-shaped field exists on the stable side), variable
      context builder, composed assembly per §2.3
- [x] T003 Named tests: AC #1 class→tier mapping; AC #3 byte-identity of the stable
      prefix across two assemblies at different world times (plus a deliberate red
      variant proving the test catches a rendered timestamp); AC #6 E6 ceiling

**Checkpoint**: the caching design exists as types and tests, independent of the SDK.

## Phase 2 — SDK client: streaming, cancel, retry, breakpoints (US3)

**Goal**: The anthropic-sdk-go wrapper meets RT-1..RT-3/RT-5..RT-7 against a mocked
transport.

**Independent test**: `go test ./llm/...` — cancellation, retry, and breakpoint
tests green with no network.

- [x] T004 Add anthropic-sdk-go to mind/go.mod (record version + URL + accessed
      date in the PR per the evidence rule); implement mind/llm/client.go: request
      construction from class config, explicit cache-control breakpoint after the
      stable prefix, streaming, retry with backoff, per-call model selection
- [x] T005 Mock the transport at the wrapper's seam; named tests: AC #5 prompt
      cancellation of an in-flight call (context cancel → prompt clean
      termination), retry-on-transient, stream delivery, breakpoint present for
      cached classes and absent for E6
- [x] T006 Implement mind/llm/structured.go: E2/E3/E6 structured-output shapes and
      parsing with a bounded failure mode (AC #4's structured-outputs half)

**Checkpoint**: RT-1..RT-7 each demonstrated by a named mocked test.

## Phase 3 — Accounting and session report (US2 close)

**Goal**: Every call counted by class from the first call.

**Independent test**: accounting tests green; a scripted session of mocked calls
produces the session-end report.

- [x] T007 Implement mind/llm/accounting.go: per-class call count, input/output
      tokens, cache read/write splits when reported, cancelled-call partial usage
      counted; session-end report (card AC #7)
- [x] T008 Named test: scripted mocked session across all six classes reports
      correct per-class figures, including a cancelled call's partials

## Phase 4 — Closure: gates, wiki, board

- [x] T009 go vet ./... + go test ./... green across the module; scope check: diff
      touches only mind/, specs/011-*, board files, runbook log row
- [x] T010 Wiki: re-verify notes whose sources this PR touches (body-protocol-seam's
      mind/ sources; overview; promptworld-lineage's llm-shape claims if amended) —
      amend honestly, re-pin, regenerate CAPSULES.md if descriptions changed
- [x] T011 Tick this file, check card ACs now true (backlog CLI in-worktree), append
      phase-done note

## Dependencies

Phase 1 → 2 → 3 → 4 serial (registry/assembler before client; client before
accounting's cancelled-call case). No dependency on M2 (parallel lane, FR-007).

## Implementation strategy

Assembly-first: the byte-identity invariant is the task's highest-value test and
needs no SDK — prove the caching design before paying for the wire.
