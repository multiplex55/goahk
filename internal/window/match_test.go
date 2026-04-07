package window

import "testing"

func TestFilter_AllRuleTypes(t *testing.T) {
	windows := []Info{
		{Title: "Visual Studio Code", Class: "Chrome_WidgetWin_1", Exe: "Code.exe", Active: true},
		{Title: "Notepad", Class: "Notepad", Exe: "notepad.exe", Active: false},
	}

	tests := []struct {
		name    string
		matcher Matcher
		want    int
	}{
		{name: "exact title", matcher: Matcher{TitleExact: "visual studio code"}, want: 1},
		{name: "contains title", matcher: Matcher{TitleContains: "studio"}, want: 1},
		{name: "regex title", matcher: Matcher{TitleRegex: `^Visual\s+Studio`}, want: 1},
		{name: "class name", matcher: Matcher{ClassName: "notepad"}, want: 1},
		{name: "exe name", matcher: Matcher{ExeName: "code.EXE"}, want: 1},
		{name: "active only", matcher: Matcher{ActiveOnly: true}, want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Filter(windows, tt.matcher)
			if err != nil {
				t.Fatalf("Filter returned error: %v", err)
			}
			if len(got) != tt.want {
				t.Fatalf("len=%d want %d", len(got), tt.want)
			}
		})
	}
}

func TestFilter_CombinedFilters(t *testing.T) {
	windows := []Info{
		{Title: "Terminal - Dev", Class: "ConsoleWindowClass", Exe: "WindowsTerminal.exe", Active: true},
		{Title: "Terminal - Prod", Class: "ConsoleWindowClass", Exe: "WindowsTerminal.exe", Active: false},
	}

	got, err := Filter(windows, Matcher{TitleContains: "terminal", ClassName: "consolewindowclass", ExeName: "windowsterminal.exe", ActiveOnly: true})
	if err != nil {
		t.Fatalf("Filter returned error: %v", err)
	}
	if len(got) != 1 || got[0].Title != "Terminal - Dev" {
		t.Fatalf("unexpected matches: %#v", got)
	}
}

func TestFilter_TitlePrecedence(t *testing.T) {
	windows := []Info{{Title: "My Window"}}
	got, err := Filter(windows, Matcher{TitleExact: "missing", TitleContains: "Window", TitleRegex: ".*"})
	if err != nil {
		t.Fatalf("Filter returned error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected exact-title precedence to win, got %#v", got)
	}
}

func TestFilter_CaseSensitivityExpectations(t *testing.T) {
	windows := []Info{{Title: "My App"}}
	got, err := Filter(windows, Matcher{TitleRegex: "my app"})
	if err != nil {
		t.Fatalf("Filter returned error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected regex to be case-sensitive unless pattern opts in")
	}

	got, err = Filter(windows, Matcher{TitleContains: "my"})
	if err != nil {
		t.Fatalf("Filter returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected contains matching to be case-insensitive")
	}
}
