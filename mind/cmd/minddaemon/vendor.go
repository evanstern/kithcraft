// Command minddaemon (this file): TASK-0023 T001 — the live Vendor over
// seam.Conn, promoted from TASK-0016's evening_test.go wireVendor adapter
// (that doc comment is the contract mind/deliberate/loop.go's Vendor
// interface names, plan.md design decision 3): stamp the wire envelope
// around a Loop's composed intent payload and write it on the body's own
// live session. wireVendor reads conn/session/seq through Runtime under
// rt.mu at send time rather than capturing them at construction, so an
// intent sent after a reconnect goes out on the body's *current*
// connection rather than a stale one.
package main

import "fmt"

// wireVendor adapts one body's live session to deliberate.Vendor.
type wireVendor struct {
	rt   *Runtime
	body string
}

func (v *wireVendor) SendIntent(payload map[string]any) error {
	v.rt.mu.Lock()
	bs := v.rt.bodies[v.body]
	if bs == nil || bs.conn == nil {
		v.rt.mu.Unlock()
		return fmt.Errorf("minddaemon: no live session for body %q", v.body)
	}
	bs.mindSeq++
	conn, env := bs.conn, map[string]any{
		"protocol": "0.1", "message": "intent", "session": bs.session,
		"seq": bs.mindSeq, "body": v.body, "world_time": bs.lastWorldTime,
		"payload": payload,
	}
	v.rt.mu.Unlock()
	return conn.WriteMessage(env)
}
