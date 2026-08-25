package main

import (
	"fmt"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"kithcraft/mind/seam"
	"kithcraft/mind/seamtest"
)

var testCapabilities = map[string]any{
	"percept_types": []any{"sighting", "act_result"},
	"origins":       []any{"saw", "acted"},
	"verbs":         []any{map[string]any{"verb": "speak", "targets": []any{"none"}}},
}

func sightingPercept(id string) map[string]any {
	return map[string]any{
		"percept_id": id, "percept_type": "sighting", "urgency": "background",
		"provenance": map[string]any{"origin": "saw", "source": nil, "observed_at": int64(1), "received_at": int64(1)},
		"place":      nil, "content": map[string]any{},
	}
}

func actResultPercept(id, intentID string) map[string]any {
	return map[string]any{
		"percept_id": id, "percept_type": "act_result", "urgency": "notable",
		"provenance": map[string]any{"origin": "acted", "source": nil, "observed_at": int64(1), "received_at": int64(1)},
		"place":      nil,
		"content": map[string]any{
			"intent_id": intentID, "verb": "speak", "outcome": "completed",
			"reason_code": nil, "reason": nil, "detail": "spoke",
		},
	}
}

func waitUntil(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition not met before timeout")
}

// wireDaemonHook installs a skeleton-plumbing OnPercept hook (T012): a
// non-act_result percept provokes one trivial "speak" intent so the
// end-to-end test has something to observe leaving the daemon; an
// act_result percept instead resolves the matching pending intent
// (T010). No deliberation exists (M5 replaces this whole mechanism) —
// this is a test-visible hook, not a policy.
func wireDaemonHook(ing *seam.Ingester, pending *seam.Pending) {
	var mindSeq int64
	var nextID int64
	ing.OnPercept = func(conn seam.Conn, body string, msg map[string]any) {
		payload, _ := msg["payload"].(map[string]any)
		if ptype, _ := payload["percept_type"].(string); ptype == "act_result" {
			content, _ := payload["content"].(map[string]any)
			intentID, _ := content["intent_id"].(string)
			pending.ResolveActResult(intentID)
			return
		}
		id := atomic.AddInt64(&nextID, 1)
		out, err := pending.Compose(fmt.Sprintf("i-%d", id), "speak", map[string]any{"type": "none"}, "reacting to a percept — skeleton plumbing, M5 replaces", "")
		if err != nil {
			return
		}
		mindSeq++
		conn.WriteMessage(map[string]any{
			"protocol": "0.1", "message": "intent", "session": "s-1", "seq": mindSeq,
			"body": body, "world_time": int64(0), "payload": out,
		})
	}
}

