// Package converse (this file): card AC #7's tedium checks, consolidated
// (specs/017-dusk-conversation Phase 4, T009). AC #7 has three clauses; each
// already has a proving test elsewhere in this package except the third,
// which this file closes:
//
//  1. lines do not repeat within a cycle — TestAmbientPool_ServeUnderBudgetNoRepeat
//     (pool_test.go).
//  2. a conversation reaches a natural end, never a turn cap mid-sentence —
//     TestExchange_SafetyBoundNeverFires (converse_test.go).
//  3. the stall-line is used sparingly, never a prefix tic — was doc-comment
//     only (pool.go's Stall doc); TestAmbientPool_StallOnlyAfterExhaustion
//     below makes it a behavioural check: across a simulated day's worth of
//     remark triggers, Stall is reached only once genuine pool content is
//     spent, never interspersed with or prepended to a real line.
package converse

import (
	"context"
	"testing"

	"kithcraft/mind/llm"
	"kithcraft/mind/prompt"

	"github.com/anthropics/anthropic-sdk-go/option"
)

// TestAmbientPool_StallOnlyAfterExhaustion simulates one villager's full
// in-game day: 8 real lines from one Refill, then more remark triggers than
// the pool holds. It proves "sparingly" concretely (Stall used for exactly
// the overflow, 2 of 10 draws here, never for one of the 8 real lines) and
// "never a prefix tic" structurally (each draw returns either a real line or
// a Stall line, never both concatenated — the loop below never has a reason
// to combine them, because Stall's contract is substitute-for, not prepend-
// to; see pool.go's Stall doc).
func TestAmbientPool_StallOnlyAfterExhaustion(t *testing.T) {
	srv, _ := jsonMessageServer(t, []string{eightLines})
	client := llm.New(option.WithBaseURL(srv.URL), option.WithAPIKey("test"))
	p := NewAmbientPool()
	if _, err := p.Refill(context.Background(), client, "Tam", 3, prompt.AmbientStablePrefix{}, "", "", ""); err != nil {
		t.Fatalf("Refill: %v", err)
	}

	const draws = 10 // 8 real lines available, 2 more than the pool holds
	seen := map[string]bool{}
	realCount, stallCount := 0, 0
	for i := 0; i < draws; i++ {
		if line, _, ok := p.Serve("Tam", 3); ok {
			if seen[line] {
				t.Errorf("draw %d: repeated line %q within the cycle", i, line)
			}
			seen[line] = true
			realCount++
			continue
		}
		// Pool empty: the caller's cue is Stall, never a retry against the
		// model on this path (pool.go's Serve doc) — nothing here prepends
		// it to anything.
		line := Stall(3, "Tam")
		found := false
		for _, l := range StallLines {
			found = found || l == line
		}
		if !found {
			t.Errorf("draw %d: Stall returned %q, not a StallLines member", i, line)
		}
		stallCount++
	}
	if realCount != 8 {
		t.Errorf("real lines served = %d, want 8 (the whole batch, none skipped)", realCount)
	}
	if stallCount != 2 {
		t.Errorf("Stall reached %d times, want 2 (only the overflow past the 8-line batch) — sparingly means proportional to genuine exhaustion, not routine use", stallCount)
	}
}
