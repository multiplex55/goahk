package inspect

import "sync"

type nodeChildrenCache struct {
	mu               sync.RWMutex
	selectedWindowID string
	entries          map[string][]TreeNodeDTO
}

func newNodeChildrenCache() *nodeChildrenCache {
	return &nodeChildrenCache{entries: map[string][]TreeNodeDTO{}}
}

func (c *nodeChildrenCache) key(windowID, nodeID string) string {
	return windowID + "::" + nodeID
}

func (c *nodeChildrenCache) window() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.selectedWindowID
}

func (c *nodeChildrenCache) setSelectedWindow(windowID string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.selectedWindowID == windowID {
		return false
	}
	c.selectedWindowID = windowID
	c.entries = map[string][]TreeNodeDTO{}
	return true
}

func (c *nodeChildrenCache) get(windowID, nodeID string) ([]TreeNodeDTO, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.selectedWindowID != windowID {
		return nil, false
	}
	children, ok := c.entries[c.key(windowID, nodeID)]
	if !ok {
		return nil, false
	}
	return append([]TreeNodeDTO(nil), children...), true
}

func (c *nodeChildrenCache) put(windowID, nodeID string, children []TreeNodeDTO) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.selectedWindowID != windowID {
		return
	}
	c.entries[c.key(windowID, nodeID)] = append([]TreeNodeDTO(nil), children...)
}

func (c *nodeChildrenCache) invalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = map[string][]TreeNodeDTO{}
}

func (c *nodeChildrenCache) invalidateNode(windowID, nodeID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, c.key(windowID, nodeID))
}
