// Package memory (this file): the §6.3 deterministic episodic admission
// gate — the filter between the percept inbox and the log deciding which
// percepts are eligible to become memories at all. It is pure and
// stateful only in its own first-sighting bookkeeping; nothing here calls
// a model (routing §1.2). A dropped percept never reaches Log.Append —
// dropping it from the episodic buffer is not belief deletion and does
// not touch RM-7, which governs the log and the store, not this gate.
// specs/010-event-sourced-memory Phase 3 (T007).
package memory

import "sync"

// AdmitRule names which §6.3 rule admitted a percept, checked in the
// order routing §6.3 lists them. The zero value names no rule and is
// never returned alongside admit=true.
type AdmitRule string

const (
	RuleUrgency       AdmitRule = "urgency"             // urgency >= notable
	RuleOtherBody     AdmitRule = "other_body_or_player" // involves another body or the player
	RuleActedReason   AdmitRule = "authored_reason"      // act_result on an intent the mind authored a reason for
	RuleToldOrText    AdmitRule = "told_fact_or_text"
	RuleFirstSighting AdmitRule = "first_sighting" // first sighting of a kind or place
)

// Gate is the §6.3 admission gate. One per villager: "first sighting" is
// per-villager knowledge, so a Gate shared across villagers would leak
// one villager's sightings into another's admission decisions. Safe for
// concurrent Decide calls against the same villager's percept stream.
type Gate struct {
	mu   sync.Mutex
	seen map[string]bool // "kind:"+token or "place:"+token already sighted
}

// NewGate returns an empty gate: nothing is a known kind or place yet.
func NewGate() *Gate { return &Gate{seen: make(map[string]bool)} }

// Decide is §6.3's gate, applied to one already wire-decoded percept
// object (percept_type, urgency, provenance, place, content — the shape
// a percept envelope's payload takes; see body-protocol-v0.md §4.1). It
// reports whether to admit the percept and, when admitted, which rule
// admitted it. Every kind or place the percept names is recorded as
// seen regardless of which rule (if any) admitted it — otherwise a
// percept that qualified on some other rule but happened to be a first
// sighting too would never let a later, truly repeated sighting of the
// same thing get correctly dropped.
func (g *Gate) Decide(percept map[string]any) (bool, AdmitRule) {
	perceptType, _ := percept["percept_type"].(string)
	content, _ := percept["content"].(map[string]any)
	place, _ := percept["place"].(map[string]any)
	provenance, _ := percept["provenance"].(map[string]any)

	g.mu.Lock()
	defer g.mu.Unlock()

	firstSighting := false
	for _, subj := range subjectsOf(place, content) {
		if !g.seen[subj] {
			firstSighting = true
		}
		g.seen[subj] = true
	}

	// Type-specific unconditional rules (act_result-with-reason,
	// told_fact/text) are checked before the general "other body or
	// player" rule: a told_fact's teller is always another body, so
	// checking OtherBody first would report the wrong rule for every
	// told_fact — right admit decision, misleading reason.
	switch {
	case urgencyAtLeastNotable(percept["urgency"]):
		return true, RuleUrgency
	case perceptType == "act_result" && hasAuthoredReason(content):
		return true, RuleActedReason
	case perceptType == "told_fact" || perceptType == "text":
		return true, RuleToldOrText
	case involvesOtherBodyOrPlayer(provenance, content):
		return true, RuleOtherBody
	case firstSighting:
		return true, RuleFirstSighting
	default:
		return false, "" // the drop rule: a repeated background sighting of an already-known thing
	}
}

// subjectsOf derives §6.3's "kind or place" subject keys from a
// percept's place and content — the percept-type-specific parsing the
// rule needs, kept generic (field-name driven, not a percept_type
// switch) since every percept type that carries a place or a thing/kind
// uses the same field names (body-protocol-v0.md §4.2-§4.7).
func subjectsOf(place, content map[string]any) []string {
	var subs []string
	if pl, ok := place["place"].(string); ok && pl != "" {
		subs = append(subs, "place:"+pl)
	}
	if thing, ok := content["thing"].(map[string]any); ok {
		if kind, ok := thing["kind"].(string); ok && kind != "" {
			subs = append(subs, "kind:"+kind)
		}
	}
	if aboutPlace, ok := content["about_place"].(map[string]any); ok {
		if pl, ok := aboutPlace["place"].(string); ok && pl != "" {
			subs = append(subs, "place:"+pl)
		}
	}
	if present, ok := content["present"].([]any); ok {
		for _, item := range present {
			if m, ok := item.(map[string]any); ok {
				if kind, ok := m["kind"].(string); ok && kind != "" {
					subs = append(subs, "kind:"+kind)
				}
			}
		}
	}
	if sk, ok := content["sound_kind"].(string); ok && sk != "" {
		subs = append(subs, "kind:"+sk)
	}
	return subs
}

// urgencyAtLeastNotable is §2.8's urgency threshold; an absent or
// unrecognized value (a future minor version's addition) is conservative
// — it does not auto-admit, the same posture §2.7 takes for origin.
func urgencyAtLeastNotable(v any) bool {
	s, _ := v.(string)
	return s == "notable" || s == "urgent"
}

// involvesOtherBodyOrPlayer is "any percept involving another body or
// the player" (§6.3): true when the provenance's source is a body or a
// person (§2.6/§4.5 — the player crosses as a person source, never a
// body), when the content names another minded body (a sighted thing
// with a non-empty body token, §2.5), or when a speech percept names
// other bodies as addressees.
func involvesOtherBodyOrPlayer(provenance, content map[string]any) bool {
	if source, ok := provenance["source"].(map[string]any); ok {
		if k, _ := source["kind"].(string); k == "body" || k == "person" {
			return true
		}
	}
	if thing, ok := content["thing"].(map[string]any); ok {
		if b, ok := thing["body"].(string); ok && b != "" {
			return true
		}
	}
	if addressed, ok := content["addressed_to"].([]any); ok && len(addressed) > 0 {
		return true
	}
	return false
}

// hasAuthoredReason is the act_result half of §6.3: `reason` is echoed
// back verbatim from the intent that produced this result (§5.2, §5.4) —
// present and non-empty iff the mind authored one.
func hasAuthoredReason(content map[string]any) bool {
	r, _ := content["reason"].(string)
	return r != ""
}
