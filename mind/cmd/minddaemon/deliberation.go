// Command minddaemon (this file): TASK-0023 phase 1 (T002/T003) — E2/E3
// trigger classification over existing percepts (docs/design/llm-routing-
// and-budget.md §2.2, mind/deliberate.TriggerE3), the per-body
// deliberation loop that composes mind/deliberate.Loop with real context
// (a bound persona, the K=10 window over this body's own memory log), and
// the §5.5 interrupt registered on that loop. Only one deliberation is
// ever in flight per body (plan.md design decision 2): each body gets one
// goroutine draining a trigger queue, so a trigger arriving mid-Run just
// waits its turn.
package main

import (
	"context"
	"fmt"
	"hash/fnv"
	"os"
	"sort"
	"strings"

	"kithcraft/mind/consolidate"
	"kithcraft/mind/deliberate"
	"kithcraft/mind/llm"
	"kithcraft/mind/memory"
	"kithcraft/mind/persona"
	"kithcraft/mind/prompt"
)

// triggerE2 classifies a percept against §2.2's E2 trigger list —
// "schedule transition, act completed with an open choice, notable
// percept at a natural break" — per plan.md design decision 1: Runtime
// classifies, no new percept type is introduced. The fourth trigger,
// "deferred urgent interrupt", is not a percept shape at all — it is
// Interrupt's own coalesced follow-up (interrupt.go), enqueued as a nil
// trigger by classifyAndTrigger below.
//
// ponytail: exactly what counts as "a natural break" is approximated —
// self_state (§4.9, the body's own condition) and a notable sighting —
// since the daemon has no schedule-transition percept shape of its own
// yet. Refine against the dev-server observation (tasks.md T009) if it
// finds this over- or under-fires.
func triggerE2(percept map[string]any) bool {
	pt, _ := percept["percept_type"].(string)
	switch pt {
	case "self_state":
		return true
	case "sighting":
		urgency, _ := percept["urgency"].(string)
		return urgency == "notable"
	case "act_result":
		content, _ := percept["content"].(map[string]any)
		outcome, _ := content["outcome"].(string)
		return outcome == "completed" // the act finished, leaving an open choice: what next?
	default:
		return false
	}
}

// classifyAndTrigger is HandlePercept's T002/T003 tail: it lazily starts
// body's deliberation-loop goroutine, then either feeds an urgent percept
// to M5's Interrupt (§5.5 — never a model call of its own) or, on an E2/
// E3 hit, queues the percept for that body's single loop.
func (rt *Runtime) classifyAndTrigger(body string, bs *bodyStore, payload map[string]any) {
	bs.delibOnce.Do(func() { rt.startDeliberationLoop(body, bs) })

	if deliberate.IsUrgent(payload) {
		bs.interrupt.Urgent(payload)
		return
	}
	if deliberate.TriggerE3(payload) || triggerE2(payload) {
		select {
		case bs.triggerCh <- payload:
		default:
			fmt.Fprintf(os.Stderr, "minddaemon: trigger queue full for body %q, dropping a trigger (a deliberation is already backed up)\n", body)
		}
	}
}

// startDeliberationLoop builds body's trigger queue and Interrupt and
// starts the one goroutine that ever runs a deliberate.Loop.Run for it.
// onEnqueue (Interrupt's coalescing signal) pushes a nil trigger — the
// loop's cue to Drain the coalesced urgents and run exactly one follow-up
// (§5.5's "enqueues one deliberation whose context includes the urgent
// percept").
func (rt *Runtime) startDeliberationLoop(body string, bs *bodyStore) {
	bs.triggerCh = make(chan map[string]any, 16)
	bs.interrupt = deliberate.NewInterrupt(func() {
		select {
		case bs.triggerCh <- nil:
		default:
		}
	})
	go func() {
		for trig := range bs.triggerCh {
			rt.runDeliberation(body, bs, trig)
		}
	}()
}

