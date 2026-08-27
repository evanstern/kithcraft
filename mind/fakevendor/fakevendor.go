// Package fakevendor implements FakeVendor, the scripted in-memory vendor
// double of docs/design/body-protocol-v0.md §10: the full protocol surface,
// nothing more, driven by a script rather than a world (S2, TASK-0015). It
// grows mind/seamtest's from-the-outside pattern — a mind under test dials
// or is dialed over the real wire (mind/seam, mind/wire), never a hand-rolled
// channel — so passing against FakeVendor means passing against any world.
// mind/seamtest stays as M1's minimal recorder; FakeVendor is its full-shape
// successor for mind-side tests that need session lifecycle, acks, and
// resolvable acts, not just a percept-in/intent-out recording double.
//
// Scope discipline (§10.5) is structural, not a comment: the exported
// surface below is exactly §10.1's shape (see fakevendor_test.go's
// API-surface test) — no read API, no autonomous behaviour. Emit and
// Resolve send exactly what the caller hands them (or, for Resolve, an
// act_result assembled only from the intent's own fields); nothing here
// ever decides what a percept says.
package fakevendor

import (
	"fmt"
	"sync"

	"kithcraft/mind/seam"
)

// Manifest builds a §6.2-valid session_open payload from the vendor's
// declared vocabulary. continuity may be nil (a fresh body, no reconnect
// to report). The manifest describes the vendor, never the world (§6.2):
// it MUST NOT vary with anything Emit is later asked to send, and nothing
// here lets it.
func Manifest(timeUnit string, capabilities, continuity map[string]any) map[string]any {
	payload := map[string]any{"time_unit": timeUnit, "capabilities": capabilities}
	if continuity != nil {
		payload["continuity"] = continuity
	}
	return payload
}

// FakeVendor is one scripted vendor connection for one body, playing the
// vendor's half of conn (a net.Pipe half or a dialed UDS connection wrapped
// by seam.NewWireConn — whatever the mind under test's other end is). It
// has no pathfinding, no simulation, no autonomous behaviour: if a `go_to`
// should succeed, the script says so via Resolve.
type FakeVendor struct {
	conn     seam.Conn
	session  string
	body     string
	manifest map[string]any

	// Strict is the V-5 posture: reject malformed rather than coerce
	// (§10.1). RestrictChangeReports is §4.10's delivery restriction.
	// Both default true (New) — the correct posture — and are read by
	// the H-tests (Phase 2/3) that flip them to reproduce a violation
	// on purpose; this package does not yet condition any behaviour on
	// either (ponytail: add when H-1/H-2/H-6 land and need it).
	Strict                bool
	RestrictChangeReports bool

	mu        sync.Mutex // guards everything below except writeMu's own critical section
	worldTime int64
	seq       int64
	nextPID   int64
	opened    bool
	closed    bool
	acts      []map[string]any
	pending   map[string]map[string]any // intent_id -> received intent payload

	// issued is every thing_id/place/body token this vendor has ever
	// mentioned in an Emit'd percept (§5.2: "a token the mind already
	// knows"). It is bookkeeping derived from what Emit is given, never
	// a decision about a percept's content (§10.5) — recordAndAck reads
	// it only to tell an issued-but-gone target from one this vendor
	// never handed out (H-5, §5.3/§5.6), never to answer "does X still
	// exist" (that would be the existence oracle §5.6 forbids).
	issued map[string]bool

	// writeMu serializes conn.WriteMessage calls: wire.WriteFrame forbids
	// concurrent writers on one connection (mind/wire/frame.go), and both
	// the caller (Emit/Resolve/Close) and the background recordAndAck
	// loop (intent_ack) write to conn.
	writeMu sync.Mutex
}

// New wires a FakeVendor to conn for one session/body identity. manifest is
// the payload Open() sends as session_open (build it with Manifest).
// Nothing is written to conn until Open().
func New(conn seam.Conn, session, body string, manifest map[string]any) *FakeVendor {
	return &FakeVendor{
		conn: conn, session: session, body: body, manifest: manifest,
		Strict: true, RestrictChangeReports: true,
		pending: map[string]map[string]any{},
		issued:  map[string]bool{},
	}
}

// Manifest returns the §6.2 payload this vendor opens with.
func (v *FakeVendor) Manifest() map[string]any { return v.manifest }

