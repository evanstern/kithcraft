---
id: TASK-0021
title: I1 - Demo configuration and the two run targets
status: To Do
assignee: []
created_date: '2026-08-21 23:40'
labels:
  - integration
  - m-0-build
milestone: m-0
dependencies:
  - TASK-0013
  - TASK-0018
  - TASK-0014
  - TASK-0019
documentation:
  - docs/design/demo-build-plan.md
  - docs/design/llm-routing-and-budget.md
  - docs/design/death-mechanics.md
priority: high
ordinal: 21000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
As an operator, I want one documented sequence that brings up a demo-ready world with the cast seeded, so that running the demo is a decision about when, not a research project.

**Scope boundary.** Two artifacts, started independently: the daemon binary and the mod jar. Documented startup ordering and the reconnect behaviour that makes **mind-restart independent of vendor-restart** (T-4). World and server config: the vanilla daylight cycle per ruling **R-1** (nine day/night cycles across the ~3-hour evening, stated as a ruling rather than left as a default — lengthening the in-game day would make the demo unrepresentative of the game anyone will actually play, and nine cycles exercise consolidation 27 times), the **grief-period** knob (R-3), and the **danger-tuning** knob (R-6) present but **off by default** — the knob exists only for a recorded take that cannot afford to lose a cast member, and using it is a per-run choice, never the default. Cast seeding: run M3's genesis for three villagers and bind them to bodies. Surface the per-class call/token counters and the E6-input-tokens instrument at session end.

**Done proves.** One documented command sequence brings up a server with three personas seeded and bound. **Restarting the daemon mid-session and reconnecting leaves the villagers with their memories** — the demo acceptance check decision-0003 promotes from aspiration to test. Counters report at session end. Every knob above is config, not a constant in code.

**Depends on.** M3, M7, V3, V5.

**References.** docs/design/demo-build-plan.md section 3.4 (I1) and its rulings R-1, R-3, R-6 are the plan of record. Ratified surfaces consumed: decision-0003 + docs/design/llm-routing-and-budget.md (T-4 mind-restart independence, section 7.1 the nine-dusk question, the per-class instrumentation), docs/design/death-mechanics.md (section 6.2's grief-period and danger-tuning open items), decision-0001 (Fabric server-side mod, two artifacts).

**Suggested tier: `sonnet` (next sweep's runbook decides).**
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 One documented command sequence brings up the daemon and the mod jar as independently started artifacts and yields a server with three personas seeded and bound
- [ ] #2 Restarting the daemon mid-session and reconnecting leaves the villagers with their memories (T-4 mind-restart independence)
- [ ] #3 The vanilla daylight cycle is kept per ruling R-1: nine day/night cycles across the evening, recorded as a ruling rather than an unexamined default
- [ ] #4 The grief-period knob (R-3) and the danger-tuning knob (R-6) exist as config, with danger tuning off by default
- [ ] #5 Per-class call/token counters and the E6-input-tokens instrument report at session end
- [ ] #6 Every knob above is config, not a constant in code
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Tests pass
- [ ] #2 Docs and wiki are updated and pass freshness tests
- [ ] #3 Spec and Backlog are in sync
<!-- DOD:END -->
