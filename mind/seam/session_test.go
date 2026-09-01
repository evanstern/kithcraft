package seam

import (
	"testing"
	"time"

	"kithcraft/mind/wire"
)

// fakeConn is an in-memory Conn double: the test scripts inbound messages
// with script() and closeStream(), and inspects everything HandleConnection
// wrote via out(). It satisfies Conn directly — no wire/UDS involved — since
// the session layer's own behavior, not the codec, is what these tests pin.
type fakeConn struct {
	in     chan map[string]any
	inErr  chan error
	outCh  chan map[string]any
	closed chan struct{}
}

func newFakeConn() *fakeConn {
	return &fakeConn{
		in:     make(chan map[string]any, 8),
		inErr:  make(chan error, 1),
		outCh:  make(chan map[string]any, 8),
		closed: make(chan struct{}, 1),
	}
}

func (f *fakeConn) script(msg map[string]any) { f.in <- msg }
func (f *fakeConn) closeStream()              { f.inErr <- wire.ErrConnectionClosed }

func (f *fakeConn) ReadMessage() (map[string]any, error) {
	select {
	case m := <-f.in:
		return m, nil
	case err := <-f.inErr:
		return nil, err
	}
}

func (f *fakeConn) WriteMessage(msg map[string]any) error {
	f.outCh <- msg
	return nil
}

func (f *fakeConn) Close() error {
	select {
	case f.closed <- struct{}{}:
	default:
	}
	return nil
}

// out drains every message written so far, without blocking.
func (f *fakeConn) out() []map[string]any {
	var got []map[string]any
	for {
		select {
		case m := <-f.outCh:
			got = append(got, m)
		default:
			return got
		}
	}
}

// run drives HandleConnection with a throwaway Ingester — every Phase 2
// test here is about session negotiation, not ingest bookkeeping, so each
// gets its own fresh, unshared one.
func run(conn Conn) <-chan error {
	return runWithIngester(conn, NewIngester())
}

func runWithIngester(conn Conn, ing *Ingester) <-chan error {
	done := make(chan error, 1)
	go func() { done <- HandleConnection(conn, ing) }()
	return done
}

func sessionOpen(body string, extra map[string]any) map[string]any {
	msg := map[string]any{
		"protocol":   "0.1",
		"message":    "session_open",
		"session":    "s-1",
		"seq":        int64(0),
		"body":       body,
		"world_time": int64(1000),
		"payload": map[string]any{
			"time_unit":    "second",
			"capabilities": map[string]any{"percept_types": []any{"sighting"}},
		},
	}
	for k, v := range extra {
		msg[k] = v
	}
	return msg
}

func waitDone(t *testing.T, done <-chan error) error {
	t.Helper()
	select {
	case err := <-done:
		return err
	case <-time.After(2 * time.Second):
		t.Fatal("HandleConnection did not return")
		return nil
	}
}

// TestHandleConnection_FirstMessageMustBeSessionOpen proves the connection
// is refused when the vendor's first frame is anything but session_open
// (seam-wire-v0.md §1.4).
func TestHandleConnection_FirstMessageMustBeSessionOpen(t *testing.T) {
	conn := newFakeConn()
	done := run(conn)
	conn.script(map[string]any{
		"protocol": "0.1", "message": "session_close", "session": "s-1",
		"seq": int64(0), "body": "b-1", "world_time": int64(1),
		"payload": map[string]any{"reason": "shutdown", "detail": nil},
	})
	if err := waitDone(t, done); err == nil {
		t.Fatal("expected an error when the first message is not session_open")
	}
}

// TestNegotiation_UnsupportedVersion_RefusesFailClosed proves version
// negotiation fails closed per seam-wire-v0.md §5.2: an unsupported MAJOR
// gets exactly one session_close/reason:error/detail:unsupported_version,
// echoing session/body/world_time/protocol from the refused frame, seq 0.
func TestNegotiation_UnsupportedVersion_RefusesFailClosed(t *testing.T) {
	conn := newFakeConn()
	done := run(conn)
	conn.script(sessionOpen("b-eda", map[string]any{"protocol": "9.0"}))

	if err := waitDone(t, done); err == nil {
		t.Fatal("expected an error for an unsupported protocol version")
	}
	out := conn.out()
	if len(out) != 1 {
		t.Fatalf("got %d outbound messages, want exactly 1 refusal: %#v", len(out), out)
	}
	reply := out[0]
	if reply["message"] != "session_close" {
		t.Fatalf("reply message = %v, want session_close", reply["message"])
	}
	if reply["session"] != "s-1" || reply["body"] != "b-eda" || reply["protocol"] != "9.0" {
		t.Fatalf("reply did not echo session/body/protocol: %#v", reply)
	}
	if reply["world_time"] != int64(1000) {
		t.Fatalf("reply did not echo world_time: %#v", reply)
	}
	if reply["seq"] != int64(0) {
		t.Fatalf("reply seq = %v, want 0", reply["seq"])
	}
	pl := reply["payload"].(map[string]any)
	if pl["reason"] != "error" || pl["detail"] != "unsupported_version" {
		t.Fatalf("reply payload = %#v, want reason:error detail:unsupported_version", pl)
	}
}

