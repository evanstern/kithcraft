// Package memory (this file): the private, provenance-stamped belief store
// (PM-1) and RM-4..RM-7 — the read/write rules everything durably known
// obeys. Belief's fields are unexported and Store's only mutators are
// Upsert and Retract (card AC #7 — the mind's own ingest/consolidation flow
// is the only write path; see beliefs_external_test.go for the API-surface
// proof). specs/010-event-sourced-memory Phase 2 (T005).
package memory

import (
	"math"
	"sync"
)

// Config is mind configuration for RM-5/RM-6's read-time arithmetic — half-
// life, floor, and freshness horizon are not protocol constants (§2.2), so
// they live here with demo defaults, keyed per belief Kind. Config is
// treated as immutable after construction: Store never writes through it.
type Config struct {
	HalfLife        map[string]int64 // belief kind -> world-time half-life for confidence decay (RM-5)
	DefaultHalfLife int64
	Horizon         map[string]int64 // belief kind -> world-time freshness horizon (RM-6)
	DefaultHorizon  int64
	ConfidenceFloor float64 // RM-5: below this, effective confidence stops driving behaviour
}

// DefaultConfig returns demo-scale defaults (ponytail: tuned values are a
// later balancing pass, not this task's scope — plan.md Complexity Tracking).
func DefaultConfig() Config {
	return Config{DefaultHalfLife: 20_000, DefaultHorizon: 6_000, ConfidenceFloor: 0.05}
}

func (c Config) halfLife(kind string) int64 {
	if v, ok := c.HalfLife[kind]; ok && v > 0 {
		return v
	}
	return c.DefaultHalfLife
}

func (c Config) horizon(kind string) int64 {
	if v, ok := c.Horizon[kind]; ok && v > 0 {
		return v
	}
	return c.DefaultHorizon
}

// UpsertInput is the not-yet-stored shape Upsert accepts. Subject is an
// opaque identity for what the belief is about (a thing_id, place, or kind
// token — AR-2 opaqueness: the store never parses it, only compares it for
// equality as a map key). Kind selects Config's per-kind half-life/horizon.
type UpsertInput struct {
	Subject    string
	Kind       string
	Content    any
	Provenance Provenance
	ObservedAt *int64  // world time observed; nil = unknown = maximally stale (§2.6)
	Confidence float64 // stamped once; RM-5 requires it never change afterward
}

// Belief is one stored, provenance-stamped fact (PM-1). The only way to
// produce one is Store.Upsert; every exported getter returns a value or a
// defensive copy, mirroring Event's immutability shape (card AC #1's
// pattern applied to beliefs).
type Belief struct {
	subject    string
	kind       string
	content    any
	provenance Provenance
	observedAt *int64
	confidence float64
}

func (b Belief) Subject() string           { return b.subject }
func (b Belief) Kind() string              { return b.kind }
func (b Belief) Content() any              { return deepCopyJSON(b.content) }
func (b Belief) Provenance() Provenance    { return b.provenance }
func (b Belief) ObservedAt() *int64        { return copyInt64(b.observedAt) }
func (b Belief) StoredConfidence() float64 { return b.confidence }

// RetractionCause is RM-7's closed vocabulary: the only reasons a belief may
// be removed from the store. Anything else is not a deletion channel.
type RetractionCause string

const (
	Correction       RetractionCause = "correction"
	Death            RetractionCause = "death"
	WitnessedRemoval RetractionCause = "witnessed_removal"
)

func (c RetractionCause) valid() bool {
	switch c {
	case Correction, Death, WitnessedRemoval:
		return true
	}
	return false
}

// Store is the private, provenance-stamped belief map (PM-1): distinct from
// any vendor resolution index, with no field or method a vendor, the
// player, or a debug command could reach (card AC #2, AC #7). Its state is
// a fold over the sequence of Upsert/Retract calls it receives — replaying
// the same sequence twice always yields the same map (US1 AC #2's reducer
// property applied to belief state).
type Store struct {
	mu      sync.Mutex
	beliefs map[string]Belief
	cfg     Config
}

