package main

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"kithcraft/mind/seam"
	"kithcraft/mind/wire"
)

// shortSockDir returns a fresh temp directory short enough to hold a socket
// path under sun_path's 103-byte limit. t.TempDir() embeds the test's name
// in the path (seam-wire-v0.md §1.1 warns exactly against a deep,
// name-derived path) and reliably overflows that limit on this host, so
// tests needing a real UDS socket use this instead.
func shortSockDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "mindtest")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

// TestListen_UnlinksStaleSocket proves decision-0004's liveness probe: a
// leftover file at the socket path (a crash, not a graceful shutdown — Go's
// net only unlinks on a listener it owns closing) is removed and bound to,
// because a dial against it fails.
func TestListen_UnlinksStaleSocket(t *testing.T) {
	path := filepath.Join(shortSockDir(t), "mind.sock")
	if err := os.WriteFile(path, []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}
	ln, err := listen(path)
	if err != nil {
		t.Fatalf("listen over a stale socket file: %v", err)
	}
	defer ln.Close()
}

// TestListen_RefusesWhenAnotherDaemonIsLive proves the other half of the
// probe: a path whose dial SUCCEEDS is a live daemon, and must not be
// stolen — the second listen must fail loudly instead of stealing it.
func TestListen_RefusesWhenAnotherDaemonIsLive(t *testing.T) {
	path := filepath.Join(shortSockDir(t), "mind.sock")
	ln1, err := listen(path)
	if err != nil {
		t.Fatalf("first listen: %v", err)
	}
	defer ln1.Close()

	if _, err := listen(path); err == nil {
		t.Fatal("expected listen to refuse a socket path another daemon holds live")
	}
}

// TestListen_RejectsOversizePath proves the sun_path length check fails
// loudly rather than truncating (decision-0004, seam-wire-v0.md §1.1).
func TestListen_RejectsOversizePath(t *testing.T) {
	long := filepath.Join(shortSockDir(t), strings.Repeat("a", maxSocketPathBytes)+".sock")
	if _, err := listen(long); err == nil {
		t.Fatal("expected listen to reject an oversize socket path")
	}
}

// TestServe_AcceptsAndNegotiates proves the accept loop and per-connection
// framing end to end over a real UDS socket: a dialed connection speaking
// an unsupported protocol version gets exactly the fail-closed refusal
// mind/seam's session layer defines.
func TestServe_AcceptsAndNegotiates(t *testing.T) {
	path := filepath.Join(shortSockDir(t), "mind.sock")
	ln, err := listen(path)
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go serve(ln, seam.NewIngester())

	conn, err := net.Dial("unix", path)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	msg := map[string]any{
		"protocol": "9.0", "message": "session_open", "session": "s-1",
		"seq": int64(0), "body": "b-1", "world_time": int64(42),
		"payload": map[string]any{
			"time_unit":    "second",
			"capabilities": map[string]any{},
		},
	}
	body, err := wire.EncodeCanonical(msg)
	if err != nil {
		t.Fatalf("EncodeCanonical: %v", err)
	}
	if err := wire.WriteFrame(conn, body); err != nil {
		t.Fatalf("WriteFrame: %v", err)
	}

	respBody, err := wire.ReadFrame(conn)
	if err != nil {
		t.Fatalf("ReadFrame: %v", err)
	}
	resp, err := wire.Decode(respBody)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	reply := resp.(map[string]any)
	if reply["message"] != "session_close" {
		t.Fatalf("reply message = %v, want session_close", reply["message"])
	}
	pl := reply["payload"].(map[string]any)
	if pl["detail"] != "unsupported_version" {
		t.Fatalf("reply detail = %v, want unsupported_version", pl["detail"])
	}
}
