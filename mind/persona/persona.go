// Package persona owns the villagers' generated natures and the structural
// half of the persona firewall (specs/013-persona-genesis, decision-0003 §8):
// this package writes a persona file exactly once, at mode 0444 (files.go's
// WriteOnce), and nothing anywhere else in this package — or anywhere else in
// the codebase — has a write path to an existing persona. Load is read-only:
// it binds a persona to the cast id it was generated for and never
// regenerates a missing or unknown one. Genesis (Phase 3, genesis.go) is the
// only producer of a Persona; validate.go (Phase 2) is the model-free half
// that rejects drift once a persona exists.
package persona

// Persona is one villager's generated nature: E1 (Opus 5, per llm.E1) emits
// Name, Values, EndogenousDesires, Anchor, and DriftMarkers at genesis
// (brief #5 — these are no longer authored constants, unlike promptworld I's
// persona.Texts/Anchors/DriftMarkers, which this package's design otherwise
// ports). Anchor and DriftMarkers are the validator's raw material: Anchor is
// the verbatim line a conforming text must restate (anchor echo), and
// DriftMarkers are this persona's own contradiction words, unioned with the
// authored cast-wide moralizing lexicon (moralizing.go) at validation time.
type Persona struct {
	// CastID binds this persona to a vendor-mod cast entry across restarts
	// (FR-004). The vendor mod's Cast.Member (mod/src/main/java/dev/
	// kithcraft/mod/cast/Cast.java) has no id distinct from its display
	// name — Name is the only identity a spawned Villager carries — so
	// CastID is that same string.
	CastID string `json:"cast_id"`
	Name   string `json:"name"`

	Values            []string `json:"values"`
	EndogenousDesires []string `json:"endogenous_desires"`

	Anchor       string   `json:"anchor"`
	DriftMarkers []string `json:"drift_markers"`

	// Profession and BiomeVariant are the vendor mod's villager profession
	// and villager-type (biome variant) ids — decision-0002's cast
	// distinctiveness pairing, carried through so the fiction and the body
	// agree (card AC #5).
	Profession   string `json:"profession"`
	BiomeVariant string `json:"biome_variant"`
}

// Verified demo-cast pairing (2026-08-28, against mod/src/main/java/dev/
// kithcraft/mod/cast/Cast.java's Cast.MEMBERS — the vendor mod's actual
// seeded cast, TASK-0014): three DISTINCT profession x biome-variant pairs,
// not the uniform "Plains; farmer/librarian/cleric" plan.md's "Key decisions
// already settled" section assumed before this check ran (deviation recorded
// in specs/013-persona-genesis/tasks.md, Phase 1 checkpoint).
//
//	CastID "Aldric" — profession "armorer",   biome variant "plains"
//	CastID "Petra"  — profession "farmer",    biome variant "desert"
//	CastID "Yenna"  — profession "fisherman", biome variant "taiga"

// DemoCast returns the three-member demo cast pairing above as Genesis
// input (TASK-0021 T001): the single source of truth for it, so the daemon
// and this package's own tests (genesis_test.go's demoCastEntries) name the
// same three entries instead of two copies drifting apart.
func DemoCast() []CastEntry {
	return []CastEntry{
		{CastID: "Aldric", Profession: "armorer", BiomeVariant: "plains"},
		{CastID: "Petra", Profession: "farmer", BiomeVariant: "desert"},
		{CastID: "Yenna", Profession: "fisherman", BiomeVariant: "taiga"},
	}
}
