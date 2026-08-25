// Package memory (this file): the §2.7 direct-perception classifier and the
// RM-2/RM-3 citation-resolution gate. Both are pure, deterministic functions
// over already-logged Events — no model call, no prose inspection — because
// body-protocol-v0.md §2.7 requires the classifier to be defined once, in
// the protocol, so mind and vendor cannot disagree about what counts as
// first-hand. specs/010-event-sourced-memory Phase 2 (T004).
package memory

// directOrigins is body-protocol-v0.md §2.7's DIRECT_ORIGINS: the closed set
// of percept origins that count as this body's own direct perception. Kept
// unexported so nothing outside the package can mutate the set; the pure
// function below is the only way in.
var directOrigins = map[string]bool{"acted": true, "saw": true, "heard": true, "felt": true}

// DirectPerception is §2.7's classifier: direct_perception(origin) = origin
// ∈ DIRECT_ORIGINS. It is a pure function of origin and nothing else — no
// hops, no source, no percept text — so hops can never promote or demote a
// claim (§2.6), and an unrecognized or absent origin (the zero value "")
// classifies secondhand, never direct (EH-2b, V-6): a map miss returns
// false.
func DirectPerception(origin string) bool { return directOrigins[origin] }

// Provenance is a belief's claimed epistemic tier — distinct from a
// percept's wire-level origin (§2.7), this is what RM-2/RM-3's citation gate
// resolves a model-authored claim down to. It never appears as a "direct"
// boolean on any wire (§2.7); it lives only in the belief store.
type Provenance string

const (
	Witnessed Provenance = "witnessed" // backed by a cited percept whose origin is direct
	Told      Provenance = "told"      // backed by a cited percept whose origin is not direct
	Inferred  Provenance = "inferred"  // no cited percept resolved to anything in the log
)

// provenanceRank orders the three tiers so ResolveCitations can compute a
// minimum; unrecognized values rank with Inferred (the conservative floor).
func provenanceRank(p Provenance) int {
	switch p {
	case Witnessed:
		return 2
	case Told:
		return 1
	default:
		return 0
	}
}

// ResolveCitations is RM-2/RM-3's gate: claimed is the provenance tier a
// model-authored belief asserts for itself, citedPerceptIDs are the percept
// IDs it names as its evidence (RM-2), and events is the log to resolve
// those citations against. The gate never rejects (RM-3) — it always
// returns a usable Provenance — and it never elevates a claim past what the
// evidence supports: the returned tier is claimed, coerced down to the best
// evidence found among the citations that actually resolve. coerced reports
// whether the returned tier differs from claimed, which is what a caller
// accumulates into RM-3's coercion count.
func ResolveCitations(events []Event, claimed Provenance, citedPerceptIDs []string) (resolved Provenance, coerced bool) {
	evidence := Inferred // RM-2: nothing resolvable degrades to inferred
	for _, id := range citedPerceptIDs {
		if id == "" {
			continue
		}
		for _, ev := range events {
			if ev.PerceptID() != id {
				continue
			}
			level := Told // a resolvable citation is at least secondhand-with-a-source
			if DirectPerception(ev.Origin()) {
				level = Witnessed
			}
			if provenanceRank(level) > provenanceRank(evidence) {
				evidence = level
			}
		}
	}

	resolved = claimed
	if provenanceRank(evidence) < provenanceRank(claimed) {
		resolved = evidence
	}
	return resolved, resolved != claimed
}
