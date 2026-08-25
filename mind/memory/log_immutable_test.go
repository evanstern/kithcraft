// Package memory_test (external, deliberately): card AC #1 / SC-004 must
// be proven from outside memory's package boundary, where an internal
// test's package-private field access would prove nothing.
package memory_test

import (
	"reflect"
	"testing"

	"kithcraft/mind/memory"
)

// TestEvent_ImmutableAtTypeLevel proves mutation is impossible by
// construction: memory.Event has zero exported fields, so there is no
// expression outside this package that assigns into one — a compile
// error, not a review discipline.
func TestEvent_ImmutableAtTypeLevel(t *testing.T) {
	typ := reflect.TypeOf(memory.Event{})
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if f.IsExported() {
			t.Fatalf("memory.Event field %q is exported: code outside this package could set it directly, defeating card AC #1", f.Name)
		}
	}
}

// TestEvent_NoMutatingMethods checks the other half: unexported fields
// don't help if a pointer-receiver method could still mutate an Event in
// place. The method set of *Event equals the method set of Event, so
// every method has a value receiver and can only ever operate on a copy.
func TestEvent_NoMutatingMethods(t *testing.T) {
	valTyp := reflect.TypeOf(memory.Event{})
	ptrTyp := reflect.TypeOf(&memory.Event{})
	if ptrTyp.NumMethod() != valTyp.NumMethod() {
		t.Fatalf("*memory.Event has %d methods but memory.Event has %d: a pointer-receiver method exists that a plain Event value cannot call — exactly where a setter would hide",
			ptrTyp.NumMethod(), valTyp.NumMethod())
	}
}
