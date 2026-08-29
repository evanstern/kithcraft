// Package deliberate (this file): reading the declared verb vocabulary
// off a session manifest (body-protocol-v0.md §6.2's
// capabilities.verbs) — the only place this package learns what verbs
// exist. No verb name is ever hardcoded anywhere else in mind/deliberate
// (spec.md FR-002, card AC #2): a Loop is handed exactly this map and
// composes through seam.Pending, which refuses anything not in it (V-4).
package deliberate

// ManifestVerbs extracts {verb: declared} from a session_open payload's
// capabilities.verbs list (§6.2's shape: a list of {verb, targets}
// entries, one per declared verb). Target shapes are not needed here — seam.Pending only
// gates on verb name (V-4); Tokens (tokens.go) gates targets separately,
// against what the mind was actually shown, not against the manifest's
// declared target *types*.
func ManifestVerbs(capabilities map[string]any) map[string]bool {
	verbs := map[string]bool{}
	list, _ := capabilities["verbs"].([]any)
	for _, v := range list {
		entry, ok := v.(map[string]any)
		if !ok {
			continue
		}
		if verb, ok := entry["verb"].(string); ok && verb != "" {
			verbs[verb] = true
		}
	}
	return verbs
}