// Open sends session_open (world_time 0, seq 0) and starts the background
// loop that records every intent the mind sends, in receipt order, acking
// each accepted:true by default and leaving it pending until Resolve
// (§10.1's default behaviour — waiting is the state minds occupy most).
func (v *FakeVendor) Open() error {
	v.mu.Lock()
	v.opened = true
	env := v.envelope("session_open", v.manifest)
	v.mu.Unlock()
	if err := v.write(env); err != nil {
		return fmt.Errorf("fakevendor: session_open: %w", err)
	}
	go v.recordAndAck()
	return nil
}

// Close sends session_close(reason) and ends the connection. A second
// Close, or any Emit/Resolve after Close, is a script error (loud failure,
// never a no-op): a closed session emits nothing.
func (v *FakeVendor) Close(reason string) error {
	v.mu.Lock()
	if v.closed {
		v.mu.Unlock()
		return fmt.Errorf("fakevendor: Close called twice")
	}
	v.closed = true
	env := v.envelope("session_close", map[string]any{"reason": reason, "detail": nil})
	v.mu.Unlock()
	err := v.write(env)
	_ = v.conn.Close()
	if err != nil {
		return fmt.Errorf("fakevendor: session_close: %w", err)
	}
	return nil
}

// Emit pushes one percept (§4.1 shape: percept_id, percept_type, urgency,
// provenance, place, content — the caller's to build, valid or not) to the
// mind. Emitting after Close is a script error.
func (v *FakeVendor) Emit(percept map[string]any) error {
	v.mu.Lock()
	if !v.opened || v.closed {
		v.mu.Unlock()
		return fmt.Errorf("fakevendor: Emit before Open or after Close (script error): a closed session emits nothing")
	}
	v.noteIssuedTokens(percept)
	env := v.envelope("percept", percept)
	v.mu.Unlock()
	if err := v.write(env); err != nil {
		return fmt.Errorf("fakevendor: emit: %w", err)
	}
	return nil
}

// Advance moves world_time forward by n and does nothing else (§10.1) —
// no percept, no autonomous resolution of any pending act.
func (v *FakeVendor) Advance(n int64) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.worldTime += n
}

// Acts returns every intent received so far, in receipt order — the
// assertion surface for tests. It is never handed to the mind.
func (v *FakeVendor) Acts() []map[string]any {
	v.mu.Lock()
	defer v.mu.Unlock()
	out := make([]map[string]any, len(v.acts))
	copy(out, v.acts)
	return out
}

// Resolve emits the act_result (§5.4) for a pending intent: outcome,
// reason_code and detail as given, verb and reason echoed from the intent
// as received. Resolving an unknown or already-resolved intent_id is a
// script error — never a silent no-op, which would let a test pass while
// asserting nothing.
func (v *FakeVendor) Resolve(intentID, outcome string, reasonCode, detail *string) error {
	v.mu.Lock()
	intent, ok := v.pending[intentID]
	if !ok {
		v.mu.Unlock()
		return fmt.Errorf("fakevendor: Resolve(%q): unknown or already-resolved intent_id (script error)", intentID)
	}
	delete(v.pending, intentID)
	v.nextPID++
	pid := fmt.Sprintf("p-fv-act-%d", v.nextPID)
	wt := v.worldTime
	content := map[string]any{
		"intent_id": intentID, "verb": intent["verb"], "outcome": outcome,
		"reason_code": ptrToAny(reasonCode), "reason": intent["reason"], "detail": ptrToAny(detail),
	}
	percept := map[string]any{
		"percept_id": pid, "percept_type": "act_result", "urgency": "notable",
		"provenance": map[string]any{"origin": "acted", "source": nil, "observed_at": wt, "received_at": wt},
		"place":      nil,
		"content":    content,
	}
	closed := v.closed
	env := v.envelope("percept", percept)
	v.mu.Unlock()
	if closed {
		return fmt.Errorf("fakevendor: Resolve after Close (script error): a closed session emits nothing")
	}
	if err := v.write(env); err != nil {
		return fmt.Errorf("fakevendor: resolve: %w", err)
	}
	return nil
}

// write serializes conn.WriteMessage against recordAndAck's own writes
// (intent_ack) — see writeMu's doc comment.
func (v *FakeVendor) write(env map[string]any) error {
	v.writeMu.Lock()
	defer v.writeMu.Unlock()
	return v.conn.WriteMessage(env)
}

