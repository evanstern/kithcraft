package consolidate

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anthropics/anthropic-sdk-go/option"

	"kithcraft/mind/llm"
	"kithcraft/mind/prompt"
)

// respond writes a minimal Messages API response with the given stop
// reason and text — the scripted server this test uses in place of any
// live call (spec.md FR-007), matching mind/llm/client_test.go's own
// httptest pattern.
func respond(w http.ResponseWriter, stopReason, text string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"id": "msg_test", "type": "message", "role": "assistant", "model": "claude-test",
		"content":     []map[string]any{{"type": "text", "text": text}},
		"stop_reason": stopReason,
		"usage":       map[string]any{"input_tokens": 10, "output_tokens": 5},
	})
}

// TestClientDigester_TextExtraction proves ClientDigester wires *llm.Client
// into Digester correctly for a normal (non-truncated) call.
func TestClientDigester_TextExtraction(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respond(w, "end_turn", `{"summary":"a quiet night"}`)
	}))
	t.Cleanup(srv.Close)

	d := ClientDigester{Client: llm.New(option.WithBaseURL(srv.URL), option.WithAPIKey("test-key"))}
	raw, truncated, err := d.Digest(context.Background(), prompt.Assembled{Variable: "hi"})
	if err != nil {
		t.Fatalf("Digest: %v", err)
	}
	if truncated {
		t.Error("truncated = true, want false")
	}
	if raw != `{"summary":"a quiet night"}` {
		t.Errorf("raw = %q, want the response text verbatim", raw)
	}
}

// TestClientDigester_TruncationDetected proves card AC #3 at the SDK
// boundary: stop_reason "max_tokens" is reported as truncated, the exact
// signal I's day-20 bug shipped without checking.
func TestClientDigester_TruncationDetected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respond(w, "max_tokens", `{"summary":"cut off mid-sen`)
	}))
	t.Cleanup(srv.Close)

	d := ClientDigester{Client: llm.New(option.WithBaseURL(srv.URL), option.WithAPIKey("test-key"))}
	_, truncated, err := d.Digest(context.Background(), prompt.Assembled{Variable: "hi"})
	if err != nil {
		t.Fatalf("Digest: %v", err)
	}
	if !truncated {
		t.Error("truncated = false, want true on stop_reason=max_tokens")
	}
}
