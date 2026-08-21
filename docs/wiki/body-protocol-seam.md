---
name: body-protocol-seam
description: The anti-corner architecture move — a world-agnostic body protocol (perceive / act / remember) between mind daemon and world, with the Minecraft mod as the first body vendor. Why minds never couple to Minecraft, what the seam buys, and the drafted v0 contract — five seam invariants, three surfaces, four core verbs. Load for any protocol or architecture work.
kind: pattern
sources:
  - docs/design/kithcraft-brief.md
  - docs/design/mod-stack-comparison.md
  - docs/design/body-protocol-v0.md
verified_against: 50bd22220d5028abb7161fce8c8ec9919e1639e2
---

# Body-protocol seam

The brief's single architecture posture: the mind daemon speaks a world-agnostic protocol
of **perceive / act / remember**, and the Minecraft mod is merely the first *body vendor*
implementing it. Minds never couple to Minecraft. This is the deliberate anti-corner
move — the one interface the project refuses to compromise early.

**The seam is now a contract, not a posture.** TASK-0002 drafted protocol v0 in
`docs/design/body-protocol-v0.md`; that document, not this note, is the contract. What
follows is the summary — load the doc for any field-level work.

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

## What v0 binds

**Five seam invariants** — break one and it is a different protocol, not a revision:
the mind has **no read access** to the world (percepts are the only path in; no query, no
lookup); every percept is **provenance-stamped at emission** from a closed vocabulary;
percepts are **self-contained when they cross** (the mind cannot call back); the body
reports what *is* while the mind owns what was expected and what to keep; **durable memory
belongs to the mind**.

**Three surfaces.** *Perceive* — a one-directional push of typed percepts (sighting,
observation, sound, speech, told_fact, text, act_result, self_state, change_report), each
carrying provenance and an urgency band. *Act* — an intent/result split: the mind sends
intents, the vendor acks receipt only, and what actually happened returns as an
`act_result` **percept**, because an act that returned world state would be a query in
disguise. Four core verbs every vendor must implement: `go_to`, `speak`, `attend`, `wait`.
*Remember* — mostly a statement of what does **not** cross: the vendor persists bodies,
the world, and the token registry; it never persists, ranks, or weights a mind's memories.

**The mechanism that makes epistemic hygiene checkable:** origins are a closed vocabulary
(`acted`/`saw`/`heard`/`felt`/`told`/`read`) and `direct_perception(origin)` is a pure
function of that value alone, defined in the protocol so mind and vendor cannot disagree.
It never reads prose — which is the only thing that stops a language model writing "I saw
it myself" into a belief. An unknown or absent origin classifies **secondhand, never
direct**, which is what makes additive versioning safe.

Two rules that look like fields and are actually invariants: `observation.vocabulary` (the
closed set of kinds scanned — without it an absence claim has no scope) and the
`change_report` **delivery restriction** (never deliver a change report to the body that
caused it or watched it happen — promptworld I measured that bug at **75% of all memories
formed**).

World concepts cross abstracted: opaque `kind` tokens a mind may compare but never parse,
meaning carried as `roles` plus prose-only `descriptor`, space as opaque place tokens plus
coarse bands (no coordinates, no arithmetic), time as integers in a vendor-declared unit
(never ticks). Transport is deliberately undecided — the shapes are data, not a wire format.

The doc proves the seam two ways: a second-vendor sketch (a trivial text world implementing
the whole surface, which exposed that optional percept channels were being assumed rather
than declared) and a specified in-memory **fake vendor** — scripted percepts in, asserted
acts out — that tests the rules no compiler can, including reproducing the 75% flood on
demand by lifting the delivery restriction.

## Connections

Mandated by [[design-brief]]; evaluated as constraint R3 in [[mod-stack-decision]];
first vendor substrate is [[villager-brain-api]]; the reflex/planner doctrine it carries,
and the promptworld I machinery it deliberately does not inherit, are in
[[promptworld-lineage]]. The doctrine port behind v0's cited rule IDs (`EH-n`/`PM-n`/`F-n`)
is `specs/002-body-protocol-v0/research/doctrine-port.md`.

## Operational notes

The protocol doc is the contract; this note is a map to it. Any change to the message
shapes, the origin vocabulary, or the seam invariants lands there first — and a change to
an invariant or to `DIRECT_ORIGINS` membership is a **major** version bump, not a revision.

Transport remains an open question (the doc's §11 Q-1), as does hearing's engine hook in
the first vendor (Q-2). When either is decided, this note's summary of "transport is
deliberately undecided" needs revisiting.
