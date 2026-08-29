---
name: promptworld-lineage
description: What Kithcraft inherits from promptworld I, and what it discards — doctrine transfers (event-sourced memory, reflex/planner split, salience/consolidation, epistemic hygiene, persona firewall; persona TASK-0013, toolloop's REQUEST/FACT/gate shape TASK-0016, and nightly consolidation TASK-0018 all now ported for real); code does not (sim, executor, governor, guardian; 62% of non-test lines). Decision-0003: the daemon does not survive the seam in any language; Go rebuild, old packages kept as source material. Load before importing I code or for mind-architecture questions.
kind: concept
sources:
  - docs/design/kithcraft-brief.md
  - specs/004-mind-daemon-routing/research/daemon-assessment.md
  - docs/design/llm-routing-and-budget.md
  - backlog/decisions/decision-0003 - Mind-daemon-language-Go-rebuilt-behind-the-seam.md
verified_against: eda461376e0bce9bce286137a8295aeb98315e0a
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

**Update (TASK-0013, 2026-08-28): `persona` is no longer source-material-only.**
`mind/persona/` now ports its two-half firewall design for real, ahead of the other
three source-material packages above: genesis (one E1/Opus-5 call per villager,
generating name, values, endogenous desires, an anchor line, and drift markers —
brief #5's "generated, not authored" delta from I — with decision-0002's profession ×
biome-variant pairing carried through, never model-invented), write-once 0444 storage
with no post-genesis write path (proven at the type level by an external-package
reflection test), and a model-free validator (anchor echo + drift lexicon, including
an authored cast-wide moralizing lexicon that vetoes the politeness-policing
spell-breaker at birth) that provably imports no `llm` code. See
specs/013-persona-genesis.

**Update (TASK-0016, 2026-08-28): `toolloop` graduates from source material too.**
`mind/deliberate/` ports its REQUEST/FACT/gate *shape* (never its promptworld-I code,
decision-0003 is unchanged on that point) onto `intent`/`intent_ack`/`act_result`: a
composed intent is the REQUEST, an `act_result` percept is the FACT, and the mind's
admission gate (`mind/memory`) decides what of it becomes memory. E3 job-board
deliberation, the §5.5 urgency interrupt, and the K=10 situated memory window (2 seeded
serendipity picks from the older half — the mind-side formativeness this note's "what
dies" section already flagged as new design, since I's world-side salience table cannot
exist under the seam) all sit above that loop. See specs/016-deliberation.

**Update (TASK-0018, 2026-08-28): the nightly-digestion heritage is also ported for
real.** `mind/consolidate/` lands the machinery *shape* the doctrine list above names —
an event-sourced nightly ledger (`ledger.go`, M2's append-only-log-plus-reducer idiom
applied to a night record), the ordinal `m1..mN` prompt convention (memories have no
stable IDs, so the convention *is* the identity mechanism, mapping accepted references
back to durable `(world_time, hash)` pairs), and the no-marker-on-failure rule (a
transport failure, cancellation, or over-limit response lands nothing; the night
retries). This is doctrine ported as shape, same posture as persona above — no I code
imported. One further wrinkle worth naming because it touches the "forbidden" item two
paragraphs below: the death-carry weighting spike (`deathcarry.go`) reuses I's old
salience table's number for a witnessed death ("witnessed death — 10★",
death-mechanics.md §3) as its next-cycle multiplier. This does not reinstate the
forbidden thing — no `salience`/`importance`/`weight` field exists on any percept, and
the number is consumed entirely mind-side, as one input to a read-time retrieval-weight
function over already-admitted memories, never as a world-side annotation. See
specs/018-consolidation.

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
