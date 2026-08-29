---
id: TASK-0021
title: I1 - Demo configuration and the two run targets
status: Done
assignee: []
created_date: '2026-08-21 23:40'
updated_date: '2026-08-29 04:18'
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

Spec: specs/021-demo-config
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 One documented command sequence brings up the daemon and the mod jar as independently started artifacts and yields a server with three personas seeded and bound
- [x] #2 Restarting the daemon mid-session and reconnecting leaves the villagers with their memories (T-4 mind-restart independence)
- [x] #3 The vanilla daylight cycle is kept per ruling R-1: nine day/night cycles across the evening, recorded as a ruling rather than an unexamined default
- [x] #4 The grief-period knob (R-3) and the danger-tuning knob (R-6) exist as config, with danger tuning off by default
- [x] #5 Per-class call/token counters and the E6-input-tokens instrument report at session end
- [x] #6 Every knob above is config, not a constant in code
- [x] #7 Spec phase: Phase 1 — The daemon runtime loop (US1 machinery + US2 machinery)
- [x] #8 Spec phase: Phase 2 — Knobs and the report (US3 + US4)
- [x] #9 Spec phase: Phase 3 — The documented sequence and the live proof (US1 + US2 live)
- [x] #10 Spec phase: Phase 4 — Gates and closure
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
claimed by sweep-0007-0022 orchestrator 2026-08-29 (lane 5; deps M3/M7/V3/V5 all merged); spec 021 stub + link ride this claim commit

tier: sonnet (default) · model cc/claude-sonnet-5[1m] · rubric: config plumbing and documented startup; every knob is ruled (R-1, R-3, R-6) — assembly of tested parts, judgment already settled (runbook lane 5)

Phase 4 (T010) AC proofs — AC #1: docs/design/demo-runbook.md (one file, §1-3 build+both start orders) followed verbatim in specs/021-demo-config/research/bringup-observation.md §1-3 — daemon and mod jar started independently, server up with a session attached, personas seeded/bound via the documented re-bind path (stub personas matching persona.DemoCast(), zero calls — this dispatch made no live API calls; a real E1 genesis run needs the operator's key per the runbook's §4).

AC #2 (T-4): bringup-observation.md §4, the headline check — one memory admitted pre-kill (world_time=100), daemon SIGTERM'd, restarted against the same rundir, the SAME body token reconnected (HandleConnection's 'matched by body token alone'), a second memory admitted post-restart (world_time=5000); session-report.log + the memory log show both records with the gap visible, nothing synthetic in between. Honest caveat: driven via a scripted mind/seamtest.DialUnix double against the live daemon, not the live mod's own session (bringup-observation.md §5 explains why: the live mod's self_state-only heartbeat produces no admissible memories, so its session alone could not exercise the check). Follow-up flagged for refactor-triage/I2, not fixed here: the live mod's Continuity.java always sends firstSession() and mints a fresh body token per attach — the daemon-side reconnect mechanism this AC proves is real, but the live mod does not yet exercise it end-to-end.

AC #3 (R-1): docs/design/demo-runbook.md §5 records the ruling (vanilla daylight cycle kept, nine cycles across the ~3h evening, 27 consolidations for the 3-villager cast, citing llm-routing-and-budget.md §7.1/A-2); mind/consolidate/cycle.go:60 CycleTicks=24000 is deliberately a constant, not a knob (plan.md design decision 4) — R-1 is recorded as a ruling in the run doc, not left as an unexamined default.

AC #4/#6: R-3 (grief period) — mod/.../death/GriefPeriod.java:23 configuredTicks() reads System property kithcraft.griefPeriodTicks, default 24000, GriefPeriodTest.configOverridableNotAConstant proves override-not-constant. R-6 (danger tuning, new) — mod/.../death/DangerTuning.java:23 enabled() reads kithcraft.dangerTuning, default OFF (Boolean.getBoolean false when unset); DangerTuningTest proves off-by-default + override-not-constant + the shouldSuppress decision table. mind/cmd/minddaemon/config.go's envOr/envOrBool (config_test.go TestEnvOr_ConfigNotConstant/TestEnvOrBool_ConfigNotConstant) cover the daemon-side socket/rundir/genesis knobs the same way. Config-not-constant audit: all four knobs read configuration at call time, none baked into a call site.

AC #5: mind/cmd/minddaemon/report.go Runtime.Report — every llm.E1..E6 row emitted zeroed rather than omitted, plus each body's admitted-count-per-villager-day series, unconditional. Both lifetimes observed: report_test.go's TestReport_ZeroCallPathEmitsZeroed (fresh runtime, no Client, no bodies) and TestReport_IncludesAdmittedInstrumentCounts (one admitted percept lands 'day 0: 1 admitted') cover the unit lifetime; bringup-observation.md's live run actually produced session-report.log with 'obs-body-1 day 0: 1 admitted' after the pre-kill admission, appended for real by the live daemon, not just asserted in a test.

Refactor-triage/I2 flags (Phase 3 findings, not fixed in this phase): (1) the live mod's Continuity.java always sends firstSession() and mints a fresh body token per attach regardless of a prior session, per its own doc — the mind-restart-independence mechanism (AC #2) is proven daemon-side and via a test double, not yet through the live mod's own reconnect path. (2) the live mod's self_state-only heartbeat produced zero body-store opens across ~2m40s of continuous ticking spanning the kill+restart in bringup-observation.md §5 — structurally unadmittable per admission.go's rules (no subject, background urgency), and whether self_state percepts even reach the daemon at all was not distinguished in this pass. Both are docs+observation findings only, no code changed for either.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
I1 delivered via PR #26 (merge 10f5fbb, merge commit, pins preserved). The daemon became a real assembled runtime: per-body M2 log + M7 ledger/archive, persona load-or-genesis with partial resume, Archived hook wired (closing TASK-0018's deferral), sleep from world_time crossings with no-marker retry proven end-to-end; TASK-0020's token-namespace finding fixed as glue (BodyTokenLookups). Config: env-mirrored daemon flags, -genesis=false rehearsal mode, R-6 danger knob OFF by default (no new Mixin, surface stays 4), R-3 verified, config-not-constant audits; unconditional session-end report (counters + E6-input instrument). docs/design/demo-runbook.md is the operator bring-up I2 follows (R-1 recorded as ruling). Live observation: rehearsal zero-call both sub-cases, both start orders, RESTART INDEPENDENCE CONFIRMED (memory pre-kill, SIGTERM, restart, same-token reconnect, gap reported not backfilled). Honest flags for refactor-triage/I2: mod-side reconnect identity gap (Continuity.java firstSession() + fresh token per attach); self_state-only heartbeat admits nothing. Spec-bridge derivation: 4 phases 10/10, Done-eligible. ~826k subagent tokens across 4 sonnet dispatches (cc/claude-sonnet-5[1m], verified per dispatch).
<!-- SECTION:FINAL_SUMMARY:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 Tests pass
- [x] #2 Docs and wiki are updated and pass freshness tests
- [x] #3 Spec and Backlog are in sync
<!-- DOD:END -->
