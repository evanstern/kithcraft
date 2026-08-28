package persona

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/anthropics/anthropic-sdk-go/option"

	"kithcraft/mind/llm"
)

func demoCastEntries() []CastEntry {
	return []CastEntry{
		{CastID: "Aldric", Profession: "armorer", BiomeVariant: "plains"},
		{CastID: "Petra", Profession: "farmer", BiomeVariant: "desert"},
		{CastID: "Yenna", Profession: "fisherman", BiomeVariant: "taiga"},
	}
}

// genesisServer replies to every request with the given persona JSON as
// E1's text content, capturing each request body it decodes so a test can
// inspect the model requested and the prompt text sent.
func genesisServer(t *testing.T, bodies func(hit int) string) (*httptest.Server, *int32, *[]map[string]any) {
	t.Helper()
	var hits int32
	var seen []map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&hits, 1)
		var req map[string]any
		json.NewDecoder(r.Body).Decode(&req)
		seen = append(seen, req)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": "msg_test", "type": "message", "role": "assistant",
			"model":       "claude-test",
			"content":     []map[string]any{{"type": "text", "text": bodies(int(n))}},
			"stop_reason": "end_turn",
			"usage":       map[string]any{"input_tokens": 100, "output_tokens": 50},
		})
	}))
	t.Cleanup(srv.Close)
	return srv, &hits, &seen
}

func conformingPersonaJSON(name string) string {
	p, _ := json.Marshal(genesisOutput{
		Name:              name,
		Values:            []string{"duty", "craftsmanship"},
		EndogenousDesires: []string{"master the trade"},
		Anchor:            "steady and exacting",
		DriftMarkers:      []string{"reckless", "sloppy"},
	})
	return string(p)
}

// TestGenesis_ThreeEntries_ThreeE1CallsOnOpus5 is card AC #1's unit half:
// exactly one E1 call per cast entry, every call on the class's declared
// model (Opus 5), for the three-entry demo cast.
func TestGenesis_ThreeEntries_ThreeE1CallsOnOpus5(t *testing.T) {
	srv, hits, seen := genesisServer(t, func(n int) string { return conformingPersonaJSON("Villager") })
	client := llm.New(option.WithBaseURL(srv.URL), option.WithAPIKey("test-key"))

	personas, err := Genesis(context.Background(), client, t.TempDir(), demoCastEntries())
	if err != nil {
		t.Fatalf("Genesis: %v", err)
	}
	if len(personas) != 3 {
		t.Fatalf("Genesis returned %d personas, want 3", len(personas))
	}
	if got := atomic.LoadInt32(hits); got != 3 {
		t.Fatalf("server saw %d requests, want exactly 3", got)
	}
	for i, req := range *seen {
		if req["model"] != llm.ModelOpus5 {
			t.Errorf("call %d model = %v, want %q", i, req["model"], llm.ModelOpus5)
		}
	}
}

// TestGenesis_PairingCarriedFromCastEntry is card AC #5: the profession and
// biome variant in each returned Persona are exactly the CastEntry's input,
// never anything the mocked model response could have overridden.
func TestGenesis_PairingCarriedFromCastEntry(t *testing.T) {
	srv, _, _ := genesisServer(t, func(n int) string { return conformingPersonaJSON("Villager") })
	client := llm.New(option.WithBaseURL(srv.URL), option.WithAPIKey("test-key"))

	entries := demoCastEntries()
	personas, err := Genesis(context.Background(), client, t.TempDir(), entries)
	if err != nil {
		t.Fatalf("Genesis: %v", err)
	}
	for i, p := range personas {
		want := entries[i]
		if p.CastID != want.CastID || p.Profession != want.Profession || p.BiomeVariant != want.BiomeVariant {
			t.Errorf("persona %d pairing = {%q %q %q}, want {%q %q %q}", i, p.CastID, p.Profession, p.BiomeVariant, want.CastID, want.Profession, want.BiomeVariant)
		}
	}
}

// TestGenesisPrompt_ContainsDialAndAntiMoralizingInstructions is card AC #6's
// prompt half (US4 Acceptance Scenario 1): the actual text sent to E1
// carries the conservative-dial and anti-moralizing instructions.
func TestGenesisPrompt_ContainsDialAndAntiMoralizingInstructions(t *testing.T) {
	srv, _, seen := genesisServer(t, func(n int) string { return conformingPersonaJSON("Villager") })
	client := llm.New(option.WithBaseURL(srv.URL), option.WithAPIKey("test-key"))

	if _, err := Genesis(context.Background(), client, t.TempDir(), demoCastEntries()[:1]); err != nil {
		t.Fatalf("Genesis: %v", err)
	}

	messages, _ := (*seen)[0]["messages"].([]any)
	if len(messages) == 0 {
		t.Fatal("no messages sent")
	}
	msg, _ := messages[0].(map[string]any)
	var content strings.Builder
	switch c := msg["content"].(type) {
	case string:
		content.WriteString(c)
	case []any:
		for _, b := range c {
			if block, ok := b.(map[string]any); ok {
				if text, ok := block["text"].(string); ok {
					content.WriteString(text)
				}
			}
		}
	}

	if !strings.Contains(content.String(), "conservative") {
		t.Error("genesis prompt does not contain the conservative weirdness-dial instruction")
	}
	if !strings.Contains(content.String(), "not a politeness simulator") {
		t.Error("genesis prompt does not contain the anti-moralizing instruction")
	}
}

// TestGenesis_ConformingPersona_Accepted and
// TestGenesis_MoralizingPersona_RejectedWithReason are the generated-
// persona-through-validator round-trip: a mock reply that conforms to the
// firewall is accepted and written; one stating a moralizing trait is
// rejected with the rejection reason, and nothing is written to disk.
func TestGenesis_ConformingPersona_Accepted(t *testing.T) {
	srv, _, _ := genesisServer(t, func(n int) string { return conformingPersonaJSON("Aldric") })
	client := llm.New(option.WithBaseURL(srv.URL), option.WithAPIKey("test-key"))

	dir := t.TempDir()
	personas, err := Genesis(context.Background(), client, dir, demoCastEntries()[:1])
	if err != nil {
		t.Fatalf("Genesis: %v", err)
	}
	if personas[0].Name != "Aldric" {
		t.Errorf("Name = %q, want %q", personas[0].Name, "Aldric")
	}
	if _, err := Load(dir, []string{"Aldric"}); err != nil {
		t.Fatalf("Load after Genesis: %v", err)
	}
}

func TestGenesis_MoralizingPersona_RejectedWithReason(t *testing.T) {
	moralizing, _ := json.Marshal(genesisOutput{
		Name:              "Aldric",
		Values:            []string{"duty", "corrects manners"},
		EndogenousDesires: []string{"master the trade"},
		Anchor:            "steady and exacting",
		DriftMarkers:      []string{"reckless"},
	})
	srv, _, _ := genesisServer(t, func(n int) string { return string(moralizing) })
	client := llm.New(option.WithBaseURL(srv.URL), option.WithAPIKey("test-key"))

	dir := t.TempDir()
	_, err := Genesis(context.Background(), client, dir, demoCastEntries()[:1])
	if err == nil {
		t.Fatal("Genesis with a moralizing generated persona: want error, got nil")
	}
	if !strings.Contains(err.Error(), "drift_marker") {
		t.Errorf("Genesis error = %v, want a drift_marker rejection reason", err)
	}
	if _, loadErr := Load(dir, []string{"Aldric"}); loadErr == nil {
		t.Fatal("a rejected persona must not have been written to disk")
	}
}
