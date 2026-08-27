// harness_test.go (this file): the cheap protocol-rule tests —
// H-1 (V-5, T004) and H-2/H-3/H-4 (V-6 and the classifier's purity, T005) —
// specs/015-fake-vendor-harness tasks.md Phase 2. Each test drives a
// scripted percept through the real FakeVendor -> seam.Conn wire (the same
// from-the-outside pattern as fakevendor_test.go) and asserts on what a
// "mind" (mindSide) actually receives, never on FakeVendor's internals.
// Each carries a mutation check: a small, clearly-wrong reference rule that
// currently disagrees with production on the same input. If the real rule
// were ever weakened to match the wrong one, the disagreement collapses and
// the mutation check itself goes red — which is what proves the primary
// assertion above it is pinned to a real guard, not a coincidence of this
// one input.
package fakevendor_test

import (
	"testing"

	"kithcraft/mind/fakevendor"
	"kithcraft/mind/seam"
	"kithcraft/mind/wire"
)

// rawPerceptEnvelope wraps payload as a full percept envelope matching what
// FakeVendor.envelope would produce — built independently here (rather than
// through FakeVendor) so the mutation checks below can run wire.Decode and
// wire.Validate directly against it, with nothing in between.
func rawPerceptEnvelope(payload map[string]any) map[string]any {
	return map[string]any{
		"protocol": "0.1", "message": "percept", "session": "s-1",
		"seq": int64(0), "body": "b-1", "world_time": int64(0),
		"payload": payload,
	}
}

// perceptWithProvenance builds a minimal, otherwise-valid percept carrying
// the given provenance — everything H-2/H-3/H-4 vary is inside provenance
// or content, never the percept's other required fields.
func perceptWithProvenance(id string, provenance map[string]any) map[string]any {
	return map[string]any{
		"percept_id": id, "percept_type": "sighting", "urgency": "background",
		"provenance": provenance, "place": nil, "content": map[string]any{},
	}
}

// originOf extracts payload.provenance.origin from a message the mind side
// actually received (never from the script that built it) — so a test
// asserting on it is asserting on what crossed the wire, not what was
// intended.
func originOf(t *testing.T, msg map[string]any) string {
	t.Helper()
	payload, ok := msg["payload"].(map[string]any)
	if !ok {
		t.Fatalf("payload missing or not an object: %#v", msg)
	}
	prov, ok := payload["provenance"].(map[string]any)
	if !ok {
		t.Fatalf("payload.provenance missing or not an object: %#v", payload)
	}
	origin, _ := prov["origin"].(string)
	return origin
}

// naiveOriginPresentIsDirect is the plausible-but-wrong classifier H-2's
// unknown-origin case guards against: treating any real-looking origin
// string as direct evidence just because it decoded fine, instead of
// checking it against the closed DIRECT_ORIGINS set (§2.7).
func naiveOriginPresentIsDirect(origin string) bool { return origin != "" }

// naiveEmptyOriginIsDirect is the plausible-but-wrong classifier H-2's
// absent-origin case guards against: defaulting "no information" to the
// body's own direct perception instead of the conservative secondhand
// floor (EH-2b).
func naiveEmptyOriginIsDirect(origin string) bool { return origin == "" }

// naiveHopsZeroIsDirect is the plausible-but-wrong classifier H-3 guards
// against: trusting hops==0 as proof of direct perception, instead of
// reading origin alone (§2.6: hops does not promote or demote origin).
func naiveHopsZeroIsDirect(hops int64) bool { return hops == 0 }

// naiveDirectFieldIsAuthoritative is the plausible-but-wrong classifier H-4
// guards against: trusting an unknown provenance.direct field as an
// override instead of ignoring it per V-1.
func naiveDirectFieldIsAuthoritative(direct bool) bool { return direct }

// TestH1_MissingProvenance_RejectedNeverDefaulted is H-1 / V-5's first case
// (§10.3): a percept with no provenance at all is rejected at the seam
// before anything reaches mind state — never given an assumed origin
// (EH-2a). What ships without this rule: "A percept quietly defaulted to
// *something*, and the 'unstamped is impossible' guarantee (EH-2) silently
// gone with the compiler that used to hold it."
func TestH1_MissingProvenance_RejectedNeverDefaulted(t *testing.T) {
	vendorSide, mind := newPair(t)
	v := fakevendor.New(vendorSide, "s-1", "b-1", coreManifest())
	v.Strict = true
	if err := v.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	mind.next(t) // session_open

	bad := map[string]any{
		"percept_id": "p-h1a", "percept_type": "sighting", "urgency": "background",
		"place": nil, "content": map[string]any{},
		// provenance: absent entirely
	}
	if err := v.Emit(bad); err != nil {
		t.Fatalf("Emit (the vendor writes whatever the script hands it): %v", err)
	}
	mind.expectNone(t) // rejected at the seam: nothing enters mind state

	// Mutation check: wire.Decode alone (no Validate) accepts this same
	// malformed payload — the rejection above is entirely Validate's rule
	// (V-5), nothing else in the pipeline catches it. If the Validate call
	// were ever dropped from wireConn.ReadMessage, this percept would
	// decode cleanly and mind.expectNone above would go red.
	body, err := wire.EncodeCanonical(rawPerceptEnvelope(bad))
	if err != nil {
		t.Fatalf("EncodeCanonical: %v", err)
	}
	decoded, err := wire.Decode(body)
	if err != nil {
		t.Fatalf("Decode alone must accept a shape-valid frame regardless of missing provenance: %v", err)
	}
	if err := wire.Validate(decoded); err == nil {
		t.Fatal("Validate must reject a percept with no provenance (V-5/EH-2a)")
	}
}

