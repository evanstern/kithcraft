package llm

import (
	"errors"
	"testing"
)

// TestParseIntentRoundTrip and TestParseIntentBoundedFailure prove AC
// scenario 2 (structured outputs are parsed values with a bounded failure
// mode): a valid payload parses, and an invalid one returns a typed
// *ParseError rather than panicking or silently returning a usable-looking
// zero value.
func TestParseIntentRoundTrip(t *testing.T) {
	got, err := ParseIntent(E2, `{"verb":"move","target":"well","reason":"thirsty"}`)
	if err != nil {
		t.Fatalf("ParseIntent: %v", err)
	}
	want := Intent{Verb: "move", Target: "well", Reason: "thirsty"}
	if got != want {
		t.Errorf("ParseIntent = %+v, want %+v", got, want)
	}
}

func TestParseIntentBoundedFailure(t *testing.T) {
	for _, raw := range []string{`not json`, `{}`, `{"verb":""}`} {
		_, err := ParseIntent(E2, raw)
		if err == nil {
			t.Fatalf("ParseIntent(%q): want error, got nil", raw)
		}
		var pe *ParseError
		if !errors.As(err, &pe) {
			t.Errorf("ParseIntent(%q) err = %T, want *ParseError", raw, err)
		}
	}
}

func TestParseDigestBoundedFailure(t *testing.T) {
	if _, err := ParseDigest(E6, `{`); err == nil {
		t.Fatal("ParseDigest(malformed): want error, got nil")
	} else {
		var pe *ParseError
		if !errors.As(err, &pe) {
			t.Errorf("ParseDigest err = %T, want *ParseError", err)
		}
	}
	got, err := ParseDigest(E6, `{"summary":"quiet night"}`)
	if err != nil {
		t.Fatalf("ParseDigest: %v", err)
	}
	if got.Summary != "quiet night" {
		t.Errorf("ParseDigest.Summary = %q, want %q", got.Summary, "quiet night")
	}
}

func TestSchemaFor(t *testing.T) {
	structured := map[Class]bool{E1: false, E2: true, E3: true, E4: false, E5: false, E6: true}
	for class, want := range structured {
		_, ok := SchemaFor(class)
		if ok != want {
			t.Errorf("SchemaFor(%s) ok = %v, want %v", class, ok, want)
		}
	}
}
