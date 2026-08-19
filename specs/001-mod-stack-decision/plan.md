# Plan 001 — Mod stack decision

**Constitution status:** `.specify/memory/constitution.md` is an unfilled template —
no ratified constitution exists. Per PDLC doctrine this plan is checked against the
project's grounding docs instead: `docs/design/kithcraft-brief.md` (ratified) and
`CLAUDE.md` (PDLC grounding). No `docs/wiki/` exists yet (pre-first-build); the brief
is the sole design authority.

## Approach

This is a research-and-writing task; no code. Three phases:

1. **Re-verify prior art.** For each load-bearing project (Fabric + Fabric API,
   Paper, Citizens2, CraftAgent, AI_NPC, and the villager brain API surface), pull
   current facts from primary sources (repos, release pages, license files, docs):
   latest release + date, supported MC versions, commit/maintenance cadence, license.
   Web research; every fact gets a URL and an accessed date.
2. **Write the comparison** (`docs/design/mod-stack-comparison.md`): per-option
   capability fit against the ratified constraints (R3 list), dependency health from
   phase 1, risks. Include the hybrid option honestly (what it would buy, what it
   costs).
3. **Draft the recommendation + decision record.** One stack, rationale mapped to
   R3, recorded via `backlog decision create` (CLI only — never hand-edit
   `backlog/`). Mark the decision as proposed-pending-operator-ratification.

## Structure

- `docs/design/mod-stack-comparison.md` — the evidence artifact (R1).
- Backlog decision record — the decision artifact (R2/R3).
- No source tree changes; no dependencies added.

## Risks / notes

- Web sources may disagree or be stale; prefer primary (repo releases, license files)
  over aggregators, and date every claim.
- The comparison must not silently narrow TASK-0003's entity-implementation options;
  where a stack choice constrains them, say so explicitly.
