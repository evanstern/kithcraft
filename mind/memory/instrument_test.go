package memory

import "testing"

// TestInstrument_BucketsByVillagerDay is card AC #6: admitted events
// bucket by world_time / dayLength, never by wall clock, and the report
// reflects exactly what was recorded.
func TestInstrument_BucketsByVillagerDay(t *testing.T) {
	in := NewInstrument(1000)
	for _, wt := range []int64{0, 500, 999, 1000, 1500, 2999} {
		in.Record(wt)
	}
	got := in.Report()
	want := map[int64]int{0: 3, 1: 2, 2: 1}
	if len(got) != len(want) {
		t.Fatalf("Report() = %v, want %v", got, want)
	}
	for day, count := range want {
		if got[day] != count {
			t.Errorf("day %d: got %d, want %d", day, got[day], count)
		}
	}
}
