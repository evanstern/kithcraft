// e2e_test.go (this file): the §10.2 canonical end-to-end (T008,
// specs/015-fake-vendor-harness Phase 4) — closing TASK-0010's AC #5 at
// its literal wording: "the section-10.2 canonical end-to-end against the
// fake vendor." mind/memory/e2e_test.go already proved this same story
// against the memory package directly; this file proves it again with
// every percept crossing FakeVendor's real seam wire first — that
// difference is the whole point of this test existing beside that one.
// Percept ingest, admission (memory.Gate), and belief formation
// (memory.Log/Store + ResolveCitations) all run the real mind machinery;
// only step 3's go_to intent has no author yet (M5's deliberation engine
// doesn't exist), so — the same license mind/memory/e2e_test.go takes for
// the belief-authoring half of this same gap — the test itself plays the
// mind's act, sent as a real intent over the real wire and ack'd/resolved
// by FakeVendor like any other.
//
// Steps 1/2/3/5a/5b/4 below follow mind/memory/e2e_test.go's own
// numbering and its reason for it: step 5's epistemic assertions are
// checked right after the told belief forms, before step 4's fresh,
// direct correction would overwrite it — otherwise "secondhand" would
// already be stale by the time it's asserted.
package fakevendor_test

import (
	"testing"

	"kithcraft/mind/fakevendor"
	"kithcraft/mind/memory"
)

