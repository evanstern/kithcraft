package seam

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"kithcraft/mind/wire"
)

// SupportedMajor is the only protocol MAJOR this daemon speaks. Per
// seam-wire-v0.md §5.2, receivers compare only MAJOR after the handshake;
// v0's wire protocol string is "0.1" (body-protocol-v0.md §7.1).
const SupportedMajor = "0"

// RefusalLinger bounds how long the mind waits after writing a refusing
// session_close before it closes the connection, so the peer reads the
// frame instead of seeing a bare EOF (seam-wire-v0.md §1.4, §5.2).
const RefusalLinger = 50 * time.Millisecond

// manifest is what the FIRST session_open on a connection fixes
// (seam-wire-v0.md §1.3): every later session_open on the same connection
// must match it exactly — capabilities compared as canonical bytes so
// "byte-identical" is a real equality check, not a deep-equal guess.
type manifest struct {
	session   string
	timeUnit  any
	capsBytes []byte
}

// HandleConnection runs one connection's whole life. The vendor's first
// frame must be session_open (seam-wire-v0.md §1.4); version negotiation is
// fail-closed (§5.2); every later session_open on the connection either
// multiplexes a new body onto the same manifest or is refused
// (§1.3 — differing session/time_unit/capabilities). Continuity
// (body-protocol-v0.md §6.3, seam-wire-v0.md §1.5) needs no state here: a
// reconnecting body is matched by its `body` token alone — the same as any
// other session_open — and previous_session is never consulted, so a
// reconnect after this daemon lost all memory of the prior connection still
// succeeds. Durable percept/body memory across restarts is M2's job, not
// this package's.
//
// Percept and intent handling (T009/T010) are Phase 3 scope: any message
// kind other than session_open/session_close is accepted and left
// unprocessed here.
func HandleConnection(conn Conn) error {
	defer conn.Close()

	first, err := conn.ReadMessage()
	if err != nil {
		if errors.Is(err, wire.ErrConnectionClosed) {
			return nil
		}
		return err
	}
	if kind := str(first["message"]); kind != "session_open" {
		// ponytail: no session_close is sent for this case (unlike the
		// version refusal below) — the protocol spells out a reply only
		// for unsupported_version (§7.1/§5.2). Add one if a vector or
		// acceptance test ever pins a reply here.
		return fmt.Errorf("seam: connection must speak session_open first, got %q", kind)
	}
	if major(str(first["protocol"])) != SupportedMajor {
		refuse(conn, first, "unsupported_version")
		return fmt.Errorf("seam: unsupported protocol version %q", first["protocol"])
	}

	pl := payloadOf(first)
	capsBytes, err := wire.EncodeCanonical(pl["capabilities"])
	if err != nil {
		return fmt.Errorf("seam: capabilities not representable: %w", err)
	}
	m := manifest{session: str(first["session"]), timeUnit: pl["time_unit"], capsBytes: capsBytes}
	attached := map[string]bool{str(first["body"]): true}

	for {
		msg, err := conn.ReadMessage()
		if err != nil {
			if errors.Is(err, wire.ErrConnectionClosed) {
				return nil // orderly disconnect between frames, §1.4/§2.3
			}
			return err // connection-fatal: framing or presence-validation failed
		}

		switch str(msg["message"]) {
		case "session_open":
			if mismatch := m.diff(msg); mismatch != "" {
				refuse(conn, msg, mismatch)
				return fmt.Errorf("seam: %s", mismatch)
			}
			attached[str(msg["body"])] = true
		case "session_close":
			delete(attached, str(msg["body"]))
		}
	}
}

// diff reports why msg's manifest disagrees with m as a refusal detail
// token, or "" if every field matches (seam-wire-v0.md §1.3).
func (m manifest) diff(msg map[string]any) string {
	if str(msg["session"]) != m.session {
		return "differing_session"
	}
	pl := payloadOf(msg)
	if !equalValue(pl["time_unit"], m.timeUnit) {
		return "differing_time_unit"
	}
	capsBytes, err := wire.EncodeCanonical(pl["capabilities"])
	if err != nil || !bytes.Equal(capsBytes, m.capsBytes) {
		return "differing_capabilities"
	}
	return ""
}

// refuse sends the mind's only uninvited message: a fail-closed refusal,
// session_close/reason:error (seam-wire-v0.md §5.2). It echoes protocol,
// session, body, and world_time from the refused frame, uses seq: 0 (the
// mind's first message in that direction), and lingers before the caller
// closes the connection so the peer reads it rather than seeing a bare EOF.
func refuse(conn Conn, refused map[string]any, detail string) {
	reply := map[string]any{
		"protocol":   refused["protocol"],
		"message":    "session_close",
		"session":    refused["session"],
		"seq":        int64(0),
		"body":       refused["body"],
		"world_time": refused["world_time"],
		"payload": map[string]any{
			"reason": "error",
			"detail": detail,
		},
	}
	_ = conn.WriteMessage(reply) // best-effort: a write failure is already connection-fatal
	time.Sleep(RefusalLinger)
}

func str(v any) string {
	s, _ := v.(string)
	return s
}

func payloadOf(msg map[string]any) map[string]any {
	p, _ := msg["payload"].(map[string]any)
	return p
}

func major(protocol string) string {
	if i := strings.IndexByte(protocol, '.'); i >= 0 {
		return protocol[:i]
	}
	return protocol
}

// equalValue compares two decoded JSON values by canonical bytes — reusing
// EncodeCanonical rather than a hand-rolled deep-equal, since it already
// defines exactly the equality this wire cares about.
func equalValue(a, b any) bool {
	ab, aerr := wire.EncodeCanonical(a)
	bb, berr := wire.EncodeCanonical(b)
	return aerr == nil && berr == nil && bytes.Equal(ab, bb)
}
