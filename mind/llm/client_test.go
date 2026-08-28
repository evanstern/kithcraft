package llm

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/anthropics/anthropic-sdk-go/option"

	"kithcraft/mind/prompt"
)

// blockingTransport is a client-side http.RoundTripper that hangs until
// the request's own context is done — the same seam the SDK's own tests
// use (anthropic-sdk-go/client_test.go's closureTransport) to prove
// cancellation without depending on TCP-level connection-close detection,
// which is unreliable to assert on within a test's timeout.
type blockingTransport struct{}

func (blockingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	<-req.Context().Done()
	return nil, req.Context().Err()
}

// minimalMessage writes a valid, minimal Anthropic Messages API response.
func minimalMessage(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"id": "msg_test", "type": "message", "role": "assistant",
		"model":       "claude-test",
		"content":     []map[string]any{{"type": "text", "text": "ok"}},
		"stop_reason": "end_turn",
		"usage":       map[string]any{"input_tokens": 10, "output_tokens": 5},
	})
}

func testClient(t *testing.T, baseURL string) *Client {
	t.Helper()
	return New(option.WithBaseURL(baseURL), option.WithAPIKey("test-key"))
}

// TestSendCancellation proves card AC #5 / RT-2: an in-flight call
// terminates promptly and cleanly when its context is cancelled — the
// §5.5 interrupt mechanism IS cancellation.
func TestSendCancellation(t *testing.T) {
	c := New(option.WithAPIKey("test-key"), option.WithHTTPClient(&http.Client{Transport: blockingTransport{}}))
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		_, err := c.Send(ctx, E2, prompt.Assembled{Variable: "hi"})
		done <- err
	}()

	time.Sleep(50 * time.Millisecond) // let the call reach the server
	start := time.Now()
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Send err = %v, want context.Canceled", err)
		}
		if elapsed := time.Since(start); elapsed > 2*time.Second {
			t.Fatalf("Send took %s to terminate after cancel, want prompt", elapsed)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Send did not terminate after context cancellation")
	}
}

// TestRetryOnTransient proves RT-7: a transient (5xx) failure is retried
// with backoff and a subsequent success is returned — the SDK's own
// retry loop, exercised through this wrapper's default posture.
func TestRetryOnTransient(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits.Add(1) < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		minimalMessage(w)
	}))
	t.Cleanup(srv.Close)

	c := testClient(t, srv.URL)
	msg, err := c.Send(context.Background(), E2, prompt.Assembled{Variable: "hi"})
	if err != nil {
		t.Fatalf("Send after transient failures: %v", err)
	}
	if msg == nil || len(msg.Content) == 0 {
		t.Fatal("Send returned no content after retry succeeded")
	}
	if got := hits.Load(); got != 3 {
		t.Errorf("server saw %d requests, want 3 (1 initial + 2 retries, DefaultMaxRetries=%d)", got, DefaultMaxRetries)
	}
}

// TestStreamDelivery proves RT-1: a streaming call delivers progressive,
// usable text before the stream ends (E4's latency budget).
func TestStreamDelivery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)
		events := []struct{ typ, data string }{
			{"message_start", `{"type":"message_start","message":{"id":"m1","type":"message","role":"assistant","model":"claude-test","content":[],"stop_reason":null,"usage":{"input_tokens":10,"output_tokens":0}}}`},
			{"content_block_start", `{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`},
			{"content_block_delta", `{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}`},
			{"content_block_stop", `{"type":"content_block_stop","index":0}`},
			{"message_delta", `{"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":1}}`},
			{"message_stop", `{"type":"message_stop"}`},
		}
		for _, e := range events {
			w.Write([]byte("event: " + e.typ + "\ndata: " + e.data + "\n\n"))
			if flusher != nil {
				flusher.Flush()
			}
		}
	}))
	t.Cleanup(srv.Close)

	c := testClient(t, srv.URL)
	stream, err := c.Stream(context.Background(), E4, prompt.Assembled{Stable: "STABLE", Variable: "hi"})
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}
	defer stream.Close()

	var gotDelta bool
	for stream.Next() {
		ev := stream.Current()
		if ev.Type == "content_block_delta" && ev.Delta.Text != "" {
			gotDelta = true
		}
	}
	if err := stream.Err(); err != nil {
		t.Fatalf("stream.Err() = %v", err)
	}
	if !gotDelta {
		t.Error("stream delivered no text_delta event — no progressive/usable first token")
	}
}

// TestBreakpointPlacement proves RT-3/§4.3: the request carries an
// explicit cache_control breakpoint on the system block for cached classes
// (E2/E3/E4), and none for classes that have a stable prefix but are not
// cached (E5/E6) or have no stable prefix at all (E1).
func TestBreakpointPlacement(t *testing.T) {
	cases := []struct {
		class      Class
		stable     string
		wantSystem bool
		wantCache  bool
	}{
		{E1, "", false, false},      // no stable prefix at all
		{E2, "STABLE-E2", true, true},
		{E3, "STABLE-E3", true, true},
		{E4, "STABLE-E4", true, true},
		{E5, "STABLE-E5", true, false}, // stable prefix, but not cached
		{E6, "STABLE-E6", true, false}, // stable prefix, deliberately not cached
	}
	for _, c := range cases {
		t.Run(string(c.class), func(t *testing.T) {
			var got map[string]any
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				json.NewDecoder(r.Body).Decode(&got)
				minimalMessage(w)
			}))
			t.Cleanup(srv.Close)

			client := testClient(t, srv.URL)
			if _, err := client.Send(context.Background(), c.class, prompt.Assembled{Stable: c.stable, Variable: "x"}); err != nil {
				t.Fatalf("Send: %v", err)
			}

			sys, hasSystem := got["system"]
			if hasSystem != c.wantSystem {
				t.Fatalf("system field present = %v, want %v (got: %v)", hasSystem, c.wantSystem, got["system"])
			}
			if !c.wantSystem {
				return
			}
			blocks, ok := sys.([]any)
			if !ok || len(blocks) != 1 {
				t.Fatalf("system = %#v, want one text block", sys)
			}
			block, ok := blocks[0].(map[string]any)
			if !ok {
				t.Fatalf("system[0] = %#v, want an object", blocks[0])
			}
			_, hasCache := block["cache_control"]
			if hasCache != c.wantCache {
				t.Errorf("cache_control present = %v, want %v (block: %v)", hasCache, c.wantCache, block)
			}
		})
	}
}

// TestModelPrefix proves ANTHROPIC_MODEL_PREFIX is applied to the request's
// model ID when set, and leaves it untouched when unset — the classes.go
// canonical IDs stay the product's choice; only the wire spelling changes.
func TestModelPrefix(t *testing.T) {
	params, err := buildParams(E1, prompt.Assembled{Variable: "x"})
	if err != nil {
		t.Fatalf("buildParams: %v", err)
	}
	if got := string(params.Model); got != ModelOpus5 {
		t.Fatalf("Model = %q, want unprefixed %q", got, ModelOpus5)
	}

	t.Setenv("ANTHROPIC_MODEL_PREFIX", "cc/")
	params, err = buildParams(E1, prompt.Assembled{Variable: "x"})
	if err != nil {
		t.Fatalf("buildParams: %v", err)
	}
	if want := "cc/" + ModelOpus5; string(params.Model) != want {
		t.Fatalf("Model = %q, want %q", params.Model, want)
	}
}
