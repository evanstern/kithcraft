// Package prompt (this file): typed stable-prefix inputs per class
// (docs/design/llm-routing-and-budget.md §2.3). This package imports
// nothing from mind/llm (plan.md, "Structure Decision") — it produces the
// values the client sends, testable without any SDK or model-tier type.
//
// It owns assembly *structure* only, not prompt content (spec.md FR-007):
// M3 owns persona genesis text, M5/M6/M7 own their class prompts. The
// fields below are the labeled slots §2.3's table names for each class's
// stable prefix — callers fill them with real text once that content
// exists.
//
// The byte-identity invariant (card AC #3) is enforced by the type surface
// first: none of these structs has a field a caller could put a world time,
// day counter, or timestamp into. A stable prefix is stable because there
// is nothing volatile it could hold, not because a caller remembered not to
// add one.
package prompt

// StablePrefix renders a class's cacheable prefix to bytes. Every
// StablePrefix implementation renders the same fields in the same order on
// every call — no map iteration, no field ever sourced from a clock — so
// identical field values produce byte-identical output (assemble.go's
// Assemble relies on this).
type StablePrefix interface {
	Render() string
}

// DeliberationStablePrefix is E2/E3/E4's shared shape: §2.3 gives all
// three the identical composition — persona (600) + standing desires/
// values (300) + verb & salient-kind manifest in role form (400) +
// instructions (700) — so one type serves all three rather than three
// structurally-identical ones.
type DeliberationStablePrefix struct {
	Persona      string // who this villager is
	Values       string // standing desires/values
	Manifest     string // verb & salient-kind manifest, in role form
	Instructions string // per-class instructions
}

func (s DeliberationStablePrefix) Render() string {
	return "PERSONA:\n" + s.Persona +
		"\n\nVALUES:\n" + s.Values +
		"\n\nMANIFEST:\n" + s.Manifest +
		"\n\nINSTRUCTIONS:\n" + s.Instructions
}

// AmbientStablePrefix is E5's shape: a persona thumbnail only (§2.3: 250
// tokens) — the ambient pool call needs far less standing context than a
// deliberation.
type AmbientStablePrefix struct {
	PersonaThumbnail string
}

func (s AmbientStablePrefix) Render() string {
	return "PERSONA_THUMBNAIL:\n" + s.PersonaThumbnail
}

// ConsolidationStablePrefix is E6's shape: persona + the persona firewall
// anchor (§2.3: 800 tokens) — E6 needs the firewall anchor present because
// its output becomes belief/narrative text that must not drift the
// villager (§1.2's persona firewall). Deliberately not cached (§4.3); that
// policy lives in llm.Registry, not here — this package doesn't know what
// caching is.
type ConsolidationStablePrefix struct {
	Persona        string
	FirewallAnchor string
}

func (s ConsolidationStablePrefix) Render() string {
	return "PERSONA:\n" + s.Persona + "\n\nFIREWALL_ANCHOR:\n" + s.FirewallAnchor
}

// E1 (persona genesis) has no stable prefix at all (§2.3: "—") — it is a
// one-shot call, ever, per villager, so there is nothing to keep stable
// across calls and no cache to break by construction.
