# Doctrine port — promptworld I epistemic hygiene & perception (spec 002 Phase 1)

**What this is.** The rules Phase 2's protocol draft cites instead of re-deriving. Every
item names its source as a **promptworld I wiki note**, never a file or symbol — doctrine
transfers, code does not ([[promptworld-lineage]]). Notes read just-in-time from
`/Users/evanstern/projects/promptworld/docs/wiki/` via its INDEX; nothing was read from I's
source tree.

**Status of I as an authority.** Reference, not authority. Where a rule of I's leaned on
machinery that died with I — the governor, the cognition horizon, tick-denominated staleness
budgets, determinism-for-replay, the guardian — the adaptation is stated inline under
**→ Kithcraft**, per plan.md's risk note and [[promptworld-lineage]].

**Reading key.** `[I: note-name]` = source pointer into promptworld I's wiki.
**PORTS** = transfers as-is. **ADAPTS** = transfers with a stated change.
**DROPS** = does not transfer, with the reason. **GAP** = Kithcraft needs a rule I never had.

---

## 1. Epistemic hygiene — the core doctrine

The brief's one-line version is "an agent knows only what it saw or was told, with
provenance." I mechanized that as five separable rules.

### EH-1 — Provenance is a closed vocabulary stamped at emission. **PORTS.**

Every memory carries an `Origin` drawn from a fixed, small vocabulary, stamped by the
*emitter* at the moment of emission — never inferred later, never parsed out of text.
I's vocabulary: own act / witnessed event / report received at distance / delivered omen /
conversation gist / nightly digest / grounded arrival observation.
`[I: agent-memory-window]`, `[I: executor-social-perception]`

**→ Kithcraft.** The vocabulary must be **re-cut**, not copied: the *omen* class has no
producer once the guardian dies ([[promptworld-lineage]]), and I had no player. Phase 2
should cut it against Kithcraft's actual channels — at minimum `acted` / `saw` / `heard` /
`told` / `read`. `read` is a **GAP**: the job-board book (brief decision 7) is an
artifact-mediated channel I never had, and it is neither seeing an event nor being told by
a speaker. It must be its own origin class, because its trust properties differ from both
(a book does not know it is lying, and it persists after its author leaves).

### EH-2 — Unstamped is structurally impossible; unknown classifies secondhand. **PORTS, with a new enforcement mechanism.**

I made the provenance parameter *required* on every memory constructor, so a new emission
site could not compile unstamped; and it made an absent/legacy origin classify as
**secondhand** — the conservative direction. `[I: agent-memory-window]`,
`[I: executor-social-perception]`

**→ Kithcraft.** The compiler guarantee does not survive a cross-process protocol seam.
The two replacements Phase 2 must specify: (a) provenance is a **required field** in the
percept schema — a percept without it is malformed and rejected at the seam, not
defaulted; and (b) the conservative-default rule becomes *more* load-bearing under
versioning, because an older or sloppier vendor is exactly the case where the field goes
missing. Rule to state in the versioning story: **an unrecognized or absent origin value
classifies secondhand and never direct.** The fake/test vendor (R4) is where this is
proven, since there is no compiler to prove it.

### EH-3 — One text-free classifier decides "direct perception". **PORTS — and is the highest-value single item in this port.**

I exposed exactly one pure function over the origin value — `DirectPerception(origin)` —
and made it the **only** signal the belief validator reads to decide whether an agent may
claim to have witnessed something. Explicitly: no text inspection, ever.
`[I: agent-memory-window]`, `[I: executor-social-perception]`,
`[I: nightly-consolidation]`

**→ Kithcraft.** Ports verbatim as a protocol-level obligation: the classifier is a pure
function of the stamped origin, defined *in the protocol document* so mind and vendor
cannot disagree about what counts as first-hand. This is what makes epistemic hygiene
checkable rather than aspirational — an LLM will happily write "I saw" into a belief, and
the only thing that stops it is a mechanical gate that never reads the prose.

### EH-4 — Claims are evidence-cited and coerced, never trusted. **PORTS.**

A model-authored belief must cite the specific memories it rests on (I: up to 4 ordinal
labels). A deterministic gate then resolves those citations and **coerces** the claim's
provenance down to what the evidence supports — a "witnessed" claim with no directly
perceived memory behind it degrades to "told", and with nothing resolvable at all to
"inferred". The gate never rejects the night for this; it corrects and records the count.
`[I: nightly-consolidation]`

