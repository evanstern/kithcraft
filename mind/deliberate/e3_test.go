package deliberate

import (
	"strings"
	"testing"

	"kithcraft/mind/prompt"
)

func TestTriggerE3_TextOriginRead_Fires(t *testing.T) {
	percept := map[string]any{
		"percept_type": "text",
		"provenance":   map[string]any{"origin": "read"},
		"content":      map[string]any{"text": "Build a shelter by the north wall. — the player"},
	}
	if !TriggerE3(percept) {
		t.Fatal("a text percept with origin:read must trigger E3 (§4.7, FR-003)")
	}
}

func TestTriggerE3_OtherPerceptTypesAndOrigins_DoNotFire(t *testing.T) {
	cases := []map[string]any{
		{"percept_type": "speech", "provenance": map[string]any{"origin": "read"}},
		{"percept_type": "text", "provenance": map[string]any{"origin": "heard"}},
		{"percept_type": "text", "provenance": map[string]any{"origin": "told"}},
		{"percept_type": "text"}, // no provenance at all
		{},
	}
	for i, c := range cases {
		if TriggerE3(c) {
			t.Errorf("case %d: %#v must not trigger E3", i, c)
		}
	}
}

// TestE3Context_VariableContext_RendersAllFourSections proves the E3
// context shape (§2.3): board contents, other villagers' claims, standing
// relationship to the player, current commitments — all four present in
// the rendered variable suffix.
func TestE3Context_VariableContext_RendersAllFourSections(t *testing.T) {
	c := E3Context{
		Board:        "Build a shelter by the north wall. — the player",
		OtherClaims:  "Mira claimed the well repair this morning.",
		Relationship: "trusts the player, mildly wary of favoritism",
		Commitments:  "promised to mind the stall until dusk",
	}
	rendered := c.VariableContext().Render()
	for _, want := range []string{c.Board, c.OtherClaims, c.Relationship, c.Commitments} {
		if !strings.Contains(rendered, want) {
			t.Errorf("rendered variable context missing %q:\n%s", want, rendered)
		}
	}
}

// TestE3Context_StablePrefixByteUntouched proves T005's stable-prefix
// invariant (plan.md decision 4, mirroring prompt.TestAssemble_
// StablePrefixByteIdentical): two E3 assemblies sharing the same
// DeliberationStablePrefix but carrying entirely different board/claims/
// relationship/commitments content produce a byte-identical stable half —
// only the variable suffix differs (§4.3's caching lever intact).
func TestE3Context_StablePrefixByteUntouched(t *testing.T) {
	stable := prompt.DeliberationStablePrefix{
		Persona:      "a villager persona",
		Values:       "standing desires and values",
		Manifest:     "claim/decline manifest, role form",
		Instructions: "E3 job-board deliberation instructions",
	}

	morning := E3Context{
		Board:        "Build a shelter by the north wall. — the player",
		OtherClaims:  "no other claims yet",
		Relationship: "trusts the player",
		Commitments:  "none pending",
	}
	evening := E3Context{
		Board:        "Repair the well before the frost. — the player",
		OtherClaims:  "Mira claimed the shelter this morning",
		Relationship: "growing wary of favoritism",
		Commitments:  "promised to mind the stall until dusk",
	}

	a1 := prompt.Assemble(stable, morning.VariableContext())
	a2 := prompt.Assemble(stable, evening.VariableContext())

	if a1.Stable != a2.Stable {
		t.Fatalf("stable prefix differs across E3 contexts:\n%q\nvs\n%q", a1.Stable, a2.Stable)
	}
	if a1.Variable == a2.Variable {
		t.Fatal("variable context is identical across different board postings — this test would prove nothing about the split")
	}
}
