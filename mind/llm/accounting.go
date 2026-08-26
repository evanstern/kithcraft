// Package llm (this file): per-class call/token accounting (FR-005, RT
// §6.2's "instrumented from the first call", card AC #7). A Client wires
// one Accounting in at New() — there is no opt-in step, so counting starts
// with the first Send or Stream. AccountedStream wraps the SDK's streaming
// response so a caller that merely ranges over Next()/Current(), exactly as
// client_test.go's existing tests already do, gets its usage recorded for
// free, including a cancelled call's partial usage (spec.md's Edge Cases).
package llm

import (
	"context"
	"errors"
	"sync"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/packages/ssestream"
)

// Stats is one class's tally: how many calls, how many of those were
// cancelled, and the token totals — including the cache read/write splits
// §4.3's caching design exists to measure, when the API reports them.
type Stats struct {
	Calls               int
	CancelledCalls      int
	InputTokens         int64
	OutputTokens        int64
	CacheReadTokens     int64
	CacheCreationTokens int64
}

// Accounting is a Client's per-class counters, safe for the concurrent use
// RT-6 asks of Client (3-6 calls at once, one goroutine per call).
type Accounting struct {
	mu    sync.Mutex
	stats map[Class]*Stats
}

func newAccounting() *Accounting {
	return &Accounting{stats: make(map[Class]*Stats)}
}

func (a *Accounting) record(class Class, input, output, cacheRead, cacheCreate int64, cancelled bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	s := a.stats[class]
	if s == nil {
		s = &Stats{}
		a.stats[class] = s
	}
	s.Calls++
	if cancelled {
		s.CancelledCalls++
	}
	s.InputTokens += input
	s.OutputTokens += output
	s.CacheReadTokens += cacheRead
	s.CacheCreationTokens += cacheCreate
}

// Report returns a snapshot of every class that has made at least one call
// — the session-end report card AC #7 asks for. A class that never ran is
// omitted rather than listed at zero.
func (a *Accounting) Report() map[Class]Stats {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make(map[Class]Stats, len(a.stats))
	for class, s := range a.stats {
		out[class] = *s
	}
	return out
}

// AccountedStream wraps the SDK's streaming response. It embeds the SDK
// stream so Next/Current/Err/Close all forward to it unchanged; Next is
// the only method overridden, to update a running usage snapshot from
// whichever event carries it — message_start's initial Usage, then
// message_delta's cumulative Usage overwriting it field-by-field, but only
// where the new value is nonzero, so a message_delta that reports only
// output_tokens (the API's traditional shape) can't clobber the input/cache
// totals message_start already gave. The moment iteration ends — drained,
// errored, or cancelled — the snapshot finalizes into Accounting exactly
// once, whether that happens inside Next (the caller's for-loop drains to
// false) or Close (a caller that breaks out early).
type AccountedStream struct {
	*ssestream.Stream[anthropic.MessageStreamEventUnion]
	class                                 Class
	acc                                   *Accounting
	input, output, cacheRead, cacheCreate int64
	finalized                             bool
}

func newAccountedStream(s *ssestream.Stream[anthropic.MessageStreamEventUnion], class Class, acc *Accounting) *AccountedStream {
	return &AccountedStream{Stream: s, class: class, acc: acc}
}

func (s *AccountedStream) Next() bool {
	if s.Stream.Next() {
		switch ev := s.Stream.Current(); ev.Type {
		case "message_start":
			s.apply(ev.Message.Usage.InputTokens, ev.Message.Usage.OutputTokens, ev.Message.Usage.CacheReadInputTokens, ev.Message.Usage.CacheCreationInputTokens)
		case "message_delta":
			s.apply(ev.Usage.InputTokens, ev.Usage.OutputTokens, ev.Usage.CacheReadInputTokens, ev.Usage.CacheCreationInputTokens)
		}
		return true
	}
	s.finalize()
	return false
}

func (s *AccountedStream) Close() error {
	s.finalize()
	return s.Stream.Close()
}

func (s *AccountedStream) apply(input, output, cacheRead, cacheCreate int64) {
	if input != 0 {
		s.input = input
	}
	if output != 0 {
		s.output = output
	}
	if cacheRead != 0 {
		s.cacheRead = cacheRead
	}
	if cacheCreate != 0 {
		s.cacheCreate = cacheCreate
	}
}

func (s *AccountedStream) finalize() {
	if s.finalized {
		return
	}
	s.finalized = true
	cancelled := errors.Is(s.Stream.Err(), context.Canceled)
	s.acc.record(s.class, s.input, s.output, s.cacheRead, s.cacheCreate, cancelled)
}
