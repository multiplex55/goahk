package window

import "testing"

func TestMatchHelpers_TitleExactAndContainsCaseInsensitive(t *testing.T) {
	windows := []Info{{Title: "Visual Studio Code"}}

	exact, err := Filter(windows, MatchTitleExact("visual studio code"))
	if err != nil {
		t.Fatalf("Filter exact error = %v", err)
	}
	if len(exact) != 1 {
		t.Fatalf("exact matches = %d, want 1", len(exact))
	}

	contains, err := Filter(windows, MatchTitleContains("STUDIO"))
	if err != nil {
		t.Fatalf("Filter contains error = %v", err)
	}
	if len(contains) != 1 {
		t.Fatalf("contains matches = %d, want 1", len(contains))
	}
}

func TestMatchHelpers_ClassAndExeCaseInsensitive(t *testing.T) {
	windows := []Info{{Class: "Chrome_WidgetWin_1", Exe: "Code.exe"}}

	classMatches, err := Filter(windows, MatchClass("chrome_widgetwin_1"))
	if err != nil {
		t.Fatalf("Filter class error = %v", err)
	}
	if len(classMatches) != 1 {
		t.Fatalf("class matches = %d, want 1", len(classMatches))
	}

	exeMatches, err := Filter(windows, MatchExe("code.EXE"))
	if err != nil {
		t.Fatalf("Filter exe error = %v", err)
	}
	if len(exeMatches) != 1 {
		t.Fatalf("exe matches = %d, want 1", len(exeMatches))
	}
}
