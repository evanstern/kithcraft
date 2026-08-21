# Death surface evidence — spec 005 Phase 1

All facts below were checked live on **2026-08-21** (accessed date for every citation unless
noted otherwise). This builds on — does not duplicate — [[engine-behavior]]
(`specs/003-entity-implementation/research/engine-behavior.md`), which already verified §1
(hostile-mob targeting of `VillagerEntity`), the raid flee-to-house/bell mechanic, despawn
rules (villagers are passive, don't distance/time-despawn), and zombie-villager conversion
odds and Mixin-suppression points. Those findings are cited by reference below, not
re-derived. Per decision-0002, the augmented vanilla villager inherits everything documented
here for free (it is `VillagerEntity`); a custom entity would inherit none of it (see 003
§1–§4 for the per-behavior Mixin-surface breakdown).

## 1. Villager damage sources and immunities

- **No hunger/starvation mechanic for villagers.** Villagers have a private `foodLevel`
  field (persisted as NBT `FoodLevel`) and a `hungry()`/`wantsMoreFood()` check, but these
  drive only breeding willingness (12+ nutrition points from food in the villager's hidden
  inventory makes it "willing" to breed) — there is no analogue of the player's
  `foodTickTimer`/starvation-damage loop on `Villager`/`AbstractVillager`. The "Starvation"
  damage type exists in the engine's generic damage-type table (armor/enchant interactions
  are defined for it), but the trigger mechanic itself — hunger bar hitting 0 and then
  periodically dealing damage — is implemented on the player only. **Villagers do not
  starve to death in vanilla**, confirmed by the absence of any such path in the Villager
  mechanics/food sources and by the Forge-mapped `Villager` class method list (`hungry()`,
  `eatUntilFull()`, `wantsMoreFood()`, `digestFood()` — all breeding-food bookkeeping, no
  health-damage call).
  https://minecraft.wiki/w/Villager (§Breeding, NBT `FoodLevel`) — accessed 2026-08-21.
  https://minecraft.wiki/w/Food — accessed 2026-08-21.
  https://nekoyue.github.io/ForgeJavaDocs-NG/javadoc/1.18.2/net/minecraft/world/entity/npc/Villager.html
  — accessed 2026-08-21 (Forge 1.18.2 mapping; method existence/role is stable vanilla
  game-logic, not re-verified against the project's Yarn build — flag per
  [[villager-brain-api]]'s own note to re-check mappings at implementation time).
- **Fall damage: standard, no special immunity found.** Most mobs (villagers included; no
  exemption found in any source consulted) take ~1HP per block fallen beyond their
  `safe_fall_distance` attribute (default 3 blocks). Water of any depth negates fall damage
  entirely.
  https://minecraft.wiki/w/Damage (§Fall damage) — accessed 2026-08-21.
- **Fire/lava: standard, no immunity.** Contact with lava deals 4HP/tick and sets the
  entity ablaze (~15s burn, 14 fire-damage ticks after leaving); no villager-specific
  exemption found (unlike undead mobs, which burn in sunlight but aren't fire-immune
  either).
  https://minecraft.wiki/w/Lava — accessed 2026-08-21.
- **Drowning: villagers are not exempt.** The drowning-immune list is aquatic mobs, undead
  mobs, and iron golems — villagers are none of these, so they have the standard ~15s
  breath meter and then take 2HP/s drowning damage like the player.
  https://minecraft.wiki/w/Water — accessed 2026-08-21.
  https://minecraft.wiki/w/Damage (§Drowning) — accessed 2026-08-21.
- **Player friendly fire: unrestricted, but costs reputation.** A player can attack or kill
  a villager directly with no engine-side protection. Doing so generates gossip that every
  other villager with line-of-sight within a 16-block box of the victim picks up:
  `minor_negative` (attack: weight −1, +25 on hit, decays −20/20min, max 200) or
  `major_negative` (kill: weight −5, +25, decays −10/20min, share penalty −10, max 100).
  **Indirect kills (fire, lava, suffocation from falling blocks) generate no negative
  gossip at all**, because "villagers can determine their source" is required for
  reputation to register — an explicit vanilla loophole.
  https://minecraft.wiki/w/Villager (§Reputation) — accessed 2026-08-21.
  https://minecraft.wiki/w/Trading (§Reputation) — accessed 2026-08-21.

## 2. Night-danger mechanics vs. villagers

Zombie/husk/drowned/zombie-villager targeting-and-pursuit, the `VillagerHostilesSensor`
distance table, raid flee-to-house-with-door-and-bed behavior, bell-ringing, and
zombie-villager conversion odds (0%/50%/100% Easy/Normal/Hard) are already verified in
[[engine-behavior]] §1 and §3 — not re-derived here. New findings this pass:

- **Panic/flee is a `PanicTask` Core-activity brain task, not a `Goal`.** It triggers when
  the brain holds a `HURT_BY` or `NEAREST_HOSTILE` memory (the latter populated by
  `VillagerHostilesSensor`, already in 003 §1); firing it forgets `PATH`/`WALK_TARGET`/
  `LOOK_TARGET`/`BREED_TARGET`/`INTERACTION_TARGET` and switches the villager to its Panic
  activity — visibly, sweat particles and shaking, frantic flight, sometimes into houses. In
  Java Edition the trigger list is zombie, zombie villager, husk, drowned, zoglin, illager,
  vex, wither, ravager.
  https://minecraft.wiki/w/Villager (§Behavior) — accessed 2026-08-21.
  https://blog.yuuta.moe/2021/09/06/minecraft-1.17.1-panic-iron-golem-spawning-mechanism/
  — accessed 2026-08-21 (Yarn-mapped 1.17.1 brain-task walkthrough; behavior described is
  consistent with the wiki's current-version prose above, cross-checked, not independently
  re-verified against a current Yarn build — flag as secondary source).
  Panic also drives iron-golem summoning: 3 panicking villagers within a 10×10×10 box, all
  having slept within the last 24000 ticks (one day) and none holding a fresh
  `GOLEM_DETECTED_RECENTLY` memory, spawn a golem (same source).
- **Doors: villagers open/close wooden doors when pathfinding** (added 1.2.1, 12w05a) —
  free per-augmented-villager behavior, nothing to wire. **Zombies attack closed wooden
  doors** (banging/shaking animation + sound) and, only on Hard/Hardcore, can break the top
  half down in ~10 seconds (bottom half cannot be broken; a zombie facing the bottom half
  can't break through). Sources disagree on the per-zombie chance of spawning with the
  door-breaking ability: the current `Wooden_Door` and `Zombie` wiki pages say 5% (Java) /
  10% (Bedrock); the `Village_mechanics` page's older prose describes it as difficulty-gated
  without a percentage. **Vindicators** (raid-spawned, Normal/Hard) can also open or break
  wooden doors, but only while pursuing a targeted player/villager/wandering trader.
  Zombie-proofing by recessing the door frame (so zombies can't get the required jump/facing)
  is a documented player counter-strategy.
  https://minecraft.wiki/w/Wooden_Door — accessed 2026-08-21.
  https://minecraft.wiki/w/Zombie (§Breaking doors) — accessed 2026-08-21.
  https://minecraft.wiki/w/Village_mechanics (§Zombie sieges) — accessed 2026-08-21.
- **Zombie sieges are a separate, Java-only event from raids/normal targeting** — not yet
  covered in 003. At midnight (tick 18000) there's a 10% nightly chance a siege is queued;
  it fires near a player standing inside a "logical village" (a village needs a
  `#without_zombie_sieges`-untagged biome and, per older-version prose not yet reconfirmed
  against 1.14+'s rewritten village detection, historically required ≥10 doors and ≥20
  villagers — flag: the population/door-count gate is described inconsistently across
  wiki revisions and needs re-check against the target MC version at implementation time).
  ~20 zombies attempt to spawn over 60 ticks at a point 32 blocks from the triggering
  player, ignoring the normal 24-block player-proximity spawn restriction; siege zombies
  behave exactly like normally-spawned zombies afterward (same targeting, panic-inducing,
  door-breaking rules above) and can still convert villagers to zombie villagers on kill.
  https://minecraft.wiki/w/Zombie_siege — accessed 2026-08-21.
  https://minecraft.wiki/w/Village_mechanics (§Zombie sieges) — accessed 2026-08-21.
- **decision-0002 implication:** because the augmented villager is `VillagerEntity`, all of
  the above (targeting, panic, door interaction, siege eligibility, conversion) is inherited
  automatically; no engine work is required to get night danger "for free" — the design
  work (Phase 2) is about what to admit/suppress/tune, not what to build.

## 3. Vanilla death aftermath

- **Villagers drop no items or XP on death, with two narrow exceptions.** A farmer
  villager holding bone meal has an 8.5% (+1%/Looting level) chance to drop it when killed
  by a player or tamed wolf; an adult villager wearing armor equipped via dispenser can
  drop that armor. Otherwise, on death **any item in the villager's hidden inventory is
  lost outright, not dropped as an item entity — even with `keepInventory` true.** A
  trade-offer item held in a villager's hand at time of death also does not drop (though it
  can be sheared off beforehand for carpets/leather horse armor/saddles). The
  `entity.villager.death` sound fires both on death and on zombie-villager conversion — it
  does not distinguish the two outcomes audibly.
  **Do not confuse this with trade-XP:** a successful trade drops 3–6 XP (8–11 while the
  villager is breeding-willing) as an unrelated, live mechanic — irrelevant to death, noted
  here only to avoid conflating the two XP-adjacent facts.
  https://minecraft.wiki/w/Villager (§Drops, §Sounds) — accessed 2026-08-21.
- **Item despawn timer (for the rare cases something does drop): 6000 ticks (5 minutes)**
  in a loaded, entity-ticking chunk; paused while the chunk is unloaded; stacking items
  inherit the longer remaining timer of the merged pair.
  https://minecraft.wiki/w/Item_(entity) — accessed 2026-08-21.
- **Bed/job-site POI release on death — directly confirmed in the vanilla `Villager`
  class's method surface.** The Forge-mapped 1.18.2 javadoc lists `die(DamageSource)` as
  an override, alongside private `releaseAllPois()` and `tellWitnessesThatIWasMurdered(Entity)`,
  and a public `releasePoi(MemoryModuleType<GlobalPos>)`. Read together with the reputation
  finding above (§1: a kill broadcasts `major_negative` gossip to every witnessing villager
  within a 16-block line-of-sight box), `tellWitnessesThatIWasMurdered` is almost certainly
  the call site that drives that gossip broadcast, and `die()` calling `releaseAllPois()`
  confirms the dying villager's bed (`minecraft:home`) and job-site claims are released back
  to the world (available for another villager to claim) as part of the death path, not left
  dangling. **Flagged as thin evidence in one specific sense:** this is a Forge 1.18.2
  mapping, not the project's Fabric/Yarn 1.21.3 build ([[villager-brain-api]]'s pin); method
  *names and behavior* at this level of vanilla game logic are stable across editions/versions
  in practice, but exact signatures should be re-confirmed against the Yarn build used at
  implementation time, consistent with [[villager-brain-api]]'s own "re-verify mappings
  against the target MC version" note.
  https://nekoyue.github.io/ForgeJavaDocs-NG/javadoc/1.18.2/net/minecraft/world/entity/npc/Villager.html
  — accessed 2026-08-21.
- **Death is logged.** In Java Edition, "the death of a player, villager, or renamed mob is
  recorded in the game's logs" — a minor, free hook if the design wants a server-log trail
  independent of the mind daemon's own memory writes.
  https://minecraft.wiki/w/Death — accessed 2026-08-21.
- **Gossip/reputation on death:** see §1 above (`major_negative`, weight −5, broadcast to
  witnesses within 16 blocks, indirect-kill loophole) — the single reputation-relevant
  effect of a villager's death found in vanilla; no other reputation/gossip side effect of
  death itself (as opposed to being attacked while alive) was found.

## Areas where evidence is thin or version-dependent (do not round off)

1. The zombie door-break spawn chance is cited as 5% (Java)/10% (Bedrock) by current
   `Wooden_Door`/`Zombie` pages but described only qualitatively (Hard/Hardcore-gated, no
   percentage) by the `Village_mechanics` page — not reconciled here, flag for Phase 2 if the
   design needs an exact number to tune against.
2. Zombie-siege village-eligibility (door/villager-count thresholds) is stated inconsistently
   across wiki revisions pre- vs. post-1.14 village-detection rewrite; needs re-check against
   the project's actual target MC version before any siege-suppression Mixin work is scoped.
3. The `die()`/`releaseAllPois()`/`tellWitnessesThatIWasMurdered()` method surface is sourced
   from a Forge 1.18.2 mapping, not independently cross-checked against the project's
   Fabric/Yarn 1.21.3 pin — same caveat [[villager-brain-api]] already carries for its own
   citations.
4. Whether `VillagerHostilesSensor`/panic-task machinery is generic `Brain<E>` substrate or
   `VillagerEntity`-specific remains open per [[engine-behavior]]'s own flag #1 — unchanged by
   this pass, still relevant to what a hypothetical non-augmented entity would have to
   reimplement (moot under decision-0002, noted for completeness).
