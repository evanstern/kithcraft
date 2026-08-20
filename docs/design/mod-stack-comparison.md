# Mod stack comparison — Fabric vs Paper/Citizens2 vs hybrid

**Spec:** 001-mod-stack-decision, Phase 2 · **Evidence base:** `specs/001-mod-stack-decision/research/prior-art.md`
(all facts re-verified 2026-08-20; cited here by reference to that file's URLs and accessed
dates unless a claim needed new sourcing, which is called out where it occurs).
**Ratified constraints checked against:** `docs/design/kithcraft-brief.md` (2026-08-19) —
villager-shaped NPCs (not bot clients), small cast (3–6) at real-time only, and a
world-agnostic body protocol with the mod as the first swappable body vendor.

## Option A — Fabric server-side mod

**What it is:** a Fabric mod running server-side, driving vanilla `Villager` entities (or a
custom entity extending the same brain substrate) directly through the engine's AI API.

**Constraint fit:**
- *Villager-shaped, not bot clients* — direct fit. Fabric mods run inside the server process
  and manipulate real server entities; there is no fake player client in the loop at all,
  which is the strongest form of this constraint among the three options.
- *Body-protocol seam* — clean. A Fabric mod is a thin server-side surface (tick loop +
  Mixin/accessor hooks); perceive/act/remember can be implemented as calls out to an external
  mind daemon with no requirement that mind logic live inside the mod. Nothing about Fabric's
  architecture forces game logic and reasoning logic into the same process.
- *Small cast, real-time only* — no friction; a 3–6 entity cast is trivial load for either the
  engine or an external LLM daemon polling at human conversational cadence.
- *Villager fiction (beds, workstations, schedules)* — direct fit, because the substrate is
  the actual vanilla villager brain. Per prior-art.md §6, reading and using existing
  `Schedule`, `Activity`, `MemoryModuleType`, and POI-sensor machinery on the vanilla
  `Brain<E>` class is plain API access, no Mixin required. Adding genuinely *new* activities,
  memory types, or POI types requires Mixin/accessor injection — demonstrated end-to-end
  against a real mod in the Fabric wiki tutorial cited there — which is standard Fabric
  practice, not a blocker.

**Dependency health (from prior-art.md):**
- Fabric Loader 0.19.3 and Fabric API 0.158.0+26.2 are both actively maintained, Apache-2.0
  licensed, current against the 26.2 Minecraft line, with parallel builds for 26.1.x/1.21.x
  (prior-art.md §1).
- The villager brain API surface itself is a vanilla engine API (exposed to Fabric mods via
  Yarn mappings), not a third-party dependency with its own maintenance risk (prior-art.md §6).

**Risks:**
- Everything beyond "use what vanilla already defines" (new activities, new memory types, new
  POI types) requires Mixin/accessor code — real engineering surface area, version-sensitive,
  and a maintenance burden the project owns directly (no upstream plugin absorbs it).
- No off-the-shelf LLM-NPC framework to build on: CraftAgent (the nearest Fabric prior art) is
  a documented dead end — last push 2026-01-06 (~7.5 months stale), single contributor, pinned
  to MC 1.20.1/1.21.8 (not the current 26.x line) (prior-art.md §4). Its more active sibling,
  SecondBrain (sailex428/SecondBrain, LGPL-3.0, last push 2026-03-31, 46 releases), is worth
  reading as design reference for the same reason — but neither is a dependency this option
  proposes to build on; Option A means writing the LLM-NPC integration layer from scratch
  against the raw brain API, with CraftAgent/SecondBrain as prior-art reading, not code reuse.
- Custom-entity vs augmented-vanilla-villager is explicitly out of scope here (TASK-0003), but
  Option A does not foreclose either sub-choice — both remain available on a Fabric base.

## Option B — Paper plugin + Citizens2

