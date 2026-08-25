package wire

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// vectorDir is relative to this package's directory (mind/wire), which is
// where `go test` sets the working directory: up two to the repo root, then
// into the shared fixture directory pinned by seam-wire-v0.md.
const vectorDir = "../../seam/vectors"

// A vector as it sits on disk. See seam/vectors/README.md for the format.
type vector struct {
	Name      string          `json:"name"`
	Direction string          `json:"direction"`
	Expect    string          `json:"expect"`
	Refusal   string          `json:"refusal"`
	Decoded   json.RawMessage `json:"decoded"`
	FrameHex  string          `json:"frame_hex"`
}

// census is contracts/vectors.md's closed list: nine percept types, three
// intent shapes, two session-lifecycle messages, three error/edge cases.
// The set is complete when every name here has a file and no file exists
// outside it.
var census = []string{
	"percept_sighting", "percept_observation", "percept_sound", "percept_speech",
	"percept_told_fact", "percept_text", "percept_act_result", "percept_self_state",
	"percept_change_report",
	"intent", "intent_ack", "cancel",
	"session_open", "session_close",
	"err_missing_provenance", "err_unknown_origin", "intent_ack_refused",
}

func loadVectors(t *testing.T) map[string]vector {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(vectorDir, "*.json"))
	if err != nil || len(paths) == 0 {
		t.Fatalf("no vectors found in %s: %v", vectorDir, err)
	}
	out := make(map[string]vector, len(paths))
	for _, p := range paths {
		raw, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		var v vector
		if err := json.Unmarshal(raw, &v); err != nil {
			t.Fatalf("parse %s: %v", p, err)
		}
		if want := strings.TrimSuffix(filepath.Base(p), ".json"); v.Name != want {
			t.Fatalf("%s declares name %q; file name and vector name must agree", p, v.Name)
		}
		out[v.Name] = v
	}
	return out
}

// TestCensus enforces the closed list in both directions: every named
// vector must have a file, and no file may exist outside the named set.
func TestCensus(t *testing.T) {
	got := loadVectors(t)
	for _, name := range census {
		if _, ok := got[name]; !ok {
			t.Errorf("missing vector %q (contracts/vectors.md requires it)", name)
		}
	}
	want := make(map[string]bool, len(census))
	for _, n := range census {
		want[n] = true
	}
	var extra []string
	for n := range got {
		if !want[n] {
			extra = append(extra, n)
		}
	}
	sort.Strings(extra)
	for _, n := range extra {
		t.Errorf("vector %q is outside the closed census — extras are scope creep", n)
	}
	t.Logf("census: %d vectors, matching contracts/vectors.md exactly", len(census))
}

