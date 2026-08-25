// Package llm (this file): the anthropic-sdk-go wrapper (RT-1..RT-3,
// RT-5..RT-7, docs/design/llm-routing-and-budget.md §6.1) — request
// construction from classes.go's Registry, explicit cache-control
// breakpoint placement (§4.3), streaming (§5.2), and per-call model
// selection (§1.3). Retry-with-backoff (RT-7) is the SDK's own behavior
// (429/5xx/connection errors, exponential backoff — see
// internal/requestconfig): this wrapper does not reimplement it, only
// makes the retry count an explicit, testable posture instead of a silent
// default. Cancellation (RT-2) is likewise the SDK's own context handling
// — Send/Stream simply pass ctx through and the SDK returns ctx.Err()
// promptly once it is cancelled.
//
// Dependency: github.com/anthropics/anthropic-sdk-go v1.58.0
// (decision-0003 §8.2, the one sanctioned new dependency for this task).
// Module: github.com/anthropics/anthropic-sdk-go. Accessed: 2026-08-25.
package llm

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/ssestream"

	"kithcraft/mind/prompt"
)

// DefaultMaxRetries is the retry count this wrapper asks the SDK to use
// (RT-7). It matches the SDK's own default (internal/requestconfig.go), so
// nothing changes if a caller never touches it — this constant exists to
// make the posture explicit and testable rather than silently inherited.
const DefaultMaxRetries = 2

// Client wraps the Messages API for the six-class registry. It holds no
// per-call state, so a single Client is safe for the concurrent use RT-6
// asks for (3-6 calls, one per villager peak) — each call builds its own
// request and the underlying SDK client is itself concurrency-safe.
type Client struct {
	sdk anthropic.Client
}

// New builds a wrapper client. opts ride straight through to
// anthropic.NewClient — API key, base URL override, or a custom transport
// (option.WithHTTPClient) — exactly as the SDK itself takes them, so tests
// configure this the same way production code does. A caller-supplied
// option.WithMaxRetries overrides DefaultMaxRetries (opts are applied
// after it).
func New(opts ...option.RequestOption) *Client {
	all := append([]option.RequestOption{option.WithMaxRetries(DefaultMaxRetries)}, opts...)
	return &Client{sdk: anthropic.NewClient(all...)}
}

// Send performs one non-streaming call for class against the assembled
// prompt (RT-5: model comes from Registry[class], so concurrent calls in
// one process each carry their own class's model). Structured-output
// parsing is structured.go's job — Send returns the raw SDK message.
func (c *Client) Send(ctx context.Context, class Class, a prompt.Assembled) (*anthropic.Message, error) {
	params, err := buildParams(class, a)
	if err != nil {
		return nil, err
	}
	return c.sdk.Messages.New(ctx, params)
}

// Stream performs a streaming call (RT-1 — E4's latency budget). The
// caller ranges over the returned stream (stream.Next()/stream.Current())
// and must Close it when done. Cancelling ctx terminates the stream
// promptly (RT-2); the stream's Err() then reports the cancellation.
func (c *Client) Stream(ctx context.Context, class Class, a prompt.Assembled) (*ssestream.Stream[anthropic.MessageStreamEventUnion], error) {
	params, err := buildParams(class, a)
	if err != nil {
		return nil, err
	}
	return c.sdk.Messages.NewStreaming(ctx, params), nil
}

// buildParams translates a class + assembled prompt into the SDK's request
// shape, entirely from Registry[class] (classes.go) — a tier or posture
// change never touches this function. The assembled stable prefix becomes
// the system prompt; the variable context becomes the user message (§2.3's
// order). A cache-control breakpoint is placed on the system block only
// when both a stable prefix exists (E1 has none) and the class's Cached
// flag is set (§4.3: E2/E3/E4 yes; E1/E5/E6 no) — RT-3's placement is
// therefore structural, not a per-call decision a caller could get wrong.
func buildParams(class Class, a prompt.Assembled) (anthropic.MessageNewParams, error) {
	cfg, ok := Registry[class]
	if !ok {
		return anthropic.MessageNewParams{}, fmt.Errorf("llm: no registry entry for class %q", class)
	}

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(cfg.Model),
		MaxTokens: int64(cfg.MaxTokens),
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(a.Variable)),
		},
	}

	if a.Stable != "" {
		sys := anthropic.TextBlockParam{Text: a.Stable}
		if cfg.Cached {
			sys.CacheControl = anthropic.NewCacheControlEphemeralParam()
		}
		params.System = []anthropic.TextBlockParam{sys}
	}

	if cfg.ThinkingOn {
		// §5.1/§5.4 both name "adaptive thinking" explicitly for the
		// classes that have it on (E2/E3/E6) — the enabled+budget_tokens
		// variant requires budget_tokens < max_tokens with a 1024 floor,
		// which E2/E3's 1024-token ceiling can't satisfy; adaptive has no
		// such conflict and matches the doc's own wording.
		params.Thinking = anthropic.ThinkingConfigParamUnion{OfAdaptive: &anthropic.ThinkingConfigAdaptiveParam{}}
	}
	if cfg.Effort != "" {
		params.OutputConfig.Effort = anthropic.OutputConfigEffort(cfg.Effort)
	}
	if cfg.StructuredOutput {
		if schema, ok := SchemaFor(class); ok {
			params.OutputConfig.Format = anthropic.JSONOutputFormatParam{Schema: schema}
		}
	}

	return params, nil
}
