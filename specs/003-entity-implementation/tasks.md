# Tasks 003 — Entity implementation decision

## Phase 1 — Engine behavior evidence

- [x] Hostile-mob targeting of villagers (zombie pursuit, raids, panic): mechanics verified for both options (dated, cited)
- [x] Village fiction mechanics (beds/sleep, workstations, POI claim, schedules): what vanilla Villager inherits free vs what a custom entity must wire (dated, cited)
- [x] Death/despawn/permadeath surface: vanilla death handling, what must be suppressed (curing, restocking) per option (dated, cited)
- [x] Unwanted vanilla behaviors (trading UI, breeding, gossip, iron-golem summoning): disable points and Mixin surface per option (dated, cited)
- [x] Client-side visibility: what a vanilla client renders for an augmented villager vs a custom entity; skin/appearance flexibility without a client mod (dated, cited)
- [x] Evidence file specs/003-entity-implementation/research/engine-behavior.md committed

## Phase 2 — Comparison document

- [ ] docs/design/entity-implementation-comparison.md drafted: both options against all six constraint areas, per-option Mixin surface and risks
- [ ] Interactions with the body-protocol seam flagged without deciding for TASK-0002
- [ ] Every claim carries a URL and accessed date

## Phase 3 — Recommendation & decision record

- [ ] Recommendation written into the comparison doc: one option, rationale mapped to the ratified constraints
- [ ] Backlog decision record created (proposed — pending operator ratification)
- [ ] Narrowing effects on TASK-0006 (demo build plan) stated explicitly
