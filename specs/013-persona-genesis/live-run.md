# Live genesis run — record (T010/T011)

**Attempted**: 2026-08-28 (re-run after the routing-prefix fix). **Result**: SUCCESS —
three real E1 calls produced three validator-accepted personas, written 0444, and all
three re-bound correctly on a simulated restart.

## Diagnosis (orchestrator-verified, not re-diagnosed by this dispatch)

The prior attempt's 404 ("no active credentials for provider: anthropic") was not a
credentials problem and not the doubled `/v1/v1/messages` path shape flagged as an
unconfirmed lead in this file's earlier revision — that path shape is correct. The
actual cause: the proxy behind the root `.env`'s `ANTHROPIC_BASE_URL` routes by
model-ID *prefix*. Bare API IDs as written in `mind/llm/classes.go`
(e.g. `claude-opus-5`) have no provider route on this host and 404; the host form
(`cc/claude-opus-5`) returns 200 — confirmed by the orchestrator via direct curl before
this dispatch, and `cc/claude-opus-4-8` also 200s.

The fix (`mind/llm/client.go`, commit 186a6ba): a new env var
`ANTHROPIC_MODEL_PREFIX` (default `""`, no-op) that `buildParams` prepends to the
request's model ID only, at request-build time. `classes.go`'s canonical IDs
(`ModelOpus5 = "claude-opus-5"`, etc.) are unchanged — they stay the product's
ratified choice; `Accounting` is keyed by `Class`, not by the wire model string, so
accounting/logs are unaffected by the prefix. This re-run set
`ANTHROPIC_MODEL_PREFIX=cc/` in the process env alongside the existing
`ANTHROPIC_API_KEY`/`ANTHROPIC_BASE_URL` exports from the root `.env` (never
printed/logged/committed).

## Run

```
cd mind && (set -a && source ../.env && set +a && export ANTHROPIC_MODEL_PREFIX=cc/ && \
  go test -tags=live -run 'TestLiveGenesis|TestLiveRestart' -v ./persona/...)
```

`TestLiveGenesis_ThreeCastEntries` (T010): three E1 calls, one per demo cast entry,
each accepted by the validator on the **first** attempt — zero rejections, zero
retries. Model served (from the API response's `msg.Model`, not just requested):
`claude-opus-5` on all three calls.

| Cast entry | Persona name | Input tokens | Output tokens |
|---|---|---:|---:|
| Aldric (armorer/plains) | Merrow Vand | 3223 | 429 |
| Petra (farmer/desert) | Sefa | 3223 | 390 |
| Yenna (fisherman/taiga) | Ottiline Vask | 3225 | 518 |

Three files written 0444 at `mind/run/persona/{Aldric,Petra,Yenna}.json`
(gitignored — `mind/.gitignore`, `run/`).

**Rejections/retries**: none. All three passed `Validate` on the first generation; the
one-sanctioned-retry path (`rejectedAtBirth`) was not exercised.

## Persona summaries

- **Aldric → Merrow Vand** (armorer, plains): a plains armorer who values repairable,
  battle-tested gear over anything decorative — anchor: *"I make things that take the
  hit so a person doesn't, and I'd rather be told what mine failed at than thanked for
  it."*
- **Petra → Sefa** (farmer, desert): a desert farmer whose life is measured in water and
  seed stock, counting every drop to keep the terrace green through the dry months —
  anchor: *"I keep things alive in a place that would rather they weren't, and I do it
  by counting."*
- **Yenna → Ottiline Vask** (fisherman, taiga): a taiga fisherman who reads ice and
  water by hand rather than by being told, still finding the lake surprising after a
  lifetime on it — anchor: *"I fish the same lake my whole life and it still surprises
  me — that's why I go back out."*

## Restart/re-bind proof (T011)

```
cd mind && go test -tags=live -run 'TestLiveRestart' -v ./persona/...
--- PASS: TestLiveRestart_LoadsAndBindsRealFiles (0.00s)
```

`TestLiveRestart_LoadsAndBindsRealFiles` now runs (does not skip) against the real
files T010 wrote — a fresh `Load` call with no in-memory state carried over binds all
three cast ids (`Aldric`, `Petra`, `Yenna`) to their generated personas. Card AC #4's
live half is proven.

## Gates

`cd mind && go vet ./... && go test -count=1 ./...` — green (all packages, no live
credentials needed; the live-tagged tests stay excluded from the default build per
their `//go:build live` tag).

## Card acceptance criteria — status after this run

(TASK-0013's AC list, board file not edited by this dispatch — recorded here for the
next spec-bridge sync to mirror.)

- **AC #1** (three E1-on-Opus-5 calls produce three personas, weirdness dial
  conservative): proven live this run — this file's table, `msg.Model ==
  "claude-opus-5"` on all three calls.
- **AC #2** (written once at 0444, no post-genesis write path to attempt): proven at
  the type level, Phase 1 (`persona_external_test.go`) — unaffected by this run; the
  three real files this run wrote are 0444 (`ls -l mind/run/persona/`).
- **AC #3** (model-free validator rejects a drifted persona without a second model
  call): proven by Phase 2's unit tests (`validate.go`); this run's three outputs all
  passed on the first attempt, so the live path did not itself exercise a rejection.
- **AC #4** (personas survive a daemon restart and re-bind to the same bodies): proven
  live this run — `TestLiveRestart_LoadsAndBindsRealFiles` PASS above, over the real
  files T010 wrote.
- **AC #5** (profession/biome-variant pairing carried into the persona): proven live —
  the table above and the JSON files' `profession`/`biome_variant` fields match each
  cast entry (Aldric/armorer/plains, Petra/farmer/desert, Yenna/fisherman/taiga).
- **AC #6** (no moralizing persona template; drift lexicon catches stated moralizing):
  proven by Phase 2/3 unit tests; none of the three live outputs tripped the
  moralizing lexicon (all three wrote cleanly through `Validate` first try).