// NewStore returns an empty belief store configured by cfg (DefaultConfig
// for demo defaults).
func NewStore(cfg Config) *Store {
	return &Store{beliefs: make(map[string]Belief), cfg: cfg}
}

// Upsert is the store's ingest path (RM-4). It always applies except in the
// one case RM-4 forbids: a secondhand claim (Provenance != Witnessed)
// arriving for a subject whose stored belief is Witnessed and no staler
// (observed_at not older) than the incoming claim — secondhand never
// overwrites fresher-or-equal firsthand. observed_at: nil is maximally
// stale (§2.6), so it never counts as "fresher." Reports whether it applied.
func (s *Store) Upsert(in UpsertInput) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.beliefs[in.Subject]
	if exists && existing.provenance == Witnessed && in.Provenance != Witnessed && !fresherThan(in.ObservedAt, existing.observedAt) {
		return false
	}
	s.beliefs[in.Subject] = Belief{
		subject: in.Subject, kind: in.Kind, content: deepCopyJSON(in.Content),
		provenance: in.Provenance, observedAt: copyInt64(in.ObservedAt), confidence: in.Confidence,
	}
	return true
}

// Get returns the subject's current stored belief, unchanged by time (RM-5:
// stored confidence never mutates) — reads Confidence/Fresh for the
// read-time-computed values.
func (s *Store) Get(subject string) (Belief, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.beliefs[subject]
	return b, ok
}

// Confidence is RM-5: effective confidence computed at read time from the
// belief's stored confidence, its age since observed_at, and its kind's
// half-life. It never mutates what is stored.
func (s *Store) Confidence(subject string, now int64) (float64, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.beliefs[subject]
	if !ok {
		return 0, false
	}
	return effectiveConfidence(b.confidence, b.observedAt, now, s.cfg.halfLife(b.kind)), true
}

// Active reports whether subject's effective confidence is at or above
// Config's floor — RM-5: below the floor a belief stops driving behaviour
// but stays revisable, i.e. it is still in the store (Get still finds it).
func (s *Store) Active(subject string, now int64) (bool, bool) {
	conf, ok := s.Confidence(subject, now)
	if !ok {
		return false, false
	}
	return conf >= s.cfg.ConfidenceFloor, true
}

// Fresh is RM-6: fresh iff now - observed_at < horizon(kind), evaluated at
// read time and never stored. observed_at: nil is maximally stale and so
// never fresh, regardless of now.
func (s *Store) Fresh(subject string, now int64) (bool, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.beliefs[subject]
	if !ok {
		return false, false
	}
	return fresh(b.observedAt, now, s.cfg.horizon(b.kind)), true
}

// Retract is the only path that removes a subject from the store (RM-7): a
// cause outside the closed vocabulary is a no-op, not a deletion — time
// alone, or a typo'd cause, never deletes. Reports whether a belief was
// removed.
func (s *Store) Retract(subject string, cause RetractionCause) bool {
	if !cause.valid() {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.beliefs[subject]; !ok {
		return false
	}
	delete(s.beliefs, subject)
	return true
}

// fresherThan reports whether a is strictly more recent than b, treating
// nil (§2.6: unknown = maximally stale) as older than any known time and
// tying nil against nil.
func fresherThan(a, b *int64) bool {
	if a == nil {
		return false
	}
	if b == nil {
		return true
	}
	return *a > *b
}

// effectiveConfidence is RM-5's read-time arithmetic: exponential decay by
// age since observedAt against halfLife. observedAt: nil (maximally stale)
// decays to zero rather than an undefined age.
func effectiveConfidence(stored float64, observedAt *int64, now, halfLife int64) float64 {
	if observedAt == nil {
		return 0
	}
	age := now - *observedAt
	if age <= 0 {
		return stored
	}
	if halfLife <= 0 {
		return 0
	}
	return stored * math.Pow(0.5, float64(age)/float64(halfLife))
}

// fresh is RM-6's read-time arithmetic: observedAt: nil (maximally stale)
// is never fresh.
func fresh(observedAt *int64, now, horizon int64) bool {
	if observedAt == nil {
		return false
	}
	age := now - *observedAt
	return age >= 0 && age < horizon
}
