---
name: prior-art
description: The verified prior-art landscape (re-checked 2026-08-20) — bot-as-player family (Mineflayer, Voyager, Mindcraft, Project Sid) vs server-mod family (Citizens2, CraftAgent, AI_NPC, SecondBrain), each project's health/license, the documented dead ends, and the gap Kithcraft claims. Load before relying on or citing any external project.
kind: concept
sources:
  - specs/001-mod-stack-decision/research/prior-art.md
  - docs/design/kithcraft-brief.md
verified_against: 50c3def435dd9326d38e51118f08944815cbe80c
---

# Prior art

Two families of Minecraft LLM-agent work exist; Kithcraft sits in the second. Every fact
here was re-verified 2026-08-20 with URLs and accessed dates in
`specs/001-mod-stack-decision/research/prior-art.md` — cite that file, not memory.

## How it works

**Bot-as-player family** (agents join as fake player clients; ruled out by
[[design-brief]] #3): Mineflayer (mature JS bot library), Voyager (LLM self-curriculum
research), Mindcraft (active multi-agent framework), Project Sid (Altera's 500–1000 agent
civilization runs; its PIANO mind architecture is worth reading).

**Server-mod family** (NPCs as real entities), with 2026-08-20 health:

- **Citizens2** — 2.0.43; active but lagging (last free-repo commit 2026-02-16,
  user-reported breakage on newest MC builds); **OSL-3.0 — copyleft, GPL-incompatible**,
  the flag that decided against the Paper option.
- **CraftAgent** — **documented dead end**: last push 2026-01-06 (~7.5 months stale),
  pinned to MC 1.20.1/1.21.8, single contributor; LGPL-3.0. Design reference only
  (SQLite conversation memory, world perception, action handlers, web dashboard).
- **SecondBrain** (sailex428) — CraftAgent's materially-more-active sibling (last push
  2026-03-31, 46 releases, LGPL-3.0). Design reference, not a dependency — an
  opinionated LLM-NPC framework, not infrastructure.
- **AI_NPC** — very new (first public release 2026-07-23); **license unverifiable** (no
  public source repo found) — documented dead end on that fact.

**Healthy infrastructure:** Fabric Loader 0.19.3 / Fabric API 0.158.0+26.2 (active,
Apache-2.0); Paper 26.2 (active, GPLv3). The vanilla villager brain API is engine
surface, no third-party health to track ([[villager-brain-api]]).

**The gap Kithcraft claims** (brief): everything above is a research benchmark or a chat
skin; nobody has shipped persistent villagers with memory, relationships, desires, and
mortality, cohabiting a survival world with an embodied player over weeks. The parts all
demonstrably work; the synthesis is unclaimed.

## Connections

Evidence base for [[mod-stack-decision]]; the family split enforces [[design-brief]]'s
villager-shaped constraint; brain-API detail in [[villager-brain-api]].

## Operational notes

The space moves fast — the 2026-08-20 check superseded 2026-08-19 citations for currency
after one day. Re-verify before any new reliance; a claim without URL + accessed date
doesn't count (project evidence rule).
