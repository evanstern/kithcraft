# LLM routing and cost envelope — a 3-villager evening at 1x

**Status:** spec 004 Phase 2 deliverable. Board task TASK-0004, milestone m-0.
**Authority:** `docs/design/kithcraft-brief.md` (ratified 2026-08-19 — reflex/planner split,
real-time-only, 3–6 villagers); `docs/design/body-protocol-v0.md` (contract accepted
2026-08-21) fixes the percept/intent surfaces this routing sits behind; decision-0002
(accepted 2026-08-21) puts the augmented vanilla villager's `Brain<E>` in charge of *doing*.
Phase 1 evidence: `specs/004-mind-daemon-routing/research/daemon-assessment.md`.

**What this is.** Which villager cognition events call a language model, at which tier, how
often, with what latency budget, and what an evening costs. It answers R2 and R3 of
`specs/004-mind-daemon-routing/spec.md`.

**What this is not.** No language or reuse recommendation — that is Phase 3, and nothing here
prefers a candidate. No transport decision (seam Q-1 stays open). No re-litigation of
real-time-only, the cast size, or the split.

**Evidence discipline.** Pricing carries URL + accessed date. Cadence and token figures are
**assumptions, labelled as such** — every one is numbered `A-n` in §3 so a later measurement
can replace it by number rather than by re-deriving the document.

---

## 0. The headline, before the arithmetic

> **An evening costs about five dollars.** The demo — 3 villagers, one continuous ~3-hour
> real-time session — runs to **≈ $5.17** at the routing below (**≈ $4.00** with prompt
> caching on the three frequent classes). About **$1.72/hour**, or **$0.57 per villager-hour**,
> across ≈ 1.63 M tokens.

And the finding that should change how the rest of the project reasons about this:

> **The entire routing decision is worth about $8.** Running every class on Opus 5 costs
> **$10.51**; running every class on Haiku 4.5 costs **$2.10**. The whole spread available to
> a tier ladder, on the demo, is the price of a sandwich.

At a 3–6 villager cast, **cost is not the binding constraint — latency and context growth
are.** Route for *quality per event*, spend the tier where the villager's character is at
stake, and put engineering effort into the two things money cannot buy back: a conversation
that does not stall (§5) and an episodic buffer that does not grow without bound (§6.3). The
tier ladder below is still worth having — halving a bill is worth an afternoon — but it should
never be allowed to argue a class *down* a tier when the class carries meaning. Cheap reflexes
for competence, expensive thoughts for meaning ([[promptworld-lineage]]) is a statement about
where thought is *warranted*, not a budget instruction.

---

## 1. The split: what calls a model, what never does

Three tiers of machinery, not two. The brief's reflex/planner split names the outer two; the
middle one is where most of the interesting engineering lives, and it is the one a reader
skimming for "which events use an LLM" will miss.

### 1.1 Engine-side reflexes — the body vendor, no model, no daemon involvement

Per decision-0002 the augmented vanilla villager's `Brain<E>` owns *doing*. None of the
following ever reaches the mind daemon at all, let alone a model:

