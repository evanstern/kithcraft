# Kithcraft grounding wiki — INDEX

Load this file first; load notes just-in-time. Whole-corpus orientation: `CAPSULES.md`.

Two real code surfaces ground the corpus: the mind daemon (`mind/`, TASK-0008) and, as of
TASK-0009, the Fabric mod vendor (`mod/`) — alongside the design surface (brief, decisions,
prior art) and the development machinery (PDLC, board, tiers, hooks).

## Design

- [[overview]] — the project's shape: what Kithcraft is, what exists so far, where truth lives.
- [[design-brief]] — the ratified brief: thesis, ten ratified decisions, spell-breakers.
- [[body-protocol-seam]] — the anti-corner move: world-agnostic perceive/act/remember; mod as first body vendor; and the wire beneath it (UDS, length-prefixed canonical JSON, golden vectors).
- [[promptworld-lineage]] — what transfers from promptworld I (doctrine) and what died with it (code).
- [[v1-demo]] — milestone m-0 "one real evening": the demo definition and its emotional load-bearing walls.

## Decisions & evidence

- [[mod-stack-decision]] — decision-0001 (accepted): Fabric server-side mod; rationale and accepted risks.
- [[prior-art]] — the verified landscape: bot-as-player vs server-mod families, dependency health, dead ends.
- [[villager-brain-api]] — the Fabric/vanilla `Brain<E>` substrate: what's plain API, what needs Mixin.

## Process & machinery

- [[pdlc-process]] — the development loop: board, specs, spec-bridge, sweeps, gates.
- [[model-tiers]] — the tier ladder: config, generated agents, pin verification doctrine.
- [[root-guard]] — root-read-only enforcement: the hook, its jurisdiction, the backlog exception.
