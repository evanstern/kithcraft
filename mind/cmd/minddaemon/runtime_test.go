package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/anthropics/anthropic-sdk-go/option"

	"kithcraft/mind/consolidate"
	"kithcraft/mind/llm"
	"kithcraft/mind/persona"
	"kithcraft/mind/prompt"
	"kithcraft/mind/seam"
	"kithcraft/mind/seamtest"
)

// TestNewRuntime_OpensArchiveUnderRunDir proves T001's store-opening half:
// NewRuntime creates the villager dir and an empty, ready Archive without
// any persona or model involvement.
func TestNewRuntime_OpensArchiveUnderRunDir(t *testing.T) {
	rt, err := NewRuntime(t.TempDir())
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}
	defer rt.Close()
	if rt.Archive.IsArchived("nobody") {
		t.Error("fresh archive reports a body archived")
	}
	if rt.Client != nil || rt.Digester != nil {
		t.Error("Client/Digester should be nil without ANTHROPIC_API_KEY (rehearsal path)")
	}
}

// TestLoadOrGenesisCast_NoKeyAndMissingPersonas_ClearFailure is spec.md's
// edge case: genesis needed, no ANTHROPIC_API_KEY, so startup must fail
// loudly naming the missing cast ids rather than binding a partial cast.
// Zero-call: no httptest server is even started.
func TestLoadOrGenesisCast_NoKeyAndMissingPersonas_ClearFailure(t *testing.T) {
	rt, err := NewRuntime(t.TempDir())
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}
	defer rt.Close()

	err = rt.LoadOrGenesisCast(context.Background())
	if err == nil {
		t.Fatal("expected a clear failure with no ANTHROPIC_API_KEY and no persisted personas")
	}
	for _, e := range persona.DemoCast() {
		if !strings.Contains(err.Error(), e.CastID) {
			t.Errorf("failure %q does not name missing cast id %q", err.Error(), e.CastID)
		}
	}
	matches, _ := filepath.Glob(filepath.Join(rt.PersonaDir, "*.json"))
	if len(matches) != 0 {
		t.Errorf("no partial cast should ever be written to disk, found %v", matches)
	}
}

// genesisServer is TestLoadOrGenesisCast_Resumes...'s mock E1 endpoint —
// mirrors mind/persona/genesis_test.go's own fake so this package's tests
// never touch the network either.
func genesisServer(t *testing.T) (*httptest.Server, *int32) {
	t.Helper()
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		out, _ := json.Marshal(struct {
			Name              string   `json:"name"`
			Values            []string `json:"values"`
			EndogenousDesires []string `json:"endogenous_desires"`
			Anchor            string   `json:"anchor"`
			DriftMarkers      []string `json:"drift_markers"`
		}{
			Name: "Generated", Values: []string{"craft"}, EndogenousDesires: []string{"build things"},
			Anchor: "steady and exacting", DriftMarkers: []string{"reckless"},
		})
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": "msg_test", "type": "message", "role": "assistant", "model": "claude-test",
			"content":     []map[string]any{{"type": "text", "text": string(out)}},
			"stop_reason": "end_turn",
			"usage":       map[string]any{"input_tokens": 10, "output_tokens": 10},
		})
	}))
	t.Cleanup(srv.Close)
	return srv, &hits
}

// TestLoadOrGenesisCast_ResumesPartialGenesis is T001's other edge case:
// two of three personas already exist (an interrupted prior run's 0444
// files); genesis must run for the missing one only, and existing files
// are re-bound via Load, never touched by WriteOnce again.
func TestLoadOrGenesisCast_ResumesPartialGenesis(t *testing.T) {
	dir := t.TempDir()
	personaDir := filepath.Join(dir, "persona")
	for _, id := range []string{"Aldric", "Petra"} {
		if err := persona.WriteOnce(personaDir, persona.Persona{CastID: id, Name: id, Anchor: "already here"}); err != nil {
			t.Fatalf("seeding %s: %v", id, err)
		}
	}

	srv, hits := genesisServer(t)
	rt, err := NewRuntime(dir)
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}
	defer rt.Close()
	rt.Client = llm.New(option.WithBaseURL(srv.URL), option.WithAPIKey("test-key"))
	rt.Digester = consolidate.ClientDigester{Client: rt.Client}

	if err := rt.LoadOrGenesisCast(context.Background()); err != nil {
		t.Fatalf("LoadOrGenesisCast: %v", err)
	}
	if got := atomic.LoadInt32(hits); got != 1 {
		t.Fatalf("E1 hits = %d, want 1 (genesis only for the missing Yenna)", got)
	}
	if len(rt.Personas) != 3 {
		t.Fatalf("Personas has %d entries, want 3", len(rt.Personas))
	}
	if rt.Personas["Aldric"].Anchor != "already here" {
		t.Error("the pre-seeded Aldric was not re-bound via Load (genesis must not have overwritten it)")
	}
}

