package converse

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/anthropics/anthropic-sdk-go/option"

	"kithcraft/mind/llm"
	"kithcraft/mind/prompt"
)

// jsonMessageServer serves one non-streaming Messages API reply per call
// (llm/client_test.go's minimalMessage, parameterized on text and counted
// — E5 is non-streaming, per the registry, so this package's SSE helpers
// don't apply to it).
func jsonMessageServer(t *testing.T, replies []string) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := calls.Add(1) - 1
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": "msg_test", "type": "message", "role": "assistant",
			"model":       "claude-test",
			"content":     []map[string]any{{"type": "text", "text": replies[i]}},
			"stop_reason": "end_turn",
			"usage":       map[string]any{"input_tokens": 10, "output_tokens": 5},
		})
	}))
	t.Cleanup(srv.Close)
	return srv, &calls
}

const eightLines = "Evening.\nRough day at the mill.\nMm.\nGood to see you.\nCold tonight.\nAlmost done here.\nLong one, this.\nSee you at the fire."

// TestE5UsesHaiku45 is card AC #5's registry assertion for this package's
// call site: E5 is Haiku 4.5, batched (not streaming) — Refill and
// Escalate both call llm.E5, so a future edit pointing either at the wrong
// class fails here even if classes_test.go's registry-level pin stays
// correct (converse_test.go's TestExchangeUsesE4Config does the same for
// E4).
func TestE5UsesHaiku45(t *testing.T) {
	cfg := llm.Registry[llm.E5]
	if cfg.Model != llm.ModelHaiku45 {
		t.Errorf("E5 Model = %q, want Haiku 4.5", cfg.Model)
	}
	if cfg.Streaming {
		t.Error("E5 must be batched (non-streaming), not a live stream")
	}
}

// TestAmbientPool_RefillProducesLines proves Refill makes one batched call
// and stores ~8 persona-flavoured lines (card AC #5).
func TestAmbientPool_RefillProducesLines(t *testing.T) {
	srv, calls := jsonMessageServer(t, []string{eightLines})
	client := llm.New(option.WithBaseURL(srv.URL), option.WithAPIKey("test"))
	stable := prompt.AmbientStablePrefix{PersonaThumbnail: "Tam, the miller"}

	p := NewAmbientPool()
	lines, err := p.Refill(context.Background(), client, "Tam", 3, stable, "tired", "Eda nearby", "roof fixed")
	if err != nil {
		t.Fatalf("Refill: %v", err)
	}
	if len(lines) != 8 {
		t.Errorf("got %d lines, want 8", len(lines))
	}
	if calls.Load() != 1 {
		t.Errorf("server saw %d calls, want 1 (one batched call)", calls.Load())
	}
}

// TestAmbientPool_ServeUnderBudgetNoRepeat proves card AC #5's serve
// clause: measured under 200 ms, and a line served this cycle is never
// served again — Serve is exhausted after exactly len(lines) draws.
func TestAmbientPool_ServeUnderBudgetNoRepeat(t *testing.T) {
	srv, _ := jsonMessageServer(t, []string{eightLines})
	client := llm.New(option.WithBaseURL(srv.URL), option.WithAPIKey("test"))
	p := NewAmbientPool()
	if _, err := p.Refill(context.Background(), client, "Tam", 3, prompt.AmbientStablePrefix{}, "", "", ""); err != nil {
		t.Fatalf("Refill: %v", err)
	}

	seen := map[string]bool{}
	for i := 0; i < 8; i++ {
		line, latency, ok := p.Serve("Tam", 3)
		if !ok {
			t.Fatalf("Serve #%d: ok = false, want true", i)
		}
		if latency >= 200*time.Millisecond {
			t.Errorf("Serve #%d latency = %s, want < 200ms", i, latency)
		}
		if seen[line] {
			t.Errorf("Serve #%d repeated line %q within the same cycle", i, line)
		}
		seen[line] = true
	}
	if _, _, ok := p.Serve("Tam", 3); ok {
		t.Error("Serve after exhausting the pool: ok = true, want false")
	}
}

