// Package deliberate (this file): the K=10 situated memory window
// (spec.md FR-006, plan.md design decision 3, routing §2.3's memory-window
// line, card AC #6) — a pure selector over a caller-supplied snapshot of
// weighted memories: the top K-2 by recency-decayed salience, plus 2
// seeded serendipity picks from the snapshot's older half. Salience is a
// mind-side concept only (body-protocol-v0.md §2.8: "No salience,
// importance, weight, or memorability field may exist on any percept"),
// so this file never reads a wire percept directly — a caller (memory
// consolidation, a later phase) is responsible for assembling the
// []WindowItem snapshot from mind/memory's store. Pure function over
// (snapshot, now, seed) per plan.md decision 3: no wall-clock, no
// wiring, testable without time mocking.
package deliberate

import (
	"math"
	"math/rand"
	"sort"
)

// WindowSize is K (§2.3, card AC #6): the situated window holds this many
// items — WindowSize-2 by decayed weight, 2 by serendipity.
const WindowSize = 10

// WindowItem is one memory in the snapshot SelectWindow chooses over: an
// opaque ID (used only for de-duplication), its base Salience (pre-decay
// weight, mind-assigned), and ObservedAt (the world_time it was
// observed — RM-5/RM-6's clock, mind/memory/beliefs.go's style).
type WindowItem struct {
	ID         string
	Salience   float64
	ObservedAt int64
}

// Snapshot is the store snapshot SelectWindow is pure over: Items plus
// DayLength, the world_time-units-per-day divisor the decay arithmetic
// needs. DayLength is mind configuration, not a protocol constant — the
// same posture as mind/memory/instrument.go's dayLength and
// mind/memory/beliefs.go's Config half-life/horizon.
type Snapshot struct {
	Items     []WindowItem
	DayLength int64
}

// SelectWindow picks the K=10 situated window (§2.3): the top K-2 items
// by salience halved once per DayLength of age since ObservedAt
// (world_time arithmetic, RM-5's read-time-decay style applied per day
// rather than per configured half-life), plus 2 serendipity picks drawn
// from the snapshot's older half (split by age) using seed — deterministic
// and reproducible in tests, and stable per villager when the caller
// seeds from persona identity. Degrades gracefully under fewer than K
// items: no duplicates (a serendipity candidate already in the top K-2 is
// excluded), no panic, and serendipity fires only when an unpicked older-
// half candidate remains (card AC #6).
func SelectWindow(snap Snapshot, now, seed int64) []WindowItem {
	if len(snap.Items) == 0 {
		return nil
	}
	dayLength := snap.DayLength
	if dayLength <= 0 {
		dayLength = 1
	}

	type scored struct {
		item   WindowItem
		weight float64
		age    int64
	}
	all := make([]scored, len(snap.Items))
	for i, it := range snap.Items {
		age := now - it.ObservedAt
		if age < 0 {
			age = 0 // future-dated observation: no decay, mirrors beliefs.go's effectiveConfidence
		}
		all[i] = scored{it, it.Salience * math.Pow(0.5, float64(age)/float64(dayLength)), age}
	}

	topN := WindowSize - 2
	if topN > len(all) {
		topN = len(all)
	}
	byWeight := append([]scored(nil), all...)
	sort.SliceStable(byWeight, func(i, j int) bool { return byWeight[i].weight > byWeight[j].weight })

	picked := make(map[string]bool, WindowSize)
	out := make([]WindowItem, 0, WindowSize)
	for _, s := range byWeight[:topN] {
		out = append(out, s.item)
		picked[s.item.ID] = true
	}

	// Older half: the snapshot's own items sorted oldest-first. Serendipity
	// candidates are whichever of that half were not already selected
	// above (no duplicates).
	byAge := append([]scored(nil), all...)
	sort.SliceStable(byAge, func(i, j int) bool { return byAge[i].age > byAge[j].age })
	olderHalf := byAge[:len(byAge)/2]

	candidates := make([]scored, 0, len(olderHalf))
	for _, s := range olderHalf {
		if !picked[s.item.ID] {
			candidates = append(candidates, s)
		}
	}
	if len(candidates) == 0 {
		return out
	}
	rng := rand.New(rand.NewSource(seed))
	rng.Shuffle(len(candidates), func(i, j int) { candidates[i], candidates[j] = candidates[j], candidates[i] })
	need := 2
	if need > len(candidates) {
		need = len(candidates)
	}
	for _, s := range candidates[:need] {
		out = append(out, s.item)
	}
	return out
}
