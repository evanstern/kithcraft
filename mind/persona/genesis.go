// Package persona (this file): E1 persona genesis (FR-002, decision-0003
// §8) — the only file in this package that imports mind/llm (package doc
// comment's structural guarantee for validate.go's llm-free half). One E1
// call per cast entry, Opus 5 per llm.Registry[E1], producing a persona's
// name, values, endogenous desires, anchor line, and drift markers as
// GENERATED content (brief #5: minds are generated, not authored) — the
// entry's profession/biome pairing is carried through untouched, never
// model-invented (decision-0002, card AC #5).
//
// classes.go's Registry sets E1's StructuredOutput=false — client.go's
// buildParams only attaches output_config.format when a class both has
// StructuredOutput=true and SchemaFor names it, and SchemaFor only names
// E2/E3/E6 (structured.go). Structured output genuinely fights E1's config,
// so this file falls back to JSON-in-text: the prompt demands a single JSON
// object and parseGenesisOutput strict-decodes it, following structured.go's
// idiom (a typed shape, no regex/best-effort scraping, a bounded failure via
// the same *llm.ParseError E2/E3/E6 already use). Recorded per plan.md's
// risk note in specs/013-persona-genesis/tasks.md (Phase 3, T007).
package persona

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"

	"kithcraft/mind/llm"
	"kithcraft/mind/prompt"
)

// CastEntry is one genesis request's input (decision-0002): a vendor-mod
// cast id paired with its profession and biome variant. Genesis carries
// Profession and BiomeVariant straight into the resulting Persona — they
// are decided facts about the villager's body and place, never part of
// what E1 is asked to invent.
type CastEntry struct {
	CastID       string
	Profession   string
	BiomeVariant string
}

// genesisOutput is E1's JSON-in-text response shape: exactly the fields
// brief #5 asks a generated persona to carry, nothing decision-0002
// already supplies.
type genesisOutput struct {
	Name              string   `json:"name"`
	Values            []string `json:"values"`
	EndogenousDesires []string `json:"endogenous_desires"`
	Anchor            string   `json:"anchor"`
	DriftMarkers      []string `json:"drift_markers"`
}

// Genesis runs one E1 call per entry (FR-002: three calls for the demo
// cast; no retries beyond the client's own DefaultMaxRetries — a rejected
// persona surfaces as an error, it does not silently regenerate) and writes
// each accepted persona via WriteOnce (FR-003: genesis against an existing
// cast id refuses, the same guarded path WriteOnce already proves — Genesis
// opens no write path of its own). A generated persona is run through
// Validate (validate.go) against its OWN generated text before it is
// written, so a moralizing trait stated at birth (US4) is rejected the same
// way later drift is rejected, before a byte reaches disk.
func Genesis(ctx context.Context, client *llm.Client, dir string, entries []CastEntry) ([]Persona, error) {
	out := make([]Persona, 0, len(entries))
	for _, e := range entries {
		p, err := genesisOne(ctx, client, e)
		if err != nil {
			return nil, fmt.Errorf("persona: Genesis: cast id %q: %w", e.CastID, err)
		}
		if err := WriteOnce(dir, p); err != nil {
			return nil, fmt.Errorf("persona: Genesis: cast id %q: %w", e.CastID, err)
		}
		out = append(out, p)
	}
	return out, nil
}

func genesisOne(ctx context.Context, client *llm.Client, e CastEntry) (Persona, error) {
	assembled := prompt.Assemble(nil, genesisVariableContext(e))
	msg, err := client.Send(ctx, llm.E1, assembled)
	if err != nil {
		return Persona{}, fmt.Errorf("E1 call: %w", err)
	}

	out, err := parseGenesisOutput(msg)
	if err != nil {
		return Persona{}, err
	}

	p := Persona{
		CastID:            e.CastID,
		Name:              out.Name,
		Values:            out.Values,
		EndogenousDesires: out.EndogenousDesires,
		Anchor:            out.Anchor,
		DriftMarkers:      out.DriftMarkers,
		Profession:        e.Profession,
		BiomeVariant:      e.BiomeVariant,
	}

	if err := Validate(p, selfNarrative(p)); err != nil {
		return Persona{}, fmt.Errorf("generated persona rejected at birth: %w", err)
	}
	return p, nil
}

