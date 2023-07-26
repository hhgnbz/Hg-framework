package registry

import (
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// 注册中心
// 1. 服务端启动后，向注册中心发送注册消息，注册中心得知该服务已经启动，处于可用状态。一般来说，服务端还需要定期向注册中心发送心跳，证明自己还活着。
// 2. 客户端向注册中心询问，当前哪些服务是可用的，注册中心将可用的服务列表返回客户端。
// 3. 客户端根据注册中心得到的服务列表，选择其中一个发起调用。

// HintRegistry 结构体，默认超时时间设置为 5 min，也就是说，任何注册的服务超过 5 min，即视为不可用状态。
type HintRegistry struct {
	timeout time.Duration
	mu      sync.Mutex
	servers map[string]*ServerItem
}

type ServerItem struct {
	Addr  string
	start time.Time
}

const (
	defaultPath    = "/_geerpc_/registry"
	defaultTimeout = time.Minute * 5
)

// NewHintRegistry 创建指定超时时间的注册中心，timeout == 0 为不设定超时时间
func NewHintRegistry(timeout time.Duration) *HintRegistry {
	return &HintRegistry{
		timeout: timeout,
		servers: make(map[string]*ServerItem),
	}
}

var DefaultHintRegistry = NewHintRegistry(defaultTimeout)

// putServer：添加服务实例，如果服务已经存在，则更新 start。
func (hr *HintRegistry) putServer(addr string) {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	if si, ok := hr.servers[addr]; ok {
		si.start = time.Now()
	} else {
		hr.servers[addr] = &ServerItem{Addr: addr, start: time.Now()}
	}
}

// aliveServers：返回可用的服务列表，如果存在超时的服务，则删除。
func (hr *HintRegistry) aliveServers() []string {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	res := make([]string, 0)
	for addr, s := range hr.servers {
		if hr.timeout == 0 || s.start.Add(hr.timeout).After(time.Now()) {
			// 服务可用
			res = append(res, s.Addr)
		} else {
			delete(hr.servers, addr)
		}
	}
	sort.Strings(res)
	return res
}

// 为了实现上的简单，HintRegistry 采用 HTTP 协议提供服务，且所有的有用信息都承载在 HTTP Header 中。
// Get：返回所有可用的服务列表，通过自定义字段 X-Hintrpc-Servers 承载。
// Post：添加服务实例或发送心跳，通过自定义字段 X-Hintrpc-Server 承载。
func (hr *HintRegistry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		w.Header().Set("X-Hintrpc-Servers", strings.Join(hr.aliveServers(), ","))
	case "POST":
		addr := req.Header.Get("X-Hintrpc-Server")
		if addr == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		hr.putServer(addr)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (hr *HintRegistry) HandleHTTP(registryPath string) {
	http.Handle(registryPath, hr)
	log.Println("rpc registry path:", registryPath)
}

// HandleHttp 默认参数调用
func HandleHttp() {
	DefaultHintRegistry.HandleHTTP(defaultPath)
}

// Heartbeat 便于服务启动时定时向注册中心发送心跳，默认周期比注册中心设置的过期时间少 1 min
func Heartbeat(registry, addr string, duration time.Duration) {
	// 保证在服务被移除前，能正常将心跳信息送达
	if duration == 0 {
		duration = defaultTimeout - time.Duration(1)*time.Minute
	}
	err := sendHeartbeat(registry, addr)
	go func() {
		t := time.NewTicker(duration)
		for err == nil {
			<-t.C
			err = sendHeartbeat(registry, addr)
		}
	}()
}

func sendHeartbeat(registry, addr string) error {
	log.Println(addr, "send heart beat to registry", registry)
	httpClient := &http.Client{}
	req, _ := http.NewRequest("POST", registry, nil)
	req.Header.Set("X-Hintrpc-Server", addr)
	if _, err := httpClient.Do(req); err != nil {
		log.Println("rpc server: heart beat err:", err)
		return err
	}
	return nil
}
