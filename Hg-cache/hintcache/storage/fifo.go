package storage

import "container/list"

const (
	FIFOStrategy = "fifo"
)

type FifoCache struct {
	maxBytes  int64
	usedBytes int64
	ll        *list.List
	cache     map[string]*list.Element // key is string type,value is pointer of list element
	OnEvicted func(key string, val Value)
}

// NewFifoCache is the constructor of LRUCache
func NewFifoCache(maxByte int64, onEvicted func(string, Value)) *FifoCache {
	return &FifoCache{
		maxBytes:  maxByte,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (fc *FifoCache) Get(key string) (val Value, ok bool) {
	//TODO implement me
	panic("implement me")
}

func (fc *FifoCache) Add(key string, val Value) {
	//TODO implement me
	panic("implement me")
}

var _ EntryInterface = (*FifoCache)(nil)
