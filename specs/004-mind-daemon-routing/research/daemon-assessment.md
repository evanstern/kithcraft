# Daemon assessment — spec 004 Phase 1

**What this is.** The evidence Phase 3's language/reuse recommendation will rest on. Three
bodies of evidence: (§1) what promptworld I's mind-side machinery actually contains and
whether it is portable behind the seam, (§2) what the accepted seam contract demands of
*any* daemon regardless of language, and (§3) the candidate languages with cited evidence.

**Phase 1 gathers; it does not recommend.** No option is preferred here. Every §3 judgement
is a per-criterion finding, and the summary tables are deliberately findings-only.

**Reading key.** `[I:<path>]` = a path in promptworld I's repo
(`/Users/evanstern/projects/promptworld`), read read-only for this assessment.
`[I-wiki:<note>]` = a note in I's grounding corpus (`docs/wiki/`), consulted INDEX-first and
just-in-time per [[promptworld-lineage]] and the corpus-loading rule.

**Source discipline.** §1 claims carry an I-repo path. §3 external claims carry a URL and an
accessed date. Everything checked live on **2026-08-21** unless stated otherwise.

**A note on reading I's source at all.** [[promptworld-lineage]] says doctrine transfers and
code does not, and spec 002's `doctrine-port.md` deliberately cited only I's *wiki notes*,
never its files. This assessment is the one place where reading the tree is the job: R1 asks
whether the code is portable, and "portable" is a property of code, not of a summary. Nothing
is imported; paths are cited as evidence about size and coupling.

---

## 1. promptworld I's mind-side machinery, classified

### 1.0 What "the daemon" actually is

I is **one Go module, one binary** (`github.com/evanstern/promptworld`, Go 1.26.4) that is
both the daemon and every client tool `[I:go.mod]`, `[I-wiki:overview]`. There is no separable
"mind daemon" artifact to keep or discard: the mind is a package inside a process whose other
packages are the world.

Measured size, excluding worktrees and tests: **71,744 lines** of non-test Go across 21
`internal/` packages `[I:internal/]`. The packages that matter here, code/test lines:

