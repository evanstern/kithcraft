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

<!-- Merge-policy ruling (2026-08-25, operator): the SWEEP SELF-MERGES its PRs after
     gates pass (merge commits, serial), including PR #13 forward. Operator review
     happens only at the named checkpoints below (decision ratifications, escalations,
     environment prerequisites) — not per-PR. Ruling recorded when the permission
     classifier flagged the first self-merge (PR #13); the operator chose
     sweep-self-merge over per-PR review. -->

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
| 2026-08-22 | TASK-0008 | #13 | dd08df6 | ~604k subagent (sonnet x4) + orchestrator | claimed (3b8521c), spec cycle (f845f70), phase ACs seeded (a912757). phases: 1 done (72b3c09, sonnet VERIFIED, ~138k; mind/ module + wire codec, 17/17 vectors, orchestrator re-ran green), 2 done (6b2ee3d, sonnet verified, ~179k; port + UDS listener + sessions + no-backfill continuity test; §1.3 last-open-wins ponytailed), 3 done (4f6e819, sonnet verified, ~186k; ingest + intents + double + both e2e ACs; seq accounting widened to all counter-sharing messages), 4 done (f51b933 + re-pins 47c9363/11b5caa, sonnet verified, ~101k; gates green, scope clean, body-protocol-seam + overview + INDEX amended — seam no longer contract-only, no-code claims fixed; CAPSULES regenerated; orchestrator additionally caught promptworld-lineage staled by decision-0003's flip — earlier probes missed it via an unquoted-path bug — amended + re-pinned 032ca6b). PR #13 MERGED dd08df6 (merge commit, pins preserved; self-merge per the 2026-08-25 operator merge-policy ruling); card synced Done (ff5c92d, board-sync at root); worktree+branch removed after merge verified. ~604k subagent tokens across 4 sonnet dispatches. |
| 2026-08-22/25 | TASK-0009 | #14 | 7ddb3d3 | ~730k subagent (sonnet x4) + orchestrator | claimed (2249ac6), spec cycle (ed22675), phase ACs seeded (cffe269). phases: 1 done (fa30968, sonnet VERIFIED, ~150k; MC 26.2 / Loader 0.19.3 / API 0.158.0 / Loom 1.17.19 cited; build + runServer green. FINDING: Yarn discontinued, MC 26.1+ unobfuscated — noted on TASK-0014), 2 done (fb48d1c, sonnet verified, ~209k; Gson parse + custom canonical writer per ruling carve-out; 26/26 tests incl. 17-vector suite + stub-mind handshake), 3 done (62f375a, sonnet verified, ~190k; L-7 static manifest byte-identity test, SavedData-backed token registry, live restart persistence observation), 4 done (378fe18, sonnet verified, ~181k; harness replaced per operator ruling — Gson parse + custom writer, seamRoundTrip Gradle task, mutation-checked red both ways; villager-brain-api marked UNVERIFIED at 26.2 for V3, body-protocol-seam + overview + INDEX amended, CAPSULES regenerated, pins resolvable). PR #14 MERGED 7ddb3d3 (merge commit, pins preserved; self-merge per ruling); card synced Done (4aec551, board-sync at root); worktree+branch removed after merge verified. ~730k subagent tokens across 4 sonnet dispatches. NOTE for future dispatches: the backlog MCP tools resolve against ROOT's backlog/ — unsafe from a worktree; use the backlog CLI in-worktree. |
| 2026-08-25 | TASK-0010 | #15 | c572509 | ~594k subagent (sonnet x4) + orchestrator | in flight (lane 2): claimed (ccf3d8e), spec cycle (652706a), phase ACs seeded (7f4e0c2). phases: 1 done (dd7d929, sonnet VERIFIED serving claude-sonnet-5, ~125k; append-only log + type-level immutability via external-package API-surface test, (world_time,hash) identity; orchestrator re-ran vet+test green), 2 done (46cad59, sonnet verified, ~152k; provenance classifier + RM-1..RM-7 belief store, external-package AC #2/#7 surface tests; orchestrator re-ran vet+test green), 3 done (9d41210, sonnet verified, ~181k; §6.3 gate + instrument + §10.2 e2e proving told-cannot-become-witnessed; rule-label order + involves-other-body extraction judgment calls flagged; e2e proves the epistemic core directly — seam has no memory wiring yet, lands in M5; orchestrator re-ran vet+test green), 4 done (2ef523b, sonnet verified, ~136k; gates green, scope clean, body-protocol-seam + overview amended + re-pinned honestly, CAPSULES regenerated; ACs #1-#4,#6,#7 ticked with citing tests; AC #5 deliberately left for S2 — fake-vendor wording — completes at TASK-0015/M5). PR #15 MERGED c572509 (merge commit, pins preserved; self-merge per the 2026-08-25 operator merge-policy ruling); card synced Done via spec-bridge derived plan (e26b620, board-sync at root); worktree+branch removed after merge verified. ~594k subagent tokens across 4 sonnet dispatches. |
| 2026-08-25 | TASK-0011 | #16 | 39c75ac | ~613k subagent (sonnet x4) + orchestrator | in flight (lane 2): claimed (d6c3876), spec cycle (17a8d29), phase ACs seeded (a974b5c). phases: 1 done (e0545e0, sonnet verified cc/claude-sonnet-5[1m], ~141k; E1..E6 registry + prompt assembly, AC #3 byte-identity + red variant, prompt imports nothing from llm; four under-specified class params filled by one documented rule, cited inline in classes.go; orchestrator re-ran vet+test green), 2 done (9833e06, sonnet verified, ~190k; anthropic-sdk-go v1.58.0 (accessed 2026-08-25, cited in client.go), Send/Stream + cancel/retry/breakpoint mocked tests, structured Intent/Digest reusing seam.Pending shape, Digest ponytailed for M7; adaptive thinking chosen over budgeted — E2/E3 1024 ceiling cannot satisfy budget_tokens>=1024<max_tokens, doc says adaptive by name; orchestrator re-ran vet+test green), 3 done (9f96c7e, sonnet verified, ~136k; Accounting + AccountedStream wired into Client, cancelled-partials counted, six-class scripted session test; orchestrator re-ran vet+test green), 4 done (102b88c, sonnet verified, ~146k; merged origin/main in post-TASK-0010 (b899d17), gates green fresh, scope clean incl. sanctioned SDK dep, overview amended + re-pinned NEEDS-REVIEW-class, ACs #1-#11 ticked with citing tests; pre-existing capsule/body size-budget failures on body-protocol-seam + promptworld-lineage confirmed to predate this branch on main — corpus debt flagged for refactor-triage, not fixed here). PR #16 MERGED 39c75ac (merge commit, pins preserved; operator merged after classifier denied self-merge); card synced Done via spec-bridge derived plan (b48f07e, board-sync at root); worktree+branch removed after merge verified. ~613k subagent tokens across 4 sonnet dispatches. |
| 2026-08-25 | TASK-0012 | #17 | 81c2b75 | ~883k subagent (sonnet x4) + orchestrator | closed (lane 2): claimed (291c64a), spec cycle (55269a2), phase ACs seeded (60ef8e7). phases: 1 done (eeb01f6, sonnet verified cc/claude-sonnet-5[1m], ~169k; Provenance/PerceptEmitter/Sightings/SelfState + recursive AR/no-salience scanner, 59 tests green, zero Mixins; nearest_hostile surfaced as ordinary k:hostile sighting — interpretation noted; seq-counter/Handshake wiring ponytailed to Phase 3/4; orchestrator re-ran gradle test green), 2 done (bd533b8, sonnet verified, ~156k; R-8 VERIFIED — MC 26.2 GameEvent/GameEventListener native hearing hook, javap-cited, no Mixin; Sounds/Testimony/ChangeReports + §4.10 restriction test, manifest declares floor+5 extras, L-7 holds, 71 tests; live listener registration ponytailed to Phase 3/4; orchestrator re-ran gradle test green), 3 done (56a5048, sonnet verified, ~177k; IntentHandler/TargetResolution/Verbs, 88 tests; Brain facts re-verified vs 26.2 jar — Yarn-name renames + getSchedule/setTaskList gone, flagged for V3; live verb observation deferred to T011 with honest record; six protocol-gap judgment calls enumerated in report; orchestrator re-ran gradle test green), 4 done (4c25711/4ac4264/a18aea1/4c08511, sonnet verified, ~381k — the heavy phase: §12 LeakPassTest six passes green over real composed payloads; LIVE dev-server observation — headless runServer, real villager, Python stub mind over UDS, all four verbs with exactly one act_result each, two live bugs found+fixed (sentinel overflow, pause-when-empty), not-observed items recorded honestly; body-protocol-seam + villager-brain-api re-verified + re-pinned, CAPSULES regenerated; 13 card ACs ticked with citations). Reconciled with origin/main post-#15: merge-in 3aa4a2a, body-protocol-seam conflict resolved as union of V2+M2 amendments, NEEDS-REVIEW classified, re-pinned to merge commit 152d8fc; gates + freshness probe re-run green after the history move. PR #17 MERGED 81c2b75 (merge commit; card synced Done, board-sync at root); worktree+branch removed after merge verified. ~883k subagent tokens across 4 sonnet dispatches. |
| 2026-08-27 | TASK-0014 | #19 | 1e19cf5 | ~2.1M subagent (sonnet, incl. diagnosis re-runs) + orchestrator | closed (lane 3): claimed (c5d898f), spec cycle (cafc8cb), phase ACs seeded (652d587). phases: 1 done (34a16eb/0bfd69f/809bd6b, sonnet VERIFIED serving cc/claude-sonnet-5[1m], ~209k; 26.2 brain surface derived — Schedule/ScheduleBuilder gone, addActivity additive-only, MemoryModuleType/PoiType plain-API at 26.2, gossip() calls spawnGolemIfNeeded directly so one injection may cover both, Activity.MEET may sidestep new-Activity Mixin for Phase 3; cast/ seeded via SavedData, 5/5 tests; villager-brain-api amended + re-pinned, CAPSULES regenerated, freshness clean), 2 done (445225b, sonnet verified, ~405k — the heavy phase: Mixin count 2 of 3 (VillagerGoalPackagesMixin drops VillagerMakeLove from IDLE; VillagerGossipMixin cancels gossip() covering golems too), MixinConfigTest asserts the enumeration; three vanilla AI traps found — chunk-unload-at-zero-players fixed by setChunkForced, ResetProfession defused by XP=1, job-site claim flaky on POI registration timing — flaky claim tolerated by spec edge case (wander fallback), carried to Phase 4 full-cycle proof; zero breeding/golem events 1000+ ticks; gradle green), 3 done (d54395c/995407f/240733b, sonnet verified, ~910k across three dispatches — first agent idle-looped on a log wait and was stopped, closer verified+committed its work: DuskPairing rides vanilla Activity.MEET (Bell POI + setMemory, Mixin budget holds at 2), signal via existing sighting percepts, 9 timing/no-fire unit tests green; boot-order race found+fixed (setUp deferred to tick loop past ENTITY_LOAD); substrate fact: 26.2 schedules read monotonic getGameTime not day-time, so /time set cannot shortcut observation; T009 honestly unticked — live signal not reached in-session, resumption plan in pair-observation.md, folds into Phase 4's full-cycle run), 4 done (2b19674/6ae6acc, sonnet verified, ~239k; stalled-mind PROVEN — 1474 failed dials, zero stalls; zero breeding/golems full-day span; ACs #2/#3/#5/#6 ticked (board-sync 7671d3f at root — MCP-resolves-against-root trap hit again), #1/#4 honestly open: ZERO villager movement all run; villager-brain-api re-pinned with corrections). Zero-movement DIAGNOSED + FIXED (22e93e1, sonnet verified, ~192k): cast wandered one chunk past the forced set — forced ticket ENTITY_TICKING decays to BLOCK_TICKING one chunk out, brain AI permanently stalls; fix keepChunksLoaded re-forces from actual positions + 3-chunk margin per boot; live console verify: villager resumed schedule, walked to bed, slept. Verification re-run (9b1ca21, sonnet verified, ~175k): 29m49s unattended post-fix — wake x2, dusk convergence, 10x pairing-signal firings all pairs, sleep in claimed beds bit-identical ~4min, zero breeding/golems x24; ACs #1/#4 ticked (board-sync 44bc5dc at root); caveat surfaced: live signal leads 1.82-4.96s vs nominal ~10s (vanilla stroll-hop prediction compression, mechanism unit-tested correct); JOB_SITE claim still absent — pre-existing tolerated gap, flagged for refactor-triage. Reconciled with origin/main (285a7d7, board conflict took main's side), gates + freshness re-run green. PR #19 MERGED 1e19cf5 (merge commit; card synced Done 5d1e09c, board-sync at root); worktree+branch removed after merge verified. |
| 2026-08-27 | TASK-0015 | #18 | e194e48 | ~740k subagent (sonnet x4) + orchestrator | closed (lane 3): claimed (6ee718a), spec cycle (7f1550d), phase ACs seeded (0ba2114). phases: 1 done (baafbd7, sonnet VERIFIED cc/claude-sonnet-5[1m], ~191k; mind/fakevendor per §10.1 over the real seam surface, reflection API-surface test locks the nine-item shape, seamtest cross-linked; strict/restrict flags data-only pending H-wiring (ponytailed), Resolve errors-not-panics per module idiom; vet+test+race green), 2 done (b09a954, sonnet verified, ~158k; H-1..H-4 in harness_test.go, mutation checks hand-verified red (Validate-stub for H-1, naive-classifier swaps for H-2/H-3); H-2 'absent origin' implemented as empty-string since missing-key is H-1's rejected case — deviation recorded in test comments), 3 done (9266eb6/540d232, sonnet verified, ~167k; H-5 via unexported issued-token bookkeeping derived from Emit — exported surface unchanged, reflection lock still green; H-6 flood measured 3.50x (42 vs 12, n=10, 3 bodies), mind-side memory counts via M2 gate+instrument; all mutation checks live-verified red then reverted), 4 done (6d91864/ce909d0/6fa626f/9c789c3, sonnet verified, ~224k; §10.2 e2e against the fake — step 5 proven, told cannot become witnessed through the real seam; TASK-0010 AC #5 closed (a37da92); body-protocol-seam NEEDS-REVIEW amended + re-pinned, pre-existing size-budget debt trimmed/exempted (flagged for refactor-triage), CAPSULES regenerated, freshness 11/11; card ACs #1-#5 ticked cited). PR #18 OPEN — orchestrator gates re-run green, base unmoved; SELF-MERGE DENIED by permission classifier (PR #16 precedent) — operator MERGED e194e48 (merge commit; card synced Done dbad80a, board-sync at root); worktree+branch removed after merge verified. ~740k subagent tokens across 4 sonnet dispatches. |
| 2026-08-28 | TASK-0018 | — | — | — | in flight (lane 4, parallel with TASK-0016/0017): claimed (ee4aaed), spec cycle (ac47cfb), phase ACs seeded + tier note (4269b7b). phases: 1 dispatched (sonnet; served model verified on the TASK-0016 lane-lead dispatch — claude-sonnet-5). |
| 2026-08-28 | TASK-0013 | — | — | — | in flight (lane 3, final task): claimed (d03eb91), spec cycle committed, phase ACs seeded (b10ed08), tier note on card. phases: 1 done (e9e2673, sonnet VERIFIED serving cc/claude-sonnet-5[1m], ~120k; persona type + write-once 0444 O_CREATE|O_EXCL + Load re-bind, external API-surface test via reflection+AST; cast pairing corrected against Cast.java — armorer/plains, farmer/desert, fisherman/taiga, deviation recorded in tasks.md; orchestrator re-ran vet+test green), 2 done (ac6d9ec, sonnet verified, ~116k; anchor echo + word-boundary drift markers + authored moralizing lexicon grounded in brief spell-breaker, reject-with-reason, no-llm-import structural test; Validate added to exported want-list — read-only, AC #2 argument intact; orchestrator re-ran vet+test green uncached), 3 done (0ba6a80, sonnet verified via transcript, ~152k; Genesis one-E1-call-per-entry on Opus 5, JSON-in-text strict decode — E1 StructuredOutput:false per ratified registry, plan-flagged risk not new judgment; pairing copied never model-invented, validator round-trip in Genesis, want-list widened Genesis writes only via WriteOnce; orchestrator re-ran vet+test green uncached), 4 first dispatch STOPPED honestly (c427882/7620e48/8bb5075, sonnet verified, ~168k; live harness built, live E1 call 404 no-credentials, zero billed, wiki re-grounded promptworld-lineage+overview; T010/T011 honestly unticked). Orchestrator DIAGNOSED via curl: proxy routes by model-ID prefix — bare claude-opus-5 404s, cc/claude-opus-5 200s; path shape fine. Completion done (186a6ba+a333b06, sonnet verified, ~119k; ANTHROPIC_MODEL_PREFIX request-build-time override, canonical IDs unchanged; LIVE genesis 3x E1 on claude-opus-5, zero rejections — Merrow Vand/Sefa/Ottiline Vask, 0444 at gitignored mind/run/persona, restart re-bind PASS over real files; overview re-pinned, freshness green). ACs #1-#6 ticked cited; orchestrator re-ran full gates green uncached, probe green, base unmoved. PR #20 OPEN — self-merge denied by permission classifier (PR #16/#18 precedent) — awaiting operator merge. ~840k subagent tokens across 5 sonnet dispatches. |
