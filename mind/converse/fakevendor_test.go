// fakevendor_test.go (this file): T004 — a scripted dusk exchange run
// against the real seam wire and mind/fakevendor, proving the speak
// intents converse.go composes actually cross the seam (not just Pending's
// in-memory bookkeeping) and that first-token latency is measured and
// under the test ceiling (card ACs #2, #3). External package, mirroring
// mind/fakevendor's own external test packages: what's proven here is
// proven from a caller's vantage, over the real wire.
package converse_test

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/anthropics/anthropic-sdk-go/option"

	"kithcraft/mind/converse"
	"kithcraft/mind/fakevendor"
	"kithcraft/mind/llm"
	"kithcraft/mind/prompt"
	"kithcraft/mind/seam"
)

func manifest() map[string]any {
	caps := map[string]any{
		"percept_types": []any{"speech"},
		"origins":       []any{"acted"},
		"verbs":         []any{map[string]any{"verb": "speak", "targets": []any{"body", "person", "none"}}},
	}
	return fakevendor.Manifest("second", caps, nil)
}

// pipeConn wires one Speaker's Out to its own FakeVendor over a net.Pipe,
// draining the vendor's own outbound traffic (session_open, intent_ack) in
// the background — converse.Speaker does not read acks in Phase 1
// (converse.go's ponytail), so something must still drain the pipe or the
// vendor's writes block.
func pipeConn(t *testing.T, session, body string) (*fakevendor.FakeVendor, seam.Conn) {
	t.Helper()
	c1, c2 := net.Pipe()
	t.Cleanup(func() { _ = c1.Close(); _ = c2.Close() })
	mindConn := seam.NewWireConn(c2)
	go func() {
		for {
			if _, err := mindConn.ReadMessage(); err != nil {
				return
			}
		}
	}()
	v := fakevendor.New(seam.NewWireConn(c1), session, body, manifest())
	if err := v.Open(); err != nil {
		t.Fatalf("fakevendor Open: %v", err)
	}
	return v, mindConn
}

func fakevendorSpeaker(t *testing.T, name string, client *llm.Client) (*converse.Speaker, *fakevendor.FakeVendor) {
	t.Helper()
	v, out := pipeConn(t, "s-dusk", "b-"+name)
	return &converse.Speaker{
		Name: name, Client: client, Pending: seam.NewPending(map[string]bool{"speak": true}),
		Out: out, Session: "s-dusk", Body: "b-" + name,
		Stable: prompt.DeliberationStablePrefix{
			Persona: name + "'s persona", Values: "values", Manifest: "manifest", Instructions: "instructions",
		},
		Interlocutor: converse.Interlocutor{Who: "the other villager", Impression: "friendly", SharedHistory: "worked together today"},
		MemoryWindow: "m1, m2, m3",
	}, v
}

// scriptedSSEServer serves one scripted reply per call in order, each
// delayed slightly before its first delta — the mock-injected latency
// plan.md design decision 6 asks for, proving the instrumentation works
// rather than merely proving a mock server is fast.
func scriptedSSEServer(t *testing.T, replies []string, delay time.Duration) *httptest.Server {
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
		time.Sleep(delay)
		delta, _ := json.Marshal(map[string]any{"type": "content_block_delta", "index": 0, "delta": map[string]any{"type": "text_delta", "text": text}})
		write("content_block_delta", string(delta))
		write("content_block_stop", `{"type":"content_block_stop","index":0}`)
		write("message_delta", `{"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":1}}`)
		write("message_stop", `{"type":"message_stop"}`)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// waitForActs polls FakeVendor.Acts() until it reaches n entries or a
// short timeout elapses — recordAndAck (fakevendor.go) records a received
// intent in its own background goroutine, a beat after the mind side's
// WriteMessage returns, so there is no other signal to synchronize on.
func waitForActs(t *testing.T, v *fakevendor.FakeVendor, n int) []map[string]any {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		if acts := v.Acts(); len(acts) >= n || time.Now().After(deadline) {
			return acts
		}
		time.Sleep(time.Millisecond)
	}
}

// TestDuskExchange_AgainstFakeVendor is T004: a multi-turn dusk exchange
// about the day, the work, and the player, run against FakeVendor and a
// mocked model client. Proves: turns alternate and mention the scripted
// topics (card AC #2), first-token latency is measured with a nonzero
// floor from the injected delay and stays under the 3s ceiling (card AC
// #3), the exchange ends on ClosingMarker rather than the safety bound
// (card AC #7), and both speakers' composed intents actually crossed the
// seam wire into FakeVendor.Acts().
func TestDuskExchange_AgainstFakeVendor(t *testing.T) {
	replies := []string{
		"What a day — the mill roof finally got fixed.",
		"Aye, good work. The player even lent a hand carrying beams.",
		"That they did. Well, I'm for bed. " + converse.ClosingMarker,
	}
	srv := scriptedSSEServer(t, replies, 20*time.Millisecond)
	client := llm.New(option.WithBaseURL(srv.URL), option.WithAPIKey("test"))

	a, vendorA := fakevendorSpeaker(t, "Tam", client)
	b, vendorB := fakevendorSpeaker(t, "Eda", client)

	turns, err := converse.Exchange(context.Background(), a, b, converse.Config{WorldTime: 918233, MaxTurns: 10})
	if err != nil {
		t.Fatalf("Exchange: %v", err)
	}

	if len(turns) != 3 {
		t.Fatalf("got %d turns, want 3", len(turns))
	}
	if len(turns) >= 10 {
		t.Fatal("safety bound fired instead of ClosingMarker (card AC #7)")
	}
	if turns[0].Speaker != "Tam" || turns[1].Speaker != "Eda" || turns[2].Speaker != "Tam" {
		t.Fatalf("turns did not alternate: %+v", turns)
	}

	full := turns[0].Text + " " + turns[1].Text + " " + turns[2].Text
	for _, want := range []string{"day", "work", "player"} {
		if !strings.Contains(strings.ToLower(full), want) {
			t.Errorf("transcript %q does not mention %q (card AC #2: day/work/player)", full, want)
		}
	}

	const ceiling = 3 * time.Second
	for _, tn := range turns {
		if tn.FirstTokenLatency <= 0 {
			t.Errorf("turn %q: FirstTokenLatency = %v, want > 0 (the mock delay proves the instrumentation fired)", tn.Speaker, tn.FirstTokenLatency)
		}
		if tn.FirstTokenLatency >= ceiling {
			t.Errorf("turn %q: FirstTokenLatency = %v, want < %v (§5.2 ceiling)", tn.Speaker, tn.FirstTokenLatency, ceiling)
		}
	}

	// Both speakers' speak intents crossed the real seam wire. Recording
	// happens in FakeVendor's own background goroutine (recordAndAck), a
	// beat after WriteMessage returns on the mind side, so wait for the
	// expected count rather than asserting immediately.
	if acts := waitForActs(t, vendorA, 2); len(acts) != 2 { // Tam speaks turns 0 and 2
		t.Fatalf("Tam's vendor recorded %d intents, want 2: %#v", len(acts), acts)
	}
	if acts := waitForActs(t, vendorB, 1); len(acts) != 1 { // Eda speaks turn 1
		t.Fatalf("Eda's vendor recorded %d intents, want 1: %#v", len(acts), acts)
	}
	for _, v := range []*fakevendor.FakeVendor{vendorA, vendorB} {
		for _, act := range v.Acts() {
			if act["verb"] != "speak" {
				t.Errorf("act verb = %v, want speak", act["verb"])
			}
		}
	}
}
