package storage

type entry struct {
	key string
	val Value
}

type EntryInterface interface {
	Get(key string) (val Value, ok bool)
	Add(key string, val Value)
}

func NewEntryInterface(cacheStrategy string, maxByte int64, onEvicted func(string, Value)) EntryInterface {
	switch cacheStrategy {
	case LRUStrategy:
		return NewLruCache(maxByte, onEvicted)
	case FIFOStrategy:
		return NewFifoCache(maxByte, onEvicted)
	default:
		return NewLruCache(maxByte, onEvicted)
	}
}