// TestEndToEnd_PerceptsInIntentsOut is card AC #5: the daemon starts on a
// real listener, a double dials it, ingests a scripted stream with a
// duplicate percept_id across a reconnect and a seq gap within a
// session, and emits an intent that gets acked and resolved by intent_id.
func TestEndToEnd_PerceptsInIntentsOut(t *testing.T) {
	path := filepath.Join(shortSockDir(t), "mind.sock")
	ing := seam.NewIngester()
	pending := seam.NewPending(map[string]bool{"speak": true})
	wireDaemonHook(ing, pending)

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

	if err := dbl.Send(seamtest.SessionOpen("s-1", "b-1", "second", testCapabilities, nil)); err != nil {
		t.Fatalf("session_open: %v", err)
	}

	// seq 1: ordinary percept.
	if err := dbl.Send(seamtest.Percept("s-1", "b-1", 1, 1001, sightingPercept("p-1"))); err != nil {
		t.Fatal(err)
	}
	// seq 4: a gap (seq 2,3 shed as background percepts, §3.3).
	if err := dbl.Send(seamtest.Percept("s-1", "b-1", 4, 1004, sightingPercept("p-2"))); err != nil {
		t.Fatal(err)
	}

	waitUntil(t, 2*time.Second, func() bool { return len(dbl.Intents()) >= 2 })
	if got := ing.ShedCount("b-1"); got != 2 {
		t.Fatalf("ShedCount(b-1) = %d, want 2 (seq 2,3 shed within this session)", got)
	}

	// Reconnect: a new connection, a new session id, same body — and a
	// retransmitted p-1 must be deduped (seam-wire-v0.md §3.4).
	dbl.Close()
	dbl2, err := seamtest.DialUnix(path)
	if err != nil {
		t.Fatalf("reconnect dial: %v", err)
	}
	defer dbl2.Close()
	if err := dbl2.Send(seamtest.SessionOpen("s-2", "b-1", "second", testCapabilities, map[string]any{
		"previous_session": "s-1", "previous_close_world_time": int64(1004), "body_continuous": true,
	})); err != nil {
		t.Fatalf("reconnect session_open: %v", err)
	}
	if err := dbl2.Send(seamtest.Percept("s-2", "b-1", 1, 2001, sightingPercept("p-1"))); err != nil {
		t.Fatal(err)
	}
	// Give the daemon a moment to (not) react to the dup, then prove it
	// produced no new intent for it.
	time.Sleep(200 * time.Millisecond)
	if got := len(dbl2.Intents()); got != 0 {
		t.Fatalf("reconnect connection recorded %d intents, want 0: a retransmitted percept_id must not provoke one", got)
	}

	// Ack and resolve the first intent the daemon emitted.
	intents := dbl.Intents()
	if len(intents) == 0 {
		t.Fatal("expected at least one emitted intent (card AC #5)")
	}
	first := intents[0]
	payload, _ := first["payload"].(map[string]any)
	intentID, _ := payload["intent_id"].(string)
	if intentID == "" {
		t.Fatalf("emitted intent had no intent_id: %#v", first)
	}
	if !pending.IsPending(intentID) {
		t.Fatalf("intent %s should still be pending before any ack/act_result", intentID)
	}

	if err := dbl2.Send(seamtest.IntentAck("s-2", "b-1", 2, intentID, true)); err != nil {
		t.Fatal(err)
	}
	if err := dbl2.Send(seamtest.Percept("s-2", "b-1", 3, 2002, actResultPercept("p-3", intentID))); err != nil {
		t.Fatal(err)
	}
	waitUntil(t, 2*time.Second, func() bool { return !pending.IsPending(intentID) })
}

// TestEndToEnd_RestartReconnect_NoBackfill is card AC #6, through the
// real listener rather than the fakeConn used by Phase 2's continuity
// test (seam/session_test.go): the daemon process is stood down (its
// listener closed) and a fresh one started at the same path, with a
// fresh Ingester — exactly what a real restart leaves behind. The
// reconnecting body is matched by its `body` token alone and the daemon
// invents nothing about the gap.
func TestEndToEnd_RestartReconnect_NoBackfill(t *testing.T) {
	path := filepath.Join(shortSockDir(t), "mind.sock")

	ing1 := seam.NewIngester()
	ln1, err := listen(path)
	if err != nil {
		t.Fatal(err)
	}
	go serve(ln1, ing1)

	dbl1, err := seamtest.DialUnix(path)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	if err := dbl1.Send(seamtest.SessionOpen("s-1", "b-eda", "second", testCapabilities, nil)); err != nil {
		t.Fatalf("session_open: %v", err)
	}
	dbl1.Close() // vendor drops the connection: the daemon "dies"
	ln1.Close()  // and its listener goes with it (Close unlinks the socket)

	// A brand-new listener and Ingester stand in for the restarted
	// daemon process: nothing from ing1 carries over.
	ing2 := seam.NewIngester()
	ln2, err := listen(path)
	if err != nil {
		t.Fatalf("listen after restart: %v", err)
	}
	defer ln2.Close()
	go serve(ln2, ing2)

	dbl2, err := seamtest.DialUnix(path)
	if err != nil {
		t.Fatalf("reconnect dial: %v", err)
	}
	defer dbl2.Close()
	err = dbl2.Send(seamtest.SessionOpen("s-2", "b-eda", "second", testCapabilities, map[string]any{
		"previous_session": "s-1", "previous_close_world_time": int64(1000), "body_continuous": true,
	}))
	if err != nil {
		t.Fatalf("reconnect session_open: %v", err)
	}

	time.Sleep(200 * time.Millisecond)
	if err := dbl2.ReadErr(); err != nil {
		t.Fatalf("reconnect with continuity was refused: %v", err)
	}
	if got := len(dbl2.Intents()); got != 0 {
		t.Fatalf("restart must invent nothing for the gap, but the daemon emitted %d intents", got)
	}
	if got := ing2.ShedCount("b-eda"); got != 0 {
		t.Fatalf("fresh Ingester after restart must report no shed count, got %d", got)
	}
}
