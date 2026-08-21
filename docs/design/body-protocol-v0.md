# Body protocol v0 — perceive / act / remember

**Status:** draft contract, v0. Spec `specs/002-body-protocol-v0/`, board TASK-0002.
**Authority:** `docs/design/kithcraft-brief.md` (ratified) mandates the seam; this document
defines it. Doctrine is cited, not re-derived — rule IDs `EH-n` / `PM-n` / `F-n` point at
`specs/002-body-protocol-v0/research/doctrine-port.md`, which in turn points at promptworld I's
wiki notes. Wiki grounding: [[body-protocol-seam]], [[promptworld-lineage]], [[design-brief]].

**What this binds.** The seam between a **mind** (one villager's cognition, a process) and a
**body vendor** (a world that gives that villager a body — the Fabric mod is the first). It
binds message shapes, provenance, and the division of labour. It does **not** bind either
side's internals: how a mind remembers, how a vendor pathfinds, and what an urgent percept
makes a mind do are all deliberately out of scope.

**How to read.** Shapes are given as JSON for legibility; they are defined as **data**, not as
a wire format — see §8, where transport is an open question, not a decision. `MUST` /
`MUST NOT` / `SHOULD` / `MAY` are load-bearing where they appear. Unmarked prose is rationale.

---

## 1. Seam invariants

The five rules everything below serves. A change that breaks one of these is not a protocol
revision; it is a different protocol.

| # | Invariant | Source |
|---|---|---|
| **SI-1** | **The mind has no read access to the world.** The perceive surface is the *only* path by which world state reaches a mind. The protocol offers no query, no lookup, no "where is the nearest X". A mind holds an inbox, not a handle. | PM-1 |
| **SI-2** | **Every percept is stamped with provenance at emission**, from a closed vocabulary, by the vendor. Provenance is never inferred, never parsed from text, never defaulted silently. | EH-1, EH-2 |
| **SI-3** | **Percepts are self-contained when they cross.** Everything the mind needs is baked in at emission. | PM-11 (see below) |
| **SI-4** | **The body reports what *is*; the mind owns what was expected, what it means, and what to keep.** This is the tie-breaker for every "which side owns this?" question. | PM-3 |
| **SI-5** | **Durable memory belongs to the mind.** The vendor persists bodies and the world; it never persists, ranks, or weights a mind's memories. | PM-12, §6 |

### SI-3 has a new reason — state it or it will be deleted

promptworld I required payloads baked at emission so live and replayed runs agreed
byte-for-byte. That justification **dies with I** ([[promptworld-lineage]]:
determinism-for-replay is gone). The rule survives on stronger ground:

> The vendor is the only party that can see ground truth, and **the mind cannot call back** —
> calling back is exactly the omniscience hole SI-1 closes. A percept that is incomplete when
> it crosses can never be completed.

Any future reader who removes this rule alongside the replay machinery has removed the thing
that makes SI-1 survivable. (PM-11.)

---

## 2. Common types

### 2.1 Envelope

Every message on the seam:

```json
{
  "protocol": "0.1",
  "message": "percept",
  "session": "s-7f3a",
  "seq": 1041,
  "body": "b-eda",
  "world_time": 918233,
  "payload": {}
}
```

| Field | Type | Req | Meaning |
|---|---|---|---|
| `protocol` | string `"MAJOR.MINOR"` | ✔ | §7. |
| `message` | enum | ✔ | `session_open` \| `session_close` \| `percept` \| `intent` \| `intent_ack` \| `cancel`. |
| `session` | string | ✔ | Opaque session id, vendor-issued. |
| `seq` | integer | ✔ | Monotonic per `(session, body, direction)`. Gaps mean loss; see §8 T-2. |
| `body` | string | ✔ | Stable body identity (§2.3). |
| `world_time` | integer | ✔ | Vendor world time at emission, in the session's declared unit (§2.2). |
| `payload` | object | ✔ | Shape determined by `message`. |

One envelope carries exactly one message. A vendor MAY multiplex several bodies over one
connection; `body` disambiguates.

### 2.2 Time

Time crosses as **integers in a unit the vendor declares once**, at `session_open`
(`time_unit`, e.g. `"second"`). Durations use the same unit.

- The protocol MUST NOT carry tick counts, frame numbers, or any vendor-internal clock
  denomination (PM-4: I's horizons were tick-denominated against machinery Kithcraft does not
  have, and were tuning constants besides).
- `world_time` MUST be monotonically non-decreasing within a session.
- Freshness horizons, decay half-lives, and staleness thresholds are **configuration of the
  mind**, never protocol constants and never vendor-supplied numbers (PM-4, EH-6).
- A time value MAY be `null` where the vendor genuinely does not know it. `null` MUST be
  treated by the mind as maximally stale — never as "now".

### 2.3 Identity and tokens

| Token | Refers to | Stability requirement |
|---|---|---|
| `body` | A minded body (a villager, the player is not one) | Stable for the lifetime of that body, **across sessions**. |
| `thing_id` | A specific thing in the world — a bed, a chest, a mob, a player | Stable while the vendor can track that instance. |
| `place` | A location, opaquely | Stable across sessions for the same location. |
| `kind` | A *class* of thing (§3) | Stable for the vendor's lifetime. |

**MUST NOT reuse.** A vendor MUST NOT issue a retired token for a different referent. A mind's
memories hold tokens; silently rebinding one corrupts every memory that mentions it, with no
channel by which the mind could learn of the corruption (SI-1). This is the single hardest
persistence obligation the protocol places on a vendor.

Tokens are **opaque**. See §3.

### 2.4 Place

```json
{ "place": "pl-3a91", "descriptor": "the well" }
```

| Field | Type | Req | Meaning |
|---|---|---|---|
| `place` | string | ✔ | Opaque token. |
| `descriptor` | string \| null | — | Human-readable noun phrase, vendor-composed. For prose and prompting **only**; MUST NOT be parsed or branched on. |

The descriptor is PM-10's situated-memory intent, moved to the side of the seam that can
actually produce it: only the vendor knows the world's features, so only the vendor can say
"the well". The *reason* half of PM-10 comes from the mind's own intent (§5.3) — the vendor
never invents a why.

No coordinates cross. Ever. See AR-4.

### 2.5 Thing

```json
{
  "thing_id": "th-882",
  "kind": "k:sleeping-place",
  "roles": ["shelter"],
  "descriptor": "a bed",
  "body": null,
  "count": 1
}
```

| Field | Type | Req | Meaning |
|---|---|---|---|
| `thing_id` | string \| null | — | Instance identity, where the vendor tracks instances. `null` for "a thing of this kind, not individuated". |
| `kind` | string | ✔ | Opaque kind token (§3). |
| `roles` | string[] | — | Open role vocabulary (§3). Minds MUST tolerate unknown roles. |
| `descriptor` | string \| null | — | Human-readable. Prose only. |
| `body` | string \| null | — | Set when this thing is a minded body; carries its `body` token. |
| `count` | integer | — | Default 1. For aggregate sightings ("three of these"). |

### 2.6 Provenance

Attached to **every** percept. Absent or malformed provenance means the percept is malformed
(EH-2a).

```json
{
  "origin": "told",
  "source": { "kind": "body", "body": "b-tam", "descriptor": "Tam" },
  "observed_at": 917400,
  "received_at": 918233,
  "hops": 1
}
```

| Field | Type | Req | Meaning |
|---|---|---|---|
| `origin` | enum | ✔ | §2.7. Closed vocabulary. |
| `source` | object \| null | ✔ | Who or what this reached the body *through*. `null` only when `origin` is `self`-ish (`acted`, `felt`) or `saw`/`heard`, where the source is the body itself. |
| `observed_at` | integer \| null | ✔ | World time at which the content was **observed by the source** — not the time it reached this body. For a told fact this is the **teller's** last-seen time (EH-5). `null` = unknown = maximally stale. |
| `received_at` | integer | ✔ | World time the percept reached this body. |
| `hops` | integer | — | 0 = the source observed it directly. ≥1 = relayed. Informational; it does **not** promote or demote origin. |

**The source is the immediate teller, not the original observer** (EH-5 companion). A fact does
not launder itself into first-hand status by being relayed: `hops` grows, `origin` stays
`told`, `observed_at` stays the original observation time.

### 2.7 Origin — the closed vocabulary and the classifier

The vocabulary is **re-cut for Kithcraft, not copied** from promptworld I: I's *delivered omen*
class has no producer once the guardian dies, and I had no player (EH-1 → Kithcraft).

| `origin` | Means | Direct? |
|---|---|---|
| `acted` | This body did it. | ✔ |
| `saw` | This body saw it. | ✔ |
| `heard` | This body heard a **sound** (§4.4 — not its cause). | ✔ |
| `felt` | This body's own condition (§4.9). | ✔ |
| `told` | A speaker told this body. | ✘ |
| `read` | This body read an artifact. | ✘ |

`read` is a **GAP** promptworld I never had: the job-board book (brief decision 7) is an
artifact-mediated channel that is neither seeing an event nor being told by a speaker, and its
trust properties differ from both — a book does not know it is lying, and it persists after its
author leaves. It gets its own class rather than being folded into `told`.

`felt` is an **addition beyond the doctrine port's stated minimum** (`acted`/`saw`/`heard`/
`told`/`read`). Interoception is direct perception but it is not sight, and folding "I am
freezing" into `saw` would make the classifier lie about what kind of evidence backs a belief.
It is direct because the body is the only possible witness to its own condition.

