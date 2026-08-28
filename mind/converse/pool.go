// Package converse (this file): the E5 ambient line pool and specific-
// remark escalation (specs/017-dusk-conversation Phase 3, US3,
// docs/design/llm-routing-and-budget.md §1.3/§2.3/§5.3). AmbientPool holds
// one villager's daily batch of ~8 persona-flavoured lines from a single
// Haiku 4.5 call (card AC #5); Serve draws from it with a map lookup, so
// the <200 ms budget is trivially met (measured anyway, per FR-003
// discipline). Named AmbientPool, not Pool, because pregen.go's Pool
// already owns that name for a different thing (Phase 2's naming note).
//
// A remark about something specific escalates to Escalate's live call
// instead (card AC #6) — the reflex/planner split applied one level down
// within E5 itself (§5.3). When the pool has nothing to serve, the caller
// reaches for Stall (card AC #7's stall clause): never a model call on the
// <200 ms path.
package converse

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/anthropics/anthropic-sdk-go"

	"kithcraft/mind/llm"
	"kithcraft/mind/prompt"
)

// AmbientPool is the daemon-facing store of daily ambient lines, keyed by
// villager. Safe for concurrent use.
type AmbientPool struct {
	mu   sync.Mutex
	day  map[string]int64
	left map[string][]string
}

func NewAmbientPool() *AmbientPool {
	return &AmbientPool{day: make(map[string]int64), left: make(map[string][]string)}
}

// Refill runs E5's one batched call for villager on day — the call the
// daemon's day-rollover wiring makes (ingest wiring is out of scope here,
// same posture as pregen.go's Pool.Begin) — and replaces whatever lines
// the pool held. Response text is split one line per persona-flavoured
// remark (§2.3: ~8 lines out); blank lines are dropped.
func (p *AmbientPool) Refill(ctx context.Context, client *llm.Client, villager string, day int64, stable prompt.AmbientStablePrefix, mood, whoAround, recentNotable string) ([]string, error) {
	vc := prompt.NewVariableContext().
		Add("MOOD", mood).
		Add("WHO_AROUND", whoAround).
		Add("RECENT_NOTABLE", recentNotable)
	msg, err := client.Send(ctx, llm.E5, prompt.Assemble(stable, *vc))
	if err != nil {
		return nil, err
	}
	lines := splitLines(messageText(msg))
	p.mu.Lock()
	p.day[villager] = day
	p.left[villager] = lines
	p.mu.Unlock()
	return lines, nil
}

// Serve draws one unserved line for villager on day, measuring service
// time (§5.3, card AC #5). A line served this cycle is removed so it is
// never served again. ok is false — the pool-empty case — when villager
// has never been refilled, when day doesn't match the pool's last Refill
// (a rollover happened and the daemon's refresh hasn't run yet: "yesterday's
// lines are gone" even if unspent, not served stale), or when every line
// this cycle is already spent. The caller's cue on false is Stall, never a
// retry against the model on this path.
func (p *AmbientPool) Serve(villager string, day int64) (line string, latency time.Duration, ok bool) {
	start := time.Now()
	p.mu.Lock()
	defer p.mu.Unlock()
	remaining := p.left[villager]
	if p.day[villager] != day || len(remaining) == 0 {
		return "", time.Since(start), false
	}
	line, p.left[villager] = remaining[0], remaining[1:]
	return line, time.Since(start), true
}

// IsTargeted reports whether an ambient-remark trigger is "about something
// specific that just happened" (§5.3) rather than a passing greeting — the
// predicate that routes to Escalate instead of Serve (card AC #6). A
// trigger carrying a subject is the targeted signal; a passing greeting
// has none.
func IsTargeted(subject string) bool {
	return strings.TrimSpace(subject) != ""
}

// Escalate performs one live E5 call (Haiku 4.5, per the registry — same
// tier as the pool, just made live instead of batched) for a targeted
// remark, bypassing the pool entirely (card AC #6). Unlike Refill this
// produces one line, not a batch: subject/detail describe the specific
// thing the remark is about.
func Escalate(ctx context.Context, client *llm.Client, stable prompt.AmbientStablePrefix, subject, detail string) (string, error) {
	vc := prompt.NewVariableContext().
		Add("SUBJECT", subject).
		Add("DETAIL", detail)
	msg, err := client.Send(ctx, llm.E5, prompt.Assemble(stable, *vc))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(messageText(msg)), nil
}

// StallLines is the sparing stall-line policy for the pool-empty edge case
// (spec.md "Edge Cases": "stall-line policy + retry next natural trigger;
// never a blocking call on the < 200 ms path"). Short and non-committal —
// never meant as a prefix tic prepended to real speech, only a substitute
// for a turn where Serve had nothing.
var StallLines = []string{"Hm.", "...", "One moment.", "Mm."}

// Stall picks a stall line deterministically from villager and day —
// repeatable in tests, and, structurally, incapable of making a model call
// (it takes no *llm.Client): the <200 ms path this covers forbids one by
// construction, not by caller discipline.
func Stall(day int64, villager string) string {
	i := (day + int64(len(villager))) % int64(len(StallLines))
	if i < 0 {
		i += int64(len(StallLines))
	}
	return StallLines[i]
}

func messageText(msg *anthropic.Message) string {
	if msg == nil || len(msg.Content) == 0 {
		return ""
	}
	return msg.Content[0].Text
}

func splitLines(s string) []string {
	var lines []string
	for _, l := range strings.Split(s, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			lines = append(lines, l)
		}
	}
	return lines
}
