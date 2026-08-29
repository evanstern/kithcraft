package deliberate

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// TestNoPolitenessPolicingInDeliberationPath is card AC #9's structural
// half, in TestNoCompiledInVerbVocabulary's style (manifest_test.go): no
// non-test source file in this package spells a compliance gate, a
// cooldown, or a player-conduct-keyed refusal mechanism. Refusal grounds
// in this package are only ever a villager's own wants/commitments/
// relationships (board_test.go's decline case), never the player's
// conduct or a rate limit on how often they may ask — kithcraft-brief.md
// #6 names both as spell-breakers. A behavioural companion is impossible
// here (there is nothing to exercise — the absence itself is the proof),
// so this grep is the whole check, same posture as T002's.
func TestNoPolitenessPolicingInDeliberationPath(t *testing.T) {
	forbidden := regexp.MustCompile(`(?i)(compliance|cooldown|cool_down|conduct_score|player_conduct|politeness|rudeness|obedien|consecutive_declin)`)
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || filepath.Ext(name) != ".go" {
			continue
		}
		if len(name) >= len("_test.go") && name[len(name)-len("_test.go"):] == "_test.go" {
			continue // this file itself, and any other test, may name what must not exist
		}
		b, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if m := forbidden.FindString(string(b)); m != "" {
			t.Fatalf("%s names %q — no compliance gate, cooldown, or player-conduct-keyed refusal mechanism may exist in the deliberation path (card AC #9)", name, m)
		}
	}
}
