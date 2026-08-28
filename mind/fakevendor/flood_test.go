// flood_test.go (this file): H-6 (T007, specs/015-fake-vendor-harness
// Phase 3, body-protocol-v0.md §4.10/§10.4) — the change_report delivery
// restriction is load-bearing, with a number. Three bodies share a place:
// one acts repeatedly (a take loop), the other two watch. The identical
// script runs twice, restrict_change_reports off then on. Without the
// restriction, the actor is told about its own act twice (act_result, then
// a change_report about the same change) and each witness is told twice
// too (the sighting it actually saw, then a change_report of it) — the
// bookkeeping channel drowns the experiential one, reproducing the 75%
// figure §4.10 cites from promptworld I. Memory counts come from the mind
// side's own admission machinery (mind/memory Gate + Instrument), never
// from the vendor (§10.5: a vendor counting memories would be a read API
// by the back door) — FakeVendor.RestrictChangeReports is read here, by
// the script, to decide delivery per body; FakeVendor itself conditions no
// behaviour on it (Phase 1's ponytail, still true after Phase 3: Emit
// sends whatever it is handed).
package fakevendor_test

import (
	"fmt"
	"testing"

	"kithcraft/mind/fakevendor"
	"kithcraft/mind/memory"
)

const floodSharedPlace = "pl-grove"
const floodThingID = "th-tree"

func floodSighting(id string) map[string]any {
	return map[string]any{
		"percept_id": id, "percept_type": "sighting", "urgency": "background",
		"provenance": map[string]any{"origin": "saw", "source": nil, "observed_at": int64(0), "received_at": int64(0)},
		"place":      map[string]any{"place": floodSharedPlace, "descriptor": "the shared grove"},
		"content":    map[string]any{"thing": map[string]any{"thing_id": floodThingID, "kind": "k:tree", "roles": []any{}, "descriptor": "the apple tree"}},
	}
}

func floodChangeReport(id string) map[string]any {
	return map[string]any{
		"percept_id": id, "percept_type": "change_report", "urgency": "notable",
		"provenance": map[string]any{"origin": "saw", "source": nil, "observed_at": int64(0), "received_at": int64(0)},
		"place":      map[string]any{"place": floodSharedPlace, "descriptor": "the shared grove"},
		"content": map[string]any{
			"change": "altered",
			"thing":  map[string]any{"thing_id": floodThingID, "kind": "k:tree", "roles": []any{}, "descriptor": "the apple tree"},
		},
	}
}

// maybeEmitChangeReport is the script-side embodiment of §4.10's delivery
// restriction: FakeVendor conditions no behaviour on RestrictChangeReports
// itself (Emit sends whatever the script hands it, unconditionally); the
// script is what must consult the flag before deciding a recipient is due
// a change_report — exactly the check a real vendor's own change-report
// sweep has to make. Returning nil without emitting, under restriction, is
// the MUST NOT (§4.10) made structural rather than promised.
func maybeEmitChangeReport(v *fakevendor.FakeVendor, cr map[string]any) error {
	if v.RestrictChangeReports {
		return nil
	}
	return v.Emit(cr)
}

func admitPercept(gate *memory.Gate, inst *memory.Instrument, msg map[string]any) {
	payload, _ := msg["payload"].(map[string]any)
	if ok, _ := gate.Decide(payload); ok {
		wt, _ := msg["world_time"].(int64)
		inst.Record(wt)
	}
}

func totalMemories(inst *memory.Instrument) int {
	total := 0
	for _, c := range inst.Report() {
		total += c
	}
	return total
}

type floodResult struct {
	memoryCount int
	crToActor   int
	crToWitness int
}