**The classifier (EH-3 — the highest-value single item in the port):**

```
DIRECT_ORIGINS = { "acted", "saw", "heard", "felt" }
direct_perception(origin) = origin ∈ DIRECT_ORIGINS
```

- This function is defined **here, in the protocol**, so mind and vendor cannot disagree about
  what counts as first-hand.
- It MUST be a pure function of the `origin` value and of nothing else. It MUST NOT inspect
  percept text, descriptors, source names, or `hops`.
- A `direct` boolean MUST NOT appear on the wire. A derived value that can be transmitted is a
  derived value that can disagree with its source.
- An **unrecognized or absent** origin classifies **secondhand, never direct** (EH-2b). This
  conservative default is what makes §7's version rules safe.

This is what makes epistemic hygiene checkable rather than aspirational. A language model will
happily write "I saw it myself" into a belief; the only thing that stops it is a mechanical
gate that never reads the prose.

### 2.8 Urgency — and the field that MUST NOT exist

```json
{ "urgency": "background" }
```

`urgency` ∈ `background` | `notable` | `urgent`. Vendor-assigned, required on every percept.

- `background` — texture. The mind may ignore it without dropping it.
- `notable` — worth attention at the next natural break.
- `urgent` — *this changes what you are doing right now*. The body is what knows you are on
  fire; that judgement is genuinely body-side (PM-12).

**The seam carries urgency, not formativeness.** No `salience`, `importance`, `weight`, or
`memorability` field may exist on any percept, in v0 or any successor.

