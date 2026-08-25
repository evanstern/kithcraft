package memory

import "testing"

func obs(v int64) *int64 { return &v }

// TestRM1_DirectPerceptionGatesOnOriginAlone is card AC #3's RM-1 test: a
// mind may durably claim direct perception only where the backing percept's
// origin is in DIRECT_ORIGINS (§2.7), and the classifier reads nothing else
// — an unrecognized or absent origin (V-6) classifies secondhand.
func TestRM1_DirectPerceptionGatesOnOriginAlone(t *testing.T) {
	for _, origin := range []string{"acted", "saw", "heard", "felt"} {
		if !DirectPerception(origin) {
			t.Errorf("DirectPerception(%q) = false, want true (§2.7 DIRECT_ORIGINS)", origin)
		}
	}
	for _, origin := range []string{"told", "read", "", "some-future-origin"} {
		if DirectPerception(origin) {
			t.Errorf("DirectPerception(%q) = true, want false — secondhand or unrecognized (EH-2b/V-6)", origin)
		}
	}
}

// TestRM2RM3_CitationGateCoercesNeverRejects builds a small log with one
// direct and one secondhand percept and drives ResolveCitations against
// both, proving the gate degrades a claim to what its citations actually
// resolve to (RM-2) and always returns a usable Provenance rather than an
// error (RM-3) — including when nothing cited resolves at all.
func TestRM2RM3_CitationGateCoercesNeverRejects(t *testing.T) {
	log, err := Open(PathFor(t.TempDir(), "b-tam"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer log.Close()

	if _, err := log.Append(EventInput{WorldTime: 1, Origin: "saw", PerceptID: "p-1", PerceptType: "sighting", ReceivedAt: 1, ObservedAt: obs(1), Content: map[string]any{"kind": "k-oak"}}); err != nil {
		t.Fatalf("Append: %v", err)
	}
	if _, err := log.Append(EventInput{WorldTime: 2, Origin: "told", PerceptID: "p-2", PerceptType: "told_fact", ReceivedAt: 2, ObservedAt: obs(2), Content: map[string]any{"kind": "k-oak"}}); err != nil {
		t.Fatalf("Append: %v", err)
	}
	events := log.Events()

	cases := []struct {
		name        string
		claimed     Provenance
		citations   []string
		wantResolve Provenance
		wantCoerced bool
	}{
		{"witnessed backed by direct percept holds", Witnessed, []string{"p-1"}, Witnessed, false},
		{"witnessed backed only by secondhand percept degrades to told", Witnessed, []string{"p-2"}, Told, true},
		{"witnessed backed by nothing resolvable degrades to inferred", Witnessed, []string{"p-nonexistent"}, Inferred, true},
		{"told with no citations degrades to inferred", Told, nil, Inferred, true},
		{"inferred never needs coercion", Inferred, nil, Inferred, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, coerced := ResolveCitations(events, c.claimed, c.citations)
			if got != c.wantResolve || coerced != c.wantCoerced {
				t.Fatalf("ResolveCitations(%v, %v) = (%v, %v), want (%v, %v)", c.claimed, c.citations, got, coerced, c.wantResolve, c.wantCoerced)
			}
		})
	}
}

// TestRM4_SecondhandNeverOverwritesFresherOrEqualFirsthand is RM-4: a told
// (or inferred) claim upserts only where existing knowledge is absent or
// strictly staler than the incoming observed_at.
func TestRM4_SecondhandNeverOverwritesFresherOrEqualFirsthand(t *testing.T) {
	s := NewStore(DefaultConfig())

	if !s.Upsert(UpsertInput{Subject: "th-1", Kind: "person", Content: "seen firsthand", Provenance: Witnessed, ObservedAt: obs(100), Confidence: 1}) {
		t.Fatal("first upsert into an absent subject must always apply")
	}

	if s.Upsert(UpsertInput{Subject: "th-1", Kind: "person", Content: "told, same time", Provenance: Told, ObservedAt: obs(100), Confidence: 1}) {
		t.Fatal("told fact at equal freshness overwrote fresher-or-equal firsthand knowledge")
	}
	b, _ := s.Get("th-1")
	if b.Provenance() != Witnessed {
		t.Fatalf("belief provenance = %v after blocked upsert, want unchanged Witnessed", b.Provenance())
	}

	if !s.Upsert(UpsertInput{Subject: "th-1", Kind: "person", Content: "told, later", Provenance: Told, ObservedAt: obs(150), Confidence: 1}) {
		t.Fatal("told fact strictly fresher than stored firsthand must apply (RM-4: own knowledge staler)")
	}
	b, _ = s.Get("th-1")
	if b.Provenance() != Told || *b.ObservedAt() != 150 {
		t.Fatalf("belief = (%v, %v) after fresher told upsert, want (Told, 150)", b.Provenance(), b.ObservedAt())
	}

	// A nil (maximally stale) observed_at is never "fresher" than anything,
	// but still applies where no belief exists yet at all.
	if !s.Upsert(UpsertInput{Subject: "th-2", Kind: "person", Content: "told, unknown when", Provenance: Told, ObservedAt: nil, Confidence: 1}) {
		t.Fatal("upsert into an absent subject must apply even with observed_at nil")
	}
}

// TestRM5_StoredConfidenceNeverMutatesEffectiveDecaysAtReadTime is RM-5:
// stored confidence is stamped once and never changes; the effective value
// a caller reads is computed fresh each time from age and half-life, and
// falls below the floor without deleting anything.
func TestRM5_StoredConfidenceNeverMutatesEffectiveDecaysAtReadTime(t *testing.T) {
	cfg := Config{DefaultHalfLife: 1000, DefaultHorizon: 1000, ConfidenceFloor: 0.05}
	s := NewStore(cfg)
	s.Upsert(UpsertInput{Subject: "th-1", Kind: "conviction", Content: nil, Provenance: Witnessed, ObservedAt: obs(1000), Confidence: 1.0})

	if c, _ := s.Confidence("th-1", 1000); c != 1.0 {
		t.Fatalf("Confidence at age 0 = %v, want 1.0", c)
	}
	if c, _ := s.Confidence("th-1", 2000); c < 0.49 || c > 0.51 {
		t.Fatalf("Confidence at one half-life = %v, want ~0.5", c)
	}
	if active, _ := s.Active("th-1", 2000+7000); active {
		t.Fatalf("Active at 7 half-lives past = true, want false (below floor %v)", cfg.ConfidenceFloor)
	}

	b, _ := s.Get("th-1")
	if b.StoredConfidence() != 1.0 {
		t.Fatalf("StoredConfidence after reads at multiple ages = %v, want unchanged 1.0", b.StoredConfidence())
	}
}

// TestRM6_FreshnessIsPerKindReadTimeArithmetic is RM-6: fresh iff
// now - observed_at < horizon(kind), evaluated at read time and per kind;
// nil observed_at is maximally stale and never fresh.
func TestRM6_FreshnessIsPerKindReadTimeArithmetic(t *testing.T) {
	cfg := Config{Horizon: map[string]int64{"place": 500}, DefaultHorizon: 100, DefaultHalfLife: 1000}
	s := NewStore(cfg)
	s.Upsert(UpsertInput{Subject: "pl-1", Kind: "place", Content: nil, Provenance: Witnessed, ObservedAt: obs(1000), Confidence: 1})
	s.Upsert(UpsertInput{Subject: "pl-2", Kind: "place", Content: nil, Provenance: Told, ObservedAt: nil, Confidence: 1})

	if fresh, _ := s.Fresh("pl-1", 1499); !fresh {
		t.Fatal("Fresh at age 499 < horizon 500 = false, want true")
	}
	if fresh, _ := s.Fresh("pl-1", 1500); fresh {
		t.Fatal("Fresh at age 500 == horizon 500 = true, want false (not strictly less)")
	}
	if fresh, _ := s.Fresh("pl-2", 1000); fresh {
		t.Fatal("Fresh with nil observed_at = true, want false — maximally stale is never fresh")
	}
}

// TestRM7_DeletionOnlyViaCorrectionDeathOrWitnessedRemoval proves staleness
// hides a belief but never deletes it, and that the only removal channel is
// Retract with a cause from RM-7's closed vocabulary.
func TestRM7_DeletionOnlyViaCorrectionDeathOrWitnessedRemoval(t *testing.T) {
	cfg := Config{DefaultHalfLife: 10, DefaultHorizon: 10, ConfidenceFloor: 0.05}
	s := NewStore(cfg)
	s.Upsert(UpsertInput{Subject: "th-1", Kind: "person", Content: "x", Provenance: Witnessed, ObservedAt: obs(0), Confidence: 1})

	far := int64(100_000)
	if fresh, _ := s.Fresh("th-1", far); fresh {
		t.Fatal("expected staleness at 100000 world-time units past a horizon of 10")
	}
	if active, _ := s.Active("th-1", far); active {
		t.Fatal("expected effective confidence below floor at that remove")
	}
	if _, ok := s.Get("th-1"); !ok {
		t.Fatal("staleness deleted the belief; RM-7 requires it only hide, never delete")
	}

	if s.Retract("th-1", RetractionCause("timer_expired")) {
		t.Fatal("Retract applied with a cause outside RM-7's closed vocabulary")
	}
	if _, ok := s.Get("th-1"); !ok {
		t.Fatal("an invalid retraction cause deleted the belief")
	}

	if !s.Retract("th-1", Death) {
		t.Fatal("Retract with a valid RM-7 cause did not apply")
	}
	if _, ok := s.Get("th-1"); ok {
		t.Fatal("belief still present after a valid Retract")
	}
}
