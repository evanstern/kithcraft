# TASK-0007..0022 m-0 build sweep — "one real evening" — sweep runbook (2026-08-21)

**You (the session reading this) are the ORCHESTRATOR** for the tasks below. Run each
through the host project's full PDLC — spec → link → worktree → delegated implementation →
PR → merge → re-ground — parallelizing within lanes, merging serially, treating merge
conflicts as routine. Direction is decided; do not re-litigate it:
`docs/design/demo-build-plan.md` (signed off 2026-08-21, PR #10) is the plan of record
these sixteen tasks were created from, and it consumes five settled surfaces that win over
any implementer instinct: `docs/design/kithcraft-brief.md` (ratified 2026-08-19),
decision-0001 (Fabric), `docs/design/body-protocol-v0.md` (accepted, PR #6),
decision-0002 + `entity-implementation-comparison.md` (augmented vanilla villager),
decision-0003 + `llm-routing-and-budget.md` (Go daemon, six classes / three tiers, real
wires only). Plan-of-record is the board; this file carries only ordering, doctrine, and
the log.

**Status:** signed-off · operator sign-off on lanes: 2026-08-21
<!-- Sign-off ruling (2026-08-21): lanes approved as authored; the proposed opus
     escalation for TASK-0007/S1 TAKEN by the operator — it runs opus
     (cc/claude-opus-5[1m], fallback cc/claude-opus-4-8[1m]). TASK-0019's trigger
     stays conditional. The escalation checkpoint the tier rubric requires is this
     ruling; transport ratification (checkpoint 2) remains the operator's at PR. -->
<!-- Only the OPERATOR flips draft → signed-off (the author never pre-fills it). An
     executing session must refuse a runbook whose status it cannot verify. -->
<!-- This draft rides the TASK-0007 branch per the operator's 2026-08-21 no-runbook-PR
     ruling: runbook commits ride task branches / the board track, never their own PR.
     The TASK-0007 claim is the next commit after sign-off. -->

## Read first (in this order)

1. `docs/design/demo-build-plan.md` — the decomposition these tasks came from; §2's nine
   rulings (R-1..R-9) are settled, §8.2 is the lane source, §5 the spell-breaker map.
2. `docs/wiki/CAPSULES.md` — corpus rollup; then just-in-time full notes:
   [[body-protocol-seam]] for S/M/V seam surface work, [[promptworld-lineage]] for M2/M3/M5/M7
   (what ports, what is reimplemented), [[villager-brain-api]] for V3/V5,
   [[model-tiers]] before any dispatch, [[root-guard]] before any commit lands anywhere.
3. Project `CLAUDE.md` — PDLC grounding, model-tier posture, task conventions.
4. `backlog task list --plain` — live state; other sessions move it while you work.
5. The task you're about to execute (`backlog task view TASK-<n> --plain`).

## State when this runbook was written (2026-08-21)

- **Done already:** TASK-0001..0006 (all decision/design phase; last merge PR #11,
  sweep-0004-0006 close). Wiki 11 notes, freshness probe green at 8fd2a89. Specs 001–006
  taken; **next free spec number is 007**. No code exists yet — this sweep writes the
  project's first Go and Java.
- **In flight in other sessions (do not duplicate; expect their merges):** none known.
- **Paused — untouched (`paused` label in the task's frontmatter `labels:`; excluded
  from lane conflict analysis; never claim, rebase, or clean their
  branches/worktrees):** none.
- **Queued (this runbook's scope):** TASK-0007 (lane 0); TASK-0008, TASK-0009 (lane 1);
  TASK-0010, TASK-0011, TASK-0012 (lane 2); TASK-0013, TASK-0014, TASK-0015 (lane 3);
  TASK-0016, TASK-0017, TASK-0018, TASK-0019 parallel then TASK-0020 (lane 4);
  TASK-0021 then TASK-0022 (lane 5, serial).
- `tiers.mjs --root . --check` exited 0 on 2026-08-21 before these lanes were authored
  (all three agent definitions `unchanged`; no regeneration, so no session restart
  required before first dispatch). Merge-drift gate: **absent** (no `scripts/` dir).

## Execution lanes (dependency-ordered; parallelize within a lane)

Rule of thumb: DEVELOP in parallel, MERGE serially. The seam is a genuine parallelism
gift: the mind lane (Go, its own module) and the vendor lane (Java/Gradle) touch disjoint
directories in different languages — within a lane they are close to conflict-free. The
shared hotspots are the board dir, the wiki, and this file (see doctrine). Lanes are the
plan's §8.2, adopted 1:1; the critical path is S1 → M1 → M2 → M5 → V4 → I1 → I2, seven
merges deep.

**Lane 0 — the wire (alone, blocks everything):**

- **TASK-0007 / S1 (PROPOSED `opus` · model `cc/claude-opus-5[1m]`, fallback
  `cc/claude-opus-4-8[1m]` — escalation proposed by the plan and by the card: the
  transport choice is a decision the spec constrains (real wires only, T-1..T-7) but
  does not settle — design work per the rubric, directly analogous to the TASK-0002 and
  TASK-0004 escalations. Taking it is the OPERATOR's ruling at sign-off of these lanes;
  if declined it runs `sonnet` `cc/claude-sonnet-5[1m]`. Ratifying the resulting
  decision record is an operator checkpoint regardless of tier)** — decide UDS/TCP/stdio
  against T-1..T-7; framing spec as the spec-002 successor; golden vectors (one per
  percept type, per intent shape, plus `session_open`); trivial Go encoder + Java
  decoder round-trip every vector. **Contract-shaped: both skeletons encode against its
  vectors; nothing else starts until it merges.**

**Lane 1 — skeletons (2 parallel, one per side; after TASK-0007 merges):**

- **TASK-0008 / M1 (`sonnet` · model `cc/claude-sonnet-5[1m]` — default tier: greenfield
  against the written contract; boundary decode, V-1..V-6, and session lifecycle are all
  specified in body-protocol-v0 and decision-0003)** — Go daemon skeleton: own module,
  vendor port declared at the consumer, decode→validate→mutate, session lifecycle,
  percept ingest, intent bookkeeping, minimal in-test double.
- **TASK-0009 / V1 (`sonnet` · model `cc/claude-sonnet-5[1m]` — default tier: mod
  skeleton to the written handshake contract; the one open verification (Yarn mappings
  at target version) is verification, not design)** — Fabric mod skeleton, transport
  client, `session_open` manifest (L-7: identical for every body), token registry
  persisted across restarts.

**Lane 2 — core surfaces (3 parallel; M-tasks after TASK-0008, V-task after TASK-0009):**

- **TASK-0010 / M2 (`sonnet` · `cc/claude-sonnet-5[1m]` — default tier: reimplement
  event-sourced memory to a written contract (RM-1..RM-7, admission gate per routing
  §6.3); judgment calls already made by decision-0003)** — memory, belief store,
  admission gate, E6-input-tokens instrument.
- **TASK-0011 / M4 (`sonnet` · `cc/claude-sonnet-5[1m]` — default tier: model client
  against RT-1..RT-7; the stable/variable prompt split is designed in routing §2.3)** —
  anthropic-sdk-go client, per-class prompt assembly, tier routing, instrumentation.
- **TASK-0012 / V2 (`sonnet` · `cc/claude-sonnet-5[1m]` — default tier: percept/intent
  conformance to the protocol's written surface; R-8 (sound hook) is verify-then-declare,
  not design)** — percepts out, intents in, provenance stamping, change_report delivery
  restriction, four verbs, §12 leak passes.

**Lane 3 — cast, persona, proofs (3 parallel):**

- **TASK-0013 / M3 (`sonnet` · `cc/claude-sonnet-5[1m]` — default tier: persona port is
  decision-0003's "cleanest carry"; the design is settled, this is the port plus the
  genesis prompt)** — E1 genesis on Opus 5 (the *product's* model choice, not the
  implementer's tier), 0444 firewall, model-free validator. After TASK-0008 + TASK-0011.
- **TASK-0014 / V3 (`sonnet` · `cc/claude-sonnet-5[1m]` — default tier: vanilla
  `Brain<E>` substrate work mapped in villager-brain-api; Mixin surface enumerated and
  bounded by decision-0002)** — schedules, cast, dusk pair formation with the ~10 s-ahead
  signal (R-7), body-keeps-moving rule. After TASK-0009.
- **TASK-0015 / S2 (`sonnet` · `cc/claude-sonnet-5[1m]` — default tier: protocol §10
  specifies FakeVendor and H-1..H-6 in detail; execution against a written standard)** —
  fake vendor + protocol-rule harness; each test red when its rule is lifted. After
  TASK-0008 + TASK-0010.

**Lane 4 — the beats (parallel across sides; TASK-0020 serializes behind TASK-0016):**

- **TASK-0016 / M5 (`sonnet` · `cc/claude-sonnet-5[1m]` — default tier: toolloop shape
  ports one-to-one per decision-0003; the urgency interrupt and K=10 window are written
  in routing §5.5/§6)** — deliberation E2/E3, job-board decision, urgency interrupt.
- **TASK-0017 / M6 (`sonnet` · `cc/claude-sonnet-5[1m]` — default tier: E4/E5 params are
  fully specified (Sonnet 5, <3 s, thinking off; Haiku pool); the latency posture is
  design already done)** — dusk conversation + ambient pool, pre-generation off V3's
  signal.
- **TASK-0018 / M7 (`sonnet` · `cc/claude-sonnet-5[1m]` — default tier: consolidation
  ports the measured machinery shape (ordinal convention, no-marker-on-failure); R-9 is
  ruled)** — nightly consolidation E6, death carry, archived-not-terminated minds.
- **TASK-0019 / V5 (`sonnet` · `cc/claude-sonnet-5[1m]` — default tier **with the
  plan's named escalation trigger**: R-4/R-5 verification comes first; if the siege
  suppression point is not where death §1 assumes, or suppression needs more than a
  targeted injection, STOP — that is an operator checkpoint and a possible opus
  escalation, never a silent Mixin-surface growth past decision-0002's bound)** — death,
  danger, graves, belongings, grief period, token discipline. After TASK-0012 +
  TASK-0014.
- **TASK-0020 / V4 (`sonnet` · `cc/claude-sonnet-5[1m]` — default tier: the board rides
  Q-6's read channel with no protocol extension; build execution is deliberately the
  thinnest possible system)** — job-board book + blueprint build, kept whole. **The
  graph's one cross-lane dependency: starts after TASK-0016 merges** (needs real claim
  behaviour, not a stub), plus TASK-0012 + TASK-0014.

**Lane 5 — integration (serial):**

- **TASK-0021 / I1 (`sonnet` · `cc/claude-sonnet-5[1m]` — default tier: config plumbing
  and documented startup; every knob is ruled (R-1, R-3, R-6))** — demo config, two run
  targets, cast seeding, mind-restart independence. After TASK-0013, TASK-0018,
  TASK-0014, TASK-0019.
- **TASK-0022 / I2 (`sonnet` · `cc/claude-sonnet-5[1m]` — default tier: run the
  checklist, measure the A-n numbers; judgment about what the findings mean returns to
  the operator)** — the evening itself. **Requires the operator at the keyboard** (~3 h
  real-time with the player present) — see checkpoints. After everything.

Tiers and their model IDs come from **`.claude/model-tiers.json`** — `tiers.mjs --root .
--check` exited 0 on 2026-08-21 (see state snapshot). `haiku` fits nothing in this sweep
(plan §1: no siblings exist yet to write to, and its 200K window under-holds the protocol
doc). Dispatches name the generated agents `sonnet-implementer` / `opus-implementer` —
the frontmatter `model:` pin is what this harness honors — and pass the model ID on the
call too. **Verify the served model from the first dispatch's transcript before launching
sibling dispatches** — one wrong pin is a rounding error; a lane of them is the budget.

Record the model tier + explicit model ID + rubric justification on each board task at
dispatch, including which model actually served (one-way escalation only; escalations are
operator checkpoints).

## Per-PR gates this project enforces (enumerated — implementers cannot miss these)

- **Merge-drift gate: absent.** No `scripts/check-merge-drift.mjs` in this repo; raw git
  discipline stands: `git fetch origin && git pull --ff-only` at root before
  claim/worktree/merge; verify merged (`gh api repos/{owner}/{repo}/pulls/{n} --jq
  .merged`) before deleting branches/worktrees; check `specs/` on fresh `origin/main`
  for number collisions before claiming (next free at authoring time: **007**; the spec
  number for TASK-00NN is NN by convention — 007..022 — renumber on collision).
- **Root-guard hook** (`.claude/hooks/root-guard-hook.mjs`, PreToolUse on Bash and
  Write/Edit): the root checkout is read-only — all work rides worktree branches under
  `.worktrees/`; this sweep runs in **no-main-push mode**: post-merge closures (tasks.md
  ticks, spec-bridge:sync, runbook log rows) ride the NEXT claimed task's branch;
  sweep-close rides a wrap-up PR. The one exception: board-sync commits scoped entirely
  to `backlog/` may commit at root via Bash (the CLI is the sanctioned editor).
- **No runbook PRs (operator ruling 2026-08-21):** runbook commits — draft, sign-off,
  log rows — ride task branches or the wrap-up PR, never a dedicated PR.
- **Rebases are forbidden repo-wide** (root-guard doctrine) — every reconcile is a merge
  of `origin/main` into the branch; **PRs land as merge commits, never squash** (branches
  carry wiki re-pins; squash rewrites the hashes those pins reference).
- **Wiki freshness probe after every history move, unconditionally:** for each note in
  `docs/wiki/`, `git log <verified_against>..HEAD -- <sources>` must be empty or the
  note re-verified. Honest re-pins only — classify every staled note against
  `git diff <old-pin>..<merge-commit> -- <sources>` as RE-PIN-ONLY or NEEDS-REVIEW
  (amend prose BEFORE bumping); the merge commit is the re-pin target, never the
  justification. Pins reference `docs/design/*` sources outside the wiki — a
  wiki-untouched diff can still be stale.
- **CAPSULES.md is derived state:** regenerate via grounding-wiki `capsules.mjs`
  whenever any note's `description:` changes; never hand-edit.
- **Evidence rule:** any claim about an external dependency's version, maintenance,
  license, or pricing carries a URL + accessed date (established TASK-0001; binds V1's
  Yarn/Fabric version re-verification and M4's SDK/pricing claims).
- **Board via CLI only,** from a checkout the CLI can write (the task worktree in this
  mode, or root for `backlog/`-scoped board-sync commits); add specific task files to
  git, never `backlog/` wholesale.
- **Code gates (new this sweep — first code in the repo):** Go work ships `go vet` +
  `go test ./...` green in its module; Java work ships `gradle build` (and `gradle test`
  where tests exist) green. Every non-trivial rule the cards name lands with its test —
  the card ACs are written as checkable proofs; the PR that cannot demonstrate its
  "Done proves" line is not ready. Dev-server proofs (V-tasks) record how they were
  verified in the PR description (test output, or a documented `gradle runServer`
  observation when no automated harness exists yet).

## Per-task artifacts required before PR

Per-TASK obligations. **No PR opens for a task until each line below checks true for
it.** The sweep's Output gate re-checks the first two lines for every scoped task at the
end.

- [ ] `specs/0NN-<slug>/` carries a real `spec.md` (problem + requirements mapped to the
      card's ACs), `plan.md` (the constitution at `.specify/memory/constitution.md` is
      an unfilled template — plan.md must state that plainly and plan against the
      grounding docs: the brief, the five settled surfaces, CLAUDE.md, `docs/wiki/`),
      and `tasks.md` (phased checkboxes the bridge derives from), committed on the
      task's branch. A claim stub reserves the number; it satisfies nothing here.
- [ ] The card carries its Spec marker from the claim commit (`spec-bridge:link` against
      the stub), and phase ACs are seeded from tasks.md (link update mode) before
      implementation dispatch.
- **Escape lines (operator-signed only):** none.
- **Host additions:**
  - TASK-0007: the transport decision lands as a Backlog decision record
    (`backlog decision create …`) in the same PR; the framing/serialization spec lands
    under `docs/design/` as the spec-002 successor; the golden vectors land as
    language-neutral fixture files both implementations can reach. Operator ratification
    of the decision follows PR review, before sync moves the card Done.
  - TASK-0009: the Yarn-mappings / target-version re-verification is recorded in the PR
    (evidence rule applies to version claims). **Operator ruling 2026-08-22:** V1
    introduces Gradle — in that same PR, replace `seam/java-roundtrip`'s hand-rolled
    JSON *parsing* with a library-based harness (the hand-roll existed only to avoid
    a build system this repo didn't have yet; once Gradle is native, the avoidance is
    pointless). The canonical *writer* may remain custom only if the chosen library
    provably cannot emit C-1..C-10 form — verified against the vectors and recorded
    either way.
  - TASK-0012: R-8's hearing-hook verification is recorded with its outcome (declared
    or unsupported) — the card's AC #3.
  - TASK-0019: R-4/R-5 verification findings are recorded BEFORE implementation
    commits; if the escalation trigger fires, stop and surface (checkpoint 4).
  - TASK-0022: the run's findings doc (beats walked, spell-breakers walked, A-n
    measured values, failures with owning task) lands under `docs/design/` in the PR.
  - All: wiki notes whose sources the PR touches are re-verified in that PR (board DoD:
    "Docs and wiki are updated and pass freshness tests"). New code directories that
    become load-bearing get wiki coverage as the corpus grows — at minimum the notes
    whose prose the PR invalidates are amended.

## Concurrency & conflict doctrine

- **Hotspots:** `backlog/tasks/*` (all lanes flip cards and tick ACs);
  `docs/design/sweep-0007-0022-runbook.md` (all lanes log rows — later merge takes
  main's side and re-adds its row); `docs/wiki/CAPSULES.md` (regenerate after merge if
  any description changed — never hand-merge); the **Go module root** (`go.mod`, shared
  packages — M2/M4 both sit on M1's skeleton and must touch different packages; M5/M6/M7
  likewise; if two mind tasks need the same file, the earlier merger wins and the later
  merges main in); the **Java mod source tree** (V2/V3 both extend V1 — same rule; V5's
  Mixin config file is shared with V3's — coordinate: V3 lands its three task-list
  overrides first, V5 adds conversion-cancel after); S1's **golden vector fixtures**
  (M1 and V1 both consume, neither edits — a vector change after S1 merges is a contract
  change and an operator checkpoint).
- **Paused tasks are not live lanes:** a task labeled `paused` (set/cleared only via
  `backlog task edit --labels`, provenance in its append-notes) is listed in the state
  snapshot and NEVER claimed, rebased, or cleaned.
- Reconcile: **merge `origin/main` into the branch** (repo-wide rebase ban); take main's
  side for anything not deliberately changed. Treat every branch as pin-carrying once
  its wiki re-verifications land — merge-commit PRs only.
- After every history move: re-run gates AND the freshness probe unconditionally.
- Within a lane, merge order is free — prefer the **vendor** side of a pair first when
  both are ready (the mind side is exercised against the fake vendor / in-test double
  and does not block on the real one); otherwise smallest-first. Two hotspot-heavy PRs
  never merge within one re-ground cycle without a reconcile between.
- **Claim before work:** the FIRST commit of each branch claims it — board card →
  In Progress, spec-number stub dir, spec-bridge link — pushed immediately
  (`git push -u origin <branch>`); never force-push a claim.
- **A rejected push means you lost the race:** fetch, re-read board and `specs/`.
  Another holder → STOP the lane, surface to operator. Unrelated rejection with
  task+number still free → merge `origin/main` in and re-push (plain push).
- Verify a PR is merged (`gh api … --jq .merged`) before deleting its branch/worktree;
  never delete+recreate a closed PR's head.
- **Dispatch phase-scoped:** one fresh implementer per tasks.md phase (grouping small
  adjacent phases is the orchestrator's recorded call); handoff between phases is the
  spec dir + tasks.md tick-state + the branch's commits, never chat context. Every
  dispatch prompt carries the turn-hygiene block (batch independent reads, minimal
  narration, lower effort on mechanical phases). At each dispatch boundary, update the
  task's in-flight log row. At a lane boundary, the orchestrator SHOULD end its session
  and resume from this file + the board (context cost is monotonic; the tail is the
  expensive part).

## Operator checkpoints (do not proceed silently)

1. **Sign-off on these lanes** — pending. Includes ruling on the one PROPOSED opus
   escalation (TASK-0007 / S1). TASK-0019's trigger is conditional, not a proposal to
   rule now.
2. **TASK-0007 transport ratification:** the operator ratifies the decision record and
   framing spec. Surface the PR and stop before merging — every later task encodes
   against it.
3. **Environment prerequisites (operator provides, before the lanes that need them):**
   a working Fabric dev environment (JDK + Gradle + target Minecraft version) for
   V-task dev-server proofs from TASK-0009 on; `ANTHROPIC_API_KEY` (and billing
   awareness — routing's baseline is ≈$5.17/evening, ceiling ~$20) for live-call proofs
   from TASK-0013/0017/0018 on (M4's unit tests should mock; genesis/latency proofs are
   real calls). Surface and pause the affected lane if either is missing when reached.
4. **TASK-0019 escalation trigger:** if R-4/R-5 verification finds the siege
   suppression point elsewhere than death §1 assumes, or suppression needs more than a
   targeted injection — stop the lane, surface findings, operator rules on escalation.
5. **TASK-0022 is operator-run:** the evening needs the player present ~3 hours at 1x.
   The sweep prepares (I1 merged, checklists staged) and stops; the operator schedules
   the run; the implementer writes up findings from the recorded run afterward.
6. Tier escalations; lane amendments (amend this file, note why, tell the operator).

## Done means

TASK-0007..0022 all Done on the board via one merged PR each (sixteen PRs); every card
still carries its Spec marker; `specs/007-*/` through `specs/022-*/` each carry real
spec.md + plan.md + tasks.md; the transport decision record exists and is
operator-ratified; TASK-0022's findings doc exists with every A-n assumption annotated
by measurement; Go tests and Gradle builds green on main; wiki freshness probe green
over all notes (honest re-pins only); no stale sweep worktrees; this file's log complete
and status flipped to done (via wrap-up PR).

## Execution log

Multi-phase dispatch stays visible in `notes` — one slot, never a second table: while a
task is in flight its row carries the phases dispatched/completed (e.g.
`phases: 1-2 done, 3 dispatched`), updated at each dispatch boundary, so a resuming
session can see where within the task the last one stopped; the closing note on merge
replaces or absorbs it. `tokens/cost` carries best-effort actuals from the
harness/transcript, so future runbook authoring budgets against real numbers.
(Prior-sweep actuals for budgeting: a 3-phase opus task ran ~397–571k subagent tokens; a
3-phase sonnet task ~350–411k. Sixteen tasks at that rate is plausibly ~6M subagent
tokens; the lane-boundary session-restart prescription above is the cost control.)

| date | task | PR | merge | tokens/cost (best-effort) | notes |
|------|------|----|-------|---------------------------|------|
| 2026-08-21/22 | TASK-0007 | #12 | 0262c8e | ~614k subagent (opus x4) + orchestrator | claimed (a9cd757), spec cycle committed (9b21361), phase ACs seeded (c03089d). phases: 1 done (6003775, opus VERIFIED serving claude-opus-5, ~139k; UDS chosen — AF_UNIX SOCK_STREAM, mind listens / vendor dials; decision-0004 proposed; SOCK_SEQPACKET host-verified unavailable), 2 done (b023927, opus verified, ~142k; seam-wire-v0.md — length-prefix + canonical JSON, one conn per vendor, byte-exact equality, seq-at-enqueue makes shedding observable), 3 done (d380acf, opus verified, ~162k; 17 vectors, Go + Java harnesses green — orchestrator re-ran both independently; mutation-checked), 4 done (10c6e02 + pin fix 4f881b6, opus verified, ~171k; six-pass leak sweep clean over spec + vectors, FR-007 scope clean at 38 files, body-protocol-seam re-verified honestly — Q-1 closed-but-proposed, CAPSULES regenerated; orchestrator caught and fixed a non-resolving pin sha the agent recorded). PR #12 MERGED 0262c8e (merge commit, pin preserved; operator ratified decision-0004 by merging — checkpoint 2 satisfied; decision flipped accepted 42c1f82, card synced Done d39dada, both board-sync at root); worktree+branch removed after merge verified. ~614k subagent tokens across 4 opus dispatches + orchestrator. Operator vector-comprehension review held pre-merge (frame_hex decoded live, redundancy-by-design confirmed); Gradle-replaces-hand-rolled-Java-harness ruling recorded on TASK-0009 + gates section (38bfa36). |
| 2026-08-22 | TASK-0008 | — | — | — | in flight: claimed (3b8521c), spec cycle (f845f70), phase ACs seeded (a912757). phases: 1 dispatched (sonnet VERIFIED serving claude-sonnet-5 from transcript). |
| 2026-08-22 | TASK-0009 | — | — | — | in flight: claimed (2249ac6), spec cycle (ed22675), phase ACs seeded (cffe269). phases: 1 dispatched (sonnet; sibling launched after TASK-0008's dispatch verified the tier pin). |
