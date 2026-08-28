// Package consolidate (this file): T004's death carry — RM-7 read as a
// weighting function over conversation-context candidate selection
// (docs/design/death-mechanics.md §3: "a recent death should surface
// disproportionately in dusk conversation for a while, then fade in
// retrieval frequency ... never be silently deleted"). Mirrors
// mind/memory/beliefs.go's effectiveConfidence read-time-arithmetic
// idiom (RM-5) — a pure function of (how long ago, now), no storage
// mutation, no randomness — applied to retrieval weight instead of
// belief confidence, so card AC #5's distribution is asserted directly
// on the curve, never on a sampled draw.
//
// This file only computes the weight. Conversation-context assembly
// (M6, mind/converse — TASK-0017, a sibling in-flight branch) is the
// consumer: SelectionWeights is landed here as an exported hook so M6
// adopts it at merge, with no cross-branch edit (plan.md "Where it
// lives").
package consolidate

import (
	"math"

	"kithcraft/mind/memory"
)

// NormalPresence is RM-7's floor made numeric: every candidate's
// baseline retrieval weight, and the value a witnessed death's weight
// decays toward but never below — "never zero, never deleted."
const NormalPresence = 1.0

// deathCarrySpike is the next-cycle spike multiplier over NormalPresence.
// promptworld I's old salience table rated a witnessed death its single
// highest band ("witnessed death — 10★", death-mechanics.md §3) — reused
// verbatim as the spike this posture carries forward, not re-derived.
const deathCarrySpike = 10.0

// deathCarryDecayPerCycle halves the spike's excess over NormalPresence
// with each elapsed cycle — the same half-life shape RM-5 already uses
// for belief confidence (mind/memory/beliefs.go), applied here to
// retrieval weight.
const deathCarryDecayPerCycle = 0.5

// DeathCarryWeight is FR-005: a witnessed death's retrieval-frequency
// weight at cycle now, given it was witnessed at deathCycle. now <=
// deathCycle (the death's own cycle, or earlier) is NormalPresence — no
// spike yet. now == deathCycle+1 ("next cycle") is the full spike.
// Each cycle after that halves the excess over NormalPresence; the
// result never drops below NormalPresence, however large age grows
// (RM-7: time alone never deletes, so the weight never reaches zero or
// goes negative — it just becomes ordinary).
func DeathCarryWeight(deathCycle, now int64) float64 {
	age := now - deathCycle
	if age <= 0 {
		return NormalPresence
	}
	w := NormalPresence + (deathCarrySpike-NormalPresence)*math.Pow(deathCarryDecayPerCycle, float64(age-1))
	if w < NormalPresence {
		return NormalPresence
	}
	return w
}

// WitnessedDeath is one death event a caller has already classified as
// such. death-mechanics.md §3: a witnessed death is an ordinary
// `sighting` sequence with no special percept type, so classification is
// the caller's read of content — card AC #5 is about the weighting
// curve, not detection, and this package does not attempt the latter.
type WitnessedDeath struct {
	ID    memory.ID
	Cycle int64 // the nightly cycle (consolidation ordinal) it was witnessed in
}

// SelectionWeights is the exported selector hook: given the witnessed
// deaths a conversation-context assembly pass already knows about and
// the current cycle, it returns each one's retrieval-frequency weight.
// An ID absent from the result carries the implicit NormalPresence
// weight — SelectionWeights never returns a value below it, so a caller
// defaulting every other candidate to NormalPresence never under-weights
// by omission.
func SelectionWeights(deaths []WitnessedDeath, now int64) map[memory.ID]float64 {
	out := make(map[memory.ID]float64, len(deaths))
	for _, d := range deaths {
		out[d.ID] = DeathCarryWeight(d.Cycle, now)
	}
	return out
}
