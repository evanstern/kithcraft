package seam

import (
	"net"
	"testing"
	"time"

	"kithcraft/mind/wire"
)

func directProvenance(origin string) map[string]any {
	return map[string]any{"origin": origin, "source": nil, "observed_at": int64(1), "received_at": int64(1)}
}

// percept builds a well-formed percept envelope; a nil provenance omits
// the key entirely, so callers can build V-5's "missing provenance" case.
func percept(body string, seq int64, id string, provenance map[string]any) map[string]any {
	payload := map[string]any{
		"percept_id": id, "percept_type": "sighting", "urgency": "background",
		"place": nil, "content": map[string]any{},
	}
	if provenance != nil {
		payload["provenance"] = provenance
	}
	return map[string]any{
		"protocol": "0.1", "message": "percept", "session": "s-1",
		"seq": seq, "body": body, "world_time": int64(1000 + seq),
		"payload": payload,
	}
}

// TestClassifyOrigin proves §2.7's classifier: the four direct origins
// are firsthand; everything else — including an origin from a future
// minor version this daemon has never heard of, and an absent one — is
// secondhand (EH-2b, card AC #4).
func TestClassifyOrigin(t *testing.T) {
	for _, direct := range []string{"acted", "saw", "heard", "felt"} {
		if got := ClassifyOrigin(direct); got != "firsthand" {
			t.Errorf("ClassifyOrigin(%q) = %q, want firsthand", direct, got)
		}
	}
	for _, indirect := range []string{"told", "read", "glimpsed" /* future minor version */, ""} {
		if got := ClassifyOrigin(indirect); got != "secondhand" {
			t.Errorf("ClassifyOrigin(%q) = %q, want secondhand", indirect, got)
		}
	}
}

// TestIngest_DuplicatePerceptID_DroppedAfterReconnect proves §3.4: within
// a session the stream cannot duplicate, but a percept_id reappearing
// after Attach resets the seq counter (a reconnect) is a retransmission
// and dedup drops it.
func TestIngest_DuplicatePerceptID_DroppedAfterReconnect(t *testing.T) {
	ing := NewIngester()
	ing.Attach("b-1", 0)
	if ing.Dedup("b-1", "p-1") {
		t.Fatal("first sighting of p-1 must not be a duplicate")
	}
	ing.Attach("b-1", 0) // reconnect: fresh seq counter, ledger preserved
	if !ing.Dedup("b-1", "p-1") {
		t.Fatal("p-1 repeated after a reconnect must be deduped")
	}
}

// TestIngest_SeqGap_RecordsShedCount proves §3.3: a gap in the per-body
// seq stream is recorded as the count of background percepts it implies
// were shed, never treated as an error.
func TestIngest_SeqGap_RecordsShedCount(t *testing.T) {
	ing := NewIngester()
	ing.Attach("b-1", 0)
	if shed := ing.Observe("b-1", 1); shed != 0 {
		t.Fatalf("consecutive seq must not be a gap, got shed=%d", shed)
	}
	if shed := ing.Observe("b-1", 4); shed != 2 { // seq 2,3 shed
		t.Fatalf("Observe seq 4 after 1 = shed %d, want 2", shed)
	}
	if got := ing.ShedCount("b-1"); got != 2 {
		t.Fatalf("cumulative ShedCount = %d, want 2", got)
	}
}

// TestIngest_MalformedPercept_MutatesNothing proves card AC #3 against the
// real wire, not a shortcut fake: a percept missing provenance fails
// wire.Validate inside wireConn.ReadMessage before HandleConnection's
// switch ever sees it (V-5), so the Ingester's state — dedup ledger, shed
// count, and OnPercept — is exactly as it was before the malformed frame
// arrived.
func TestIngest_MalformedPercept_MutatesNothing(t *testing.T) {
	ing := NewIngester()
	onPercept := 0
	ing.OnPercept = func(Conn, string, map[string]any) { onPercept++ }

	a, b := net.Pipe()
	defer a.Close()
	done := runWithIngester(NewWireConn(b), ing)

	write := func(msg map[string]any) {
		body, err := wire.EncodeCanonical(msg)
		if err != nil {
			t.Fatalf("EncodeCanonical: %v", err)
		}
		if err := wire.WriteFrame(a, body); err != nil {
			t.Fatalf("WriteFrame: %v", err)
		}
	}

	write(sessionOpen("b-1", nil))
	write(percept("b-1", 1, "p-1", directProvenance("saw")))
	malformed := percept("b-1", 2, "p-2", nil) // missing provenance, V-5
	write(malformed)

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected HandleConnection to end with an error after the malformed percept")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("HandleConnection did not return")
	}

	if onPercept != 1 {
		t.Fatalf("OnPercept ran %d times, want exactly 1 (only for the valid percept)", onPercept)
	}
	if got := ing.ShedCount("b-1"); got != 0 {
		t.Fatalf("ShedCount = %d, want 0: the malformed percept must not touch seq accounting", got)
	}
	if ing.Dedup("b-1", "p-2") {
		t.Fatal("p-2 must not be in the dedup ledger: the malformed percept never reached ingest")
	}
}
