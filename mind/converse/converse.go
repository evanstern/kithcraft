// Package converse (this file): the E4 two-mind dusk exchange
// (docs/design/llm-routing-and-budget.md §1.3/§2.3/§5.2, specs/017-dusk-
// conversation Phase 1). A conversation is two Speakers alternating turns:
// each turn assembles E4's context (transcript so far, the speaker's own
// interlocutor slice, memory window) in the variable suffix, streams a
// reply from the registry's E4 config, and composes the result as a
// `speak` intent through the seam (body-protocol-v0.md §5.5). Termination
// is a closing-marker convention the speaker's own text carries on its
// final turn, not a turn cap (card AC #7) — MaxTurns is a safety bound,
// engineering hygiene rather than the ending mechanism (plan.md design
// decision 2).
package converse

import (
	"context"
	"fmt"
	"strings"
	"time"

	"kithcraft/mind/llm"
	"kithcraft/mind/prompt"
	"kithcraft/mind/seam"
)

// ClosingMarker is the convention a speaker's prompt instructs it to end
// its final turn with. The exchange loop detects it and strips it before
// the line joins the transcript or is spoken — this is the termination
// mechanism (T003); DefaultMaxTurns below is not.
const ClosingMarker = "[CONVERSATION_END]"

// DefaultMaxTurns is a generous safety bound, used when Config.MaxTurns is
// zero. It exists so a runaway exchange (a model that never emits
// ClosingMarker) still ends — card AC #7 forbids a turn cap as the
// mechanism, not a bound that exists but provably doesn't fire; T003's
// tests assert it never fires against a scripted exchange.
const DefaultMaxTurns = 40

// Interlocutor is a speaker's own slice of who it is talking to — "who
// this is, what I think of them, shared history" (§2.3's E4 row) — built
// from THIS speaker's belief store. The two sides of a conversation each
// carry their own Interlocutor; they are not shared or symmetric.
type Interlocutor struct {
	Who           string
	Impression    string
	SharedHistory string
}

func (i Interlocutor) render() string {
	return "WHO: " + i.Who + "\nIMPRESSION: " + i.Impression + "\nSHARED_HISTORY: " + i.SharedHistory
}

// Speaker is one villager's side of the exchange.
type Speaker struct {
	Name    string // body id — the transcript label and speak-intent source
	Client  *llm.Client
	Pending *seam.Pending // must declare "speak" (V-4)

	// Out, Session, and Body carry a composed speak intent onto the real
	// seam wire (a seam.Conn — FakeVendor in T004's test). Out may be nil:
	// intents are still composed into Pending (and returned to the
	// caller), just not written anywhere — a caller earlier in the
	// pipeline than a live seam connection.
	//
	// ponytail: no session_open handshake or ack/act_result handling
	// lives here — Phase 1 only proves the intent is composed and sent.
	// Wiring this into a running daemon's actual session lifecycle is a
	// later phase's job.
	Out     seam.Conn
	Session string
	Body    string

	Stable       prompt.DeliberationStablePrefix // persona/values/manifest/instructions
	Interlocutor Interlocutor
	MemoryWindow string

	seq int64
}

// assemble builds this turn's E4 call: transcript-so-far, the speaker's
// interlocutor slice, and the memory window, all in the variable suffix
// (T002) — s.Stable never varies across calls, so Assembled.Stable is
// byte-identical turn to turn (converse_test.go).
func (s *Speaker) assemble(transcript string) prompt.Assembled {
	vc := prompt.NewVariableContext().
		Add("TRANSCRIPT", transcript).
		Add("INTERLOCUTOR", s.Interlocutor.render()).
		Add("MEMORY", s.MemoryWindow)
	return prompt.Assemble(s.Stable, *vc)
}

// stream performs one E4 call and returns the accumulated reply text plus
// the wall-clock latency to its first text delta (§5.2's ceiling,
// instrumented — plan.md design decision 6).
func (s *Speaker) stream(ctx context.Context, transcript string) (text string, firstToken time.Duration, err error) {
	strm, err := s.Client.Stream(ctx, llm.E4, s.assemble(transcript))
	if err != nil {
		return "", 0, err
	}
	defer strm.Close()

	start := time.Now()
	var sb strings.Builder
	for strm.Next() {
		ev := strm.Current()
		if ev.Type == "content_block_delta" && ev.Delta.Text != "" {
			if firstToken == 0 {
				firstToken = time.Since(start)
			}
			sb.WriteString(ev.Delta.Text)
		}
	}
	if err := strm.Err(); err != nil {
		return "", 0, err
	}
	return sb.String(), firstToken, nil
}

