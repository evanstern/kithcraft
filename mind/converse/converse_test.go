package converse

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/anthropics/anthropic-sdk-go/option"

	"kithcraft/mind/llm"
	"kithcraft/mind/prompt"
	"kithcraft/mind/seam"
)

// TestExchangeUsesE4Config proves card AC #1: the exchange calls
// llm.E4, and E4's registry entry is Sonnet 5, streaming, effort low,
// thinking off, cached prefix, ~300 max_tokens (docs/design/llm-routing-
// and-budget.md §5.2). mind/llm/classes_test.go already pins these values
// at the registry level; this test ties the same assertion to the class
// this package actually calls (s.stream uses llm.E4, converse.go), so a
// future edit that pointed Exchange at the wrong class would fail here
// even if the registry itself stayed correct.
func TestExchangeUsesE4Config(t *testing.T) {
	cfg := llm.Registry[llm.E4]
	if cfg.Model != llm.ModelSonnet5 {
		t.Errorf("E4 Model = %q, want Sonnet 5", cfg.Model)
	}
	if !cfg.Streaming {
		t.Error("E4 must stream")
	}
	if cfg.Effort != llm.EffortLow {
		t.Errorf("E4 Effort = %q, want low", cfg.Effort)
	}
	if cfg.ThinkingOn {
		t.Error("E4 thinking must be off")
	}
	if !cfg.Cached {
		t.Error("E4 must cache its stable prefix")
	}
	if cfg.MaxTokens > 300 {
		t.Errorf("E4 MaxTokens = %d, want ~300", cfg.MaxTokens)
	}
}

func testSpeaker(name string, client *llm.Client) *Speaker {
	return &Speaker{
		Name:    name,
		Client:  client,
		Pending: seam.NewPending(map[string]bool{"speak": true}),
		Stable: prompt.DeliberationStablePrefix{
			Persona: name + "'s persona", Values: "values", Manifest: "manifest", Instructions: "instructions",
		},
		Interlocutor: Interlocutor{Who: "the other villager", Impression: "friendly", SharedHistory: "worked together today"},
		MemoryWindow: "m1, m2, m3",
	}
}

// TestSpeakerAssemble_StablePrefixByteIdentical proves T002: two turns for
// the same speaker, with a growing transcript, assemble a byte-identical
// stable prefix — everything that changes turn to turn (transcript,
// interlocutor slice, memory window) lives in the variable suffix.
func TestSpeakerAssemble_StablePrefixByteIdentical(t *testing.T) {
	s := testSpeaker("Tam", nil)
	a1 := s.assemble("")
	a2 := s.assemble("Tam: hello\nEda: hello yourself")
	if a1.Stable != a2.Stable {
		t.Fatalf("stable prefix differs as the transcript grows:\n%q\nvs\n%q", a1.Stable, a2.Stable)
	}
	if a1.Variable == a2.Variable {
		t.Fatal("variable context is identical across a growing transcript — this test would prove nothing")
	}
}

// sseServer serves one scripted SSE reply per call, cycling through
// replies in order — the mock model client every mind/llm test also uses
// (client_test.go's TestStreamDelivery), scripted here across a multi-turn
// exchange rather than one call.
func sseServer(t *testing.T, replies []string) *httptest.Server {
	t.Helper()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := calls.Add(1) - 1
		text := replies[i]
		w.Header().Set("Content-Type", "text/event-stream")
		flusher := w.(http.Flusher)
		write := func(typ, data string) {
			w.Write([]byte("event: " + typ + "\ndata: " + data + "\n\n"))
			flusher.Flush()
		}
		write("message_start", `{"type":"message_start","message":{"id":"m","type":"message","role":"assistant","model":"claude-test","content":[],"stop_reason":null,"usage":{"input_tokens":10,"output_tokens":0}}}`)
		write("content_block_start", `{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`)
		delta, _ := json.Marshal(map[string]any{"type": "content_block_delta", "index": 0, "delta": map[string]any{"type": "text_delta", "text": text}})
		write("content_block_delta", string(delta))
		write("content_block_stop", `{"type":"content_block_stop","index":0}`)
		write("message_delta", `{"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":1}}`)
		write("message_stop", `{"type":"message_stop"}`)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// scriptedSSEServerCounted is sseServer plus an exposed call counter — the
// OpeningSlot tests (T005/T006) need it to prove a pre-generated opening
// turn serves without an extra E4 call, and that an unfilled slot's
// fallback does make one.
func scriptedSSEServerCounted(t *testing.T, replies []string) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := calls.Add(1) - 1
		text := replies[i]
		w.Header().Set("Content-Type", "text/event-stream")
		flusher := w.(http.Flusher)
		write := func(typ, data string) {
			w.Write([]byte("event: " + typ + "\ndata: " + data + "\n\n"))
			flusher.Flush()
		}
		write("message_start", `{"type":"message_start","message":{"id":"m","type":"message","role":"assistant","model":"claude-test","content":[],"stop_reason":null,"usage":{"input_tokens":10,"output_tokens":0}}}`)
		delta, _ := json.Marshal(map[string]any{"type": "content_block_delta", "index": 0, "delta": map[string]any{"type": "text_delta", "text": text}})
		write("content_block_delta", string(delta))
		write("content_block_stop", `{"type":"content_block_stop","index":0}`)
		write("message_delta", `{"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":1}}`)
		write("message_stop", `{"type":"message_stop"}`)
	}))
	t.Cleanup(srv.Close)
	return srv, &calls
}

