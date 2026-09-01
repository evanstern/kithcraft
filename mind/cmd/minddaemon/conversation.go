// Command minddaemon (this file): TASK-0023 phase 2 — the dusk exchange
// (T005, US3) and the ambient pool (T006, US4), both driven off
// classifications of the ordinary `sighting` percepts a live session
// already carries (plan.md design decision 1's "triggers are
// classifications of existing percepts," applied to M6's packages the way
// deliberation.go already applied it to M5's).
//
// Pair detection (T005) consumes V3's pairing-signal sighting exactly as
// mod/src/main/java/dev/kithcraft/mod/brain/PairingSignal.java composes it
// (that class's own doc explains why `sighting` and not a new percept
// type): a `k:person` thing whose `descriptor` is the OTHER cast member's
// display name and `body` is their token, `doing` fixed to "walking to the
// gathering place". DuskPairing.java fires this TWICE per approaching
// pair, once per perceiver, each one naming the OTHER member — so two
// signals converge one pair. This build's bounded heuristic (recorded,
// per plan.md's Risks note and spec.md's edge cases): the FIRST signal for
// a (pairID, day) starts the pregen Fill for its sender (the designated
// opener — always live-attached by construction, since it just sent us
// this percept); the SECOND, from the other side, IS convergence (the wire
// carries no separate "arrived"/"met" percept yet) and runs the live
// Exchange. A pair that only ever signals once — the other side never
// confirms — is discarded unspoken after pairConvergeTimeout
// (abort-discard).
//
// Ambient (T006) claims every OTHER `k:person` sighting handlePairSignal's
// exact shape doesn't match: the smallest live trigger this build has for
// "a player passing a villager" (spec.md US4). A generic one (no `doing`
// text — Sightings.sightingContent's `doing` MAY be null) serves the
// AmbientPool; one carrying specific `doing` text is "about something
// specific" (M6's IsTargeted) and escalates to a live call instead.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"kithcraft/mind/consolidate"
	"kithcraft/mind/converse"
	"kithcraft/mind/deliberate"
	"kithcraft/mind/prompt"
	"kithcraft/mind/seam"
)

// pairSignalDoing is the exact `doing` text PairingSignal.content composes
// (its own doc: "~10s before arrival"). Matching it verbatim is what makes
// isPairSignal a classification of an EXISTING percept, not a new shape.
const pairSignalDoing = "walking to the gathering place"

// pairConvergeTimeout bounds how long a pending pair waits for its OTHER
// side to signal before its pregen Slot is discarded unspoken (T005's
// abort-discard case, spec.md Edge Cases). The wire carries no separate
// "meeting failed" percept (DuskPairing.java never sends one), so a
// generous timeout well past PairingSignal.LEAD_SECONDS (10s) stands in.
// Var, not const, so conversation_test.go can shorten it for a fast test.
var pairConvergeTimeout = 30 * time.Second

// pendingPair is one pair's bookkeeping between its first and (hoped-for)
// second pairing-signal sighting. Every field is written once, before pp
// is published into Runtime.pairs — safe to read without a lock once
// obtained via a locked map read (mu is the publication edge). Guarded as
// a whole by Runtime.mu, same as bodyStore's fields.
type pendingPair struct {
	firstBody string
	slot      *converse.Slot
	timer     *time.Timer
}

// personSighting extracts a `sighting` percept's `k:person` thing content
// (its body token, descriptor, and the sighting's own `doing` text), or
// ok=false if payload isn't that shape at all — a different percept_type,
// or a sighting of something other than a person.
func personSighting(payload map[string]any) (thingBody, descriptor, doing string, ok bool) {
	if pt, _ := payload["percept_type"].(string); pt != "sighting" {
		return "", "", "", false
	}
	content, _ := payload["content"].(map[string]any)
	thing, _ := content["thing"].(map[string]any)
	if kind, _ := thing["kind"].(string); kind != "k:person" {
		return "", "", "", false
	}
	thingBody, _ = thing["body"].(string)
	descriptor, _ = thing["descriptor"].(string)
	doing, _ = content["doing"].(string)
	return thingBody, descriptor, doing, true
}

