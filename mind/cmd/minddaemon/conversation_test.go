// TASK-0023 phase 2 (T004-T007, card AC #3/#4/#5): proofs that the dusk
// exchange, the ambient pool, and FirstTokenLatency reporting are
// reachable through the REAL daemon binary — the real listener, the real
// seam wire, the real Runtime.HandlePercept/HandleSessionOpen hooks — not
// only mind/converse's own package tests. No live model call: scripted
// httptest servers, the same posture deliberation_test.go and
// runtime_test.go already take.
package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/anthropics/anthropic-sdk-go/option"

	"kithcraft/mind/consolidate"
	"kithcraft/mind/converse"
	"kithcraft/mind/llm"
	"kithcraft/mind/persona"
	"kithcraft/mind/seam"
	"kithcraft/mind/seamtest"
)

// pairingSignalPercept is the exact shape mod/.../PairingSignal.java
// composes via Sightings.sightingContent (see conversation.go's package
// doc): a sighting of a k:person doing "walking to the gathering place",
// descriptor carrying the OTHER member's cast name, body carrying their
// token.
func pairingSignalPercept(id, otherBody, otherName string) map[string]any {
	return map[string]any{
		"percept_id": id, "percept_type": "sighting", "urgency": "background",
		"provenance": map[string]any{"origin": "saw", "source": nil, "observed_at": int64(1), "received_at": int64(1)},
		"place":      nil,
		"content": map[string]any{
			"thing":    map[string]any{"thing_id": nil, "kind": "k:person", "roles": []any{}, "descriptor": otherName, "body": otherBody, "count": int64(1)},
			"distance": "near", "doing": pairSignalDoing,
		},
	}
}

// inertActResultPercept crosses no trigger at all (not a completed act, not
// text, not a sighting, not urgent) — used purely to advance world_time
// past a day boundary without incidentally firing E2/pairing/ambient.
func inertActResultPercept(id string) map[string]any {
	return map[string]any{
		"percept_id": id, "percept_type": "act_result", "urgency": "background",
		"provenance": map[string]any{"origin": "acted", "source": nil, "observed_at": int64(1), "received_at": int64(1)},
		"place":      nil,
		"content": map[string]any{
			"intent_id": "i-none", "verb": "wait", "outcome": "in_progress",
			"reason_code": nil, "reason": nil, "detail": nil,
		},
	}
}

// personSightingPercept is an ordinary (non-pairing) sighting of a person —
// T006's ambient trigger shape. doing == "" is the generic/ambient case;
// any other doing is IsTargeted's escalate case.
func personSightingPercept(id, doing string) map[string]any {
	content := map[string]any{
		"thing":    map[string]any{"thing_id": nil, "kind": "k:person", "roles": []any{}, "descriptor": "a passerby", "body": "b-passerby", "count": int64(1)},
		"distance": "near",
	}
	if doing != "" {
		content["doing"] = doing
	}
	return map[string]any{
		"percept_id": id, "percept_type": "sighting", "urgency": "background",
		"provenance": map[string]any{"origin": "saw", "source": nil, "observed_at": int64(1), "received_at": int64(1)},
		"place":      nil, "content": content,
	}
}

// sseTextServer always answers with reply after delay (mind/converse's own
// countingSSEServer precedent) — used for E4 dusk-exchange calls.
func sseTextServer(t *testing.T, reply string, delay time.Duration) (*httptest.Server, *atomic.Int32) {
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

// textMessageServer answers each call in turn with replies[i%len(replies)]
// as a plain Messages-API response — used for E5 (mind/converse.AmbientPool
// uses Client.Send, not Stream).
func textMessageServer(t *testing.T, replies ...string) *httptest.Server {
	t.Helper()
	var i int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&i, 1) - 1
		text := replies[int(n)%len(replies)]
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": "msg_test", "type": "message", "role": "assistant", "model": "claude-test",
			"content":     []map[string]any{{"type": "text", "text": text}},
			"stop_reason": "end_turn",
			"usage":       map[string]any{"input_tokens": 10, "output_tokens": 10},
		})
	}))
	t.Cleanup(srv.Close)
	return srv
}

// startConversationDaemon wires a Runtime (rt.Personas set before serving,
// per startDeliberationDaemon's own T004 fix) around client, serves it on a
// real UDS listener, and returns the dialed Double.
func startConversationDaemon(t *testing.T, client *llm.Client, personas map[string]persona.Persona) (*Runtime, *seamtest.Double) {
	t.Helper()
	rt, err := NewRuntime(t.TempDir())
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}
	t.Cleanup(func() { rt.Close() })
	rt.Client = client
	rt.Personas = personas

	ing := seam.NewIngester()
	ing.OnPercept = rt.HandlePercept
	ing.OnSessionOpen = rt.HandleSessionOpen

	path := filepath.Join(shortSockDir(t), "mind.sock")
	ln, err := listen(path)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { ln.Close() })
	go serve(ln, ing)

	dbl, err := seamtest.DialUnix(path)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { dbl.Close() })
	return rt, dbl
}

