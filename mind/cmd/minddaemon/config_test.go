package main

import "testing"

// TestEnvOr_ConfigNotConstant is T004's audit (card ACs #4/#6): the flag
// defaults read from the environment, not a baked-in literal — an env var
// override changes the value with no rebuild.
func TestEnvOr_ConfigNotConstant(t *testing.T) {
	if got := envOr("MINDDAEMON_TEST_UNSET", "fallback"); got != "fallback" {
		t.Errorf("unset env: got %q, want the default", got)
	}
	t.Setenv("MINDDAEMON_TEST_UNSET", "from-env")
	if got := envOr("MINDDAEMON_TEST_UNSET", "fallback"); got != "from-env" {
		t.Errorf("set env: got %q, want the env value, not the default", got)
	}
}

func TestEnvOrBool_ConfigNotConstant(t *testing.T) {
	if got := envOrBool("MINDDAEMON_TEST_UNSET_BOOL", true); got != true {
		t.Errorf("unset env: got %v, want the default", got)
	}
	t.Setenv("MINDDAEMON_TEST_UNSET_BOOL", "false")
	if got := envOrBool("MINDDAEMON_TEST_UNSET_BOOL", true); got != false {
		t.Errorf("set env: got %v, want false from the env, not the true default", got)
	}
	t.Setenv("MINDDAEMON_TEST_UNSET_BOOL", "not-a-bool")
	if got := envOrBool("MINDDAEMON_TEST_UNSET_BOOL", true); got != true {
		t.Errorf("unparseable env: got %v, want the default rather than a crash", got)
	}
}
