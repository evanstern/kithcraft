---
name: overview
description: The system's shape — what Kithcraft is (Minecraft mod giving the player LLM villagers as company), what exists at this stage (design artifacts, PDLC machinery, the mind daemon with event-sourced memory, persona firewall, bounded deliberation loop, nightly consolidation and archival, and the first vendor-side Fabric mod code now including death/remains and job-board/blueprint-build machinery), and where each kind of truth lives. Load first for orientation.
kind: concept
sources:
  - README.md
  - docs/design/kithcraft-brief.md
  - CLAUDE.md
  - specs/019-death-remains/spec.md
  - specs/020-job-board/spec.md
verified_against: 6055fd0fd5b3949d0e65e99ab009eb1c675cc0c2
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
involvement — since TASK-0017 (M6) also `mind/converse/`: the dusk exchange (E4, Sonnet
5, thinking off, streaming, ~300 max_tokens under the design's one hard latency
ceiling) between two villagers, pre-generated off V3's pair-formation signal so the
scene opens instantly, and the E5 ambient pool (Haiku 4.5, one batched call per
villager per in-game day serving in-process in well under the 200 ms budget) with
specific-remark escalation to a live call — since TASK-0018 (M7) also
`mind/consolidate/`: a nightly ledger
consolidating each day's admitted buffer under Opus 5 (E6) into a digest, the ordinal
`m1..mN` prompt convention resolving accepted references back to durable identity,
no-marker-on-failure retry semantics, a death-carry retrieval-weighting curve (RM-7:
faded, never deleted), and R-9 archival (no new session for a dead mind, its log stays
readable, its body token retired and never reissued) — and, as of TASK-0009 (V1), the
**first vendor-side code**: a real Fabric mod in `mod/` implementing the wire client,
`session_open` handshake, capability manifest, and token registry
([[body-protocol-seam]]). TASK-0016 (M5) then added
`mind/deliberate/`: the bounded E2/E3 deliberation loop, porting toolloop's REQUEST/FACT/gate
shape (decision-0003, [[promptworld-lineage]]) onto intent/intent_ack/act_result, plus the
§5.5 urgency interrupt and the K=10 situated memory window — a villager now claims or
declines a posted job with an authored reason of its own, rather than executing a command. Since TASK-0019 (V5) the mod
also carries death/remains machinery: zombie sieges suppressed and zombie-conversion
cancelled to death-equivalent (two more targeted Mixins, [[villager-brain-api]]), a
named grave + belongings chest placed at the death site with no villager agency
required, a configurable grief period holding the dead villager's bed/job-site
unclaimed, an optional job-board posting, and the dead villager's body token retired
and never reissued. Since TASK-0020 (V4) the mod also carries a job-board/blueprint-build
surface: a fixed-position lectern accepting any free-text posting (no form, no syntax), a
read cycle riding `Activity.MEET` exactly as V5's dusk pairing already does, a `claim` verb
added to the manifest via §5.5's declared-extras half ([[body-protocol-seam]]) so a mind
claims a posting with its own authored reason, a tiny deterministic blueprint parser
(shape/size/material, heavily ponytailed), and block-by-block placement on a tick budget
during `Activity.WORK` with schedule/danger interrupt-resume via a persisted cursor — no new
Mixin (the whole surface is plain-JDK/vanilla-API). V5's grave posting now rides this same
board for real, closing the seam that task's own doc had flagged as pending.

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
in `mind/llm/` and `mind/prompt/`, TASK-0013/M3's persona genesis and firewall in
`mind/persona/`, TASK-0016/M5's deliberation loop in `mind/deliberate/`,
TASK-0017/M6's dusk exchange and ambient pool in `mind/converse/`, and TASK-0018/M7's
nightly consolidation, death carry, and archival in `mind/consolidate/`) and the
Fabric mod vendor (`mod/`, TASK-0009/V1) —
[[body-protocol-seam]] cites both. This note's claims were re-verified against `mod/`'s,
`mind/memory/`'s, `mind/llm/`'s, `mind/prompt/`'s, `mind/persona/`'s,
`mind/deliberate/`'s, `mind/converse/`'s, and `mind/consolidate/`'s existence in this
pass (2026-08-28). `mind/converse/`, `mind/deliberate/`, and `mind/consolidate/` prove
against the fake vendor and scripted/mocked models only (`go vet` and `go test -race`
green in `mind/`); the real < 3 s first-token ceiling, the daemon's live wiring of V3's
pair-formation signal, and a live E6 digest call are unproven here by design — I2's
evening run is where that lands. The death-carry selector `mind/consolidate/` exports
is now consumable by `mind/converse/`'s context assembly — both merged; actual wiring is
daemon-integration work. The `mind/persona/` live-genesis proof
(specs/013-persona-genesis/live-run.md) succeeded in an earlier pass: three real E1 calls on
Opus 5 produced three validator-accepted personas, written 0444 at
`mind/run/persona/` and re-bound correctly on a simulated daemon restart. The mod's
Yarn-mapped brain-API surface ([[villager-brain-api]]) is flagged, not yet
re-derived, against MC 26.2's official mappings — that re-derivation is V3's
(TASK-0014) scoped work.

**Re-verified 2026-08-28 (TASK-0020 Phase 4, T012):** the job-board/blueprint-build paragraph
above is new; everything else re-confirmed still true against current code. TASK-0020's own
live dev-server pass (`specs/020-job-board/research/board-observation.md`) proved posting,
reading, and claiming live (the claim proof also reconfirmed persistence across a server
restart) but did NOT reach a live build placement — a token-namespace mismatch in the current
single-body live-attach wiring (`LiveBuildExecution`'s claimant lookup vs. `BodySession`'s
generic attach token) makes build structurally unreachable through that path today, unrelated
to schedule timing; the live half of build/interrupt/resume is deferred to I2 (TASK-0022).
