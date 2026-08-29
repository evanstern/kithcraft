package deliberate

import "testing"

// monotonicItems returns n items whose age increases with index (m1
// newest, age 0; mN oldest, age (n-1)*dayLength) at equal salience, so
// weight ranking exactly matches recency — the fixture every ordering
// assertion below relies on.
func monotonicItems(n int, now, dayLength, salience int64) []WindowItem {
	items := make([]WindowItem, n)
	for i := 0; i < n; i++ {
		items[i] = WindowItem{
			ID:         idFor(i),
			Salience:   float64(salience),
			ObservedAt: now - int64(i)*dayLength,
		}
	}
	return items
}

func idFor(i int) string { return string(rune('a' + i)) }

func containsID(items []WindowItem, id string) bool {
	for _, it := range items {
		if it.ID == id {
			return true
		}
	}
	return false
}

// TestSelectWindow_TopKMinus2ByDecayedWeight proves FR-006/card AC #6's
// main clause: with 12 monotonically-aged items, the 8 youngest (highest
// decayed weight, salience held constant) are exactly the top selection.
func TestSelectWindow_TopKMinus2ByDecayedWeight(t *testing.T) {
	items := monotonicItems(12, 10_000, 100, 10)
	out := SelectWindow(Snapshot{Items: items, DayLength: 100}, 10_000, 1)

	if len(out) != WindowSize {
		t.Fatalf("len(out) = %d, want %d", len(out), WindowSize)
	}
	for i := 0; i < 8; i++ {
		if !containsID(out, idFor(i)) {
			t.Fatalf("expected top-8 item %q (youngest, highest decayed weight) in result, got %+v", idFor(i), out)
		}
	}
}

// TestSelectWindow_SerendipityFromOlderHalf proves the 2 extra picks come
// from the store's older half (items m9..m12, the 4 oldest of 12) and
// never duplicate a top-8 pick.
func TestSelectWindow_SerendipityFromOlderHalf(t *testing.T) {
	items := monotonicItems(12, 10_000, 100, 10)
	out := SelectWindow(Snapshot{Items: items, DayLength: 100}, 10_000, 42)

	seen := map[string]int{}
	for _, it := range out {
		seen[it.ID]++
		if seen[it.ID] > 1 {
			t.Fatalf("item %q selected twice — no duplicates permitted (card AC #6)", it.ID)
		}
	}
	extra := 0
	for i := 8; i < 12; i++ {
		if seen[idFor(i)] == 1 {
			extra++
		}
	}
	if extra != 2 {
		t.Fatalf("got %d serendipity picks from the older half (m9..m12), want exactly 2", extra)
	}
}

// TestSelectWindow_Deterministic proves the serendipity draw is
// reproducible: identical snapshot + seed yields an identical selection.
func TestSelectWindow_Deterministic(t *testing.T) {
	items := monotonicItems(12, 10_000, 100, 10)
	a := SelectWindow(Snapshot{Items: items, DayLength: 100}, 10_000, 7)
	b := SelectWindow(Snapshot{Items: items, DayLength: 100}, 10_000, 7)

	if len(a) != len(b) {
		t.Fatalf("len mismatch: %d vs %d", len(a), len(b))
	}
	for i := range a {
		if a[i].ID != b[i].ID {
			t.Fatalf("same seed produced different selections at index %d: %q vs %q", i, a[i].ID, b[i].ID)
		}
	}
}

// TestSelectWindow_GracefulUnderK is US4 AC #3: fewer than K memories
// selects all of them, no duplicates, no panic, no serendipity (nothing
// is left over to draw from once everything is already picked).
func TestSelectWindow_GracefulUnderK(t *testing.T) {
	items := monotonicItems(3, 10_000, 100, 10)
	out := SelectWindow(Snapshot{Items: items, DayLength: 100}, 10_000, 1)

	if len(out) != 3 {
		t.Fatalf("len(out) = %d, want 3 (all available memories)", len(out))
	}
	for i := 0; i < 3; i++ {
		if !containsID(out, idFor(i)) {
			t.Fatalf("expected item %q among the %d available memories", idFor(i), 3)
		}
	}
}

// TestSelectWindow_PartialSerendipity_NoPanic covers the middle case: 9
// items leaves exactly 1 unpicked candidate in the older half, so only 1
// (not 2) serendipity slot fills — graceful degradation, not a panic.
func TestSelectWindow_PartialSerendipity_NoPanic(t *testing.T) {
	items := monotonicItems(9, 10_000, 100, 10)
	out := SelectWindow(Snapshot{Items: items, DayLength: 100}, 10_000, 1)

	if len(out) != 9 {
		t.Fatalf("len(out) = %d, want 9 (all items: top-8 plus the lone remaining candidate)", len(out))
	}
}

// TestSelectWindow_Empty proves an empty snapshot degrades to nil, not a
// panic.
func TestSelectWindow_Empty(t *testing.T) {
	if out := SelectWindow(Snapshot{}, 10_000, 1); out != nil {
		t.Fatalf("SelectWindow(empty) = %+v, want nil", out)
	}
}

// TestSelectWindow_DecayHalvesPerDayOfAge proves the decay arithmetic
// itself (FR-006: "salience halved per day of age"), not just an
// age-monotonic ordering: a salience-8 item 2 days old outweighs a
// salience-1 item observed now, and ties a salience-8 item 3 days old —
// only true under exact per-day halving.
func TestSelectWindow_DecayHalvesPerDayOfAge(t *testing.T) {
	const now, day = int64(10_000), int64(100)
	items := []WindowItem{
		{ID: "young-low", Salience: 1, ObservedAt: now},        // weight 1
		{ID: "mid-high", Salience: 8, ObservedAt: now - 2*day}, // weight 8*0.25 = 2
		{ID: "old-high", Salience: 8, ObservedAt: now - 3*day}, // weight 8*0.125 = 1
	}
	out := SelectWindow(Snapshot{Items: items, DayLength: day}, now, 1)
	if len(out) == 0 || out[0].ID != "mid-high" {
		t.Fatalf("top pick = %+v, want mid-high (weight 2 under exact per-day halving)", out)
	}
}