// encodeFrame is EncodeCanonical plus the length prefix, via WriteFrame — the
// production write path this codec actually exposes.
func encodeFrame(v any) ([]byte, error) {
	body, err := EncodeCanonical(v)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := WriteFrame(&buf, body); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// TestRoundTrip is the whole obligation of seam-wire-v0.md §6: for every
// vector, read the pinned frame through this package's ReadFrame, decode
// its body, compare the result structurally against the pinned decoded
// form, re-encode canonically through WriteFrame, and compare bytes over
// the whole frame including the length word. Error vectors additionally pin
// what Validate does with them.
func TestRoundTrip(t *testing.T) {
	vectors := loadVectors(t)
	names := make([]string, 0, len(vectors))
	for n := range vectors {
		names = append(names, n)
	}
	sort.Strings(names)

	for _, name := range names {
		v := vectors[name]
		t.Run(name, func(t *testing.T) {
			want, err := Decode(v.Decoded)
			if err != nil {
				t.Fatalf("the vector's own decoded form is not usable: %v", err)
			}

			frame, err := hex.DecodeString(v.FrameHex)
			if err != nil {
				t.Fatalf("frame_hex is not lowercase unseparated hex: %v", err)
			}
			if lower := strings.ToLower(v.FrameHex); lower != v.FrameHex {
				t.Errorf("frame_hex must be lowercase (§6 fixture convention)")
			}

			// 1. The frame reads, and consumes exactly its own bytes.
			r := bytes.NewReader(frame)
			body, err := ReadFrame(r)
			if err != nil {
				t.Fatalf("ReadFrame: %v", err)
			}
			if r.Len() != 0 {
				t.Errorf("frame declares fewer bytes than the fixture carries: %d left over", r.Len())
			}

			got, err := Decode(body)
			if err != nil {
				t.Fatalf("Decode: %v", err)
			}

			// 2. Meaning: structural equality against the pinned decoded form.
			if !reflect.DeepEqual(got, want) {
				t.Errorf("decoded value differs from the vector's declared form\n got: %s\nwant: %s",
					mustCanonical(got), mustCanonical(want))
			}

			// 3. Bytes: canonical re-encode equals the pinned frame, length word included.
			reencoded, err := encodeFrame(got)
			if err != nil {
				t.Fatalf("re-encode: %v", err)
			}
			if h := hex.EncodeToString(reencoded); h != v.FrameHex {
				t.Errorf("re-encoded frame differs from the pinned bytes\n%s", hexDiff(v.FrameHex, h))
			}

			// 4. Validation behaves as the vector pins.
			verr := Validate(got)
			switch v.Expect {
			case "roundtrip", "decode_ok":
				if verr != nil {
					t.Errorf("vector must be accepted but validation refused it: %v", verr)
				}
			case "refuse":
				if verr == nil {
					t.Fatalf("vector must be refused (%s) but validation accepted it", v.Refusal)
				}
				if verr.Error() != v.Refusal {
					t.Errorf("refusal mismatch\n got: %s\nwant: %s", verr.Error(), v.Refusal)
				}
			default:
				t.Fatalf("unknown expect %q", v.Expect)
			}
		})
	}
}

// TestFramingRefusals covers the frame-layer rules the vectors cannot carry
// as data: a fixture for a malformed frame would be a fixture that is not a
// frame.
func TestFramingRefusals(t *testing.T) {
	cases := []struct {
		name  string
		frame []byte
	}{
		{"zero length", []byte{0, 0, 0, 0}},
		{"short header", []byte{0, 0, 1}},
		{"truncated body", []byte{0, 0, 0, 8, '{', '}'}},
		{"over the 1 MiB cap", []byte{0, 0x20, 0, 1}},
	}
	for _, c := range cases {
		if _, err := ReadFrame(bytes.NewReader(c.frame)); err == nil {
			t.Errorf("%s: expected refusal, got acceptance", c.name)
		}
	}

	// An empty read (stream ends exactly between frames) is an orderly
	// disconnect, not a framing error.
	if _, err := ReadFrame(bytes.NewReader(nil)); err != ErrConnectionClosed {
		t.Errorf("empty stream: got %v, want ErrConnectionClosed", err)
	}

	// Two frames back to back: the next length word begins at the byte
	// after the previous body's last byte (§2.1), with no delimiter.
	one, err := encodeFrame(map[string]any{"a": int64(1)})
	if err != nil {
		t.Fatal(err)
	}
	two, err := encodeFrame(map[string]any{"b": int64(2)})
	if err != nil {
		t.Fatal(err)
	}
	r := bytes.NewReader(append(append([]byte{}, one...), two...))

	b1, err := ReadFrame(r)
	if err != nil {
		t.Fatalf("first frame of a stream: %v", err)
	}
	b2, err := ReadFrame(r)
	if err != nil {
		t.Fatalf("second frame of a stream: %v", err)
	}
	v1, _ := Decode(b1)
	v2, _ := Decode(b2)
	if !reflect.DeepEqual(v1, map[string]any{"a": int64(1)}) || !reflect.DeepEqual(v2, map[string]any{"b": int64(2)}) {
		t.Errorf("back-to-back frames did not decode independently")
	}
}

// TestReceiverAcceptsNonCanonical is §2.4's asymmetry: senders emit
// canonical form, receivers accept any conforming JSON encoding of the same
// value. A receiver that only accepted canonical bytes would fail against
// every conforming peer.
func TestReceiverAcceptsNonCanonical(t *testing.T) {
	scruffy := []byte("{ \"b\" : 2,\n  \"a\" : 1 }")
	got, err := Decode(scruffy)
	if err != nil {
		t.Fatalf("receiver refused conforming non-canonical JSON: %v", err)
	}
	canon, err := EncodeCanonical(got)
	if err != nil {
		t.Fatal(err)
	}
	if string(canon) != `{"a":1,"b":2}` {
		t.Errorf("canonical form of the same value = %s, want {\"a\":1,\"b\":2}", canon)
	}
}

func mustCanonical(v any) []byte {
	b, err := EncodeCanonical(v)
	if err != nil {
		return []byte(fmt.Sprintf("<unencodable: %v>", err))
	}
	return b
}

// hexDiff reports the byte offset of the first difference, which is why §6
// pins hex rather than base64: two characters is one byte, so the offset is
// computable by hand from a failing diff.
func hexDiff(want, got string) string {
	n := min(len(want), len(got))
	for i := 0; i < n; i += 2 {
		if want[i:i+2] != got[i:i+2] {
			return fmt.Sprintf("first difference at byte %d: want %s, got %s\n want: …%s…\n  got: …%s…",
				i/2, want[i:i+2], got[i:i+2], window(want, i), window(got, i))
		}
	}
	return fmt.Sprintf("frames agree for %d bytes then differ in length: want %d bytes, got %d",
		n/2, len(want)/2, len(got)/2)
}

func window(s string, i int) string {
	lo, hi := max(0, i-16), min(len(s), i+16)
	return s[lo:hi]
}