This is the protocol's deliberate divergence from promptworld I, and a reader who knows I will
expect the old behaviour. In I, the **world** minted memories from a fixed salience table
(talk 3★ … witnessed death 10★) and salience did double duty: ranking the working-memory window
*and* gating an interrupt band. Under this seam, memory belongs to the mind (§6), so a
world-side salience table is a layer violation — the body would be deciding what is formative
about a life it is not living. The three uses split: *urgency* stays body-side as the field
above; *formativeness* is mind-side and does not cross; I's interrupt **mechanism** (a
generation counter superseding in-flight thoughts) was entangled with the cognition-horizon
machinery that dies, and is left entirely to the mind daemon (TASK-0004). The protocol
specifies only that the body can mark a percept urgent. (PM-7, PM-12.)

---

## 3. The abstraction rule for world-specific concepts

**R3's requirement, stated once and applied throughout.** The failure this prevents is not
aesthetic: the moment a mind branches on `minecraft:oak_log`, the second body vendor is a
rewrite of the mind, and the seam bought nothing.

| # | Rule |
|---|---|
| **AR-1** | No vendor-native type names, registry identifiers, class names, coordinate conventions, or units cross the seam in any field the mind branches on. |
| **AR-2** | **`kind` tokens are opaque handles.** A mind MAY compare them for equality and use them as map keys. A mind MUST NOT parse, split, substring-match, namespace-match, or otherwise derive meaning from a token's spelling. A vendor MAY spell tokens however it likes; a vendor SHOULD NOT spell them as its native identifiers, precisely so that a mind violating AR-2 breaks loudly rather than quietly. |
| **AR-3** | **Meaning crosses as `roles` plus `descriptor`.** `roles` is an open vocabulary of affordances a mind may reason about — v0 seeds `shelter`, `sleeping_place`, `workplace`, `food_source`, `water`, `light_source`, `storage`, `danger`, `boundary`, `gathering_place`, `readable`. Minds MUST tolerate unknown roles (a role they do not recognize is a role they do not act on, not an error). `descriptor` is for prose and prompting only. |
| **AR-4** | **Space crosses as opaque place tokens plus coarse bands.** No coordinates, no vectors, no distances in world units, no arithmetic. `extent` and `distance` ∈ `here` \| `near` \| `middling` \| `far`; `bearing` ∈ a vendor-declared small set of relative directions, or `null`. **The mind performs no spatial arithmetic**: "the nearest known well" is resolved by the vendor, from a place the mind named (F-7). This is exactly the reflex half of the reflex/planner split ([[promptworld-lineage]]) — and it is why no geometry needed porting from I. |
| **AR-5** | **Time crosses per §2.2** — integers in a declared unit, never ticks. |
| **AR-6** | **Condition and quantity cross as bands or counts**, never as vendor-native scales. `none` \| `mild` \| `severe` \| `critical`; `count` as a plain integer. No hit points, no hunger bars, no durability. |

**Applied — a worked pair:**

```json
{ "kind": "minecraft:red_bed", "pos": {"x": 118, "y": 64, "z": -37}, "health": 14 }
```

is a violation of AR-1, AR-2, AR-4, and AR-6 in one object. The same content, admissibly:

```json
{
  "thing": { "thing_id": "th-882", "kind": "k:sleeping-place", "roles": ["shelter", "sleeping_place"], "descriptor": "a bed" },
  "place": { "place": "pl-3a91", "descriptor": "Tam's house" },
  "distance": "near"
}
```

A text-world vendor produces the second shape as easily as a Minecraft one. That is the test.

---

## 4. The **perceive** surface (world → mind)

One-directional push. The mind holds an inbox and nothing else (SI-1).

### 4.1 Percept envelope

```json
{
  "percept_id": "p-9c21",
  "percept_type": "sighting",
  "urgency": "background",
  "provenance": { "origin": "saw", "source": null, "observed_at": 918233, "received_at": 918233 },
  "place": { "place": "pl-3a91", "descriptor": "the well" },
  "content": {}
}
```

| Field | Type | Req | Meaning |
|---|---|---|---|
| `percept_id` | string | ✔ | Unique within the session. Used for dedup (§8 T-2) and cited by the mind's durable claims (§6.4). |
| `percept_type` | enum | ✔ | §4.2–§4.9. |
| `urgency` | enum | ✔ | §2.8. |
| `provenance` | object | ✔ | §2.6. Missing → **malformed**, rejected at the seam, never defaulted (EH-2a). |
| `place` | object \| null | ✔ | Where this happened, as far as the body can tell (PM-10). `null` only where genuinely placeless. |
| `content` | object | ✔ | Shape determined by `percept_type`. |

A percept MUST be complete on arrival (SI-3). A vendor MUST NOT emit a percept expecting the
mind to ask a follow-up question; there is no channel for one.

### 4.2 `sighting` — a report about one thing

```json
{
  "percept_id": "p-9c21", "percept_type": "sighting", "urgency": "background",
  "provenance": { "origin": "saw", "source": null, "observed_at": 918233, "received_at": 918233 },
  "place": { "place": "pl-3a91", "descriptor": "the well" },
  "content": {
    "thing": { "thing_id": "th-401", "kind": "k:person", "roles": [], "descriptor": "Tam", "body": "b-tam" },
    "distance": "near",
    "activity": "drawing water"
  }
}
```

`activity` is an optional vendor-composed human-readable phrase (prose only, AR-3).

**A sighting is a bounded claim about one thing.** It says nothing about what else was or was
not there. That is §4.3's job, and the distinction is load-bearing.

### 4.3 `observation` — a bounded, exhaustive claim about a place

The single most important perception fix promptworld I made (PM-2). I's original sweep only
ever reported what *is* there, which made confabulated place-beliefs **unfalsifiable** — nothing
ever recorded what a place *lacked*.

