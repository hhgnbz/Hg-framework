package storage

import (
	"fmt"
	"testing"
)

func TestLFUCache(t *testing.T) {
	lfu := newLfuCache(15, nil)
	lfu.Add("test1", String("1"))
	lfu.Add("test2", String("2"))
	fmt.Println(lfu.Get("test1"))
	fmt.Println(lfu.Get("test1"))
	fmt.Println(lfu.Get("test2"))
	lfu.Add("test3", String("3"))
	fmt.Println(lfu.Get("test1"))
	fmt.Println(lfu.Get("test2"))
	fmt.Println(lfu.Get("test3"))
}
