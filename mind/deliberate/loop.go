// Package deliberate: the bounded deliberation loop (spec.md FR-001,
// plan.md design decision 1) — porting only the *shape* of promptworld
// I's toolloop (decision-0003: its code does not survive the seam) onto
// body-protocol-v0.md's own request/fact split: a tool call there is a
// REQUEST; here, an intent is the REQUEST (§5.2), an act_result is the
// FACT (§5.4), and the admission gate (mind/memory) decides what of the
// fact becomes memory. This package owns none of that gate's logic — it
// only guarantees the fact it hands outward derives from act_result and
// nothing else (card AC #1).
//
// No changes to the wire or protocol surface: this is pure mind-side
// composition above mind/seam (Pending, verb refusal), mind/llm
// (structured-output decode, A-9), and mind/memory (the gate a Loop's
// OnFact typically wires to). No live model call happens in this package
// or its tests — Proposer is supplied by the caller, scripted in tests
// exactly as mind/llm's own tests script the SDK transport.
package deliberate

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"kithcraft/mind/llm"
	"kithcraft/mind/seam"
)

// DefaultMaxIterations bounds a Run when Config.MaxIterations is left at
// zero — the toolloop-shape safety valve (FR-001): a deliberation that
// keeps proposing must still terminate. ponytail: not derived from a
// measured E2/E3 round count (no live runs exist yet); raise it if the
// scripted-evening check (tasks.md T010) finds a legitimate deliberation
// tripping it.
const DefaultMaxIterations = 5

// ErrDone is a Proposer's signal that this deliberation has nothing
// further to add — toolloop's "model done" termination, mapped onto a
// single return value since A-9's structured output is one intent per
// round and an empty verb is already ParseIntent's own failure mode.
var ErrDone = errors.New("deliberate: nothing further to propose")

// ErrBoundExceeded is returned when Run reaches Config.MaxIterations
// without ErrDone — the bounded-loop safety valve tripping, not a crash.
var ErrBoundExceeded = errors.New("deliberate: max iterations reached without the deliberation signalling done")

// Proposer produces one round's raw structured-output text (A-9): the
// model's parsed-later intent. A real Proposer wraps an *llm.Client.Send
// call for l.Class plus its response text; tests supply a scripted queue
// directly, the same posture mind/llm/client_test.go takes toward the SDK
// transport (no live API calls, spec.md Success Criteria).
type Proposer func(ctx context.Context) (raw string, err error)

// Vendor is a Loop's outbound half: hand a composed intent payload
// (seam.Pending.Compose's return value) to whatever sends it — a real
// wiring wraps seam.Conn.WriteMessage with envelope stamping (protocol,
// session, seq, body, world_time), which is connection bookkeeping this
// package deliberately does not own (plan.md: "no changes to the wire or
// protocol surface").
type Vendor interface {
	SendIntent(payload map[string]any) error
}

// VendorFunc adapts a plain func to Vendor, mirroring http.HandlerFunc.
type VendorFunc func(payload map[string]any) error

func (f VendorFunc) SendIntent(payload map[string]any) error { return f(payload) }

// Config is one Loop's fixed setup. Verbs MUST come from ManifestVerbs
// against the body's actual session_open capabilities (FR-002) — nothing
// in this package supplies a fallback vocabulary. Tokens defaults to an
// empty set (NewTokens) when nil, which refuses every non-nil target
// until Observe has been fed real percepts.
type Config struct {
	Verbs         map[string]bool
	Tokens        *Tokens
	Vendor        Vendor
	Class         llm.Class
	OnFact        func(percept map[string]any) // called ONLY from Deliver — see Deliver's doc (card AC #1)
	MaxIterations int
}

// Result reports how a Run ended: how many rounds it ran, and whether the
// Proposer signalled ErrDone (Done=true) rather than the bound tripping.
type Result struct {
	Iterations int
	Done       bool
}

