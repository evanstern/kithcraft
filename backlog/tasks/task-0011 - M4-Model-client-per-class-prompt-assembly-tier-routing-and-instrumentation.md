---
id: TASK-0011
title: >-
  M4 - Model client, per-class prompt assembly, tier routing, and
  instrumentation
status: In Progress
assignee: []
created_date: '2026-08-21 23:37'
updated_date: '2026-08-25 21:01'
labels:
  - mind
  - m-0-build
milestone: m-0
dependencies:
  - TASK-0008
documentation:
  - docs/design/demo-build-plan.md
  - docs/design/llm-routing-and-budget.md
priority: high
ordinal: 11000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
As an operator, I want every model call routed by class, cached at a genuinely stable prefix, and counted from the first call, so that the demo's cost and latency assumptions are measurements rather than estimates.

**Scope boundary.** `anthropic-sdk-go` against RT-1..RT-7: streaming with usable first-token latency, clean in-flight cancellation, **prompt caching with explicit breakpoint placement**, structured outputs for E2/E3/E6, per-call model selection (six classes across three tiers in one process), modest concurrency (3-6, peaking at one per villager), retry with backoff, and a **truncation-aware ceiling on E6** — budget 4,096 against an expected 1,200, and never set E6's `max_tokens` near its expected output. **Per-class prompt assembly as a first-class component**, not concatenation at the call site: six classes x the stable/variable split (routing section 2.3) *is* the caching design and the cost model. Per-class call and token accounting from the first call.

**Done proves.** Each class routes to its declared tier (E1/E6 Opus 5, E2/E3/E4 Sonnet 5, E5 Haiku 4.5). A test asserts the **stable prefix is byte-identical across calls for a given villager and class** — specifically that no date, day counter, or timestamp is rendered into it, which silently destroys 23% of the bill *and* part of E4's latency budget at once. Cancelling an in-flight call terminates it promptly and cleanly. Per-class counters report at session end.

**Depends on.** M1.

**References.** docs/design/demo-build-plan.md section 3.2 (M4) is the plan of record. Ratified surfaces consumed: decision-0003 + docs/design/llm-routing-and-budget.md (six LLM classes over three tiers, RT-1..RT-7, the stable/variable prompt split in section 2.3, the cost model and its A-n assumptions).

**Suggested tier: `sonnet` (next sweep's runbook decides).**

Spec: specs/011-model-client-routing
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Each of the six LLM classes routes to its declared tier: E1/E6 Opus 5, E2/E3/E4 Sonnet 5, E5 Haiku 4.5
- [x] #2 Per-class prompt assembly is a first-class component implementing the stable/variable split, not concatenation at the call site
- [x] #3 A test asserts the stable prefix is byte-identical across calls for a given villager and class, with no date, day counter or timestamp rendered into it
- [x] #4 Streaming, explicit cache-breakpoint placement, structured outputs for E2/E3/E6 and retry with backoff are in place per RT-1..RT-7
- [x] #5 Cancelling an in-flight call terminates it promptly and cleanly
- [x] #6 E6 carries a truncation-aware max_tokens ceiling (4096 against an expected 1200), never set near its expected output
- [x] #7 Per-class call and token counters report at session end
- [x] #8 Spec phase: Phase 1 — Class registry and prompt assembly (US2 groundwork + US1 data)
- [x] #9 Spec phase: Phase 2 — SDK client: streaming, cancel, retry, breakpoints (US3)
- [x] #10 Spec phase: Phase 3 — Accounting and session report (US2 close)
- [x] #11 Spec phase: Phase 4 — Closure: gates, wiki, board
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Phase 4 closure (T009-T011). AC proofs: #1 mind/llm/classes_test.go TestClassTierMapping (E1/E6 Opus5, E2/E3/E4 Sonnet5, E5 Haiku4.5). #2 mind/prompt/assemble.go+shapes.go as a first-class component (no import from mind/llm), composed assembly proven by mind/prompt/assemble_test.go's TestAssemble_* suite. #3 mind/prompt/assemble_test.go TestAssemble_StablePrefixByteIdentical (plus TestAssemble_CatchesRenderedTimestamp, the deliberate red variant per SC-003). #4 mind/llm/client_test.go TestStreamDelivery (streaming), TestBreakpointPlacement (cache breakpoints present for E2/E3/E4, absent for E1/E5/E6), TestRetryOnTransient (backoff); mind/llm/structured_test.go TestParseIntentRoundTrip/TestParseIntentBoundedFailure/TestParseDigestBoundedFailure (structured outputs E2/E3/E6). #5 mind/llm/client_test.go TestSendCancellation. #6 mind/llm/classes_test.go TestE6Ceiling (4096 vs expected 1200, >=2x margin). #7 mind/llm/accounting_test.go TestSessionAccounting (all six classes, cancelled-call partial usage). Phases 1-4 (tasks.md T001-T011) all checked; go vet + go test -count=1 ./... green in mind/ (module kithcraft/mind); scope diff origin/main...HEAD touches only mind/{llm,prompt,go.mod,go.sum}, specs/011-model-client-routing/, backlog/tasks/ — clean. SDK evidence: github.com/anthropics/anthropic-sdk-go v1.58.0, module https://github.com/anthropics/anthropic-sdk-go, accessed 2026-08-25 (decision-0003's one sanctioned new dependency; recorded in mind/go.mod/go.sum and mind/llm/client.go's package doc). Wiki: docs/wiki/overview.md prose amended to name TASK-0011/M4's mind/llm and mind/prompt landing, re-pinned verified_against to b899d179 (this branch's pre-Phase-4 HEAD, the merge of TASK-0010's mind/memory landing); docs/wiki/body-protocol-seam.md and docs/wiki/promptworld-lineage.md reviewed — this branch's own commits (T001-T008, none of which touch docs/wiki or those notes' sources) invalidate no claim in either, so left as-is (not re-pinned). Freshness probe (grounding-wiki cli.mjs freshness) reports 0 new staleness/budget issues from this branch; 3 pre-existing capsule/body-size-budget failures on body-protocol-seam.md and promptworld-lineage.md are unchanged from origin/main (predate TASK-0011, out of this task's scope) — confirmed identical on a clean origin/main checkout.
<!-- SECTION:NOTES:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Tests pass
- [ ] #2 Docs and wiki are updated and pass freshness tests
- [ ] #3 Spec and Backlog are in sync
<!-- DOD:END -->

## Comments

<!-- COMMENTS:BEGIN -->
created: 2026-08-25 20:14
---
Sweep runbook tier: sonnet · model cc/claude-sonnet-5[1m] (default tier — model client against RT-1..RT-7; stable/variable prompt split designed in routing §2.3). Served model recorded at dispatch.
---
<!-- COMMENTS:END -->
