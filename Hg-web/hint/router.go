package hint

import (
	"log"
	"net/http"
	"strings"
)

// 将路由相关的方法和结构提取出来，方便对 router 的功能进行增强
// 例如，提供动态路由的支持(trie树实现)
// 使用 roots 来存储每种请求方式的Trie树根节点。使用 handlers 存储每种请求方式的HandlerFunc。

// router struct
type router struct {
	roots    map[string]*trieNode   // roots key e.g. roots['GET'] roots['POST']
	handlers map[string]HandlerFunc // handlers key e.g. handlers['GET-/p/:lang/doc'], handlers['POST-/p/book']
}

// constructor of Router
func newRouter() *router {
	return &router{
		roots:    make(map[string]*trieNode),
		handlers: make(map[string]HandlerFunc),
	}
}

// Only one * is allowed
func parsePattern(pattern string) []string {
	ss := strings.Split(pattern, "/")
	parts := make([]string, 0)
	for _, item := range ss {
		if item != "" {
			parts = append(parts, item)
			if item[0] == '*' {
				break
			}
		}
	}
	return parts
}

// handle router
func (r *router) handle(c *Context) {
	n, params := r.getRoute(c.Method, c.Path)
	if n != nil {
		c.Params = params
		key := c.Method + "-" + n.pattern
		c.handlers = append(c.handlers, r.handlers[key])
	} else {
		c.handlers = append(c.handlers, func(c *Context) {
			c.String(http.StatusNotFound, "404 Not Found: %s\n", c.Path)
		})
	}
	c.Next()
}

// inside func for users to add router
// m -> http method(get/post)
// p -> full path(pattern)
// h -> handler func
// roots key e.g. roots['GET'] roots['POST']
// handlers key e.g. handlers['GET-/p/:lang/doc'], handlers['POST-/p/book']
func (r *router) addRouter(m string, p string, h HandlerFunc) {
	log.Printf("Route %4s - %s", m, p)
	parts := parsePattern(p)

	key := m + "-" + p
	_, ok := r.roots[m]
	if !ok {
		r.roots[m] = &trieNode{}
	}
	r.roots[m].insert(p, parts, 0)
	r.handlers[key] = h
}

// getRoute returns a map that key is the suffix of ":" or "*",value is the same part's pattern
// e.g. /p/go/doc -> /p/:lang/doc -> {lang: "go"}
// e.g. /static/css/hb.css -> /static/*filepath -> {filepath: "css/hb.css"}
func (r *router) getRoute(m string, p string) (*trieNode, map[string]string) {
	searchParts := parsePattern(p)
	params := make(map[string]string)
	root, ok := r.roots[m]
	if !ok {
		return nil, nil
	}
	n := root.search(searchParts, 0)
	if n != nil {
		parts := parsePattern(n.pattern)
		for index, part := range parts {
			if part[0] == ':' {
				params[part[1:]] = searchParts[index]
			}
			if part[0] == '*' && len(part) > 1 {
				params[part[1:]] = strings.Join(searchParts[index:], "/")
				break
			}
		}
		return n, params
	}
	return nil, nil
}
