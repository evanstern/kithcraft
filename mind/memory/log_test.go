package memory

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func observedAt(v int64) *int64 { return &v }

func sampleInput(worldTime int64) EventInput {
	return EventInput{
		WorldTime: worldTime, Origin: "saw", PerceptID: "p-1", PerceptType: "sighting",
		ReceivedAt: worldTime, ObservedAt: observedAt(worldTime),
		Content: map[string]any{"kind": "k-oak", "place": "pl-1"},
	}
}

func TestLog_AppendComputesWorldTimeHashIdentity(t *testing.T) {
	log, err := Open(PathFor(t.TempDir(), "b-tam"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer log.Close()

	ev, err := log.Append(sampleInput(100))
	if err != nil {
		t.Fatalf("Append: %v", err)
	}
	if ev.WorldTime() != 100 {
		t.Fatalf("WorldTime = %d, want 100", ev.WorldTime())
	}
	if ev.Hash() == "" {
		t.Fatal("Hash is empty; T003 identity requires a non-empty content hash")
	}
	if ev.ID() != (ID{WorldTime: 100, Hash: ev.Hash()}) {
		t.Fatalf("ID() = %+v, want (world_time, hash) pair matching the event", ev.ID())
	}

	// Two events at the same world_time with different content must not
	// collide — the hash half of the pair is what disambiguates them.
	ev2, err := log.Append(EventInput{
		WorldTime: 100, Origin: "heard", PerceptID: "p-2", PerceptType: "sound",
		ReceivedAt: 100, Content: map[string]any{"sound_kind": "footsteps"},
	})
	if err != nil {
		t.Fatalf("Append: %v", err)
	}
	if ev2.ID() == ev.ID() {
		t.Fatal("distinct content at the same world_time produced the same identity pair")
	}
}

// TestLog_ReplayReproducesStateByteForByte is US1 AC #2 / SC-004's other
// half: state is a reducer over the log, so a fresh Log opened against
// the same file a prior Log wrote must reduce to the identical state —
// compared here at the byte level (Raw), not merely by field equality.
func TestLog_ReplayReproducesStateByteForByte(t *testing.T) {
	path := PathFor(t.TempDir(), "b-tam")

	original, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	var want []Event
	for i := int64(0); i < 5; i++ {
		ev, err := original.Append(sampleInput(1000 + i))
		if err != nil {
			t.Fatalf("Append: %v", err)
		}
		want = append(want, ev)
	}
	if err := original.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	replayed, err := Open(path)
	if err != nil {
		t.Fatalf("Open (replay): %v", err)
	}
	defer replayed.Close()

	got := replayed.Events()
	if len(got) != len(want) {
		t.Fatalf("replayed %d events, want %d", len(got), len(want))
	}
	for i := range want {
		if !reflectEqualBytes(got[i].Raw(), want[i].Raw()) {
			t.Fatalf("event %d: replayed bytes differ from appended bytes\n got: %s\nwant: %s", i, got[i].Raw(), want[i].Raw())
		}
		if got[i].ID() != want[i].ID() {
			t.Fatalf("event %d: replayed ID %+v != appended ID %+v", i, got[i].ID(), want[i].ID())
		}
	}
}

func reflectEqualBytes(a, b []byte) bool { return reflect.DeepEqual(a, b) }

func TestOpen_MissingFileReplaysEmpty(t *testing.T) {
	log, err := Open(filepath.Join(t.TempDir(), "no-such.jsonl"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer log.Close()
	if len(log.Events()) != 0 {
		t.Fatalf("Events() = %v, want empty for a log with no prior file", log.Events())
	}
}

func TestOpen_TamperedLineFailsReplay(t *testing.T) {
	path := PathFor(t.TempDir(), "b-tam")
	log, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if _, err := log.Append(sampleInput(1)); err != nil {
		t.Fatalf("Append: %v", err)
	}
	if err := log.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// Hand-edit the file: append-only means outside the package's own
	// write path there is no API for this, only the filesystem — and a
	// hand edit must not replay silently (§6.4's "cannot be edited" is
	// enforced by the hash, since Go's type system can't reach the file).
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	tampered := []byte(string(data)[:len(data)-2] + "x\n")
	if err := os.WriteFile(path, tampered, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if _, err := Open(path); err == nil {
		t.Fatal("Open replayed a tampered log line without error")
	}
}

func TestLog_ContentIsolatedFromCallerMutation(t *testing.T) {
	log, err := Open(PathFor(t.TempDir(), "b-tam"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer log.Close()

	content := map[string]any{"kind": "k-oak"}
	in := EventInput{WorldTime: 1, Origin: "saw", PerceptID: "p-1", PerceptType: "sighting", ReceivedAt: 1, Content: content}
	ev, err := log.Append(in)
	if err != nil {
		t.Fatalf("Append: %v", err)
	}

	// Mutating the map the caller passed in must not reach the stored
	// event: content came in as a reference type, and Append must copy.
	content["kind"] = "tampered"
	if got := ev.Content().(map[string]any)["kind"]; got != "k-oak" {
		t.Fatalf("stored content changed via caller's map: kind = %v", got)
	}

	// Mutating a value handed back by Content() must not reach the
	// stored event either — Content() must return a fresh copy each call.
	returned := ev.Content().(map[string]any)
	returned["kind"] = "also-tampered"
	if got := ev.Content().(map[string]any)["kind"]; got != "k-oak" {
		t.Fatalf("stored content changed via Content()'s returned map: kind = %v", got)
	}
}

func TestLog_EventsReturnsDefensiveCopy(t *testing.T) {
	log, err := Open(PathFor(t.TempDir(), "b-tam"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer log.Close()
	if _, err := log.Append(sampleInput(1)); err != nil {
		t.Fatalf("Append: %v", err)
	}

	got := log.Events()
	got[0] = Event{} // mutate the caller's copy
	if log.Events()[0].WorldTime() != 1 {
		t.Fatal("mutating the slice Events() returned reached the log's internal state")
	}
}
