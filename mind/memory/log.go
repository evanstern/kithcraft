// Package memory (this file): the append-only memory-event log — the
// durable record every later remember-surface rule (body-protocol-v0.md
// §6.4, RM-1..RM-7) will be a reducer or a read-time function over. Card
// AC #1 requires mutation to be impossible at the type level, not by
// review convention: Event's fields are unexported, Log.Append is the
// only constructor, and every getter returns a value or a defensive copy
// — nothing exported ever hands out a pointer or map into an Event's own
// state. specs/010-event-sourced-memory Phase 1 (T001-T003).
package memory

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"kithcraft/mind/wire"
)

// EventInput is the not-yet-durable shape Append accepts: everything a
// caller supplies to record one memory event, generally derived from an
// admitted percept (§2.6's provenance plus its content). It carries no
// identity of its own — Append computes the (WorldTime, Hash) pair T003
// wires in as durable identity, and the Event it returns is the only
// record that lasts.
type EventInput struct {
	WorldTime   int64
	Origin      string // §2.7 origin, verbatim from provenance.origin
	PerceptID   string
	PerceptType string
	ReceivedAt  int64
	ObservedAt  *int64 // nil = unknown = maximally stale (§2.6)
	Content     any    // decoded percept content, wire.Decode's value shape
}

// ID is a memory event's durable identity: the (world_time, hash) pair
// T003 wires in as the convention M7's consolidation will address stored
// memories by. Multiple events can share a world_time (more than one
// percept lands in the same tick); Hash — computed over everything but
// world_time — disambiguates them. ID is comparable and usable as a map
// key directly.
type ID struct {
	WorldTime int64
	Hash      string
}

// Event is one immutable, durably-identified memory event. The only way
// to produce one is Log.Append (or replaying a log Append already wrote);
// nothing in this package's exported API can modify one afterward (card
// AC #1). See log_immutable_test.go for the API-surface proof.
type Event struct {
	raw         []byte // the canonical JSON line as appended/read, sans '\n'
	worldTime   int64
	hash        string
	origin      string
	perceptID   string
	perceptType string
	receivedAt  int64
	observedAt  *int64
	content     any
}

func (e Event) ID() ID              { return ID{WorldTime: e.worldTime, Hash: e.hash} }
func (e Event) WorldTime() int64    { return e.worldTime }
func (e Event) Hash() string        { return e.hash }
func (e Event) Origin() string      { return e.origin }
func (e Event) PerceptID() string   { return e.perceptID }
func (e Event) PerceptType() string { return e.perceptType }
func (e Event) ReceivedAt() int64   { return e.receivedAt }
func (e Event) ObservedAt() *int64  { return copyInt64(e.observedAt) }
func (e Event) Content() any        { return deepCopyJSON(e.content) }
func (e Event) Raw() []byte         { return append([]byte(nil), e.raw...) }

// Log is one villager's append-only memory-event log: a JSONL file plus
// the in-memory event list that replaying it reduces to (US1 AC #2). The
// file is the durable record; the slice is a read cache Open rebuilds by
// replay and Append extends — nothing else ever writes to either.
type Log struct {
	mu     sync.Mutex
	file   *os.File
	events []Event
}

// PathFor returns the JSONL log path for villagerID under dir — one file
// per villager (T001).
func PathFor(dir, villagerID string) string {
	return filepath.Join(dir, villagerID+".jsonl")
}

// Open replays path (a no-such-file path replays to an empty log — a
// villager's first session has no history yet) and returns a Log ready
// to Append to. Callers must Close it when done.
func Open(path string) (*Log, error) {
	events, err := replay(path)
	if err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("memory: open log %s: %w", path, err)
	}
	return &Log{file: f, events: events}, nil
}

// Close closes the underlying file. It does not touch the in-memory
// event list — a Log that outlives its file handle is still a valid
// read-only view.
func (l *Log) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.file.Close()
}

// Append durably records in as one new Event: computes its (world_time,
// hash) identity, writes the canonical JSON line, and extends the
// in-memory state. This is the package's only write path (card AC #7 —
// the belief store built on top of this log gets no other way in).
func (l *Log) Append(in EventInput) (Event, error) {
	hash, err := computeHash(in.Origin, in.PerceptID, in.PerceptType, in.ReceivedAt, in.ObservedAt, in.Content)
	if err != nil {
		return Event{}, fmt.Errorf("memory: cannot hash event content: %w", err)
	}
	record := map[string]any{
		"world_time":   in.WorldTime,
		"hash":         hash,
		"origin":       in.Origin,
		"percept_id":   in.PerceptID,
		"percept_type": in.PerceptType,
		"received_at":  in.ReceivedAt,
		"observed_at":  int64OrNil(in.ObservedAt),
		"content":      in.Content,
	}
	raw, err := wire.EncodeCanonical(record)
	if err != nil {
		return Event{}, fmt.Errorf("memory: cannot encode event: %w", err)
	}
	ev := Event{
		raw: raw, worldTime: in.WorldTime, hash: hash, origin: in.Origin,
		perceptID: in.PerceptID, perceptType: in.PerceptType, receivedAt: in.ReceivedAt,
		observedAt: copyInt64(in.ObservedAt), content: deepCopyJSON(in.Content),
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	if _, err := l.file.Write(append(append([]byte(nil), raw...), '\n')); err != nil {
		return Event{}, fmt.Errorf("memory: append write failed: %w", err)
	}
	if err := l.file.Sync(); err != nil {
		return Event{}, fmt.Errorf("memory: append sync failed: %w", err)
	}
	l.events = append(l.events, ev)
	return ev, nil
}

// Events returns a defensive copy of the log's reduced state in append
// order — the reducer US1 AC #2 requires (Phase 2 folds this into belief
// state; Phase 1 proves the log itself is what state reduces over).
func (l *Log) Events() []Event {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]Event, len(l.events))
	copy(out, l.events)
	return out
}

