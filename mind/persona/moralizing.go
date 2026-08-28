package persona

// Moralizing is the cast-wide moralizing lexicon (FR-005, decision-0003 §8):
// the politeness-policing spell-breaker the brief names explicitly —
// "villagers taking offense at a player being a jerk (this is not a
// politeness simulator)" (docs/design/kithcraft-brief.md, "Spell-breakers to
// design against"). It is doctrine, not character: authored once here, never
// emitted by E1, and unioned into every persona's own DriftMarkers at
// validation time (validate.go) regardless of what that persona's genesis
// produced. Deliberately small and blunt — this list vetoes a moralizer at
// birth, it is not a taxonomy of politeness.
var Moralizing = []string{
	"scolds",
	"lectures",
	"chides",
	"moralizes",
	"admonishes",
	"reprimands",
	"corrects manners",
	"polices manners",
	"takes offense",
}
