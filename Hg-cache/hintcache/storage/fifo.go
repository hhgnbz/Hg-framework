package storage

import "container/list"

const (
	FIFOStrategy = "fifo"
)

type FIFOCache struct {
	maxBytes  int64
	usedBytes int64
	ll        *list.List
	cache     map[string]*list.Element // key is string type,value is pointer of list element
	OnEvicted func(key string, val Value)
}

// newFifoCache is the constructor of LRUCache
func newFifoCache(maxByte int64, onEvicted func(string, Value)) *FIFOCache {
	return &FIFOCache{
		maxBytes:  maxByte,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (fc *FIFOCache) Get(key string) (val Value, ok bool) {
	if ele, ok := fc.cache[key]; ok {
		kv := ele.Value.(*entry)
		return kv.val, true
	}
	return
}

func (fc *FIFOCache) Add(key string, val Value) {
	if ele, ok := fc.cache[key]; ok {
		fc.ll.MoveToFront(ele)
		ele.Value = val
		kv := ele.Value.(*entry)
		fc.usedBytes += int64(len(kv.key)) + int64(kv.val.Len())
	} else {
		insertELe := fc.ll.PushFront(&entry{key: key, val: val})
		fc.cache[key] = insertELe
		fc.usedBytes += int64(len(key)) + int64(val.Len())
	}
	if fc.maxBytes != 0 && fc.usedBytes > fc.maxBytes {
		fc.removeBack()
	}
}

func (fc *FIFOCache) removeBack() {
	deleteEle := fc.ll.Back()
	if deleteEle != nil {
		fc.ll.Remove(deleteEle)
		ele := deleteEle.Value.(*entry)
		delete(fc.cache, ele.key)
		fc.usedBytes -= int64(len(ele.key)) + int64(ele.val.Len())
		if fc.OnEvicted != nil {
			fc.OnEvicted(ele.key, ele.val)
		}
	}
}

func (fc *FIFOCache) Len() int {
	return fc.ll.Len()
}

var _ EntryInterface = (*FIFOCache)(nil)
