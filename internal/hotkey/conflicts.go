package hotkey

import "fmt"

type Binding struct {
	ID    string
	Raw   string
	Chord Chord
}

func ParseBinding(id, hotkey string) (Binding, error) {
	chord, err := Parse(hotkey)
	if err != nil {
		return Binding{}, fmt.Errorf("binding %q: %w", id, err)
	}
	return Binding{ID: id, Raw: hotkey, Chord: chord}, nil
}

func DetectConflicts(bindings []Binding) error {
	seen := make(map[string]string, len(bindings))
	for _, b := range bindings {
		norm := b.Chord.String()
		if existingID, exists := seen[norm]; exists {
			return fmt.Errorf("hotkey conflict: %q and %q both map to %s", existingID, b.ID, norm)
		}
		seen[norm] = b.ID
	}
	return nil
}
