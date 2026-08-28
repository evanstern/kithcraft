---
id: TASK-0017
title: 'M6 - Dusk conversation and the ambient pool (E4, E5)'
status: In Progress
assignee: []
created_date: '2026-08-21 23:39'
updated_date: '2026-08-28 19:47'
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

Spec: specs/017-dusk-conversation
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 E4 conversation turns run on Sonnet 5 with streaming on, effort low, thinking off, cached prefix and max_tokens ~300
- [x] #2 A scripted dusk exchange between two minds against the fake vendor produces a multi-turn conversation about the day, the work and the player
- [x] #3 First-token latency is measured and is under the 3 s ceiling
- [x] #4 V3's pair-formation signal is consumed to pre-generate the opening turn during the walk
- [x] #5 E5 ambient runs on Haiku 4.5 as one batched call per villager per in-game day producing ~8 lines, served from the pool in under 200 ms and refreshed daily
- [x] #6 A remark about something specific escalates to a live call rather than drawing from the pool
- [x] #7 Design check (tedium): lines do not repeat within a cycle, conversations reach a natural end rather than looping or hitting a turn cap mid-sentence, and the stall-line is used sparingly
- [x] #8 Design check (politeness-policing): a villager may resent, grumble and say so, but never lectures, moralizes, or gates anything on the player's conduct
- [x] #9 Spec phase: Phase 1 — The dusk exchange (US1)
- [x] #10 Spec phase: Phase 2 — Pre-generation (US2)
- [x] #11 Spec phase: Phase 3 — The ambient pool (US3)
- [x] #12 Spec phase: Phase 4 — Spell-breaker checks, gates, and closure
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
claimed by sweep-0007-0022 orchestrator 2026-08-28 (lane 4); spec 017 stub + link ride this claim commit

tier: sonnet (default) · model cc/claude-sonnet-5[1m] · rubric: E4/E5 params fully specified (Sonnet 5 <3s thinking off; Haiku pool); the latency posture is design already done (runbook lane 4)

Phase 4 (T009-T013) closes the card. AC citations, all against the fake vendor / scripted or mocked model, mind/vet+test green:
AC#1 mind/converse/converse_test.go TestExchangeUsesE4Config (Sonnet 5, streaming, effort low, thinking off, cached, max_tokens<=300).
AC#2 mind/converse/fakevendor_test.go TestDuskExchange_AgainstFakeVendor (multi-turn transcript asserted to mention day/work/player).
AC#3 same test's measured first-token latency assertion, plus TestSpeaker latency plumbing in converse.go's stream(); mocked delay, ceiling instrumented, not assumed. HONEST LIMIT: the real <3s ceiling against the live API is I2's evening run, not proven here.
AC#4 mind/converse/pregen_test.go + converse_test.go TestExchange_OpeningSlotServesWithoutNewCall / TestExchange_OpeningSlotFallsBackWhenUnfilled / TestSlot_* (Pool.Begin models the call the daemon's V3-signal ingest will make). HONEST LIMIT: the slot/serve/discard/fallback mechanism is proven against a modeled signal; wiring V3's actual pair-formation percept into Pool.Begin is the daemon integration's job, not this package's.
AC#5 mind/converse/pool_test.go TestE5UsesHaiku45, TestAmbientPool_RefillProducesLines, TestAmbientPool_ServeUnderBudgetNoRepeat, TestAmbientPool_DayRolloverClearsYesterday.
AC#6 mind/converse/pool_test.go TestIsTargeted, TestEscalate_LiveCallBypassesPool.
AC#7 mind/converse/pool_test.go TestAmbientPool_ServeUnderBudgetNoRepeat (no repeats) + converse_test.go TestExchange_SafetyBoundNeverFires (natural end, not a cap) + tedium_test.go TestAmbientPool_StallOnlyAfterExhaustion (stall used only for the overflow past the 8-line batch, not routinely — the missing structural check Phase 3 flagged).
AC#8 politeness_test.go TestPromptAssembly_NoModeratingLexicon (structural: converse.go/pool.go carry none of persona.Moralizing's lexicon) + TestExchange_ResentfulContentPassesThroughUnfiltered + TestAmbientPool_ResentfulLinePassesThroughUnfiltered (scripted grumbling content passes verbatim; no conduct-gating mechanism exists).
AC#9-12 spec phases 1-4 all complete: specs/017-dusk-conversation/tasks.md T001-T013 all checked.
Gates (T011): go vet ./... and go test -race -count=1 ./... green in mind/ (commit e86913c). Wiki (T012): docs/wiki/overview.md amended and re-pinned to e86913c (commit 72ffa64) — the only note whose prose tracks per-milestone build status; all other notes' declared sources are untouched by this branch.
<!-- SECTION:NOTES:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Tests pass
- [ ] #2 Docs and wiki are updated and pass freshness tests
- [ ] #3 Spec and Backlog are in sync
<!-- DOD:END -->
