package folders

import "testing"

func TestNormalizeAndDedupe_DedupeByPathAndKeepDiagnostics(t *testing.T) {
	in := []rawFolderInfo{
		{Path: `C:\\Work`, Title: "Work", PID: 100, HWND: "0xB"},
		{Path: `c:\\work`, Title: "Work duplicate", PID: 101, HWND: "0xA"},
		{Path: `C:\\Users`, Title: "Users", PID: 200, HWND: "0xC"},
		{Diagnostic: "unresolvable handle", HWND: "0xD"},
	}

	got := normalizeAndDedupe(in, true)
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	var hasUsers bool
	var hasDiagnostic bool
	for _, entry := range got {
		if entry.Path == `C:\\Users` {
			hasUsers = true
		}
		if entry.Diagnostic != "" {
			hasDiagnostic = true
		}
	}
	if !hasUsers {
		t.Fatalf("expected C:\\\\Users entry: %#v", got)
	}
	if !hasDiagnostic {
		t.Fatalf("expected diagnostic entry: %#v", got)
	}
}

func TestParsePowerShellFolderResults_MapsNativePayload(t *testing.T) {
	raw := []byte(`[{"path":"C:\\\\Temp","title":"Temp","pid":42,"hwnd":"0x2A"}]`)
	got, err := parsePowerShellFolderResults(raw)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].Path != `C:\\Temp` || got[0].PID != 42 || got[0].HWND != "0x2A" {
		t.Fatalf("entry = %#v", got[0])
	}
}
