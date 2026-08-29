# Feature Specification: Demo configuration and the two run targets (I1)

**Feature Branch**: `task-0021-demo-config` · **Spec dir**: `specs/021-demo-config`

**Created**: 2026-08-29 · **Status**: Draft

**Input**: TASK-0021 / I1 — one documented sequence bringing up a demo-ready
world: the daemon binary and the mod jar as independently started artifacts,
mind-restart independence (T-4), the R-1/R-3/R-6 knobs as config, cast seeding
via M3's genesis, and session-end instrumentation. Consumes decision-0003 +
llm-routing-and-budget.md (T-4, §7.1 nine-dusk ruling, per-class
instrumentation), death-mechanics.md §6.2 (grief-period + danger-tuning),
decision-0001 (two artifacts). Plan of record: demo-build-plan.md §3.4 I1.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - One sequence, demo-ready (Priority: P1)

As an operator, I want one documented sequence that brings up a demo-ready
world with the cast seeded, so that running the demo is a decision about when,
not a research project.

**Independent Test**: The documented command sequence, followed verbatim,
yields a running server with three personas seeded and bound (card AC #1).

**Acceptance Scenarios**:

1. **Given** the repo at main, **When** the operator follows the documented
   sequence (a RUNBOOK/README under docs/ or the spec dir — one file, ordered
   commands, prerequisites named incl. ANTHROPIC_API_KEY for genesis),
   **Then** the daemon binary and the mod jar start as two independently
   started artifacts (decision-0001) and the server comes up with the cast
   present.
2. **Given** no persisted personas, **When** the daemon starts with genesis
   enabled, **Then** M3's genesis runs for the three cast members (E1 on
   Opus 5, real calls — the operator's key; zero-call dry path documented for
   rehearsal) and binds them to bodies (persona-to-body binding via the
   existing cast/token machinery).
3. **Given** persisted personas from a prior run, **Then** startup re-binds
   (M3's Load re-bind path) without re-genesis.

### User Story 2 - The mind restarts; the villagers keep their memories (Priority: P1)

As an operator, I want to restart the daemon mid-session without the villagers
losing themselves, so that a crash is an inconvenience, not a lobotomy.

**Independent Test**: Restarting the daemon mid-session and reconnecting
leaves the villagers with their memories — the decision-0003 acceptance check
promoted to a test (card AC #2, T-4).

**Acceptance Scenarios**:

1. **Given** a live session with admitted memories, **When** the daemon is
   killed and restarted, **Then** the vendor's reconnect (V1's dial-retry,
   proven at 1474 failed dials) re-opens sessions with `body_continuous` and
   the mind's persisted state (M2's log, M3's personas, M7's ledger) is
   re-loaded — memories survive.
2. **Given** the gap while the mind was down, **Then** it is reported as a
   gap, never backfilled (M1's continuity rule).

### User Story 3 - The rulings are knobs, and the knobs are config (Priority: P2)

As an operator, I want every tunable the design ruled on to be config with the
ruled default, so that a per-run choice never needs a rebuild.

**Independent Test**: R-1, R-3, R-6 all inspectable as config; danger tuning
off by default; nothing above is a code constant (card ACs #3, #4, #6).

**Acceptance Scenarios**:

1. **Given** the world config, **Then** the vanilla daylight cycle is kept and
   RECORDED as ruling R-1 in the run doc (nine cycles/evening — representative
   play, 27 consolidations), not left as an unexamined default.
2. **Given** the config surface, **Then** the grief-period knob (R-3, exists
   since V5: `kithcraft.griefPeriodTicks`) and a danger-tuning knob (R-6) are
   present; danger tuning is OFF by default with its purpose documented (a
   recorded-take-only choice).
3. **Given** the code, **Then** a config audit shows each knob reads
   configuration, not a constant.

### User Story 4 - The session tells you what it cost (Priority: P2)

As an operator, I want the per-class counters and the E6-input instrument
reported at session end, so that the budget model meets reality.

**Independent Test**: Session end (daemon shutdown or session_close) emits the
per-class call/token counters (M4's Accounting) and the E6-input-tokens
instrument (M2's) in a readable report (card AC #5).

**Acceptance Scenarios**:

1. **Given** a session with model calls, **When** it ends, **Then** a report
   lists calls/input/output tokens per class (E1..E6) and the E6-input
   instrument's buffer-size-per-villager-day series.
2. **Given** a session with zero calls (stub/rehearsal), **Then** the report
   still emits, zeroed — the reporting path is unconditional.

### Edge Cases

- Daemon starts before the server / after it: either order works (the vendor
  dials and retries; the daemon listens — document both).
- Genesis interrupted mid-cast: already-written personas persist (0444);
  restart resumes genesis for the missing members only.
- Key absent at genesis: fail with a clear message naming the prerequisite;
  no partial cast bound.

## Requirements *(mandatory)*

- **FR-001**: One documented command sequence (single file), two independently
  started artifacts, prerequisites named.
- **FR-002**: Cast seeding runs M3's genesis + body binding; re-bind on
  restart; resume on partial genesis.
- **FR-003**: Mind-restart independence proven live: kill/restart daemon
  mid-session, memories survive, gap reported not backfilled.
- **FR-004**: R-1 recorded as a ruling in the run doc; R-3 and R-6 knobs as
  config; danger tuning off by default; config-not-constant audit.
- **FR-005**: Session-end report: per-class counters + E6-input instrument,
  unconditional.
- **FR-006**: No new Mixins (surface capped at 4); no protocol extension.
- **FR-007**: Unit tests mock model calls; the live genesis path exists for
  the operator's run (M3's live harness precedent) — rehearsal path zero-call.

## Success Criteria *(mandatory)*

- All six card ACs demonstrated: tests + a live documented bring-up
  observation (research doc, honest not-observed records where live proof
  falls short — the restart-independence check observed for real).
- Wiki notes whose sources this PR touches re-verified honestly; CAPSULES
  regenerated if descriptions change; board/spec in sync at PR time.
