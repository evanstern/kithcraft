// Package memory (this file): the E6-input-tokens instrument (routing
// §6.3, §4.4/A-8) — admitted episodic-buffer size per villager-day (card
// AC #6). It counts, it does not decide: what "day" means is world_time
// arithmetic on the session's declared time unit (§2.2), never wall
// clock (harness T-b). specs/010-event-sourced-memory Phase 3 (T008).
package memory

import "sync"

// Instrument counts admitted memory events per villager-day. One per
// villager, matching Gate's granularity.
type Instrument struct {
	mu        sync.Mutex
	dayLength int64
	counts    map[int64]int
}

// NewInstrument returns an empty instrument. dayLength is world_time
// units per villager-day, in the session's declared time_unit — mind
// configuration (§2.2: horizons and cadences are never protocol
// constants), same posture as Config's half-life/horizon.
func NewInstrument(dayLength int64) *Instrument {
	return &Instrument{dayLength: dayLength, counts: make(map[int64]int)}
}

// Record counts one admitted event at worldTime toward its villager-day.
// Call it once per Log.Append that Gate.Decide admitted.
func (in *Instrument) Record(worldTime int64) {
	in.mu.Lock()
	defer in.mu.Unlock()
	in.counts[dayIndex(worldTime, in.dayLength)]++
}

// Report returns admitted buffer size per villager-day (card AC #6),
// keyed by day index — call at session end.
func (in *Instrument) Report() map[int64]int {
	in.mu.Lock()
	defer in.mu.Unlock()
	out := make(map[int64]int, len(in.counts))
	for k, v := range in.counts {
		out[k] = v
	}
	return out
}

func dayIndex(worldTime, dayLength int64) int64 {
	if dayLength <= 0 {
		return 0
	}
	return worldTime / dayLength
}
