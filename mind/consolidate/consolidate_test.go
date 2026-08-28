package consolidate

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"kithcraft/mind/llm"
	"kithcraft/mind/memory"
	"kithcraft/mind/prompt"
)

// scriptedDigester replays a fixed sequence of Digest results — the
// mock/scripted model client this package's tests use in place of any
// live call (spec.md FR-007), mirroring mind/llm's own httptest scripting
// one layer up (no SDK involved at all here).
type scriptedDigester struct {
	calls   int
	replies []digestReply
}

type digestReply struct {
	raw       string
	truncated bool
	err       error
}

func (s *scriptedDigester) Digest(ctx context.Context, a prompt.Assembled) (string, bool, error) {
	r := s.replies[s.calls]
	s.calls++
	return r.raw, r.truncated, r.err
}

func openTestLog(t *testing.T) *memory.Log {
	t.Helper()
	log, err := memory.Open(memory.PathFor(t.TempDir(), "b-tam"))
	if err != nil {
		t.Fatalf("memory.Open: %v", err)
	}
	t.Cleanup(func() { log.Close() })
	return log
}

func openTestLedger(t *testing.T) *Ledger {
	t.Helper()
	l, err := OpenLedger(PathFor(t.TempDir(), "b-tam"))
	if err != nil {
		t.Fatalf("OpenLedger: %v", err)
	}
	t.Cleanup(func() { l.Close() })
	return l
}

func appendEvent(t *testing.T, log *memory.Log, worldTime int64, perceptID string) memory.Event {
	t.Helper()
	ev, err := log.Append(memory.EventInput{
		WorldTime: worldTime, Origin: "saw", PerceptID: perceptID, PerceptType: "sighting",
		ReceivedAt: worldTime, Content: map[string]any{"note": perceptID},
	})
	if err != nil {
		t.Fatalf("log.Append: %v", err)
	}
	return ev
}

// TestE6IsOpus5AndOffline proves card AC #1: E6's registry config, as this
// package actually consumes it (ClientDigester.Digest calls llm.E6), is
// Opus 5 — the ratified tier (docs/design/llm-routing-and-budget.md §1.3).
func TestE6IsOpus5AndOffline(t *testing.T) {
	cfg := llm.Registry[llm.E6]
	if cfg.Model != llm.ModelOpus5 {
		t.Errorf("E6 model = %q, want %q", cfg.Model, llm.ModelOpus5)
	}
	if !cfg.StructuredOutput {
		t.Error("E6 StructuredOutput = false, want true (RT-4/A-9)")
	}
}

// TestNoBatchAPIPath is card AC #1's negative half: no Batch API path
// exists in this package to misuse (§5.4: "Do not move E6 to the Batch
// API"). A design absence has no positive assertion to make, so this
// checks the one thing that would falsify it — the package's own source
// naming the SDK's batch service (BetaMessageBatchService / MessageBatch,
// anthropic-sdk-go/betamessagebatch.go). Matched case-sensitively against
// the SDK's actual identifier so prose discussing the absence (this file
// included) never trips it — only real usage would.
func TestNoBatchAPIPath(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || filepath.Ext(name) != ".go" || strings.HasSuffix(name, "_test.go") {
			continue
		}
		b, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("ReadFile(%s): %v", name, err)
		}
		if strings.Contains(string(b), "MessageBatch") {
			t.Errorf("%s references the SDK's batch service — E6 must never gain a Batch API path (§5.4)", name)
		}
	}
}

// TestRunNight_OrdinalMappingRoundTrip is card AC #2: the admitted buffer
// is rendered under the m1..mN convention and E6's accepted references
// ("m1", "m3") map back to the exact durable (world_time, hash) pair of
// the buffer event each ordinal named.
func TestRunNight_OrdinalMappingRoundTrip(t *testing.T) {
	log := openTestLog(t)
	ledger := openTestLedger(t)
	ev1 := appendEvent(t, log, 100, "p-1")
	appendEvent(t, log, 150, "p-2")
	ev3 := appendEvent(t, log, 200, "p-3")

	d := &scriptedDigester{replies: []digestReply{
		{raw: `{"summary":"a quiet night","references":["m1","m3","m9"]}`},
	}}
	if err := RunNight(context.Background(), log, ledger, d, nil, 200); err != nil {
		t.Fatalf("RunNight: %v", err)
	}

	recs := ledger.Records()
	if len(recs) != 1 {
		t.Fatalf("Records() len = %d, want 1", len(recs))
	}
	rec := recs[0]
	if rec.Empty {
		t.Fatal("record marked Empty, want a digest")
	}
	want := []memory.ID{ev1.ID(), ev3.ID()}
	if len(rec.References) != len(want) || rec.References[0] != want[0] || rec.References[1] != want[1] {
		t.Errorf("References = %+v, want %+v (m9 out of range must be dropped, not erroring)", rec.References, want)
	}
}

