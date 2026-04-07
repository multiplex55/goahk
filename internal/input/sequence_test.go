package input

import (
	"reflect"
	"testing"
)

func TestTokenizeSequence(t *testing.T) {
	tokens, err := TokenizeSequence("ctrl+c {enter} alt+shift+x")
	if err != nil {
		t.Fatalf("tokenize err: %v", err)
	}
	got := make([][]string, 0, len(tokens))
	for _, tok := range tokens {
		got = append(got, tok.Keys)
	}
	want := [][]string{{"ctrl", "c"}, {"enter"}, {"alt", "shift", "x"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("tokens=%v want=%v", got, want)
	}
}

func TestTokenizeSequence_Errors(t *testing.T) {
	cases := []string{"{", "{}", "ctrl++a"}
	for _, tc := range cases {
		if _, err := TokenizeSequence(tc); err == nil {
			t.Fatalf("expected error for %q", tc)
		}
	}
}
