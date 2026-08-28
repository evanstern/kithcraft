# Feature Specification: Death, danger, and what remains (V5)

**Feature Branch**: `task-0019-death-remains` · **Spec dir**: `specs/019-death-remains`

**Created**: 2026-08-28 · **Status**: Draft

**Input**: TASK-0019 / V5 — verified preconditions (R-4, R-5) first; siege
suppression; conversion-cancel Mixin; authored remains (grave, belongings
bundle, optional tend-grave board entry); grief period as config (R-3); token
discipline. Consumes docs/design/death-mechanics.md (§1 admitted/suppressed,
§2 remains, §3 memory carry, §6.2 open items), decision-0002 +
entity-implementation-comparison.md (bounded Mixin budget), body-protocol-v0.md
(§4.10 change_report restriction, token discipline), kithcraft-brief.md
(micromanagement, politeness-policing). Plan of record: demo-build-plan.md §3.3
V5, rulings R-3/R-4/R-5.

**Escalation trigger (named, binding)**: if the siege suppression point is not
where death §1 assumes, or suppression needs more than a targeted injection —
STOP after recording the verification findings; that is an operator checkpoint
(runbook checkpoint 4), never a silent Mixin-surface growth past
decision-0002's bound.

## User Scenarios & Testing *(mandatory)*

### User Story 0 - Verify before building (Priority: P0)

As a future implementer, I want R-4/R-5 verified against the actual 26.2 engine
before any death code is written, so that the design stands on checked surface.

**Independent Test**: Findings recorded in the spec dir BEFORE implementation
commits (card AC #1; runbook host-addition line).

**Acceptance Scenarios**:

1. **Given** the 26.2 server jar, **Then** it is verified (javap/source
   evidence, same standard as brain-26.2.md) whether POI re-claim has natural
   lag after `releaseAllPois()` (R-4), where the zombie-siege trigger sits and
   whether it can be Mixin-suppressed with one targeted injection, and whether
   a 3-villager cast meets village-eligibility thresholds at all (R-5).
2. **Given** the findings, **Then** they land as `research/death-26.2.md` in
   this spec dir; if the escalation trigger fires, the lane stops there.

### User Story 1 - Walls that matter, deaths that are earned (Priority: P1)

As a player, I want the walls and torches I built to be the reason my friends
are still here in the morning.

**Independent Test**: On a dev server, no siege ever fires (card AC #2);
admitted death vectors remain untouched vanilla (card ACs #10, #11).

**Acceptance Scenarios**:

1. **Given** the siege trigger point verified in US0, **When** the suppression
   Mixin lands, **Then** sieges are suppressed regardless of eligibility and a
   dev-server observation window records zero sieges (card AC #2).
2. **Given** zombie conversion (ruled equivalent to death), **Then** one
   conversion-cancel Mixin makes conversion terminal; the total Mixin surface
   (V3's two + V5's additions) stays inside decision-0002's committed bound
   (~4), asserted by MixinConfigTest's enumeration (card AC #3).
3. **Given** the design checks, **Then** nothing is added to villager
   self-preservation (no feeding/escort/vigilance surface — card AC #10) and
   no engine guardrail exists on friendly fire (card AC #11) — structural
   absence checks.

### User Story 2 - Remains, authored because free is invisible (Priority: P1)

As a survivor's neighbour, I want a grave and their things to be there in the
morning, so that the loss has a place.

**Independent Test**: A villager killed by a zombie leaves a mod-placed named
grave at the death site (or nearest safe buildable surface) with a belongings
bundle beside it, no villager agency required (card ACs #4, #5).

**Acceptance Scenarios**:

1. **Given** a villager death at a location, **Then** the mod places a named
   grave marker there or at the nearest safe buildable surface — always, with
   no villager involvement (card AC #4).
2. **Given** the villager's hidden inventory, **Then** it is captured BEFORE
   vanilla destroys it and placed at the grave as an ordinary
   `roles: ["storage"]` thing named for its owner (card AC #5).
3. **Given** V4's job-board mechanism (not yet merged — see plan for the
   decoupling), **Then** an optional "tend <name>'s grave" entry is offered
   through the same channel a posted job rides, takeable or ignorable
   (card AC #6).

### User Story 3 - The grief period and the token discipline (Priority: P2)

As a villager who lost a neighbour, I want their bed left alone for a while,
so that the space they occupied stays theirs a little longer.

**Independent Test**: The dead villager's bed and job site stay unclaimed for
the configured grief period (default one in-game cycle per R-3), config not
constant (card AC #7); the dead token is never reissued (card AC #9).

**Acceptance Scenarios**:

1. **Given** a death, **Then** the bed and job-site POIs are held unclaimed for
   the grief period — a config value defaulting to one day/night cycle — using
   whatever R-4 verification found about natural POI re-claim lag.
2. **Given** the dead villager's body token, **Then** it is retired and never
   reissued; the grave (or converted mob) gets a NEW body token (card AC #9).

### User Story 4 - Death travels the ordinary channels (Priority: P2)

As an absent villager, I want to learn of a death the way I learn anything —
by coming home to it — so that my memory of it is honestly secondhand.

**Independent Test**: A witnessing villager receives ordinary `sighting`
percepts (no magic death broadcast); an absent one receives a `change_report`
with `change: "gone"` on return plus a `sighting` of the grave (card AC #8).

**Acceptance Scenarios**:

1. **Given** a witnessed death, **Then** the witness's percepts are ordinary
   sightings of the dying villager (thing/doing/origin:"saw") — no death
   percept type exists (death §3).
2. **Given** an absent villager returning, **Then** §4.10's delivery
   restriction produces a `change_report` (`change: "gone"`) plus a sighting of
   the grave-thing.

### Edge Cases

- Death over lava/void: grave at nearest safe buildable surface (the "or
  nearest" clause is load-bearing).
- Death with empty inventory: grave still placed; bundle placed empty or
  omitted — recorded as a deviation either way.
- Two deaths same night: independent graves, bundles, grief holds, tokens.
- Conversion mid-flight (already-converting entity at suppression landing):
  cancel path still terminal; no cured-villager path exists in v1.

## Requirements *(mandatory)*

- **FR-001**: R-4/R-5 verification findings recorded before implementation
  commits; escalation trigger honored (STOP, surface) if it fires.
- **FR-002**: Siege suppression via targeted Mixin; zero sieges on dev server.
- **FR-003**: Conversion-cancel Mixin; total Mixin surface within
  decision-0002's bound, enumerated in MixinConfigTest.
- **FR-004**: Mod-placed named grave + pre-capture belongings bundle
  (`roles: ["storage"]`, owner-named), no villager agency.
- **FR-005**: Optional tend-grave board entry riding the job-board channel.
- **FR-006**: Grief period config (default one cycle) holding bed + job site.
- **FR-007**: Token discipline — retired-never-reissued dead token, new token
  for grave/converted mob.
- **FR-008**: Death percepts ride existing channels only (sightings,
  change_report); no new percept type, no protocol extension.
- **FR-009**: Structural absence checks for the two spell-breakers.

## Success Criteria *(mandatory)*

- All eleven card ACs demonstrated: unit/integration tests where automatable
  (`gradle build` + `gradle test` green), documented dev-server observation
  where live behaviour is the proof (recorded in the PR per the runbook's
  dev-server-proofs gate).
- Wiki notes whose sources this PR touches re-verified honestly; CAPSULES
  regenerated if descriptions change; board/spec in sync at PR time.
