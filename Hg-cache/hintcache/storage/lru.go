package storage

import (
	"container/list"
)

const (
	LRUStrategy = "lru"
)

type entry struct {
	key string
	val Value
}

type LRUCache struct {
	maxBytes  int64
	usedBytes int64
	ll        *list.List
	cache     map[string]*list.Element // key is string type,value is pointer of list element
	OnEvicted func(key string, val Value)
}

// NewLruCache is the constructor of LRUCache
func NewLruCache(maxByte int64, onEvicted func(string, Value)) *LRUCache {
	return &LRUCache{
		maxBytes:  maxByte,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Get search key's value
// 从字典中找到对应的双向链表的节点
// 将该节点移动到队尾
// 如果键对应的链表节点存在，则将对应节点移动到队尾，并返回查找到的值。
// c.ll.MoveToFront(ele)，即将链表中的节点 ele 移动到队尾（双向链表作为队列，队首队尾是相对的，在这里约定 front 为队尾）
func (c *LRUCache) Get(key string) (val Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.val, true
	}
	return
}

// RemoveTheOldest removes the oldest item
// c.ll.Back() 取到队首节点，从链表中删除。
// delete(c.cache, kv.key)，从字典中 c.cache 删除该节点的映射关系。
// 更新当前所用的内存 c.usedBytes。
// 如果回调函数 OnEvicted 不为 nil，则调用回调函数。
func (c *LRUCache) RemoveTheOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.usedBytes -= int64(len(kv.key)) + int64(kv.val.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.val)
		}
	}
}

// Add adds a value to the cache.
// 如果键存在，则更新对应节点的值，并将该节点移到队尾。
// 不存在则是新增场景，首先队尾添加新节点 &entry{key, value}, 并字典中添加 key 和节点的映射关系。
// 更新 c.usedBytes，如果超过了设定的最大值 c.maxBytes，则移除最少访问的节点。
func (c *LRUCache) Add(key string, val Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		ele.Value = val
		kv := ele.Value.(*entry)
		c.usedBytes += int64(len(kv.key)) + int64(kv.val.Len())
	} else {
		insertELe := c.ll.PushFront(&entry{key: key, val: val})
		c.cache[key] = insertELe
		c.usedBytes += int64(len(key)) + int64(val.Len())
	}
	if c.maxBytes != 0 && c.usedBytes > c.maxBytes {
		c.RemoveTheOldest()
	}
}

func (c *LRUCache) Len() int {
	return c.ll.Len()
}
