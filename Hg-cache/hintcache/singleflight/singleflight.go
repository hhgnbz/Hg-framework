package singleflight

import "sync"

// call 代表正在进行中，或已经结束的请求。使用 sync.WaitGroup 锁避免重入。
// 并发协程之间不需要消息传递，非常适合 sync.WaitGroup。
// wg.Add(1) 锁加1。
// wg.Wait() 阻塞，直到锁被释放。
// wg.Done() 锁减1。
type call struct {
	wg  sync.WaitGroup
	val any
	err error
}

// Group 是 singleflight 的主数据结构，管理不同 key 的请求(call)。
type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

// Do 接收 2 个参数，第一个参数是 key，第二个参数是一个函数 fn
// Do 的作用是，针对相同的 key，无论 Do 被调用多少次，函数 fn 都只会被调用一次，等待 fn 调用结束了，返回返回值或错误。
func (g *Group) Do(key string, fn func() (any, error)) (any, error) {
	g.mu.Lock()
	// 懒加载
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	// 有请求正在进行
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		// 等待进行中的请求
		c.wg.Wait()
		// 请求完成，直接返回
		return c.val, c.err
	}
	// 多并发中首个进行到这里的请求
	c := new(call)
	// 阻塞后续请求
	c.wg.Add(1)
	// 先填入新的call，后续请求在code#27行查询到ok = true，直接进入等待取值
	g.m[key] = c
	g.mu.Unlock()
	// 调用fn请求节点key对应数据
	c.val, c.err = fn()
	// 首个请求完成，其他同个key的异步请求可进行直接取值
	c.wg.Done()
	// singleflight不承载存值功能，只做高并发下的快速返回取值
	// 当并发取值结束，将map中数据清除
	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()
	// 首次的结果返回
	return c.val, c.err
}