// selfNarrative is the candidate a freshly generated persona is Validated
// against: its own anchor (satisfying anchor echo trivially — a persona
// always restates itself) plus its own values and endogenous desires, so a
// moralizing word stated anywhere in the generated content — not only the
// anchor — trips the cast-wide lexicon (US4, moralizing.go) at birth.
func selfNarrative(p Persona) string {
	parts := append([]string{p.Anchor}, p.Values...)
	parts = append(parts, p.EndogenousDesires...)
	return strings.Join(parts, " ")
}

// parseGenesisOutput strict-decodes E1's JSON-in-text response. A malformed
// payload or a missing name/anchor is a bounded *llm.ParseError (structured.
// go's idiom), never a panic or a Persona a caller could mistake for real.
func parseGenesisOutput(msg *anthropic.Message) (genesisOutput, error) {
	var raw strings.Builder
	for _, block := range msg.Content {
		if block.Type == "text" {
			raw.WriteString(block.Text)
		}
	}
	text := raw.String()

	var out genesisOutput
	if err := json.Unmarshal([]byte(text), &out); err != nil {
		return genesisOutput{}, &llm.ParseError{Class: llm.E1, Raw: text, Err: err}
	}
	if out.Name == "" || out.Anchor == "" {
		return genesisOutput{}, &llm.ParseError{Class: llm.E1, Raw: text, Err: fmt.Errorf("missing required field(s): name and anchor must be non-empty")}
	}
	return out, nil
}

// worldPremise is E1's shared, unvarying world-context section (§2.3's
// "world premise" column) — genesis text is this package's to own
// (plan.md: "M3 owns persona genesis text"), not a caller-supplied string a
// consumer that doesn't exist yet would have to invent.
const worldPremise = `Kithcraft is a Minecraft-shaped survival-crafting world where the villagers are generated minds, not scripted NPCs. The whole point of the project is that they are people the player did not write, and their company only counts as company because who they are is not up for revision by the player, the model, or anyone else. Villagers have their own schedules, wants, and relationships already in motion; the player's orders land on top of a life, not into a blank slot.`

// weirdnessDialInstruction pins the dial conservative (brief #5): "Weirdness
// dial starts conservative." Genesis exists to produce distinct, grounded
// people, not surrealism.
const weirdnessDialInstruction = `Weirdness dial: conservative. Write a distinct, grounded, believable person shaped by an ordinary life — not a surreal, whimsical, or joke character. Distinctiveness comes from specific values and wants, not from strangeness.`

// antiMoralizingInstruction is US4's prompt half: no character whose nature
// is correcting/scolding/policing the player's manners (the brief's named
// spell-breaker, "this is not a politeness simulator").
const antiMoralizingInstruction = `Do not generate a character whose nature is correcting, scolding, lecturing, or policing the player's manners. This is not a politeness simulator: a villager may have its own wants, reluctance, and grumbling, but it never takes offense at how the player behaves and never exists to teach the player manners.`

// outputInstruction demands the JSON-in-text shape parseGenesisOutput
// decodes, and explains drift_markers in the prompt itself, per T008: they
// are trait words that would CONTRADICT this character, the opposite of who
// they are — not a description of them.
const outputInstruction = `Invent this villager's name and inner life, then respond with a single JSON object and nothing else — no prose before or after, no markdown fences. The object must have exactly these fields:

{
  "name": string — the villager's given name,
  "values": [string, ...] — what this person cares about,
  "endogenous_desires": [string, ...] — what this person wants, independent of any player request,
  "anchor": string — one sentence this character would say about themselves; it is the fixed line their identity is measured against for the rest of their life,
  "drift_markers": [string, ...] — single trait words that would CONTRADICT this character (for example, a patient, careful farmer's drift markers might include "reckless" and "careless") — these are the opposite of who this person is, not a description of them
}`

// genesisVariableContext composes E1's whole call (§2.3: E1 has no stable
// prefix, so everything — world premise, cast pairing, dial, output demand —
// is variable context).
func genesisVariableContext(e CastEntry) prompt.VariableContext {
	v := prompt.NewVariableContext()
	v.Add("WORLD_PREMISE", worldPremise)
	v.Add("CAST_PAIRING", fmt.Sprintf("This villager's profession is %q and biome variant is %q — a fixed fact about their body and place in the world. Invent the person shaped by that life; do not invent a different job or place.", e.Profession, e.BiomeVariant))
	v.Add("WEIRDNESS_DIAL", weirdnessDialInstruction)
	v.Add("ANTI_MORALIZING", antiMoralizingInstruction)
	v.Add("OUTPUT", outputInstruction)
	return *v
}