**What it is:** a Paper plugin driving Citizens2 NPCs (or Paper's own entity API directly, per
AI_NPC's precedent) as the villager substrate.

**Constraint fit:**
- *Villager-shaped, not bot clients* — fit, with a caveat. Citizens2 NPCs are real server
  entities (not fake player clients), so the family match holds. But Citizens2's NPC substrate
  is not the vanilla villager brain — it is its own goals/behavior-tree + navigation system
  layered onto whatever entity type it's told to impersonate (brief, "prior art" section). Riding
  the actual village fiction (beds, workstations, vanilla `Schedule`/POI machinery) means either
  bypassing Citizens2's own behavior system to reach into the vanilla villager brain underneath,
  or reimplementing schedule/POI logic in the plugin — neither is what Citizens2 is built for.
- *Body-protocol seam* — fit, same as Fabric: a Paper plugin is also a thin server-side surface,
  and nothing requires mind logic to live in the plugin process.
- *Small cast, real-time only* — no friction, same reasoning as Option A.
- *AI_NPC as closer prior art:* AI_NPC (Paper plugin, prior-art.md §5) already demonstrates
  walk-up conversational NPCs with function-calling actions and off-main-thread LLM calls on
  this exact stack family — but it is "first public release" (2.0.0, 2026-07-23, one month old
  at check time, 34 downloads), too new to assess a maintenance track record, and **its license
  could not be verified** (no license field on its Hangar listing, no discoverable source repo)
  (prior-art.md §5). It cannot be relied on as a dependency until its license is confirmed
  directly from the author or a distributed jar.

**Dependency health (from prior-art.md):**
- Paper 26.2 build #112 is actively maintained, current against 26.2, GPLv3-licensed (some
  contributor code dual-MIT, but the distributable artifact is governed by GPLv3)
  (prior-art.md §2).
- Citizens2 2.0.43 is actively maintained but lagging: last free-repo commit 2026-02-16 (about
  six months before the 2026-08-20 check), no formal GitHub Releases (CI + Spigot-premium
  builds are the distribution channel), and Spigot-page user reviews as recently as the check
  report breakage on 1.21.10/1.21.11 requiring reloads, with the maintainer confirming
  Paper-build compatibility issues are being worked (prior-art.md §3).
- **License flag (load-bearing):** Citizens2 is OSL-3.0 — a copyleft license with network-use
  provisions, *not* on the FSF/OSI list of licenses considered GPL-compatible for combined
  works. Any code statically/tightly integrated with Citizens2 needs its own license review
  before combining with a differently-licensed codebase (prior-art.md §3). This is a direct
  risk to a body-protocol architecture that wants a clean, freely-relicensable mod-side seam:
  tight integration with Citizens2 would pull the mind-daemon-facing code under OSL-3.0's
  copyleft terms unless the integration stays at arm's length (e.g. Citizens2 as an external
  plugin process communicated with over its command/API surface, not compiled directly against).

**Risks:**
- The OSL-3.0/GPL-incompatibility flag above is the single biggest risk specific to this
  option — it constrains how tightly the body-vendor code can couple to Citizens2 without
  inheriting copyleft obligations.
- Citizens2's lag behind current MC builds (six-month-old last commit, active user-reported
  breakage) is an operational risk for a fast-moving MC version line.
- Citizens2's own NPC/behavior substrate is not the vanilla villager brain, so "ride the
  village fiction" (beds, workstations, vanilla schedules) is friction, not a given — Option B
  either fights Citizens2's abstraction to reach vanilla brain machinery, or forgoes it.

## Option C — Hybrid (Fabric core + Paper/Citizens2-style NPC layer, or vice versa)

**What it would buy:** in principle, the vanilla villager brain substrate's schedule/POI/memory
machinery (Fabric-side strength) combined with Citizens2's more mature NPC-authoring
conveniences (goal/behavior-tree API, persistence traits) if those were layered on top. It could
also mean running Fabric for the villager entities and a separate Paper-based service for
player-facing conveniences unrelated to the NPCs themselves.

**What it costs:**
- Fabric and Paper/Spigot are different server implementations with different plugin/mod APIs;
  there is no supported "run both on one server" story in the prior art gathered here — a
  hybrid in the literal sense (one server process, both stacks) is not what either ecosystem is
  built for. The realistic reading of "hybrid" is either (a) picking Fabric as the base and
  hand-rolling Citizens2-like conveniences on the vanilla brain API, which is just Option A with
  extra self-imposed scope, or (b) running two separate processes/servers bridged somehow, which
  multiplies operational and body-protocol-seam complexity for a v1 "one real evening" demo with
  a 3–6-NPC cast that neither option's dependency set actually requires.
- It inherits the worse of both dependency-health pictures: if Citizens2 is anywhere in the
  stack, the OSL-3.0 copyleft flag (Option B) still applies; if Fabric mixins into vanilla brain
  internals are anywhere in the stack, that engineering burden (Option A) still applies. A
  hybrid does not average the risks away — it accumulates them.
- No prior art gathered in Phase 1 demonstrates this combination working end-to-end (CraftAgent
  and SecondBrain are pure Fabric; AI_NPC is pure Paper; Citizens2 is pure Paper/Spigot). Betting
  the v1 demo on an unprecedented integration is a schedule and reliability risk against the
  "one real evening" milestone that neither pure option carries.

