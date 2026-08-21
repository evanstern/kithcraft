# Plan 002 — Body protocol v0

**Constitution status:** `.specify/memory/constitution.md` is an unfilled template —
no ratified constitution exists. Per PDLC doctrine this plan is checked against the
project's grounding docs instead: `docs/design/kithcraft-brief.md` (ratified),
`CLAUDE.md`, and `docs/wiki/` (load [[body-protocol-seam]], [[promptworld-lineage]],
[[mod-stack-decision]], [[villager-brain-api]] just-in-time).

## Approach

Design-and-writing task; no code. Three phases:

1. **Port the doctrine.** Read promptworld I's `docs/wiki/` INDEX-first and pull the
   epistemic-hygiene and perception doctrine just-in-time (origin/provenance
   classifier, situated memory, mental-maps/place-facts, salience posture as it
   touches percepts). Distill into a short doctrine-port note in the spec dir
   (`research/doctrine-port.md`) with per-item source pointers (note names, not code),
   so the protocol doc cites ported rules instead of re-deriving them. Feasibility
   cross-check against the first vendor's substrate ([[villager-brain-api]]): every
   perception-channel commitment must be implementable with vanilla sensors/memory
   modules plus owned Mixin surface.
2. **Draft the protocol** (`docs/design/body-protocol-v0.md`): the three surfaces with
   field-level message shapes, the perception model with provenance attachment, the
   versioning story, the abstraction rule for world concepts. Keep shapes
   serialization-neutral (defined as data, JSON-representable); transport is a named
   open question.
3. **Prove the seam.** Add the second-vendor sketch (R3) and the fake/test vendor
   spec (R4) to the protocol doc; sweep the doc for Minecraft type/identifier leaks;
   grow [[body-protocol-seam]]'s sources with the protocol doc and re-verify that
   note (same PR — its own Operational notes require it).

## Structure

- `specs/002-body-protocol-v0/research/doctrine-port.md` — phase 1 output.
- `docs/design/body-protocol-v0.md` — the protocol document (R1–R4).
- `docs/wiki/body-protocol-seam.md` — source list grown, note re-verified.
- No source tree changes; no dependencies added.

## Risks / notes

- **Over-specification is the failure mode:** v0 binds only the seam. Where a choice
  belongs to the mind (memory internals) or the vendor (pathfinding), the protocol
  names the boundary, not the mechanism.
- The perception model must not assume the entity choice (TASK-0003 in flight,
  parallel lane) — channels are defined against "a villager-shaped body", not against
  vanilla-villager or custom-entity specifics.
- promptworld I doctrine is reference, not authority — where I's rules assumed the
  governor/tick machinery that died with it ([[promptworld-lineage]]), adapt to
  real-time-only and say so in the doctrine-port note.
