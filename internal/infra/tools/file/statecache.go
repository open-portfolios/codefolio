package file

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type StateCache struct {
	mu      sync.Mutex
	entries map[string]*stateEntry
}

type stateEntry struct {
	Content string
	Mtime   int64
}

func NewStateCache() *StateCache {
	return &StateCache{
		entries: make(map[string]*stateEntry),
	}
}

func (c *StateCache) Record(filePath string, content string, mtime int64) {
	abs := normalizePath(filePath)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[abs] = &stateEntry{Content: content, Mtime: mtime}
}

func (c *StateCache) Check(filePath string) (bool, string) {
	abs := normalizePath(filePath)
	c.mu.Lock()
	entry, exists := c.entries[abs]
	c.mu.Unlock()

	if !exists {
		return false, fmt.Sprintf("Error: file has not been read yet. Read it first before editing.")
	}

	info, err := os.Stat(abs)
	if err != nil {
		return true, ""
	}
	currentMtime := info.ModTime().UnixMilli()
	if currentMtime > entry.Mtime {
		return false, fmt.Sprintf("Error: file has been modified since last read. Read it again before editing.")
	}

	return true, ""
}

func (c *StateCache) Update(filePath string, newContent string) {
	abs := normalizePath(filePath)
	info, err := os.Stat(abs)
	if err != nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[abs] = &stateEntry{
		Content: newContent,
		Mtime:   info.ModTime().UnixMilli(),
	}
}

func normalizePath(p string) string {
	abs, err := filepath.Abs(p)
	if err != nil {
		return p
	}
	return abs
}