// e2eAbsentFromObservation is §4.3's absence mechanism, duplicated here
// (see mind/memory/e2e_test.go's identical helper) rather than exported
// from memory: each test package proves the rule for itself against its
// own wire-shaped content, never trusting a shared implementation to
// still be applying it correctly.
func e2eAbsentFromObservation(content map[string]any, kind string) bool {
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

// e2eAppend folds one wire-decoded percept message the mind side actually
// received into log — the same translation a real ingest flow performs
// after Gate.Decide admits (mirrors mind/memory/e2e_test.go's mustAppend,
// reading from a decoded wire envelope instead of a script's percept map,
// which is the point: this is what crossed the wire, not what the script
// intended to send).
func e2eAppend(t *testing.T, log *memory.Log, msg map[string]any) {
	t.Helper()
	payload, _ := msg["payload"].(map[string]any)
	provenance, _ := payload["provenance"].(map[string]any)
	origin, _ := provenance["origin"].(string)
	var observedAt *int64
	if v, ok := provenance["observed_at"].(int64); ok {
		observedAt = &v
	}
	perceptID, _ := payload["percept_id"].(string)
	perceptType, _ := payload["percept_type"].(string)
	worldTime, _ := msg["world_time"].(int64)
	if _, err := log.Append(memory.EventInput{
		WorldTime: worldTime, Origin: origin, PerceptID: perceptID, PerceptType: perceptType,
		ReceivedAt: worldTime, ObservedAt: observedAt, Content: payload["content"],
	}); err != nil {
		t.Fatalf("Append(%s): %v", perceptID, err)
	}
}

// TestE2E_ProtocolSection10_2_AgainstFakeVendor is US2's independent test
// (spec.md) and T008: the §10.2 script, run against FakeVendor over the
// real seam wire, closing TASK-0010's AC #5. What ships without it: the
// carry TASK-0010 left deliberately open across PR #15 — a mind that
// forms secondhand beliefs correctly when driven directly against
// mind/memory, unproven when driven the way an actual body will be, over
// the wire.
func TestE2E_ProtocolSection10_2_AgainstFakeVendor(t *testing.T) {
	vendorSide, mind := newPair(t)
	v := fakevendor.New(vendorSide, "s-e2e", "b-1", coreManifest())
	if err := v.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	mind.next(t) // session_open

	log, err := memory.Open(memory.PathFor(t.TempDir(), "b-e2e"))
	if err != nil {
		t.Fatalf("memory.Open: %v", err)
	}
	defer log.Close()
	gate := memory.NewGate()
	store := memory.NewStore(memory.DefaultConfig())

	// 1. The mind learns a place and a person exist.
	sighting := map[string]any{
		"percept_id": "p-sight-tam", "percept_type": "sighting", "urgency": "background",
		"provenance": map[string]any{"origin": "saw", "source": nil, "observed_at": int64(1000), "received_at": int64(1000)},
		"place":      map[string]any{"place": "pl-3a91", "descriptor": "the well"},
		"content": map[string]any{
			"thing":    map[string]any{"thing_id": "th-401", "kind": "k:person", "body": "b-tam", "descriptor": "Tam"},
			"distance": "near",
		},
	}
	if err := v.Emit(sighting); err != nil {
		t.Fatalf("Emit sighting: %v", err)
	}
	got := mind.next(t)
	payload, _ := got["payload"].(map[string]any)
	if admit, rule := gate.Decide(payload); !admit || rule != memory.RuleOtherBody {
		t.Fatalf("sighting of Tam: Decide = (%v, %v), want (true, RuleOtherBody)", admit, rule)
	}
	e2eAppend(t, log, got)

	observation := map[string]any{
		"percept_id": "p-obs-well", "percept_type": "observation", "urgency": "notable",
		"provenance": map[string]any{"origin": "saw", "source": nil, "observed_at": int64(1001), "received_at": int64(1001)},
		"place":      map[string]any{"place": "pl-3a91", "descriptor": "the well"},
		"content": map[string]any{
			"extent":     "near",
			"vocabulary": []any{"k:sleeping-place", "k:water-source", "k:person"},
			"present":    []any{map[string]any{"thing_id": "th-9", "kind": "k:water-source", "descriptor": "the well itself", "count": int64(1)}},
		},
	}
	if err := v.Emit(observation); err != nil {
		t.Fatalf("Emit observation: %v", err)
	}
	got = mind.next(t)
	payload, _ = got["payload"].(map[string]any)
	if admit, rule := gate.Decide(payload); !admit || rule != memory.RuleUrgency {
		t.Fatalf("observation: Decide = (%v, %v), want (true, RuleUrgency)", admit, rule)
	}
	if !e2eAbsentFromObservation(payload["content"].(map[string]any), "k:sleeping-place") {
		t.Fatal("§4.3 absence claim: k:sleeping-place is in vocabulary but not present, so it must read as absent")
	}
	e2eAppend(t, log, got)

	// 2. Tam tells the mind about somewhere the mind has never been.
	toldFact := map[string]any{
		"percept_id": "p-told-orchard", "percept_type": "told_fact", "urgency": "background",
		"provenance": map[string]any{
			"origin": "told", "source": map[string]any{"kind": "body", "body": "b-tam", "descriptor": "Tam"},
			"observed_at": int64(902430), "received_at": int64(918251), "hops": int64(1),
		},
		"place": map[string]any{"place": "pl-3a91", "descriptor": "the well"},
		"content": map[string]any{
			"about_place": map[string]any{"place": "pl-51c", "descriptor": "the old orchard"},
			"thing":       map[string]any{"thing_id": nil, "kind": "k:food-source", "descriptor": "apple trees"},
			"assertion":   "present",
		},
	}
	if err := v.Emit(toldFact); err != nil {
		t.Fatalf("Emit told_fact: %v", err)
	}
	got = mind.next(t)
	payload, _ = got["payload"].(map[string]any)
	if admit, rule := gate.Decide(payload); !admit || rule != memory.RuleToldOrText {
		t.Fatalf("told_fact: Decide = (%v, %v), want (true, RuleToldOrText)", admit, rule)
	}
	e2eAppend(t, log, got)

	const subject = "pl-51c:k:food-source" // opaque subject key: the food source at the old orchard

	// 3. The mind honestly forms a "told" belief citing the told_fact
	// (RM-1: it does not claim to have seen the orchard itself), then acts
	// on that secondhand knowledge — a go_to whose target is a place this
	// body has never itself visited, only heard of.
	events := log.Events()
	resolved, coerced := memory.ResolveCitations(events, memory.Told, []string{"p-told-orchard"})
	if resolved != memory.Told || coerced {
		t.Fatalf("honest told claim: ResolveCitations = (%v, %v), want (Told, false)", resolved, coerced)
	}
	tellerObservedAt := int64(902430)
	if !store.Upsert(memory.UpsertInput{Subject: subject, Kind: "food-source", Content: toldFact["content"], Provenance: resolved, ObservedAt: &tellerObservedAt, Confidence: 0.6}) {
		t.Fatal("first upsert into an absent subject must apply")
	}

	// No deliberation engine exists yet (M5): this intent has no model
	// author, so the test plays that part directly — but it still crosses
	// the real seam wire and is ack'd/resolved by FakeVendor like any
	// other intent.
	mind.send(t, map[string]any{
		"protocol": "0.1", "message": "intent", "session": "s-e2e",
		"seq": int64(0), "body": "b-1", "world_time": int64(918251),
		"payload": map[string]any{
			"intent_id": "i-go-orchard", "verb": "go_to",
			"target": map[string]any{"type": "place", "place": "pl-51c"},
			"reason":  "Tam said there are apple trees at the old orchard",
		},
	})
	ack := mind.next(t)
	ackPayload, _ := ack["payload"].(map[string]any)
	if ack["message"] != "intent_ack" || ackPayload["accepted"] != true {
		t.Fatalf("go_to ack = %v/%#v, want intent_ack/accepted:true — a mind acting on secondhand knowledge is not refused for it", ack["message"], ackPayload)
	}
	acts := v.Acts()
	last := acts[len(acts)-1]
	if last["verb"] != "go_to" {
		t.Fatalf("acts[-1].verb = %v, want go_to", last["verb"])
	}
	target, _ := last["target"].(map[string]any)
	if target["place"] != "pl-51c" {
		t.Fatalf("acts[-1].target.place = %v, want pl-51c", target["place"])
	}
	if reason, _ := last["reason"].(string); reason == "" {
		t.Fatal("acts[-1].reason must not be empty — §5.2: the mind authored a why")
	}

	// 5a. Step 5's first half: the belief is secondhand.
	belief, ok := store.Get(subject)
	if !ok || belief.Provenance() != memory.Told {
		t.Fatalf("belief at %q = (%v, %v), want (Told, true)", subject, belief.Provenance(), ok)
	}

	// 5b. Step 5's point: a claim that the mind *saw* apple trees there,
	// citing only the told_fact, is mechanically coerced — never rejected
	// (RM-3), never honored as witnessed (RM-1/RM-2). No witnessed claim
	// about apple trees at the old orchard exists.
	sawResolved, sawCoerced := memory.ResolveCitations(events, memory.Witnessed, []string{"p-told-orchard"})
	if sawResolved != memory.Told || !sawCoerced {
		t.Fatalf("claimed-witnessed citing only a told percept: ResolveCitations = (%v, %v), want (Told, true)", sawResolved, sawCoerced)
	}
	store.Upsert(memory.UpsertInput{Subject: subject, Kind: "food-source", Content: "apple trees", Provenance: sawResolved, ObservedAt: &tellerObservedAt, Confidence: 0.6})
	if b, _ := store.Get(subject); b.Provenance() == memory.Witnessed {
		t.Fatal("belief durably claims Witnessed from a told percept alone — card AC #5 violated")
	}

	// 4. It gets there and finds the telling was stale: advance, resolve
	// the go_to, and a real, direct observation of the orchard, bare.
	v.Advance(600)
	if err := v.Resolve("i-go-orchard", "completed", nil, nil); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	got = mind.next(t) // the act_result
	payload, _ = got["payload"].(map[string]any)
	if payload["percept_type"] != "act_result" {
		t.Fatalf("payload.percept_type = %v, want act_result", payload["percept_type"])
	}

	orchardObservation := map[string]any{
		"percept_id": "p-obs-orchard", "percept_type": "observation", "urgency": "notable",
		"provenance": map[string]any{"origin": "saw", "source": nil, "observed_at": int64(918851), "received_at": int64(918851)},
		"place":      map[string]any{"place": "pl-51c", "descriptor": "the old orchard"},
		"content":    map[string]any{"extent": "near", "vocabulary": []any{"k:food-source", "k:water-source"}, "present": []any{}},
	}
	if err := v.Emit(orchardObservation); err != nil {
		t.Fatalf("Emit orchard observation: %v", err)
	}
	got = mind.next(t)
	payload, _ = got["payload"].(map[string]any)
	if admit, rule := gate.Decide(payload); !admit || rule != memory.RuleUrgency {
		t.Fatalf("orchard observation: Decide = (%v, %v), want (true, RuleUrgency)", admit, rule)
	}
	if !e2eAbsentFromObservation(payload["content"].(map[string]any), "k:food-source") {
		t.Fatal("the orchard observation must record k:food-source as absent — the telling was stale")
	}
	e2eAppend(t, log, got)

	freshAt := int64(918851)
	if !store.Upsert(memory.UpsertInput{Subject: subject, Kind: "food-source", Content: "absent, directly observed", Provenance: memory.Witnessed, ObservedAt: &freshAt, Confidence: 1.0}) {
		t.Fatal("a direct, fresher observation must always apply (RM-4)")
	}
	b, _ := store.Get(subject)
	if b.Provenance() != memory.Witnessed {
		t.Fatalf("after the direct observation, belief provenance = %v, want Witnessed", b.Provenance())
	}
	if b.Content() == "apple trees" {
		t.Fatal("the witnessed correction must not still read the earlier told content")
	}
}
