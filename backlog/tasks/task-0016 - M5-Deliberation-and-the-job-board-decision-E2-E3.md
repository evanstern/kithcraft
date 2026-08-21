---
id: TASK-0016
title: 'M5 - Deliberation and the job-board decision (E2, E3)'
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
  - docs/design/body-protocol-v0.md
  - docs/design/kithcraft-brief.md
priority: high
ordinal: 16000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
As a villager, I want work orders to arrive on top of a life I am already living, so that taking one — or not — is a choice I made rather than a command I executed.

**Scope boundary.** Port `toolloop`'s bounded-loop shape — *a tool call is a REQUEST; an event is the FACT; the gate decides* — onto `intent`/`intent_ack`/`act_result`, the one-to-one map decision-0003 identified. Verb vocabulary from the runtime manifest, **not** from a compiled-in list. E2 routine deliberation at schedule transitions and open choices (~8/villager/cycle); E3 job-board deliberation on a `text` percept with `origin: read`, carrying its own context shape (board contents, other villagers' claims, standing relationship to the player, current commitments). The **urgency interrupt** exactly as routing section 5.5 states it: an `urgent` percept **cancels the in-flight deliberation**, **does not itself trigger a model call**, and **enqueues one deliberation** whose context includes it — because the body's reflex has already run. Memory window K=10 situated: top K-2 by recency-decayed weight plus **2 seeded serendipity picks from the older half** (the thing that stops a villager's context collapsing onto its five loudest days). Structured output, so an intent is a value rather than a text to interpret.

**Done proves.** Against the fake vendor: a scripted board posting yields a claim-or-decline intent carrying an **authored `reason`** (section 5.2 requires the mind to have a why). A decline is reachable and reads as this persona's decline, not a generic refusal. An `urgent` percept mid-deliberation cancels the call and produces exactly one follow-up deliberation — not three, and not zero. No intent names a target by description ("the nearest bed"); every target is a token the mind was given.

**Depends on.** M2, M4.

**Design check — micromanagement.** *Reluctance is the product* (brief #6, routing E3), but reluctance is not non-compliance forever: a villager who never takes a posted job turns the board into a chore the player must keep re-issuing, which is the failure mode inverted rather than avoided. The check: across a scripted evening's postings, work gets done without the player re-posting, and the refusals that do occur are legible as *this* villager's.

**Design check — politeness-policing.** A refusal must be grounded in the villager's own wants, commitments or relationship — never in the player's conduct. There is no compliance gate, no cooldown, and no "you were rude to me so I won't work" mechanic anywhere in this task. The player can be a jerk; that costs them a relationship, not an API.

**References.** docs/design/demo-build-plan.md section 3.2 (M5) is the plan of record. Ratified surfaces consumed: decision-0003 + docs/design/llm-routing-and-budget.md (E2/E3 classes, the urgency interrupt in section 5.5, the K=10 situated window with serendipity picks, the toolloop-to-intent map), docs/design/body-protocol-v0.md (intent/intent_ack/act_result, verbs from the runtime manifest, tokens-not-descriptions, Q-6's read channel), docs/design/kithcraft-brief.md (#6 reluctance; the micromanagement and politeness-policing spell-breakers).

**Suggested tier: `sonnet` (next sweep's runbook decides).**
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 The bounded deliberation loop maps onto intent/intent_ack/act_result: a tool call is a request, an act_result is the fact, and the gate decides
- [ ] #2 The verb vocabulary is read from the runtime manifest, never a compiled-in list
- [ ] #3 Against the fake vendor a scripted board posting yields a claim-or-decline intent carrying an authored reason
- [ ] #4 A decline is reachable and reads as this persona's decline, not a generic refusal
- [ ] #5 An urgent percept mid-deliberation cancels the in-flight call, triggers no model call of its own, and enqueues exactly one follow-up deliberation whose context includes it
- [ ] #6 The K=10 situated memory window is top K-2 by recency-decayed weight plus 2 seeded serendipity picks from the older half
- [ ] #7 No intent names a target by description: every target is a token the mind was given
- [ ] #8 Design check (micromanagement): across a scripted evening's postings work gets done without the player re-posting, and refusals are legible as this villager's
- [ ] #9 Design check (politeness-policing): refusals are grounded in the villager's wants, commitments or relationship, never the player's conduct; no compliance gate, cooldown or lockout exists
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Tests pass
- [ ] #2 Docs and wiki are updated and pass freshness tests
- [ ] #3 Spec and Backlog are in sync
<!-- DOD:END -->
