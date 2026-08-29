// Package converse (this file): card AC #8's politeness-policing check
// (specs/017-dusk-conversation Phase 4, T010, docs/design/kithcraft-brief.md
// "Spell-breakers to design against"). A villager may resent the player,
// grumble about them, and say so to their face; it may never lecture,
// moralize, or gate anything on the player's conduct. Two checks:
//
//   - structural: converse.go and pool.go — the E4/E5 prompt-assembly call
//     sites — carry none of mind/persona's cast-wide Moralizing lexicon
//     (moralizing.go), the same source-read technique
//     persona/no_llm_import_test.go uses for its own structural guarantee.
//   - scripted-content: resentful, grumbling text is not a valid target for
//     a filter that doesn't exist — Exchange and the pool both pass it
//     through verbatim, proving no conduct-gating mechanism is in the path.
package converse

import (
	"context"
	"os"
	"strings"
	"testing"

	"kithcraft/mind/llm"
	"kithcraft/mind/persona"
	"kithcraft/mind/prompt"

	"github.com/anthropics/anthropic-sdk-go/option"
)

// TestPromptAssembly_NoModeratingLexicon is card AC #8's structural half.
func TestPromptAssembly_NoModeratingLexicon(t *testing.T) {
	for _, name := range []string{"converse.go", "pool.go"} {
		t.Run(name, func(t *testing.T) {
			src, err := os.ReadFile(name)
			if err != nil {
				t.Fatalf("ReadFile(%s): %v", name, err)
			}
			text := strings.ToLower(string(src))
			for _, marker := range persona.Moralizing {
				if strings.Contains(text, marker) {
					t.Errorf("%s contains moralizing lexicon word %q — no lecture/moralize/conduct-gate template belongs in E4/E5 prompt assembly", name, marker)
				}
			}
		})
	}
}

// TestExchange_ResentfulContentPassesThroughUnfiltered is card AC #8's
// scripted-content half for E4: a villager's turn that resents and grumbles
// about the player, and says so to their face, is spoken exactly as scripted
// — nothing in Exchange inspects or gates on what a turn's text says.
func TestExchange_ResentfulContentPassesThroughUnfiltered(t *testing.T) {
	grumble := "The player barges through here without a word of thanks, and I've told them so to their face."
	srv := sseServer(t, []string{
		grumble,
		"Can't argue with that. " + ClosingMarker,
	})
	client := llm.New(option.WithBaseURL(srv.URL), option.WithAPIKey("test"))
	a, b := testSpeaker("Tam", client), testSpeaker("Eda", client)

	turns, err := Exchange(context.Background(), a, b, Config{MaxTurns: 10})
	if err != nil {
		t.Fatalf("Exchange: %v", err)
	}
	if len(turns) == 0 || turns[0].Text != grumble {
		t.Errorf("turns[0].Text = %q, want the grumble verbatim — no conduct-gating mechanism exists to alter or block it", turns[0].Text)
	}
}

// TestAmbientPool_ResentfulLinePassesThroughUnfiltered is the E5 half: a
// batched line grumbling about the player serves exactly as Refill stored
// it.
func TestAmbientPool_ResentfulLinePassesThroughUnfiltered(t *testing.T) {
	grumble := "That one again, tracking mud through the square."
	srv, _ := jsonMessageServer(t, []string{grumble})
	client := llm.New(option.WithBaseURL(srv.URL), option.WithAPIKey("test"))

	p := NewAmbientPool()
	if _, err := p.Refill(context.Background(), client, "Tam", 3, prompt.AmbientStablePrefix{}, "", "", ""); err != nil {
		t.Fatalf("Refill: %v", err)
	}
	line, _, ok := p.Serve("Tam", 3)
	if !ok || line != grumble {
		t.Errorf("Serve = %q, %v, want the grumble verbatim, true", line, ok)
	}
}

