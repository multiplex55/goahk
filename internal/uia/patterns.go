package uia

import "sort"

// DiscoverPatterns converts raw pattern support flags into deterministic pattern names.
// It only performs discovery; behavior adapters can be added later.
func DiscoverPatterns(raw map[string]bool) []string {
	if len(raw) == 0 {
		return nil
	}
	out := make([]string, 0, len(raw))
	for name, supported := range raw {
		if supported {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}