// TestConsolidationPrefixFor_CarriesPersonaText is T004's E6 half: the
// bound persona's name and anchor render into the stable prefix RunNight
// sends, closing I1's empty-ConsolidationStablePrefix ponytail.
func TestConsolidationPrefixFor_CarriesPersonaText(t *testing.T) {
	prefix := consolidationPrefixFor(persona.Persona{Name: "Aldric", Anchor: "steady and exacting"})
	rendered := prefix.Render()
	if !strings.Contains(rendered, "Aldric") {
		t.Errorf("consolidation prefix %q missing persona name", rendered)
	}
	if !strings.Contains(rendered, "steady and exacting") {
		t.Errorf("consolidation prefix %q missing firewall anchor", rendered)
	}
}

// TestDuskExchange_PairSignalConvergesAndRecordsLatency is card AC #3
// (spec.md US3): two live sessions' pairing-signal sightings converge into
// a live E4 exchange over their own sessions (T005), each speaker's Stable
// prefix is built from ITS bound persona (T004 — exchangeSpeaker would
// fail closed otherwise, per the assertions below), and the spoken turn's
// FirstTokenLatency lands in the session report (T007).
func TestDuskExchange_PairSignalConvergesAndRecordsLatency(t *testing.T) {
	const reply = "Evening, friend. " + converse.ClosingMarker
	srv, _ := sseTextServer(t, reply, 5*time.Millisecond)
	client := llm.New(option.WithBaseURL(srv.URL), option.WithAPIKey("test-key"))

	rt, dbl := startConversationDaemon(t, client, map[string]persona.Persona{
		"Aldric": {CastID: "Aldric", Name: "Aldric", Anchor: "steady", Values: []string{"craft"}},
		"Petra":  {CastID: "Petra", Name: "Petra", Anchor: "practical", Values: []string{"harvest"}},
	})

	if err := dbl.Send(seamtest.SessionOpen("s-1", "Aldric", "second", testCapabilities, nil)); err != nil {
		t.Fatalf("session_open Aldric: %v", err)
	}
	if err := dbl.Send(seamtest.SessionOpen("s-1", "Petra", "second", testCapabilities, nil)); err != nil {
		t.Fatalf("session_open Petra: %v", err)
	}
	if err := dbl.Send(seamtest.Percept("s-1", "Aldric", 1, 100, pairingSignalPercept("p-1", "Petra", "Petra"))); err != nil {
		t.Fatalf("Aldric's pairing signal: %v", err)
	}
	if err := dbl.Send(seamtest.Percept("s-1", "Petra", 1, 100, pairingSignalPercept("p-2", "Aldric", "Aldric"))); err != nil {
		t.Fatalf("Petra's pairing signal (convergence): %v", err)
	}

	waitUntil(t, 2*time.Second, func() bool { return len(dbl.Intents()) >= 1 })
	intent := dbl.Intents()[0]
	if intent["body"] != "Aldric" {
		t.Fatalf("spoke on body %v, want Aldric (the designated opener — the first side to signal)", intent["body"])
	}
	payload, _ := intent["payload"].(map[string]any)
	if payload["verb"] != "speak" {
		t.Fatalf("intent verb = %v, want speak", payload["verb"])
	}
	target, _ := payload["target"].(map[string]any)
	if got := target["text"]; got != "Evening, friend." {
		t.Fatalf("spoken text = %q, want the reply with ClosingMarker stripped", got)
	}

	// The exchange goroutine records FirstTokenLatency after Exchange
	// returns, a beat after the intent itself is written.
	waitUntil(t, 2*time.Second, func() bool {
		rt.mu.Lock()
		defer rt.mu.Unlock()
		return len(rt.bodies["Aldric"].turnLatencies) >= 1
	})
	report := rt.reportText()
	if !strings.Contains(report, "Aldric: ") || strings.Contains(report, "Aldric: (no dusk exchange turns)") {
		t.Fatalf("session report missing Aldric's FirstTokenLatency:\n%s", report)
	}
}

