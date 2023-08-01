package storage

import "fmt"

type entry struct {
	key string
	val Value
	//updateTime *time.Time
}

//func (e *entry) expired(duration time.Duration) bool {
//	if e.updateTime == nil {
//		return false
//	} else {
//		// updateTime + ttl < curTime -> false -> 未过期
//		return e.updateTime.Add(duration).Before(time.Now())
//	}
//}
//
//func (e *entry) updateEntryTTL() {
//	curTime := time.Now()
//	e.updateTime = &curTime
//}

type EntryInterface interface {
	Get(key string) (val Value, ok bool)
	Add(key string, val Value)
}

func NewEntryInterface(cacheStrategy string, maxByte int64, onEvicted func(string, Value)) EntryInterface {
	switch cacheStrategy {
	case LRUStrategy:
		fmt.Println("creating lru cache...")
		return newLruCache(maxByte, onEvicted)
	case FIFOStrategy:
		fmt.Println("creating fifo cache...")
		return newFifoCache(maxByte, onEvicted)
	case LFUStrategy:
		fmt.Println("creating lfu cache...")
		return newLfuCache(maxByte, onEvicted)
	default:
		fmt.Println("creating default(lru) cache...")
		return newLruCache(maxByte, onEvicted)
	}
}
