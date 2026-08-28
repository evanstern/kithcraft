# Tasks: Persona genesis and the persona firewall

**Input**: specs/013-persona-genesis/ (spec.md, plan.md)
**Prerequisites**: mind/ module (M1/M2/M4 merged), decision-0002, decision-0003,
llm-routing-and-budget.md §1.3 (E1), kithcraft-brief.md #5,
promptworld/internal/persona (source material, read-only)

**Organization**: Phases map 1:1 to phase-scoped dispatches.

## Phase 1 — Persona type, write-once storage, re-bind (US2)

**Goal**: A persona exists on disk exactly once, at 0444, and comes back bound
to the same cast id after restart; the exported surface provably offers no
post-genesis write path.

**Independent test**: `go test ./persona/...` — storage + external
API-surface tests green, no LLM code in the loop.

- [x] T001 Implement mind/persona/persona.go: Persona type (cast id, name,
      values, endogenous desires, anchor line, drift markers, profession,
      biome variant); verify the demo cast pairing against the vendor mod's
      seeded cast (TASK-0014 cast/) and record it
- [x] T002 Implement mind/persona/files.go: WriteOnce (0444, refuses existing),
      Load (read-only, binds by cast id, errors on unknown/missing — never
      regenerates)
- [x] T003 External API-surface test (persona_external_test): reflection over
      the exported surface asserts no exported mutation path for an existing
      persona (AC #2's "no code path to attempt"); restart/re-bind unit test
      over temp files (AC #4 unit half)

**Checkpoint**: the structural half of the firewall exists and is proven at
the type level.

**Deviation from plan.md (T001 verification), recorded here per plan.md's
own instruction:** plan.md's "Key decisions already settled" section assumed
the demo cast is uniform Plains with farmer/librarian/cleric professions.
Reading the vendor mod's actual seeded cast
(`mod/src/main/java/dev/kithcraft/mod/cast/Cast.java`, `Cast.MEMBERS`) shows
three DISTINCT profession × biome-variant pairs instead: Aldric —
armorer/plains, Petra — farmer/desert, Yenna — fisherman/taiga. This is the
real decision-0002 pairing and is what `mind/persona/persona.go` records in
its doc comment; no code changed as a result (Persona's fields already carry
Profession/BiomeVariant generically), only the assumed values were corrected
at the source of truth.

## Phase 2 — Model-free validator (US3 + US4 lexicon half)

**Goal**: Drift is rejected by arithmetic — anchor echo + drift lexicon +
authored moralizing lexicon — with zero model involvement, provable.

**Independent test**: `go test ./persona/...` — validator tests green,
including the no-llm-import structural test.

- [ ] T004 Implement mind/persona/validate.go: anchor echo (verbatim under
      whitespace/case normalization) + drift-marker matching (word-boundary,
      case-insensitive) + reject-with-reason; document the honest limit
      (stated drift only; subtle drift parked for a model-judged validator)
- [ ] T005 Implement mind/persona/moralizing.go: authored cast-wide moralizing
      lexicon (politeness-policing spell-breaker words); validator applies it
      to every persona in union with the persona's own markers
- [ ] T006 Named tests: anchor-echo reject/accept, drift-marker reject at word
      boundary any case (AC #3), moralizing reject cast-wide (AC #6 lexicon
      half), plus the structural test that validate.go imports no llm code

**Checkpoint**: the validatory half of the firewall exists; rejection is a
testable 100% guarantee.

## Phase 3 — E1 genesis (US1 + US4 prompt half)

**Goal**: Three cast entries in, three Opus-5 E1 calls out, each producing a
structured persona with anchor + markers + pairing; the prompt pins the dial
conservative and instructs away from moralizers.

**Independent test**: `go test ./persona/...` — genesis tests green against a
mocked client (no network).

- [ ] T007 Implement mind/persona/genesis.go: per-entry E1 call via the llm
      client at class E1's config (Opus 5, no cache, pre-session), structured
      persona output (schema following mind/llm/structured.go's idiom; JSON-
      in-text strict-decode fallback recorded if structured fights E1 config),
      profession/biome pairing carried into the persona
- [ ] T008 The genesis prompt: conservative weirdness dial (brief #5), values
      + endogenous desires demanded, anchor line + drift markers demanded as
      output fields, explicit anti-moralizing instruction
- [ ] T009 Named tests with mocked client: exactly three calls for three
      entries on E1 (AC #1 unit half), pairing present in output (AC #5),
      prompt-content tests for dial + anti-moralizing instruction (AC #6
      prompt half); generated-persona-through-validator round-trip

**Checkpoint**: genesis is complete and cheap to prove; only the live run
remains.

## Phase 4 — Live genesis, restart proof, gates, closure (US1/US2 live halves)

**Goal**: The real thing: three Opus-5 calls produce three 0444 personas that
survive a daemon restart; all gates green; wiki honest.

**Independent test**: live run recorded (PR description); `go vet ./...` +
`go test ./...` green; freshness probe green.

- [ ] T010 Live genesis run (ANTHROPIC_API_KEY from root .env): three real E1
      calls, three 0444 files, validator accepts all three; record model
      served, token counts, and any drift-rejection retries in the PR (AC #1
      live half)
- [ ] T011 Restart/re-bind proof over the real files: reload binds each
      persona to its cast id (AC #4 live half); tick card ACs #1–#6 with
      citing tests/observations
- [ ] T012 Gates + re-ground: go vet + go test green; wiki notes whose sources
      this PR touched re-verified honestly (promptworld-lineage §firewall,
      overview); CAPSULES regenerated if any description changed; freshness
      probe green
