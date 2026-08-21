# Death mechanics — what kills, what remains, how the living remember

**Spec:** 005-death-mechanics · **Phase:** 2 (design). Ratified constraints this design must
not relitigate: permadeath is real and should sting, replenishment is punted in v1
([[design-brief]] #4, #10); the augmented villager is `VillagerEntity` (decision-0002); memory
crosses the seam as percepts with provenance ([[body-protocol-seam]]). All facts cited below
are carried from `specs/005-death-mechanics/research/death-surface.md` (Phase 1, checked live
2026-08-21) by section reference — not re-derived.

## 1. Causes of death — admitted vs. suppressed

The governing test for every ruling: **does this death read as a consequence of neglect
(preventable, base-building-relevant) or as arbitrary bad luck?** The thesis wants the former
(it stings *because* it was earnable) and rejects the latter (arbitrary death is cheap, not
sad).

| Source | Ruling | Argument (cites death-surface.md) |
|---|---|---|
| Hostile-mob night targeting (zombie/husk/drowned melee) | **ADMIT** | The central admitted danger. §2 confirms targeting, panic/flee, and door-breaking are all inherited free from `VillagerEntity`. This is precisely the mechanic the brief's v1-demo line means by "the player's walls and torches protect their friends" — a death here is legible as "the wall had a gap" or "the door wasn't reinforced," which is neglect, not luck. |
| Zombie-villager conversion | **ADMIT, ruled equivalent to true death in v1** | §2 gives 0%/50%/100% conversion odds by difficulty; §3 notes the death sound doesn't distinguish conversion from death, and `die()`/`releaseAllPois()` fires either way. A converted friend wandering as a mindless hostile is *different, not lesser* emotional material than a corpse — but curing a zombie villager back is not established anywhere in the Phase 1 evidence, and inventing a curing path here would quietly undo "permadeath is real" without a ratified reason to. Ruling: conversion is treated identically to death at the body-protocol layer (a new body token, `body_continuous: false` for that identity, memories close out) for v1. If a later pass verifies a curing mechanic worth designing around, that is a new ratified decision, not a default this doc assumes. |
| Falls | **ADMIT** | §1: standard ~1HP/block beyond 3-block safe distance, no villager exemption, free from `VillagerEntity`. Villager pathfinding doesn't walk them off ledges by default, so a fall death reads as "built near an unfenced drop" — neglect, and zero Mixin cost either way. |
| Drowning | **ADMIT** | §1: villagers are not on the drowning-immune list (aquatic/undead/iron golems only); standard breath meter and damage. A preventable consequence of building near water without care — on-thesis, free. |
| Fire / lava | **ADMIT** | §1: no villager-specific immunity found. Same base-building-diligence argument as falls/drowning (don't build next to lava, put out fires) — free, preventable, on-thesis. |
| Hunger / starvation | **N/A — not a death vector to rule on** | §1 confirms there is no starvation-damage loop on `Villager` at all; `foodLevel` only gates breeding willingness. There is nothing to admit or suppress. This absence is load-bearing for §4 below: the single most common companion-game micromanagement vector (feeding) simply does not exist as a survival requirement. |
| Player direct kill (friendly fire) | **ADMIT, no engine guardrail added** | §1: unrestricted mechanically; a direct kill generates `major_negative` gossip (weight −5) broadcast to witnesses within 16 blocks. The embodied-player decision ([[design-brief]] #1) means the player has real physical agency, including the capacity to do real harm — adding a guardrail would be a fiction about who the player is. What this death *means* to survivors is carried by §3's memory channel, not by vanilla's reputation system (see the reputation note in §3). |
| Indirect player-caused kill (lure into lava/fall/suffocation) | **ADMIT, flagged loophole** | §1: indirect kills generate **no** gossip at all, because vanilla's reputation system requires "the villager can determine the source." This is a vanilla loophole, not a design choice — a player can cause a death with zero in-fiction social consequence via vanilla's own reputation bookkeeping. §3's memory carry does not have this loophole (a witnessed death is a witnessed death regardless of cause), which is the reason that channel, not vanilla reputation, is the one this design leans on. |
| Zombie sieges | **SUPPRESS (recommend Mixin-suppress the trigger)** | §2: a *separate* Java-only event, ~20 zombies spawning over 60 ticks at 32 blocks, ignoring the normal player-proximity spawn restriction — on top of, not instead of, ordinary targeting. §2's own flag notes the village-eligibility thresholds (door/population count) are inconsistently documented and unverified against the target MC version. Independent of whether a 3-villager household even qualifies: an event that can overwhelm well-built defenses with a once-a-night dice roll is exactly the "arbitrary, cheap death" this doc's governing test rejects — it defeats the "your walls failed" legibility that makes ordinary night danger on-thesis. Ordinary per-mob targeting and raids stay admitted; the siege escalation on top of them does not. TASK-0006 verifies the Mixin-suppression point and whether a 3-villager cast would ever trigger one at all. |
| Raids (Bad Omen, illager patrols) | **ADMIT, but flag as likely outside v1-demo scope** | Player-triggered (attacking a raid patrol / Bad Omen), so it is player agency creating stakes at home, on-thesis. Whether "one real evening" ([[v1-demo]]) includes raid content at all is a build-plan (TASK-0006) scoping call, not a permadeath-mechanics ruling — noted here so TASK-0006 doesn't have to re-derive it. |

## 2. Remains

Governing test: vanilla leaves almost nothing behind (§3: no drops except two narrow farmer/
armor exceptions, hidden inventory lost outright even with `keepInventory`). **Free is
invisible** — a death that leaves no trace costs the thesis nothing to suppress and gains it
nothing to admit, so remains have to be *authored*, not inherited.

### Grave

**Who digs it:** nobody is ordered to — there is no direct control ([[design-brief]] #6). Two
things happen, not one, so the marker never depends on a busy or resentful villager's
compliance:

1. **Guaranteed baseline, mod-placed.** On death the mod immediately places a modest marker
   (a mound/sign naming the villager) at the death location, with no villager agency required.
   This is a reliability floor: with a 2–3 survivor cast, waiting on someone to volunteer risks
   the marker never appearing at all, and the memorial gesture is the entire point of this
   section — it cannot be optional.
2. **Optional elaboration, villager-driven.** The mod also posts a job-board entry ("tend
   Tam's grave" / similar) that a survivor can pick up on their own initiative — a grieving
   close friend takes it readily; a rival or a stranger ignores it. That refusal or acceptance
   *is* the character material ([[design-brief]] #6: reluctance and grumbling are relationship,
   not bugs), and it rides the existing job-board mechanism (#7) rather than inventing a new
   one.

**Placement:** at the actual death location, defaulting to the nearest safe, buildable surface
if death occurred somewhere a marker can't sensibly sit (mid-water, in lava). A grave where
someone actually died ("she died holding the north wall") carries more weight than a
generic cemetery plot, and it is a `place` token like any other — a mind can `go_to` it later
("visit Tam's grave") using the existing core verb (body-protocol-v0 §5.5), no new affordance
needed.

### Belongings

**Design overrides vanilla here, deliberately.** §3's finding — everything in a villager's
hidden inventory is destroyed outright on death, not dropped — is exactly the kind of
invisible default that costs the thesis a real beat for free. Ruling: the mod SHOULD capture
the dying villager's hidden inventory before vanilla's loss and place it as a "belongings
bundle" (a chest or equivalent) at the grave site, tagged with the owner's name. This needs no
protocol extension: it is an ordinary `k:storage`-kind thing with `roles: ["storage"]` and a
descriptor like "Tam's things" (body-protocol-v0 §2.5, AR-3) — survivors and the player
perceive it exactly like any other storage container, just one that means something.

### World markers

§3 confirms `die()` calls `releaseAllPois()` — the bed and job-site claim are released back to
the world for free, no engine work required. The design choice is what to do with that for-free
release, not how to get it:

- **Leave the bed and job-site unclaimed for a grief period** before another villager (or the
  mod) reassigns them, rather than instantly backfilling. Instant reassignment is the same
  "free-is-invisible" trap as the drops above — an empty bed the player notices while walking
  past is a beat; an efficiently-recycled one is bookkeeping. The length of that period is a
  tuning knob for TASK-0006, not a number this doc invents (per plan.md's own risk note: state
  the posture, not fake precision).
- **Whether POI re-claim has any natural lag after `releaseAllPois()` is unverified** —
  §3 establishes the release happens, not its claim-eligibility timing. Flagged for TASK-0006
  to check before relying on "instant" being the actual default to override.

## 3. Memory and conversation carry

**No protocol extension is proposed or needed.** Every shape below already exists in
body-protocol-v0; the design work is choosing which existing percept a witness or a survivor
gets, not inventing a new one.

- **A witness's own experience of a death is an ordinary `sighting` sequence, not a special
  "death" percept type.** A villager who sees a friend attacked gets `sighting` percepts
  (`thing`: the dying villager, `doing`: "being attacked" / "collapsing", `origin: "saw"`) the
  same way it would see anything else happen nearby (body-protocol-v0 §4.2). There is
  deliberately no magic death-broadcast: the witness saw it happen, moment to moment, which is
  what makes the resulting memory durably "witnessed" under RM-1 rather than a bookkeeping
  event.
- **An absent survivor learns of a death through the channel that already exists for exactly
  this case.** §4.10's `change_report` restriction fires only for bodies that were absent,
  asleep, or out of range — precisely a survivor who wasn't there. They get a `change_report`
  (`thing`: the person, `change: "gone"`) on return, plus an ordinary `sighting`/`observation`
  of the new grave-thing at the place they return to. Neither requires a new percept type.
- **Being told at dusk is an ordinary `speech`/`told_fact`, correctly secondhand.** A witness
  relaying the death to an absent villager at dusk conversation is `origin: "told"` from that
  witness — RM-1/RM-2 correctly prevent the told-to villager from ever durably claiming they
  saw it. This is the seam doing its job, not a gap.
- **Vanilla's death-triggered reputation/gossip system (§1, §3: `major_negative`, weight −5,
  16-block witness broadcast) is a separate, narrower mechanism from this memory channel and
  should not be conflated with it.** Reputation gates trading prices; it is silent on indirect
  kills (the §1 loophole) and says nothing at all about non-player-caused deaths. The
  conversational memory of a death — grief, gossip, stories — rides the percept/memory layer
  above, which fires for every death regardless of cause and is what actually serves the
  loneliness-cure thesis. The two systems may both be true of the same event without needing to
  agree.

**Dusk-conversation surfacing and decay — posture, not algorithm.** promptworld I's old
salience table rated a witnessed death as its single highest band (body-protocol-v0 §2.8 cites
it directly: "witnessed death — 10★"). That table is gone with I and formativeness is
deliberately mind-side now (SI-5), but the underlying intuition should carry into whatever
consolidation the mind daemon (TASK-0004) implements: **a recent death should surface
disproportionately in dusk conversation for a while, then fade in retrieval frequency as
formativeness/recency weighting shifts — never be silently deleted (RM-7: only a correction, a
death, or a witnessed removal deletes a fact; time alone never does).** The dead stay
conversationally alive, and fade; this document states that posture and leaves the half-life,
retrieval weighting, and consolidation mechanism to TASK-0004, which owns durable memory
(SI-5).

**One flag for the seam's owner, not decided here:** what happens to a dead villager's own mind
process — archived, terminated, something else — is a mind-daemon lifecycle question, not a
per-body-session question this protocol's scope covers. Flagging for TASK-0004 rather than
ruling on it.

## 4. The micromanagement spell-breaker (card AC #2)

**Failure mode named:** feeding schedules, escort missions, and constant vigilance turning
company into chores — the brief's spell-breaker list names this explicitly.

**Why the design mostly doesn't have to answer it:** self-preservation competence already
exists, for free, verified in Phase 1:

- **Villagers cannot starve** (§1) — the single most common companion-game micromanagement
  vector (feeding) is not a survival requirement at all, by vanilla's own construction. Nothing
  to add.
- **Panic/flee is autonomous** (§2: `PanicTask`, triggered by the brain's own `HURT_BY`/
  `NEAREST_HOSTILE` memory) — villagers flee danger on their own initiative, including into
  houses.
- **Door use and nightly shelter-seeking are autonomous** (§2: pathfinding opens/closes doors;
  the ratified wake/work/socialize/sleep schedule ([[v1-demo]]) already drives villagers to
  their claimed bed at night via the inherited POI system).

**What the design adds, beyond leaving this alone:** nothing new to the villager's own
behavior. The one thing this doc adds is the §1 ruling to **suppress zombie sieges** — not
because sieges threaten a well-defended household (they don't, uniquely; ordinary targeting
does too), but because an event that can kill a competent, well-walled villager regardless of
the player's building is functionally indistinguishable from a micromanagement demand: it
teaches the player that walls don't help and only constant intervention does. Suppressing it
keeps the fragility posture consistent — **death is a function of neglect, not dice** — which
is what makes "the player manages flow; drama emerges from interactions" ([[design-brief]] #6)
still true under night danger.

**What the player's protective play looks like:** ordinary Minecraft base-building — walls,
torches to suppress hostile spawns, hung and ideally recessed doors (§2's documented
zombie-door-break counter-strategy) — asked to do double emotional duty, not a second system
layered on top. No feeding UI, no escort quest, no babysitting surface is introduced by this
design.

## 5. Shrinking-cast consequence (card AC #3)

A permanently shrinking cast (replenishment punted, [[design-brief]] #10) means a 3-villager
demo that loses one becomes a 2-villager evening for the rest of that session. **Accepted, not
scoped — with the mitigation living elsewhere, not here.**

- [[v1-demo]] defines "one real evening" as a single-session showcase, not an ongoing
  campaign. A death from a legible, preventable cause (a wall gap, a lava mistake) partway
  through that evening is on-thesis dramatic material for exactly this kind of demo — it makes
  the rest of the evening more poignant, not less demonstrable.
- The brief's "~3–6, a household not a city" framing ([[design-brief]] #4) means a single death
  still leaves company — two villagers is still not alone. Only *cascading* deaths in a short
  demo would threaten the thesis, and that risk is what §1's and §4's fragility rulings
  (suppressed sieges, neglect-not-luck causes) are built to keep low.
- If a scripted/recorded demo run specifically needs to avoid losing a cast member mid-take,
  the correct lever is **demo-config danger tuning at TASK-0006** (fewer hostile spawns,
  daytime-only staging, whatever knob the build plan names) — not a change to the permadeath
  rules themselves, and not a reopening of the replenishment punt.

## 6. Design checks (Phase 3 — ratification prep)

### 6.1 Cross-check against the brief's spell-breakers and [[v1-demo]]

| Spell-breaker / definition | Check | Result |
|---|---|---|
| **Tedious interactions** | Does any death mechanic add a required player action? | **PASS.** The grave marker is mod-placed unconditionally (§2); the belongings bundle is captured automatically; the only player-facing addition is an *optional* job-board entry a survivor may or may not pick up — the player is never required to do anything to get the memorial beat. |
| **Micromanagement** | Does keeping a villager alive require feeding/escort/vigilance? | **PASS**, argued in full in §4: no starvation loop, autonomous panic/flee/door-use/shelter-seeking all inherited free. The one addition (suppressing zombie sieges) *removes* an arbitrary-death vector rather than adding a management surface. |
| **Politeness-policing** | Does any death rule make a villager gate or judge the player's conduct? | **PASS.** §1's ruling on player direct/indirect kills is explicit: no engine guardrail, no reputation gate added by this design — the embodied player's capacity for real harm is left alone, and its social weight rides the existing memory/gossip channels, not a new judgment layer. |
| **[[v1-demo]] definition** ("one real evening," walls protect friends, dusk conversation) | Does the design serve or contradict the demo's required texture? | **Serves it.** §1's admitted-causes table is built around exactly the demo's stated mechanic ("vanilla night danger means the player's walls and torches protect their friends"); §3's memory carry is what makes the dead surface in the demo's dusk conversation. Nothing in this design asks for anything outside the demo's three-villager, one-evening frame. |

### 6.2 What TASK-0006 inherits

**Settled here — TASK-0006 implements as-is, no further design call needed:**

- Admitted causes: hostile-mob night targeting, zombie-villager conversion (ruled equivalent
  to death), falls, drowning, fire/lava, player direct and indirect kills — all admitted per §1,
  no suppression beyond sieges.
- Suppressed cause: zombie sieges (recommend Mixin-suppress the trigger) — §1.
- Grave: mod-placed baseline marker at the death location (or nearest safe surface) +
  optional job-board "tend the grave" entry — §2.
- Belongings: capture the dying villager's hidden inventory before vanilla destroys it;
  place as a named, `roles: ["storage"]` bundle at the grave — §2.
- World markers: bed/job-site release via `releaseAllPois()` is free; hold them unclaimed
  for a grief period before reassignment — §2.
- Memory/conversation carry: witness sighting, absent-survivor `change_report`, dusk
  retelling as `speech`/`told_fact` — all existing percept shapes, no new plumbing — §3.
- Fragility posture: rely entirely on existing autonomous villager competence (no feeding
  UI, no escort quest) — §4.
- Shrinking cast: accepted as on-thesis dramatic material, not mitigated by default — §5.
- **Token discipline:** when a villager dies or converts, the vendor MUST issue a new body
  token for the grave/converted-mob identity and MUST NOT reissue or reuse the dead
  villager's retired `body` token (body-protocol-v0 §2.3) — an implementation obligation the
  design relies on but does not itself enforce.

**Open — TASK-0006 must resolve or verify before relying on it:**

1. Verify the Mixin-suppression approach for zombie sieges, and whether a 3-villager cast
   would ever satisfy the (unverified, version-dependent) village-eligibility thresholds
   that trigger one at all (§1).
2. Decide whether raid content (Bad Omen, illager patrols) is in v1-demo scope at all (§1)
   — flagged here, not ruled on.
3. Set the grief-period duration before bed/job-site reassignment — a tuning knob this doc
   deliberately leaves unnumbered (§2).
4. Verify whether POI re-claim has any natural lag after `releaseAllPois()`, before relying
   on "instant" being the actual default this design overrides (§2).
5. If a specific scripted/recorded demo run needs to guarantee no cast loss mid-take, apply
   demo-config danger tuning (fewer spawns, daytime-only staging) rather than touching the
   permadeath rules (§5).

### 6.3 Seam-tension audit — the "no protocol extension" claim

§3 above claims every memory/conversation shape this design needs already exists in
body-protocol-v0 and none is extended. Audited against `docs/design/body-protocol-v0.md`
directly (percept types, the `change_report` delivery restriction, and the remember
surface), mechanic by mechanic:

- **Witness sighting** (a friend attacked/dying) → ordinary `sighting` (§4.2 of the
  protocol): `thing` + `doing` + `origin: "saw"`. No new percept type; `doing` is already
  prose-only per AR-3. Fits.
- **Absent survivor learning of a death** → `change_report` with `change: "gone"` (§4.10).
  The protocol's delivery restriction (never to the actor or a witness) is exactly what this
  design leans on to *not* double-report a death to the people who already lived through it
  — confirmed compatible, not merely unblocked by it.
- **The new grave-thing a returning survivor perceives** → an ordinary `sighting`/
  `observation` of a `thing` with an as-yet-unseeded role (something like `"memorial"` or
  `"grave"`). `roles` is an explicitly **open** vocabulary (AR-3: "minds MUST tolerate
  unknown roles") — introducing a new role value is vendor content, not a protocol change.
  No extension.
- **Belongings bundle** → `roles: ["storage"]` is an already-seeded role (AR-3). No new
  vocabulary needed at all.
- **Dusk retelling** → `speech`/`told_fact`, `origin: "told"` (§4.5–4.6) — the ordinary
  secondhand path, unmodified.
- **Zombie-villager conversion ruled equivalent to death** → uses `body_continuous: false`
  at `session_open` continuity exactly as specified in §6.3 of the protocol ("this is a
  different body"). A clean, intended use of an existing field, not a repurposing.
- **Dead villager's own mind-process lifecycle** and **dusk-surfacing decay/consolidation**
  are both explicitly left to the mind side (SI-5, remember surface) — this design states a
  posture and defers the mechanism to TASK-0004, which is exactly where SI-5 puts that
  decision. No protocol involvement to audit.
- **`target_gone` as the sanctioned non-existence channel** (§5.6 of the protocol) is not in
  tension with the richer death-reporting channels above: `target_gone` is the fallback for
  a mind that still acts on stale knowledge of someone now dead; `sighting`/`change_report`
  are the primary, richer channels for learning the death happened as an event. The two
  are complementary, not competing.

**Result: audit found no tension.** Every mechanic in §§1–3 maps onto an existing percept
type, an existing open-vocabulary field (`roles`), or an existing continuity field, with no
new percept type, no new envelope field, and no narrowing of an existing one. This is a
genuine re-check, not a restatement of §3's own claim — flagged here for the protocol owner
to spot-check independently rather than accepted on this document's say-so.

### 6.4 Wiki re-verification

Checked this branch's design against the two wiki notes the brief's death posture touches,
even though neither currently lists `docs/design/death-mechanics.md` as a source:

- **[[v1-demo]]** — its prose states the demo's required texture and spell-breakers but
  only *references* death mechanics as an open question feeding the demo ("death mechanics
  (TASK-0005)"). This design resolves that open question without contradicting anything
  the note asserts — night danger's role in the demo ("the player's walls and torches
  protect their friends") is reinforced, not changed. **Verdict: extended, not
  contradicted.** No re-pin; `death-mechanics.md` is not added to its `sources` because
  the note's own claims remain accurate as written.
- **[[design-brief]]** — its prose compresses ratified decision #4 ("permadeath is real and
  should sting... mechanics deferred") and decision #6 (no direct control, spell-breakers).
  This design is precisely that deferred mechanics work, landing inside both ratified
  constraints (permadeath stays real; the micromanagement spell-breaker is explicitly
  answered in §4). **Verdict: extended, not contradicted.** No re-pin.

Neither note's `description:` frontmatter field was changed, so no `CAPSULES.md`
regeneration is needed from this check. If a future pass decides either note should cite
this design doc directly (e.g. because TASK-0006 wants `[[v1-demo]]` to name the death
posture explicitly), that is a normal wiki-update, not implied by this finding.