// TestDuskExchange_UnconfirmedPairAbortsAfterTimeout is spec.md's abort-
// discard edge case (T005): a pairing signal with no reciprocal confirm
// (the other side never signals) discards its pregen Slot after
// pairConvergeTimeout rather than leaking a pending pair or ever speaking.
func TestDuskExchange_UnconfirmedPairAbortsAfterTimeout(t *testing.T) {
	old := pairConvergeTimeout
	pairConvergeTimeout = 50 * time.Millisecond
	defer func() { pairConvergeTimeout = old }()

	srv, hits := sseTextServer(t, "Evening.", 0)
	client := llm.New(option.WithBaseURL(srv.URL), option.WithAPIKey("test-key"))
	rt, dbl := startConversationDaemon(t, client, map[string]persona.Persona{
		"Aldric": {CastID: "Aldric", Name: "Aldric", Anchor: "steady"},
	})

	if err := dbl.Send(seamtest.SessionOpen("s-1", "Aldric", "second", testCapabilities, nil)); err != nil {
		t.Fatalf("session_open: %v", err)
	}
	if err := dbl.Send(seamtest.Percept("s-1", "Aldric", 1, 100, pairingSignalPercept("p-1", "Petra", "Petra"))); err != nil {
		t.Fatalf("pairing signal: %v", err)
	}

	waitUntil(t, 2*time.Second, func() bool { return hits.Load() >= 1 }) // the fill started
	waitUntil(t, 2*time.Second, func() bool {
		rt.mu.Lock()
		defer rt.mu.Unlock()
		return len(rt.pairs) == 0
	})
	if got := len(dbl.Intents()); got != 0 {
		t.Fatalf("an unconfirmed pair sent %d intents, want 0 (never spoken)", got)
	}
}

// TestAmbientTrigger_ServeAndEscalate is card AC #4 (spec.md US4, T006):
// a day-rollover refills the AmbientPool (one batched E5 call), a generic
// person sighting serves a pool line as a speak intent, and a sighting
// carrying specific `doing` text escalates to its own live call instead
// (M6's IsTargeted/Escalate).
func TestAmbientTrigger_ServeAndEscalate(t *testing.T) {
	srv := textMessageServer(t,
		"Fine morning, isn't it?\nMorning to you.",
		"Oh, that broken cart? I saw it this morning.")
	client := llm.New(option.WithBaseURL(srv.URL), option.WithAPIKey("test-key"))
	_, dbl := startConversationDaemon(t, client, map[string]persona.Persona{
		"Aldric": {CastID: "Aldric", Name: "Aldric", Anchor: "steady"},
	})

	if err := dbl.Send(seamtest.SessionOpen("s-1", "Aldric", "second", testCapabilities, nil)); err != nil {
		t.Fatalf("session_open: %v", err)
	}
	// Cross a day boundary (consolidate.CycleTicks) to trigger the refill —
	// inert percepts (not sightings) so only the world_time crossing itself
	// is exercised here, not an incidental ambient/pairing trigger.
	if err := dbl.Send(seamtest.Percept("s-1", "Aldric", 1, 100, inertActResultPercept("p-1"))); err != nil {
		t.Fatalf("baseline percept: %v", err)
	}
	if err := dbl.Send(seamtest.Percept("s-1", "Aldric", 2, consolidate.CycleTicks+1, inertActResultPercept("p-2"))); err != nil {
		t.Fatalf("day-crossing percept: %v", err)
	}
	time.Sleep(200 * time.Millisecond) // let the refill's HTTP round trip finish

	if err := dbl.Send(seamtest.Percept("s-1", "Aldric", 3, consolidate.CycleTicks+2, personSightingPercept("p-3", ""))); err != nil {
		t.Fatalf("ambient (generic) percept: %v", err)
	}
	waitUntil(t, 2*time.Second, func() bool { return len(dbl.Intents()) >= 1 })

	if err := dbl.Send(seamtest.Percept("s-1", "Aldric", 4, consolidate.CycleTicks+3, personSightingPercept("p-4", "asking about the broken cart"))); err != nil {
		t.Fatalf("ambient (targeted) percept: %v", err)
	}
	waitUntil(t, 2*time.Second, func() bool { return len(dbl.Intents()) >= 2 })

	intents := dbl.Intents()
	first, _ := intents[0]["payload"].(map[string]any)
	firstTarget, _ := first["target"].(map[string]any)
	if got := firstTarget["text"]; got != "Fine morning, isn't it?" {
		t.Errorf("first (served) line = %q, want the pool's first refilled line", got)
	}
	second, _ := intents[1]["payload"].(map[string]any)
	secondTarget, _ := second["target"].(map[string]any)
	if got := secondTarget["text"]; got != "Oh, that broken cart? I saw it this morning." {
		t.Errorf("second (escalated) line = %q, want the scripted escalate reply", got)
	}
}
