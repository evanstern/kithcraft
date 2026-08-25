// Package wire implements the seam's framing and canonical JSON codec per
// docs/design/seam-wire-v0.md. Every protocol message crosses the socket as
// exactly one frame: a 4-byte big-endian length prefix followed by that many
// bytes of UTF-8 JSON (§2.1).
package wire

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// MaxFrameBody is the wire's size cap (§2.5): 1 MiB per frame body. It bounds
// a corrupt length word, not content — see §2.5's rationale. It is not a
// percept budget and must never be used as a shedding input.
const MaxFrameBody = 1 << 20

// ErrConnectionClosed indicates an orderly disconnect between frames: the
// stream ended while reading a fresh length header, with no bytes of it read
// yet. Per §2.3 this is a normal end-of-session, not a framing error.
var ErrConnectionClosed = errors.New("wire: connection closed between frames")

// ReadFrame reads exactly one frame from r (§2.1, §2.3): a 4-byte big-endian
// length prefix followed by that many body bytes, looping internally on
// short reads. The length is checked against MaxFrameBody before the body
// buffer is allocated, so an oversize or corrupt length word never causes a
// large allocation.
//
// Every returned error other than ErrConnectionClosed is connection-fatal
// per §2.3/§2.5: the caller MUST close the connection and MUST NOT attempt
// to resynchronize, because a stream whose framing is in doubt has every
// subsequent boundary in doubt too.
func ReadFrame(r io.Reader) ([]byte, error) {
	var hdr [4]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		if errors.Is(err, io.EOF) {
			// Zero bytes read: the stream ended between frames (§2.3).
			return nil, ErrConnectionClosed
		}
		return nil, fmt.Errorf("wire: truncated frame: reading length header: %w", err)
	}

	n := binary.BigEndian.Uint32(hdr[:])
	if n == 0 {
		return nil, fmt.Errorf("wire: malformed frame: length 0 (an empty body is not a JSON value)")
	}
	if n > MaxFrameBody {
		return nil, fmt.Errorf("wire: frame body %d exceeds the 1 MiB cap", n)
	}

	body := make([]byte, n)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, fmt.Errorf("wire: truncated frame: body declared %d bytes: %w", n, err)
	}
	return body, nil
}

// WriteFrame serializes one frame — a 4-byte big-endian length prefix plus
// body — into a single buffer and writes it to w in one call, per §4.2's
// whole-frame writer discipline: a frame is never split across writes, and a
// write failure mid-frame is connection-fatal (the caller must close).
//
// Callers MUST NOT call WriteFrame concurrently on the same w — §4.2
// requires exactly one writer per connection for the connection's whole
// life; enforcing that is the caller's responsibility, not this function's.
func WriteFrame(w io.Writer, body []byte) error {
	if len(body) == 0 {
		return fmt.Errorf("wire: malformed frame: length 0 (an empty body is not a JSON value)")
	}
	if len(body) > MaxFrameBody {
		return fmt.Errorf("wire: frame body %d exceeds the 1 MiB cap", len(body))
	}

	frame := make([]byte, 4+len(body))
	binary.BigEndian.PutUint32(frame, uint32(len(body)))
	copy(frame[4:], body)

	if _, err := w.Write(frame); err != nil {
		return fmt.Errorf("wire: write failed mid-frame (connection-fatal, §4.2): %w", err)
	}
	return nil
}