// runDeliberation runs at most one deliberate.Loop.Run for body. trig is
// the triggering percept, or nil for an interrupt-driven follow-up.
// Rehearsal mode (no ANTHROPIC_API_KEY) and a body with no bound persona
// (spec.md Edge Cases) both log and skip — never a model call, never a
// crash, matching runNight's existing nil-Digester posture.
func (rt *Runtime) runDeliberation(body string, bs *bodyStore, trig map[string]any) {
	urgents := bs.interrupt.Drain()

	if rt.Client == nil {
		fmt.Fprintf(os.Stderr, "minddaemon: deliberation trigger for body %q but no ANTHROPIC_API_KEY — skipped (rehearsal mode)\n", body)
		return
	}
	p, ok := rt.personaFor(body)
	if !ok {
		fmt.Fprintf(os.Stderr, "minddaemon: deliberation trigger for body %q but no bound persona (stub cast) — skipped\n", body)
		return
	}

	rt.mu.Lock()
	caps, worldTime := bs.capabilities, bs.lastWorldTime
	rt.mu.Unlock()

	verbs := deliberate.ManifestVerbs(caps)
	class, variable := deliberationContext(trig, urgents, windowFor(body, bs, worldTime))
	a := prompt.Assemble(stablePrefixFor(p, verbs), variable)
	l := deliberate.New(deliberate.Config{Verbs: verbs, Vendor: &wireVendor{rt: rt, body: body}, Class: class})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bs.interrupt.Register(cancel)
	defer bs.interrupt.Register(nil)

	rt.mu.Lock()
	bs.loop = l
	rt.mu.Unlock()
	defer func() {
		rt.mu.Lock()
		bs.loop = nil
		rt.mu.Unlock()
	}()

	// A cancelled Run (ctx.Err() != nil, §5.5) is not logged as a failure:
	// the coalesced follow-up Interrupt already enqueued is the correct
	// next step, not a crash.
	if _, err := l.Run(ctx, singleRoundProposer(rt.Client, class, a)); err != nil && ctx.Err() == nil {
		fmt.Fprintf(os.Stderr, "minddaemon: deliberation for body %q: %v\n", body, err)
	}
}

// personaFor is Phase 1's placeholder cast-entry <-> body-token binding
// (spec.md Edge Cases; plan.md decision 5 is Phase 2's real one, closing
// I1's ponytail): a body token that literally names a loaded CastID is
// bound; anything else has no persona yet, so deliberation skips with a
// log line rather than guessing one. rt.Personas is populated once at
// startup (LoadOrGenesisCast) before serving begins, so reading it here
// needs no lock.
func (rt *Runtime) personaFor(body string) (persona.Persona, bool) {
	p, ok := rt.Personas[body]
	return p, ok
}

// deliberationContext picks E2 vs E3 (deliberate.TriggerE3 on trig — a
// nil trig, the interrupt follow-up case, is always E2 per §2.2's fourth
// trigger) and renders that class's variable context: E3's four-field
// board shape (Board filled from the percept; ponytail — OtherClaims/
// Relationship/Commitments need board-state, relationship, and commitment
// tracking this daemon doesn't keep yet, ManifestVerbs/window are what
// Phase 1 actually has) or a generic E2 context, plus the K=10 window and
// any coalesced urgent percepts either way.
func deliberationContext(trig map[string]any, urgents []map[string]any, window string) (llm.Class, prompt.VariableContext) {
	var vc *prompt.VariableContext
	var class llm.Class
	if trig != nil && deliberate.TriggerE3(trig) {
		content, _ := trig["content"].(map[string]any)
		text, _ := content["text"].(string)
		e3vc := deliberate.E3Context{Board: text}.VariableContext()
		vc, class = &e3vc, llm.E3
	} else {
		vc, class = prompt.NewVariableContext(), llm.E2
		if trig != nil {
			vc.Add("TRIGGER", fmt.Sprintf("%v", trig["content"]))
		}
	}
	vc.Add("WINDOW", window)
	if len(urgents) > 0 {
		vc.Add("URGENT", renderPercepts(urgents))
	}
	return class, *vc
}