**→ Kithcraft.** Ports as a **remember**-surface obligation (R1's third surface), not a
perceive-surface one — it constrains what the mind may durably assert, not what the body
reports. Phase 2 should carry the coerce-don't-reject posture with it: rejection throws
away a whole night's digestion over a bookkeeping error; coercion keeps the content and
downgrades the epistemic claim.

### EH-5 — Secondhand is never fresher than firsthand. Staleness *is* the trust model. **PORTS.**

When knowledge is passed between agents, the fact arrives stamped with the **teller's**
last-seen time, not the telling time, and the receiver upserts it only where the receiver's
own knowledge is absent or staler. A receiver's own fresher knowledge never loses to
secondhand. `[I: mental-map-propagation]`, `[I: social-fabric]`

**→ Kithcraft.** Ports as-is and is cheap; it is the whole anti-telephone-game mechanism
and costs one timestamp field on a told fact plus a comparison rule. Phase 2 must also
carry the companion: **the source of a told fact is the immediate teller, not the original
observer** — the chain is the provenance `[I: social-fabric]`, so a fact does not launder
itself into first-hand status by being relayed.

### EH-6 — Confidence decays on read; the store never mutates and never logs decay. **PORTS.**

A belief's stored confidence never changes with time; *effective* confidence is computed
at read time from an age and a half-life (I: 8 game-days, an order of magnitude slower
than a memory's one-day recency half-life — "convictions outlive vividness"). Below a
floor a belief stops driving behavior but stays revisable rather than being deleted.
`[I: nightly-belief-decay]`

**→ Kithcraft.** The *rule* ports; the *constants* do not (see PM-4). One of I's reasons
for read-time-only decay was snapshot churn under a determinism-for-replay regime, which
dies with I. The reason that survives is better: across a seam, a mind that mutates nothing
on a timer has no clock to keep in sync with the vendor's.

---

## 2. Perception doctrine

### PM-1 — Private knowledge, not ground truth. Two agents see different worlds. **PORTS. This is the spine.**

I's spec 041 retired villager omniscience: before it, resolvers and prompts read world
ground truth directly, so every agent knew where everything was. After it, each agent
carries a private map — explored terrain plus known place-facts with provenance and a
last-seen time — and *every* read (target resolution, prompt rendering, who-can-I-talk-to)
goes through that map instead. `[I: mental-maps]`, `[I: mental-map-model]`

**→ Kithcraft.** This is the doctrine that most directly shapes the protocol seam, and it
ports as a **structural** rule rather than a policy one: the perceive surface is the
*only* way world state reaches a mind. If the protocol offers any query that returns
ground truth — "where is the nearest bed", "list nearby villagers" — omniscience is back
and every rule above is decoration. Phase 2 should state this as a hard seam invariant:
**the mind has no read access to the world, only an inbox of percepts.** The vendor may
still resolve *action targets* against ground truth (that is the reflex half of the
reflex/planner split), but it resolves them for a mind that named a place it already knows.

### PM-2 — Perception of ABSENCE, exhaustive within a radius. **PORTS — the single most important perception fix I made.**

I's original sweep only ever reported what IS there, which made confabulated place-beliefs
**unfalsifiable** — nothing ever recorded what a place *lacked*. The fix: on arrival at a
chosen target, emit the **complete sorted set** of feature kinds within a scan radius.
Absence is implied by exhaustiveness — anything not in the list was not there.
`[I: executor-perception-observation]`

**→ Kithcraft.** Ports, and Phase 2 must build the perceive surface so an "observation"
percept is distinguishable from a "sighting" percept — one is a bounded claim about a
whole place, the other is a report about one thing. **Feasibility red flag:** exhaustive
enumeration is cheap over I's 2-D tile grid with a ~dozen-kind vocabulary and expensive
over Minecraft's 3-D block volume. See F-6 below; the mitigation is that exhaustiveness is
scoped to a **closed salient-kind vocabulary**, never "all blocks".

### PM-3 — Expectation lives mind-side; the world reports only what is. **PORTS. Best seam rule in the corpus.**

I explicitly refused an "absence_of" field on the observation payload, with the reason:
the world cannot know what an agent expected. Reconciling an observation against a held
belief — confirmation boost, bounded disconfirmation decay, silence untouched — ran
entirely mind-side. `[I: executor-perception-observation]`, `[I: nightly-belief-decay]`

**→ Kithcraft.** Ports as the load-bearing division of labour across the seam, and Phase 2
should quote it as the tie-breaker for every "which side owns this?" question: **the body
reports what is; the mind owns what was expected, what it means, and what to keep.** It is
also the rule that keeps a second body vendor (R3) cheap — a vendor implements observation,
not cognition.

### PM-4 — Freshness is evaluated at read time, per-kind, and staleness is never removal. **PORTS as a rule; the numbers DROP.**

A known fact is fresh iff `now − last_seen < horizon(kind)`, tested only when read. Time
never mutates a fact. Volatile kinds (a fire, a ground pile) get a short horizon; durable
kinds get a long one. A stale fact stays stored and invisible rather than being deleted;
only a correction, a death, or a witnessed removal deletes. `[I: mental-map-model]`

**→ Kithcraft.** The rule ports. **The constants do not**, on two counts: (a) I's horizons
were denominated in ticks where 1 tick = 1 game second `[I: game-clock]`, and Kithcraft
runs real-time-only at 1x on Minecraft's own clock — there is no shared tick semantics
across the seam; (b) I's own note records the horizons as **tuning, soak-validated rather
than derived** `[I: mental-maps]`, so importing the numbers would be importing someone
else's playtest. Phase 2 rule: durations cross the seam as **vendor-supplied world-time
values with a stated unit**, never as tick counts, and horizon values are named as
vendor/mind configuration, not protocol constants.

### PM-5 — Knowledge only ever grows or corrects, through recorded channels. No silent forgetting. **PORTS.**

There is no decay-driven deletion; a villager's knowledge changes only via a recorded
perception, a telling, or a witnessed removal. `[I: mental-maps]`

**→ Kithcraft.** Ports. Its consequence for the protocol is a testability property worth
stating in R4: because every knowledge change has a named channel, the fake vendor can
drive a mind's entire epistemic state by scripting percepts alone.

### PM-6 — Act-time removal beats sweep-time correction. **PORTS — with a measured warning attached.**

I shipped the naive version first: completing a harvest cleared the world but left the
stale fact in every map, including the actor's own, so the ambient sweep "corrected" the
map moments later and minted a third-person-voiced loss memory *for the agent who swung
the axe* and for every bystander who watched it fall. **Measured live: 75% of all memories
formed.** The fix was to remove the fact at the act, for the actor and every in-radius
witness, silently — and give the actor a first-person act memory instead. The
correction channel then only ever fires for agents who were absent, asleep, or out of
range: the genuine return-and-discover narrative it was built for.
`[I: mental-map-perception]`, `[I: executor-perception-observation]`

**→ Kithcraft.** Ports as a design warning Phase 2 should encode rather than a mechanism
to copy: **a percept channel that reports a change to someone who caused or watched the
change is a memory-flooding bug, not a feature.** The generalized rule for the perceive
surface: the vendor attributes an event to its actor and its in-range witnesses at the
moment it happens; a later diff-based channel reports only to those who missed it. Phase 2
should carry the 75% figure as the evidence — it is the kind of finding that is disbelieved
until it is cited.

### PM-7 — Ambient perception is bounded and prioritized: at most one of each kind per beat, and the chatty channel is not a trigger. **PORTS.**

I capped its sweep at one "saw" plus one "correction" per agent per beat, staggered agents
across the sweep, made the chatty channel digest-only (no narration, no planner wake), and
made the correction channel wake the planner **only** when the removed fact matched that
agent's own current target. `[I: mental-map-perception]`, `[I: executor-perception-observation]`

**→ Kithcraft.** Ports, and the reasoning strengthens: percepts crossing a process
boundary and eventually reaching a token budget cost far more here than in-process events
did in I. Phase 2 should give the perceive surface an explicit **salience/urgency** notion
so the seam distinguishes *this changes what you are doing right now* from *this is
background texture* — because the mind must be able to ignore the second without dropping
it. **ADAPTS in one respect:** I's cadence knobs were tick-denominated and partly justified
by hot-path relief in a fixed-timestep loop; here budgeting is a vendor-side concern and
the protocol must **not** expose a tick number or a sweep cadence.

### PM-8 — Peer positions are last-*seen*, never live. **PORTS.**

"Talk to X" and "go find X" resolved against a remembered sighting, so an agent walks
honestly to where it last saw someone and finds them gone; only liveness stayed a live
check. `[I: mental-map-model]`, `[I: mental-map-perception]`

**→ Kithcraft.** Ports. Note the deliberate exception it proves: I kept *death* as a live
check because acting on a remembered-alive peer is a correctness bug, not drama. Phase 2
should name the same narrow exception class explicitly rather than letting vendors invent
their own.

### PM-9 — Directions are exchanged in conversation, capped, with the teller's freshness. **PORTS.**

Every founded conversation exchanged a small cap (I: 2) of fresh facts per direction that
the other party lacked or held staler, chosen freshest → nearest-to-listener → stable
order, riding alongside the gossip slot. `[I: mental-map-propagation]`

**→ Kithcraft.** Ports as the concrete mechanism behind "or was told". Two Kithcraft
adaptations: (a) the **player** is now a teller, so the told-channel's source must be able
to name a non-villager speaker; (b) I's divine reveal channel DROPS with the guardian, and
must not be quietly re-used as the player's channel — a player telling a villager where the
well is is `told` with a player source, and carries a player's fallibility, not a granted
truth.

### PM-10 — Situate every memory: where it happened, and why the agent was there. **PORTS as intent; the mechanism ADAPTS.**

Every memory in I baked a *place* (coordinates plus a deterministic nearest-feature noun
phrase — "the fire", "the woods") and, for a driven act, the **reason** of the intent that
drove it; witness memories carry no reason, because a witness did not drive the act. The
composed text was baked once at emission and never re-derived downstream.
`[I: agent-memory-window]`, `[I: executor-social-perception]`

**→ Kithcraft.** The intent ports and matters (unsituated memory is what makes an LLM
villager sound like a chatbot). The mechanism adapts twice: (a) place description is a
**vendor** duty, since only the vendor knows the world's features — and it must cross the
seam as an opaque place identity plus an optional human-readable descriptor, never as
Minecraft coordinates the mind does arithmetic on (see F-7); (b) the "reason" half comes
from the mind's own act, so it is the mind that joins it — the vendor never invents a why.

### PM-11 — Bake the payload at emission; never re-derive downstream. **ADAPTS — same rule, new justification.**

I required every event payload to be complete at emission and never recomputed at
consumption, so live and replayed runs agreed byte-for-byte.
`[I: mental-map-model]`, `[I: executor-social-perception]`

**→ Kithcraft.** The original justification — determinism-for-replay — **dies with I**
([[promptworld-lineage]]). The rule survives on a different and arguably stronger ground:
across the seam the vendor is the only party that can see ground truth, and **the mind
cannot call back**, because calling back is the omniscience hole PM-1 closes. So a percept
must be self-contained when it crosses. Phase 2 should state this reason explicitly rather
than inheriting I's, or a future reader will delete the rule along with the replay
machinery.

### PM-12 — Salience is assigned where the memory is minted, on a fixed table with a reserved interrupt band. **DOES NOT TRANSFER CLEANLY — flagged for Phase 2.**

In I, the **world** minted memories from a fixed salience table (talk 3★ … witnessed death
10★), and salience did double duty: it ranked the working-memory window *and* gated an
interrupt band (values at/above the interrupt threshold superseded an in-flight thought).
Most entries were deliberately kept *below* the interrupt threshold; the one deliberate
exception was the "you are dangerously cold and have done nothing about it" percept, whose
entire job is to break a mis-scheduling mind out of a loop.
`[I: agent-memory-window]`

**→ Kithcraft — the divergence.** Under this project's seam, **memory belongs to the mind**
(it is the third surface, "remember"), so a world-side salience table is a layer violation:
the body would be deciding what is formative. Two of the three uses still need a home:

- *Urgency* (should this interrupt what the mind is doing?) is partly a **body** judgement —
  the body is what knows you are on fire — so the perceive surface needs an urgency/priority
  field. That is PM-7's ask.
- *Formativeness* (should this become a durable memory, and how heavily weighted?) is a
  **mind** judgement and must not cross the seam as a number the vendor chose.
- I's interrupt mechanism itself (a generation counter superseding in-flight thoughts) was
  entangled with the cognition-horizon/thought-scheduling machinery that **dies**; Phase 2
  should specify only the seam half — the body can mark a percept urgent — and leave what a
  mind does about it to TASK-0004.

### PM-13 — Hearing. **GAP — no doctrine to port.**

I had no hearing channel. Its proximity model was sight (a witness radius) plus a
conversation founded at adjacency, with a courtesy "hail" pause so a hailer could close
distance. `[I: executor-social-perception]`, `[I: social-fabric]`

**→ Kithcraft.** Hearing is genuinely new and must be **derived**, not ported. The
epistemic rules above are what constrain it: a heard thing is direct perception of a
*sound*, not of its cause (hearing a mob is not seeing which mob), so Phase 2 should keep
"heard" a distinct origin from "saw" precisely so a mind cannot launder an inference from
a noise into a witnessed fact. Occlusion and range asymmetry (sound passes walls, sight
does not) is the property that makes hearing worth having at all.

---

## 3. Feasibility cross-check — channels vs the Fabric brain substrate

Substrate facts below are **already verified** and are cited, not re-researched:
[[villager-brain-api]] (the note) and `specs/001-mod-stack-decision/research/prior-art.md`
§6 (the citations, checked 2026-08-20). No new web research was performed for this phase.

**The two substrate buckets** ([[villager-brain-api]], prior-art §6):

- **Plain API — no Mixin needed to *use* an existing brain:** `Schedule` get/set;
  `Activity` queries and per-activity task-list assignment; `MemoryModuleType` read/write
  (`hasMemoryModule` / `getOptionalMemory` / `remember` / `setMemory` / `forget` /
  `isMemoryInState` / `hasMemoryModuleWithValue`); `Sensor`-driven memory refresh each tick.
- **Mixin/accessor injection required to *extend*:** registering new `Activity` values,
  new `MemoryModuleType`s, new `PointOfInterestType`s, and wiring a custom Activity into an
  entity's brain init and a `ScheduleBuilder`-modified `Schedule`. Standard Fabric practice,
  demonstrated end-to-end in the Fabric community wiki's villager-activities tutorial;
  recorded as **owned engineering surface** and an accepted risk of decision-0001
  ([[mod-stack-decision]]).

| # | Perception channel | What the protocol likely commits to | Substrate mapping | Bucket | Feasibility verdict |
|---|---|---|---|---|---|
| F-1 | **Sight — entities** (villagers, player, mobs) | Percepts naming other bodies seen, with a last-seen place; PM-8's last-seen-not-live rule | Vanilla brains are already refreshed by `Sensor`s writing into `MemoryModuleType`s; reading those modules is plain API ([[villager-brain-api]]; prior-art §6) | **Plain API**, most likely | **Green, with one verification.** The *class* of surface is verified; the specific vanilla sensor/memory-module names for visible entities are **not** verified in our notes. Verify names against the target MC version at implementation (the note's own operational caveat: Yarn mappings shift). Fallback if a needed module is absent: a mod-side entity scan, which is cheap and bounded. |
| F-2 | **Sight — occlusion / line of sight** | A stated posture (does a wall block sight?), not necessarily full ray-cast fidelity in v0 | Not a brain-API concern; a server-side world/ray query the mod owns | **Owned mod surface** (not Mixin) | **Green.** No brain extension needed. Cost is per-query, so PM-7's budget rule applies. Keep v0's commitment to a *posture* ("sight is occluded") and leave fidelity to the vendor — an occlusion model is not a seam contract. |
| F-3 | **Sight — places/features** (workstations, beds, fires, chests, structures) | I's place-facts (PM-1): known features with provenance + last-seen | POI machinery covers the village's own well-known sites; arbitrary block-level features have **no** vanilla memory-module representation, so the mod scans and keeps its own store | Mixed: POI reads are plain API; **new** `PointOfInterestType`s and any new `MemoryModuleType` need **Mixin** ([[villager-brain-api]]) | **Yellow.** Feasible and squarely inside the accepted risk of decision-0001, but this is where owned Mixin surface actually gets spent. Mitigation available: keep the mental map **vendor-side in the mod's own store** rather than in vanilla memory modules — nothing in the protocol requires a villager's knowledge to live in `Brain`. Phase 2 should not commit to brain-resident knowledge. |
| F-4 | **Hearing** | A distinct `heard` origin (PM-13): sounds with a direction/place and a kind, deliberately weaker than sight | No verified vanilla brain surface for general sound reception in our notes; sound/vibration events are engine-side and the mod would hook them server-side | **Owned mod surface**; likely no Mixin, possibly a Mixin if no public hook exists | **Yellow — the least-verified channel.** Nothing in [[villager-brain-api]] or prior-art §6 speaks to sound at all, so this row is an honest unknown rather than a verified green. **Phase 2 must not write a hearing schema that presumes a specific engine event source.** Define hearing by its *epistemic* properties (a sound of kind K from roughly direction D, cause unknown) so any hook satisfies it, and card the engine-hook verification as implementation work. |
| F-5 | **Being told** (villager↔villager, player→villager) | PM-9's capped, freshness-stamped fact exchange; player speech as a `told` percept with a player source | Chat/interaction is entirely mod-side; no brain involvement | **Owned mod surface**, trivial | **Green.** Cheapest channel and the one that carries the most product value (the brief's thesis is company). No substrate risk. |
| F-6 | **Observation of absence** (PM-2) | Exhaustive set of salient kinds within a scan radius on arrival, plus a dedup window | Mod-side scan over a bounded volume; no brain surface involved | **Owned mod surface** | **Yellow — the sharpest cost red flag.** I's version was exhaustive over a 2-D tile disc with a ~dozen-kind vocabulary; Minecraft is a 3-D block volume and "exhaustive" is unbounded if taken literally. **Constraint Phase 2 must respect:** exhaustiveness is defined over a **closed, small salient-kind vocabulary**, never over blocks — "everything of the kinds I care about, within radius R" — otherwise the falsifiability property (PM-2) is bought at an unpayable price. Carry I's dedup rule too (an unchanged place re-observed inside a window emits nothing). |
| F-7 | **Place identity across the seam** (PM-10) | An opaque place token + optional human-readable descriptor | Vendor-internal; the mod resolves tokens back to positions | **Owned mod surface** | **Green, and it is an R3 requirement.** Minecraft's coordinate convention must not cross the seam (spec R3). Corollary Phase 2 should make explicit: **the mind performs no coordinate arithmetic** — "nearest known X" resolution is the vendor's job, which is exactly the reflex half of the reflex/planner split ([[promptworld-lineage]]). This also removes any need to port I's BFS/Manhattan geometry. |
| F-8 | **Self-action results** (`acted` origin) | Act-surface results feeding back as percepts, incl. PM-6's actor-attributed act memory | Mod-side: the vendor executes and reports outcome | **Owned mod surface** | **Green.** Note it is the *act* surface's result half that produces this origin — Phase 2's intent/result split (tasks.md Phase 2, line 2) is where PM-6's "attribute to the actor at the moment it happens" rule lands. |

