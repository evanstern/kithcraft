package seam

import "testing"

func TestPending_Compose_RefusesUndeclaredVerb(t *testing.T) {
	p := NewPending(map[string]bool{"go_to": true})
	if _, err := p.Compose("i-1", "fly", nil, "because", ""); err == nil {
		t.Fatal("expected Compose to refuse a verb the manifest doesn't declare (V-4)")
	}
	if p.IsPending("i-1") {
		t.Fatal("a refused intent must not enter the pending set")
	}
}

func TestPending_Compose_AddsToPendingSet(t *testing.T) {
	p := NewPending(map[string]bool{"go_to": true})
	if _, err := p.Compose("i-1", "go_to", map[string]any{"type": "place", "place": "pl-1"}, "why", ""); err != nil {
		t.Fatalf("Compose: %v", err)
	}
	if !p.IsPending("i-1") {
		t.Fatal("a composed intent must be pending")
	}
}

func TestPending_Supersedes_ReplacesPredecessor(t *testing.T) {
	p := NewPending(map[string]bool{"go_to": true})
	if _, err := p.Compose("i-1", "go_to", nil, "first", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := p.Compose("i-2", "go_to", nil, "changed my mind", "i-1"); err != nil {
		t.Fatal(err)
	}
	if p.IsPending("i-1") {
		t.Fatal("i-1 should have been replaced by i-2's supersedes")
	}
	if !p.IsPending("i-2") {
		t.Fatal("i-2 must be pending")
	}
}

func TestPending_ResolveActResult_MatchesByIntentID(t *testing.T) {
	p := NewPending(map[string]bool{"go_to": true})
	if _, err := p.Compose("i-1", "go_to", nil, "why", ""); err != nil {
		t.Fatal(err)
	}
	intent, ok := p.ResolveActResult("i-1")
	if !ok || intent["verb"] != "go_to" {
		t.Fatalf("ResolveActResult(i-1) = %#v, %v; want the composed intent, true", intent, ok)
	}
	if p.IsPending("i-1") {
		t.Fatal("a resolved intent must leave the pending set")
	}
}

func TestPending_ResolveActResult_UnknownIntentID_NotOK(t *testing.T) {
	p := NewPending(nil)
	if _, ok := p.ResolveActResult("i-ghost"); ok {
		t.Fatal("resolving an unknown intent_id must report ok=false, not fabricate a match")
	}
}

func TestPending_Cancel_RemovesFromPendingSet(t *testing.T) {
	p := NewPending(map[string]bool{"wait": true})
	if _, err := p.Compose("i-1", "wait", nil, "pausing", ""); err != nil {
		t.Fatal(err)
	}
	payload, ok := p.Cancel("i-1")
	if !ok || payload["intent_id"] != "i-1" {
		t.Fatalf("Cancel(i-1) = %#v, %v", payload, ok)
	}
	if p.IsPending("i-1") {
		t.Fatal("a canceled intent must leave the pending set")
	}
}

func TestPending_Cancel_NotPending_NotOK(t *testing.T) {
	p := NewPending(nil)
	if _, ok := p.Cancel("i-ghost"); ok {
		t.Fatal("canceling a non-pending intent must report ok=false")
	}
}
