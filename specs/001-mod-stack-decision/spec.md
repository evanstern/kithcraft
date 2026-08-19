# Spec 001 — Mod stack decision: Fabric vs Paper/Citizens2 vs hybrid

**Board task:** TASK-0001 · **Milestone:** One real evening (v1 demo)
**Direction source:** docs/design/kithcraft-brief.md (ratified 2026-08-19 — do not relitigate)

## Problem

Kithcraft is a Minecraft server mod that gives an embodied player LLM villagers as
company. Every downstream decision — entity implementation (TASK-0003), the body
protocol's first vendor (TASK-0002), the demo build plan (TASK-0006) — depends on which
mod stack the project builds on. The brief names the choice as the first open question:
**Fabric vs Paper/Citizens2 vs hybrid.** The prior-art links in the brief were verified
2026-08-19; the modding and LLM-agent spaces move fast, so the evaluation must re-verify
before relying on anything.

## Requirements (mapped to the card's acceptance criteria)

### R1 — Evidence-backed comparison (card AC #1)

A comparison document at `docs/design/mod-stack-comparison.md` covering, for each of
Fabric, Paper (with Citizens2), and hybrid approaches:

- Current stable version and target Minecraft version support.
- Maintenance status of load-bearing prior art: Citizens2, CraftAgent, AI_NPC, and the
  Fabric villager brain API surface (activity/schedule injection, POI, memory modules).
- License of every dependency the option would rely on, with compatibility noted.
- Fit against the ratified constraints (see R3).
- Every factual claim dated and cited with a URL (claims are re-verified now, not
  inherited from the brief).

### R2 — Recommendation as a decision record (card AC #2)

A single recommended stack with rationale, recorded as a Backlog decision record (via
`backlog decision create`, in the same PR). The recommendation is **proposed** by this
task and **ratified** by the operator — ratification is an operator checkpoint at the
end of the sweep, not something this spec's execution performs.

### R3 — Constraint fit stated explicitly (card AC #3)

The recommendation must explicitly address:

- **Body-protocol seam:** the mod is the first *body vendor* implementing
  perceive/act/remember; minds never couple to Minecraft. The chosen stack must not
  force mind logic into the mod layer.
- **Villager-shaped, not bot clients:** server-mod architecture family (NPCs as real
  server entities), not Mineflayer-style fake player clients.
- Secondary ratified constraints where they discriminate between options: small cast
  (3–6), real-time only, vanilla night danger threatening villagers, drop-in
  multiplayer, riding the village fiction (beds, workstations, schedules).

## Out of scope

- The entity implementation choice itself (custom entity vs augmented villager) —
  that is TASK-0003; this task only ensures the chosen stack doesn't foreclose it.
- The body protocol draft (TASK-0002).
- Any implementation, scaffolding, or dependency installation.

## Done means

Comparison doc exists with dated, cited evidence; decision record exists proposing one
stack with rationale addressing R3; both land in this task's single PR; operator
ratification recorded on the decision before the card syncs Done.