// isPairSignal reports whether payload is V3's dusk pairing signal (see
// this file's package doc). otherBody/otherName name the OTHER pair
// member — the one real wire signal (besides body==CastID) that ever
// carries a persona's identity; conversation.go's bindPersonaIfUnbound
// uses it.
func isPairSignal(payload map[string]any) (otherBody, otherName string, ok bool) {
	body, name, doing, ok := personSighting(payload)
	if !ok || doing != pairSignalDoing || body == "" {
		return "", "", false
	}
	return body, name, true
}

// canonicalPairID orders two body tokens so either side's own sighting of
// the other derives the SAME PairKey — plain string comparison, since body
// tokens carry no other ordering the wire defines.
func canonicalPairID(a, b string) string {
	if a < b {
		return a + "+" + b
	}
	return b + "+" + a
}

// bindPersonaIfUnbound binds bs's persona to castName if bs has none yet
// and castName names a loaded CastID — the opportunistic binding source a
// pairing-signal sighting gives for the OTHER body (see this file's
// package doc and runtime.go's top-of-file comment). Never overwrites an
// already-bound persona.
func (rt *Runtime) bindPersonaIfUnbound(bs *bodyStore, castName string) {
	p, ok := rt.Personas[castName]
	if !ok {
		return
	}
	rt.mu.Lock()
	if !bs.hasPersona {
		bs.persona, bs.hasPersona = p, true
	}
	rt.mu.Unlock()
}

// exchangeSpeaker builds one side's converse.Speaker for the dusk exchange
// (T005): the SAME body-to-persona binding and stable-prefix construction
// E2/E3 use (deliberation.go's stablePrefixFor — plan.md design decision
// 5's "E2/E3 and E4 prefixes get the same binding"), plus this side's live
// seam.Conn/session/verbs and an Interlocutor naming the other side. false
// means no bound persona (spec.md's stub-cast edge case) — the caller's
// cue to discard rather than compose a guessed one.
func (rt *Runtime) exchangeSpeaker(body string, bs *bodyStore, partnerName string) (*converse.Speaker, bool) {
	p, ok := rt.personaFor(bs)
	if !ok {
		return nil, false
	}
	rt.mu.Lock()
	caps, conn, session := bs.capabilities, bs.conn, bs.session
	rt.mu.Unlock()
	verbs := deliberate.ManifestVerbs(caps)
	return &converse.Speaker{
		Name:         body,
		Client:       rt.Client,
		Pending:      seam.NewPending(verbs),
		Out:          conn,
		Session:      session,
		Body:         body,
		Stable:       stablePrefixFor(p, verbs),
		Interlocutor: converse.Interlocutor{Who: partnerName},
	}, true
}

// castNameOf returns bs's bound persona's CastID, or "" if unbound.
func (rt *Runtime) castNameOf(bs *bodyStore) string {
	if p, ok := rt.personaFor(bs); ok {
		return p.CastID
	}
	return ""
}