// TestH1_MissingProvenanceOrigin_RejectedNeverDefaulted is H-1's second
// case: provenance present but without origin is exactly as malformed as no
// provenance at all (EH-2a) — a provenance nobody stamped an origin on must
// not be treated as usable.
func TestH1_MissingProvenanceOrigin_RejectedNeverDefaulted(t *testing.T) {
	vendorSide, mind := newPair(t)
	v := fakevendor.New(vendorSide, "s-1", "b-1", coreManifest())
	v.Strict = true
	if err := v.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	mind.next(t) // session_open

	bad := perceptWithProvenance("p-h1b", map[string]any{
		// origin: absent
		"source": nil, "observed_at": int64(0), "received_at": int64(0),
	})
	if err := v.Emit(bad); err != nil {
		t.Fatalf("Emit: %v", err)
	}
	mind.expectNone(t) // rejected at the seam: nothing enters mind state

	// Mutation check: same shape as above, at the provenance sub-object
	// level. Decode alone accepts it; only Validate's deeper provenance
	// loop (mind/wire/decode.go) rejects a provenance missing origin.
	body, err := wire.EncodeCanonical(rawPerceptEnvelope(bad))
	if err != nil {
		t.Fatalf("EncodeCanonical: %v", err)
	}
	decoded, err := wire.Decode(body)
	if err != nil {
		t.Fatalf("Decode alone must accept this shape: %v", err)
	}
	if err := wire.Validate(decoded); err == nil {
		t.Fatal("Validate must reject provenance missing origin (V-5/EH-2a)")
	}
}

// TestH2_UnknownOrigin_ClassifiesSecondhand is H-2 / V-6's first case
// (§10.3): an origin value this daemon has never heard of — as a future
// minor-version vendor might mint — is V-2-tolerated (accepted, not
// rejected) and classifies secondhand, never direct. What ships without
// this rule: "a v0.2 vendor's new origin read as first-hand by a v0.1 mind
// — an omniscience bug arriving through a *non-breaking* change."
func TestH2_UnknownOrigin_ClassifiesSecondhand(t *testing.T) {
	vendorSide, mind := newPair(t)
	v := fakevendor.New(vendorSide, "s-1", "b-1", coreManifest())
	v.Strict = false // §10.3's H-2 posture (FakeVendor does not yet condition behaviour on it — Phase 1 ponytail)
	if err := v.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	mind.next(t) // session_open

	p := perceptWithProvenance("p-h2a", map[string]any{
		"origin": "dreamt", "source": nil, "observed_at": int64(0), "received_at": int64(0),
	})
	if err := v.Emit(p); err != nil {
		t.Fatalf("Emit: %v", err)
	}
	got := mind.next(t) // unrecognized origin is accepted, not rejected (V-2)
	origin := originOf(t, got)
	class := seam.ClassifyOrigin(origin)
	if class != "secondhand" {
		t.Fatalf("ClassifyOrigin(%q) = %q, want secondhand", origin, class)
	}

	// Mutation check: see naiveOriginPresentIsDirect's doc comment.
	if naive, real := naiveOriginPresentIsDirect(origin), class == "firsthand"; naive == real {
		t.Fatalf("mutation check: naive-non-empty-is-direct(%q)=%v no longer diverges from the real classifier — if production ever matched it, this test's classification assertion above would go red", origin, naive)
	}
}

// TestH2_AbsentOrigin_ClassifiesSecondhand is H-2's second case: origin
// present but empty (the classifier's zero value — mind/memory/provenance.go:
// "a map miss returns false") classifies secondhand exactly like an unknown
// value. This is a distinct wire shape from H-1's missing-origin-key case:
// the key is present (so V-5's presence check passes and the percept
// arrives), it simply carries no information — EH-2b's conservative default
// is what demotes it.
func TestH2_AbsentOrigin_ClassifiesSecondhand(t *testing.T) {
	vendorSide, mind := newPair(t)
	v := fakevendor.New(vendorSide, "s-1", "b-1", coreManifest())
	v.Strict = false
	if err := v.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	mind.next(t)

	p := perceptWithProvenance("p-h2b", map[string]any{
		"origin": "", "source": nil, "observed_at": int64(0), "received_at": int64(0),
	})
	if err := v.Emit(p); err != nil {
		t.Fatalf("Emit: %v", err)
	}
	got := mind.next(t) // origin present-but-empty passes V-5; V-6 is what demotes it
	origin := originOf(t, got)
	class := seam.ClassifyOrigin(origin)
	if class != "secondhand" {
		t.Fatalf("ClassifyOrigin(%q) = %q, want secondhand", origin, class)
	}

	// Mutation check: see naiveEmptyOriginIsDirect's doc comment.
	if naive, real := naiveEmptyOriginIsDirect(origin), class == "firsthand"; naive == real {
		t.Fatalf("mutation check: naive-empty-defaults-direct(%q)=%v no longer diverges from the real classifier — if production ever matched it, this test's classification assertion above would go red", origin, naive)
	}
}

