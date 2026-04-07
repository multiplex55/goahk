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
