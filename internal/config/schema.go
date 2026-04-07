package config

import "time"

// Config is the root runtime configuration.
type Config struct {
	App            AppConfig              `json:"app"`
	Logging        LoggingConfig          `json:"logging"`
	Hotkeys        []HotkeyBinding        `json:"hotkeys"`
	Flows          []Flow                 `json:"flows,omitempty"`
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
	Flow   string `json:"flow,omitempty"`
	Steps  []Step `json:"steps,omitempty"`
}

type Step struct {
	Action string            `json:"action"`
	Params map[string]string `json:"params,omitempty"`
}

type Flow struct {
	ID      string     `json:"id"`
	Timeout Duration   `json:"timeout,omitempty"`
	Steps   []FlowStep `json:"steps"`
}

type FlowStep struct {
	Name      string            `json:"name,omitempty"`
	Action    string            `json:"action,omitempty"`
	Params    map[string]string `json:"params,omitempty"`
	Timeout   Duration          `json:"timeout,omitempty"`
	If        *FlowIf           `json:"if,omitempty"`
	WaitUntil *FlowWaitUntil    `json:"waitUntil,omitempty"`
	Repeat    *FlowRepeat       `json:"repeat,omitempty"`
}

type FlowIf struct {
	WindowMatches *string    `json:"windowMatches,omitempty"`
	ElementExists *string    `json:"elementExists,omitempty"`
	Then          []FlowStep `json:"then,omitempty"`
	Else          []FlowStep `json:"else,omitempty"`
}

type FlowWaitUntil struct {
	WindowMatches *string  `json:"windowMatches,omitempty"`
	ElementExists *string  `json:"elementExists,omitempty"`
	Timeout       Duration `json:"timeout,omitempty"`
	Interval      Duration `json:"interval,omitempty"`
}

type FlowRepeat struct {
	Times int        `json:"times"`
	Steps []FlowStep `json:"steps"`
}

type Duration time.Duration

func (d Duration) ToStd() time.Duration { return time.Duration(d) }

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
