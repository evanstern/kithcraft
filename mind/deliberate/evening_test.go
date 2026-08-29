// Package deliberate (this file): T010 — the scripted-evening
// micromanagement design check (tasks.md T010, card AC #8) — across
// several board postings in one evening, run against the real FakeVendor
// over the real seam wire (mind/fakevendor, the same from-the-outside
// pattern mind/fakevendor/e2e_test.go uses, rather than board_test.go's
// direct VendorFunc doubles), work gets done without the player
// re-posting (each posting is answered exactly once, and a claimed job
// resolves without the vendor ever re-emitting it), and every response
// carries a reason legible as this villager's own — grounded in its own
// posting and its own persona reasoning, never a boilerplate refusal
// (spec.md FR-007, kithcraft-brief.md #6).
package deliberate

import (
	"context"
	"net"
	"testing"
	"time"

	"kithcraft/mind/fakevendor"
	"kithcraft/mind/llm"
	"kithcraft/mind/seam"
)

// wireVendor adapts a mind-side seam.Conn to deliberate.Vendor exactly as
// loop.go's Vendor doc describes a real wiring: stamp the envelope
// (protocol, session, seq, body, world_time) around Compose's payload and
// write it over the wire — in place of Phases 1-3's VendorFunc doubles,
// which hand the payload straight to a test channel with no wire in
// between.
type wireVendor struct {
	conn    seam.Conn
	session string
	body    string
	seq     int64
	sent    chan map[string]any // test introspection: every payload sent
}

func (v *wireVendor) SendIntent(payload map[string]any) error {
	env := map[string]any{
		"protocol": "0.1", "message": "intent", "session": v.session,
		"seq": v.seq, "body": v.body, "world_time": int64(0),
		"payload": payload,
	}
	v.seq++
	if v.sent != nil {
		v.sent <- payload
	}
	return v.conn.WriteMessage(env)
}

// boardPosting is one scripted evening posting: the text percept FakeVendor
// emits and the villager's scripted (claim or decline) structured-output
// response (A-9) to it.
type boardPosting struct {
	perceptID string
	text      string
	rawIntent string
	wantVerb  string
}

