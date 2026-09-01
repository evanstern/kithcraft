// Command minddaemon (this file): TASK-0021 T001/T002 — the daemon's real
// runtime, assembled from already-tested packages (mind/persona,
// mind/memory, mind/consolidate, mind/llm). Per the plan's design decision
// 2, this wires the MINIMUM that makes I1's card true: persona load-or-
// genesis at start, per-body memory log + nightly ledger opened lazily
// under a run dir, the shared archive consulted on session_open (closing
// TASK-0018's named daemon-wiring deferral), and RunNight invoked off the
// sleep signal (consolidate.SleepTriggered) as percepts flow through. No
// deliberation/converse wiring lives here (M5) — this file never composes
// an intent.
//
// ponytail: persona identity (fixed CastIDs: Aldric/Petra/Yenna) and body
// identity (opaque per-boot session tokens) are disjoint namespaces in this
// build (see T003's fix in mod/ for the same fact on the vendor side) — a
// live session's memory/ledger stores are keyed by its `body` token, not by
// a CastID, matching consolidate/archive.go's own mind-identity-is-body-
// token convention. Binding a body's stream to a specific persona's name
// for E6's stable prefix is deliberation-adjacent work M5's real mind-
// identity layer will do; RunNight below uses an empty ConsolidationStable
// Prefix, which still exercises the real E6 call/parse/ledger path end to
// end.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"kithcraft/mind/consolidate"
	"kithcraft/mind/deliberate"
	"kithcraft/mind/llm"
	"kithcraft/mind/memory"
	"kithcraft/mind/persona"
	"kithcraft/mind/prompt"
	"kithcraft/mind/seam"
)

// bodyStore is one live body's durable state, opened lazily the first time
// that body is seen. lastWorldTime is -1 until the first message, so the
// first sighting can never itself read as a sleep-boundary crossing.
type bodyStore struct {
	log           *memory.Log
	ledger        *consolidate.Ledger
	gate          *memory.Gate
	instrument    *memory.Instrument
	lastWorldTime int64

	// TASK-0023 phase 1 (T001-T003): this body's live manifest/session,
	// set at session_open (Runtime.HandleSessionOpen) and read by
	// wireVendor and the deliberation loop — all fields below are guarded
	// by Runtime.mu, the same lock lastWorldTime above already uses.
	capabilities map[string]any
	conn         seam.Conn
	session      string
	mindSeq      int64
	loop         *deliberate.Loop // the Loop currently in flight, if any (HandlePercept.Deliver's target)

	// The per-body deliberation loop: one goroutine draining triggerCh so
	// at most one deliberate.Loop.Run is ever in flight for this body
	// (plan.md design decision 2). Built lazily, once, on this body's
	// first E2/E3/urgent trigger (deliberation.go).
	delibOnce sync.Once
	interrupt *deliberate.Interrupt
	triggerCh chan map[string]any
}

// Runtime is the daemon's assembled, non-skeleton state. Client/Digester
// are nil together when ANTHROPIC_API_KEY is unset (spec.md FR-007's
// zero-call rehearsal path): genesis then refuses loudly if any persona is
// actually missing, and a sleep-boundary crossing logs and skips rather
// than panicking.
type Runtime struct {
	VillagerDir string
	PersonaDir  string
	Archive     *consolidate.Archive
	Client      *llm.Client
	Digester    consolidate.Digester
	Personas    map[string]persona.Persona

	mu     sync.Mutex
	bodies map[string]*bodyStore
}

// NewRuntime opens runDir's stores (T001): the shared archive eagerly (its
// IsArchived hook must be live before the first session_open), per-body
// logs/ledgers lazily in bodyOrOpen. It makes no model calls — Client/
// Digester are wired from ANTHROPIC_API_KEY alone, never dialed here.
func NewRuntime(runDir string) (*Runtime, error) {
	villagerDir := filepath.Join(runDir, "villagers")
	personaDir := filepath.Join(runDir, "persona")
	if err := os.MkdirAll(villagerDir, 0o755); err != nil {
		return nil, fmt.Errorf("minddaemon: %w", err)
	}
	archive, err := consolidate.OpenArchive(consolidate.ArchivePathFor(villagerDir))
	if err != nil {
		return nil, fmt.Errorf("minddaemon: %w", err)
	}
	rt := &Runtime{
		VillagerDir: villagerDir,
		PersonaDir:  personaDir,
		Archive:     archive,
		bodies:      map[string]*bodyStore{},
	}
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		rt.Client = llm.New()
		rt.Digester = consolidate.ClientDigester{Client: rt.Client}
	}
	return rt, nil
}

