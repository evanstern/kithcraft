// Package llm (this file): structured-output shapes and parsing for
// E2/E3/E6 (A-9, RT-4) — output_config.format bounds these classes to a
// parsed value instead of prose, and a parse failure here is the typed
// ParseError, never a panic or a silently empty value a caller could
// mistake for "nothing to do" (spec.md User Story 3, Acceptance Scenario 2).
//
// Scope (spec.md FR-007): this file owns the structured-output *mechanism*
// and the minimal shape RT-4 needs to exist at all. Intent's fields mirror
// mind/seam's own intent shape (intents.go's Pending.Compose(intentID,
// verb, target, reason, supersedes)) so a deliberation's parsed output
// plugs directly into intent composition — no invented business logic, no
// translation layer. Digest is deliberately minimal: M7 owns nightly
// consolidation's actual belief/narrative content (ponytail: this is the
// wire shape only; widen its fields when M7 defines what a consolidation
// actually writes).
package llm

import (
	"encoding/json"
	"fmt"
)

// ParseError is structured-output's bounded failure mode: Class names
// which call failed, Raw preserves the text that didn't parse (for
// logging/debugging), and Err is the underlying decode error (Unwrap
// exposes it to errors.Is/As).
type ParseError struct {
	Class Class
	Raw   string
	Err   error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("llm: %s structured output parse failed: %v", e.Class, e.Err)
}

func (e *ParseError) Unwrap() error { return e.Err }

// Intent is E2/E3's structured decision (§1.3/A-9): a verb to enact, its
// target, and why.
type Intent struct {
	Verb       string `json:"verb"`
	Target     any    `json:"target,omitempty"`
	Reason     string `json:"reason,omitempty"`
	Supersedes string `json:"supersedes,omitempty"`
}

// IntentSchema is the JSON Schema sent as E2/E3's output_config.format.
var IntentSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"verb":       map[string]any{"type": "string"},
		"target":     map[string]any{},
		"reason":     map[string]any{"type": "string"},
		"supersedes": map[string]any{"type": "string"},
	},
	"required":             []string{"verb"},
	"additionalProperties": false,
}

// ParseIntent parses E2/E3's structured output text into an Intent. An
// empty verb is refused (an intent with no verb is not composable by
// seam.Pending.Compose, which requires a declared verb — V-4) — that
// refusal is folded into the bounded-failure mode here rather than left
// for a caller to discover deeper in the seam.
func ParseIntent(class Class, raw string) (Intent, error) {
	var v Intent
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return Intent{}, &ParseError{Class: class, Raw: raw, Err: err}
	}
	if v.Verb == "" {
		return Intent{}, &ParseError{Class: class, Raw: raw, Err: fmt.Errorf("empty verb")}
	}
	return v, nil
}

// Digest is E6's structured output (A-9): nightly consolidation's parsed
// result — a summary, the beliefs it formed, and the ordinal m1..mN
// references (body-protocol-v0.md §2.3, mind/consolidate T002) it cites as
// evidence. References is the field M4 ponytailed and M7 (mind/consolidate)
// completes: the prompt convention IS the identity mechanism, so the model
// cites "m3" rather than a (world_time, hash) pair it was never given, and
// mind/consolidate maps accepted references back to durable identity.
// Beliefs stays a bag of fields — its shape is still M7's to define beyond
// the reference convention.
type Digest struct {
	Summary    string           `json:"summary"`
	Beliefs    []map[string]any `json:"beliefs,omitempty"`
	References []string         `json:"references,omitempty"`
}

// DigestSchema is the JSON Schema sent as E6's output_config.format.
var DigestSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"summary":    map[string]any{"type": "string"},
		"beliefs":    map[string]any{"type": "array", "items": map[string]any{"type": "object"}},
		"references": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
	},
	"required":             []string{"summary"},
	"additionalProperties": false,
}

// ParseDigest parses E6's structured output text into a Digest.
func ParseDigest(class Class, raw string) (Digest, error) {
	var v Digest
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return Digest{}, &ParseError{Class: class, Raw: raw, Err: err}
	}
	return v, nil
}

// SchemaFor returns the output_config.format schema for a structured class
// (E2/E3/E6) and ok=true; a class with StructuredOutput=false (classes.go's
// Registry) returns ok=false — client.go uses this to decide whether to
// set OutputConfig.Format on the request at all.
func SchemaFor(class Class) (schema map[string]any, ok bool) {
	switch class {
	case E2, E3:
		return IntentSchema, true
	case E6:
		return DigestSchema, true
	default:
		return nil, false
	}
}
