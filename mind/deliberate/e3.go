// Package deliberate (this file): E3's trigger and context shape (spec.md
// FR-003, plan.md design decision 4) — the job-board deliberation fires on
// a `text` percept with `origin: read` (body-protocol-v0.md §4.7; routing
// doc line 100), and its variable context is exactly §2.3's row: board
// contents, other villagers' claims, standing relationship to the player,
// current commitments. This file owns the trigger check and the struct
// that renders into mind/prompt's variable suffix — never the stable
// prefix (persona/values/manifest/instructions), which is the same
// DeliberationStablePrefix E2/E4 already share and stays untouched here
// (§4.3's caching lever).
package deliberate

import "kithcraft/mind/prompt"

// TriggerE3 reports whether percept is E3's trigger: a `text` percept
// whose provenance.origin is "read" (§4.7's board-reading shape; FR-003,
// card AC #3's Given clause). Any other percept_type or origin — including
// a `text` percept heard rather than read — does not fire E3.
func TriggerE3(percept map[string]any) bool {
	if pt, _ := percept["percept_type"].(string); pt != "text" {
		return false
	}
	prov, _ := percept["provenance"].(map[string]any)
	origin, _ := prov["origin"].(string)
	return origin == "read"
}

// E3Context is E3's variable context (§2.3's row, plan.md decision 4):
// the board text itself, other villagers' claims against it, this
// villager's standing relationship to the player, and current
// commitments. Nothing here is stable across calls — that's why it
// renders through mind/prompt's VariableContext rather than
// DeliberationStablePrefix.
type E3Context struct {
	Board        string // the read percept's content.text (§4.7)
	OtherClaims  string
	Relationship string
	Commitments  string
}

// VariableContext renders c into mind/prompt's labeled-section shape,
// ready for prompt.Assemble alongside a class's DeliberationStablePrefix.
func (c E3Context) VariableContext() prompt.VariableContext {
	return *prompt.NewVariableContext().
		Add("BOARD", c.Board).
		Add("OTHER_CLAIMS", c.OtherClaims).
		Add("RELATIONSHIP", c.Relationship).
		Add("COMMITMENTS", c.Commitments)
}
