package hintcache

import (
	"hintcache/storage"
	"sync"
)

type cache struct {
	mu             sync.Mutex
	entryInterface storage.EntryInterface
	cacheStrategy  string
	cacheBytes     int64
}

func (c *cache) add(key string, val ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Lazy Initialization
	if c.entryInterface == nil {
		c.entryInterface = storage.NewEntryInterface(c.cacheStrategy, c.cacheBytes, nil)
	}
	c.entryInterface.Add(key, val)
}

func (c *cache) get(key string) (val ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.entryInterface == nil {
		return
	}
	if v, ok := c.entryInterface.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}