**Constraint fit:** inherits whichever half of the hybrid actually carries the villager-fiction
and body-protocol-seam requirements — there is no fit unique to hybridization itself; it is
additive complexity layered on one (or both) of Options A/B's existing fit profile.

## Adjacent prior art (context, not a fourth option)

sailex428/SecondBrain — a Fabric mod in the same design space as CraftAgent (LLM-driven NPCs),
materially more active (last push 2026-03-31, 46 releases, latest v3.1.6 2026-03-25, LGPL-3.0)
(prior-art.md §4). It is CraftAgent's more maintained sibling and worth reading for design
patterns if Option A is chosen; it is not evaluated here as a stack option because it is an
opinionated LLM-NPC framework, not infrastructure the mod stack decision is choosing between —
the same category CraftAgent occupies, one rung more alive.

## Recommendation

**PROPOSED — pending operator ratification.** Decision record: `backlog/decisions/decision-0001
- Mod-stack-Fabric-server-side-mod-proposed.md`.

**Recommended: Option A — Fabric server-side mod**, driving vanilla `Villager` entities (or a
custom entity on the same brain substrate, per TASK-0003) through the engine's brain API.

**Rationale, mapped to R3:**

- **Body-protocol seam.** Both Option A and Option B keep this seam clean — a Fabric mod and a
  Paper plugin are equally thin server-side surfaces, and neither forces mind logic into the
  mod/plugin process (see each option's "Constraint fit" above). The seam alone doesn't
  discriminate between them; dependency health does. Option B's only viable NPC substrate,
  Citizens2, is OSL-3.0 — a copyleft license not on the FSF/OSI GPL-compatible list — so a
  body-vendor implementation tightly coupled to it risks pulling that code under copyleft terms
  unless integration stays at arm's length. Option A has no such license entanglement: Fabric
  Loader and Fabric API are both Apache-2.0, and the villager brain surface is vanilla engine
  API, not a third-party dependency at all (prior-art.md §§1, 6). A clean, freely-relicensable
  body-vendor seam favors Option A.
- **Villager-shaped, not bot clients.** Both options satisfy this at the family level (real
  server entities, no fake-player client). Option A is the strongest fit of the three: it drives
  the actual vanilla villager entity/brain directly, with no intermediate NPC-behavior
  abstraction in between. Option B's Citizens2 layer is its own goals/behavior-tree system
  impersonating an entity type — a real server entity, but not the vanilla villager brain.
- **Village fiction (beds, workstations, schedules) — the deciding factor.** This is where the
  options diverge most concretely. Option A's fit is direct: reading/using the vanilla
  `Schedule`/`Activity`/`MemoryModuleType`/POI machinery is plain API access on the real
  villager brain (prior-art.md §6). Option B's fit requires either bypassing Citizens2's own
  behavior system to reach into the vanilla brain underneath, or reimplementing schedule/POI
  logic in the plugin — friction the comparison above documents as inherent to Citizens2's
  design, not incidental.
- **Option C (hybrid) rejected.** No prior art demonstrates a Fabric+Citizens2 (or similar)
  combination working end-to-end; a hybrid accumulates both options' risks (the OSL-3.0 flag if
  Citizens2 is anywhere in the stack, the Mixin engineering burden if Fabric-brain internals are
  anywhere in the stack) rather than averaging them, for a v1 cast of 3–6 NPCs that doesn't
  require either option's dependency set stretched this way.

**Accepted risks (carried into TASK-0002/0003/0006, not resolved by this decision):** no
off-the-shelf LLM-NPC framework to build on (CraftAgent is a documented dead end; SecondBrain is
design reference only); extending the brain substrate beyond what vanilla defines requires
Mixin/accessor code the project owns directly. Both are detailed under Option A's Risks above.

**What this narrows for TASK-0003 (entity implementation).** TASK-0003 chooses between a custom
entity and an augmented vanilla villager — that choice remains fully open; this decision does
not pick a side. What it does foreclose is the *substrate* that choice is made within: both
TASK-0003 sub-options now sit on the vanilla Fabric brain API (Schedule/Activity/MemoryModuleType/
POI, prior-art.md §6), not on a Citizens2-style goals/behavior-tree system. A "custom entity"
under this decision means a Fabric entity class still wired into the vanilla brain substrate
(extending or composing with `Brain<E>`), not a Citizens2-authored NPC with its own behavior
layer — that path is closed off by choosing Option A over Option B. TASK-0003 therefore inherits
Mixin/accessor access as the mechanism for anything beyond vanilla brain capabilities, for
either sub-option it picks.
