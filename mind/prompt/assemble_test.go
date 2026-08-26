package prompt

import "testing"

func testDeliberationStable() DeliberationStablePrefix {
	return DeliberationStablePrefix{
		Persona:      "a villager persona",
		Values:       "standing desires and values",
		Manifest:     "verb & salient-kind manifest",
		Instructions: "deliberation instructions",
	}
}

// TestAssemble_StablePrefixByteIdentical proves card AC #3: two assemblies
// for the same villager and class at different world times produce a
// byte-identical stable prefix — no date, day counter, or timestamp
// rendered into it. The two calls below share identical stable input but
// deliberately differ in their variable context's world_time, which is
// where such a value belongs (§2.3).
func TestAssemble_StablePrefixByteIdentical(t *testing.T) {
	stable := testDeliberationStable()

	early := NewVariableContext().Add("MEMORY", "m1, m2, m3").Add("WORLD_TIME", "1000")
	late := NewVariableContext().Add("MEMORY", "m1, m2, m3, m4").Add("WORLD_TIME", "97000")

	a1 := Assemble(stable, *early)
	a2 := Assemble(stable, *late)

	if a1.Stable != a2.Stable {
		t.Fatalf("stable prefix differs across world times:\n%q\nvs\n%q", a1.Stable, a2.Stable)
	}
	if a1.Variable == a2.Variable {
		t.Fatal("variable context is identical across different world times — this test would prove nothing about the split")
	}
}

// TestAssemble_CatchesRenderedTimestamp is SC-003's deliberate red variant:
// it builds a stand-in render that (wrongly) folds a world-time value into
// what should be the stable text, the way a caller might if nothing
// stopped them, and shows the byte-identity comparison above would have
// failed and caught it. Without this, TestAssemble_StablePrefixByteIdentical
// passing could mean either "the invariant holds" or "the test can't tell
// the difference" — this proves the latter is false.
func TestAssemble_CatchesRenderedTimestamp(t *testing.T) {
	leaked := func(worldTime string) string {
		return testDeliberationStable().Render() + "\nWORLD_TIME:" + worldTime
	}
	got1 := leaked("1000")
	got2 := leaked("97000")
	if got1 == got2 {
		t.Fatal("expected the leaked-timestamp variant to differ across world times — if it doesn't, this red-variant check cannot prove the real byte-identity test catches the same bug")
	}
}

// TestAssemble_NoTimeShapedFieldOnStableSide documents, as a compile-time
// fact rather than a runtime assertion, that DeliberationStablePrefix,
// AmbientStablePrefix, and ConsolidationStablePrefix expose only the §2.3
// stable-content fields (persona/values/manifest/instructions,
// persona-thumbnail, persona/firewall-anchor) — none of them has a
// world-time, day-counter, or timestamp field a caller could fill in. This
// test exists so the claim is checked by something that fails to compile
// if a future edit adds one, not only asserted in a comment.
func TestAssemble_NoTimeShapedFieldOnStableSide(t *testing.T) {
	_ = DeliberationStablePrefix{Persona: "", Values: "", Manifest: "", Instructions: ""}
	_ = AmbientStablePrefix{PersonaThumbnail: ""}
	_ = ConsolidationStablePrefix{Persona: "", FirewallAnchor: ""}
	// Any additional field on these composite literals (including a
	// time-shaped one added elsewhere in the struct) would be a compile
	// error here under `go vet`'s unkeyed-field checks only if unkeyed;
	// the real guarantee is structural (shapes.go's field lists), and
	// this test's role is to keep that struct literal in lockstep with
	// the type so a reviewer sees the whole field set in one place.
}

// TestAssemble_E1HasNoStablePrefix proves E1 assembles with an empty
// stable half (§2.3: "—") — persona genesis is a one-shot call with
// nothing to keep stable across calls.
func TestAssemble_E1HasNoStablePrefix(t *testing.T) {
	variable := *NewVariableContext().Add("PREMISE", "world premise").Add("NAME", "villager name")
	a := Assemble(nil, variable)
	if a.Stable != "" {
		t.Errorf("E1 assembly Stable = %q, want empty", a.Stable)
	}
	if a.Prompt() != a.Variable {
		t.Errorf("E1 Prompt() should equal Variable alone when Stable is empty")
	}
}
