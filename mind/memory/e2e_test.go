// Package memory (this file): the protocol §10.2 canonical end-to-end,
// driven against this package's own API. M4/M5's model-authored belief
// step and the act surface are out of scope here (spec FR-007 — no
// deliberation, no consolidation prompt); this test exercises the
// memory-surface half §10.2 actually specifies for this task: the
// percepts the script sends, gated by Gate, logged by Log, and folded
// into the belief store exactly as a later consolidation pass would.
// Step 5 is the point: a claim that the mind *saw* apple trees at the
// old orchard, citing only the told_fact that mentioned them, is
// mechanically coerced down to "told" and never durably stored as
// witnessed (card AC #5). specs/010-event-sourced-memory Phase 3 (T010).
package memory

import "testing"

func TestEndToEnd_ProtocolSection10_2_ToldCannotBecomeWitnessed(t *testing.T) {
	log, err := Open(PathFor(t.TempDir(), "b-eda"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer log.Close()
	gate := NewGate()
	store := NewStore(DefaultConfig())

	// 1. The mind learns a place and a person exist.
	sighting := percept("sighting", "background", selfProvenance("saw"),
		map[string]any{"place": "pl-3a91", "descriptor": "the well"},
		map[string]any{"thing": map[string]any{"thing_id": "th-401", "kind": "k:person", "body": "b-tam", "descriptor": "Tam"}, "distance": "near"})
	if admit, rule := gate.Decide(sighting); !admit || rule != RuleOtherBody {
		t.Fatalf("sighting of Tam: Decide = (%v, %v), want (true, RuleOtherBody)", admit, rule)
	}
	mustAppend(t, log, sighting, 1000)

	observation := percept("observation", "notable", selfProvenance("saw"),
		map[string]any{"place": "pl-3a91", "descriptor": "the well"},
		map[string]any{
			"extent":     "near",
			"vocabulary": []any{"k:sleeping-place", "k:water-source", "k:person"},
			"present":    []any{map[string]any{"thing_id": "th-9", "kind": "k:water-source", "descriptor": "the well itself", "count": int64(1)}},
		})
	if admit, rule := gate.Decide(observation); !admit || rule != RuleUrgency {
		t.Fatalf("observation: Decide = (%v, %v), want (true, RuleUrgency)", admit, rule)
	}
	mustAppend(t, log, observation, 1001)
	if !absentFromObservation(observation["content"].(map[string]any), "k:sleeping-place") {
		t.Fatal("§4.3 absence claim: k:sleeping-place is in vocabulary but not present, so it must read as absent")
	}

	// 2. Tam tells the mind about somewhere the mind has never been.
	tellerProvenance := map[string]any{
		"origin": "told", "source": map[string]any{"kind": "body", "body": "b-tam", "descriptor": "Tam"},
		"observed_at": int64(902430), "received_at": int64(918251), "hops": int64(1),
	}
	toldFact := percept("told_fact", "background", tellerProvenance,
		map[string]any{"place": "pl-3a91", "descriptor": "the well"},
		map[string]any{
			"about_place": map[string]any{"place": "pl-51c", "descriptor": "the old orchard"},
			"thing":       map[string]any{"thing_id": nil, "kind": "k:food-source", "descriptor": "apple trees"},
			"assertion":   "present",
		})
	toldFact["percept_id"] = "p-told-orchard"
	if admit, rule := gate.Decide(toldFact); !admit || rule != RuleToldOrText {
		t.Fatalf("told_fact: Decide = (%v, %v), want (true, RuleToldOrText)", admit, rule)
	}
	mustAppend(t, log, toldFact, 918251)

	const subject = "pl-51c:k:food-source" // opaque subject key: the food source at the old orchard
	events := log.Events()

	// 3. The mind honestly forms a "told" belief citing the told_fact —
	// RM-1: it does not claim to have seen the orchard itself.
	resolved, coerced := ResolveCitations(events, Told, []string{"p-told-orchard"})
	if resolved != Told || coerced {
		t.Fatalf("honest told claim: ResolveCitations = (%v, %v), want (Told, false)", resolved, coerced)
	}
	tellerObservedAt := int64(902430)
	if !store.Upsert(UpsertInput{Subject: subject, Kind: "food-source", Content: toldFact["content"], Provenance: resolved, ObservedAt: &tellerObservedAt, Confidence: 0.6}) {
		t.Fatal("first upsert into an absent subject must apply")
	}

	// 5a. Step 5's first half: the belief is secondhand.
	belief, ok := store.Get(subject)
	if !ok || belief.Provenance() != Told {
		t.Fatalf("belief at %q = (%v, %v), want (Told, true)", subject, belief.Provenance(), ok)
	}

	// 5b. Step 5's point: a claim that the mind *saw* apple trees there,
	// citing only the told_fact, is mechanically coerced — never
	// rejected (RM-3), never honored as witnessed (RM-1/RM-2).
	sawResolved, sawCoerced := ResolveCitations(events, Witnessed, []string{"p-told-orchard"})
	if sawResolved != Told || !sawCoerced {
		t.Fatalf("claimed-witnessed citing only a told percept: ResolveCitations = (%v, %v), want (Told, true)", sawResolved, sawCoerced)
	}
	store.Upsert(UpsertInput{Subject: subject, Kind: "food-source", Content: "apple trees", Provenance: sawResolved, ObservedAt: &tellerObservedAt, Confidence: 0.6})
	if b, _ := store.Get(subject); b.Provenance() == Witnessed {
		t.Fatal("belief durably claims Witnessed from a told percept alone — card AC #5 violated")
	}

	// 4. It gets there and finds the telling was stale: a real, direct
	// observation of the orchard, bare.
	orchardObservation := percept("observation", "notable", selfProvenance("saw"),
		map[string]any{"place": "pl-51c", "descriptor": "the old orchard"},
		map[string]any{"extent": "near", "vocabulary": []any{"k:food-source", "k:water-source"}, "present": []any{}})
	if admit, rule := gate.Decide(orchardObservation); !admit || rule != RuleUrgency {
		t.Fatalf("orchard observation: Decide = (%v, %v), want (true, RuleUrgency)", admit, rule)
	}
	mustAppend(t, log, orchardObservation, 918851)
	if !absentFromObservation(orchardObservation["content"].(map[string]any), "k:food-source") {
		t.Fatal("the orchard observation must record k:food-source as absent — the telling was stale")
	}

	freshAt := int64(918851)
	if !store.Upsert(UpsertInput{Subject: subject, Kind: "food-source", Content: "absent, directly observed", Provenance: Witnessed, ObservedAt: &freshAt, Confidence: 1.0}) {
		t.Fatal("a direct, fresher observation must always apply (RM-4)")
	}
	b, _ := store.Get(subject)
	if b.Provenance() != Witnessed {
		t.Fatalf("after the direct observation, belief provenance = %v, want Witnessed", b.Provenance())
	}
	if b.Content() == "apple trees" {
		t.Fatal("the witnessed correction must not still read the earlier told content")
	}
}

// mustAppend logs percept as one memory event at worldTime, using its
// own provenance fields for origin/observed_at/received_at — the same
// translation an ingest flow does after Gate.Decide admits.
func mustAppend(t *testing.T, log *Log, p map[string]any, worldTime int64) {
	t.Helper()
	provenance, _ := p["provenance"].(map[string]any)
	origin, _ := provenance["origin"].(string)
	var observedAt *int64
	if v, ok := provenance["observed_at"].(int64); ok {
		observedAt = &v
	}
	perceptID, _ := p["percept_id"].(string)
	perceptType, _ := p["percept_type"].(string)
	if _, err := log.Append(EventInput{
		WorldTime: worldTime, Origin: origin, PerceptID: perceptID, PerceptType: perceptType,
		ReceivedAt: worldTime, ObservedAt: observedAt, Content: p["content"],
	}); err != nil {
		t.Fatalf("Append(%s): %v", perceptID, err)
	}
}

// absentFromObservation is §4.3's absence mechanism: a kind is absent
// from an observation iff it is in vocabulary and not in present.
func absentFromObservation(content map[string]any, kind string) bool {
	vocab, _ := content["vocabulary"].([]any)
	inVocab := false
	for _, v := range vocab {
		if s, _ := v.(string); s == kind {
			inVocab = true
		}
	}
	if !inVocab {
		return false
	}
	present, _ := content["present"].([]any)
	for _, item := range present {
		m, _ := item.(map[string]any)
		if s, _ := m["kind"].(string); s == kind {
			return false
		}
	}
	return true
}
