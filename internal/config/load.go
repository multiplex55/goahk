package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func LoadFile(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	return LoadReader(f)
}

func LoadReader(r io.Reader) (Config, error) {
	var cfg Config
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	if dec.More() {
		return Config{}, fmt.Errorf("decode config: multiple JSON values")
	}

	ApplyDefaults(&cfg)
	if err := Validate(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
