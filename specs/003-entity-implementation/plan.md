# Plan 003 — Entity implementation decision

**Constitution status:** `.specify/memory/constitution.md` is an unfilled template —
no ratified constitution exists. Per PDLC doctrine this plan is checked against the
project's grounding docs instead: `docs/design/kithcraft-brief.md` (ratified),
`CLAUDE.md`, and `docs/wiki/` (load [[mod-stack-decision]], [[villager-brain-api]],
[[design-brief]], [[v1-demo]] just-in-time).

## Approach

Research-and-writing task; no code. Three phases, mirroring spec 001's shape:

1. **Verify engine behavior.** For each constraint area in R1, pull current facts
   about vanilla behavior and Fabric modding surface from primary sources (Minecraft
   wiki mechanics pages, Fabric docs, Yarn-mapped source references): mob targeting of
   villagers, sleep/workstation/POI mechanics, death/despawn rules, breeding/trading/
   gossip behaviors and their disable points, client-side rendering of custom entities
   vs villager variants. Every fact gets a URL and accessed date. Store in
   `specs/003-entity-implementation/research/engine-behavior.md`.
2. **Write the comparison** (`docs/design/entity-implementation-comparison.md`):
   per-option fit against the six constraint areas, the Mixin surface each option
   owns, risks. Build on — do not duplicate — [[villager-brain-api]]'s substrate
   findings and spec 001's research.
3. **Draft the recommendation + decision record.** One option, rationale mapped to
   the constraints, recorded via `backlog decision create` (CLI only), proposed —
   pending operator ratification. State narrowing effects on TASK-0006 (demo build
   plan) explicitly.

## Structure

- `specs/003-entity-implementation/research/engine-behavior.md` — phase 1 evidence.
- `docs/design/entity-implementation-comparison.md` — the comparison (R1).
- Backlog decision record — the decision artifact (R2).
- No source tree changes; no dependencies added.

## Risks / notes

- Vanilla behavior claims are version-sensitive — pin claims to the MC version the
  prior research targeted (yarn-1.21.3 era per [[villager-brain-api]]) and note where
  behavior differs across recent versions.
- The night-danger and multiplayer constraints are emotionally load-bearing
  ([[design-brief]], [[v1-demo]]) — where evidence is thin (e.g. exact custom-entity
  client rendering limits), say so rather than round off; a wrong claim here
  invalidates the decision.
- Do not decide protocol matters (TASK-0002's lane); flag interactions only.
