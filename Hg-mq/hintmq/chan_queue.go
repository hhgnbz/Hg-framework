package hintmq

// chan 需要在创建时指定容量，不可扩展
// mq 理论上，不考虑实际内存以及硬盘空间情况下，可支持无限大的容量
// 使用 链表结构 将一个个指定容量的 chan 连接起来，实现可扩展性
// 链表头部 用于读取数据
// 尾部 用于写入数据

//                   mq
//                   |
//      -------------------------------
//      ↓                             ↓
//    head                           tail
//    chan1---->chan2----->chan3--->chan4
//     ↓                               ↓
//    read message                    write

// TODO 池化，复用空闲的chan

type chanQueue struct {
	channel chan []byte
	next    *chanQueue
}

func newChanQueue(size uint) *chanQueue {
	return &chanQueue{
		channel: make(chan []byte, size),
		next:    nil,
	}
}

func (cq *chanQueue) pop() <-chan []byte {
	return cq.channel
}
func (cq *chanQueue) append(b []byte) bool {
	select {
	case cq.channel <- b:
		return true
	default:
		return false
	}
}
func (cq *chanQueue) count() uint {
	return uint(len(cq.channel))
}
func (cq *chanQueue) capacity() uint {
	return uint(cap(cq.channel))
}
func (cq *chanQueue) empty() bool {
	return len(cq.channel) == 0
}
func (cq *chanQueue) full() bool {
	return len(cq.channel) == cap(cq.channel)
}
