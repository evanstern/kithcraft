package deliberate

import "testing"

func TestTokens_ObservePercept_RecordsPlaceThingBody(t *testing.T) {
	tok := NewTokens()
	tok.Observe(map[string]any{
		"percept_type": "sighting",
		"place":        map[string]any{"place": "pl-3a91", "descriptor": "the well"},
		"content": map[string]any{
			"thing": map[string]any{"thing_id": "th-401", "kind": "k:person", "body": "b-tam"},
		},
	})
	if !tok.Known("place", "pl-3a91") {
		t.Error("place token not recorded")
	}
	if !tok.Known("thing_id", "th-401") {
		t.Error("thing_id token not recorded")
	}
	if !tok.Known("body", "b-tam") {
		t.Error("body token not recorded")
	}
	if tok.Known("place", "pl-unseen") {
		t.Error("an unobserved token must not be known")
	}
}

func TestTokens_ValidateTarget_NilAndNoneAlwaysPass(t *testing.T) {
	tok := NewTokens()
	if err := tok.ValidateTarget(nil); err != nil {
		t.Errorf("nil target: %v", err)
	}
	if err := tok.ValidateTarget(map[string]any{"type": "none"}); err != nil {
		t.Errorf("type none target: %v", err)
	}
}

func TestTokens_ValidateTarget_RejectsDescriptiveString(t *testing.T) {
	tok := NewTokens()
	if err := tok.ValidateTarget("the nearest bed"); err == nil {
		t.Fatal("expected a bare descriptive string target to be rejected")
	}
}

func TestTokens_ValidateTarget_RejectsUnknownToken(t *testing.T) {
	tok := NewTokens()
	if err := tok.ValidateTarget(map[string]any{"type": "place", "place": "pl-1"}); err == nil {
		t.Fatal("expected an unobserved place token to be rejected")
	}
}

func TestTokens_ValidateTarget_AcceptsKnownToken(t *testing.T) {
	tok := NewTokens()
	tok.mark("place", "pl-1")
	if err := tok.ValidateTarget(map[string]any{"type": "place", "place": "pl-1"}); err != nil {
		t.Errorf("expected a known place token to pass: %v", err)
	}
}

func TestTokens_ValidateTarget_RejectsUnknownTargetType(t *testing.T) {
	tok := NewTokens()
	if err := tok.ValidateTarget(map[string]any{"type": "coordinate", "x": 1}); err == nil {
		t.Fatal("expected an unrecognized target type to be rejected")
	}
}
