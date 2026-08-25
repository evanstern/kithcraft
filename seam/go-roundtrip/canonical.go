// Package roundtrip is a throwaway-grade proof of docs/design/seam-wire-v0.md.
// It is not the transport: M1 (TASK-0008) and V1 (TASK-0009) build those. Its only
// job is to make the golden vectors executable, so the wire is pinned by agreement
// between two independent implementations rather than by prose alone.
//
// Go stdlib only, by design: a dependency here would be a dependency on the wire.
package roundtrip

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// maxFrameBody is seam-wire-v0.md §2.5's cap: 1 MiB.
const maxFrameBody = 1 << 20

// ---------------------------------------------------------------------------
// Decoding
// ---------------------------------------------------------------------------

// DecodeFrame reads one frame from b — a 4-byte big-endian length prefix followed
// by that many bytes of UTF-8 JSON (§2.1) — and returns the normalized message plus
// the number of bytes consumed.
func DecodeFrame(b []byte) (any, int, error) {
	if len(b) < 4 {
		return nil, 0, fmt.Errorf("truncated frame: %d bytes, need at least the 4-byte length word", len(b))
	}
	n := binary.BigEndian.Uint32(b[:4])
	if n == 0 {
		return nil, 0, fmt.Errorf("malformed frame: length 0 (an empty body is not a JSON value)")
	}
	if n > maxFrameBody {
		return nil, 0, fmt.Errorf("frame body %d exceeds the 1 MiB cap", n)
	}
	if uint64(len(b)) < 4+uint64(n) {
		return nil, 0, fmt.Errorf("truncated frame: body declared %d bytes, %d available", n, len(b)-4)
	}
	v, err := Normalize(b[4 : 4+n])
	if err != nil {
		return nil, 0, err
	}
	return v, int(4 + n), nil
}

// Normalize parses JSON text into the comparison form used throughout: objects as
// map[string]any, arrays as []any, numbers as int64, and strings/bools/null as
// themselves. Numbers become int64 because C-6 makes every v0 number an integer and
// requires a receiver to refuse one outside signed-64-bit range rather than lose it
// to a double.
func Normalize(text []byte) (any, error) {
	if !utf8.Valid(text) {
		return nil, fmt.Errorf("body is not valid UTF-8")
	}
	if bytes.HasPrefix(text, []byte{0xEF, 0xBB, 0xBF}) {
		return nil, fmt.Errorf("body carries a BOM (C-1)")
	}
	dec := json.NewDecoder(bytes.NewReader(text))
	dec.UseNumber()
	var raw any
	if err := dec.Decode(&raw); err != nil {
		return nil, fmt.Errorf("not valid JSON: %w", err)
	}
	if dec.More() {
		return nil, fmt.Errorf("trailing content after the message (one frame, one message)")
	}
	return convert(raw)
}

func convert(v any) (any, error) {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, mv := range t {
			cv, err := convert(mv)
			if err != nil {
				return nil, err
			}
			out[k] = cv
		}
		return out, nil
	case []any:
		out := make([]any, len(t))
		for i, av := range t {
			cv, err := convert(av)
			if err != nil {
				return nil, err
			}
			out[i] = cv
		}
		return out, nil
	case json.Number:
		n, err := strconv.ParseInt(t.String(), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("number %q is not a signed 64-bit integer (C-6)", t.String())
		}
		return n, nil
	default: // string, bool, nil
		return v, nil
	}
}

// ---------------------------------------------------------------------------
// Canonical encoding (§2.4, C-1..C-10)
// ---------------------------------------------------------------------------

// EncodeFrame serializes v in canonical form and prefixes the 4-byte big-endian length.
func EncodeFrame(v any) ([]byte, error) {
	body, err := EncodeCanonical(v)
	if err != nil {
		return nil, err
	}
	if len(body) > maxFrameBody {
		return nil, fmt.Errorf("frame body %d exceeds the 1 MiB cap", len(body))
	}
	frame := make([]byte, 4, 4+len(body))
	binary.BigEndian.PutUint32(frame, uint32(len(body)))
	return append(frame, body...), nil
}

