package clipboard

import "sync"

type History struct {
	mu    sync.RWMutex
	buf   []string
	head  int
	size  int
	dedup bool
}

func NewHistory(capacity int, dedup bool) *History {
	if capacity < 0 {
		capacity = 0
	}
	return &History{buf: make([]string, capacity), dedup: dedup}
}

func (h *History) Push(value string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.buf) == 0 {
		return
	}
	if h.dedup && h.size > 0 {
		last := h.buf[(h.head-1+len(h.buf))%len(h.buf)]
		if last == value {
			return
		}
	}
	h.buf[h.head] = value
	h.head = (h.head + 1) % len(h.buf)
	if h.size < len(h.buf) {
		h.size++
	}
}

func (h *History) Values() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]string, 0, h.size)
	for i := 0; i < h.size; i++ {
		idx := (h.head - 1 - i + len(h.buf)) % len(h.buf)
		out = append(out, h.buf[idx])
	}
	return out
}