// TestEveningOfPostings_WorkGetsDoneWithoutRepostingOrPolicing is the
// scripted-evening design check itself: three board postings cross the
// wire from a real FakeVendor to one villager's Loop; the claims resolve
// to completed act_results and the decline carries its own persona reason
// — the vendor never has to re-emit a posting to get an answer, and no two
// postings share a canned response (card AC #8).
func TestEveningOfPostings_WorkGetsDoneWithoutRepostingOrPolicing(t *testing.T) {
	c1, c2 := net.Pipe()
	t.Cleanup(func() { _ = c1.Close(); _ = c2.Close() })
	vendorConn, mindConn := seam.NewWireConn(c1), seam.NewWireConn(c2)

	capabilities := map[string]any{"verbs": []any{
		map[string]any{"verb": "claim", "targets": []any{}},
		map[string]any{"verb": "decline", "targets": []any{}},
	}}
	v := fakevendor.New(vendorConn, "s-evening", "b-villager", fakevendor.Manifest("minute", capabilities, nil))

	sent := make(chan map[string]any, 1)
	l := New(Config{
		Verbs:  ManifestVerbs(capabilities),
		Vendor: &wireVendor{conn: mindConn, session: "s-evening", body: "b-villager", sent: sent},
		Class:  llm.E3,
	})

	// The mind-side read loop must be running before Open, exactly like
	// mind/fakevendor/fakevendor_test.go's newPair: net.Pipe's Write blocks
	// until something Reads, and Open's session_open is the first write.
	// acks is drained so the test can wait for FakeVendor's intent_ack —
	// the signal that recordAndAck has actually admitted the intent into
	// its pending set, so a Resolve right after never races it (§10.1).
	postings := make(chan map[string]any, 8)
	acks := make(chan map[string]any, 8)
	go func() {
		for {
			msg, err := mindConn.ReadMessage()
			if err != nil {
				return
			}
			payload, _ := msg["payload"].(map[string]any)
			switch msg["message"] {
			case "percept":
				if payload["percept_type"] == "act_result" {
					l.Deliver(payload)
					continue
				}
				postings <- payload
			case "intent_ack":
				acks <- payload
			}
		}
	}()

	if err := v.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}

	evening := []boardPosting{
		{
			perceptID: "p-post-1", text: "Build a shelter by the north wall. — the player",
			rawIntent: `{"verb":"claim","reason":"Building shelters is exactly my trade, and that wall has bothered me all week."}`,
			wantVerb:  "claim",
		},
		{
			perceptID: "p-post-2", text: "Mind the well while I'm away. — the player",
			rawIntent: `{"verb":"decline","reason":"I promised Mira I'd cover the stall until dusk, and the well can't wait on hands that are already full."}`,
			wantVerb:  "decline",
		},
		{
			perceptID: "p-post-3", text: "Clear the fallen branches from the orchard path. — the player",
			rawIntent: `{"verb":"claim","reason":"I was already heading that way to check on the trees."}`,
			wantVerb:  "claim",
		},
	}

	var reasons []string
	for _, p := range evening {
		posting := map[string]any{
			"percept_id": p.perceptID, "percept_type": "text", "urgency": "notable",
			"provenance": map[string]any{"origin": "read", "source": nil, "observed_at": int64(0), "received_at": int64(0)},
			"place":      nil,
			"content":    map[string]any{"text": p.text, "attributed_to": "the player"},
		}
		if err := v.Emit(posting); err != nil {
			t.Fatalf("Emit(%s): %v", p.perceptID, err)
		}

		var received map[string]any
		select {
		case received = <-postings:
		case <-time.After(2 * time.Second):
			t.Fatalf("%s: timed out waiting for the posting to reach the mind", p.perceptID)
		}
		if !TriggerE3(received) {
			t.Fatalf("%s: TriggerE3 = false, want true for a text/origin:read percept", p.perceptID)
		}
		content, _ := received["content"].(map[string]any)
		e3ctx := E3Context{Board: content["text"].(string)}
		if e3ctx.VariableContext().Render() == "" {
			t.Fatalf("%s: E3Context rendered empty — the board content must reach the deliberation context", p.perceptID)
		}

		done := make(chan struct{})
		var runErr error
		go func() {
			_, runErr = l.Run(context.Background(), queueProposer(p.rawIntent))
			close(done)
		}()

		var payload map[string]any
		select {
		case payload = <-sent:
		case <-time.After(2 * time.Second):
			t.Fatalf("%s: timed out waiting for the villager's intent", p.perceptID)
		}
		if payload["verb"] != p.wantVerb {
			t.Fatalf("%s: sent verb = %v, want %v", p.perceptID, payload["verb"], p.wantVerb)
		}
		reason, _ := payload["reason"].(string)
		if reason == "" {
			t.Fatalf("%s: intent carries no authored reason (card AC #8: a refusal must be legible as this villager's)", p.perceptID)
		}
		reasons = append(reasons, reason)

		intentID, _ := payload["intent_id"].(string)
		select {
		case ack := <-acks:
			if ack["intent_id"] != intentID || ack["accepted"] != true {
				t.Fatalf("%s: intent_ack = %#v, want accepted:true for %q", p.perceptID, ack, intentID)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("%s: timed out waiting for intent_ack", p.perceptID)
		}
		if err := v.Resolve(intentID, "completed", nil, nil); err != nil {
			t.Fatalf("%s: Resolve: %v", p.perceptID, err)
		}

		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatalf("%s: Run did not complete after its act_result was resolved", p.perceptID)
		}
		if runErr != nil {
			t.Fatalf("%s: Run: %v", p.perceptID, runErr)
		}
	}

	// Work got done without re-posting: the vendor recorded exactly one
	// intent per posting — it never had to re-emit a posting to get an
	// answer out of the villager.
	acts := v.Acts()
	if len(acts) != len(evening) {
		t.Fatalf("vendor recorded %d intents for %d postings, want exactly one intent per posting (no re-posting needed)", len(acts), len(evening))
	}

	// No two postings share a reason: the decline (posting 2) and both
	// claims are each grounded in their own posting, not a canned string
	// standing in for actual persona reasoning.
	seen := map[string]bool{}
	for _, r := range reasons {
		if seen[r] {
			t.Fatalf("reason %q repeated across postings — a generic response, not persona-grounded (card AC #8)", r)
		}
		seen[r] = true
	}
}
