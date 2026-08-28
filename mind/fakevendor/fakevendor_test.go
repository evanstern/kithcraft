// Package fakevendor_test: Phase 1's shape, default-behaviour, and
// API-surface tests (specs/015-fake-vendor-harness tasks T002/T003). It is
// external (mirrors mind/memory/beliefs_external_test.go) because card AC
// #4's proof — no read API, no capability beyond §10.1 — must hold from
// outside the package boundary, the same vantage a real caller has.
//
// Every test here drives FakeVendor over a real seam.Conn (net.Pipe), with
// the "mind" side played by the test itself sending raw envelopes and
// reading FakeVendor's replies — the same from-the-outside pattern
// mind/seamtest uses, so what's proven here is proven of the real wire, not
// a hand-rolled shortcut. Both sides of the pipe need a continuously
// reading background goroutine (net.Pipe's Write blocks until something
// Reads); FakeVendor supplies its own once Open() runs, and mindSide below
// supplies the other, mirroring mind/seamtest.Double's record loop.
package fakevendor_test

import (
	"net"
	"reflect"
	"sort"
	"testing"
	"time"

	"kithcraft/mind/fakevendor"
	"kithcraft/mind/seam"
)

// mindSide plays the mind's half of the connection directly: it drains
// every inbound message into a channel (so writes on either side of the
// net.Pipe never block waiting for a reader) and lets the test send raw
// envelopes and assert on what arrives.
type mindSide struct {
	conn seam.Conn
	ch   chan map[string]any
}

func newPair(t *testing.T) (vendorSide seam.Conn, mind *mindSide) {
	t.Helper()
	c1, c2 := net.Pipe()
	t.Cleanup(func() { _ = c1.Close(); _ = c2.Close() })
	m := &mindSide{conn: seam.NewWireConn(c2), ch: make(chan map[string]any, 16)}
	go func() {
		for {
			msg, err := m.conn.ReadMessage()
			if err != nil {
				return
			}
			m.ch <- msg
		}
	}()
	return seam.NewWireConn(c1), m
}

func (m *mindSide) send(t *testing.T, msg map[string]any) {
	t.Helper()
	if err := m.conn.WriteMessage(msg); err != nil {
		t.Fatalf("mindSide.send: %v", err)
	}
}

func (m *mindSide) next(t *testing.T) map[string]any {
	t.Helper()
	select {
	case msg := <-m.ch:
		return msg
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for a message from FakeVendor")
		return nil
	}
}

func (m *mindSide) expectNone(t *testing.T) {
	t.Helper()
	select {
	case msg := <-m.ch:
		t.Fatalf("expected no message, got %#v", msg)
	case <-time.After(100 * time.Millisecond):
	}
}

func coreManifest() map[string]any {
	caps := map[string]any{
		"percept_types": []any{"sighting", "observation", "sound", "speech", "act_result"},
		"origins":       []any{"acted", "saw", "heard", "felt", "told", "read"},
		"verbs": []any{
			map[string]any{"verb": "go_to", "targets": []any{"place", "thing", "body"}},
			map[string]any{"verb": "speak", "targets": []any{"body", "person", "none"}},
			map[string]any{"verb": "attend", "targets": []any{"place", "none"}},
			map[string]any{"verb": "wait", "targets": []any{"none"}},
		},
		"salient_kinds":  []any{},
		"bearings":       []any{"ahead", "behind", "left", "right"},
		"distance_bands": []any{"here", "near", "middling", "far"},
	}
	return fakevendor.Manifest("second", caps, nil)
}

func intentEnvelope(session, body, intentID, verb string, seq int64) map[string]any {
	return map[string]any{
		"protocol": "0.1", "message": "intent", "session": session,
		"seq": seq, "body": body, "world_time": int64(0),
		"payload": map[string]any{
			"intent_id": intentID, "verb": verb,
			"target": map[string]any{"type": "none"}, "reason": "because",
		},
	}
}

