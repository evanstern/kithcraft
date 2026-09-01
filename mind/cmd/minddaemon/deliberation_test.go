// TASK-0023 phase 1 (T002/T003, card ACs #1/#2): proofs that E2/E3
// composition and the §5.5 interrupt are reachable through the REAL
// daemon binary — the real listener, the real seam wire, the real
// Runtime.HandlePercept/HandleSessionOpen hooks — not only mind/deliberate's
// package tests. No live model call: the "model" is a scripted httptest
// server, the same posture runtime_test.go's genesisServer already takes
// toward E1.
package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/anthropics/anthropic-sdk-go/option"

	"kithcraft/mind/llm"
	"kithcraft/mind/persona"
	"kithcraft/mind/seam"
	"kithcraft/mind/seamtest"
)

var boardCapabilities = map[string]any{
	"percept_types": []any{"text", "self_state", "act_result"},
	"origins":       []any{"read", "felt", "acted"},
	"verbs": []any{
		map[string]any{"verb": "claim", "targets": []any{"none"}},
		map[string]any{"verb": "decline", "targets": []any{"none"}},
	},
}

func boardPostingPercept(id, text string) map[string]any {
	return map[string]any{
		"percept_id": id, "percept_type": "text", "urgency": "notable",
		"provenance": map[string]any{"origin": "read", "source": nil, "observed_at": int64(0), "received_at": int64(0)},
		"place":      nil,
		"content":    map[string]any{"text": text, "attributed_to": "the player"},
	}
}

func actResultForIntent(id, intentID string) map[string]any {
	return map[string]any{
		"percept_id": id, "percept_type": "act_result", "urgency": "notable",
		"provenance": map[string]any{"origin": "acted", "source": nil, "observed_at": int64(1), "received_at": int64(1)},
		"place":      nil,
		"content": map[string]any{
			"intent_id": intentID, "verb": "claim", "outcome": "completed",
			"reason_code": nil, "reason": "because", "detail": "did it",
		},
	}
}

func urgentSelfStatePercept(id string) map[string]any {
	return map[string]any{
		"percept_id": id, "percept_type": "self_state", "urgency": "urgent",
		"provenance": map[string]any{"origin": "felt", "source": nil, "observed_at": int64(0), "received_at": int64(0)},
		"place":      nil,
		"content":    map[string]any{"condition": "threatened", "level": "severe", "trend": "worsening"},
	}
}

// writeIntentResponse writes one Messages-API-shaped response whose sole
// content block is a raw structured-output Intent (llm.ParseIntent's
// input) — the same fake-model shape genesisServer (runtime_test.go)
// takes for E1, applied to E2/E3.
func writeIntentResponse(w http.ResponseWriter, verb, reason string) {
	raw, _ := json.Marshal(map[string]any{"verb": verb, "reason": reason})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"id": "msg_test", "type": "message", "role": "assistant", "model": "claude-test",
		"content":     []map[string]any{{"type": "text", "text": string(raw)}},
		"stop_reason": "end_turn",
		"usage":       map[string]any{"input_tokens": 10, "output_tokens": 10},
	})
}

// simpleIntentServer always answers with the same scripted intent.
func simpleIntentServer(t *testing.T, verb, reason string) (*httptest.Server, *int32) {
	t.Helper()
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		writeIntentResponse(w, verb, reason)
	}))
	t.Cleanup(srv.Close)
	return srv, &hits
}

// gatedIntentServer blocks its FIRST request on release — every later
// request answers immediately. reached closes once the first request has
// actually arrived and is blocked, so a test can wait for it before firing
// whatever is meant to cancel that call.
//
// The handler blocks on release rather than r.Context().Done(): the
// deliberation's own ctx cancellation is proven client-side (Loop.Run
// returning promptly with a cancellation error, mind/deliberate's own
// contract) — whether the local net/http server's request context
// observes a client-side cancel promptly is a platform-timing detail this
// test does not need to depend on. The caller MUST close release (directly
// or via t.Cleanup, registered after this call so it runs first — cleanups
// are LIFO) so the blocked handler's connection does not wedge
// httptest.Server.Close().
func gatedIntentServer(t *testing.T, verb, reason string) (srv *httptest.Server, reached <-chan struct{}, release chan<- struct{}, hits *int32) {
	t.Helper()
	reachedCh := make(chan struct{})
	releaseCh := make(chan struct{})
	var once sync.Once
	var n int32
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := atomic.AddInt32(&n, 1)
		if c == 1 {
			once.Do(func() { close(reachedCh) })
			select {
			case <-releaseCh:
			case <-r.Context().Done():
			}
			return
		}
		writeIntentResponse(w, verb, reason)
	}))
	t.Cleanup(s.Close)
	return s, reachedCh, releaseCh, &n
}

