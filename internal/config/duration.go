package config

import (
	"encoding/json"
	"fmt"
	"time"
)

func (d *Duration) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*d = 0
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("duration must be string: %w", err)
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", s, err)
	}
	*d = Duration(parsed)
	return nil
}
