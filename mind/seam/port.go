// Package seam implements the mind daemon's session lifecycle: fail-closed
// version negotiation, session_open/session_close, per-body multiplexing on
// one connection, and continuity across a reconnect
// (docs/design/body-protocol-v0.md §6, docs/design/seam-wire-v0.md §1/§5.2).
package seam

import (
	"errors"
	"io"

	"kithcraft/mind/wire"
)

// Conn is the vendor connection port: a framed message transport a vendor
// connection provides, declared here at the consumer (FR-001, card AC #1).
// The real UDS listener wraps each accepted connection as a Conn via
// NewWireConn; an in-test double may do the same over net.Pipe or implement
// Conn directly in memory. Nothing above this package touches a socket or
// mind/wire's byte-level API.
type Conn interface {
	// ReadMessage blocks for the next decoded, presence-validated message.
	// wire.ErrConnectionClosed means an orderly disconnect between frames
	// (seam-wire-v0.md §1.4/§2.3); every other error is connection-fatal.
	ReadMessage() (map[string]any, error)
	// WriteMessage canonically encodes and frames one message.
	WriteMessage(msg map[string]any) error
	Close() error
}

// wireConn adapts any byte stream (a UDS net.Conn, a net.Pipe half, …) to
// Conn using mind/wire's framing and canonical codec — decision-0004's
// "framing is defined over an abstract stream" applied at the seam layer.
type wireConn struct {
	rw io.ReadWriteCloser
}

// NewWireConn wraps rw as a Conn.
func NewWireConn(rw io.ReadWriteCloser) Conn {
	return &wireConn{rw: rw}
}

func (c *wireConn) ReadMessage() (map[string]any, error) {
	body, err := wire.ReadFrame(c.rw)
	if err != nil {
		return nil, err
	}
	v, err := wire.Decode(body)
	if err != nil {
		return nil, err
	}
	if err := wire.Validate(v); err != nil {
		return nil, err
	}
	msg, ok := v.(map[string]any)
	if !ok {
		return nil, errors.New("seam: decoded message is not a JSON object")
	}
	return msg, nil
}

func (c *wireConn) WriteMessage(msg map[string]any) error {
	body, err := wire.EncodeCanonical(msg)
	if err != nil {
		return err
	}
	return wire.WriteFrame(c.rw, body)
}

func (c *wireConn) Close() error { return c.rw.Close() }
