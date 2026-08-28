// Package consolidate (this file): T005's archival, ruling R-9 (archived,
// not terminated — docs/design/demo-build-plan.md's R-9 row). Archival is
// state, not lifecycle: it never touches memory.Log or Ledger — a dead
// mind's durable log, and any nights it already consolidated, stay
// exactly as replay left them, forever readable through Log/Ledger's
// ordinary API (spec.md US4 AC #2). What Archive itself owns is two
// small, closed, append-only facts per mind, mirroring Ledger's own
// replay-then-append shape (T001): whether a new session may ever open
// for it again (no), and whether its body token may ever be reissued
// (no).
//
// mind/seam's session-open path consults IsArchived before honoring any
// session_open naming an archived mind — same session_close/reason:error
// refusal shape session.go's manifest-mismatch refusal already uses
// (T005 "hook into the session-open path the way the existing refusal
// machinery works").
//
// ponytail: mind identity and body token are the same opaque string
// below — mind/seam has no separate body-to-mind-identity resolution
// layer yet (the whole ingest skeleton is replaced by M5). If a real
// resolution layer arrives, key archival on mind identity and look up
// its *current* body token there instead of assuming they coincide.
package consolidate

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// ArchiveRecord is one landed archival: the durable fact that MindID
// (== BodyToken, see package doc) may never open a new session again and
// BodyToken may never be reissued, as of WorldTime.
type ArchiveRecord struct {
	MindID    string `json:"mind_id"`
	BodyToken string `json:"body_token"`
	WorldTime int64  `json:"world_time"`
}

// Archive is the archived-set: an append-only JSONL file of
// ArchiveRecords, one per archived mind, shared across every mind
// this daemon process ever hosts (unlike Log/Ledger's per-villager
// file) — a session_open can name any body, so the refusal check needs
// one registry, not one file per mind it hasn't identified yet.
type Archive struct {
	mu      sync.Mutex
	file    *os.File
	records []ArchiveRecord
	ids     map[string]bool
	tokens  map[string]bool
}

// ArchivePathFor returns the shared archive path under dir, alongside
// per-villager ledgers and logs (plan.md design decision 5: "persisted
// alongside the ledger").
func ArchivePathFor(dir string) string {
	return filepath.Join(dir, "archive.jsonl")
}

// OpenArchive replays path (a no-such-file path replays to an empty
// archive — nothing has ever been archived) and returns an Archive ready
// to Archive into. Callers must Close it when done.
func OpenArchive(path string) (*Archive, error) {
	records, err := replayArchive(path)
	if err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("consolidate: open archive %s: %w", path, err)
	}
	a := &Archive{file: f, ids: map[string]bool{}, tokens: map[string]bool{}}
	for _, r := range records {
		a.records = append(a.records, r)
		a.ids[r.MindID] = true
		a.tokens[r.BodyToken] = true
	}
	return a, nil
}

// Close closes the underlying file, mirroring Ledger.Close.
func (a *Archive) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.file.Close()
}

// Records returns a defensive copy of every landed record, in append
// order.
func (a *Archive) Records() []ArchiveRecord {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]ArchiveRecord, len(a.records))
	copy(out, a.records)
	return out
}

// IsArchived reports whether mindID has been archived — the check
// mind/seam's session-open refusal consults (body and mind identity
// coincide in this skeleton; see package doc).
func (a *Archive) IsArchived(mindID string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.ids[mindID]
}

// TokenRetired reports whether bodyToken has been retired into the
// never-reissue set. Distinct from IsArchived only in the (currently
// hypothetical, see package doc) case a mind's body token differs from
// its mind identity.
func (a *Archive) TokenRetired(bodyToken string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.tokens[bodyToken]
}

// Archive lands rec durably. It is idempotent by MindID: archival is a
// one-way, one-time fact, so re-archiving an already-archived mind is a
// no-op rather than a second record — the never-reissue set can only
// ever grow, never re-land the same token.
func (a *Archive) Archive(rec ArchiveRecord) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.ids[rec.MindID] {
		return nil
	}
	raw, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("consolidate: cannot encode archive record: %w", err)
	}
	if _, err := a.file.Write(append(raw, '\n')); err != nil {
		return fmt.Errorf("consolidate: archive append write failed: %w", err)
	}
	if err := a.file.Sync(); err != nil {
		return fmt.Errorf("consolidate: archive append sync failed: %w", err)
	}
	a.records = append(a.records, rec)
	a.ids[rec.MindID] = true
	a.tokens[rec.BodyToken] = true
	return nil
}

func replayArchive(path string) ([]ArchiveRecord, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("consolidate: read archive %s: %w", path, err)
	}
	var records []ArchiveRecord
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for i := 1; scanner.Scan(); i++ {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var rec ArchiveRecord
		if err := json.Unmarshal(line, &rec); err != nil {
			return nil, fmt.Errorf("consolidate: archive %s line %d: %w", path, i, err)
		}
		records = append(records, rec)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("consolidate: read archive %s: %w", path, err)
	}
	return records, nil
}
