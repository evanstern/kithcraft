package consolidate

import (
	"testing"

	"kithcraft/mind/memory"
)

// TestDeathCarryWeight_NoSpikeUntilNextCycle covers now <= deathCycle:
// the death's own cycle, and any cycle before it, carries no spike yet.
func TestDeathCarryWeight_NoSpikeUntilNextCycle(t *testing.T) {
	for _, now := range []int64{5, 6, 7} { // deathCycle == 7
		if got := DeathCarryWeight(7, now); got != NormalPresence {
			t.Errorf("DeathCarryWeight(7, %d) = %v, want NormalPresence %v", now, got, NormalPresence)
		}
	}
}

// TestDeathCarryWeight_SpikeNextCycle is card AC #5's "disproportionately
// high next cycle": the cycle immediately after the death carries the
// full spike.
func TestDeathCarryWeight_SpikeNextCycle(t *testing.T) {
	if got := DeathCarryWeight(0, 1); got != deathCarrySpike {
		t.Fatalf("DeathCarryWeight(0, 1) = %v, want spike %v", got, deathCarrySpike)
	}
}

// TestDeathCarryWeight_LowerTwoCyclesLater is card AC #5's "lower two
// cycles later": the weighting curve strictly decreases past the spike,
// however "two cycles" is counted — from the death's own cycle or from
// the spike's — and never falls back to (or below) the spike itself.
func TestDeathCarryWeight_LowerTwoCyclesLater(t *testing.T) {
	spike := DeathCarryWeight(0, 1)
	twoAfterDeath := DeathCarryWeight(0, 2)
	twoAfterSpike := DeathCarryWeight(0, 3)
	if !(spike > twoAfterDeath && twoAfterDeath >= twoAfterSpike) {
		t.Fatalf("weights not monotonically non-increasing: spike=%v, +2=%v, +3=%v", spike, twoAfterDeath, twoAfterSpike)
	}
	if twoAfterSpike >= spike {
		t.Fatalf("weight two cycles after the spike (%v) not lower than the spike (%v)", twoAfterSpike, spike)
	}
}

// TestDeathCarryWeight_FloorsAtNormalPresence_NeverZero is RM-7: the
// death "remains present — not deleted — well after" the spike fades,
// and the weight never reaches zero or drops below normal presence, no
// matter how many cycles pass.
func TestDeathCarryWeight_FloorsAtNormalPresence_NeverZero(t *testing.T) {
	for _, age := range []int64{10, 100, 1_000_000} {
		got := DeathCarryWeight(0, age)
		if got < NormalPresence {
			t.Errorf("DeathCarryWeight(0, %d) = %v, below the NormalPresence floor %v", age, got, NormalPresence)
		}
		if got == 0 {
			t.Errorf("DeathCarryWeight(0, %d) = 0, RM-7 forbids ever reaching zero", age)
		}
	}
}

// TestDeathCarryWeight_Decreasing pins the whole curve shape (a
// deterministic distribution assertion, not a sampled draw): every
// additional elapsed cycle after the spike weighs no more than the one
// before it.
func TestDeathCarryWeight_Decreasing(t *testing.T) {
	prev := DeathCarryWeight(0, 1)
	for now := int64(2); now <= 20; now++ {
		got := DeathCarryWeight(0, now)
		if got > prev {
			t.Fatalf("DeathCarryWeight(0, %d) = %v > previous %v: curve is not non-increasing", now, got, prev)
		}
		prev = got
	}
}

// TestSelectionWeights_MatchesDeathCarryWeight is the exported selector
// hook M6 will consume: it returns exactly DeathCarryWeight per death,
// keyed by ID, with nothing else in the map.
func TestSelectionWeights_MatchesDeathCarryWeight(t *testing.T) {
	deaths := []WitnessedDeath{
		{ID: memory.ID{WorldTime: 0, Hash: "a"}, Cycle: 3},
		{ID: memory.ID{WorldTime: 0, Hash: "b"}, Cycle: 10},
	}
	got := SelectionWeights(deaths, 4)
	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2", len(got))
	}
	if want := DeathCarryWeight(3, 4); got[deaths[0].ID] != want {
		t.Errorf("weight for a = %v, want %v", got[deaths[0].ID], want)
	}
	if want := DeathCarryWeight(10, 4); got[deaths[1].ID] != want {
		t.Errorf("weight for b = %v, want %v (before its own cycle: NormalPresence)", got[deaths[1].ID], want)
	}
}