// runFloodScenario scripts the §10.4 scenario once: n take-loop iterations
// by the actor, witnessed by two other bodies, with restrict controlling
// RestrictChangeReports on all three. It returns the mind-side admitted
// memory count summed across the three bodies and how many change_reports
// actually reached the actor and the witnesses.
func runFloodScenario(t *testing.T, restrict bool, n int) floodResult {
	t.Helper()

	actorConn, actorMind := newPair(t)
	w1Conn, w1Mind := newPair(t)
	w2Conn, w2Mind := newPair(t)

	actorV := fakevendor.New(actorConn, "s-flood", "b-actor", coreManifest())
	w1V := fakevendor.New(w1Conn, "s-flood", "b-w1", coreManifest())
	w2V := fakevendor.New(w2Conn, "s-flood", "b-w2", coreManifest())
	for _, v := range []*fakevendor.FakeVendor{actorV, w1V, w2V} {
		v.RestrictChangeReports = restrict
		if err := v.Open(); err != nil {
			t.Fatalf("Open: %v", err)
		}
	}
	actorMind.next(t) // session_open
	w1Mind.next(t)
	w2Mind.next(t)

	actorGate, actorInst := memory.NewGate(), memory.NewInstrument(86400)
	w1Gate, w1Inst := memory.NewGate(), memory.NewInstrument(86400)
	w2Gate, w2Inst := memory.NewGate(), memory.NewInstrument(86400)

	var crToActor, crToW1, crToW2 int

	for i := 0; i < n; i++ {
		intentID := fmt.Sprintf("i-take-%d", i)
		actorMind.send(t, intentEnvelope("s-flood", "b-actor", intentID, "take", int64(i+1)))
		actorMind.next(t) // intent_ack — not a percept, not admitted

		// The witnesses see the take happen. Same kind/place every
		// iteration: only the first is a first_sighting admit, later
		// ones are a repeated background sighting of an already-known
		// thing (dropped) — identical in both runs, so it never
		// contributes to the flood ratio either way.
		if err := w1V.Emit(floodSighting(fmt.Sprintf("p-w1-see-%d", i))); err != nil {
			t.Fatalf("w1 Emit: %v", err)
		}
		admitPercept(w1Gate, w1Inst, w1Mind.next(t))
		if err := w2V.Emit(floodSighting(fmt.Sprintf("p-w2-see-%d", i))); err != nil {
			t.Fatalf("w2 Emit: %v", err)
		}
		admitPercept(w2Gate, w2Inst, w2Mind.next(t))

		detail := "took an apple"
		if err := actorV.Resolve(intentID, "completed", nil, &detail); err != nil {
			t.Fatalf("Resolve: %v", err)
		}
		admitPercept(actorGate, actorInst, actorMind.next(t)) // the honest act_result

		if err := maybeEmitChangeReport(actorV, floodChangeReport(fmt.Sprintf("p-cr-actor-%d", i))); err != nil {
			t.Fatalf("actor change_report: %v", err)
		}
		if !restrict {
			admitPercept(actorGate, actorInst, actorMind.next(t))
			crToActor++
		}
		if err := maybeEmitChangeReport(w1V, floodChangeReport(fmt.Sprintf("p-cr-w1-%d", i))); err != nil {
			t.Fatalf("w1 change_report: %v", err)
		}
		if !restrict {
			admitPercept(w1Gate, w1Inst, w1Mind.next(t))
			crToW1++
		}
		if err := maybeEmitChangeReport(w2V, floodChangeReport(fmt.Sprintf("p-cr-w2-%d", i))); err != nil {
			t.Fatalf("w2 change_report: %v", err)
		}
		if !restrict {
			admitPercept(w2Gate, w2Inst, w2Mind.next(t))
			crToW2++
		}
	}

	if restrict {
		// Mutation check: if maybeEmitChangeReport's guard were ever
		// dropped (or inverted), a change_report would actually have
		// been written to each of these three connections above, and
		// this would find it waiting instead of finding nothing —
		// turning red exactly when the restriction stops holding.
		actorMind.expectNone(t)
		w1Mind.expectNone(t)
		w2Mind.expectNone(t)
	}

	return floodResult{
		memoryCount: totalMemories(actorInst) + totalMemories(w1Inst) + totalMemories(w2Inst),
		crToActor:   crToActor,
		crToWitness: crToW1 + crToW2,
	}
}

// TestH6_ChangeReportFlood_RestrictionCutsMemoryLoad is H-6 (§10.4): the
// change_report delivery restriction (§4.10's MUST NOT) is what keeps the
// bookkeeping channel from drowning the experiential one. What ships
// without it: the measured 75% bug — three quarters of a mind's memory as
// third-person narration of its own acts and its own sightings, shipped
// again.
func TestH6_ChangeReportFlood_RestrictionCutsMemoryLoad(t *testing.T) {
	const n = 10

	restricted := runFloodScenario(t, true, n)
	flooded := runFloodScenario(t, false, n)

	if restricted.crToActor != 0 {
		t.Fatalf("H-6: restricted run delivered %d change_report(s) to the actor; §4.10 MUST NOT", restricted.crToActor)
	}
	if restricted.crToWitness != 0 {
		t.Fatalf("H-6: restricted run delivered %d change_report(s) to witnesses; §4.10 MUST NOT", restricted.crToWitness)
	}
	if flooded.memoryCount <= 3*restricted.memoryCount {
		t.Fatalf("H-6: flooded.memory_count (%d) must exceed 3x restricted.memory_count (%d) — ratio=%.2fx, want >3x",
			flooded.memoryCount, restricted.memoryCount, float64(flooded.memoryCount)/float64(restricted.memoryCount))
	}

	ratio := float64(flooded.memoryCount) / float64(restricted.memoryCount)
	t.Logf("H-6 flood ratio: flooded=%d restricted=%d ratio=%.2fx (rule requires >3x)", flooded.memoryCount, restricted.memoryCount, ratio)
}
