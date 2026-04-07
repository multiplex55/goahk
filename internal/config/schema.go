package config

// Config is the root runtime configuration.
type Config struct {
	App            AppConfig              `json:"app"`
	Logging        LoggingConfig          `json:"logging"`
	Hotkeys        []HotkeyBinding        `json:"hotkeys"`
	WindowMatchers []WindowMatcher        `json:"windowMatchers"`
	UIASelectors   map[string]UIASelector `json:"uiaSelectors,omitempty"`
	Clipboard      ClipboardConfig        `json:"clipboard"`
	Startup        StartupConfig          `json:"startup"`
}

type AppConfig struct {
	Name string `json:"name,omitempty"`
}

type LoggingConfig struct {
	Level  string `json:"level,omitempty"`
	Format string `json:"format,omitempty"`
}

type HotkeyBinding struct {
	ID     string `json:"id"`
	Hotkey string `json:"hotkey"`
	Steps  []Step `json:"steps"`
}

type Step struct {
	Action string            `json:"action"`
	Params map[string]string `json:"params,omitempty"`
}

type WindowMatcher struct {
	Name          string `json:"name,omitempty"`
	TitleContains string `json:"titleContains,omitempty"`
	Exe           string `json:"exe,omitempty"`
}

type ClipboardConfig struct {
	EnableHistory bool `json:"enableHistory"`
	HistorySize   int  `json:"historySize,omitempty"`
}

type StartupConfig struct {
	RunAtLogin bool `json:"runAtLogin"`
}

type UIASelector struct {
	AutomationID string        `json:"automationId,omitempty"`
	Name         string        `json:"name,omitempty"`
	ControlType  string        `json:"controlType,omitempty"`
	Ancestors    []UIASelector `json:"ancestors,omitempty"`
}
