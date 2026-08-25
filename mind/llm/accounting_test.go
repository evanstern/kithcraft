package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"kithcraft/mind/prompt"
)

// call is one scripted step of TestSessionAccounting's session: which class
// makes the call, what usage the mock server reports for it, and whether
// it's the streaming call the test cancels mid-flight.
type call struct {
	class                                 Class
	input, output, cacheRead, cacheCreate int64
	stream                                bool
}

// TestSessionAccounting proves card AC #7 (FR-005): a scripted mocked
// session touching all six classes — E2 called twice, to prove
// per-class totals accumulate rather than overwrite — reports correct
// per-class call counts and token totals at session end, including a
// cancelled streaming call's partial usage (spec.md's Edge Cases; T008).
func TestSessionAccounting(t *testing.T) {
	script := []call{
		{class: E1, input: 500, output: 300},
		{class: E2, input: 2000, output: 150, cacheRead: 1800},
		{class: E2, input: 2000, output: 140, cacheRead: 1800},
		{class: E3, input: 2100, output: 180, cacheCreate: 1900},
		{class: E5, input: 300, output: 100},
		{class: E6, input: 1200, output: 900},
		{class: E4, input: 900, output: 40, cacheRead: 800, stream: true}, // cancelled before completion
	}

	var idx atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := script[int(idx.Add(1))-1]
		if c.stream {
			w.Header().Set("Content-Type", "text/event-stream")
			flusher := w.(http.Flusher)
			write := func(typ, data string) {
				w.Write([]byte("event: " + typ + "\ndata: " + data + "\n\n"))
				flusher.Flush()
			}
			write("message_start", fmt.Sprintf(`{"type":"message_start","message":{"id":"m","type":"message","role":"assistant","model":"claude-test","content":[],"stop_reason":null,"usage":{"input_tokens":%d,"output_tokens":0,"cache_read_input_tokens":%d,"cache_creation_input_tokens":%d}}}`, c.input, c.cacheRead, c.cacheCreate))
			write("content_block_start", `{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`)
			write("message_delta", fmt.Sprintf(`{"type":"message_delta","delta":{"stop_reason":null},"usage":{"input_tokens":%d,"output_tokens":%d,"cache_read_input_tokens":%d,"cache_creation_input_tokens":%d}}`, c.input, c.output, c.cacheRead, c.cacheCreate))
			<-r.Context().Done() // hang — the test cancels; message_stop never arrives
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": "msg", "type": "message", "role": "assistant", "model": "claude-test",
			"content":     []map[string]any{{"type": "text", "text": "ok"}},
			"stop_reason": "end_turn",
			"usage": map[string]any{
				"input_tokens": c.input, "output_tokens": c.output,
				"cache_read_input_tokens": c.cacheRead, "cache_creation_input_tokens": c.cacheCreate,
			},
		})
	}))
	t.Cleanup(srv.Close)

	client := testClient(t, srv.URL)
	for _, c := range script {
		if !c.stream {
			if _, err := client.Send(context.Background(), c.class, prompt.Assembled{Stable: "S", Variable: "v"}); err != nil {
				t.Fatalf("Send(%s): %v", c.class, err)
			}
			continue
		}

		ctx, cancel := context.WithCancel(context.Background())
		stream, err := client.Stream(ctx, c.class, prompt.Assembled{Stable: "S", Variable: "v"})
		if err != nil {
			t.Fatalf("Stream(%s): %v", c.class, err)
		}
		for stream.Next() {
			if stream.Current().Type == "message_delta" {
				cancel() // partial usage is in hand; sever the stale thought (RT-2)
			}
		}
		if err := stream.Err(); !errors.Is(err, context.Canceled) {
			t.Fatalf("stream.Err() = %v, want context.Canceled", err)
		}
		cancel()
	}

	report := client.Accounting().Report()

	want := map[Class]Stats{
		E1: {Calls: 1, InputTokens: 500, OutputTokens: 300},
		E2: {Calls: 2, InputTokens: 4000, OutputTokens: 290, CacheReadTokens: 3600},
		E3: {Calls: 1, InputTokens: 2100, OutputTokens: 180, CacheCreationTokens: 1900},
		E5: {Calls: 1, InputTokens: 300, OutputTokens: 100},
		E6: {Calls: 1, InputTokens: 1200, OutputTokens: 900},
		E4: {Calls: 1, CancelledCalls: 1, InputTokens: 900, OutputTokens: 40, CacheReadTokens: 800},
	}
	if len(report) != len(want) {
		t.Fatalf("Report() has %d classes, want %d: %+v", len(report), len(want), report)
	}
	for class, w := range want {
		got, ok := report[class]
		if !ok {
			t.Errorf("Report() missing class %s", class)
			continue
		}
		if got != w {
			t.Errorf("Report()[%s] = %+v, want %+v", class, got, w)
		}
	}
}
