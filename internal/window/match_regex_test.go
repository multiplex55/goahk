package window

import (
	"strings"
	"testing"
	"time"
)

func TestFilter_InvalidRegex(t *testing.T) {
	_, err := Filter([]Info{{Title: "x"}}, Matcher{TitleRegex: "("})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "invalid title regex") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMatchesWithTimeout(t *testing.T) {
	ok, err := MatchesWithTimeout(Info{Title: "abcdef"}, Matcher{TitleRegex: "abc"}, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected match")
	}

	_, err = MatchesWithTimeout(Info{Title: "abcdef"}, Matcher{TitleRegex: "("}, 50*time.Millisecond)
	if err == nil {
		t.Fatal("expected invalid regex error")
	}
}
