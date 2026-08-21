# Plan 004 — Mind daemon language & LLM routing/budget

**Constitution status:** `.specify/memory/constitution.md` is an unfilled template —
no ratified constitution exists. Per PDLC doctrine this plan is checked against the
project's grounding docs instead: `docs/design/kithcraft-brief.md` (ratified),
`CLAUDE.md`, and `docs/wiki/` (load [[promptworld-lineage]], [[body-protocol-seam]],
[[design-brief]], [[v1-demo]] just-in-time; promptworld I's own `docs/wiki/` corpus
INDEX-first for the daemon assessment).

## Approach

Research-and-writing task; no code. Three phases, mirroring specs 001/003's shape:

1. **Assess promptworld I's daemon against the seam.** Read I's wiki corpus
   (INDEX-first, notes just-in-time) and, where the wiki points, its code layout — to
   classify the mind-side machinery: what implements transferred doctrine (portable),
   what is entangled with the dead world engine (dies), and what the seam contract
   (body-protocol-v0.md) would require of a daemon regardless of language. Output:
   `specs/004-mind-daemon-routing/research/daemon-assessment.md`, every code/doc claim
   carrying a path or URL reference.
2. **Draft the routing sketch + cost envelope**
   (`docs/design/llm-routing-and-budget.md`): event → tier mapping with cadence and
   latency posture per event class (R2), then the cost math (R3) with current model
   pricing (URL + accessed date; the claude-api reference skill is the primary source
   for Anthropic pricing). The routing sketch respects the reflex/planner split: the
   engine-side brain (decision-0002's augmented villager) owns doing; the daemon owns
   choosing and relating.
3. **Draft the recommendation + decision record.** One language/reuse choice,
   rationale mapped to R1's criteria and the routing sketch's demands, recorded via
   `backlog decision create` (CLI only), proposed — pending operator ratification.
   State narrowing effects on TASK-0006 (demo build plan) explicitly.

## Structure

- `specs/004-mind-daemon-routing/research/daemon-assessment.md` — Phase 1 evidence.
- `docs/design/llm-routing-and-budget.md` — Phase 2 deliverable (routing + cost).
- `backlog/decisions/decision-0003 - …` — Phase 3 decision record (via CLI).
- Wiki: [[promptworld-lineage]]'s Operational notes anticipate exactly this task —
  its prose and sources are re-verified in this PR if the daemon assessment adds
  sources or contradicts its summary.

## Risks / notes

- The daemon assessment reads a *sibling repo* (promptworld). Read-only: nothing in
  promptworld is modified.
- Cost numbers rot; the envelope states its price-accessed dates and the sensitivity
  note bounds the estimate rather than pretending precision.
- Language choice could tempt transport decisions (seam Q-1). Out of scope: flag
  implications, decide nothing.
