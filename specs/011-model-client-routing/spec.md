# Feature Specification: Model client, prompt assembly, tier routing, instrumentation

**Feature Branch**: `task-0011-model-client-routing` · **Spec dir**: `specs/011-model-client-routing`

**Created**: 2026-08-25 · **Status**: Draft

**Input**: TASK-0011 / M4 — `anthropic-sdk-go` against RT-1..RT-7, per-class prompt
assembly as a first-class component, six classes over three tiers, per-class
accounting from the first call. Consumes decision-0003 +
docs/design/llm-routing-and-budget.md (§1.3, §2.3, §4.3, §5, §6.1–6.2).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Every class on its declared tier (Priority: P1)

As an operator, I want each of the six LLM event classes routed to its declared
model tier, so that quality lands where character is at stake and latency where the
scene demands it.

**Why this priority**: Routing §1.3 is the ratified tier table; a class on the wrong
tier is either a quality failure (E1/E6 down) or a latency failure (E4 up).

**Independent Test**: A named test asserts the class→model mapping: E1/E6 → Opus 5,
E2/E3/E4 → Sonnet 5, E5 → Haiku 4.5 (card AC #1), with per-call model selection in
one process (RT-5).

**Acceptance Scenarios**:

1. **Given** a call for any class, **Then** the request carries that class's declared
   model and per-class configuration (E4: streaming, low effort, thinking off, tight
   max_tokens; E6: thinking on, max_tokens 4096 against expected 1200).
2. **Given** E6, **Then** its `max_tokens` ceiling is truncation-aware and never near
   expected output (card AC #6, RT-7's inherited lesson).

### User Story 2 - A prefix that never changes and a bill that gets counted (Priority: P1)

As an operator, I want prompt assembly to keep the stable prefix byte-identical and
every call counted by class, so that caching works and every A-n assumption becomes
a measurement.

**Independent Test**: Card AC #3's byte-identity test; per-class counters report at
session end (card AC #7).

**Acceptance Scenarios**:

1. **Given** two assemblies for the same villager and class at different world times,
   **Then** the stable prefix is byte-identical — no date, day counter, or timestamp
   rendered into it (card AC #3; the 23%-of-the-bill invariant).
2. **Given** prompt assembly, **Then** it is a first-class component implementing
   routing §2.3's stable/variable split per class, not concatenation at call sites
   (card AC #2); cache breakpoints are placed explicitly after the stable prefix
   (RT-3), and E6's prefix is deliberately NOT cached (routing §4.3 — every call
   would be a cache write past TTL).
3. **Given** a session end, **Then** per-class call counts and input/output token
   counts (including cache read/write splits when present) report (card AC #7).

### User Story 3 - Calls that stream, cancel, retry, and parse (Priority: P2)

As a villager's mind, I want in-flight calls cancellable and outputs structured, so
that an urgent percept can supersede a stale thought and an intent is a value, not
prose to interpret.

**Independent Test**: Cancellation test — an in-flight call terminates promptly and
cleanly on context cancellation (card AC #5); structured-output round-trip for
E2/E3/E6 shapes.

**Acceptance Scenarios**:

1. **Given** an in-flight call and a cancelled context, **Then** the call terminates
   promptly, resources are released, and the cancellation is observable (RT-2 — the
   §5.5 interrupt mechanism IS cancellation).
2. **Given** E2/E3/E6, **Then** outputs are structured (RT-4/A-9): parsed values with
   a bounded failure mode, not free prose.
3. **Given** a transient failure, **Then** retry with backoff applies (RT-1/RT-7);
   streaming delivers usable first tokens for E4.
4. **Given** concurrency, **Then** the client supports 3–6 concurrent calls, one per
   villager peak (RT-6), goroutine-per-call with context.

### Edge Cases

- Unit tests mock the transport — no live API calls in `go test` (operator
  checkpoint 3: live-call proofs belong to M3/M6/M7; M4's tests mock).
- A cancelled call's partial usage still lands in the per-class accounting.
- The assembler must make the volatile-content-after-last-breakpoint rule structural:
  the stable-prefix type simply has no field a caller could put a timestamp in.

## Requirements *(mandatory)*

- **FR-001**: `anthropic-sdk-go` client wrapper: streaming, retry with backoff,
  per-call model selection, context-based cancellation, explicit cache-control
  breakpoint placement (RT-1..RT-3, RT-5..RT-7).
- **FR-002**: The six-class registry (E1..E6) binding class → model, latency posture,
  thinking/effort config, max_tokens, structured-output schema where applicable, and
  cache policy (E2/E3/E4 cached; E1/E5/E6 not).
- **FR-003**: Per-class prompt assembly as a first-class component: a stable-prefix
  builder (persona, standing desires/values, manifest in role form, instructions —
  per class) and a variable-context builder, composed per routing §2.3; byte-identity
  of the stable prefix testable by construction.
- **FR-004**: Structured outputs for E2/E3/E6 (A-9): typed result values.
- **FR-005**: Per-class call/token accounting from the first call, reported at
  session end; cache reads/writes counted when the API reports them.
- **FR-006**: Unit tests run against a mocked transport; no network in `go test`.
- **FR-007**: Scope: no prompt *content* beyond assembly structure (M3 owns persona
  genesis text, M5/M6/M7 own their class prompts); no memory reads (M2's store is a
  parallel lane — the assembler takes already-selected context as input).

## Success Criteria *(mandatory)*

- **SC-001**: `go vet` + `go test ./...` green in the `mind/` module.
- **SC-002**: Every card AC (#1–#7) demonstrated by a named test.
- **SC-003**: The byte-identity test fails if anyone renders a timestamp into a
  stable prefix (proven by a deliberate red variant during development).
- **SC-004**: Cancellation terminates an in-flight mocked call promptly and cleanly.

## Assumptions

- decision-0003 fixed Go + `anthropic-sdk-go`; this is the one sanctioned new
  dependency (the evidence rule applies: record version + URL + accessed date).
- Routing §2.3's token figures are estimates (A-10) — the assembler ships the
  *structure*; re-baselining figures is playtest work, out of scope.
- M2 develops in parallel (same lane): the assembler consumes selected-context
  values, not the belief store's API — no import of mind/memory from this task.
