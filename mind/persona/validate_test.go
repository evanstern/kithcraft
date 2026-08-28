package persona

import (
	"errors"
	"testing"
)

func testPersona() Persona {
	return Persona{
		CastID:       "Aldric",
		Name:         "Aldric",
		Anchor:       "steady and exacting",
		DriftMarkers: []string{"reckless", "sloppy"},
		Profession:   "armorer",
		BiomeVariant: "plains",
	}
}

// TestValidate_AnchorEcho covers US3 AC #1 and #3: a candidate restating the
// anchor under whitespace/case normalization is accepted; anything else is
// rejected with a RejectionReason naming "anchor_echo".
func TestValidate_AnchorEcho(t *testing.T) {
	p := testPersona()
	tests := []struct {
		name      string
		candidate string
		wantErr   bool
	}{
		{"exact restate", "steady and exacting", false},
		{"case and whitespace normalized", "  Steady   AND Exacting.  ", false},
		{"restate embedded in a longer narrative", "Today Aldric was steady and exacting at the forge.", false},
		{"paraphrase fails", "Aldric is calm and precise", true},
		{"empty text fails", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(p, tt.candidate)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate(%q) error = %v, wantErr %v", tt.candidate, err, tt.wantErr)
			}
			var reason *RejectionReason
			if tt.wantErr && (!errors.As(err, &reason) || reason.Check != "anchor_echo") {
				t.Fatalf("Validate(%q) error = %v, want RejectionReason{Check: anchor_echo}", tt.candidate, err)
			}
		})
	}
}

// TestValidate_DriftMarker_WordBoundary is US3 AC #2 plus the word-boundary
// proof: a drift marker matches whole-word, case-insensitively, anywhere in
// the text, but must NOT match as a substring inside a larger unrelated
// word.
func TestValidate_DriftMarker_WordBoundary(t *testing.T) {
	p := testPersona()
	tests := []struct {
		name      string
		candidate string
		wantWord  string // non-empty means a drift_marker rejection is expected
	}{
		{"exact case match", "steady and exacting, but reckless today", "reckless"},
		{"case-insensitive match", "steady and exacting, but RECKLESS today", "reckless"},
		{"mixed-case match at boundary", "steady and exacting; SlOpPy work", "sloppy"},
		{"substring inside a larger word does not match", "steady and exacting, recklessness aside", ""},
		{"substring as a prefix of a larger word does not match", "steady and exacting, sloppyish work", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(p, tt.candidate)
			if tt.wantWord == "" {
				if err != nil {
					t.Fatalf("Validate(%q) = %v, want accept (no word-boundary match)", tt.candidate, err)
				}
				return
			}
			var reason *RejectionReason
			if !errors.As(err, &reason) || reason.Check != "drift_marker" || reason.Word != tt.wantWord {
				t.Fatalf("Validate(%q) error = %v, want RejectionReason{Check: drift_marker, Word: %q}", tt.candidate, err, tt.wantWord)
			}
		})
	}
}

// TestValidate_Moralizing_RejectsRegardlessOfOwnMarkers is US4 AC #2 (card
// AC #6 lexicon half): the moralizing lexicon is cast-wide and authored — it
// rejects a moralizing persona even when that word never appears among the
// persona's own generated DriftMarkers.
func TestValidate_Moralizing_RejectsRegardlessOfOwnMarkers(t *testing.T) {
	p := testPersona() // DriftMarkers: reckless, sloppy — no moralizing words
	tests := []string{
		"steady and exacting, but she lectures the player about manners",
		"steady and exacting, though he scolds anyone who leaves a mess",
		"steady and exacting; he takes offense easily",
	}
	for _, candidate := range tests {
		t.Run(candidate, func(t *testing.T) {
			err := Validate(p, candidate)
			var reason *RejectionReason
			if !errors.As(err, &reason) || reason.Check != "drift_marker" {
				t.Fatalf("Validate(%q) error = %v, want a drift_marker rejection from the cast-wide moralizing lexicon", candidate, err)
			}
		})
	}
}

// TestValidate_Accepts_ConformingText is US3 AC #3: a candidate that echoes
// the anchor and contains no drift or moralizing word is accepted.
func TestValidate_Accepts_ConformingText(t *testing.T) {
	p := testPersona()
	if err := Validate(p, "Aldric remained steady and exacting through the long shift at the forge."); err != nil {
		t.Fatalf("Validate() = %v, want accept", err)
	}
}