// Close closes every store this runtime opened. Errors are joined so a
// shutdown reports every failure rather than only the first.
func (rt *Runtime) Close() error {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	var errs []error
	for _, bs := range rt.bodies {
		errs = append(errs, bs.log.Close(), bs.ledger.Close())
	}
	errs = append(errs, rt.Archive.Close())
	return errors.Join(errs...)
}

// LoadOrGenesisCast is T001's persona half (M3's Load/Genesis, FR-002/
// FR-003): every demo cast id already on disk re-binds via Load (no
// re-genesis); anything missing genesis-runs, resuming only the missing
// ids (0444 files from an interrupted prior run persist and are never
// touched again — Load, not Genesis, is what a present file gets). With
// nothing missing this makes zero model calls, satisfying the rehearsal
// path. With something missing and no key, it fails loudly naming exactly
// which cast ids need it — no partial cast is ever bound in that case.
func (rt *Runtime) LoadOrGenesisCast(ctx context.Context) error {
	entries := persona.DemoCast()
	loaded := make(map[string]persona.Persona, len(entries))
	var missing []persona.CastEntry
	for _, e := range entries {
		p, err := persona.Load(rt.PersonaDir, []string{e.CastID})
		if err != nil {
			missing = append(missing, e)
			continue
		}
		loaded[e.CastID] = p[e.CastID]
	}
	if len(missing) == 0 {
		rt.Personas = loaded
		return nil
	}
	if rt.Client == nil {
		ids := make([]string, len(missing))
		for i, e := range missing {
			ids[i] = e.CastID
		}
		return fmt.Errorf("minddaemon: persona genesis needed for %v but ANTHROPIC_API_KEY is not set — export it and rerun, or pre-seed %s from a prior live run", ids, rt.PersonaDir)
	}
	generated, err := persona.Genesis(ctx, rt.Client, rt.PersonaDir, missing)
	if err != nil {
		return fmt.Errorf("minddaemon: persona genesis: %w", err)
	}
	for _, p := range generated {
		loaded[p.CastID] = p
	}
	rt.Personas = loaded
	return nil
}

// bodyOrOpen returns body's store, opening its log/ledger under VillagerDir
// on first sight (M2/M7's per-villager file convention, PathFor).
func (rt *Runtime) bodyOrOpen(body string) (*bodyStore, error) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	if bs, ok := rt.bodies[body]; ok {
		return bs, nil
	}
	log, err := memory.Open(memory.PathFor(rt.VillagerDir, body))
	if err != nil {
		return nil, err
	}
	ledger, err := consolidate.OpenLedger(consolidate.PathFor(rt.VillagerDir, body))
	if err != nil {
		log.Close()
		return nil, err
	}
	// TASK-0021 T005 (card AC #5): one Instrument per body, day-length
	// matching R-1's kept cycle (consolidate.CycleTicks) — the same
	// villager-day granularity RunNight's own sleep boundary uses.
	bs := &bodyStore{log: log, ledger: ledger, gate: memory.NewGate(), instrument: memory.NewInstrument(consolidate.CycleTicks), lastWorldTime: -1}
	rt.bodies[body] = bs
	return bs, nil
}

