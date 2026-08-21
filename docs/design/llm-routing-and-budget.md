# LLM routing and cost envelope — a 3-villager evening at 1x

**Status:** spec 004 Phase 2 + Phase 3 deliverable. Board task TASK-0004, milestone m-0.
**Authority:** `docs/design/kithcraft-brief.md` (ratified 2026-08-19 — reflex/planner split,
real-time-only, 3–6 villagers); `docs/design/body-protocol-v0.md` (contract accepted
2026-08-21) fixes the percept/intent surfaces this routing sits behind; decision-0002
(accepted 2026-08-21) puts the augmented vanilla villager's `Brain<E>` in charge of *doing*.
Phase 1 evidence: `specs/004-mind-daemon-routing/research/daemon-assessment.md`.

**What this is.** Which villager cognition events call a language model, at which tier, how
often, with what latency budget, and what an evening costs (§1–§7, answering R2 and R3 of
`specs/004-mind-daemon-routing/spec.md`) — and, in **§8, the language/reuse recommendation**
those demands and Phase 1's evidence together support (R1).

**What this is not.** No transport decision (seam Q-1 stays open — §7 item 4, §8.6). No
re-litigation of real-time-only, the cast size, or the split. §1–§7 were written **before**
the recommendation and deliberately prefer no candidate; §8 is the only section that chooses.

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

---

## 8. Recommendation — the mind daemon's language and reuse posture

**Status: PROPOSED.** Recorded as `backlog/decisions/decision-0003`, pending operator
ratification at the TASK-0004 PR checkpoint (R4). Until ratified it is a recommendation, not
settled fact.

Evidence: `specs/004-mind-daemon-routing/research/daemon-assessment.md` (Phase 1, checked live
2026-08-21) and §1–§7 above (Phase 2). This section weighs; it does not re-gather.

### 8.1 The recommendation

> **Write the mind daemon in Go, as a new program, reusing promptworld I's four portable
> assets as source material rather than as a codebase.**

Three clauses, each load-bearing, and the second is the one that gets misread:

1. **Go** — the language, its Anthropic SDK, and the operator's fluency carry forward.
2. **A new program, architected inside-out around the vendor port.** This is *not* "lift
   `internal/mind` behind the seam", and the option named "keep Go" in the spec does not
   survive contact with the evidence. Phase 1 §1.2 measured `internal/mind` as a **co-process
   of the world engine**: it holds a replica of `sim.State` and applies every event to it,
   reads that state directly for prompt assembly, lands output through in-process injection
   doors typed in world-engine terms (`InjectIntent(sim.InjectArgs)`), and bakes the cast size
   into ten fixed-size array types over `sim.AgentCount = 8`. Under the seam all four are
   replaced — replica → percept inbox, direct state reads → a private provenance-stamped
   belief store, injection doors → `intent`/`act_result`, fixed arrays → a variable cast with
   permadeath. **The daemon does not survive the seam in any language**, so the real question
   was never keep-vs-rebuild; it was which language carries the doctrine, the portable assets,
   and the operator at the lowest total cost. The answer to *that* is Go, and the answer to
   keep-vs-rebuild is rebuild — in every candidate, including the incumbent.
