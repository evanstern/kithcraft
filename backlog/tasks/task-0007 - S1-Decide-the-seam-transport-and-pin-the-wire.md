---
id: TASK-0007
title: S1 - Decide the seam transport and pin the wire
status: In Progress
assignee: []
created_date: '2026-08-21 23:35'
updated_date: '2026-08-22 04:01'
labels:
  - seam
  - m-0-build
milestone: m-0
dependencies: []
documentation:
  - docs/design/demo-build-plan.md
  - docs/design/body-protocol-v0.md
  - docs/design/llm-routing-and-budget.md
priority: high
ordinal: 7000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
As a future implementer, I want the mind<->vendor wire chosen and its framing pinned by fixtures, so that the Go and Java sides can be built independently without discovering they disagree at first contact.

**Scope boundary.** Choose among **real wires only** — UDS, TCP, stdio — per decision-0003's one-way narrowing (a Go daemon against a JVM mod forecloses in-process deliberately, because it makes an SI-1 breach structurally impossible rather than merely forbidden). Weigh against T-1..T-7: push not pull; ordered-or-reorderable per body; long-lived sessions; mind restart independent of vendor restart; message-oriented with a schema the fake vendor can produce without an engine; backpressure that sheds `background` only and never splits an `observation`; process-separable but not process-required. Deliverable: a decision record, a framing/serialization spec as the spec-002 successor, and **golden message vectors** — one fixture per percept type, per intent shape, and the `session_open` handshake — that both implementations run against.

**Done proves.** The wire is decided with T-1..T-7 answered one by one; the golden vectors exist and are language-neutral; a trivial Go encoder and a trivial Java decoder each round-trip every vector. Nothing else is built.

**Depends on.** Nothing — this is the contract-shaped head of the graph.

**References.** docs/design/demo-build-plan.md section 3.1 (S1) is the plan of record. Ratified surfaces consumed: docs/design/body-protocol-v0.md (Q-1, T-1..T-7, the seam invariants), decision-0003 + docs/design/llm-routing-and-budget.md (transport narrowed to real wires; the decomposition splits at the seam), decision-0001 (Fabric server-side mod).

