package consolidate

import (
	"context"
	"testing"

	"kithcraft/mind/memory"
)

func openTestArchive(t *testing.T, dir string) *Archive {
	t.Helper()
	a, err := OpenArchive(ArchivePathFor(dir))
	if err != nil {
		t.Fatalf("OpenArchive: %v", err)
	}
	t.Cleanup(func() { a.Close() })
	return a
}

// TestArchive_UnknownMindNotArchived is the default: nothing is archived
// until Archive lands a record for it.
func TestArchive_UnknownMindNotArchived(t *testing.T) {
	a := openTestArchive(t, t.TempDir())
	if a.IsArchived("b-tam") {
		t.Error("IsArchived(b-tam) = true before any Archive call")
	}
	if a.TokenRetired("b-tam") {
		t.Error("TokenRetired(b-tam) = true before any Archive call")
	}
}

// TestArchive_LandsAndPersists is card AC #6's core: archival lands as a
// durable fact, readable back through a fresh OpenArchive against the
// same path — mirroring Ledger's own replay round trip (T001).
func TestArchive_LandsAndPersists(t *testing.T) {
	dir := t.TempDir()
	path := ArchivePathFor(dir)

	a, err := OpenArchive(path)
	if err != nil {
		t.Fatalf("OpenArchive: %v", err)
	}
	if err := a.Archive(ArchiveRecord{MindID: "b-tam", BodyToken: "b-tam", WorldTime: 500}); err != nil {
		t.Fatalf("Archive: %v", err)
	}
	if !a.IsArchived("b-tam") {
		t.Fatal("IsArchived(b-tam) = false right after Archive")
	}
	if !a.TokenRetired("b-tam") {
		t.Fatal("TokenRetired(b-tam) = false right after Archive")
	}
	a.Close()

	reopened, err := OpenArchive(path)
	if err != nil {
		t.Fatalf("re-OpenArchive: %v", err)
	}
	defer reopened.Close()
	if !reopened.IsArchived("b-tam") {
		t.Error("archival did not survive a reopen")
	}
	recs := reopened.Records()
	if len(recs) != 1 || recs[0].MindID != "b-tam" || recs[0].WorldTime != 500 {
		t.Errorf("Records() = %+v, want one record for b-tam at world_time 500", recs)
	}
}

// TestArchive_IdempotentByMindID: re-archiving an already-archived mind
// lands no second record — archival is a one-way, one-time fact.
func TestArchive_IdempotentByMindID(t *testing.T) {
	a := openTestArchive(t, t.TempDir())
	rec := ArchiveRecord{MindID: "b-tam", BodyToken: "b-tam", WorldTime: 500}
	if err := a.Archive(rec); err != nil {
		t.Fatalf("first Archive: %v", err)
	}
	if err := a.Archive(ArchiveRecord{MindID: "b-tam", BodyToken: "b-tam", WorldTime: 999}); err != nil {
		t.Fatalf("second Archive: %v", err)
	}
	if got := a.Records(); len(got) != 1 || got[0].WorldTime != 500 {
		t.Errorf("Records() = %+v, want exactly the first landed record", got)
	}
}

// TestArchive_UnaffectedMindStaysOpen: archiving one mind never archives
// another — the set is keyed, not global.
func TestArchive_UnaffectedMindStaysOpen(t *testing.T) {
	a := openTestArchive(t, t.TempDir())
	if err := a.Archive(ArchiveRecord{MindID: "b-tam", BodyToken: "b-tam", WorldTime: 500}); err != nil {
		t.Fatalf("Archive: %v", err)
	}
	if a.IsArchived("b-nell") {
		t.Error("IsArchived(b-nell) = true after archiving an unrelated mind b-tam")
	}
}

// TestArchive_DeathBeforeOwnConsolidation is the edge case named in
// spec.md's Edge Cases: a villager dies mid-night, before its own
// consolidation runs. Archival wins — its last night is never
// consolidated (no session opens for it, so RunNight is never invoked
// again), and its log remains exactly as it was: readable, unconsolidated
// tail and all.
func TestArchive_DeathBeforeOwnConsolidation(t *testing.T) {
	dir := t.TempDir()
	log, err := openLogAt(t, dir, "b-tam")
	if err != nil {
		t.Fatalf("open log: %v", err)
	}
	ledger, err := OpenLedger(PathFor(dir, "b-tam"))
	if err != nil {
		t.Fatalf("OpenLedger: %v", err)
	}
	t.Cleanup(func() { ledger.Close() })
	archive := openTestArchive(t, dir)

	// Night one consolidates normally.
	appendEvent(t, log, 100, "p-1")
	d := &scriptedDigester{replies: []digestReply{{raw: `{"summary":"night one"}`}}}
	if err := RunNight(context.Background(), log, ledger, d, nil, 100); err != nil {
		t.Fatalf("night one RunNight: %v", err)
	}

	// The villager dies partway through night two: more events land in
	// the log (its "last night"), but death arrives before anything ever
	// triggers a second RunNight for it.
	lastEvent := appendEvent(t, log, 150, "p-2")
	if err := archive.Archive(ArchiveRecord{MindID: "b-tam", BodyToken: "b-tam", WorldTime: 150}); err != nil {
		t.Fatalf("Archive: %v", err)
	}

	// Archival wins: the ledger gained no second record (its last night
	// was never consolidated)...
	recs := ledger.Records()
	if len(recs) != 1 {
		t.Fatalf("ledger Records() len = %d, want 1 (no consolidation after archival)", len(recs))
	}
	// ...and the log's unconsolidated tail remains exactly as it was,
	// still readable through the existing store API.
	events := log.Events()
	if len(events) != 2 {
		t.Fatalf("log Events() len = %d, want 2", len(events))
	}
	if events[1].ID() != lastEvent.ID() {
		t.Error("the unconsolidated night-two event did not survive archival unchanged")
	}
	if !archive.IsArchived("b-tam") {
		t.Error("IsArchived(b-tam) = false after Archive")
	}
}

// openLogAt mirrors openTestLog but at a caller-chosen directory, so the
// log, ledger, and archive can share one dir the way ArchivePathFor
// documents ("persisted alongside the ledger").
func openLogAt(t *testing.T, dir, villagerID string) (*memory.Log, error) {
	t.Helper()
	log, err := memory.Open(memory.PathFor(dir, villagerID))
	if err != nil {
		return nil, err
	}
	t.Cleanup(func() { log.Close() })
	return log, nil
}