// scriptedDigester mirrors mind/consolidate's own test double: a fixed
// sequence of Digest results, no SDK, no network — the mock-model test
// T002 asks for, driven through the real listener/Ingester/HandlePercept
// path rather than calling RunNight directly.
type scriptedDigester struct {
	n       int
	replies []struct {
		raw string
		err error
	}
}

func (d *scriptedDigester) Digest(ctx context.Context, a prompt.Assembled) (string, bool, error) {
	r := d.replies[d.n]
	d.n++
	return r.raw, false, r.err
}

// notablePercept is admitted by mind/memory.Gate unconditionally
// (RuleUrgency) — unlike e2e_test.go's sightingPercept (background, admits
// nothing), this test needs a non-empty buffer at each crossing so RunNight
// actually calls Digest instead of taking the empty-night shortcut.
func notablePercept(id string) map[string]any {
	return map[string]any{
		"percept_id": id, "percept_type": "sighting", "urgency": "notable",
		"provenance": map[string]any{"origin": "saw", "source": nil, "observed_at": int64(1), "received_at": int64(1)},
		"place":      nil, "content": map[string]any{},
	}
}

// TestSleepTriggerEndToEnd_NoMarkerOnFailureThenRetrySucceeds is T002's
// card proof: a real listener drives percepts through Runtime.HandlePercept
// (wired as Ingester.OnPercept), crossing one cycle boundary whose E6 call
// fails (no ledger record lands — the buffer is not lost) and a second
// boundary whose call succeeds, landing one record that covers both
// windows — the no-marker-on-failure retry semantics, reachable end to end
// through the daemon loop rather than by calling consolidate.RunNight
// directly.
func TestSleepTriggerEndToEnd_NoMarkerOnFailureThenRetrySucceeds(t *testing.T) {
	rt, err := NewRuntime(t.TempDir())
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}
	defer rt.Close()
	rt.Digester = &scriptedDigester{replies: []struct {
		raw string
		err error
	}{
		{err: errors.New("transport failure, first boundary")},
		{raw: `{"summary":"a quiet day","references":[]}`},
	}}

	ing := seam.NewIngester()
	ing.OnPercept = rt.HandlePercept
	ing.Archived = rt.Archive.IsArchived

	path := filepath.Join(shortSockDir(t), "mind.sock")
	ln, err := listen(path)
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go serve(ln, ing)

	dbl, err := seamtest.DialUnix(path)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer dbl.Close()
	if err := dbl.Send(seamtest.SessionOpen("s-1", "b-sleep", "second", testCapabilities, nil)); err != nil {
		t.Fatalf("session_open: %v", err)
	}

	send := func(seq, worldTime int64) {
		t.Helper()
		p := notablePercept(fmt.Sprintf("p-%d", seq))
		if err := dbl.Send(seamtest.Percept("s-1", "b-sleep", seq, worldTime, p)); err != nil {
			t.Fatalf("percept seq %d: %v", seq, err)
		}
	}
	recordCount := func() int {
		rt.mu.Lock()
		bs, ok := rt.bodies["b-sleep"]
		rt.mu.Unlock()
		if !ok {
			return -1
		}
		return len(bs.ledger.Records())
	}

	send(1, 100)                      // before any boundary
	send(2, consolidate.CycleTicks+1) // crosses boundary 1: digest fails, no marker
	waitUntil(t, 2*time.Second, func() bool { return recordCount() == 0 })
	send(3, 2*consolidate.CycleTicks+1) // crosses boundary 2: digest succeeds
	waitUntil(t, 2*time.Second, func() bool { return recordCount() == 1 })

	rt.mu.Lock()
	bs := rt.bodies["b-sleep"]
	rt.mu.Unlock()
	records := bs.ledger.Records()
	if got := records[0].WindowEnd; got != 2*consolidate.CycleTicks+1 {
		t.Errorf("landed record WindowEnd = %d, want %d (the retry covers the failed window too)", got, 2*consolidate.CycleTicks+1)
	}
}
