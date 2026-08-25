package memory

import "testing"

func percept(perceptType, urgency string, provenance, place, content map[string]any) map[string]any {
	return map[string]any{
		"percept_id": "p-x", "percept_type": perceptType, "urgency": urgency,
		"provenance": provenance, "place": place, "content": content,
	}
}

func selfProvenance(origin string) map[string]any {
	return map[string]any{"origin": origin, "source": nil, "observed_at": int64(1), "received_at": int64(1)}
}

func bodyProvenance(origin, body string) map[string]any {
	return map[string]any{
		"origin": origin, "source": map[string]any{"kind": "body", "body": body, "descriptor": "Tam"},
		"observed_at": int64(1), "received_at": int64(1),
	}
}

// TestGate_AdmitsOnUrgencyAtLeastNotable is §6.3's first rule: urgency
// notable or urgent admits regardless of anything else about the percept.
func TestGate_AdmitsOnUrgencyAtLeastNotable(t *testing.T) {
	g := NewGate()
	for _, urgency := range []string{"notable", "urgent"} {
		p := percept("self_state", urgency, selfProvenance("felt"), nil, map[string]any{"condition": "cold"})
		if admit, rule := g.Decide(p); !admit || rule != RuleUrgency {
			t.Errorf("Decide(urgency=%s) = (%v, %v), want (true, RuleUrgency)", urgency, admit, rule)
		}
	}
	if admit, _ := g.Decide(percept("self_state", "background", selfProvenance("felt"), nil, map[string]any{"condition": "cold"})); admit {
		t.Error("background self_state with no other qualifying rule admitted; want false")
	}
}

// TestGate_AdmitsPerceptInvolvingOtherBodyOrPlayer covers both channels
// §6.3 means by "involving another body or the player": a sighting of a
// minded body, and a percept whose provenance source is a body or a
// person (the player crosses as a person source, §4.5).
func TestGate_AdmitsPerceptInvolvingOtherBodyOrPlayer(t *testing.T) {
	g := NewGate()
	sighting := percept("sighting", "background", selfProvenance("saw"),
		map[string]any{"place": "pl-3a91"},
		map[string]any{"thing": map[string]any{"thing_id": "th-401", "kind": "k:person", "body": "b-tam"}, "distance": "near"})
	if admit, rule := g.Decide(sighting); !admit || rule != RuleOtherBody {
		t.Fatalf("Decide(sighting of another body) = (%v, %v), want (true, RuleOtherBody)", admit, rule)
	}

	g2 := NewGate()
	fromPlayer := percept("speech", "background", bodyProvenance("told", "b-tam"), nil, map[string]any{"utterance": "hello"})
	if admit, rule := g2.Decide(fromPlayer); !admit || rule != RuleOtherBody {
		t.Fatalf("Decide(speech sourced from a body) = (%v, %v), want (true, RuleOtherBody)", admit, rule)
	}
}

// TestGate_AdmitsActResultWithAuthoredReason: only when `reason` is
// present and non-empty (§5.2, §5.4) — an act with no authored reason
// gets no special pass.
func TestGate_AdmitsActResultWithAuthoredReason(t *testing.T) {
	g := NewGate()
	withReason := percept("act_result", "background", selfProvenance("acted"), nil,
		map[string]any{"intent_id": "i-1", "verb": "go_to", "outcome": "completed", "reason": "to see the orchard"})
	if admit, rule := g.Decide(withReason); !admit || rule != RuleActedReason {
		t.Fatalf("Decide(act_result with reason) = (%v, %v), want (true, RuleActedReason)", admit, rule)
	}

	withoutReason := percept("act_result", "background", selfProvenance("acted"), nil,
		map[string]any{"intent_id": "i-2", "verb": "wait", "outcome": "completed", "reason": nil})
	if admit, rule := g.Decide(withoutReason); admit {
		t.Fatalf("Decide(act_result without a reason, background) = (%v, %v), want (false, \"\")", admit, rule)
	}
}

