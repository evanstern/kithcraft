// Package consolidate (this file): T001's nightly ledger — an append-only,
// event-sourced record of E6 consolidation attempts (M2 idioms:
// mind/memory/log.go's append-only-log-plus-reducer shape, applied to a
// night record rather than a memory event). Unlike Log, a NightRecord
// carries no wire-protocol identity to verify, so this file uses plain
// encoding/json rather than mind/wire's canonical encoder — that encoder
// exists to make a percept's (world_time, hash) identity reproducible
// (body-protocol-v0.md §2's C-1..C-10), a property this internal-only file
// has no need of. specs/018-consolidation Phase 1 (T001, T002, T003).
package consolidate

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"kithcraft/mind/llm"
	"kithcraft/mind/memory"
)

// NightRecord is one landed ledger entry: either a consolidated night
// (Digest set, or Empty true for a night with nothing admitted) — a
// failed attempt (T003) never produces a NightRecord at all, so its
// absence from the ledger is the no-marker-on-failure rule made concrete.
type NightRecord struct {
	TriggerWorldTime int64       `json:"trigger_world_time"` // the sleep event's world_time
	WindowStart      int64       `json:"window_start"`       // exclusive: buffer covers (WindowStart, WindowEnd]
	WindowEnd        int64       `json:"window_end"`         // == TriggerWorldTime
	Empty            bool        `json:"empty"`              // true: nothing admitted, still a consolidated marker
	Digest           *llm.Digest `json:"digest,omitempty"`
	References       []memory.ID `json:"references,omitempty"` // accepted m1..mN, mapped to durable identity
}

// Ledger is T001's nightly ledger. One per villager, mirroring
// mind/memory.Log's per-villager file shape.
type Ledger struct {
	mu      sync.Mutex
	file    *os.File
	records []NightRecord
}

// PathFor returns the ledger path for villagerID under dir — one file per
// villager, distinct from mind/memory.Log's own file (T001).
func PathFor(dir, villagerID string) string {
	return filepath.Join(dir, villagerID+".ledger.jsonl")
}

// OpenLedger replays path (a no-such-file path replays to an empty
// ledger — no night has ever consolidated) and returns a Ledger ready for
// RunNight. Callers must Close it when done.
func OpenLedger(path string) (*Ledger, error) {
	records, err := replayLedger(path)
	if err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("consolidate: open ledger %s: %w", path, err)
	}
	return &Ledger{file: f, records: records}, nil
}

// Close closes the underlying file, mirroring mind/memory.Log.Close.
func (l *Ledger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.file.Close()
}

// Records returns a defensive copy of every landed record, in append
// order.
func (l *Ledger) Records() []NightRecord {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]NightRecord, len(l.records))
	copy(out, l.records)
	return out
}

// Watermark is the reducer T002's windowing runs on: the highest
// WindowEnd across every landed record. A night that failed never landed
// a record (T003), so it never advances this — which is exactly how a
// retry naturally re-covers the same window, and how repeated failures
// accumulate rather than lose a window. ok is false when no night has
// ever consolidated (world_time 0 is a legitimate world_time, so "no
// watermark yet" cannot be spelled as 0).
func (l *Ledger) Watermark() (worldTime int64, ok bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, r := range l.records {
		if !ok || r.WindowEnd > worldTime {
			worldTime, ok = r.WindowEnd, true
		}
	}
	return worldTime, ok
}

// append is the ledger's only write path — unexported because landing a
// record is RunNight's decision (cycle.go), never a caller's.
func (l *Ledger) append(rec NightRecord) error {
	raw, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("consolidate: cannot encode night record: %w", err)
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	if _, err := l.file.Write(append(raw, '\n')); err != nil {
		return fmt.Errorf("consolidate: ledger append write failed: %w", err)
	}
	if err := l.file.Sync(); err != nil {
		return fmt.Errorf("consolidate: ledger append sync failed: %w", err)
	}
	l.records = append(l.records, rec)
	return nil
}

func replayLedger(path string) ([]NightRecord, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("consolidate: read ledger %s: %w", path, err)
	}
	var records []NightRecord
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for i := 1; scanner.Scan(); i++ {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var rec NightRecord
		if err := json.Unmarshal(line, &rec); err != nil {
			return nil, fmt.Errorf("consolidate: ledger %s line %d: %w", path, i, err)
		}
		records = append(records, rec)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("consolidate: read ledger %s: %w", path, err)
	}
	return records, nil
}