**Suggested tier: `opus` — proposed escalation (next sweep's runbook decides.** A decision the spec constrains but does not settle, analogous to the TASK-0002/TASK-0004 escalations. Taking it is the operator's checkpoint at sign-off; ratifying the resulting decision is an operator checkpoint regardless of tier.

Spec: specs/007-seam-transport
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The transport is decided among real wires only (UDS/TCP/stdio) and T-1..T-7 are each answered explicitly in a decision record
- [x] #2 A framing/serialization spec exists as the spec-002 successor
- [x] #3 Golden message vectors exist and are language-neutral: one per percept type, one per intent shape, and the session_open handshake
- [x] #4 A trivial Go encoder and a trivial Java decoder each round-trip every golden vector
- [x] #5 Nothing beyond the decision record, framing spec and vectors is built
- [ ] #6 Spec phase: Phase 1 — The decision and the T-matrix (US1)
- [ ] #7 Spec phase: Phase 2 — The framing/serialization spec (US2)
- [ ] #8 Spec phase: Phase 3 — Golden vectors and both round-trips (US3)
- [ ] #9 Spec phase: Phase 4 — Closure: leak audit, wiki, board
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Claimed by sweep-0007-0022 orchestrator 2026-08-21 on branch task-0007-seam-transport (worktree .worktrees/task-0007). Tier: opus (operator-escalated at runbook sign-off 2026-08-21 — transport choice is design work the spec constrains but does not settle). Model ID: cc/claude-opus-5[1m], fallback cc/claude-opus-4-8[1m]. Served model recorded at dispatch.

Phase 1 done (6003775, opus verified from transcript — claude-opus-5, ~139k subagent tokens): wire chosen — UDS (AF_UNIX SOCK_STREAM), mind listens / vendor dials; decision-0004 created (proposed, operator ratifies at PR); T-matrix filled in research.md with host-verified evidence (SOCK_SEQPACKET unavailable on macOS 26.5.1; JDK 26 UDS via SocketChannel; go1.26.4 present); stdio rejected on structural T-4 failure, TCP on reachability/permission grounds.

Phase 2 done (b023927, opus verified — claude-opus-5, ~142k subagent tokens): docs/design/seam-wire-v0.md written as the spec-002 successor. Framing: one connection per vendor with per-body session_open/close multiplexed by envelope body; 4-byte big-endian length prefix + canonical UTF-8 JSON (C-1..C-10), 1 MiB cap; byte-exact round-trip equality (asymmetric: emit canonical, accept conforming); seq assigned at enqueue so gaps count shed background percepts; vectors append-only within a MAJOR. Two flagged judgment calls recorded in-doc: continuity matched by body token not previous_session (T-4), and the spec-side leak audit written complete with Phase 4 extending it over vectors.

Phase 3 done (d380acf, opus verified — claude-opus-5, ~162k subagent tokens): 17 vectors in seam/vectors/ (9 percept + 3 intent + 2 session + 3 error) matching contracts/vectors.md exactly, census enforced both directions. Go harness green (go1.26.4: ok kithcraft/seam/go-roundtrip) and Java harness green (OpenJDK 26.0.2: 91 passed, 0 failed, over 17 vectors) — both re-run independently by the orchestrator. Mutation-checked: corrupted byte and decoded-only drift both caught. No framing-spec fix needed; note: Go encoding/json cannot be the canonical encoder (C-7), both harnesses hand-roll the writer.

Phase 4 done (opus — claude-opus-5): leak sweep, scope check, wiki re-ground, board sync.

T013 leak sweep — ran body-protocol-v0 §12's six passes over docs/design/seam-wire-v0.md AND all 17 vectors, recorded in seam-wire-v0.md §7 (now §7.1 document / §7.2 vectors / §7.3 combined result). Vector sweep: all six passes CLEAN, no leak, no vector edited, no fix required. Method beyond grep: full census of all 131 distinct string values, all 66 JSON keys, and every numeric key across the set; plus mechanical checks (every frame_hex = 4-byte BE length + exactly that many bytes, nothing trailing; every body one well-formed UTF-8 JSON object; zero \\u escapes; one literal-UTF-8 em dash per C-7). Five findings recorded as verified-clean (V-1..V-5), the sharpest being V-1: sound_kind 'k:snarl' is the one k: token absent from the manifest's salient_kinds — correct, since salient_kinds is the report-at-a-place vocabulary and the protocol declares no sound-kind manifest field; whether a future MINOR wants one is a protocol question, left open. Harnesses re-run green after the edit (go test -count=1 ok; java 91 passed/0 failed over 17 vectors).

T014 FR-007 scope — CLEAN. 38 files in git diff origin/main...HEAD, none outside backlog/ (decision-0004 + this card), docs/design/seam-wire-v0.md, docs/design/sweep-0007-0022-runbook.md, seam/, specs/007-seam-transport/, .specify/feature.json, and docs/wiki/ (this phase's own re-grounding). AC #5 checked. AC #1 checked (true since Phase 1).

T015 wiki — docs/wiki/body-protocol-seam.md re-verified claim by claim against its sources, then amended: the two stale spots removed ('transport is deliberately undecided'; 'transport remains an open question'), new 'What the wire binds' section added (UDS + listen/dial + why T-4/T-3 decided it, length-prefixed canonical JSON, bare header, session=connection, no resume/ack, 17 vectors + two agreeing harnesses), operational notes now carry the vectors-append-only rule and the standing obligation to re-run the six passes on any added vector, and Q-1 restated as closed-but-proposed (operator ratifies at PR). Sources grew by docs/design/seam-wire-v0.md and decision-0004's file; description changed, so docs/wiki/CAPSULES.md regenerated; INDEX.md one-liner updated. Re-pinned verified_against to d380acf (the last CONTENT commit — Phase 3's vectors; the Phase 4 audit edit lands in the same commit as this pin). Grepped every other note for transport/wire/Q-1 staleness: none found (v1-demo's only mention is the tier-escalation note, still accurate).

T016 tasks.md T013-T016 ticked. Not touched: Definition of Done items, status (both the orchestrator's).
<!-- SECTION:NOTES:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Tests pass
- [ ] #2 Docs and wiki are updated and pass freshness tests
- [ ] #3 Spec and Backlog are in sync
<!-- DOD:END -->
