# Implementation Plan: Fake body vendor and the protocol-rule harness

**Branch**: `task-0015-fake-vendor-harness` | **Date**: 2026-08-27 | **Spec**: specs/015-fake-vendor-harness/spec.md

## Summary

Grow M1's `mind/seamtest` double into the full `FakeVendor` of protocol §10.1, then
land the six H-tests (each with a mutation check), H-6's flood reproduction with the
printed ratio, the §10.2 canonical end-to-end (closing TASK-0010's AC #5), and an
API-surface test enforcing §10.5's no-read-API discipline.

## Technical Context

**Language/Version**: Go (the mind module's toolchain; `mind/go.mod`). No mod/
changes — this is entirely mind-side test infrastructure.
**Primary Dependencies**: `mind/seam` (Conn, session, ingest, intents),
`mind/seamtest` (the double this task grows — its doc comment names S2 as
successor), `mind/wire` (framing/canonical JSON), `mind/memory` (Gate + Store +
instrument for H-6 counts and §10.2 step 5). Stdlib testing only — no frameworks
(house rule).
**Storage**: none — in-memory, deterministic; `world_time` advances only by script.
**Testing**: `go vet` + `go test ./...` in mind/. Mutation checks follow house
precedent (TASK-0007/0009): lift the rule under a build-tag-free test helper or a
switchable code path and assert the test would fail — recorded in the test itself.
**Constraints**: §10.5 scope discipline is structural: FakeVendor's exported
surface is exactly §10.1; assertions read `.acts` from the test, never through the
mind. H-6's memory counting rides M2's admission gate + instrument, not the vendor.
**Scale/Scope**: test infrastructure; three scripted bodies max.

## Constitution Check

The constitution at `.specify/memory/constitution.md` is an **unfilled template**
(stated plainly; house precedent specs 004–012). Checked against grounding docs:

- body-protocol-v0 §10: shape, default behaviour, H-tests, flood scenario, scope
  discipline implemented verbatim; no protocol semantics altered — PASS.
- decision-0003: the harness is how mind work proceeds before the mod exists —
  first-class task, mind-side, Go — PASS.
- kithcraft-brief minds-are-others: §10.5's no-read-API is that constraint's
  structural defence; enforced by an API-surface test, not a comment — PASS.
- One-task-one-PR; M5/M7/mod scope excluded (spec FR-007) — PASS.

**Post-design re-check**: structure below adds no scope beyond FR-001..FR-007 — PASS.

## Project Structure

```
mind/fakevendor/            # FakeVendor (exported for mind tests across packages)
  fakevendor.go             # §10.1 shape; session over the real seam surface
  fakevendor_test.go        # shape + default-behaviour + API-surface tests
  harness_test.go           # H-1..H-5 with mutation checks
  flood_test.go             # H-6 (§10.4) — ratio printed
  e2e_test.go               # §10.2 canonical end-to-end (closes TASK-0010 AC #5)
```

`mind/seamtest` stays (M1's tests use it); FakeVendor builds on the same
drive-from-outside pattern rather than replacing it — the doc comments cross-link.

## Phase ordering rationale

The vendor's shape first (everything else scripts against it), then the cheap
H-tests (H-1..H-4 are classifier/decode rules), then the two structural ones (H-5
needs intent lifecycle + advance; H-6 needs the memory wiring), then the §10.2
end-to-end and closure — it exercises every prior piece at once.