// startDeliberationDaemon wires a Runtime around srv's fake model client,
// sets rt.Personas (TASK-0023 phase 2, T004: HandleSessionOpen binds a
// body's persona from rt.Personas ONCE at attach, so the whole cast map
// must be set before session_open is ever sent — the same "populated once
// at startup, before serving begins" invariant LoadOrGenesisCast's own doc
// already states, now load-bearing for a race detector too, not just
// documentation), and serves it on a real UDS listener. Returns the dialed
// Double and a cleanup-free teardown left to t.Cleanup.
func startDeliberationDaemon(t *testing.T, srv *httptest.Server, body string, personas map[string]persona.Persona) (*Runtime, *seamtest.Double) {
	t.Helper()
	rt, err := NewRuntime(t.TempDir())
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}
	t.Cleanup(func() { rt.Close() })
	rt.Client = llm.New(option.WithBaseURL(srv.URL), option.WithAPIKey("test-key"))
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
	if err := dbl.Send(seamtest.SessionOpen("s-1", body, "second", boardCapabilities, nil)); err != nil {
		t.Fatalf("session_open: %v", err)
	}
	return rt, dbl
}

// TestLiveE3Deliberation_PerceptInIntentOutWithAuthoredReason is card AC
// #1 (spec.md US1): a board text/origin:read percept, through the real
// daemon binary, reaches a live E3 mind/deliberate.Loop composed with a
// bound persona and the K=10 window, and comes back out on the wire as a
// claim intent carrying the model's own authored reason.
func TestLiveE3Deliberation_PerceptInIntentOutWithAuthoredReason(t *testing.T) {
	const reason = "Building shelters is exactly my trade, and that wall has bothered me all week."
	srv, hits := simpleIntentServer(t, "claim", reason)

	_, dbl := startDeliberationDaemon(t, srv, "b-villager", map[string]persona.Persona{
		"b-villager": {Name: "Aldric", Anchor: "steady and exacting", Values: []string{"craft", "care"}},
	})

	if err := dbl.Send(seamtest.Percept("s-1", "b-villager", 1, 100, boardPostingPercept("p-1", "Build a shelter by the north wall. — the player"))); err != nil {
		t.Fatalf("percept: %v", err)
	}

	waitUntil(t, 2*time.Second, func() bool { return len(dbl.Intents()) >= 1 })
	if got := atomic.LoadInt32(hits); got < 1 {
		t.Fatalf("fake model server hit %d times, want at least 1", got)
	}

	intent := dbl.Intents()[0]
	payload, _ := intent["payload"].(map[string]any)
	if payload["verb"] != "claim" {
		t.Fatalf("intent verb = %v, want claim", payload["verb"])
	}
	if payload["reason"] != reason {
		t.Fatalf("intent reason = %q, want the model's own authored reason %q", payload["reason"], reason)
	}
	if payload["intent_id"] == "" || payload["intent_id"] == nil {
		t.Fatal("intent carries no intent_id")
	}

	// Resolve it so the Run completes cleanly (ErrDone on round 2) rather
	// than leaking a goroutine blocked on Deliver for the test's duration.
	intentID, _ := payload["intent_id"].(string)
	if err := dbl.Send(seamtest.Percept("s-1", "b-villager", 2, 101, actResultForIntent("p-2", intentID))); err != nil {
		t.Fatalf("act_result: %v", err)
	}
}

