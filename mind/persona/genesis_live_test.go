//go:build live

// Live genesis proof (T010/T011, specs/013-persona-genesis/tasks.md Phase 4).
// Build-tag gated so `go test ./...` never touches the network or spends a
// real E1 call (M4 idiom) — the daemon's normal build/test loop is
// untouched. Run explicitly, with ANTHROPIC_API_KEY/ANTHROPIC_BASE_URL
// exported into the process env from the repo root's .env:
//
//	go test -tags=live -run TestLiveGenesis -v ./...
//	go test -tags=live -run TestLiveRestart -v ./...
//
// personaRunDir (mind/run/persona/, gitignored — mind/.gitignore) is the
// daemon's runtime data dir for generated personas. No such convention
// existed before this task; this file is where it's established, since
// minddaemon/main.go does not yet wire persona loading (that's a later
// milestone's job — this task only proves genesis and re-bind work for real).
package persona

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"

	"kithcraft/mind/llm"
	"kithcraft/mind/prompt"
)

// rejectedAtBirth reports whether err is liveGenesisOne's validator-rejection
// wrap (validate.go's RejectionReason) as opposed to a transport/auth/parse
// failure. Only a validator rejection gets the one sanctioned retry — an
// auth/billing-shaped failure must STOP, never retry into a bill (dispatch
// rule), so the two error shapes must not be treated alike.
func rejectedAtBirth(err error) bool {
	var reason *RejectionReason
	return errors.As(err, &reason)
}

func personaRunDir() string { return filepath.Join("..", "run", "persona") }

// liveGenesisOne mirrors genesisOne (genesis.go) but also returns the raw
// API message, so the live-run record (specs/013-persona-genesis/live-run.md)
// can cite the model and token counts the API actually reported rather than
// only the model requested. genesis.go itself is unchanged — this exists
// only for this live-run proof's logging.
func liveGenesisOne(ctx context.Context, client *llm.Client, e CastEntry) (Persona, *anthropic.Message, error) {
	assembled := prompt.Assemble(nil, genesisVariableContext(e))
	msg, err := client.Send(ctx, llm.E1, assembled)
	if err != nil {
		return Persona{}, nil, fmt.Errorf("E1 call: %w", err)
	}
	out, err := parseGenesisOutput(msg)
	if err != nil {
		return Persona{}, msg, err
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
		return Persona{}, msg, fmt.Errorf("generated persona rejected at birth: %w", err)
	}
	return p, msg, nil
}

// TestLiveGenesis_ThreeCastEntries is T010: three real E1 calls, one per
// demo cast entry. A validator rejection gets exactly ONE regeneration
// attempt (sanctioned per the dispatch — recorded via t.Logf); a second
// rejection stops the test rather than looping into more spend.
func TestLiveGenesis_ThreeCastEntries(t *testing.T) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set — live genesis is opt-in, see file doc comment")
	}
	dir := personaRunDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll %s: %v", dir, err)
	}
	client := llm.New()
	ctx := context.Background()

	for _, e := range demoCastEntries() {
		p, msg, err := liveGenesisOne(ctx, client, e)
		if err != nil {
			if !rejectedAtBirth(err) {
				t.Fatalf("cast %s: E1 call failed (not a validator rejection — no retry, per the auth/billing STOP rule): %v", e.CastID, err)
			}
			t.Logf("cast %s: attempt 1 rejected by the validator: %v — one sanctioned retry (spec.md FR-002)", e.CastID, err)
			p, msg, err = liveGenesisOne(ctx, client, e)
			if err != nil {
				t.Fatalf("cast %s: retry also rejected, stopping (no bill-multiplying loop): %v", e.CastID, err)
			}
		}
		if err := WriteOnce(dir, p); err != nil {
			t.Fatalf("cast %s: WriteOnce: %v", e.CastID, err)
		}
		t.Logf("cast %s: model=%s name=%q input_tokens=%d output_tokens=%d",
			e.CastID, msg.Model, p.Name, msg.Usage.InputTokens, msg.Usage.OutputTokens)
	}
}

// TestLiveRestart_LoadsAndBindsRealFiles is T011: the process-restart
// simulation over the REAL files T010 wrote — a fresh Load call, no
// in-memory state carried from genesis, binds each of the three real
// persona files to its cast id (FR-004).
func TestLiveRestart_LoadsAndBindsRealFiles(t *testing.T) {
	dir := personaRunDir()
	ids := []string{"Aldric", "Petra", "Yenna"}
	loaded, err := Load(dir, ids)
	if err != nil {
		t.Skipf("no live persona files at %s — run TestLiveGenesis_ThreeCastEntries first: %v", dir, err)
	}
	for _, id := range ids {
		p, ok := loaded[id]
		if !ok || p.CastID != id {
			t.Errorf("restart re-bind: cast id %q did not bind correctly", id)
		} else {
			t.Logf("cast %s: re-bound to %q on reload", id, p.Name)
		}
	}
}
