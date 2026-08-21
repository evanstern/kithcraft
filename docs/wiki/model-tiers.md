---
name: model-tiers
description: The model-tier ladder for dispatched work — .claude/model-tiers.json declares haiku/sonnet/opus tiers with host-form IDs (cc/…[1m]), tiers.mjs generates the agent definitions, opus is escalation-gated. The verify-served-model doctrine and the restart-after-regenerate trap. Load before any dispatch or tier-config change.
kind: component
sources:
  - .claude/model-tiers.json
  - CLAUDE.md
verified_against: 50c3def435dd9326d38e51118f08944815cbe80c
---

# Model tiers

The posture: **thinking is Opus/Fable-tier, execution is Sonnet/Haiku-tier.** The
orchestrator plans and gates at the top of the ladder; implementing a written spec runs
at the cheapest tier that holds it. The ladder is host config, not doctrine:
`.claude/model-tiers.json` declares it; `pdlc/scripts/tiers.mjs` regenerates
`.claude/agents/<tier>-implementer.md` from it; the generated frontmatter `model:` is
what actually runs at dispatch.

## How it works

Current config (configVersion 1, defaultTier `sonnet`, escalationTier `opus`):

- **haiku** — `cc/claude-haiku-4-5-20251001`, 200K context. Narrow mechanical slices;
  its caution: sized for one phase-scoped dispatch, escalate rather than truncate.
- **sonnet** — `cc/claude-sonnet-5[1m]`, 1M. The default implementer — work to an
  existing pattern or written spec.
- **opus** — `cc/claude-opus-5[1m]`, fallback `cc/claude-opus-4-8[1m]`, 1M,
  `escalation: true`. Design work and real judgment calls; reaching for it is an
  operator checkpoint recorded before dispatch, never an implementer's own call.

Model IDs are written in **this host's form** (`cc/…[1m]` routing-proxy spelling — bare
IDs and aliases are rejected here). The load-bearing rules, each field-proven:

- **Verify the served model from the first dispatch's transcript** before launching
  siblings — neither the frontmatter pin nor the dispatch `model` parameter is proof
  (both observed failing on different hosts, 2026-07-31 and 2026-08-10). TASK-0001's
  sweep verified `cc/claude-sonnet-5[1m]` from each phase's transcript.
- **Regenerate then restart:** the agent registry is read at session start; a newly
  generated definition dispatches as "agent type not found" and an edited one keeps its
  old pin. TASK-0001's sweep hit exactly this — Phase 1 blocked until session restart.
- Never hand-edit a generated definition; `tiers.mjs --check` (CI and sweep
  precondition) reports drift.

## Connections

Dispatch context: [[pdlc-process]] (sweeps name `<tier>-implementer` agents); the config
is planted/refreshed by pdlc:bootstrap per [[overview]]'s process plane.

## Operational notes

Regenerate after every config edit: `node <pdlc>/scripts/tiers.mjs --root .`;
`--check` exits nonzero on drift. A bare tier name in a runbook is invalid — always
tier + explicit model ID + justification, recorded on the board task at dispatch.
