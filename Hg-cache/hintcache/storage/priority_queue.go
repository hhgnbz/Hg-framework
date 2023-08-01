package storage

import (
	"container/heap"
)

type PriorityQueue []*pqEntry

type pqEntry struct {
	index int
	entry *entry
	freq  int
}

func (pq PriorityQueue) Len() int {
	return len(pq)
}

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].freq < pq[j].freq
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index, pq[j].index = i, j
}

func (pq *PriorityQueue) Push(x any) {
	e := x.(*pqEntry)
	e.index = len(*pq)
	*pq = append(*pq, e)
}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	res := old[n-1]
	old[n-1] = nil
	newPq := old[:n-1]
	for i := 0; i < len(newPq); i++ {
		newPq[i].index = i
	}
	*pq = newPq
	return res
}

func (pqe *pqEntry) update() {
	pqe.freq++
}

var _ heap.Interface = (*PriorityQueue)(nil)
