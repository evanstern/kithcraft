# Implementation Plan: Demo configuration and the two run targets (I1)

**Spec dir**: `specs/021-demo-config` · **Branch**: `task-0021-demo-config`

**Constitution check**: `.specify/memory/constitution.md` is an unfilled template —
stated plainly per runbook gate. Planned against the grounding docs:
decision-0001/0003, llm-routing-and-budget.md, death-mechanics.md §6.2,
demo-build-plan.md §3.4, `docs/wiki/` ([[overview]], [[body-protocol-seam]],
[[villager-brain-api]] — chunk-ticket trap binds the run doc's world setup).

## Where it lives

Mostly assembly, not new machinery — I1 is the task that finally WIRES what
the sweep built: `mind/cmd/minddaemon` gains its real runtime loop (persona
load/genesis at start, memory + ledger + archive wiring, session-end report) —
closing the daemon-wiring deferral TASK-0018 named for refactor-triage, here
because I1's card is exactly that wiring; `mod/` gains only config polish
(danger-tuning knob; grief knob exists) — no new Mixins. The documented
sequence lands as `docs/design/demo-runbook.md` (the operator-facing artifact
I2 will follow). The TASK-0020 token-namespace finding (findClaimant vs
BodySession tokens) is IN SCOPE to fix here if small — it is exactly
daemon/vendor integration wiring — else explicitly handed to I2 with the
observation cited.

## Design decisions (settled surfaces restated)

1. **Two artifacts** (decision-0001): `go build ./cmd/minddaemon` and
   `./gradlew build` jar; either start order works (vendor dials with retry —
   proven; daemon listens).
2. **Daemon runtime loop**: on start — load personas (genesis for missing,
   M3's Genesis + WriteOnce; ANTHROPIC_API_KEY + prefix override per M3's
   live-run precedent), open stores (M2 log, M7 ledger + archive), serve
   sessions (M1's listener), consult Archive on session_open (M7's hook),
   run RunNight on sleep triggers, emit the session-end report on shutdown.
   Wire the MINIMUM that makes I1's card true — deliberation/converse
   wiring beyond what the ACs need stays for I2's findings.
3. **Config**: daemon flags/env for socket path, run dir, genesis on/off;
   mod system-properties for grief (exists) + danger tuning (new, off by
   default — the R-6 knob scales hostile spawn pressure near the cast or
   equivalent smallest lever; document what it does and that using it is a
   per-run choice).
4. **R-1** is a recorded ruling in demo-runbook.md, not code.
5. **Session-end report**: M4 Accounting snapshot + M2 instrument dump,
   printed on daemon shutdown and appended to a run log file.
6. **Restart independence live proof**: scripted in the observation — boot,
   admit memories via stub activity, kill daemon, restart, verify re-bind +
   memory survival + gap-reported.

## Risks / open items

- The daemon wiring is the largest single integration step so far; keep it
  assembly of tested parts (every component has its own green suite) — new
  logic only where glue demands it.
- Live genesis needs the operator's key; the rehearsal path must be zero-call
  (personas pre-seeded from M3's live-run artifacts or stub personas).

## Phase map

Phase 1 — daemon runtime loop: stores + personas + listener + archive hook (US1 partial, US2 machinery).
Phase 2 — config surface: knobs, flags, R-1 ruling text; session-end report (US3 + US4).
Phase 3 — demo-runbook.md + live bring-up observation incl. restart independence (US1 + US2 live).
Phase 4 — gates, wiki re-ground, board close.
