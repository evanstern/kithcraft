# Demo build plan — "one real evening"

**Spec:** 006-evening-demo-build-plan · **Phase:** 1 (decomposition) · **Board task:**
TASK-0006 · **Milestone:** m-0.

**What this is.** The decomposition of milestone m-0 into deliverable, one-PR-shaped board
tasks: sixteen of them across four groups, with dependencies, per-task scope boundaries,
suggested tiers, and the spell-breaker checks each must satisfy. Phase 2 creates these on
the board; this document is the plan of record they are created from.

**What this is not.** No implementation, and no re-litigation of a ratified decision. The
plan *consumes* five settled surfaces and treats each as fact:

| Surface | What it fixes for this plan |
|---|---|
| `docs/design/kithcraft-brief.md` (ratified 2026-08-19) | Thesis, the two load-bearing constraints, the demo's beats, the spell-breakers. |
| decision-0001 | Fabric server-side mod. |
| `docs/design/body-protocol-v0.md` (accepted) | The seam contract: five invariants, three surfaces, four core verbs, the fake vendor, Q-1…Q-6. |
| decision-0002 + `entity-implementation-comparison.md` | Augmented vanilla `VillagerEntity`; a **bounded ~4-injection Mixin surface**; no client jar; cast distinctiveness via profession × biome variant. |
| decision-0003 + `llm-routing-and-budget.md` | Go daemon rebuilt behind the seam; six LLM classes over three tiers; **the decomposition splits at the seam**; two porting tasks and only two; no "port the daemon" task; transport narrowed to real wires. |
| `docs/design/death-mechanics.md` | Admitted/suppressed causes, grave + belongings + POI grief period, memory carry — and §6.2's five open items, which land in §2 below. |

---

## 1. The shape of the decomposition, and the razor

**Every task lands wholly on one side of the seam.** Decision-0003's consequence, taken
literally: the demo is a two-language, two-artifact project (a Fabric mod jar, a Go daemon
binary), and a task spanning both has smuggled coupling across SI-1 and is split. The four
groups fall straight out of that rule plus the two things that belong to neither side:

- **Seam (S1–S2)** — the contract's own remaining work: the wire, and the conformance
  harness that proves the rules no compiler can.
- **Mind (M1–M7)** — Go. Testable against the fake vendor, *before the mod exists*.
- **Vendor (V1–V5)** — Java/Fabric. Testable on a dev server.
- **Integration (I1–I2)** — the two artifacts run together, and the evening itself.

**The razor for granularity: does this PR give a reviewer a real decision to make?** Too
big and the PR recreates the undifferentiated demo; too small and it is a diff for its own
sake. Sixteen tasks is where the material landed, slightly above the 8–15 the spec
anticipated, and the two places the count could have come down are both places the razor
says no:

