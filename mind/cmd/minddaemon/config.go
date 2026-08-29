// Command minddaemon (this file): TASK-0021 T004 — the daemon-side half of
// the config surface (spec.md FR-004, card ACs #4/#6). Every flag mirrors
// an env var of the same name (MINDDAEMON_ prefix) so a run dir/socket/
// genesis choice never needs a rebuild: the env var sets the flag's
// default, an explicit flag still wins — the standard Go CLI idiom, not a
// second parallel config path.
package main

import (
	"os"
	"strconv"
)

// envOr returns the env var's value if set, else def — the flag default a
// flag.String(...) call reads before flag.Parse() applies any explicit
// -flag override.
func envOr(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}

// envOrBool is envOr for a boolean flag default (MINDDAEMON_GENESIS): an
// unparseable value falls back to def rather than panicking, since a
// config typo should never crash startup before flag.Parse() even runs.
func envOrBool(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}
