# Spec 006 — Build plan for the "one real evening" demo

**Board task:** TASK-0006 · **Milestone:** One real evening (v1 demo)
**Direction source:** docs/design/kithcraft-brief.md (ratified 2026-08-19 — do not
relitigate) and the four ratified decisions it now has: decision-0001 (Fabric
server-side mod), the body protocol v0 (contract, PR #6), decision-0002 (augmented
vanilla villager), decision-0003 (mind daemon in Go, rebuilt behind the seam) with its
routing/budget doc (docs/design/llm-routing-and-budget.md), plus the ratified death
design (docs/design/death-mechanics.md, PR #8). This task PLANS the demo build; it
implements nothing.

## Problem

Every architectural input the demo needs is now decided, but the demo is one large
undifferentiated ambition ("an evening with three villagers"). Building it requires a
decomposition into board tasks that (a) are each one-PR deliverables, (b) order by real
dependency so a sweep can lane them, and (c) cover every demo beat without gold-plating
beyond the demo. A wrong decomposition taxes every task of the next sweep.

## Requirements (mapped to the card's acceptance criteria)

### R1 — Deliverable task decomposition (card AC #1)

A build plan document that decomposes the demo into deliverable tasks created on the
board, each: one-PR-shaped (a coherent, reviewable, mergeable deliverable), carrying a
user story per project convention, with dependencies stated so the next sweep can
derive lanes. The plan states for each task what "done" proves (testable against the
fake vendor where mind-side, against a dev server where mod-side).

### R2 — Full demo-beat coverage (card AC #2)

The plan covers all demo beats: personas/desires generation (persona genesis per the
routing doc), schedules (wake/work/socialize/sleep on the vanilla Brain substrate),
persistent memory (event-sourced, mind-side), the diegetic job-board book, the
blueprint build alongside the player, dusk conversation, and night danger (with the
death-design's admitted/suppressed causes and its Mixin obligations). Plus the
infrastructure those beats imply: the Go daemon skeleton against the seam, the Fabric
mod as first body vendor, the fake-vendor test harness, and the transport decision
(seam Q-1 — now narrowed to real wires by decision-0003) which must be scheduled,
not left floating.

### R3 — Constraints honored, spell-breakers named (card AC #3)

The plan honors the two load-bearing constraints (loneliness-cure thesis,
minds-are-others) in its scoping choices, and names the spell-breakers (tedium,
micromanagement, politeness-policing) as design checks on the tasks where each risk
lives.

### R4 — Operator sign-off (card AC #4)

The plan and its created board tasks land in the PR; the operator signs off at PR
review before the card moves Done. The next sweep does not start under this task.

## Non-goals

- No implementation of any demo beat.
- No re-litigation of ratified decisions (stack, entity, language, routing tiers,
  death rulings).
- No replenishment, curing, or post-demo features; demo scope only.
- Transport (Q-1) gets a decision *task* in the plan, not a decision here.