- **V1 + V2 (mod skeleton + vendor conformance) stay split.** Merged, the PR is the Gradle
  and Mixin toolchain, the transport client, the manifest, the token registry, nine percept
  types and four verbs — unreviewable as one decision. Split, V1 carries a real and narrow
  one: *does the handshake describe the vendor rather than the world* (L-7 — SI-1 defeated
  at `session_open`, before a percept crosses) and *does the token registry survive a
  restart* (the vendor's hardest obligation, §6.1).
- **M1 + M4 (daemon skeleton + model client) stay split.** They are independent axes — seam
  plumbing and LLM plumbing — that happen to live in one binary. Merged they are one large
  PR with two unrelated approval questions.

And one place the razor says **merge**, against the instinct to split: **V4 keeps the job
board and the blueprint build together.** They are one demo beat ("the player posts a
blueprint; one villager builds it while the player builds alongside"). Split, the first PR
delivers a board that is read and nothing built — a deliverable with no observable payoff,
which is exactly the diff-for-its-own-sake the doctrine rejects.

**Tier floor is `sonnet` across all sixteen, with one proposed escalation.** Every task here
is greenfield work against a written contract — the judgment calls are already made by the
five surfaces above, which is the sonnet scope exactly. `haiku` fits nothing: its scope is
"work written to a sibling standard," and until M1/V1 land there are no siblings, and its
200K window is under-sized for tasks that must hold the protocol doc open. The single
proposed escalation is **S1 (transport)** — a decision the spec deliberately does not settle
— and per doctrine taking it is the operator's checkpoint at runbook sign-off, not an
implementer's call. **V5 (death/danger) carries a named escalation trigger** rather than a
tier bump: it is the one task standing on unverified engine surface.

---

## 2. Scoping rulings this plan owes

Five surfaces flagged open items *for this task*. Ruling them here means no downstream task
re-derives them, and each ruling names the task that implements it.

| # | Open item | Ruling | Lands in |
|---|---|---|---|
| **R-1** | **Nine dusks or one?** (routing §7.1 — the vanilla cycle gives ~9 day/night cycles in 3 real hours) | **Accept nine.** Lengthening the in-game day makes the demo unrepresentative of the game anyone will actually play, and "the evening" is a *recording* framing, not a build constraint — I2 cuts its highlight from the dusk that landed best. Nine cycles also exercise consolidation 27 times, which is the machinery most likely to break and least likely to be exercised by a single-cycle run. Routing already models this case, so no number moves. | I1 (config), I2 (run) |
| **R-2** | **Raids in v1-demo scope?** (death §6.2.2) | **Out of scope, and nothing to build.** Raids are player-triggered (Bad Omen / attacking a patrol); a demo run that does not attack a patrol never sees one. No task budgets for raid content and no task suppresses it — an absence with a cause, not an oversight. | nobody (recorded) |
| **R-3** | **Grief period before bed/job-site reassignment** (death §6.2.3 — deliberately unnumbered there) | **One full in-game day/night cycle (~20 real minutes at 1x).** Long enough that the player walks past the empty bed in daylight at least once — which is the entire point of not backfilling — and short enough that a nine-cycle evening does not end permanently down a bed. Exposed as config, not a constant. | V5 |
| **R-4** | **POI re-claim lag after `releaseAllPois()`** (death §6.2.4 — unverified) | Verification, **before** implementing the hold. If vanilla already lags, the task shrinks to tuning it; if re-claim is instant, the hold is real work. Do not build against the assumption. | V5 (precondition) |
| **R-5** | **Siege Mixin-suppression point + whether a 3-villager cast triggers one at all** (death §6.2.1) | Verification-first, same posture as R-4, and **suppress regardless of eligibility.** If a 3-villager household turns out never to qualify, the suppression is cheap insurance against a cast that grows to the brief's 6; if it does qualify, it is load-bearing. Either way the answer is known before the code. | V5 (precondition) |
| **R-6** | **Demo danger tuning** (death §6.2.5) | **Config knob, default off.** The neglect-not-luck posture wants real danger in the demo; the knob exists only for a recorded take that cannot afford to lose a cast member, and using it is a per-run choice, never the default. | I1 |
| **R-7** | **Conversation initiation** (routing §7.2 — what makes two villagers decide to talk) | **Engine-side pair formation**, as a scheduled Activity: at dusk, villagers path to a shared gathering place, and the vendor emits the pairing signal **~10 s ahead of arrival** so the mind can pre-generate the opening turn (routing §5.2 lever 2). The mind never decides *whether* a conversation starts; it decides what is said. | V3 (signal), M6 (consumes) |
| **R-8** | **Hearing's engine hook in the first vendor** (protocol Q-2 — no verified vanilla surface) | Verification work inside the percept surface, not a separate task. `sound` is an **optional** percept type (§6.2); if no hook verifies, the vendor declares it unsupported and the demo loses nothing — the four-type floor does not include it. | V2 |
| **R-9** | **A dead villager's own mind process** (death §3, flagged to the seam's owner and never ruled) | **Archive, do not terminate.** The dead mind opens no new session and its body token is retired and never reissued (death §6.2, token discipline), but its durable log survives — because survivors' memories cite it, and because "stories told about them" (brief #4) is the whole texture permadeath exists to produce. Termination would make the dead unreconstructable for no saving. | M7 |

Two further inheritances, recorded so no task re-decides them: **the mod plans no client
jar** (decision-0002 — "a friend can drop in" is satisfied by the entity choice alone), and
**cast distinctiveness is profession × biome variant plus names, dialogue and job-board
role** — no custom skins, no resource-pack spike inside m-0.

---

## 3. The task graph

Labels are plan-local (`S1`, `M1`, …); board ids are assigned at Phase 2. Dependencies are
by label. "Done proves" is written to be checkable: mind-side against the fake vendor,
vendor-side on a dev server, integration against a running pair.

### 3.1 Seam

---

#### S1 — Decide the seam transport and pin the wire

> As a **future implementer**, I want the mind↔vendor wire chosen and its framing pinned by
> fixtures, so that the Go and Java sides can be built independently without discovering
> they disagree at first contact.

**Scope.** Choose among **real wires only** — UDS, TCP, stdio — per decision-0003's one-way
narrowing (a Go daemon against a JVM mod forecloses in-process, deliberately, because it
makes an SI-1 breach structurally impossible rather than merely forbidden). Weigh against
T-1…T-7 (push not pull; ordered-or-reorderable per body; long-lived sessions; **mind restart
independent of vendor restart**; message-oriented with a schema the fake vendor can produce
without an engine; backpressure that sheds `background` only and never splits an
`observation`; process-separable but not process-required). Deliverable: a decision record,
a framing/serialization spec as the spec-002 successor, and **golden message vectors** — one
fixture per percept type, per intent shape, and the `session_open` handshake — that both
implementations run against.

**Done proves.** The wire is decided with T-1…T-7 answered one by one; the golden vectors
exist and are language-neutral; a trivial Go encoder and a trivial Java decoder each round-trip
every vector. Nothing else is built.

**Depends on.** Nothing. This is the contract-shaped head of the graph.

**Tier.** **`opus` — proposed escalation.** A decision the spec constrains but does not
settle, directly analogous to the TASK-0002 and TASK-0004 escalations. Taking it is the
operator's checkpoint at sign-off; ratifying the decision is an operator checkpoint
regardless of tier.

---

#### S2 — Fake body vendor and the protocol-rule harness

> As a **future implementer**, I want the seam's unprovable rules turned into failing-on-
> violation tests, so that a rule that looks like a stylistic preference in prose cannot be
> deleted by the next refactor.

**Scope.** `FakeVendor` per protocol §10.1 — manifest, open/close, `emit`, `advance`,
`.acts`, `resolve`, and the two deliberate misbehaviour switches (`strict`,
`restrict_change_reports`). The six named tests: **H-1** (malformed rejected, never
defaulted, before any state mutation), **H-2** (unknown or absent origin classifies
secondhand), **H-3** (the classifier is pure — a percept whose prose says "I saw this myself"
is still secondhand), **H-4** (`direct` never appears on the wire), **H-5** (`target_gone` is
the only non-existence channel; an *unissued* token refuses at ack, a *known-but-gone* one
must accept and fail after a walk), **H-6** (the 75% flood, reproduced on purpose:
`flooded.memory_count > 3 × restricted.memory_count`). Plus §10.5's scope discipline, enforced:
no autonomous behaviour, **no read API for the mind**, no capability the real vendors lack.

**Done proves.** All six tests green, and each one red when its rule is lifted. H-6 prints
the ratio. The fake vendor exposes no method by which a mind can learn world state without
acting.

**Depends on.** M1 (the vendor port is declared at the consumer), M2 (H-6 counts memories;
the canonical end-to-end asserts a belief's origin class).

**Tier.** `sonnet` — protocol §10 specifies these tests in detail; this is execution against
a written standard.

**Spell-breaker check.** None directly. But §10.5's "no read API" is the **minds-are-others**
constraint's structural defence: the fake vendor is the single most convenient place in the
codebase to add `vendor.things_near(body)` for a test's benefit, and doing so builds an
omniscience bug into the reference implementation.

---

### 3.2 Mind side (Go)

---

#### M1 — Mind daemon skeleton: process, vendor port, boundary decode, session lifecycle

> As a **future implementer**, I want a daemon that can hold a session and refuse a malformed
> percept before it touches any state, so that every later mind component is built on a
> boundary that cannot be talked past.

**Scope.** Its own module and `go.mod`, outside the mod's Gradle build. The vendor port as an
interface **declared at the consumer**. The boundary-decode component decision-0003 schedules
explicitly, in harness H-1's order: presence-checked decode → validate → mutate, never
interleaved. V-1…V-6 (ignore unknown fields; fall back on unknown enum values; retain but never
interpret an unknown `percept_type`; refuse an unknown verb; **a missing required field is
malformed, not defaulted**; unrecognized or absent `origin` classifies secondhand). Session
lifecycle: `session_open` handshake and manifest ingest, continuity handling (`body_continuous`,
and a gap reported as a gap — never backfilled), `session_close`. Percept ingest: dedup by
`percept_id`, `seq` gap detection. Intent bookkeeping: the pending set, `supersedes`, matching
`act_result` to `intent_id`, `cancel`. A minimal in-test double sufficient to drive this;
S2 grows it into the conformance harness.

**Done proves.** The daemon binary starts, opens a session against the double, ingests a
scripted percept stream with duplicates and a `seq` gap, and emits intents. A percept missing
`provenance` mutates nothing. An `origin` value from a future minor version classifies
secondhand. Restarting the daemon mid-session re-opens with continuity and does not invent
what happened in the gap.

**Depends on.** S1.

**Tier.** `sonnet`.

---

#### M2 — Event-sourced memory, the belief store, and the episodic admission gate

> As a **villager**, I want to know only what I saw or was told, with provenance, so that what
> I say at dusk is honestly mine and not something I was handed.

**Scope.** Reimplement event-sourced memory to the contract (~400 lines per decision-0003):
append-only log, immutability **enforced in the schema, not in convention**, state as a
reducer. Deliberately not carried: promptworld I's world-event vocabulary (the mind never sees
it under SI-1) and its `log_format_version` migration chain (whose determinism-for-replay
justification died with I). The private, provenance-stamped map — PM-1's spine, the reason two
villagers see different worlds — kept distinct from the vendor's resolution index, which the
mind never reads. RM-1…RM-7 reimplemented to the contract, porting nothing: witnessed claims,
coerce-never-reject, secondhand never beating fresher firsthand, read-time confidence and
freshness as arithmetic on `world_time`, and **only a correction, a death, or a witnessed
removal deletes a fact — time alone never does**. The deterministic **episodic admission gate**
(routing §6.3): admit on urgency ≥ `notable`, any percept involving another body or the player,
any `act_result` on an intent the mind authored a `reason` for, any `told_fact` or `text`, any
first sighting of a `kind` or `place`; drop repeated `background` sightings of already-known
things. And the one long-run instrument: **E6 input tokens per villager over time**, not dollars.

**Done proves.** The canonical end-to-end (protocol §10.2) against the fake vendor, including
its step 5: a mind told about the orchard cannot durably claim it *saw* apple trees there. The
admission gate's instrument reports buffer size per villager-day. An attempt to mutate a logged
event fails at the type level, not at review.

**Depends on.** M1.

**Tier.** `sonnet`.

**Spell-breaker check — none, but the constraint bites.** **Minds-are-others** is structural
here: the belief store has no external write path. Nothing outside the mind — not the vendor,
not the player, not a debug command — may author a belief.

---

#### M3 — Persona genesis and the persona firewall

> As a **player**, I want the villagers to be people I did not write, so that their company
> counts as company.

**Scope.** E1 on **Opus 5**, three calls total, pre-session and unbounded in latency. Port
`persona`'s two-half design — the cleanest carry in the project: **structural impossibility**
(written once at mode 0444; *no post-genesis write path exists*) plus a **model-free validator**
(anchor echo + drift lexicon), because rejection is a testable 100% guarantee only when no
second model call is involved. Generated names, values and **endogenous desires** for the
three-villager cast, weirdness dial conservative (brief #5). Cast distinctiveness pairs each
persona with a profession and biome variant (decision-0002) so the fiction and the body agree.

**Done proves.** Three personas generated and written at mode 0444. A test that attempts a
post-genesis write **fails to find a code path to attempt**, not merely fails at runtime. The
validator rejects a drifted persona without a model call. Personas survive a daemon restart
and re-bind to the same bodies.

**Depends on.** M1, M4.

**Tier.** `sonnet` — decision-0003's doctrine-transfer checklist settles the design; this is
the port plus the genesis prompt.

**Spell-breaker check — politeness-policing.** The genesis prompt must not generate a
moralizing persona template. A villager may hold values the player offends, and may say so;
a villager whose *generated character* is "corrects the player's manners" has shipped the
spell-breaker at birth, before any conversation code runs. The drift lexicon is the place to
catch it.

**Constraint — minds-are-others, in its strongest form.** This task is where the brief's
load-bearing constraint is either structural or decorative. Honest limit, inherited and
carried: the lexicon catches *stated* drift.

---

#### M4 — Model client, per-class prompt assembly, tier routing, and instrumentation

> As an **operator**, I want every model call routed by class, cached at a genuinely stable
> prefix, and counted from the first call, so that the demo's cost and latency assumptions are
> measurements rather than estimates.

**Scope.** `anthropic-sdk-go` against RT-1…RT-7: streaming with usable first-token latency,
clean in-flight cancellation, **prompt caching with explicit breakpoint placement**, structured
outputs for E2/E3/E6, per-call model selection (six classes across three tiers in one process),
modest concurrency (3–6, peaking at one per villager), retry with backoff and a
**truncation-aware ceiling on E6** — budget 4,096 against an expected 1,200 and never set E6's
`max_tokens` near its expected output. **Per-class prompt assembly as a first-class component**,
not concatenation at the call site: six classes × the stable/variable split (routing §2.3) *is*
the caching design and the cost model. Per-class call and token accounting from the first call.

**Done proves.** Each class routes to its declared tier (E1/E6 Opus 5, E2/E3/E4 Sonnet 5, E5
Haiku 4.5). A test asserts the **stable prefix is byte-identical across calls for a given
villager and class** — specifically that no date, day counter, or timestamp is rendered into
it, which silently destroys 23% of the bill *and* part of E4's latency budget at once.
Cancelling an in-flight call terminates it promptly and cleanly. Per-class counters report at
session end.

**Depends on.** M1.

**Tier.** `sonnet`.

---

#### M5 — Deliberation and the job-board decision (E2, E3)

> As a **villager**, I want work orders to arrive on top of a life I am already living, so
> that taking one — or not — is a choice I made rather than a command I executed.

**Scope.** Port `toolloop`'s bounded-loop shape — *a tool call is a REQUEST; an event is the
FACT; the gate decides* — onto `intent`/`intent_ack`/`act_result`, which is the one-to-one map
decision-0003 identified. Verb vocabulary from the runtime manifest, not from a compiled-in
list. E2 routine deliberation at schedule transitions and open choices (8/villager/cycle);
E3 job-board deliberation on a `text` percept with `origin: read`, carrying its own context
shape (board contents, other villagers' claims, standing relationship to the player, current
commitments). The **urgency interrupt** exactly as routing §5.5 states it: an `urgent` percept
**cancels the in-flight deliberation**, **does not itself trigger a model call**, and
**enqueues one deliberation** whose context includes it — because the body's reflex has already
run. Memory window K=10 situated, top K−2 by recency-decayed weight plus **2 seeded serendipity
picks from the older half** (the thing that stops a villager's context collapsing onto its five
loudest days). Structured output, so an intent is a value rather than a text to interpret.

**Done proves.** Against the fake vendor: a scripted board posting yields a claim-or-decline
intent carrying an **authored `reason`** (§5.2 requires the mind to have a why). A decline is
reachable and reads as this persona's decline, not a generic refusal. An `urgent` percept
mid-deliberation cancels the call and produces exactly one follow-up deliberation — not three,
and not zero. No intent names a target by description ("the nearest bed"); every target is a
token the mind was given.

**Depends on.** M2, M4.

**Tier.** `sonnet`.

**Spell-breaker check — micromanagement.** *Reluctance is the product* (brief #6, routing E3),
but reluctance is not non-compliance forever: a villager who never takes a posted job turns the
board into a chore the player must keep re-issuing, which is the failure mode inverted rather
than avoided. The check: across a scripted evening's postings, work gets done without the player
re-posting, and the refusals that do occur are legible as *this* villager's.

**Spell-breaker check — politeness-policing.** A refusal must be grounded in the villager's own
wants, commitments or relationship — never in the player's conduct. There is no compliance gate,
no cooldown, and no "you were rude to me so I won't work" mechanic anywhere in this task. The
player can be a jerk; that costs them a relationship, not an API.

---

#### M6 — Dusk conversation and the ambient pool (E4, E5)

> As a **player**, I want to overhear my neighbours talking about the day, the work, and me, so
> that the base I built stops being a place and starts being a household.

**Scope.** E4 conversation turns on **Sonnet 5**, under the design's only hard latency ceiling:
**< 3 s to first token** — streaming on, `effort: low`, **thinking off**, cached prefix,
`max_tokens` ~300. Extended thinking, which every other class benefits from, is a direct tax on
the one class that cannot pay it, and Opus is *too slow* here: this tier is constrained from
above. Consume V3's pair-formation signal (R-7) to **pre-generate the opening turn during the
walk**, so the scene opens instantly. The interlocutor model — who this is, what I think of
them, shared history — as its own context slice. E5 ambient on **Haiku 4.5**: one batched call
per villager per in-game day producing ~8 persona-flavoured lines, served from the pool in
< 200 ms and **refreshed daily**; a remark about something specific escalates to a live call
instead of drawing from the pool.

**Done proves.** Against the fake vendor: a scripted dusk exchange between two minds produces a
multi-turn conversation about the day, the work and the player, with **measured** first-token
latency < 3 s. The pool serves in < 200 ms and does not repeat a line within a cycle. A
conversation **ends** — it has a termination condition, not a turn cap that leaves two villagers
mid-sentence.

**Depends on.** M2, M4.

**Tier.** `sonnet`.

**Spell-breaker check — tedium.** This is the task where tedium lives. Three concrete checks:
lines do not repeat within a cycle (the pool refreshes daily for exactly this reason); a
conversation reaches a natural end rather than looping; and the ambient stall-line ("Hm.") is
used sparingly, because a villager who says "Hm." before every sentence is worse than one who
pauses. A villager the player learns to walk past is the spell broken.

**Spell-breaker check — politeness-policing.** A villager may resent the player, grumble about
them at the fire, and say so to their face. What it may not do is lecture, moralize, or gate
anything on the player's conduct. Grumbling at the fire is relationship; a reprimand is a
politeness simulator.

---

#### M7 — Nightly consolidation, and how the dead stay conversationally alive (E6)

> As a **villager**, I want to wake up having kept what mattered about yesterday — including
> who is no longer here — so that a day with the player accumulates into a history instead of
> evaporating.

**Scope.** E6 on **Opus 5**, triggered by the sleep event and timed against `world_time`, never
a wall clock (harness T-b). Port the machinery *shape* and its measured lessons: the nightly
ledger, the ordinal `m1..mN` prompt convention (memories have no stable IDs and slice indexes
are unstable, so the prompt convention *is* the identity mechanism, with accepted references
mapped back to durable `(tick, hash)` pairs), and the rule that **a transport failure lands no
marker** — the night is retried, not lost, and a villager waking with yesterday undigested is
recoverable and invisible. Runs inside the ~5.8-minute sleep window; **not** on the Batch API,
where a digest arriving after the villager has woken is a correctness problem rather than a
latency one. v1 has **no formativeness scoring pass** — the admission gate decides eligibility,
E6 decides what mattered. The death carry (death §3): a recent death surfaces disproportionately
in dusk conversation, then **fades in retrieval frequency** — never silently deleted (RM-7).
Implements ruling **R-9**: a dead villager's mind is archived, opens no new session, and its
body token is retired.

**Done proves.** Against the fake vendor: a scripted day's admitted buffer consolidates into a
digest whose references resolve back to `(tick, hash)` pairs. A consolidation failed mid-call
leaves **no marker** and is retried on the next attempt. A witnessed death is retrieved at high
frequency in the following cycle's conversation context and at lower frequency two cycles later,
and is still present — not deleted — well after that. A dead villager's log is readable; no
session opens for it.

**Depends on.** M2, M4.

**Tier.** `sonnet`.

---

### 3.3 Vendor side (Java / Fabric)

---

#### V1 — Fabric mod skeleton, vendor session, capability manifest, token registry

> As a **future implementer**, I want a mod that can hold a protocol session and hand out
> tokens that still mean the same thing after a restart, so that the mind's memories survive
> the world they were formed in.

**Scope.** The mod jar and its dev-server setup against the target Minecraft version — and
**re-verify the Yarn mappings and version-dependent facts here**, since `villager-brain-api`'s
symbol names were checked at yarn-1.21.3+build.1 and routing's A-2 daylight arithmetic is
flagged "verify against the target version before building." The transport client per S1. The
`session_open` handshake: `time_unit` (a declared unit, **never ticks**), continuity,
capabilities — declared `percept_types` (at minimum the four-type floor: `act_result`,
`observation`, `sighting`, `speech`), origins, verbs with target shapes, `salient_kinds` in
role-annotated form, bearings, distance bands. The **token registry** — `body`/`place`/
`thing_id`/`kind` → referent, persisted across sessions: the vendor's hardest obligation, and
tokens are never reused. No client jar (decision-0002).

**Done proves.** On a dev server: the mod loads, a villager's session opens against the daemon
(or a stub mind) and the handshake round-trips S1's golden vectors. **The manifest is identical
for every body and does not vary with world state** — the L-7 test, because a vendor populating
`salient_kinds` from what is nearby has made `session_open` a "what is around me" query and
defeated SI-1 before the first percept. Tokens issued before a server restart still resolve to
the same referents after it.

**Depends on.** S1.

**Tier.** `sonnet`.

---

#### V2 — Body-vendor conformance: percepts out, intents in

> As a **villager**, I want the world to tell me only what my body could actually have
> perceived, so that everything I later believe has an honest origin.

**Scope.** The two halves of the seam surface, which cannot be split because `act_result` is
both. **Perceive:** the four-type floor plus whatever else is declared — `sighting` (with
`doing`, prose-only), `observation` (with its `vocabulary` — the kinds *scanned*, which for a
3-D volume is necessarily a subset of the manifest, and without which an absence claim has no
scope), `speech`, `act_result`, and optionally `sound` (**R-8**: verify a hearing hook; declare
it unsupported if none verifies), `told_fact`, `text`, `self_state`, `change_report`.
Provenance **stamped at emission** from the closed origin vocabulary. Urgency bands — and **no
`salience`/`importance`/`weight` field may exist**, world-side salience being forbidden. The
`change_report` **delivery restriction**: never to the body that caused the change or watched
it happen. The abstraction rule throughout: opaque `kind` tokens, meaning as `roles` plus
prose-only `descriptor`, space as place tokens plus coarse bands (**no coordinates, no
arithmetic**). `nearest_hostile` exposed as the free, already-computed danger signal
decision-0002 confirmed. **Act:** the four core verbs — `go_to` (targeting a body resolves to
its **last-seen** place, never its live position), `speak`, `attend`, `wait` — plus `cancel`
and the intent/ack split: the ack acknowledges receipt only, and what happened returns as an
`act_result` percept. `unknown_target` refuses only an **unissued** token; a known-but-gone
referent MUST be accepted and fail with `target_gone` after a walk, because a synchronous
"gone?" answer is an existence oracle a mind can poll without moving.

**Done proves.** On a dev server: a villager body emits provenance-stamped percepts a mind
ingests without rejection. An `observation` yields a falsifiable absence claim. **No
`change_report` reaches the actor or a witness** — the rule with a $1.30/evening price tag that
grows with the cast. The four verbs execute and each returns exactly one `act_result`. Protocol
§12's six leak passes run clean over captured payloads: no engine-native type, identifier, or
coordinate convention in any message shape.

**Depends on.** V1.

**Tier.** `sonnet`.

---

#### V3 — The augmented villager: brain, schedule, cast, and dusk pair formation

> As a **player**, I want three named neighbours who live a full day without me, so that the
> village is somewhere I arrived rather than something I have to run.

**Scope.** `Schedule` get/set for wake/work/socialize/sleep on the vanilla `Brain<E>`
substrate; `Activity` registration and task-list assignment; memory modules; POI bed and
job-site claim, sleep pathing, door use — all inherited free per decision-0002. The bounded
Mixin surface, **enumerated and no larger**: up to three task-list overrides suppressing
breeding, gossip and iron-golem summoning (the conversion-cancel injection belongs to V5).
Cast setup: three named villagers distinguished by profession × biome variant plus nameplates.
The dusk **pair-formation Activity** implementing R-7: villagers path to a shared gathering
place at dusk, and the pairing signal is emitted **~10 s ahead of arrival** so M6 can
pre-generate. And the rule that makes the whole latency posture survivable: **the scheduled
activity keeps the body busy while the mind thinks** — a villager standing motionless awaiting
a response has converted a 20-second thought into a 20-second bug, and no tier change fixes it.

**Done proves.** On a dev server, three named villagers run a **full day/night cycle
unattended**: wake, work, socialize at dusk, and sleep in their claimed beds. No breeding
occurs, no gossip-driven golem is summoned, and no player action is required at any point. Two
villagers converge on the gathering place at dusk and the pairing signal precedes arrival by
~10 s. With a deliberately stalled mind, bodies keep moving.

**Depends on.** V1.

**Tier.** `sonnet`.

**Spell-breaker check — micromanagement.** "A full cycle unattended" is not a nice-to-have in
the scope boundary; it *is* the spell-breaker check, made testable. The moment keeping a
villager fed, escorted, or on-task requires the player, the demo has shipped the failure mode
the brief names.

---

#### V4 — The job-board book and the blueprint build

> As a **player**, I want to post a blueprint on a board and have a neighbour take it up and
> build it beside me, so that giving work feels like asking a person rather than issuing a
> command.

**Scope.** The demo's centrepiece and the v1 soul of the order interface (brief #7), kept
whole. The diegetic in-world board — a book/lectern the player writes a blueprint into,
readable by villagers, with other villagers' claims visible on it so they can argue about and
prioritize orders. Reading it emits a `text` percept with `origin: "read"`, carrying the
blueprint **as text**: protocol Q-6 records that v0's `target` shape is deliberately thin and
that the read channel is how a blueprint crosses in v0 — so this task does not extend the
protocol to carry a structured build plan. Engine-side build execution for a claimed job:
block placement against the blueprint, material sourcing, interruptible by schedule transition
or danger and resumable afterwards. The mind is consulted on **taking** the job and at decision
points; the building itself is engine-side (AR-4 — the mind names a token, the engine resolves
it).

**Done proves.** On a dev server with a mind attached: the player writes a blueprint into the
board; a villager walks to it, reads it, and a `text` percept crosses the seam; a claim becomes
visible to the other villagers; the claimed blueprint is built block by block **while the
player builds alongside**; interrupting at dusk leaves a partial build that resumes the next
work period.

**Depends on.** V2, V3, M5.

**Tier.** `sonnet`.

**Spell-breaker check — tedium.** Posting an order must be one diegetic gesture (write in the
book), not a form to fill in or a syntax to learn. If the player has to phrase a blueprint
carefully to be understood, the interface has become a chore and the diegetic framing is
decoration.

**Spell-breaker check — micromanagement.** Once claimed, a build proceeds without the player
re-issuing, supervising, or hand-feeding materials. The player manages flow; they do not run
the site.

**Constraint — minds-are-others.** The board posts an *order*, not a command. The claim is the
villager's to make, and the plan supports no path by which the player forces one.

---

#### V5 — Death, danger, and what remains

> As a **player**, I want the walls and torches I built to be the reason my friends are still
> here in the morning, so that base-building carries the weight of protecting people rather
> than protecting loot.

**Scope.** **Preconditions, verified before any implementation** (R-4, R-5): whether POI
re-claim has natural lag after `releaseAllPois()`, and where the zombie-siege trigger can be
Mixin-suppressed plus whether a 3-villager cast satisfies the (version-dependent, inconsistently
documented) village-eligibility thresholds at all. **Admitted, no work needed:** hostile-mob
night targeting, falls, drowning, fire/lava, player direct and indirect kills — all inherited
free, all legible as neglect rather than luck. **Suppressed:** zombie sieges, per death §1 —
an event that can overwhelm well-built defences on a nightly dice roll teaches the player that
walls do not help, which is the fragility posture inverted. **One conversion-cancel Mixin**
(permadeath; conversion is ruled equivalent to death in v1). **Remains, authored because free
is invisible:** a mod-placed grave marker at the death location (or nearest safe buildable
surface) with **no villager agency required** — a 2–3 survivor cast cannot be relied on to
volunteer, and the memorial gesture cannot be optional; a **belongings bundle** capturing the
hidden inventory *before* vanilla destroys it, placed at the grave as an ordinary
`roles: ["storage"]` thing named for its owner; an **optional** job-board entry ("tend Tam's
grave") a survivor may take up or ignore, riding V4's existing mechanism. **Grief period**
per R-3: bed and job-site held unclaimed for one in-game cycle, exposed as config. **Token
discipline:** a new body token for the grave or converted mob; the dead villager's token is
retired and never reissued.

**Done proves.** On a dev server: a villager killed by a zombie leaves a named grave at the
death site with a belongings chest beside it; their bed stays unclaimed for the configured
grief period; **no siege ever fires**; a witnessing villager receives ordinary `sighting`
percepts (not a magic death broadcast) and an absent one receives a `change_report` with
`change: "gone"` on return, plus a `sighting` of the grave; the dead body's token is never
reissued.

**Depends on.** V2, V3.

**Tier.** `sonnet` — **with a named escalation trigger.** This is the one task standing on
unverified engine surface (R-4, R-5, and decision-0002's flagged-thin `GossipManager`
genericity finding). If the siege suppression point is not where death §1 assumes, or if
suppression turns out to need more than a targeted injection, that is an architecture question
outside this tier's scope: stop and escalate rather than growing the Mixin surface past
decision-0002's committed bound.

**Spell-breaker check — micromanagement.** This design *adds nothing* to villager
self-preservation, on purpose: villagers cannot starve (there is no starvation loop at all),
panic and flee autonomously, and seek shelter on their own schedule. No feeding UI, no escort
quest, no babysitting surface. The siege suppression is the one addition, and it **removes** an
arbitrary-death vector rather than adding a management surface.

**Spell-breaker check — politeness-policing.** No engine guardrail on friendly fire. The
embodied player has real physical agency including the capacity to do real harm, and adding a
guardrail would be a fiction about who the player is. What the death *means* rides the memory
channel (M7), not a judgment layer.

---

### 3.4 Integration

---

#### I1 — Demo configuration and the two run targets

> As an **operator**, I want one documented sequence that brings up a demo-ready world with the
> cast seeded, so that running the demo is a decision about when, not a research project.

**Scope.** Two artifacts, started independently: the daemon binary and the mod jar. Documented
startup ordering and the reconnect behaviour that makes **mind-restart independent of
vendor-restart** (T-4). World and server config: the vanilla daylight cycle per **R-1** (nine
cycles, stated as a ruling rather than left as a default), the **grief-period** knob (R-3), and
the **danger-tuning** knob (R-6) present but **off by default**. Cast seeding: run M3's genesis
for three villagers and bind them to bodies. Surface the per-class call/token counters and the
E6-input-tokens instrument at session end.

**Done proves.** One documented command sequence brings up a server with three personas seeded
and bound. **Restarting the daemon mid-session and reconnecting leaves the villagers with their
memories** — the demo acceptance check decision-0003 promotes from aspiration to test. Counters
report at session end. Every knob above is config, not a constant in code.

**Depends on.** M3, M7, V3, V5.

**Tier.** `sonnet`.

---

#### I2 — The evening: run it, measure it, check it against the brief

> As an **operator**, I want a full evening run judged against every beat and every
> spell-breaker with numbers attached, so that "the demo works" is a finding rather than an
> impression.

**Scope.** A ~3-hour real-time evening at 1x with the player present throughout and three
villagers alive. Walk the beat checklist (§4 below) and the spell-breaker checklist (§5).
Measure and **replace the A-n assumptions by number**: calls per class, tokens per call,
cost against the **~$20 ceiling** (baseline ≈$5.17, ≈$4.00 cached — cost is not the binding
constraint and this run is not an optimization exercise), E4 first-token latency against its
< 3 s ceiling, and E6 input tokens per villager per cycle against the ~80/day assumption and
the ~150 upgrade trigger. Record which dusk landed best.

**Done proves.** A recorded evening in which every beat in §4's coverage map is observed, no
spell-breaker check in §5 fails, and each A-n assumption is annotated with its measured value.
Where a check fails, the finding is written down with the task that owns the fix — this run is
allowed to *find* problems; it is not allowed to fail silently.

**Depends on.** All fifteen preceding tasks.

**Tier.** `sonnet`.

---

## 4. Coverage map — every beat, and the infrastructure it rides

Beats are the brief's v1-demo paragraph as mirrored in `[[v1-demo]]`. Every one traces to at
least one task; every task traces to at least one beat or to infrastructure a beat requires.

| Demo beat | Mind side | Vendor side | Seam / integration |
|---|---|---|---|
| **Three villagers with names, generated personas and desires** | **M3** (E1 genesis, firewall, endogenous desires) | **V3** (cast: profession × biome variant, nameplates) | **I1** (seeding and binding) |
| **Schedules — wake / work / socialize / sleep** | — (engine-owned; the mind is consulted only at transitions, **M5**) | **V3** (`Schedule`/`Activity`, POI bed claim, sleep pathing) | **I2** (a full cycle observed) |
| **Persistent memory** | **M2** (event-sourced log, private provenance-stamped map, admission gate) + **M7** (nightly consolidation, what a day meant) | — (SI-5: the vendor never persists, ranks or weights a mind's memories) | **I1** (mind-restart independence proves persistence is real) |
| **The job-board book** | **M5** (E3: read, weigh, claim or decline, with a reason) | **V4** (the diegetic book, `text` percepts, visible claims) | **I2** |
| **The blueprint build alongside the player** | **M5** (taking the job; decision points) | **V4** (block placement, interruption, resumption) | **I2** |
| **Dusk conversation — about the day, the work, the player** | **M6** (E4 turns under the < 3 s ceiling; E5 ambient pool) | **V3** (pair formation + the ~10 s-ahead signal), **V2** (`speak` → `speech` in earshot) | **I2** (measured first-token latency) |
| **Night danger; walls and torches protect friends** | **M5** (the urgency interrupt: the mind is consulted *after* the reflex, at the next break), **M7** (the dead stay conversationally alive, then fade) | **V5** (admitted causes, siege suppression, grave, belongings, grief period, token discipline), **V2** (`nearest_hostile`, witness `sighting`, absent-survivor `change_report`) | **I2** |

| Infrastructure | Task | Why it is not optional |
|---|---|---|
| **Transport decision (seam Q-1)** | **S1** | Narrowed to real wires by decision-0003, not answered. Scheduled rather than left floating — both skeletons block on it. |
| **Go daemon skeleton** | **M1** | The boundary-decode component decision-0003 schedules explicitly, so V-5's cost is paid once and up front rather than discovered late. |
| **Model client and routing** | **M4** | Six classes across three tiers in one process; the cacheable prefix is 23% of the bill *and* part of E4's latency budget. |
| **Fabric mod skeleton** | **V1** | The token registry is the vendor's hardest obligation, and the manifest is where SI-1 is most easily lost. |
| **Body-vendor conformance** | **V2** | The `change_report` restriction alone is a protocol rule with a measured $1.30/evening price tag. |
| **Fake-vendor harness** | **S2** | Decision-0003 makes it a first-class task: it is how mind work proceeds *before the mod exists*, and it gates the mind side of every beat. |
| **Demo config and run targets** | **I1** | Two artifacts, started independently; mind-restart independence is an acceptance check. |
| **The evening itself** | **I2** | The beats are only claimed once observed together, at 1x, for three hours. |

**Deliberately absent, with a cause:** raid content (**R-2** — player-triggered, so nothing to
build and nothing to suppress); a client jar (decision-0002 — "a friend can drop in" is
satisfied by the entity choice alone); custom per-individual skins (bounded to profession × 7
biome variants; a resource-pack spike is a later, separately-scoped option); a per-percept
formativeness scoring pass (routing §1.3 — the admission gate plus E6 already make that
judgment, and scoring duplicates it at N× the call count); a "port the daemon" task
(decision-0003 — there is no such task, and Phase 1's measurement is why).

---

## 5. Constraints and spell-breakers

### 5.1 The two load-bearing constraints, where they bit

**Loneliness-cure — the AI serves the feeling of company, not the game loop.** This is the
constraint that did the most *scoping* work, and always in the direction of cutting:

- **No villager competence project.** Nothing here builds crafting, mining, or tech-tree
  climbing for villagers — the Voyager-shaped work that would be the obvious next thing to
  build and would consume the whole demo. The embodied loop is already fun; the AI budget goes
  to choosing and relating.
- **V4's build execution is deliberately the thinnest possible build system** — one blueprint,
  engine-side placement, interruptible and resumable. It exists to put a neighbour beside the
  player while they work, not to be a construction feature.
- **The two tasks the plan will not trade away are M6 (conversation) and M7 (consolidation).**
  Conversation is the beat the thesis is made of; consolidation is what makes the villagers
  remember the player between cycles, which is the difference between company and a chat skin.
  If the demo has to shed scope, it sheds from V4's build fidelity, never from these.

**Minds-are-others — nothing may let the player edit a villager's mind.** This constraint is
structural in four places rather than stated in one:

- **M3** — the persona firewall's first half is *structural impossibility*: mode 0444, and no
  post-genesis write path exists to attempt. The test asserts the absence of a code path, not a
  runtime refusal.
- **M2** — the belief store has no external write path. Not the vendor, not the player, not a
  debug command.
- **V4** — the board posts an order; the claim is the villager's. There is no forcing path.
- **S2 / the seam itself** — SI-1 means the mind's private map is unreadable from outside, and
  §10.5 forbids the fake vendor from growing a read API "for a test's benefit," which is where
  the violation would first be built and from where it would be copied into a real vendor.

### 5.2 Spell-breakers, attached where the risk actually lives

Each spell-breaker is a named design check on specific tasks, plus a checklist item at I2.
A spell-breaker with no task attached is a slogan.

| Spell-breaker | Where the risk lives | The check |
|---|---|---|
| **Tedious interactions with the player** | **M6** (conversation and ambient), **V4** (the board) | M6: lines do not repeat within a cycle; conversations **end**; the stall-line is rare. V4: posting an order is one diegetic gesture, not a form or a syntax. **The test that matters: does the player start walking past?** |
| **Micromanagement to keep villagers alive or productive** | **V3** (schedules), **V5** (death and danger), **M5** (deliberation), **V4** (the build) | V3: a **full cycle unattended** is the scope boundary, not an aspiration. V5: nothing added to self-preservation — no feeding UI, no escort, no vigilance; siege suppression *removes* the dice-death that teaches "walls don't help." M5: reluctance is the product, but work gets done without the player re-posting. V4: a claimed build proceeds without supervision or hand-fed materials. |
| **Villagers taking offense at a player being a jerk** | **M3** (genesis), **M5** (refusals), **M6** (conversation), **V5** (friendly fire) | M3: genesis must not produce a moralizing persona template — the drift lexicon catches stated moralizing. M5: a refusal is grounded in the villager's wants and commitments, never the player's conduct; no compliance gate, no cooldown, no lockout. M6: a villager may resent, grumble, and say so — but never lectures. V5: **no engine guardrail on friendly fire**; the embodied player's capacity to do harm is left alone and its weight rides memory, not a judgment layer. |

---

## 6. Suggested lanes for the next sweep

Dependency-ordered, contract-shaped first. **Develop in parallel, merge serially** — and the
seam is a genuine parallelism gift here: the mind lane and the vendor lane touch disjoint
directories in different languages, so within a lane they are close to conflict-free. The
board directory, the wiki, and the runbook are the shared hotspots, as usual.

**Lane 0 — the wire (alone, blocks everything):**
- **S1** transport decision. `opus`, proposed escalation. Nothing else starts until the framing
  and golden vectors exist, because both skeletons encode against them.

**Lane 1 — skeletons (2 parallel, one per side):**
- **M1** mind daemon skeleton · **V1** Fabric mod skeleton.
- Maximally parallel-safe: different languages, different build systems, no shared file.

**Lane 2 — core surfaces (3 parallel):**
- **M2** memory and admission gate · **M4** model client and routing · **V2** vendor conformance.
- M2 and M4 both sit on M1 and touch different packages; V2 sits on V1.

**Lane 3 — the cast, the persona, the proofs (3 parallel):**
- **M3** persona genesis · **V3** brain, schedule, cast, pair formation · **S2** fake-vendor
  harness.
- S2 lands here rather than earlier because H-6 counts memories and the canonical end-to-end
  asserts a belief's origin class — both need M2. This is a deliberate departure from reading
  decision-0003's "first-class task" as "first task": the harness is first-class and named, but
  it cannot precede the store whose rules it proves.

**Lane 4 — the beats (parallel across sides; V4 serializes behind M5):**
- **M5** deliberation and job-board decision · **M6** conversation and ambient · **M7**
  consolidation — all three parallel on M2 + M4.
- **V5** death, danger and remains — parallel with the mind trio, on V2 + V3.
- **V4** job board and blueprint build — **the one cross-lane dependency in the graph**: it
  needs M5's claim behaviour to be demonstrable end-to-end, so it starts after M5 merges. Left
  as a real dependency rather than faked with a stub, because the beat is only observable when
  both halves exist.

**Lane 5 — integration (serial):**
- **I1** demo config and run targets, then **I2** the evening.

**Merge-order notes for the runbook author.** S1 merges before anything. Within lanes 1–4,
merge order is free — prefer merging the **vendor** side of a pair first when both are ready,
since the mind side is exercised against the fake vendor and does not block on the real one.
I1 and I2 merge last and in that order. Sixteen tasks over six lanes puts the **critical path
at six merges deep** (S1 → M1 → M2 → M5 → V4 → I1 → I2 is the longest chain, seven), which is
the number to argue with if the sweep needs to be shorter: the only real lever is splitting V4,
and §1 explains why that trade is bad.

**Escalation checkpoints for the operator at sign-off.** One proposed escalation (S1 → `opus`)
and one named escalation trigger (V5, if the siege suppression point is not where the design
assumes). Both are the operator's call, not an implementer's — and the tier ladder's own
warning applies: **verify the served model from the first dispatch's transcript before
launching siblings.**

---

## 7. What this plan does not cover

Restated so a later reader does not mistake an absence for an oversight: no replenishment,
curing, or wanderer mechanism (brief #10, punted); no multi-settlement or open-world
multiplayer (brief #9, deferred to V2); no second body vendor (the seam makes it possible, the
demo does not need it); no formativeness scoring pass (routing §1.3's upgrade path, gated on a
measured instrument); no client jar or custom skins; and no post-demo hardening. Each of those
is a task the *next* plan may want. None of them is on the path to one real evening.
