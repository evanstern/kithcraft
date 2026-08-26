# Implementation Plan: Model client, prompt assembly, tier routing, instrumentation

**Branch**: `task-0011-model-client-routing` | **Date**: 2026-08-25 | **Spec**: specs/011-model-client-routing/spec.md

## Summary

The mind's model layer: an `anthropic-sdk-go` wrapper meeting RT-1..RT-7, a
six-class registry binding each E-class to its tier and posture, per-class prompt
assembly implementing the stable/variable split (the caching design), structured
outputs for E2/E3/E6, and per-class accounting from the first call. All unit tests
against a mocked transport.

## Technical Context

**Language/Version**: Go (module `kithcraft/mind`, go1.26.4 — the M1 module).
**Primary Dependencies**: `anthropic-sdk-go` — the one sanctioned new dependency
(decision-0003 §8.2: proven at this exact workload in promptworld I, including
caching and stop-reason mapping). Version + URL + accessed date recorded in the PR
per the evidence rule. Everything else stdlib.
**Storage**: None — accounting is in-memory, reported at session end.
**Testing**: `go test` with a mocked transport (the SDK's option.WithHTTPClient or
an interface seam at our wrapper boundary — chosen at implementation; no network).
**Constraints**: routing §5's per-class latency postures are *configuration
correctness* here (the right flags on the right class), not measured latency — live
latency proofs belong to M6's dispatch (operator checkpoint 3).
**Scale/Scope**: 6 classes, 3 tiers, 3–6 concurrent calls.

## Constitution Check

The constitution at `.specify/memory/constitution.md` is an **unfilled template**
(stated plainly; house precedent specs 004–010). Checked against grounding docs:

- llm-routing-and-budget.md §6.1: RT-1..RT-7 each land as a requirement/test;
  §1.3's tier table is consumed as ratified, not re-decided — PASS.
- §2.3/§4.3: the stable/variable split is the component's shape; E6's no-cache rule
  encoded in the class registry — PASS.
- decision-0003: Go + anthropic-sdk-go; `llm` package's provider/breaker/budget
  *shape* is source material, not a vendored port — PASS.
- One-task-one-PR; no prompt content, no memory reads (FR-007) — PASS.

**Post-design re-check**: structure below adds no scope beyond FR-001..FR-007 — PASS.

## Project Structure

### Documentation (this feature)

```text
specs/011-model-client-routing/
├── README.md  spec.md  plan.md  tasks.md  (this cycle)
└── (no research.md — SDK fitness was Phase-1-assessed under TASK-0004; routing is ratified)
```

### Source Code (repository root)

```text
mind/
├── llm/
│   ├── client.go         # SDK wrapper: streaming, retry/backoff, cancellation, breakpoints
│   ├── client_test.go    # mocked-transport tests: cancel, retry, stream, breakpoint placement
│   ├── classes.go        # E1..E6 registry: model, posture flags, max_tokens, cache policy
│   ├── classes_test.go   # AC #1 routing test; E6 ceiling test; E4 posture test
│   ├── structured.go     # structured-output shapes + parse for E2/E3/E6 (A-9)
│   ├── accounting.go     # per-class call/token counters, session-end report
│   └── accounting_test.go
└── prompt/
    ├── assemble.go       # per-class assembler: StablePrefix + Variable, composed
    ├── assemble_test.go  # byte-identity across calls/world-times; no-timestamp-field proof
    └── shapes.go         # typed inputs per class (persona block, manifest-in-roles, etc.)
```

**Structure Decision**: two packages. `llm` owns the wire to Anthropic; `prompt`
owns assembly and imports nothing from `llm` (the assembler produces values the
client sends — testable without any SDK type). The stable-prefix type carries only
stable fields by construction: the byte-identity invariant is enforced by the type
surface first and the test second. Class configuration is data (one table in
classes.go), mirroring routing §1.3 so a tier change is a one-line diff.

## Complexity Tracking

No violations. Deliberate simplifications: no circuit breaker (ponytail: ceiling is
demo-scale call volume where retry-with-backoff suffices; add the breaker shape from
promptworld I's llm package if sustained-failure behaviour matters post-demo). The
accounting reports at session end only — no live dashboard.
