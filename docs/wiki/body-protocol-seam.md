---
name: body-protocol-seam
description: The anti-corner architecture move — a world-agnostic body protocol (perceive / act / remember) between mind daemon and world, with the Minecraft mod as the first body vendor. Why minds never couple to Minecraft, what the seam buys, and how it constrained the mod-stack decision. Load for any protocol or architecture work (TASK-0002 especially).
kind: pattern
sources:
  - docs/design/kithcraft-brief.md
  - docs/design/mod-stack-comparison.md
verified_against: 50c3def435dd9326d38e51118f08944815cbe80c
---

# Body-protocol seam

The brief's single architecture posture: the mind daemon speaks a world-agnostic protocol
of **perceive / act / remember**, and the Minecraft mod is merely the first *body vendor*
implementing it. Minds never couple to Minecraft. This is the deliberate anti-corner
move — the one interface the project refuses to compromise early.

## How it works

The seam's guaranteed consequences (from the brief):

- Whether promptworld I's Go mind daemon survives or the mind layer is rebuilt fresh is
  an implementation detail *behind* the seam, not an early commitment.
- A future owned world engine (V2) is a second body vendor, not a rewrite.
- Minds are testable without booting Minecraft.

The seam also did real work in the mod-stack decision: the comparison
(`docs/design/mod-stack-comparison.md`) found both Fabric and Paper keep the seam equally
clean — a mod/plugin is a thin server-side surface with no requirement that mind logic
live in-process — so the seam alone didn't discriminate; dependency health did. What the
seam *did* rule on: Citizens2's OSL-3.0 copyleft posed a license-entanglement risk to a
tightly-coupled body-vendor implementation, while Fabric (Apache-2.0) plus vanilla engine
API leaves the body vendor freely licensable.

The protocol itself is **not yet drafted** — that is TASK-0002 ("body protocol v0"), the
next contract-shaped work on the board. Under [[mod-stack-decision]], its first vendor
implements against the Fabric brain substrate ([[villager-brain-api]]).

## Connections

Mandated by [[design-brief]]; evaluated as constraint R3 in [[mod-stack-decision]];
first vendor substrate is [[villager-brain-api]]; the reflex/planner doctrine it will
carry is in [[promptworld-lineage]].

## Operational notes

When TASK-0002 lands the protocol draft, this note's sources must grow to include the
protocol artifact, and the note re-verified — the seam moves from posture to contract.
