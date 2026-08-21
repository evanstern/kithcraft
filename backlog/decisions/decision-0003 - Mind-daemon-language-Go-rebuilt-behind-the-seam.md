---
id: decision-0003
title: 'Mind daemon language: Go, rebuilt behind the seam'
date: '2026-08-21 22:09'
status: proposed
---
## Context

Spec 004 (TASK-0004) resolves two coupled questions that block the first line of mind-layer
code: **(a)** the mind daemon's language and its reuse-vs-rebuild posture toward promptworld
I's Go daemon, and **(b)** LLM routing and the cost envelope for a 3-villager evening at 1x.
They are coupled because the routing's demands on a model client (streaming, cancellation,
prompt caching, structured outputs, per-call model selection) and its latency posture are what
a language must actually satisfy — and because what is portable from promptworld I is mostly
*mind-layer* code whose value depends on the same choice.

This decision answers (a). (b) is answered in `docs/design/llm-routing-and-budget.md` §1–§7
(3-tier routing over six LLM event classes; ≈$5.17 per demo evening, ≈$4.00 with caching;
conversation's <3 s first-token ceiling as the one hard latency constraint).

Full evidence: `specs/004-mind-daemon-routing/research/daemon-assessment.md` (Phase 1, checked
live 2026-08-21 — what I's mind-side code contains and whether it is portable, what the seam
contract demands of any daemon, and four candidate languages with cited SDK/async/effort/
testability/maintainability evidence) and `docs/design/llm-routing-and-budget.md` §8
(the recommendation, the criterion-by-criterion weighing, and the doctrine-transfer checklist).
Sits on decision-0001 (Fabric server-side mod, accepted) and decision-0002 (augmented vanilla
`VillagerEntity`, accepted), and behind the accepted `docs/design/body-protocol-v0.md` seam.

**Status: PROPOSED — pending operator ratification at the TASK-0004 PR checkpoint.** Not yet
in force.

## Decision

**Write the mind daemon in Go, as a new program architected inside-out around the vendor port,
reusing promptworld I's four portable assets as source material rather than as a codebase.**

In the spec's candidate terms this is **rebuild-Go**. Three clauses:

1. **Go** — the language, its Anthropic SDK, and the operator's demonstrated fluency carry
   forward.
2. **A new program, not a lifted daemon.** The "keep Go" option does not survive the evidence:
   Phase 1 measured `internal/mind` as a **co-process of the world engine** — it holds a
   replica of `sim.State` and applies every event to it, reads that state directly for prompt
   assembly, lands output through in-process injection doors typed in world-engine terms
   (`InjectIntent(sim.InjectArgs)`), and bakes the cast size into ten fixed-size array types
   over `sim.AgentCount = 8`. Under the seam all four are replaced: replica → percept inbox,
   direct state reads → a private provenance-stamped belief store, injection doors →
   `intent`/`act_result`, fixed arrays → a variable cast with permadeath. **The daemon does
   not survive the seam in any language.** Keep-vs-rebuild was never the real question; the
   real question was which language carries the doctrine, the portable assets, and the
   operator at the lowest total cost.
3. **Four assets are source material, read and lifted selectively — not vendored.**
   `internal/toolloop` (994 lines, sim-agnostic by construction), `internal/persona`
   (475 lines, `personas.go` imports nothing at all), `internal/tool`'s registry *mechanism*
   (its vocabulary replaced wholesale by `session_open`'s runtime manifest), and
   `internal/llm`'s provider/breaker/budget shape **minus** its dead `internal/cognition`
   dependency — the staleness router whose justification dies with real-time-only (brief #8).

### Rationale, mapped to the spec's R1 criteria

**Two of the five criteria are ties, and saying so is the finding.** Phase 1 measured all four
candidate Anthropic SDKs as first-party, MIT, Stainless-generated from one spec, and pushed
within hours of each other on 2026-08-21 — **SDK maturity does not discriminate.** Nor does
async: the load is 3–6 concurrent calls at human cadence with no CPU-bound work in the mind
layer, and every candidate has a first-class cancellation primitive. The decision rests on the
other three criteria plus the portable assets.

- **LLM-orchestration fit / SDK — tie on features, broken by proof.** promptworld I ships
  `anthropic-sdk-go` in production at *this exact workload*: full tool-use wire, prompt caching
  via `NewCacheControlEphemeralParam`, `Stop` reason mapping. The routing doc's RT-1…RT-7
  (streaming, cancellation, caching, structured outputs, per-call model selection, modest
  concurrency, truncation-aware retry) are therefore demonstrated rather than assumed in Go.
  No other candidate offers evidence stronger than a README.
- **Async — tie.** Goroutines/`context` proven under `-race` at a heavier load than the demo
  asks for, with in-flight cancellation already implemented on `agent.slept`/`agent.died`. The
  routing doc's interrupt mechanism *is* cancellation; `context.Context` on every SDK call makes
  it the same primitive as everything else.
- **Seam-contract effort — the one criterion Go loses, bounded.** V-5 ("malformed, never
  defaulted") is against Go's grain, since zero values are indistinguishable from absent for
  non-pointer types; TypeScript's discriminated unions give V-3 exhaustiveness at compile time
  and Pydantic makes V-5/V-6 near-declarative. **But harness H-1 requires seam-boundary
  validation as a distinct stage before any state mutation in every language** — the component
  exists regardless. Go's cost is that it is written explicitly rather than derived from types:
  one file of boring code, paid once. Phase 1's finding stands: no candidate has a *structural*
  difficulty with the contract.
- **Fake-vendor testability — Go, narrowly.** The harness's synthesized demand is a daemon
  designed inside-out around a scriptable vendor port; in Go a port is an interface declared at
  the *consumer* and satisfied without declaration, which is already how I tests
  (`Submitter`/`Injector`/`SocialInjector`; `scriptedModel`). Test-to-code ratio 1.6:1 in
  `mind`, 2.3:1 in `toolloop`, stdlib runner, no framework. Recorded caution against TS: the
  ecosystem's module-mocking reflex is in tension with the contract's §10.5 prohibition on a
  mind-readable test API.
- **Operator maintainability — Go, decisively. This is the largest genuine differential.** The
  operator wrote and maintains ~72k non-test lines of Go *at this problem shape* — the LLM
  orchestrator, tool loop, consolidation driver, and persona firewall this project rebuilds —
  and has run precisely this process shape before (long-running daemon, pidfile, recovery,
  signal handling). Single static binary, no runtime to install, 7 direct dependencies. No
  evidence in either repo of the operator running a TypeScript or Python service.

### Alternatives rejected, with the strongest counter-argument named

- **Keep-Go (lift I's daemon behind the seam)** — rejected because the thing it keeps is not
  portable. Reframed and rejected as an option, not as a language.
- **Rebuild-TypeScript** — the strongest counter-argument in the whole assessment, and it is
  real: the higher-level **Claude Agent SDK ships for TypeScript and Python only**, with no Go
  or Java equivalent. It does not carry because **this daemon does not want a generic agent
  loop.** promptworld I deliberately wrote its own bounded loop encoding doctrine the framework
  loop does not have — *"a tool call is a REQUEST; an event is the FACT; the gate decides"* —
  and that maps exactly onto the seam's `intent`/`intent_ack`/`act_result` split. A framework
  whose loop executes tools and returns results would have to be prevented from doing the one
  thing it exists to do, because under this contract the mind never learns the outcome of its
  own act except as a percept the vendor sends back. Rejected on **fit**, not availability.
- **Rebuild-JVM (co-located with the Fabric mod)** — one-language-project is a genuine
  maintainability argument and the body vendor is already a JVM artifact. Rejected on three
  counts, the first decisive: **(1) SI-1 breach risk** — co-location makes the seam a method
  call, which makes an in-process fake vendor trivial *and* makes handing the mind a world
  handle trivial; the contract names this exact hazard ("a vendor that lets a mind read the
  resolution index has reintroduced omniscience no matter how the protocol is worded"), and
  separate processes make the breach **structurally impossible** rather than merely forbidden.
  **(2)** "One language" overstates the saving — the mod's Gradle/Mixin/Yarn-mappings toolchain
  is a different maintenance world from a plain JVM service, so two build systems either way.
  **(3)** The Fabric path's JDK target is unpinned by any evidence in this project.
- **Rebuild-Python** — enumerated for honesty (best-supported SDK by adoption, Pydantic makes
  boundary validation near-declarative); weakest of the four on runtime and dependency
  management for a long-running always-on daemon, and no evidence of operator use.

### Doctrine-transfer checklist

The spec requires the winning choice to state how each transferred doctrine item is carried.
Full detail in `docs/design/llm-routing-and-budget.md` §8.3; the summary is that **most of the
doctrine is already carried by the protocol rather than by any code decision** — which is
itself evidence that the language choice is a smaller decision than it looked.

| Doctrine item | Carried how |
|---|---|
| **Event-sourced memory** | **Reimplement to contract** (~400 lines): append-only log, immutability enforced in the schema not in convention, state as a reducer. Deliberately **not** carried: I's world-event vocabulary (the mind never sees it under SI-1) and the `log_format_version` migration chain, whose determinism-for-replay justification dies with I. |
| **Reflex / planner split** | Reflex half **not carried** — decision-0002's `Brain<E>`/`Schedule`/POI stack *is* the reflex half now, and AR-4 forbids mind-side spatial arithmetic. Planner half: **port `toolloop`'s shape** (its REQUEST/FACT/gate doctrine maps one-to-one onto `intent`/`intent_ack`/`act_result`) plus `tool`'s registry mechanism, with the verb vocabulary **already in the protocol** via the runtime manifest. |
| **Salience / consolidation / situated memory** | Three-way split. World-side salience **forbidden** by the contract (no `salience`/`importance`/`weight` field on any percept). Consolidation: **port the machinery shape and its measured lessons** — nightly ledger, ordinal `m1..mN` prompt convention, `(tick, hash)` durable identity, the truncation lesson, "transport failure lands no marker". Formativeness: **new design** — v1 has no scoring pass; a deterministic admission gate feeds one Opus-tier nightly consolidation. Situated `Where`/`Why` **already in the protocol** (vendor composes the place descriptor; mind supplies the intent reason). |
| **Epistemic hygiene** | **Already in the protocol, not in code.** `direct_perception(origin)` is a contract obligation with `DIRECT_ORIGINS` defined in-contract, a MUST NOT against a `direct` boolean on the wire, and harness H-3 proving the classifier ignores prose; coerce-never-reject is RM-2/RM-3. Mechanisms were always tiny (8 and 30 lines) — the value was the knowledge, now binding text. **Reimplement RM-1…RM-7 to the contract; port nothing.** |
| **Persona firewall** | **Port `persona`** (475 lines, no imports) — the cleanest carry. The doctrine is the two-half design: structural impossibility (written once at mode 0444, no post-genesis write path exists) plus a model-free validator (anchor echo + drift lexicon), so rejection is a testable 100% guarantee a second model call would downgrade to a guarantee about a distribution. Brief #5 changes the *source* (generated at birth) not the firewall; the honest limit transfers with it (the lexicon catches *stated* drift). |

### What would change this

Named so a future reader can check rather than re-argue: a second maintainer whose language is
not Go (the decisive criterion is a fact about this operator, not about Go); the Agent SDK
becoming load-bearing (test: does the framework let a gate sit between the model's tool call
and the fact the mind may believe?); transport resolving to in-process-with-the-mod; the
percept surface growing far beyond the contract's current shapes (which un-bounds Go's V-5
cost); or `anthropic-sdk-go` falling materially behind on streaming, caching breakpoints, or
structured outputs. None is true today.

## Consequences

- **The demo is a two-language, two-artifact project:** a Fabric mod jar and a Go daemon
  binary, started independently. TASK-0006's decomposition **splits at the seam** — every
  deliverable task lands wholly on one side, and a task spanning both has smuggled coupling
  across SI-1 and should be split. The mind daemon is its own module with its own `go.mod`;
  it does not live inside the mod's Gradle build.
- **TASK-0006 must not budget a "port the daemon" task — there is no such task.** Exactly two
  porting tasks exist (the `toolloop` bounded-loop shape onto intent/ack/act_result, and
  `persona`'s two-half firewall); everything else in the mind is written fresh against the
  contract. One additional Go-specific task is scheduled up front: the boundary-decode
  component (presence-checked decode → validate → mutate, in that order, per harness H-1).
- **The fake-vendor harness is a Go test suite and a first-class demo-plan task.** It is how
  mind work proceeds *before the mod exists*, and it gates the mind side of every demo beat.
  The vendor-facing port is an interface declared at the consumer.
- **Mind-restart independence becomes a demo acceptance check**, not an aspiration: "restart
  the daemon mid-session and the villagers keep their memories" is directly testable once mind
  and vendor are separate processes.
- **Transport (seam Q-1) is narrowed, deliberately and one-way:** a Go daemon against a JVM mod
  **forecloses the in-process option**, so Q-1 is a choice among real wires (UDS, TCP, stdio),
  not between a wire and a method call. This is a narrowing of the option set, not an answer —
  the wire remains open for the spec 002 successor. The narrowing is intentional: it makes an
  SI-1 breach structurally impossible rather than merely forbidden.
- **The routing/budget half of TASK-0004 narrows TASK-0006 independently of language:** six LLM
  event classes over three tiers (persona genesis and nightly consolidation on Opus 5;
  deliberation, job-board, and conversation on Sonnet 5; ambient lines on Haiku 4.5), with
  pathing/building/schedules/panic staying engine-side at zero cost. Plan against a **ceiling
  of ~$20 per demo evening** (baseline $5.17, ≈$4.00 cached; realistic worst case at 6 chatty
  villagers $16–17). Cost is **not** the binding constraint — the whole tier ladder is worth
  about $8 on the demo — so effort goes to conversation latency (<3 s to first token) and
  episodic-buffer growth instead.
- **Ratification: pending.** Proposed by TASK-0004; becomes settled fact for TASK-0006 (and for
  any mind-layer implementation task) only once the operator ratifies at the PR checkpoint. Per
  decision-0001/0002 precedent, status flips by direct edit — the backlog CLI's decision
  command supports only `create`, no edit verb.
