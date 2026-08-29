# T007 — live bring-up observation (card ACs #1, #2)

**Scope**: Phase 3. `docs/design/demo-runbook.md` followed verbatim, both start orders
exercised across two daemon lifetimes, rehearsal mode throughout (**zero API calls** — no
`ANTHROPIC_API_KEY` ever exported this session). One doc fix folded back in while following
it (§0 below); one honest, unresolved open question recorded rather than guessed at (§4).

## 0. Runbook fix made while following it

Order A (daemon first) on a *fresh* checkout fails at listen time: `mod/run/` does not exist
yet (gitignored, absent on checkout) and the daemon's `listen()` does not create the socket's
parent directory. §2 (`mkdir -p mind/run mod/run`) already covers this — it was added to the
runbook from this exact dry run, not discovered as a gap after the fact.

The bigger fix: a fresh `./gradlew runServer` boot **paused after 60 seconds** with no player
connected (`server.properties`' vanilla `pause-when-empty-seconds=60` default) — the whole
tick loop, and with it every attached mind session, stops advancing. This silently starves
the headline check (no ticks ⇒ no heartbeats ⇒ nothing to observe). Folded into the runbook
(§0's "Unattended/no-player runs only" note) rather than left as a footgun for the next
operator running rehearsal solo — the same fix `specs/014-augmented-villager` and
`specs/019-death-remains`'s own unattended observations already needed.

## 1. Build + first-time setup — confirmed exactly as documented

```
cd mind && go build -o minddaemon ./cmd/minddaemon        # OK
cd mod && ./gradlew build                                  # OK, no test failures
mkdir -p mind/run mod/run && echo eula=true > mod/run/eula.txt
```

## 2. Rehearsal mode — both sub-cases observed, zero API calls throughout

**No personas, `-genesis=false`** (fresh `mind/run`, nothing pre-seeded):

```
$ mind/minddaemon -socket "$SOCK" -rundir "$(pwd)/mind/run" -genesis=false
minddaemon: persona genesis needed for [Aldric Petra Yenna] but ANTHROPIC_API_KEY is not
set — export it and rerun, or pre-seed .../mind/run/persona from a prior live run
```

Fails loudly, names exactly the missing ids, binds nothing partial — spec.md's Edge Cases
row proven live, exit code 1, zero network calls (no key was ever exported to fail against).

**Personas present, `-genesis=false`**: TASK-0013's live-run personas (`mind/run/persona/`)
are gitignored and were not present in this fresh worktree (checked — absent both here and
in the root checkout's `mind/run`), so this run seeded **stub** JSON files by hand for the
three demo cast ids (`cast_id`/`name`/`profession`/`biome_variant` matching
`persona.DemoCast()`, written 0444) rather than running a real genesis — the fallback this
dispatch's brief names explicitly. With those present:

```
$ mind/minddaemon -socket "$SOCK" -rundir "$(pwd)/mind/run" -genesis=false
minddaemon: listening on .../mod/run/kithcraft.sock
```

`LoadOrGenesisCast` re-bound all three via `Load`, zero model calls — this is what §4's
"preferred rehearsal path" claims, confirmed. **This run does not exercise real genesis
output** (a live E1 call, real persona text) — that half of card AC #1's scenario 2 is
carried by `specs/013-persona-genesis/live-run.md`'s already-recorded live run, not
re-proven here (this dispatch's zero-API-call constraint forbids re-running it).

## 3. Sequence yields a running server + daemon with sessions attached — confirmed, both orders

Order A (daemon first) and order B (server first) were both exercised across the two daemon
lifetimes below; in both, `WireClient.dialWithBackoff`'s retry closed whichever gap existed.
Mod log, second boot (reused persisted world, cast already seeded — idempotent, no re-spawn
lines):

```
[live] token registry live entries at boot: {...}
[live] attach scan at tick 1200: 3 villager(s), 4 total entities: [...]
[live] session_open sent: session=s-live-... body=b-14 place=pl-15
[live] attached to villager a0f29acd-c1e6-4aab-8177-c86990ac0b75
```

Daemon log: `minddaemon: listening on .../mod/run/kithcraft.sock`. Both artifacts up,
independently started, session attached — card AC #1's structural half, confirmed live.

## 4. Headline check — mind-restart independence (card AC #2, T-4)

**Driven via a scripted double** (`mind/seamtest.DialUnix`, the same test double the
codebase's own `TestEndToEnd_PerceptsInIntentsOut` uses — a throwaway `go run` program,
deleted before this commit, never part of the shipped tree) dialed at the **same live daemon
socket** the mod server was simultaneously attached to, for a distinct body token
(`obs-body-1`) — see §5 for why the live mod's own session could not be used for this half.

1. **Admit a memory pre-kill.** `session_open` for `obs-body-1`, then one `notable`-urgency
   percept at `world_time=100` (`RuleUrgency`, §6.3). Confirmed admitted:
   `mind/run/villagers/obs-body-1.jsonl` — one line, `world_time: 100`.
2. **Kill the daemon mid-session.** `kill <pid>` (SIGTERM, graceful — the runbook's §6 path).
   `mind/run/session-report.log` gained a new entry: `obs-body-1 day 0: 1 admitted` —
   card AC #5's report, unconditional, captured before the process actually exited. The
   socket file (`mod/run/kithcraft.sock`) was removed on exit, matching `listen()`'s own
   documented behavior.
3. **Restart the daemon** — same binary, same `-rundir`, same stub personas (re-bind, zero
   calls, per §2).
4. **Reconnect the same body token, post-restart.** A fresh scripted double, same
   `obs-body-1`, one more `notable` percept at `world_time=5000`. `HandleConnection`'s own
   doc claims a reconnecting body is matched by its `body` token alone, with no memory of the
   prior connection required — confirmed: the second daemon process accepted it and appended.
5. **Memories survive; the gap is a gap, never backfilled.** `mind/run/villagers/
   obs-body-1.jsonl` now reads, in full:
   ```json
   {"world_time":100, ..., "percept_id":"obs-p1", ...}
   {"world_time":5000, ..., "percept_id":"obs-p2", ...}
   ```
   Two records, nothing else — the daemon never invented anything for the 4900-tick window
   it was down (M1's continuity rule, structural: there is no code path in `memory.Log` or
   `Runtime.HandlePercept` that could write a record for a percept nobody sent). The gap
   itself is visible in the data (the jump from 100 to 5000), which is what "reported, never
   backfilled" cashes out to here — there is no separate gap-report field on the wire this
   build populates (`HandleConnection`'s comment: `previous_session` is read by nobody yet;
   see §5).
6. The second daemon's own session-report (a fresh in-memory `Instrument`, not reloaded from
   disk) correctly read `obs-body-1 day 0: 1 admitted` for *its own* lifetime — the
   cumulative count lives in the log file, not the report, which is honest per-process
   accounting rather than a bug.

**Card AC #2 / US2 / T-4: proven** — the mechanism decision-0003 promised (daemon restart is
recoverable at the session boundary, no data invented for the downtime) holds at the data
level, driven directly against the live daemon process exactly as a reconnecting vendor
would be received.

## 5. What was NOT observed, and why — the live mod's own reconnect

Two things did not happen during ~2m40s of continuous, unpaused ticking with the mod's
`BodySession` attached (spanning the daemon kill and restart in §4):

- **No `self_state` heartbeat was ever admitted**, and no `mind/run/villagers/<its body
  token>.jsonl` file was ever created for it — `bodyOrOpen` opens a body's store on the
  *first* `HandlePercept` call regardless of admission, so this means the daemon never even
  saw one. `SelfState.content()` carries only `condition`/`level`/`trend` — no `thing`,
  `place`, `sound_kind`, or `present` key — so `memory.Gate`'s `subjectsOf` finds nothing to
  call a first sighting, and `background` urgency + `felt` provenance fail every other §6.3
  rule too. **Structurally, this build's only live percept type can never be admitted** —
  that much is provable by reading `admission.go` and is not in doubt. Whether the heartbeat
  is even reaching the daemon at all (as opposed to being sent-but-dropped) was not
  distinguished within this observation's budget.
- **The mod never logged noticing the daemon's outage or reattaching.** No `"reader ended"`,
  `"session tick failed"`, `"dial failed, will retry"`, or a second `"attached to villager"`
  line appeared — checked both mid-run and again after a full graceful `stop` (ruling out
  log buffering as the explanation; a clean JVM shutdown flushes everything, and nothing new
  about the mind session appeared).

**This is why §4 used a scripted double instead of the live mod's own session**: the live V1
wiring (`BodySession`, T011's dev-server proof) did not, this run, exercise session-loss
detection or reconnection on its own — root cause not isolated here (out of this phase's
scope: T007 is a documentation-and-observation task, zero API calls, no code changes). Worth
a dedicated follow-up session-lifecycle observation.

Separately, reading `Continuity.java` (not something this run needed to exercise to notice):
`BodySession.open` always sends `Continuity.firstSession()` and always mints a **fresh** body
token via `ground.issueBody(...)` on every attach — its own class doc says as much ("Until
[the token registry] exists, every body opens as `firstSession()`"). So even once the mod
*does* reattach, today's live wiring would not yet carry a body's identity across the
reconnect the way §4's scripted double (same token, by hand) did — that registry is named as
Phase 3/T008's prerequisite in `Continuity.java`'s own doc, not this task's to build.

## Gates

`go build -o minddaemon ./cmd/minddaemon` and `./gradlew build` both green before this run
(§1). No code changed by this observation — the throwaway scripted double
(`mind/cmd/obsdouble`) was deleted before this commit; `git status` clean of it.