```json
{
  "percept_id": "p-9c22", "percept_type": "observation", "urgency": "notable",
  "provenance": { "origin": "saw", "source": null, "observed_at": 918240, "received_at": 918240 },
  "place": { "place": "pl-77b", "descriptor": "the north clearing" },
  "content": {
    "extent": "near",
    "vocabulary": ["k:sleeping-place", "k:workbench", "k:fire", "k:water-source", "k:storage", "k:person"],
    "present": [
      { "thing_id": "th-903", "kind": "k:fire", "roles": ["light_source"], "descriptor": "a campfire", "count": 1 },
      { "thing_id": null, "kind": "k:storage", "roles": ["storage"], "descriptor": "chests", "count": 2 }
    ]
  }
}
```

| Field | Req | Meaning |
|---|---|---|
| `extent` | ✔ | Band (AR-4) the scan covered. |
| `vocabulary` | ✔ | The **exact closed set of kinds scanned**. |
| `present` | ✔ | Every thing of a vocabulary kind found, in a stable sorted order. |

**Absence is `vocabulary` minus `present`**, and nothing else. This is the mechanism, and it is
also the mitigation for the sharpest cost red flag in the feasibility pass (F-6):

> Exhaustiveness MUST be defined over a **closed, small, vendor-declared salient-kind
> vocabulary** — never over "everything". promptworld I's version was exhaustive over a 2-D
> tile disc with a ~dozen-kind vocabulary; a 3-D block volume makes literal exhaustiveness
> unbounded and unaffordable. Scoping the claim buys the falsifiability property at a payable
> price.

`vocabulary` MUST be on the wire. Without it the mind cannot know the scope of the absence
claim, and an unscoped absence claim is worse than none.