// TestAmbientPool_DayRolloverClearsYesterday proves the daily refresh:
// Serve on a new day, before that day's Refill has run, reports pool-empty
// rather than serving yesterday's still-unspent lines; after the new day's
// Refill, its lines serve.
func TestAmbientPool_DayRolloverClearsYesterday(t *testing.T) {
	srv, _ := jsonMessageServer(t, []string{eightLines, "New day line one.\nNew day line two."})
	client := llm.New(option.WithBaseURL(srv.URL), option.WithAPIKey("test"))
	p := NewAmbientPool()
	if _, err := p.Refill(context.Background(), client, "Tam", 3, prompt.AmbientStablePrefix{}, "", "", ""); err != nil {
		t.Fatalf("Refill day 3: %v", err)
	}
	// Serve only one of day 3's eight lines — most are left unspent.
	if _, _, ok := p.Serve("Tam", 3); !ok {
		t.Fatal("Serve on day 3 before rollover: ok = false, want true")
	}

	if _, _, ok := p.Serve("Tam", 4); ok {
		t.Error("Serve on day 4 before day 4's Refill: ok = true, want false — yesterday's unspent lines must not serve")
	}

	if _, err := p.Refill(context.Background(), client, "Tam", 4, prompt.AmbientStablePrefix{}, "", "", ""); err != nil {
		t.Fatalf("Refill day 4: %v", err)
	}
	line, _, ok := p.Serve("Tam", 4)
	if !ok || !strings.HasPrefix(line, "New day") {
		t.Errorf("Serve on day 4 after Refill = %q, %v, want a day-4 line, true", line, ok)
	}
	if _, _, ok := p.Serve("Tam", 3); ok {
		t.Error("Serve on day 3 after day 4's Refill: ok = true, want false — the old day is simply gone")
	}
}

// TestIsTargeted proves card AC #6's routing predicate.
func TestIsTargeted(t *testing.T) {
	if IsTargeted("") {
		t.Error("empty subject: IsTargeted = true, want false (a passing greeting)")
	}
	if !IsTargeted("the broken fence") {
		t.Error("non-empty subject: IsTargeted = false, want true")
	}
}

// TestEscalate_LiveCallBypassesPool proves card AC #6: a targeted remark
// makes a live E5 call and never touches the pool (no Refill/Serve
// involved at all).
func TestEscalate_LiveCallBypassesPool(t *testing.T) {
	srv, calls := jsonMessageServer(t, []string{"That fence again? Third time this week."})
	client := llm.New(option.WithBaseURL(srv.URL), option.WithAPIKey("test"))

	line, err := Escalate(context.Background(), client, prompt.AmbientStablePrefix{PersonaThumbnail: "Tam"}, "the broken fence", "player just walked past it a third time")
	if err != nil {
		t.Fatalf("Escalate: %v", err)
	}
	if line != "That fence again? Third time this week." {
		t.Errorf("Escalate line = %q, want the scripted reply", line)
	}
	if calls.Load() != 1 {
		t.Errorf("server saw %d calls, want 1 (one live call)", calls.Load())
	}
}

// TestStall_NeverCallsModel proves the pool-empty edge case's stall-line
// policy never touches the model on the < 200 ms path: Stall's signature
// takes no client, so this is a compile-time guarantee — this test proves
// the runtime side, that it returns quickly and only from StallLines.
func TestStall_NeverCallsModel(t *testing.T) {
	start := time.Now()
	line := Stall(3, "Tam")
	if elapsed := time.Since(start); elapsed >= 200*time.Millisecond {
		t.Errorf("Stall took %s, want well under the 200ms budget it covers", elapsed)
	}
	found := false
	for _, l := range StallLines {
		if l == line {
			found = true
		}
	}
	if !found {
		t.Errorf("Stall returned %q, not a member of StallLines", line)
	}
}

// TestStall_Deterministic proves Stall is repeatable (same day+villager,
// same line) rather than randomized — a property the daemon's tests and
// this test both rely on.
func TestStall_Deterministic(t *testing.T) {
	if a, b := Stall(5, "Eda"), Stall(5, "Eda"); a != b {
		t.Errorf("Stall(5, Eda) = %q then %q, want identical", a, b)
	}
}
