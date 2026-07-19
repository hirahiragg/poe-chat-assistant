package translation

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/hirahiragg/poe-chat-assistant/internal/chat"
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

func (c *Cache) Get(dir Direction, message string, context []chat.Message) (string, bool) {
	key := cacheKey(dir, message, context)
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.entries[key]
	return val, ok
}

func (c *Cache) Set(dir Direction, message string, context []chat.Message, translation string) {
	key := cacheKey(dir, message, context)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = translation
}

func cacheKey(dir Direction, message string, context []chat.Message) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s\x00%s", dir, message)
	for _, m := range context {
		fmt.Fprintf(h, "\x00%s\x00%s", m.Player, m.Body)
	}
	return hex.EncodeToString(h.Sum(nil))
}
