# Spec 003 — Entity implementation: custom entity vs augmented vanilla villager

**Board task:** TASK-0003 · **Milestone:** One real evening (v1 demo)
**Direction source:** docs/design/kithcraft-brief.md (ratified 2026-08-19 — do not
relitigate); decision-0001 (Fabric, accepted 2026-08-20) fixes the substrate — both
sub-options sit on the vanilla Fabric brain/Mixin surface; a "custom entity" means a
Fabric entity class wired into `Brain<E>`, never a Citizens2-authored NPC.

## Problem

Kithcraft villagers must be real inhabitants of the world — sleeping in beds, working
at stations, threatened by the same night the player is — so protecting them feels
like protecting friends, not managing props. Two implementation shapes fit the ratified
"villager-shaped, smarter, riding the village fiction" decision: **augment vanilla
`Villager` entities** (activity/schedule injection, POI, memory modules on the stock
entity) or **a custom Fabric entity** on the same brain substrate. The choice gates
entity work in the demo build plan (TASK-0006) and shapes the first body vendor's
implementation surface.

## Requirements (mapped to the card's acceptance criteria)

### R1 — Trade-off writeup against the ratified constraints (card AC #1)

A comparison document at `docs/design/entity-implementation-comparison.md` evaluating
both options against each ratified constraint, with evidence:

- **Village fiction reuse:** beds, workstations, schedules, POI — what each option
  inherits free vs must reimplement or Mixin-wire.
- **Real night danger:** hostile-mob targeting must threaten villagers — how each
  option participates in vanilla mob AI targeting, panic, and door/raid mechanics.
- **Permadeath:** death handling, despawn rules, and what each option must suppress
  or own (e.g. vanilla breeding/restocking/curing behaviors that break the fiction).
- **Drop-in multiplayer:** a friend joining vanilla-client sees them as real entities —
  client-side requirements of each option (custom entity ⇒ what renders without a
  client mod; augmented villager ⇒ what vanilla clients already render).
- **Behavior control:** how much of vanilla villager behavior (trading, gossip,
  breeding, restocking) must be disabled/overridden per option, and the Mixin surface
  each requires.
- **Rendering/skin flexibility:** distinguishable villagers (the cast must read as
  individuals) — what each option allows without a client mod.
- Every engine-behavior claim cited (vanilla mechanics docs/wiki or mapped source),
  dated; the evidence rule from TASK-0001 binds (URL + accessed date).

### R2 — Recommendation as a ratified decision record (card AC #2)

A single recommended option with rationale, recorded as a Backlog decision record in
the same PR (proposed status). The recommendation is **proposed** by this task and
**ratified** by the operator at the PR checkpoint — ratification is the operator's
act, not this spec's execution.

## Out of scope

- Implementing either option; any code, scaffolding, or dependency work.
- The body protocol's shapes (TASK-0002, parallel lane) — this comparison must not
  bind protocol decisions; where entity choice touches the seam (e.g. what percepts
  are cheap to emit), state the interaction without deciding for 0002.
- Mind-side concerns (TASK-0004).

## Done means

Comparison doc exists with cited, dated evidence for all six constraint areas; a
decision record exists proposing one option with rationale; operator has ratified;
wiki notes touched by this work ([[villager-brain-api]] at minimum, if its claims are
extended) re-verified in the same PR.
