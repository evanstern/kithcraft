package deliberate

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestManifestVerbs_ReadsDeclaredSet(t *testing.T) {
	caps := map[string]any{
		"verbs": []any{
			map[string]any{"verb": "go_to", "targets": []any{"place", "thing", "body"}},
			map[string]any{"verb": "carry", "targets": []any{"thing"}},
		},
	}
	got := ManifestVerbs(caps)
	if !got["go_to"] || !got["carry"] || len(got) != 2 {
		t.Fatalf("ManifestVerbs = %#v, want exactly {go_to, carry}", got)
	}
}

// TestManifestVerbs_DifferentManifestsYieldDifferentVocabularies is the
// behavioural half of card AC #2: the verb set a Loop will accept tracks
// the manifest it was built from, not a fixed vocabulary — two vendors
// declaring different verb sets produce two different accepted sets.
func TestManifestVerbs_DifferentManifestsYieldDifferentVocabularies(t *testing.T) {
	minimal := ManifestVerbs(map[string]any{"verbs": []any{
		map[string]any{"verb": "wait", "targets": []any{"none"}},
	}})
	extended := ManifestVerbs(map[string]any{"verbs": []any{
		map[string]any{"verb": "wait", "targets": []any{"none"}},
		map[string]any{"verb": "craft", "targets": []any{"thing"}},
	}})
	if minimal["craft"] {
		t.Fatal("a manifest that never declared craft must not yield it")
	}
	if !extended["craft"] {
		t.Fatal("a manifest that declared craft must yield it")
	}
}

func TestManifestVerbs_EmptyOrMalformedCapabilities(t *testing.T) {
	if got := ManifestVerbs(nil); len(got) != 0 {
		t.Fatalf("ManifestVerbs(nil) = %#v, want empty", got)
	}
	if got := ManifestVerbs(map[string]any{"verbs": "not a list"}); len(got) != 0 {
		t.Fatalf("ManifestVerbs with malformed verbs = %#v, want empty", got)
	}
}

// TestNoCompiledInVerbVocabulary is card AC #2's structural half: no
// non-test source file in this package spells a core verb as a string
// literal. This is a lightweight guard alongside the behavioural proof
// above (the real proof, since a grep can always be dodged by spelling a
// literal differently) — it exists to catch a future hardcoded fallback
// list being added by accident.
func TestNoCompiledInVerbVocabulary(t *testing.T) {
	coreVerb := regexp.MustCompile(`"(go_to|speak|attend|wait)"`)
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || filepath.Ext(name) != ".go" || filepath.Base(name) != name {
			continue
		}
		if len(name) >= len("_test.go") && name[len(name)-len("_test.go"):] == "_test.go" {
			continue // test fixtures are allowed to name verbs
		}
		b, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if m := coreVerb.FindString(string(b)); m != "" {
			t.Fatalf("%s hardcodes verb literal %s — verbs must come from ManifestVerbs only (FR-002, card AC #2)", name, m)
		}
	}
}
