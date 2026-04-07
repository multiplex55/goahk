package config

// DefaultConfig returns the baseline configuration values.
func DefaultConfig() Config {
	return Config{
		App: AppConfig{
			Name: "goahk",
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
		},
		Clipboard: ClipboardConfig{
			EnableHistory: false,
			HistorySize:   50,
		},
		Startup: StartupConfig{
			RunAtLogin: false,
		},
	}
}

// ApplyDefaults mutates cfg with default values for omitted fields.
func ApplyDefaults(cfg *Config) {
	def := DefaultConfig()

	if cfg.App.Name == "" {
		cfg.App.Name = def.App.Name
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = def.Logging.Level
	}
	if cfg.Logging.Format == "" {
		cfg.Logging.Format = def.Logging.Format
	}
	if cfg.Clipboard.HistorySize == 0 {
		cfg.Clipboard.HistorySize = def.Clipboard.HistorySize
	}
}
