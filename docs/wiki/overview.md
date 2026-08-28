---
name: overview
description: The system's shape — what Kithcraft is (Minecraft mod giving the player LLM villagers as company), what exists at this stage (design artifacts, one ratified decision, PDLC machinery, the mind daemon with its own event-sourced memory, its E1-generated persona firewall, and the first vendor-side Fabric mod code), and where each kind of truth lives. Load first for orientation.
kind: concept
sources:
  - README.md
  - docs/design/kithcraft-brief.md
  - CLAUDE.md
verified_against: c4278829d8dc3b019e4f73fb25b724395c961ac8
---

# Overview

Kithcraft (working lineage label: promptworld II) is a Minecraft server mod that gives an
embodied player LLM villagers as company. The thesis: survival-crafting games are already
fun; what they have never had is company — the AI layer is load-bearing for the *feeling*
of an alive world, not for the game loop. What exists is a ratified design brief, one
ratified architecture decision (Fabric server-side mod), verified prior-art research, the
development machinery that produced them, the mind daemon skeleton in `mind/` (TASK-0008,
M1) — since TASK-0010 (M2) carrying its own event-sourced memory: an append-only log, a
private provenance-stamped belief store, and the deterministic episodic admission gate
([[body-protocol-seam]]) — since TASK-0011 (M4) also an LLM client (`mind/llm/`) routing
the six villager-cognition classes across three model tiers, with per-class prompt
assembly (`mind/prompt/`) keeping each class's stable prefix byte-identical for caching,
and per-class call/token accounting — since TASK-0013 (M3) also `mind/persona/`: E1
(Opus 5) generates each villager's name, values, endogenous desires, anchor line, and
drift markers at birth, written once at mode 0444 with no post-genesis write path, and
guarded by a model-free validator that rejects stated drift with zero LLM
involvement — and, as of TASK-0009 (V1), the **first vendor-side code**: a real Fabric
mod in `mod/` implementing the wire client, `session_open` handshake, capability
manifest, and token registry ([[body-protocol-seam]]).

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

Two real code surfaces now exist alongside the docs/config majority of the wiki's
sources: the mind daemon (`mind/`, TASK-0008/M1, now carrying TASK-0010/M2's
event-sourced memory in `mind/memory/`, TASK-0011/M4's LLM client and prompt assembly
in `mind/llm/` and `mind/prompt/`, and TASK-0013/M3's persona genesis and firewall in
`mind/persona/`) and the Fabric mod vendor (`mod/`, TASK-0009/V1) —
[[body-protocol-seam]] cites both. This note's claims were re-verified against `mod/`'s,
`mind/memory/`'s, `mind/llm/`'s, `mind/prompt/`'s, and `mind/persona/`'s existence in
this pass (2026-08-28). The `mind/persona/` live-genesis proof
(specs/013-persona-genesis/live-run.md) is a recorded, environment-blocked attempt as
of this pass — unit tests against a mocked client are green, real personas are not yet
on disk. The mod's
Yarn-mapped brain-API surface ([[villager-brain-api]]) is flagged, not yet
re-derived, against MC 26.2's official mappings — that re-derivation is V3's
(TASK-0014) scoped work.
