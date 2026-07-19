package translation

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
)

type Cache struct {
	mu      sync.RWMutex
	entries map[string]string
}

func NewCache() *Cache {
	return &Cache{
		entries: make(map[string]string),
	}
}

func (c *Cache) Get(req Request) (string, bool) {
	key := cacheKey(req)
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.entries[key]
	return val, ok
}

func (c *Cache) Set(req Request, translation string) {
	key := cacheKey(req)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = translation
}

func cacheKey(req Request) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s\x00%s\x00%s", req.Direction, req.Message, req.TargetLang)
	for _, m := range req.Context {
		fmt.Fprintf(h, "\x00%s\x00%s", m.Player, m.Body)
	}
	return hex.EncodeToString(h.Sum(nil))
}
