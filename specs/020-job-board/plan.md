# Implementation Plan: The job-board book and the blueprint build (V4)

**Spec dir**: `specs/020-job-board` · **Branch**: `task-0020-job-board`

**Constitution check**: `.specify/memory/constitution.md` is an unfilled template —
stated plainly per runbook gate. Planned against the grounding docs:
kithcraft-brief.md, body-protocol-v0.md (Q-6, AR-4), decision-0002,
demo-build-plan.md §3.3, `docs/wiki/` ([[body-protocol-seam]],
[[villager-brain-api]] — including the chunk-ticket trap and the schedule
substrate facts TASK-0014 derived).

## Where it lives

`mod/` (Java/Fabric): new package `dev.kithcraft.mod.board/` (lectern board,
postings, claims, read cycle) and `dev.kithcraft.mod.build/` (blueprint
interpretation, block placement, interrupt/resume). Consumes V2's percept
emitters (text/origin:read via the existing Testimony/read machinery), V3's
cast/schedule substrate, V5's tend-grave posting seam (GraveBoardEntry now
rides this board for real — closing that decoupling). Mind-side: NOTHING new —
M5's mind/deliberate is merged and its E3 loop is the claim driver in tests.
No wire/protocol changes (Q-6 is the whole point).

## Design decisions (settled surfaces restated)

1. **Board = lectern + written book.** Vanilla blocks, zero new items; the
   "board" is a designated lectern (config: position or first-placed
   convention). Claims render as appended lines in the board's readable
   content (the villager-visible surface is the text itself — same channel
   the blueprint rides, Q-6).
2. **Read cycle** rides V3's schedule substrate: a board-visit behavior in the
   existing Activity plumbing (no new Activity minting — Activity.MEET/WORK
   packages per villager-brain-api's addActivity-additive facts). If a new
   Activity value were needed it would need a Mixin — it is not; reuse.
3. **Claim registration**: the claim intent's verb rides the existing four-verb
   manifest surface... UNLESS the manifest floor doesn't cover it — then the
   manifest's declared-extras mechanism (V2 declares floor+5 extras, L-7
   holds) carries a claim verb. No protocol change either way: verbs are
   manifest-declared by design.
4. **Blueprint interpretation — the fidelity floor**: a tiny deterministic
   parser over free text extracting (shape, size, material) with generous
   defaults; anything unparseable still claims/declines cleanly. This is the
   deliberately-shed-first surface; `ponytail:` it thoroughly.
5. **Build execution**: block-by-block placement on a tick budget during work
   periods; material sourcing simplified to "from the village's storage
   things or creative-style provision" — pick the simplest defensible rule,
   record it. Interrupt = stop placing + persist cursor (SavedData, V1's
   registry idiom); resume = pick up cursor at next work period.
6. **No force-claim path** — structural absence test in V5's
   StructuralAbsenceTest style.
7. **No new Mixins expected** (board is blocks + behaviors + SavedData). If
   one becomes genuinely necessary: STOP (surface is at decision-0002's
   ceiling of 4).

## Risks / open items

- JOB_SITE-claim flakiness (TASK-0014 known gap) may affect work-period
  scheduling of builds; if hit, work around via the same tolerated-wander
  posture, record honestly.
- The live full-beat proof needs the dev server + stub mind; M5's loop drives
  claim logic in unit tests (Go side is merged; the Java side only sees
  intents), so "mind attached" in the card's Done-proves is satisfiable with
  the stub emitting an M5-shaped claim — record the split honestly.

## Phase map

Phase 1 — board, postings, read percept, claims visible (US1 + US2 read half).
Phase 2 — claim registration, no-force-path, tend-grave seam closure (US2).
Phase 3 — blueprint interpretation + build execution + interrupt/resume (US3+US4).
Phase 4 — live full-beat observation, gates, wiki re-ground, board close.