// TestLiveInterrupt_CancelsInFlightAndProducesOneFollowUp is card AC #2
// (spec.md US2, §5.5), proven through the daemon path with -race: an
// urgent percept arriving mid-deliberation cancels the in-flight model
// call (no intent for it ever reaches the wire), fires no call of its own
// (the hit count settles rather than jumping the instant the urgent
// lands), and produces exactly one follow-up call, which comes back out
// as a live intent.
//
// Scope note: mind/deliberate/interrupt_test.go's own
// TestInterrupt_CoalescesMultipleUrgentsIntoOneEnqueue already proves,
// deterministically and package-internally, that several urgents arriving
// before a Drain coalesce into one enqueue — that is Interrupt's own
// invariant, not this daemon's wiring. Racing two urgent percepts over the
// wire against exactly when this body's goroutine happens to Drain is a
// timing bet, not a proof, so this test sends one and proves what T003
// actually adds: that a live percept reaches Interrupt at all.
func TestLiveInterrupt_CancelsInFlightAndProducesOneFollowUp(t *testing.T) {
	srv, reached, release, hits := gatedIntentServer(t, "claim", "the follow-up's own reason")
	// Registered after gatedIntentServer's own t.Cleanup(srv.Close): cleanups
	// run LIFO, so this unblocks the first (cancelled) request's still-open
	// handler before Close() waits on it.
	t.Cleanup(func() { close(release) })
	_, dbl := startDeliberationDaemon(t, srv, "b-villager", map[string]persona.Persona{
		"b-villager": {Name: "Aldric", Anchor: "steady", Values: []string{"craft"}},
	})

	if err := dbl.Send(seamtest.Percept("s-1", "b-villager", 1, 100, boardPostingPercept("p-1", "Build a shelter. — the player"))); err != nil {
		t.Fatalf("percept: %v", err)
	}

	select {
	case <-reached:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for the in-flight deliberation call to reach the fake server")
	}
	if got := len(dbl.Intents()); got != 0 {
		t.Fatalf("no intent should be sent before the (blocked) call ever returns, got %d", got)
	}

	if err := dbl.Send(seamtest.Percept("s-1", "b-villager", 2, 101, urgentSelfStatePercept("p-2"))); err != nil {
		t.Fatalf("urgent percept: %v", err)
	}

	waitUntil(t, 2*time.Second, func() bool { return atomic.LoadInt32(hits) >= 2 })
	// Settle briefly: the urgent itself must never fire a call of its own
	// (§5.5's middle clause) — only the follow-up does.
	time.Sleep(100 * time.Millisecond)
	if got := atomic.LoadInt32(hits); got != 2 {
		t.Fatalf("fake model server hit %d times, want exactly 2 (the cancelled call + one follow-up)", got)
	}

	waitUntil(t, 2*time.Second, func() bool { return len(dbl.Intents()) >= 1 })
	payload, _ := dbl.Intents()[0]["payload"].(map[string]any)
	if payload["verb"] != "claim" {
		t.Fatalf("follow-up intent verb = %v, want claim", payload["verb"])
	}

	// Resolve the follow-up so its Run completes cleanly (ErrDone on round
	// 2) rather than leaking a goroutine blocked on Deliver.
	intentID, _ := payload["intent_id"].(string)
	if err := dbl.Send(seamtest.Percept("s-1", "b-villager", 3, 102, actResultForIntent("p-3", intentID))); err != nil {
		t.Fatalf("act_result: %v", err)
	}
}

// TestDeliberation_NilClient_LogsAndSkips is spec.md's rehearsal-mode edge
// case: an E3 trigger with no ANTHROPIC_API_KEY (rt.Client == nil) never
// calls a model and never panics — the daemon just has nothing to send.
func TestDeliberation_NilClient_LogsAndSkips(t *testing.T) {
	rt, err := NewRuntime(t.TempDir())
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}
	defer rt.Close()
	rt.Personas = map[string]persona.Persona{"b-villager": {Name: "Aldric"}}

	ing := seam.NewIngester()
	ing.OnPercept = rt.HandlePercept
	ing.OnSessionOpen = rt.HandleSessionOpen
	path := filepath.Join(shortSockDir(t), "mind.sock")
	ln, err := listen(path)
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go serve(ln, ing)

	dbl, err := seamtest.DialUnix(path)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer dbl.Close()
	if err := dbl.Send(seamtest.SessionOpen("s-1", "b-villager", "second", boardCapabilities, nil)); err != nil {
		t.Fatalf("session_open: %v", err)
	}
	if err := dbl.Send(seamtest.Percept("s-1", "b-villager", 1, 100, boardPostingPercept("p-1", "Build a shelter. — the player"))); err != nil {
		t.Fatalf("percept: %v", err)
	}

	time.Sleep(200 * time.Millisecond)
	if got := len(dbl.Intents()); got != 0 {
		t.Fatalf("rehearsal mode (nil Client) sent %d intents, want 0", got)
	}
	if err := dbl.ReadErr(); err != nil {
		t.Fatalf("connection ended unexpectedly (a crash?): %v", err)
	}
}

// TestDeliberation_NoBoundPersona_LogsAndSkips is spec.md's stub-cast edge
// case: a body whose token names no loaded CastID (deliberation.go's
// personaFor, bound once at attach per TASK-0023 phase 2 T004) skips
// deliberation with a log line rather than crashing, even with a live
// model client configured.
func TestDeliberation_NoBoundPersona_LogsAndSkips(t *testing.T) {
	srv, hits := simpleIntentServer(t, "claim", "should never be seen")
	_, dbl := startDeliberationDaemon(t, srv, "b-stub-nobody", map[string]persona.Persona{}) // no bound persona for any body

	if err := dbl.Send(seamtest.Percept("s-1", "b-stub-nobody", 1, 100, boardPostingPercept("p-1", "Build a shelter. — the player"))); err != nil {
		t.Fatalf("percept: %v", err)
	}

	time.Sleep(200 * time.Millisecond)
	if got := len(dbl.Intents()); got != 0 {
		t.Fatalf("a body with no bound persona sent %d intents, want 0", got)
	}
	if got := atomic.LoadInt32(hits); got != 0 {
		t.Fatalf("fake model server hit %d times, want 0 (never called for an unbound body)", got)
	}
	if err := dbl.ReadErr(); err != nil {
		t.Fatalf("connection ended unexpectedly (a crash?): %v", err)
	}
}