func renderPercepts(percepts []map[string]any) string {
	var sb strings.Builder
	for i, p := range percepts {
		if i > 0 {
			sb.WriteByte('\n')
		}
		fmt.Fprintf(&sb, "%v", p["content"])
	}
	return sb.String()
}

// windowFor renders body's K=10 situated window (routing §2.3) from its
// own memory log, seeded deterministically per body so serendipity picks
// are stable villager-to-villager (window.go's SelectWindow doc).
func windowFor(body string, bs *bodyStore, worldTime int64) string {
	snap, byID := windowSnapshot(bs.log.Events())
	chosen := deliberate.SelectWindow(snap, worldTime, seedFor(body))
	var sb strings.Builder
	for i, w := range chosen {
		if i > 0 {
			sb.WriteByte('\n')
		}
		ev := byID[w.ID]
		fmt.Fprintf(&sb, "- [%s] %v", ev.PerceptType(), ev.Content())
	}
	return sb.String()
}

// windowSnapshot maps a body's memory log into deliberate.SelectWindow's
// input shape. mind/memory.Log.Events() (specs/010-event-sourced-memory
// Phase 1) already is the "smallest exported read" plan.md calls for —
// M5's noted deviation (no enumeration existed when that note was
// written) is closed by this pre-existing method, not by a new export.
// Salience is a uniform placeholder: no formativeness scoring exists yet
// (routing doc §1.3's own ponytail — v1 has no scoring pass at all), so
// every admitted event counts the same until one does.
func windowSnapshot(events []memory.Event) (deliberate.Snapshot, map[string]memory.Event) {
	items := make([]deliberate.WindowItem, len(events))
	byID := make(map[string]memory.Event, len(events))
	for i, ev := range events {
		id := fmt.Sprintf("%d:%s", ev.WorldTime(), ev.Hash())
		observedAt := ev.WorldTime()
		if oa := ev.ObservedAt(); oa != nil {
			observedAt = *oa
		}
		items[i] = deliberate.WindowItem{ID: id, Salience: 1.0, ObservedAt: observedAt}
		byID[id] = ev
	}
	return deliberate.Snapshot{Items: items, DayLength: consolidate.CycleTicks}, byID
}

func seedFor(body string) int64 {
	h := fnv.New64a()
	h.Write([]byte(body))
	return int64(h.Sum64())
}

// stablePrefixFor renders a bound persona into E2/E3's shared stable
// prefix shape (§2.3's DeliberationStablePrefix).
func stablePrefixFor(p persona.Persona, verbs map[string]bool) prompt.DeliberationStablePrefix {
	names := make([]string, 0, len(verbs))
	for v := range verbs {
		names = append(names, v)
	}
	sort.Strings(names)
	return prompt.DeliberationStablePrefix{
		Persona:      p.Name + " — " + p.Anchor,
		Values:       strings.Join(p.Values, "; "),
		Manifest:     strings.Join(names, ", "),
		Instructions: "Decide one act and author your own reason (§5.2) — never a boilerplate refusal.",
	}
}

// singleRoundProposer wraps one live llm.Client.Send call as a Proposer:
// the first round sends the real request; every later round signals
// ErrDone (M5's convention, exercised deterministically).
//
// ponytail: Phase 1 wires one decision per trigger. A real multi-round
// deliberation needs the model itself to decide whether to continue,
// which needs prompt content (a "propose again or stop" instruction) this
// phase does not own — add it once E2/E3's real prompts exist.
func singleRoundProposer(client *llm.Client, class llm.Class, a prompt.Assembled) deliberate.Proposer {
	sent := false
	return func(ctx context.Context) (string, error) {
		if sent {
			return "", deliberate.ErrDone
		}
		sent = true
		msg, err := client.Send(ctx, class, a)
		if err != nil {
			return "", err
		}
		if len(msg.Content) == 0 {
			return "", fmt.Errorf("minddaemon: %s response has no content blocks", class)
		}
		return msg.Content[0].Text, nil
	}
}
