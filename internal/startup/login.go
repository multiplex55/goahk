package startup

import (
	"fmt"
	"strings"
)

type Entry struct {
	Name    string
	Command string
}

func BuildInstallCommand(executablePath, configPath string) (Entry, error) {
	executablePath = strings.TrimSpace(executablePath)
	if executablePath == "" {
		return Entry{}, fmt.Errorf("executable path is required")
	}
	cmd := fmt.Sprintf("\"%s\" -config \"%s\"", executablePath, strings.TrimSpace(configPath))
	return Entry{Name: "goahk", Command: cmd}, nil
}

func BuildUninstallCommand() Entry {
	return Entry{Name: "goahk"}
}
