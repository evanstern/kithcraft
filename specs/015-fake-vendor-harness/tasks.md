# Tasks: Fake body vendor and the protocol-rule harness

**Input**: specs/015-fake-vendor-harness/ (spec.md, plan.md)
**Prerequisites**: mind/seam + mind/seamtest (M1, PR #13), mind/memory (M2, PR #15),
body-protocol-v0.md §2.7/§4.10/§5.3/§5.6/§10

**Organization**: Phases map 1:1 to phase-scoped dispatches.

## Phase 1 — FakeVendor shape and scope discipline (US2 + US3 groundwork)

**Goal**: The §10.1 vendor exists, indistinguishable from a world, structurally
unable to leak.

**Independent test**: `go vet` + `go test ./...` green in mind/; API-surface test
proves the exported shape is exactly §10.1.

- [x] T001 Implement mind/fakevendor: §10.1's full shape (manifest, open/close,
      emit, advance, .acts, resolve, strict, restrict_change_reports) driving the
      mind through the real seam surface per seamtest's from-the-outside pattern
      (FR-001); default intent behaviour ack-record-wait (FR-002); loud script
      errors (resolve on unknown/resolved id, emit after close)
- [x] T002 Shape + default-behaviour tests: manifest is §6.2-valid, advance moves
      time and nothing else, intents stay pending until resolved
- [x] T003 API-surface test (M2's external-package precedent): exported surface is
      exactly §10.1 — no read API, no autonomous behaviour, no beyond-contract
      capability (FR-006, card AC #4/#5)

**Checkpoint**: a scripted world that cannot be told from a real one.

## Phase 2 — The cheap rules: H-1..H-4 (US1 groundwork)

**Goal**: The classifier/decode rules bound to failing-on-violation tests.

**Independent test**: `go test` green; each of H-1..H-4 has a mutation check
proving red-when-lifted.

- [x] T004 H-1 (V-5): strict mode — missing provenance, then missing
      provenance.origin, both rejected before any state mutation; mutation check
- [x] T005 H-2 (V-6) + H-3 (pure classifier) + H-4 (no direct on the wire): unknown
      /absent origin classifies secondhand; prose swearing firsthand stays
      secondhand; "direct": true ignored per V-1; mutation checks for each

**Checkpoint**: four rules that can no longer be deleted silently.

## Phase 3 — The structural rules: H-5 and H-6 (US1 close)

**Goal**: The existence-oracle and flood rules, reproduced and bound.

**Independent test**: `go test` green; H-6 prints the ratio.

- [ ] T006 H-5: issued-but-gone token accepted at ack, fails target_gone only after
      advance(); unissued token refused unknown_target at ack; mutation check
      (FR-003)
- [ ] T007 H-6 (§10.4): three-body flood scenario, identical script with
      restrict_change_reports off then on; memory counts via M2's gate +
      instrument; assert flooded > 3× restricted, zero to actor, zero to
      witnesses; print the ratio; mutation check (FR-004, card AC #3)

**Checkpoint**: the two rules that look stylistic and are structural, proven.

## Phase 4 — Canonical end-to-end, gates, wiki, board (US2 close + closure)

**Goal**: §10.2 runs against the fake; TASK-0010's AC #5 closes; every gate green.

**Independent test**: full `go vet` + `go test ./...` green; freshness probe green;
card ACs ticked with citations.

- [ ] T008 The §10.2 canonical end-to-end against FakeVendor: the five-step script
      through a real session, step 5's epistemic assertion (secondhand origin
      class, no witnessed claim) (FR-005)
- [ ] T009 Close TASK-0010 AC #5 with the T008 test as citation (backlog CLI
      in-worktree, note referencing the deliberate carry from PR #15)
- [ ] T010 Gates + wiki + board: go vet/test green; body-protocol-seam.md (and any
      touched notes) re-verified and honestly re-pinned — the seam note's harness
      claims now real; CAPSULES regenerated if descriptions changed; freshness
      probe green; card ACs ticked with citations; runbook log row updated

**Checkpoint**: S2 done — the seam's rules fail on violation, on demand.