// speak composes text as a `speak` intent (target type "none" — a dusk
// exchange is spoken into earshot of whoever is at the gathering place,
// body-protocol-v0.md §6.4's "speak — append the utterance to the room";
// there is no separate content field on the intent schema for a speak's
// text, so it rides inside target alongside type, the only place the
// schema leaves for it) and, if s.Out is set, writes it onto the seam wire.
func (s *Speaker) speak(worldTime int64, text string) error {
	target := map[string]any{"type": "none", "text": text}
	payload, err := s.Pending.Compose(fmt.Sprintf("i-%s-%d", s.Name, s.seq), "speak", target, "dusk conversation turn", "")
	if err != nil {
		return err
	}
	if s.Out == nil {
		return nil
	}
	s.seq++
	return s.Out.WriteMessage(map[string]any{
		"protocol": "0.1", "message": "intent", "session": s.Session,
		"seq": s.seq, "body": s.Body, "world_time": worldTime,
		"payload": payload,
	})
}

// Turn is one spoken line, with the latency its call measured.
type Turn struct {
	Speaker           string
	Text              string
	FirstTokenLatency time.Duration
}

// Config is Exchange's tuning: WorldTime rides on every composed intent
// (mind-side arithmetic only, never a clock — §6.2 of the routing doc);
// MaxTurns overrides DefaultMaxTurns when nonzero.
type Config struct {
	WorldTime int64
	MaxTurns  int

	// OpeningSlot, if set, is consulted for a's opening turn (T005/US2):
	// filled earlier off V3's pair-formation signal (pregen.go), it lets
	// the first turn serve without a new E4 call (card AC #4). If the fill
	// isn't ready within OpeningWait (zero: check now, don't wait), the
	// opening turn falls back to a's normal live stream call (T006) — the
	// ceiling still holds because the fallback is just an ordinary E4
	// call, same as any other turn.
	OpeningSlot *Slot
	OpeningWait time.Duration
}

// Exchange runs a dusk conversation between a and b, alternating turns
// starting with a. Each turn assembles context (T002), streams an E4 reply
// (T001), composes and sends the resulting speak intent, and extends the
// shared transcript both speakers see on their next turn. The loop ends
// when a turn's text carries ClosingMarker (T003) or the safety bound
// (Config.MaxTurns, else DefaultMaxTurns) is reached — the latter is
// hygiene, not the mechanism; T003's tests prove it does not fire against
// a scripted exchange.
func Exchange(ctx context.Context, a, b *Speaker, cfg Config) ([]Turn, error) {
	max := cfg.MaxTurns
	if max <= 0 {
		max = DefaultMaxTurns
	}
	speakers := [2]*Speaker{a, b}
	var transcript strings.Builder
	var turns []Turn
	for i := 0; i < max; i++ {
		s := speakers[i%2]
		var text string
		var latency time.Duration
		var err error
		var served bool
		if i == 0 && cfg.OpeningSlot != nil {
			text, latency, served = cfg.OpeningSlot.Take(cfg.OpeningWait)
		}
		if !served {
			text, latency, err = s.stream(ctx, transcript.String())
		}
		if err != nil {
			return turns, fmt.Errorf("converse: %s's turn: %w", s.Name, err)
		}
		done := strings.Contains(text, ClosingMarker)
		text = strings.TrimSpace(strings.Replace(text, ClosingMarker, "", 1))
		if err := s.speak(cfg.WorldTime, text); err != nil {
			return turns, fmt.Errorf("converse: %s: speak intent: %w", s.Name, err)
		}
		turns = append(turns, Turn{Speaker: s.Name, Text: text, FirstTokenLatency: latency})
		if transcript.Len() > 0 {
			transcript.WriteString("\n")
		}
		transcript.WriteString(s.Name + ": " + text)
		if done {
			break
		}
	}
	return turns, nil
}
