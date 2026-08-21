# Plan 005 — Death mechanics design

**Constitution status:** `.specify/memory/constitution.md` is an unfilled template —
no ratified constitution exists. Per PDLC doctrine this plan is checked against the
project's grounding docs instead: `docs/design/kithcraft-brief.md` (ratified),
`CLAUDE.md`, and `docs/wiki/` (load [[design-brief]], [[v1-demo]],
[[villager-brain-api]], [[body-protocol-seam]] just-in-time).

## Approach

Design-writing task; no code. Three phases, mirroring specs 001/003's shape:

1. **Verify the vanilla death surface.** What decision-0002's augmented villager
   inherits: villager damage sources and immunities, mob targeting at night, panic/
   flee behavior, hunger (villagers don't starve in vanilla — verify), death drops,
   despawn rules. Every fact URL + accessed date (evidence rule). Build on — do not
   duplicate — specs/003's engine-behavior research where it already covers death/
   despawn. Output: `specs/005-death-mechanics/research/death-surface.md`.
2. **Write the design** (`docs/design/death-mechanics.md`): causes admitted/
   suppressed, remains (grave/belongings/markers), and memory/conversation carry —
   each choice argued from the loneliness-cure thesis and minds-are-others
   tie-breakers; the micromanagement answer (R2) and shrinking-cast stance (R3) as
   named sections. Memory-entry shapes must be expressible as body-protocol percepts
   (saw vs told provenance) without extending the protocol — flag any tension for
   the seam's owner rather than deciding.
3. **Ratification prep.** Cross-check the design against the brief's spell-breakers
   and the demo definition ([[v1-demo]]); state what TASK-0006's build plan inherits
   (which mechanics are demo-scope vs later). Operator ratifies at PR review.

## Structure

- `specs/005-death-mechanics/research/death-surface.md` — Phase 1 evidence.
- `docs/design/death-mechanics.md` — Phase 2 deliverable.
- Wiki: notes citing the brief's death posture ([[design-brief]], [[v1-demo]]) are
  re-verified in this PR only if the design contradicts or extends their prose.

## Risks / notes

- Emotional-design choices (grave form, story cadence) are taste calls: the doc
  argues them from the two tie-breakers and leaves ratification to the operator.
- Fragility tuning has no playtest data yet; the doc states tunable defaults with
  the tuning knob named, not fake precision.
