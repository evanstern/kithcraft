# Spec 005 — Death mechanics: what kills, what remains, how the living remember

**Board task:** TASK-0005 · **Milestone:** One real evening (v1 demo)
**Direction source:** docs/design/kithcraft-brief.md (ratified 2026-08-19 — do not
relitigate). Ratified posture: permadeath is real and should sting ([[design-brief]]
#4); replenishment is punted in v1 (#10), so death permanently shrinks the cast.
decision-0002 (augmented vanilla villager) fixes the entity substrate death happens to;
the body protocol v0 fixes how a death reaches other minds (percepts with provenance).

## Problem

The brief ratifies that villagers can die for real, but defers the mechanics. Without
them, night danger is stakes-free theater and TASK-0006 cannot plan the demo's danger
beats. Three questions need design answers: **what can kill** a villager (night danger
certainly; hunger? falls?), **what remains** in the world (grave, belongings), and
**how survivors carry the dead** (memory entries, dusk-conversation material, stories).
The design must thread two ratified constraints: death must sting (loneliness-cure
thesis — these are friends, not props) without making survival babysitting (the
micromanagement spell-breaker).

## Requirements (mapped to the card's acceptance criteria)

### R1 — Death-mechanics design doc (card AC #1)

A design document at `docs/design/death-mechanics.md` covering:

- **Causes of death:** which vanilla damage sources apply to villagers (hostile mobs,
  environment) and which the design admits or suppresses for the cast (hunger, falls,
  drowning, fire, friendly fire from the player), with the vanilla mechanics verified
  (dated, cited) for what decision-0002's augmented villager inherits.
- **What remains:** the grave (form, placement, who digs it), belongings (inventory
  drops vs preserved bundle), and any marker the world keeps (bed left empty,
  workstation unclaimed) — each chosen for emotional weight per the thesis.
- **How the living remember:** what memory entries a death writes for witnesses vs
  told-about survivors (provenance per the seam: saw vs told), how the dead surface in
  dusk conversation, and how long the dead stay conversationally alive (decay/
  consolidation posture, from the salience doctrine).

### R2 — The micromanagement spell-breaker (card AC #2)

The design explicitly addresses fragility: villagers competent enough at self-
preservation (fleeing, sheltering at night, eating) that keeping them alive is
base-building and walls — the player's normal play — not feeding schedules and escort
missions. The doc names the failure mode and states the design's specific answers.

### R3 — Shrinking cast consequence (card AC #3)

The design states what a permanently shrinking cast means for the v1 demo (a 3-villager
demo where one dies is a 2-villager evening) and either accepts that with rationale or
scopes a mitigation (e.g. demo-config cast size) — without reopening ratified
replenishment punting.

### R4 — Operator ratification (card AC #4)

The design lands as a doc in the PR; the operator ratifies at PR review before the
card moves Done.

## Non-goals

- No implementation, no Mixin inventory beyond what causes-of-death verification
  requires (that's demo build-plan territory, TASK-0006).
- No replenishment design (ratified punt).
- No player-death design — vanilla handles the player.