// HandleSessionOpen is the real listener's Ingester.OnSessionOpen hook
// (TASK-0023 T001/T002): records the manifest capabilities and the live
// conn/session a body's Vendor and deliberation Loop read from —
// capabilities arrive on session_open only, never on a percept, and
// mind/deliberate/loop.go's Config.Verbs doc requires the body's actual
// declared verb set, not an invented one.
func (rt *Runtime) HandleSessionOpen(conn seam.Conn, session, body string, capabilities map[string]any) {
	bs, err := rt.bodyOrOpen(body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "minddaemon: opening stores for body %q at session_open: %v\n", body, err)
		return
	}
	rt.mu.Lock()
	bs.capabilities, bs.conn, bs.session = capabilities, conn, session
	rt.mu.Unlock()
}

// HandlePercept is the real listener's Ingester.OnPercept hook (T001/T002,
// TASK-0023 T002/T003): runs the admission gate (mind/memory.Gate) and
// appends admitted percepts to that body's log, routes an act_result to
// whichever deliberate.Loop.Run is waiting on it (Loop.Deliver — M5's ONLY
// path to a Run's OnFact), classifies the percept against the live E2/E3/
// urgent triggers, then checks the sleep signal on the message's own
// world_time and runs a night's consolidation when it fires.
func (rt *Runtime) HandlePercept(conn seam.Conn, body string, msg map[string]any) {
	worldTime, _ := msg["world_time"].(int64)
	payload, _ := msg["payload"].(map[string]any)

	bs, err := rt.bodyOrOpen(body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "minddaemon: opening stores for body %q: %v\n", body, err)
		return
	}

	if pt, _ := payload["percept_type"].(string); pt == "act_result" {
		rt.mu.Lock()
		l := bs.loop
		rt.mu.Unlock()
		if l != nil {
			l.Deliver(payload)
		}
	}

	if admit, _ := bs.gate.Decide(payload); admit {
		if _, err := bs.log.Append(eventInputFromPercept(worldTime, payload)); err != nil {
			fmt.Fprintf(os.Stderr, "minddaemon: memory append for body %q: %v\n", body, err)
		} else {
			bs.instrument.Record(worldTime)
		}
	}

	rt.mu.Lock()
	prev := bs.lastWorldTime
	bs.lastWorldTime = worldTime
	rt.mu.Unlock()

	rt.classifyAndTrigger(body, bs, payload)

	if consolidate.SleepTriggered(prev, worldTime) {
		rt.runNight(body, bs, worldTime)
	}
}

// runNight is T002's trigger action. A nil Digester (no ANTHROPIC_API_KEY)
// logs and skips rather than failing loudly — the same window is retried
// at the next boundary a Digester is available for (consolidate.RunNight's
// own no-marker-on-failure rule, extended by construction: skipping lands
// no marker either).
func (rt *Runtime) runNight(body string, bs *bodyStore, worldTime int64) {
	if rt.Digester == nil {
		fmt.Fprintf(os.Stderr, "minddaemon: sleep boundary at world_time %d for body %q but no ANTHROPIC_API_KEY — consolidation skipped, window not marked (will retry)\n", worldTime, body)
		return
	}
	if err := consolidate.RunNight(context.Background(), bs.log, bs.ledger, rt.Digester, prompt.ConsolidationStablePrefix{}, worldTime); err != nil {
		fmt.Fprintf(os.Stderr, "minddaemon: RunNight for body %q: %v\n", body, err)
	}
}

// eventInputFromPercept builds M2's EventInput from a percept payload's
// provenance/content — the field names §2.6/§4.1 fix (session.go's
// payloadOf already validated presence before this hook ever runs).
func eventInputFromPercept(worldTime int64, payload map[string]any) memory.EventInput {
	provenance, _ := payload["provenance"].(map[string]any)
	origin, _ := provenance["origin"].(string)
	receivedAt, _ := provenance["received_at"].(int64)
	var observedAt *int64
	if v, ok := provenance["observed_at"].(int64); ok {
		observedAt = &v
	}
	perceptID, _ := payload["percept_id"].(string)
	perceptType, _ := payload["percept_type"].(string)
	return memory.EventInput{
		WorldTime:   worldTime,
		Origin:      origin,
		PerceptID:   perceptID,
		PerceptType: perceptType,
		ReceivedAt:  receivedAt,
		ObservedAt:  observedAt,
		Content:     payload["content"],
	}
}
