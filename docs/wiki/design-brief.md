---
name: design-brief
description: The ratified design brief — loneliness-cure thesis, the ten ratified decisions (embodied player, villager-shaped NPCs, small cast, generated minds, no direct control, diegetic orders, real-time only), the two load-bearing tie-breaker constraints, and the spell-breakers to design against. Load before any design-adjacent work.
kind: concept
sources:
  - docs/design/kithcraft-brief.md
verified_against: 50c3def435dd9326d38e51118f08944815cbe80c
---

# Design brief (ratified)

`docs/design/kithcraft-brief.md` is the project's sole ratified design authority
(operator-ratified 2026-08-19, distilled from a live design session in the promptworld I
repo). It is written as a self-contained handoff: an agent starting fresh needs only this
file. Its decisions are settled — work does not relitigate them.

## How it works

The thesis: **the LLM villagers exist to cure the loneliness** of survival-crafting play.
The AI serves the feeling of company; the mine/build/craft/explore loop carries the game
on its own.

The ten ratified decisions, compressed:

1. **Embodied player** — physical presence, no god view.
2. **Vehicle: Minecraft server mod** — world/physics/multiplayer come free; V2 owned
   engine possible later via the body-protocol seam.
3. **Villager-shaped, smarter** — real villagers riding the village fiction (beds,
   workstations, schedules), never Mineflayer-style bot clients.
4. **Small cast, deep bonds** — ~3–6 named villagers; permadeath is real and should sting.
5. **Minds are generated, not edited** — personas and endogenous desires generated at
   birth; player-editable minds would make villagers puppets and the player alone again.
6. **No direct control** — orders land on lives in motion; reluctance and grumbling are
   relationship, not bugs.
7. **Diegetic order interface** — the v1 soul is the job-board book.
8. **Real time only (1x)** — deletes promptworld I's most expensive subsystem; a villager
   taking 20 seconds to decide is a person mulling.
9. **Multiplayer v1** = my server, me and my villagers, a friend can drop in.
10. **Replenishment punted** — no spawning mechanism in v1.

Tie-breakers for forks the brief doesn't cover: (a) the loneliness-cure thesis, (b)
minds-are-others. Spell-breakers to design against: tedious interactions,
micromanagement, villagers policing player politeness.

## Connections

The architecture posture it mandates is [[body-protocol-seam]]; its doctrine inheritance
is [[promptworld-lineage]]; its demo definition is [[v1-demo]]; its "open questions" list
seeded the board's tasks, the first resolved by [[mod-stack-decision]] on evidence in
[[prior-art]].

## Operational notes

The brief's own prior-art citations are dated 2026-08-19 and superseded for currency by
`specs/001-mod-stack-decision/research/prior-art.md` (2026-08-20); no contradictions,
only version drift.