// TestH3_ClassifierIgnoresProseHopsAndSource is H-3 (§10.3): a told percept
// whose prose swears "I saw this myself, directly, firsthand", with hops:0
// and source.descriptor:"saw", is still secondhand — the classifier reads
// origin and nothing else (§2.7). What ships without this rule: "A gate
// that can be talked out of its verdict by the text it is gating — which is
// precisely the LLM failure mode EH-3 exists to stop."
func TestH3_ClassifierIgnoresProseHopsAndSource(t *testing.T) {
	vendorSide, mind := newPair(t)
	v := fakevendor.New(vendorSide, "s-1", "b-1", coreManifest())
	if err := v.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	mind.next(t)

	p := map[string]any{
		"percept_id": "p-h3", "percept_type": "told_fact", "urgency": "notable",
		"provenance": map[string]any{
			"origin":      "told",
			"source":      map[string]any{"kind": "body", "body": "b-tam", "descriptor": "saw"},
			"observed_at": int64(0), "received_at": int64(0), "hops": int64(0),
		},
		"place":   nil,
		"content": map[string]any{"utterance": "I saw this myself, directly, firsthand"},
	}
	if err := v.Emit(p); err != nil {
		t.Fatalf("Emit: %v", err)
	}
	got := mind.next(t)
	origin := originOf(t, got)
	class := seam.ClassifyOrigin(origin)
	if class != "secondhand" {
		t.Fatalf("ClassifyOrigin(%q) = %q, want secondhand despite hops:0/descriptor \"saw\"/swearing prose", origin, class)
	}

	// Mutation check: see naiveHopsZeroIsDirect's doc comment.
	if naive, real := naiveHopsZeroIsDirect(0), class == "firsthand"; naive == real {
		t.Fatalf("mutation check: naive-hops-zero-is-direct=%v no longer diverges from the real classifier — if production ever trusted hops, this test's classification assertion above would go red", naive)
	}
}

// TestH4_UnknownDirectFieldIgnored is H-4 (§10.3): a "direct": true field
// riding alongside origin:"told" is tolerated as an unknown field (V-1) and
// ignored — classification still comes from origin alone. What ships
// without this rule: "A derived boolean that can disagree with its source,
// and a vendor able to grant first-hand status by asserting it."
func TestH4_UnknownDirectFieldIgnored(t *testing.T) {
	vendorSide, mind := newPair(t)
	v := fakevendor.New(vendorSide, "s-1", "b-1", coreManifest())
	if err := v.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	mind.next(t)

	p := map[string]any{
		"percept_id": "p-h4", "percept_type": "told_fact", "urgency": "notable",
		"provenance": map[string]any{
			"origin": "told", "direct": true,
			"source":      map[string]any{"kind": "body", "body": "b-tam", "descriptor": "Tam"},
			"observed_at": int64(0), "received_at": int64(0),
		},
		"place":   nil,
		"content": map[string]any{"utterance": "it happened"},
	}
	if err := v.Emit(p); err != nil {
		t.Fatalf("Emit: %v", err)
	}
	got := mind.next(t) // V-1: unknown field tolerated, not rejected
	origin := originOf(t, got)
	class := seam.ClassifyOrigin(origin)
	if class != "secondhand" {
		t.Fatalf("ClassifyOrigin(%q) = %q, want secondhand — provenance.direct must be ignored", origin, class)
	}

	// Confirm the unknown field actually rode the wire uninterpreted
	// (V-1 tolerates, never strips) rather than merely going unread by
	// coincidence.
	payload, _ := got["payload"].(map[string]any)
	prov, _ := payload["provenance"].(map[string]any)
	if d, ok := prov["direct"]; !ok || d != true {
		t.Fatalf("payload.provenance.direct = %#v, want true still present on the wire", prov["direct"])
	}

	// Mutation check: see naiveDirectFieldIsAuthoritative's doc comment.
	if naive, real := naiveDirectFieldIsAuthoritative(true), class == "firsthand"; naive == real {
		t.Fatalf("mutation check: naive-direct-field-is-authoritative=%v no longer diverges from the real classifier — if production ever read provenance.direct, this test's classification assertion above would go red", naive)
	}
}
