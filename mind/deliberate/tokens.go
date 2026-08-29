// Package deliberate (this file): the token-only-targets rule (§5.2, card
// AC #7) — a set of place/thing_id/body identity tokens (§2.3) this mind
// has actually been shown, and the check that a composed intent's target
// names one of them rather than a description to search for.
package deliberate

import (
	"fmt"
	"sync"
)

// Tokens is every place/thing_id/body token seen so far, kind-scoped
// ("place:pl-1" != "thing_id:pl-1" even if the strings collide). It is
// fed by Observe, called once per inbound percept payload; a target is
// legitimate only if it names something already in this set (card AC #7:
// "every composed target traceable to a token the mind was given").
type Tokens struct {
	mu   sync.Mutex
	seen map[string]bool
}

// NewTokens returns an empty set — nothing is known yet.
func NewTokens() *Tokens { return &Tokens{seen: map[string]bool{}} }

// Observe walks a decoded percept payload (or any nested JSON-shaped
// value — map[string]any / []any, mind/wire's decode shape) and records
// every place/thing_id/body token found, keyed on field name rather than
// percept_type: every percept shape that carries an identity token uses
// one of these three field names somewhere in its tree
// (body-protocol-v0.md §2.4/§2.5/§4.5's source). This is the generic form
// of mind/memory/admission.go's subjectsOf, widened to all three token
// kinds because a target (§5.2) may name a thing or a body, not only a
// place.
func (t *Tokens) Observe(v any) {
	switch val := v.(type) {
	case map[string]any:
		for k, sub := range val {
			if k == "place" || k == "thing_id" || k == "body" {
				if s, ok := sub.(string); ok && s != "" {
					t.mark(k, s)
				}
			}
			t.Observe(sub)
		}
	case []any:
		for _, item := range val {
			t.Observe(item)
		}
	}
}

func (t *Tokens) mark(kind, token string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.seen[kind+":"+token] = true
}

// Known reports whether token has been Observed under kind ("place",
// "thing_id", or "body").
func (t *Tokens) Known(kind, token string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.seen[kind+":"+token]
}

// ValidateTarget enforces §5.2: nil (a verb needing no target at all)
// always passes; anything else must be a structured
// {"type": "place"|"thing"|"body", ...} reference naming a token this set
// has actually seen. A bare string or any other shape is a description in
// a target's clothing ("the nearest bed") and is rejected here, before
// compose ever sees it (card AC #7; body-protocol-v0.md §5.2, §12 finding
// L-5).
func (t *Tokens) ValidateTarget(target any) error {
	if target == nil {
		return nil
	}
	m, ok := target.(map[string]any)
	if !ok {
		return fmt.Errorf("deliberate: target %#v is not a structured token reference (§5.2 forbids a descriptive target)", target)
	}
	typ, _ := m["type"].(string)
	var field, kind string
	switch typ {
	case "", "none":
		return nil
	case "place":
		field, kind = "place", "place"
	case "thing":
		field, kind = "thing_id", "thing_id"
	case "body":
		field, kind = "body", "body"
	default:
		return fmt.Errorf("deliberate: target type %q is not one of place|thing|body|none (§5.2)", typ)
	}
	token, _ := m[field].(string)
	if token == "" {
		return fmt.Errorf("deliberate: target type %q is missing its %q token", typ, field)
	}
	if !t.Known(kind, token) {
		return fmt.Errorf("deliberate: target %s %q is not a token this mind was given (§5.2 — a target must already be known, never a description to search for)", kind, token)
	}
	return nil
}
