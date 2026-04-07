package tray

import "fmt"

const (
	CommandShowStatus = "show_status"
	CommandOpenLogs   = "open_logs"
	CommandReload     = "reload_config"
	CommandExit       = "exit"
)

type StatusProvider interface {
	StatusText() string
}

type Menu struct {
	Status     StatusProvider
	OnOpenLogs func() error
	OnReload   func() error
	OnExit     func() error
}

func (m Menu) Execute(command string) (string, error) {
	switch command {
	case CommandShowStatus:
		if m.Status == nil {
			return "status unavailable", nil
		}
		return m.Status.StatusText(), nil
	case CommandOpenLogs:
		if m.OnOpenLogs == nil {
			return "", nil
		}
		return "", m.OnOpenLogs()
	case CommandReload:
		if m.OnReload == nil {
			return "", nil
		}
		return "", m.OnReload()
	case CommandExit:
		if m.OnExit == nil {
			return "", nil
		}
		return "", m.OnExit()
	default:
		return "", fmt.Errorf("unknown tray command %q", command)
	}
}
