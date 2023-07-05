package hintcache

import (
	"fmt"
	"hintcache/consistenthash"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	basePath        = "/_hint/"
	defaultReplicas = 50
)

// HTTPPool implements PeerPicker for a pool of HTTP peers.
type HTTPPool struct {
	// baseUrl e.g. http://localhost:8080/
	self        string
	curBasePath string
	mu          sync.Mutex
	peers       *consistenthash.Map
	httpGetters map[string]*httpGetter
}

// NewHTTPPool initializes an HTTP pool of peers.
func NewHTTPPool(baseUrl string) *HTTPPool {
	return &HTTPPool{
		self:        baseUrl,
		curBasePath: basePath,
	}
}

type httpGetter struct {
	baseUrl string
}

func (hg *httpGetter) Get(group string, key string) (val []byte, err error) {
	u := fmt.Sprintf(
		"%v%v/%v",
		hg.baseUrl,
		url.QueryEscape(group),
		url.QueryEscape(key),
	)
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	return bytes, nil
}

// 确保类型实现了这个接口 如果没有实现会报错
var _ PeerGetter = (*httpGetter)(nil)

// Log info with server name
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// Set updates the pool's list of peers.
// 实例化了一致性哈希算法，并且添加了传入的节点，并为每一个节点创建了一个 HTTP 客户端 httpGetter。
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseUrl: peer + p.curBasePath}
	}
}

// PeerGetter picks a peer according to key
// 包装了一致性哈希算法的 Get() 方法，根据具体的 key，选择节点，返回节点对应的 HTTP 客户端。
func (p *HTTPPool) PeerGetter(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

// 确保类型实现了这个接口 如果没有实现会报错
var _ PeerPicker = (*HTTPPool)(nil)

// ServeHTTP handle all http requests
// 1. 判断访问路径的前缀是否是 basePath，不是返回错误。
// 2. 约定访问路径格式为 /<basepath>/<groupname>/<key>，通过 groupname 得到 group 实例，再使用 group.Get(key) 获取缓存数据。
// 3. 最终使用 w.Write() 将缓存值作为 httpResponse 的 body 返回。
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.curBasePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	// /<basepath>/<groupname>/<key> required
	parts := strings.SplitN(r.URL.Path[len(p.curBasePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	groupName := parts[0]
	key := parts[1]
	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}
