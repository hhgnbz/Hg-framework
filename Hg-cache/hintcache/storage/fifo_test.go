package storage

import (
	"fmt"
	"reflect"
	"testing"
)

func TestFifoGet(t *testing.T) {
	fifo := NewFifoCache(int64(0), nil)
	fifo.Add("key1", String("1234"))
	if v, ok := fifo.Get("key1"); !ok || string(v.(String)) != "1234" {
		t.Fatalf("cache hit key1=1234 failed")
	}
	if _, ok := fifo.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}

func TestRemoveBack(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "k3"
	v1, v2, v3 := "value1", "value2", "v3"
	cap := len(k1 + k2 + v1 + v2)
	fifo := NewFifoCache(int64(cap), nil)
	fifo.Add(k1, String(v1))
	fifo.Add(k2, String(v2))
	fifo.Add(k3, String(v3))

	if _, ok := fifo.Get("key1"); ok || fifo.Len() != 2 {
		t.Fatalf("RemoveBack key1 failed")
	}
}

func TestFifoOnEvicted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}
	fifo := NewFifoCache(int64(10), callback)
	fifo.Add("key1", String("123456"))
	fifo.Add("k2", String("k2"))
	fifo.Add("k3", String("k3"))
	fifo.Add("k4", String("k4"))

	expect := []string{"key1", "k2"}

	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s", expect)
	}
}

func TestFifoGetPrint(t *testing.T) {
	fifo := NewFifoCache(int64(10), nil)
	fifo.Add("k1", String("11"))
	if val, ok := fifo.Get("k1"); ok {
		fmt.Println(val)
	}
	fifo.Add("k2", String("22"))
	if _, ok := fifo.Get("k1"); !ok {
		fmt.Println("k1 deleted")
	}
	if val, ok := fifo.Get("k2"); ok {
		fmt.Println(val)
	}
	fifo.Add("k3", String("33333333"))
	if _, ok := fifo.Get("k1"); !ok {
		fmt.Println("k1 deleted")
	}
	if val, ok := fifo.Get("k2"); ok {
		fmt.Println(val)
	}
	if val, ok := fifo.Get("k3"); ok {
		fmt.Println(val)
	}
}
