// h5_test.go (this file): H-5 (T006, specs/015-fake-vendor-harness Phase
// 3) — target_gone is the only non-existence channel
// (body-protocol-v0.md §5.3/§5.6, §10.3). An intent naming a token this
// vendor issued, whose referent is gone, is accepted at ack and fails
// only after advance() via Resolve's target_gone; an intent naming a
// token never issued is refused unknown_target synchronously, at ack.
// Same from-the-outside pattern as harness_test.go and flood_test.go.
package fakevendor_test

import (
	"testing"

	"kithcraft/mind/fakevendor"
)

func strPtr(s string) *string { return &s }

// sightingWithThing issues thingID to the mind, the same way §10.2's step
// 1 does — a target naming this token later is not "unissued".
func sightingWithThing(id, thingID string) map[string]any {
	return map[string]any{
		"percept_id": id, "percept_type": "sighting", "urgency": "background",
		"provenance": map[string]any{"origin": "saw", "source": nil, "observed_at": int64(0), "received_at": int64(0)},
		"place":      map[string]any{"place": "pl-3a91", "descriptor": "the well"},
		"content":    map[string]any{"thing": map[string]any{"thing_id": thingID, "kind": "k:person", "roles": []any{}, "descriptor": "Tam"}},
	}
}

func thingTargetIntent(session, body, intentID, thingID string, seq int64) map[string]any {
	return map[string]any{
		"protocol": "0.1", "message": "intent", "session": session,
		"seq": seq, "body": body, "world_time": int64(0),
		"payload": map[string]any{
			"intent_id": intentID, "verb": "go_to",
			"target": map[string]any{"type": "thing", "thing": thingID},
			"reason":  "to find Tam",
		},
	}
}

// TestH5_IssuedTokenGone_AcceptedThenFailsAfterAdvance is H-5's first half
// (§5.3/§5.6): an intent targeting th-401 — a token this vendor issued via
// an earlier sighting — is acked accepted:true immediately, stays pending
// across advance(), and fails only then, as act_result/failed/target_gone.
// What ships without this rule: a vendor that peeks at current world state
// before acking and refuses a gone target synchronously — an existence
// oracle a mind could poll without moving a foot (§5.3, §5.6). Mutation
// check: that exact bug (a synchronous "gone" answer at ack) is what would
// turn the accepted:true assertion below red.
func TestH5_IssuedTokenGone_AcceptedThenFailsAfterAdvance(t *testing.T) {
	vendorSide, mind := newPair(t)
	v := fakevendor.New(vendorSide, "s-1", "b-1", coreManifest())
	if err := v.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	mind.next(t) // session_open

	// Issue th-401 by sighting it, same as §10.2's step 1.
	if err := v.Emit(sightingWithThing("p-h5-sight", "th-401")); err != nil {
		t.Fatalf("Emit: %v", err)
	}
	mind.next(t) // the sighting itself

	mind.send(t, thingTargetIntent("s-1", "b-1", "i-gone", "th-401", 1))
	ack := mind.next(t)
	if ack["message"] != "intent_ack" {
		t.Fatalf("reply = %v, want intent_ack", ack["message"])
	}
	apl, _ := ack["payload"].(map[string]any)
	if apl["accepted"] != true {
		t.Fatalf("ack for an issued (even if now-gone) token = %#v, want accepted:true — existence is not an ack-time question (§5.3/§5.6)", apl)
	}

	mind.expectNone(t) // no act_result yet: the failure never arrives synchronously

	v.Advance(600)
	if err := v.Resolve("i-gone", "failed", strPtr("target_gone"), strPtr("Tam was gone")); err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	result := mind.next(t)
	rpl, _ := result["payload"].(map[string]any)
	if rpl["percept_type"] != "act_result" {
		t.Fatalf("payload.percept_type = %v, want act_result", rpl["percept_type"])
	}
	content, _ := rpl["content"].(map[string]any)
	if content["outcome"] != "failed" || content["reason_code"] != "target_gone" {
		t.Fatalf("act_result content = %#v, want outcome:failed reason_code:target_gone", content)
	}
}

// TestH5_UnissuedToken_RefusedUnknownTargetAtAck is H-5's second half: a
// token this vendor never mentioned in any percept is refused
// unknown_target synchronously, at ack — never accepted, never pending,
// never producing an act_result. What ships without this rule: a vendor
// that accepts any string as a target, indistinguishable from one that has
// actually verified the token names something real (§5.3).
func TestH5_UnissuedToken_RefusedUnknownTargetAtAck(t *testing.T) {
	vendorSide, mind := newPair(t)
	v := fakevendor.New(vendorSide, "s-1", "b-1", coreManifest())
	if err := v.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	mind.next(t) // session_open

	mind.send(t, thingTargetIntent("s-1", "b-1", "i-unknown", "th-999", 1))
	ack := mind.next(t)
	apl, _ := ack["payload"].(map[string]any)
	if apl["accepted"] != false || apl["reason_code"] != "unknown_target" {
		t.Fatalf("ack for an unissued token = %#v, want accepted:false reason_code:unknown_target", apl)
	}

	mind.expectNone(t) // refused at ack: nothing further ever arrives for it

	// Mutation check: this test's sibling asserts accepted:true for an
	// *issued*-but-gone token. If the ack refusal above ever collapsed
	// (an unissued token wrongly accepted), it would also become pending,
	// and Resolve would succeed instead of erroring — silently losing the
	// H-5 distinction between the two channels.
	if err := v.Resolve("i-unknown", "completed", nil, nil); err == nil {
		t.Fatal("an unissued, refused intent_id must never be pending — Resolve on it must error, not succeed")
	}
}