// TestRunNight_ConsolidatedWindowExcludedNextNight is card AC #2's window
// rule: once a night lands, its events never appear in a later night's
// buffer.
func TestRunNight_ConsolidatedWindowExcludedNextNight(t *testing.T) {
	log := openTestLog(t)
	ledger := openTestLedger(t)
	appendEvent(t, log, 100, "p-1")

	d := &scriptedDigester{replies: []digestReply{
		{raw: `{"summary":"night one"}`},
		{raw: `{"summary":"night two"}`},
	}}
	if err := RunNight(context.Background(), log, ledger, d, nil, 100); err != nil {
		t.Fatalf("night one: %v", err)
	}

	appendEvent(t, log, 200, "p-2")
	if err := RunNight(context.Background(), log, ledger, d, nil, 200); err != nil {
		t.Fatalf("night two: %v", err)
	}

	recs := ledger.Records()
	if len(recs) != 2 {
		t.Fatalf("Records() len = %d, want 2", len(recs))
	}
	if recs[1].WindowStart != 100 {
		t.Errorf("night two WindowStart = %d, want 100 (excludes night one's window)", recs[1].WindowStart)
	}
}

// TestRunNight_EmptyNightLandsMarker is card AC #3: nothing admitted is a
// consolidated night, not a failed one — no model call is even made.
func TestRunNight_EmptyNightLandsMarker(t *testing.T) {
	log := openTestLog(t)
	ledger := openTestLedger(t)
	d := &scriptedDigester{replies: []digestReply{{err: errors.New("must not be called")}}}

	if err := RunNight(context.Background(), log, ledger, d, nil, 500); err != nil {
		t.Fatalf("RunNight: %v", err)
	}
	if d.calls != 0 {
		t.Errorf("Digest called %d times on an empty buffer, want 0", d.calls)
	}
	recs := ledger.Records()
	if len(recs) != 1 || !recs[0].Empty {
		t.Fatalf("Records() = %+v, want one Empty record", recs)
	}
}

// TestRunNight_TransportFailureLandsNoMarker is card AC #3: a transport
// failure lands no marker and the buffer stays intact for retry.
func TestRunNight_TransportFailureLandsNoMarker(t *testing.T) {
	log := openTestLog(t)
	ledger := openTestLedger(t)
	appendEvent(t, log, 100, "p-1")

	d := &scriptedDigester{replies: []digestReply{{err: errors.New("connection reset")}}}
	if err := RunNight(context.Background(), log, ledger, d, nil, 100); err == nil {
		t.Fatal("RunNight: want error on transport failure, got nil")
	}
	if len(ledger.Records()) != 0 {
		t.Fatalf("Records() len = %d, want 0 (no marker on failure)", len(ledger.Records()))
	}
}

// TestRunNight_CancellationLandsNoMarker is card AC #3's cancellation
// case — the sleep-interrupted-by-wake edge (spec.md Edge Cases).
func TestRunNight_CancellationLandsNoMarker(t *testing.T) {
	log := openTestLog(t)
	ledger := openTestLedger(t)
	appendEvent(t, log, 100, "p-1")

	d := &scriptedDigester{replies: []digestReply{{err: context.Canceled}}}
	if err := RunNight(context.Background(), log, ledger, d, nil, 100); !errors.Is(err, context.Canceled) {
		t.Fatalf("RunNight err = %v, want wrapping context.Canceled", err)
	}
	if len(ledger.Records()) != 0 {
		t.Fatalf("Records() len = %d, want 0", len(ledger.Records()))
	}
}

// TestRunNight_OverLimitLandsNoMarker is card AC #3's truncation case —
// I's day-20 lesson: an over-limit response is a failure, never a stored
// digest, however plausible its (truncated) text looks.
func TestRunNight_OverLimitLandsNoMarker(t *testing.T) {
	log := openTestLog(t)
	ledger := openTestLedger(t)
	appendEvent(t, log, 100, "p-1")

	d := &scriptedDigester{replies: []digestReply{{raw: `{"summary":"cut off mid-sen`, truncated: true}}}
	if err := RunNight(context.Background(), log, ledger, d, nil, 100); err == nil {
		t.Fatal("RunNight: want error on truncated response, got nil")
	}
	if len(ledger.Records()) != 0 {
		t.Fatalf("Records() len = %d, want 0 (truncated digest must never be stored)", len(ledger.Records()))
	}
}