- The vendor SHOULD suppress a re-observation of an unchanged place within a dedup window
  (I's rule, ported). The window is vendor configuration, not a protocol constant.
- **No `absent` or `absence_of` field.** The world cannot know what an agent expected (PM-3).
  Reconciling an observation against a held belief — confirmation, bounded disconfirmation,
  silence leaving things untouched — is entirely mind-side. This is SI-4 in its sharpest form,
  and it is what keeps a second vendor cheap: a vendor implements observation, not cognition.

### 4.4 `sound` — heard, cause unknown

Hearing is a **GAP**: promptworld I had none, so it is derived, not ported (PM-13). It is also
the least-verified channel against the first vendor's substrate (F-4) — nothing in
[[villager-brain-api]] speaks to sound at all. **This shape therefore specifies hearing
epistemically and MUST NOT presume any particular engine event source**, so that whatever hook
turns out to exist can satisfy it.

```json
{
  "percept_id": "p-9c23", "percept_type": "sound", "urgency": "urgent",
  "provenance": { "origin": "heard", "source": null, "observed_at": 918244, "received_at": 918244 },
  "place": null,
  "content": {
    "sound_kind": "k:snarl",
    "bearing": "behind",
    "distance": "near",
    "descriptor": "something snarling in the dark"
  }
}
```

Normative:

- The content MUST NOT name the thing that made the sound. Not as `thing_id`, not as `kind` of
  the source, not in the descriptor. **A heard thing is direct perception of a *sound*, not of
  its cause** — hearing a mob is not seeing which mob. If the body also saw the causer, that is
  a separate `sighting` percept, and joining them is the mind's inference to make and own.
- `heard` is a distinct origin from `saw` precisely so a mind **cannot launder an inference
  from a noise into a witnessed fact**. Collapsing the two origins would silently defeat the
  classifier.
- `place` is typically `null`: a sound tells you a direction, not a location. `bearing` and
  `distance` are bands (AR-4) and MAY be `null` when the vendor cannot tell.

Occlusion asymmetry — sound passes walls, sight does not — is what makes hearing worth having
at all. The protocol states the posture; fidelity is vendor business (F-2).

### 4.5 `speech` — a speaker said something

```json
{
  "percept_id": "p-9c24", "percept_type": "speech", "urgency": "notable",
  "provenance": {
    "origin": "told",
    "source": { "kind": "body", "body": "b-tam", "descriptor": "Tam" },
    "observed_at": 918250, "received_at": 918250, "hops": 0
  },
  "place": { "place": "pl-3a91", "descriptor": "the well" },
  "content": { "utterance": "I'm not going out there tonight.", "addressed_to": ["b-eda"] }
}
```

`source.kind` ∈ `body` | `person` | `unknown`. **The player is a `person` source, not a body**
— the player has no mind daemon behind the seam, and a player's speech is `told` from a
fallible teller. promptworld I's divine-reveal channel **dies with the guardian and MUST NOT be
quietly re-used as the player's channel** (PM-9b): a player telling a villager where the well is
is a `told` fact carrying a player's fallibility, not a granted truth.

### 4.6 `told_fact` — knowledge passed between agents

The concrete mechanism behind "or was told" (PM-9). A vendor exchanges a small cap of facts per
direction during a founded conversation; the cap and ordering are vendor configuration.

```json
{
  "percept_id": "p-9c25", "percept_type": "told_fact", "urgency": "background",
  "provenance": {
    "origin": "told",
    "source": { "kind": "body", "body": "b-tam", "descriptor": "Tam" },
    "observed_at": 902430, "received_at": 918251, "hops": 1
  },
  "place": { "place": "pl-3a91", "descriptor": "the well" },
  "content": {
    "about_place": { "place": "pl-51c", "descriptor": "the old orchard" },
    "thing": { "thing_id": null, "kind": "k:food-source", "roles": ["food_source"], "descriptor": "apple trees" },
    "assertion": "present"
  }
}
```

`assertion` ∈ `present` | `gone`.

**`observed_at` is the teller's last-seen time, not the telling time** (EH-5). This one field
plus one comparison rule is the entire anti-telephone-game mechanism:

> **A receiver's own fresher knowledge never loses to secondhand.** The mind upserts a told fact
> only where its own knowledge is absent or staler. Secondhand is never fresher than firsthand.

The comparison is mind-side (SI-4); the protocol's obligation is to put an honest teller
timestamp on the wire so the comparison is possible at all.

### 4.7 `text` — read from an artifact

```json
{
  "percept_id": "p-9c26", "percept_type": "text", "urgency": "notable",
  "provenance": {
    "origin": "read",
    "source": { "kind": "artifact", "thing_id": "th-12", "descriptor": "the job board" },
    "observed_at": null, "received_at": 918260
  },
  "place": { "place": "pl-002", "descriptor": "the village square" },
  "content": { "text": "Build a shelter by the north wall. — the player", "attributed_to": "the player" }
}
```

`attributed_to` is what the artifact **claims**, not what the vendor verified. A mind MUST NOT
treat an attribution as a `told` provenance from that party. `observed_at: null` is normal for
an artifact of unknown authorship time, and means maximally stale (§2.2).

### 4.8 `act_result` — the result half of the act surface

Origin is always `acted`. See §5.4.

### 4.9 `self_state` — the body's own condition

```json
{
  "percept_id": "p-9c27", "percept_type": "self_state", "urgency": "urgent",
  "provenance": { "origin": "felt", "source": null, "observed_at": 918270, "received_at": 918270 },
  "place": { "place": "pl-77b", "descriptor": "the north clearing" },
  "content": { "condition": "cold", "level": "severe", "trend": "worsening" }
}
```

`condition` is drawn from an open vocabulary (v0 seeds `hunger`, `injury`, `cold`, `heat`,
`fatigue`, `encumbered`, `threatened`); minds MUST tolerate unknown conditions. `level` per
AR-6. `trend` ∈ `worsening` | `steady` | `easing` | `null`.

This channel is where `urgent` earns its keep: promptworld I kept nearly every entry *below* its
interrupt threshold, with one deliberate exception — "you are dangerously cold and have done
nothing about it" — whose entire job was to break a mis-scheduling mind out of a loop (PM-12).

### 4.10 `change_report` — and the 75% warning

The diff channel: something changed at a place, reported to a body that was not there.

```json
{
  "percept_id": "p-9c28", "percept_type": "change_report", "urgency": "notable",
  "provenance": { "origin": "saw", "source": null, "observed_at": 918300, "received_at": 918300 },
  "place": { "place": "pl-51c", "descriptor": "the old orchard" },
  "content": {
    "change": "gone",
    "thing": { "thing_id": "th-77", "kind": "k:tree", "roles": [], "descriptor": "the big oak" }
  }
}
```

`change` ∈ `appeared` | `gone` | `altered`.

**The delivery restriction is the whole point of this section, and it is normative:**

> A vendor **MUST NOT** deliver a `change_report` to a body that caused the change, or that was
> in range to witness it. Those bodies were already told: the actor by its `act_result`
> (§5.4), the witnesses by a `sighting` at the moment it happened.

This is a design warning encoded rather than a mechanism copied (PM-6). promptworld I shipped
the naive version first: completing a harvest cleared the world but left the stale fact in every
map — *including the actor's own* — so the ambient correction sweep "corrected" the map moments
later and minted a third-person-voiced loss memory for the agent who swung the axe, and for
every bystander who watched it fall.

> **Measured live in promptworld I: 75% of all memories formed came from that channel.**

The figure is carried here because it is the kind of finding that is disbelieved until it is
cited. Three quarters of a mind's memory — and therefore three quarters of the token budget its
recall spends — was bookkeeping noise voiced as experience. The generalized rule: **a percept
channel that reports a change to someone who caused or watched the change is a
memory-flooding bug, not a feature.** With the restriction in place, this channel fires only for
bodies that were absent, asleep, or out of range — the genuine return-and-discover narrative it
was built for.

### 4.11 Budget and shedding

Ambient perception is bounded and prioritized (PM-7). Percepts crossing a process boundary and
eventually reaching a token budget cost far more here than in-process events did in I.

- Cadence, per-beat caps, and staggering are **vendor concerns**. The protocol MUST NOT expose a
  tick number, a sweep cadence, or a per-beat cap (PM-7 adaptation — I's knobs were
  tick-denominated and justified by hot-path relief in a fixed-timestep loop that Kithcraft does
  not have).
- Under pressure a vendor MAY shed `background` percepts. It MUST NOT shed `urgent` ones.
- An `observation` MUST be shed whole or delivered whole. A partially delivered observation is a
  false absence claim (§4.3).

---

## 5. The **act** surface (mind → world)

### 5.1 The intent/result split

**The mind sends intents; the body reports results — and results are themselves percepts,
carrying provenance like any other.** (F-8.)

This is not a stylistic preference. Under SI-1 the mind has no read access to the world, so
there is no such thing as an act that "returns" world state. What a mind learns from acting, it
learns the same way it learns everything: something arrives in the inbox, stamped. An act that
returned a value would be a query wearing a costume, and would reopen the omniscience hole.

```
mind ──intent──▶ vendor          (a request; may be refused)
mind ◀─intent_ack── vendor       (received / admissible — NOT success)
mind ◀─percept(act_result)── vendor   (what actually happened; origin: acted)
```

### 5.2 `intent`

```json
{
  "intent_id": "i-4410",
  "verb": "go_to",
  "target": { "type": "place", "place": "pl-51c" },
  "reason": "to see whether the orchard still has fruit before dusk",
  "supersedes": "i-4402",
  "not_after": 918900
}
```

| Field | Type | Req | Meaning |
|---|---|---|---|
| `intent_id` | string | ✔ | Mind-issued, unique within the session. |
| `verb` | string | ✔ | From the vendor's declared verb set (§5.5). |
| `target` | object \| null | ✔ | `{type: "place"\|"thing"\|"body"\|"none", ...}` per the verb's declared target shape. |
| `reason` | string \| null | ✔ | **Mind-authored, opaque to the vendor.** Echoed back on the `act_result` and nowhere else. |
| `supersedes` | string \| null | — | An earlier `intent_id` this replaces. |
| `not_after` | integer \| null | — | World time past which the intent is pointless; the vendor SHOULD abandon it with `expired`. |

`reason` is PM-10's second half, placed where it belongs: the "why" comes from the mind's own
act, so **the mind joins it; the vendor never invents a why**. A vendor MUST NOT parse, act on,
or branch on `reason` — it exists so the resulting memory is situated rather than a bare log
line, which is a large part of what keeps an LLM villager from sounding like a chatbot.

**Targets name things the mind already knows.** A `place` or `thing_id` in a target MUST be a
token the mind received in a percept. The vendor resolves the token against ground truth — that
is the legitimate reflex half of the split (PM-1) — but it resolves it *for a mind that named a
place it already knows*. A vendor MUST NOT accept a target expressed as a description to search
for ("the nearest bed"), because that is a world query in intent form.

### 5.3 `intent_ack`

```json
{ "intent_id": "i-4410", "accepted": true, "reason_code": null }
```

Acknowledges receipt and admissibility only. **Acceptance is not success**; the only report of
what happened is the `act_result` percept. The ack exists because without it a mind cannot
distinguish "still walking" from "message lost".

`reason_code` on refusal ∈ `unknown_verb` | `unknown_target` | `malformed` | `unsupported_version`
| `busy`.

### 5.4 `act_result` (a percept, §4.8)

```json
{
  "percept_id": "p-9d02", "percept_type": "act_result", "urgency": "notable",
  "provenance": { "origin": "acted", "source": null, "observed_at": 918410, "received_at": 918410 },
  "place": { "place": "pl-51c", "descriptor": "the old orchard" },
  "content": {
    "intent_id": "i-4410",
    "verb": "go_to",
    "outcome": "completed",
    "reason_code": null,
    "reason": "to see whether the orchard still has fruit before dusk",
    "detail": "walked to the old orchard"
  }
}
```

`outcome` ∈ `completed` | `failed` | `interrupted` | `superseded` | `expired`.
`reason_code` (on non-completion) ∈ `unreachable` | `target_gone` | `not_capable` | `blocked` |
`interrupted_by_urgency` | `refused_by_world`. Receivers MUST have a fallback branch for
unrecognized codes (§7).

Every intent that was acked as accepted MUST eventually produce exactly one `act_result`. A
vendor that drops one leaves a mind waiting forever with no channel to ask.

The actor's own act is attributed **to the actor, at the moment it happens** — this percept —
which is precisely what makes §4.10's delivery restriction implementable (PM-6).

### 5.5 Verbs — a small core plus a vendor-declared set

The abstraction rule (§3) forbids the protocol from enumerating a world's affordances. But a
mind portable across vendors needs *some* fixed floor. The split:

**Core verbs — every vendor MUST implement:**

| Verb | Target | Meaning |
|---|---|---|
| `go_to` | `place` \| `thing` \| `body` | Move to a known place. Targeting a `body` resolves to that body's **last-seen** place, never its live position (PM-8) — a villager walks honestly to where it last saw someone and finds them gone. |
| `speak` | `body` \| `person` \| `none` | Say something. Content is the mind's text. |
| `attend` | `place` \| `none` | Look deliberately at where the body is; SHOULD produce an `observation` (§4.3). |
| `wait` | `none` | Do nothing on purpose, for a stated duration or until superseded. |

Four verbs is enough to exercise the whole seam end-to-end, which is what Phase 3's test vendor
will need.

**Extended verbs** are declared in the capability manifest (§6.2) with their target shapes.
A mind MAY use only declared verbs; an undeclared verb is refused with `unknown_verb`. Adding a
verb is an additive change (§7).

### 5.6 The one live-check exception, resolved without a query

PM-8 kept peer positions last-seen but made **death** a live check, because acting on a
remembered-alive peer is a correctness bug rather than drama. A live check is a world query,
which SI-1 forbids. The resolution:

> **The mind learns of death by failing to act, not by asking.** An intent targeting a body
> that no longer exists MUST fail with `outcome: "failed"`, `reason_code: "target_gone"` — and
> that failure is a percept with `origin: "acted"`, which is honest: the body found out by
> trying.

This keeps the correctness property without opening a read path, and it means vendors do not
invent their own exception classes. `target_gone` is the **only** sanctioned discovery channel
for a referent's non-existence.

### 5.7 `cancel`

```json
{ "intent_id": "i-4410" }
```

Requests abandonment. Produces an `act_result` with `outcome: "superseded"`. A vendor MAY
complete an already-committed action anyway and report `completed`; cancellation is a request,
not a guarantee.

---

## 6. The **remember** surface (the mind's durable-memory contract)

The third surface is the least like the other two: it is mostly a **statement of what does
*not* cross**. Memory belongs to the mind (SI-5), so the protocol's job here is to define the
persistence split, the session boundary, and the floor on what a mind may durably claim.

### 6.1 Who stores what

| Concern | Owner | Notes |
|---|---|---|
| Percept history, beliefs, relationships, personas, plans | **Mind** | The vendor MUST NOT persist these, MUST NOT require them to replay, and MUST NOT rank or weight them. |
| Body existence, identity binding, physical condition, inventory, position | **Vendor** | The world's state is the world's business. |
| The **token registry** — `body`/`place`/`thing_id`/`kind` → referent | **Vendor** | Persisted across sessions per §2.3. This is the vendor's hardest obligation. |
| The **resolution index** — which token is where, what is at a place *right now* | **Vendor** | Ground truth, used for reflex target resolution (§5.2). Never readable by the mind. |
| The mind's **private map** — known places and facts, each with provenance and a last-seen time | **Mind** | This is PM-1's spine: two villagers see different worlds. |

**The two-store distinction is deliberate and easy to collapse by accident.** The vendor's
resolution index and the mind's private map look similar and are not: one is ground truth used
to steer a body, the other is remembered, provenance-stamped, possibly-wrong knowledge used to
decide. A vendor that lets a mind read the first has reintroduced omniscience no matter how the
protocol is worded.

The feasibility pass adds a vendor-side recommendation that this contract deliberately does
**not** mandate: the first vendor SHOULD keep its store in mod-owned code rather than in the
engine's own memory modules (F-3). **Nothing in this protocol requires a villager's knowledge to
live in the engine's brain**, and binding it there would spend owned Mixin surface for no seam
benefit while making the epistemic layer non-portable to a second vendor.

### 6.2 `session_open` — the handshake

```json
{
  "protocol": "0.1",
  "message": "session_open",
  "session": "s-7f3a",
  "seq": 0,
  "body": "b-eda",
  "world_time": 918233,
  "payload": {
    "time_unit": "second",
    "continuity": {
      "previous_session": "s-6b02",
      "previous_close_world_time": 902100,
      "body_continuous": true
    },
    "capabilities": {
      "percept_types": ["sighting", "observation", "sound", "speech", "told_fact", "text", "act_result", "self_state", "change_report"],
      "origins": ["acted", "saw", "heard", "felt", "told", "read"],
      "verbs": [
        { "verb": "go_to", "targets": ["place", "thing", "body"] },
        { "verb": "speak", "targets": ["body", "person", "none"] },
        { "verb": "attend", "targets": ["place", "none"] },
        { "verb": "wait", "targets": ["none"] },
        { "verb": "carry", "targets": ["thing"] }
      ],
      "salient_kinds": [
        { "kind": "k:sleeping-place", "roles": ["shelter", "sleeping_place"], "descriptor": "a bed" },
        { "kind": "k:fire", "roles": ["light_source", "gathering_place"], "descriptor": "a fire" }
      ],
      "bearings": ["ahead", "behind", "left", "right"],
      "distance_bands": ["here", "near", "middling", "far"]
    }
  }
}
```

The manifest is how AR-2's opacity stays workable: a mind learns the vendor's vocabulary at
session start, in role-annotated form, and never has to know what a token spells.

### 6.3 Session continuity — the gap is a gap

`continuity` reports **that** the body was unattended and for how long. It MUST NOT report what
happened during the gap.

- `body_continuous: false` means this is a different body (the previous one died, or the vendor
  cannot vouch for continuity). The mind is entitled to know its body changed; it is not
  entitled to a free account of the interval.
- A vendor MAY, after resume, emit ordinary `change_report` percepts for things the body finds
  changed — this is the legitimate return-and-discover case (§4.10), since the body was
  genuinely absent. Those percepts MUST carry `received_at` at resume time and an honest
  `observed_at` (usually resume time, since the body is discovering it now, not witnessing when
  it happened).
- The vendor MUST NOT backfill the gap as though the body had been watching. Epistemic hygiene
  does not suspend at a process boundary; a mind that comes back knowing what it did not see is
  omniscient by the back door.

`session_close` carries `{ "reason": "shutdown" | "body_lost" | "error", "detail": null }`.

### 6.4 What a mind may durably claim

These constrain the mind, not the vendor. They are stated here because the seam is where they
become checkable — the classifier and the origins are protocol property, so a fake vendor
(Phase 3) can prove them without a compiler.

| # | Rule | Source |
|---|---|---|
| **RM-1** | A mind MUST NOT durably assert direct perception of anything not backed by a retained percept whose `origin` satisfies `direct_perception` (§2.7). The gate reads the origin and never the prose. | EH-3 |
| **RM-2** | A model-authored belief MUST cite the specific percepts it rests on. A deterministic gate resolves the citations and **coerces** the claim's provenance down to what the evidence supports: a "witnessed" claim with no direct percept behind it degrades to "told"; with nothing resolvable, to "inferred". | EH-4 |
| **RM-3** | The gate **coerces, it does not reject.** Rejection throws away a whole consolidation over a bookkeeping error; coercion keeps the content and downgrades the epistemic claim. The count of coercions is recorded. | EH-4 |
| **RM-4** | Secondhand never overwrites fresher firsthand. Upsert a told fact only where own knowledge is absent or staler (§4.6). | EH-5 |
| **RM-5** | Stored confidence never changes with time. **Effective** confidence is computed at read time from age and a half-life; below a floor a belief stops driving behaviour but stays revisable rather than being deleted. The half-life for convictions is far longer than for memory vividness — convictions outlive vividness. | EH-6 |
| **RM-6** | Freshness is evaluated **at read time, per kind**: fresh iff `now − last_seen < horizon(kind)`. Time never mutates a fact. Horizons are mind configuration, not protocol constants (§2.2). | PM-4 |
| **RM-7** | **No silent forgetting.** Knowledge changes only through a recorded channel: a percept, a telling, or a witnessed removal. Staleness hides a fact; it never deletes one. Only a correction, a death, or a witnessed removal deletes. | PM-5 |

RM-5 and RM-6 have a seam consequence worth naming: **a mind that mutates nothing on a timer has
no clock to keep in sync with the vendor's.** promptworld I's original reason for read-time-only
decay was snapshot churn under determinism-for-replay, which dies with I; this reason is better
and is the one to keep.

RM-7 has a testability consequence for Phase 3: because every knowledge change has a named
channel, a fake vendor can drive a mind's entire epistemic state **by scripting percepts alone**
(PM-5).

---

## 7. Versioning

### 7.1 The field

Every message carries `protocol: "MAJOR.MINOR"`. v0 is `"0.1"`.

Negotiation is minimal and fails closed: the vendor states its version in `session_open`; the
mind either proceeds or replies `session_close` with `reason: "error"`, `detail:
"unsupported_version"`. There is no downgrade dance in v0 — a mind that cannot speak the
vendor's version does not get a body.

### 7.2 Additive (MINOR bump)

- A new **optional** field on an existing shape.
- A new `percept_type`.
- A new `origin` value.
- A new verb (core stays fixed; extended verbs are manifest-declared and adding one is free).
- A new `kind` token, `role`, `condition`, `sound_kind`, or `reason_code`.

### 7.3 Breaking (MAJOR bump)

- Removing or renaming any field; making an optional field required.
- Narrowing a type, or changing the meaning of an existing enum value.
- **Changing `DIRECT_ORIGINS` membership** (§2.7) — this silently rewrites what every stored
  belief was entitled to claim.
- Changing the meaning of an urgency band.
- Weakening a token-stability guarantee (§2.3).
- Removing the `change_report` delivery restriction (§4.10) or the `vocabulary` field (§4.3) —
  both look like fields and are actually invariants.

### 7.4 Receiver rules that make additive changes safe

| # | Rule |
|---|---|
| **V-1** | Receivers MUST ignore unknown fields and MUST NOT reject a message for containing them. Senders MUST NOT rely on unknown fields being preserved or echoed. |
| **V-2** | Receivers MUST have a fallback branch for unknown enum values in open vocabularies (`role`, `condition`, `kind`, `sound_kind`, `reason_code`) — an unrecognized value is a value not acted on, not an error. |
| **V-3** | An unknown `percept_type` MAY be retained and MUST NOT be interpreted. A mind MUST NOT guess a percept's meaning from its fields. |
| **V-4** | An unknown `verb` is refused with `unknown_verb` (§5.3). |
| **V-5** | A **missing required field is malformed** and is rejected at the seam, never defaulted (EH-2a — the compiler guarantee that made unstamped emission impossible in promptworld I does not survive a cross-process seam, so rejection replaces it). |
| **V-6** | **Unrecognized or absent `origin` classifies secondhand, never direct** (EH-2b). This is the one place where a conservative default is mandated rather than rejection, and it is what makes §7.2 safe: a new origin added in a minor version is classified secondhand by an older mind — which is the harmless direction of error. An older or sloppier vendor is exactly the case where this field goes missing, so the rule gets *more* load-bearing under versioning, not less. |

V-5 and V-6 are not in tension: a strict implementation rejects the malformed percept outright;
any implementation that chooses leniency at the boundary MUST fall to secondhand rather than to
direct. Phase 3's fake vendor is where both are proven, since there is no compiler to prove them.

---

## 8. Transport — open question, with constraints

**Not decided in v0.** Per spec 002's out-of-scope list, the wire choice (gRPC, WebSocket,
stdio, an in-process channel for the test vendor) is named here as an open question. The shapes
above are defined as data and MUST remain serialization-neutral; anything that can carry ordered
messages with these fields is a candidate.

