// Package consolidate (this file): T002's digest cycle and T003's
// no-marker-on-failure rule — the sleep-event-triggered E6 call itself.
// All windowing is world_time arithmetic (docs/design/llm-routing-and-
// budget.md §6.2, harness T-b); nothing in this file reads a wall clock.
// There is deliberately no per-percept formativeness scoring pass here or
// anywhere else in this package (card AC #4, §1.3): the admission gate
// (mind/memory.Gate) already decided eligibility before an event ever
// reaches Log, and this file's only judgment call is E6's — what mattered
// among what was already admitted.
package consolidate

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"

	"kithcraft/mind/llm"
	"kithcraft/mind/memory"
	"kithcraft/mind/prompt"
)

// Digester is what RunNight needs from a model client for E6: send the
// assembled prompt and report back raw structured-output text, whether the
// response hit its token ceiling, or a transport/cancellation error.
// Declared at the consumer (mirroring mind/seam's Submitter/Injector
// idiom) so tests script it without any SDK or live call (spec.md FR-007).
// There is no batch variant of this interface, by construction — no
// Batch API path exists for RunNight to accidentally take (card AC #1).
type Digester interface {
	Digest(ctx context.Context, a prompt.Assembled) (raw string, truncated bool, err error)
}

// ClientDigester adapts *llm.Client to Digester for E6 — the only call
// this package makes to the model client. Truncated is StopReason ==
// "max_tokens" (I's silent-truncation lesson, docs/design/llm-routing-
// and-budget.md §2.3/RT-7): an over-limit response is reported, never
// parsed as though it were complete.
type ClientDigester struct{ Client *llm.Client }

func (d ClientDigester) Digest(ctx context.Context, a prompt.Assembled) (string, bool, error) {
	msg, err := d.Client.Send(ctx, llm.E6, a)
	if err != nil {
		return "", false, err
	}
	truncated := msg.StopReason == anthropic.StopReasonMaxTokens
	if len(msg.Content) == 0 {
		return "", truncated, fmt.Errorf("consolidate: E6 response has no content blocks")
	}
	return msg.Content[0].Text, truncated, nil
}

// RunNight is T001's sleep-event trigger, invoked once per villager sleep
// event with worldTime taken from that event — never from a wall clock.
// It selects the admitted buffer since the ledger's watermark (T002: the
// consolidated window's exclusion from the next night's buffer), renders
// it under the m1..mN ordinal convention, and either lands a consolidated-
// night marker or, on any failure, lands nothing at all (T003): the
// buffer stays intact and the same window is retried on the next trigger.
//
// An empty buffer is not a failure: it lands an Empty marker (T003), so
// consolidation stays cheap and the watermark still advances.
func RunNight(ctx context.Context, log *memory.Log, ledger *Ledger, d Digester, stable prompt.StablePrefix, worldTime int64) error {
	watermark, _ := ledger.Watermark() // ok=false reads as 0 here: bufferSince treats "never consolidated" as "since the beginning"
	buffer := bufferSince(log.Events(), watermark, worldTime)

	if len(buffer) == 0 {
		return ledger.append(NightRecord{TriggerWorldTime: worldTime, WindowStart: watermark, WindowEnd: worldTime, Empty: true})
	}

	a := prompt.Assemble(stable, *prompt.NewVariableContext().Add("BUFFER", renderBuffer(buffer)))
	raw, truncated, err := d.Digest(ctx, a)
	if err != nil {
		return fmt.Errorf("consolidate: E6 call failed, night %d not consolidated (will retry): %w", worldTime, err)
	}
	if truncated {
		return fmt.Errorf("consolidate: E6 response truncated at max_tokens, night %d not consolidated (will retry) — see docs/design/llm-routing-and-budget.md §2.3", worldTime)
	}

	digest, err := llm.ParseDigest(llm.E6, raw)
	if err != nil {
		return fmt.Errorf("consolidate: E6 digest did not parse, night %d not consolidated (will retry): %w", worldTime, err)
	}

	return ledger.append(NightRecord{
		TriggerWorldTime: worldTime, WindowStart: watermark, WindowEnd: worldTime,
		Digest: &digest, References: mapReferences(buffer, digest.References),
	})
}

// bufferSince is T002's window: every log event landing after the
// watermark (exclusive) and no later than this trigger (inclusive) — pure
// world_time arithmetic, no wall clock anywhere in this package.
func bufferSince(events []memory.Event, watermark, worldTime int64) []memory.Event {
	var buf []memory.Event
	for _, e := range events {
		if e.WorldTime() > watermark && e.WorldTime() <= worldTime {
			buf = append(buf, e)
		}
	}
	return buf
}

// renderBuffer is T002's ordinal m1..mN prompt convention: each admitted
// event gets one line labeled by its position in the buffer, never by its
// (world_time, hash) identity — memories have no stable IDs in the prompt
// (docs/design/llm-routing-and-budget.md §2.3). mapReferences is this
// convention's other half, resolving an accepted "m3" back to the durable
// pair renderBuffer deliberately did not expose to the model.
func renderBuffer(buffer []memory.Event) string {
	var sb strings.Builder
	for i, ev := range buffer {
		if i > 0 {
			sb.WriteByte('\n')
		}
		fmt.Fprintf(&sb, "m%d: [%s] %v", i+1, ev.PerceptType(), ev.Content())
	}
	return sb.String()
}

// mapReferences maps E6's accepted ordinal references back to the durable
// (world_time, hash) identity of the buffer event each one named — the
// convention IS the identity mechanism (card AC #2). An unrecognized or
// out-of-range label is dropped rather than failing the whole digest,
// matching this codebase's coerce-not-reject posture for a model's
// output (RM-2/RM-3).
func mapReferences(buffer []memory.Event, refs []string) []memory.ID {
	ids := make([]memory.ID, 0, len(refs))
	for _, r := range refs {
		n, ok := strings.CutPrefix(r, "m")
		if !ok {
			continue
		}
		idx, err := strconv.Atoi(n)
		if err != nil || idx < 1 || idx > len(buffer) {
			continue
		}
		ids = append(ids, buffer[idx-1].ID())
	}
	return ids
}
