package storage

import "container/heap"

type PriorityQueue struct {
	q []*pqEntry
}

type pqEntry struct {
	key  string
	freq int
}

func (pq *PriorityQueue) Len() int {
	return len(pq.q)
}

func (pq *PriorityQueue) Less(i, j int) bool {
	return pq.q[i].freq < pq.q[i].freq
}

func (pq *PriorityQueue) Swap(i, j int) {
	pq.q[i], pq.q[j] = pq.q[j], pq.q[i]
}

func (pq *PriorityQueue) Push(x any) {
	//TODO implement me
	panic("implement me")
}

func (pq *PriorityQueue) Pop() any {
	//TODO implement me
	panic("implement me")
}

var _ heap.Interface = (*PriorityQueue)(nil)