// envelope builds one outbound message, assigning the next seq in this
// body's vendor->mind counter (seam-wire-v0.md §3.2: it starts at 0 with
// session_open and every later message shares it). Caller holds v.mu.
func (v *FakeVendor) envelope(kind string, payload map[string]any) map[string]any {
	env := map[string]any{
		"protocol": "0.1", "message": kind, "session": v.session,
		"seq": v.seq, "body": v.body, "world_time": v.worldTime,
		"payload": payload,
	}
	v.seq++
	return env
}

// recordAndAck is the read loop playing the vendor's reactive half: every
// intent is recorded (Acts) and acked accepted:true, then left pending for
// Resolve. Nothing else is inspected or acted on (no autonomous behaviour).
func (v *FakeVendor) recordAndAck() {
	for {
		msg, err := v.conn.ReadMessage()
		if err != nil {
			return // connection closed or fatal; nothing further to record
		}
		if msg["message"] != "intent" {
			continue
		}
		payload, _ := msg["payload"].(map[string]any)
		intentID, _ := payload["intent_id"].(string)

		v.mu.Lock()
		v.acts = append(v.acts, payload)
		var ack map[string]any
		if v.targetUnissued(payload) {
			// H-5 / §5.3: a token this vendor never issued is refused
			// here, synchronously, and never joins pending — no
			// act_result will ever follow for it. An issued token whose
			// referent is gone takes the other branch below and fails
			// only later, at Resolve (§5.6).
			ack = v.envelope("intent_ack", map[string]any{
				"intent_id": intentID, "accepted": false, "reason_code": "unknown_target",
			})
		} else {
			v.pending[intentID] = payload
			ack = v.envelope("intent_ack", map[string]any{
				"intent_id": intentID, "accepted": true, "reason_code": nil,
			})
		}
		v.mu.Unlock()
		_ = v.write(ack) // best-effort: a write failure surfaces via the next ReadMessage
	}
}

// targetUnissued reports whether payload's target names a thing/place/body
// token this vendor never mentioned in an Emit'd percept (§5.2). It is the
// only synchronous ack-time refusal H-5 sanctions — it checks nothing about
// whether the referent still exists, only whether the identifier was ever
// handed out; an issued token whose referent is gone is NOT unissued and
// takes the accepted branch, failing only later at Resolve's target_gone
// (§5.3/§5.6 — this is what keeps the ack from becoming an existence
// oracle). Caller holds v.mu.
func (v *FakeVendor) targetUnissued(payload map[string]any) bool {
	target, ok := payload["target"].(map[string]any)
	if !ok {
		return false
	}
	key, _ := target["type"].(string)
	if key != "thing" && key != "place" && key != "body" {
		return false // "none" or unrecognized: no token to check
	}
	id, _ := target[key].(string)
	if id == "" {
		return false
	}
	return !v.issued[id]
}

// noteIssuedTokens records every thing_id/place identifier percept mentions
// as now known to the mind (§5.2) — called on every Emit so targetUnissued
// can later tell an issued-but-gone token from one never handed out. This
// is bookkeeping derived from what Emit is given; it does not gate, alter,
// or interpret the percept (§10.5) — nothing here decides what a percept
// says, only records identifiers already present in it. Caller holds v.mu.
func (v *FakeVendor) noteIssuedTokens(percept map[string]any) {
	if place, ok := percept["place"].(map[string]any); ok {
		if id, ok := place["place"].(string); ok && id != "" {
			v.issued[id] = true
		}
	}
	content, _ := percept["content"].(map[string]any)
	for _, key := range []string{"thing", "about_place"} {
		m, ok := content[key].(map[string]any)
		if !ok {
			continue
		}
		for _, idKey := range []string{"thing_id", "place", "body"} {
			if id, ok := m[idKey].(string); ok && id != "" {
				v.issued[id] = true
			}
		}
	}
	if present, ok := content["present"].([]any); ok {
		for _, item := range present {
			if m, ok := item.(map[string]any); ok {
				if id, ok := m["thing_id"].(string); ok && id != "" {
					v.issued[id] = true
				}
			}
		}
	}
}

func ptrToAny(s *string) any {
	if s == nil {
		return nil
	}
	return *s
}
