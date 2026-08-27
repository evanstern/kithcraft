# Feature Specification: Fake body vendor and the protocol-rule harness

**Feature Branch**: `task-0015-fake-vendor-harness` · **Spec dir**: `specs/015-fake-vendor-harness`

**Created**: 2026-08-27 · **Status**: Draft

**Input**: TASK-0015 / S2 — `FakeVendor` per body-protocol-v0 §10.1 and the six
protocol-rule tests H-1..H-6 (§10.3), including H-6's deliberate 75%-flood
reproduction (§10.4), under §10.5's scope discipline. Grows M1's `mind/seamtest`
double (its doc comment names S2 as the successor). Consumes body-protocol-v0.md
§10 + §2.7 + §4.10 + §5.3/5.6, decision-0003 (the harness as a first-class task),
and closes TASK-0010's deliberately-open AC #5 (the §10.2 canonical end-to-end
against the fake vendor).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Rules that fail on violation instead of living in prose (Priority: P1)

As a future implementer, I want the seam's unprovable rules turned into
failing-on-violation tests, so that a rule that looks like a stylistic preference in
prose cannot be deleted by the next refactor.

**Why this priority**: This is the card's whole reason to exist — V-5/V-6, the
change_report restriction, and SI-1's edges are rules no compiler can check across a
process seam.

**Independent Test**: All six H-tests green, and each turns red when its rule is
lifted (card AC #2) — the red proof is a mutation check per house precedent
(TASK-0007's vector harnesses, TASK-0009's seamRoundTrip).

**Acceptance Scenarios**:

1. **H-1** (V-5): with `strict` on, a percept missing `provenance`, then one missing
   `provenance.origin`, are both rejected at the seam before any state mutation.
2. **H-2** (V-6): with `strict` off, `origin: "dreamt"` and absent `origin` both
   classify secondhand (`direct_perception` false).
3. **H-3**: a `told` percept whose prose swears "I saw this myself, directly,
   firsthand" (`hops: 0`, `source.descriptor: "saw"`) is still secondhand — the
   classifier reads origin and nothing else (§2.7).
4. **H-4**: a percept carrying `"direct": true` alongside `origin: "told"` — the
   unknown field is ignored (V-1) and classification is still secondhand.
5. **H-5**: an intent on an issued-but-gone token is acked `accepted: true` and
   fails as `act_result`/`failed`/`target_gone` only after `advance()` — never
   synchronously; an unissued token is refused `unknown_target` at ack.
6. **H-6** (§10.4): identical three-body script run twice, `restrict_change_reports`
   off then on; asserts `flooded.memory_count > 3 × restricted.memory_count`, zero
   change_reports to actor and to witnesses under restriction, and prints the ratio
   (card AC #3).

### User Story 2 - A vendor double a mind cannot tell from a world (Priority: P1)

As a mind under test, I want the fake vendor to present exactly the protocol surface
and nothing more, so that passing against it means passing against any world.

**Independent Test**: The §10.2 canonical end-to-end runs against FakeVendor
end-to-end — scripted percepts through a real session, asserted acts out, step 5's
epistemic assertion — closing TASK-0010's AC #5 at its literal wording.

**Acceptance Scenarios**:

1. **Given** the §10.2 script (sighting, observation with absence claim, told_fact,
   go_to on secondhand knowledge, advance + resolve, bare-orchard observation),
   **Then** the mind's belief about food at the old orchard has origin class
   secondhand and the mind cannot claim it witnessed apple trees there.
2. **Given** any intent received, **Then** the default is ack `accepted: true`,
   record in `.acts`, and nothing else — the act stays pending until the script
   resolves it (§10.1: waiting is the state minds occupy most).

### User Story 3 - Scope discipline, enforced not promised (Priority: P1)

As the project, I want the fake vendor structurally unable to leak world state to a
mind, so that the minds-are-others constraint's most convenient violation point is
closed (card design check).

**Independent Test**: An API-surface test (house precedent: M2's external-package
surface tests) proves the vendor's exported surface is exactly §10.1's shape — no
read API, no query method, nothing a mind could call to learn world state without
acting (card AC #4).

**Acceptance Scenarios**:

1. **Given** the vendor's exported API, **Then** it is exactly: manifest, open,
   close, emit, advance, acts, resolve, strict, restrict_change_reports — and the
   assertion surface (`.acts`) is for tests, never handed to the mind.
2. **Given** the protocol contract, **Then** the fake has no autonomous behaviour
   (nothing emits unless the script says so; advance moves time and nothing else)
   and no capability the real vendors lack (card AC #5).

### Edge Cases

- `resolve` on an unknown or already-resolved intent_id: script error, loud failure
  (a silent no-op would let a test pass while asserting nothing).
- `emit` after `close`: script error — a closed session emits nothing.
- H-6 memory_count: counted at the mind's admission gate (M2's instrument), not by
  the vendor — the vendor counting memories would be a read API by the back door.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `FakeVendor` in Go under `mind/` implementing §10.1's shape exactly:
  manifest (§6.2 payload), open/close, emit, advance (world_time += n, nothing
  else), `.acts` in receipt order, resolve(intent_id, outcome, reason_code, detail),
  `strict`, `restrict_change_reports` (card AC #1). It drives the mind through the
  real seam surface (growing `mind/seamtest`'s from-the-outside pattern), so the
  mind under test cannot tell it from a world.
- **FR-002**: Default intent behaviour: ack accepted, record, do nothing until
  resolved (§10.1).
- **FR-003**: The six H-tests as named tests (H-1..H-6), each green, each with a
  mutation check proving it turns red when its rule is lifted (card AC #2).
- **FR-004**: H-6 per §10.4: three bodies, one acting repeatedly, identical script
  twice, restriction off/on; the three assertions; ratio printed (card AC #3).
  Memory counts come from the mind side's admission machinery (M2), not the vendor.
- **FR-005**: The §10.2 canonical end-to-end against FakeVendor, including step 5 —
  and on green, TASK-0010's AC #5 is ticked with this test as the citation.
- **FR-006**: Scope discipline (§10.5), enforced by an API-surface test: no read
  API, no autonomous behaviour, no capability beyond the contract (card AC #4, #5).
- **FR-007**: Out of scope: any real-world vendor change (mod/ untouched),
  deliberation (M5), consolidation (M7), protocol extensions.

### Key Entities

- **FakeVendor**: the scripted in-memory vendor — the protocol surface §10.1
  specifies, nothing more.
- **H-test**: a named protocol-rule test binding a rule to a reproduction and a
  mutation check.
- **The flood scenario**: H-6's three-body script — the measured 75% bug,
  reconstructed on demand.

## Success Criteria *(mandatory)*

- **SC-001**: Card AC #1 — FakeVendor implements §10.1's full shape.
- **SC-002**: Card AC #2 — H-1..H-6 green; each red when its rule is lifted.
- **SC-003**: Card AC #3 — H-6 reproduces the flood and prints the ratio.
- **SC-004**: Card AC #4 — no read API; proven by API-surface test.
- **SC-005**: Card AC #5 — no autonomous behaviour, no beyond-contract capability.
- **SC-006**: TASK-0010 AC #5 closed by the §10.2 end-to-end against the fake.
