// Package seam (this file): percept ingest bookkeeping — origin
// classification (V-6), percept_id dedup scoped to a reconnect, and
// seq-gap shed accounting (docs/design/body-protocol-v0.md §2.7/§4.1,
// docs/design/seam-wire-v0.md §3). Presence validation (V-5) already runs
// in wireConn.ReadMessage before any message reaches HandleConnection's
// switch (port.go) — nothing here ever sees a percept that failed it, so
// the "validate before mutate" ordering (card AC #2/#3) is a property of
// the call graph, not something this file has to re-check.
package seam

import "sync"

// directOrigins is body-protocol-v0.md §2.7's closed first-hand set.
var directOrigins = map[string]bool{
	"acted": true, "saw": true, "heard": true, "felt": true,
}

// ClassifyOrigin implements §2.7's classifier: a pure function of origin
// alone (EH-3 — it MUST NOT be extended to look at anything else on the
// percept). The four direct origins classify firsthand; everything
// else — including an origin this daemon has never heard of, as a future
// minor version might mint — classifies secondhand, never direct
// (EH-2b, card AC #4).
func ClassifyOrigin(origin string) string {
	if directOrigins[origin] {
		return "firsthand"
	}
	return "secondhand"
}

// bodyIngest is one body's ingest bookkeeping. seen is the percept_id
// dedup ledger, which spans a reconnect by design (seam-wire-v0.md §3.4:
// within one session the stream cannot duplicate, so an id reappearing
// is unambiguously a retransmission). haveSeq/lastSeq/shed are seq-gap
// accounting and are reset on every Attach, because seq restarts at 0 in
// a new session (§3.2) — carrying the old counter across a reconnect
// would manufacture a bogus gap.
type bodyIngest struct {
	seen    map[string]bool
	haveSeq bool
	lastSeq int64
	shed    int64
}

// Ingester is one daemon process's percept-ingest state, shared across
// every connection it accepts so percept_id dedup can span a reconnect
// (seam-wire-v0.md §3.4). It is deliberately NOT durable: a daemon
// restart starts a fresh, empty Ingester — durable memory is M2's job
// (FR-008), and losing the dedup ledger on restart is honest, not a
// shortcut: nothing here ever claims continuity it cannot back with
// evidence (matches HandleConnection's continuity handling, T008).
type Ingester struct {
	mu     sync.Mutex
	bodies map[string]*bodyIngest

	// OnPercept, if set, runs synchronously after a percept is admitted
	// (validated by the time it reaches here, and not a duplicate).
	// Phase 3's end-to-end test uses this to drive an intent out (T012)
	// and to resolve act_results against a Pending set (T010) — there is
	// no real deliberation here or anywhere in this daemon (M5 replaces
	// this whole mechanism). Nil is the safe, do-nothing default.
	OnPercept func(conn Conn, body string, msg map[string]any)

	// Archived, if set, reports whether body names an archived mind
	// (specs/018-consolidation T005, ruling R-9): a true here refuses the
	// session_open naming it, the same session_close/reason:error shape
	// HandleConnection already uses for a manifest mismatch. Nil (the
	// default) never refuses — Ingester itself has no notion of
	// archival; a real daemon wires *consolidate.Archive.IsArchived here.
	Archived func(body string) bool
}

// NewIngester returns an empty Ingester ready to use.
func NewIngester() *Ingester {
	return &Ingester{bodies: map[string]*bodyIngest{}}
}

func (ing *Ingester) bodyOrCreate(body string) *bodyIngest {
	b := ing.bodies[body]
	if b == nil {
		b = &bodyIngest{seen: map[string]bool{}}
		ing.bodies[body] = b
	}
	return b
}

// Attach (re)starts body's seq counter at seq and preserves its dedup
// ledger. Call it whenever body's session_open is processed — the first
// attach on a connection or a later reconnect — since seam-wire-v0.md
// §3.2 starts the vendor→mind counter for a body at 0 with that
// session_open, and every message afterward (percepts, intent_ack,
// session_close, …) shares that one counter.
func (ing *Ingester) Attach(body string, seq int64) {
	ing.mu.Lock()
	defer ing.mu.Unlock()
	b := ing.bodyOrCreate(body)
	b.haveSeq, b.lastSeq, b.shed = true, seq, 0
}

// Observe updates seq-gap accounting for any vendor→mind message naming
// body, returning the shed count this call's gap accounts for (0 if
// none). It must run for every message in that direction for body, not
// only percepts — a gap can only mean deliberate background-percept
// shedding (seam-wire-v0.md §3.3) if nothing else quietly consumed a seq
// number in between; an intent_ack sharing the same per-body counter
// would otherwise be misread as a shed percept.
func (ing *Ingester) Observe(body string, seq int64) (shed int64) {
	ing.mu.Lock()
	defer ing.mu.Unlock()
	b := ing.bodyOrCreate(body)
	if b.haveSeq && seq > b.lastSeq+1 {
		shed = seq - b.lastSeq - 1
		b.shed += shed
	}
	b.haveSeq, b.lastSeq = true, seq
	return shed
}

// Dedup reports whether perceptID has already been seen for body — a
// dup means it must be dropped and mutate nothing further (§3.4's one
// job for percept_id). It records perceptID as seen either way isn't
// true: a duplicate is not re-recorded, but a first sighting is.
func (ing *Ingester) Dedup(body, perceptID string) (dup bool) {
	ing.mu.Lock()
	defer ing.mu.Unlock()
	b := ing.bodyOrCreate(body)
	if b.seen[perceptID] {
		return true
	}
	b.seen[perceptID] = true
	return false
}

// ShedCount reports the cumulative shed-percept count observed for body
// in its current session (reset by the last Attach).
func (ing *Ingester) ShedCount(body string) int64 {
	ing.mu.Lock()
	defer ing.mu.Unlock()
	if b := ing.bodies[body]; b != nil {
		return b.shed
	}
	return 0
}
