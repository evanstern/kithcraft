---
id: TASK-0004
title: Decide mind daemon language and LLM routing/budget for a small cast at 1x
status: In Progress
assignee: []
created_date: '2026-08-19 18:37'
updated_date: '2026-08-21 22:12'
labels:
  - design-decision
  - architecture
milestone: m-0
dependencies:
  - TASK-0002
documentation:
  - docs/design/kithcraft-brief.md
priority: medium
ordinal: 4000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
As the operator, I want the mind daemon's language settled and an LLM routing/budget sketch for 3–6 villagers at 1x real time, so that an evening with my villagers has a known cost envelope and the mind layer can start being built.

Context: two coupled open questions from the brief. (a) Mind daemon language: whether promptworld I's Go daemon survives behind the body-protocol seam or the mind layer is rebuilt fresh — an implementation detail behind the seam, but one that must be decided to start building. (b) LLM routing and budget: which cognition events call which model tier, expected call cadence (the reflex/planner split means scripted competence for doing, LLM only for choosing and relating), cost envelope for an evening of play, and latency posture (a villager taking 20s to decide is a person mulling — real time deletes the governor/speed-ladder machinery). Doctrine reference: promptworld I docs/wiki (event-sourced memory, reflex/planner split, salience+consolidation, persona firewall).

Spec: specs/004-mind-daemon-routing
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A decision on mind daemon language / reuse-vs-rebuild is recorded as a Backlog decision record with rationale
- [x] #2 An LLM routing sketch exists: which villager cognition events call an LLM, at what tier, at what expected cadence for a 3-villager evening
- [x] #3 A cost envelope estimate for the one-real-evening demo is written down with its assumptions
- [ ] #4 Operator has ratified the decisions
- [x] #5 Spec phase: Phase 1 — Daemon assessment
- [x] #6 Spec phase: Phase 2 — Routing sketch & cost envelope
- [x] #7 Spec phase: Phase 3 — Recommendation & decision record
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Tier: opus (operator-escalated at runbook sign-off 2026-08-21 — coupled architecture decisions the spec cannot pre-settle). Model ID: cc/claude-opus-5[1m], fallback cc/claude-opus-4-8[1m]. Served model to be verified from first dispatch transcript.

Phase 1 dispatch served model VERIFIED: cc/claude-opus-5[1m] (from implementer report, 2026-08-21).

Phase 2 done (46cb88d, opus verified): docs/design/llm-routing-and-budget.md — demo evening ≈ $5.17 (≈ $3.99 cached), 435 calls / ~1.63M tokens; 6 LLM event classes across 3 tiers (Opus: persona genesis + consolidation; Sonnet: deliberation/job-board/conversation; Haiku: ambient pool); reflexes and deterministic mind-side machinery explicitly non-LLM; 9 day/night cycles per 3h evening → consolidation ×27; sensitivity: 6 villagers ≈ $10.1, chattier ≈ $8.6.

Phase 3 done (7acc232 + re-pin bf947cf, opus verified, ~146k tokens): recommendation rebuild-Go (daemon rebuilt behind the seam, four promptworld assets reused as source material); decision-0003 created (proposed). Q-1 transport narrowed one-way (in-process foreclosed) — flagged, not decided. CAPSULES regenerated. Card ACs 1-3 satisfied; AC 4 (operator ratification) pending at PR review.
<!-- SECTION:NOTES:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Tests pass
- [ ] #2 Docs and wiki are updated and pass freshness tests
- [ ] #3 Spec and Backlog are in sync
<!-- DOD:END -->
