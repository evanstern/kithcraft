// Package deliberate (this file): the §5.5 urgency interrupt (spec.md
// FR-005, plan.md design decision 2, card AC #5) — an `urgent` percept
// (body-protocol-v0.md §2.8: top-level urgency == "urgent") cancels
// whichever in-flight deliberation call is registered, fires no model
// call of its own (Interrupt holds no Proposer/Vendor reference — it
// cannot call one), and coalesces every urgent percept that arrives
// before the follow-up actually drains into exactly one enqueued
// follow-up deliberation. Loop.Run already returns ctx.Err() cleanly on
// context cancellation (Phase 1); this file supplies the cancel func and
// the coalescing buffer around it, never touching loop.go.
package deliberate

import (
	"context"
	"sync"
)

// IsUrgent reports whether percept's §2.8 urgency band is "urgent" — the
// only band this interrupt reacts to (`notable`/`background` are the
// admission gate's concern, mind/memory/admission.go, not this one's).
func IsUrgent(percept map[string]any) bool {
	urgency, _ := percept["urgency"].(string)
	return urgency == "urgent"
}

// Interrupt is the mutex'd state machine plan.md's Risks note calls for:
// it settles the cancel/complete race by construction — cancelling an
// already-finished call is a no-op (context.CancelFunc's own contract,
// RT-2), and Enqueue fires at most once per coalescing window regardless
// of whether Urgent or the call's own completion happens first.
type Interrupt struct {
	mu        sync.Mutex
	cancel    context.CancelFunc
	pending   []map[string]any
	queued    bool
	onEnqueue func() // fired at most once per coalescing window
}

// NewInterrupt returns an Interrupt that calls onEnqueue (may be nil)
// the first time a coalescing window opens — the caller's cue to start
// one follow-up Run once convenient, then Drain its context.
func NewInterrupt(onEnqueue func()) *Interrupt {
	return &Interrupt{onEnqueue: onEnqueue}
}

// Register associates the in-flight call's cancel func with i — call it
// right before starting a Loop.Run(ctx, ...) with ctx from
// context.WithCancel, and Register(nil) once that Run returns, so a
// later Urgent never cancels a stale, already-superseded call.
func (i *Interrupt) Register(cancel context.CancelFunc) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.cancel = cancel
}

// Urgent feeds one urgent percept to the interrupt (§5.5, card AC #5):
// cancels the registered in-flight call if any, buffers percept for the
// next Drain, and fires onEnqueue exactly once per coalescing window —
// every urgent percept arriving before the next Drain shares that one
// follow-up's context (US3 AC #3's coalescing scenario). It never calls
// a Proposer or Vendor: no model call fires from this method.
func (i *Interrupt) Urgent(percept map[string]any) {
	i.mu.Lock()
	if i.cancel != nil {
		i.cancel()
	}
	i.pending = append(i.pending, percept)
	fire := !i.queued
	i.queued = true
	i.mu.Unlock()

	if fire && i.onEnqueue != nil {
		i.onEnqueue()
	}
}

// Drain atomically takes every urgent percept buffered since the last
// Drain and closes this coalescing window (a following Urgent opens a
// fresh one). Call it exactly once, right before actually starting the
// enqueued follow-up Run, so any urgent percept that arrived between
// onEnqueue firing and this call still lands in that same follow-up's
// context. An empty return means no urgent is currently pending.
func (i *Interrupt) Drain() []map[string]any {
	i.mu.Lock()
	defer i.mu.Unlock()
	out := i.pending
	i.pending = nil
	i.queued = false
	return out
}
