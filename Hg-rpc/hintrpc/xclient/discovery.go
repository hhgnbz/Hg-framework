package xclient

import (
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"
)

// SelectMode 代表不同的负载均衡策略，目前仅实现 Random 和 RoundRobin 两种策略。
type SelectMode int

const (
	Random SelectMode = iota
	RoundRobin
)

// Discovery 是一个接口类型，包含了服务发现所需要的最基本的接口。
type Discovery interface {
	Refresh() error                      // 从注册中心更新服务列表
	Update(servers []string) error       // 手动更新服务列表
	Get(mode SelectMode) (string, error) // 根据负载均衡策略，选择一个服务实例
	GetAll() ([]string, error)           // 返回所有的服务实例
}

// MultiServersDiscovery 不需要注册中心，服务列表由手工维护的服务发现的结构体
type MultiServersDiscovery struct {
	r       *rand.Rand
	index   int
	servers []string
	mu      sync.RWMutex
}

// Refresh 对于 MultiServersDiscovery 无意义
func (m *MultiServersDiscovery) Refresh() error {
	return nil
}

func (m *MultiServersDiscovery) Update(servers []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.servers = servers
	return nil
}

func (m *MultiServersDiscovery) Get(mode SelectMode) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := len(m.servers)
	if n == 0 {
		return "", errors.New("rpc discovery: no available server")
	}
	switch mode {
	case Random:
		return m.servers[m.r.Intn(n)], nil
	case RoundRobin:
		s := m.servers[m.index%n]
		m.index = (m.index + 1) % n
		return s, nil
	default:
		return "", errors.New("rpc discovery: select mode error")
	}
}

func (m *MultiServersDiscovery) GetAll() ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// copy
	// 防止用户访问到内部服务列表修改
	res := make([]string, len(m.servers), len(m.servers))
	copy(res, m.servers)
	return res, nil
}

func NewMultiServersDiscovery(servers []string) *MultiServersDiscovery {
	d := &MultiServersDiscovery{
		servers: servers,
		r:       rand.New(rand.NewSource(time.Now().UnixNano())), // 初始化时使用时间戳设定随机数种子，避免每次产生相同的随机数序列
	}
	d.index = d.r.Intn(math.MaxInt32 - 1) // index 记录 Round Robin 算法已经轮询到的位置，为了避免每次从 0 开始，初始化时随机设定一个值。
	return d
}

var _ Discovery = (*MultiServersDiscovery)(nil)
