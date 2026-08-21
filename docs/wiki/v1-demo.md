---
name: v1-demo
description: Milestone m-0 "One real evening" — the v1 demo definition: three villagers with generated personas, schedules and persistent memory; a job-board blueprint built alongside the player; dusk conversation; night danger making the player's walls protect friends. Load when scoping or building toward the demo; docs/design/demo-build-plan.md decomposes it into TASK-0007..0022.
kind: concept
sources:
  - docs/design/kithcraft-brief.md
  - docs/design/demo-build-plan.md
  - backlog/milestones/m-0 - one-real-evening-(v1-demo).md
verified_against: a6bc69488fe2356dd0710cee5c22cdc3e303a699
---

# v1 demo — "one real evening"

Milestone m-0, the project's first deliverable experience: three villagers on a survival
server with the player, one real evening of cohabitation. It is deliberately small — a
household, not a city — and every board task on m-0 exists to reach it: TASK-0001..0006 were
the decision phase, TASK-0007..0022 are the build phase those decisions bought.

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

Defined by [[design-brief]]; board milestone m-0 groups the path to it; decomposed by
`docs/design/demo-build-plan.md` (TASK-0006); runs on the stack chosen in
[[mod-stack-decision]] over the substrate in [[villager-brain-api]], across the seam in
[[body-protocol-seam]].

## Operational notes

**The demo's open questions are all resolved; it now has a ratified build plan.** The five
that fed it closed in order — body protocol (TASK-0002 → `docs/design/body-protocol-v0.md`),
entity implementation (TASK-0003 → decision-0002), mind daemon language + LLM budget
(TASK-0004 → decision-0003), death mechanics (TASK-0005 → `docs/design/death-mechanics.md`),
and the build plan itself (TASK-0006 → `docs/design/demo-build-plan.md`).

The plan decomposes m-0 into **sixteen one-PR-shaped board tasks, TASK-0007…TASK-0022**,
split at the seam: seam (S1–S2), mind/Go (M1–M7), vendor/Java (V1–V5), integration (I1–I2).
Its §8 traces every beat above to the tasks that deliver it, and its §6/§8.2 give the six
suggested lanes for the next sweep. Two operator checkpoints are named there rather than
decided: TASK-0007 (transport) is a proposed `opus` escalation, and TASK-0019 (death and
danger) carries an escalation trigger if the siege-suppression point is not where the design
assumes. The plan also rules nine open items (R-1…R-9) so no downstream task re-derives them.
