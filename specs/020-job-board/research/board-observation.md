# T010 — dev-server board/build observation (card ACs #2, #5, #6 live halves)

**Scope**: Phase 4. Two `./gradlew runServer` boots against the persisted `mod/run/world`
(cast already seeded — Aldric/Petra/Yenna), no player connected, console commands via
`run/stdin.fifo` (held open with a background `sleep 999999 > run/stdin.fifo` writer, the
same idiom `death-observation.md` used). A throwaway Python "stub mind" (`mod/run/
stub_mind*.py`, not committed, `.gitignore`d under `mod/run/`) stands in for the mind daemon
at `kithcraft.sock`, speaking the real length-prefixed wire protocol — precedent:
`specs/012-vendor-conformance/research/verb-observation.md`. Two boots were needed because the
first session's claim exchange left a stale, half-dead Unix-socket connection the mod's
single-body attach loop never detected (recorded honestly below, not glossed over).

Commits under test: HEAD at observation time `70e093b` (Phase 3 code unchanged since
`eab23a9`; the two intervening commits are runbook bookkeeping only).

## The beat, step by step

### 1. Board setup + book — confirmed, persisted across boots

The lectern is placed at a fixed offset from world spawn (`BoardSetup.placeBoard`,
`Cast.MEMBERS.size()+2, 0, 2` → `(5, 119, 2)` for this world's origin), confirmed present at
every boot this session touched. The posting text `"Build a small stone wall, please."` was
planted by the prior (stopped) Phase 4 agent's session via `/data merge block 5 119 2
{Book:{...writable_book_content...}}` and **persisted through this session's two full server
restarts** — `BoardData`'s `SavedData` round-trip works exactly as `BoardTest`/T001 already
proved in unit tests, now also proven live across a real restart boundary, not just an
in-process round-trip.

### 2. The read — confirmed live, repeatedly, this session

`BoardVisit` rides `Activity.MEET` exactly as designed (T002). Both Aldric and Petra read the
board multiple times in this session, each producing the exact composed percept envelope
(`BoardVisit.java:82`'s `LOGGER.info("[board] {} reads the board: {}", ...)` line):

```
[23:03:02] [board] Aldric reads the board: {..., payload={percept_id=p-1, percept_type=text,
  urgency=notable, provenance={origin=read, source={kind=artifact, thing_id=th-16,
  descriptor=the job board}, ...}, place={place=pl-17, descriptor=the job board},
  content={text=Build a small stone wall, please. ...
[23:03:05] [board] Petra reads the board: {... world_time=10289, ...}
[23:03:21] [board] Petra reads the board: {... world_time=10609, ...}
[23:03:33] [board] Petra reads the board: {... world_time=10849, ...}
[23:03:40] [board] Petra reads the board: {... world_time=10989, ...}
```

Content/provenance composed through the pre-existing `Testimony`/`PerceptEmitter` exactly as
T002's structural zero-protocol-extension test already proved statically — this is that same
composition, now observed firing repeatedly from real dusk-arrival behavior across two
different villagers. `thing_id=th-16` here is this boot's own freshly minted board token (see
§4's finding below for why that number matters).

### 3. The claim — confirmed live (a genesis claim, plus this session's persistence reconfirmation)

**A live claim was already proven completing** in the prior (stopped) agent's session, same
code, same world, captured in that session's own log (retained on disk, `mod/run/
console.log`/`stub_mind.log` from before this closer session started): a stub-mind body
(`b-6`) sent `{"verb":"claim","target":{"type":"thing","thing_id":"th-0"}, "reason":"Building
shelters is exactly my trade, and this looks like honest work."}` and received back
`intent_ack{accepted:true}` followed by `act_result{outcome:"completed", detail:"claimed
it"}` — driving the real engine pipeline (`IntentHandler` → `BoardClaims` → `Board.tryClaim`)
end to end, not a unit test standing in.

**This session reconfirmed the claim's persistence and the first-accepted-wins rule live,
across a restart**: a fresh stub-mind body (issued this boot, `b-30`) sent the identical claim
shape against this boot's own board token (`th-24`, resolved correctly — the target-resolution
step matched) and got back `act_result{outcome:"failed", reason_code:"blocked",
detail:"someone else already claimed it"}`. This is `Board.tryClaim`'s persisted `claimedBy`
field surviving two full server restarts and correctly rejecting a second claimant — a
stronger live proof than a second isolated "it completed" run would have been, since it
exercises persistence (`BoardData`'s codec) and the first-wins rule (`Board.tryClaim`'s
`claimedBy.equals(claimantBody)` branch) that a single-boot demo can't reach.

**Honest nuance, found while re-driving**: `TokenRegistry.issue` (`TokenRegistry.java:65`)
mints a brand-new, never-reused token on every call, including at every server boot's board
setup (`KithcraftMod.onServerStarted`'s `boardThingToken = pendingTokens.issue(THING, "the job
board")` runs unconditionally, with no "reuse if referent already registered" check). So the
board's own thing-token *string* changes every restart (`th-0` → `th-8` → `th-16` → `th-24`
across this world's four boots so far) even though it always names the same board — a claim
attempt against a stale `th-N` from a previous boot resolves (the old token is never retired)
but fails target-matching in `BoardClaims.claim` (`!boardThingToken.equals(thingToken)` →
`UNKNOWN_POSTING`/`not_capable`), while a thing-token that was never issued at all fails one
level earlier, at `intent_ack{accepted:false, reason_code:"unknown_target"}`, before any
`act_result` is even composed. Neither failure mode is a code bug — just a fact worth a future
session knowing before it hardcodes a token value into a script the way this session's first
attempt did (see below).

### 4. Build placement — NOT observed, root cause identified (not a timing question)

**Two bounded checks, both negative**: `/execute if block 5 119 5 minecraft:air run say
BUILD_SITE_EMPTY` (the build site's origin block, `BuildSetup.siteOrigin`) returned
`BUILD_SITE_EMPTY` both times checked (once ~76s after the original genesis claim in the prior
session's log, once again in this session ~90s after this session's reconfirmed-blocked claim
attempt). No block was ever placed at the build site across either session.

**Root cause, found by reading the live wiring rather than waiting longer**: build placement
is structurally unreachable through the current single-body live attach path, independent of
`Activity.WORK` timing or the open 24000-tick schedule question
(`docs/wiki/villager-brain-api.md`). `LiveBuildExecution.tick`'s `findClaimant`
(`LiveBuildExecution.java:81-86`) resolves the board's `claimedBy` token against
`bodyTokenLookup`, which `KithcraftMod.onServerTickUnsafe` wires to `duskPairing::bodyTokenFor`
— **`DuskPairing`'s own per-cast seat-token map** (`DuskPairing.java:216-221`, populated only
for Aldric/Petra/Yenna's dusk-gathering seats). But the body that actually sends a live claim
intent is `BodySession.open`'s **own, separate, generic single-attach token**
(`BodySession.java:67`, `ground.issueBody("a villager", ...)`) — a different token namespace
entirely, never registered with `DuskPairing`. `board.claimedBy()` therefore always holds a
token `duskPairing.bodyTokenFor(uuid)` can never produce for any villager, for any UUID, in
any session — `findClaimant` returns `null` every tick, forever, once a claim has landed
through this live path. This is not a race with the schedule substrate; it's two disjoint
token spaces on either side of one `equals` check. Recorded here as an honest, specific
non-closure — not something this Phase 4 closer pass is scoped to fix (T010 is observation,
not a new build), but precise enough that a future task doesn't have to re-derive it from
scratch the way this session did.

Because build never started, `Activity.WORK`'s own timing (whether it was active at all
during either session's window) was never actually load-bearing for this particular
not-observed result — worth stating plainly so a future reader doesn't spend a bounded-check
budget on the wrong hypothesis.

### JOB_SITE — reconfirmed absent, consistent with the known TASK-0014 gap

`/data get entity @e[type=minecraft:villager,name=Aldric,limit=1]
Brain.memories."minecraft:job_site"` returned `Found no elements matching ...` in **both**
sessions this pass touched (matching the prior stopped session's three identical results) —
the same already-documented gap `death-observation.md` and `docs/wiki/villager-brain-api.md`
/`full-cycle-observation.md` record; not new, not this task's to close.

### 5. Interrupt/resume — NOT observed live (no build ever started to interrupt)

Since no build placement ever began (§4), there was nothing live to interrupt or resume.
**Covered structurally, not empirically, by `BuildEngineTest`**
(`scheduleTransitionInterruptsPlacementAndCursorPersistsUnchanged`,
`dangerPanicInterruptsPlacementAndCursorPersistsUnchanged`,
`resumesFromTheCursorWithoutRestartingOrReclaiming` — T009, all green) — the live half lands
at I2 (TASK-0022) once the §4 wiring gap is closed and a build can actually start to be
interrupted. This is the same honest "unit-proven, live half deferred" pattern
`death-observation.md` used for its own not-yet-reachable live proofs.

## What was NOT observed

- **Villager WORK-period build placement.** Root cause identified (§4) — a live-wiring token-
  namespace gap, not a timing miss. Not fixed here (out of T010's scope).
- **Interrupt/resume of a live, in-progress build.** No build ever started (§4) to interrupt.
  Unit-proven only (`BuildEngineTest`); live half deferred to TASK-0022 (I2).
- **A second, independently-completing live claim in a single boot.** The board models
  exactly one posting at a time by design (Phase 2's own finding) — a second claim attempt in
  the same world can only ever observe the reject path once the first has landed, which is
  exactly what this session's §3 reconfirmation shows, and is itself the intended behavior,
  not a gap.

## Gates

`./gradlew compileJava` green before this run (worktree was clean, no uncommitted changes —
Phase 3's `eab23a9` is what actually ran). Full `./gradlew build test` is T011, run separately
after this observation.
