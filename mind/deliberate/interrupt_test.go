package deliberate

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"kithcraft/mind/llm"
)

// TestIsUrgent proves the §2.8 urgency-band check this file's Urgent is
// meant to be fed from.
func TestIsUrgent(t *testing.T) {
	if !IsUrgent(map[string]any{"urgency": "urgent"}) {
		t.Fatal("urgency:urgent must report IsUrgent true")
	}
	for _, band := range []string{"notable", "background", "", "URGENT"} {
		if IsUrgent(map[string]any{"urgency": band}) {
			t.Fatalf("urgency:%q must not report IsUrgent true", band)
		}
	}
}

// TestInterrupt_CancelsInFlightCall_NoOwnModelCall is card AC #5's first
// two clauses, race order A ("cancel wins"): an urgent percept arriving
// while a call is genuinely in-flight cancels it via ctx (RT-2), and
// Urgent itself never reaches a Vendor — Interrupt holds no Vendor/
// Proposer reference, so no model call can fire from it.
func TestInterrupt_CancelsInFlightCall_NoOwnModelCall(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	it := NewInterrupt(nil)
	it.Register(cancel)

	vendor := VendorFunc(func(map[string]any) error {
		t.Fatal("SendIntent must never be reached: propose blocks on ctx and is cancelled before returning an intent")
		return nil
	})
	l := New(Config{Verbs: map[string]bool{"wait": true}, Vendor: vendor, Class: llm.E2})

	inFlight := make(chan struct{})
	propose := func(ctx context.Context) (string, error) {
		close(inFlight)
		<-ctx.Done()
		return "", ctx.Err()
	}

	runErr := make(chan error, 1)
	go func() {
		_, err := l.Run(ctx, propose)
		runErr <- err
	}()

	<-inFlight // the call is now genuinely in-flight
	it.Urgent(map[string]any{"urgency": "urgent", "percept_type": "sound"})

	if err := <-runErr; !errors.Is(err, context.Canceled) {
		t.Fatalf("Run err = %v, want wrapping context.Canceled", err)
	}
	if drained := it.Drain(); len(drained) != 1 {
		t.Fatalf("Drain() = %v, want exactly the one urgent percept buffered", drained)
	}
}

// TestInterrupt_CompletionWinsRace_StillCoalescesFollowup is card AC #5's
// race order B ("complete wins"): the in-flight round already resolved
// normally (via Deliver, matching Loop's own request/fact split) before
// Urgent is ever called. Cancelling an already-resolved context is a
// no-op per context.CancelFunc's own contract, so Run's normal result is
// untouched — but the urgent percept still opens (and is present in) the
// one coalesced follow-up.
func TestInterrupt_CompletionWinsRace_StillCoalescesFollowup(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	it := NewInterrupt(nil)
	it.Register(cancel)

	sent := make(chan map[string]any, 1)
	l := New(Config{
		Verbs:  map[string]bool{"wait": true},
		Vendor: VendorFunc(func(payload map[string]any) error { sent <- payload; return nil }),
		Class:  llm.E2,
	})

	type outcome struct {
		res Result
		err error
	}
	runDone := make(chan outcome, 1)
	go func() {
		res, err := l.Run(ctx, queueProposer(`{"verb":"wait","reason":"still deciding"}`))
		runDone <- outcome{res, err}
	}()

	payload := <-sent
	l.Deliver(map[string]any{"content": map[string]any{"intent_id": payload["intent_id"], "outcome": "completed"}})

	got := <-runDone // Run's own natural end — Done=true from queueProposer's ErrDone
	if got.err != nil || !got.res.Done {
		t.Fatalf("Run result = %+v err=%v, want a normal completion unaffected by any later cancel", got.res, got.err)
	}

	// The round already finished; Urgent's cancel() call is now a no-op —
	// exercised here so the race's other order is actually covered, not
	// just asserted by comment.
	it.Urgent(map[string]any{"urgency": "urgent", "percept_type": "sound"})

	if drained := it.Drain(); len(drained) != 1 {
		t.Fatalf("Drain() = %v, want exactly the one urgent percept that arrived after completion", drained)
	}
}

// TestInterrupt_CoalescesMultipleUrgentsIntoOneEnqueue is US3 AC #3: N
// urgent percepts arriving before the follow-up drains still fire
// onEnqueue exactly once, and Drain hands back every one of them.
func TestInterrupt_CoalescesMultipleUrgentsIntoOneEnqueue(t *testing.T) {
	var enqueued int32
	it := NewInterrupt(func() { atomic.AddInt32(&enqueued, 1) })

	const n = 5
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			it.Urgent(map[string]any{"urgency": "urgent", "percept_id": idFor(i)})
		}(i)
	}
	wg.Wait()

	if got := atomic.LoadInt32(&enqueued); got != 1 {
		t.Fatalf("onEnqueue fired %d times, want exactly 1 (coalesced, card AC #5)", got)
	}
	if drained := it.Drain(); len(drained) != n {
		t.Fatalf("Drain() returned %d percepts, want all %d coalesced urgents", len(drained), n)
	}
}

// TestInterrupt_DrainOpensNewCoalescingWindow proves Drain resets state:
// a fresh Urgent after Drain fires onEnqueue again rather than staying
// silent forever after the first window.
func TestInterrupt_DrainOpensNewCoalescingWindow(t *testing.T) {
	var enqueued int32
	it := NewInterrupt(func() { atomic.AddInt32(&enqueued, 1) })

	it.Urgent(map[string]any{"urgency": "urgent"})
	it.Drain()
	it.Urgent(map[string]any{"urgency": "urgent"})

	if got := atomic.LoadInt32(&enqueued); got != 2 {
		t.Fatalf("onEnqueue fired %d times across two windows, want 2", got)
	}
}

// TestInterrupt_DrainWithNothingPending proves an idle Drain is a
// well-defined no-op, not a panic.
func TestInterrupt_DrainWithNothingPending(t *testing.T) {
	it := NewInterrupt(nil)
	if drained := it.Drain(); len(drained) != 0 {
		t.Fatalf("Drain() on an idle Interrupt = %v, want empty", drained)
	}
}
