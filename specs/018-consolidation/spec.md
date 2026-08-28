# Feature Specification: Nightly consolidation and the archived dead (E6)

**Feature Branch**: `task-0018-consolidation` · **Spec dir**: `specs/018-consolidation`

**Created**: 2026-08-28 · **Status**: Draft

**Input**: TASK-0018 / M7 — E6 nightly consolidation on Opus 5 (sleep-event
trigger, `world_time` timing), the ported machinery shape (nightly ledger,
ordinal `m1..mN` convention, no-marker-on-failure), the death carry (death
mechanics §3), and ruling R-9 (archived-not-terminated minds). Consumes
decision-0003 + docs/design/llm-routing-and-budget.md (E6 row, §5.4 sleep
window and retry posture, no Batch API, harness T-b), death-mechanics.md §3,
body-protocol-v0.md (RM-7), kithcraft-brief.md (#4 stories told about them).
Plan of record: demo-build-plan.md §3.2 M7.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The night that keeps what mattered (Priority: P1)

As a villager, I want to wake up having kept what mattered about yesterday, so
that a day with the player accumulates into a history instead of evaporating.

**Why this priority**: Consolidation is where a day becomes durable memory —
the whole accumulation thesis rides on it.

**Independent Test**: Against the fake vendor with a scripted model, a scripted
day's admitted buffer consolidates into a digest whose accepted references
resolve back to durable `(tick, hash)` pairs (card AC #2).

**Acceptance Scenarios**:

1. **Given** E6's class config, **Then** it is Opus 5, triggered by the sleep
   event, timed against `world_time` never a wall clock (harness T-b), runs
   inside the sleep window, and is not on the Batch API (card AC #1).
2. **Given** the day's admitted buffer (M2's gate decided eligibility), **When**
   consolidation runs, **Then** memories are presented under the ordinal
   `m1..mN` prompt convention and accepted references map back to `(tick, hash)`
   pairs — the convention IS the identity mechanism (card AC #2).
3. **Given** v1's posture, **Then** no formativeness scoring pass exists: the
   admission gate decides eligibility, E6 decides what mattered (card AC #4).
4. **Given** a digest, **Then** it lands with a marker recording the night as
   consolidated; the consolidated window is excluded from the next night's
   buffer.

### User Story 2 - The night that fails safely (Priority: P1)

As a villager whose mind hit a transport failure at 3am, I want yesterday
undigested rather than half-digested, so that the night is retried, not lost.

**Independent Test**: A consolidation failing mid-call lands no marker; the next
attempt retries the same night (card AC #3).

**Acceptance Scenarios**:

1. **Given** a mid-call failure (transport error, cancelled stream), **Then** no
   marker lands, no partial digest is stored, and the buffer is intact.
2. **Given** the next consolidation trigger, **Then** the unconsolidated night
   is retried; a villager waking undigested is recoverable and invisible.
3. **Given** the truncation lesson (I's digest silently outgrew a 1,024-token
   cap), **Then** the digest's max_tokens posture follows the registry and an
   over-limit response is detected as a failure (no marker), never silently
   truncated into a stored digest.

### User Story 3 - The dead stay conversationally alive, and fade (Priority: P2)

As a player, I want a lost villager to be talked about — a lot at first, less
later, never forgotten — so that permadeath produces stories instead of a
respawn.

**Independent Test**: A witnessed death is retrieved at high frequency in the
following cycle's conversation context, at lower frequency two cycles later,
and is still present — not deleted — well after that (card AC #5).

**Acceptance Scenarios**:

1. **Given** a witnessed death in the buffer (ordinary `sighting` sequence per
   death §3 — no special percept type), **When** cycles pass, **Then** its
   retrieval frequency in conversation-context selection is disproportionately
   high next cycle, lower two cycles later, and the memory remains present
   (RM-7: time alone never deletes) (card AC #5).
2. **Given** the fade, **Then** it is retrieval-frequency weighting, not
   deletion or mutation of the stored record.

### User Story 4 - Archived, not terminated (Priority: P2)

As a survivor, I want my memories of the dead to keep pointing at something
real, so that the stories stay grounded.

**Independent Test**: A dead villager's mind archives: its durable log is
readable, no session opens for it, and its body token is retired and never
reissued (card AC #6, ruling R-9).

**Acceptance Scenarios**:

1. **Given** a villager's death, **When** archival runs, **Then** the mind opens
   no new session; an inbound session attempt for it is refused.
2. **Given** the archive, **Then** the durable log survives and is readable
   (survivors' memories cite it).
3. **Given** the body token, **Then** it is retired and never reissued.

### Edge Cases

- Sleep interrupted mid-consolidation (wake event): in-flight call cancelled →
  no marker → retry next night (same path as transport failure).
- Empty admitted buffer: consolidation is a no-op that still lands a marker
  (nothing to digest is a consolidated night, not a failed one).
- Multiple nights unconsolidated (repeated failures): next success consolidates
  the accumulated buffer; the ledger records which window each digest covers.
- Death of a villager mid-night before its own consolidation: archival wins;
  the dead mind's last night is never consolidated (no session opens for it) —
  its log remains as-is.

## Requirements *(mandatory)*

- **FR-001**: E6 uses the registry config (Opus 5); trigger is the sleep event;
  all timing arithmetic on `world_time`; no Batch API path exists.
- **FR-002**: Nightly ledger with per-night markers; ordinal `m1..mN` prompt
  convention with acceptance mapping to `(tick, hash)`; digest stored as
  event-sourced state (M2 idioms — append-only, reducer).
- **FR-003**: No-marker-on-failure, covering transport failure, cancellation,
  and over-limit responses; retry on next trigger.
- **FR-004**: No formativeness scoring pass anywhere in the path.
- **FR-005**: Death carry as retrieval-frequency weighting over conversation-
  context selection (high → lower → present), no deletion (RM-7).
- **FR-006**: Archival per R-9 — no new session, readable log, token retired and
  never reissued.
- **FR-007**: No live API calls in tests; the model client is mocked/scripted.

## Success Criteria *(mandatory)*

- All six card ACs demonstrated by named tests (`go vet` + `go test ./...`
  green in `mind/`).
- Wiki notes whose sources this PR touches re-verified honestly; CAPSULES
  regenerated if descriptions change; board/spec in sync at PR time.