3. **The four assets are source material.** `toolloop` (994 lines, sim-agnostic by
   construction — its only project imports are `llm` and `tool`), `persona` (475 lines,
   `personas.go` imports nothing at all), `tool`'s registry *mechanism* (vocabulary replaced
   wholesale by the manifest), and `llm`'s provider/breaker/budget shape minus its dead
   `cognition` dependency. Read them, lift what earns its place, and delete the rest — do not
   vendor the packages wholesale. `llm` in particular is 3,770 lines of which a large fraction
   exists to serve the staleness router that dies with real-time-only (brief #8); what this
   daemon needs from it is providers, retries, a budget meter, and concurrency.

**Which candidate this is, in the spec's own terms: rebuild-Go.** Keep-Go is rejected because
the thing it proposes to keep is not portable. Rebuild-TypeScript and rebuild-JVM are
rejected on the balance below.

### 8.2 Rationale, mapped to R1's criteria

Phase 1's central negative finding sets the frame: **SDK maturity does not discriminate.** All
four candidate SDKs are first-party, MIT, generated from one spec by the same Stainless
toolchain, and were pushed within hours of each other on 2026-08-21 (assessment §3.1). Nor does
the async story: the load is 3–6 concurrent calls at human cadence with no CPU-bound work, and
every candidate has a first-class cancellation primitive (§3.2). So two of the five criteria are
**ties**, and the decision rests on the other three plus the assets.

| R1 criterion | Finding | Weight |
|---|---|---|
| **LLM-orchestration fit / SDK** | **Tie on the client SDK — broken by proof, not by features.** promptworld I ships `anthropic-sdk-go` in production at *this exact workload*: full tool-use wire, `NewCacheControlEphemeralParam` prompt caching, `Stop` reason mapping (`[I:internal/llm/providers.go:592-800]`). §6.1's RT-1…RT-7 are therefore not a bet in Go; RT-3 (caching, 23% of the bill) and RT-7 (the truncation ceiling) are *already demonstrated* against this SDK. No other candidate offers evidence stronger than a README. | **Go, on evidence** |
| **Async story** | Tie. Goroutines/`context` proven under `-race` at this load with in-flight cancellation on `agent.slept`/`agent.died`; `AbortSignal` and structured concurrency are equally adequate. §5.5's interrupt *is* cancellation, and `context.Context` on every SDK call makes it the same primitive as everything else. | tie |
| **Seam-contract effort** | **Go's one real weakness, and it lands where the contract already mandates a component.** V-5 ("malformed, never defaulted") is against Go's grain — zero values are indistinguishable from absent for non-pointer types, and I hit exactly this (pointer fields + `omitempty`, on the record). TS discriminated unions give V-3 exhaustiveness at compile time; Pydantic makes V-5/V-6 near-declarative. **But harness H-1 requires seam-boundary validation as a distinct stage *before any state mutation* in every language.** The component exists regardless; Go's cost is that it must be written explicitly rather than derived from types. That is one file of boring code, paid once, against a criterion where the alternatives' advantage is ergonomic rather than structural (assessment §3.3: "no candidate has a structural difficulty with the contract"). | **TS/Python edge, small and bounded** |
| **Fake-vendor testability** | **Go's structural interfaces are the idiom §2.2 asks for.** The synthesized demand is a daemon "designed inside-out around a scriptable vendor port"; in Go a port is an interface declared *at the consumer*, satisfied without declaration — which is exactly how I already tests (`Submitter`/`Injector`/`SocialInjector` declared in `internal/mind`; `scriptedModel` returning queued replies). Test-to-code ratio 1.6:1 in `mind`, 2.3:1 in `toolloop`, no framework. One caution recorded against TS: the ecosystem's module-mocking reflex is in direct tension with §10.5's prohibition on a mind-readable test API. | **Go, narrowly** |
| **Operator maintainability** | **The largest genuine differential in the whole assessment.** The operator wrote and maintains ~72k non-test lines of Go *at this problem shape* — the LLM orchestrator, the tool loop, the consolidation driver, and the persona firewall this project rebuilds. Single static binary, no runtime to install, 7 direct dependencies. Against that: no evidence in either repo of the operator running a TS or Python service. This is a long-running always-on daemon holding durable per-villager state, and the operator has run precisely that shape before (daemon + pidfile + recovery + signal handling). | **Go, decisively** |

**The Agent SDK asymmetry, weighed honestly.** It is real — `@anthropic-ai/claude-agent-sdk`
and its Python sibling have no Go or Java equivalent (assessment §3.1) — and it is the single
strongest argument for TypeScript. It does not carry, for one reason: **this daemon does not
want a generic agent loop.** promptworld I deliberately wrote its own bounded loop encoding
doctrine the framework loop does not have — *"a tool call is a REQUEST; an event is the FACT;
the gate decides"* — and that doctrine maps **exactly** onto the seam's
`intent` / `intent_ack` / `act_result` split (`body-protocol-v0.md` §5.1). A framework whose
loop executes tools and returns results would have to be prevented from doing the one thing it
exists to do, because under this contract the mind **never** learns the outcome of its own act
except as a percept the vendor sends back. The asymmetry would be decisive for a daemon that
wanted the framework's loop; for this one, adopting it means fighting it. Recorded as the
strongest counter-argument and rejected on fit, not on availability.

**The JVM co-location option, weighed honestly.** One-language-project is a genuine
maintainability argument and the body vendor is already a JVM artifact (decision-0001). Three
things sink it:

- **SI-1 breach risk is the decisive one.** Co-locating mind and vendor in one process makes
  the seam a method call — which makes T-7's in-process fake vendor trivial *and* makes it
  trivially easy to hand the mind a world handle. The contract names this exact hazard: "a
  vendor that lets a mind read the [resolution index] has reintroduced omniscience no matter
  how the protocol is worded" (§6.1). SI-1 is the invariant the entire project's architecture
  rests on — the reason minds never couple to Minecraft and a V2 world is a second vendor, not
  a rewrite. **A separate process makes the breach structurally impossible rather than merely
  forbidden**, and that is worth more than a saved language.
- **"One language" overstates the saving.** The mod's toolchain is Gradle/Mixin/Yarn-mappings
  with mappings that churn across MC versions — a different maintenance world from a plain JVM
  service. The operator would be running two build systems either way.
- **The JDK target is unpinned by any evidence in this project** (assessment §3.5): the
  prior-art pass carries JVM requirements for the *surrounding* stack (PaperMC 26.1+ → Java 25;
  CraftAgent → Java 17/21), not for the chosen Fabric path. Choosing JVM on a
  one-language argument means committing to a version nobody has checked.

**And Phase 2's demands do not move the ranking.** RT-1…RT-7 are demands on the *model client*,
and §3.1's finding holds: every candidate's SDK ships streaming, retries, caching, and
per-call model selection. What §5 adds is an *architectural* demand — §5.1's "a deliberation
call in flight must never block the body", and §5.4's three-simultaneous-Opus-calls spike —
and Go's goroutine-per-call-with-`context` shape is the one this project has already run under
`-race` at a heavier load than the demo asks for.

### 8.3 The doctrine-transfer checklist

R1 requires the winning choice to state how each transferred doctrine item is carried. The
mechanisms are `port asset X`, `reimplement to contract`, and `already in the protocol` — and
the honest summary is that **most of the doctrine is already carried by the protocol, not by
any code decision**, which is itself the strongest evidence that the language choice is a
smaller decision than it looked.

| Doctrine item | How Go carries it | Mechanism |
|---|---|---|
| **Event-sourced memory** | **Reimplement to contract, ~400 lines.** The pattern — append-only log, immutability in the schema (SQLite `no_update`/`no_delete` triggers raising `ABORT`) not in convention, state as a reducer over the log — ports as doctrine and is small enough that porting is not the operation. What is **deliberately not carried**: I's event *vocabulary* (world events the mind never sees under SI-1), and the whole `log_format_version` migration chain, whose justification was determinism-for-replay and dies with it. A Kithcraft mind's log is a memory store, not a replayable world. | reimplement to contract |
| **Reflex / planner split** | **Reflex half: not carried — the engine owns it now.** Under decision-0002 the augmented vanilla villager's `Brain<E>`/`Schedule`/`Activity`/POI stack *is* the reflex half, and AR-4 forbids the mind doing spatial arithmetic at all, so I's BFS pathfinder and 2-D survival ladder are doubly dead. **Planner half: port `toolloop`'s shape** (994 lines, sim-agnostic by construction) — its REQUEST/FACT/gate doctrine maps onto `intent`/`intent_ack`/`act_result` one-to-one, which is the strongest single portability finding in the assessment. `tool`'s registry *mechanism* ports; its vocabulary is replaced by `session_open`'s runtime manifest, which is a better design than I's compiled-in registry. §1.1–1.3 above are this doctrine item applied to spend: engine reflexes are free, the daemon's deterministic machinery is free, and only the six E-classes cost money. | port asset (`toolloop`, `tool` mechanism) + already in protocol (the manifest) |
| **Salience / consolidation** | **Split three ways.** (a) *Salience-as-a-world-side-table:* **forbidden** — §2.8 bans a `salience`/`importance`/`weight` field on any percept; urgency stays body-side, formativeness is mind-side and does not cross. (b) *Consolidation machinery:* **port the shape and its measured lessons** — the once-per-night event-sourced ledger, the ordinal `m1..mN` prompt convention, the `(tick, hash)` durable identity mapping, the truncation lesson (I's digest silently outgrew a 1,024-token cap and every night was rejected from day 20 on), and "a transport failure lands **no marker**". Encoded above as E6's tier, §2.3's `max_tokens` rule, and §5.4's retry posture. (c) *Formativeness:* **new design, stated in §1.3** — v1 has no scoring pass; a deterministic admission gate (§6.3) feeds a single Opus-tier E6 that decides what mattered. The situated-memory `Where`/`Why` split is already relocated correctly by the contract (§2.4/§5.2): the vendor composes `place.descriptor`, the mind supplies `intent.reason`. | port asset (`mind/consolidate.go` shape) + new design (formativeness) + already in protocol (situated split) |
| **Epistemic hygiene** | **Already carried — into the protocol, not into code.** `direct_perception(origin)` is a contract obligation (§2.7) with `DIRECT_ORIGINS` defined in the contract so mind and vendor cannot disagree, a MUST NOT against a `direct` boolean on the wire, and harness H-3 proving the classifier ignores prose. `enforceProvenance`'s coerce-never-reject posture is RM-2/RM-3. The mechanisms were always tiny (`DirectPerception` 8 lines, `enforceProvenance` 30) — the value was the knowledge, and the knowledge is now binding text. `mentalmap.go` is the most self-contained file in I and the least portable in substance: its content is 2-D coordinates AR-4 forbids crossing the seam. **Reimplement RM-1…RM-7 to the contract; port nothing.** §1.2 above hard-codes the consequence: none of these may become a model call. | already in protocol |
| **Persona firewall** | **Port `persona` (475 lines, no imports) — the cleanest carry in the assessment.** The doctrine is the two-half design: *structural* impossibility (written once at mode 0444, no post-genesis write path exists anywhere) plus *validatory* mechanism (a model-free anchor echo under normalization, plus an authored drift lexicon matched word-boundary — so rejection is a testable 100% guarantee, which a second model call would downgrade to a guarantee about a distribution). Brief #5 changes the *source* — personas generated at birth (E1, Opus 5, §1.3), not authored at genesis for a fixed cast — and the firewall is unchanged by that: decision 5's demand *is* the structural half. The honest limit transfers with it: the lexicon catches *stated* drift; subtle drift needs the parked model-judged validator. | port asset (`persona`) |

**Carried by the choice, not by the doctrine list: two operational assets.** I's LLM layer
contributes its provider registry, circuit breakers, and budget-meter shape (minus the
`cognition` staleness router that dies), which is directly what §6.2's "cost and token
accounting instrumented from the first call" requires. And I ran this exact process shape —
long-running daemon, pidfile, recovery, signal handling — which is a maintainability asset
independent of any package.

### 8.4 What would change this recommendation

Named so a future reader can check them rather than re-argue the whole decision. Any one of
these is grounds to reopen; none is true today.

1. **A second maintainer whose language is not Go.** The maintainability criterion is the
   decisive one and it is a fact about *this operator*, not about Go. If the project acquires
   a contributor fluent in TypeScript and not Go, the largest term in the balance changes sign.
2. **The Agent SDK becoming load-bearing.** If the mind's design moves toward wanting a
   framework-provided loop — sub-agents, MCP tooling inside the mind, a managed runtime —
   TypeScript's advantage stops being about a loop this daemon rejects. The test is concrete:
   does the framework let a gate sit between the model's tool call and the fact that the mind
   is allowed to believe? If yes, re-weigh.
3. **Transport (Q-1) resolving to in-process-with-the-mod.** Then JVM's argument becomes
   structural rather than stylistic. Note the direction of causation in §8.6: this
   recommendation *forecloses* that resolution rather than waiting on it, and does so
   deliberately.
4. **The percept surface growing far beyond §2.1's shapes.** Go's V-5 weakness is bounded
   because the boundary is small. If a successor protocol multiplies percept types and nested
   optional structures, the "one file of boring code" grows into a hand-rolled schema layer,
   and Pydantic or a TS validator starts to earn its keep.
5. **`anthropic-sdk-go` falling materially behind.** The SDKs are Stainless-generated from one
   spec and shipped same-day today. If Go's lags on a capability §6.1 depends on — streaming,
   caching breakpoints, structured outputs — the tie in the first criterion breaks the other
   way. This is the cheapest of the five to monitor and the least likely.

### 8.5 What this narrows for TASK-0006 (demo build plan)

The three TASK-0004 outputs — the language choice, the tier routing, and the cost envelope —
each fix something TASK-0006 would otherwise have to decide. Stated explicitly so the demo plan
inherits them rather than reopening them.

**From the language choice:**

- **Mind tasks are Go tasks; body tasks are Java/Fabric tasks; the split is the seam.** The
  demo is a **two-language, two-artifact project**: a Fabric mod jar and a Go daemon binary.
  TASK-0006's decomposition splits at the seam and **every deliverable task lands wholly on one
  side of it** — a task that spans both is a task that has smuggled coupling across SI-1 and
  should be split. The mind daemon is its own module (its own repo or a top-level directory
  with its own `go.mod`); it does not live inside the mod's Gradle build.
- **The demo needs two run targets and an ordering.** A daemon binary and a mod jar, started
  independently — T-7 requires mind-restart to be independent of vendor-restart, so "restart
  the daemon mid-session and the villagers keep their memories" is a demo-plan acceptance
  check, not an aspiration.
- **The fake-vendor harness is a Go test suite, and it is a demo-plan task, not an afterthought.**
  `body-protocol-v0.md` §10's six tests (H-1…H-6) plus the canonical end-to-end scenario are the
  mind's development environment: they are how mind work proceeds **before the mod exists**, and
  they gate the mind side of every demo beat. The vendor-facing port is an interface declared at
  the consumer; the fake satisfies it in-process.
- **Two named porting tasks, and only two.** Lifting `toolloop`'s bounded-loop shape onto the
  intent/ack/act_result split, and lifting `persona`'s two-half firewall. Everything else in
  the mind is written fresh against the contract. TASK-0006 should not budget a "port the
  daemon" task — there is no such task, and Phase 1's measurement is why.
- **One boundary-decode component, explicitly scheduled.** Go's V-5 cost (§8.2) is one file
  that must exist before any percept handling: presence-checked decode, then validate, then
  mutate — harness H-1's ordering, made a task so it is not discovered late.

**From the routing tiers (§1.3), beat by beat.** Every demo beat in the brief, resolved to its
call profile:

| Demo beat | LLM? | Class / tier | Note |
|---|---|---|---|
| Personas, desires generated at birth | **Yes** | E1 — **Opus 5**, 3 calls total | Pre-session, unbounded latency, mode-0444 output. The persona firewall's structural half is a demo requirement, not a later hardening. |
| Schedules (wake / work / socialize / sleep) | **No** | engine `Schedule`/`Activity` | Zero cost, zero latency. Decision-0002's inheritance. |
| Walking, building, mining, pathing, door use | **No** | engine `Brain<E>` + mod handlers | AR-4: the mind names a token, the engine resolves it. |
| Player posts a blueprint on the job board | **Yes** | E3 — **Sonnet 5**, ~24 calls | The centerpiece. *Reluctance is the product* (brief #6): a villager that always cheerfully accepts has broken the demo. 20–30 s latency is the scene, not a cost. |
| One villager builds it alongside the player | **Mostly no** | E2 — **Sonnet 5** at decision points only | The *building* is engine-side. The mind is consulted on taking the job, on interruptions, and at schedule transitions — 8/villager/cycle. |
| Dusk conversation about the day, the work, the player | **Yes** | E4 — **Sonnet 5**, ~138 calls | **The one hard latency ceiling: <3 s to first token.** Streaming on, `effort: low`, thinking off, cached prefix, `max_tokens` ~300. Opus is *too slow* here — the tier is constrained from above. Pre-generating the opening turn during pair-formation pathing is a demo-plan task. |
| Grumbling at the fire, greetings, passing remarks | **Yes, offline** | E5 — **Haiku 4.5**, 27 batched calls | A daily pre-generated pool served in <200 ms. Never a live call on the critical path. |
| Persistent memory across the evening | **Yes, offline** | E6 — **Opus 5**, 27 calls | Nightly, in the ~5.8-minute sleep window. Plus the deterministic admission gate (§6.3) — **not** a model call, and the highest-leverage component in the cost model. |
| Night danger; walls and torches protect friends | **No** | vanilla goal/sensor stack | Panic and flee are inherited. The mind is consulted *after*, at the next natural break (§5.5) — an urgent percept cancels an in-flight deliberation and enqueues one; it does not fire its own call. |

Two demo-plan consequences fall straight out: **the villager must never freeze while thinking**
(§5.1 — a 20-second thought with a motionless body is a 20-second bug, and no tier change fixes
it), and **"nine dusks or one"** (§7 item 1) is TASK-0006's call — the vanilla cycle gives 9
day/night cycles in 3 hours, so the demo either accepts nine dusks or lengthens the in-game
day. Routing models the vanilla case; either choice only makes the bill smaller.

**From the cost envelope:**

- **Plan against a ceiling of ~$20 per demo evening**, not the $5.17 baseline. Baseline is
  $5.17 (≈$4.00 with caching on E2/E3/E4); the realistic worst case within the brief's own
  bounds — 6 villagers, chatty — is **$16–17**. The ceiling is the number a demo plan budgets
  against; the baseline is the number it measures toward.
- **Cost is not the binding constraint, so do not let TASK-0006 optimize for it.** The entire
  tier ladder is worth about $8 on the demo. Latency (§5) and episodic-buffer growth (§6.3)
  are the constraints; engineering effort goes there.
- **Three items are demo-plan tasks because money cannot buy them back later:** prompt caching
  with a genuinely stable prefix (23% of the bill *and* part of E4's latency budget — and
  rendering "today is day 4" into the prefix silently destroys both); the `change_report`
  delivery restriction of `body-protocol-v0.md` §4.10 (A-8 — a protocol rule with a
  $1.30/evening price tag that grows with the cast and the weeks); and **per-class call and
  token instrumentation from the first call**, since every A-n assumption is a number the first
  playtest replaces only if calls are counted by class.
- **The one metric to instrument for the long run is E6 input tokens per villager over time**,
  not dollars per evening. RM-7 forbids silent forgetting, so growth is by design; what bounds
  cost is what the mind *loads*, not what it stores.

### 8.6 One consequence for transport (Q-1) — flagged, and the direction is one-way

The spec's non-goals say to flag a language finding that bears on transport, not to decide
transport. The finding: **choosing a Go daemon against a JVM mod forecloses the in-process
option.** Mind and vendor are separate processes, so Q-1 is a choice among real wires (UDS,
TCP, stdio) rather than a choice between a wire and a method call.

This is a narrowing of Q-1's option set, not an answer to it, and it is a **deliberate** one:
T-7 requires the seam be "process-separable but not process-required", and separate processes
make an SI-1 breach structurally impossible rather than merely forbidden. Nothing in §1–§7
excludes any remaining candidate — RT-1…RT-7 are demands on the model client, not on the
mind↔vendor wire, and E4's <3 s budget is mind-internal latency to which the seam adds only the
`speak` intent hop. **The wire itself remains open for the spec 002 successor.**
