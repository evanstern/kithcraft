package persona

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// WriteOnce writes p to dir as a single 0444 JSON file and refuses — returns
// an error, never overwrites — if a persona for p.CastID already exists.
// os.O_EXCL makes the existence check and the write atomic (no stat-then-
// write race), which is this function's only reason to exist: it is the one
// write path to a persona file in the entire codebase (package doc comment).
func WriteOnce(dir string, p Persona) error {
	if p.CastID == "" {
		return fmt.Errorf("persona: WriteOnce: empty cast id")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("persona: WriteOnce: %w", err)
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("persona: WriteOnce: %w", err)
	}
	path := filePath(dir, p.CastID)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o444)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("persona: WriteOnce: %s already exists — genesis runs once, refusing to overwrite", path)
		}
		return fmt.Errorf("persona: WriteOnce: %w", err)
	}
	defer f.Close()
	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("persona: WriteOnce: %w", err)
	}
	return nil
}

// Load reads dir and binds a Persona to each of castIDs (FR-004): a missing
// file, or a file whose own CastID doesn't match the id it's read for, is a
// startup error — never a silent regenerate. Load performs no writes.
func Load(dir string, castIDs []string) (map[string]Persona, error) {
	out := make(map[string]Persona, len(castIDs))
	for _, id := range castIDs {
		path := filePath(dir, id)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("persona: Load: cast id %q: %w", id, err)
		}
		var p Persona
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("persona: Load: cast id %q: %w", id, err)
		}
		if p.CastID != id {
			return nil, fmt.Errorf("persona: Load: file at %s is bound to cast id %q, not %q", path, p.CastID, id)
		}
		out[id] = p
	}
	return out, nil
}

func filePath(dir, castID string) string {
	return filepath.Join(dir, castID+".json")
}
