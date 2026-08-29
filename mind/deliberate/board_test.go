package deliberate

import (
	"context"
	"testing"

	"kithcraft/mind/llm"
)

// TestLoop_E3ClaimIntent_CarriesAuthoredReason proves card AC #3: an E3
// deliberation's claim intent reaches Vendor.SendIntent carrying the
// model's own authored reason, unaltered — the plumbing this phase's
// design decision 5 says the (scripted, no-live-call) tests verify.
func TestLoop_E3ClaimIntent_CarriesAuthoredReason(t *testing.T) {
	const reason = "Building shelters is exactly my trade, and that north wall has bothered me all week."

	sent := make(chan map[string]any, 1)
	l := New(Config{
		Verbs:  map[string]bool{"claim": true},
		Tokens: tokensWith("thing_id", "th-post-1"),
		Vendor: VendorFunc(func(payload map[string]any) error { sent <- payload; return nil }),
		Class:  llm.E3,
	})

	done := make(chan struct{})
	go func() {
		l.Run(context.Background(), queueProposer(
			`{"verb":"claim","target":{"type":"thing","thing_id":"th-post-1"},"reason":"`+reason+`"}`,
		))
		close(done)
	}()

	payload := <-sent
	if payload["reason"] != reason {
		t.Fatalf("sent reason = %q, want %q", payload["reason"], reason)
	}
	intentID, _ := payload["intent_id"].(string)
	l.Deliver(map[string]any{"content": map[string]any{"intent_id": intentID, "outcome": "completed"}})
	<-done
}

// TestLoop_E3DeclineIntent_ReachableWithPersonaGroundedReason proves card
// AC #4: a decline is reachable (Run succeeds, no error, no special-cased
// rejection of the "decline" verb) and its reason — scripted here as this
// persona's own commitment, not a generic refusal — flows through exactly
// as authored.
func TestLoop_E3DeclineIntent_ReachableWithPersonaGroundedReason(t *testing.T) {
	const reason = "I promised Mira I'd mind the stall until dusk, and the wall can wait for hands that aren't already full."

	sent := make(chan map[string]any, 1)
	l := New(Config{
		Verbs:  map[string]bool{"decline": true},
		Tokens: tokensWith("thing_id", "th-post-1"),
		Vendor: VendorFunc(func(payload map[string]any) error { sent <- payload; return nil }),
		Class:  llm.E3,
	})

	done := make(chan struct{})
	var runErr error
	go func() {
		_, runErr = l.Run(context.Background(), queueProposer(
			`{"verb":"decline","target":{"type":"thing","thing_id":"th-post-1"},"reason":"`+reason+`"}`,
		))
		close(done)
	}()

	payload := <-sent
	intentID, _ := payload["intent_id"].(string)
	l.Deliver(map[string]any{"content": map[string]any{"intent_id": intentID, "outcome": "completed"}})
	<-done

	if runErr != nil {
		t.Fatalf("a decline must be reachable, got Run error: %v", runErr)
	}
	if payload["verb"] != "decline" {
		t.Fatalf("sent verb = %v, want decline", payload["verb"])
	}
	if payload["reason"] != reason {
		t.Fatalf("sent reason = %q, want the persona-grounded %q, not a generic refusal", payload["reason"], reason)
	}
}

// TestLoop_EmptyReasonIntent_RejectedBeforeCompose proves FR-004 (§5.2:
// "the mind must have a why"): an intent with no reason never reaches
// Vendor.SendIntent.
func TestLoop_EmptyReasonIntent_RejectedBeforeCompose(t *testing.T) {
	var sent bool
	vendor := VendorFunc(func(map[string]any) error { sent = true; return nil })
	l := New(Config{Verbs: map[string]bool{"claim": true}, Vendor: vendor, Class: llm.E3})

	_, err := l.Run(context.Background(), queueProposer(`{"verb":"claim"}`))
	if err == nil {
		t.Fatal("expected Run to reject an intent with no authored reason")
	}
	if sent {
		t.Fatal("an intent with no reason must never reach Vendor.SendIntent")
	}
}
