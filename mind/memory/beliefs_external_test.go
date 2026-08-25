// Package memory_test (external, deliberately, mirroring log_immutable_test.go):
// card AC #2 and AC #7 must be proven from outside memory's package
// boundary — an internal test's package-private field access would prove
// nothing about what an outside caller (a vendor, the player, a debug
// command) can reach.
package memory_test

import (
	"reflect"
	"testing"

	"kithcraft/mind/memory"
)

// TestStore_NoExternalWritePathBeyondIngestAPI is card AC #7: memory.Store
// has zero exported fields, so nothing outside the package can reach the
// private map directly (PM-1) — Upsert and Retract, the mind's own
// ingest/consolidation entry points, are the only way in.
func TestStore_NoExternalWritePathBeyondIngestAPI(t *testing.T) {
	typ := reflect.TypeOf(memory.Store{})
	for i := 0; i < typ.NumField(); i++ {
		if f := typ.Field(i); f.IsExported() {
			t.Fatalf("memory.Store field %q is exported: code outside this package could write the belief map directly, defeating card AC #7", f.Name)
		}
	}
}

// TestStore_NoReadPathToVendorState is card AC #2: no exported method on
// *memory.Store takes or returns a type from a vendor-facing package
// (kithcraft/mind/seam or kithcraft/mind/seamtest) — the belief store is
// structurally distinct from, and has no read path into, any vendor
// resolution index.
func TestStore_NoReadPathToVendorState(t *testing.T) {
	typ := reflect.TypeOf(&memory.Store{})
	for i := 0; i < typ.NumMethod(); i++ {
		m := typ.Method(i)
		ft := m.Func.Type()
		for j := 0; j < ft.NumIn(); j++ {
			checkNoVendorPkg(t, m.Name, "parameter", ft.In(j))
		}
		for j := 0; j < ft.NumOut(); j++ {
			checkNoVendorPkg(t, m.Name, "return value", ft.Out(j))
		}
	}
}

func checkNoVendorPkg(t *testing.T, method, role string, typ reflect.Type) {
	t.Helper()
	for typ.Kind() == reflect.Ptr || typ.Kind() == reflect.Slice || typ.Kind() == reflect.Array || typ.Kind() == reflect.Map || typ.Kind() == reflect.Chan {
		typ = typ.Elem()
	}
	if pkg := typ.PkgPath(); pkg == "kithcraft/mind/seam" || pkg == "kithcraft/mind/seamtest" {
		t.Fatalf("Store.%s %s has type %v from vendor-facing package %q: the belief store must have no read path to vendor state (card AC #2)", method, role, typ, pkg)
	}
}
