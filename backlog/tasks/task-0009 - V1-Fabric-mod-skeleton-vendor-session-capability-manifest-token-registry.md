---
id: TASK-0009
title: 'V1 - Fabric mod skeleton, vendor session, capability manifest, token registry'
status: To Do
assignee: []
created_date: '2026-08-21 23:36'
labels:
  - vendor
  - m-0-build
milestone: m-0
dependencies:
  - TASK-0007
documentation:
  - docs/design/demo-build-plan.md
  - docs/design/entity-implementation-comparison.md
  - docs/design/body-protocol-v0.md
priority: high
ordinal: 9000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
As a future implementer, I want a mod that can hold a protocol session and hand out tokens that still mean the same thing after a restart, so that the mind's memories survive the world they were formed in.

**Scope boundary.** The mod jar and its dev-server setup against the target Minecraft version — and **re-verify the Yarn mappings and version-dependent facts here**, since `villager-brain-api`'s symbol names were checked at yarn-1.21.3+build.1 and routing's A-2 daylight arithmetic is flagged "verify against the target version before building." The transport client per S1. The `session_open` handshake: `time_unit` (a declared unit, **never ticks**), continuity, capabilities — declared `percept_types` (at minimum the four-type floor: `act_result`, `observation`, `sighting`, `speech`), origins, verbs with target shapes, `salient_kinds` in role-annotated form, bearings, distance bands. The **token registry** — `body`/`place`/`thing_id`/`kind` -> referent, persisted across sessions: the vendor's hardest obligation, and tokens are never reused. No client jar (decision-0002).

**Done proves.** On a dev server: the mod loads, a villager's session opens against the daemon (or a stub mind) and the handshake round-trips S1's golden vectors. **The manifest is identical for every body and does not vary with world state** — the L-7 test, because a vendor populating `salient_kinds` from what is nearby has made `session_open` a "what is around me" query and defeated SI-1 before the first percept. Tokens issued before a server restart still resolve to the same referents after it.

**Depends on.** S1 (the wire and its golden vectors).

**References.** docs/design/demo-build-plan.md section 3.3 (V1) is the plan of record. Ratified surfaces consumed: decision-0001 (Fabric server-side mod), decision-0002 + docs/design/entity-implementation-comparison.md (augmented vanilla VillagerEntity, bounded ~4-injection Mixin surface, no client jar), docs/design/body-protocol-v0.md (session_open, capability manifest, token discipline, SI-1).

**Suggested tier: `sonnet` (next sweep's runbook decides).**
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 On a dev server the mod loads and a villager's session opens against a daemon or stub mind, round-tripping S1's golden vectors
- [ ] #2 The capability manifest declares at minimum the four-type percept floor (act_result, observation, sighting, speech), origins, verbs with target shapes, role-annotated salient_kinds, bearings and distance bands
- [ ] #3 The manifest is identical for every body and does not vary with world state (the L-7 test)
- [ ] #4 time_unit is a declared unit, never raw ticks
- [ ] #5 The token registry persists across sessions: tokens issued before a server restart resolve to the same referents after it, and tokens are never reused
- [ ] #6 Yarn mappings and version-dependent facts are re-verified against the target Minecraft version
- [ ] #7 No client jar is produced
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Tests pass
- [ ] #2 Docs and wiki are updated and pass freshness tests
- [ ] #3 Spec and Backlog are in sync
<!-- DOD:END -->
