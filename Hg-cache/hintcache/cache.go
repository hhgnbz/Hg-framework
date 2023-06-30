package hintcache

import (
	"hintcache/storage"
	"sync"
)

type cache struct {
	mu         sync.Mutex
	lru        *storage.LRUCache
	cacheBytes int64
}

func (c *cache) add(key string, val ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Lazy Initialization
	if c.lru == nil {
		c.lru = storage.NewLruCache(c.cacheBytes, nil)
	}
	c.lru.Add(key, val)
}

func (c *cache) get(key string) (val ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}