Constraints any candidate must satisfy:

| # | Constraint |
|---|---|
| **T-1** | **Push, not pull, for percepts.** The percept stream is one-directional. A transport whose natural idiom is request/response for world state invites SI-1 violations and must be used against the grain. |
| **T-2** | **Ordered per body**, or percepts must be reorderable by `seq`. `percept_id` supports dedup under at-least-once delivery. |
| **T-3** | **Long-lived sessions** with explicit open/close, carrying the handshake of §6.2. |
| **T-4** | **A mind restart must not require a vendor restart**, and vice versa. The session boundary is the recovery unit; §6.3's continuity report is how a mind rejoins. |
| **T-5** | **Message-oriented**, with a schema the fake vendor can produce without embedding a game engine. |
| **T-6** | **Backpressure that can shed `background` only** (§4.11), and that never splits an `observation`. |
| **T-7** | **Process-separable but not process-required.** The test vendor (Phase 3) should be able to sit in-process behind the same shapes; the seam is a contract, not a network hop. |

---

## 9. Open questions carried forward

| # | Question | Notes |
|---|---|---|
| **Q-1** | Transport (§8). | Constraints stated; decision deferred. |
| **Q-2** | Hearing's engine hook in the first vendor. | F-4: no verified vanilla surface. §4.4 is deliberately written so any hook satisfies it; verification is implementation work and should be carded. |
| **Q-3** | The entity implementation (TASK-0003, parallel lane). | This protocol is neutral to it by construction — channels are defined against "a villager-shaped body", never against a specific entity type. |
| **Q-4** | Conversation founding and turn-taking. | v0 carries `speech` and `told_fact` percepts and a `speak` verb; who may talk to whom, and when a conversation is "founded", is vendor policy in v0. |
| **Q-5** | Multi-body minds / one mind attending several bodies. | The envelope's `body` field admits it; nothing else in v0 supports it. Left open, not designed. |
| **Q-6** | Vendor-declared verbs with rich parameters (crafting recipes, building plans). | v0's `target` shape is deliberately thin. The brief's job-board blueprint (decision 7) will press on this; the `read` channel carries the blueprint as text in v0. |

---

**Next (Phase 3, not this document):** the second-vendor sketch (R3), the fake/test body vendor
spec (R4), a leak sweep against §3, and re-verifying [[body-protocol-seam]] with this file as a
source.
