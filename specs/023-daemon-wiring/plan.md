# Implementation Plan: Wire deliberation and conversation into the live daemon (M8)

**Spec dir**: `specs/023-daemon-wiring` · **Branch**: `task-0023-daemon-wiring`

**Constitution check**: `.specify/memory/constitution.md` is an unfilled template —
stated plainly per runbook gate. Planned against the grounding docs:
llm-routing-and-budget.md (§2.2, §5.5, §5.2), the merged specs' recorded
conventions (016 Proposer/ErrDone + WindowItem; 017 Slot/Pool + AmbientPool +
naming; 021 Runtime + world_time crossings + report), `docs/wiki/`
([[overview]]'s not-closed addendum names exactly this gap;
[[body-protocol-seam]]).

## Where it lives

`mind/cmd/minddaemon/` almost exclusively — Runtime grows the composition
layer that calls the four packages. Expected package touches are additive
glue only: a Vendor implementation over the live `seam.Conn` (the contract
M5's loop declared; the fake-vendor tests scripted it — the daemon now
implements it for real), a store→`[]deliberate.WindowItem` snapshot adapter
(the enumeration M5 noted the store lacks — smallest exported read that
serves it, likely in mind/memory), body-token→persona binding in Runtime
(closing I1's ponytail), and FirstTokenLatency capture into the report
(mind/converse already computes it; the report already exists). `mod/` is
expected untouched (triggers are existing percepts); if a mod change is
genuinely needed it is a recorded deviation.

## Design decisions (settled surfaces restated)

1. **Triggers are classifications of existing percepts.** E2: schedule
   transitions and open choices arrive as the percepts the vendor already
   emits (sightings/self_state at breaks; act_result leaving an open
   choice); Runtime classifies — no new percept type. E3: `TriggerE3`
   verbatim. Urgency: `IsUrgent` on the ingest path feeding M5's Interrupt.
2. **One deliberation at a time per villager** (routing's serialization
   posture — the interrupt's "enqueue one" only makes sense against it);
   a per-body deliberation loop goroutine with a trigger queue.
3. **The live Vendor** stamps wire envelopes exactly as TASK-0016's
   evening-test `wireVendor` adapter prescribed (its doc comment is the
   contract) — promote that shape from test scaffolding to
   `cmd/minddaemon`, not a new design.
4. **Pair detection** consumes V3's signal percepts (sightings of the
   partner en route to the meet — the shape TASK-0014 shipped and
   TASK-0017's pregen modeled); Runtime keys Slot fills off them and runs
   Exchange at convergence (both bodies' sessions live in one daemon —
   in-process exchange, speak intents out per side).
5. **Persona binding**: cast entry ↔ body token mapping established at
   attach (the manifest/handshake already carries what's needed — else the
   cast/persona name-matching I1's seeding used); E6's stable prefix gains
   persona text (closing the empty-prefix ponytail); E2/E4 prefixes get the
   same binding.
6. **Rehearsal safety**: every new path no-ops with a log line when the
   client is nil (E6's existing posture).
7. **No new deps, no new Mixins, no protocol extension.**

## Risks / open items

- In-process two-mind Exchange vs per-session isolation: both sessions live
  in one Runtime, so the exchange can run internally — but each speaker's
  speak intent must go out on its OWN session (V2's earshot delivery does
  the rest). Watch the at-most-once slot semantics under live timing.
- The dev-server observation depends on Activity.MEET/WORK timing (known
  substrate questions) — bounded checks, honest records, per precedent.

## Phase map

Phase 1 — the live Vendor + E2/E3 composition + window snapshot + interrupt (US1+US2).
Phase 2 — persona binding + dusk exchange + pregen + ambient pool + latency surfacing (US3+US4).
Phase 3 — real-binary fake-vendor proofs; dev-server observation (FR-006/007).
Phase 4 — gates, wiki re-ground (overview's addendum shrinks), board close.
