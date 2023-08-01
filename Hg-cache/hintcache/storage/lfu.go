package storage

import "container/heap"

const (
	LFUStrategy = "lfu"
)

type LFUCache struct {
	maxBytes  int64
	usedBytes int64
	pq        *PriorityQueue
	cache     map[string]*pqEntry // key is string type,value is pointer of list element
	OnEvicted func(key string, val Value)
}

func newLfuCache(maxByte int64, onEvicted func(string, Value)) *LFUCache {
	pq := PriorityQueue(make([]*pqEntry, 0))
	return &LFUCache{
		maxBytes:  maxByte,
		pq:        &pq,
		cache:     make(map[string]*pqEntry),
		OnEvicted: onEvicted,
	}
}

func (c *LFUCache) Get(key string) (val Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		ele.update()
		heap.Fix(c.pq, ele.index)
		return ele.entry.val, true
	}
	return
}

func (c *LFUCache) Add(key string, val Value) {
	if ele, ok := c.cache[key]; ok {
		//更新value
		c.usedBytes += int64(val.Len()) - int64(ele.entry.val.Len())
		for c.maxBytes != 0 && c.maxBytes < c.usedBytes {
			c.removeMinFreq()
		}
		ele.entry.val = val
		ele.update()
		heap.Fix(c.pq, ele.index)
	} else {
		ele := &pqEntry{0, &entry{key, val}, 0}
		ele.update()
		// 先移除，再插入
		// 避免元素刚加入即移除
		// e.g. kv1 -> freq:3,kv2 -> freq:2,此时加入kv3 -> freq:1,先插入会导致kv3直接被移除,实际应该为kv2被移除
		c.usedBytes += int64(len(ele.entry.key)) + int64(ele.entry.val.Len())
		for c.maxBytes != 0 && c.maxBytes < c.usedBytes {
			c.removeMinFreq()
		}
		heap.Push(c.pq, ele)
		c.cache[key] = ele
	}
}

func (c *LFUCache) removeMinFreq() {
	e := heap.Pop(c.pq).(*pqEntry)
	if e != nil {
		delete(c.cache, e.entry.key)
		c.usedBytes -= int64(len(e.entry.key)) + int64(e.entry.val.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(e.entry.key, e.entry.val)
		}
	}
}

func (c *LFUCache) Len() int {
	return c.pq.Len()
}

var _ EntryInterface = (*LFUCache)(nil)
