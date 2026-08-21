---
name: v1-demo
description: Milestone m-0 "One real evening" — the v1 demo definition: three villagers with generated personas, schedules and persistent memory; a job-board blueprint built alongside the player; dusk conversation; night danger making the player's walls protect friends. Load when scoping or building toward the demo (TASK-0006).
kind: concept
sources:
  - docs/design/kithcraft-brief.md
  - backlog/milestones/m-0 - one-real-evening-(v1-demo).md
verified_against: 50c3def435dd9326d38e51118f08944815cbe80c
---

# v1 demo — "one real evening"

Milestone m-0, the project's first deliverable experience: three villagers on a survival
server with the player, one real evening of cohabitation. It is deliberately small — a
household, not a city — and every board task (TASK-0001..0006) exists to reach it.

## How it works

The demo's required texture (brief, "v1 demo" section; mirrored in the m-0 milestone
description):

- Three villagers with **names, generated personas and desires**, schedules
  (wake/work/socialize/sleep), and persistent memory.
- The player posts a simple blueprint on the **job-board book** (the diegetic order
  interface, [[design-brief]] #7); one villager builds it **while the player builds
  alongside** — company in the work itself.
- At dusk the villagers **talk to each other** — about the day, the work, the player.
- Vanilla night danger means the player's walls and torches protect their *friends* —
  base-building becomes emotionally load-bearing.

Spell-breakers the demo must not exhibit: tedious player interactions, micromanagement
required to keep villagers alive or productive, villagers taking offense at a player
being a jerk.

The demo is the sizing anchor for engineering decisions: a 3–6 NPC cast at 1x real time
is trivial load for either the engine or an external LLM daemon (per the mod-stack
comparison), which is why the hybrid stack's added complexity bought nothing
([[mod-stack-decision]]).

## Connections

Defined by [[design-brief]]; board milestone m-0 groups the path to it; TASK-0006 (build
plan) will decompose it; runs on the stack chosen in [[mod-stack-decision]] over the
substrate in [[villager-brain-api]].

## Operational notes

The board's remaining open questions feeding the demo: body protocol (TASK-0002), entity
implementation (TASK-0003), mind daemon language + LLM budget (TASK-0004), death
mechanics (TASK-0005), build plan (TASK-0006).
