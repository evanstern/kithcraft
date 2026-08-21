---
name: overview
description: The system's shape — what Kithcraft is (Minecraft mod giving the player LLM villagers as company), what exists at this stage (design artifacts, one ratified decision, PDLC machinery, no code), and where each kind of truth lives. Load first for orientation.
kind: concept
sources:
  - README.md
  - docs/design/kithcraft-brief.md
  - CLAUDE.md
verified_against: 50c3def435dd9326d38e51118f08944815cbe80c
---

# Overview

Kithcraft (working lineage label: promptworld II) is a Minecraft server mod that gives an
embodied player LLM villagers as company. The thesis: survival-crafting games are already
fun; what they have never had is company — the AI layer is load-bearing for the *feeling*
of an alive world, not for the game loop. The project is **pre-code**: no mod source
exists yet. What exists is a ratified design brief, one ratified architecture decision
(Fabric server-side mod), verified prior-art research, and the development machinery that
produced them.

## How it works

The repo's planes of truth:

- **Design authority:** `docs/design/kithcraft-brief.md` — ratified 2026-08-19, not to be
  relitigated. See [[design-brief]].
- **Decisions:** `backlog/decisions/` via the `backlog` CLI. decision-0001 (accepted)
  chose the Fabric stack. See [[mod-stack-decision]].
- **Plan of record:** the Backlog.md board (`backlog/tasks/`, CLI-only). Milestone m-0 is
  the v1 demo ([[v1-demo]]); tasks TASK-0002..0006 remain queued.
- **Specs:** `specs/NNN-<slug>/` (Spec Kit), bridged to the board by spec-bridge. One spec
  exists: 001-mod-stack-decision, complete.
- **Process doctrine:** `CLAUDE.md` (PDLC grounding) — see [[pdlc-process]],
  [[model-tiers]], [[root-guard]].

The architecture posture that shapes everything downstream: a world-agnostic
**body protocol** (perceive/act/remember) with the Minecraft mod as the first body
vendor — minds never couple to Minecraft ([[body-protocol-seam]]). Doctrine transfers
from promptworld I; code does not ([[promptworld-lineage]]).

## Connections

Every other note grounds one region of this map: design ([[design-brief]],
[[body-protocol-seam]], [[v1-demo]]), evidence ([[prior-art]], [[villager-brain-api]]),
decisions ([[mod-stack-decision]]), machinery ([[pdlc-process]], [[model-tiers]],
[[root-guard]]).

## Operational notes

Pre-code means the wiki's sources are docs and config, not `src/`. The first
implementation PR (expected from TASK-0002/0003/0006) will obligate a wiki-update pass
and new component notes.
