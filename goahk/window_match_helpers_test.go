package goahk

import "testing"

func TestMatchHelpersAndEncoding(t *testing.T) {
	matcher := MatchTitleExact("Editor")
	matcher.ClassName = MatchClass("Chrome_WidgetWin_1").ClassName
	matcher.ExeName = MatchExe("Code.exe").ExeName
	matcher.ActiveOnly = true

	got := encodeMatcher(matcher)
	want := "title_exact:Editor,class:Chrome_WidgetWin_1,exe:Code.exe,active:true"
	if got != want {
		t.Fatalf("encodeMatcher() = %q, want %q", got, want)
	}
}
