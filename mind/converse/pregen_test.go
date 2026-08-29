package converse

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/anthropics/anthropic-sdk-go/option"

	"kithcraft/mind/llm"
)

// countingSSEServer serves reply to every call after delay (sseServer's
// single-script write logic, reused here with an exposed call counter —
// T005/T006 need to prove a slot serves *without* a new call, and that
// counter is how).
func countingSSEServer(t *testing.T, reply string, delay time.Duration) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "text/event-stream")
		flusher := w.(http.Flusher)
		write := func(typ, data string) {
			w.Write([]byte("event: " + typ + "\ndata: " + data + "\n\n"))
			flusher.Flush()
		}
		write("message_start", `{"type":"message_start","message":{"id":"m","type":"message","role":"assistant","model":"claude-test","content":[],"stop_reason":null,"usage":{"input_tokens":10,"output_tokens":0}}}`)
		time.Sleep(delay)
		delta, _ := json.Marshal(map[string]any{"type": "content_block_delta", "index": 0, "delta": map[string]any{"type": "text_delta", "text": reply}})
		write("content_block_delta", string(delta))
		write("content_block_stop", `{"type":"content_block_stop","index":0}`)
		write("message_delta", `{"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":1}}`)
		write("message_stop", `{"type":"message_stop"}`)
	}))
	t.Cleanup(srv.Close)
	return srv, &calls
}

// TestSlot_ServeAtConvergence is T005's core proof, order 1 of the race:
// the fill completes before convergence asks for it, so Take serves the
// pre-generated text without a new E4 call (card AC #4) — and a second
// Take on the same slot always misses (at-most-one opening spoken).
func TestSlot_ServeAtConvergence(t *testing.T) {
	srv, calls := countingSSEServer(t, "Evening, friend.", 5*time.Millisecond)
	client := llm.New(option.WithBaseURL(srv.URL), option.WithAPIKey("test"))
	speaker := testSpeaker("Tam", client)

	sl := newSlot()
	sl.Fill(context.Background(), speaker)
	<-sl.done // wait for the fill to actually finish — this test proves the "already ready" order

	text, latency, ok := sl.Take(0)
	if !ok {
		t.Fatal("Take: ok = false, want true — fill had already completed")
	}
	if text != "Evening, friend." {
		t.Errorf("Take text = %q, want the pre-generated line", text)
	}
	if latency <= 0 {
		t.Error("Take latency = 0, want the fill's measured first-token latency")
	}
	if calls.Load() != 1 {
		t.Errorf("server saw %d calls, want 1 (Take must not make a new call)", calls.Load())
	}
	if _, _, ok := sl.Take(0); ok {
		t.Error("second Take on the same slot: ok = true, want false (at-most-one opening spoken)")
	}
}

// TestSlot_LiveFallbackDiscardsLateFill is T006, order 2 of the race:
// convergence asks before the fill is ready, so Take misses immediately
// (the caller's cue to fall back to a live stream); when the fill finishes
// afterward, it is discarded — nothing can claim it a second time.
func TestSlot_LiveFallbackDiscardsLateFill(t *testing.T) {
	srv, calls := countingSSEServer(t, "Evening, friend.", 150*time.Millisecond)
	client := llm.New(option.WithBaseURL(srv.URL), option.WithAPIKey("test"))
	speaker := testSpeaker("Tam", client)

	sl := newSlot()
	sl.Fill(context.Background(), speaker)

	text, _, ok := sl.Take(10 * time.Millisecond)
	if ok {
		t.Fatalf("Take: ok = true (text %q), want false — the fill had not finished yet", text)
	}

	<-sl.done // let the late fill actually land
	if _, _, ok := sl.Take(0); ok {
		t.Error("Take after a late fill landed: ok = true, want false — the late fill must be discarded")
	}
	if calls.Load() != 1 {
		t.Errorf("server saw %d calls, want 1 (only the fill — Take never calls the model itself)", calls.Load())
	}
}

