package seam

import (
	"errors"
	"net"
	"reflect"
	"testing"

	"kithcraft/mind/wire"
)

func sampleSessionOpen() map[string]any {
	return map[string]any{
		"protocol":   "0.1",
		"message":    "session_open",
		"session":    "s-7f3a",
		"seq":        int64(0),
		"body":       "b-eda",
		"world_time": int64(918233),
		"payload": map[string]any{
			"time_unit":    "second",
			"capabilities": map[string]any{"percept_types": []any{"sighting"}},
		},
	}
}

// TestWireConn_RoundTrip proves NewWireConn's WriteMessage/ReadMessage round
// a message byte-exactly through mind/wire's framing and codec — the real
// UDS listener's transport, exercised here over an in-process net.Pipe so
// the test needs no socket.
func TestWireConn_RoundTrip(t *testing.T) {
	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()
	ca, cb := NewWireConn(a), NewWireConn(b)

	msg := sampleSessionOpen()
	errCh := make(chan error, 1)
	go func() { errCh <- ca.WriteMessage(msg) }()

	got, err := cb.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("WriteMessage: %v", err)
	}
	if !reflect.DeepEqual(got, msg) {
		t.Fatalf("round-trip mismatch:\n got  %#v\n want %#v", got, msg)
	}
}

// TestWireConn_ReadMessage_RefusesMalformed proves ReadMessage surfaces
// wire.Validate's presence-checked refusal (V-5) rather than handing the
// session layer a message missing a required field.
func TestWireConn_ReadMessage_RefusesMalformed(t *testing.T) {
	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()

	msg := sampleSessionOpen()
	delete(msg, "world_time") // required top-level field, V-5
	body, err := wire.EncodeCanonical(msg)
	if err != nil {
		t.Fatalf("EncodeCanonical: %v (bad test fixture)", err)
	}
	go wire.WriteFrame(a, body)

	if _, err := NewWireConn(b).ReadMessage(); err == nil {
		t.Fatal("expected ReadMessage to refuse a message missing a required field")
	}
}

// TestWireConn_ReadMessage_ConnectionClosed proves an orderly disconnect
// between frames surfaces as wire.ErrConnectionClosed, not a generic error.
func TestWireConn_ReadMessage_ConnectionClosed(t *testing.T) {
	a, b := net.Pipe()
	defer b.Close()
	a.Close()

	_, err := NewWireConn(b).ReadMessage()
	if !errors.Is(err, wire.ErrConnectionClosed) {
		t.Fatalf("got %v, want wire.ErrConnectionClosed", err)
	}
}
