---
id: TASK-0010
title: 'M2 - Event-sourced memory, the belief store, and the episodic admission gate'
status: To Do
assignee: []
created_date: '2026-08-21 23:37'
labels:
  - mind
  - m-0-build
milestone: m-0
dependencies:
  - TASK-0008
documentation:
  - docs/design/demo-build-plan.md
  - docs/design/body-protocol-v0.md
  - docs/design/llm-routing-and-budget.md
priority: high
ordinal: 10000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
As a villager, I want to know only what I saw or was told, with provenance, so that what I say at dusk is honestly mine and not something I was handed.

**Scope boundary.** Reimplement event-sourced memory to the contract (~400 lines per decision-0003): append-only log, immutability **enforced in the schema, not in convention**, state as a reducer. Deliberately **not** carried: promptworld I's world-event vocabulary (the mind never sees it under SI-1) and its `log_format_version` migration chain (whose determinism-for-replay justification died with I). The private, provenance-stamped map — PM-1's spine, the reason two villagers see different worlds — kept distinct from the vendor's resolution index, which the mind never reads. RM-1..RM-7 reimplemented to the contract, porting nothing: witnessed claims, coerce-never-reject, secondhand never beating fresher firsthand, read-time confidence and freshness as arithmetic on `world_time`, and **only a correction, a death, or a witnessed removal deletes a fact — time alone never does**. The deterministic **episodic admission gate** (routing section 6.3): admit on urgency >= `notable`, any percept involving another body or the player, any `act_result` on an intent the mind authored a `reason` for, any `told_fact` or `text`, any first sighting of a `kind` or `place`; drop repeated `background` sightings of already-known things. Plus the one long-run instrument: **E6 input tokens per villager over time**, not dollars.

**Done proves.** The canonical end-to-end (protocol section 10.2) against the fake vendor, including its step 5: a mind told about the orchard cannot durably claim it *saw* apple trees there. The admission gate's instrument reports buffer size per villager-day. An attempt to mutate a logged event fails at the type level, not at review.

**Depends on.** M1.

**Design check — minds-are-others (constraint, structural).** The belief store has **no external write path**. Nothing outside the mind — not the vendor, not the player, not a debug command — may author a belief.

**References.** docs/design/demo-build-plan.md section 3.2 (M2) is the plan of record. Ratified surfaces consumed: docs/design/body-protocol-v0.md (SI-1, SI-5, PM-1, RM-1..RM-7, the canonical end-to-end in section 10.2), decision-0003 + docs/design/llm-routing-and-budget.md (event-sourced memory reimplemented not ported; the episodic admission gate in section 6.3; the E6-input-tokens instrument).

**Suggested tier: `sonnet` (next sweep's runbook decides).**
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Event-sourced memory is append-only with immutability enforced in the schema, not by convention: an attempt to mutate a logged event fails at the type level
- [ ] #2 The private, provenance-stamped map is kept distinct from the vendor's resolution index, which the mind never reads
- [ ] #3 RM-1..RM-7 hold: witnessed claims, coerce-never-reject, secondhand never beats fresher firsthand, confidence and freshness computed at read time from world_time, and only a correction, a death or a witnessed removal deletes a fact
- [ ] #4 The deterministic episodic admission gate admits per routing section 6.3 and drops repeated background sightings of already-known things
- [ ] #5 The canonical end-to-end (protocol section 10.2) passes against the fake vendor, including step 5: a mind told about the orchard cannot durably claim it saw apple trees there
- [ ] #6 The E6-input-tokens-per-villager instrument reports admitted buffer size per villager-day
- [ ] #7 Design check (minds-are-others): the belief store has no external write path from the vendor, the player or a debug command
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Tests pass
- [ ] #2 Docs and wiki are updated and pass freshness tests
- [ ] #3 Spec and Backlog are in sync
<!-- DOD:END -->
