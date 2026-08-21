---
name: promptworld-lineage
description: What Kithcraft inherits from promptworld I and what it deliberately discards — doctrine transfers (event-sourced memory, reflex/planner split, salience/consolidation, epistemic hygiene, persona firewall), code does not (sim engine, executor, governor, guardian all die). Load when a design question touches mind architecture or when tempted to import promptworld I code.
kind: concept
sources:
  - docs/design/kithcraft-brief.md
verified_against: 50c3def435dd9326d38e51118f08944815cbe80c
---

# Promptworld lineage

Kithcraft is "promptworld II": promptworld I is retroactively the R&D project where the
theory of agent minds got worked out. II throws away the world engine and keeps the
minds — as **doctrine, not code**. Nothing in II imports I's code; I's repo (its
`docs/wiki/` corpus) is grounded reference only.

## How it works

**Doctrine that transfers** (brief, "Architecture posture"):

- **Event-sourced memory.**
- **The reflex/planner split** — scripted competence for *doing* (pathfind, chop, place
  blocks), LLM for *choosing and relating*. Cheap reflexes for competence, expensive
  thoughts for meaning; the brief calls this the single most transferable lesson of I.
- **Salience + consolidation + situated memory** — the nightly-digestion heritage.
- **Epistemic hygiene** — an agent knows only what it saw or was told, with provenance.
- **The persona firewall.**

**What dies with promptworld I:** the sim engine, executor, reducer, terrain generation,
TUI, determinism-for-replay machinery, the governor/speed ladder, and the guardian (the
player-as-intermediary concept as a whole). The real-time-only decision ([[design-brief]]
#8) is what deletes the governor family: at 1x there is no cognition horizon, no speed
ladders, no staleness budgets.

The inversion driving the split: I built a fascinating simulation with nothing for the
player to do; II puts the player in a game that's already fun and gives them neighbors.

## Connections

Doctrine list lives in [[design-brief]]; whether I's Go daemon survives is an open detail
behind [[body-protocol-seam]]; the reflex side of the split maps onto the engine's own
brain machinery ([[villager-brain-api]]) under [[mod-stack-decision]].

## Operational notes

When mind-daemon work starts (TASK-0004 decides its language), the transferred doctrine
items become checkable requirements against I's wiki corpus — consult I's `docs/wiki/`
INDEX just-in-time rather than porting files.