// replay reads path's JSONL lines back into Events, recomputing and
// verifying each one's hash — a line altered after the fact fails to
// replay rather than replaying as if it were never touched.
func replay(path string) ([]Event, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("memory: read log %s: %w", path, err)
	}
	lines := bytes.Split(data, []byte("\n"))
	events := make([]Event, 0, len(lines))
	for i, line := range lines {
		if len(line) == 0 { // final split element after the trailing '\n'
			continue
		}
		ev, err := decodeLine(line)
		if err != nil {
			return nil, fmt.Errorf("memory: log %s line %d: %w", path, i+1, err)
		}
		events = append(events, ev)
	}
	return events, nil
}

func decodeLine(line []byte) (Event, error) {
	v, err := wire.Decode(line)
	if err != nil {
		return Event{}, err
	}
	m, ok := v.(map[string]any)
	if !ok {
		return Event{}, fmt.Errorf("log line is not a JSON object")
	}

	worldTime, ok := m["world_time"].(int64)
	if !ok {
		return Event{}, fmt.Errorf("missing or malformed world_time")
	}
	hash, ok := m["hash"].(string)
	if !ok {
		return Event{}, fmt.Errorf("missing or malformed hash")
	}
	origin, ok := m["origin"].(string)
	if !ok {
		return Event{}, fmt.Errorf("missing or malformed origin")
	}
	perceptID, ok := m["percept_id"].(string)
	if !ok {
		return Event{}, fmt.Errorf("missing or malformed percept_id")
	}
	perceptType, ok := m["percept_type"].(string)
	if !ok {
		return Event{}, fmt.Errorf("missing or malformed percept_type")
	}
	receivedAt, ok := m["received_at"].(int64)
	if !ok {
		return Event{}, fmt.Errorf("missing or malformed received_at")
	}
	observedAt, err := decodeObservedAt(m["observed_at"])
	if err != nil {
		return Event{}, err
	}
	content, hasContent := m["content"]
	if !hasContent {
		return Event{}, fmt.Errorf("missing content")
	}

	wantHash, err := computeHash(origin, perceptID, perceptType, receivedAt, observedAt, content)
	if err != nil {
		return Event{}, fmt.Errorf("cannot verify hash: %w", err)
	}
	if wantHash != hash {
		return Event{}, fmt.Errorf("hash mismatch: stored %s, recomputed %s (log line altered after append)", hash, wantHash)
	}

	return Event{
		raw: append([]byte(nil), line...), worldTime: worldTime, hash: hash, origin: origin,
		perceptID: perceptID, perceptType: perceptType, receivedAt: receivedAt,
		observedAt: observedAt, content: content,
	}, nil
}

func decodeObservedAt(v any) (*int64, error) {
	if v == nil {
		return nil, nil
	}
	n, ok := v.(int64)
	if !ok {
		return nil, fmt.Errorf("malformed observed_at")
	}
	return &n, nil
}

// computeHash is T003's identity function: a sha256 over the canonical
// JSON encoding of everything an event carries except world_time and the
// hash itself, so the pair (world_time, hash) disambiguates events that
// land in the same tick without world_time feeding its own disambiguator.
func computeHash(origin, perceptID, perceptType string, receivedAt int64, observedAt *int64, content any) (string, error) {
	b, err := wire.EncodeCanonical(map[string]any{
		"origin":       origin,
		"percept_id":   perceptID,
		"percept_type": perceptType,
		"received_at":  receivedAt,
		"observed_at":  int64OrNil(observedAt),
		"content":      content,
	})
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

func int64OrNil(p *int64) any {
	if p == nil {
		return nil
	}
	return *p
}

func copyInt64(p *int64) *int64 {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}

// deepCopyJSON copies a wire.Decode-shaped value (nil, bool, int64,
// string, []any, map[string]any) so a caller mutating what Content()
// returned, or Append mutating what it was handed after the call, can
// never reach an Event's own state.
func deepCopyJSON(v any) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, e := range t {
			out[k] = deepCopyJSON(e)
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i, e := range t {
			out[i] = deepCopyJSON(e)
		}
		return out
	default: // nil, bool, int64, string are already immutable values
		return t
	}
}
