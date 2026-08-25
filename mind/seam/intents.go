// Package seam (this file): the mind's own intent bookkeeping — the
// pending set an intent joins when composed, supersedes replacement,
// act_result matching by intent_id, and cancel
// (docs/design/body-protocol-v0.md §5.1-§5.7). V-4 — refusing an intent
// naming a verb the manifest doesn't declare — is enforced here, at
// composition: verbs bind OUTBOUND intents, unlike ingest.go's V-5/V-6
// which bind inbound percepts.
package seam

import (
	"fmt"
	"sync"
)

// Pending is one body's outstanding-intent bookkeeping: which intents the
// mind is still waiting on an act_result for. Acceptance is not success
// (§5.1/§5.3) — an intent_ack never resolves an entry; only a matching
// act_result, or the mind's own cancel/supersedes, does.
type Pending struct {
	mu    sync.Mutex
	verbs map[string]bool
	open  map[string]map[string]any
}

// NewPending scopes a Pending to one body's declared verb set — the
// session manifest's capabilities.verbs (body-protocol-v0.md §5.5/§6.2).
// A nil verbs refuses every verb (an empty manifest declares nothing).
func NewPending(verbs map[string]bool) *Pending {
	return &Pending{verbs: verbs, open: map[string]map[string]any{}}
}

// Compose builds an intent payload (§5.2) and admits it to the pending
// set. An undeclared verb is refused here, before anything mutates
// (V-4) — the pending set and any supersedes target are left untouched.
// A non-empty supersedes removes that predecessor from the pending set:
// the mind has moved on from it, even though the vendor may still report
// its own eventual act_result (outcome: superseded) for it.
func (p *Pending) Compose(intentID, verb string, target any, reason, supersedes string) (map[string]any, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.verbs[verb] {
		return nil, fmt.Errorf("seam: verb %q is not declared in this body's manifest (V-4)", verb)
	}
	payload := map[string]any{
		"intent_id": intentID, "verb": verb, "target": target,
		"reason": nilIfEmpty(reason), "supersedes": nilIfEmpty(supersedes),
		"not_after": nil,
	}
	if supersedes != "" {
		delete(p.open, supersedes)
	}
	p.open[intentID] = payload
	return payload, nil
}

// Cancel requests abandonment of a pending intent (§5.7), removing it
// from the pending set and returning the cancel payload to send. ok is
// false when intentID is not pending (already resolved, superseded, or
// never issued) — a no-op, not an error.
func (p *Pending) Cancel(intentID string) (payload map[string]any, ok bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, present := p.open[intentID]; !present {
		return nil, false
	}
	delete(p.open, intentID)
	return map[string]any{"intent_id": intentID}, true
}

// ResolveActResult matches an act_result percept's intent_id against the
// pending set and removes it — every accepted intent produces exactly
// one act_result (§5.4), and this is that resolution. ok is false when
// intentID is not pending (already resolved, superseded, or unknown to
// this body — a vendor bug or a stale message, not a reason to mutate
// anything further).
func (p *Pending) ResolveActResult(intentID string) (intent map[string]any, ok bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	intent, ok = p.open[intentID]
	if ok {
		delete(p.open, intentID)
	}
	return intent, ok
}

// IsPending reports whether intentID is still awaiting a result.
func (p *Pending) IsPending(intentID string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	_, ok := p.open[intentID]
	return ok
}

func nilIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
