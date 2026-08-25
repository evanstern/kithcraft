---
id: TASK-0011
title: >-
  M4 - Model client, per-class prompt assembly, tier routing, and
  instrumentation
status: In Progress
assignee: []
created_date: '2026-08-21 23:37'
updated_date: '2026-08-25 20:14'
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
- [ ] #1 Each of the six LLM classes routes to its declared tier: E1/E6 Opus 5, E2/E3/E4 Sonnet 5, E5 Haiku 4.5
- [ ] #2 Per-class prompt assembly is a first-class component implementing the stable/variable split, not concatenation at the call site
- [ ] #3 A test asserts the stable prefix is byte-identical across calls for a given villager and class, with no date, day counter or timestamp rendered into it
- [ ] #4 Streaming, explicit cache-breakpoint placement, structured outputs for E2/E3/E6 and retry with backoff are in place per RT-1..RT-7
- [ ] #5 Cancelling an in-flight call terminates it promptly and cleanly
- [ ] #6 E6 carries a truncation-aware max_tokens ceiling (4096 against an expected 1200), never set near its expected output
- [ ] #7 Per-class call and token counters report at session end
<!-- AC:END -->

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
