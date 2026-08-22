package roundtrip

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

var update = flag.Bool("update", false,
	"rewrite each vector's frame_hex from its decoded form (authoring only — never in CI). "+
		"The independent check is the Java harness, which only ever verifies.")

const vectorDir = "../vectors"

// A vector as it sits on disk. See ../vectors/README.md for the format.
type vector struct {
	Name      string          `json:"name"`
	Direction string          `json:"direction"`
	Expect    string          `json:"expect"`
	Refusal   string          `json:"refusal"`
	Decoded   json.RawMessage `json:"decoded"`
	FrameHex  string          `json:"frame_hex"`
}

// census is contracts/vectors.md's closed list. The set is complete when every name
// here has a file and no file has a name outside it — an extra vector is scope creep,
// and a missing one is a hole in the pinning.
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

// TestCensus enforces the closed list.
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
	extra := []string{}
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

// TestRoundTrip is the whole obligation: for every vector, decode the pinned frame,
// compare its meaning structurally against the pinned decoded form, re-encode
// canonically, and compare bytes over the WHOLE frame including the length word.
// Error vectors additionally pin what validation does with them.
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
			want, err := Normalize(v.Decoded)
			if err != nil {
				t.Fatalf("the vector's own decoded form is not usable: %v", err)
			}

			if *update {
				frame, err := EncodeFrame(want)
				if err != nil {
					t.Fatalf("encode: %v", err)
				}
				rewriteFrameHex(t, filepath.Join(vectorDir, name+".json"), hex.EncodeToString(frame))
				return
			}

			frame, err := hex.DecodeString(v.FrameHex)
			if err != nil {
				t.Fatalf("frame_hex is not lowercase unseparated hex: %v", err)
			}
			if lower := strings.ToLower(v.FrameHex); lower != v.FrameHex {
				t.Errorf("frame_hex must be lowercase (§6 fixture convention)")
			}

			// 1. The frame decodes, and consumes exactly its own bytes.
			got, n, err := DecodeFrame(frame)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if n != len(frame) {
				t.Errorf("frame declares %d bytes but the fixture carries %d", n, len(frame))
			}

			// 2. Meaning: structural equality against the pinned decoded form.
			if !reflect.DeepEqual(got, want) {
				t.Errorf("decoded value differs from the vector's declared form\n got: %s\nwant: %s",
					mustCanonical(got), mustCanonical(want))
			}

			// 3. Bytes: canonical re-encode equals the pinned frame, length word included.
			reencoded, err := EncodeFrame(got)
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

// TestFramingRefusals covers the frame-layer rules the vectors cannot carry as data:
// a fixture for a malformed frame would be a fixture that is not a frame.
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
		if _, _, err := DecodeFrame(c.frame); err == nil {
			t.Errorf("%s: expected refusal, got acceptance", c.name)
		}
	}

	// Two frames back to back: the next length word begins at the byte after the
	// previous body's last byte (§2.1), with no delimiter between them.
	one, err := EncodeFrame(map[string]any{"a": int64(1)})
	if err != nil {
		t.Fatal(err)
	}
	two, err := EncodeFrame(map[string]any{"b": int64(2)})
	if err != nil {
		t.Fatal(err)
	}
	stream := append(append([]byte{}, one...), two...)
	v1, n1, err := DecodeFrame(stream)
	if err != nil || n1 != len(one) {
		t.Fatalf("first frame of a stream: %v (consumed %d, want %d)", err, n1, len(one))
	}
	v2, _, err := DecodeFrame(stream[n1:])
	if err != nil {
		t.Fatalf("second frame of a stream: %v", err)
	}
	if !reflect.DeepEqual(v1, map[string]any{"a": int64(1)}) || !reflect.DeepEqual(v2, map[string]any{"b": int64(2)}) {
		t.Errorf("back-to-back frames did not decode independently")
	}
}

// TestReceiverAcceptsNonCanonical is §2.4's asymmetry: senders emit canonical form,
// receivers accept any conforming JSON encoding of the same value. A receiver that
// only accepted canonical bytes would fail against every conforming peer.
func TestReceiverAcceptsNonCanonical(t *testing.T) {
	scruffy := []byte("{ \"b\" : 2,\n  \"a\" : 1 }")
	got, err := Normalize(scruffy)
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

// hexDiff reports the byte offset of the first difference, which is why §6 pins hex
// rather than base64: two characters is one byte, so the offset is computable by hand.
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

// rewriteFrameHex replaces the frame_hex line in place, leaving the rest of the file
// byte-identical — the decoded form is hand-authored and must never be reformatted by
// a tool.
func rewriteFrameHex(t *testing.T, path, h string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	lines := strings.Split(string(raw), "\n")
	found := false
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), `"frame_hex"`) {
			indent := line[:len(line)-len(strings.TrimLeft(line, " "))]
			lines[i] = fmt.Sprintf("%s\"frame_hex\": \"%s\"", indent, h)
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("%s has no frame_hex line", path)
	}
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	t.Logf("updated frame_hex (%d frame bytes)", len(h)/2)
}
