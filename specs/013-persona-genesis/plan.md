# Implementation Plan: Persona genesis and the persona firewall

**Spec dir**: `specs/013-persona-genesis` · **Branch**: `task-0013-persona-genesis`

**Constitution**: `.specify/memory/constitution.md` is an unfilled template —
stated plainly per the runbook's per-task artifact rule. This plan is checked
against the grounding docs instead: docs/design/kithcraft-brief.md (ratified),
decision-0002/0003, docs/design/llm-routing-and-budget.md,
docs/design/demo-build-plan.md §3.2, docs/wiki/ ([[promptworld-lineage]],
[[body-protocol-seam]]), and project CLAUDE.md.

## Where it lands

New package `mind/persona/` in the existing Go module. Consumers-to-be (M5
deliberation, M7 consolidation) are NOT wired here — M3 delivers the package,
its files, and its proofs; wiring into prompt assembly's persona slot happens
in the tasks that consume it (the `prompt` package already types a persona
stable-prefix input; this task only has to produce content that slot can take).

## Structure

```
mind/persona/
  genesis.go      — E1 calls via llm.Client; structured persona output; cast pairing
  persona.go      — Persona type: name, values, desires, anchor, drift markers, cast id, profession/biome
  files.go        — write-once 0444 storage; Load/re-bind; refuse-regenerate
  validate.go     — model-free validator: anchor echo + lexicon (NO llm import)
  moralizing.go   — authored cast-wide moralizing lexicon (doctrine, not generated)
  *_test.go       — unit tests (mocked client), external API-surface test package
```

The API-surface guarantee (AC #2, AC #3's no-model-call) follows the M2/S2
idiom: an external test package (`persona_external_test`) reflects over the
exported surface and asserts no mutation path exists; a second check greps the
validator's imports (structural: `validate.go`'s file-level imports exclude
`mind/llm` — enforced by a test reading the source, same as S2's reflection
lock, since Go has no per-file import visibility).

## Key decisions already settled (no re-litigation)

- E1 on Opus 5, 3 calls, pre-session, unbounded latency, not cached — already
  encoded in `mind/llm/classes.go` (E1 row). Genesis uses `llm.E1` config
  as-is; the *product's* model choice is not the implementer's tier.
- Two-half firewall — decision-0003 §8 "cleanest carry": port the design from
  `promptworld/internal/persona` (read as source material, import nothing).
- Generated-not-authored personas (brief #5) — the delta from I: E1 emits
  anchor + drift markers as structured output; the moralizing lexicon stays
  authored (spec's "What changes" section).
- Weirdness dial conservative — a genesis-prompt constant, tested by name.
- Cast pairing = profession × biome variant (decision-0002); the three-entry
  demo cast matches TASK-0014's seeded cast (Plains; farmer/librarian/cleric
  professions per its cast/ package) — verified at implementation against
  `vendor mod`'s cast seeding, recorded in tasks.md phase 1.

## Risks / open items

- Structured output for E1: `mind/llm/structured.go` exists (M4's Intent/
  Digest shapes); persona genesis adds its own schema. If structured output
  fights the E1 no-cache/effort config, fall back to JSON-in-text with strict
  decode — a recorded implementation note, not a design change.
- Live genesis needs `ANTHROPIC_API_KEY` (root `.env` per operator, checkpoint
  3 satisfied 2026-08-28). Unit tests never touch the network.

## Phases (mirror tasks.md)

1. Persona type, storage, re-bind — write-once 0444, refuse-regenerate,
   external API-surface test (ACs #2, #4 groundwork).
2. Model-free validator — anchor echo + drift lexicon + authored moralizing
   lexicon, no-llm-import structural test (ACs #3, #6 lexicon half).
3. Genesis — E1 prompt (conservative dial, anti-moralizing instruction, named
   prompt tests), structured output, cast pairing, mocked-client tests
   (ACs #1, #5, #6 prompt half).
4. Live run + closure — real 3-call genesis, restart/re-bind proof over real
   files, gates, wiki re-verification, AC ticks (ACs #1, #4 live halves).
