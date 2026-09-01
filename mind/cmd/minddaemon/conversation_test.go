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

// sseTextServer answers each request in turn with replies[i % len(replies)]
// after delay (mind/converse's own countingSSEServer precedent, cycling the
// same way textMessageServer already does for E5) — used for E4 dusk-
// exchange calls. A single reply is the common case (every call gets the
// same text); T008's multi-turn proof passes more than one so a later call
// can carry converse.ClosingMarker while an earlier one doesn't.
func sseTextServer(t *testing.T, delay time.Duration, replies ...string) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1) - 1
		reply := replies[int(n)%len(replies)]
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

// TestDuskExchange_PairSignalConvergesAndRecordsLatency is card AC #3/#6
// (spec.md US3, FR-006 proof set item (c)): two live sessions' pairing-
// signal sightings converge into a live E4 exchange over their own
// sessions (T005), each speaker's Stable prefix is built from ITS bound
// persona (T004 — exchangeSpeaker would fail closed otherwise, per the
// assertions below), the opening turn is genuinely pre-generated (no
// live call for it — hits stays at exactly 2, not 3), the exchange runs a
// SECOND, alternating-body turn rather than stopping at one (the
// ClosingMarker only appears on the reply this test hands to the second
// call), and both turns' FirstTokenLatency land in the session report
// (T007).
//
// The 50ms pause between the two signals is deliberate, not padding: it
// gives the pregen Fill an in-process-server's worth of head start to
// finish before convergence's Take(0) ("check now, don't wait" —
// pregen.go's own doc) checks it, so this test exercises the pregen-served
// path deterministically rather than racing the equally-valid, equally-
// documented live-fallback path (T005's package doc) — a race this test
// isn't trying to settle.
func TestDuskExchange_PairSignalConvergesAndRecordsLatency(t *testing.T) {
	const opening = "Evening, friend."
	const reply = "And to you — fair skies tonight. " + converse.ClosingMarker
	srv, hits := sseTextServer(t, 5*time.Millisecond, opening, reply)
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
	time.Sleep(50 * time.Millisecond) // let the pregen Fill finish before convergence checks it
	if err := dbl.Send(seamtest.Percept("s-1", "Petra", 1, 100, pairingSignalPercept("p-2", "Aldric", "Aldric"))); err != nil {
		t.Fatalf("Petra's pairing signal (convergence): %v", err)
	}

	waitUntil(t, 2*time.Second, func() bool { return len(dbl.Intents()) >= 2 })
	intents := dbl.Intents()

	opener := intents[0]
	if opener["body"] != "Aldric" {
		t.Fatalf("first turn spoke on body %v, want Aldric (the designated opener — the first side to signal)", opener["body"])
	}
	openerPayload, _ := opener["payload"].(map[string]any)
	if openerPayload["verb"] != "speak" {
		t.Fatalf("opener intent verb = %v, want speak", openerPayload["verb"])
	}
	openerTarget, _ := openerPayload["target"].(map[string]any)
	if got := openerTarget["text"]; got != opening {
		t.Fatalf("opener spoken text = %q, want %q", got, opening)
	}

	responder := intents[1]
	if responder["body"] != "Petra" {
		t.Fatalf("second turn spoke on body %v, want Petra — the multi-turn exchange must alternate speakers", responder["body"])
	}
	responderPayload, _ := responder["payload"].(map[string]any)
	responderTarget, _ := responderPayload["target"].(map[string]any)
	if got := responderTarget["text"]; got != "And to you — fair skies tonight." {
		t.Fatalf("responder spoken text = %q, want the reply with ClosingMarker stripped", got)
	}

	if got := hits.Load(); got != 2 {
		t.Fatalf("fake model server hit %d times, want exactly 2 (the pregen fill + Petra's one live call — no wasted fallback call for the opener)", got)
	}

	// The exchange goroutine records FirstTokenLatency after Exchange
	// returns, a beat after the intents themselves are written.
	waitUntil(t, 2*time.Second, func() bool {
		rt.mu.Lock()
		defer rt.mu.Unlock()
		return len(rt.bodies["Aldric"].turnLatencies) >= 1 && len(rt.bodies["Petra"].turnLatencies) >= 1
	})
	report := rt.reportText()
	if !strings.Contains(report, "Aldric: ") || strings.Contains(report, "Aldric: (no dusk exchange turns)") {
		t.Fatalf("session report missing Aldric's FirstTokenLatency:\n%s", report)
	}
	if !strings.Contains(report, "Petra: ") || strings.Contains(report, "Petra: (no dusk exchange turns)") {
		t.Fatalf("session report missing Petra's FirstTokenLatency (the multi-turn responder):\n%s", report)
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

	srv, hits := sseTextServer(t, 0, "Evening.")
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

// TestFR006_ProofSet is not itself a proof — it is FR-006's named entry
// point (card AC #6, T008): `go test -run TestFR006_ProofSet -v` prints
// this doc without exercising anything twice. Every fake-vendor proof
// spec.md's US1-US4 requires runs through the REAL daemon binary (a real
// net.Listener, a real seam.Ingester, a real dialed seamtest.Double) —
// never only a mind/deliberate or mind/converse package double:
//
//	(a) board text percept -> live claim/decline intent, authored reason
//	    TestLiveE3Deliberation_PerceptInIntentOutWithAuthoredReason
//	    (deliberation_test.go). Decline rides the identical wire mechanism
//	    (seam.Pending/ManifestVerbs is verb-agnostic — it composes whatever
//	    verb the model's structured output names); proven for "claim" here
//	    and for both verbs at the package level by mind/deliberate's own
//	    tests, not duplicated at this layer.
//	(b) urgent mid-deliberation -> cancel + exactly one coalesced follow-up
//	    TestLiveInterrupt_CancelsInFlightAndProducesOneFollowUp
//	    (deliberation_test.go)
//	(c) scripted pair signal -> pre-generated opening + multi-turn exchange,
//	    FirstTokenLatency recorded for both speakers
//	    TestDuskExchange_PairSignalConvergesAndRecordsLatency (this file)
//	(d) ambient pool day-crossing refill -> generic serve + targeted escalate
//	    TestAmbientTrigger_ServeAndEscalate (this file)
func TestFR006_ProofSet(t *testing.T) {
	t.Skip("doc-only entry point — see the FR-006 SET enumerated in this test's doc comment; each proof runs as its own named test")
}
