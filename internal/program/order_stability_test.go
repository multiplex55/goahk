package program

import (
	"reflect"
	"testing"
)

func TestNormalizeBindingOrderStableAcrossMapPermutations(t *testing.T) {
	t.Parallel()

	input := map[string]BindingSpec{
		"three": {ID: "three", Hotkey: "ctrl+3", Steps: []StepSpec{{Action: "system.log"}}},
		"one":   {ID: "one", Hotkey: "ctrl+1", Steps: []StepSpec{{Action: "system.log"}}},
		"two":   {ID: "two", Hotkey: "ctrl+2", Steps: []StepSpec{{Action: "system.log"}}},
	}

	var expected []string
	for i := 0; i < 30; i++ {
		p := Program{Bindings: make([]BindingSpec, 0, len(input))}
		for _, b := range input {
			p.Bindings = append(p.Bindings, b)
		}
		normalized := Normalize(p)
		got := make([]string, 0, len(normalized.Bindings))
		for _, b := range normalized.Bindings {
			got = append(got, b.ID)
		}
		if i == 0 {
			expected = got
			continue
		}
		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("Normalize() order = %v, want %v", got, expected)
		}
	}
}