**Net substrate verdict.** No planned perception channel is blocked by the Fabric
substrate. Everything except F-1 lives in mod-owned code rather than in vanilla brain
extension, which *reduces* Mixin exposure relative to what decision-0001 already accepted.
The recommended Phase 2 posture — **keep villager knowledge in the mod's own store, not in
vanilla memory modules** (F-3) — keeps the brain substrate for what it is genuinely good
at (schedules, activities, the village fiction: the reflex half) and keeps the epistemic
layer portable to the second vendor (R3).

---

## 4. What Phase 2 must carry forward

Rules to cite rather than re-derive: EH-1…EH-6, PM-1…PM-11.
Divergences to state explicitly, because a reader who knows promptworld I will expect the
old behaviour: **PM-12** (salience moves mind-side; the seam carries urgency, not
formativeness), **PM-9(b)** (no divine-reveal channel; the player is a fallible teller),
**PM-4** (horizons are configuration, not protocol constants), **PM-11** (the bake-at-
emission rule keeps its shape and loses its original reason).
Red flags to respect: **F-6** (exhaustiveness must be vocabulary-scoped or it is
unaffordable), **F-4** (hearing has no verified engine hook — specify it epistemically,
not mechanically), **F-3** (do not put the mental map in vanilla memory modules),
**PM-6** (a change-report channel that reports to the actor floods memory — 75%, measured).
