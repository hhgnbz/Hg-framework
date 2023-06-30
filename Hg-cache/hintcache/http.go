package hintcache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

const basePath = "/_hint/"

// HTTPPool implements PeerPicker for a pool of HTTP peers.
type HTTPPool struct {
	// baseUrl e.g. http://localhost:8080/
	curBaseUrl  string
	curBasePath string
}

// NewHTTPPool initializes an HTTP pool of peers.
func NewHTTPPool(baseUrl string) *HTTPPool {
	return &HTTPPool{
		curBaseUrl:  baseUrl,
		curBasePath: basePath,
	}
}

// Log info with server name
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.curBaseUrl, fmt.Sprintf(format, v...))
}

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
