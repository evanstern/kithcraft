---
name: promptworld-lineage
description: What Kithcraft inherits from promptworld I and what it deliberately discards — doctrine transfers (event-sourced memory, reflex/planner split, salience/consolidation, epistemic hygiene, persona firewall), code does not (sim engine, executor, governor, guardian all die; 62% of I's non-test lines, measured). Carries decision-0003's resolution: the daemon does not survive the seam in any language, so the mind is rebuilt in Go with four named packages as source material. Load when a design question touches mind architecture or when tempted to import promptworld I code.
kind: concept
sources:
  - docs/design/kithcraft-brief.md
  - specs/004-mind-daemon-routing/research/daemon-assessment.md
  - docs/design/llm-routing-and-budget.md
  - backlog/decisions/decision-0003 - Mind-daemon-language-Go-rebuilt-behind-the-seam.md
verified_against: 7acc232f5cd5cc1008acb56dbab5017337196174
---

# Promptworld lineage

Kithcraft is "promptworld II": promptworld I is retroactively the R&D project where the
theory of agent minds got worked out. II throws away the world engine and keeps the
minds — as **doctrine, not code**. Nothing in II imports I's code; I's repo (its
`docs/wiki/` corpus) is grounded reference only.

## How it works

**Doctrine that transfers** (brief, "Architecture posture"):

- **Event-sourced memory.**
- **The reflex/planner split** — scripted competence for *doing* (pathfind, chop, place
  blocks), LLM for *choosing and relating*. Cheap reflexes for competence, expensive
  thoughts for meaning; the brief calls this the single most transferable lesson of I.
- **Salience + consolidation + situated memory** — the nightly-digestion heritage.
- **Epistemic hygiene** — an agent knows only what it saw or was told, with provenance.
- **The persona firewall.**

**What dies with promptworld I:** the sim engine, executor, reducer, terrain generation,
TUI, determinism-for-replay machinery, the governor/speed ladder, and the guardian (the
player-as-intermediary concept as a whole). The real-time-only decision ([[design-brief]]
#8) is what deletes the governor family: at 1x there is no cognition horizon, no speed
ladders, no staleness budgets.

The inversion driving the split: I built a fascinating simulation with nothing for the
player to do; II puts the player in a game that's already fun and gives them neighbors.

**Measured, not asserted.** TASK-0004's assessment counted I's tree: `sim` (22,035 non-test
lines), `tui` (14,044), `guardian` (7,372), and `cognition` (996) die outright — **44,447 of
71,744 non-test lines, 62% of the codebase**, before counting the `mind`/`llm` code written to
serve them.

## Does I's daemon survive? No — resolved by decision-0003

The brief left this as "an implementation detail behind the seam". TASK-0004 answered it, and
the answer is sharper than keep-or-rebuild: **the daemon does not survive the seam in any
language.** I's `internal/mind` is not a daemon that happens to live in the same repo as a
world — it is a **co-process of the world engine**. It holds a replica of `sim.State` and
applies every event to it, reads that state directly for prompt assembly, lands output through
in-process injection doors typed in world-engine terms (`InjectIntent(sim.InjectArgs)`), and
bakes the cast size into ten fixed-size array types over `sim.AgentCount = 8`. Behind
[[body-protocol-seam]] all four are replaced — replica → percept inbox, direct state reads → a
private provenance-stamped belief store, injection doors → `intent`/`act_result`, fixed arrays
→ a variable cast with permadeath.

So the real question was never keep-vs-rebuild; it was **which language carries the doctrine,
the portable assets, and the operator at the lowest cost.** decision-0003 (accepted 2026-08-21,
PR #9) answers **Go, rebuilt** — with four packages as *source material, not a
codebase*: `toolloop` (994 lines, sim-agnostic by construction — its REQUEST/FACT/gate doctrine
maps one-to-one onto `intent`/`intent_ack`/`act_result`), `persona` (475 lines, no imports),
`tool`'s registry mechanism (vocabulary replaced by the runtime manifest), and `llm`'s provider
layer minus its dead `cognition` dependency.

Two findings worth carrying, because they change how the doctrine list above should be read:

- **Most of the doctrine is already carried by the protocol, not by any code decision.**
  Epistemic hygiene in particular is *fully* ported — into `body-protocol-v0`'s §2.7 classifier
  and RM-1…RM-7, not into files. The mechanisms were always tiny (8 and 30 lines); the value
  was the knowledge, and the knowledge is now binding contract text.
- **Some doctrine is carried by the engine now.** The reflex half of the split is Minecraft's
  own `Brain<E>`/`Schedule`/POI stack under [[mod-stack-decision]] and decision-0002, so I's
  pathfinder and survival ladder are dead code *and* a solved problem. And one item is
  **forbidden** rather than transferred: I's world-side salience table cannot exist under the
  seam (no `salience`/`importance`/`weight` field on any percept), which leaves mind-side
  formativeness as new design.

## Connections

Doctrine list lives in [[design-brief]]; the seam the rebuilt mind sits behind is
[[body-protocol-seam]]; the reflex side of the split maps onto the engine's own brain
machinery ([[villager-brain-api]]) under [[mod-stack-decision]].

## Operational notes

TASK-0004 decided the language (decision-0003, **accepted** — operator ratified by merging
PR #9, 2026-08-21). With mind-daemon work now unblocked, the
transferred doctrine items are checkable requirements against I's wiki corpus — consult I's
`docs/wiki/` INDEX just-in-time rather than porting files. The one sanctioned exception to
"never read I's tree" is on the record and was exercised once: TASK-0004's assessment read
paths as *evidence about size and coupling*, because "portable" is a property of code and not
of a summary. Nothing was imported.