// EncodeCanonical writes the canonical JSON body. It is hand-rolled rather than
// delegated to encoding/json because the stdlib encoder violates C-7 twice: it
// escapes <, >, & and U+2028/U+2029 that RFC 8259 does not require escaped, and it
// spells \b and \f as /.
func EncodeCanonical(v any) ([]byte, error) {
	var sb strings.Builder
	if err := writeValue(&sb, v); err != nil {
		return nil, err
	}
	return []byte(sb.String()), nil
}

func writeValue(sb *strings.Builder, v any) error {
	switch t := v.(type) {
	case nil:
		sb.WriteString("null") // C-10
	case bool:
		if t {
			sb.WriteString("true")
		} else {
			sb.WriteString("false")
		}
	case int64:
		sb.WriteString(strconv.FormatInt(t, 10)) // C-6
	case string:
		return writeString(sb, t)
	case []any:
		sb.WriteByte('[')
		for i, e := range t { // C-5: array order is never touched
			if i > 0 {
				sb.WriteByte(',')
			}
			if err := writeValue(sb, e); err != nil {
				return err
			}
		}
		sb.WriteByte(']')
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys) // C-3: ascending by the key's UTF-8 bytes
		sb.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				sb.WriteByte(',') // C-2: no insignificant whitespace
			}
			if err := writeString(sb, k); err != nil {
				return err
			}
			sb.WriteByte(':')
			if err := writeValue(sb, t[k]); err != nil {
				return err
			}
		}
		sb.WriteByte('}')
	default:
		return fmt.Errorf("value of type %T is not representable on this wire", v)
	}
	return nil
}

// writeString applies C-7 and C-8: escape exactly what RFC 8259 requires and nothing
// more, non-ASCII as literal UTF-8, and never a lone surrogate.
func writeString(sb *strings.Builder, s string) error {
	if !utf8.ValidString(s) {
		return fmt.Errorf("string %q is not valid UTF-8 (C-8: a lone surrogate MUST NOT be emitted)", s)
	}
	sb.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			sb.WriteString(`\"`)
		case '\\':
			sb.WriteString(`\\`)
		case '\b':
			sb.WriteString(`\b`)
		case '\t':
			sb.WriteString(`\t`)
		case '\n':
			sb.WriteString(`\n`)
		case '\f':
			sb.WriteString(`\f`)
		case '\r':
			sb.WriteString(`\r`)
		default:
			if r < 0x20 {
				fmt.Fprintf(sb, `\u%04x`, r) // lowercase hex
			} else {
				sb.WriteRune(r) // literal UTF-8; / is not escaped
			}
		}
	}
	sb.WriteByte('"')
	return nil
}

// ---------------------------------------------------------------------------
// Message validation — only what the error vectors pin
// ---------------------------------------------------------------------------

// requiredPayloadFields lists, per message type, the payload members the protocol
// marks required (✔ in its tables). Required-and-nullable counts as present carrying
// null (C-9); optional members are absent entirely and are not listed here.
var requiredPayloadFields = map[string][]string{
	"percept":       {"percept_id", "percept_type", "urgency", "provenance", "place", "content"},
	"intent":        {"intent_id", "verb", "target", "reason"},
	"intent_ack":    {"intent_id", "accepted", "reason_code"},
	"cancel":        {"intent_id"},
	"session_open":  {"time_unit", "capabilities"},
	"session_close": {"reason", "detail"},
}

// Validate applies the required-field rules a receiver must apply before acting on a
// message (V-5). It deliberately stops there: enum membership is NOT checked, because
// V-2 requires an unrecognized value to decode and be classified above the codec —
// err_unknown_origin is exactly that case.
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
	// EH-2a: provenance is required on every percept and, being an object, is
	// checked one level deeper — a provenance without an origin is as unusable as
	// none at all, and neither may be defaulted.
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
