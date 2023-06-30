package hintcache

import (
	"fmt"
	"log"
	"testing"
)

var db = map[string]string{
	"hhg": "creator",
	"cow": "mou",
	"cat": "meow",
}

func TestGet(t *testing.T) {
	loadCounts := make(map[string]int, len(db))
	h := NewGroup("names", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				loadCounts[key] += 1
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
	h.mainCache.add("hhg", ByteView{b: []byte("creator")})
	// one key,search twice
	for k, v := range db {
		if view, err := h.Get(k); err != nil || view.String() != v {
			t.Fatal("failed to get value")
		} // load from callback function
		if _, err := h.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		} // cache hit
	}

	if view, err := h.Get("unknown"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	}
}
