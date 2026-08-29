# Demo runbook — one documented sequence, two artifacts

**Spec:** `specs/021-demo-config` (I1) · **Board task:** TASK-0021 · **Consumes:**
decision-0001 (two artifacts), decision-0003 §T-4 (mind-restart independence),
`llm-routing-and-budget.md` §7.1/A-2 (R-1), `death-mechanics.md` §6.2 (R-3/R-6).

**What this is.** The one ordered command sequence an operator follows cold to bring up a
demo-ready world: the daemon binary and the mod jar as two independently started artifacts
(decision-0001), cast seeded and bound, restartable without losing the villagers' memories
(T-4). Followed verbatim by `specs/021-demo-config/research/bringup-observation.md`.

## 0. Prerequisites

- **JDK 25** (`mod/build.gradle`'s `sourceCompatibility`) and the Gradle wrapper — no
  separate Gradle install; `mod/gradlew` is checked in.
- **Go 1.26+** (`mind/go.mod`).
- **`ANTHROPIC_API_KEY`** — only if genesis will actually run (missing personas, genesis
  mode on). The rehearsal path (§3) needs it for nothing; zero API calls.
- If the host routes model IDs by prefix (TASK-0013's live-run precedent), also export
  `ANTHROPIC_MODEL_PREFIX` (e.g. `cc/`) — confirm once per host with a one-off curl before a
  real genesis run.

## 1. Build the two artifacts

Independently built, independently started (decision-0001) — neither needs the other to
build:

```sh
cd mind && go build -o minddaemon ./cmd/minddaemon
cd mod && ./gradlew build
```

## 2. First-time run-directory setup

Both sides' run directories are gitignored (`mind/.gitignore`'s `run/`, `mod/.gitignore`'s
`run/`) and absent on a fresh checkout. Create them once before the first boot — the daemon's
`listen()` does not create the socket's parent directory, and Loom's dev server needs
`eula.txt` accepted before it will boot at all:

```sh
mkdir -p mind/run mod/run
echo eula=true > mod/run/eula.txt
```

**Unattended/no-player runs only** (a rehearsal, an observation, a recorded take with nobody
connected): after the first boot creates `mod/run/server.properties`, set
`pause-when-empty-seconds=-1` in it — the vanilla default (60) pauses the whole tick loop,
and with it every villager and every mind session, once no player has been online for a
minute (the same setting `specs/014-augmented-villager`'s and `specs/019-death-remains`'s
unattended observations already establish). A demo run with a player connected the whole
time needs no change here.

## 3. Start the two artifacts — either order works

The socket path is the join point (decision-0004): the daemon **listens**, the mod **dials
with retry** (`WireClient.dialWithBackoff`, 250ms doubling to 10s, proven at 1474 failed
dials in `specs/019-death-remains`). Point the daemon's `-socket` at the mod's own default
(`kithcraft.socket` system property, default `kithcraft.sock` relative to the server's run
directory — `BodySession`'s doc) so no `-D` override is needed on the mod side:

```sh
SOCK="$(pwd)/mod/run/kithcraft.sock"   # run from the repo root
```

**Order A — daemon first:**

```sh
mind/minddaemon -socket "$SOCK" -rundir "$(pwd)/mind/run" -genesis=false &
(cd mod && ./gradlew runServer)
```

**Order B — server first:**

```sh
(cd mod && ./gradlew runServer) &
mind/minddaemon -socket "$SOCK" -rundir "$(pwd)/mind/run" -genesis=false
```

Either way: the mod's `KithcraftMod.onServerTick` retries the dial about once a second until
the daemon's socket exists (`"[live] mind dial failed, will retry"` in the server log until
it succeeds), and the daemon's `listen()` accepts the first connection whenever it arrives.
Nothing about startup order is load-bearing.

`-genesis=false` (equivalently `MINDDAEMON_GENESIS=0`) is the rehearsal path — see §4. Drop
it (or set it `true`, the default) for a real genesis run with `ANTHROPIC_API_KEY` exported.

## 4. Rehearsal mode — zero API calls

`-genesis=false` forces the zero-call path **unconditionally**, regardless of whether
`ANTHROPIC_API_KEY` happens to be exported (spec.md US1 scenario 2; decoupled from "key
absent" on purpose). With it:

- No persisted personas under `<rundir>/persona/`: `LoadOrGenesisCast` finds every demo cast
  id missing, `rt.Client`/`rt.Digester` are nil, and the daemon **fails loudly** naming the
  missing ids rather than binding a partial cast — this is FR-002/the Edge Cases section's
  "no partial cast bound" rule, not a bug. Pre-seed personas (below) before running rehearsal
  cold, or run one real genesis first.
- Persisted personas present (`<rundir>/persona/{Aldric,Petra,Yenna}.json`, 0444 — either
  from a prior live genesis run or copied in): `LoadOrGenesisCast` **re-binds via `Load`**,
  zero model calls, zero genesis needed. This is the preferred rehearsal path — re-bind, not
  re-genesis.
- A sleep-boundary crossing with no `Digester` logs and skips (no consolidation, no marker
  written — retried at the next boundary once a key is available). No E1/E6 calls happen at
  any point.

## 5. The rulings and the knobs

**R-1 (ruling, not a knob).** The vanilla daylight cycle is kept —
`consolidate.CycleTicks = 24000` is a constant, not config (plan.md's design decision 4). At
20 ticks/second that is **9 day/night cycles across a ~3-hour evening**, each with its own
dusk and consolidation — **27 consolidations total** for a 3-villager cast
(`llm-routing-and-budget.md` §7.1/A-2). This is recorded here as a deliberate ruling: nine
dusks is representative play, not an oversight and not the only lever (a slowed cycle was the
alternative considered and rejected for the demo default).

**R-3 — grief-period knob**, config since V5, unchanged here:
`-Dkithcraft.griefPeriodTicks=<ticks>` on the mod's JVM (default 24000, one day/night cycle;
`GriefPeriod.configuredTicks()`). Loom's `runServer` task does not forward `-D` flags from
the `gradlew` invocation — pass it via `JAVA_TOOL_OPTIONS`, which the `java` launcher reads
directly regardless of who forked it (confirmed live in `specs/019-death-remains`):

```sh
JAVA_TOOL_OPTIONS="-Dkithcraft.griefPeriodTicks=1200" ./gradlew runServer
```

**R-6 — danger-tuning knob**, new this task, **OFF by default**:
`-Dkithcraft.dangerTuning=true` (same `JAVA_TOOL_OPTIONS` idiom) despawns a newly-spawned
hostile landing within 24 blocks of a cast member (`DangerTuning.SUPPRESSION_RADIUS`). This
is a **recorded-take-only choice** — it exists so a scripted/recorded demo take cannot lose a
cast member to an unlucky spawn mid-take. It does not touch permadeath or the admitted-causes
table (`death-mechanics.md` §6.2 item 5 is explicit the fix for a losable take is thinning
environmental danger, not changing the rules). Leave it off for representative play.

## 6. Session-end report

Written to stdout **and** appended to `<rundir>/session-report.log` (the `-rundir` passed in
§3) on daemon shutdown — SIGINT/SIGTERM or a clean exit, always, even a zero-call rehearsal
session (every `llm.E1..E6` row printed zeroed, never omitted; `mind/cmd/minddaemon/report.go`).
A `kill <pid>` on the daemon is a graceful SIGTERM: `serve` returns, `main`'s deferred
`rt.Report` then `rt.Close` run in that order before the process exits.

## 7. Shutting down

Stop the mod server first (console `stop`, or SIGTERM the `runServer` JVM) so the world
saves cleanly, then `kill` the daemon (§6 covers what that emits). Killing the daemon alone
mid-session, leaving the mod server running, is exactly T-4's restart-independence case —
see `specs/021-demo-config/research/bringup-observation.md`.
