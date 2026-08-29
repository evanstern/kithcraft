---
id: TASK-0022
title: 'I2 - The evening: run it, measure it, check it against the brief'
status: In Progress
assignee: []
created_date: '2026-08-21 23:40'
updated_date: '2026-08-29 04:18'
labels:
  - integration
  - m-0-build
milestone: m-0
dependencies:
  - TASK-0007
  - TASK-0008
  - TASK-0009
  - TASK-0010
  - TASK-0011
  - TASK-0012
  - TASK-0013
  - TASK-0014
  - TASK-0015
  - TASK-0016
  - TASK-0017
  - TASK-0018
  - TASK-0019
  - TASK-0020
  - TASK-0021
documentation:
  - docs/design/demo-build-plan.md
  - docs/design/kithcraft-brief.md
  - docs/design/llm-routing-and-budget.md
priority: high
ordinal: 22000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
As an operator, I want a full evening run judged against every beat and every spell-breaker with numbers attached, so that "the demo works" is a finding rather than an impression.

**Scope boundary.** A ~3-hour real-time evening at 1x with the player present throughout and three villagers alive. Walk the **beat checklist** (demo-build-plan section 4's coverage map: personas and desires; schedules; persistent memory; the job-board book; the blueprint build alongside the player; dusk conversation; night danger) and the **spell-breaker checklist** (section 5.2: tedium, micromanagement, politeness-policing). Measure and **replace the A-n assumptions by number**: calls per class, tokens per call, cost against the **~$20 ceiling** (baseline ~$5.17, ~$4.00 cached — cost is not the binding constraint and this run is not an optimization exercise), E4 first-token latency against its < 3 s ceiling, and E6 input tokens per villager per cycle against the ~80/day assumption and the ~150 upgrade trigger. Record which dusk landed best (per ruling R-1, "the evening" is a *recording* framing: the highlight is cut from the dusk that landed best, not engineered by lengthening the day).

**Done proves.** A recorded evening in which every beat in the coverage map is observed, no spell-breaker check fails, and each A-n assumption is annotated with its measured value. Where a check fails, the finding is written down with the task that owns the fix — this run is allowed to *find* problems; it is not allowed to fail silently.

**Depends on.** All fifteen preceding tasks.

**References.** docs/design/demo-build-plan.md sections 3.4 (I2), 4 (coverage map) and 5 (constraints and spell-breakers) are the plan of record. Ratified surfaces consumed: docs/design/kithcraft-brief.md (the v1-demo beats and the three spell-breakers), decision-0003 + docs/design/llm-routing-and-budget.md (the A-n assumptions, the cost baseline and ~$20 ceiling, the E4 latency ceiling, the E6 upgrade trigger).

**Suggested tier: `sonnet` (next sweep's runbook decides).**

Spec: specs/022-the-evening
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A ~3-hour evening at 1x is run with the player present throughout and three villagers alive
- [ ] #2 Every beat in the plan's coverage map is observed during the run: personas and desires, schedules, persistent memory, the job-board book, the blueprint build alongside the player, dusk conversation, night danger
- [ ] #3 Every spell-breaker check (tedium, micromanagement, politeness-policing) is walked and none fails
- [ ] #4 Each A-n assumption is annotated with its measured value: calls per class, tokens per call, total cost against the ~ ceiling, E4 first-token latency against the 3 s ceiling, E6 input tokens per villager-cycle against the ~80/day assumption and ~150 upgrade trigger
- [ ] #5 Where a check fails, the finding is written down with the task that owns the fix - the run is allowed to find problems but not to fail silently
- [ ] #6 The dusk that landed best is recorded for the highlight cut
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
claimed by sweep-0007-0022 orchestrator 2026-08-29 (lane 5 final; all fifteen deps merged); spec 022 stub + link ride this claim commit. OPERATOR CHECKPOINT 5: the evening itself is operator-run (~3h, player present) — the sweep prepares and stops
<!-- SECTION:NOTES:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Tests pass
- [ ] #2 Docs and wiki are updated and pass freshness tests
- [ ] #3 Spec and Backlog are in sync
<!-- DOD:END -->
