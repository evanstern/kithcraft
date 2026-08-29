package deliberate

import (
	"context"
	"errors"
	"sync"
	"testing"

	"kithcraft/mind/llm"
	"kithcraft/mind/memory"
)

// queueProposer replays raw responses in order, then returns ErrDone.
func queueProposer(raw ...string) Proposer {
	i := 0
	return func(ctx context.Context) (string, error) {
		if i >= len(raw) {
			return "", ErrDone
		}
		r := raw[i]
		i++
		return r, nil
	}
}

func tokensWith(kind, token string) *Tokens {
	t := NewTokens()
	t.mark(kind, token)
	return t
}

// TestLoop_TreatsOnlyActResultAsFact proves card AC #1: the loop's fact
// sink fires only once Deliver hands it a real act_result, never merely
// because an intent was composed and sent — the REQUEST/FACT split this
// whole package exists to enforce.
func TestLoop_TreatsOnlyActResultAsFact(t *testing.T) {
	sent := make(chan map[string]any, 1)
	vendor := VendorFunc(func(payload map[string]any) error {
		sent <- payload
		return nil
	})

	var mu sync.Mutex
	var facts []map[string]any

	l := New(Config{
		Verbs:  map[string]bool{"go_to": true},
		Tokens: tokensWith("place", "pl-1"),
		Vendor: vendor,
		Class:  llm.E2,
		OnFact: func(percept map[string]any) {
			mu.Lock()
			facts = append(facts, percept)
			mu.Unlock()
		},
	})

	propose := queueProposer(`{"verb":"go_to","target":{"type":"place","place":"pl-1"},"reason":"to look"}`)

	done := make(chan struct{})
	var res Result
	var runErr error
	go func() {
		res, runErr = l.Run(context.Background(), propose)
		close(done)
	}()

	payload := <-sent
	intentID, _ := payload["intent_id"].(string)
	if intentID == "" {
		t.Fatal("sent payload carries no intent_id")
	}

	mu.Lock()
	gotBeforeFact := len(facts)
	mu.Unlock()
	if gotBeforeFact != 0 {
		t.Fatal("OnFact fired before any act_result was delivered — a fact must derive from act_result, never from having sent the intent (card AC #1)")
	}

	l.Deliver(map[string]any{
		"percept_type": "act_result",
		"content":      map[string]any{"intent_id": intentID, "verb": "go_to", "outcome": "completed", "reason": "to look"},
	})

	<-done
	if runErr != nil {
		t.Fatalf("Run: %v", runErr)
	}
	if !res.Done || res.Iterations != 1 {
		t.Fatalf("Result = %+v, want Done=true Iterations=1", res)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(facts) != 1 {
		t.Fatalf("OnFact called %d times, want exactly 1 (once per act_result)", len(facts))
	}
	content, _ := facts[0]["content"].(map[string]any)
	if content["outcome"] != "completed" {
		t.Fatalf("fact content = %#v, want outcome completed", content)
	}
}

// TestLoop_ActResultForUnknownIntentID_Ignored is the T004 edge case:
// act_result naming an intent this Loop never composed must not panic and
// must not fire OnFact.
func TestLoop_ActResultForUnknownIntentID_Ignored(t *testing.T) {
	l := New(Config{Verbs: map[string]bool{"wait": true}, Vendor: VendorFunc(func(map[string]any) error { return nil })})
	var called bool
	l.cfg.OnFact = func(map[string]any) { called = true }
	l.Deliver(map[string]any{"content": map[string]any{"intent_id": "i-ghost", "outcome": "completed"}})
	if called {
		t.Fatal("OnFact fired for an intent_id this Loop never composed")
	}
}

// TestLoop_UndeclaredVerb_RefusedAsDeliberationFailure proves card AC #2
// via the loop's actual behaviour: a verb outside Config.Verbs is refused
// at compose (V-4, seam.Pending's existing behaviour) and surfaced as a
// returned error — never a crash, and no intent is sent.
func TestLoop_UndeclaredVerb_RefusedAsDeliberationFailure(t *testing.T) {
	var sent bool
	vendor := VendorFunc(func(map[string]any) error { sent = true; return nil })
	l := New(Config{Verbs: map[string]bool{"go_to": true}, Vendor: vendor, Class: llm.E2})

	_, err := l.Run(context.Background(), queueProposer(`{"verb":"fly","reason":"why not"}`))
	if err == nil {
		t.Fatal("expected Run to return an error for an undeclared verb")
	}
	if sent {
		t.Fatal("an intent refused at compose must never reach Vendor.SendIntent")
	}
}

// TestLoop_DescriptiveTarget_RejectedBeforeCompose proves card AC #7: a
// target naming a description rather than a known token is rejected
// before it ever reaches seam.Pending.Compose (no intent sent).
func TestLoop_DescriptiveTarget_RejectedBeforeCompose(t *testing.T) {
	var sent bool
	vendor := VendorFunc(func(map[string]any) error { sent = true; return nil })
	l := New(Config{Verbs: map[string]bool{"go_to": true}, Tokens: NewTokens(), Vendor: vendor, Class: llm.E2})

	_, err := l.Run(context.Background(), queueProposer(`{"verb":"go_to","target":"the nearest bed","reason":"tired"}`))
	if err == nil {
		t.Fatal("expected Run to reject a descriptive (bare-string) target")
	}
	if sent {
		t.Fatal("a rejected target must never reach Vendor.SendIntent")
	}
}

// TestLoop_UnknownTokenTarget_RejectedBeforeCompose is AC #7's other half:
// a structurally valid target naming a token this mind was never shown is
// rejected the same way a description is.
func TestLoop_UnknownTokenTarget_RejectedBeforeCompose(t *testing.T) {
	vendor := VendorFunc(func(map[string]any) error { t.Fatal("must not send"); return nil })
	l := New(Config{Verbs: map[string]bool{"go_to": true}, Tokens: NewTokens(), Vendor: vendor, Class: llm.E2})

	_, err := l.Run(context.Background(), queueProposer(`{"verb":"go_to","target":{"type":"place","place":"pl-never-seen"},"reason":"why"}`))
	if err == nil {
		t.Fatal("expected Run to reject a target token this mind was never given")
	}
}

// TestLoop_MaxIterations_BoundsWithoutADone proves FR-001: a Proposer
// that never signals ErrDone still terminates the loop, bounded by
// Config.MaxIterations rather than running forever.
func TestLoop_MaxIterations_BoundsWithoutADone(t *testing.T) {
	vendor := VendorFunc(func(map[string]any) error { return nil })
	l := New(Config{
		Verbs:         map[string]bool{"wait": true},
		Vendor:        vendor,
		Class:         llm.E2,
		MaxIterations: 2,
	})
	// Auto-deliver so Run never blocks: every SendIntent immediately
	// resolves its own intent, letting the loop actually reach the bound.
	l.cfg.Vendor = VendorFunc(func(payload map[string]any) error {
		go l.Deliver(map[string]any{"content": map[string]any{"intent_id": payload["intent_id"], "outcome": "completed"}})
		return nil
	})

	calls := 0
	propose := func(ctx context.Context) (string, error) {
		calls++
		return `{"verb":"wait","reason":"still deciding"}`, nil
	}

	res, err := l.Run(context.Background(), propose)
	if !errors.Is(err, ErrBoundExceeded) {
		t.Fatalf("Run err = %v, want ErrBoundExceeded", err)
	}
	if res.Iterations != 2 || res.Done {
		t.Fatalf("Result = %+v, want Iterations=2 Done=false", res)
	}
	if calls != 2 {
		t.Fatalf("propose called %d times, want exactly MaxIterations=2", calls)
	}
}

// TestLoop_FactWiresIntoAdmissionGate proves T004's other half: what
// reaches mind/memory's admission gate is exactly the act_result percept
// Deliver was given, unwrapped nowhere else — a real Gate.Decide call
// sees an authored reason and admits it on RuleActedReason.
func TestLoop_FactWiresIntoAdmissionGate(t *testing.T) {
	gate := memory.NewGate()
	var admitted bool
	var rule memory.AdmitRule

	sent := make(chan map[string]any, 1)
	l := New(Config{
		Verbs:  map[string]bool{"go_to": true},
		Tokens: tokensWith("place", "pl-1"),
		Vendor: VendorFunc(func(payload map[string]any) error { sent <- payload; return nil }),
		Class:  llm.E2,
		OnFact: func(percept map[string]any) { admitted, rule = gate.Decide(percept) },
	})

	done := make(chan struct{})
	go func() {
		l.Run(context.Background(), queueProposer(`{"verb":"go_to","target":{"type":"place","place":"pl-1"},"reason":"to see the orchard"}`))
		close(done)
	}()

	payload := <-sent
	intentID, _ := payload["intent_id"].(string)
	l.Deliver(map[string]any{
		"percept_type": "act_result",
		"urgency":      "background",
		"provenance":   map[string]any{"origin": "acted", "source": nil, "observed_at": int64(0), "received_at": int64(0)},
		"place":        nil,
		"content":      map[string]any{"intent_id": intentID, "verb": "go_to", "outcome": "completed", "reason": "to see the orchard"},
	})
	<-done

	if !admitted || rule != memory.RuleActedReason {
		t.Fatalf("gate.Decide(fact) = admitted=%v rule=%v, want admitted=true rule=%v", admitted, rule, memory.RuleActedReason)
	}
}
