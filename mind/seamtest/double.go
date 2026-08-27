// Package seamtest is the minimal in-test double for mind/seam's vendor
// port: it drives seam.Conn from the OUTSIDE, the vendor's role, so
// Phase 3's tests exercise the real wire and session layers rather than a
// hand-rolled channel fake. It scripts a percept stream (including
// duplicates and a seq gap) and records the intents the daemon emits.
//
// Deliberately minimal (T011): mind/fakevendor (S2, TASK-0015) is its
// full-shape successor — session lifecycle, acks, resolvable acts, and
// the H-test harness — for tests that need more than a percept-in/
// intent-out recording double. No general read API and no autonomous
// behavior — the double does only what a test tells it to, plus the
// passive recording a live socket requires (there is no non-blocking peek
// on a net.Conn).
package seamtest

import (
	"net"
	"sync"

	"kithcraft/mind/seam"
)

// Double is one vendor connection's test double.
type Double struct {
	conn seam.Conn

	mu      sync.Mutex
	intents []map[string]any
	readErr error
	done    chan struct{}
}

// DialUnix dials the daemon's real UDS listener at path (or a net.Pipe
// half wrapped the same way — both satisfy T-7) and starts recording
// every message it sends, in arrival order.
func DialUnix(path string) (*Double, error) {
	c, err := net.Dial("unix", path)
	if err != nil {
		return nil, err
	}
	return NewDouble(seam.NewWireConn(c)), nil
}

// NewDouble wraps an already-established seam.Conn (e.g. one half of a
// net.Pipe) as a Double.
func NewDouble(conn seam.Conn) *Double {
	d := &Double{conn: conn, done: make(chan struct{})}
	go d.record()
	return d
}

func (d *Double) record() {
	defer close(d.done)
	for {
		msg, err := d.conn.ReadMessage()
		if err != nil {
			d.mu.Lock()
			d.readErr = err
			d.mu.Unlock()
			return
		}
		if msg["message"] == "intent" {
			d.mu.Lock()
			d.intents = append(d.intents, msg)
			d.mu.Unlock()
		}
	}
}

// Intents returns every intent recorded so far, in arrival order.
func (d *Double) Intents() []map[string]any {
	d.mu.Lock()
	defer d.mu.Unlock()
	out := make([]map[string]any, len(d.intents))
	copy(out, d.intents)
	return out
}

// ReadErr reports the error (if any) that ended the recording loop — a
// connection close or a connection-fatal refusal from the daemon.
func (d *Double) ReadErr() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.readErr
}

// Send writes one message as-is — the caller composes the full envelope
// (session_open, percept, intent_ack, act_result, cancel, …); the double
// adds no protocol logic of its own. Use the builder functions below for
// the common envelopes.
func (d *Double) Send(msg map[string]any) error { return d.conn.WriteMessage(msg) }

// Close ends the connection.
func (d *Double) Close() error { return d.conn.Close() }

// SessionOpen builds a session_open envelope for one body attaching to
// session (body-protocol-v0.md §6.2, seam-wire-v0.md §1.3). continuity
// may be nil.
func SessionOpen(session, body, timeUnit string, capabilities, continuity map[string]any) map[string]any {
	payload := map[string]any{"time_unit": timeUnit, "capabilities": capabilities}
	if continuity != nil {
		payload["continuity"] = continuity
	}
	return map[string]any{
		"protocol": "0.1", "message": "session_open", "session": session,
		"seq": int64(0), "body": body, "world_time": int64(0), "payload": payload,
	}
}

// Percept builds a percept envelope (body-protocol-v0.md §4.1). seq is
// the caller's to choose — scripting a repeat or a gap is the point.
func Percept(session, body string, seq, worldTime int64, percept map[string]any) map[string]any {
	return map[string]any{
		"protocol": "0.1", "message": "percept", "session": session, "seq": seq,
		"body": body, "world_time": worldTime, "payload": percept,
	}
}

// IntentAck builds an intent_ack envelope (§5.3).
func IntentAck(session, body string, seq int64, intentID string, accepted bool) map[string]any {
	return map[string]any{
		"protocol": "0.1", "message": "intent_ack", "session": session, "seq": seq,
		"body": body, "world_time": int64(0),
		"payload": map[string]any{"intent_id": intentID, "accepted": accepted, "reason_code": nil},
	}
}
