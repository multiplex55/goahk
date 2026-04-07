package window

import "testing"

func TestFilter_TitleContainsOperator(t *testing.T) {
	got, err := Filter([]Info{{Title: "Visual Studio Code"}}, Matcher{TitleContains: "studio"})
	if err != nil {
		t.Fatalf("Filter returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len=%d want 1", len(got))
	}
}

func TestFilter_TitleExactOperator(t *testing.T) {
	got, err := Filter([]Info{{Title: "Visual Studio Code"}}, Matcher{TitleExact: "visual studio code"})
	if err != nil {
		t.Fatalf("Filter returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len=%d want 1", len(got))
	}
}

func TestFilter_ExecutableNameOperator(t *testing.T) {
	got, err := Filter([]Info{{Exe: "Code.exe"}}, Matcher{ExeName: "code.EXE"})
	if err != nil {
		t.Fatalf("Filter returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len=%d want 1", len(got))
	}
}

func TestFilter_ClassNameOperator(t *testing.T) {
	got, err := Filter([]Info{{Class: "Chrome_WidgetWin_1"}}, Matcher{ClassName: "chrome_widgetwin_1"})
	if err != nil {
		t.Fatalf("Filter returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len=%d want 1", len(got))
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