func percept(kind string, worldTime int64) map[string]any {
	return map[string]any{
		"percept_id": "p-1", "percept_type": kind, "urgency": "background",
		"provenance": map[string]any{"origin": "saw", "source": nil, "observed_at": worldTime, "received_at": worldTime},
		"place":      nil, "content": map[string]any{},
	}
}

// TestFakeVendor_Manifest_IsWireValid is T002: the session_open FakeVendor
// opens with round-trips through the real wire's V-5 presence validation
// (mind/wire.Validate, run inside seam.NewWireConn.ReadMessage) without
// error — §6.2's required shape, proven against the real seam surface
// rather than asserted in prose.
func TestFakeVendor_Manifest_IsWireValid(t *testing.T) {
	vendorSide, mind := newPair(t)
	v := fakevendor.New(vendorSide, "s-1", "b-1", coreManifest())
	if err := v.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}

	msg := mind.next(t)
	if msg["message"] != "session_open" {
		t.Fatalf("first message = %v, want session_open", msg["message"])
	}
	pl, _ := msg["payload"].(map[string]any)
	if pl["time_unit"] != "second" {
		t.Fatalf("payload.time_unit = %v, want %q", pl["time_unit"], "second")
	}
	if _, ok := pl["capabilities"].(map[string]any); !ok {
		t.Fatalf("payload.capabilities missing or not an object: %#v", pl)
	}
	if !reflect.DeepEqual(v.Manifest(), pl) {
		t.Fatalf("Manifest() = %#v, want the payload actually sent %#v", v.Manifest(), pl)
	}
}

// TestFakeVendor_Advance_MovesTimeAndNothingElse is T002: Advance changes
// world_time on the next envelope and emits nothing of its own.
func TestFakeVendor_Advance_MovesTimeAndNothingElse(t *testing.T) {
	vendorSide, mind := newPair(t)
	v := fakevendor.New(vendorSide, "s-1", "b-1", coreManifest())
	if err := v.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	mind.next(t) // session_open

	v.Advance(600)
	mind.expectNone(t) // advance alone must produce no message

	if err := v.Emit(percept("sighting", 0)); err != nil {
		t.Fatalf("Emit: %v", err)
	}
	msg := mind.next(t)
	if msg["world_time"] != int64(600) {
		t.Fatalf("world_time = %v, want 600 after Advance(600)", msg["world_time"])
	}
}

// TestFakeVendor_DefaultIntentBehaviour_AckRecordWait is T002 + FR-002: an
// intent is acked accepted:true and recorded, and produces no act_result
// until Resolve is called.
func TestFakeVendor_DefaultIntentBehaviour_AckRecordWait(t *testing.T) {
	vendorSide, mind := newPair(t)
	v := fakevendor.New(vendorSide, "s-1", "b-1", coreManifest())
	if err := v.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	mind.next(t) // session_open

	mind.send(t, intentEnvelope("s-1", "b-1", "i-1", "wait", 1))

	ack := mind.next(t)
	if ack["message"] != "intent_ack" {
		t.Fatalf("reply = %v, want intent_ack", ack["message"])
	}
	apl, _ := ack["payload"].(map[string]any)
	if apl["intent_id"] != "i-1" || apl["accepted"] != true {
		t.Fatalf("intent_ack payload = %#v, want accepted:true for i-1", apl)
	}

	acts := v.Acts()
	if len(acts) != 1 || acts[0]["intent_id"] != "i-1" || acts[0]["verb"] != "wait" {
		t.Fatalf("Acts() = %#v, want exactly the i-1/wait intent recorded", acts)
	}

	mind.expectNone(t) // no act_result until Resolve

	detail := "done waiting"
	if err := v.Resolve("i-1", "completed", nil, &detail); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	result := mind.next(t)
	if result["message"] != "percept" {
		t.Fatalf("resolve produced %v, want percept", result["message"])
	}
	rpl, _ := result["payload"].(map[string]any)
	if rpl["percept_type"] != "act_result" {
		t.Fatalf("payload.percept_type = %v, want act_result", rpl["percept_type"])
	}
	content, _ := rpl["content"].(map[string]any)
	if content["intent_id"] != "i-1" || content["outcome"] != "completed" || content["verb"] != "wait" || content["detail"] != "done waiting" {
		t.Fatalf("act_result content = %#v", content)
	}
}

