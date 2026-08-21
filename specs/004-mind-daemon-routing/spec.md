# Spec 004 — Mind daemon language & LLM routing/budget for a small cast at 1x

**Board task:** TASK-0004 · **Milestone:** One real evening (v1 demo)
**Direction source:** docs/design/kithcraft-brief.md (ratified 2026-08-19 — do not
relitigate); body protocol v0 (`docs/design/body-protocol-v0.md`, contract accepted
2026-08-21) fixes the seam the mind daemon sits behind; decision-0002 (augmented vanilla
villager, accepted 2026-08-21) fixes the first body vendor's shape. The reflex/planner
split is settled doctrine ([[promptworld-lineage]]): scripted competence for doing, LLM
only for choosing and relating.

## Problem

The mind layer cannot start being built until two coupled questions are answered. (a)
**Language / reuse-vs-rebuild:** promptworld I's Go daemon exists; whether it survives
behind the body-protocol seam or the mind layer is rebuilt fresh is an implementation
detail *behind* the seam — but it must be decided to write the first line of mind code.
(b) **LLM routing and budget:** which villager cognition events call an LLM, at which
model tier, at what cadence — and what an evening of play with 3–6 villagers costs.
Real time (1x) is load-bearing: a villager taking 20s to decide is a person mulling, so
there is no governor/speed-ladder machinery and latency tolerance is generous.

## Requirements (mapped to the card's acceptance criteria)

### R1 — Language / reuse-vs-rebuild decision (card AC #1)

A decision recorded as a Backlog decision record with rationale, weighing at minimum:

- **What promptworld I's daemon actually contains** vs what the seam contract needs —
  assessed from I's repo (`/Users/evanstern/projects/promptworld`, its `docs/wiki/`
  corpus INDEX-first), classifying I's mind-side code as portable-behind-the-seam vs
  entangled with the dead world engine (sim engine, executor, governor, guardian).
- **Candidate languages** (at minimum: keep Go / rebuild in Go / rebuild in
  TypeScript or another candidate the evidence motivates), each judged on: fit for an
  LLM-orchestration daemon (SDK maturity, async story), seam-contract implementation
  effort, testability against the fake vendor spec, and operator maintainability.
- The doctrine-transfer checklist: whichever choice wins must state how each
  transferred doctrine item (event-sourced memory, reflex/planner split,
  salience/consolidation, epistemic hygiene, persona firewall) is carried.

### R2 — LLM routing sketch (card AC #2)

A routing sketch covering, for a 3-villager evening at 1x:

- **Event → tier mapping:** which cognition events call an LLM (e.g. job-board
  decision, dusk conversation turns, salience scoring, nightly consolidation,
  greeting/reaction lines) and which stay scripted reflexes (pathfinding, block
  placement, schedule following — engine-side per decision-0002).
- **Expected cadence** per event class for a 3-villager evening (calls/hour, tokens
  per call, context shape), with the assumptions stated.
- **Latency posture** per event class: what can take 20s (a decision mulled), what
  needs a faster tier or pre-generation (conversation flow), what runs offline
  (consolidation during sleep).

### R3 — Cost envelope (card AC #3)

A written cost estimate for the one-real-evening demo (3 villagers, one ~3-hour
evening) with assumptions explicit: model prices carry URL + accessed date (evidence
rule), token estimates per event class, cadence math shown, and a sensitivity note
(what happens at 6 villagers, what happens if conversation is chattier than assumed).

### R4 — Operator ratification (card AC #4)

The decision record lands proposed; the operator ratifies at PR review before the card
moves Done. Ratification is the human gate on both the language choice and the
routing/budget posture.

## Non-goals

- No mind-daemon code, no protocol changes (the seam contract is accepted; transport
  remains open question Q-1 there and is NOT decided here unless the language choice
  forces a note — flag, don't decide).
- No re-litigation of real-time-only, the small cast, or the reflex/planner split.
- No model fine-tuning or local-model hosting design; v1 assumes hosted APIs.
