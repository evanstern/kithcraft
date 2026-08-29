# Implementation Plan: Death, danger, and what remains (V5)

**Spec dir**: `specs/019-death-remains` · **Branch**: `task-0019-death-remains`

**Constitution check**: `.specify/memory/constitution.md` is an unfilled template —
stated plainly per runbook gate. Planned against the grounding docs:
death-mechanics.md, decision-0002 + entity-implementation-comparison.md,
body-protocol-v0.md, kithcraft-brief.md, `docs/wiki/` ([[villager-brain-api]],
[[body-protocol-seam]]), and specs/014's research/brain-26.2.md standard.

## Where it lives

`mod/` (Java/Fabric) — new package `dev.kithcraft.mod.death/` plus Mixins in
`dev.kithcraft.mod.mixin` and lines in `kithcraft.mixins.json`. Consumes V2's
percept/change_report machinery and V3's cast/token registry, both merged.
Vendor-side only; no mind or protocol changes. Disjoint from the three
in-flight mind-side branches (016/017/018) except the two shared hotspots the
runbook names: `kithcraft.mixins.json` (V3 landed its overrides; V5 adds after
— we are the later merger and take main's side on conflict) and the runbook log.

## The V4 decoupling (tend-grave entry, card AC #6)

V4 (TASK-0020, the job-board) is NOT merged and starts only after TASK-0016.
The card says the tend-grave entry "rides V4's existing mechanism"; the
runbook's lane graph nevertheless schedules V5 before/parallel to V4. Resolution
without re-litigating the lanes: the board is Q-6's read channel — a book/text
surface V2 already carries percepts for. V5 lands the tend-grave entry as a
posting through that same read channel (the text content a board would carry),
behind a small seam (`GraveBoardEntry` producing the posting text + a takeable
marker). When V4 merges its board book, the entry rides it with no rework — the
seam is the posting content, not the book block. Recorded as a deviation note on
AC #6 if the wording is judged to under-satisfy "rides V4's mechanism"; the
honest alternative (hold AC #6 open until V4 merges) is the fallback the
orchestrator can pick at merge time.

## Design decisions (settled surfaces restated)

1. **Verification first (US0/Phase 1)** — R-4 (POI re-claim lag after
   `releaseAllPois()`), R-5 (siege trigger location, suppressibility with one
   injection, 3-villager eligibility), evidenced at the brain-26.2.md standard
   (javap + decompiled-source citations), landing as research/death-26.2.md.
   The named escalation trigger binds: wrong suppression point or >1 targeted
   injection → STOP, surface to operator (runbook checkpoint 4).
2. **Suppress regardless of eligibility** — even if a 3-villager cast never
   qualifies, the Mixin lands (the card says so; eligibility is
   version-dependent and inconsistently documented).
3. **Conversion-cancel** — one Mixin making conversion terminal-equivalent:
   cancel the conversion outcome, route through the same death path (grave,
   bundle, token retirement, `body_continuous: false`).
4. **Grave placement** — death location if buildable, else nearest safe
   buildable surface (bounded search, deterministic tie-break); grave is a
   named thing with a NEW body token; belongings captured in the death event
   (before vanilla drop/destroy) into a storage-role thing named for the owner.
5. **Grief period** — config (default one day/night cycle ≈ 24000 ticks),
   holding bed + job-site POIs unclaimed; mechanism informed by R-4's finding
   (natural lag may carry part of it; the config is the guarantee).
6. **Percepts** — zero new machinery: witnesses see ordinary sightings (V2's
   emitters fire on what happens near them); the absent get §4.10
   change_reports (V2's restriction logic already gates delivery); the grave is
   an ordinary thing that gets sighted.
7. **Structural absence tests** — no feeding/escort/vigilance surface; no
   friendly-fire guardrail: grep-style structural tests in the pattern V2/V3
   used for their no-Mixin/no-salience claims.

## Risks / open items

- R-5 may fire the escalation trigger — that is a designed outcome, not a
  failure; Phase 1 ends the lane at a checkpoint if so.
- Live dev-server proofs need the Fabric env (present since TASK-0009) but not
  the API key (no mind involvement required — a stub mind suffices, per V2's
  Phase 4 pattern).

## Phase map

Phase 1 — R-4/R-5 verification, findings recorded; STOP/GO decision (US0).
Phase 2 — siege suppression + conversion-cancel Mixins; budget test (US1).
Phase 3 — grave, belongings, tend-grave posting, grief period, tokens (US2+US3).
Phase 4 — percept-channel proofs, dev-server observation, gates, wiki, close (US4).
