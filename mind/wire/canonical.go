package wire

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// EncodeCanonical writes v — built from the value shapes Decode produces
// (map[string]any, []any, int64, string, bool, nil) — as canonical JSON per
// §2.4, rules C-1..C-10: UTF-8 with no BOM, no insignificant whitespace,
// object keys sorted ascending by UTF-8 byte, integers only, minimal string
// escaping with literal (never \u-escaped) UTF-8, and lowercase literals.
//
// This is hand-rolled rather than delegated to encoding/json because the
// stdlib encoder is a known non-conformant writer for this wire (TASK-0007
// finding): it escapes <, >, &, and U+2028/U+2029 that C-7 does not require
// escaped, and it spells \b and \f as \u00xx instead of the two-character
// form. encoding/json remains fine for decoding (Decode uses it), since a
// tolerant reader has no canonical-form obligation.
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
		sb.WriteString(strconv.FormatInt(t, 10)) // C-6: integer, no fraction/exponent
	case string:
		return writeString(sb, t)
	case []any:
		sb.WriteByte('[')
		for i, e := range t { // C-5: array order is semantic and never touched
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
		return fmt.Errorf("wire: value of type %T is not representable on this wire", v)
	}
	return nil
}

// writeString applies C-7 (escape exactly what RFC 8259 requires and
// nothing more, non-ASCII as literal UTF-8) and C-8 (a lone surrogate MUST
// NOT be emitted).
func writeString(sb *strings.Builder, s string) error {
	if !utf8.ValidString(s) {
		return fmt.Errorf("wire: string %q is not valid UTF-8 (C-8: a lone surrogate MUST NOT be emitted)", s)
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
				fmt.Fprintf(sb, `\u%04x`, r) // lowercase hex, C-7
			} else {
				sb.WriteRune(r) // literal UTF-8; '/' is not escaped
			}
		}
	}
	sb.WriteByte('"')
	return nil
}
