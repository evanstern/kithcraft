# Spec 002 — Body protocol v0 (perceive / act / remember)

**Board task:** TASK-0002 · **Milestone:** One real evening (v1 demo)
**Direction source:** docs/design/kithcraft-brief.md (ratified 2026-08-19 — do not
relitigate); decision-0001 (Fabric, accepted 2026-08-20) fixes the first vendor's
substrate.

## Problem

The brief's single architecture posture — the anti-corner move — is a world-agnostic
protocol between mind daemon and world: **perceive / act / remember**, with the
Minecraft mod as merely the first *body vendor*. The protocol does not exist yet; until
it does, every mind-side and vendor-side task (TASK-0004's daemon, TASK-0006's demo
plan) has no contract to build against, and the seam is posture instead of contract.
This task drafts v0: the message shapes, the perception model, and the versioning story.

## Requirements (mapped to the card's acceptance criteria)

### R1 — Protocol document with message shapes and versioning (card AC #1)

A v0 protocol document at `docs/design/body-protocol-v0.md` covering:

- The three surfaces — **perceive** (world → mind), **act** (mind → world), and
  **remember** (the mind's durable memory contract) — each with concrete message
  shapes (field-level schemas, not prose sketches).
- A versioning story: how v0 evolves without breaking vendors (version negotiation or
  explicit version field, what constitutes a breaking change, additive-change rules).
- Transport-agnostic definition: shapes are defined as data (JSON schema or equivalent
  serialization-neutral form), with transport an explicit non-commitment of v0.

### R2 — Perception model with epistemic hygiene (card AC #2)

The perception model specifies what a villager sees and hears, and how provenance is
attached:

- Sensory channels and their limits (sight range/occlusion posture, hearing, being
  told), at a level the first vendor can feasibly implement on the Fabric brain
  substrate (sensors, memory modules).
- **Epistemic hygiene ported from promptworld I doctrine:** an agent knows only what
  it saw or was told, with provenance on every percept and memory. Port the doctrine
  from promptworld I's `docs/wiki/` (INDEX-first, just-in-time notes — e.g. its
  origin-provenance classifier and situated-memory rules); port doctrine, never code.

### R3 — Demonstrably world-agnostic (card AC #3)

- No Minecraft types, identifiers, or coordinate conventions leak across the seam;
  the doc states the abstraction rule for world-specific concepts (blocks, items,
  mobs) and shows it applied.
- The doc demonstrates world-agnosticism concretely: a sketch of how a **second body
  vendor** (the V2 owned engine, or a trivial text world) would implement the same
  surface.

### R4 — Testable without Minecraft (card AC #4)

- The doc specifies a **fake/test body vendor**: a minimal in-memory vendor
  implementing the full protocol surface, sufficient to exercise a mind end-to-end
  (scripted percepts in, asserted acts out) without booting Minecraft.

## Out of scope

- Implementing the protocol, the fake vendor, or any code — v0 is a contract document.
- The mind daemon's language/runtime (TASK-0004) and internals; the protocol
  constrains only the seam.
- The entity implementation choice (TASK-0003) — the protocol must be neutral to it.
- Transport/wire choice (gRPC vs WebSocket vs stdio) — named as an open question with
  constraints, not decided.

## Done means

Protocol doc exists with all four requirement areas covered; card ACs #1–#4 map to it;
[[body-protocol-seam]] wiki note grows the protocol doc as a source and is re-verified
in the same PR (its Operational notes demand this — the seam moves from posture to
contract).