// TestFakeVendor_Resolve_UnknownIntentID_LoudError proves the edge case in
// spec.md's Edge Cases: resolving an intent_id the vendor never issued a
// pending act for is a script error, never a silent no-op.
func TestFakeVendor_Resolve_UnknownIntentID_LoudError(t *testing.T) {
	vendorSide, _ := newPair(t)
	v := fakevendor.New(vendorSide, "s-1", "b-1", coreManifest())
	if err := v.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := v.Resolve("i-ghost", "completed", nil, nil); err == nil {
		t.Fatal("Resolve on an unknown intent_id must return an error, not no-op")
	}
}

// TestFakeVendor_Resolve_Twice_LoudError proves the same for an
// already-resolved intent_id.
func TestFakeVendor_Resolve_Twice_LoudError(t *testing.T) {
	vendorSide, mind := newPair(t)
	v := fakevendor.New(vendorSide, "s-1", "b-1", coreManifest())
	if err := v.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	mind.next(t) // session_open
	mind.send(t, intentEnvelope("s-1", "b-1", "i-1", "wait", 1))
	mind.next(t) // intent_ack

	if err := v.Resolve("i-1", "completed", nil, nil); err != nil {
		t.Fatalf("first Resolve: %v", err)
	}
	mind.next(t) // act_result

	if err := v.Resolve("i-1", "completed", nil, nil); err == nil {
		t.Fatal("resolving i-1 a second time must return an error, not no-op")
	}
}

// TestFakeVendor_Emit_AfterClose_LoudError proves the other named edge
// case: a closed session emits nothing, loudly.
func TestFakeVendor_Emit_AfterClose_LoudError(t *testing.T) {
	vendorSide, mind := newPair(t)
	v := fakevendor.New(vendorSide, "s-1", "b-1", coreManifest())
	if err := v.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	mind.next(t) // session_open
	if err := v.Close("shutdown"); err != nil {
		t.Fatalf("Close: %v", err)
	}
	mind.next(t) // session_close

	if err := v.Emit(percept("sighting", 0)); err == nil {
		t.Fatal("Emit after Close must return an error, not no-op")
	}
}

// TestFakeVendor_ExportedSurface_IsExactlySpec101 is T003 / FR-006 / card
// AC #4-#5: *FakeVendor's exported method set and field set are exactly
// §10.1's shape. Any addition — a read/query method, a capability the
// contract doesn't name — fails this test, which is the point: the fake
// vendor is the most convenient place in the codebase to add a read API by
// accident (§10.5), and this test is what makes that accident loud instead
// of a silent SI-1 violation.
func TestFakeVendor_ExportedSurface_IsExactlySpec101(t *testing.T) {
	typ := reflect.TypeOf(&fakevendor.FakeVendor{})
	var methods []string
	for i := 0; i < typ.NumMethod(); i++ {
		methods = append(methods, typ.Method(i).Name)
	}
	sort.Strings(methods)
	wantMethods := []string{"Acts", "Advance", "Close", "Emit", "Manifest", "Open", "Resolve"}
	sort.Strings(wantMethods)
	if !reflect.DeepEqual(methods, wantMethods) {
		t.Fatalf("*FakeVendor methods = %v, want exactly %v (§10.1)", methods, wantMethods)
	}

	elem := typ.Elem()
	var fields []string
	for i := 0; i < elem.NumField(); i++ {
		if f := elem.Field(i); f.IsExported() {
			fields = append(fields, f.Name)
		}
	}
	sort.Strings(fields)
	wantFields := []string{"RestrictChangeReports", "Strict"}
	sort.Strings(wantFields)
	if !reflect.DeepEqual(fields, wantFields) {
		t.Fatalf("FakeVendor exported fields = %v, want exactly %v (§10.1)", fields, wantFields)
	}
}
