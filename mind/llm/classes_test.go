package llm

import "testing"

// TestClassTierMapping proves card AC #1: each class carries its declared
// model (§1.3) — E1/E6 on Opus 5, E2/E3/E4 on Sonnet 5, E5 on Haiku 4.5.
func TestClassTierMapping(t *testing.T) {
	want := map[Class]string{
		E1: ModelOpus5,
		E2: ModelSonnet5,
		E3: ModelSonnet5,
		E4: ModelSonnet5,
		E5: ModelHaiku45,
		E6: ModelOpus5,
	}
	for class, model := range want {
		cfg, ok := Registry[class]
		if !ok {
			t.Fatalf("Registry has no entry for %s", class)
		}
		if cfg.Model != model {
			t.Errorf("Registry[%s].Model = %q, want %q", class, cfg.Model, model)
		}
	}
	if len(Registry) != len(want) {
		t.Errorf("Registry has %d classes, want %d (E1..E6)", len(Registry), len(want))
	}
}

// TestE4Posture proves the AC #1 scenario's config half: E4 carries
// streaming, low effort, thinking off, and a tight max_tokens (§5.2).
func TestE4Posture(t *testing.T) {
	cfg := Registry[E4]
	if !cfg.Streaming {
		t.Error("E4 must stream (§5.2)")
	}
	if cfg.ThinkingOn {
		t.Error("E4 must have thinking off (§5.2) — it is the one latency-critical class")
	}
	if cfg.Effort != EffortLow {
		t.Errorf("E4 effort = %q, want %q (§5.2)", cfg.Effort, EffortLow)
	}
	if cfg.MaxTokens > 300 {
		t.Errorf("E4 MaxTokens = %d, want <= 300 (§5.2: tight, ~300)", cfg.MaxTokens)
	}
}

// TestE6Ceiling proves card AC #6: E6's max_tokens ceiling is
// truncation-aware and never near its expected output (RT-7's inherited
// lesson, §2.3 — "budget 4,096 against an expected 1,200").
func TestE6Ceiling(t *testing.T) {
	cfg := Registry[E6]
	if cfg.MaxTokens != 4096 {
		t.Errorf("E6 MaxTokens = %d, want 4096 (§2.3)", cfg.MaxTokens)
	}
	if cfg.ExpectedOutputTokens != 1200 {
		t.Errorf("E6 ExpectedOutputTokens = %d, want 1200 (§2.3)", cfg.ExpectedOutputTokens)
	}
	// "never near" — require at least a 2x margin, well under the
	// doc's own ~3.4x, so a future edit that narrows the ceiling
	// materially trips this before it reaches I's silent-truncation bug.
	if cfg.MaxTokens < 2*cfg.ExpectedOutputTokens {
		t.Errorf("E6 MaxTokens (%d) is not a safe multiple of ExpectedOutputTokens (%d): RT-7 requires the ceiling stay clear of expected output", cfg.MaxTokens, cfg.ExpectedOutputTokens)
	}
}

// TestCachePolicy proves §4.3/FR-002: E2/E3/E4 are cached, E1/E5/E6 are
// not (E6 explicitly — caching its once-per-20-minutes prefix would be a
// cache *write* past TTL every call, a cost increase, not a saving).
func TestCachePolicy(t *testing.T) {
	cached := map[Class]bool{E1: false, E2: true, E3: true, E4: true, E5: false, E6: false}
	for class, want := range cached {
		if got := Registry[class].Cached; got != want {
			t.Errorf("Registry[%s].Cached = %v, want %v", class, got, want)
		}
	}
}

// TestStructuredOutputPolicy proves A-9: E2/E3/E6 parse a structured
// value; E1/E4/E5 do not.
func TestStructuredOutputPolicy(t *testing.T) {
	structured := map[Class]bool{E1: false, E2: true, E3: true, E4: false, E5: false, E6: true}
	for class, want := range structured {
		if got := Registry[class].StructuredOutput; got != want {
			t.Errorf("Registry[%s].StructuredOutput = %v, want %v", class, got, want)
		}
	}
}