// TestSlot_AbortDiscard is T006's abort path: the pair-formation signal
// fired (Fill started) but the meeting never converged, so the daemon
// calls Discard. Even once the in-flight fill lands, Take never serves it.
func TestSlot_AbortDiscard(t *testing.T) {
	srv, _ := countingSSEServer(t, "Evening, friend.", 20*time.Millisecond)
	client := llm.New(option.WithBaseURL(srv.URL), option.WithAPIKey("test"))
	speaker := testSpeaker("Tam", client)

	sl := newSlot()
	sl.Fill(context.Background(), speaker)
	sl.Discard()
	<-sl.done

	if _, _, ok := sl.Take(0); ok {
		t.Error("Take after Discard: ok = true, want false — an aborted meeting's slot must never be spoken")
	}
}

// TestSlot_ConcurrentFillAndTakeRace exercises both race orders under -race
// across many trials with randomized timing, proving the mutex-guarded
// one-shot claim holds regardless of scheduling (T006's "cover the race,
// both orders").
func TestSlot_ConcurrentFillAndTakeRace(t *testing.T) {
	srv, _ := countingSSEServer(t, "hi", 0)
	client := llm.New(option.WithBaseURL(srv.URL), option.WithAPIKey("test"))

	for i := 0; i < 25; i++ {
		sl := newSlot()
		speaker := testSpeaker("Tam", client)
		var wg sync.WaitGroup
		wg.Add(2)
		var okCount int32
		go func() {
			defer wg.Done()
			sl.Fill(context.Background(), speaker)
		}()
		go func() {
			defer wg.Done()
			if _, _, ok := sl.Take(time.Duration(i%3) * time.Millisecond); ok {
				atomic.AddInt32(&okCount, 1)
			}
		}()
		wg.Wait()
		<-sl.done
		if _, _, ok := sl.Take(0); ok {
			atomic.AddInt32(&okCount, 1)
		}
		if okCount > 1 {
			t.Fatalf("trial %d: slot served more than once (okCount=%d)", i, okCount)
		}
	}
}

// TestPool_BeginTakeServesKeyedByPairDay proves the Pool wiring T005 asks
// for: Begin fills a slot keyed by (pairID, day); Take on that key serves
// it and removes it, so a repeat Take on the same key (or an unknown key)
// misses.
func TestPool_BeginTakeServesKeyedByPairDay(t *testing.T) {
	srv, calls := countingSSEServer(t, "Evening, friend.", 5*time.Millisecond)
	client := llm.New(option.WithBaseURL(srv.URL), option.WithAPIKey("test"))
	speaker := testSpeaker("Tam", client)

	p := NewPool()
	key := PairKey{PairID: "Tam+Eda", Day: 3}
	sl := p.Begin(context.Background(), key, speaker)
	<-sl.done

	text, _, ok := p.Take(key, 0)
	if !ok || text != "Evening, friend." {
		t.Fatalf("Pool.Take(key) = %q, %v; want the pre-generated line, true", text, ok)
	}
	if calls.Load() != 1 {
		t.Errorf("server saw %d calls, want 1", calls.Load())
	}
	if _, _, ok := p.Take(key, 0); ok {
		t.Error("second Pool.Take on the same key: ok = true, want false (the slot was removed)")
	}
	if _, _, ok := p.Take(PairKey{PairID: "other", Day: 3}, 0); ok {
		t.Error("Pool.Take on an unbegun key: ok = true, want false")
	}
}

// TestPool_DiscardDropsUnspoken is T006's abort path at the Pool level.
func TestPool_DiscardDropsUnspoken(t *testing.T) {
	srv, _ := countingSSEServer(t, "Evening, friend.", 5*time.Millisecond)
	client := llm.New(option.WithBaseURL(srv.URL), option.WithAPIKey("test"))
	speaker := testSpeaker("Tam", client)

	p := NewPool()
	key := PairKey{PairID: "Tam+Eda", Day: 3}
	sl := p.Begin(context.Background(), key, speaker)
	p.Discard(key)
	<-sl.done

	if _, _, ok := p.Take(key, 0); ok {
		t.Error("Pool.Take after Discard: ok = true, want false")
	}
}