// TestGate_AdmitsToldFactAndText: both percept types always admit,
// regardless of urgency (routing §6.3 lists them unconditionally).
func TestGate_AdmitsToldFactAndText(t *testing.T) {
	g := NewGate()
	told := percept("told_fact", "background", bodyProvenance("told", "b-tam"),
		map[string]any{"place": "pl-3a91"},
		map[string]any{"about_place": map[string]any{"place": "pl-51c"}, "thing": map[string]any{"kind": "k:food-source"}, "assertion": "present"})
	if admit, rule := g.Decide(told); !admit || rule != RuleToldOrText {
		t.Fatalf("Decide(told_fact) = (%v, %v), want (true, RuleToldOrText)", admit, rule)
	}

	text := percept("text", "background", map[string]any{"origin": "read", "source": nil, "observed_at": nil, "received_at": int64(1)},
		map[string]any{"place": "pl-002"}, map[string]any{"text": "Build a shelter", "attributed_to": "the player"})
	if admit, rule := g.Decide(text); !admit || rule != RuleToldOrText {
		t.Fatalf("Decide(text) = (%v, %v), want (true, RuleToldOrText)", admit, rule)
	}
}

// TestGate_AdmitsFirstSightingOfKindOrPlace: a background sighting of a
// kind and place the gate has never seen before admits on that rule
// alone.
func TestGate_AdmitsFirstSightingOfKindOrPlace(t *testing.T) {
	g := NewGate()
	p := percept("sighting", "background", selfProvenance("saw"),
		map[string]any{"place": "pl-9"}, map[string]any{"thing": map[string]any{"kind": "k:sleeping-place"}, "distance": "near"})
	if admit, rule := g.Decide(p); !admit || rule != RuleFirstSighting {
		t.Fatalf("Decide(first sighting) = (%v, %v), want (true, RuleFirstSighting)", admit, rule)
	}
}

// TestGate_DropsRepeatedBackgroundSightingOfKnownThing is the drop rule:
// once a kind/place has been sighted, a later background sighting of the
// same thing (no other qualifying rule) is dropped.
func TestGate_DropsRepeatedBackgroundSightingOfKnownThing(t *testing.T) {
	g := NewGate()
	sightingOf := func() map[string]any {
		return percept("sighting", "background", selfProvenance("saw"),
			map[string]any{"place": "pl-9"}, map[string]any{"thing": map[string]any{"kind": "k:tree"}, "distance": "near"})
	}
	if admit, rule := g.Decide(sightingOf()); !admit || rule != RuleFirstSighting {
		t.Fatalf("first sighting: Decide = (%v, %v), want (true, RuleFirstSighting)", admit, rule)
	}
	if admit, rule := g.Decide(sightingOf()); admit {
		t.Fatalf("repeated background sighting of a known kind/place = (%v, %q), want (false, \"\") — the §6.3 drop rule", admit, rule)
	}
}

// TestGate_Deterministic_SameStreamSameAdmissions is §6.3's own closing
// requirement: no model call anywhere, so the same percept stream fed to
// two fresh gates must produce identical admit/rule decisions in order.
func TestGate_Deterministic_SameStreamSameAdmissions(t *testing.T) {
	stream := []map[string]any{
		percept("sighting", "background", selfProvenance("saw"), map[string]any{"place": "pl-1"}, map[string]any{"thing": map[string]any{"kind": "k:tree"}}),
		percept("sighting", "background", selfProvenance("saw"), map[string]any{"place": "pl-1"}, map[string]any{"thing": map[string]any{"kind": "k:tree"}}),
		percept("observation", "notable", selfProvenance("saw"), map[string]any{"place": "pl-1"}, map[string]any{"vocabulary": []any{"k:tree"}, "present": []any{}}),
		percept("told_fact", "background", bodyProvenance("told", "b-tam"), nil, map[string]any{"about_place": map[string]any{"place": "pl-2"}}),
		percept("act_result", "background", selfProvenance("acted"), nil, map[string]any{"reason": nil}),
		percept("self_state", "urgent", selfProvenance("felt"), nil, map[string]any{"condition": "cold"}),
	}

	run := func() []bool {
		g := NewGate()
		got := make([]bool, len(stream))
		for i, p := range stream {
			admit, _ := g.Decide(p)
			got[i] = admit
		}
		return got
	}

	a, b := run(), run()
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("percept %d: run 1 admit=%v, run 2 admit=%v — the gate is not deterministic", i, a[i], b[i])
		}
	}
}
