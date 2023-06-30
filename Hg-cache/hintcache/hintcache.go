package hintcache

import (
	"fmt"
	"log"
	"sync"
)

//                                  是
// 接收 key --> 检查是否被缓存 -----> 返回缓存值 ⑴
//          |  否                        是
//          |-----> 是否应当从远程节点获取 -----> 与远程节点交互 --> 返回缓存值 ⑵
//                            |  否
//                            |-----> 调用回调函数，获取值并添加到缓存 --> 返回缓存值 ⑶

// 如果缓存不存在，应从数据源（文件，数据库等）获取数据并添加到缓存中。
// Q：框架内是否应该支持多数据源配置？
// A：不应该，一是数据源种类太多，无法一一列举。二是扩展性不好。从哪获取，如何获取，具体业务操作等已经跳脱出缓存框架应该实现的范畴，应该将功能决定权交给用户。
// Solution：设计一个回调方法，方便用户自行处理回调业务逻辑
// 定义一个函数类型 GetterFunc，并且实现接口 Getter 的方法，然后在这个方法中调用自己。
// 这是 Go 语言中将其他函数（参数返回值定义与 F 一致）转换为接口 A 的常用技巧。

// Getter loads data for key (every callback should implement Get method)
type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (gf GetterFunc) Get(key string) ([]byte, error) {
	return gf(key)
}

// A Group is a cache namespace and associated data loaded spread over
// 项目中的核心结构
type Group struct {
	name      string
	getter    Getter
	mainCache cache
}

var (
	mu sync.RWMutex
	// key : group name
	// value : pointer of named group
	groups = make(map[string]*Group)
)

// NewGroup create a new instance of Group
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil getter!")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	// group resister
	groups[name] = g
	return g
}

// GetGroup returns the named group previously created with NewGroup, or
// nil if there's no such group.
func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	g := groups[name]
	return g
}

// Get value for a key from cache
// 从 mainCache 中查找缓存，如果存在则返回缓存值。
// 缓存不存在，则调用 load 方法，load 调用 getLocally（分布式场景下会调用 getFromPeer 从其他节点获取）
// getLocally 调用用户回调函数获取源数据，并且将源数据添加到缓存 mainCache 中（通过 populateCache 方法）
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key shouldn't be blank")
	}
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[HintCache]key ", key, " hit")
		return v, nil
	}
	// 没命中缓存
	return g.load(key)
}

// createCache 当缓存未命中，在 远程节点查找击中 或 用户自定义Getter方法查找击中 后调用，将其放入缓存中
func (g *Group) createCache(key string, val ByteView) {
	g.mainCache.add(key, val)
}

// load 当缓存未命中，先尝试远程节点查找，后尝试用户自定义Getter方法查找
func (g *Group) load(key string) (ByteView, error) {
	return g.getLocally(key)
}

// getLocally 调用用户自定义Getter方法查找缓存
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	// Deep copy
	val := ByteView{b: cloneSlice(bytes)}
	// Create k-v data in cache
	g.createCache(key, val)
	return val, nil
}