// Loop is one body's bounded deliberation: compose → send → await fact →
// hand to OnFact → repeat, bounded by Config.MaxIterations. It owns a
// seam.Pending scoped to Config.Verbs (so verb refusal is the existing,
// already-tested V-4 behaviour, not reimplemented here) and the
// intent_id -> waiting-round bookkeeping Deliver resolves against.
type Loop struct {
	cfg     Config
	pending *seam.Pending
	seq     atomic.Int64

	mu      sync.Mutex
	waiters map[string]chan map[string]any
}

// New builds a Loop from cfg. cfg.Verbs should be ManifestVerbs(...) of
// the body's actual manifest — see manifest.go.
func New(cfg Config) *Loop {
	if cfg.Tokens == nil {
		cfg.Tokens = NewTokens()
	}
	return &Loop{
		cfg:     cfg,
		pending: seam.NewPending(cfg.Verbs),
		waiters: map[string]chan map[string]any{},
	}
}

// Run drives the bounded loop: each round asks propose for one round's
// raw structured output, decodes it as an llm.Intent (A-9), validates its
// target is a known token (card AC #7), composes it through seam.Pending
// (refusing an undeclared verb per V-4, card AC #2), hands it to
// Config.Vendor, and blocks for the matching act_result via Deliver
// before continuing. propose returning ErrDone ends the run cleanly;
// reaching Config.MaxIterations without it returns ErrBoundExceeded — a
// deliberation failure, never a panic (spec.md Edge Cases).
func (l *Loop) Run(ctx context.Context, propose Proposer) (Result, error) {
	max := l.cfg.MaxIterations
	if max <= 0 {
		max = DefaultMaxIterations
	}

	var res Result
	for ; res.Iterations < max; res.Iterations++ {
		raw, err := propose(ctx)
		if errors.Is(err, ErrDone) {
			res.Done = true
			return res, nil
		}
		if err != nil {
			return res, fmt.Errorf("deliberate: propose: %w", err)
		}

		intent, err := llm.ParseIntent(l.cfg.Class, raw)
		if err != nil {
			return res, err
		}
		if err := l.cfg.Tokens.ValidateTarget(intent.Target); err != nil {
			return res, err
		}

		id := fmt.Sprintf("i-%d", l.seq.Add(1))
		payload, err := l.pending.Compose(id, intent.Verb, intent.Target, intent.Reason, intent.Supersedes)
		if err != nil {
			return res, err
		}

		ch := make(chan map[string]any, 1)
		l.mu.Lock()
		l.waiters[id] = ch
		l.mu.Unlock()

		if err := l.cfg.Vendor.SendIntent(payload); err != nil {
			l.mu.Lock()
			delete(l.waiters, id)
			l.mu.Unlock()
			return res, err
		}

		select {
		case percept := <-ch:
			if l.cfg.OnFact != nil {
				l.cfg.OnFact(percept)
			}
		case <-ctx.Done():
			return res, ctx.Err()
		}
	}
	return res, ErrBoundExceeded
}

// Deliver feeds one act_result percept (body-protocol-v0.md §5.4's full
// payload shape — percept_type/urgency/provenance/place/content) to
// whichever Run call is waiting on its content.intent_id. This is the
// ONLY path by which Config.OnFact is ever invoked (card AC #1: the fact
// derives from act_result, never from an intent having merely been sent —
// SendIntent above never touches OnFact). A percept naming an intent_id
// this Loop never composed, or already resolved, is silently dropped —
// matching seam.Pending.ResolveActResult's existing ok=false posture
// (spec.md Edge Cases: "act_result for an unknown/expired intent:
// already ignored by seam bookkeeping").
func (l *Loop) Deliver(percept map[string]any) {
	content, _ := percept["content"].(map[string]any)
	id, _ := content["intent_id"].(string)
	if id == "" {
		return
	}
	l.pending.ResolveActResult(id)

	l.mu.Lock()
	ch, ok := l.waiters[id]
	delete(l.waiters, id)
	l.mu.Unlock()

	if ok {
		ch <- percept
	}
}