| Reflex | Owner |
|---|---|
| Pathfinding, navigation, door/gate use, obstacle handling | vanilla `Brain<E>` + POI/pathing |
| Block placement, mining, item pickup, container use, crafting execution | mod action handlers |
| Schedule following — wake / work / socialize / rest / sleep | vanilla `Schedule` + `Activity` |
| POI claim: bed, workstation | vanilla memory modules |
| Panic, flee, self-preservation, hostile sensing | vanilla goal/sensor stack (decision-0002's load-bearing inheritance) |
| Look-at, gaze, idle animation | vanilla |
| **Target resolution** — turning a `place`/`thing_id` the mind named into a route | vendor (`body-protocol-v0.md` §5.2, AR-4) |
| Percept emission: sweeps, dedup windows, `background` shedding | vendor (§4.11) |

AR-4 is worth restating because it is the largest single cost saving in the design and it is
already won: **the mind performs no spatial arithmetic.** "Walk to the nearest well" is not a
prompt, it is `go_to` against a token the mind already holds, resolved by the engine. Every
project that has put spatial reasoning in the model has paid for it per step.

### 1.2 Mind-side deterministic machinery — the daemon, still no model

These run inside the mind daemon and **MUST NOT** become model calls. Naming them is not
pedantry: each one is a place where "just ask the model" is the obvious shortcut, and each is a
place where doing so destroys a property the project is relying on.

| Mechanism | Why it must stay model-free |
|---|---|
| `direct_perception(origin)` (§2.7) | The contract requires a pure function of the origin value. A model asked "did you really see this?" reads the prose — which is exactly failure H-3 in the fake-vendor harness (§10.3). |
| Provenance coercion — RM-2/RM-3 | A deterministic gate that resolves citations and coerces claims down. Its guarantee is *mechanical*; a model-judged version is a guarantee about a distribution. |
| The persona firewall validator | Anchor echo + drift lexicon, model-free by inherited design (assessment §1.1 D5). Rejection is a testable 100% guarantee **only** because no second model call is involved. |
| RM-4 freshness comparison (secondhand never beats fresher firsthand) | One timestamp comparison. |
| RM-5/RM-6 read-time confidence and freshness | Arithmetic on `world_time` integers — the only arithmetic the mind is supposed to do (§L-9). |
| Percept ingest: dedup by `percept_id`, `seq` gap detection, V-1…V-6 validation | Seam-boundary validation must happen *before* any state mutation (harness H-1). |
| Intent bookkeeping: pending set, `supersedes`, matching `act_result` to `intent_id` | Bookkeeping. |
| Episodic admission gate (§6.3) | The cheap filter that decides which percepts are even *eligible* to become memories. Deliberately not a model — see §6.3. |
| Ambient line selection from the daily pool (§2, E5) | A pick from a pre-generated pool; the generation was the model call. |

### 1.3 LLM event classes — the six things a villager actually thinks about

| # | Class | Trigger | Tier | Why this tier |
|---|---|---|---|---|
| **E1** | **Persona genesis** | Villager birth (once, ever) | **Opus 5** | Rarest and most consequential call in the system. Its output is who this villager *is* for the rest of the project, is written once at mode-0444 and never revised (brief #5, the persona firewall), and costs 3 calls total. There is no argument for saving forty cents here. |
| **E2** | **Routine deliberation** | Schedule transition, act completed with an open choice, `notable` percept at a natural break, deferred `urgent` interrupt | **Sonnet 5** | The volume class (42% of the bill). Choosing among options the engine has already made physically possible — genuine judgment, but bounded by a manifest-declared verb set and a memory window. Sonnet 5 with adaptive thinking at medium effort holds this; Haiku produces intents that are *valid but characterless*, which is the exact failure the project cannot tolerate. |
| **E3** | **Job-board deliberation** | A `text` percept from the board (§4.7, `origin: read`), or re-reading a live posting | **Sonnet 5** | The demo's centerpiece and the diegetic order interface (brief #7). Distinct from E2 because its context shape differs (board contents, other villagers' claims, standing relationship to the player) and because *reluctance is the product*: "villagers read, argue about, and prioritize" orders. A model that always cheerfully accepts has broken decision 6. |
| **E4** | **Conversation turn** | Dusk conversation, player speech, an argument at the board | **Sonnet 5** | The thesis. "At dusk the villagers talk to each other — about the day, about the work, about the player" is the demo's emotional payload; this is the class the whole project exists to make good. It is also the one class with a hard latency ceiling (§5), which constrains the tier from *above*: Opus 5 with thinking is too slow for conversational flow, so quality here is bought by prompt and context design, not by climbing the ladder. |
| **E5** | **Ambient line pool** | Once per villager per in-game day, batched | **Haiku 4.5** | Greetings, passing remarks, grumbles at the fire. Individually near-worthless, collectively the texture that stops a villager reading as a state machine. Generated as a **daily pool of ~8 persona-flavored lines in one batched call**, then served with zero latency (§5, E5) — the pool refreshes daily so it does not repeat itself into a spell-breaker. A remark *about something specific* escalates to a live Haiku call instead of drawing from the pool. |
| **E6** | **Nightly consolidation** | Villager sleeps (once per in-game night) | **Opus 5** | The single highest-leverage call in the design: it decides what a villager *keeps* of a day, and its output has the longest half-life of anything the mind produces. It is also fully offline (§5) and fires 27 times an evening. Rare + consequential + zero latency pressure is the exact profile the top tier exists for. |

**The class that is deliberately absent: per-percept formativeness scoring.** Phase 1 surfaced
this as an unowned design item (assessment §4.2) — the world-side salience table is forbidden by
§2.8, so formativeness is mind-side with no inherited answer. The obvious implementation is an
LLM scoring each percept, and it is the single most expensive mistake available here: it turns
every percept into a call. **The recommendation is that v1 has no scoring pass at all.** A
deterministic admission gate (§6.3) decides what enters the episodic buffer; E6 reads the whole
buffer and decides what *mattered*. The expensive thought about what a day meant already exists
and already has all the evidence — scoring duplicates it at N× the call count and worse context.

> `ponytail:` v1 ships consolidation-only formativeness. **Ceiling:** it holds while a day's
> admitted buffer fits comfortably in one E6 call. **Upgrade when** the buffer routinely exceeds
> ~150 admitted memories per villager-day or E6 output starts truncating — then add a batched
> Haiku scoring pass at each schedule transition (~5 calls/villager/day, ~$0.02/evening) that
> pre-ranks the buffer, and keep E6 as the arbiter.

---

## 2. Cadence — a 3-villager, ~3-hour evening

### 2.1 The clock, which is not the clock you expect

**A "one real evening" demo is not one in-game night.** At 1x with a vanilla daylight cycle a
Minecraft day is 24,000 ticks at 20 ticks/second = 1,200 s = **20 real minutes**. A ~3-hour
session therefore spans **≈ 9 full day/night cycles** (A-2), each with its own dusk, its own
sleep, and its own consolidation.

This is the most consequential single fact in the cadence model and it cuts both ways. It
inflates every per-cycle class ninefold — most visibly consolidation, which fires 27 times
rather than 3. It also means the demo's "at dusk the villagers talk" happens **nine times**,
which is either a gift (nine chances for the scene to land) or a problem (the scene is supposed
to feel like *the* evening). Two levers exist — lengthen the in-game day for the demo, or accept
nine cycles — and choosing between them is a demo-design question for TASK-0006, not a routing
one. **This document models the vanilla 9-cycle case**, which is the conservative direction:
a slowed daylight cycle only makes the numbers smaller, roughly in proportion.

### 2.2 Calls per class

| Class | Rate | Per villager, evening | **Evening total** | Calls/hour (cast) |
|---|---|---|---|---|
| E1 Persona genesis | once ever | 1 | **3** | — (pre-session) |
| E2 Routine deliberation | 8 / villager / cycle (A-4) | 72 | **216** | 72 |
| E3 Job-board deliberation | ~8 / villager / evening (A-5) | 8 | **24** | 8 |
| E4 Conversation turn | 12 / dusk (cast) + 30 player-directed (A-6) | ~46 | **138** | 46 |
| E5 Ambient pool | 1 batched call / villager / cycle (A-7) | 9 | **27** | 9 |
| E6 Consolidation | 1 / villager / in-game night (A-2) | 9 | **27** | 9 |
| | | | **435** | **144/hr** |

E2's rate deserves its derivation, because the one measured figure available does not transfer.
promptworld I measured ~16 planner calls per game-hour for **eight** agents at 4x
(`[I-wiki:agent-mind]`), i.e. ~2 per agent per game-hour — but I's game-hours were its own
denomination and its agents had **no engine-owned schedule**, so the planner answered "what do I
do now" from scratch. Under decision-0002 the `Schedule`/`Activity` ladder answers that question
for free most of the time, and the mind is consulted only at genuine decision points. Counting
them directly: ~5 schedule transitions per cycle (wake / work / socialize / idle / rest) plus ~3
opportunistic (an act completed leaving an open choice, a deferred urgent interrupt, a `notable`
percept at a break) = **8** (A-4). That is one deliberation per villager every ~2.5 minutes of
play — which reads, correctly, as a person who mostly gets on with their day and occasionally
stops to think.

### 2.3 Context shape and tokens per call

The split matters as much as the total: the **stable prefix** is identical across every call for
a given villager and class, so it is the prompt-caching unit (§4.3). Everything else is per-call.

| Class | Stable prefix (cacheable) | Variable context | **Input** | **Output** |
|---|---|---|---|---|
| **E1** Persona genesis | — | world premise, cast context, weirdness dial, name (1,200) | **1,200** | **1,500** |
| **E2** Deliberation | persona (600) + standing desires/values (300) + verb & salient-kind manifest in role form (400) + instructions (700) = **2,000** | memory window, K=10 situated (400) + belief/relationship slice (300) + percepts since last call (500) + current situation & available targets (300) = 1,500 | **3,500** | **200** |
| **E3** Job-board | same 2,000 | board text + other villagers' claims (300) + memory & standing relationship to the player (600) + current commitments (300) = 1,200 | **3,200** | **250** |
| **E4** Conversation | same 2,000 | transcript so far, avg over a 6-turn exchange (800) + interlocutor model: who this is, what I think of them, shared history (500) + memory window (400) = 1,700 | **3,700** | **150** |
| **E5** Ambient pool | persona thumbnail (250) | day's mood, who is around, recent notable (150) | **400** | **400** (≈8 lines) |
| **E6** Consolidation | persona + firewall anchor (800) | day's admitted episodic buffer, ~80 memories × 40 tok (3,200) + existing beliefs & narrative-so-far (1,200) + instructions & output schema (600) = 5,000 | **5,800** | **1,200** |

Three notes on shape, each inherited rather than invented:

- **The memory window is K=10, situated.** promptworld I's selector: salience halved per day of
  age, top K−2, plus 2 seeded serendipity picks from the older half (assessment §1.1 D3). The
  serendipity picks are what stop a villager's context collapsing onto its five loudest days.
- **E6 addresses memories by ordinal label `m1..mN`**, not by ID — memories have no stable IDs
  and slice indexes are unstable, so the prompt convention is the identity mechanism, and
  accepted references map back to durable `(tick, hash)` pairs afterwards. Inherited from I,
  where it was expensive to discover.
- **E6 needs a generous `max_tokens`.** I's truncation bug — a late world's digest outgrew a
  1,024-token cap and *every night was silently rejected from day 20 on* — was fixed with a
  1024→2048→4096 retry ladder. With current output limits the ladder is unnecessary; the lesson
  survives as: **never set E6's `max_tokens` near its expected output.** Budget 4,096 against an
  expected 1,200, and keep I's other rule verbatim: a transport failure lands **no marker**, so
  the night is retried rather than lost.

### 2.4 Assumptions, numbered

Every figure above rests on one of these. They are estimates, not measurements; the point of
numbering them is that a first playtest replaces them individually.

| # | Assumption |
|---|---|
| **A-1** | 3 villagers, one continuous ~3-hour real-time session at 1x, player present throughout. |
| **A-2** | Vanilla daylight cycle: 24,000 ticks ÷ 20 tps = 20 real min/day → **9 day/night cycles** in 3 hours. Night is ~7,000 ticks ≈ **5.8 real minutes** of sleep window. *Verify against the target Minecraft version before building.* |
| **A-3** | Prompt caching is available and the stable prefix is byte-identical across calls (§4.3). The headline is quoted **without** caching; caching is the lever, not the plan. |
| **A-4** | 8 deliberation calls per villager per cycle (5 schedule transitions + 3 opportunistic). Derived in §2.2, **not** ported from I's 2/agent/game-hour figure. |
| **A-5** | ~8 job-board deliberations per villager across the evening — concentrated in the ~5 cycles where a posting is live, then rare. One blueprint is posted (the demo's script). |
| **A-6** | Conversation: 2 pairwise conversations per dusk × ~6 turns = 12 calls/dusk across the cast, over 9 dusks; plus 30 player-directed replies across the evening (~10/hour). One turn = one call by the speaker. |
| **A-7** | Ambient lines are served from a daily pre-generated pool (1 batched call/villager/cycle), not generated per encounter. The per-encounter alternative is 6/villager/cycle = 162 calls and costs $0.097 instead of $0.065 — the pool is chosen for **latency** (§5), and being marginally cheaper is a coincidence. |
| **A-8** | The §4.10 `change_report` delivery restriction is implemented. Without it the episodic buffer is ~4× larger (§6.3) and E6's cost roughly doubles. |
| **A-9** | Structured outputs (`output_config.format`) are used for E2/E3/E6, so output is a parsed intent or digest rather than prose. This both bounds output tokens and removes a parse-failure mode. |
| **A-10** | Token estimates in §2.3 are order-of-magnitude engineering estimates, not `count_tokens` measurements against a real prompt. Re-baseline them against the first working prompt; note that Opus 5 / Sonnet 5 use a newer tokenizer producing ~30% more tokens for the same text than the 4.6-and-earlier family (pricing page, accessed 2026-08-21), so any figure carried over from a promptworld I measurement is an undercount. |

---

## 3. Pricing

All figures per million tokens (MTok), Anthropic first-party API, **accessed 2026-08-21** from
`https://docs.claude.com/en/docs/about-claude/pricing` (canonical:
`https://platform.claude.com/docs/en/about-claude/pricing`).

| Model | Base input | 5-min cache write | Cache hit / refresh | Output |
|---|---|---|---|---|
| **Claude Opus 5** (`claude-opus-5`) | $5.00 | $6.25 | $0.50 | $25.00 |
| **Claude Sonnet 5** (`claude-sonnet-5`) | $2.00 | $2.50 | $0.20 | $10.00 |
| **Claude Haiku 4.5** (`claude-haiku-4-5`) | $1.00 | $1.25 | $0.10 | $5.00 |

Same source, same date, two notes that bear on the estimate:

- Sonnet 5's $2/$10 was announced as introductory pricing through 2026-08-31; the page states
  the scheduled increase to $3/$15 **will not occur** and $2/$10 is now standard. Had it
  occurred, the evening would have cost ≈ $6.9 rather than $5.17.
- Claude 4.7-and-later models use a newer tokenizer producing **~30% more tokens for the same
  text**. Every token figure in §2.3 is an estimate against *these* models, but any number
  inherited from promptworld I (which ran on the older tokenizer) must be scaled up before use.

Costs rot. This section is the thing to re-check; §2's cadence is the thing to re-measure.

---

## 4. The cost envelope

### 4.1 The evening, class by class

`calls × tokens ÷ 1,000,000 × price`, input and output separately.

| Class | Tier | Calls | Input tok | Input $ | Output tok | Output $ | **Total** |
|---|---|---|---|---|---|---|---|
| **E2** Deliberation | Sonnet 5 | 216 | 216 × 3,500 = 756,000 | 0.756 × $2 = **$1.512** | 216 × 200 = 43,200 | 0.0432 × $10 = **$0.432** | **$1.944** |
| **E6** Consolidation | Opus 5 | 27 | 27 × 5,800 = 156,600 | 0.1566 × $5 = **$0.783** | 27 × 1,200 = 32,400 | 0.0324 × $25 = **$0.810** | **$1.593** |
| **E4** Conversation | Sonnet 5 | 138 | 138 × 3,700 = 510,600 | 0.5106 × $2 = **$1.021** | 138 × 150 = 20,700 | 0.0207 × $10 = **$0.207** | **$1.228** |
| **E3** Job-board | Sonnet 5 | 24 | 24 × 3,200 = 76,800 | 0.0768 × $2 = **$0.154** | 24 × 250 = 6,000 | 0.006 × $10 = **$0.060** | **$0.214** |
| **E1** Persona genesis | Opus 5 | 3 | 3 × 1,200 = 3,600 | 0.0036 × $5 = **$0.018** | 3 × 1,500 = 4,500 | 0.0045 × $25 = **$0.113** | **$0.131** |
| **E5** Ambient pool | Haiku 4.5 | 27 | 27 × 400 = 10,800 | 0.0108 × $1 = **$0.011** | 27 × 400 = 10,800 | 0.0108 × $5 = **$0.054** | **$0.065** |
| | | **435** | **1,514,400** | **$3.499** | **117,600** | **$1.676** | **$5.175** |

> **≈ $5.17 for the demo evening.** $1.72/hour. $0.57 per villager-hour. ≈ 1.63 M tokens.
> ≈ 1.2¢ per call, averaged.

Interrupt waste is inside the rounding: cancelling an in-flight E2 call on an `urgent` percept
(§5.5) does not refund its input tokens. At ~5 interrupts per villager per evening that is ~15
partial calls ≈ **$0.14**, already implicitly counted since those interrupts are among A-4's
opportunistic triggers.

### 4.2 Where the money is

| Class | Share | Read |
|---|---|---|
| E2 Deliberation | 38% | Volume × a real context. The only class where trimming context has leverage. |
| E6 Consolidation | 31% | 27 calls carrying 31% of the bill — the top tier earning its price on rarity. |
| E4 Conversation | 24% | The thesis, at a quarter of the bill. |
| E3 + E1 + E5 | 8% | Noise. |

**67% of the bill is input tokens.** That is the profile of a system whose thinking is cheap and
whose *remembering* is expensive — which is correct for this design and is the reason §4.3 and
§6.3 are the two places worth engineering.

### 4.3 The caching lever

The stable prefix (§2.3) is byte-identical across every call for a villager+class, so it is a
cache prefix. Cache hits cost 10% of base input; writes cost 1.25×.

| Class | Cached prefix | Prefix tokens | Uncached $ | With caching (≈90% hit rate) | Saving |
|---|---|---|---|---|---|
| E2 | 2,000 × 216 | 432,000 | $0.864 | $0.186 | **$0.678** |
| E4 | 2,000 × 138 | 276,000 | $0.552 | $0.119 | **$0.433** |
| E3 | 2,000 × 24 | 48,000 | $0.096 | $0.021 | **$0.075** |
| | | | | | **$1.186** |

> **Cached total ≈ $3.99** — a 23% saving, essentially free, and the same mechanism that keeps
> the conversation latency budget (§5) reachable.

**Do not cache E6's prefix.** Consolidation fires once per in-game night — every 20 real minutes
— which is past the 5-minute TTL, so every call would be a cache *write* at 1.25× base:
$0.108 → $0.135, a 25% **increase**. The 1-hour TTL is worse still. This is the general shape:
caching pays for the frequent classes and taxes the rare one.

The 90% hit-rate assumption holds only if the prefix is genuinely stable. The standard silent
invalidators apply and one is specific to this design: **a persona block is stable, but a
"current world time" or "today is day 4" line rendered into the prefix invalidates it on every
call.** Volatile content goes after the last cache breakpoint. Verify with
`usage.cache_read_input_tokens` rather than by inspection.

### 4.4 Sensitivity

| Scenario | Cost | Δ | Derivation |
|---|---|---|---|
| **Baseline** (3 villagers, 3 h, as modelled) | **$5.17** | — | §4.1 |
| **6 villagers** (brief's upper bound) | **≈ $10.1** | +95% | Everything scales linearly except the ~30 player-directed conversation replies ($0.267), which are capped by one player's attention, not by cast size: (5.175 − 0.267) × 2 + 0.267. |
| **Chattier evening** — dusks run long, the board is argued over, the player talks constantly | **≈ $8.6** | +66% | Conversation calls 3× (138 → 414) *and* +30% input per call as transcripts grow (3,700 → 4,810): E4 $1.23 → $4.60. Larger ambient pools add ~$0.03. |
| **Both** (6 villagers, chatty) | **≈ $16–17** | +3× | The realistic worst case for the demo's stated bounds. |
| **A-8 violated** — `change_report` restriction not implemented | **≈ $6.47** | +25% | promptworld I measured **75% of all memories formed** coming from that channel (`body-protocol-v0.md` §4.10). A 4× episodic buffer takes E6 input from 5,800 → ~15,400: E6 $1.59 → $2.89. **A protocol rule with a $1.30/evening price tag** — and it grows with the cast, and with the weeks. |
| **All-Opus 5** (no tier ladder) | **$10.51** | +103% | Every class at $5/$25. |
| **All-Haiku 4.5** (maximum thrift) | **$2.10** | −59% | Every class at $1/$5. Not recommended at any price — E6 and E1 are where a villager's character is decided. |
| **Fast mode on conversation** (Opus 5, $10/$50) | **+≈ $5** | +97% | Priced and **rejected**: Sonnet 5 with streaming already meets the §5 latency budget, so this buys nothing the design needs and roughly doubles the bill. |

**The horizon beyond one evening.** The brief's ambition is "over weeks." Twenty evenings at
baseline ≈ **$104** — still not the binding constraint. But the term that grows is E6's input:
a villager's belief store and narrative accumulate, and consolidation reads them. If the
belief+narrative slice grows from 1,200 to 6,000 tokens over a month, E6's per-call input rises
to ~10,600 and the class goes $1.59 → $2.24 (+41%) *per evening, forever after*. **The thing to
instrument from day one is not dollars
per evening; it is E6 input tokens per villager over time.** RM-7's no-silent-forgetting rule
means growth is by design — what bounds cost is what the mind *loads*, not what it stores.

---

## 5. Latency posture

Real-time-only (brief #8) is the decision that makes this section short. There is no cognition
horizon, no governor, no speed ladder, no staleness budget — a villager taking 20 seconds to
decide is a person mulling. All that machinery dies, and what replaces it is one rule per class.

| Class | Posture | Budget | Mechanism |
|---|---|---|---|
| **E1** Persona genesis | **Offline, pre-session** | unbounded (minutes) | Never in the player's path. Runs at villager creation. |
| **E2** Deliberation | **Mulling-tolerant** | **20 s** | The brief's own number. The villager is never frozen while it thinks — the `Brain<E>` schedule keeps the body busy (§5.1 below). Adaptive thinking on, effort medium. |
| **E3** Job-board | **Mulling-tolerant, diegetically** | **20–30 s** | The one class where latency is a *feature*: a villager standing at the board, reading, visibly weighing it, is the scene. Spend the time. |
| **E4** Conversation | **Conversation-flow — the only hard ceiling** | **< 3 s to first token** | See §5.2. Streaming, low effort, thinking off, cached prefix. |
| **E5** Ambient | **Immediate** | **< 200 ms** | Which is why it is a pool, not a call. See §5.3. |
| **E6** Consolidation | **Fully offline** | ~6 min window | See §5.4. |

### 5.1 The rule that makes 20 seconds survivable

**A deliberation call in flight must never block the body.** This is the architectural
consequence of the whole latency posture and it is worth stating as a requirement rather than an
observation: while the mind thinks, the vendor's `Brain<E>` continues the villager's scheduled
activity. The villager walks to work, tends the field, looks around. When the intent arrives it
supersedes the ambient behaviour. A design in which the villager stands motionless awaiting a
response has converted a 20-second thought into a 20-second bug, and no tier change fixes it.

The corollary for the daemon: the act surface is asynchronous by construction
(`body-protocol-v0.md` §5.1), and the fake-vendor harness requires the mind to be **correct while
indefinitely blocked** on a pending act (§10.1, T-d). The same property covers a slow model call.

### 5.2 Conversation, the one class with a real ceiling

A reply that takes 20 seconds does not read as mulling; it reads as a broken NPC, and it breaks
the scene the demo exists to produce. Three mechanisms, in order of preference:

1. **Stream, and render progressively.** Minecraft chat and nameplate text arrive over time
   anyway. A first token at ~1–2 s, streamed out at reading pace, reads as speech. This alone
   meets the budget on Sonnet 5 for a ~150-token reply and is why the tier is constrained from
   *above*: Opus 5 with adaptive thinking is slower to first token than the scene tolerates.
2. **Pre-generate the opening turn.** Pair formation at dusk is engine-side and predictable
   ~10 s ahead (two villagers pathing to the same gathering place). The first turn of a dusk
   conversation can be generated during the walk, so the scene opens instantly.
3. **Cover the gap with the ambient pool.** A pooled "Hm." or a nod-line lands in <200 ms while
   the substantive turn generates. Used sparingly — it is a stall tactic, and a villager who
   says "Hm." before every sentence is worse than one who pauses.

Configuration for E4: **streaming on, `effort: low`, thinking off, cached prefix, `max_tokens`
tight (~300).** This class is the one place in the design where the usual defaults are wrong:
extended thinking, which every other class benefits from, is a direct latency tax on the only
class that cannot pay it.

### 5.3 Ambient is a pool because it cannot be a call

A greeting must land within ~1 second of the encounter or it lands after the villager has walked
past — and a Haiku call plus network is not reliably that fast, whatever its median. So the model
call is moved off the critical path: **one batched call per villager per in-game day generates
~8 persona-flavored lines; encounters draw from the pool with zero latency.** The pool refreshes
daily, so it does not repeat itself into a spell-breaker.

The split within the class is the reflex/planner split applied one level down: generic lines come
from the pool (reflex); a remark *about something specific that just happened* escalates to a
live Haiku call (planner) and accepts the latency, because a specific remark arriving a beat late
still lands, whereas a generic one does not.

### 5.4 Consolidation runs in the sleep window

At A-2's vanilla cycle, night gives a **~5.8-minute** real-time sleep window. An Opus 5
consolidation with adaptive thinking runs comfortably inside it. Three notes:

- **All three villagers sleep at roughly the same in-game time**, so this is the design's only
  concurrency spike — 3 simultaneous Opus calls, every 20 minutes. Trivial for any rate limit at
  this cast size, and worth staggering anyway (villagers do not reach their beds on the same
  tick, so the natural stagger is free if the trigger is the sleep event rather than a clock).
- **If it does not finish before wake, nothing breaks.** I's ledger rule ports verbatim: a
  transport failure lands **no marker**, so the night was never consolidated and is retried. The
  villager wakes with yesterday undigested — recoverable, and invisible to the player.
- **Do not move E6 to the Batch API.** It is 50% cheaper and completely wrong here: a digest that
  arrives after the villager has woken and started acting on beliefs it does not yet hold is a
  correctness problem, not a latency one. Saving $0.80 an evening is not worth it.

### 5.5 Urgency and the interrupt — the inherited open item

Phase 1 flagged this as unowned (assessment §4.1): the seam carries `urgency` but explicitly
leaves the interrupt *mechanism* to the mind daemon (§2.8), and I's generation-counter
supersession died with the cognition-horizon machinery. The latency posture forces an answer, so
this document states one:

> An `urgent` percept **cancels the in-flight deliberation call** (whatever cancellation
> primitive the Phase 3 language provides), **does not itself trigger a model call**, and
> **enqueues one deliberation** whose context includes the urgent percept.

The middle clause is the load-bearing one. The body's reflex has *already* handled the physical
response — panic, flee, hostile sensing are all inherited vanilla behaviour under decision-0002.
The mind does not need to be consulted about whether to run from a zombie; it needs to be
consulted about what to do *now that it has run*, and that can wait for the next natural break.
The alternative — an urgent percept firing an immediate call — produces three concurrent panic
calls the moment a mob spawns near the cast, for no behavioural gain.

Cost: a cancelled call is still billed for input and partial output (~$0.14/evening, §4.1). This
is the correct trade — cancelling a stale thought is cheaper than acting on it.

---

## 6. What this routing demands of any daemon

Language-independent, in the shape of Phase 1 §2's obligation list, so Phase 3 can price these
against candidates without re-deriving them. **This is not a language recommendation and does not
prefer one.**

### 6.1 Of the model client

| # | Demand | Driven by |
|---|---|---|
| **RT-1** | **Streaming**, with usable first-token latency and progressive delivery. | §5.2 — non-negotiable for E4. |
| **RT-2** | **Cancellation of an in-flight call**, cleanly and promptly. | §5.5 — the interrupt mechanism *is* cancellation. |
| **RT-3** | **Prompt caching** with explicit breakpoint placement. | §4.3 — 23% of the bill, and part of E4's latency budget. |
| **RT-4** | **Structured outputs** for E2/E3/E6. | A-9 — bounds output tokens, removes a parse-failure mode, and makes an intent a value rather than a text to interpret. |
| **RT-5** | **Per-call model selection**, since six classes span three tiers in one process. | §1.3. |
| **RT-6** | **Modest concurrency** — 3–6 concurrent calls, peaking at one simultaneous call per villager. | §5.4. This is a low bar; Phase 1 §3.2 found no candidate fails it. |
| **RT-7** | **Retry with backoff, and a truncation-aware ceiling on E6.** | §2.3 — I's silent-truncation bug, whose lesson survives even though its ladder is obsolete. |

### 6.2 Of the daemon's shape

- **Per-class prompt assembly is a first-class component**, not string concatenation at the call
  site. Six classes × a stable/variable split (§2.3) is the caching design, the cost model, and
  the thing that will be tuned most often after the first playtest.
- **Cost and token accounting is instrumented from the first call.** Every assumption in §2.4 is
  a number a playtest replaces; that only happens if calls are counted by class and tokens are
  recorded per call. The harness already requires memory formation to be *countable* (harness
  H-6); counting calls is the same discipline one layer up.
- **No wall clock for anything semantic** (harness T-b). Every cadence in §2 is expressed against
  `world_time` from percepts. Real elapsed time is for latency budgets and timeouts only — never
  for "is it time to consolidate", which reads `world_time`.

### 6.3 The episodic admission gate — the one piece of new machinery

Between the percept inbox and E6 sits a deterministic filter deciding which percepts are eligible
to become memories at all. It is the cheapest and highest-leverage component in the whole cost
model, because it is the sole input to the class carrying 31% of the bill and the only term that
grows without bound over weeks.

It is deterministic by design (§1.2), and it is **not** a formativeness judgment — it is the
cheap gate that makes the expensive judgment affordable. Admit on: `urgency` ≥ `notable`; any
percept involving another body or the player; any `act_result` on an intent the mind authored a
`reason` for; any `told_fact` or `text`; any first sighting of a `kind` or `place`. Drop —
without forgetting, since RM-7 forbids silent deletion of *knowledge*, and dropping a percept
from the episodic buffer is not the same as deleting a belief — repeated `background` sightings
of already-known things, which is the bulk of the stream.

The 75% finding (§4.4, A-8) is the reason this gate exists rather than being assumed: promptworld
I shipped without one and three quarters of every villager's memory was bookkeeping noise voiced
as experience. The protocol's §4.10 restriction removes the largest single source; the admission
gate handles the rest.

> `ponytail:` v1's gate is the fixed rule list above. **Ceiling:** it holds while admitted volume
> stays near A-2's ~80/villager-day. **Upgrade when** §4.4's E6-input-over-time instrument shows
> the buffer trending past ~150 — then tighten the rules before reaching for the batched scoring
> pass of §1.3, since a cheaper gate beats a smarter one.

---

## 7. Open items this routing leaves, and who owns them

| # | Item | Owner |
|---|---|---|
| 1 | **Nine dusks or one?** §2.1's clock mismatch is a demo-design question — accept 9 cycles or lengthen the in-game day. Routing models the vanilla case; either choice only makes the bill smaller. | TASK-0006 (demo build plan) |
| 2 | **Conversation initiation.** This document prices *turns*; what makes two villagers decide to talk (an engine-side pair-formation at a gathering place, per §5.2's pre-generation lever) is behaviour design, not routing. | TASK-0006 |
| 3 | **Formativeness, if consolidation-only proves insufficient.** §1.3's upgrade path, gated on the §6.3 instrument. | deferred, measured |
| 4 | **Transport (seam Q-1) stays open.** Nothing in this routing forces it: RT-1…RT-7 are demands on the *model client*, not on the mind↔vendor wire. One observation for the record — E4's <3 s budget is a mind-internal latency, and the seam adds only the `speak` intent hop, so no candidate transport in §8's constraint list is excluded by anything here. **Flagged, not decided.** | spec 002 successor |
| 5 | **Re-baseline every A-n against the first playtest**, and re-check §3's prices, which rot. | first playtest |
