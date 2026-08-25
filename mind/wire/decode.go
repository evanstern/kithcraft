package wire

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"unicode/utf8"
)

// Decode parses one frame body into the wire's comparison value shape:
// objects as map[string]any, arrays as []any, numbers as int64 (C-6 — every
// v0 number is an integer; a value outside signed-64-bit range is refused
// rather than silently widened to a float), and strings/bools/null as
// themselves.
//
// Decode is deliberately tolerant, per the decode half of V-1..V-3: it does
// not check enum membership or interpret any field's meaning, so an unknown
// field (V-1), an unrecognized enum value such as an unfamiliar origin
// (V-2), or an unfamiliar percept_type (V-3) all decode successfully and
// are retained uninterpreted. Classifying what any of them means — V-2's
// fallback handling, V-6's secondhand classification — is the seam layer's
// job, above this package. Presence-checking required fields (V-5) is
// Validate's job, deliberately kept separate so decode never rejects on
// meaning, only on shape.
func Decode(body []byte) (any, error) {
	if !utf8.Valid(body) {
		return nil, fmt.Errorf("malformed: body is not valid UTF-8")
	}
	if bytes.HasPrefix(body, []byte{0xEF, 0xBB, 0xBF}) {
		return nil, fmt.Errorf("malformed: body carries a BOM (C-1)")
	}

	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	var raw any
	if err := dec.Decode(&raw); err != nil {
		return nil, fmt.Errorf("malformed: not valid JSON: %w", err)
	}
	if dec.More() {
		return nil, fmt.Errorf("malformed: trailing content after the message (one frame, one message)")
	}
	return normalize(raw)
}

func normalize(v any) (any, error) {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, mv := range t {
			cv, err := normalize(mv)
			if err != nil {
				return nil, err
			}
			out[k] = cv
		}
		return out, nil
	case []any:
		out := make([]any, len(t))
		for i, av := range t {
			cv, err := normalize(av)
			if err != nil {
				return nil, err
			}
			out[i] = cv
		}
		return out, nil
	case json.Number:
		n, err := strconv.ParseInt(t.String(), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("malformed: number %q is not a signed 64-bit integer (C-6)", t.String())
		}
		return n, nil
	default: // string, bool, nil
		return v, nil
	}
}

// requiredPayloadFields lists, per message type, the payload members the
// protocol (body-protocol-v0.md) marks required (✔ in its tables).
// Required-and-nullable counts as present carrying null (C-9); optional
// members are simply absent and are not listed here. Enum-valued fields
// (origin, percept_type, verb, …) are checked for presence only, never for
// membership — V-2/V-3 tolerance is the point.
var requiredPayloadFields = map[string][]string{
	"percept":       {"percept_id", "percept_type", "urgency", "provenance", "place", "content"},
	"intent":        {"intent_id", "verb", "target", "reason"},
	"intent_ack":    {"intent_id", "accepted", "reason_code"},
	"cancel":        {"intent_id"},
	"session_open":  {"time_unit", "capabilities"},
	"session_close": {"reason", "detail"},
}

// Validate applies V-5's presence-checked required-field rules to a decoded
// message: a required field that is absent is malformed and MUST be
// refused, never defaulted (EH-2a is the sharpest instance — a percept with
// no provenance must not be given an assumed origin). Validate deliberately
// stops there: it does not check any field's value, because V-2/V-3 require
// an unrecognized enum value or percept_type to be accepted, not refused —
// that judgment belongs above this package.
//
// ponytail: duplicate-key detection (C-4) is not implemented — encoding/json
// silently last-wins on a duplicate object key, and no vector in
// seam/vectors/ exercises it. Add if a vector ever pins C-4's refusal.
func Validate(v any) error {
	msg, ok := v.(map[string]any)
	if !ok {
		return fmt.Errorf("malformed:message is not a JSON object")
	}
	for _, f := range []string{"protocol", "message", "session", "seq", "body", "world_time", "payload"} {
		if _, present := msg[f]; !present {
			return fmt.Errorf("missing_required_field:%s", f)
		}
	}

	kind, ok := msg["message"].(string)
	if !ok {
		return fmt.Errorf("malformed:message is not a string")
	}
	payload, ok := msg["payload"].(map[string]any)
	if !ok {
		return fmt.Errorf("malformed:payload is not a JSON object")
	}
	fields, known := requiredPayloadFields[kind]
	if !known {
		return fmt.Errorf("malformed:unknown message type %q", kind)
	}
	for _, f := range fields {
		if _, present := payload[f]; !present {
			return fmt.Errorf("missing_required_field:payload.%s", f)
		}
	}

	// EH-2a: provenance is required on every percept and, being an object,
	// is checked one level deeper — a provenance without an origin is as
	// unusable as none at all, and neither may be defaulted.
	if kind == "percept" {
		prov, ok := payload["provenance"].(map[string]any)
		if !ok {
			return fmt.Errorf("malformed:payload.provenance is not a JSON object")
		}
		for _, f := range []string{"origin", "source", "observed_at", "received_at"} {
			if _, present := prov[f]; !present {
				return fmt.Errorf("missing_required_field:payload.provenance.%s", f)
			}
		}
	}
	return nil
}
