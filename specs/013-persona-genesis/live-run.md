# Live genesis run — record (T010)

**Attempted**: 2026-08-28. **Result**: STOPPED — no persona was generated.

## What happened

`ANTHROPIC_API_KEY` and `ANTHROPIC_BASE_URL` were exported from the repo root's `.env`
into the process env (never printed/logged/committed) and
`go test -tags=live -run 'TestLiveGenesis|TestLiveRestart' -v ./persona/...` was run
(`mind/persona/genesis_live_test.go`, the live-tagged harness this task adds).

The first E1 call (cast entry `Aldric`) failed:

```
POST ".../v1/v1/messages": 404 Not Found
{"error":{"message":"No active credentials for provider: anthropic","type":"invalid_request_error","code":"model_not_found"}}
```

Per the dispatch's rule ("if a call fails on auth/billing, STOP and report — do not
retry into a bill"), the harness's one sanctioned regeneration attempt fired
automatically on this first run (its retry path did not yet distinguish a validator
rejection from a transport/auth error) and failed the same way. **No third call was
made.** Genesis stopped there. The harness was corrected afterward
(`rejectedAtBirth` in `genesis_live_test.go`) so a future run's retry only fires on an
actual validator rejection (`*RejectionReason`) — any other error, including this
auth-shaped one, now fails fast with no retry.

- **Cast entries attempted**: 1 of 3 (`Aldric`) — 2 calls, both rejected before reaching
  a model (no `usage` in either response).
- **Personas generated / written**: 0. `mind/run/persona/` was created (empty,
  gitignored) and holds no files.
- **Model served**: unknown — no call reached the model; the request path is what's
  diagnostic (below).
- **Tokens**: none billed by either call (both are proxy-level 404s with no `usage`
  block in the response).

## Diagnostic clue (unconfirmed — not a fix, not applied)

The Anthropic Go SDK (`v1.58.0`, `message.go` lines 71/97) hardcodes its request path
as the literal string `"v1/messages"` and joins it onto `ANTHROPIC_BASE_URL`. The
failing request's URL shows a doubled `/v1/v1/messages` segment. If the configured
`ANTHROPIC_BASE_URL` already ends in `/v1`, that doubling is exactly what a base URL
without a trailing `/v1` would avoid. This is offered as a lead, not a diagnosis — the
`.env` value itself was never read or altered by this task (only exported into the
process env, per the "never print/echo" rule), so whether this is the actual cause of
the "no active credentials" response, or a coincidental red herring next to a genuinely
unconfigured credential, is unverified.

## What this blocks

- T010 (this task): not done — zero real personas exist.
- T011 (restart/re-bind proof over real files): blocked on T010 — the live-tagged
  `TestLiveRestart_LoadsAndBindsRealFiles` test correctly skips (no files to load) and
  was not forced green.
- Phase 4's checkpoint ("real thing: three Opus-5 calls produce three 0444 personas
  that survive a daemon restart") is not reached.

## What is NOT blocked

- Phases 1-3 (already merged commits e9e2673, ac6d9ec, 0ba6a80) are unaffected — all
  mocked-client unit tests still pass (`go test -count=1 ./...` green, no live
  credentials needed).
- `go vet ./...` / `go test -count=1 ./...` (T012's gate half) are green independent of
  this blocker.

## Recommendation

An operator needs to confirm the root `.env`'s `ANTHROPIC_BASE_URL`/credentials are
live and reachable (possibly by checking the trailing-`/v1` hypothesis above) before
re-running:

```
cd mind && (set -a && source ../.env && set +a && go test -tags=live -run 'TestLiveGenesis|TestLiveRestart' -v ./persona/...)
```

This is an environment/credentials question, not a code judgment call within this
dispatch's scope — flagging rather than guessing further, per the sweep's escalation
rule.
