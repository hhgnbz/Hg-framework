package storage

import (
	"container/heap"
	"fmt"
	"testing"
)

func TestPriorityQueue(t *testing.T) {
	pq := PriorityQueue(make([]*pqEntry, 0))
	for i := 10; i >= 0; i-- {
		heap.Push(&pq, &pqEntry{0, &entry{}, i})
	}
	for pq.Len() != 0 {
		e := heap.Pop(&pq).(*pqEntry)
		fmt.Println(e.freq)
	}
}
