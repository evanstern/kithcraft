// Package persona_test (external, deliberately, mirroring
// mind/memory/beliefs_external_test.go): card AC #2's "no code path to
// attempt" must be proven from outside persona's package boundary — an
// internal test's access to unexported helpers would prove nothing about
// what an outside caller (the daemon, a debug command, a future consumer)
// can reach. specs/013-persona-genesis Phase 1 (T003).
package persona_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"

	"kithcraft/mind/persona"
)

// TestPersona_NoMethodsOnPersonaType is card AC #2: Persona is a plain data
// struct — reflecting over *persona.Persona proves no method (a hypothetical
// (*Persona).Save, say) exists to write a persona back to disk from the
// value itself.
func TestPersona_NoMethodsOnPersonaType(t *testing.T) {
	typ := reflect.TypeOf(&persona.Persona{})
	if n := typ.NumMethod(); n != 0 {
		var names []string
		for i := 0; i < n; i++ {
			names = append(names, typ.Method(i).Name)
		}
		t.Fatalf("*persona.Persona has %d exported method(s) %v: any method on the persona value is a potential write path outside files.go's WriteOnce", n, names)
	}
}

// TestPersona_ExportedFunctionSurface_IsExactlyLoadAndWriteOnce is card AC
// #2's "no code path to attempt": reflect cannot enumerate a package's
// top-level functions (there's no receiver to hang a reflect.Type off of),
// so — following this same feature's validate.go no-llm-import check
// (plan.md: "a test reading the source, same as S2's reflection lock, since
// Go has no per-file import visibility") — this test parses the package's
// non-test source and asserts its exported top-level function set is
// exactly {Genesis, Load, Validate, WriteOnce}. WriteOnce's own refusal is
// proven at runtime by TestWriteOnce_RefusesExistingFile below; Validate
// (Phase 2) is read-only over its Persona argument and has no write path of
// its own. Genesis (Phase 3) DOES write, but only by calling WriteOnce
// internally (genesis.go) — it opens no write path WriteOnce doesn't
// already guard, so it is not a second, independent way to overwrite an
// existing persona. This test proves there is no OTHER exported function
// that could.
func TestPersona_ExportedFunctionSurface_IsExactlyLoadAndWriteOnce(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	dir := filepath.Dir(thisFile)

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatalf("parser.ParseDir(%s): %v", dir, err)
	}
	pkg, ok := pkgs["persona"]
	if !ok {
		t.Fatalf("package %q not found parsing %s (found: %v)", "persona", dir, pkgNames(pkgs))
	}

	var funcs []string
	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Recv != nil || !fd.Name.IsExported() {
				continue
			}
			funcs = append(funcs, fd.Name.Name)
		}
	}
	sort.Strings(funcs)
	want := []string{"Genesis", "Load", "Validate", "WriteOnce"}
	if !reflect.DeepEqual(funcs, want) {
		t.Fatalf("persona package's exported top-level functions = %v, want exactly %v — any addition is a new candidate write path to an existing persona", funcs, want)
	}
}

func pkgNames(pkgs map[string]*ast.Package) []string {
	var names []string
	for name := range pkgs {
		names = append(names, name)
	}
	return names
}

// TestWriteOnce_ThenLoad_RebindsSameCastID is card AC #4's unit half: after
// a simulated restart (a fresh Load call over the same directory), the
// persona binds to the exact cast id it was written for.
func TestWriteOnce_ThenLoad_RebindsSameCastID(t *testing.T) {
	dir := t.TempDir()
	p := persona.Persona{
		CastID:            "Aldric",
		Name:              "Aldric",
		Values:            []string{"duty", "craftsmanship"},
		EndogenousDesires: []string{"master the forge"},
		Anchor:            "steady and exacting",
		DriftMarkers:      []string{"reckless", "sloppy"},
		Profession:        "armorer",
		BiomeVariant:      "plains",
	}
	if err := persona.WriteOnce(dir, p); err != nil {
		t.Fatalf("WriteOnce: %v", err)
	}

	got, err := persona.Load(dir, []string{"Aldric"})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !reflect.DeepEqual(got["Aldric"], p) {
		t.Fatalf("Load after restart = %#v, want the exact persona written %#v", got["Aldric"], p)
	}
}

// TestWriteOnce_RefusesExistingFile is FR-003: a second WriteOnce for a
// cast id that already has a persona is a refused no-op, never a
// regenerate — the file on disk, and its 0444 mode, must be unchanged.
func TestWriteOnce_RefusesExistingFile(t *testing.T) {
	dir := t.TempDir()
	p := persona.Persona{CastID: "Petra", Name: "Petra", Anchor: "patient", Profession: "farmer", BiomeVariant: "desert"}
	if err := persona.WriteOnce(dir, p); err != nil {
		t.Fatalf("first WriteOnce: %v", err)
	}

	changed := p
	changed.Anchor = "reckless and hasty"
	if err := persona.WriteOnce(dir, changed); err == nil {
		t.Fatal("second WriteOnce for the same cast id must refuse, not overwrite")
	}

	got, err := persona.Load(dir, []string{"Petra"})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got["Petra"].Anchor != "patient" {
		t.Fatalf("Anchor after refused overwrite = %q, want unchanged %q", got["Petra"].Anchor, "patient")
	}

	info, err := os.Stat(filepath.Join(dir, "Petra.json"))
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Mode().Perm() != 0o444 {
		t.Fatalf("persona file mode = %o, want 0444", info.Mode().Perm())
	}
}

// TestLoad_MissingOrUnknownCastID_IsError is FR-004: a cast id with no
// persona file — whether the whole set is unknown or just one entry among
// otherwise-known ids — is a startup error, never a silent regenerate.
func TestLoad_MissingOrUnknownCastID_IsError(t *testing.T) {
	dir := t.TempDir()
	p := persona.Persona{CastID: "Yenna", Name: "Yenna", Anchor: "watchful", Profession: "fisherman", BiomeVariant: "taiga"}
	if err := persona.WriteOnce(dir, p); err != nil {
		t.Fatalf("WriteOnce: %v", err)
	}

	tests := []struct {
		name    string
		castIDs []string
	}{
		{"missing file entirely", []string{"NoSuchVillager"}},
		{"mixed known and unknown", []string{"Yenna", "Ghost"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := persona.Load(dir, tt.castIDs); err == nil {
				t.Fatalf("Load(%v) must error, never silently regenerate", tt.castIDs)
			}
		})
	}
}
