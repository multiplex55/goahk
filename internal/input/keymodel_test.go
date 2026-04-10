package input

import (
	"errors"
	"reflect"
	"testing"
)

func TestBuildChordKeyInputs_CommonChords(t *testing.T) {
	tests := []struct {
		name string
		keys []string
		want []keyInput
	}{
		{
			name: "ctrl+c",
			keys: []string{"ctrl", "c"},
			want: []keyInput{{vk: vkControl}, {vk: 'C'}, {vk: 'C', flags: keyeventfKeyUp}, {vk: vkControl, flags: keyeventfKeyUp}},
		},
		{
			name: "alt+tab",
			keys: []string{"alt", "tab"},
			want: []keyInput{{vk: vkMenu}, {vk: 0x09}, {vk: 0x09, flags: keyeventfKeyUp}, {vk: vkMenu, flags: keyeventfKeyUp}},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := buildChordKeyInputs(tc.keys)
			if err != nil {
				t.Fatalf("buildChordKeyInputs err=%v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got=%v want=%v", got, tc.want)
			}
		})
	}
}

func TestBuildMouseButtonInputs_Sequences(t *testing.T) {
	got, err := buildMouseButtonInputs(MouseButtonLeft, "double")
	if err != nil {
		t.Fatalf("buildMouseButtonInputs err=%v", err)
	}
	want := []mouseInput{{flags: mouseeventfLeftDown}, {flags: mouseeventfLeftUp}, {flags: mouseeventfLeftDown}, {flags: mouseeventfLeftUp}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got=%v want=%v", got, want)
	}
}

func TestBuildChordKeyInputs_InvalidArgument(t *testing.T) {
	_, err := buildChordKeyInputs([]string{"ctrl", ""})
	if !errors.Is(err, ErrInvalidInputArgument) {
		t.Fatalf("expected ErrInvalidInputArgument got=%v", err)
	}
}

func TestBuildTextKeyInputs_WhitespaceMatchesLegacyNoop(t *testing.T) {
	if got := buildTextKeyInputs("   "); len(got) != 0 {
		t.Fatalf("expected no events, got %v", got)
	}
}
