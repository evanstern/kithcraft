package persona

import (
	"go/parser"
	"go/token"
	"testing"
)

// TestValidator_ImportsNoLLMCode is US3 AC #3 and FR-001/FR-005: the
// validator half of the persona firewall must be structurally incapable of
// issuing a model call. Go has no per-file import visibility (plan.md), so
// this test parses each validator source file directly and asserts its
// import list excludes kithcraft/mind/llm — the same source-read technique
// persona_external_test.go's exported-surface test uses.
func TestValidator_ImportsNoLLMCode(t *testing.T) {
	const forbidden = "kithcraft/mind/llm"
	files := []string{"validate.go", "moralizing.go"}
	fset := token.NewFileSet()
	for _, name := range files {
		t.Run(name, func(t *testing.T) {
			f, err := parser.ParseFile(fset, name, nil, parser.ImportsOnly)
			if err != nil {
				t.Fatalf("parser.ParseFile(%s): %v", name, err)
			}
			for _, imp := range f.Imports {
				if imp.Path.Value == `"`+forbidden+`"` {
					t.Fatalf("%s imports %s: the validator must never be able to issue a model call", name, forbidden)
				}
			}
		})
	}
}
