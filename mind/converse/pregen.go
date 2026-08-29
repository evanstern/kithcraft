// Package converse (this file): the pre-generation slot (specs/017-dusk-
// conversation Phase 2, plan.md design decision 3). V3's dusk pair-
// formation signal (a sighting percept, per docs/design/demo-build-plan.md
// R-7) fires while a pair is still walking toward the gathering place —
// measured live at 1.82-4.96 s lead (TASK-0014, vs a nominal ~10 s), so the
// opening turn's E4 call can start before the scene opens instead of after
// it. A Slot holds that call's outcome; Pool keys slots by (pairID, day) so
// the daemon's signal-handling code (out of scope here — see Pool.Begin's
// doc) has one place to fill, serve, or discard them. At most one opening
// turn is ever spoken from a Slot: Take is a one-shot claim, whichever
// caller reaches it first (a completed fill at convergence, or convergence
// itself timing out and falling back to a live call) — this is what makes
// the race in T006 (fill completing concurrently with convergence) safe in
// both orders.
package converse

import (
	"context"
	"sync"
	"time"
)

// Slot is one pair's opening-turn pre-generation outcome. Zero value is not
// ready to use; construct with newSlot (or via Pool.Begin).
type Slot struct {
	mu      sync.Mutex
	ready   bool
	text    string
	latency time.Duration
	err     error
	spoken  bool // one-shot guard: Take/Discard set this; a second claim always misses
	done    chan struct{}
}

func newSlot() *Slot { return &Slot{done: make(chan struct{})} }

// Fill runs speaker's opening-turn E4 call (an empty transcript — this is
// the first turn) in the background. This is the call the daemon's V3
// signal-handling code makes when a sighting percept promotes to a
// converging pair (spec US2 AC#1); ingesting that percept itself is not
// this phase's job (plan.md: "model the trigger as a call the daemon
// wiring will make").
func (sl *Slot) Fill(ctx context.Context, speaker *Speaker) {
	go func() {
		text, latency, err := speaker.stream(ctx, "")
		sl.mu.Lock()
		sl.text, sl.latency, sl.err, sl.ready = text, latency, err, true
		sl.mu.Unlock()
		close(sl.done) // after Unlock: Take's post-receive Lock() is then guaranteed to see the write
	}()
}

// Take claims the slot, waiting up to wait for an in-flight Fill to finish
// (wait <= 0: check now, don't wait — with V3's lead sometimes under 2s,
// blocking here would spend the very ceiling pre-generation exists to
// protect). ok is false if the slot was never filled, filling failed, the
// wait elapsed first, or the slot was already claimed (by a prior Take or
// Discard) — every false case is the caller's cue to fall back to a live
// stream call (T006). Take is one-shot: it always marks the slot spoken
// before returning, win or miss, so a Fill that completes after a missed
// Take is discarded rather than served late.
func (sl *Slot) Take(wait time.Duration) (text string, latency time.Duration, ok bool) {
	if wait > 0 {
		timer := time.NewTimer(wait)
		defer timer.Stop()
		select {
		case <-sl.done:
		case <-timer.C:
		}
	} else {
		select {
		case <-sl.done:
		default:
		}
	}

	sl.mu.Lock()
	defer sl.mu.Unlock()
	if sl.spoken || !sl.ready || sl.err != nil {
		sl.spoken = true
		return "", 0, false
	}
	sl.spoken = true
	return sl.text, sl.latency, true
}

// Discard drops the slot unspoken: the pair-formation signal fired but the
// meeting aborted before convergence (T006's abort-discard path). A Fill
// still in flight is left to run to completion (there is no ctx to cancel
// here — the caller's ctx does that if it wants to); Discard just ensures
// its result, once ready, is never claimed.
func (sl *Slot) Discard() {
	sl.mu.Lock()
	sl.spoken = true
	sl.mu.Unlock()
}

// PairKey identifies one pre-generation slot: a converging pair, on one
// in-game day (plan.md decision 3). Day is world_time arithmetic (M2
// convention, §6.2 of the routing doc) — computing it is the daemon
// wiring's job, not this package's.
type PairKey struct {
	PairID string
	Day    int64
}

// Pool is the daemon-facing store of pre-generation slots, keyed by
// PairKey — T005's "per-pair slot keyed (pairID, day)". A Pool is safe for
// concurrent use.
type Pool struct {
	mu    sync.Mutex
	slots map[PairKey]*Slot
}

func NewPool() *Pool { return &Pool{slots: make(map[PairKey]*Slot)} }

// Begin starts filling key's slot with speaker's opening turn — the call
// the daemon makes when V3's pair-formation signal arrives for this pair
// (card AC #4's fill side). It overwrites any existing slot for key
// (a re-fired signal for the same pair-day restarts the fill).
func (p *Pool) Begin(ctx context.Context, key PairKey, speaker *Speaker) *Slot {
	sl := newSlot()
	p.mu.Lock()
	p.slots[key] = sl
	p.mu.Unlock()
	sl.Fill(ctx, speaker)
	return sl
}

// take removes and returns key's slot, or nil if none is pending — shared
// by Take and Discard so a pair-day's slot is claimed exactly once.
func (p *Pool) take(key PairKey) *Slot {
	p.mu.Lock()
	defer p.mu.Unlock()
	sl := p.slots[key]
	delete(p.slots, key)
	return sl
}

// Take claims key's slot at convergence (card AC #4's serve side: ready,
// no new call). See Slot.Take for the wait/fallback semantics.
func (p *Pool) Take(key PairKey, wait time.Duration) (text string, latency time.Duration, ok bool) {
	sl := p.take(key)
	if sl == nil {
		return "", 0, false
	}
	return sl.Take(wait)
}

// Discard drops key's slot unspoken (T006's abort-discard path: the signal
// fired but the meeting never converged).
func (p *Pool) Discard(key PairKey) {
	if sl := p.take(key); sl != nil {
		sl.Discard()
	}
}
