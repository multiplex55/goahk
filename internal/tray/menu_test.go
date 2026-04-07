package tray

import (
	"errors"
	"testing"
)

type stubStatus string

func (s stubStatus) StatusText() string { return string(s) }

func TestMenuExecute_RoutesCommands(t *testing.T) {
	called := map[string]int{}
	menu := Menu{
		Status:     stubStatus("ready"),
		OnOpenLogs: func() error { called[CommandOpenLogs]++; return nil },
		OnReload:   func() error { called[CommandReload]++; return nil },
		OnExit:     func() error { called[CommandExit]++; return nil },
	}

	status, err := menu.Execute(CommandShowStatus)
	if err != nil || status != "ready" {
		t.Fatalf("show status = (%q, %v)", status, err)
	}
	for _, command := range []string{CommandOpenLogs, CommandReload, CommandExit} {
		if _, err := menu.Execute(command); err != nil {
			t.Fatalf("execute %s: %v", command, err)
		}
	}
	for _, command := range []string{CommandOpenLogs, CommandReload, CommandExit} {
		if called[command] != 1 {
			t.Fatalf("command %s called %d times", command, called[command])
		}
	}
}

func TestMenuExecute_UnknownAndCallbackError(t *testing.T) {
	wantErr := errors.New("boom")
	menu := Menu{OnReload: func() error { return wantErr }}
	if _, err := menu.Execute(CommandReload); !errors.Is(err, wantErr) {
		t.Fatalf("reload error = %v, want %v", err, wantErr)
	}
	if _, err := menu.Execute("bad"); err == nil {
		t.Fatal("expected error for unknown command")
	}
}
