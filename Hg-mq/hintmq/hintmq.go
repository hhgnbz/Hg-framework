package hintmq

import "sync"

type HintMQ struct {
	head   *chanQueue
	tail   *chanQueue
	mu     sync.Mutex
	buffer chan []byte
	idle   chan *chanQueue // 存放正在使用的队列
	size   uint
	caps   uint
}

func NewHintMQ() *HintMQ {
	hmq := &HintMQ{
		head:   newChanQueue(256),
		mu:     sync.Mutex{},
		buffer: make(chan []byte, 128),
		idle:   make(chan *chanQueue, 8),
		size:   0,
		caps:   256,
	}
	hmq.tail = hmq.head
	return hmq
}

func (hmq *HintMQ) extend() {
	hmq.mu.Lock()
	defer hmq.mu.Unlock()
	if !hmq.tail.full() {
		return
	}
	select {
	case q := <-hmq.idle:
		hmq.tail.next = q
		hmq.tail = q
		hmq.caps += q.capacity()
	default:
		size := hmq.tail.capacity()
		if size <= 1024 {
			size *= 2
		} else {
			size += 2048
		}
		q := newChanQueue(size)
		hmq.tail.next = q
		hmq.tail = q
		hmq.caps += size
	}
}

func (hmq *HintMQ) reduce() bool {
	hmq.mu.Lock()
	defer hmq.mu.Unlock()
	if !hmq.head.empty() {
		return true
	}
	if hmq.head == hmq.tail {
		return false
	}

	q := hmq.head
	hmq.head = q.next
	hmq.caps -= q.capacity()
	select {
	case hmq.idle <- q:
	default:
	}
	return true
}

func (hmq *HintMQ) Write(message []byte) {
	if !hmq.tail.append(message) {
		hmq.extend()
		hmq.Write(message)
	} else {
		hmq.size += 1
	}
}

func (hmq *HintMQ) Read() <-chan []byte {
	select {
	case message := <-hmq.head.pop():
		hmq.buffer <- message
		hmq.size -= 1
		return hmq.buffer
	default:
		if hmq.reduce() {
			return hmq.Read()
		} else {
			return hmq.buffer
		}
	}
}