// handlePairSignal is HandlePercept's T005 tail. See this file's package
// doc for the convergence heuristic; nil client or no bound persona for
// the opener just means the pair proceeds with no pregen Fill — Exchange's
// own live-fallback (Config.OpeningSlot nil/unfilled) already covers it,
// and V3's live pregen fallback is documented as first-class (research/
// specs/014-augmented-villager: measured lead 1.82-4.96s).
func (rt *Runtime) handlePairSignal(body string, bs *bodyStore, worldTime int64, payload map[string]any) {
	otherBody, otherName, ok := isPairSignal(payload)
	if !ok {
		return
	}
	otherBS, err := rt.bodyOrOpen(otherBody)
	if err != nil {
		fmt.Fprintf(os.Stderr, "minddaemon: opening stores for paired body %q: %v\n", otherBody, err)
		return
	}
	rt.bindPersonaIfUnbound(otherBS, otherName)

	key := converse.PairKey{PairID: canonicalPairID(body, otherBody), Day: worldTime / consolidate.CycleTicks}

	rt.mu.Lock()
	pp, seenBefore := rt.pairs[key]
	rt.mu.Unlock()

	if seenBefore && pp.firstBody != body {
		// Convergence: the OTHER side already signaled. Claim it exactly
		// once — a re-observed convergence (e.g. a duplicate percept)
		// finds nothing left to claim.
		rt.mu.Lock()
		claimed := rt.pairs[key] == pp
		if claimed {
			delete(rt.pairs, key)
		}
		rt.mu.Unlock()
		if !claimed {
			return
		}
		if pp.timer != nil {
			pp.timer.Stop()
		}
		go rt.runExchange(pp.slot, pp.firstBody, otherBS, body, bs, worldTime)
		return
	}
	if seenBefore {
		return // a re-fired signal from the same side — not convergence
	}

	// First signal for this pair: body (the sender, live by construction)
	// is the designated opener. A concurrent first signal from the OTHER
	// side racing this one resolves by last-write-wins on rt.pairs[key];
	// the loser's Fill/timer become inert (abortPair checks pp identity,
	// not just the key, so a superseded timer never touches the winner).
	pp = &pendingPair{firstBody: body}
	if speaker, ok := rt.exchangeSpeaker(body, bs, otherName); ok {
		pp.slot = rt.convPool.Begin(context.Background(), key, speaker)
	}
	pp.timer = time.AfterFunc(pairConvergeTimeout, func() { rt.abortPair(key, pp) })

	rt.mu.Lock()
	rt.pairs[key] = pp
	rt.mu.Unlock()
}

// abortPair discards pp's pregen Slot unspoken if it is still the current
// entry for key (T005's abort-discard case) — identity-checked so a timer
// from a pair that already converged, or lost a first-signal race, never
// touches a different pending pair that happens to share the same key.
func (rt *Runtime) abortPair(key converse.PairKey, pp *pendingPair) {
	rt.mu.Lock()
	current := rt.pairs[key] == pp
	if current {
		delete(rt.pairs, key)
	}
	rt.mu.Unlock()
	if current && pp.slot != nil {
		pp.slot.Discard()
	}
}

// runExchange is convergence: both sides' Speakers are built for real
// (live Out/Session, so each speak intent rides that speaker's OWN
// session — plan.md's Risks note), the opener's turn consults sl (the
// pregen Slot, possibly nil — Exchange's own documented live fallback),
// and every turn's FirstTokenLatency is recorded for the session report
// (T007).
func (rt *Runtime) runExchange(sl *converse.Slot, openerBody string, openerBS *bodyStore, responderBody string, responderBS *bodyStore, worldTime int64) {
	opener, ok := rt.exchangeSpeaker(openerBody, openerBS, rt.castNameOf(responderBS))
	if !ok {
		if sl != nil {
			sl.Discard()
		}
		return
	}
	responder, ok := rt.exchangeSpeaker(responderBody, responderBS, rt.castNameOf(openerBS))
	if !ok {
		if sl != nil {
			sl.Discard()
		}
		return
	}

	turns, err := converse.Exchange(context.Background(), opener, responder, converse.Config{
		WorldTime:   worldTime,
		OpeningSlot: sl,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "minddaemon: dusk exchange %s<->%s: %v\n", openerBody, responderBody, err)
	}
	rt.recordLatencies(turns)
}

// recordLatencies files every turn's FirstTokenLatency under its own
// speaker's bodyStore (T007) — a session-end report reads them back
// (report.go).
func (rt *Runtime) recordLatencies(turns []converse.Turn) {
	for _, t := range turns {
		bs, err := rt.bodyOrOpen(t.Speaker)
		if err != nil {
			continue
		}
		rt.mu.Lock()
		bs.turnLatencies = append(bs.turnLatencies, t.FirstTokenLatency)
		rt.mu.Unlock()
	}
}

