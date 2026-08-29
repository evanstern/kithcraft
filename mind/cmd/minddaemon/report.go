// Command minddaemon (this file): TASK-0021 T005 — the session-end report
// (spec.md FR-005, card AC #5): M4's per-class call/token counters (llm.
// Accounting) and M2's E6-input instrument (admitted buffer size per
// villager-day), in readable text. Unconditional: a nil Client (rehearsal
// path, no ANTHROPIC_API_KEY) or a session with zero admitted percepts
// still produces a report with every row zeroed, never omitted — spec.md's
// zero-call acceptance scenario.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"kithcraft/mind/llm"
)

// classOrder is every class in registry order (llm.classes.go's own E1..E6
// comment order) — Accounting.Report omits a class that never ran (its own
// documented behavior), so the report fills every row in rather than
// mirroring that omission.
var classOrder = []llm.Class{llm.E1, llm.E2, llm.E3, llm.E4, llm.E5, llm.E6}

// Report writes the session-end report to w and appends the same text to
// <runDir>/session-report.log (creating it if absent). A file-append
// failure is reported on w rather than returned, matching this file's
// other handlers — a report that can't be filed is still worth printing.
func (rt *Runtime) Report(w io.Writer, runDir string) {
	text := rt.reportText()
	io.WriteString(w, text)

	path := filepath.Join(runDir, "session-report.log")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		fmt.Fprintf(w, "minddaemon: opening %s for the session report: %v\n", path, err)
		return
	}
	defer f.Close()
	if _, err := io.WriteString(f, text); err != nil {
		fmt.Fprintf(w, "minddaemon: writing %s: %v\n", path, err)
	}
}

func (rt *Runtime) reportText() string {
	var b []byte
	line := func(format string, args ...any) {
		b = append(b, []byte(fmt.Sprintf(format, args...)+"\n")...)
	}

	line("=== minddaemon session report — %s ===", time.Now().UTC().Format(time.RFC3339))

	line("-- per-class calls/tokens (M4 Accounting) --")
	var stats map[llm.Class]llm.Stats
	if rt.Client != nil {
		stats = rt.Client.Accounting().Report()
	}
	for _, class := range classOrder {
		s := stats[class] // zero value when absent (rehearsal path or a class that never ran)
		line("  %s: calls=%d cancelled=%d input_tokens=%d output_tokens=%d cache_read=%d cache_creation=%d",
			class, s.Calls, s.CancelledCalls, s.InputTokens, s.OutputTokens, s.CacheReadTokens, s.CacheCreationTokens)
	}

	line("-- E6-input admitted buffer size per villager-day (M2 Instrument) --")
	rt.mu.Lock()
	bodies := make([]string, 0, len(rt.bodies))
	for body := range rt.bodies {
		bodies = append(bodies, body)
	}
	sort.Strings(bodies)
	if len(bodies) == 0 {
		line("  (no bodies seen this session)")
	}
	for _, body := range bodies {
		days := rt.bodies[body].instrument.Report()
		if len(days) == 0 {
			line("  %s: (no admitted events)", body)
			continue
		}
		dayIdx := make([]int64, 0, len(days))
		for d := range days {
			dayIdx = append(dayIdx, d)
		}
		sort.Slice(dayIdx, func(i, j int) bool { return dayIdx[i] < dayIdx[j] })
		for _, d := range dayIdx {
			line("  %s day %d: %d admitted", body, d, days[d])
		}
	}
	rt.mu.Unlock()

	return string(b)
}