// TestRunNight_MultiNightAccumulationAfterFailures is spec.md's edge case:
// repeated failures accumulate the buffer rather than losing any of it,
// and the next success consolidates everything at once.
func TestRunNight_MultiNightAccumulationAfterFailures(t *testing.T) {
	log := openTestLog(t)
	ledger := openTestLedger(t)
	ev1 := appendEvent(t, log, 100, "p-1")

	d := &scriptedDigester{replies: []digestReply{{err: errors.New("fail 1")}}}
	if err := RunNight(context.Background(), log, ledger, d, nil, 100); err == nil {
		t.Fatal("want failure on night one")
	}

	ev2 := appendEvent(t, log, 200, "p-2")
	d.replies = append(d.replies, digestReply{err: errors.New("fail 2")})
	if err := RunNight(context.Background(), log, ledger, d, nil, 200); err == nil {
		t.Fatal("want failure on night two")
	}

	appendEvent(t, log, 300, "p-3") // ev3, referenced by ordinal below
	d.replies = append(d.replies, digestReply{raw: `{"summary":"finally","references":["m1","m2","m3"]}`})
	if err := RunNight(context.Background(), log, ledger, d, nil, 300); err != nil {
		t.Fatalf("night three: %v", err)
	}

	recs := ledger.Records()
	if len(recs) != 1 {
		t.Fatalf("Records() len = %d, want 1 (only the success lands)", len(recs))
	}
	if recs[0].WindowStart != 0 {
		t.Errorf("WindowStart = %d, want 0 (accumulated since the very first, never-consolidated window)", recs[0].WindowStart)
	}
	if len(recs[0].References) != 3 {
		t.Fatalf("References len = %d, want 3 (all three nights' events)", len(recs[0].References))
	}
	if recs[0].References[0] != ev1.ID() || recs[0].References[1] != ev2.ID() {
		t.Errorf("References = %+v, want to start with ev1, ev2", recs[0].References)
	}
}

// TestLedger_ReplayRoundTrip proves the ledger is genuinely event-sourced
// (M2 idiom): closing and reopening reconstructs the same reduced state.
func TestLedger_ReplayRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := PathFor(dir, "b-tam")

	l1, err := OpenLedger(path)
	if err != nil {
		t.Fatalf("OpenLedger: %v", err)
	}
	if err := l1.append(NightRecord{TriggerWorldTime: 100, WindowEnd: 100, Empty: true}); err != nil {
		t.Fatalf("append: %v", err)
	}
	if err := l1.append(NightRecord{
		TriggerWorldTime: 200, WindowStart: 100, WindowEnd: 200,
		Digest:     &llm.Digest{Summary: "s"},
		References: []memory.ID{{WorldTime: 150, Hash: "h1"}},
	}); err != nil {
		t.Fatalf("append: %v", err)
	}
	l1.Close()

	l2, err := OpenLedger(path)
	if err != nil {
		t.Fatalf("reopen OpenLedger: %v", err)
	}
	defer l2.Close()

	recs := l2.Records()
	if len(recs) != 2 {
		t.Fatalf("Records() len = %d, want 2", len(recs))
	}
	if wm, ok := l2.Watermark(); !ok || wm != 200 {
		t.Errorf("Watermark() = (%d, %v), want (200, true)", wm, ok)
	}
	if recs[1].Digest == nil || recs[1].Digest.Summary != "s" {
		t.Errorf("Digest not preserved across replay: %+v", recs[1].Digest)
	}
	if len(recs[1].References) != 1 || recs[1].References[0] != (memory.ID{WorldTime: 150, Hash: "h1"}) {
		t.Errorf("References not preserved across replay: %+v", recs[1].References)
	}
}

// TestLedger_WatermarkEmpty proves a fresh ledger has no watermark — the
// first night's buffer must start from the beginning of the log.
func TestLedger_WatermarkEmpty(t *testing.T) {
	l := openTestLedger(t)
	if _, ok := l.Watermark(); ok {
		t.Error("Watermark() ok = true on a fresh ledger, want false")
	}
}