| Package | Code | Test | Role |
|---|---|---|---|
| `internal/sim` | 22,035 | 32,406 | **the world engine** — state, reducer, executor, reflex, mental maps, memory, salience |
| `internal/mind` | 6,515 | 10,451 | the planner/consolidation/conversation driver |
| `internal/llm` | 3,770 | 6,012 | provider registry, routing, budget, breakers, leases |
| `internal/tui` | 14,044 | 13,873 | the terminal client — **dies** |
| `internal/guardian` | 7,372 | 9,662 | the player-as-intermediary — **dies** |
| `internal/toolloop` | 994 | 2,291 | bounded tool-use loop driver |
| `internal/tool` | 2,255 | 1,425 | tool/verb registry and schemas |
| `internal/cognition` | 996 | 1,751 | cognition horizon, router, governor — **dies** (brief #8) |
| `internal/persona` | 475 | 327 | write-once persona files |
| `internal/store` | 427 | 320 | append-only SQLite event log |

The three largest packages are `sim`, `tui`, and `guardian` — and per
[[promptworld-lineage]] **two of those three die outright**, along with `cognition`.

### 1.1 The doctrine items, one at a time

#### D1 — Event-sourced memory

**What implements it.** Two layers, and they are not the same layer.

*The log itself* is `internal/store`: an append-only SQLite `events` table with
`events_no_update` / `events_no_delete` triggers raising `ABORT` in-schema, contiguous seq
enforcement, and canonical struct-marshaled JSON payloads `[I:internal/store/store.go]`,
`[I:internal/store/schema.go]`, `[I-wiki:event-log]`. This layer is **427 lines** and imports
nothing of I's above it.

*The event vocabulary and the reducer* are `internal/sim` — `State.Apply`, the single mutation
path, split across 143 files with an event taxonomy that I's own wiki needed **sixteen child
notes** to catalog `[I-wiki:event-types]` and its children.

**Portable?** **Split verdict, and the split is the important finding.**

- The *pattern* — append-only log, immutability enforced in schema not convention, state as a
  reducer over the log — is doctrine and ports for free. It is also ~400 lines of anyone's
  language, not a thing to port.
- The *vocabulary* does not port. I's event types are world-engine events: `agent.moved`,
  `agent.foraged`, `agent.chopped`, `agent.wall_chipped`, `gru.*`, `guardian.*`,
  `curriculum.*`. Under the seam these are **body-vendor concerns** — the mind never sees
  them, and half their producers (gru, guardian, curriculum) do not exist in Kithcraft.
- One entanglement is load-bearing and must be named: I's log carries a
  **`log_format_version` gate with a rename-requires-migration doctrine**
  `[I:internal/store/format.go]` and a whole migration chain v1→v5
  `[I-wiki:world-migration]` — machinery whose entire justification is
  **determinism-for-replay**, which dies with I ([[promptworld-lineage]]). A Kithcraft mind's
  log is a memory store, not a replayable world; it does not need byte-identical replay, and
  every constraint that flows from that requirement (canonical field order, emitter-computed
  payloads, the spec-092 replay-hazard audit `[I-wiki:sim-state-reducer-replay-hazards]`) is
  cost with no remaining benefit.

**Verdict: doctrine ports, code does not, and the pattern is small enough that "porting" is
not the operation.** Note the seam agrees independently: RM-5/RM-6 remove the last
determinism-flavored reason for read-time-only decay and replace it with a better one
(`body-protocol-v0.md` §6.4, "a mind that mutates nothing on a timer has no clock to keep in
sync with the vendor's").

#### D2 — The reflex/planner split

**What implements it.** The reflex half is `internal/sim/policy.go` + `path.go`:
`decideIntent`'s four classified rung groups (`survivalDecision` / `directiveDecision` /
`prepDecision` / `wanderDecision`) over a BFS pathfinder with fixed N/E/S/W neighbor order and
deterministic tie-breaking `[I-wiki:reflex-policy]`, `[I-wiki:reflex-pathfinding]`. The
planner half is `internal/mind` driving `internal/toolloop` against
`internal/tool`'s roster `[I-wiki:agent-mind]`, `[I-wiki:tool-loop]`.

**Portable?** **The doctrine is the single most transferable item (the brief says so) and the
reflex code is 100% dead.**

- The reflex ladder is a 2-D grid survival policy: `forage`/`chop`/`hunt`/`build_fire`/
  `refuel_fire`/`goto_warmth`/`warm_up`/`bathe`/`quarry`/`craft_planks`/`build_wall_stone`…
  `[I:internal/tool/registry.go:353-381]`. Every one of those verbs is a promptworld-I world
  mechanic. Kithcraft's reflex half is **Minecraft's own villager brain** —
  `Schedule`/`Activity`/`MemoryModuleType`/POI on `Brain<E>` ([[villager-brain-api]]) — under
  decision-0002's augmented vanilla villager. The reflex code does not port because the
  reflexes now live in someone else's engine.
- The BFS pathfinder is doubly dead: it pathfinds a 64×64 tile map, and AR-4 forbids the mind
  from doing spatial arithmetic at all (`body-protocol-v0.md` §3: "the mind performs no
  spatial arithmetic … this is exactly the reflex half of the split — and it is why no
  geometry needed porting from I").
- The *planner* half is the genuinely reusable shape: `internal/toolloop` (994 lines) is
  explicitly "transport-agnostic and sim-agnostic … it imports only `internal/llm` (the wire)
  and `internal/tool` (the schema/roster source)" `[I:internal/toolloop/loop.go:1-19]` —
  verified: its only project imports are those two `[I:internal/toolloop/loop.go]`. Its
  doctrine ("a tool call is a REQUEST; an event is the FACT; the gate decides") maps
  **exactly** onto the seam's intent/`intent_ack`/`act_result` split (`body-protocol-v0.md`
  §5.1). This is the strongest single portability finding in §1.
- `internal/tool`'s **registry mechanism** is portable and imports nothing of I's (verified: no
  project imports in any non-test file `[I:internal/tool/]`); its **vocabulary** is entirely
  promptworld-I verbs and is replaced wholesale by the vendor's `session_open` manifest
  (`body-protocol-v0.md` §6.2) — which is a better design than I's compiled-in registry,
  because the verb set now arrives at runtime from the vendor.

**Verdict: reflex code dead (the engine owns it now); planner/tool-loop *shape* highly
portable, its *vocabulary* replaced by the manifest.**

#### D3 — Salience, consolidation, situated memory

**What implements it.** Three pieces:

1. **The salience table** — `salTalk`=3 … `salWitnessDeath`=10, plus ~15 later additions
   `[I:internal/sim/memory.go:252-304]`, `[I-wiki:agent-memory-window]`.
2. **The working window** — `SelectMemories`: salience halved per game-day of age, top K−2,
   plus 2 seeded serendipity picks from the oldest half, K=`WindowK`(10)
   `[I:internal/sim/memory.go]`, `[I-wiki:agent-memory-window]`.
3. **Nightly consolidation** — sleep-triggered single cloud call digesting the episodic
   buffer into promotes/fades/beliefs/narrative, landing atomically
   `[I:internal/mind/consolidate.go]` (416 lines), `[I:internal/sim/consolidate.go]` (428),
   `[I-wiki:nightly-consolidation]`.
4. **Situated memory** — `situatedMemoryEvent`/`situatedMemoryAboutEvent`/
   `situatedMemoryToned`, each baking a `Where` (`PlaceAt`→`describePlace`, a deterministic
   Manhattan feature scan) and a `Why` (the driving intent's reason verbatim)
   `[I:internal/sim/memory.go]`, `[I-wiki:executor-social-perception]`.

**Portable?** **Mixed, with one hard entanglement and one already-resolved conflict.**

- **The salience table is world-side in I and MUST NOT be in Kithcraft.** In I the *world*
  minted memories from a fixed table, and salience did double duty: ranking the working-memory
  window *and* gating an interrupt band (`GenerationBumpSalience`=9,
  `[I-wiki:agent-memory-window]`). `body-protocol-v0.md` §2.8 explicitly forbids this:
  "**no `salience`, `importance`, `weight`, or `memorability` field may exist on any percept,
  in v0 or any successor**". The three uses split — urgency stays body-side, formativeness is
  mind-side and does not cross, and I's interrupt *mechanism* (a generation counter superseding
  in-flight thoughts, `Agent.Generation`, `[I:internal/sim/agents.go:184+]`) was entangled with
  the cognition-horizon machinery that dies and is "left entirely to the mind daemon
  (TASK-0004)". **This is an open design item this task inherits, not a port.**
- **The consolidation shape ports strongly.** Its non-obvious, expensive-to-rediscover parts
  are all mind-side and vendor-independent: the once-per-night event-sourced ledger
  (`ConsolidationDue` on `NightIndex(tick) > LastConsolidatedNight` plus a 12-game-hour gap),
  the ordinal-label `m1..mN` prompt convention (memories have no IDs and slice indexes are
  unstable), the `(tick, hash)` durable identity mapping accepted refs back to events, the
  **truncation-aware retry ladder** (1024→2048→4096 on `StopMaxTokens`, at most 2 retries,
  a parsing reply never consuming one) `[I:internal/mind/retry.go]`, and "transport failure
  lands **no marker** — the attempt never happened". The retry ladder in particular is a
  *measured* fix (playtest-1: a late world's digest outgrew 1024 tokens and every night was
  silently rejected from day 20 on, `[I-wiki:nightly-consolidation]`) — exactly the kind of
  finding that is cheap to carry and expensive to rediscover.
- **Situated memory half-ports, and the seam already relocated the halves.** I's `Where` was
  computed by the *world* (a feature scan over the tile map) and its `Why` by the *mind* (the
  intent's reason). `body-protocol-v0.md` §2.4 and §5.2 make this explicit and correct: the
  vendor composes `place.descriptor` ("the well") because only the vendor can see the world,
  and the mind supplies `intent.reason` because only the mind knows why. So `describePlace`
  (the code) is dead; the *composition grammar* (`<base> at <desc> — <why>`, composed once at
  emission so no consumer re-derives) is portable and is a genuinely good idea.
- The window selector itself (`SelectMemories` and the spec-042 relevance-blended variant) is
  ~1 file of ranking arithmetic. Portable in principle; small enough that porting is not the
  operation.

**Verdict: consolidation *machinery-shape* and its measured lessons port well; the salience
table's world-side placement is forbidden by the accepted contract and its interrupt half is
an open item this spec inherits.**

#### D4 — Epistemic hygiene

**What implements it.** Four mechanisms:

1. **`Origin`, a closed vocabulary stamped at emission**, a required parameter on every
   situated constructor so a new unstamped emission site cannot compile
   `[I:internal/sim/memory.go:43-55]`.
2. **`DirectPerception(origin)`** — a pure switch over the stored field, 8 lines
   `[I:internal/sim/memory.go:62-70]`.
3. **`enforceProvenance`** — the evidence-cited coercion gate: resolves each belief's
   ordinal evidence refs to durable `(tick, hash)` identities, asks `DirectPerception` of each,
   and **coerces** a `witnessed` claim with no direct backing down to `told` (some secondhand
   resolved) or `inferred` (nothing resolved), counting coercions and **never rejecting**
   `[I:internal/mind/validate.go:97-126]` — 30 lines.
4. **Mental maps** — per-agent private `MentalMap{Explored, Facts []PlaceFact, Peers
   []PeerSighting}`, each fact carrying provenance, a source, and a last-seen tick, with
   **read-time-only** freshness horizons (`factHorizonVolatileTicks` 12 game-hours,
   `factHorizonDurableTicks` 4 game-days) and no silent forgetting
   `[I:internal/sim/mentalmap.go:24-110]`, `[I-wiki:mental-maps]`.

**Portable?** **This is the most portable doctrine item and simultaneously the one that has
already been ported.** Two findings:

- **The mechanisms are tiny.** `DirectPerception` is 8 lines; `enforceProvenance` is 30. The
  *value* was never in the code — it was in knowing that a pure, text-free classifier is the
  only thing that stops a language model writing "I saw it myself" into a belief. That
  knowledge is now a **protocol obligation** in `body-protocol-v0.md` §2.7, with
  `DIRECT_ORIGINS` defined in the contract "so mind and vendor cannot disagree", a MUST NOT
  against a `direct` boolean on the wire, and H-3 in the fake-vendor harness proving the
  classifier ignores prose (§10.3).
- **`enforceProvenance`'s coerce-never-reject posture is the highest-value carry** and it is
  already RM-2/RM-3 in the contract. `mentalmap.go` is the one substantial file in the whole
  assessment that is **self-contained** — its only imports are `encoding/base64` and `sort`
  `[I:internal/sim/mentalmap.go]`, zero project imports. But its *content* is 2-D grid facts
  (`X`, `Y` ints, an `Explored` bitmap over a tile map), and AR-4 forbids coordinates crossing
  the seam at all. So the most portable file in I is portable in *structure* and dead in
  *substance*: what survives is "a private, provenance-stamped, read-time-fresh fact store per
  mind", which the contract already mandates as the mind's private map (§6.1).

**Verdict: fully ported already — into the protocol, not into code. Nothing to reuse; nothing
lost.**

#### D5 — The persona firewall

**What implements it.** Two halves, deliberately:

- *Structural*: `persona.md` written exactly once by `promptworld new` at **mode 0444**, with
  no post-genesis write path anywhere in the codebase
  `[I:internal/persona/personas.go:1-5]`, pinned by a test asserting the mode
  `[I:internal/persona/persona_test.go:25]`.
- *Validatory*: a deterministic, **model-free** validator — an anchor echo compared under
  normalization (lowercase, trimmed, trailing `.`/`!` stripped, whitespace collapsed) plus an
  authored per-villager drift lexicon matched word-boundary/case-insensitively against the
  narrative and any self-belief; any hit rejects the night
  `[I:internal/mind/validate.go]`, `[I-wiki:nightly-consolidation]`.

**Portable?** **Fully portable, entirely mind-side, and the smallest package in the
assessment** (`internal/persona`: 475 code lines, and `personas.go` has **no imports at all**
`[I:internal/persona/personas.go]`).

Two carries worth naming:

- **The two-half design is the doctrine**: structural impossibility (no write path exists)
  plus mechanical validation (no second model call, so rejection is a testable 100%
  guarantee). Either half alone is weaker.
- **The honest limit is on the record and transfers with it**: "the lexicon catches *stated*
  drift; subtle drift needs the parked model-judged validator"
  `[I-wiki:nightly-consolidation]`.

One Kithcraft-specific adaptation is forced by the brief and is not a portability problem:
brief decision 5 requires personas be **generated at birth**, where I's were authored
write-once at genesis for a fixed cast of eight. The firewall's structural half (no edit path)
is what decision 5 actually demands ("the moment a player can open a villager's head and edit
its values, the villager is a puppet"); the generation source changes, the firewall does not.

### 1.2 The entanglement measurement

The question R1 asks is whether I's mind code can be lifted behind the seam. Measured:

- `internal/mind` (non-test) references **151 distinct `sim.*` symbols**
  `[I:internal/mind/]`. `sim` is the dead world engine. The top references are not incidental:
  `sim.State` (the mind holds a **live replica of world state** and applies every event to it,
  `[I:internal/mind/mind.go:200-310]`), `sim.NewState`, `sim.Memory`, `sim.PlaceFact`,
  `sim.Belief`, `sim.Needs`, `sim.Intent`, `sim.PlanStep`, and ~30 distinct `*Payload` types.
- The mind's **prompt assembly reads world state directly** — `s.Agents[idx]`, other agents'
  names, relationship edges, debts, norms `[I:internal/mind/prompt.go:90-384]`.
- The mind's two seams to the world are `Loop.InjectIntent(sim.InjectArgs)` and
  `Loop.InjectSocial([]store.Event)` `[I:internal/mind/mind.go:33-35]`,
  `[I:internal/mind/convo.go:51]` — **both take `sim` types**, and `InjectSocial` takes raw
  event rows. These are in-process function calls into a single-goroutine reducer, not a
  protocol.
- **The cast size is a compile-time constant baked into array types.**
  `const AgentCount = 8` `[I:internal/sim/agents.go:14]`, and `internal/mind` declares **ten
  fixed-size arrays over it** — `personas [sim.AgentCount]string`, `nextDue`, `lastPlanned`,
  `pending`, `pendingSeq`, `planInFlight`, `planCancel`, `consolInFlight`, `world
  [sim.AgentCount]agentSnap`, and it appears in `mind.New`'s signature
  `[I:internal/mind/mind.go:65-199,594]`. Kithcraft's cast is 3–6 and villagers can die
  permanently (brief #4). This is a small, mechanical change — but it is a change *inside the
  mind package's core types*, which is evidence about how thoroughly the mind was written
  against one specific world.

**The finding.** `internal/mind` is not a daemon behind a seam that happens to sit in the same
repo. It is a **co-process of the world engine**: it holds a replica of world state, reads
that state directly for prompting, and lands its output through in-process injection doors
typed in world-engine terms. Lifting it behind `body-protocol-v0` is not a port — it is a
rewrite of its inputs (replica → percept inbox), its outputs (`InjectArgs` → `intent`), its
state (`sim.State` → the mind's own private store), and its prompt assembly (world reads →
percept-derived belief store). What survives that operation is the *design*, not the file.

**The three genuine exceptions**, each verified by imports:

| Package | Code lines | Project imports | Note |
|---|---|---|---|
| `internal/toolloop` | 994 | `llm`, `tool` only | explicitly sim-agnostic by design `[I:internal/toolloop/loop.go:14-18]` |
| `internal/tool` | 2,255 | **none** | mechanism portable, vocabulary replaced by the manifest |
| `internal/persona` | 475 | **none** | `personas.go` imports nothing at all |

`internal/llm` (3,770 lines) is a fourth near-exception: its only project import is
`internal/cognition` `[I:internal/llm/llm.go]` — which is the **cognition horizon, a package
that dies with I** (brief #8, [[promptworld-lineage]]). So `llm` is one dead dependency away
from portable. But note what that dependency *is*: `cognition.Estimator`, `Profile`,
`ClassForKind`, `ValidateKinds` — the seconds-per-point/staleness-budget machinery whose
entire purpose was surviving speed ladders. At 1x, an LLM layer needs providers, retries, a
budget meter, and concurrency; it does not need a staleness router. Much of `llm`'s 3,770
lines exists to serve the thing that dies.

### 1.3 Summary table

| Doctrine item | I's implementation | Portable behind the seam? |
|---|---|---|
| **Event-sourced memory** | `store` (427 ln, append-only SQLite, in-schema triggers) + `sim`'s 16-note event taxonomy + a format-version/migration chain | **Pattern yes (and trivial); vocabulary no** — I's events are world events the mind never sees under SI-1. Replay-determinism machinery dies with its justification. |
| **Reflex/planner split** | reflex: `sim/policy.go`+`path.go` (2-D grid ladder, BFS). planner: `mind`→`toolloop`→`tool` | **Reflex code fully dead** — Minecraft's `Brain<E>` is the reflex half now, and AR-4 forbids mind-side geometry. **Planner shape strongly portable** — `toolloop` is sim-agnostic by construction and its request/fact/gate doctrine maps onto intent/ack/act_result. |
| **Salience + consolidation + situated memory** | salience table + `SelectMemories` window + `mind/consolidate.go` (416) / `sim/consolidate.go` (428) + situated constructors | **Consolidation shape + measured lessons port well** (ledger, ordinal refs, `(tick,hash)` identity, truncation retry ladder, no-marker-on-transport-failure). **Salience table's world-side placement is forbidden** by §2.8; its interrupt-band half is an open item inherited here. Situated `Where`/`Why` already re-split correctly by §2.4/§5.2. |
| **Epistemic hygiene** | `Origin` vocab + `DirectPerception` (8 ln) + `enforceProvenance` (30 ln) + `mentalmap.go` (582 ln, zero project imports) | **Already ported — into the contract, not into code.** §2.7 defines the classifier; RM-1…RM-7 carry the rest. Mechanisms are tiny; the value was the knowledge. `mentalmap.go`'s content is 2-D coordinates, which AR-4 forbids crossing. |
| **Persona firewall** | `persona` (475 ln, no imports) mode-0444 + model-free anchor/drift-lexicon validator | **Fully portable, entirely mind-side.** Two-half design (structural + validatory) is the doctrine. Brief #5 changes the *source* (generated at birth) not the firewall. Honest limit transfers with it. |

**What dies, per [[promptworld-lineage]], confirmed by measurement:** `sim` (22,035),
`tui` (14,044), `guardian` (7,372), `cognition` (996) — **44,447 of 71,744 non-test lines,
62% of the codebase, before counting the `mind`/`llm` code written to serve them.**

---

## 2. What the seam contract demands of any daemon

Extracted from `docs/design/body-protocol-v0.md` (contract accepted 2026-08-21). These are
language-independent obligations — the cost floor any candidate in §3 pays.

### 2.1 Surfaces to implement

**Perceive (§4) — an inbox, and nothing else.**

- Accept a **one-directional push** of typed percepts (`sighting`, `observation`, `sound`,
  `speech`, `told_fact`, `text`, `act_result`, `self_state`, `change_report`), each with an
  envelope (§2.1: `protocol`/`message`/`session`/`seq`/`body`/`world_time`/`payload`).
- **Hold no world handle.** SI-1: the percept stream is the only path by which world state
  reaches the mind. There is no query, no lookup, no callback. Implication for the daemon's
  shape: everything the mind knows must live in the mind's own store, reconstructed from
  percepts.
- **Tolerate the vendor's declared subset.** Only four percept types are floor (`act_result`,
  `observation`, `sighting`, `speech`); `sound`, `told_fact`, `text`, `self_state`,
  `change_report` are optional and **"a mind MUST run correctly against a vendor that
  declares none of them"** (§6.2). So no percept handler may be structurally required.
- **Reconcile absence.** `observation` carries `vocabulary` (the closed set scanned) and
  `present`; absence is `vocabulary` minus `present`, and reconciling it against held beliefs
  — confirmation, bounded disconfirmation, silence leaving things untouched — is **entirely
  mind-side** (§4.3). This is the single largest piece of cognition the contract assigns to
  the daemon.
- **Dedup and reorder.** `percept_id` for dedup under at-least-once delivery; `seq` monotonic
  per `(session, body, direction)`, gaps mean loss (§2.1, §8 T-2).

**Act (§5) — intents out, results back as percepts.**

- Emit `intent{intent_id, verb, target, reason, supersedes?, not_after?}`; handle
  `intent_ack{accepted, reason_code}` — **acceptance is not success**; consume exactly one
  `act_result` percept per accepted intent (§5.4). The daemon must therefore track pending
  intents and tolerate an **indefinitely** pending act — the fake vendor's default is to ack
  and do nothing until the script resolves it, precisely because "waiting" is the state a mind
  occupies most of the time (§10.1).
- **Author a `reason` on every intent** (§5.2) — mind-authored, opaque to the vendor, echoed
  back on the result. Not optional decoration: it is what situates the resulting memory.
- **Name only known tokens.** A target must be a token the mind received in a percept; a
  description to search for ("the nearest bed") is forbidden (§5.2, L-5).
- **Learn of death by failing** (§5.6): `target_gone` on an `act_result` is the *only*
  sanctioned non-existence discovery channel. No polling, no probing.
- Support `cancel` and `supersedes`.
- Implement against the **four core verbs** (`go_to`, `speak`, `attend`, `wait`) plus whatever
  the manifest declares — so verb handling is **runtime-dispatched from the manifest**, never
  a compiled-in enum. (This is the direct contrast with I's compiled `internal/tool` registry.)

**Remember (§6) — the mind owns everything durable.**

- **Own the whole durable store**: percept history, beliefs, relationships, personas, plans,
  and the mind's private map (known places and facts, each with provenance and a last-seen
  time). The vendor persists none of it and MUST NOT rank or weight it (SI-5, §6.1).
- Handle `session_open`'s manifest (§6.2): `time_unit`, `continuity`, and `capabilities`
  (`percept_types`, `origins`, `verbs` with target shapes, `salient_kinds` with roles and
  descriptors, `bearings`, `distance_bands`). **The vocabulary arrives at runtime.**
- Handle **session continuity as a gap** (§6.3): `body_continuous: false` means a different
  body; the mind is entitled to know its body changed and *not* entitled to an account of the
  interval.
- Enforce **RM-1…RM-7** on itself: no durable direct-perception claim without a retained
  direct percept; model-authored beliefs cite percepts and are **coerced, not rejected**;
  secondhand never overwrites fresher firsthand; stored confidence never changes with time
  (effective confidence computed at read time from age and a half-life, with convictions
  outliving vividness); freshness evaluated at read time per kind; **no silent forgetting**.

**Cross-cutting.**

- Implement `direct_perception(origin)` as a **pure function of the origin value alone** —
  never reading percept text, descriptors, source names, or `hops` (§2.7). An unrecognized or
  absent origin classifies **secondhand** (V-6).
- Receiver rules V-1…V-6: ignore unknown fields; fallback branch for unknown enum values in
  open vocabularies; retain-but-never-interpret an unknown `percept_type`; treat a missing
  required field as malformed.
- Treat every world concept as opaque: never parse a `kind` token's spelling (AR-2), never do
  spatial arithmetic (AR-4), never branch on prose fields (`descriptor`, `doing`, `detail`).
- Do **time arithmetic** on `world_time` integers — legitimately, and it is the only
  arithmetic the mind is supposed to do (L-9), because RM-6's read-time freshness requires it.
- Version negotiation fails closed (§7.1): a mind that cannot speak the vendor's version
  replies `session_close`/`unsupported_version`.

**Not demanded — flagged, not decided.** Transport is open (§8, Q-1). Any candidate must be
able to satisfy T-1…T-7: push-not-pull, ordered-or-reorderable per body, long-lived sessions
with explicit open/close, mind-restart independent of vendor-restart, message-oriented with a
schema, backpressure shedding only `background`, and **process-separable but not
process-required** (T-7). Per this spec's non-goals, a language finding that bears on
transport is *flagged here* and decided elsewhere.

### 2.2 What the fake-vendor spec requires of a daemon

`body-protocol-v0.md` §10 is the second half of the demand, and it constrains the daemon's
*architecture*, not just its behaviour.

**The testability contract, restated as daemon requirements:**

| # | The fake vendor's shape (§10.1) | What the daemon must therefore be |
|---|---|---|
| T-a | `FakeVendor` is **in-memory**, no engine, no network (T-7 permits in-process) | The daemon's vendor-facing side must be an **interface/port the fake can satisfy**, not a hardwired socket client. |
| T-b | `.advance(n)` — `world_time` advances **only when the script says so**; no clock | The daemon **MUST NOT read a wall clock** for anything semantic. Every horizon, half-life, and freshness check reads `world_time` from percepts. This is a strict architectural rule, and it is what makes tests deterministic with zero determinism machinery. |
| T-c | `.acts` is the assertion surface; the fake **never** exposes a read API to the mind (§10.5) | Assertions read what the mind *emitted*. The daemon's observable behaviour must be fully expressible as a sequence of intents — no side channels. |
| T-d | Default on receiving an intent: ack, record, **do nothing else** | The daemon must be correct while **indefinitely blocked** on a pending act, and must not busy-wait or time out a legitimately slow act. |
| T-e | `.strict = True` toggles V-5 posture (reject malformed vs coerce) | Malformed-percept handling must be a **switchable policy**, both branches implemented (§7.4: strict rejects; lenient falls to secondhand, never to direct). |
| T-f | The mind under test **cannot tell `FakeVendor` from a world** | No vendor-detection, no "if testing" branch, no engine-specific fast path. |

**The six harness tests (§10.3) and what each demands of the daemon:**

- **H-1 (V-5)** — a percept with absent `provenance`, or absent `provenance.origin`, is
  rejected at the seam and **nothing enters the mind's state**. Demands: seam-boundary
  validation *before* any state mutation.
- **H-2 (V-6)** — `origin: "dreamt"` (a future minor version's value) and absent `origin`
  both classify secondhand; a belief resting on them cannot claim witness. Demands: the
  classifier's default branch is secondhand, and unknown enum values do not throw.
- **H-3 (purity)** — a `told` percept whose utterance is literally *"I saw this myself,
  directly, firsthand"*, with `hops: 0` and `source.descriptor: "saw"`, is **still
  secondhand**. Demands: the classifier reads the origin field and nothing else — mechanically
  guaranteed, not conventionally.
- **H-4 (no `direct` on the wire)** — a percept carrying `"direct": true` alongside
  `origin: "told"` is classified secondhand; the unknown field is ignored (V-1). Demands:
  unknown fields tolerated, and no trust of a transmitted derived value.
- **H-5 (`target_gone`)** — an intent targeting an issued-but-gone token is **acked
  `accepted: true`** and fails later as `act_result`/`failed`/`target_gone`, *after*
  `advance()`, not synchronously; an *unissued* token is refused `unknown_target` at ack.
  Demands: the daemon distinguishes the two, and never treats an ack as an existence answer.
- **H-6 (the 75% flood)** — with `restrict_change_reports = False`, run an identical script
  twice and assert `flooded.memory_count > 3 * restricted.memory_count`. Demands: the daemon
  exposes a **countable notion of "percepts that would become durable memories"** — i.e. its
  memory-formation path must be observable and countable from a test, not buried.

**The canonical end-to-end test (§10.2)** additionally demands: the daemon can act on
secondhand knowledge (`go_to` a place it has only been *told* about, with a mind-authored
`reason`), and — step 5, "what no amount of prose in this document can establish" — expose
`mind.belief(...).origin_class` and `mind.claims_witnessed(...)` as **assertable surfaces**.

**The synthesized architectural demand.** Taken together, §10 requires a daemon whose entire
vendor coupling is one injectable port, whose time source is the percept stream, whose memory
formation is countable, and whose belief store is queryable by a test. **A daemon that is
merely "testable" does not satisfy this; a daemon designed inside-out around a scriptable
vendor port does.** Any candidate language is therefore judged not on whether it *can* be
tested but on how cheaply it expresses: an injectable transport port, deterministic
script-driven time, and an assertable belief store.

---

## 3. Candidate languages — evidence

**Scope note.** The candidates are the three named in the spec (keep-Go / rebuild-Go /
rebuild-TypeScript) plus **one the evidence motivates**: rebuild-in-Java/Kotlin, because
decision-0001 fixes the body vendor as a Fabric mod (a JVM artifact) and a single-language
project is a real maintainability variable. Python is noted for completeness because it is the
best-supported Anthropic SDK by adoption, and omitting it would make the enumeration
dishonest. **No recommendation is made here.**

### 3.1 Anthropic SDK maturity — measured

All figures fetched from the GitHub and package registry APIs on **2026-08-21**.

| Language | Package | Latest | Released | Stars | Last push | License |
|---|---|---|---|---|---|---|
| Python | `anthropic` | **1.0.0** | 2026-08-20 | 3,842 | 2026-08-21 | MIT |
| TypeScript | `@anthropic-ai/sdk` | **0.120.0** | 2026-08-19 | 2,095 | 2026-08-21 | MIT |
| Go | `anthropic-sdk-go` | **v1.66.0** | 2026-08-19 | 1,179 | 2026-08-21 | MIT |
| Java | `anthropic-sdk-java` | **v2.57.0** | 2026-08-19 | 366 | 2026-08-21 | MIT |

Sources — all accessed 2026-08-21:
`https://api.github.com/repos/anthropics/anthropic-sdk-{go,typescript,python,java}` (release
tag, publish date, stars, `pushed_at`, license);
`https://registry.npmjs.org/@anthropic-ai/sdk` (dist-tags/time);
`https://pypi.org/pypi/anthropic/json` (version, `requires_python`);
`https://github.com/anthropics/anthropic-sdk-go` (README: install, `Requirements: Go 1.24+`).

**The finding that matters: all four are first-party, MIT, and were pushed within hours of
each other on the same day.** They are generated from one spec by the same Stainless
toolchain (`stainless-app[bot]` is the top contributor on both the Go and TypeScript repos —
504 contributions on TS, top contributor on Go). **SDK maturity does not discriminate between
these candidates.** Star counts measure language-ecosystem size, not SDK quality; every one of
the four ships sync/async clients with streaming, retries, and error handling.

`https://platform.claude.com/docs/en/api/client-sdks` — accessed 2026-08-21 — confirms
official client SDKs in **seven** languages (Python, TypeScript, C#, Go, Java, PHP, Ruby),
each described as providing "idiomatic interfaces, type safety, and built-in support for
streaming, retries, and error handling".

**One real asymmetry exists, and it is not in the client SDKs.** The higher-level **Claude
Agent SDK** (agent loop, tool execution, runtime — a distinct product per the same docs page)
ships for TypeScript and Python only:

- `@anthropic-ai/claude-agent-sdk` — latest **0.3.239**, published 2026-08-21, 272 versions
  since 2025-09-27 (`https://registry.npmjs.org/@anthropic-ai/claude-agent-sdk`, accessed
  2026-08-21).
- `claude-agent-sdk` (PyPI) — latest **0.2.144** (`https://pypi.org/pypi/claude-agent-sdk/json`,
  accessed 2026-08-21).
- **No Go or Java equivalent** is listed at
  `https://platform.claude.com/docs/en/api/client-sdks` (accessed 2026-08-21).

Whether that asymmetry *matters* depends on whether the mind daemon wants a framework-provided
agent loop at all — and note that promptworld I deliberately wrote its own bounded loop
(`internal/toolloop`) with doctrine the Agent SDK's generic loop does not encode ("a tool call
is a REQUEST; an event is the FACT; the gate decides"). Recorded as evidence; the weighing is
Phase 3's.

**Proof that the Go SDK is production-viable for exactly this workload:** promptworld I ships
`github.com/anthropics/anthropic-sdk-go v1.58.0` as a direct dependency `[I:go.mod]`, using it
for the full tool-use wire — `MessageNewParams`, `ToolUnionParam`/`ToolParam` with raw
`ToolInputSchemaParam`, `NewToolUseBlock`/`NewToolResultBlock` round-tripping, prompt caching
via `NewCacheControlEphemeralParam`, and `Stop` reason mapping
`[I:internal/llm/providers.go:592-800]`. This is the strongest single piece of
language-fit evidence available for any candidate, because it is a working system rather than
a README.

### 3.2 Async story

- **Go** — goroutines + channels + `context.Context` cancellation, in the stdlib. Evidence
  from I at the exact workload: per-provider worker goroutines with slot-aware admission,
  a conversation priority lane, in-flight cancellation on `agent.slept`/`agent.died`, and
  `atomic` mirrors letting workers read tick state without touching the absorb-owned replica
  `[I-wiki:llm-concurrency-leases]`, `[I-wiki:agent-mind]`,
  `[I:internal/mind/mind.go:78-146]`. The whole test suite runs under `-race`
  `[I-wiki:llm-orchestrator]`. The Go SDK takes `context.Context` on every call
  (`https://github.com/anthropics/anthropic-sdk-go`, accessed 2026-08-21) — cancellation is
  the same primitive as everything else.
- **TypeScript** — single-threaded event loop; `async`/`await` with `AbortSignal`. A daemon
  orchestrating 3–6 concurrent LLM calls at conversational cadence is squarely inside what the
  event loop handles; there is no CPU-bound work in the mind layer (the reflex/planner split
  puts all the compute in the engine). `AbortController` is the cancellation primitive. The TS
  SDK supports Node 20 LTS+, Deno, Bun, Cloudflare Workers, and the Vercel Edge Runtime
  (`https://github.com/anthropics/anthropic-sdk-typescript`, accessed 2026-08-21).
- **Java/Kotlin** — the Java SDK uses the builder pattern with `CompletableFuture` async
  (`https://platform.claude.com/docs/en/api/client-sdks`, accessed 2026-08-21); Kotlin adds
  structured-concurrency coroutines. On a modern JDK, virtual threads make blocking calls
  cheap. Untested by this project at this workload.
- **Python** — sync and async clients, `asyncio`; requires Python ≥3.10
  (`https://pypi.org/pypi/anthropic/json`, accessed 2026-08-21).

**Finding: the async requirement here is undemanding and no candidate fails it.** The load is
3–6 villagers at human conversational cadence (I measured ~16 planner calls/game-hour for
**eight** agents at 4x `[I-wiki:agent-mind]`), and the brief's real-time-only decision (#8)
deletes the entire class of latency-under-speed-multiplier problems that made I's concurrency
hard. What discriminates is not throughput but **cancellation and supersession** — the seam's
`supersedes`/`cancel` (§5.7) and the interrupt mechanism §2.8 explicitly leaves to the daemon
— and every candidate has a first-class primitive for it (`context`, `AbortSignal`,
structured concurrency).

### 3.3 Seam-contract implementation effort

Judged against §2's obligations. The contract is defined as **data, not a wire format** (§8),
so the work is: model the shapes, validate at the boundary, dispatch on runtime-declared
vocabulary, keep a durable store.

- **Go** — struct tags + `encoding/json` in the stdlib; the required "reject missing, never
  default" posture (V-5) is *against* Go's grain, since zero values are indistinguishable from
  absent fields for non-pointer types. I hit exactly this and its workaround is on the record:
  `Journal` and `Map` are **pointer** fields with `omitempty` specifically because "a value
  struct would always serialize `journal:{}` (encoding/json omitempty is a no-op on non-pointer
  structs)" `[I:internal/sim/agents.go:184+]`. So V-5 compliance in Go means explicit
  presence-checking, pointer fields, or a validation layer — known, mechanical, not free.
  Runtime-declared verb dispatch (map of string → handler) is idiomatic. Existing proof: I's
  JSON-lines-over-UDS protocol `[I:internal/ipc/protocol.go]`, `[I:internal/ipc/socket.go]`.
- **TypeScript** — the contract is specified in JSON and TS models JSON natively;
  discriminated unions on `percept_type` give exhaustiveness checking at compile time, and the
  V-3 requirement ("an unknown `percept_type` MAY be retained and MUST NOT be interpreted") is
  a `default:` arm the compiler can force. Runtime validation (V-5's "malformed, never
  defaulted") needs a schema validator, which is a dependency decision — but the type/value
  gap is exactly what such libraries exist for, and it is the most-travelled path in the
  ecosystem. Structural typing fits the "tolerate unknown fields" rule (V-1) with no
  ceremony.
- **Java/Kotlin** — Jackson/Gson records model the shapes; Kotlin's sealed classes give the
  same exhaustiveness as TS unions, and nullability is in the type system (which helps V-5 and
  the `null`-means-maximally-stale rule of §2.2). More ceremony per shape than either
  alternative. **The one structural advantage: if the daemon is in-process with the Fabric
  mod, the "seam" is a method call and T-7's in-process test vendor is trivial** — no
  serialization at all, which is exactly what §8's "defined as data, not a wire format"
  permits. That is also the biggest risk in the option: co-locating mind and vendor in one
  process makes it very easy to accidentally hand the mind a world handle, which is SI-1
  defeated silently. The contract anticipates this exact hazard (§6.1: "a vendor that lets a
  mind read the [resolution index] has reintroduced omniscience no matter how the protocol is
  worded"). Flagged as a transport-adjacent observation for Q-1; **not decided here**.
- **Python** — Pydantic is the canonical answer to boundary validation and would make V-5/V-6
  compliance close to declarative.

**Finding: no candidate has a structural difficulty with the contract; the differences are in
where boundary validation lands** (stdlib-with-care in Go, dependency-with-good-ergonomics in
TS/Python, verbose-but-checked in Java/Kotlin).

**Effort floor, independent of language.** Every candidate must build the §2.1 obligations
from scratch. Nothing in I implements this contract — I's mind read a state replica and landed
through in-process injection doors (§1.2), which is the opposite architecture. **Even
"keep Go" is a rewrite of the mind package's inputs, outputs, state, and prompt assembly**; it
keeps the language, the SDK integration, `toolloop`'s shape, and the persona firewall — not
the daemon.

### 3.4 Fake-vendor testability

Judged against §2.2's six architectural demands.

- **Go** — interfaces are satisfied structurally with no declaration, so "one injectable
  vendor port" is the language's default idiom. I already tests exactly this way: scripted
  model doubles (`scriptedModel` returning queued replies then erroring
  `[I:internal/mind/convo_test.go:17-24]`) and interface test seams declared at the consumer
  (`Submitter`, `Injector`, `SocialInjector`, all defined in `internal/mind`
  `[I:internal/mind/mind.go:27-35]`). Test-to-code ratio in the mind package is **1.6:1**
  (10,451 test lines to 6,515) and in `toolloop` **2.3:1**. The stdlib test runner needs no
  framework. **T-b (no wall clock) is a discipline, not a language property** — and I's own
  history shows the failure mode is real and catchable: its determinism harness exists
  precisely to pin it.
- **TypeScript** — mature mocking and fake-timer ecosystems; the fake vendor is an object
  literal satisfying an interface. Same structural-typing advantage as Go. The one caution:
  the ecosystem's default reflex is to mock *modules*, which would let a test reach past the
  vendor port; the contract's §10.5 prohibition ("a read API for the mind … will be copied
  into a real vendor by someone reasonably assuming the test double models the contract")
  applies with extra force where module mocking is idiomatic.
- **Java/Kotlin** — interface injection is the ecosystem's native pattern and JUnit is
  universal; in-process-with-the-mod makes the fake vendor a plain class. Most ceremony, least
  ambiguity.
- **Python** — `pytest` plus a plain class; the §10.1 pseudocode is *already written in
  Python-shaped syntax* (`.strict = True`, `v.emit(...)`, `v.advance(600)`), which is a
  presentational coincidence rather than evidence, but it does mean the harness translates
  with essentially zero interpretation.

**Finding: every candidate can express the fake vendor cleanly, and the differences are
ergonomic rather than structural.** The discriminating question is not language capability but
whether the daemon is *architected* inside-out around the port (§2.2's synthesized demand) —
which is a design decision the recommendation must state, in any language.

### 3.5 Operator maintainability

Recorded as evidence, not weighed:

- **Go** — the operator wrote and maintained ~72k non-test lines of it in promptworld I
  `[I:internal/]`, including the LLM orchestrator, tool loop, consolidation driver, and
  persona firewall this project would rebuild. Demonstrated fluency at exactly this problem
  shape is the strongest evidence available for any candidate. Single static binary, no
  runtime to install, stdlib-heavy (I's whole dep list is 7 direct requires `[I:go.mod]`).
- **TypeScript** — largest LLM-tooling ecosystem; the Agent SDK's primary language. Requires a
  Node runtime and a dependency tree; the ecosystem's churn rate is materially higher than
  Go's. No evidence in either repo of the operator running a TS service.
- **Java/Kotlin** — **the body vendor is already a JVM artifact.** Decision-0001 fixes Fabric,
  so whatever the mind is written in, **the operator is already maintaining a JVM build** for
  the mod. Recorded caveat on the JVM *version*: spec 001's prior-art pass carries JVM
  requirements only for the surrounding stack, not for the chosen Fabric path directly —
  PaperMC (the rejected option) 26.1+ needs **Java 25**, and CraftAgent (the Fabric-based
  dead-end reference mod) targeted **Java 17/21 depending on MC version**
  (`specs/001-mod-stack-decision/research/prior-art.md` — accessed 2026-08-20). The
  Fabric mod's own target JDK is unpinned by any evidence in this project and would need
  checking against the target MC version before it could weigh in a decision. One-language-
  project is a genuine maintainability argument; against it, the mod's toolchain is
  Gradle/Mixin/Yarn-mappings — a different maintenance world from a plain JVM service — and
  the mappings churn across MC versions ([[villager-brain-api]] operational note: "exact
  symbol names can shift across MC versions").
- **Python** — largest AI ecosystem; no evidence of operator use in either repo; runtime and
  dependency management are the weakest of the four for a long-running always-on daemon.

**One measured cross-cutting fact.** Whichever language wins, the mind is a **long-running,
always-on daemon holding durable per-villager state**, not a request/response service. I ran
exactly that shape (daemon + pidfile + recovery + signal handling, `[I-wiki:daemon-lifecycle]`)
and the operational notes record what it cost: a local single-operator posture, unauthenticated
protocol, darwin/arm64 target `[I-wiki:overview]`. That shape is a maintainability variable
independent of language and should be weighed as one.

### 3.6 Findings table (no recommendation)

| Criterion | Keep Go (lift I's daemon) | Rebuild Go | Rebuild TypeScript | Rebuild JVM (Java/Kotlin) |
|---|---|---|---|---|
| **Anthropic SDK** | Proven in production at this exact workload: `anthropic-sdk-go v1.58.0`, full tool-use wire + prompt caching `[I:internal/llm/providers.go]` | Same; v1.66.0 current (2026-08-19) | `@anthropic-ai/sdk` 0.120.0 (2026-08-19), 2,095★ — largest ecosystem; **also the Agent SDK's language** | `anthropic-sdk-java` v2.57.0 (2026-08-19), 366★, builder + `CompletableFuture` |
| **Async** | Goroutines/channels/`context`, proven under `-race` at this load | Same | Event loop + `AbortSignal`; no CPU-bound work in the mind layer | `CompletableFuture` / coroutines / virtual threads; untested here |
| **Seam effort** | **Rewrite of inputs, outputs, state, and prompt assembly regardless** — the existing mind is a co-process of the world (§1.2) | Stdlib JSON; V-5 needs explicit presence handling (I's pointer-field workaround is on the record) | Discriminated unions give V-3 exhaustiveness free; boundary validation is a dependency choice | Sealed classes + nullable types fit V-5/§2.2; most ceremony. In-process-with-mod makes T-7 trivial **and SI-1 easier to breach** — flag for Q-1 |
| **Fake vendor** | Structural interfaces are the idiom; I's own seams + `scriptedModel` are the pattern already `[I:internal/mind/mind.go:27-35]` | Same | Same structural advantage; caution: module-mocking reflex vs §10.5 | Interface injection is native; most verbose, least ambiguous |
| **Maintainability** | Operator wrote ~72k lines of it; single binary, 7 direct deps | Same | Largest LLM ecosystem; new runtime + dep tree for the operator; no evidence of prior operator use | **One-language project with the Fabric mod** (already JVM: Java 17–25 per prior-art) vs Gradle/Mixin/mappings churn |
| **What actually carries over** | `toolloop` shape (994 ln, sim-agnostic by construction), `persona` (475 ln, no imports), `tool`'s registry mechanism, `llm`'s provider/breaker/budget shape minus its dead `cognition` dep, consolidation's measured lessons | The same *design*, re-expressed | The same *design*, re-expressed | The same *design*, re-expressed |

### 3.7 The finding that frames the choice

**Measured, not argued: "keep Go" and "rebuild in Go" are much closer together than the
option names suggest.** §1.2 shows `internal/mind` holds a replica of world state
(`sim.NewState`, `Apply` per event), reads it directly for prompt assembly
(`s.Agents[idx]`, relationship edges, debts, norms), lands output through in-process
injection doors typed in world-engine terms (`InjectIntent(sim.InjectArgs)`,
`InjectSocial([]store.Event)`), and bakes the cast size into ten fixed-size array types
(`[sim.AgentCount]`, where `AgentCount = 8`). Under the seam, **every one of those four is
replaced**: replica → percept inbox, direct state reads → a private provenance-stamped belief
store, injection doors → `intent`/`act_result`, fixed arrays → a variable cast with
permadeath.

So the real question Phase 3 answers is not "keep the daemon or rebuild it" — the daemon as
such does not survive the seam in any language. It is: **which language carries the doctrine,
the four genuinely portable assets (`toolloop`'s shape, `persona`, `tool`'s mechanism,
`llm`'s provider layer minus `cognition`), and the operator, at the lowest total cost** —
weighed against the fact that the body vendor is already a JVM artifact and the Agent SDK is
TypeScript/Python-only.

---

## 4. Open items this assessment surfaces (for Phases 2–3)

1. **The interrupt mechanism is unowned.** `body-protocol-v0.md` §2.8 explicitly leaves it to
   the mind daemon: I's generation-counter supersession (`Agent.Generation`,
   `GenerationBumpSalience` = 9) was entangled with the cognition-horizon machinery that dies.
   The seam carries `urgency` but not formativeness; **how an `urgent` percept supersedes an
   in-flight thought is a Kithcraft design decision with no inherited answer.** Phase 2's
   latency posture touches this directly.
2. **Formativeness (mind-side salience) has no design.** I's world-side table is forbidden
   (§2.8). The mind must decide what is formative about its own life, from percepts alone.
   Bears on the memory/consolidation cost estimate in Phase 2.
3. **Persona generation vs the firewall.** Brief #5 requires generation at birth; I's personas
   were authored write-once. The firewall's structural half (no edit path) is what decision 5
   demands; the generation call itself is a new LLM event class Phase 2's routing sketch must
   account for (one-time per villager, not per-evening cadence).
4. **Transport (Q-1) stays open**, but §3.3 records one language-adjacent observation: a JVM
   daemon co-located with the Fabric mod makes the T-7 in-process test vendor trivial *and*
   makes an SI-1 breach easy to introduce silently. **Flagged, not decided**, per this spec's
   non-goals.
5. **No re-verification of [[promptworld-lineage]] is needed on the evidence above.** Every
   measurement in §1 confirms the note's summary — sim/executor/governor/guardian die
   (measured: 62% of non-test lines), doctrine transfers, code does not. The note's
   Operational notes anticipated this task by name and are satisfied by it. Phase 3 should
   re-verify its `verified_against` pin only if the recommendation adds sources.