// handleAmbientTrigger is HandlePercept's T006 tail (US4 AC #2). See this
// file's package doc for the trigger definition.
func (rt *Runtime) handleAmbientTrigger(body string, bs *bodyStore, worldTime int64, payload map[string]any) {
	_, _, doing, ok := personSighting(payload)
	if !ok || doing == pairSignalDoing {
		return // not a sighting, or handlePairSignal already claims this exact shape
	}
	if rt.Client == nil {
		return // rehearsal mode — nothing to call
	}
	p, ok := rt.personaFor(bs)
	if !ok {
		fmt.Fprintf(os.Stderr, "minddaemon: ambient trigger for body %q but no bound persona — skipped\n", body)
		return
	}
	stable := prompt.AmbientStablePrefix{PersonaThumbnail: p.Name + " — " + p.Anchor}
	day := worldTime / consolidate.CycleTicks

	if converse.IsTargeted(doing) {
		line, err := converse.Escalate(context.Background(), rt.Client, stable, doing, "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "minddaemon: ambient escalate for body %q: %v\n", body, err)
			return
		}
		rt.speakLine(body, bs, line)
		return
	}
	if line, _, ok := rt.ambient.Serve(body, day); ok {
		rt.speakLine(body, bs, line)
	} else {
		rt.speakLine(body, bs, converse.Stall(day, body))
	}
}

// refillAmbient is T006's day-crossing action (US4 AC #1): one batched E5
// call per villager per day-rollover — the same world_time boundary
// runNight already reacts to. Nil client / unbound persona: log-and-skip,
// matching runNight's own rehearsal posture.
//
// ponytail: mood/whoAround/recentNotable ride empty — deriving them from
// this body's own memory window is prompt CONTENT (E5's variable
// context), not this phase's job (matches deliberationContext's own
// OtherClaims/Relationship/Commitments ponytail, deliberation.go).
func (rt *Runtime) refillAmbient(body string, bs *bodyStore, worldTime int64) {
	if rt.Client == nil {
		return
	}
	p, ok := rt.personaFor(bs)
	if !ok {
		fmt.Fprintf(os.Stderr, "minddaemon: ambient refill for body %q but no bound persona — skipped\n", body)
		return
	}
	stable := prompt.AmbientStablePrefix{PersonaThumbnail: p.Name + " — " + p.Anchor}
	day := worldTime / consolidate.CycleTicks
	if _, err := rt.ambient.Refill(context.Background(), rt.Client, body, day, stable, "", "", ""); err != nil {
		fmt.Fprintf(os.Stderr, "minddaemon: ambient refill for body %q: %v\n", body, err)
	}
}

// speakLine composes line as a `speak` intent — target {"type":"none",
// "text":line}, §5.2's only slot for a speak's text, the same shape
// converse.Speaker.speak uses — and sends it on body's live session
// through the same wireVendor T001 already wired.
func (rt *Runtime) speakLine(body string, bs *bodyStore, line string) {
	rt.mu.Lock()
	verbs := deliberate.ManifestVerbs(bs.capabilities)
	rt.mu.Unlock()
	payload, err := seam.NewPending(verbs).Compose(
		fmt.Sprintf("i-ambient-%s-%d", body, time.Now().UnixNano()),
		"speak", map[string]any{"type": "none", "text": line}, "ambient remark", "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "minddaemon: composing ambient speak intent for body %q: %v\n", body, err)
		return
	}
	if err := (&wireVendor{rt: rt, body: body}).SendIntent(payload); err != nil {
		fmt.Fprintf(os.Stderr, "minddaemon: sending ambient speak intent for body %q: %v\n", body, err)
	}
}
