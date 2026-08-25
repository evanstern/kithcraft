// Package prompt (this file): the variable-context builder and composed
// assembly (§2.3) — the stable/variable split that is the caching design
// (§4.3), built as a first-class component rather than concatenation at a
// call site (spec.md FR-003).
package prompt

import "strings"

// Section is one labeled block of variable context — memory window,
// board text, transcript, current world time, whatever a class's §2.3 row
// lists. Order is caller-controlled and preserved: it is not sorted, since
// unlike the wire's canonical JSON (mind/wire), a prompt's section order is
// authored, not a wire invariant.
type Section struct {
	Label string
	Text  string
}

// VariableContext is a class's per-call, non-cacheable remainder — every
// §2.3 row's "Variable context" column, including anything time-shaped
// (world_time, "current situation", a transcript so far). Nothing about
// this type is stable across calls; that's the point of splitting it from
// StablePrefix.
type VariableContext struct {
	Sections []Section
}

// Add appends a labeled section and returns the receiver, so callers can
// chain: prompt.NewVariableContext().Add("MEMORY", mem).Add("SITUATION", sit).
func (v *VariableContext) Add(label, text string) *VariableContext {
	v.Sections = append(v.Sections, Section{Label: label, Text: text})
	return v
}

// NewVariableContext returns an empty VariableContext ready for Add calls.
func NewVariableContext() *VariableContext {
	return &VariableContext{}
}

func (v VariableContext) Render() string {
	var sb strings.Builder
	for i, s := range v.Sections {
		if i > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString(s.Label)
		sb.WriteString(":\n")
		sb.WriteString(s.Text)
	}
	return sb.String()
}

// Assembled is one composed call, stable prefix first, cache breakpoint
// after it (RT-3), variable context after that (§2.3's order for every
// class that has a stable prefix at all).
type Assembled struct {
	Stable   string // empty for a class with no StablePrefix (E1)
	Variable string
}

// Prompt concatenates the two halves into the text a call actually sends.
// The split is kept visible on Assembled (Stable/Variable separately)
// specifically so a cache-breakpoint placement (Phase 2's client) and a
// byte-identity test (assemble_test.go) can each address the stable half
// without re-deriving it from the concatenation.
func (a Assembled) Prompt() string {
	if a.Stable == "" {
		return a.Variable
	}
	return a.Stable + "\n\n" + a.Variable
}

// Assemble composes a class's stable prefix and variable context per
// §2.3. stable may be nil for a class with none (E1) — variable then
// carries the whole call.
func Assemble(stable StablePrefix, variable VariableContext) Assembled {
	a := Assembled{Variable: variable.Render()}
	if stable != nil {
		a.Stable = stable.Render()
	}
	return a
}
