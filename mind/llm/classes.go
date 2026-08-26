// Package llm (this file): the six-class registry — E1..E6 bound to their
// ratified model, latency posture, thinking/effort config, max_tokens, cache
// policy, and structured-output flag (docs/design/llm-routing-and-budget.md
// §1.3, §2.3, §4.3, §5). This is data, not behavior: a tier change is a
// one-line diff to Registry, matching the plan's stated goal
// (specs/011-model-client-routing/plan.md, "Structure Decision"). The SDK
// wrapper that actually sends these calls arrives in Phase 2 (T004+); this
// file has no dependency on it or on anything outside the stdlib.
package llm

// Class is one of the six villager cognition events that reach a model
// (§1.3). The zero value is not a valid class — always use the E1..E6
// constants.
type Class string

const (
	E1 Class = "E1" // Persona genesis
	E2 Class = "E2" // Routine deliberation
	E3 Class = "E3" // Job-board deliberation
	E4 Class = "E4" // Conversation turn
	E5 Class = "E5" // Ambient line pool
	E6 Class = "E6" // Nightly consolidation
)

// Model IDs, §3's pricing table. Kept as named constants rather than
// string literals in Registry so a tier reassignment reads as "E4:
// ModelSonnet5" rather than a bare string easy to typo.
const (
	ModelOpus5   = "claude-opus-5"
	ModelSonnet5 = "claude-sonnet-5"
	ModelHaiku45 = "claude-haiku-4-5"
)

// Effort levels §5 names explicitly. "" means the class's posture gives no
// effort config (offline classes, or classes with thinking off) — left
// unset rather than invented, since inventing a level nothing in the
// routing doc states would be a claim this registry cannot back.
const (
	EffortLow    = "low"
	EffortMedium = "medium"
)

// Config is one class's full call posture: everything §1.3/§5's tables
// specify plus the cache and structured-output policy of §4.3/A-9.
type Config struct {
	Model string

	// Latency posture (§5). Streaming and ThinkingOn/Effort are the flags
	// §5's per-class rows and §5.1-§5.4's prose actually name; nothing
	// here is a measured latency (that's M6's job, per the plan's
	// "Constraints" note) — it is configuration correctness.
	Streaming  bool
	ThinkingOn bool
	Effort     string // EffortLow, EffortMedium, or "" (see above)

	// MaxTokens is this class's request ceiling; ExpectedOutputTokens is
	// §2.3's output estimate. RT-7's inherited lesson (promptworld I's
	// silent truncation bug, §2.3) is: never set MaxTokens near
	// ExpectedOutputTokens. Only E4 and E6 have config the doc states
	// explicitly (E4's tight ~300 is the latency budget itself; E6's
	// 4,096-against-1,200 is the truncation lesson by name). The other
	// four get a uniform, generous multiple of their §2.3 expected
	// output instead of a per-class invented number: 4096 for the two
	// Opus-tier, offline-latency classes (E1, E6 — E6's is the one the
	// doc gives directly, E1 shares it since both are unbounded-latency
	// Opus calls), 1024 for the three Sonnet/Haiku classes that aren't
	// E4's deliberately tight exception.
	MaxTokens            int
	ExpectedOutputTokens int

	Cached           bool // §4.3: stable prefix gets an explicit cache breakpoint
	StructuredOutput bool // A-9: E2/E3/E6 parse a value, not prose
}

// Registry binds every class to its Config. §1.3 is the tier table this
// mirrors; §5 is the latency-posture table; §4.3 is the cache policy;
// A-9 is the structured-output list.
var Registry = map[Class]Config{
	E1: {
		Model:                ModelOpus5, // §1.3: rarest, most consequential call
		Streaming:            false,
		ThinkingOn:           false, // §5: "offline, pre-session, unbounded" — no effort/thinking config stated
		Effort:               "",
		MaxTokens:            4096,  // shares E6's Opus-tier, offline ceiling; see Config doc
		ExpectedOutputTokens: 1500,  // §2.3
		Cached:               false, // one-shot call, ever — no reuse to cache (§4.3's policy list)
		StructuredOutput:     false,
	},
	E2: {
		Model:                ModelSonnet5,
		Streaming:            false,
		ThinkingOn:           true,
		Effort:               EffortMedium, // §5.1: "Adaptive thinking on, effort medium."
		MaxTokens:            1024,
		ExpectedOutputTokens: 200,  // §2.3
		Cached:               true, // §4.3: the volume class, 42% of the bill
		StructuredOutput:     true, // A-9
	},
	E3: {
		Model:     ModelSonnet5,
		Streaming: false,
		// §5's table gives E3 no separate thinking/effort line; it shares
		// E2's mulling-tolerant posture and identical stable-prefix shape
		// (§2.3: "same 2,000"), so this mirrors E2's explicit config
		// rather than inventing a distinct one for a class the doc
		// treats as structurally the same.
		ThinkingOn:           true,
		Effort:               EffortMedium,
		MaxTokens:            1024,
		ExpectedOutputTokens: 250, // §2.3
		Cached:               true,
		StructuredOutput:     true, // A-9
	},
	E4: {
		Model: ModelSonnet5,
		// §5.2, verbatim config: "streaming on, effort: low, thinking
		// off, cached prefix, max_tokens tight (~300)." The one class
		// where the ceiling is deliberately tight rather than generous —
		// tight max_tokens is part of the latency budget, not a
		// truncation risk, because E4's expected output (150) already
		// sits close to it by design.
		Streaming:            true,
		ThinkingOn:           false,
		Effort:               EffortLow,
		MaxTokens:            300,
		ExpectedOutputTokens: 150, // §2.3
		Cached:               true,
		StructuredOutput:     false, // A-9 lists E2/E3/E6, not E4 — E4 stays prose for streaming
	},
	E5: {
		Model:                ModelHaiku45,
		Streaming:            false, // batched, once/villager/day — not on any critical path (§5.3)
		ThinkingOn:           false,
		Effort:               "",
		MaxTokens:            1024,
		ExpectedOutputTokens: 400,   // §2.3, ~8 lines
		Cached:               false, // batched once/day — no reuse to cache
		StructuredOutput:     false,
	},
	E6: {
		Model:      ModelOpus5,
		Streaming:  false,
		ThinkingOn: true, // §5.4: "Opus 5 with adaptive thinking runs comfortably inside" the sleep window
		Effort:     "",   // §5.4 doesn't name a level
		// §2.3, by name: "Budget 4,096 against an expected 1,200" — the
		// truncation lesson (RT-7) applied directly, not derived.
		MaxTokens:            4096,
		ExpectedOutputTokens: 1200,
		Cached:               false, // §4.3: past the cache TTL every call — caching E6 would be a 25% cost *increase*
		StructuredOutput:     true,  // A-9
	},
}