// TestSessionOpen_ArchivedMind_RefusedOnFirstOpen is specs/018-consolidation
// T005 / ruling R-9: the very first session_open on a connection is
// refused, fail-closed, when Ingester.Archived reports the named body is
// an archived mind — the same session_close/reason:error shape the
// version and manifest refusals already use.
func TestSessionOpen_ArchivedMind_RefusedOnFirstOpen(t *testing.T) {
	ing := NewIngester()
	ing.Archived = func(body string) bool { return body == "b-dead" }
	conn := newFakeConn()
	done := runWithIngester(conn, ing)
	conn.script(sessionOpen("b-dead", nil))

	if err := waitDone(t, done); err == nil {
		t.Fatal("expected an error for an archived mind's session_open")
	}
	out := conn.out()
	if len(out) != 1 || out[0]["message"] != "session_close" {
		t.Fatalf("expected exactly one session_close refusal, got %#v", out)
	}
	pl := out[0]["payload"].(map[string]any)
	if pl["reason"] != "error" || pl["detail"] != "archived_mind" {
		t.Fatalf("refusal payload = %#v, want reason:error detail:archived_mind", pl)
	}
}

// TestSessionOpen_ArchivedMind_RefusedOnMultiplex proves the same refusal
// fires for a later, multiplexed session_open on an already-open
// connection — an archived body can never attach, first frame or not.
func TestSessionOpen_ArchivedMind_RefusedOnMultiplex(t *testing.T) {
	ing := NewIngester()
	ing.Archived = func(body string) bool { return body == "b-dead" }
	conn := newFakeConn()
	done := runWithIngester(conn, ing)
	conn.script(sessionOpen("b-1", nil))
	conn.script(sessionOpen("b-dead", nil))

	if err := waitDone(t, done); err == nil {
		t.Fatal("expected an error for an archived mind's multiplexed session_open")
	}
	out := conn.out()
	if len(out) != 1 || out[0]["message"] != "session_close" {
		t.Fatalf("expected exactly one session_close refusal, got %#v", out)
	}
	pl := out[0]["payload"].(map[string]any)
	if pl["detail"] != "archived_mind" {
		t.Fatalf("refusal detail = %v, want archived_mind", pl["detail"])
	}
}

// TestSessionOpen_NotArchived_Unaffected proves a nil Archived hook (the
// zero-value Ingester) and a non-archived body both behave exactly as
// before T005 — archival's absence changes nothing.
func TestSessionOpen_NotArchived_Unaffected(t *testing.T) {
	conn := newFakeConn()
	done := run(conn) // NewIngester()'s zero-value Archived is nil
	conn.script(sessionOpen("b-1", nil))
	conn.closeStream()

	if err := waitDone(t, done); err != nil {
		t.Fatalf("HandleConnection returned an error with no Archived hook set: %v", err)
	}
	if out := conn.out(); len(out) != 0 {
		t.Fatalf("expected no refusal, got %#v", out)
	}
}

// TestSessionOpen_Multiplex_SameManifestAttachesSecondBody proves several
// bodies multiplex onto one connection when every session_open carries the
// same session/time_unit/capabilities (seam-wire-v0.md §1.3).
func TestSessionOpen_Multiplex_SameManifestAttachesSecondBody(t *testing.T) {
	conn := newFakeConn()
	done := run(conn)
	conn.script(sessionOpen("b-1", nil))
	conn.script(sessionOpen("b-2", nil))
	conn.closeStream()

	if err := waitDone(t, done); err != nil {
		t.Fatalf("HandleConnection returned an error for a matching second body: %v", err)
	}
	if out := conn.out(); len(out) != 0 {
		t.Fatalf("expected no refusal, got %#v", out)
	}
}

// TestSessionOpen_DifferingCapabilities_Refused proves a later session_open
// on the same connection with capabilities that differ even slightly from
// the first is refused rather than silently accepted (seam-wire-v0.md §1.3,
// W-8's mechanically-checkable manifest).
func TestSessionOpen_DifferingCapabilities_Refused(t *testing.T) {
	conn := newFakeConn()
	done := run(conn)
	conn.script(sessionOpen("b-1", nil))
	conn.script(sessionOpen("b-2", map[string]any{
		"payload": map[string]any{
			"time_unit":    "second",
			"capabilities": map[string]any{"percept_types": []any{"sighting", "sound"}},
		},
	}))

	if err := waitDone(t, done); err == nil {
		t.Fatal("expected an error for a differing capabilities manifest")
	}
	out := conn.out()
	if len(out) != 1 || out[0]["message"] != "session_close" {
		t.Fatalf("expected exactly one session_close refusal, got %#v", out)
	}
	pl := out[0]["payload"].(map[string]any)
	if pl["detail"] != "differing_capabilities" {
		t.Fatalf("refusal detail = %v, want differing_capabilities", pl["detail"])
	}
}

