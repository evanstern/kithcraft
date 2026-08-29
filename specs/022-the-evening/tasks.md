# Tasks: The evening (I2)

**Spec dir**: `specs/022-the-evening` · **Branch**: `task-0022-the-evening`

## Phase 1 — Stage the run kit (sweep)

- [x] T001 run-kit.md: the beat checklist (plan §4's seven beats, each with
      what-to-look-for + capture evidence), the spell-breaker checklist
      (§5.2's three, with per-task checks and the walking-past test), and
      the A-n measurement sheet (A-1..A-8 → capture source → threshold)
      — `specs/022-the-evening/run-kit.md` §1–§3.
- [x] T002 Prerequisites verified: build green at this branch; demo-runbook.md
      current (follow-read, not re-run); the operator-facing needs named
      (key for genesis unless re-bind, ~3h with player, recording setup,
      dangerTuning available for the take)
      — `go build ./cmd/minddaemon` and `./gradlew build` both green this
      session (no code changed); demo-runbook.md's flags/env vars/system
      properties cross-checked live against `mind/cmd/minddaemon/{main,
      config}.go` and `mod/.../death/{DangerTuning,GriefPeriod}.java` — no
      drift, no fix needed; run-kit.md §0 names the operator-facing needs
      plus the M5/M6 live-wiring decision point found while verifying.
- [x] T003 Watch list compiled from merged tasks' honest flags (0014 signal
      lead + JOB_SITE; 0020 build/WORK timing; 0021 reconnect identity +
      heartbeat admissibility), each with where-to-look during the run
      — run-kit.md §4, items 2–5; plus two additional items (#1 M5/M6 never
      wired into `mind/cmd/minddaemon/runtime.go`, #6 E4 latency has no
      capture path) found by reading the live code while staging this kit.

## Phase 2 — The evening (OPERATOR — runbook checkpoint 5)

- [ ] T004 The ~3-hour run at 1x, player present, kit in hand; logs +
      session-report.log + recording captured (card ACs #1, #2, #3, #6 raw)

## Phase 3 — Write-up and close (sweep, post-run)

- [ ] T005 docs/design/evening-findings.md: beats walked, spell-breakers
      walked, A-n annotated with measured values, best dusk named, every
      failure owned by a task (card ACs #2–#6 resolved)
- [ ] T006 Failures carded: new tasks (or refactor-triage handoff) for any
      finding without an owner
- [ ] T007 Gates green; wiki re-ground (v1-demo — the demo now ran; overview);
      CAPSULES if descriptions changed; freshness green
- [ ] T008 Card ACs ticked with citing proofs; board/spec synced at PR time
