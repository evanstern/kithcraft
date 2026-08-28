package persona

import (
	"fmt"
	"regexp"
	"strings"
)

// RejectionReason is validate.go's reject-with-reason (FR-005): which
// firewall check failed, and — for a drift-marker rejection — which word
// matched. It satisfies error so a caller that only checks err != nil still
// works; errors.As recovers the structured reason.
type RejectionReason struct {
	Check string // "anchor_echo" or "drift_marker"
	Word  string // the matched drift/moralizing word; empty for anchor_echo
}

func (r *RejectionReason) Error() string {
	if r.Word != "" {
		return fmt.Sprintf("persona: validate: %s rejected: matched %q", r.Check, r.Word)
	}
	return fmt.Sprintf("persona: validate: %s rejected", r.Check)
}

var spaceRun = regexp.MustCompile(`\s+`)

// normalize collapses whitespace runs and lowercases: the anchor-echo check
// cares about content, not typography — a model that adds a trailing period
// or capitalizes while restating faithfully still passes; a paraphrase still
// fails (promptworld I's validateConsolidation carried the same rule).
func normalize(s string) string {
	return spaceRun.ReplaceAllString(strings.ToLower(strings.TrimSpace(s)), " ")
}

// Validate is the model-free half of the persona firewall (FR-005, US3).
// candidate is villager-generated text (e.g. a self-narrative) that must (a)
// restate p's Anchor verbatim, under whitespace/case normalization, and (b)
// contain none of p's own DriftMarkers or the cast-wide Moralizing lexicon
// (moralizing.go), matched word-boundary and case-insensitive. A nil error
// means candidate may land.
//
// Honest limit (FR-006): this catches STATED drift only — a persona that
// literally says "I am reckless now" or "he lectures the player". A persona
// that drifts by inference, tone, or implication without saying a marker
// word passes this layer. Catching that needs a model judging the text
// against the persona, which reintroduces exactly the model call this layer
// exists to avoid — parked for a model-judged validator, post-demo.
func Validate(p Persona, candidate string) error {
	if p.Anchor == "" || !strings.Contains(normalize(candidate), normalize(p.Anchor)) {
		return &RejectionReason{Check: "anchor_echo"}
	}

	markers := make([]string, 0, len(p.DriftMarkers)+len(Moralizing))
	markers = append(markers, p.DriftMarkers...)
	markers = append(markers, Moralizing...)
	for _, marker := range markers {
		re := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(marker) + `\b`)
		if re.MatchString(candidate) {
			return &RejectionReason{Check: "drift_marker", Word: marker}
		}
	}
	return nil
}