// TestSessionClose_DetachesBody proves session_close for one body on a
// multi-body connection is accepted and does not end the connection.
func TestSessionClose_DetachesBody(t *testing.T) {
	conn := newFakeConn()
	done := run(conn)
	conn.script(sessionOpen("b-1", nil))
	conn.script(sessionOpen("b-2", nil))
	conn.script(map[string]any{
		"protocol": "0.1", "message": "session_close", "session": "s-1",
		"seq": int64(1), "body": "b-1", "world_time": int64(1001),
		"payload": map[string]any{"reason": "shutdown", "detail": nil},
	})
	conn.closeStream()

	if err := waitDone(t, done); err != nil {
		t.Fatalf("HandleConnection returned an error after a body's session_close: %v", err)
	}
	if out := conn.out(); len(out) != 0 {
		t.Fatalf("expected no refusal after an ordinary session_close, got %#v", out)
	}
}

// TestContinuity_ReconnectAfterRestart_NoBackfillAndBodyTokenMatch proves
// body-protocol-v0.md §6.3 / seam-wire-v0.md §1.5: a body reconnects across
// a daemon restart matched by its `body` token alone. Restart is modeled
// honestly here — HandleConnection is called fresh with a brand-new fakeConn
// and no state carried over from the first call, which is exactly what a
// real restart leaves behind at this layer (no durable memory exists until
// M2). The reconnect's session_open carries continuity.previous_session
// naming a session this "new daemon" has never heard of; the rejoin must
// still succeed (previous_session is a MAY-use check, never required), and
// nothing the mind sends may invent what happened in the gap.
func TestContinuity_ReconnectAfterRestart_NoBackfillAndBodyTokenMatch(t *testing.T) {
	firstConn := newFakeConn()
	firstDone := run(firstConn)
	firstConn.script(sessionOpen("b-eda", nil))
	firstConn.closeStream() // the daemon "restarts": this connection just ends
	if err := waitDone(t, firstDone); err != nil {
		t.Fatalf("first connection: %v", err)
	}

	// A brand-new HandleConnection call stands in for the restarted daemon:
	// nothing from firstConn's state is passed in.
	reconnect := newFakeConn()
	reconnectDone := run(reconnect)
	reconnect.script(sessionOpen("b-eda", map[string]any{
		"payload": map[string]any{
			"time_unit": "second",
			"capabilities": map[string]any{
				"percept_types": []any{"sighting"},
			},
			"continuity": map[string]any{
				"previous_session":          "s-unknown-to-this-daemon",
				"previous_close_world_time": int64(999),
				"body_continuous":           true,
			},
		},
	}))
	reconnect.closeStream()

	if err := waitDone(t, reconnectDone); err != nil {
		t.Fatalf("reconnect with continuity was refused: %v", err)
	}
	if out := reconnect.out(); len(out) != 0 {
		t.Fatalf("reconnect must not invent anything for the gap; mind sent %#v", out)
	}
}

// TestOnSessionOpen_FiresForFirstFrameAndMultiplex proves TASK-0023 T001/
// T002's manifest hook: it fires with the capabilities the frame actually
// declared for the connection's first session_open AND for a later
// multiplexed one — the only two places a body's declared verb set can be
// learned (mind/deliberate/loop.go's Config.Verbs doc).
func TestOnSessionOpen_FiresForFirstFrameAndMultiplex(t *testing.T) {
	conn := newFakeConn()
	ing := NewIngester()
	type call struct {
		body string
		caps map[string]any
	}
	calls := make(chan call, 2)
	ing.OnSessionOpen = func(c Conn, session, body string, capabilities map[string]any) {
		if c != conn {
			t.Errorf("OnSessionOpen conn = %v, want the connection itself", c)
		}
		if session != "s-1" {
			t.Errorf("OnSessionOpen session = %q, want s-1", session)
		}
		calls <- call{body, capabilities}
	}
	done := runWithIngester(conn, ing)
	conn.script(sessionOpen("b-1", nil))
	conn.script(sessionOpen("b-2", nil))

	// Wait for both hooks to fire BEFORE closing the stream: fakeConn's
	// ReadMessage selects between its in/inErr channels, so scripting a
	// close before HandleConnection has drained the queued frames would
	// race the close against them (Go's select picks pseudo-randomly
	// among ready cases).
	var got []call
	for len(got) < 2 {
		select {
		case c := <-calls:
			got = append(got, c)
		case <-time.After(2 * time.Second):
			t.Fatalf("OnSessionOpen fired %d times, want 2 (first frame + multiplex)", len(got))
		}
	}
	conn.closeStream()
	if err := waitDone(t, done); err != nil {
		t.Fatalf("HandleConnection: %v", err)
	}

	if got[0].body != "b-1" || got[1].body != "b-2" {
		t.Fatalf("OnSessionOpen bodies = %v, want [b-1 b-2]", got)
	}
	wantCaps := map[string]any{"percept_types": []any{"sighting"}}
	for _, g := range got {
		if !equalValue(g.caps, wantCaps) {
			t.Errorf("OnSessionOpen(%s) capabilities = %#v, want %#v", g.body, g.caps, wantCaps)
		}
	}
}