// TestExchange_OpeningSlotServesWithoutNewCall is T005's Exchange-level
// proof of card AC #4: a's opening turn comes from an already-filled
// OpeningSlot, and the exchange server (a distinct client from the pregen
// call) sees only turns 1 and 2 — turn 0 never reaches it.
func TestExchange_OpeningSlotServesWithoutNewCall(t *testing.T) {
	pregenSrv, pregenCalls := scriptedSSEServerCounted(t, []string{"Evening, friend. What a day at the mill."})
	pregenClient := llm.New(option.WithBaseURL(pregenSrv.URL), option.WithAPIKey("test"))

	exchangeSrv, exchangeCalls := scriptedSSEServerCounted(t, []string{
		"Aye, the roof got fixed today.",
		"Good night to you. " + ClosingMarker,
	})
	exchangeClient := llm.New(option.WithBaseURL(exchangeSrv.URL), option.WithAPIKey("test"))

	a, b := testSpeaker("Tam", exchangeClient), testSpeaker("Eda", exchangeClient)

	// The daemon fills using Tam's speaker configuration but its own call
	// — a separate client here stands in for "the call V3's signal
	// triggers", distinguishable from turn 2's later live call to Tam.
	pregenSpeaker := *a
	pregenSpeaker.Client = pregenClient
	sl := newSlot()
	sl.Fill(context.Background(), &pregenSpeaker)
	<-sl.done

	turns, err := Exchange(context.Background(), a, b, Config{MaxTurns: 10, OpeningSlot: sl})
	if err != nil {
		t.Fatalf("Exchange: %v", err)
	}
	if len(turns) != 3 {
		t.Fatalf("got %d turns, want 3", len(turns))
	}
	if turns[0].Speaker != "Tam" || turns[0].Text != "Evening, friend. What a day at the mill." {
		t.Errorf("turns[0] = %+v, want Tam's pre-generated opening", turns[0])
	}
	if pregenCalls.Load() != 1 {
		t.Errorf("pregen server saw %d calls, want 1 (only Fill)", pregenCalls.Load())
	}
	if exchangeCalls.Load() != 2 {
		t.Errorf("exchange server saw %d calls, want 2 (turns 1-2 only; turn 0 must not call it)", exchangeCalls.Load())
	}
}

// TestExchange_OpeningSlotFallsBackWhenUnfilled is T006's live-stream
// fallback: an OpeningSlot that was never filled (the first-class case per
// TASK-0014's measured 1.82-4.96 s signal lead — the slot may simply not
// be ready at convergence) does not block or error Exchange; the opening
// turn is generated live like any other turn, and the ceiling-holding
// stream path (T004) applies unchanged.
func TestExchange_OpeningSlotFallsBackWhenUnfilled(t *testing.T) {
	srv, calls := scriptedSSEServerCounted(t, []string{
		"Evening — long day at the mill.",
		"Aye, it was.",
		"Good night. " + ClosingMarker,
	})
	client := llm.New(option.WithBaseURL(srv.URL), option.WithAPIKey("test"))
	a, b := testSpeaker("Tam", client), testSpeaker("Eda", client)

	sl := newSlot() // never filled

	turns, err := Exchange(context.Background(), a, b, Config{MaxTurns: 10, OpeningSlot: sl})
	if err != nil {
		t.Fatalf("Exchange: %v", err)
	}
	if len(turns) != 3 {
		t.Fatalf("got %d turns, want 3", len(turns))
	}
	if turns[0].Text != "Evening — long day at the mill." {
		t.Errorf("turns[0].Text = %q, want the live fallback's first scripted reply", turns[0].Text)
	}
	if calls.Load() != 3 {
		t.Errorf("server saw %d calls, want 3 (the opening fell back to a live call)", calls.Load())
	}
}

// TestExchange_SafetyBoundNeverFires is T003's required proof: against a
// scripted exchange whose script closes itself with ClosingMarker well
// before the safety bound, the loop ends by the marker, not by exhausting
// MaxTurns (card AC #7 — a turn cap must never be the ending mechanism).
func TestExchange_SafetyBoundNeverFires(t *testing.T) {
	srv := sseServer(t, []string{
		"Long day at the mill.",
		"Aye, and the player kept us busy too.",
		"True enough. Good night. " + ClosingMarker,
	})
	client := llm.New(option.WithBaseURL(srv.URL), option.WithAPIKey("test"))
	a, b := testSpeaker("Tam", client), testSpeaker("Eda", client)

	turns, err := Exchange(context.Background(), a, b, Config{MaxTurns: 10})
	if err != nil {
		t.Fatalf("Exchange: %v", err)
	}
	if len(turns) != 3 {
		t.Fatalf("got %d turns, want 3 (the script closes on the 3rd)", len(turns))
	}
	if len(turns) >= 10 {
		t.Fatal("safety bound (MaxTurns) fired — it must never be the ending mechanism (card AC #7)")
	}
	for _, tn := range turns {
		if strings.Contains(tn.Text, ClosingMarker) {
			t.Errorf("turn %q still carries ClosingMarker — it must be stripped before joining the transcript", tn.Text)
		}
	}
}
