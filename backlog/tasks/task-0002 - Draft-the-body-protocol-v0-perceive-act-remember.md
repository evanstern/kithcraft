---
id: TASK-0002
title: Draft the body protocol v0 (perceive / act / remember)
status: In Progress
assignee: []
created_date: '2026-08-19 18:36'
updated_date: '2026-08-21 19:49'
labels:
  - design-decision
  - architecture
milestone: m-0
dependencies:
  - TASK-0001
documentation:
  - docs/design/kithcraft-brief.md
priority: high
ordinal: 2000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
As a future body vendor implementer (Minecraft mod today, an owned engine in V2), I want a world-agnostic body protocol v0 (perceive / act / remember), so that minds never couple to Minecraft and a new world is a second vendor, not a rewrite.

Context: this is the brief's anti-corner move. Draft v0 of the protocol: the perceive/act/remember surface, message shapes, and the perception model (what a villager sees/hears), porting promptworld I's epistemic hygiene rules (an agent knows only what it saw or was told, with provenance). Reference doctrine lives in promptworld I's docs/wiki/ (start from INDEX.md, load notes just-in-time); nothing imports I's code. Choice of mod stack (TASK-0001) informs what the first body vendor can feasibly expose.

Spec: specs/002-body-protocol-v0
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A v0 protocol document exists covering perceive, act, and remember message shapes with a versioning story
- [ ] #2 The perception model is specified: what a villager sees/hears and how provenance is attached (epistemic hygiene ported from promptworld I doctrine)
- [ ] #3 The protocol is demonstrably world-agnostic: no Minecraft types leak across the seam, and the doc shows how a second body vendor would implement it
- [ ] #4 Minds are testable without booting Minecraft: the doc specifies a test/fake body vendor
- [ ] #5 Spec phase: Phase 1 — Doctrine port & feasibility
- [ ] #6 Spec phase: Phase 2 — Protocol draft
- [ ] #7 Spec phase: Phase 3 — Prove the seam
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Dispatch tier: opus (cc/claude-opus-5[1m], fallback cc/claude-opus-4-8[1m]) — ESCALATED by operator at runbook sign-off 2026-08-21: protocol drafting is design work on the seam (the project's one-way door), a judgment call the spec does not settle. Served model to be verified from Phase 1 transcript.
<!-- SECTION:NOTES:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Tests pass
- [ ] #2 Docs and wiki are updated and pass freshness tests
- [ ] #3 Spec and Backlog are in sync
<!-- DOD:END -->
