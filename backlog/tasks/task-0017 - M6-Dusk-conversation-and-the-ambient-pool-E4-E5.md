---
id: TASK-0017
title: 'M6 - Dusk conversation and the ambient pool (E4, E5)'
status: To Do
assignee: []
created_date: '2026-08-21 23:39'
labels:
  - mind
  - m-0-build
milestone: m-0
dependencies:
  - TASK-0010
  - TASK-0011
documentation:
  - docs/design/demo-build-plan.md
  - docs/design/llm-routing-and-budget.md
  - docs/design/kithcraft-brief.md
priority: high
ordinal: 17000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
As a player, I want to overhear my neighbours talking about the day, the work, and me, so that the base I built stops being a place and starts being a household.

**Scope boundary.** E4 conversation turns on **Sonnet 5**, under the design's only hard latency ceiling: **< 3 s to first token** — streaming on, `effort: low`, **thinking off**, cached prefix, `max_tokens` ~300. Extended thinking, which every other class benefits from, is a direct tax on the one class that cannot pay it, and Opus is *too slow* here: this tier is constrained from above. Consume V3's pair-formation signal (ruling R-7) to **pre-generate the opening turn during the walk**, so the scene opens instantly. The interlocutor model — who this is, what I think of them, shared history — as its own context slice. E5 ambient on **Haiku 4.5**: one batched call per villager per in-game day producing ~8 persona-flavoured lines, served from the pool in **< 200 ms** and **refreshed daily**; a remark about something specific escalates to a live call instead of drawing from the pool.

**Done proves.** Against the fake vendor: a scripted dusk exchange between two minds produces a multi-turn conversation about the day, the work and the player, with **measured** first-token latency < 3 s. The pool serves in < 200 ms and does not repeat a line within a cycle. A conversation **ends** — it has a termination condition, not a turn cap that leaves two villagers mid-sentence.

**Depends on.** M2, M4.

**Design check — tedium.** This is the task where tedium lives. Three concrete checks: lines do not repeat within a cycle (the pool refreshes daily for exactly this reason); a conversation reaches a natural end rather than looping; and the ambient stall-line ("Hm.") is used sparingly, because a villager who says "Hm." before every sentence is worse than one who pauses. A villager the player learns to walk past is the spell broken.

**Design check — politeness-policing.** A villager may resent the player, grumble about them at the fire, and say so to their face. What it may not do is lecture, moralize, or gate anything on the player's conduct. Grumbling at the fire is relationship; a reprimand is a politeness simulator.

**References.** docs/design/demo-build-plan.md section 3.2 (M6) is the plan of record. Ratified surfaces consumed: decision-0003 + docs/design/llm-routing-and-budget.md (E4 on Sonnet 5 with thinking off under the < 3 s ceiling, E5 ambient pool on Haiku 4.5, section 5.2 lever 2 pre-generation), docs/design/kithcraft-brief.md (the dusk-conversation beat; the tedium and politeness-policing spell-breakers), docs/design/body-protocol-v0.md (speak -> speech in earshot).

**Suggested tier: `sonnet` (next sweep's runbook decides).**
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 E4 conversation turns run on Sonnet 5 with streaming on, effort low, thinking off, cached prefix and max_tokens ~300
- [ ] #2 A scripted dusk exchange between two minds against the fake vendor produces a multi-turn conversation about the day, the work and the player
- [ ] #3 First-token latency is measured and is under the 3 s ceiling
- [ ] #4 V3's pair-formation signal is consumed to pre-generate the opening turn during the walk
- [ ] #5 E5 ambient runs on Haiku 4.5 as one batched call per villager per in-game day producing ~8 lines, served from the pool in under 200 ms and refreshed daily
- [ ] #6 A remark about something specific escalates to a live call rather than drawing from the pool
- [ ] #7 Design check (tedium): lines do not repeat within a cycle, conversations reach a natural end rather than looping or hitting a turn cap mid-sentence, and the stall-line is used sparingly
- [ ] #8 Design check (politeness-policing): a villager may resent, grumble and say so, but never lectures, moralizes, or gates anything on the player's conduct
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Tests pass
- [ ] #2 Docs and wiki are updated and pass freshness tests
- [ ] #3 Spec and Backlog are in sync
<!-- DOD:END -->
