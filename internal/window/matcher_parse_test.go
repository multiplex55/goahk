package window

import "testing"

func TestParseMatcherString(t *testing.T) {
	got := ParseMatcherString("title:Code,class:Chrome_WidgetWin_1,exe:Code.exe,active:true")
	if got.TitleContains != "Code" {
		t.Fatalf("TitleContains = %q", got.TitleContains)
	}
	if got.ClassName != "Chrome_WidgetWin_1" {
		t.Fatalf("ClassName = %q", got.ClassName)
	}
	if got.ExeName != "Code.exe" {
		t.Fatalf("ExeName = %q", got.ExeName)
	}
	if !got.ActiveOnly {
		t.Fatal("ActiveOnly = false, want true")
	}
}

func TestParseMatcherString_DefaultFallback(t *testing.T) {
	got := ParseMatcherString("notepad")
	if got.TitleContains != "notepad" {
		t.Fatalf("TitleContains = %q", got.TitleContains)
	}
}
