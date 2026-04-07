package testutil

import "sync"

type LogEntry struct {
	Level  string
	Msg    string
	Fields map[string]any
}

type FakeLogger struct {
	mu      sync.Mutex
	Entries []LogEntry
}

func (f *FakeLogger) Info(msg string, fields map[string]any) {
	f.append("info", msg, fields)
}

func (f *FakeLogger) Error(msg string, fields map[string]any) {
	f.append("error", msg, fields)
}

func (f *FakeLogger) append(level, msg string, fields map[string]any) {
	f.mu.Lock()
	defer f.mu.Unlock()
	cpy := map[string]any{}
	for k, v := range fields {
		cpy[k] = v
	}
	f.Entries = append(f.Entries, LogEntry{Level: level, Msg: msg, Fields: cpy})
}
