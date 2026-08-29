package main

import (
	"strings"
	"testing"

	"kithcraft/mind/llm"
)

// TestReport_ZeroCallPathEmitsZeroed is spec.md's zero-call acceptance
// scenario (card AC #5): a fresh runtime with no Client and no bodies
// still produces a report listing every class zeroed, not omitted.
func TestReport_ZeroCallPathEmitsZeroed(t *testing.T) {
	rt, err := NewRuntime(t.TempDir())
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}
	defer rt.Close()

	text := rt.reportText()
	for _, class := range []llm.Class{llm.E1, llm.E2, llm.E3, llm.E4, llm.E5, llm.E6} {
		want := string(class) + ": calls=0"
		if !strings.Contains(text, want) {
			t.Errorf("report missing zeroed row %q:\n%s", want, text)
		}
	}
	if !strings.Contains(text, "no bodies seen this session") {
		t.Errorf("report should note no bodies seen:\n%s", text)
	}
}

// TestReport_IncludesAdmittedInstrumentCounts proves the E6-input
// instrument's per-villager-day counts reach the report once a body has
// admitted at least one percept.
func TestReport_IncludesAdmittedInstrumentCounts(t *testing.T) {
	rt, err := NewRuntime(t.TempDir())
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}
	defer rt.Close()

	rt.HandlePercept(nil, "b-report", map[string]any{
		"world_time": int64(100),
		"payload":    notablePercept("p-report"),
	})

	text := rt.reportText()
	if !strings.Contains(text, "b-report day 0: 1 admitted") {
		t.Errorf("report missing admitted count for b-report:\n%s", text)
	}
}
